package statusreport

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// Every parser here is best-effort by design: it is reading another program's
// output, and a field that moved or a shape that changed must cost one line of
// the report rather than the whole run. A document that cannot be decoded at
// all is an error (the caller turns it into a note); a document that decodes
// but lacks a field leaves that field empty, and the renderer drops the line.

// ---- kubernetes --------------------------------------------------------------

// kubeList is the envelope `kubectl get <kind> -o json` returns for a
// multi-object query. A single-object query returns the object itself with no
// items array, which is why Items is checked for nil rather than for length.
type kubeList struct {
	Items []json.RawMessage `json:"items"`
}

// kubePod is the subset of a Pod this report reads. Anything not named here is
// ignored by encoding/json, so a newer API version adding fields changes
// nothing.
type kubePod struct {
	Metadata struct {
		Name              string `json:"name"`
		Namespace         string `json:"namespace"`
		CreationTimestamp string `json:"creationTimestamp"`
	} `json:"metadata"`
	Spec struct {
		NodeName   string          `json:"nodeName"`
		Containers []kubeContainer `json:"containers"`
		Volumes    []struct {
			Name      string `json:"name"`
			ConfigMap *struct {
				Name string `json:"name"`
			} `json:"configMap"`
			Secret *struct {
				SecretName string `json:"secretName"`
			} `json:"secret"`
			PVC *struct {
				ClaimName string `json:"claimName"`
			} `json:"persistentVolumeClaim"`
		} `json:"volumes"`
		ImagePullSecrets []struct {
			Name string `json:"name"`
		} `json:"imagePullSecrets"`
	} `json:"spec"`
	Status struct {
		Phase             string `json:"phase"`
		StartTime         string `json:"startTime"`
		ContainerStatuses []struct {
			Name         string             `json:"name"`
			Ready        bool               `json:"ready"`
			RestartCount int                `json:"restartCount"`
			Image        string             `json:"image"`
			ImageID      string             `json:"imageID"`
			State        kubeContainerState `json:"state"`
			LastState    kubeContainerState `json:"lastState"`
		} `json:"containerStatuses"`
	} `json:"status"`
}

type kubeContainer struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	VolumeMounts []struct {
		Name      string `json:"name"`
		MountPath string `json:"mountPath"`
	} `json:"volumeMounts"`
	EnvFrom []struct {
		ConfigMapRef *struct {
			Name string `json:"name"`
		} `json:"configMapRef"`
		SecretRef *struct {
			Name string `json:"name"`
		} `json:"secretRef"`
	} `json:"envFrom"`
	Resources struct {
		Limits struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"limits"`
	} `json:"resources"`
	ReadinessProbe json.RawMessage `json:"readinessProbe"`
}

type kubeContainerState struct {
	Running *struct {
		StartedAt string `json:"startedAt"`
	} `json:"running"`
	Terminated *struct {
		ExitCode  int    `json:"exitCode"`
		Reason    string `json:"reason"`
		StartedAt string `json:"startedAt"`
	} `json:"terminated"`
	Waiting *struct {
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"waiting"`
}

// ParsePods turns `kubectl get pods -o json` (a List, or one Pod) into
// instances, in the order kubectl listed them. imageMatch, when non-empty,
// keeps only pods with a container whose image contains it -- the --all
// discovery filter, applied here so the same parse serves both paths.
//
// now is the run's clock, injected so a test can assert an age.
func ParsePods(doc string, now time.Time, imageMatch string) ([]Instance, error) {
	raws, err := kubeItems(doc)
	if err != nil {
		return nil, err
	}
	out := make([]Instance, 0, len(raws))
	for _, raw := range raws {
		var p kubePod
		if err := json.Unmarshal(raw, &p); err != nil {
			// One unreadable pod in a list is skipped, not fatal: the rest of the
			// report is still worth printing (fail loud vs skip).
			continue
		}
		if imageMatch != "" && !podMatchesImage(p, imageMatch) {
			continue
		}
		out = append(out, instanceFromPod(p, now))
	}
	return out, nil
}

// kubeItems returns the object documents in doc: the items of a List, or the
// single object itself.
func kubeItems(doc string) ([]json.RawMessage, error) {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return nil, fmt.Errorf("empty response")
	}
	var list kubeList
	if err := json.Unmarshal([]byte(doc), &list); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if list.Items != nil {
		return list.Items, nil
	}
	return []json.RawMessage{json.RawMessage(doc)}, nil
}

func podMatchesImage(p kubePod, match string) bool {
	for _, c := range p.Spec.Containers {
		if strings.Contains(c.Image, match) {
			return true
		}
	}
	for _, cs := range p.Status.ContainerStatuses {
		if strings.Contains(cs.Image, match) {
			return true
		}
	}
	return false
}

// connectorIndex picks the container this report is about: the one this tool
// names in the manifests it renders, else the only container there is. A pod
// with several containers and none of them named that is reported at pod level
// only -- guessing which of them is the connector would be worse than saying
// nothing about any of them.
func connectorIndex(names []string) int {
	if len(names) == 1 {
		return 0
	}
	for i, n := range names {
		if n == spec.ConnectorContainerName {
			return i
		}
	}
	return -1
}

func instanceFromPod(p kubePod, now time.Time) Instance {
	inst := Instance{Name: p.Metadata.Name, Namespace: p.Metadata.Namespace}
	c := &Container{State: StateUnknown, Health: NotApplicable, Node: p.Spec.NodeName}

	specNames := make([]string, len(p.Spec.Containers))
	for i, sc := range p.Spec.Containers {
		specNames[i] = sc.Name
	}
	statusNames := make([]string, len(p.Status.ContainerStatuses))
	for i, cs := range p.Status.ContainerStatuses {
		statusNames[i] = cs.Name
	}
	si := connectorIndex(specNames)
	ci := connectorIndex(statusNames)

	// Record which container connectorIndex settled on, so a caller that needs
	// to address it by name (kubectl's -c) reuses this decision instead of
	// making its own and risking a different answer. The spec list is preferred
	// because it exists before any container status does.
	switch {
	case si >= 0:
		inst.ContainerName = specNames[si]
	case ci >= 0:
		inst.ContainerName = statusNames[ci]
	}

	// Pod phase is the fallback for a pod with no container status yet
	// (Pending, unschedulable): there is a real state to report even before a
	// container exists.
	switch p.Status.Phase {
	case "Pending":
		c.State = StateWaiting
	case "Running":
		c.State = StateRunning
	case "Succeeded", "Failed":
		c.State = StateExited
	}
	c.StartedAt = p.Status.StartTime
	if c.StartedAt == "" {
		c.StartedAt = p.Metadata.CreationTimestamp
	}

	if ci >= 0 {
		cs := p.Status.ContainerStatuses[ci]
		c.Restarts = cs.RestartCount
		c.Image = cs.Image
		c.Digest = digestFrom(cs.ImageID)
		switch {
		case cs.State.Running != nil:
			c.State = StateRunning
			c.StartedAt = cs.State.Running.StartedAt
		case cs.State.Terminated != nil:
			c.State = StateExited
			c.ExitCode = exitCode(cs.State.Terminated.ExitCode)
			c.Reason = cs.State.Terminated.Reason
		case cs.State.Waiting != nil:
			// A back-off is the engine waiting between restarts, which is what
			// "restarting" means to an operator; every other waiting reason
			// (ContainerCreating, ImagePullBackOff, CreateContainerConfigError) is
			// a container that has not started at all.
			if strings.Contains(cs.State.Waiting.Reason, "CrashLoopBackOff") {
				c.State = StateRestarting
			} else {
				c.State = StateWaiting
			}
			c.Reason = cs.State.Waiting.Reason
		}
		// The previous termination is the useful one on a pod that is up again or
		// looping: it carries the code that killed it (137 for an OOM kill).
		if c.ExitCode == nil && cs.LastState.Terminated != nil {
			c.ExitCode = exitCode(cs.LastState.Terminated.ExitCode)
			if c.Reason == "" {
				c.Reason = cs.LastState.Terminated.Reason
			}
		}
		if cs.Ready {
			c.Ready = "yes"
		} else {
			c.Ready = "no"
		}
	}

	if si >= 0 {
		sc := p.Spec.Containers[si]
		if c.Image == "" {
			c.Image = sc.Image
		}
		if lim := sc.Resources.Limits; lim.CPU != "" || lim.Memory != "" {
			// Limits are recorded now and the used side filled in later from
			// `kubectl top`, so a details run with no metrics API still reports
			// what the pod is allowed to use.
			if lim.CPU != "" {
				c.CPU = &Resource{Limit: lim.CPU}
			}
			if lim.Memory != "" {
				c.Memory = &Resource{Limit: lim.Memory}
			}
		}
		// Readiness is only a verdict when the pod actually declares a probe;
		// without one, kubernetes reports ready as soon as the container runs,
		// which says nothing the state column has not already said.
		if len(sc.ReadinessProbe) == 0 {
			c.Ready = NotApplicable
		}
		c.Components = podComponents(p, sc)
	}

	c.Age = Age(c.StartedAt, now)
	inst.Container = c
	return inst
}

// exitCode records a termination code only when it is non-zero. A clean exit
// says nothing the state does not already say, and rendering it as
// "exited (exit code 0)" would put noise in the column that exists to carry a
// diagnosis -- while an engine that reports 0 for a container that never ran
// would otherwise look as though it had exited cleanly.
func exitCode(code int) *int {
	if code == 0 {
		return nil
	}
	return &code
}

// digestFrom extracts the image digest from a kubernetes imageID, which comes
// in several shapes ("repo@sha256:...", "docker-pullable://repo@sha256:...",
// or a bare "sha256:..." for an image with no registry digest). Everything
// before the digest is dropped: the repository is already on the image line.
func digestFrom(imageID string) string {
	if i := strings.Index(imageID, "sha256:"); i >= 0 {
		return imageID[i:]
	}
	return ""
}

// podComponents lists the objects the pod references, so the components view
// works for an instance this tool never deployed: whatever the pod spec
// mounts or reads its environment from is what gets checked, whoever created
// it. Status is left empty here for the caller to fill from a presence probe.
func podComponents(p kubePod, sc kubeContainer) []Component {
	mountPaths := make(map[string]string, len(sc.VolumeMounts))
	for _, m := range sc.VolumeMounts {
		mountPaths[m.Name] = m.MountPath
	}
	var out []Component
	for _, v := range p.Spec.Volumes {
		comp := Component{Detail: mountPaths[v.Name]}
		switch {
		case v.ConfigMap != nil:
			comp.Kind, comp.Name = "configmap", v.ConfigMap.Name
		case v.Secret != nil:
			comp.Kind, comp.Name = "secret", v.Secret.SecretName
		case v.PVC != nil:
			comp.Kind, comp.Name = "persistentvolumeclaim", v.PVC.ClaimName
		default:
			// An emptyDir or projected volume has no object to check.
			continue
		}
		out = append(out, comp)
	}
	for _, ef := range sc.EnvFrom {
		switch {
		case ef.ConfigMapRef != nil:
			out = append(out, Component{Kind: "configmap", Name: ef.ConfigMapRef.Name, Detail: "envFrom"})
		case ef.SecretRef != nil:
			out = append(out, Component{Kind: "secret", Name: ef.SecretRef.Name, Detail: "envFrom"})
		}
	}
	for _, ps := range p.Spec.ImagePullSecrets {
		out = append(out, Component{Kind: "secret", Name: ps.Name, Detail: "imagePullSecret"})
	}
	return out
}

// kubeDeployment / kubeService are the subsets the workload summary reads.
type kubeDeployment struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Replicas *int `json:"replicas"`
	} `json:"spec"`
	Status struct {
		ReadyReplicas     int `json:"readyReplicas"`
		UpdatedReplicas   int `json:"updatedReplicas"`
		AvailableReplicas int `json:"availableReplicas"`
	} `json:"status"`
}

type kubeService struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		Ports []struct {
			Port     int    `json:"port"`
			Protocol string `json:"protocol"`
		} `json:"ports"`
	} `json:"spec"`
}

// ParseDeployment reads the replica counts a rollout is judged by. Desired
// defaults to 1 when the manifest omits replicas, matching the API default (and
// spec.ParseEnv's own default).
func ParseDeployment(doc string) (*Workload, error) {
	var d kubeDeployment
	if err := json.Unmarshal([]byte(strings.TrimSpace(doc)), &d); err != nil {
		return nil, fmt.Errorf("decoding deployment: %w", err)
	}
	w := &Workload{
		Deployment: d.Metadata.Name,
		Ready:      d.Status.ReadyReplicas,
		UpToDate:   d.Status.UpdatedReplicas,
		Available:  d.Status.AvailableReplicas,
		Desired:    1,
	}
	if d.Spec.Replicas != nil {
		w.Desired = *d.Spec.Replicas
	}
	return w, nil
}

// MergeService adds the fronting service to an existing workload summary,
// creating one when the deployment could not be read so a service-only answer
// is still reported.
func MergeService(w *Workload, doc string) (*Workload, error) {
	var s kubeService
	if err := json.Unmarshal([]byte(strings.TrimSpace(doc)), &s); err != nil {
		return w, fmt.Errorf("decoding service: %w", err)
	}
	if w == nil {
		w = &Workload{}
	}
	w.Service = s.Metadata.Name
	for _, p := range s.Spec.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "TCP" // the API default, omitted from a manifest that does not set it
		}
		w.ServicePorts = append(w.ServicePorts, strconv.Itoa(p.Port)+"/"+proto)
	}
	return w, nil
}

// ObjectExists reports whether a `get <kind> <name> -o json` document describes
// a live object, and the one word worth showing about it: a volume claim's
// phase (Bound/Pending) says more than "present", and is the only status here
// that can be bad while the object exists.
func ObjectExists(doc string) (bool, string) {
	var o struct {
		Kind   string `json:"kind"`
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(doc)), &o); err != nil || o.Kind == "" {
		return false, ""
	}
	if o.Status.Phase != "" {
		return true, o.Status.Phase
	}
	return true, "present"
}

// ApplyTop fills the used side of the CPU/memory resources from
// `kubectl top pod --containers --no-headers`, whose columns are
// "<pod> <container> <cpu> <memory>". Rows for a container this report is not
// about are ignored, and a pod with no row keeps its limits and no usage.
func ApplyTop(insts []Instance, out string) {
	type usage struct{ cpu, mem string }
	byPod := map[string]usage{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 4 {
			continue
		}
		pod, container, cpu, mem := f[0], f[1], f[2], f[3]
		// Keep the connector container's row; for a single-container pod the name
		// is whatever it is, so the first row for that pod wins.
		if _, seen := byPod[pod]; seen && container != spec.ConnectorContainerName {
			continue
		}
		byPod[pod] = usage{cpu, mem}
	}
	for i := range insts {
		u, ok := byPod[insts[i].Name]
		if !ok || insts[i].Container == nil {
			continue
		}
		insts[i].Container.CPU = withUsed(insts[i].Container.CPU, u.cpu)
		insts[i].Container.Memory = withUsed(insts[i].Container.Memory, u.mem)
	}
}

// withUsed fills a resource's used side, creating it when only usage is known
// (a pod with no limits set), and computes the percentage when both sides read.
func withUsed(r *Resource, used string) *Resource {
	if used == "" {
		return r
	}
	if r == nil {
		r = &Resource{}
	}
	r.Used = used
	r.Percent = Percent(r.Used, r.Limit)
	return r
}

// ---- docker / podman ---------------------------------------------------------

// engineContainer is the subset of a `docker|podman inspect` object this report
// reads. Both engines share these paths; where they differ (the healthcheck
// key) both spellings are decoded and whichever answered is used.
type engineContainer struct {
	Name    string `json:"Name"`
	Created string `json:"Created"`
	State   struct {
		Status      string                   `json:"Status"`
		Running     bool                     `json:"Running"`
		Restarting  bool                     `json:"Restarting"`
		Paused      bool                     `json:"Paused"`
		ExitCode    int                      `json:"ExitCode"`
		StartedAt   string                   `json:"StartedAt"`
		OOMKilled   bool                     `json:"OOMKilled"`
		Error       string                   `json:"Error"`
		Health      *struct{ Status string } `json:"Health"`
		Healthcheck *struct{ Status string } `json:"Healthcheck"`
	} `json:"State"`
	RestartCount int    `json:"RestartCount"`
	Image        string `json:"Image"`
	Config       struct {
		Image  string            `json:"Image"`
		Labels map[string]string `json:"Labels"`
	} `json:"Config"`
	HostConfig struct {
		NanoCpus int64 `json:"NanoCpus"`
		Memory   int64 `json:"Memory"`
	} `json:"HostConfig"`
	Mounts []struct {
		Type        string `json:"Type"`
		Name        string `json:"Name"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
	} `json:"Mounts"`
	NetworkSettings struct {
		Networks map[string]struct{} `json:"Networks"`
	} `json:"NetworkSettings"`
	// Podman lists the secrets a container was created with; docker does not
	// (its compose secrets arrive as mounts), so an absent key is normal.
	Secrets []struct {
		Name string `json:"Name"`
	} `json:"Secrets"`
}

// ComposeProjectLabel is the label docker compose stamps on every container it
// creates, naming the project the container was brought up as part of.
const ComposeProjectLabel = "com.docker.compose.project"

// ParseInspect turns a `docker|podman inspect <names...>` array into instances,
// in the order the engine returned them. imageMatch, when non-empty, keeps only
// containers whose configured image contains it (the --all filter).
func ParseInspect(doc string, now time.Time, imageMatch string) ([]Instance, error) {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return nil, fmt.Errorf("empty response")
	}
	var cs []engineContainer
	if err := json.Unmarshal([]byte(doc), &cs); err != nil {
		return nil, fmt.Errorf("decoding inspect response: %w", err)
	}
	out := make([]Instance, 0, len(cs))
	for _, ec := range cs {
		if imageMatch != "" && !strings.Contains(ec.Config.Image, imageMatch) {
			continue
		}
		out = append(out, instanceFromInspect(ec, now))
	}
	return out, nil
}

func instanceFromInspect(ec engineContainer, now time.Time) Instance {
	// docker reports a container name with a leading slash; podman does not.
	inst := Instance{Name: strings.TrimPrefix(ec.Name, "/")}
	inst.Group = ec.Config.Labels[ComposeProjectLabel]

	c := &Container{
		Restarts:  ec.RestartCount,
		Image:     ec.Config.Image,
		StartedAt: ec.State.StartedAt,
		Ready:     NotApplicable,
		Health:    NotApplicable,
	}
	// The engine's status word is already the vocabulary this report uses, with
	// two exceptions worth normalising: "created" has not started, and
	// "dead"/"removing" are terminal states an operator reads as exited.
	switch ec.State.Status {
	case "running":
		c.State = StateRunning
	case "restarting":
		c.State = StateRestarting
	case "paused":
		c.State = StatePaused
	case "created":
		c.State = StateWaiting
	case "exited", "dead", "removing", "stopped":
		c.State = StateExited
	default:
		c.State = StateUnknown
	}
	c.ExitCode = exitCode(ec.State.ExitCode)
	switch {
	case ec.State.OOMKilled:
		c.Reason = "OOMKilled"
	case ec.State.Error != "":
		c.Reason = ec.State.Error
	}
	if h := healthStatus(ec); h != "" {
		c.Health = h
	}
	if c.StartedAt == "" {
		c.StartedAt = ec.Created
	}
	c.Age = Age(c.StartedAt, now)
	if lim := Cores(ec.HostConfig.NanoCpus); lim != "" {
		// The stats percentage is relative to the whole host, so the container's
		// own quota is the ceiling worth naming beside it -- spelled with its
		// unit here, once, rather than assembled at render time.
		c.CPU = &Resource{Limit: lim + " cpu"}
	}
	if lim := Bytes(ec.HostConfig.Memory); lim != "" {
		c.Memory = &Resource{Limit: lim}
	}
	c.Components = engineComponents(ec)
	inst.Container = c
	return inst
}

// healthStatus returns the engine's healthcheck verdict, from whichever key
// this engine version uses. An empty verdict means the container defines no
// healthcheck at all, which is the common case here: the compose and quadlet
// artifacts this tool generates declare none, so the line usually reads n/a
// unless the image itself carries a HEALTHCHECK.
func healthStatus(ec engineContainer) string {
	if ec.State.Health != nil && ec.State.Health.Status != "" {
		return ec.State.Health.Status
	}
	if ec.State.Healthcheck != nil && ec.State.Healthcheck.Status != "" {
		return ec.State.Healthcheck.Status
	}
	return ""
}

// engineComponents lists what the container references: its mounts (the
// configs and libs this tool bind-mounts, or a named volume), its networks,
// and on podman the secrets it was created with. Everything here is already
// present by definition -- the container exists with it attached -- so status
// is "attached" rather than a presence probe the engine cannot answer.
func engineComponents(ec engineContainer) []Component {
	var out []Component
	for _, m := range ec.Mounts {
		name := m.Name
		if name == "" {
			name = m.Source
		}
		kind := m.Type
		if kind == "" {
			kind = "mount"
		}
		out = append(out, Component{Kind: kind, Name: name, Status: "attached", Detail: m.Destination})
	}
	for n := range ec.NetworkSettings.Networks {
		out = append(out, Component{Kind: "network", Name: n, Status: "attached"})
	}
	for _, s := range ec.Secrets {
		out = append(out, Component{Kind: "secret", Name: s.Name, Status: "attached"})
	}
	// Networks come out of a map, so the order is not the engine's: sort the
	// whole list so a repeated run prints the same block.
	sortComponents(out)
	return out
}

// sortComponents orders a component list by kind then name, so a block built
// partly from a map (docker's networks) prints the same way on every run.
func sortComponents(cs []Component) {
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[i].Kind != cs[j].Kind {
			return cs[i].Kind < cs[j].Kind
		}
		return cs[i].Name < cs[j].Name
	})
}

// ParseImageDigest reads the registry digest from `docker|podman image inspect
// <ref>`: the first RepoDigest, which is the digest of the manifest this image
// was pulled by. An image built locally and never pushed has none, which is
// not an error -- the digest line is simply left out.
func ParseImageDigest(doc string) string {
	var imgs []struct {
		RepoDigests []string `json:"RepoDigests"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(doc)), &imgs); err != nil {
		return ""
	}
	for _, im := range imgs {
		for _, rd := range im.RepoDigests {
			if d := digestFrom(rd); d != "" {
				return d
			}
		}
	}
	return ""
}

// ApplyStats fills the used side of CPU/memory from
// `docker|podman stats --no-stream --format <name>\t<cpu%>\t<mem>\t<mem%>`.
//
// The percentages are taken as the engine computed them rather than recomputed
// here: docker's memory percentage is already relative to the container's
// limit (or the host total when it has none), and its CPU percentage is
// relative to the whole host, which no ceiling in the inspect output describes.
func ApplyStats(insts []Instance, out string) {
	type sample struct{ cpu, mem, memPct string }
	byName := map[string]sample{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(f) < 4 {
			continue
		}
		name := strings.TrimSpace(f[0])
		if name == "" {
			continue
		}
		byName[name] = sample{strings.TrimSpace(f[1]), strings.TrimSpace(f[2]), strings.TrimSpace(f[3])}
	}
	for i := range insts {
		s, ok := byName[insts[i].Name]
		if !ok || insts[i].Container == nil {
			continue
		}
		c := insts[i].Container
		if s.cpu != "" {
			if c.CPU == nil {
				c.CPU = &Resource{}
			}
			// A percentage is the whole measurement; the quota it runs against
			// was already recorded as the limit by ParseInspect.
			c.CPU.Used = s.cpu
		}
		if s.mem != "" {
			if c.Memory == nil {
				c.Memory = &Resource{}
			}
			// "512MiB / 15.6GiB" already carries both sides; keep it whole rather
			// than splitting and re-joining it into a different spelling.
			c.Memory.Used = s.mem
			c.Memory.Limit = ""
			c.Memory.Percent = s.memPct
		}
	}
}

// EngineNamesByImage filters `docker|podman ps --all --format
// <name>\t<image>` down to the containers whose image contains match. Used by
// --all, where the operator names no target and the image reference is the only
// thing that identifies a connector instance.
func EngineNamesByImage(out, match string) []string {
	var names []string
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\t")
		if len(f) < 2 {
			continue
		}
		name, image := strings.TrimSpace(f[0]), strings.TrimSpace(f[1])
		if name != "" && strings.Contains(image, match) {
			names = append(names, name)
		}
	}
	return names
}

// ---- the in-container script's report ----------------------------------------

// Line prefixes the status script emits. They are a parsing contract between
// this package and internal/statusscript: a line the script renames stops
// being reported until this list follows it, which is what
// TestAppLinePrefixesMatchScript gates.
const (
	prefixLEMode       = "leader-election mode:"
	prefixLEState      = "leader-election state:"
	prefixHealth       = "health:"
	prefixHealthDetail = "health-detail:"
	prefixUptime       = "uptime:"
	prefixVersion      = "version:"
	prefixJava         = "java:"
	prefixConfig       = "config:"
	prefixHeap         = "heap:"
	prefixNote         = "status:"
	blockWorkflows     = "workflows:"
	blockHealthComps   = "health components:"
)

// ParseApplication reads the status script's output into the application half
// of the model. The script's stdout and stderr arrive on one stream (the exec
// seam returns combined output), so its "status:" notes are interleaved with
// the report and are collected as notes here.
//
// An unrecognised line is kept as a note rather than dropped, so an instance
// carrying a newer script still reports everything it printed and an operator
// is never silently shown less than the script said.
func ParseApplication(out string) *Application {
	app := &Application{}
	lines := strings.Split(strings.ReplaceAll(out, "\r\n", "\n"), "\n")
	block := ""
	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		// An indented line belongs to the block header above it. Any other
		// indentation ends the block.
		if line != strings.TrimLeft(line, " \t") {
			body := strings.TrimSpace(line)
			switch block {
			case blockWorkflows:
				if id, state, ok := splitKV(body); ok {
					app.Workflows = append(app.Workflows, Workflow{ID: id, State: state})
					continue
				}
			case blockHealthComps:
				if name, status, ok := splitKV(body); ok {
					app.HealthComponents = append(app.HealthComponents, NameStatus{Name: name, Status: status})
					continue
				}
			}
			app.Notes = append(app.Notes, body)
			continue
		}
		block = ""
		switch {
		case line == blockWorkflows:
			block = blockWorkflows
		case line == blockHealthComps:
			block = blockHealthComps
		case strings.HasPrefix(line, prefixLEMode):
			app.LeaderElectionMode = value(line, prefixLEMode)
		case strings.HasPrefix(line, prefixLEState):
			app.LeaderElectionState = value(line, prefixLEState)
		case strings.HasPrefix(line, prefixHealthDetail):
			app.HealthDetail = value(line, prefixHealthDetail)
		case strings.HasPrefix(line, prefixHealth):
			app.Health = value(line, prefixHealth)
		case strings.HasPrefix(line, prefixUptime):
			app.Uptime = value(line, prefixUptime)
		case strings.HasPrefix(line, prefixVersion):
			app.Version = value(line, prefixVersion)
		case strings.HasPrefix(line, prefixJava):
			app.Java = value(line, prefixJava)
		case strings.HasPrefix(line, prefixConfig):
			app.Config = value(line, prefixConfig)
		case strings.HasPrefix(line, prefixHeap):
			app.Heap = parseHeap(value(line, prefixHeap))
		case strings.HasPrefix(line, prefixNote):
			app.Notes = append(app.Notes, line)
		default:
			app.Notes = append(app.Notes, line)
		}
	}
	return app
}

// value returns what follows a known prefix, trimmed. The prefix is checked by
// the caller, so a mismatch here would be a programming error rather than bad
// input.
func value(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}

// splitKV splits an indented "key: value" child line. A child with no colon is
// not a pair and is reported as a note instead.
func splitKV(s string) (string, string, bool) {
	i := strings.Index(s, ":")
	if i < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(s[:i])
	val := strings.TrimSpace(s[i+1:])
	if key == "" || val == "" {
		return "", "", false
	}
	return key, val, true
}

// parseHeap reads the script's "<used> of <max>" heap line, or a bare used
// value when the JVM reports no maximum (an unbounded heap answers -1, which
// the script leaves out rather than printing as a limit).
func parseHeap(s string) *Resource {
	used, limit, found := strings.Cut(s, " of ")
	r := &Resource{Used: heapValue(used)}
	if found {
		r.Limit = heapValue(limit)
	}
	if r.Used == "" {
		return nil
	}
	r.Percent = Percent(r.Used, r.Limit)
	return r
}

// heapValue renders one side of the heap line. The script hands the metric over
// verbatim -- a plain byte count, or the scientific notation Jackson may
// serialise a large double as -- because busybox arithmetic can be trusted with
// neither, so a bare number is converted here into the same binary units every
// other size in this report uses. A value that already carries a unit is passed
// through as the script wrote it.
func heapValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.TrimLeft(s, "0123456789.eE+-") != "" {
		return s
	}
	n, ok := ParseQuantity(s)
	if !ok {
		return s
	}
	return Bytes(int64(n))
}
