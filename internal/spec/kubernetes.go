package spec

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Secret sources.
const (
	SourceEnv  = "env"
	SourceFile = "file"
)

// Syslog protocols.
const (
	SyslogUDP = "udp"
	SyslogTCP = "tcp"
)

// Resources is the container's cpu/memory. requests and limits are emitted with
// the same value (guaranteed QoS), so only one cpu/memory pair is specified.
type Resources struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// Deployment mirrors the kubernetes.deployment section of env.yaml.
type Deployment struct {
	Name      string    `yaml:"name"`
	Namespace string    `yaml:"namespace"`
	Image     string    `yaml:"image"`
	Replicas  int       `yaml:"replicas"`
	Resources Resources `yaml:"resources"`
	Timezone  string    `yaml:"timezone"`
}

// Service mirrors the kubernetes.service section of env.yaml.
type Service struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// CredCreate builds the credentials Secret.
type CredCreate struct {
	Name       string   `yaml:"name"`
	Source     string   `yaml:"source"`      // SourceEnv | SourceFile
	Variables  []string `yaml:"variables"`   // required when Source == env
	ValuesFile string   `yaml:"values-file"` // required when Source == file
}

// CredentialsSecret is the env-var Secret (envFrom). Create XOR Existing.
type CredentialsSecret struct {
	Create   *CredCreate `yaml:"create"`
	Existing string      `yaml:"existing"`
}

// StoreCreate embeds the .jks files from env.yaml tls.*.file.
type StoreCreate struct {
	Name string `yaml:"name"`
}

// StoresSecret is the truststore/keystore Secret (volume mount). Create XOR Existing.
type StoresSecret struct {
	Create   *StoreCreate `yaml:"create"`
	Existing string       `yaml:"existing"`
}

// Secrets groups the two optional secret wirings.
type Secrets struct {
	Credentials *CredentialsSecret `yaml:"credentials"`
	Stores      *StoresSecret      `yaml:"stores"`
}

// Syslog mirrors the kubernetes.logging.syslog section of env.yaml.
type Syslog struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Protocol string `yaml:"protocol"` // SyslogUDP (default) | SyslogTCP
}

// Logging mirrors the kubernetes.logging section of env.yaml.
type Logging struct {
	Syslog *Syslog `yaml:"syslog"`
}

// NFS locates the export backing a created PersistentVolume.
type NFS struct {
	Server string `yaml:"server"`
	Path   string `yaml:"path"`
}

// PVCCreate emits an NFS PersistentVolume + PersistentVolumeClaim pair.
type PVCCreate struct {
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

// DefaultKubeCommand is the CLI used to apply/delete manifests when the
// kubernetes.command key is unset.
const DefaultKubeCommand = "kubectl"

// Kubernetes is the parsed kubernetes section of env.yaml.
type Kubernetes struct {
	Command    string     `yaml:"command"` // deploy CLI (default kubectl; e.g. "oc" or "kubectl --context prod")
	Deployment Deployment `yaml:"deployment"`
	Service    Service    `yaml:"service"`
	Logging    *Logging   `yaml:"logging"`
	Libs       *Libs      `yaml:"libs"`
	Secrets    Secrets    `yaml:"secrets"`
}

// ParseKubernetes decodes a standalone kubernetes document (env.yaml reuses
// applyKubeDefaults via ParseEnv). Replicas defaults to 1 when unset.
func ParseKubernetes(data []byte) (*Kubernetes, error) {
	var k Kubernetes
	if err := yaml.Unmarshal(data, &k); err != nil {
		return nil, fmt.Errorf("env.yaml: %v", err)
	}
	applyKubeDefaults(&k)
	return &k, nil
}

// applyKubeDefaults fills in the defaults the connector runtime expects: command
// kubectl, replicas 1, udp syslog, 1Gi libs storage, busybox download image.
func applyKubeDefaults(k *Kubernetes) {
	if k.Command == "" {
		k.Command = DefaultKubeCommand
	}
	if k.Deployment.Replicas == 0 {
		k.Deployment.Replicas = 1
	}
	if l := k.Logging; l != nil && l.Syslog != nil && l.Syslog.Protocol == "" {
		l.Syslog.Protocol = SyslogUDP
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
