package spec

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultRunAs is the identity the connector container runs as on every
// platform unless that platform's section says otherwise. The connector image
// declares no USER, so left to itself it runs as root -- root on the host, too,
// for docker and rootful podman, which run containers without a user
// namespace. 1000 is an ordinary unprivileged uid that sits inside a stock
// rootless-podman subuid range (65536 ids), and gid 0 is the group the image's
// own files belong to, so everything the image ships stays readable.
const DefaultRunAs = "1000:0"

// RunAsImage is the run-as value that keeps the image's own user: no identity
// is written into the artifact, so the engine decides -- root for the
// connector image, or on OpenShift whatever the SCC assigns.
const RunAsImage = "image"

// maxRunAsID is the largest uid or gid a run-as may name: kubernetes caps
// runAsUser, runAsGroup and fsGroup at 2147483647, and docker, podman and the
// kernel accept that range as well.
const maxRunAsID = 2147483647

// RunAsID is a parsed run-as value: either Image, or a uid and gid.
type RunAsID struct {
	Image    bool // keep the image's own user; UID and GID are unset
	UID, GID int
}

// String renders the id the way the artifacts write it: "uid:gid", or "image".
func (r RunAsID) String() string {
	if r.Image {
		return RunAsImage
	}
	return strconv.Itoa(r.UID) + ":" + strconv.Itoa(r.GID)
}

// ParseRunAs reads a platform section's run-as: "image", or a decimal uid with
// an optional ":gid" (a bare uid runs with gid 0, the image's group). The ids
// are written into compose files, quadlet units and manifests as-is, so the
// grammar is strict -- digits only, no sign, no leading zero, no whitespace --
// and a value that fails it never reaches an artifact.
func ParseRunAs(s string) (RunAsID, error) {
	if s == RunAsImage {
		return RunAsID{Image: true}, nil
	}
	uidPart, gidPart, hasGID := strings.Cut(s, ":")
	uid, err := parseRunAsID(uidPart)
	if err != nil {
		return RunAsID{}, fmt.Errorf("the uid %w", err)
	}
	gid := 0
	if hasGID {
		if gid, err = parseRunAsID(gidPart); err != nil {
			return RunAsID{}, fmt.Errorf("the gid %w", err)
		}
	}
	return RunAsID{UID: uid, GID: gid}, nil
}

// parseRunAsID reads one id of a run-as: decimal digits, no leading zero, at
// most maxRunAsID.
func parseRunAsID(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("is empty")
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("%q is not a decimal number", s)
		}
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("%q has a leading zero", s)
	}
	n, err := strconv.Atoi(s)
	if err != nil || n > maxRunAsID {
		return 0, fmt.Errorf("%q is larger than %d", s, maxRunAsID)
	}
	return n, nil
}
