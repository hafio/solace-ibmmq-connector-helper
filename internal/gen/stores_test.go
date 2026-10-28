package gen

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/certfixture"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

const storePW = "store-pass-123"

// storeFiles is a ReadFile over an in-memory set of source files.
func storeFiles(files map[string][]byte) func(string) ([]byte, error) {
	return func(p string) ([]byte, error) {
		if b, ok := files[p]; ok {
			return b, nil
		}
		return nil, fs.ErrNotExist
	}
}

// storeEnv answers CA_PW with the store password, EMPTY with "", every other
// variable with "v", and nothing for NOPE.
func storeEnv(name string) (string, bool) {
	switch name {
	case "CA_PW":
		return storePW, true
	case "EMPTY":
		return "", true
	case "NOPE":
		return "", false
	}
	return "v", true
}

// createdCA is a created truststore listing files, its password in CA_PW.
func createdCA(files ...string) *spec.Store {
	c := &spec.StoreSources{}
	for _, f := range files {
		c.Certs.Items = append(c.Certs.Items, spec.CertSource{File: f})
	}
	return &spec.Store{PasswordEnv: "CA_PW", Create: c}
}

// BuildStore reads a created store's sources through the resolver, builds it
// with the password the environment holds, and hands the builder's findings
// back as env.yaml warnings.
func TestBuildStore(t *testing.T) {
	now := time.Now()
	ca := certfixture.CA(t, "Root")
	soon := certfixture.CA(t, "Soon", certfixture.Validity(now.Add(-time.Hour), now.Add(5*24*time.Hour)))
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(ca, soon)}), Now: func() time.Time { return now }}
	data, warns, err := BuildStore(StoreBuild{Name: "made-ca", Store: createdCA("./ca.pem"), Trusted: true}, res)
	if err != nil {
		t.Fatal(err)
	}
	if certs, err := pkcs12.DecodeTrustStore(data, storePW); err != nil || len(certs) != 2 {
		t.Errorf("decoded %d certificates, err %v; want both", len(certs), err)
	}
	if len(warns) != 1 || warns[0].File != fileEnv || !strings.Contains(warns[0].Msg, "expires on") {
		t.Errorf("warns = %v, want the expiring certificate", warns)
	}
}

// Every source reaches the builder named by its env.yaml position: a named
// certs entry, the key entry's cert, chain and key, a jks.
func TestBuildStoreReadsEverySourceByPosition(t *testing.T) {
	root := certfixture.CA(t, "Root")
	leaf := certfixture.Issue(t, root, "client")
	files := map[string][]byte{"./c.pem": certfixture.PEM(leaf), "./root.pem": certfixture.PEM(root), "./c.key": certfixture.KeyPKCS8(t, leaf)}
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(files)}
	key := &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Chain: []string{"./root.pem"}, Key: "./c.key"}}}}
	data, _, err := BuildStore(StoreBuild{Name: "client", Store: key, Private: true}, res)
	if err != nil {
		t.Fatal(err)
	}
	if _, gotLeaf, chain, err := pkcs12.DecodeChain(data, storePW); err != nil || gotLeaf.Subject.CommonName != "client" || len(chain) != 1 {
		t.Errorf("decoded leaf/chain wrong, err %v", err)
	}
	for _, c := range []struct {
		name  string
		store *spec.Store
		want  string
	}{
		{"named certs", &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{Certs: spec.CertSources{Named: true, Items: []spec.CertSource{{Name: "root", File: "./gone.pem"}}}}}, `reading tls.stores.s.create.certs.root "./gone.pem"`},
		{"key cert", &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./gone.pem", Key: "./c.key"}}}}, `reading tls.stores.s.create.keys[0].cert "./gone.pem"`},
		{"key", &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Key: "./gone.key"}}}}, `reading tls.stores.s.create.keys[0].key "./gone.key"`},
		{"chain", &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{Keys: []spec.KeySource{{Cert: "./c.pem", Chain: []string{"./gone.pem"}, Key: "./c.key"}}}}, `reading tls.stores.s.create.keys[0].chain[0] "./gone.pem"`},
		{"jks", &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{JKS: "./gone.jks"}}, `reading tls.stores.s.create.jks "./gone.jks"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := BuildStore(StoreBuild{Name: "s", Store: c.store}, res)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
		})
	}
}

func TestBuildStoreRejects(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	leaf := certfixture.Issue(t, ca, "client")
	files := map[string][]byte{
		"./ca.pem":    certfixture.PEM(ca),
		"./junk.pem":  []byte("hello"),
		"./key.jks":   certfixture.JKS(t, storePW, nil, []certfixture.Key{{Alias: "client", Chain: []certfixture.Cert{leaf}}}),
		"./trust.jks": certfixture.JKS(t, storePW, []certfixture.Trusted{{Alias: "ca", Cert: ca}}, nil),
	}
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(files)}
	withPW := func(env string, st *spec.Store) *spec.Store { st.PasswordEnv = env; return st }
	for _, c := range []struct {
		name string
		sb   StoreBuild
		res  Resolver
		want string
	}{
		{"an unset password variable", StoreBuild{Name: "s", Store: withPW("NOPE", createdCA("./ca.pem"))}, res, "tls.stores.s.password-env names NOPE, which is not set; export it before deploying"},
		{"an empty password variable", StoreBuild{Name: "s", Store: withPW("EMPTY", createdCA("./ca.pem"))}, res, "tls.stores.s.password-env names EMPTY, which is set but empty; a created store is encrypted with it"},
		{"no password", StoreBuild{Name: "s", Store: withPW("", createdCA("./ca.pem"))}, res, "tls.stores.s is created and encrypted with its own password, but sets none; add password-env"},
		{"no environment access", StoreBuild{Name: "s", Store: createdCA("./ca.pem")}, Resolver{ReadFile: res.ReadFile}, "tls.stores.s.password-env names CA_PW, but there is no environment access to read it"},
		{"an unexpanded variable", StoreBuild{Name: "s", Store: createdCA("${CERTS}/ca.pem")}, res, `tls.stores.s.create.certs[0] "${CERTS}/ca.pem" still holds an unexpanded ${...}`},
		{"no file access", StoreBuild{Name: "s", Store: createdCA("./ca.pem")}, Resolver{Env: storeEnv}, `cannot read tls.stores.s.create.certs[0] "./ca.pem" (no file access)`},
		{"a source the builder refuses", StoreBuild{Name: "s", Store: createdCA("./junk.pem")}, res, "tls.stores.s.create.certs[0] (./junk.pem) holds no certificate"},
		{"a JKS key selected as truststore", StoreBuild{Name: "s", Store: &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{JKS: "./key.jks"}}, Trusted: true}, res, "tls.stores.s is created from a JKS holding a key, so it holds no trusted certificates, but a connection selects it as a truststore"},
		{"a JKS truststore selected as keystore", StoreBuild{Name: "s", Store: &spec.Store{PasswordEnv: "CA_PW", Create: &spec.StoreSources{JKS: "./trust.jks"}}, Private: true}, res, "tls.stores.s is created from a JKS holding only trusted certificates, so it holds no private key, but a connection selects it as a keystore"},
		{"not a created store", StoreBuild{Name: "s", Store: &spec.Store{Existing: &spec.StoreExisting{PKCS12: "./t.p12"}}}, res, "tls.stores.s has no create: block to build from"},
	} {
		t.Run(c.name, func(t *testing.T) {
			data, _, err := BuildStore(c.sb, c.res)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("err = %v, want %q", err, c.want)
			}
			if data != nil {
				t.Errorf("a failed build returned %d bytes", len(data))
			}
		})
	}
}

// createdEnv is an env.yaml whose one store, ts, is created from ./ca.pem,
// followed by section.
func createdEnv(section string) Request {
	env := "image:\n  name: img\n  tag: v1\ntls:\n  stores:\n    ts:\n      password-env: CA_PW\n      create:\n        certs:\n          - ./ca.pem\n" + section
	return Request{Env: &File{Name: "env.yaml", Data: []byte(env)}, Workflows: synthTLSWorkflowFiles(1, "ts", "")}
}

// A created store travels as a secret on every platform: kubernetes builds it
// into the stores Secret it creates -- the same bytes on every run -- podman
// plans it as a build that deploy carries out, and docker as a compose secret
// whose base64 deploy puts in the environment, with a digest that stays put
// while the store does. A store kept in an operator's Secret is read from it on
// kubernetes, so nothing is built there; generate config references it at its
// mount path, like any store, and reads nothing.
func TestCreatedStoresOnEachPlatform(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	var reads []string
	res := Resolver{Env: storeEnv, Rand: fixedStatusRand, ReadFile: func(p string) ([]byte, error) {
		reads = append(reads, p)
		return storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(ca)})(p)
	}}

	kube := "kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      create: true\n"
	first, errs, _ := GenerateKubernetes(createdEnv(kube), res, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("kubernetes: %v", errs)
	}
	at := strings.Index(first, "\n  ts: ")
	if at < 0 {
		t.Fatalf("the stores Secret carries no ts key:\n%s", first)
	}
	enc := first[at+len("\n  ts: "):]
	enc = enc[:strings.IndexByte(enc, '\n')]
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatal(err)
	}
	if certs, err := pkcs12.DecodeTrustStore(raw, storePW); err != nil || len(certs) != 1 {
		t.Errorf("the Secret's ts is not the built truststore: %d certificates, err %v", len(certs), err)
	}
	if again, _, _ := GenerateKubernetes(createdEnv(kube), res, KubeOpts{}); again != first {
		t.Error("an unchanged created store must render the same manifest on every run")
	}

	reads = nil
	held := createdEnv(kube)
	held.Env.Data = []byte(strings.Replace(string(held.Env.Data), "      create:\n", "      existing:\n        secret:\n          name: my-tls\n      create:\n", 1))
	out, errs, _ := GenerateKubernetes(held, res, KubeOpts{})
	if len(errs) > 0 {
		t.Fatalf("kubernetes with the store in an operator's Secret: %v", errs)
	}
	if len(reads) != 0 {
		t.Errorf("a store kept in an operator's Secret is not built on kubernetes, but read %v", reads)
	}
	if strings.Contains(out, "  name: c-stores\n  namespace") || !strings.Contains(out, "- secret:\n                  name: my-tls\n") {
		t.Errorf("want no tool stores Secret and the operator's projected:\n%s", out)
	}

	plan, errs, _ := GeneratePodman(createdEnv("podman:\n  command: podman\n  name: c\n"), res)
	if len(errs) > 0 {
		t.Fatalf("podman: %v", errs)
	}
	var store *PodmanFile
	for i := range plan.Files {
		if plan.Files[i].StoreName == "c-tls-ts" {
			store = &plan.Files[i]
		}
	}
	if store == nil || store.Build == nil || store.Source != "" || store.Field != "tls.stores.ts.create" || !store.Build.Trusted {
		t.Fatalf("podman plan store = %+v, want a trusted build with no source", store)
	}
	kvs, _, err := ResolvePodmanFiles([]PodmanFile{*store}, res)
	if err != nil || len(kvs) != 1 {
		t.Fatalf("ResolvePodmanFiles: %v %v", kvs, err)
	}
	if _, err := pkcs12.DecodeTrustStore([]byte(kvs[0].Val), storePW); err != nil {
		t.Errorf("the podman secret is not the built truststore: %v", err)
	}

	dp, errs, _ := GenerateDocker(createdEnv("docker:\n  command: docker\n  name: c\n"), res)
	if len(errs) > 0 {
		t.Fatalf("docker: %v", errs)
	}
	if len(dp.Stores) != 1 || dp.Stores[0].Secret != "_GEN_TLS_TS_STORE" || !dp.Stores[0].Build.Trusted || dp.Image != "img:v1" || len(dp.Mounts) != 0 {
		t.Fatalf("docker plan = %+v, want one trusted created store, no bind mount", dp)
	}
	if !strings.Contains(dp.Compose, "      - source: _GEN_TLS_TS_STORE\n        target: /app/external/var/stores/ts\n") {
		t.Errorf("the compose does not carry the created store:\n%s", dp.Compose)
	}
	env, _, err := ResolveDockerStores(dp, res, 131072, "linux")
	if err != nil || len(env) != 2 || env[0].Key != "_GEN_TLS_TS_STORE" || env[1].Key != "_GEN_TLS_STORES_SHA256" || len(env[1].Val) != 64 {
		t.Fatalf("docker store environment = %v, err %v", env, err)
	}
	raw, err = base64.StdEncoding.DecodeString(env[0].Val)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pkcs12.DecodeTrustStore(raw, storePW); err != nil {
		t.Errorf("the docker secret is not the built truststore: %v", err)
	}
	if again, _, _ := ResolveDockerStores(dp, res, 131072, "linux"); again[1].Val != env[1].Val {
		t.Error("an unchanged created store must keep its digest, or every deploy recreates the container")
	}
	reads = nil
	if out, errs, _ := Config(createdEnv(""), res); len(errs) != 0 || !strings.Contains(out, "location: /app/external/classpath/truststores/ts\n") || len(reads) != 0 {
		t.Errorf("config: errs = %v, reads = %v, want the store at its mount path:\n%s", errs, reads, out)
	}
}

// A created store's findings reach the operator on each path that builds it.
func TestCreatedStoreWarningsSurface(t *testing.T) {
	now := time.Now()
	soon := certfixture.CA(t, "Soon", certfixture.Validity(now.Add(-time.Hour), now.Add(5*24*time.Hour)))
	res := Resolver{Env: storeEnv, Rand: fixedStatusRand, Now: func() time.Time { return now }, ReadFile: storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(soon)})}
	kube := "kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      create: true\n    stores:\n      create: true\n"
	_, errs, warns := GenerateKubernetes(createdEnv(kube), res, KubeOpts{})
	if len(errs) > 0 || !hasIssue(warns, "expires on") {
		t.Errorf("kubernetes: errs %v, warns %v; want the expiry warning", errs, warns)
	}
	plan, _, _ := GeneratePodman(createdEnv("podman:\n  command: podman\n  name: c\n"), res)
	if _, pw, err := ResolvePodmanFiles(plan.Files, res); err != nil || !hasIssue(pw, "expires on") {
		t.Errorf("podman: err %v, warns %v; want the expiry warning", err, pw)
	}
}

func hasIssue(issues []Issue, sub string) bool {
	for _, i := range issues {
		if strings.Contains(i.Msg, sub) {
			return true
		}
	}
	return false
}

// A docker store over what one environment variable holds on the host fails
// the deploy naming the store, its sizes and the host, instead of docker's bare
// "argument list too long"; a build failure stops it too; a plan without
// selected store needs no environment at all.
func TestResolveDockerStoresLimits(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(ca)})}
	plan := DockerPlan{Digest: true, Stores: []DockerStore{{Build: StoreBuild{Name: "ts", Store: createdCA("./ca.pem"), Trusted: true}, Secret: "_GEN_TLS_TS_STORE"}}}
	if _, _, err := ResolveDockerStores(plan, res, 100, "windows"); err == nil || !strings.Contains(err.Error(), "once base64-encoded, over the 100 an environment variable holds on windows; trim the certificates, or deploy with podman or kubernetes") || !strings.HasPrefix(err.Error(), "tls.stores.ts is ") {
		t.Errorf("err = %v", err)
	}
	bad := DockerPlan{Digest: true, Stores: []DockerStore{{Build: StoreBuild{Name: "ts", Store: createdCA("./gone.pem")}, Secret: "_GEN_TLS_TS_STORE"}}}
	if _, _, err := ResolveDockerStores(bad, res, 131072, "linux"); err == nil || !strings.Contains(err.Error(), `reading tls.stores.ts.create.certs[0] "./gone.pem"`) {
		t.Errorf("a build failure must stop the deploy: %v", err)
	}
	if env, warns, err := ResolveDockerStores(DockerPlan{}, res, 131072, "linux"); env != nil || warns != nil || err != nil {
		t.Errorf("no selected store: %v %v %v", env, warns, err)
	}
}

// The docker stores digest covers every selected store: an existing store's
// bytes, read from the path the compose file mounts, change it as a created
// store's do, so deploying after an in-place rotation recreates the container.
// The existing store is opened with its password on the way; a file this user
// cannot read only warns and stays out of the digest, and a missing one is the
// host checks' to report.
func TestResolveDockerStoresDigestsExistingStores(t *testing.T) {
	root := certfixture.CA(t, "Root")
	one, two := certfixture.TrustStorePKCS12(t, storePW, root), certfixture.TrustStorePKCS12(t, storePW, certfixture.CA(t, "Other"))
	files := map[string][]byte{"/abs/t.p12": one}
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(files)}
	ex := DockerExisting{Use: consolidate.StoreUse{Name: "ts", File: "./t.p12", Trusted: true}, Path: "/abs/t.p12", Password: spec.Cred{EnvVar: "CA_PW"}}
	plan := DockerPlan{Digest: true, Existing: []DockerExisting{ex}}
	digest := func() string {
		t.Helper()
		env, warns, err := ResolveDockerStores(plan, res, 131072, "linux")
		if err != nil || len(warns) != 0 || len(env) != 1 || env[0].Key != "_GEN_TLS_STORES_SHA256" {
			t.Fatalf("env %v, warns %v, err %v", env, warns, err)
		}
		return env[0].Val
	}
	first := digest()
	if digest() != first {
		t.Error("an unchanged store must keep its digest")
	}
	files["/abs/t.p12"] = two
	if digest() == first {
		t.Error("a store replaced in place must change the digest")
	}
	res.Env = func(string) (string, bool) { return "not-it", true }
	if _, _, err := ResolveDockerStores(plan, res, 131072, "linux"); err == nil || !strings.Contains(err.Error(), "does not open with the value of CA_PW") {
		t.Errorf("a store its password does not open stops the deploy, got %v", err)
	}
	res = Resolver{Env: storeEnv, ReadFile: func(string) ([]byte, error) { return nil, fs.ErrPermission }}
	if env, warns, err := ResolveDockerStores(plan, res, 131072, "linux"); err != nil || len(env) != 1 || !hasIssue(warns, "so it is left out of the stores digest") {
		t.Errorf("an unreadable store warns and stays out: env %v, warns %v, err %v", env, warns, err)
	}
	res.ReadFile = func(string) ([]byte, error) { return nil, fs.ErrNotExist }
	if _, warns, err := ResolveDockerStores(plan, res, 131072, "linux"); err != nil || len(warns) != 0 {
		t.Errorf("a missing store is the host checks' to report: warns %v, err %v", warns, err)
	}
}

// compose down resolves every secret the file declares without reading them,
// so remove hands it placeholders: no store is built and no source is needed.
// A store's secret is named inside the generated-name prefix, after the store.
func TestDockerStorePlaceholders(t *testing.T) {
	plan := DockerPlan{Digest: true, Stores: dockerStores([]consolidate.StoreUse{{Name: "mq-ca", Trusted: true, Created: true}, {Name: "have", File: "./t.p12"}}, spec.TLSConfig{})}
	got := DockerStorePlaceholders(plan)
	want := []KV{{Key: "_GEN_TLS_MQ_CA_STORE", Val: "-"}, {Key: "_GEN_TLS_STORES_SHA256", Val: "-"}}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("placeholders = %v, want %v (the existing store is not a docker secret)", got, want)
	}
	if DockerStorePlaceholders(DockerPlan{}) != nil {
		t.Error("no selected store needs no placeholder")
	}
	existingOnly := DockerStorePlaceholders(DockerPlan{Digest: true})
	if len(existingOnly) != 1 || existingOnly[0] != (KV{Key: "_GEN_TLS_STORES_SHA256", Val: "-"}) {
		t.Errorf("existing stores alone need the digest placeholder only, got %v", existingOnly)
	}
}

// One environment variable holds 32767 characters on Windows; 131072 bytes is
// Linux's cap on one string, kept everywhere else.
func TestEnvVarMax(t *testing.T) {
	for goos, want := range map[string]int{"windows": 32767, "linux": 131072, "darwin": 131072} {
		if got := EnvVarMax(goos); got != want {
			t.Errorf("EnvVarMax(%s) = %d, want %d", goos, got, want)
		}
	}
}

// validateEnv is an env.yaml with the image, the tls.stores entries in stores
// (each indented as under tls.stores), and the platform sections.
func validateEnv(stores string, sections ...string) Request {
	env := "image:\n  name: img\n  tag: v1\ntls:\n  stores:\n" + stores + strings.Join(sections, "")
	return Request{Env: &File{Name: "env.yaml", Data: []byte(env)}, Workflows: synthTLSWorkflowFiles(1, "ts", "ks")}
}

// validate reads and builds what deploy would, for each platform env.yaml has
// a section for, and reports every failure in one run: an existing store that
// cannot be read, a created one that does not build -- even with its password
// unset, built with a stand-in, unless a JKS needs that password to open -- and
// every platform's size limit. A store is read once however many platforms
// carry it; one nothing selects, one an existing kubernetes Secret holds, and
// a docker bind mount (the verb checks those on its host) are never read.
func TestValidateReadsWhatDeployReads(t *testing.T) {
	const (
		kube   = "kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      existing: their-creds\n"
		podman = "podman:\n  command: podman\n  name: c\n"
		docker = "docker:\n  command: docker\n  name: c\n"
	)
	existing := func(name, file string) string {
		return "    " + name + ":\n      password-env: CA_PW\n      existing:\n        pkcs12: " + file + "\n"
	}
	created := func(name, pwEnv, file string) string {
		return "    " + name + ":\n      password-env: " + pwEnv + "\n      create:\n        certs:\n          - " + file + "\n"
	}
	// held puts a store in an operator's kubernetes Secret as well.
	held := func(store string) string {
		return strings.Replace(store, "      password-env:", "      existing:\n        secret:\n          name: my-tls\n      password-env:", 1)
	}
	heldExisting := func(name, file string) string {
		return "    " + name + ":\n      password-env: CA_PW\n      existing:\n        pkcs12: " + file + "\n        secret:\n          name: my-tls\n"
	}
	now := time.Now()
	ca := certfixture.PEM(certfixture.CA(t, "Root"))
	soon := certfixture.PEM(certfixture.CA(t, "Soon", certfixture.Validity(now.Add(-time.Hour), now.Add(5*24*time.Hour))))
	var many []certfixture.Cert
	for i := 0; i < 80; i++ {
		many = append(many, certfixture.CA(t, fmt.Sprintf("CA %d", i)))
	}
	root := certfixture.CA(t, "Store Root")
	files := map[string][]byte{
		"./ca.pem": ca, "./soon.pem": soon, "./many.pem": certfixture.PEM(many...), "./junk.pem": []byte("not a certificate"),
		// Real stores the store password opens, and PKCS12 shells sized for
		// the limits (opened with the password, they only warn).
		"./t.p12": certfixture.TrustStorePKCS12(t, storePW, root), "./k.p12": certfixture.KeyStorePKCS12(t, storePW, certfixture.Issue(t, root, "client"), root),
		"./big.p12": pfxShell(t, 800*1024), "./mid.p12": pfxShell(t, 600000),
		"./jks.p12": certfixture.JKS(t, storePW, []certfixture.Trusted{{Alias: "root", Cert: root}}, nil),
	}
	for _, c := range []struct {
		name      string
		req       Request
		goos      string
		errs      []string // each reported exactly once, and nothing else
		warns     []string // each reported exactly once
		readsNone bool
	}{
		{name: "a readable existing store", req: validateEnv(existing("ts", "./t.p12")+existing("ks", "./k.p12"), kube, podman)},
		{name: "a missing existing store, kubernetes", req: validateEnv(existing("ts", "./gone.p12")+existing("ks", "./k.p12"), kube),
			errs: []string{`reading tls.stores.ts.existing.pkcs12 "./gone.p12"`}},
		{name: "a missing existing store, podman", req: validateEnv(existing("ts", "./gone.p12")+existing("ks", "./k.p12"), podman),
			errs: []string{`reading tls.stores.ts.existing.pkcs12 "./gone.p12"`}},
		{name: "missing under two platforms, reported once", req: validateEnv(existing("ts", "./gone.p12")+existing("ks", "./k.p12"), kube, podman),
			errs: []string{`reading tls.stores.ts.existing.pkcs12 "./gone.p12"`}},
		{name: "two bad stores, both reported", req: validateEnv(existing("ts", "./gone.p12")+existing("ks", "./lost.p12"), podman),
			errs: []string{`reading tls.stores.ts.existing.pkcs12 "./gone.p12"`, `reading tls.stores.ks.existing.pkcs12 "./lost.p12"`}},
		{name: "a docker existing store, missing, is the host checks' to report", req: validateEnv(existing("ts", "./gone.p12")+existing("ks", "./gone.p12"), docker)},
		{name: "a docker existing store is opened with its password", req: validateEnv(existing("ts", "./jks.p12")+existing("ks", "./k.p12"), docker),
			errs: []string{`tls.stores.ts.existing.pkcs12 "./jks.p12" is a JKS keystore, not a PKCS12`}},
		{name: "a created store that does not build", req: validateEnv(created("ts", "CA_PW", "./junk.pem")+existing("ks", "./k.p12"), docker),
			errs: []string{"holds no certificate"}},
		{name: "unset password: built with a stand-in", req: validateEnv(created("ts", "NOPE", "./junk.pem")+existing("ks", "./k.p12"), podman),
			errs: []string{"holds no certificate"}, warns: []string{"tls.stores.ts.password-env names NOPE, which is not set in this environment"}},
		{name: "unset password: a JKS is not opened, but read", req: validateEnv("    ts:\n      password-env: NOPE\n      create:\n        jks: ./gone.jks\n"+existing("ks", "./k.p12"), kube),
			errs:  []string{`reading tls.stores.ts.create.jks "./gone.jks"`},
			warns: []string{"tls.stores.ts.password-env names NOPE, which is not set in this environment"}},
		{name: "unset password: a JKS that is not one", req: validateEnv("    ts:\n      password-env: NOPE\n      create:\n        jks: ./t.p12\n"+existing("ks", "./k.p12"), kube),
			errs: []string{`tls.stores.ts.create.jks (./t.p12) is already a PKCS12 file, not a JKS`}},
		{name: "unset password: a JKS beside certs, the certs still built", req: validateEnv("    ts:\n      password-env: NOPE\n      create:\n        certs:\n          - ./junk.pem\n        jks: ./jks.p12\n"+existing("ks", "./k.p12"), kube),
			errs: []string{"holds no certificate"}},
		{name: "a password variable set but empty", req: validateEnv(created("ts", "EMPTY", "./ca.pem")+existing("ks", "./k.p12"), kube),
			errs: []string{"tls.stores.ts.password-env names EMPTY, which is set but empty; a created store is encrypted with it"}},
		{name: "a store in an operator's Secret is not read on kubernetes", req: validateEnv(held(created("ts", "CA_PW", "./gone.pem"))+heldExisting("ks", "./gone.p12"), kube), readsNone: true},
		{name: "a store in an operator's Secret is still read for podman", req: validateEnv(held(created("ts", "CA_PW", "./ca.pem"))+heldExisting("ks", "./gone.p12"), kube, podman),
			errs: []string{`reading tls.stores.ks.existing.pkcs12 "./gone.p12"`}},
		{name: "an expiring certificate, warned once", req: validateEnv(created("ts", "CA_PW", "./soon.pem")+existing("ks", "./k.p12"), kube, podman, docker),
			warns: []string{"expires on"}},
		{name: "a JKS named as a PKCS12", req: validateEnv(existing("ts", "./jks.p12")+existing("ks", "./k.p12"), kube, podman),
			errs: []string{`tls.stores.ts.existing.pkcs12 "./jks.p12" is a JKS keystore, not a PKCS12`}},
		{name: "a store its password does not open", req: validateEnv(strings.Replace(existing("ts", "./t.p12"), "CA_PW", "OTHER", 1)+existing("ks", "./k.p12"), podman),
			errs: []string{`tls.stores.ts.existing.pkcs12 "./t.p12" does not open with the value of OTHER`}},
		{name: "over the kubernetes Secret", req: validateEnv(existing("ts", "./big.p12")+existing("ks", "./k.p12"), kube),
			errs: []string{"the stores Secret would hold"}},
		{name: "over a podman secret", req: validateEnv(existing("ts", "./mid.p12")+existing("ks", "./k.p12"), podman),
			errs: []string{`tls.stores.ts.existing.pkcs12 "./mid.p12" is 600000 bytes, over the 511999 bytes podman's secret store holds`}},
		{name: "over a windows environment variable", goos: "windows", req: validateEnv(created("ts", "CA_PW", "./many.pem")+existing("ks", "./k.p12"), docker),
			errs: []string{"an environment variable holds on windows"}},
		{name: "within a linux environment variable", goos: "linux", req: validateEnv(created("ts", "CA_PW", "./many.pem")+existing("ks", "./k.p12"), docker)},
	} {
		t.Run(c.name, func(t *testing.T) {
			var reads []string
			res := Resolver{Env: storeEnv, Now: func() time.Time { return now }, GOOS: c.goos, ReadFile: func(p string) ([]byte, error) {
				reads = append(reads, p)
				return storeFiles(files)(p)
			}}
			// The unselected store sits beside every case's: never read.
			c.req.Env.Data = []byte(strings.Replace(string(c.req.Env.Data), "  stores:\n", "  stores:\n"+existing("idle", "./idle.p12"), 1))
			errs, warns := Validate(c.req, res)
			if len(errs) != len(c.errs) {
				t.Errorf("errs = %v, want %d: %q", errs, len(c.errs), c.errs)
			}
			for _, want := range c.errs {
				if n := countIssues(errs, want); n != 1 {
					t.Errorf("%q reported %d times in %v", want, n, errs)
				}
			}
			for _, want := range c.warns {
				if n := countIssues(warns, want); n != 1 {
					t.Errorf("%q warned %d times in %v", want, n, warns)
				}
			}
			for _, p := range reads {
				if p == "./idle.p12" || c.readsNone {
					t.Errorf("read %s, which deploy would not", p)
				}
			}
		})
	}
}

func countIssues(issues []Issue, sub string) int {
	n := 0
	for _, i := range issues {
		if strings.Contains(i.Msg, sub) {
			n++
		}
	}
	return n
}

// Without file access validate says so, rather than pass an existing store it
// never read. Without environment access a created store is still built, with
// the stand-in password; a literal password is used as written.
func TestValidateStoreBytesWithoutAccess(t *testing.T) {
	if _, _, err := validateStoreBytes(consolidate.StoreUse{Name: "ts", File: "./t.p12"}, nil, Resolver{}); err == nil || err.Error() != `cannot read tls.stores.ts.existing.pkcs12 "./t.p12" (no file access)` {
		t.Errorf("no file access: err = %v", err)
	}
	read := storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(certfixture.CA(t, "Root"))})
	use := consolidate.StoreUse{Name: "ts", Trusted: true, Created: true}
	data, _, err := validateStoreBytes(use, &StoreBuild{Name: "ts", Store: createdCA("./ca.pem"), Trusted: true}, Resolver{ReadFile: read})
	if err != nil {
		t.Fatalf("no environment access: %v", err)
	}
	if _, err := pkcs12.DecodeTrustStore(data, storePasswordStandIn); err != nil {
		t.Errorf("the store was not built with the stand-in: %v", err)
	}
	literal := &spec.Store{Password: "lit-pw", Create: createdCA("./ca.pem").Create}
	if data, _, err = validateStoreBytes(use, &StoreBuild{Name: "ts", Store: literal, Trusted: true}, Resolver{ReadFile: read}); err != nil {
		t.Fatalf("literal password: %v", err)
	}
	if _, err := pkcs12.DecodeTrustStore(data, "lit-pw"); err != nil {
		t.Errorf("the store was not built with its literal password: %v", err)
	}
}

// A kubernetes teardown resolves nothing: no credential is read from the
// environment, no store file opened or built, and every Secret the tool owns
// reaches kubectl delete by name alone. A retired stores existing: does not
// stop it either -- the teardown set needs no store -- and is only a warning.
func TestKubernetesTeardownResolvesNothing(t *testing.T) {
	kube := "kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      create: true\n    stores:\n      existing: my-tls\n"
	var reads []string
	res := Resolver{Rand: fixedStatusRand, ReadFile: func(p string) ([]byte, error) {
		reads = append(reads, p)
		return nil, fmt.Errorf("must not be read")
	}}
	out, errs, warns := GenerateKubernetes(createdEnv(kube), res, KubeOpts{Teardown: true})
	if len(errs) > 0 {
		t.Fatalf("teardown: %v", errs)
	}
	if len(reads) > 0 {
		t.Errorf("a teardown opens no store, read %v", reads)
	}
	if !hasIssue(warns, "kubernetes.secrets.stores.existing is no longer read") {
		t.Errorf("the retired existing: is a warning on teardown, got %v", warns)
	}
	for _, want := range []string{"kind: Secret\nmetadata:\n  name: c-credentials\n  namespace: ns\n", "kind: Secret\nmetadata:\n  name: c-stores\n  namespace: ns\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("teardown missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "stringData:") || strings.Contains(out, "\n  ts: ") {
		t.Errorf("a teardown carries no Secret payload:\n%s", out)
	}
	if _, errs, _ := GenerateKubernetes(createdEnv(kube), res, KubeOpts{}); !hasIssue(errs, "kubernetes.secrets.stores.existing is no longer read") {
		t.Errorf("deploy refuses the retired existing:, got %v", errs)
	}
}

// The probes are the operator Secrets the pod reads: one per distinct Secret a
// selected store names, with every key the volume lists from it, then an
// existing credentials Secret with every credential's mount name. A Secret the
// tool writes is never probed.
func TestKubernetesSecretProbes(t *testing.T) {
	env := "image:\n  name: img\n  tag: v1\ntls:\n  stores:\n" +
		"    ts:\n      password-env: CA_PW\n      existing:\n        secret:\n          name: broker-cert\n          key: truststore.p12\n" +
		"    ks:\n      password-env: CA_PW\n      existing:\n        secret:\n          name: broker-cert\n          key: keystore.p12\n" +
		"kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      existing: their-creds\n"
	req := Request{Env: &File{Name: "env.yaml", Data: []byte(env)}, Workflows: synthTLSWorkflowFiles(1, "ts", "ks")}
	var probes []SecretProbe
	if _, errs, _ := GenerateKubernetes(req, Resolver{Env: storeEnv, Rand: fixedStatusRand}, KubeOpts{Probes: &probes}); len(errs) > 0 {
		t.Fatalf("generate: %v", errs)
	}
	if len(probes) != 2 {
		t.Fatalf("probes = %+v, want the store Secret and the credentials Secret", probes)
	}
	if p := probes[0]; p.Name != "broker-cert" || strings.Join(p.Keys, ",") != "truststore.p12,keystore.p12" || p.Field != "tls.stores.ts.existing.secret" {
		t.Errorf("store probe = %+v", p)
	}
	if p := probes[1]; p.Name != "their-creds" || p.Field != "kubernetes.secrets.credentials.existing" || !slices.Contains(p.Keys, "CA_PW") {
		t.Errorf("credentials probe = %+v, want every mount name, CA_PW among them", p)
	}
}

// The ConfigMap is held to the same kubectl apply cap as the stores Secret,
// counting all it carries; a set that fits passes.
func TestKubeConfigMapSizeError(t *testing.T) {
	if err := kubeConfigMapSizeError("c-config", strings.Repeat("a", 200_000), "s", ""); err != nil {
		t.Errorf("200000 bytes fit: %v", err)
	}
	err := kubeConfigMapSizeError("c-config", strings.Repeat("a", 250_000), strings.Repeat("s", 9_000), "<x/>")
	if err == nil || !strings.Contains(err.Error(), "the ConfigMap c-config would hold 259004 bytes (application.yml 250000, status script 9000, logback config 4)") || !strings.Contains(err.Error(), "over the 262144 bytes kubectl apply carries") {
		t.Errorf("want the ConfigMap size error, got %v", err)
	}
}
