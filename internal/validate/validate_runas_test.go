package validate

import (
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// runAsCtx is a valid context for one platform section with its run-as set.
func runAsCtx(section, runAs string) Context {
	c := Context{Workflows: wfOK(), Defaults: &spec.Defaults{}, Image: imageOK()}
	switch section {
	case PlatformKubernetes:
		c.Kube = &spec.Kubernetes{Deployment: baseKubeDeploy(), Command: spec.DefaultKubeCommand, RunAs: runAs}
		c.CheckKubernetes = true
	case PlatformDocker:
		d := dockerOK()
		d.RunAs = runAs
		c.Docker, c.CheckDocker = d, true
	case PlatformPodman:
		p := podmanOK()
		p.RunAs = runAs
		c.Podman, c.CheckPodman = p, true
	}
	return c
}

// TestRunAsCheckedOnEveryPlatform pins the run-as findings section by section:
// the default, an unset value, image and an explicit non-root id are quiet; a
// value outside the grammar is an error naming the section and the reason; and
// uid 0 warns that the connector runs as root -- which on docker and rootful
// podman, run without a user namespace, is root on the host as well.
func TestRunAsCheckedOnEveryPlatform(t *testing.T) {
	for _, section := range []string{PlatformKubernetes, PlatformDocker, PlatformPodman} {
		t.Run(section, func(t *testing.T) {
			for _, v := range []string{"", spec.DefaultRunAs, spec.RunAsImage, "1001:1001"} {
				errs, warns := Run(runAsCtx(section, v))
				if hasErr(errs, "run-as") || hasErr(warns, "run-as") {
					t.Errorf("run-as %q should be quiet, got errors %v, warnings %v", v, errs, warns)
				}
			}
			errs, _ := Run(runAsCtx(section, "root"))
			want := section + `.run-as "root" is not a uid[:gid] or "image": the uid "root" is not a decimal number. Write a decimal uid with an optional :gid, such as "1000:0" (the default), or "image" to keep the image's own user`
			if !hasErr(errs, want) {
				t.Errorf("want %q, got %v", want, errs)
			}
			_, warns := Run(runAsCtx(section, "0:0"))
			if !hasErr(warns, section+`.run-as "0:0" runs the connector as root`) {
				t.Errorf("uid 0 should warn, got %v", warns)
			}
			if onHost := hasErr(warns, "so that is root on the host"); onHost != (section != PlatformKubernetes) {
				t.Errorf("host-root note present = %v, want it only on docker and podman: %v", onHost, warns)
			}
		})
	}
}

// TestPodmanRunAsBeyondTheSubuidRangeWarns pins the podman-only validate
// warning: rootless podman maps container ids 1..65536 into a stock
// 65536-wide subordinate range, so an id past 65536 is unmapped and the unit
// would not start. 65536 itself still fits, the gid counts as well as the uid,
// docker and kubernetes, which map no ids, stay quiet -- and generate and
// deploy (no Lint) stay quiet too, deploy checking the host's real ranges.
func TestPodmanRunAsBeyondTheSubuidRangeWarns(t *testing.T) {
	for _, c := range []struct {
		section, runAs string
		lint, warns    bool
	}{
		{PlatformPodman, "65536:65536", true, false},
		{PlatformPodman, "65537", true, true},
		{PlatformPodman, "1000:70000", true, true},
		{PlatformPodman, spec.RunAsImage, true, false},
		{PlatformPodman, "65537", false, false},
		{PlatformDocker, "70000:70000", true, false},
		{PlatformKubernetes, "70000:70000", true, false},
	} {
		ctx := runAsCtx(c.section, c.runAs)
		ctx.Lint = c.lint
		_, warns := Run(ctx)
		if got := hasErr(warns, "names an id above 65536"); got != c.warns {
			t.Errorf("%s run-as %q (lint %v): subuid warning = %v, want %v (%v)", c.section, c.runAs, c.lint, got, c.warns, warns)
		}
	}
}
