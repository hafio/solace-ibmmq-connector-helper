# Critical review - solmq-conn-util

Date: 2026-11-11
Commit: 298cccf on main, working tree dirty
Previous report: none (the 2026-10-06 report and ledger on branch `review` were discarded at the owner's request; IDs restart at CR-0001)
Scope: all areas
Stacks: Go (go.md), bash (bash.md), GitHub Actions (github-actions.md), containers (containers.md, applied to the rendered compose, quadlet and Kubernetes artifacts); PowerShell (scripts/dev.ps1) and the HTML/JavaScript generator page on the core checklist only
Graph: reused (built from 298cccf9, which is HEAD); 2937 nodes, 12881 links, 130 communities
Dynamic checks: offered, none run

## Summary

| Severity | New | Still open | Fixed since last | Total open |
|---|---|---|---|---|
| Blocker | 1 | 0 | 0 | 1 |
| Major | 9 | 0 | 0 | 9 |
| Minor | 39 | 0 | 0 | 39 |
| Nit | 12 | 0 | 0 | 12 |
| Suspected (uncounted) | 5 | 0 | 0 | 5 |

solmq-conn-util is in good structural shape: the import graph is clean, every engine call goes through one argv guard, the rendered artifacts state their posture and tests pin it, and the suite is large and behavioural. The findings cluster in three places. First, the verbatim passthrough blocks: their keys reach application.yml unquoted and unchecked (CR-0005), and aliased or misspelled input is dropped in silence (CR-0010, CR-0046). Second, the seam between `gen` and `cmd`: the engine calls read an unexpanded env.yaml while validate and the artifacts read the expanded one (CR-0006), and the three platforms resolve, check and tear down in different orders (CR-0008, CR-0002). Third, what `validate` cannot see: an empty workflow folder (CR-0022), an incomplete jar closure (CR-0023) and a namespace gate that proves less than its invariant says (CR-0007). This is the first run on a fresh ledger, so there is no trend. Note that the IDs restart here, so the five `CR-NNNN` comments left in the tests now name unrelated findings (CR-0052).

## Top risks

1. CR-0005 [blocker] Passthrough map keys are written into application.yml unquoted and unchecked; a key holding a newline adds lines - internal/render/render.go:392-403 - one workflow file can add or break keys anywhere in the instance's config, and validate passes it.
2. CR-0006 [major] Engine calls use the unexpanded env.yaml while validate and the rendered artifacts use the expanded one - cmd/solmq-conn-util/main.go:1701-1745 - `namespace: ${NS}` validates and renders correctly, then every kubectl call uses the literal.
3. CR-0007 [major] remove can delete a namespace that still holds kinds its listing never asks for, also under --no-prompt - cmd/solmq-conn-util/namespace.go:103-142 - RBAC, NetworkPolicy, Ingress and custom resources are never listed, so a teardown can cascade over them.
4. CR-0008 [major] docker remove resolves every -env credential and refuses when one is not exported - cmd/solmq-conn-util/main.go:732-737 - a teardown from a shell without the deploy-time secrets exits 2 and leaves the instance up.
5. CR-0022 [major] A folder with no workflow files validates, generates and deploys - internal/validate/validate.go:214-224 - a wrong `file_pattern` gives a green validate and a connector that runs nothing.
6. CR-0023 [major] download jar drops a dependency's subtree when its POM cannot be fetched and still reports success - internal/libs/maven.go:224-236 - one transient repository error yields an incomplete classpath.
7. CR-0030 [major] No gate executes the generator page's JavaScript port - internal/gen/htmlgolden_test.go:38-68 - the page can lint and preview differently from the CLI with every gate green.
8. CR-0041 [major] The kubernetes libs.download init container fetches jars with wget, with no checksum and with http accepted - internal/deploy/deploy.go:592-601 - code on the connector's classpath is fetched unverified.
9. CR-0045 [major] cli with a one-shot command and logs exit 0 without acting when several instances match - cmd/solmq-conn-util/instances.go:571-575 - a scripted check on a two-replica deployment succeeds without running.
10. CR-0046 [major] Misspelled keys in env.yaml sections and workflow files are ignored without a message - internal/spec/env.go:52-56 - a one-letter typo takes the default for identity, replicas or restart policy.

## Findings by area

### design

#### CR-0001 [minor] [new] remove on docker and podman needs a fully valid spec, so a broken workflow file blocks teardown

Location: cmd/solmq-conn-util/main.go:727-731, cmd/solmq-conn-util/main.go:882-886, internal/gen/gen.go:233-255
Confidence: verified
Evidence:

    main.go:727  plan, errs, warns := gen.GenerateDocker(req, res)
    main.go:729  if len(errs) > 0 {
    main.go:730      return failFast(errs)
    main.go:882  plan, errs, warns := gen.GeneratePodman(req, res)
    main.go:884  if len(errs) > 0 {
    gen.go:247   Teardown:        opts.Teardown,      (validate.Run gets a teardown mode on kubernetes only)

Both paths run for remove as well as deploy.

Why it matters: Remove needs names, the command and the secret names. A workflow file edited into a parse or validate error after the deploy makes remove exit 2 and leaves the instance up. `start`, `stop` and `restart` already read names only (lifecycle.go:26-30).

Change: Give docker and podman remove a names-only path like runLifecycle's, or pass a teardown mode to validate.Run for them as kubernetes does, so only what remove uses is checked.

#### CR-0002 [minor] [new] Platform orchestration is three hand-sequenced pipelines in cmd, and their step order differs

Location: cmd/solmq-conn-util/main.go:646-1056, cmd/solmq-conn-util/lifecycle.go:70-116
Confidence: verified
Evidence:

    main.go:734   kvs, cerr := gen.ResolveCredentials(plan.Secrets, res)        (docker: resolve first)
    main.go:765   if code, ok := preflight(r, action, validate.PlatformDocker, ...
    main.go:887   if code, ok := preflight(r, action, validate.PlatformPodman, ...   (podman: preflight first)
    main.go:922   kvs, cerr := gen.ResolveCredentials(plan.Secrets, res)        (then resolve, inside podmanDeploy)
    main.go:931-998   the podman secret diff, the stale-secret list and the write closure live in cmd
    lifecycle.go:71   switch name {                                             (a second per-platform switch)

Why it matters: An unreachable engine plus an unset variable reports a different error on each platform, and a new platform or per-platform step is a hand edit across actTargets, the lifecycle switch, the status collector and the instance helpers. CR-0008 and CR-0047 are drift this layout allowed.

Change: Introduce one per-platform value (resolve, host checks, preflight, apply, remove, lifecycle) below cmd, built once per run, and fix one documented order: validate and resolve, preflight, act. Move the podman secret diff and the docker image and digest checks into internal/runner with it.

#### CR-0003 [nit] [new] actTargets doc comment sits above the actionOpts type

Location: cmd/solmq-conn-util/main.go:514-545
Confidence: verified
Evidence:

    main.go:514  // actTargets maps each of platformNames to its deploy/remove implementation.
    main.go:518  // actionOpts is one deploy/remove invocation, as an options struct rather than
    main.go:525  type actionOpts struct {
    main.go:545  var actTargets = map[string]func(o actionOpts, r gdurunner.Runner) int{

Why it matters: go doc attaches the actTargets paragraph to actionOpts, and the map has no comment.

Suggested patch (not applied):

```diff
--- a/cmd/solmq-conn-util/main.go
+++ b/cmd/solmq-conn-util/main.go
@@ -511,10 +511,6 @@
 
 // ---- deploy / remove ---------------------------------------------------------
 
-// actTargets maps each of platformNames to its deploy/remove implementation.
-// Gated against platformNames by TestPlatformMapsCoverThreeNames (deploy/
-// remove no longer model a positional Targets list in commands.go -- the
-// platform is resolved by resolvePlatform, not looked up from args[0]).
 // actionOpts is one deploy/remove invocation, as an options struct rather than
 // a growing positional list: removeNamespace applies to exactly one platform,
 // and threading it as a fourth argument through all three would put a
@@ -542,6 +538,10 @@
 	prompter *confirm.Prompter
 }
 
+// actTargets maps each of platformNames to its deploy/remove implementation.
+// Gated against platformNames by TestPlatformMapsCoverThreeNames (deploy/
+// remove no longer model a positional Targets list in commands.go -- the
+// platform is resolved by resolvePlatform, not looked up from args[0]).
 var actTargets = map[string]func(o actionOpts, r gdurunner.Runner) int{
 	tgtKubernetes: actKubernetes,
 	tgtDocker:     actDocker,
```

#### CR-0004 [nit] [new] status takes its clock from time.Now() instead of the injected app clock

Location: cmd/solmq-conn-util/status.go:202-208, cmd/solmq-conn-util/main.go:141-184
Confidence: verified
Evidence:

    status.go:207  now: time.Now(),
    main.go:176    // progressFrameRate and now are the two seams under the progress line: how
    main.go:182    now               func() time.Time

Why it matters: A test that pins app.now does not pin what statusOpts.now feeds.

Change: Set statusOpts.now from the app's clock where the options are built.

### correctness

#### CR-0005 [blocker] [new] Passthrough map keys are written into application.yml unquoted and unchecked; a key holding a newline adds lines

Location: internal/render/render.go:392-403, internal/render/render.go:369-371, internal/consolidate/consolidate.go:550-557, internal/validate/transform.go:88-98, internal/validate/extra.go:23-25
Confidence: verified
Evidence:

    render.go:396      k := n.Content[i].Value
    render.go:399      writeScalar(w, indent, k, consolidate.FormatScalar(v))
    render.go:401      w.Line(indent, k+":")
    render.go:38       w.Line(indent, key+": "+val)                     (writeScalar, value on one line)
    yamlwriter.go:19   w.b.WriteString(strings.Repeat(" ", indent))     (then the text as given, then '\n')
    transform.go:90    h, v := exprs.Content[i].Value, exprs.Content[i+1]   (duplicate and value-kind checks follow; none on h)
    extra.go:25        var extraKeyRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.\-\[\]]*$`)

A workflow file containing

    transform-headers:
      expressions:
        "x\n    enabled": "1"

is valid YAML. yaml.v3 decodes the key as `x`, a newline, four spaces and `enabled`. checkTransformHeaders never looks at the header name. renderContainer hands the key to writeScalar, the value has no newline, and one Line call writes `x` at indent 12 followed by `    enabled: "1"` at indent 4: a sibling of `workflows:` under `solace.connector`. A key such as `a: b` or `#x` corrupts the document with no newline at all. The same raw-key path serves api-properties, additional-properties, consumer, producer, leader-election.fail-over, logging.level (render.go:370), solace-defaults and transform. Looked for a refuting guard: extraKeyRE is applied at extra.go:109 to the top-level keys of a `solace:`/`mq:` block and of mq-defaults only, affinityKeyRE (validate.go:1422) to the affinity block only, and apiprops.go reads keys without checking them.

Why it matters: The author of one workflow file can add or break keys anywhere in the instance's application.yml, and validate, the gate that must report everything, passes it. This is the class of defect the project already closed for the Kubernetes manifest.

Change: Hold every passthrough key to a plain-key grammar in validate, walking api-properties, additional-properties, consumer, producer, fail-over, logging.level, solace-defaults, transform and transform-headers at every depth as the affinity check does. Quote a key that is not plain in render.renderContainer, renderLogging and consolidate.nodeToProps. Add a render test with a key holding a newline and one holding `: `, and a validate test for each block.

#### CR-0006 [major] [new] Engine calls use the unexpanded env.yaml while validate and the rendered artifacts use the expanded one

Location: cmd/solmq-conn-util/main.go:1701-1745, cmd/solmq-conn-util/main.go:670-711, internal/gen/gen.go:1091-1101, internal/spec/expand.go:30-44, docs/configuration.md:79-82
Confidence: verified
Evidence:

    main.go:1706       e, err := spec.ParseEnv(data)                     (loadEnv: no Expand; this e drives deploy and remove)
    main.go:1728       wfDir := e.Workflows.Dir                          (scanned unexpanded)
    main.go:1181       return spec.ParseEnv(data)                        (loadEnvFile: lifecycle, status, logs, cli)
    gen.go:1091        pe, err := spec.ParseEnv(data)
    gen.go:1096        spec.Expand(spec.Expander{                        (a second parse; only this copy is expanded)
    main.go:671        ns := k.Deployment.Namespace
    main.go:696        if code, ok := ensureNamespace(r, base, k.Command, ns, approvals); !ok {
    lifecycle.go:130   name, ns := k.Deployment.Name, k.Deployment.Namespace     (waitForRollout)
    configuration.md:81  Any non-credential string field may reference the environment `solmq-conn-util`

With `kubernetes.deployment.namespace: ${NS}` and NS=prod, validate sees `prod` and passes and the manifest says `prod`. The permission probes, the namespace check and create, the rollout wait and the namespace gate then run with the literal `${NS}`. With `deployment.name: ${APP}` the apply succeeds and the rollout wait asks for `deployment/${APP}`. docker.name and podman.name feed the redeploy check and the podman secret names the same way (main.go:807, main.go:942). `spec.Expand` has one caller outside tests, gen.go:1096.

Why it matters: A spec that validates clean and renders correct artifacts fails mid-deploy or addresses a differently named object. A per-environment `${NS}` is the natural use of the documented feature.

Change: Expand the Env that loadEnv and loadEnvFile return with the Lookup gen.parse uses, before workflows.dir and file_pattern are read, or have gen return its expanded Env and use that one everywhere. Add a dispatch test that deploys with `${NS}` in the namespace and asserts the argv.

#### CR-0007 [major] [new] remove can delete a namespace that still holds kinds its listing never asks for, also under --no-prompt

Location: cmd/solmq-conn-util/namespace.go:103-142, cmd/solmq-conn-util/namespace.go:37-42, internal/runner/runner.go:866-876, docs/DESIGN.md:186-189
Confidence: verified
Evidence:

    namespace.go:37    // The invariant this half exists to hold: a namespace containing anything this
    namespace.go:38    // release does not own is never removed. Not interactively, not under --no-prompt,
    namespace.go:103   doc, lerr := runner.KubernetesListJSON(bg(), r, base, ns)
    namespace.go:109   occupants, perr := nsgate.Occupants(doc, k.CreatedNames())
    namespace.go:129   if !o.noPrompt {
    namespace.go:142   out, derr := runner.KubernetesDeleteNamespace(bg(), r, base, ns)
    runner.go:866      // KubernetesListJSON lists what is in a namespace in one call (nsgate.List: `get
    runner.go:867      // all,persistentvolumeclaims,secrets,configmaps -n <ns> -o json`), for the namespace
    DESIGN.md:187      it only when it is absent.** Deleting a Namespace cascades. A namespace holding
    DESIGN.md:188      anything the release does not own is never removed, on any path, with or without

kubectl's `all` category covers pods, services and the workload, job and autoscaler kinds. Roles, RoleBindings, NetworkPolicies, Ingresses, non-default ServiceAccounts, PodDisruptionBudgets and every custom resource are outside it and are never listed. ensureNamespace returns as soon as the namespace exists (namespace.go:58-60) and records nothing, so remove cannot tell a namespace the tool created from one it found. Deploy into a prepared namespace that holds only such objects, then run `remove --no-prompt` with an identity that may delete namespaces: the listing shows no occupant, the question is skipped, and the delete cascades over all of them. Interactively the question asserts "holds nothing else" (namespace.go:162) on the same listing. Looked for a refuting guard: nsgate.Protected covers `default` and the `kube-*` namespaces only. operations.md:262-263 does state the exact listing, so the user guide is accurate; DESIGN.md and the code comment are not.

Why it matters: A teardown can delete objects the tool never created, without a question under `--no-prompt`. The stated invariant is stronger than the check behind it.

Change: The listing is `nsgate.List` in go-deploy-utility, so widening it (for example over `kubectl api-resources --verbs=list --namespaced -o name`) belongs there. In this repo: offer the namespace delete only for a namespace the tool created (label it in ensureNamespace and read that back), do not let `--no-prompt` answer for one it did not create, and until then reword namespace.go:37-42 and DESIGN.md:186-189 to name the kinds that are checked.

#### CR-0008 [major] [new] docker remove resolves every -env credential and refuses when one is not exported

Location: cmd/solmq-conn-util/main.go:732-737, cmd/solmq-conn-util/main.go:752-773, internal/gen/gen.go:794-812, docs/operations.md:275-276
Confidence: verified
Evidence:

    main.go:734    kvs, cerr := gen.ResolveCredentials(plan.Secrets, res)       (no check that action is deploy)
    main.go:735    if cerr != nil {
    main.go:736        return usageExit(cerr)
    main.go:756    storeKVs := gen.DockerStorePlaceholders(plan)                (stores: placeholders on remove)
    main.go:773    env := envPairs(append(kvs, storeKVs...))
    main.go:791    out, rerr := runner.Docker(bg(), r, e.Docker.Command, action, plan.Compose, envDir, env, approvals)
    gen.go:794     // An -env variable that is unset or set but empty is an error: deploying with a
    operations.md:275  - docker: `compose down` against the configured project. Store secrets are

Why it matters: `remove --platform docker` from a shell without the deploy-time variables exits 2 and leaves the instance up. kubernetes and podman remove resolve nothing. The resolved values are also handed to a `compose down` that never reads them.

Change: Resolve credentials for deploy only. For remove build the credential entries as placeholders the way DockerStorePlaceholders does for stores, say so in operations.md section 5, and add a dispatch test that removes with the variables unset.

#### CR-0009 [minor] [new] cipher and key-alias reach application.yml unquoted

Location: internal/consolidate/consolidate.go:277-279, internal/tls/tls.go:68-70, internal/render/render.go:375-378
Confidence: verified
Evidence:

    consolidate.go:278  props = append(props, Prop{Key: "WMQ_SSL_CIPHER_SUITE", Val: a.cipher})
    tls.go:69           out = append(out, KV{"SSL_PRIVATE_KEY_ALIAS", sel.KeyAlias})
    consolidate.go:250  props = append(props, Prop{Key: kv.Key, Val: kv.Val})
    render.go:377       writeScalar(w, indent, p.Key, p.Val)
    render.go:25        // one would not read back unchanged. Every value that originates in the spec
    validate.go:562     if (s.Cipher != "" || s.KeyAlias != "") && !s.TLS {       (presence and dependency are the only checks)

Why it matters: A key alias or cipher holding `: `, ` #` or a leading YAML indicator yields an invalid or truncated application.yml. An alias is free text in a keystore, so this is reachable, if unusual.

Suggested patch (not applied), for the cipher:

```diff
--- a/internal/consolidate/consolidate.go
+++ b/internal/consolidate/consolidate.go
@@ -275,7 +275,7 @@
 			mb.Extras = overlayExtras(mqDefaults, dropManaged(a.extra, spec.SystemMQ, &warns, binderOwner(a.name)))
 			var props []Prop
 			if a.cipher != "" {
-				props = append(props, Prop{Key: "WMQ_SSL_CIPHER_SUITE", Val: a.cipher})
+				props = append(props, Prop{Key: "WMQ_SSL_CIPHER_SUITE", Val: QuoteScalar(a.cipher)})
 			}
 			props = appendPassthrough(props, a.pass2, &warns, binderOwner(a.name))
 			mb.AddlProps = props
```

Change: Quote the alias where consolidate copies the tls props (consolidate.go:249-251), since package tls cannot import consolidate. QuoteScalar leaves ordinary values bare, so the goldens should not move.

#### CR-0010 [minor] [new] An aliased or non-mapping passthrough block is dropped without a message

Location: internal/consolidate/consolidate.go:545-548, internal/render/render.go:363-366, internal/validate/extra.go:100-103
Confidence: verified
Evidence:

    consolidate.go:546  if node == nil || node.Kind != yaml.MappingNode {
    consolidate.go:547      return nil
    render.go:364       if level == nil || level.Kind != yaml.MappingNode || len(level.Content) == 0 {
    extra.go:101        if at.extra == nil || at.extra.Kind != yaml.MappingNode {
    spec.go:133         APIProps      yaml.Node `yaml:"api-properties"`

yaml.v3 hands a yaml.Node field the node as written, so `api-properties: *shared` arrives with Kind AliasNode. nodeToProps returns nil and the block is not rendered. spec/extra.go:224 has a deref helper, used for the `solace:`/`mq:` block itself and for tls, but not for these nodes.

Why it matters: A second connection that reuses a block through a YAML anchor silently loses its properties and runs on defaults.

Change: deref the passthrough nodes where they are captured, and have validate report a non-mapping value for api-properties, additional-properties, consumer, producer, fail-over, solace-defaults, mq-defaults and logging.level.

#### CR-0011 [minor] [new] solace-defaults is passed through unchecked while mq-defaults rejects credentials and tool-managed keys

Location: internal/validate/extra.go:68-73, internal/consolidate/consolidate.go:232-244
Confidence: verified
Evidence:

    extra.go:71         if d.MQDefaults != nil {
    extra.go:72             checkExtraKeys(add, warn, extraSite{file: fileEnv, label: "mq-defaults", ...
    consolidate.go:233  // once, not per binder (solace-defaults is left as written, as it always was).
    consolidate.go:234  mqDefaults := dropManaged(nodeToProps(d.MQDefaults), spec.SystemMQ, &warns, "mq-defaults")
    consolidate.go:244  sb.Extras = overlayExtras(nodeToProps(d.SolaceDefaults), dropManaged(a.extra, ...

Why it matters: `solace-defaults: {client-password: x}` puts a literal credential into application.yml (a ConfigMap on kubernetes) beside the tool's own key, and `host` or `msg-vpn` there duplicate keys the tool writes. The same input is an error under mq-defaults. The comment shows the asymmetry is known; its cost is not recorded.

Change: Add a solace-defaults site to checkAllExtraKeys with `defaults: true`, and filter it with dropManaged as mq-defaults is.

#### CR-0012 [minor] [new] podman remove deletes the instance's secrets even when stopping or removing the unit failed

Location: cmd/solmq-conn-util/main.go:1038-1055, internal/runner/runner.go:790-810
Confidence: verified
Evidence:

    main.go:1038   out, rerr := runner.PodmanRemove(bg(), r, sc, []string{plan.Service}, []string{plan.Unit.Filename})
    main.go:1039   // The secrets go last, after the unit that mounted them is gone: the
    main.go:1050   so, serr := runner.PodmanSecretRemove(bg(), r, p.Command, names, approvals)      (runs whatever rerr is)
    runner.go:795  if err != nil {
    runner.go:796      return out.String(), fmt.Errorf("systemctl stop %s: %w", s, err)             (before the unit file is removed)

Why it matters: When the stop fails the unit file stays installed and its secrets are deleted, so the next start fails. The error names only the stop.

Change: Skip the secret removal while the unit file is still on disk, or say in the error that the secrets were removed. CR-0066 is the repeat-remove case.

#### CR-0013 [minor] [new] Names reach the argv of status, logs and cli with a leading dash allowed; lifecycle verbs hand-roll a second name grammar

Location: cmd/solmq-conn-util/instances.go:114-127, cmd/solmq-conn-util/instances.go:219-223, cmd/solmq-conn-util/instances.go:281-291, cmd/solmq-conn-util/lifecycle.go:19-24, internal/validate/validate.go:1863-1873
Confidence: verified
Evidence:

    instances.go:119   if !validate.SafeToken(p) {                       (--pod; --container at 124)
    instances.go:222   name := s.env.Kubernetes.Deployment.Name          (from ParseEnv, not validated)
    instances.go:286   return s.env.Docker.Name, nil
    validate.go:1869   // contexts) use only letters, digits, and - . / : _ = @ , +, none of which are
    lifecycle.go:23    // start with '-' and be read as a flag.
    lifecycle.go:24    var lifecycleNameRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

Why it matters: status, logs and cli pass a `--pod`/`--container` value or an env.yaml name that starts with `-` into the engine argv as a positional, where the engine reads it as an option. start, stop and restart refuse the same name with a private copy of the DNS-1123 grammar. Two rules for one concept, and the looser one guards three verbs.

Suggested patch (not applied); lifecycle keeps its behaviour and drops its private grammar:

```diff
--- a/cmd/solmq-conn-util/lifecycle.go
+++ b/cmd/solmq-conn-util/lifecycle.go
@@ -3,7 +3,6 @@
 import (
 	"fmt"
 	"os"
-	"regexp"
 	"strings"
 
 	"github.com/hafio/go-deploy-utility/execguard"
@@ -16,13 +15,6 @@
 	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/validate"
 )
 
-// lifecycleNameRE is the DNS-1123 label every name a lifecycle verb puts in an
-// argv must be: validate holds deployment.name, deployment.namespace and the
-// docker/podman names to it, and these verbs read env.yaml without validating
-// the rest of it, so the names are held to it here -- above all, none may
-// start with '-' and be read as a flag.
-var lifecycleNameRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
-
 // runLifecycle implements start, stop and restart: the instance env.yaml names
 // on the resolved platform, acted on in place. It reads env.yaml for the
 // section's names and command only -- nothing is rendered, no credential is
@@ -61,8 +53,13 @@
 	if perr != nil {
 		return usageExit(perr)
 	}
+	// Every name a lifecycle verb puts in an argv is a DNS-1123 label, the rule
+	// validate holds deployment.name, deployment.namespace and the docker/podman
+	// names to (kube.ValidLabel). These verbs read env.yaml without validating
+	// the rest of it, so the names are held to it here -- above all, none may
+	// start with '-' and be read as a flag.
 	checkName := func(field, v string) error {
-		if !lifecycleNameRE.MatchString(v) || len(v) > 63 {
+		if !kube.ValidLabel(v) {
 			return fmt.Errorf("%s %q is not a valid DNS-1123 label; run validate", field, v)
 		}
 		return nil
```

Change: Apply the same check in resolveInstanceSession to `--pod`/`--container` values and to the names read from env.yaml (kube.ValidLabel on kubernetes, the container-name rule on docker and podman), with a test that the runner is not called.

#### CR-0014 [minor] [new] The kubernetes manifest writes deployment and namespace names bare, so a name such as 123 changes type

Location: internal/deploy/deploy.go:537-550
Confidence: verified
Evidence:

    deploy.go:537  w.Line(2, "name: "+name)
    deploy.go:538  w.Line(2, "namespace: "+ns)
    deploy.go:546  w.Line(6, "app: "+name)
    deploy.go:585  w.Line(8, consolidate.QuoteScalar(k)+": "+consolidate.QuoteScalar(sel[k]))     (nodeSelector is quoted)

Why it matters: kube.ValidLabel accepts `123`. The manifest then carries `name: 123`, a YAML integer, and kubectl refuses the apply after validate passed. Compose quotes every name for this reason (DESIGN.md:45-49).

Change: Write names, namespaces and label values through consolidate.QuoteScalar in internal/deploy, and add a test with an all-digit name. QuoteScalar leaves ordinary names bare.

#### CR-0015 [minor] [new] validate accepts replicas and management.port values that only the cluster or the connector rejects

Location: internal/spec/kubernetes.go:441-443, internal/validate/validate.go:942-948
Confidence: verified
Evidence:

    kubernetes.go:441  if k.Deployment.Replicas == 0 {
    kubernetes.go:442      k.Deployment.Replicas = 1                      (an explicit `replicas: 0` becomes 1)
    validate.go:946    if (mode == "" || mode == spec.LeaderStandalone) && dep.Replicas > 1 {     (the only replicas check)
    validate.go:942    checkPortRange(add, "kubernetes.service.port", ctx.Kube.Service.Port)      (management.port has none)
    render.go:341      w.Line(4, "port: "+strconv.Itoa(mg.Port))

Why it matters: `replicas: -1` passes validate and fails at the API server. `management.port: 70000` is rendered as written on docker, podman and `generate config`.

Change: Check replicas >= 1 (use a pointer if an explicit 0 should mean something) and management.port 1-65535 under its own key.

#### CR-0016 [minor] [new] Binder names can collide with the reserved undefined binder and with their own numbered suffixes

Location: internal/consolidate/names.go:105-114, internal/consolidate/consolidate.go:286-287
Confidence: verified
Evidence:

    names.go:106        used := map[string]int{}
    names.go:112        used[base]++
    names.go:113        return fmt.Sprintf("%s-%d", base, used[base])
    consolidate.go:287  m.Binders = append(m.Binders, &Binder{Name: "undefined", Kind: "undefined"})

Why it matters: A connection named `undefined`, or one literally named like another's generated suffix (`edge` twice plus `edge-2`), yields two binders under one key, and the connector loads one of them.

Suggested patch (not applied):

```diff
--- a/internal/consolidate/names.go
+++ b/internal/consolidate/names.go
@@ -103,14 +103,23 @@
 // sol-conn-N / mq-conn-N, with a -2/-3 suffix disambiguating any clash between
 // two different binders.
 func assignBinderNames(accs []*acc) {
-	used := map[string]int{}
+	// "undefined" is the internal binder Build appends last; no connection may take it.
+	used := map[string]int{"undefined": 1}
 	assign := func(base string) string {
 		if used[base] == 0 {
 			used[base] = 1
 			return base
 		}
-		used[base]++
-		return fmt.Sprintf("%s-%d", base, used[base])
+		// A suffixed name can be taken too (a connection literally named
+		// "edge-2"), so count on until the candidate is free.
+		for {
+			used[base]++
+			name := fmt.Sprintf("%s-%d", base, used[base])
+			if used[name] == 0 {
+				used[name] = 1
+				return name
+			}
+		}
 	}
 	var solN, mqN int
 	for _, a := range accs {
```

#### CR-0017 [minor] [new] FormatScalar compares Style for equality, so a tagged quoted scalar is written bare

Location: internal/consolidate/consolidate.go:574-588
Confidence: verified
Evidence:

    consolidate.go:575  switch n.Style {
    consolidate.go:576  case yaml.DoubleQuotedStyle:
    consolidate.go:585  default:
    consolidate.go:586      return n.Value

yaml.v3 ORs TaggedStyle into the quoting style, so `!!str "a: b"` has Style 3, matches no case and is emitted as `a: b`.

Why it matters: A quoted passthrough value that also carries a tag breaks the document or changes type on regeneration.

Suggested patch (not applied); it restores the quoting, and a plain tagged scalar still loses its tag:

```diff
--- a/internal/consolidate/consolidate.go
+++ b/internal/consolidate/consolidate.go
@@ -572,7 +572,9 @@
 // folded SpEL expression such as "x = #joinString(...)" would otherwise lose
 // everything from its " #" on.
 func FormatScalar(n *yaml.Node) string {
-	switch n.Style {
+	// yaml.v3 ORs TaggedStyle into the quoting style, so a tagged scalar
+	// (!!str "a: b") would match no case below and be written bare.
+	switch n.Style &^ yaml.TaggedStyle {
 	case yaml.DoubleQuotedStyle:
 		return strconv.Quote(n.Value)
 	case yaml.SingleQuotedStyle:
```

#### CR-0018 [minor] [new] Only the first YAML document of env.yaml is read; later documents are ignored

Location: internal/spec/env.go:52-56
Confidence: verified
Evidence:

    env.go:53  var raw rawEnv
    env.go:54  if err := yaml.Unmarshal(data, &raw); err != nil {

yaml.Unmarshal decodes one document. The review pass over internal/spec reports the same call in ParseWorkflow.

Why it matters: Two workflows joined by `---`, or a trailing document that sets `enabled: false`, load as the first document alone with no error.

Change: Decode with yaml.NewDecoder and refuse a second document, in ParseEnv and ParseWorkflow, naming the file.

#### CR-0019 [minor] [new] generate -o writes in place and keeps an existing file's wider mode; a failed stdout write exits 0

Location: cmd/solmq-conn-util/main.go:1771-1780
Confidence: verified
Evidence:

    main.go:1773  fmt.Print(content)                                             (error ignored)
    main.go:1776  if err := os.WriteFile(out, []byte(content), 0o600); err != nil {
    main.go:790   // `generate --platform docker -o` writes one (0600) for a hand-run compose.

os.WriteFile applies the mode only when it creates the file, and it truncates before writing.

Why it matters: Regenerating over an existing 0644 file leaves the compose document, which inlines application.yml with the status account's password, world-readable. A failed write leaves a short file in place of the last good one, and `generate > file` on a full disk exits 0.

Change: Write to a 0600 temp file in the target directory and rename it over the target; check the stdout write error.

#### CR-0020 [nit] [new] A double-dash separator protects only the first token after it

Location: cmd/solmq-conn-util/main.go:1682-1695
Confidence: verified
Evidence:

    main.go:1685  if err := fs.Parse(args); err != nil {
    main.go:1691  pos = append(pos, fs.Arg(0))
    main.go:1692  args = fs.Args()[1:]                (the rest is parsed as flags again)

Why it matters: `examples -- -a -f` takes `-a` as a directory and `-f` as the force flag. `cli` splits at the separator itself (shell.go) and is not affected.

Change: When the parse consumed a `--`, append every remaining argument to pos and stop.

#### CR-0021 [nit] [new] Rootless podman deploy enables linger before the redeploy question, then says nothing changed

Location: cmd/solmq-conn-util/main.go:899-906, cmd/solmq-conn-util/main.go:1627-1632
Confidence: verified
Evidence:

    main.go:900   if err := ensureLinger(r, sc, h); err != nil {
    main.go:903   if err := checkSubIDs(r, e.Podman, plan.RunAs, h, approvals); err != nil {
    runner.go:687 cmd := sc.WithEnv(systemd.EnableLinger())
    main.go:1630  fmt.Fprintln(os.Stderr, "deploy: cancelled; nothing changed")

Why it matters: After a declined redeploy, or a failed sub-id check, linger is already on for the account. The note at rootless.go:43 does say so, so nothing is hidden; the closing line is simply not true.

Change: Run checkSubIDs first and enable linger only once the question is answered yes.

### resilience

#### CR-0022 [major] [new] A folder with no workflow files validates, generates and deploys

Location: internal/validate/validate.go:214-224, internal/scan/scan.go:25-50, cmd/solmq-conn-util/main.go:1728-1743
Confidence: verified
Evidence:

    validate.go:219  if n <= MaxWorkflows {
    validate.go:220      return
    scan.go:50       return res, nil                                    (an empty match is not an error)
    main.go:1732     sr, err := scan.Scan(wfDir, e.Workflows.FilePattern, absEnv)
    render.go:249    w.Line(4, "workflows:")
    render.go:250    for _, wf := range m.Workflows {

Looked for a refuting check: a grep for a zero-workflow test in cmd and internal finds only the status script's runtime note (statusscript.go:586).

Why it matters: `file_pattern: "*.yml"` over `.yaml` files, or a mistyped `dir`, gives a clean validate and a deployed connector that runs nothing.

Change: Make zero workflows a validate error that names workflows.dir, file_pattern and the folder scanned, so generate and deploy stop at the same gate.

#### CR-0023 [major] [new] download jar drops a dependency's subtree when its POM cannot be fetched and still reports success

Location: internal/libs/maven.go:224-236, internal/libs/maven.go:175-181, internal/libs/libs.go:394-397
Confidence: verified
Evidence:

    maven.go:225  pom, pomErr := fetchPOM(d, cur.Coord, cur.Version)
    maven.go:226  if pomErr != nil {
    maven.go:230      continue
    maven.go:233  props, depMgmt, chainErr := resolveParentChain(d, pom)
    maven.go:234  if chainErr != nil {
    maven.go:235      continue
    maven.go:175  // known, an unreachable dependency (or parent) POM only stops that one
    maven.go:176  // branch from expanding further; the artifact stays in the closure at the
    libs.go:394   artifacts, err := resolveClosureAt(doer, seed, version)

`continue` skips the loop that queues that artifact's own dependencies, and nothing reaches the report. One failed POM fetch gives a shorter jar list that is downloaded, checksum-verified and reported as complete. The comment's reasoning covers the artifact's own jar, not its children.

Why it matters: A transient repository error yields an incomplete classpath reported as success. It surfaces later as a NoClassDefFoundError in the connector.

Change: Return the unresolved branches from resolveClosureAt and record them in the report, failing the run for the mq and syslog seeds. Retry a POM fetch a small, capped number of times. Add a test whose Doer fails one dependency POM.

#### CR-0024 [minor] [new] podman deploy removes stale secrets before the new unit is written, and writes the unit in place

Location: cmd/solmq-conn-util/main.go:976-998, internal/runner/runner.go:116-126
Confidence: verified
Evidence:

    main.go:987    o, err := runner.PodmanSecretRemove(bg(), r, p.Command, stale, approvals)
    main.go:993    if err := runner.WriteFile(unitPath, plan.Unit.Content, 0o644); err != nil {
    runner.go:122  if err := os.WriteFile(path, []byte(content), mode); err != nil {

Why it matters: If the unit write fails, the installed unit still names the secrets just deleted and the next start fails. The write truncates in place, so an interrupted write leaves a short unit for the next daemon-reload.

Change: Remove stale secrets after the unit is written and the reload succeeded, as a warning on failure. Write the unit through a temp file in the same directory and rename it.

#### CR-0025 [minor] [new] No engine call carries a deadline

Location: cmd/solmq-conn-util/main.go:207-210
Confidence: verified
Evidence:

    main.go:207  // bg is the context every runner helper call outside logs --follow runs under.
    main.go:208  // Ctrl-C keeps its default effect there: the process ends, and Exec's children
    main.go:210  func bg() context.Context { return context.Background() }

Why it matters: This is a recorded choice and it is sound at a terminal. An unattended run against a hung engine socket (`docker info`, `compose up`, `systemctl start`) waits forever; only the rollout wait is bounded.

Change: Give preflight and the read probes a bounded context, and thread a signal-derived context into deploy, remove and the lifecycle verbs.

### quality

#### CR-0026 [minor] [new] Output goes to process-wide os.Stdout and os.Stderr; tests swap them without defer and none runs in parallel

Location: cmd/solmq-conn-util/main_test.go:202-240, cmd/solmq-conn-util/main.go:1782-1792, cmd/solmq-conn-util/main.go:141-184
Confidence: verified
Evidence:

    main_test.go:204  old := os.Stdout
    main_test.go:209  os.Stdout = w
    main_test.go:215  f()
    main_test.go:217  os.Stdout = old                                   (no defer; captureStderr is the same)
    main.go:1788      fmt.Fprintf(os.Stderr, "error: %s %s failed: %v\n", action, platform, err)

The app struct (main.go:141-184) carries the runner, host, prompter, download seam and clock, and no writer. main.go alone has 64 direct references to the two streams. `grep -c 't.Parallel()'` over the test tree is 0.

Why it matters: The output streams are the last process-wide seam, so the CLI suite swaps globals and cannot run in parallel. A t.Fatal inside f leaves os.Stderr pointing at the pipe for every later test.

Change: Add stdout and stderr writers to app. Until then restore the originals with defer in captureStdout and captureStderr.

#### CR-0027 [nit] [new] Small helpers are defined more than once

Location: internal/dockergen/dockergen.go:297, internal/podmangen/podmangen.go:113, internal/validate/transform.go:278, internal/spec/defaults.go:559, internal/deploy/deploy_test.go:362, internal/render/render_test.go:275, internal/gen/golden_test.go:578
Confidence: verified
Evidence:

    dockergen.go:297   func seconds(n int) string { return strconv.Itoa(n) + "s" }
    podmangen.go:113   func seconds(n int) string { return strconv.Itoa(n) + "s" }
    transform.go:278   func isNull(n *yaml.Node) bool {           (also spec/defaults.go:559)
    deploy_test.go:362 func lineDiff(want, got string) string {   (also render_test.go:275, golden_test.go:578)

Why it matters: Each copy changes on its own; the three lineDiff copies shape the failure output of three golden tests.

Change: Keep one isNull in spec (validate imports it), one seconds where both generators can reach it, and one lineDiff in a shared test helper.

#### CR-0028 [nit] [new] checkSecretName's doc comment sits above checkJavaOptions

Location: internal/validate/validate.go:1448-1453, internal/validate/validate.go:1612
Confidence: verified
Evidence:

    validate.go:1448  // checkSecretName gates a Secret name that is emitted verbatim into the manifest
    validate.go:1451  // checkJavaOptions gates the top-level java-options: block. It cannot take
    validate.go:1612  func checkSecretName(add func(string, string, ...any), field, name string) {

Why it matters: go doc shows the Secret-name text on checkJavaOptions, and checkSecretName has none.

Suggested patch (not applied):

```diff
--- a/internal/validate/validate.go
+++ b/internal/validate/validate.go
@@ -1445,9 +1445,6 @@
 	return len(name) <= 63 && labelNameRE.MatchString(name)
 }
 
-// checkSecretName gates a Secret name that is emitted verbatim into the manifest
-// (metadata.name, secretRef.name, volumes[].secret.secretName). An empty name is
-// its own message: a create block without one renders a nameless object.
 // checkJavaOptions gates the top-level java-options: block. It cannot take
 // SafeToken like the other shared values: each variable is several options by
 // design, so spaces separate them, and shell metacharacters such as the '*' in
@@ -1609,6 +1606,9 @@
 	}
 }
 
+// checkSecretName gates a Secret name that is emitted verbatim into the manifest
+// (metadata.name, secretRef.name, volumes[].secret.secretName). An empty name is
+// its own message: a create block without one renders a nameless object.
 func checkSecretName(add func(string, string, ...any), field, name string) {
 	if name == "" {
 		add(fileEnv, "%s is required", field)
```

#### CR-0029 [nit] [new] Two package-level variables in libs are test seams written at run time

Location: internal/libs/libs.go:137, internal/libs/image.go:104
Confidence: verified
Evidence:

    libs.go:137   var maxArtifactBytes int64 = 32 * 1024 * 1024
    image.go:104  var embeddedListFS fs.FS = embeddedListFiles

The quality pass reports both are overwritten by tests (libs_test.go:442-444, libs_test.go:1080).

Why it matters: The house rule is no global state, and commit 9f03a4a removed the same pattern from cmd. Tests that write these cannot run in parallel.

Change: Carry both in the libs Input, the way cmd threads app.

### tests

#### CR-0030 [major] [new] No gate executes the generator page's JavaScript port

Location: internal/gen/htmlgolden_test.go:38-68, internal/gen/htmlgolden_test.go:194-212, docs/DEVELOPMENT.md:91-100
Confidence: verified
Evidence:

    htmlgolden_test.go:38    func TestGeneratorPageGoldenInSync(t *testing.T) {
    htmlgolden_test.go:115   func TestGeneratorPageFindingsGoldenInSync(t *testing.T) {
    htmlgolden_test.go:203   {"SOLACE_KEYS", spec.KnownKeys(spec.SystemSolace)},      (six key lists are pinned; no section keys)
    DEVELOPMENT.md:95        **Self-test** button diffs the sample set against both.
    DEVELOPMENT.md:99        - Changing any ported function means re-running Self-test (**Load sample**,

A grep for node, goja or chromedp in scripts/, .github/workflows and go.mod finds nothing. The Go tests compare the page's embedded copies with the Go goldens; they never run the port.

Why it matters: The page carries a port of validate, consolidate, render and the name derivation. A schema change is a hand port in six places (DEVELOPMENT.md:101-103), verified only when someone clicks Self-test, so the page can lint and preview differently from the CLI with every gate green.

Change: Add a gate that runs the page's script against its embedded golden and findings: a dev-script task that uses node when present and says so when absent, or an embedded JavaScript engine in a Go test. Until then state in DEVELOPMENT.md that no gate executes the port.

#### CR-0031 [minor] [new] The docs/test.md gate compares three counts and never a test name

Location: cmd/solmq-conn-util/testcatalog_test.go:56-91
Confidence: verified
Evidence:

    testcatalog_test.go:75  if gotFuncs != wantFuncs {
    testcatalog_test.go:81  if gotRows != wantRows {
    testcatalog_test.go:86  if gotPackages != wantPackages {

Why it matters: A renamed test, or one removed and another added, keeps the totals and the gate green while the catalogue lists a test that no longer exists. The conventions pass compared the names by hand and found the two sets identical today (1074 functions), so this is a gap in the gate, not drift.

Change: Compare the set of `func Test` names with the first-column names in the doc, and report the names missing on either side.

#### CR-0032 [minor] [new] main_test.go holds the tests for six source files, so the test tree does not mirror the source tree

Location: cmd/solmq-conn-util/main_test.go:1-8096
Confidence: verified
Evidence:

    git ls-files cmd/solmq-conn-util: main_test.go is 8096 lines; there is no status_test.go,
    logs_test.go, shell_test.go, instances_test.go, namespace_test.go or progress_test.go,
    while status.go (1123 lines), instances.go (602), shell.go (314), logs.go (311),
    progress.go (297) and namespace.go (163) all exist.

Why it matters: The house rule is that the test tree mirrors the source layout. Finding the tests for one source file, or seeing which file lacks failure-path cover, means searching one 8,000-line file.

Change: Split main_test.go along the source files, leaving the shared fakes and helpers in one helper file. It is a move; docs/test.md rows keep their names.

#### CR-0033 [nit] [new] Tests that need openssl, keytool or sh skip when the tool is absent, in CI too

Location: internal/gen/existingstore_test.go:79, internal/storebuild/storebuild_test.go:565, internal/statusscript/statusscript_test.go:853
Confidence: verified
Evidence:

    existingstore_test.go:79   t.Skip("openssl is not on PATH")
    storebuild_test.go:565     t.Skip("keytool is not on PATH; this Java interop check runs where a JDK is installed")
    statusscript_test.go:853   t.Skip("no POSIX sh on PATH or in Git for Windows, so the rendered health parse cannot be run here")

Why it matters: If a runner image drops one of these tools, the interop and health-script checks stop running and the pipeline stays green.

Change: Fail instead of skipping when `CI` is set, as the dev scripts already test for it.

### cicd

#### CR-0034 [minor] [new] tag.yml lacks the template's tag-format check and latest handling, so any v* tag publishes as Latest

Location: .github/workflows/tag.yml:38-40, .github/workflows/tag.yml:53-66, .github/workflows/tag.yml:291-295
Confidence: verified
Evidence:

    tag.yml:40   tags: ["v*"]
    tag.yml:291  gh release create "$TAG" \
    tag.yml:292    --verify-tag \
    init-devops assets/tag.yml (not in this file):
        if ! printf '%s' "$GITHUB_REF_NAME" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
        --latest="$LATEST" \

Why it matters: `v1.0`, `vfoo` or `v2.0.0-rc1` runs the whole pipeline and is published as the Latest release, and a hotfix tag on an older line takes Latest from the newest one.

Change: Re-deploy tag.yml from the current init-devops template; the repo's newer action pins (docker/login-action v4.7.0) carry over.

#### CR-0035 [minor] [new] CI runs all scan where the contract and the template say all cov scan, and tag.yml's header still lists cov

Location: .github/workflows/ci.yml:58-65, .github/workflows/tag.yml:8
Confidence: verified
Evidence:

    ci.yml:60    run: ./scripts/dev.sh all scan
    ci.yml:65    run: ./scripts/dev.ps1 all scan
    tag.yml:8    #     -> gates    build vet test cov scan, ubuntu + windows
    dev-script-contract.md:25  CI runs `all cov scan`, so the coverage thresholds gate every release rather

Why it matters: The header misdescribes what gates a release, and the coverage run is never exercised on a runner. DEVELOPMENT.md:29 documents `all scan`, so the docs match the file but not the contract.

Change: Adopt the template's `all cov scan`, or record in DEVELOPMENT.md that cov is local only and correct the tag.yml header in the template.

#### CR-0036 [minor] [new] The dev scripts are behind their contract: no race detector, no cov-queued skip, gofmt taken from PATH

Location: scripts/dev.sh:84-105, scripts/dev.ps1:101-107
Confidence: verified
Evidence:

    dev.sh:105   task_test() { run test go test -count=1 ./...; }
    dev.sh:87    if ! command -v gofmt >/dev/null 2>&1; then
    dev.sh:93    unformatted=$(gofmt -l .)
    dev.ps1:107  $unformatted = @(& gofmt -l . | Where-Object { "$_" -ne '' })
    dev-script-contract.md:12  ... use the selected toolchain's gofmt, `$(go env GOROOT)/bin/gofmt`
    dev-script-contract.md:13  ... Go runs them under the race detector (`CGO_ENABLED=1 go test -race -count=1 ./...`)

Neither script mentions `-race` or `COV_QUEUED`.

Why it matters: No gate runs the race detector although the tool has a spinner goroutine, parallel status workers and a watch loop. `full` runs the suite twice. The format check uses whichever gofmt is on PATH while `go` switches to the pinned toolchain.

Change: Bring both scripts up to the contract from the init-devops skeletons, mirrored.

#### CR-0037 [minor] [new] Local gates build the sibling checkout through go.work while CI builds the pin, and this repo's docs do not say so

Location: scripts/dev.sh:70-130, docs/DEVELOPMENT.md:15-47
Confidence: verified
Evidence:

    go.work:5 (untracked, .gitignore:26)   replace github.com/hafio/go-deploy-utility => ../go-deploy-utility
    go.mod:6                               github.com/hafio/go-deploy-utility v0.1.2
    dev.sh:72-73, 105, 114, 129            the go build, test, cov and scan invocations; none sets GOWORK
    ../go-deploy-utility is at v0.1.2-1-g83fa381; the one commit past the tag changes docs and graph files only.
    go-deploy-utility docs/consuming.md:88  5. In this repo, run `GOWORK=off ./scripts/dev.sh full` on a clean checkout and

A grep for go.work or GOWORK in scripts/, docs/, README.md and .github finds nothing.

Why it matters: The hand-off this project uses after every change (`./scripts/dev.sh all scan`) compiles whatever is checked out next door, uncommitted work included, and main is gated by CI only at tag time. Today the difference is documentation, so nothing is wrong yet. The gap is that this repo neither guards nor documents it; the sibling's release notes do.

Change: Add `GOWORK=off ./scripts/dev.sh all scan` to the release checklist in DEVELOPMENT.md, and have each task log say whether a workspace was in effect (`go env GOWORK`). Forcing GOWORK=off in every task would break the adopt-through-go.work flow consuming.md describes, so make it the default for `full` only.

#### CR-0038 [minor] [new] Dependabot covers GitHub Actions only and no scan is scheduled, so module advisories surface at tag time

Location: .github/dependabot.yml:8-24, .github/workflows/ci.yml:12-24
Confidence: verified
Evidence:

    dependabot.yml:9   - package-ecosystem: github-actions            (the only ecosystem)
    ci.yml:13          workflow_call:
    ci.yml:14          workflow_dispatch:
    ci.yml:19          pull_request:                                  (paths: workflows, dependabot.yml, the two dev scripts)

Why it matters: Running nothing on push is a recorded choice. Its cost here is that a new advisory for a Go module or the toolchain is seen only when someone runs `scan` or a tag pipeline fails, and the go-deploy-utility pin never gets a proposed bump.

Change: Add a gomod block to dependabot.yml and `go.mod`/`go.sum` to ci.yml's pull_request paths (a template change), or schedule `scan` weekly.

#### CR-0039 [minor] [new] The release version stamp comes from git describe, not from the tag being released

Location: scripts/dev.sh:62-73, .github/workflows/tag.yml:214-218
Confidence: verified
Evidence:

    dev.sh:67    VERSION="$(git describe --tags --dirty 2>/dev/null)" || VERSION="dev"
    dev.sh:73    go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$DIST/$BIN_NAME" ./cmd/solmq-conn-util
    tag.yml:216  TARGET_OS: ${{ matrix.target.os }}                   (the tag is not passed to the build)

Why it matters: With two tags on one commit (a re-tag after an infrastructure failure) describe can print the other tag, and a describe failure falls back to `dev`. Either ships binaries whose `version` does not name the release, and no step compares the two.

Change: Pass the tag to the build as the version when a tag ref is being built, and add a step that runs the linux/amd64 binary's `version` and compares it with the tag. Both belong in the template and skeletons.

#### CR-0040 [nit] [new] dev.sh takes the host OS from uname, which is not a GOOS under Git Bash

Location: scripts/dev.sh:54-59
Confidence: verified
Evidence:

    dev.sh:54  host_os()   { uname -s | tr '[:upper:]' '[:lower:]'; }
    dev.sh:59  [ "$T_OS" = "windows" ] && BIN_NAME="$BIN_NAME.exe"

Why it matters: Under Git Bash `uname -s` prints a `mingw64_nt-...` name, so a local `dev.sh build` on Windows fails. CI never runs dev.sh on Windows (ci.yml:59).

Change: Map `mingw*`, `msys*` and `cygwin*` to `windows` in host_os, in the skeleton.

### build

#### CR-0041 [major] [new] The kubernetes libs.download init container fetches jars with wget, with no checksum and with http accepted

Location: internal/deploy/deploy.go:592-601, internal/validate/validate.go:1069-1078, internal/validate/validate.go:1104-1112
Confidence: verified
Evidence:

    deploy.go:596     cmds = append(cmds, "wget -O '/libs/"+spec.BaseName(u)+"' '"+u+"'")
    deploy.go:601     w.Line(10, `command: ["sh", "-c", `+strconv.Quote(strings.Join(cmds, " && "))+`]`)
    validate.go:1076  add(fileEnv, "libs.download url %q must be http(s) and contain no spaces, quotes, or control characters", u)
    validate.go:1108  if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
    DESIGN.md:720     sidecar 404s is written and reported unverified. https is mandatory on every hop,

Why it matters: Jars that run inside the connector JVM are fetched at pod start with no integrity check, and over cleartext when the URL says http. The CLI downloader holds itself to https on every hop and a checksum; this path, which feeds the same classpath, holds neither.

Change: Require https in safeLibsURL, add a per-URL sha256 to libs.download and render `sha256sum -c` after each wget so a mismatch fails the init container. Test the http refusal and pin the checksum line in the deploy golden.

#### CR-0042 [minor] [new] No artifact states a read-only root or a network mode, and DESIGN does not say the root must stay writable

Location: internal/podmangen/podmangen.go:147-156, internal/deploy/deploy.go:605
Confidence: verified
Evidence:

    podmangen.go:153  // Identity and privileges are stated, never inherited. The image declares
    deploy.go:605     renderContainerSecurity(w, 10)
    grep -E 'read_only:|ReadOnly=|readOnlyRootFilesystem|network_mode|Network=|hostNetwork' over internal/ and testdata/: no hit
    grep -i 'read-only root|writable root|bridge' docs/DESIGN.md: no hit

The posture pass confirmed the rest is emitted and pinned: privileged off, every capability dropped, no-new-privileges, an explicit identity, secrets as files at absolute targets.

Why it matters: The house rule asks for bridge networking to be stated and the checklist for a read-only root to be emitted or documented as unsupported. Here both rest on engine defaults that no test pins.

Change: Emit the network mode, and either a read-only root with tmpfs for what the JVM writes or one line in DESIGN.md section 7 saying the image needs a writable root. Add both to the posture test.

#### CR-0043 [minor] [new] Quadlet values are written without percent escaping

Location: internal/podmangen/podmangen.go:148-151, internal/podmangen/podmangen.go:199-201
Confidence: verified
Evidence:

    podmangen.go:148  w.Line(0, "Image="+inst.Image)
    podmangen.go:149  w.Line(0, "ContainerName="+inst.Name)
    podmangen.go:200  w.Line(0, "Volume="+in.Libs.Source+":"+in.Libs.Target+":ro")
    go-deploy-utility quadlet.go:20  func EscapePercent(s string) string { return strings.ReplaceAll(s, "%", "%%") }
    go-deploy-utility execguard.go:81  const UnsafeChars = "\"'`$;|&<>()*?[]{}#!\\"       (no '%')

A grep for EscapePercent in cmd/ and internal/ finds no use. safeHostPath (validate.go:1165) is built from UnsafeChars, so a `%` passes, and the libs source is made absolute from the env.yaml directory, which validate never sees.

Why it matters: systemd expands `%` specifiers in a unit. A libs path or an env.yaml directory with a percent sign mounts a different directory or breaks the unit.

Change: Pass the Volume=, Image=, ContainerName= and Label= values through quadlet.EscapePercent, and add a podmangen test with a `%` in the libs path.

#### CR-0044 [minor] [new] The status script passes the actuator Basic credential on wget's command line

Location: internal/statusscript/statusscript.go:251-257
Confidence: verified
Evidence:

    statusscript.go:253  wget -qO- --header="Authorization: Basic $(printf %s:%s "$USER_NAME" "$PASS" | base64 | tr -d '\n')" "$1" 2>/dev/null

Why it matters: The header is an argument of the wget process, so the encoded credential is readable in the process list inside the container, and on the host for docker and rootful podman, on every health probe. The house rule keeps secrets off argv. The reach is limited: the account is the read-only status account.

Change: Record the decision and the account's reach in DESIGN.md section 10, or fetch through a client that takes the header from a file or stdin.

### interfaces

#### CR-0045 [major] [new] cli with a one-shot command and logs exit 0 without acting when several instances match

Location: cmd/solmq-conn-util/instances.go:571-575, cmd/solmq-conn-util/shell.go:193-201, cmd/solmq-conn-util/logs.go:222-230, docs/DESIGN.md:132-134
Confidence: verified
Evidence:

    instances.go:574  printInstancePicker(os.Stdout, sess, line, refs)
    instances.go:575  return ref, false, nil
    shell.go:197      if !found {
    shell.go:200          return exitcode.OK
    shell.go:167      // one-shot form is exactly what a script should be using instead.
    shell.go:238      // what a script wants back from `cli -- <command>` is the command's answer. It
    DESIGN.md:132     **`logs` and `cli` read one instance.** When several match and none is named,
    DESIGN.md:133     they print the matching commands, ready to paste, and exit 0. A guess or an

Why it matters: With two replicas, which the leader-election modes run, a scripted `cli -- <command>` runs nothing and exits 0, and `logs > app.log` writes the listing into the file. The recorded decision is about not guessing. It does not need exit 0, or stdout, for a run that cannot be interactive.

Change: Keep the listing. For a one-shot `cli --`, and for `logs` when stdout is not a terminal, print it to stderr and exit 2. Update DESIGN.md section 2 and operations.md.

#### CR-0046 [major] [new] Misspelled keys in env.yaml sections and workflow files are ignored without a message

Location: internal/spec/env.go:52-56, internal/validate/extra.go:107-112
Confidence: verified
Evidence:

    env.go:54            if err := yaml.Unmarshal(data, &raw); err != nil {       (no KnownFields)
    env.go:47            FilePattern string `yaml:"file_pattern"`                 (the one underscore key)
    javaoptions.go:38    // ParseEnv decodes without KnownFields, so without this a misspelled sub-key,
    kubernetes.go:166    // without KnownFields, so a deleted field would drop the block in silence.
    extra.go:153         warn(at.file, "%s: %q is not a key this tool knows; did you mean %q? ...      (passthrough blocks only)

Why it matters: `workflows.file-pattern`, `kubernetes.deployment.replica` or a mistyped `run-as` takes the default, and validate, deploy and status all stay quiet. The targeted scans for retired keys show the hazard is known; it is closed one key at a time.

Change: Add a reporting pass that walks the decoded node tree against the struct tags, as spec/extra.go does for the `solace:`/`mq:` blocks, for env.yaml's own sections and the workflow top level. Report an unknown key as a validate error, or a warning with a near-match hint.

#### CR-0047 [minor] [new] Refusals the documented rule makes exit 2 exit 1, and the same mistake exits differently by platform

Location: cmd/solmq-conn-util/main.go:584-600, cmd/solmq-conn-util/shell.go:168-170, cmd/solmq-conn-util/instances.go:484-488, cmd/solmq-conn-util/main.go:744-747, cmd/solmq-conn-util/main.go:926-929, docs/operations.md:61-72
Confidence: verified
Evidence:

    main.go:586        return usageExit(perr)                    (platform menu with no terminal: exit 2)
    main.go:599        return errExit(cerr)                      (remove confirmation with no terminal: exit 1)
    shell.go:169       return errExit(fmt.Errorf("cli opens an interactive shell and stdin is not a terminal; ...
    instances.go:486   return "", fmt.Errorf("%s %s is out of range: ...          (a bad argument, returned through errExit)
    main.go:746        return failAll(herrs)                     (docker: a store file env.yaml names is missing: exit 1)
    main.go:929        return usageExit(ferr)                    (podman: the same file: exit 2)
    operations.md:68   refused it (a preflight, a host check, a failed platform command); 2 refused

Why it matters: "stdin is not a terminal and no flag was passed" exits 2 for the platform menu and 1 for remove, deploy and cli. A missing store file exits 1 on docker and 2 on podman. A wrapper that retries on 1 retries calls that can never succeed. operations.md:229-230 documents exit 1 for the deploy question, so part of this is written down; the rule in section 1 says otherwise.

Change: Mark these refusals as usage errors, or name the exceptions in the exit-code paragraph. Pin each with a test.

### observability

#### CR-0048 [minor] [new] The closing ok and error lines name the action and platform only

Location: cmd/solmq-conn-util/main.go:1782-1792
Confidence: verified
Evidence:

    main.go:1788  fmt.Fprintf(os.Stderr, "error: %s %s failed: %v\n", action, platform, err)
    main.go:1791  fmt.Fprintf(os.Stderr, "ok: %s %s\n", action, platform)

Why it matters: With several env files, instances and contexts, `ok: deploy kubernetes` in a captured log does not say which object or file was changed.

Change: Have report, or one "acting on" line before the first engine call, name the instance, namespace, env.yaml path and image.

### conventions

#### CR-0049 [minor] [new] The README quick start validates and deploys a sample that ships none of the four PKCS12 stores its env.yaml selects

Location: README.md:27-36, internal/examples/files/env.yaml:81-98, internal/examples/examples.go:19-53, cmd/solmq-conn-util/main.go:1304-1305
Confidence: verified
Evidence:

    README.md:33     solmq-conn-util validate -e examples/env.yaml         # lint it
    README.md:35     solmq-conn-util deploy --platform docker -e examples/env.yaml
    env.yaml:85      existing:                           # a PKCS12 you provide (every store is PKCS12)
    env.yaml:86        pkcs12: ./certs/solace-ca.p12                       (four such stores, lines 83-98)
    examples.go:16   //go:embed files/*.yaml                               (only YAML is written)
    main.go:1305     fmt.Fprintf(os.Stderr, "next: solmq-conn-util generate config -e %s\n", ...

Why it matters: On a fresh sample the second and fourth quick-start commands report four missing store files. Nothing in the README says the stores and their password variables come first.

Change: Say so in the quick start, or stop it at `generate config`, which the sample does pass. Have `examples` print the same note.

#### CR-0050 [minor] [new] Messages, the starter files and the generator page point to a user guide that does not exist

Location: internal/validate/transform.go:45, internal/validate/transform.go:53, internal/validate/transform.go:209, internal/validate/extra.go:170, cmd/solmq-conn-util/commands.go:773, internal/examples/files/env.yaml:253, solmq-conn-util-generator.html:1655
Confidence: verified
Evidence:

    transform.go:45   ... = <SpEL expression>" (user guide section 6.7), or drop transform: to stay on ...
    transform.go:53   ... unless an expression sets it (user guide section 6.7)")
    env.yaml:253      # user guide, section 8.

`grep -i 'user guide'` finds 18 references outside tests: validate (4), commands.go (2), status.go (1), the shipped env.yaml (7) and workflow-0.yaml (1), the golden workflow-0.yaml (1) and the generator page (2). docs/ holds no user guide, and the section numbers do not map to configuration.md.

Why it matters: A validate warning and the files `examples` writes send the reader to a document that is not there.

Change: Replace each with the configuration.md section it means, updating the tests that pin the strings and the page copies in the same change.

#### CR-0051 [minor] [new] The shipped env.yaml and configuration.md describe behaviour that was removed

Location: internal/examples/files/env.yaml:192-193, internal/examples/files/env.yaml:252-253, docs/configuration.md:756, internal/gen/gen.go:198-211
Confidence: verified
Evidence:

    env.yaml:192           name: solmq-connector                 # DNS-1123 label (suffixed -1, -2, ... when >20 workflows split)
    env.yaml:193           namespace: solace-connectors          # deploy also emits a `kind: Namespace` for it
    env.yaml:252           # Whether the tool should run that login itself is an open decision -- see the
    configuration.md:756   ... `generate kubernetes` emits a `kind: Namespace` document first ...
    gen.go:198             // allowlist check; plain `generate kubernetes` leaves it empty, so an exotic
    main.go:573            "%s: %q is no longer a positional platform; pass --platform %s instead\n"
    DESIGN.md:56           and never run, with nothing saying why. The tool therefore refuses a folder with

Why it matters: The starter file every user copies says workflows are split and that deploy emits a Namespace; the tool refuses more than 20 workflows and deploy applies no Namespace. The documented `generate kubernetes` is rejected by the parser.

Suggested patch (not applied), starter file:

```diff
--- a/internal/examples/files/env.yaml
+++ b/internal/examples/files/env.yaml
@@ -189,8 +189,8 @@
 kubernetes:
   command: kubectl                        # or "oc", or "kubectl --context prod -n solace-connectors"
   deployment:
-    name: solmq-connector                 # DNS-1123 label (suffixed -1, -2, ... when >20 workflows split)
-    namespace: solace-connectors          # deploy also emits a `kind: Namespace` for it
+    name: solmq-connector                 # DNS-1123 label
+    namespace: solace-connectors          # deploy creates it only when it is absent
     replicas: 2                           # active_standby: 1 leader + standbys (standalone needs 1)
     resources:
       cpu: "1"
@@ -249,8 +249,7 @@
 # Compose has no field for it at all, and podman reads an auth file rather than a
 # secret, so both rely on the host already being logged in:
 #   docker login registry.internal       /  podman login registry.internal
-# Whether the tool should run that login itself is an open decision -- see the
-# user guide, section 8.
+# The tool never runs that login itself.
 docker:
   command: docker                         # or "docker --context foo"
   name: connector                         # same name the kubernetes pod gives the container
```

Suggested patch (not applied), configuration.md:

```diff
--- a/docs/configuration.md
+++ b/docs/configuration.md
@@ -753,7 +753,7 @@
 
 | Key | Meaning |
 |---|---|
-| `deployment.name`, `namespace` | required DNS-1123 labels. `generate kubernetes` emits a `kind: Namespace` document first, so a manual apply works; `deploy` applies none and creates the namespace on its own only when it is absent. `<name>-credentials` and `<namespace>-<name>-libs-pv` must stay within 63 characters |
+| `deployment.name`, `namespace` | required DNS-1123 labels. `generate --platform kubernetes` emits a `kind: Namespace` document first, so a manual apply works; `deploy` applies none and creates the namespace on its own only when it is absent. `<name>-credentials` and `<namespace>-<name>-libs-pv` must stay within 63 characters |
 | `deployment.replicas` | default `1`; `standalone` requires `1` |
 | `deployment.resources` | `cpu` and `memory`, Kubernetes quantities, written as identical requests and limits |
 | `deployment.node-selector` | node labels the node must carry; keys written sorted |
```

Change: Update the comments at gen.go:198 and gen.go:211 to `generate --platform kubernetes` when those lines are next edited.

#### CR-0052 [minor] [new] Tests and docs cite the deleted review ledger, stage labels and numbered rules that no file defines

Location: cmd/solmq-conn-util/main_test.go:3105, cmd/solmq-conn-util/main_test.go:6721, internal/gen/gen_extra_test.go:1494, internal/validate/validate_extra_test.go:1031, internal/validate/validate_extra_test.go:1230, docs/test.md:12-30, internal/spec/expand.go:23-32
Confidence: verified
Evidence:

    main_test.go:3105            // TestDeployDockerSeamPipesComposeAndRuns is CR-0003: the compose file inlines
    main_test.go:6721            // TestReadVerbsAskKubernetesNoPermissionFirst is CR-0006 taken to its end: status,
    gen_extra_test.go:1494       // TestGenerateKubernetesRefusesAnInjectedQuantity is CR-0001 end to end: a
    validate_extra_test.go:1031  // TestRestartIsEachEnginesOwnWord is CR-0025: docker's restart lands in the
    test.md:19                   S7 (containercli and inspect, the redeploy prompt) adds a test for every branch it introduces and
    expand.go:23                 // expandRE matches only the braced form: a bare $VAR is literal text (rule 4).

Why it matters: The house rule is to cross-reference by file and test name, never by ticket number. The ledger these IDs came from is gone, and this run restarts the numbering, so each of the five comments now names an unrelated finding in this report. The S7-S9 labels and "rule N" point at numbering no file in the repo defines.

Change: Replace each ID with the test's file and name, or with the guarantee in words. Say what each numbered rule is. Reduce docs/test.md lines 12-30 to the current floor and what it covers.

#### CR-0053 [minor] [new] Two validate messages and the four starter workflow files carry non-ASCII characters

Location: internal/validate/validate.go:382, internal/validate/validate.go:867, internal/examples/files/workflow-0.yaml:1-54, testdata/golden/specs/workflow-0.yaml:1-54
Confidence: verified
Evidence:

    validate.go:382  add(file, "%s (%s) uses conn-ref %q, so it may set only queue/topic <em dash> move the connection details ...
    validate.go:867  add(file, "conflicting password for the same binder: this side uses %s, %s uses a different one <em dash> one ...
    grep -cIP '[^\x00-\x7F]' per tracked file: internal/examples/files/workflow-0.yaml 5, workflow-1.yaml 2,
    workflow-2.yaml 3, workflow-3.yaml 3, the four testdata/golden/specs twins the same; three Go comments
    (consolidate.go, uuid.go, spec.go) and validate.go:1105.

Why it matters: The rule is plain ASCII in everything written, script output and generated files included. These two strings reach terminals, PowerShell 5.1 among them, and the workflow files are written to the user's disk by `examples`.

Change: Replace the em dashes in the two messages with ` -- ` and update the tests that pin them. Convert the four starter workflows and their golden twins together. Leave the comment lines for their next edit.

#### CR-0054 [minor] [new] docs/test.md carries no support notice and the list that enforces it is explicit, so a new doc is exempt

Location: cmd/solmq-conn-util/support_test.go:15-28, docs/DEVELOPMENT.md:64-68, docs/test.md:1-10
Confidence: verified
Evidence:

    support_test.go:19   var supportNoticeDocs = []string{                (eight paths; docs/test.md is not one)
    DEVELOPMENT.md:66    the shared module's `notice` package) verbatim under the H1 of every
    DEVELOPMENT.md:67    hand-written doc, and `TestSupportNoticeInGeneratorPage` holds the
    test.md:1            # solmq-conn-util test catalogue
    test.md:3            Every test in the suite, grouped by package and expanded to individual cases, so you can

Why it matters: The enforced rule is narrower than the documented one, and a doc added later is exempt until someone remembers the list.

Change: Add the notice to docs/test.md and its path to the list, or name the exemption in DEVELOPMENT.md. Better, check every docs/*.md and README.md except an explicit exempt list.

#### CR-0055 [minor] [new] The README says every artifact runs non-root; run-as image runs the image's own user

Location: README.md:23, docs/configuration.md:935
Confidence: verified
Evidence:

    README.md:23           - Every artifact runs the connector non-root with every capability dropped.
    configuration.md:935   | `image` | the image's own user (root); no ids written. Use on OpenShift, where the SCC assigns them |

Why it matters: A reader who takes the posture from the README assumes non-root for every configuration.

Suggested patch (not applied):

```diff
--- a/README.md
+++ b/README.md
@@ -20,7 +20,8 @@
   connector as mounted files, never as environment variables or literals.
 - TLS and mTLS are wired per connection from PKCS12 stores you provide, stores
   `deploy` builds from your certificate files, or Secrets you already manage.
-- Every artifact runs the connector non-root with every capability dropped.
+- Every artifact drops every capability and runs the connector as a non-root
+  user, unless you set `run-as: image`, which leaves the identity to the image.
 - `status`, `logs` and `cli` report, read and enter the running instances on
   any of the three platforms.
 
```

#### CR-0056 [minor] [new] The project CLAUDE.md tells the agent to run graphify update, against the user-run rule and the dev-script task

Location: CLAUDE.md:9, scripts/dev.sh:145-150
Confidence: verified
Evidence:

    CLAUDE.md:9   - After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
    dev.sh:149    run graphify graphify update .
    global rule:  Graph refreshes -- incremental or full -- are user-run like every gate (S1); when the graph is stale, propose the task.

Why it matters: Following either file breaks the other. This review reused the graph for that reason.

Suggested patch (not applied):

```diff
--- a/CLAUDE.md
+++ b/CLAUDE.md
@@ -6,4 +6,4 @@
 - For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
 - If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
 - Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
-- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
+- After modifying code, propose `./scripts/dev.sh graphify` to refresh the graph (AST-only, no API cost). The user runs it, like every gate.
```

#### CR-0057 [minor] [new] File-read errors flatten the cause with %v where the rule asks for the chain

Location: internal/gen/stores.go:102, internal/gen/gen.go:1183, internal/spec/spec.go:191, internal/spec/env.go:55
Confidence: verified
Evidence:

    stores.go:102  return storebuild.Source{}, fmt.Errorf("reading %s %q: %v", where, path, err)
    gen.go:1183    return nil, nil, fmt.Errorf("reading %s %q: %v", StoreField(st.Name), st.File, err)
    spec.go:191    return nil, fmt.Errorf("%s: %v", filepath.Base(path), err)
    env.go:55      return nil, fmt.Errorf("env.yaml: %v", err)

The conventions pass counted about 55 such wraps against 97 with %w in non-test code.

Why it matters: The rule is to preserve the cause when wrapping. With %v a caller cannot tell a missing store file from an unreadable one with errors.Is.

Change: Use %w at the file-read sites. Where a YAML parse error is flattened on purpose, say so in a comment.

#### CR-0058 [minor] [new] 241 graphify cache files and a settings backup are tracked although the cache path is ignored

Location: .gitignore:18, graphify-out/cache/last_query_stamp:1, .claude/settings.json.graphify-bak:1-23
Confidence: verified
Evidence:

    .gitignore:18     graphify-out/cache
    git ls-files graphify-out/cache | wc -l   -> 241
    git ls-files .claude                      -> settings.json, settings.json.graphify-bak
    git status --short (after this review's graph queries)  ->  M graphify-out/cache/last_query_stamp

Why it matters: An ignore rule does not untrack what is already committed. Every graph query dirties the tree, this review's own included, and the backup is leftover tool state.

Change: You run `git rm -r --cached graphify-out/cache .claude/settings.json.graphify-bak` and commit; the ignore rule then holds.

#### CR-0059 [nit] [new] DEVELOPMENT.md's dependency list omits go-deploy-utility

Location: docs/DEVELOPMENT.md:17-19, go.mod:5-10
Confidence: verified
Evidence:

    DEVELOPMENT.md:17  Requires Go 1.27+ (the `go` directive in `go.mod`). Dependencies:
    DEVELOPMENT.md:18  `gopkg.in/yaml.v3`, `software.sslmate.com/src/go-pkcs12` (writes created
    go.mod:6           github.com/hafio/go-deploy-utility v0.1.2

Why it matters: The module that holds the runner, the guards and the engine argv builders is the one not named.

Change: Add it to the sentence and point to DESIGN.md section 12 (with CR-0037).

#### CR-0060 [nit] [new] .gitignore keeps unanchored scratch names and *.p12 has no binary attribute

Location: .gitignore:12-23, .gitattributes:1-4
Confidence: verified
Evidence:

    .gitignore:12     eg
    .gitignore:15     to-be-added.md
    .gitignore:23     tmp
    .gitattributes:3  * text=auto eol=lf
    .gitattributes:4  *.jks binary

Why it matters: An unanchored `eg` or `tmp` ignores a file or directory of that name at any depth, a future fixture included. The two tracked .p12 fixtures are detected as binary by content today; an attribute would not depend on that.

Suggested patch (not applied), attributes only; anchor or drop the scratch names by hand:

```diff
--- a/.gitattributes
+++ b/.gitattributes
@@ -2,3 +2,5 @@
 # regardless of core.autocrlf. .jks fixtures are binary.
 * text=auto eol=lf
 *.jks binary
+*.p12 binary
+*.pfx binary
```

#### CR-0061 [nit] [new] The module path differs from the repository and the README states no licence

Location: go.mod:1, README.md:1-10
Confidence: verified
Evidence:

    go.mod:1   module github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
    origin:    github.com:hafio/solace-ibmmq-connector-helper
    LICENSE:1  GNU GENERAL PUBLIC LICENSE (Version 3)

Why it matters: Nothing breaks today, since the docs tell users to take a release binary. A `go install` from the repository path would fail, and a reader has to open LICENSE to learn the terms.

Change: Add one line to DEVELOPMENT.md that the path is deliberate, and a licence line to the README.

## Suspected (needs confirmation)

#### CR-0062 [minor] [new] kubernetes restart after stop reports ok while no replica runs

Location: internal/runner/runner.go:233-250, cmd/solmq-conn-util/lifecycle.go:89-93
Confidence: suspected
Evidence:

    runner.go:241     c = kube.Scale(argv, "deployment", name, namespace, 0)            (stop)
    runner.go:245     c = kube.RolloutRestart(argv, "deployment", name, namespace)      (restart)
    lifecycle.go:91   if err == nil && action != runner.ActionStop {
    lifecycle.go:92       out, err = waitForRollout(r, k, out, approvals)

Why it matters: After `stop` the Deployment has zero replicas. `rollout restart` only patches the pod template, and the rollout wait then succeeds with nothing running, while docker and podman `restart` start a stopped instance.

To confirm: on a cluster, `stop`, then `restart`, then `kubectl get deploy <name>`.

#### CR-0063 [minor] [new] status --all --details reads kubectl top output as four columns, but --all-namespaces adds a fifth

Location: internal/statusreport/parse.go:497-521, internal/runner/runner.go:886-898
Confidence: suspected
Evidence:

    runner.go:891   argv = append(argv, "--all-namespaces")
    parse.go:505    pod, container, cpu, mem := f[0], f[1], f[2], f[3]
    parse.go:514    u, ok := byPod[insts[i].Name]

Why it matters: With a leading NAMESPACE column, f[0] is the namespace and no row matches an instance name, so `--details` shows no usage and says nothing. Two pods of one name in different namespaces would also share a map entry.

To confirm: `kubectl top pod -A --containers --no-headers | head -2` on a cluster with metrics-server.

#### CR-0064 [minor] [new] kubernetes status strings skip the control-character filter the docker and podman path applies

Location: internal/statusreport/parse.go:240-246, internal/statusreport/parse.go:569-574
Confidence: suspected
Evidence:

    parse.go:243   c.Image = cs.Image                                (kubernetes: as supplied)
    parse.go:569   inst := Instance{Name: inspect.Printable(ec.Name)}       (docker and podman: filtered)
    parse.go:574   Image:     inspect.Printable(ec.Image),

Why it matters: `status --all` on a shared cluster prints strings taken from other tenants' pods. If the API server admits control characters in any of them, they reach the terminal unfiltered.

To confirm: check whether the API server accepts an escape byte in a container image or a waiting reason, then render one through `status --all`.

#### CR-0065 [minor] [new] The init-container download names its file after the raw URL, query string included

Location: internal/deploy/deploy.go:594-597
Confidence: suspected
Evidence:

    deploy.go:596   cmds = append(cmds, "wget -O '/libs/"+spec.BaseName(u)+"' '"+u+"'")

Why it matters: For `https://repo/a.jar?token=x` the file would be `/libs/a.jar?token=x`, which a `*.jar` classpath does not pick up, with no error at deploy time.

To confirm: read spec.BaseName for query handling and how the image builds its classpath.

#### CR-0066 [minor] [new] A repeat podman remove fails at systemctl stop once the unit is gone

Location: internal/runner/runner.go:790-798
Confidence: suspected
Evidence:

    runner.go:793   o, err := gdurunner.Combined(ctx, r, sc.WithEnv(systemd.Stop(sc.UserMode, s)))
    runner.go:796   return out.String(), fmt.Errorf("systemctl stop %s: %w", s, err)

Why it matters: systemd answers a stop for a unit that is not loaded with a non-zero status, so a second `remove`, a normal cleanup step, would report failure for something already gone. kubernetes (`--ignore-not-found`) and docker (`compose down`) are idempotent.

To confirm: `systemctl --user stop solmq-absent.service; echo $?`.

## Dynamic checks

None run. Offered once, with the plan; the owner runs gates:

- `GOWORK=off ./scripts/dev.sh all scan` - shows main passes its gates against the pinned shared module (CR-0037).
- `GOWORK=off CGO_ENABLED=1 go test -race -count=1 ./...` - no gate runs the race detector (CR-0036).
- `systemctl --user stop solmq-absent.service; echo $?` - settles CR-0066.
- `kubectl top pod -A --containers --no-headers | head -2` - settles CR-0063.
- `shellcheck scripts/dev.sh` and `actionlint`, if installed.

The suggested patches were checked with `git apply --check` against 298cccf; none was applied.

## Fixed since last report

First run.

## Accepted (not re-litigated)

None.

## Coverage

- Areas reviewed: design, correctness, resilience, quality, tests, cicd, build, interfaces, observability, conventions; skipped: none.
- Read closely: every non-test Go file under cmd/solmq-conn-util and internal/ except completion.go and abbreviation.go (seven area passes; the config pipeline, the artifact, store, download and status packages, and cmd with internal/runner were each read in full by one pass); .github/workflows, .github/dependabot.yml, scripts/dev.sh, scripts/dev.ps1, go.mod and the init-devops contract and templates. Every blocker and major was re-read in this session at the cited lines and around them, including the deploy and remove paths (main.go:510-1080), namespace.go, lifecycle.go, env.go, expand.go, render.go, validate/transform.go, dev.sh and both workflows; each minor and nit was spot-checked at its cited lines.
- Sampled: the test tree (helper and fake sections in full, bodies by grep and sample); the docs (about 45 claims checked against code; configuration.md from section 9.5 on was not read line by line); internal/validate/validate.go, internal/gen/gen.go, internal/runner/runner.go and cmd status.go for the quality area.
- Not read: the JavaScript in solmq-conn-util-generator.html (string and hygiene greps only); completion.go beyond its use of the command model; go-deploy-utility, which a parallel review owns (its pinned v0.1.2 copy was read only to judge this app's use of it).
- Stacks with core checklist only: PowerShell (scripts/dev.ps1), HTML/JavaScript (the generator page).
- Dynamic checks not run: all five listed above.
- Limits: static review; the graph was reused, not refreshed (it was built from HEAD); `graphify query` writes the tracked graphify-out/cache/last_query_stamp, so this review's queries show in `git status` (CR-0058); one cross-repo lead is recorded in CR-0007 (the namespace listing in go-deploy-utility kube/nsgate); one agent claim was dropped after checking ("the release can publish with missing assets": gh creates the release as a draft while assets upload); line numbers are those of 298cccf.
