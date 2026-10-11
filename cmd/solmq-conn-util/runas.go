package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// mountKind is what a bind-mounted host path holds, which decides both what
// the container needs of it and what to suggest when it cannot get it.
type mountKind int

const (
	mountPublic  mountKind = iota // a file anyone may read: a truststore, a jar
	mountPrivate                  // a file holding a private key: the keystore
	mountDir                      // a directory mounted whole: libs.dir
)

// mountCheck is one host path a container bind-mounts and has to read.
type mountCheck struct {
	label string // what env.yaml calls it, e.g. "docker mounts tls.truststore.file"
	path  string // the resolved host path
	kind  mountKind
}

// mountChecks turns the host paths a plan bind-mounts into deploy's checks,
// labelled as env.yaml names them ("docker mounts libs.dir"): the keystore
// holds a private key, libs.dir is a directory of jars, and anything else is a
// file anyone may read.
func mountChecks(section string, ms []gen.HostMount) []mountCheck {
	out := make([]mountCheck, 0, len(ms))
	for _, m := range ms {
		kind := mountPublic
		switch m.Field {
		case gen.MountKeystore:
			kind = mountPrivate
		case gen.MountLibs:
			kind = mountDir
		}
		out = append(out, mountCheck{label: section + " mounts " + m.Field, path: m.Path, kind: kind})
	}
	return out
}

// hostReader is the identity a container's process reads its bind mounts as,
// in host terms. On docker and rootful podman the container shares the host's
// ids, so the run-as uid and gid are the host's own -- and run-as image is the
// image's root, uid 0. Rootless podman maps container id 0 to the invoking user
// and every other id into the invoker's subordinate ranges, which own no host
// file. An id of -1 matches no file.
//
// Root gets no exemption: the artifacts drop every capability, so even uid 0
// has no CAP_DAC_OVERRIDE and is held to the mode bits like any other uid.
// run-as image is judged as uid 0 and gid 0 because the connector image
// declares no USER, so its own user is root.
type hostReader struct {
	runAs    spec.RunAsID
	uid, gid int
}

// sharedIDs is the reader for an engine that runs the container with the
// host's ids: docker, and podman deployed as root.
func sharedIDs(ra spec.RunAsID) hostReader {
	return hostReader{runAs: ra, uid: ra.UID, gid: ra.GID}
}

// rootlessIDs is the reader for rootless podman: container id 0 is the
// invoking user (and group), any other id a subordinate one no host file
// belongs to.
func rootlessIDs(ra spec.RunAsID) hostReader {
	h := hostReader{runAs: ra, uid: -1, gid: -1}
	if ra.Image || ra.UID == 0 {
		h.uid = os.Getuid()
	}
	if ra.Image || ra.GID == 0 {
		h.gid = os.Getgid()
	}
	return h
}

// who names the identity for a message.
func (h hostReader) who() string {
	if h.runAs.Image {
		return "root, the image's own user (run-as image)"
	}
	return h.runAs.String()
}

// unreadableMounts returns a warning for each mounted host path the
// container's run-as identity cannot read, for deploy to print before it
// starts anything: a connector that cannot read its truststore or its MQ jars
// starts anyway and fails later, far from the cause. Nothing is checked off
// Linux, where Docker Desktop's file sharing does not carry the host's modes
// into the container.
func unreadableMounts(h hostReader, section string, checks []mountCheck) []string {
	if runtime.GOOS != "linux" {
		return nil
	}
	return h.unreadable(section, checks)
}

// unreadable is unreadableMounts without the platform gate. A path that
// cannot be stat'ed is skipped: whether it exists is the engine's to report
// when it mounts it.
func (h hostReader) unreadable(section string, checks []mountCheck) []string {
	var out []string
	warn := func(label, path string, info fs.FileInfo, kind mountKind) {
		out = append(out, fmt.Sprintf("%s %s (mode %04o) is not readable by the connector, which runs as %s with no capability to bypass file permissions: it would start and then fail when it needs it. %s",
			label, path, info.Mode().Perm(), h.who(), h.remedy(kind, section)))
	}
	for _, c := range checks {
		info, err := os.Stat(c.path)
		if err != nil {
			continue
		}
		if c.kind != mountDir {
			if !h.can(info, 0o4) {
				warn(c.label, c.path, info, c.kind)
			}
			continue
		}
		// A directory has to be listable (r) and enterable (x) for the JVM to
		// find the jars in it.
		if !h.can(info, 0o5) {
			warn(c.label, c.path, info, mountDir)
			continue
		}
		entries, err := os.ReadDir(c.path)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jar") {
				continue
			}
			if fi, err := e.Info(); err == nil && !h.can(fi, 0o4) {
				warn(c.label+" jar", filepath.Join(c.path, e.Name()), fi, mountPublic)
			}
		}
	}
	return out
}

// remedy is the fix to suggest for a path of kind the reader cannot use: open
// a directory with r and x, a public file with r, and hand a private key to the
// reader's group rather than to everyone. The group is named only when the
// reader's gid is a host group at all.
func (h hostReader) remedy(kind mountKind, section string) string {
	group := ""
	if h.gid >= 0 {
		group = strconv.Itoa(h.gid)
	}
	switch {
	case kind == mountDir && group != "":
		return fmt.Sprintf("Make it listable: chmod o+rx, or give it to group %s with g+rx -- or change %s.run-as", group, section)
	case kind == mountDir:
		return fmt.Sprintf("Make it listable: chmod o+rx -- or change %s.run-as", section)
	case kind == mountPrivate && group != "":
		return fmt.Sprintf("Give it to group %s and chmod g+r rather than opening a private key to everyone, or set %s.run-as to its owner's uid", group, section)
	case kind == mountPrivate:
		return fmt.Sprintf("Set %s.run-as to a gid it can be given to (chgrp, chmod g+r) rather than opening a private key to everyone", section)
	case group != "":
		return fmt.Sprintf("Make it readable: chmod o+r, or give it to group %s with g+r -- or change %s.run-as", group, section)
	default:
		return fmt.Sprintf("Make it readable: chmod o+r -- or change %s.run-as", section)
	}
}

// can reports whether h may access info for want (an "other"-position mask:
// 4 read, 1 execute) by the kernel's rule: the owner bits when h's uid owns
// the file, else the group bits when h's gid is its group, else the other bits.
// Supplementary groups are not counted; the container is given none.
func (h hostReader) can(info fs.FileInfo, want fs.FileMode) bool {
	perm := info.Mode().Perm()
	uid, gid, ok := fileOwner(info)
	switch {
	case ok && uid == h.uid:
		return perm&(want<<6) == want<<6
	case ok && gid == h.gid:
		return perm&(want<<3) == want<<3
	default:
		return perm&want == want
	}
}

// subIDWarnings checks that rootless podman can map the run-as ids at all.
// Container id 0 is the invoking user and ids 1..n are the n subordinate ids
// /etc/subuid and /etc/subgid grant that user (several ranges add up), so a
// non-zero id needs a range at least that wide; with none at all the user
// namespace holds only the invoker, and the unit fails to start where root in
// the container used to work. A grant can also come from elsewhere (an NSS
// subid plugin), which is why this only warns. read is os.ReadFile; uid and
// name identify the invoking user.
func subIDWarnings(ra spec.RunAsID, read func(string) ([]byte, error), uid int, name string) []string {
	if ra.Image || (ra.UID == 0 && ra.GID == 0) || !subIDsFromFiles(read) {
		return nil
	}
	var out []string
	for _, c := range []struct {
		file, what string
		id         int
	}{{"/etc/subuid", "uid", ra.UID}, {"/etc/subgid", "gid", ra.GID}} {
		if c.id == 0 {
			continue // id 0 is the invoking user or group itself
		}
		n := subIDCount(read, c.file, uid, name)
		switch {
		case n >= c.id:
			continue
		case n == 0:
			out = append(out, fmt.Sprintf("podman.run-as %s needs a subordinate %s range, but %s grants %s none, so rootless podman cannot map %s %d and the unit will not start. Add a range no other user holds (for example sudo usermod --add-sub%ss 100000-165535 %s, then podman system migrate), deploy as root, or set podman.run-as: image",
				ra, c.what, c.file, name, c.what, c.id, c.what, name))
		default:
			out = append(out, fmt.Sprintf("podman.run-as %s needs %s %d mapped, but %s grants %s %d subordinate ids, which map container ids up to %d, so the unit will not start. Widen the range (sudo usermod --add-sub%ss, then podman system migrate), pick a smaller id, or deploy as root",
				ra, c.what, c.id, c.file, name, n, n, c.what))
		}
	}
	return out
}

// subIDCount totals the subordinate ids file grants the user named name (or
// numbered uid): lines are owner:start:count, and an owner may hold several.
// An unreadable file grants none.
func subIDCount(read func(string) ([]byte, error), file string, uid int, name string) int {
	data, err := read(file)
	if err != nil {
		return 0
	}
	total := 0
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Split(strings.TrimSpace(line), ":")
		if len(f) != 3 || (f[0] != name && f[0] != strconv.Itoa(uid)) {
			continue
		}
		if n, err := strconv.Atoi(f[2]); err == nil && n > 0 {
			total += n
		}
	}
	return total
}

// subIDsFromFiles reports whether the subordinate ids come from /etc/subuid
// and /etc/subgid: yes unless /etc/nsswitch.conf names another source on a
// subid: line. shadow-utils reads one source, so an SSSD or other plugin there
// replaces the files, and their contents say nothing.
func subIDsFromFiles(read func(string) ([]byte, error)) bool {
	data, err := read("/etc/nsswitch.conf")
	if err != nil {
		return true
	}
	for _, line := range strings.Split(string(data), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(key) != "subid" {
			continue
		}
		src := strings.Fields(val)
		return len(src) == 0 || (len(src) == 1 && src[0] == "files")
	}
	return true
}

// currentUserName is the invoking user's login name, which /etc/subuid lines
// usually start with, or its uid in decimal when the name cannot be looked up
// (subIDCount matches either).
func currentUserName() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return strconv.Itoa(os.Getuid())
}
