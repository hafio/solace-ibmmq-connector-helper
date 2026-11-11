package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hafio/go-deploy-utility/execguard"
	"github.com/hafio/go-deploy-utility/rootless"
	gdurunner "github.com/hafio/go-deploy-utility/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/runner"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// hostFacts is what the podman paths read from the process and the host: the
// effective uid, the environment, the home directory, the subordinate-id files
// and the user's name. A field of app rather than a lookup of its own, so a test
// supplies its own rather than depending on the euid or the files of the machine
// it runs on.
type hostFacts struct {
	scope    runner.ScopeEnv
	readFile func(string) ([]byte, error)
	userName func() string
}

// liveHost reads the facts above from this process and machine; it is what
// liveApp puts in app.host.
func liveHost() hostFacts {
	return hostFacts{scope: runner.LiveScopeEnv(), readFile: os.ReadFile, userName: currentUserName}
}

// ensureLinger makes the rootless user's manager outlive its login, enabling
// linger for the user's own account when it is off. Deploy runs it before it
// writes anything, because a unit that stops at logout and never starts at boot
// is a deploy that looks finished and is not.
func ensureLinger(r gdurunner.Runner, sc runner.QuadletScope, h hostFacts) error {
	name := h.userName()
	enabled, err := runner.EnsureLinger(bg(), r, sc, name)
	if err != nil {
		return err
	}
	if enabled {
		fmt.Fprintf(os.Stderr, "note: enabled linger for %s, so the rootless unit keeps running after logout and starts at boot\n", name)
	}
	return nil
}

// checkLinger is ensureLinger for validate, which changes nothing: it refuses a
// user that does not linger, and a probe that cannot run.
func checkLinger(r gdurunner.Runner, sc runner.QuadletScope) error {
	state, err := runner.ProbeLinger(bg(), r, sc)
	if err != nil {
		return err
	}
	return rootless.CheckLinger(sc.Euid, state)
}

// checkSubIDs refuses a numeric run-as whose uid or gid the invoking user's
// subordinate ranges do not reach. Rootless podman maps container id 0 to the
// user itself and every other id into the ranges, so an id outside them cannot be
// mapped and the unit would not start. podman is asked for the mapping rather
// than the files read, since a range can also come from an NSS source such as
// SSSD. run-as: image and id 0 need no range. Deploy and validate share it.
func checkSubIDs(r gdurunner.Runner, p *spec.Podman, ra spec.RunAsID, h hostFacts, approvals execguard.Approvals) error {
	if ra.Image {
		return nil
	}
	asWritten := p.RunAs
	if asWritten == "" {
		asWritten = ra.String()
	}
	for _, c := range []struct {
		file string
		gid  bool
		id   int
	}{{"subuid", false, ra.UID}, {"subgid", true, ra.GID}} {
		if c.id == 0 {
			continue // id 0 is the invoking user or group itself
		}
		out, err := runner.PodmanIDMap(bg(), r, p.Command, c.gid, approvals)
		if err != nil {
			return fmt.Errorf("cannot check podman.run-as %s against this user's %s ranges: %w; fix what stopped "+
				"podman, or set podman.run-as: image", asWritten, c.file, err)
		}
		if err := rootless.CheckIDMap(out, c.file, c.id, "podman.run-as", asWritten); err != nil {
			return subIDRefusal(err, ra, h)
		}
	}
	return nil
}

// subIDRefusal is cause followed by the fix: the usermod line for a range no
// other user holds, computed from /etc/subuid and /etc/subgid, then this tool's own
// alternative. With either file unreadable the line says the range was not
// computed.
func subIDRefusal(cause error, ra spec.RunAsID, h hostFacts) error {
	var files strings.Builder
	computed := true
	for _, f := range []string{"/etc/subuid", "/etc/subgid"} {
		b, err := h.readFile(f)
		if err != nil {
			computed = false
			continue
		}
		files.Write(b)
		files.WriteByte('\n')
	}
	start, end := rootless.NextFreeRange(files.String(), max(ra.UID, ra.GID)+1)
	return fmt.Errorf("%w.\n%s\n  Or set podman.run-as to an id inside the range, or to image, which keeps the "+
		"image's own user", cause, rootless.SubIDFix(h.userName(), start, end, computed))
}

// rootlessHostErrors are validate's host errors for a rootless podman unit on
// Linux: linger off or unreadable, and a run-as outside the subordinate ranges.
// Both are reported, not just the first.
func rootlessHostErrors(r gdurunner.Runner, p *spec.Podman, ra spec.RunAsID, sc runner.QuadletScope, h hostFacts) []error {
	var errs []error
	if err := checkLinger(r, sc); err != nil {
		errs = append(errs, err)
	}
	// validate takes no --allow-command, so a command that needs an approval is
	// refused here and reported as the host error it is.
	var noApprovals execguard.Approvals
	if err := checkSubIDs(r, p, ra, h, noApprovals); err != nil {
		errs = append(errs, err)
	}
	return errs
}
