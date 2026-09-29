package spec

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Resources is the container's cpu/memory. requests and limits are emitted with
// the same value (guaranteed QoS), so only one cpu/memory pair is specified.
type Resources struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// Deployment mirrors the kubernetes.deployment section of env.yaml.
//
// Image and Timezone are parsed but rejected: both moved to their own
// top-level keys so one declaration serves every platform. The fields stay so a
// stale config fails loudly in validate rather than being silently ignored.
type Deployment struct {
	Name      string    `yaml:"name"`
	Namespace string    `yaml:"namespace"`
	Image     string    `yaml:"image"` // removed; non-empty is a validation error
	Replicas  int       `yaml:"replicas"`
	Resources Resources `yaml:"resources"`
	Timezone  string    `yaml:"timezone"` // removed; non-empty is a validation error
}

// Service mirrors the kubernetes.service section of env.yaml. Port.Host is
// the Service's own port; Port.Container is the container targetPort it
// forwards to -- the same scalar / "host:container" shape docker and podman
// ports accept, via Port.UnmarshalYAML.
type Service struct {
	Enabled bool `yaml:"enabled"`
	Port    Port `yaml:"port"`
}

// CredCreate builds the credentials Secret. Its contents are no longer declared
// here: the keys are every credential the config references, derived from the
// spec itself, and their values come from the literals and `-env` variables
// those positions name. Nor is its name: the Secret is
// <deployment.name>-credentials (Kubernetes.CredentialsSecretName).
//
// It is written create: true. The older mapping form, create: {name: ...},
// still means "create" so an old env.yaml keeps deploying.
type CredCreate struct {
	// Name is retired: generate and deploy ignore it quietly, and only validate
	// reports it, asking for it to be removed. A user-chosen name let two
	// instances in one namespace render the same Secret, so each deploy
	// overwrote the other's credentials and either remove deleted both.
	Name string `yaml:"name"`

	// Removed keys, kept only to fail loudly. yaml.v3 ignores unknown fields, so
	// without these an old env.yaml would parse cleanly and silently drop its
	// entire credential configuration.
	Source     string   `yaml:"source"`
	Variables  []string `yaml:"variables"`
	ValuesFile string   `yaml:"values-file"`
}

// RemovedKeys names any key that no longer has meaning, so the caller can reject
// a stale config instead of quietly ignoring it.
func (c *CredCreate) RemovedKeys() []string {
	if c == nil {
		return nil
	}
	var out []string
	if c.Source != "" {
		out = append(out, "source")
	}
	if len(c.Variables) > 0 {
		out = append(out, "variables")
	}
	if c.ValuesFile != "" {
		out = append(out, "values-file")
	}
	return out
}

// CredentialsSecret is the credentials Secret, mounted as a volume at
// /run/secrets with one file per credential name -- never envFrom, so no
// credential is ever an environment variable in the container. Create XOR
// Existing, enforced by validate.
type CredentialsSecret struct {
	Create   *CredCreate `yaml:"create"`
	Existing string      `yaml:"existing"`
}

// UnmarshalYAML accepts create: in both spellings (see decodeCreate), so Create
// is non-nil exactly when the tool is to build the Secret.
func (c *CredentialsSecret) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Create   yaml.Node `yaml:"create"`
		Existing string    `yaml:"existing"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	var cc CredCreate
	create, err := decodeCreate(&raw.Create, "kubernetes.secrets.credentials.create", &cc)
	if err != nil {
		return err
	}
	if create {
		c.Create = &cc
	}
	c.Existing = raw.Existing
	return nil
}

// StoreCreate embeds the .jks files from env.yaml tls.*.file into a Secret
// named <deployment.name>-stores (Kubernetes.StoresSecretName). It is written
// create: true; the older create: {name: ...} still means "create".
type StoreCreate struct {
	// Name is retired, for the reason CredCreate.Name is: ignored by generate
	// and deploy, reported by validate.
	Name string `yaml:"name"`
}

// StoresSecret is the truststore/keystore Secret (volume mount). Create XOR
// Existing, enforced by validate. An Existing Secret's keys are the base
// filenames of tls.truststore.file / tls.keystore.file, which is what the
// rendered config points at under the stores mount.
type StoresSecret struct {
	Create   *StoreCreate `yaml:"create"`
	Existing string       `yaml:"existing"`
}

// UnmarshalYAML accepts create: in both spellings (see decodeCreate), so Create
// is non-nil exactly when the tool is to build the Secret.
func (s *StoresSecret) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Create   yaml.Node `yaml:"create"`
		Existing string    `yaml:"existing"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	var sc StoreCreate
	create, err := decodeCreate(&raw.Create, "kubernetes.secrets.stores.create", &sc)
	if err != nil {
		return err
	}
	if create {
		s.Create = &sc
	}
	s.Existing = raw.Existing
	return nil
}

// decodeCreate reads a secrets create: value. true means the tool builds the
// Secret and false, empty or absent means it does not. The older mapping form
// (create: {name: ...}) still means "create", so an env.yaml written before the
// names were derived keeps deploying; its keys are decoded into legacy so
// validate can report the ones that are retired.
func decodeCreate(n *yaml.Node, field string, legacy any) (bool, error) {
	if n.Kind == yaml.AliasNode && n.Alias != nil {
		n = n.Alias
	}
	switch n.Kind {
	case 0:
		return false, nil // the key is absent
	case yaml.MappingNode:
		if err := n.Decode(legacy); err != nil {
			return false, err
		}
		return true, nil
	case yaml.ScalarNode:
		if n.ShortTag() == "!!null" {
			return false, nil
		}
		var b bool
		if err := n.Decode(&b); err != nil {
			return false, fmt.Errorf("%s must be true or false, got %q", field, n.Value)
		}
		return b, nil
	}
	return false, fmt.Errorf("%s must be true or false, got a %s", field, YAMLKind(n))
}

// Secrets groups the optional secret wirings.
type Secrets struct {
	Credentials *CredentialsSecret `yaml:"credentials"`
	Stores      *StoresSecret      `yaml:"stores"`
	ImagePull   *ImagePullSecret   `yaml:"image-pull"`
}

// ImagePullSecret wires the registry credential the kubelet pulls the image
// with. Name alone references a Secret the operator manages; Create instead has
// the tool render one from the top-level image block, named
// <deployment.name>-image-pull (Kubernetes.ImagePullSecretName). Either way the
// Secret reaches the pod template as an imagePullSecrets entry.
//
// With Create, Name is retired: generate and deploy ignore it quietly and only
// validate reports it, for the reason CredCreate.Name is.
//
// Create defaults to false -- absent and false mean the same thing, which is
// why a plain bool is enough. Building a Secret is a mutation, and naming one
// you manage yourself must not overwrite it; asking the tool to own it is the
// explicit choice, matching how Credentials makes create/existing explicit.
type ImagePullSecret struct {
	Name   string `yaml:"name"`
	Create bool   `yaml:"create"`
}

// NFS locates the export backing a created PersistentVolume.
type NFS struct {
	Server string `yaml:"server"`
	Path   string `yaml:"path"`
}

// PVCCreate emits an NFS PersistentVolume + PersistentVolumeClaim pair, named
// after deployment.name (Kubernetes.LibsPVCName, LibsPVName). Several instances
// may point at the same export: each gets its own PV and claim onto it.
type PVCCreate struct {
	// Name is retired: generate and deploy ignore it quietly, and only validate
	// reports it. A user-chosen claim name let two instances in one namespace
	// share one claim, so remove of either one hung -- kubectl delete waits on
	// kubernetes.io/pvc-protection while the other instance's pod still
	// mounts it -- and the other instance lost its libs at its next restart.
	Name    string `yaml:"name"`
	Storage string `yaml:"storage"` // default 1Gi
	NFS     NFS    `yaml:"nfs"`
}

// LibsPVC mounts a PVC that already holds the jars. Create XOR Existing.
type LibsPVC struct {
	Create   *PVCCreate `yaml:"create"`
	Existing string     `yaml:"existing"`
}

// LibsDownload downloads the jars in an initContainer at pod start.
type LibsDownload struct {
	URLs  []string `yaml:"urls"`
	Image string   `yaml:"image"` // default busybox:1.37 (needs wget)
	PVC   string   `yaml:"pvc"`   // optional existing PVC; empty = emptyDir
}

// Libs provides the IBM MQ java libraries at /app/external/libs. Exactly one mode.
type Libs struct {
	PVC      *LibsPVC      `yaml:"pvc"`
	Download *LibsDownload `yaml:"download"`
}

// LibsPVName is the name of the PersistentVolume backing a created libs PVC.
//
// It carries the namespace because a PersistentVolume is cluster-scoped while
// the claim naming it is not: the claim name is only unique within a namespace,
// so deriving the PV name from it alone would let two releases in different
// namespaces fight over one PV object -- the second apply rebinding it, and the
// first release's pods then left unable to schedule on a claim that would
// never bind.
//
// The result is a single DNS-1123 label, so validate caps the combined length
// rather than letting the API server reject it mid-apply.
func LibsPVName(namespace, claim string) string { return namespace + "-" + claim + "-pv" }

// The suffixes of the objects the tool creates and names after
// kubernetes.deployment.name. A Deployment name is already unique within its
// namespace, so every derived name is too: two instances can never render,
// and then tear down, the same Secret or claim.
const (
	ConfigMapSuffix         = "-config"
	CredentialsSecretSuffix = "-credentials"
	StoresSecretSuffix      = "-stores"
	ImagePullSecretSuffix   = "-image-pull"
	LibsPVCSuffix           = "-libs"
)

// CredentialsSecretName is the credentials Secret the pod mounts, and whether
// the tool creates it: the derived name when it does, the existing: name when
// it only references one, "" when there is none.
func (k *Kubernetes) CredentialsSecretName() (name string, created bool) {
	c := k.Secrets.Credentials
	switch {
	case c == nil:
		return "", false
	case c.Create != nil:
		return k.Deployment.Name + CredentialsSecretSuffix, true
	}
	return c.Existing, false
}

// StoresSecretName is CredentialsSecretName for the truststore/keystore Secret.
func (k *Kubernetes) StoresSecretName() (name string, created bool) {
	s := k.Secrets.Stores
	switch {
	case s == nil:
		return "", false
	case s.Create != nil:
		return k.Deployment.Name + StoresSecretSuffix, true
	}
	return s.Existing, false
}

// ImagePullSecretName is the registry Secret the pod pulls with, and whether
// the tool creates it: the derived name with create, the operator's own name
// without, "" when there is no image-pull block.
func (k *Kubernetes) ImagePullSecretName() (name string, created bool) {
	ip := k.Secrets.ImagePull
	switch {
	case ip == nil:
		return "", false
	case ip.Create:
		return k.Deployment.Name + ImagePullSecretSuffix, true
	}
	return ip.Name, false
}

// LibsPVCName is the claim the tool creates for libs.pvc.create, or "" when it
// creates none. Its PersistentVolume is LibsPVName(namespace, LibsPVCName()).
func (k *Kubernetes) LibsPVCName() string {
	if lb := k.Libs; lb != nil && lb.PVC != nil && lb.PVC.Create != nil {
		return k.Deployment.Name + LibsPVCSuffix
	}
	return ""
}

// CreatedNames lists every namespaced object the tool creates for this
// instance, in a fixed order: the Deployment (and its Service, which shares the
// name), the ConfigMap, then each Secret and claim the config asks it to build.
// validate holds each to the DNS-1123 label limit, and remove uses the list to
// tell a straggler of its own from someone else's object.
func (k *Kubernetes) CreatedNames() []string {
	name := k.Deployment.Name
	out := []string{name, name + ConfigMapSuffix}
	if n, created := k.CredentialsSecretName(); created {
		out = append(out, n)
	}
	if n, created := k.StoresSecretName(); created {
		out = append(out, n)
	}
	if n, created := k.ImagePullSecretName(); created {
		out = append(out, n)
	}
	if n := k.LibsPVCName(); n != "" {
		out = append(out, n)
	}
	return out
}

// DefaultKubeCommand is the CLI used to apply/delete manifests when the
// kubernetes.command key is unset.
const DefaultKubeCommand = "kubectl"

// Kubernetes is the parsed kubernetes section of env.yaml.
//
// Logging is parsed but rejected: syslog moved to the top-level logging: block,
// beside logging.level, so one declaration serves every platform. The field
// stays so a stale config fails loudly in validate rather than being silently
// ignored -- ParseEnv decodes non-strict, so a deleted field would simply be
// dropped and the instance would come up with no syslog and no diagnostic.
type Kubernetes struct {
	Command    string     `yaml:"command"` // deploy CLI (default kubectl; e.g. "oc" or "kubectl --context prod")
	Deployment Deployment `yaml:"deployment"`
	Service    Service    `yaml:"service"`
	Logging    *Logging   `yaml:"logging"` // removed; non-nil is a validation error
	Libs       *Libs      `yaml:"libs"`
	Secrets    Secrets    `yaml:"secrets"`
}

// ParseKubernetes decodes a standalone kubernetes document (env.yaml reuses
// applyKubeDefaults via ParseEnv). Replicas defaults to 1 when unset; there is
// no defaults.Management in this standalone path, so the service port falls
// back to DefaultMgmtPort.
func ParseKubernetes(data []byte) (*Kubernetes, error) {
	var k Kubernetes
	if err := yaml.Unmarshal(data, &k); err != nil {
		return nil, fmt.Errorf("env.yaml: %v", err)
	}
	applyKubeDefaults(&k, DefaultMgmtPort)
	return &k, nil
}

// applyKubeDefaults fills in the defaults the connector runtime expects:
// command kubectl, replicas 1, 1Gi libs storage, busybox download image. mgmtPort is the effective management.port (see
// EffectiveManagementPort): an unset service.port defaults to publishing that
// port to itself, since that is the only port the pod actually listens on.
func applyKubeDefaults(k *Kubernetes, mgmtPort int) {
	if k.Command == "" {
		k.Command = DefaultKubeCommand
	}
	if k.Deployment.Replicas == 0 {
		k.Deployment.Replicas = 1
	}
	if k.Service.Port == (Port{}) {
		k.Service.Port = Port{Host: mgmtPort, Container: mgmtPort}
	}
	if lb := k.Libs; lb != nil {
		if lb.PVC != nil && lb.PVC.Create != nil && lb.PVC.Create.Storage == "" {
			lb.PVC.Create.Storage = "1Gi"
		}
		if lb.Download != nil && lb.Download.Image == "" {
			lb.Download.Image = "busybox:1.37"
		}
	}
}
