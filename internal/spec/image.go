package spec

import "strings"

// DefaultRegistry is the auths key both docker and the kubelet use for Docker
// Hub when an image names no registry. It is the historical v1 URL rather than
// a hostname because that is the key `docker login` writes and the kubelet
// looks credentials up under.
const DefaultRegistry = "https://index.docker.io/v1/"

// Image is the top-level image: block -- one declaration every platform reads,
// replacing the per-platform image keys that used to drift apart (the shipped
// example had pinned two different versions).
//
// Repo is the registry host alone. Everything after it is the repository path
// and belongs in Name, so a Docker Hub image is
// name: solace/solace-pubsub-connector-ibmmq with no repo at all -- putting the
// "solace" namespace in Repo would render the same reference but look up
// credentials under a registry that does not exist.
//
// The registry account is consulted only when the tool is asked to build a pull
// secret (kubernetes.secrets.image-pull.create); referencing a secret you made
// yourself needs neither field, which is why both are optional.
//
// Both are literal/-env pairs like every other credential here, so they get the
// same treatment from validate.checkCred and gen.ResolveCredentials: one form or
// the other but never both, the variable name checked, and a warning when it is
// not exported.
type Image struct {
	Repo string `yaml:"repo"` // registry host (and port); empty means Docker Hub. A Hub namespace belongs in Name
	Name string `yaml:"name"`
	Tag  string `yaml:"tag"`

	User    string `yaml:"user" expand:"no"`     // credential: out of scope (rule 5)
	UserEnv string `yaml:"user-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
	Pass    string `yaml:"pass" expand:"no"`     // credential: out of scope (rule 5)
	PassEnv string `yaml:"pass-env" expand:"no"` // names a host var; expanding it would defeat the -env indirection
}

// UserCred and PassCred expose the registry account as the same Cred pair every
// other credential position uses, so one set of checks and one resolver serve
// all of them.
func (i *Image) UserCred() Cred {
	if i == nil {
		return Cred{}
	}
	return Cred{i.User, i.UserEnv}
}

func (i *Image) PassCred() Cred {
	if i == nil {
		return Cred{}
	}
	return Cred{i.Pass, i.PassEnv}
}

// Ref assembles the reference the container engines pull -- repo/name:tag, with
// the repo dropped when unset. A tag naming a digest joins with '@' rather than
// ':', so pinning by digest renders as name@sha256:... instead of the
// name:sha256:... that no engine would accept.
func (i *Image) Ref() string {
	if i == nil || i.Name == "" {
		return ""
	}
	ref := i.Name
	if i.Repo != "" {
		ref = strings.TrimRight(i.Repo, "/") + "/" + ref
	}
	switch {
	case i.Tag == "":
		return ref
	case strings.HasPrefix(i.Tag, "sha256:"):
		return ref + "@" + i.Tag
	default:
		return ref + ":" + i.Tag
	}
}

// Registry is the key this image's credentials are stored under in an auths
// map -- the registry host, which is exactly what Repo holds. A namespace such
// as Docker Hub's "solace/" is part of the repository path and belongs in Name,
// so it never reaches the credential lookup.
func (i *Image) Registry() string {
	if i == nil || i.Repo == "" {
		return DefaultRegistry
	}
	return strings.TrimRight(i.Repo, "/")
}
