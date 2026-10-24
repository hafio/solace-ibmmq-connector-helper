package gen

import (
	"encoding/base64"
	"io/fs"
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
// while the store does. An existing stores Secret is the operator's, so
// nothing is built for it; generate config, which has no secret to carry it,
// refuses it by name.
func TestCreatedStoresOnEachPlatform(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	var reads []string
	res := Resolver{Env: storeEnv, Rand: fixedStatusRand, ReadFile: func(p string) ([]byte, error) {
		reads = append(reads, p)
		return storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(ca)})(p)
	}}

	kube := "kubernetes:\n  command: kubectl\n  deployment:\n    name: c\n    namespace: ns\n  secrets:\n    credentials:\n      create: true\n    stores:\n      create: true\n"
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
	existing := strings.Replace(kube, "    stores:\n      create: true\n", "    stores:\n      existing: my-tls\n", 1)
	if _, errs, _ := GenerateKubernetes(createdEnv(existing), res, KubeOpts{}); len(errs) > 0 {
		t.Fatalf("kubernetes with an existing stores Secret: %v", errs)
	}
	if len(reads) != 0 {
		t.Errorf("an existing stores Secret builds nothing, but read %v", reads)
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
	if _, errs, _ := Config(createdEnv(""), res); len(errs) != 1 || !strings.Contains(errs[0].Msg, "tls.stores.ts is created at deploy and has no file for generate config to reference; use existing: {pkcs12: <file>} for a connector you run yourself") {
		t.Errorf("config: errs = %v", errs)
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
// created stores needs no environment at all.
func TestResolveDockerStoresLimits(t *testing.T) {
	ca := certfixture.CA(t, "Root")
	res := Resolver{Env: storeEnv, ReadFile: storeFiles(map[string][]byte{"./ca.pem": certfixture.PEM(ca)})}
	plan := DockerPlan{Stores: []DockerStore{{Build: StoreBuild{Name: "ts", Store: createdCA("./ca.pem"), Trusted: true}, Secret: "_GEN_TLS_TS_STORE"}}}
	if _, _, err := ResolveDockerStores(plan, res, 100, "windows"); err == nil || !strings.Contains(err.Error(), "once base64-encoded, over the 100 an environment variable holds on windows; trim the certificates, or deploy with podman or kubernetes") || !strings.HasPrefix(err.Error(), "tls.stores.ts is ") {
		t.Errorf("err = %v", err)
	}
	bad := DockerPlan{Stores: []DockerStore{{Build: StoreBuild{Name: "ts", Store: createdCA("./gone.pem")}, Secret: "_GEN_TLS_TS_STORE"}}}
	if _, _, err := ResolveDockerStores(bad, res, 131072, "linux"); err == nil || !strings.Contains(err.Error(), `reading tls.stores.ts.create.certs[0] "./gone.pem"`) {
		t.Errorf("a build failure must stop the deploy: %v", err)
	}
	if env, warns, err := ResolveDockerStores(DockerPlan{}, res, 131072, "linux"); env != nil || warns != nil || err != nil {
		t.Errorf("no created store: %v %v %v", env, warns, err)
	}
}

// compose down resolves every secret the file declares without reading them,
// so remove hands it placeholders: no store is built and no source is needed.
// A store's secret is named inside the generated-name prefix, after the store.
func TestDockerStorePlaceholders(t *testing.T) {
	plan := DockerPlan{Stores: dockerStores([]consolidate.StoreUse{{Name: "mq-ca", Trusted: true, Created: true}, {Name: "have", File: "./t.p12"}}, spec.TLSConfig{})}
	got := DockerStorePlaceholders(plan)
	want := []KV{{Key: "_GEN_TLS_MQ_CA_STORE", Val: "-"}, {Key: "_GEN_TLS_STORES_SHA256", Val: "-"}}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("placeholders = %v, want %v (the existing store is not a docker secret)", got, want)
	}
	if DockerStorePlaceholders(DockerPlan{}) != nil {
		t.Error("no created store needs no placeholder")
	}
}
