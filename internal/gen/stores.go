package gen

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"syscall"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/storebuild"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// StoreBuild is a created store (create:) that deploy builds as it loads the
// store into the platform's secret: its name, its definition, and the roles
// connections select it in -- which a store built from a JKS alone, whose role
// only its content tells, must match.
type StoreBuild struct {
	Name    string
	Store   *spec.Store
	Trusted bool
	Private bool
}

// storeBuild returns the build of a created store, nil for an existing one.
func storeBuild(st consolidate.StoreUse, stores spec.TLSConfig) *StoreBuild {
	if !st.Created {
		return nil
	}
	return &StoreBuild{Name: st.Name, Store: stores.Store(st.Name), Trusted: st.Trusted, Private: st.Private}
}

// storeField is the env.yaml key a selected store comes from, for messages:
// its create: block, or its existing PKCS12.
func storeField(st consolidate.StoreUse) string {
	if st.Created {
		return "tls.stores." + st.Name + ".create"
	}
	return StoreField(st.Name)
}

// BuildStore builds a created store in memory: it resolves the store's
// password, reads every source through res -- resolved against env.yaml, as an
// existing store is read -- and returns the PKCS12 with the findings the build
// raised (an expiring certificate, one left out of a chain) as warnings. The
// bytes are secret material: callers hand them straight to the platform's
// secret, never to the host's disk.
func BuildStore(sb StoreBuild, res Resolver) ([]byte, []Issue, error) {
	in, err := storeInput(sb, res)
	if err != nil {
		return nil, nil, err
	}
	built, err := storebuild.Build(in)
	if err != nil {
		return nil, nil, err
	}
	switch {
	case built.Keystore && sb.Trusted:
		return nil, nil, fmt.Errorf("tls.stores.%s is created from a JKS holding a key, so it holds no trusted certificates, but a connection selects it as a truststore; create a second store from the CA certificates and select that", sb.Name)
	case !built.Keystore && sb.Private:
		return nil, nil, fmt.Errorf("tls.stores.%s is created from a JKS holding only trusted certificates, so it holds no private key, but a connection selects it as a keystore; create the keystore from one key", sb.Name)
	}
	var warns []Issue
	for _, w := range built.Warnings {
		warns = append(warns, Issue{File: fileEnv, Msg: w})
	}
	return built.Data, warns, nil
}

// storeInput reads a created store's password and sources into the builder's
// input, naming each source by its env.yaml position.
func storeInput(sb StoreBuild, res Resolver) (storebuild.Input, error) {
	st := sb.Store
	if !st.Created() {
		return storebuild.Input{}, fmt.Errorf("tls.stores.%s has no create: block to build from", sb.Name)
	}
	pw, err := storePassword(sb.Name, st.Secret(), res)
	if err != nil {
		return storebuild.Input{}, err
	}
	read := func(field, path string) (storebuild.Source, error) {
		where := "tls.stores." + sb.Name + "." + field
		switch {
		case strings.Contains(path, "${"):
			// Expand passes an unset, defaultless variable through verbatim;
			// reading it would fail on a path nobody wrote.
			return storebuild.Source{}, fmt.Errorf("%s %q still holds an unexpanded ${...}; export the variable or give it a default", where, path)
		case res.ReadFile == nil:
			return storebuild.Source{}, fmt.Errorf("cannot read %s %q (no file access)", where, path)
		}
		b, err := res.ReadFile(path)
		if err != nil {
			return storebuild.Source{}, fmt.Errorf("reading %s %q: %v", where, path, err)
		}
		return storebuild.Source{Field: field, Path: path, Data: b}, nil
	}
	in := storebuild.Input{Store: sb.Name, Password: pw, Now: res.now()}
	c := st.Create
	for i, cs := range c.Certs.Items {
		field, alias := fmt.Sprintf("create.certs[%d]", i), ""
		if c.Certs.Named {
			field, alias = "create.certs."+cs.Name, cs.Name
		}
		src, err := read(field, cs.File)
		if err != nil {
			return storebuild.Input{}, err
		}
		in.Certs = append(in.Certs, storebuild.CertSource{Alias: alias, Source: src})
	}
	for i, k := range c.Keys {
		cert, err := read(fmt.Sprintf("create.keys[%d].cert", i), k.Cert)
		if err != nil {
			return storebuild.Input{}, err
		}
		key, err := read(fmt.Sprintf("create.keys[%d].key", i), k.Key)
		if err != nil {
			return storebuild.Input{}, err
		}
		ks := storebuild.KeySource{Cert: cert, Key: key}
		for j, ch := range k.Chain {
			src, err := read(fmt.Sprintf("create.keys[%d].chain[%d]", i, j), ch)
			if err != nil {
				return storebuild.Input{}, err
			}
			ks.Chain = append(ks.Chain, src)
		}
		in.Keys = append(in.Keys, ks)
	}
	if c.JKS != "" {
		src, err := read("create.jks", c.JKS)
		if err != nil {
			return storebuild.Input{}, err
		}
		in.JKS = &src
	}
	return in, nil
}

// storePassword resolves a created store's password: a literal as written, an
// -env form from the environment. A created store is encrypted with it, so an
// empty one is an error; messages name the variable, never a value.
func storePassword(name string, c spec.Cred, res Resolver) (string, error) {
	switch {
	case c.Literal != "":
		return c.Literal, nil
	case c.EnvVar == "":
		return "", fmt.Errorf("tls.stores.%s is created and encrypted with its own password, but sets none; add password-env", name)
	case res.Env == nil:
		return "", fmt.Errorf("tls.stores.%s.password-env names %s, but there is no environment access to read it", name, c.EnvVar)
	}
	v, ok := res.Env(c.EnvVar)
	switch {
	case !ok:
		return "", fmt.Errorf("tls.stores.%s.password-env names %s, which is not set; export it before deploying", name, c.EnvVar)
	case v == "":
		return "", fmt.Errorf("tls.stores.%s.password-env names %s, which is set but empty; a created store is encrypted with it", name, c.EnvVar)
	}
	return v, nil
}

// existingPKCS12 is the outer shape of a PKCS#12 file (RFC 7292 PFX), read
// here rather than through go-pkcs12 so "this is not a PKCS12 at all" is
// decided by structure, never by a library's error text.
type existingPKCS12 struct {
	Version  int
	AuthSafe struct {
		ContentType asn1.ObjectIdentifier
		Content     asn1.RawValue `asn1:"tag:0,explicit,optional"`
	}
	MacData asn1.RawValue `asn1:"optional"`
}

// storePasswordValue is the store password a check can open the store with: a
// literal, or a set and non-empty -env variable. ok is false when there is none
// to use -- validate's unset-variable warning covers that case.
func storePasswordValue(c spec.Cred, res Resolver) (string, bool) {
	if c.Literal != "" {
		return c.Literal, true
	}
	if c.EnvVar == "" || res.Env == nil {
		return "", false
	}
	v, ok := res.Env(c.EnvVar)
	return v, ok && v != ""
}

// checkExistingStore is what the tool can tell about an existing store's
// PKCS12 before the connector opens it, run wherever a platform reads one:
// validate, kubernetes generate and deploy, podman deploy, docker deploy.
// Every store is PKCS12, so a JKS or JCEKS keystore, or a file that is not a
// PKCS12 at all (a PEM, a certificate renamed .p12), is an error with or
// without the password. With the password, a store it does not open is an
// error; a store it opens but whose contents this tool cannot read is only a
// warning -- Java may still accept it -- and so is a truststore holding a
// certificate without Java's trusted-certificate attribute, which Java loads
// as an empty truststore.
func checkExistingStore(st consolidate.StoreUse, data []byte, pw spec.Cred, res Resolver) ([]Issue, error) {
	where := fmt.Sprintf("%s %q", StoreField(st.Name), st.File)
	if len(data) >= 4 {
		switch binary.BigEndian.Uint32(data) {
		case 0xFEEDFEED:
			return nil, fmt.Errorf("%s is a JKS keystore, not a PKCS12: convert it (keytool -importkeystore -srcstoretype JKS -deststoretype PKCS12), or have the tool build the store from it with create: {jks: <file>}", where)
		case 0xCECECECE:
			return nil, fmt.Errorf("%s is a JCEKS keystore, not a PKCS12: convert it with keytool -importkeystore -srcstoretype JCEKS -deststoretype PKCS12", where)
		}
	}
	var pfx existingPKCS12
	if rest, err := asn1.Unmarshal(data, &pfx); err != nil || len(rest) != 0 || pfx.Version != 3 {
		return nil, fmt.Errorf("%s is not a PKCS12 file: every store is PKCS12 -- name a PKCS12 here, or have the tool build the store from certificate files with create:", where)
	}
	password, ok := storePasswordValue(pw, res)
	if !ok {
		return nil, nil
	}
	var err error
	if st.Private {
		_, _, _, err = pkcs12.DecodeChain(data, password)
	} else {
		_, err = pkcs12.DecodeTrustStore(data, password)
	}
	switch {
	case err == nil:
		return nil, nil
	case errors.Is(err, pkcs12.ErrIncorrectPassword) || errors.Is(err, pkcs12.ErrDecryption):
		field := "tls.stores." + st.Name + ".password"
		if pw.EnvVar != "" {
			field = "the value of " + pw.EnvVar + " (tls.stores." + st.Name + ".password-env)"
		}
		return nil, fmt.Errorf("%s does not open with %s; the connector would fail to load it", where, field)
	case strings.Contains(err.Error(), "not marked as trusted"):
		return []Issue{{File: fileEnv, Msg: fmt.Sprintf("%s holds a certificate without Java's trusted-certificate attribute, as a truststore made with openssl pkcs12 -export -nokeys before OpenSSL 3.2 (or without -jdktrust) does: Java loads such a store as empty, and the connection fails with \"the trustAnchors parameter must be non-empty\". Build it from the CA certificates with create: {certs: [...]} instead", where)}}, nil
	}
	return []Issue{{File: fileEnv, Msg: fmt.Sprintf("%s opens with its password, but this tool cannot read its contents (%v); Java may still accept it, so it is not refused", where, err)}}, nil
}

// DockerStore is a created store a compose file carries: its build, and the
// compose secret -- read from the environment variable of the same name -- its
// base64-encoded bytes travel in.
type DockerStore struct {
	Build  StoreBuild
	Secret string
}

// dockerStoreSecret names a created store's compose secret and environment
// variable: inside spec.GeneratedNamePrefix, which no operator's -env variable
// may use, and with a suffix no credential's derived name has.
func dockerStoreSecret(store string) string {
	return spec.GeneratedNamePrefix + "TLS_" + strings.ToUpper(strings.ReplaceAll(store, "-", "_")) + "_STORE"
}

// dockerStores lists the created stores among the selected ones, in their
// order.
func dockerStores(stores []consolidate.StoreUse, defs spec.TLSConfig) []DockerStore {
	var out []DockerStore
	for _, st := range stores {
		if sb := storeBuild(st, defs); sb != nil {
			out = append(out, DockerStore{Build: *sb, Secret: dockerStoreSecret(st.Name)})
		}
	}
	return out
}

// DockerExisting is an existing store a compose file bind-mounts: its use,
// the host path the compose file mounts (expanded and absolute), and its
// password, for deploy to fold its bytes into the stores digest and open it.
type DockerExisting struct {
	Use      consolidate.StoreUse
	Path     string
	Password spec.Cred
}

// dockerExisting lists the existing stores among the selected ones, in their
// order, at the paths the compose file mounts them from.
func dockerExisting(stores []consolidate.StoreUse, defs spec.TLSConfig, res Resolver) []DockerExisting {
	var out []DockerExisting
	for _, st := range stores {
		if st.Created || st.File == "" {
			continue
		}
		out = append(out, DockerExisting{Use: st, Path: res.abs(st.File), Password: defs.Store(st.Name).Secret()})
	}
	return out
}

// dockerOnlyRead reports what a failure to read a store only docker reads
// comes to. Missing, or a directory where the PKCS12 belongs, is nothing here:
// the host checks report it, with the remedy, before anything runs. Any other
// failure -- a file the invoking user may not read, above all -- leaves the
// store out of the digest and unopened, with a warning saying so: the
// connector reads it as its own run-as identity, which the host checks judge.
func dockerOnlyRead(st consolidate.StoreUse, err error) []Issue {
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.EISDIR) {
		return nil
	}
	return []Issue{{File: fileEnv, Msg: fmt.Sprintf("cannot read %s %q here (%v), so it is left out of the stores digest -- a store replaced in place will not recreate the container -- and is not opened with its password", StoreField(st.Name), st.File, err)}}
}

// dockerCreated projects the plan's created stores onto what the compose
// renderer names.
func dockerCreated(stores []DockerStore) []dockergen.CreatedStore {
	var out []dockergen.CreatedStore
	for _, ds := range stores {
		out = append(out, dockergen.CreatedStore{Name: ds.Build.Name, Secret: ds.Secret})
	}
	return out
}

// ResolveDockerStores builds the created stores a docker plan carries and
// returns the environment `docker compose up` runs with for them: each store
// base64-encoded under its secret's name -- an environment variable cannot hold
// the NUL bytes of a PKCS12 -- and the digest the service's stores label is
// filled from. A value longer than envMax, the most one variable holds on host,
// is an error naming the store, rather than docker's bare "argument list too
// long". The result is secret material: callers hand it to the child
// environment only.
//
// The digest covers every selected store: each existing one is read from the
// path the compose file bind-mounts, and opened with its password
// (checkExistingStore), so replacing it in place and deploying recreates the
// container as a changed created store does. A file it cannot read follows
// dockerOnlyRead; the host checks, which run first, have already refused a
// missing one.
func ResolveDockerStores(plan DockerPlan, res Resolver, envMax int, host string) ([]KV, []Issue, error) {
	if !plan.Digest {
		return nil, nil, nil
	}
	var out []KV
	var warns []Issue
	sum := sha256.New()
	for _, ds := range plan.Stores {
		data, w, err := BuildStore(ds.Build, res)
		if err != nil {
			return nil, nil, err
		}
		warns = append(warns, w...)
		enc := b64(data)
		if err := dockerEnvSizeError(ds, len(data), len(enc), envMax, host); err != nil {
			return nil, nil, err
		}
		out = append(out, KV{Key: ds.Secret, Val: enc})
		sum.Write([]byte(ds.Build.Name + "\x00"))
		sum.Write(data)
	}
	for _, ex := range plan.Existing {
		if res.ReadFile == nil {
			return nil, nil, fmt.Errorf("cannot read %s %q (no file access)", StoreField(ex.Use.Name), ex.Use.File)
		}
		data, err := res.ReadFile(ex.Path)
		if err != nil {
			warns = append(warns, dockerOnlyRead(ex.Use, err)...)
			continue
		}
		w, err := checkExistingStore(ex.Use, data, ex.Password, res)
		if err != nil {
			return nil, nil, err
		}
		warns = append(warns, w...)
		sum.Write([]byte(ex.Use.Name + "\x00"))
		sum.Write(data)
	}
	out = append(out, KV{Key: dockergen.StoresDigestVar, Val: hex.EncodeToString(sum.Sum(nil))})
	return out, warns, nil
}

// dockerEnvSizeError refuses a created store whose base64 encoding, n raw
// bytes and enc once encoded, overflows the environment variable that carries
// it: NAME=value and its terminating NUL, as the OS holds it, past envMax.
func dockerEnvSizeError(ds DockerStore, n, enc, envMax int, host string) error {
	if len(ds.Secret)+1+enc+1 <= envMax {
		return nil
	}
	return fmt.Errorf("tls.stores.%s is %d bytes, %d once base64-encoded, over the %d an environment variable holds on %s; trim the certificates, or deploy with podman or kubernetes", ds.Build.Name, n, enc, envMax, host)
}

// EnvVarMax is the most one environment variable -- NAME=value and its
// terminating NUL -- may hold on goos: Windows caps a variable at 32767
// characters; Linux caps one string at 131072 bytes (MAX_ARG_STRLEN), which is
// also kept as the bound everywhere else.
func EnvVarMax(goos string) int {
	if goos == "windows" {
		return 32767
	}
	return 131072
}

// DockerStorePlaceholders is the environment `docker compose down` runs with
// for the stores: compose resolves every secret a file declares even to tear
// it down, but never reads them, so a placeholder stands in for each created
// store's, and for the digest -- remove neither needs a store's sources nor
// builds or reads one.
func DockerStorePlaceholders(plan DockerPlan) []KV {
	if !plan.Digest {
		return nil
	}
	var out []KV
	for _, ds := range plan.Stores {
		out = append(out, KV{Key: ds.Secret, Val: "-"})
	}
	return append(out, KV{Key: dockergen.StoresDigestVar, Val: "-"})
}

// storePasswordStandIn encrypts a created store validate builds without its
// real password. The store is discarded, so only the build's checks of its
// sources matter, and those do not depend on the password -- except a JKS
// source's, which opens with it.
const storePasswordStandIn = "validate-only-stand-in"

// validateStores does for validate what deploy does with the stores the
// rendered connections select, on every platform env.yaml has a section for:
// it reads each existing store a platform loads, builds each created one in
// memory and discards it, and holds the bytes to each platform's size limit --
// the kubernetes stores Secret (every store not kept in an operator's Secret;
// one that is, kubernetes never opens), a podman secret, docker's environment
// variable for a created store. A docker existing store is a bind mount whose
// source the validate verb checks on its own host. Unlike deploy it stops at no
// failure, and it reads or builds a store once however many platforms carry
// it.
//
// Credentials are not resolved: an unset -env variable stays validate.Run's
// warning. A created store whose password is unset is built with a stand-in,
// so every source check still runs, unless it comes from a JKS, which opens
// with that very password.
func validateStores(e *spec.Env, stores []consolidate.StoreUse, res Resolver) (errs, warns []Issue) {
	kube := e.Kubernetes != nil
	docker, podman := e.Docker != nil, e.Podman != nil
	host := res.goos()
	fail := func(err error) { errs = append(errs, Issue{File: fileEnv, Msg: err.Error()}) }
	total := 0
	var sizes []string
	for _, st := range stores {
		sb := storeBuild(st, e.Defaults.TLS)
		kubeReads := kube && !st.ClusterHeld()
		if !kubeReads && !podman && !docker {
			continue
		}
		data, w, err := validateStoreBytes(st, sb, res)
		warns = append(warns, w...)
		switch {
		case err != nil && sb == nil && !kubeReads && !podman && res.ReadFile != nil:
			// Only docker reads it, as deploy does: the same read rule.
			warns = append(warns, dockerOnlyRead(st, err)...)
			continue
		case err != nil:
			fail(err)
			continue
		case data == nil:
			continue
		}
		if sb == nil {
			cw, cerr := checkExistingStore(st, data, e.Defaults.TLS.Store(st.Name).Secret(), res)
			warns = append(warns, cw...)
			if cerr != nil {
				fail(cerr)
				continue
			}
		}
		enc := base64.StdEncoding.EncodedLen(len(data))
		if kubeReads {
			total += enc
			sizes = append(sizes, fmt.Sprintf("%s %d bytes", st.Name, enc))
		}
		if podman {
			what := storeField(st)
			if sb == nil {
				what = fmt.Sprintf("%s %q", what, st.File)
			}
			if err := podmanSecretSizeError(what, len(data)); err != nil {
				fail(err)
			}
		}
		if docker && sb != nil {
			if err := dockerEnvSizeError(DockerStore{Build: *sb, Secret: dockerStoreSecret(st.Name)}, len(data), enc, EnvVarMax(host), host); err != nil {
				fail(err)
			}
		}
	}
	if total > 0 {
		if err := kubeStoresSizeError(total, sizes); err != nil {
			fail(err)
		}
	}
	return errs, warns
}

// validateStoreBytes reads an existing store, or builds a created one (sb),
// for validateStores. A created store whose password variable is set but empty
// is deploy's error. With the password unset the store is built with a
// stand-in, so every source check still runs -- except a JKS's, which opens
// with that very password: then every check that needs no password runs
// (validateJKSWithoutPassword), and no bytes come back.
func validateStoreBytes(st consolidate.StoreUse, sb *StoreBuild, res Resolver) ([]byte, []Issue, error) {
	if sb == nil {
		if res.ReadFile == nil {
			return nil, nil, fmt.Errorf("cannot read %s %q (no file access)", StoreField(st.Name), st.File)
		}
		b, err := res.ReadFile(st.File)
		if err != nil {
			return nil, nil, fmt.Errorf("reading %s %q: %w", StoreField(st.Name), st.File, err)
		}
		return b, nil, nil
	}
	c := sb.Store.Secret()
	if c.Literal == "" && c.EnvVar != "" && res.Env != nil {
		if v, ok := res.Env(c.EnvVar); ok && v == "" {
			_, err := storePassword(sb.Name, c, res)
			return nil, nil, err
		}
	}
	if c.Literal == "" && !envSet(res, c.EnvVar) {
		// The store's password is the one variable a build reads.
		res.Env = func(string) (string, bool) { return storePasswordStandIn, true }
		if sb.Store.Create.JKS != "" {
			return validateJKSWithoutPassword(*sb, res)
		}
	}
	return BuildStore(*sb, res)
}

// validateJKSWithoutPassword checks a store created from a JKS while its
// password is unset: it reads every source, holds the JKS to the format a
// created store can read (storebuild.SniffJKS), and builds the store from its
// certs alone when it lists any. Only the JKS's integrity check and its
// entries wait for the password; with no bytes, no size is checked either.
func validateJKSWithoutPassword(sb StoreBuild, res Resolver) ([]byte, []Issue, error) {
	in, err := storeInput(sb, res)
	if err != nil {
		return nil, nil, err
	}
	if err := storebuild.SniffJKS(*in.JKS, sb.Name); err != nil {
		return nil, nil, err
	}
	if len(in.Certs) == 0 {
		return nil, nil, nil
	}
	in.JKS = nil
	built, err := storebuild.Build(in)
	if err != nil {
		return nil, nil, err
	}
	var warns []Issue
	for _, w := range built.Warnings {
		warns = append(warns, Issue{File: fileEnv, Msg: w})
	}
	return nil, warns, nil
}

// envSet reports whether the environment variable name holds a value.
func envSet(res Resolver, name string) bool {
	if res.Env == nil || name == "" {
		return false
	}
	v, ok := res.Env(name)
	return ok && v != ""
}
