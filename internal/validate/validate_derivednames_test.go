package validate

import (
	"strings"
	"testing"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// retiredNamesKube is an instance whose env.yaml predates derived names: every
// object the tool creates still carries the name: key it used to be given.
func retiredNamesKube() *spec.Kubernetes {
	return &spec.Kubernetes{
		Deployment: spec.Deployment{Name: "solmq", Namespace: "ns", Replicas: 1},
		Secrets: spec.Secrets{
			Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{Name: "solmq-credentials"}},
			Stores:      &spec.StoresSecret{Create: &spec.StoreCreate{Name: "solmq-tls"}},
			ImagePull:   &spec.ImagePullSecret{Name: "regcred", Create: true},
		},
		Libs: &spec.Libs{PVC: &spec.LibsPVC{Create: &spec.PVCCreate{
			Name: "jar-libs-pvc", Storage: "1Gi", NFS: spec.NFS{Server: "nfs1", Path: "/libs"},
		}}},
	}
}

// registryImage is imageOK plus the registry account an image-pull create
// needs, so the retired-name fixture validates clean apart from its names.
func registryImage() *spec.Image {
	img := imageOK()
	img.User = "svc"
	img.PassEnv = "REGISTRY_PASSWORD"
	return img
}

// TestRetiredCreateNamesReportedOnlyByValidate pins the upgrade contract for
// the name keys the tool no longer honours. generate and deploy ignore them
// quietly, so an env.yaml written before names were derived keeps deploying;
// validate alone (Lint) reports each one as no longer accepted, naming the
// object it no longer controls and how to remove it.
func TestRetiredCreateNamesReportedOnlyByValidate(t *testing.T) {
	ctx := func(lint bool) Context {
		return Context{
			Workflows: wfOK(), Defaults: defsWithStores(), Image: registryImage(), Kube: retiredNamesKube(),
			CheckKubernetes: true, Lint: lint, Env: func(string) (string, bool) { return "v", true },
		}
	}

	errs, _ := Run(ctx(false))
	if hasErr(errs, "no longer accepted") {
		t.Errorf("generate and deploy must ignore the retired names quietly, got %v", errs)
	}

	errs, _ = Run(ctx(true))
	for _, c := range []struct {
		field, derived, ignored, fix string
	}{
		{"kubernetes.secrets.credentials.create.name", "solmq-credentials", "solmq-credentials", "Replace the create: block with create: true"},
		{"kubernetes.secrets.stores.create.name", "solmq-stores", "solmq-tls", "Replace the create: block with create: true"},
		{"kubernetes.secrets.image-pull.name", "solmq-image-pull", "regcred", "remove create instead"},
		{"libs.pvc.create.name", "solmq-libs", "jar-libs-pvc", "Remove the name key"},
	} {
		var msg string
		for _, e := range errs {
			if strings.Contains(e.Msg, c.field+" is no longer accepted") {
				msg = e.Msg
			}
		}
		if msg == "" {
			t.Errorf("validate must report %s as no longer accepted, got %v", c.field, errs)
			continue
		}
		for _, w := range []string{`"` + c.derived + `"`, `"` + c.ignored + `"`, c.fix} {
			if !strings.Contains(msg, w) {
				t.Errorf("%s: message %q should contain %q", c.field, msg, w)
			}
		}
	}
}

// TestRetiredCreateNamesNeedAKey pins what validate does NOT report: a created
// object without a name key is the new shape, and an image-pull name without
// create is a reference to the operator's own Secret -- still the only way to
// name one -- so neither is a retired key.
func TestRetiredCreateNamesNeedAKey(t *testing.T) {
	k := retiredNamesKube()
	k.Secrets.Credentials.Create.Name = ""
	k.Secrets.Stores.Create.Name = ""
	k.Secrets.ImagePull = &spec.ImagePullSecret{Name: "regcred"}
	k.Libs.PVC.Create.Name = ""
	errs, _ := Run(Context{
		Workflows: wfOK(), Defaults: defsWithStores(), Image: imageOK(), Kube: k,
		CheckKubernetes: true, Lint: true, Env: func(string) (string, bool) { return "v", true },
	})
	if hasErr(errs, "no longer accepted") {
		t.Errorf("nothing here is a retired key, got %v", errs)
	}
}

// TestDerivedNamesMustFitALabel covers what deriving the names costs: every
// created object's name is deployment.name plus a suffix, so the longest one
// the config actually asks for has to fit a DNS-1123 label -- checked on every
// run, since the API server would otherwise refuse it part-way through an
// apply.
func TestDerivedNamesMustFitALabel(t *testing.T) {
	// 52 + len("-credentials") = 64, while 52 + len("-config") = 59 fits.
	name := strings.Repeat("a", 52)
	run := func(sec spec.Secrets) []Issue {
		k := &spec.Kubernetes{Deployment: spec.Deployment{Name: name, Namespace: "ns", Replicas: 1}, Secrets: sec}
		errs, _ := Run(Context{Workflows: wfOK(), Defaults: defsWithStores(), Image: imageOK(), Kube: k, CheckKubernetes: true, Env: func(string) (string, bool) { return "v", true }})
		return errs
	}
	if errs := run(spec.Secrets{Credentials: &spec.CredentialsSecret{Create: &spec.CredCreate{}}}); !hasErr(errs, `derives the name "`+name+spec.CredentialsSecretSuffix+`"`) {
		t.Errorf("an over-long derived Secret name must be refused, naming it, got %v", errs)
	}
	// Referencing a Secret derives nothing, so the same deployment name fits.
	if errs := run(spec.Secrets{Credentials: &spec.CredentialsSecret{Existing: "their-creds"}}); hasErr(errs, "derives the name") {
		t.Errorf("a referenced Secret derives no name, got %v", errs)
	}
}
