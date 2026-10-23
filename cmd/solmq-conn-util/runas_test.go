package main

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/gen"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// writeMode writes a file and sets its mode exactly (umask aside).
func writeMode(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

// skipOffUnix skips where file modes carry no Unix owner or permission bits.
func skipOffUnix(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Windows file modes carry no Unix owner or permission bits")
	}
}

// TestHostReaderFollowsTheKernelsRule pins how a container's identity is
// judged against a host file, as the kernel does: the owner bits when the uid
// owns the file, else the group bits when the gid is its group, else the
// other bits. The files belong to the test's own uid and gid, so a reader with
// different ids sees only the other bits.
func TestHostReaderFollowsTheKernelsRule(t *testing.T) {
	skipOffUnix(t)
	dir := t.TempDir()
	me, myGroup := os.Getuid(), os.Getgid()
	stranger := hostReader{uid: me + 1, gid: myGroup + 1}
	owner := hostReader{uid: me, gid: myGroup + 1}
	group := hostReader{uid: me + 1, gid: myGroup}
	for _, c := range []struct {
		mode                     os.FileMode
		stranger, owner, inGroup bool
	}{
		{0o600, false, true, false},
		{0o640, false, true, true},
		{0o604, true, true, false},
		{0o644, true, true, true},
		{0o044, true, false, true}, // the owner bits decide for the owner, however open the rest is
	} {
		p := filepath.Join(dir, c.mode.String())
		writeMode(t, p, c.mode)
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range []struct {
			name   string
			reader hostReader
			want   bool
		}{{"stranger", stranger, c.stranger}, {"owner", owner, c.owner}, {"group", group, c.inGroup}} {
			if got := r.reader.can(info, 0o4); got != r.want {
				t.Errorf("mode %04o, %s: can read = %v, want %v", c.mode, r.name, got, r.want)
			}
		}
	}
}

// TestMountChecksLabelEachPathByItsKey pins how a plan's host mounts become
// deploy's checks: labelled "<section> mounts <env.yaml key>", at the plan's
// path unchanged, with a store used as a keystore judged as a private key,
// libs.dir as a directory and any other store as a file anyone may read.
func TestMountChecksLabelEachPathByItsKey(t *testing.T) {
	got := mountChecks("docker", []gen.HostMount{
		{Field: gen.StoreField("mq-ca"), Path: "/certs/t.jks"},
		{Field: gen.StoreField("mq-client"), Path: "/certs/k.jks", Private: true},
		{Field: gen.MountLibs, Path: "/srv/jars"},
	})
	want := []mountCheck{
		{label: "docker mounts tls.stores.mq-ca.file", path: "/certs/t.jks", kind: mountPublic},
		{label: "docker mounts tls.stores.mq-client.file", path: "/certs/k.jks", kind: mountPrivate},
		{label: "docker mounts libs.dir", path: "/srv/jars", kind: mountDir},
	}
	if !slices.Equal(got, want) {
		t.Errorf("mountChecks = %+v, want %+v", got, want)
	}
	if got := mountChecks("podman", nil); len(got) != 0 {
		t.Errorf("no mounts should yield no checks, got %+v", got)
	}
}

// TestUnreadableMountsNamesWhatTheConnectorCannotRead covers the deploy
// warning: a store file and a libs directory the run-as identity cannot read,
// and a jar inside a readable directory, are each named with their mode and a
// remedy that fits what the path holds -- r and x for a directory, r for a
// truststore or jar, and the reader's group rather than everyone for a private
// key. Readable paths, non-jar files and paths that do not exist (the engine
// reports those) are not named.
func TestUnreadableMountsNamesWhatTheConnectorCannotRead(t *testing.T) {
	skipOffUnix(t)
	dir := t.TempDir()
	open, trust, key := filepath.Join(dir, "open.jks"), filepath.Join(dir, "trust.jks"), filepath.Join(dir, "key.jks")
	writeMode(t, open, 0o644)
	writeMode(t, trust, 0o600)
	writeMode(t, key, 0o600)
	libs := filepath.Join(dir, "libs")
	if err := os.Mkdir(libs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(libs, 0o755); err != nil { // umask aside
		t.Fatal(err)
	}
	writeMode(t, filepath.Join(libs, "ok.jar"), 0o644)
	writeMode(t, filepath.Join(libs, "shut.jar"), 0o600)
	writeMode(t, filepath.Join(libs, "notes.txt"), 0o600)
	shutDir := filepath.Join(dir, "shut")
	if err := os.Mkdir(shutDir, 0o700); err != nil {
		t.Fatal(err)
	}
	ra := spec.RunAsID{UID: os.Getuid() + 1, GID: os.Getgid() + 1}
	g := strconv.Itoa(ra.GID)
	got := sharedIDs(ra).unreadable("docker", []mountCheck{
		{label: "docker mounts tls.stores.ca.file", path: open, kind: mountPublic},
		{label: "docker mounts tls.stores.ca.file", path: trust, kind: mountPublic},
		{label: "docker mounts tls.stores.client.file", path: key, kind: mountPrivate},
		{label: "docker mounts libs.dir", path: libs, kind: mountDir},
		{label: "docker mounts libs.dir", path: shutDir, kind: mountDir},
		{label: "docker mounts tls.stores.ca.file", path: filepath.Join(dir, "missing.jks"), kind: mountPublic},
	})
	const because = " is not readable by the connector, which runs as "
	want := []struct{ prefix, suffix string }{
		{"docker mounts tls.stores.ca.file " + trust + " (mode 0600)" + because + ra.String(),
			"Make it readable: chmod o+r, or give it to group " + g + " with g+r -- or change docker.run-as"},
		{"docker mounts tls.stores.client.file " + key + " (mode 0600)" + because + ra.String(),
			"Give it to group " + g + " and chmod g+r rather than opening a private key to everyone, or set docker.run-as to its owner's uid"},
		{"docker mounts libs.dir jar " + filepath.Join(libs, "shut.jar") + " (mode 0600)",
			"Make it readable: chmod o+r, or give it to group " + g + " with g+r -- or change docker.run-as"},
		{"docker mounts libs.dir " + shutDir + " (mode 0700)",
			"Make it listable: chmod o+rx, or give it to group " + g + " with g+rx -- or change docker.run-as"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d warnings, want %d: %q", len(got), len(want), got)
	}
	for i, w := range want {
		if !strings.HasPrefix(got[i], w.prefix) || !strings.HasSuffix(got[i], w.suffix) {
			t.Errorf("warning %d = %q, want it to start %q and end %q", i, got[i], w.prefix, w.suffix)
		}
	}
}

// TestUnreadableMountsJudgesRootByTheModeBits pins that root gets no pass: the
// artifacts drop every capability, so the connector running as uid 0 -- the
// image's own user under run-as image -- cannot read a 0600 file another uid
// owns any more than a non-root uid can. The platform gate checks on Linux
// only, where the engine sees the host's modes.
func TestUnreadableMountsJudgesRootByTheModeBits(t *testing.T) {
	skipOffUnix(t)
	if os.Getuid() == 0 {
		t.Skip("the test's files must belong to a uid other than root")
	}
	dir := t.TempDir()
	shut, open := filepath.Join(dir, "key.jks"), filepath.Join(dir, "trust.jks")
	writeMode(t, shut, 0o600)
	writeMode(t, open, 0o644)
	checks := []mountCheck{
		{label: "docker mounts tls.stores.client.file", path: shut, kind: mountPrivate},
		{label: "docker mounts tls.stores.ca.file", path: open, kind: mountPublic},
	}
	for _, c := range []struct {
		ra  spec.RunAsID
		who string
	}{
		{spec.RunAsID{Image: true}, "root, the image's own user (run-as image)"},
		{spec.RunAsID{UID: 0, GID: 0}, "0:0"},
	} {
		got := sharedIDs(c.ra).unreadable("docker", checks)
		if len(got) != 1 || !strings.Contains(got[0], shut+" (mode 0600) is not readable by the connector, which runs as "+c.who+" with no capability") {
			t.Errorf("run-as %s: want one warning for the 0600 file naming %q, got %q", c.ra, c.who, got)
		}
		gated := unreadableMounts(sharedIDs(c.ra), "docker", checks)
		if (runtime.GOOS == "linux" && len(gated) != 1) || (runtime.GOOS != "linux" && gated != nil) {
			t.Errorf("run-as %s on %s: unreadableMounts = %q", c.ra, runtime.GOOS, gated)
		}
	}
}

// TestRootlessReaderSeesTheInvokersGroup pins rootless podman's mapping: a
// non-zero container uid never owns a host file, and container gid 0 is the
// invoking user's own group -- so a 0640 jar of the invoker's is readable
// through the group bits, a 0600 one is not, and a non-zero gid maps to no host
// group. Container root -- uid 0, or the image's own -- is the invoking user.
func TestRootlessReaderSeesTheInvokersGroup(t *testing.T) {
	skipOffUnix(t)
	dir := t.TempDir()
	group, private := filepath.Join(dir, "g.jar"), filepath.Join(dir, "p.jar")
	writeMode(t, group, 0o640)
	writeMode(t, private, 0o600)
	gi, _ := os.Stat(group)
	pi, _ := os.Stat(private)
	r := rootlessIDs(spec.RunAsID{UID: 1000, GID: 0})
	if !r.can(gi, 0o4) || r.can(pi, 0o4) {
		t.Errorf("rootless gid 0: 0640 readable = %v (want true), 0600 readable = %v (want false)", r.can(gi, 0o4), r.can(pi, 0o4))
	}
	if r2 := rootlessIDs(spec.RunAsID{UID: 1000, GID: 1000}); r2.gid != -1 || r2.uid != -1 || r2.can(gi, 0o4) {
		t.Errorf("rootless gid 1000 must map to no host group: %+v", r2)
	}
	// Container root, uid 0 or the image's own, is the invoking user.
	for _, ra := range []spec.RunAsID{{UID: 0, GID: 0}, {Image: true}} {
		if r3 := rootlessIDs(ra); r3.uid != os.Getuid() || r3.gid != os.Getgid() || !r3.can(pi, 0o4) {
			t.Errorf("rootless %s must be the invoking user and read its 0600 file: %+v", ra, r3)
		}
	}
}

// fakeHostFiles is a read func over fixed file contents; a missing name is an
// unreadable file.
func fakeHostFiles(files map[string]string) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		if data, ok := files[name]; ok {
			return []byte(data), nil
		}
		return nil, os.ErrNotExist
	}
}

// TestSubIDWarningsNeedARangeWideEnough pins the rootless deploy check: a
// non-zero run-as uid or gid needs a subordinate range at least that wide in
// /etc/subuid or /etc/subgid, matched by the invoking user's name or uid, with
// several ranges adding up. Id 0 (the invoker itself) and run-as image need
// none, and an nsswitch.conf that takes subids from elsewhere leaves the files
// unjudged.
func TestSubIDWarningsNeedARangeWideEnough(t *testing.T) {
	const stock = "deployer:100000:65536\n"
	for _, c := range []struct {
		name  string
		ra    spec.RunAsID
		files map[string]string
		want  []string // one substring per warning, in order
	}{
		{"default fits a stock range", spec.RunAsID{UID: 1000}, map[string]string{"/etc/subuid": stock}, nil},
		{"the range's last id fits", spec.RunAsID{UID: 65536, GID: 65536}, map[string]string{"/etc/subuid": stock, "/etc/subgid": stock}, nil},
		{"one past the range", spec.RunAsID{UID: 65537}, map[string]string{"/etc/subuid": stock},
			[]string{"podman.run-as 65537:0 needs uid 65537 mapped, but /etc/subuid grants deployer 65536 subordinate ids, which map container ids up to 65536, so the unit will not start. Widen the range (sudo usermod --add-subuids, then podman system migrate)"}},
		{"ranges add up", spec.RunAsID{UID: 70000}, map[string]string{"/etc/subuid": "deployer:100000:40000\nother:200000:65536\ndeployer:300000:40000\n"}, nil},
		{"matched by uid", spec.RunAsID{UID: 1000}, map[string]string{"/etc/subuid": "4242:100000:65536\n"}, nil},
		{"no range at all", spec.RunAsID{UID: 1000}, map[string]string{"/etc/subuid": "other:100000:65536\n"},
			[]string{"podman.run-as 1000:0 needs a subordinate uid range, but /etc/subuid grants deployer none, so rootless podman cannot map uid 1000 and the unit will not start. Add a range no other user holds (for example sudo usermod --add-subuids 100000-165535 deployer, then podman system migrate)"}},
		{"no files at all, gid too", spec.RunAsID{UID: 1000, GID: 1000}, nil,
			[]string{"needs a subordinate uid range, but /etc/subuid grants deployer none", "needs a subordinate gid range, but /etc/subgid grants deployer none"}},
		{"root needs no range", spec.RunAsID{UID: 0, GID: 0}, nil, nil},
		{"image needs no range", spec.RunAsID{Image: true}, nil, nil},
		{"gid 0 needs no subgid", spec.RunAsID{UID: 1000}, map[string]string{"/etc/subuid": stock}, nil},
		{"files named by nsswitch", spec.RunAsID{UID: 1000}, map[string]string{"/etc/nsswitch.conf": "passwd: files\nsubid:   files\n"},
			[]string{"needs a subordinate uid range"}},
		{"subids from sssd", spec.RunAsID{UID: 1000}, map[string]string{"/etc/nsswitch.conf": "# subid: files\nsubid: sss\n"}, nil},
	} {
		got := subIDWarnings(c.ra, fakeHostFiles(c.files), 4242, "deployer")
		if len(got) != len(c.want) {
			t.Errorf("%s: got %d warnings, want %d: %q", c.name, len(got), len(c.want), got)
			continue
		}
		for i, w := range c.want {
			if !strings.Contains(got[i], w) {
				t.Errorf("%s: warning %d = %q, want it to contain %q", c.name, i, got[i], w)
			}
		}
	}
	if currentUserName() == "" {
		t.Error("currentUserName must fall back to the uid, never return empty")
	}
}
