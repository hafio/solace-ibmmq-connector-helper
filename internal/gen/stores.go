package gen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/consolidate"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/dockergen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/storebuild"
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

// createdStoreErrors reports every selected created store as an error for a
// command that cannot carry one, with why.
func createdStoreErrors(stores []consolidate.StoreUse, format string) []Issue {
	var out []Issue
	for _, st := range stores {
		if st.Created {
			out = append(out, Issue{File: fileEnv, Msg: fmt.Sprintf(format, st.Name)})
		}
	}
	return out
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
func ResolveDockerStores(plan DockerPlan, res Resolver, envMax int, host string) ([]KV, []Issue, error) {
	if len(plan.Stores) == 0 {
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
		// The variable as the OS holds it: NAME=value and its terminating NUL.
		if n := len(ds.Secret) + 1 + len(enc) + 1; n > envMax {
			return nil, nil, fmt.Errorf("tls.stores.%s is %d bytes, %d once base64-encoded, over the %d an environment variable holds on %s; trim the certificates, or deploy with podman or kubernetes", ds.Build.Name, len(data), len(enc), envMax, host)
		}
		out = append(out, KV{Key: ds.Secret, Val: enc})
		sum.Write([]byte(ds.Build.Name + "\x00"))
		sum.Write(data)
	}
	out = append(out, KV{Key: dockergen.StoresDigestVar, Val: hex.EncodeToString(sum.Sum(nil))})
	return out, warns, nil
}

// DockerStorePlaceholders is the environment `docker compose down` runs with
// for the created stores: compose resolves every secret a file declares even to
// tear it down, but never reads them, so a placeholder stands in -- remove
// neither needs a store's sources nor builds it.
func DockerStorePlaceholders(plan DockerPlan) []KV {
	if len(plan.Stores) == 0 {
		return nil
	}
	var out []KV
	for _, ds := range plan.Stores {
		out = append(out, KV{Key: ds.Secret, Val: "-"})
	}
	return append(out, KV{Key: dockergen.StoresDigestVar, Val: "-"})
}
