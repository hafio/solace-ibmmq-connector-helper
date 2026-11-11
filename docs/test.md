# solmq-conn-util test catalogue

Every test in the suite, grouped by package and expanded to individual cases, so you can
see what behavior is covered and jump to the test that covers it. This is a living
document: when a test or case is added, removed, or renamed, update the matching row in
the same change. For build and release details see [DEVELOPMENT.md](DEVELOPMENT.md);
for using the tool, see [getting-started.md](getting-started.md).

Run the suite with `./scripts/dev.sh test` (`.\scripts\dev.ps1 test` on Windows);
measure coverage with the `cov` task.

Last recorded `cov` total, from `scripts/logs/cov.log` (2026-11-08): **96.3%**, after S9. An earlier
adoption of the shared `containerart` packages deleted the covered helpers
`composeQuote`, `composeEscape`, `flowList`, `renderSecrets` and `systemdEnv`; their tests
moved to go-deploy-utility with them. Removing covered statements lowers the ratio. No new uncovered
statement came in: `internal/dockergen`, `internal/podmangen` and `javaOptionsProblem` are
fully covered.

S7 (containercli and inspect, the redeploy prompt) adds a test for every branch it introduces and
removes the covered `healthStatus`, `EngineNamesByImage` and the hand-built engine argv. The total
rose from 96.0% to 96.2%, which is the new floor.

S8 (the shared `watch` loop under `status -w`) replaces the untested `watchStatus` ticker and
`clearScreen` with `gduwatch.Loop` behind a test seam, and covers it with the `TestWatchStatus*`
tests and `TestWatchTitleNamesTheViewAndItsFlags`. The total rose from 96.2% to 96.3%, which is
the new floor.

S9 (the shared `kube`, `kube/rbac` and `kube/nsgate` packages) removes the covered local `can-i`
parser (`canIDenied`, `canIProbe`), the namespace occupancy parsers (`namespaceOccupants`, `isOurs`,
`isClusterDefault`) and the hand-built kubectl argv, whose tests moved to go-deploy-utility with
them (`TestNamespaceOccupants*` are deleted here for that reason). It adds a test for every
branch it introduces: the probe sets (`TestKubernetesProbesDeployAndRemoveSets`,
`TestKubernetesProbeSets`), the namespace flow (`TestDeployNamespaceFlow`), the rollout wait
(`TestDeployWaitsForTheRollout`, `TestLifecycleRolloutWait`) and the namespace gate's new steps
(`TestRemoveNamespaceAsksThePermissionBeforeTheQuestion`, `TestRemoveNamespaceDeleteOutcome`).
`deploy.Render` now writes the Namespace document through `kube.NamespaceManifest` and returns its
error, so the local copy is gone (`TestRenderRefusesANamespaceThatIsNotALabel`). Validation's
`isDNS1123` gave way to `kube.ValidLabel`, the same rule.
The first `cov` read 96.2%, under the 96.3% floor: the moved code was well covered, and five new
guards cannot be reached. `TestKubernetesRolloutStatusRejectsUnsafeCommand` covers the rollout
wait's guard, and `ExecArgv` no longer repeats `kube.Exec`'s empty-command check, so
`TestExecArgvRefusesAnEmptyCommand` now reaches the shared refusal. With them the total is back at
96.3%, which stays the floor. These stay uncovered on purpose, because validation refuses the
input first: the probe-set errors in `actKubernetes` and the lifecycle path (an unknown action or kind),
the command guard in `actKubernetes` (`validate.Run` already held the command to it), and the Namespace
document's refusal in `ensureNamespace` and `GenerateKubernetes`.

The seven package-level seams in `cmd/solmq-conn-util` (the host facts, the prompter, the stderr
terminal check, the library downloader, the status worker count and the progress line's frame rate and clock) moved onto one explicit `app`
value that `run` builds once and threads through `dispatch` to the code that reads them. No
behaviour changed and no test was added, removed or renamed; a test builds its own with `testApp`
and sets the field it needs instead of overwriting a package variable. The floor stays 96.3%, to be
confirmed by the next `cov` run.

## Contents

- [internal/scan](#internalscan)
- [internal/spec](#internalspec)
- [internal/consolidate](#internalconsolidate)
- [internal/tls](#internaltls)
- [internal/storebuild](#internalstorebuild)
- [internal/yamlwriter](#internalyamlwriter)
- [internal/render](#internalrender)
- [internal/logback](#internallogback)
- [internal/statusscript](#internalstatusscript)
- [internal/storescript](#internalstorescript)
- [internal/deploy](#internaldeploy)
- [internal/dockergen](#internaldockergen)
- [internal/podmangen](#internalpodmangen)
- [internal/runner](#internalrunner)
- [internal/validate](#internalvalidate)
- [internal/examples](#internalexamples)
- [internal/gen](#internalgen)
- [internal/libs](#internallibs)
- [internal/statusreport](#internalstatusreport)
- [cmd/solmq-conn-util](#cmdsolmq-conn-util)
  - [logs](#logs)
  - [cli](#cli)
  - [remove / instance resolution](#remove--instance-resolution)

## How the suite is built

- **Table-driven tests** iterate a list of cases; each case is its own row below.
- **Golden-file tests** assert generated output byte-for-byte against fixtures under
  [`testdata/golden/`](../testdata/golden) (driven by `internal/gen/golden_test.go`);
  the deterministic ordered emitters make that output stable.
- **The exec seam** is the shared `runner.Runner` (`github.com/hafio/go-deploy-utility/runner`),
  faked by `fakeRunner` (and `queueRunner`, `streamRunner`, `attachRunner`, `splitRunner`),
  which record the argv and stdin crossing the boundary instead of starting a process and
  write their scripted output to the writers `Run` is handed. The real `os/exec` path, and
  its child-process tests, live in the shared module's own suite.
- **The app seams** are the other process-wide inputs a command reads: the host facts, the
  prompter, the stderr terminal check, the library downloader, the status worker count and the progress frame rate and clock.
  They are the fields of `app` (main.go), built live by `liveApp` and per test by `testApp`
  (main_test.go), which starts from an ordinary user (euid 1000, no session), a prompter over the
  real stdin, four workers, a 120ms frame and the real clock. A test that needs another value sets
  the field through `withPrompter`, `useHost`, `useFakeDownload`, `withStderrTerminal`,
  `withScriptWorkers`, `withProgressFrameRate` or `withProgressClock`
  and dispatches with that `app`; nothing is package-wide, so nothing is restored on cleanup.
- **Columns**: a `-` in the Case column means the test runs once; otherwise Case carries
  the subtest name or the case's label/input value, in source order where the test defines
  one.
- Tests are cross-referenced by file and test name only -- no line numbers (they rot as
  tests move).

_Snapshot: 1074 test functions, 1534 case rows across 20 packages. (Functions counted from `func Test` in the source; case rows are the data rows of the tables below, not a suite run -- human, please confirm against `./scripts/dev.sh test` / `cov` output.)_

## internal/scan

Discover workflow files -- YAML-only, env-file exclusion, wildcard matching, and metacharacter rejection.

Tests: [scan_test.go](../internal/scan/scan_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestScanSortsYAMLOnly | - | result sorted to [10.yml, 20.yaml], non-yaml and dirs ignored, Dir set to input dir |
| TestScanSortsLikeLs | - | 19/2/10/9.yaml come back as [10, 19, 2, 9] -- the order LC_ALL=C ls lists them, which is the numbering the connector and the generator page both use |
| TestScanSortsPrefixedNamesLikeLs | - | the same ordering when the digits follow a shared prefix, as in the workflow-N.yaml names the examples verb writes: workflow-1, workflow-10, workflow-2 |
| TestScanExcludesEnvFile | - | env.yaml excluded even with pattern '*', only workflow-0.yaml returned |
| TestScanEnvFileExcludedRegardlessOfPattern | - | pattern 'env*' still excludes env.yaml, only envoy.yaml remains |
| TestScanEmptyPatternDefaultsToStar | - | empty pattern behaves as '*', matches a.yaml |
| TestScanErrorMissingDir | - | scanning nonexistent directory returns an error |
| TestScanPatternWildcards | workflow-* | trailing star matches workflow-0.yaml and workflow-1.yaml only |
| TestScanPatternWildcards | `*hoc*` | mid-string star matches only adhoc.yaml |
| TestScanPatternWildcards | *-1.yaml | leading star matches only workflow-1.yaml |
| TestScanPatternNoMatchIsEmptyNotError | - | non-matching pattern 'nope*' yields empty results, no error |
| TestScanRejectsNonStarMetachars | [bad | pattern with bracket metachar is rejected with error |
| TestScanRejectsNonStarMetachars | wf?.yaml | pattern with '?' metachar is rejected with error |
| TestScanRejectsNonStarMetachars | a]b | pattern with ']' metachar is rejected with error |
| TestScanRejectsNonStarMetachars | a\b | pattern with backslash metachar is rejected with error |
| TestMatchStar | *,anything.yaml | bare star matches any name -> true |
| TestMatchStar | exact.yaml,exact.yaml | identical literal pattern and name match -> true |
| TestMatchStar | exact.yaml,other.yaml | literal pattern mismatch -> false |
| TestMatchStar | pre*,prefix.yaml | trailing star prefix match -> true |
| TestMatchStar | *.yaml,x.yaml | leading star suffix match -> true |
| TestMatchStar | *.yaml,x.yml | leading star suffix mismatch -> false |
| TestMatchStar | `a*b*c,axxbyyc` | multiple stars match interspersed segments -> true |
| TestMatchStar | `a*b*c,axxc` | multiple stars but missing required segment -> false |
| TestMatchStar | **,anything | consecutive stars still match any name -> true |
| TestIsYAML | a.yaml | isYAML true for .yaml extension |
| TestIsYAML | a.yml | isYAML true for .yml extension |
| TestIsYAML | A.YAML | isYAML true case-insensitively |
| TestIsYAML | a.txt | isYAML false for .txt extension |
| TestIsYAML | yaml | isYAML false when no extension present |
| TestIsYAML | a.yamlx | isYAML false for non-exact extension match |

## internal/spec

Parse env.yaml into the typed model -- workflows, defaults, named connections, the kubernetes/docker/podman platform sections, and ports -- and apply section defaults.

Tests: [spec_test.go](../internal/spec/spec_test.go), [env_test.go](../internal/spec/env_test.go), [targets_test.go](../internal/spec/targets_test.go), [expand_test.go](../internal/spec/expand_test.go), [defaults_test.go](../internal/spec/defaults_test.go), [image_test.go](../internal/spec/image_test.go), [javaoptions_test.go](../internal/spec/javaoptions_test.go), [kubernetes_test.go](../internal/spec/kubernetes_test.go), [transform_test.go](../internal/spec/transform_test.go), [extra_test.go](../internal/spec/extra_test.go), [runas_test.go](../internal/spec/runas_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestParseWorkflowSolaceAndMQ | - | parses full solace source and mq target with dest kind, tls, truststore/keystore selections, key alias, props |
| TestParseWorkflowConnRef | - | mq source with conn-ref resolves ConnRef, DestKind queue, Dest A.IN; SetsConnFields false |
| TestBaseName | url / posix / windows / bare / empty | one shared BaseName splits on both separators, so a Windows-authored path resolves identically on Linux |
| TestWorkflowFileLess | 10 vs 2 / workflow-10 vs workflow-2 / 1 vs 10 | names compare byte by byte, the order a directory listing shows, so 10.yaml takes an earlier id than 2.yaml and a name that is a prefix of another sorts first |
| TestWorkflowFileLess | workflow-02 vs workflow-10 / 10.yaml vs 10.yml / B vs a | zero-padded names keep their numbers in step with their ids, equal stems fall through to the suffix, and the order is bytes, not a locale's collation (upper case first) |
| TestWorkflowFileLess | reversed pairs / x vs x | the reverse of every pair is false and no name is less than itself -- the strict order sort.Slice requires |
| TestConnRefSideMayTuneBinding | - | consumer block parses on a conn-ref side, SetsConnFields ignores it, and Resolve keeps it alongside the referenced tuple and destination |
| TestParseDefaultsConnectionsAndLeaderElection | - | parses 2 named connections and leader-election active_standby with fail-over |
| TestParseDefaultsLeaderSession | - | an inline `session:` block parses the full solace tuple, api-properties included, and leaves the SolaceKey marker unset |
| TestKnownKeysReadsYAMLTags | - | the schema is read off a raw struct's yaml tags the way yaml.v3 names them: the tag's first element, options dropped, an untagged field lower-cased, a "-" tag or an unexported field skipped |
| TestToolManagedKeysAreSchemaKeys | solace / mq / unknown system | every credential key is tool-managed and every tool-managed key is a sorted schema key -- except mq's ssl-bundle, which is derived rather than read; an unknown system has no keys |
| TestCanonicalKey | kebab / camelCase with an acronym / upper snake / mixed / dots and brackets | CanonicalKey folds case and the - and _ separators the way Spring's relaxed binding does, and nothing else |
| TestParseWorkflowCapturesExtraKeys | block style / flow style / known keys only | every key a side's block carries that the schema does not name is captured in file order with its value as written -- a bool, a camelCase name, a double-quoted string, a nested mapping and a list whole -- known keys and a transform key (the scan's) left out, and a block of known keys alone captures nothing |
| TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys | - | a << merge is expanded in place with the block's own key winning over a merged one and a merge source's own key over what it merges in turn, a block written as an alias reads as the block it names, and an aliased value is the value it names |
| TestParseDefaultsCapturesExtraKeys | connection mq / connection solace / both systems / session / mq-defaults / empty | each env.yaml site read through the side structs captures its other keys, a connection naming both systems carries nothing, mq-defaults parses as a block of its own, and an empty env.yaml has none |
| TestSideSetsConnFieldsCountsExtraKeys | - | an extra key, or a truststore/keystore selection, beside conn-ref is a connection field; a bare conn-ref side sets none |
| TestResolveCarriesTheConnectionsExtraKeys | - | a conn-ref side resolves to the connection's other keys along with its tuple, keeping its own destination |
| TestMisplacedEnvTransformsCoversLeaderSession | - | a transform under the inline leader-election session is reported as misplaced through ParseEnv, and is not captured as an extra key |
| TestParseWorkflowTypeErrorInsideASideStillNamesTheLine | - | a wrong scalar type on a known key still fails with yaml's own error naming the line, so the capture never hides it |
| TestExpandLeavesExtraKeysAlone | - | a ${...} inside another key is never expanded (rule 6), reaching Spring as typed |
| TestParseDefaultsLeaderSolaceKeyRetired | mapping / scalar / list | a `solace:` key under leader-election parses whatever shape it holds, sets SolaceKey and never populates Session, so validate can error naming `session:` |
| TestSideBindingFields | bare tuple | no destination and no tuning means no binding fields |
| TestSideBindingFields | queue / topic | the destination kind is reported |
| TestSideBindingFields | consumer / producer / both | per-binding tuning is reported, in schema order |
| TestResolveConnRef | known ref edge | resolves host/msg-vpn/key-alias from connections map, keeps dest |
| TestResolveConnRef | unknown ref nope | returned unchanged with ConnRef nope and empty Host |
| TestParseWorkflowEnabledDefaultsTrue | - | enabled defaults true and target stays unset when absent |
| TestParseWorkflowAmbiguousSystemAndDest | solace and mq both set | HasSystem returns false when both systems present |
| TestParseWorkflowAmbiguousSystemAndDest | queue and topic both set | DestKind is empty string when queue+topic ambiguous |
| TestParseWorkflowSyntaxError | - | malformed yaml returns non-nil error |
| TestParseDefaultsFull | - | parses named tls stores (mq-ca, mq-client) naming their existing PKCS12, management port 8090 with exposure health (not a configurable key, but parsed anyway so validate can reject it), security enabled false with 1 user, leader-election standalone, logging/solace-defaults nodes captured |
| TestParseSecurityUserRoles | absent / one / several | security.users[].roles parses to no roles (the connector's read-only default), a single role, and several in authored order |
| TestParseDefaultsSecurityEnabledKeyOmittedStaysNil | - | security.enabled is not a configurable key: an omitted key parses to Security.Enabled nil rather than being defaulted |
| TestParseDefaultsEmpty | - | empty input yields a zero-valued Management (Management{}), Security.Enabled nil with no users, and TLS.Truststore nil |
| TestParseDefaultsError | - | malformed tls yaml returns non-nil error |
| TestParseKubernetesReplicasDefault | - | deployment without replicas defaults Replicas to 1, and an unset run-as to the non-root default |
| TestParseRunAs | valid: uid:gid / bare uid / 0 / the maximum / image. refused: empty / no uid / no gid / leading zero / sign / exponent / whitespace / a name / three fields / past 2147483647 / Image | the run-as grammar: decimal ids only, a bare uid runs with gid 0, image keeps the image's user, and every refusal says which id and why, since the ids are written into artifacts as-is |
| TestDefaultRunAsIsNonRootAndParses | - | the default parses, is not root, and fits a stock rootless subuid/subgid range of 65536 ids |
| TestParseKubernetesFull | - | parses replicas 2, service enabled port 8090, credentials create name, stores create present; source and variables set on credentials.create parse structurally regardless, so RemovedKeys can report them |
| TestParseKubernetesError | - | deployment as sequence instead of map returns non-nil error |
| TestParseKubernetesNodeSelector | - | deployment.node-selector decodes to its label map; true and 2 arrive as the text written, an empty value as "", and an unset selector stays nil |
| TestParseKubernetesScheduling | - | tolerations decode in Kubernetes' shape (tolerationSeconds as a pointer), spread is kept as written, and affinity is captured as its YAML node |
| TestParseKubernetesResources | - | parses deployment resources CPU '1' and Memory 1Gi |
| TestParseKubernetesLoggingLibsDefaults | syslog and libs download present | syslog Protocol defaults to udp and libs download Image defaults to busybox:1.37 |
| TestParseKubernetesLoggingLibsDefaults | libs pvc create without storage | pvc create Storage defaults to 1Gi |
| TestParseKubernetesLoggingLibsDefaults | no logging or libs block | Logging and Libs stay nil when absent |
| TestParseEnvEmpty | - | empty file yields Workflows dir '.' pattern '*', Kubernetes/Docker/Podman nil, defaults zero-valued |
| TestWorkflowsFromRawDefaultWhenAbsent | - | workflows section absent defaults dir '.' and file pattern '*' |
| TestWorkflowsFromRawDirOverride | - | dir override /custom/dir applied, file pattern stays default '*' |
| TestWorkflowsFromRawFilePatternOverride | - | file_pattern override *.yaml applied, dir stays default '.' |
| TestParseEnvUnknownKeyIgnored | - | unknown top-level key is silently ignored, no error, docker section parses normally |
| TestParseEnvWrongScalarTypeErrors | - | non-integer management.port errors, message contains 'cannot unmarshal' |
| TestParseEnvPortsValid | bare int 8090 | parses Host=8090 Container=8090 String()='8090:8090' |
| TestParseEnvPortsValid | host:container 8080:8090 | parses Host=8080 Container=8090 String()='8080:8090' |
| TestParseEnvPortsValid | padded host:container with spaces | trims spaces to Host=8080 Container=8090 String()='8080:8090' |
| TestParseEnvPortsInvalid | non-integer 'abc' | error 'env.yaml: ports entry "abc" must be an integer or "host:container"' |
| TestParseEnvPortsInvalid | more than one colon '1:2:3' | error 'env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)' |
| TestParseEnvPortsInvalid | non-integer host and container 'a:b' | error 'env.yaml: ports entry "a:b" must be "host:container" with integer ports' |
| TestParseEnvPortsInvalid | mapping node {a: 1} | error 'env.yaml: ports entry must be an integer or "host:container", got a !!map' |
| TestApplyDockerDefaultsFillsMissing | - | docker defaults command/name/project-name/restart/run-as applied (run-as the non-root default, since the image's own user is root), ports stay empty (publishing is opt-in), stores/libs stay nil |
| TestApplyDockerDefaultsOverrideWins | - | explicit command/name/project-name/restart/run-as/ports override defaults exactly as given; a custom name does not drag project-name with it |
| TestApplyPodmanDefaultsFillsMissing | - | podman defaults command/name/restart (systemd's always, not compose's unless-stopped)/run-as applied, ports stay empty (publishing is opt-in), and both rejected keys left alone: mode empty and Quadlet nil. Defaulting either would trip validate's rejection for something the operator never wrote, and nothing dereferences Quadlet -- the unit dir comes from the invoking uid |
| TestApplyPodmanDefaultsOverrideWins | - | explicit command/name/restart/run-as/ports override defaults exactly, and a present quadlet: block still decodes non-nil so validate can reject it by name |
| TestRemovedMountKeysDecodeButAreNotDefaulted | - | libs dir kept verbatim while mount-path is left empty (defaulting it would trip validate's rejection for a value nobody wrote), and a present stores: still decodes non-nil so validate can name it |
| TestPortDefaultsFollowManagementPort | - | management.port 9091 with docker/podman/kubernetes present: kubernetes Service.Port defaults to {9091,9091}; docker/podman publish nothing with ports: omitted |
| TestPortDefaultsFallBackWhenManagementPortUnset | - | no management.port set: kubernetes service.port falls back to {DefaultMgmtPort,DefaultMgmtPort} (8090); docker/podman still publish nothing |
| TestKubernetesServicePortAcceptsBareAndHostContainerForms | bare int 8090 | kubernetes service.port parses to Host=8090 Container=8090 |
| TestKubernetesServicePortAcceptsBareAndHostContainerForms | host:container 8080:8090 | kubernetes service.port parses to Host=8080 Container=8090 |
| TestKubernetesServicePortRejectsInvalidForms | multi-colon 1:2:3 | error 'env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)' |
| TestKubernetesServicePortRejectsInvalidForms | mapping node {a: 1} | error 'env.yaml: ports entry must be an integer or "host:container", got a !!map' |
| TestWrittenLibsMountPathSurvivesDecoding | - | a mount-path the operator wrote is preserved through decoding so validate can reject it by name, rather than being silently discarded |
| TestExpandBracedVar | - | `${HOST}` in Side.Host expands from Lookup |
| TestExpandDefaultVarSetUsesValue | - | `${VPN:fallback}` uses the looked-up value when VPN is set |
| TestExpandDefaultVarUnsetUsesDefault | - | `${VPN:fallback}` falls back to the default when VPN is unset |
| TestExpandUnsetNoDefaultPassesThroughWithWarning | - | unset defaultless `${TYPO}` passes through verbatim and Warn is called exactly once naming TYPO |
| TestExpandStorePathsNotPasswordsOrAliases | - | a store's paths expand -- existing.pkcs12, every certs file, the key entry's cert, chain and key, the jks -- while password, password-env and a certs alias stay as written |
| TestExpandBareDollarVarUntouched | - | bare `$VPN` (no braces) is left untouched even though VPN is set |
| TestExpandCredentialFieldLeftAlone | - | Side.Password/PasswordEnv (`expand:"no"`) never expand |
| TestExpandYAMLNodePassthroughLeftAlone | - | a `*yaml.Node` field (APIProps) is never walked or rewritten |
| TestExpandDefaultsConnectionsMapEntry | - | a `${HOST}` value inside Defaults.Connections (map[string]Side) expands via read-modify-write |
| TestExpandSecurityUserRole | - | a `${VAR}` in security.users[].roles expands (a role is an identity, not a credential), proving Expand reaches a []string inside a slice of structs under Defaults |
| TestParseEnvJavaOptionsSpellings | plain strings / folded block / literal block / only one sub-key | java-options.tool and .jdk parse from a plain string, a >- folded block and a literal block, and all three normalize to the same single line of options |
| TestParseEnvJavaOptionsAbsent | - | an env.yaml without java-options, or with the key left empty, sets no JVM options variable at all |
| TestParseEnvJavaOptionsShapeErrors | a bare string / a misspelled sub-key / a list of options / a mapping as a value / a sub-key given twice | each fails at parse naming the key and what it should be -- ParseEnv has no KnownFields, so without this they would be dropped in silence |
| TestParseEnvJavaOptionsAlias | - | a YAML alias to a scalar is followed rather than reported as the wrong shape |
| TestJavaEnvMergesTheTLSFlag | nothing set / TLS only / TLS with an empty block / operator only / TLS then operator / jdk never gets the TLS flag / whitespace-only values / normalized before merging | the MQ TLS flag comes first in JAVA_TOOL_OPTIONS with the operator's options after it (so theirs win), JDK_JAVA_OPTIONS never carries it, and an empty variable is not set |
| TestExpandJavaOptions | - | both sub-keys take ${VAR} and ${VAR:default} like every other non-credential value |
| TestExpandNilLookupDisablesEverything | - | nil Lookup makes Expand a no-op, leaving `${HOST}` untouched |
| TestLeaderElectionEffectiveMode | empty | an empty Mode defaults EffectiveMode to standalone |
| TestLeaderElectionEffectiveMode | standalone / active_active / active_standby | an explicit Mode passes through EffectiveMode unchanged |
| TestEffectiveHealthShowDetails | nil receiver / unset / set | EffectiveHealthShowDetails falls back to DefaultHealthShowDetails (always) for a nil Defaults and an unset key, and returns a configured value unchanged |
| TestEffectiveManagementPort | nil receiver / unset port / set port | EffectiveManagementPort falls back to DefaultMgmtPort (8090) for a nil Defaults and an unset port, and returns a configured port unchanged |
| TestCredEmptyBothKeyDescribe | unset / literal only / env only / both set | Cred.Empty/Both/Key/Describe resolve deterministically for every shape; both-set resolves to the env side (validate rejects it separately) rather than panicking |
| TestSideUsernameSecretBothSystems | - | Side.Username/Secret dispatch by System, not by whichever credential pair is non-empty: solace returns client-user/-pass, mq returns user/password |
| TestStoreSecretNilSafe | - | nil *Store yields an empty Cred; literal and -env stores yield the matching Cred side |
| TestUserSecretLiteralAndEnv | - | a security user's Secret() carries the literal or the -env variable, matching what was set |
| TestCredCreateRemovedKeys | - | RemovedKeys reports each of source, variables, values-file alone and all three in order; a nil receiver and a bare create.name report none |
| TestImageRef | hub / private registry / digest / trailing slash / no tag / no name / nil | Ref() assembles repo/name:tag, drops an unset repo, and joins a sha256: tag with `@` so a digest pin is a reference an engine accepts |
| TestImageRegistry | hub fallback / private / nil / trailing slash | the auths key is the registry host, falling back to Docker Hub's v1 URL -- a Hub namespace lives in name and never reaches the lookup |
| TestRetiredPerPlatformImageStillParses | - | kubernetes.deployment.image, docker.image and podman.image parse into their fields, which is what lets validate reject them instead of yaml dropping them silently |
| TestImagePullSecretCreateDefaultsFalse | absent / explicit true | create defaults to false so naming a Secret only references it; an explicit true is honoured |
| TestSecretsCreateSpellings | create: true / the retired mapping form / an empty mapping / create: false / create left empty / only existing | credentials and stores create: takes true, and the mapping an older env.yaml carries still means create (its name decoded only so validate can report it); false, empty and absent build nothing, and existing: survives the custom decoding |
| TestSecretsCreateRejectsOtherShapes | a word / a list / stores, a word / a retired mapping with a list for its name (credentials, stores) / the block as a bare value (credentials, stores) | a create: value that is neither a boolean nor the retired mapping fails at parse naming the key, a wrongly-typed field inside the retired mapping is still a parse error, and a secrets block must itself be a mapping |
| TestSecretsCreateFollowsAnAlias | - | a YAML alias (create: *on) is resolved before create: is read, so an aliased true creates the Secret |
| TestDerivedObjectNames | - | every created Secret and the libs claim are named after deployment.name (-credentials, -stores, -image-pull, -libs) with any retired name key ignored, a referenced object keeps the operator's own name, no block yields nothing, and the stores Secret's name is the instance's whatever the retired stores block says |
| TestCreatedNames | - | the Deployment, its Service and ConfigMap always, then exactly the Secrets and claim the config asks the tool to build, each with its kind (the list the namespace gate and validate share) -- a referenced object is not the tool's, and the cluster-scoped PersistentVolume is not listed |
| TestParseWorkflowTransformHeaders | - | a top-level transform-headers block is captured verbatim -- header order and each expression's quoting intact -- and is not reported as misplaced; an absent or empty block is none |
| TestParseWorkflowTransform | - | a top-level transform block is captured verbatim -- key order, the list of one-key mappings and each expression's quoting intact -- beside a top-level transform-headers block, and neither is reported as misplaced; an absent or empty block is none |
| TestMisplacedWorkflowTransforms | - | a transform-header: typo, the legacy transform-payload section (either spelling), and transform or transform-headers under a side, its solace:/mq: block (also when that block is an alias and the key reaches it through a << merge), or that block's consumer:/producer: are each reported as the dotted path they were found at, in file order, while both top-level blocks still parse |
| TestMisplacedEnvTransforms | - | every transform key in env.yaml (transform included) -- top level, a connection, or the connection's solace:/mq: block -- is reported, identically through ParseEnv and ParseDefaults, and a clean or empty env.yaml reports none |
| TestParseEnvTopLevelSyslog | present / absent | logging.syslog parses beside logging.level at the top level, protocol defaults to udp, and an absent block stays nil (presence is what turns syslog on) |
| TestStoreConfigured | nil | a nil store is not configured |
| TestStoreConfigured | no source | a block with only a password is not configured |
| TestStoreConfigured | existing without a path | an existing block naming no PKCS12 is not configured |
| TestStoreConfigured | existing | an existing PKCS12 is configured, and Path is it |
| TestStoreConfigured | created | a created store is configured, Created, and has no Path |
| TestStoreConfigured | secret only | a store naming only a kubernetes Secret is configured, with no Path |
| TestStoreConfigured | file and secret | a PKCS12 beside a Secret is configured, and Path is the file |
| TestStoreConfigured | created with a secret | sources beside a Secret are configured and Created |
| TestStoreSecretParses | - | existing.secret's name and key decode onto the store, the key defaults to the store's name, a file beside it is kept, and SecretRef / EffectiveKey are nil-safe |
| TestTLSConfigStoreLookup | - | an empty or undefined store name is no store, and StoreNames lists the defined names sorted |
| TestParseTLSIsStrict | retired truststore | tls.truststore is an error naming tls.stores and the per-connection selectors |
| TestParseTLSIsStrict | retired keystore | tls.keystore is the same error |
| TestParseTLSIsStrict | unknown key | any other key under tls: is an error |
| TestParseTLSIsStrict | retired file | a store's file: is an error naming existing: {pkcs12: <file>}, existing: {secret: ...} and create: |
| TestParseTLSIsStrict | retired type | so is its type: |
| TestParseTLSIsStrict | unknown store key | a key a store does not take (passwrd-env) is an error naming tls.stores.<name>.<key> |
| TestParseTLSIsStrict | existing not a mapping / unknown existing key | existing: must be a mapping holding only pkcs12 and secret |
| TestParseTLSIsStrict | secret a scalar / unknown secret key | existing.secret must be a mapping holding only name and key |
| TestParseTLSIsStrict | create not a mapping / unknown create key | create: must be a mapping holding only certs, keys and jks |
| TestParseTLSIsStrict | certs a scalar | certs must be a list or a map |
| TestParseTLSIsStrict | keys a mapping / key entry a scalar / unknown key-entry key / chain a scalar | keys is a list of mappings holding only cert, chain and key, and chain is a list |
| TestParseTLSIsStrict | not a mapping / stores not a mapping | tls: and tls.stores as lists are errors; a valid store and one written with no block both parse |
| TestParseCreateBlock | - | certs decode as a list (in order, unnamed) or a map (in written order, named), one key entry with its chain, a jks; a create: with nothing in it is no create block |
| TestStoreKeyListsAreSorted | - | the store, create, key-entry, existing and existing.secret key lists come back sorted, for the generator page to be checked against |

## internal/consolidate

Build the consolidated binder model from the workflows -- dedup connections, TLS bundles, destination roles, store-path rewriting, leader election, and durable-name UUIDs.

Tests: [consolidate_test.go](../internal/consolidate/consolidate_test.go), [consolidate_extra_test.go](../internal/consolidate/consolidate_extra_test.go), [consolidate_extrakeys_test.go](../internal/consolidate/consolidate_extrakeys_test.go), [names_test.go](../internal/consolidate/names_test.go), [uuid_test.go](../internal/consolidate/uuid_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestBinderDedupAcrossWorkflows | two workflows, inline sides | binders = [sol-conn-1, mq-conn-1, undefined], numbered per system |
| TestSolaceToSolaceSingleBinder | solace->solace | binders = [sol-conn-1, undefined]; input-0 and output-0 both reference sol-conn-1 |
| TestMQToMQSingleBinder | mq->mq | binders = [mq-conn-1, undefined]; input-0 and output-0 both reference mq-conn-1 |
| TestConnRefNamingAndClashSuffix | svc.a vs svc/a conn-refs sanitize to same base | binders = [svc-a, svc-a-2, undefined], second disambiguated |
| TestConnRefDedupCollapsesToOneBinder | two workflows share edge/qm conn-refs | binders = [edge, qm, undefined], collapse to one binder each |
| TestDerivedDestinationTypes | mq topic consumer durable -> solace topic producer | input-0 jms is consumer/topic with non-empty Durable |
| TestDerivedDestinationTypes | solace producer to topic | output-0 emits no solace binding |
| TestDerivedDestinationTypes | solace queue -> mq queue producer | output-1 jms is producer/queue |
| TestBuildCarriesEachWorkflowsTransformBlocks | - | a workflow's transform and transform-headers blocks reach its own WorkflowEnable entry untouched and never a neighbour's, and a workflow with neither carries neither |
| TestFormatScalarQuoting | plain passthrough | FormatScalar returns plain |
| TestFormatScalarQuoting | double-quoted requoting | FormatScalar returns "dq" |
| TestFormatScalarQuoting | single-quote-doubling escape | FormatScalar returns 'a''b' |
| TestFormatScalarQuoting | depth>=2 nested quoted passthrough from parsed YAML | FormatScalar preserves quoting, returns "R1" |
| TestFormatScalarQuoting | one-line folded block, safe plain | a one-line block scalar that reads back the same as plain text stays plain |
| TestFormatScalarQuoting | one-line folded block holding a comment marker | a one-line block scalar is written on the key's line, so one holding " #" is double-quoted rather than cut off there |
| TestFormatScalarQuoting | multi-line literal block left to the render layer | a value spanning lines is returned unchanged for render.writeScalar to re-emit as a block |
| TestSanitizeAndIsTCPS | sanitize A_b.2-x/y | returns A-b-2-x-y |
| TestSanitizeAndIsTCPS | isTCPS tcps:// and tcp:// | tcps:// true, tcp:// false |
| TestDisplayName | acc with name set | returns the name |
| TestDisplayName | acc with only connName | falls back to connection name |
| TestDisplayName | acc solace kind with vpn | returns solace:myvpn |
| TestDisplayName | acc mq kind with qm | returns mq:QM1 |
| TestMergeProp | append new key b | list grows to 2 entries, no warnings |
| TestMergeProp | overwrite existing key a with new value | value updated to 9, one warning emitted |
| TestMergeProp | overwrite key a with same value again | no additional warning |
| TestMergePropNestedValues | - | the same nested mapping arriving again is not a change and neither is the same node, a different one warns once and the last writer wins |
| TestSameNode | nil and nil / nil and a node / same value / quoting differs / same mapping / different length / different kind | the structural compare behind mergeProp: nil only equals nil, a scalar by its formatted value so quoting counts, a container element by element |
| TestOverlayExtras | replaces in place / another spelling replaces too / a new key is appended / no extras / nil over nil / extras over no defaults | a connection's other keys lay over a defaults block by Spring's reading of the name, replacing in place or appending, and nothing to lay on returns the defaults untouched |
| TestBuildCarriesExtraKeysOntoBothBinders | - | an inline side's other keys reach its own binder -- an MQ one as Extras beside additional-properties, a Solace one laid over solace-defaults with the connection winning in place -- with no warning |
| TestBuildMergesExtraKeysAcrossSidesOfOneBinder | a union / a disagreement | two sides on one tuple merge their other keys into the one binder: a union without a warning, last (by filename) wins a disagreement with the passthrough warning |
| TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet | - | a connection carrying a nested block, reused by three workflows through conn-ref, merges it once per side without a set-more-than-once warning |
| TestBuildDropsExtraKeysTheToolManages | - | the consolidate backstop: ssl-bundle and another spelling of queue-manager or msg-vpn are dropped with the passthrough warning and the tool's value stays |
| TestBuildLeaderElectionSessionCarriesExtraKeys | - | the session's connection's other keys lay over solace-defaults as a binder's do, and a managed-key spelling is dropped under the session's own label |
| TestBuildMQDefaultsOverriddenByConnection | - | mq-defaults reach every MQ binder, a connection's own key replaces a default in place, and a managed key in the block is dropped once rather than once per binder |
| TestAppendPassthroughCollision | passthrough T collides with existing T | output has 2 props and 1 collision warning reading `binder "bndr": passthrough overrides tool-managed key "T"; tool value kept` byte for byte, since the owner label is caller-supplied now |
| TestNodeToProps | mapping with scalar and nested keys | 2 props, first key k1/val v1, second has Sub set |
| TestNodeToProps | nil node | nodeToProps returns nil |
| TestBuildMQmTLSBundle | mq TLS side selecting mq-ca and mq-client with keyAlias plus a solace target selecting mq-ca | MQTLS true, 1 bundle with truststore and keystore, KeyAlias mc, both typed PKCS12; Stores lists mq-ca (Trusted) then mq-client (Private); the shared mq-ca password is one secret |
| TestBuildCipherConflictWarning | two mq sources with different ciphers C1/C2 | warnings contain conflicting cipher |
| TestBuildMessageLoopWarning | same side used as source and target dest `SAME` | warnings contain message loop |
| TestBuildSolaceTopicSourceEmitsConsumerTopic | solace topic source -> mq queue target | input-0 solace binding is consumer with DestType topic; consolidate renders whatever it is given, so this stays covered even though validate now rejects the combination |
| TestBuildStorePathsAreMountPaths | - | TruststoreLoc and KeystoreLoc are /app/external/classpath/truststores/mq-ca and .../mq-client, the selected stores' names, never the PKCS12 paths env.yaml gives |
| TestBuildLeaderElection | dangling conn-ref missing | le non-nil with Mode/Queue set but Session nil, no guard warning |
| TestBuildLeaderElection | conn-ref happy path tcps host | Session host/vpn set and APIProps has SSL_TRUST_STORE, SSL_KEY_STORE, SSL_PRIVATE_KEY_ALIAS mounted paths and alias sc; Extras carry solace-defaults in authored order and the connection api-properties land last, after the tool TLS keys |
| TestBuildLeaderElection | inline session non-tcps host | Session set from inline fields with no TLS APIProps, but the inline api-properties and solace-defaults still come through |
| TestBuildLeaderElection | leader election truststore | the session's store is at /app/external/classpath/truststores/edge-ca, not its env.yaml path |
| TestBuildLeaderElectionSessionPassthroughCollision | - | a session passthrough key colliding with a tool TLS key keeps the tool value and warns as `leader-election session`, never as a binder |
| TestBuildLeaderElectionWarningsReachBuild | - | the session collision warning escapes Build, proving the warns slice is threaded into buildLeaderElection |
| TestBuildLeaderElectionSharesBinderSecretNames | shared connection | a session resolving to a workflow binder carries that binder CLIENT_USERNAME/PASSWORD and mounts no second secret |
| TestBuildLeaderElectionSharesBinderSecretNames | management-only broker | a session no workflow binds falls back to the fixed LEADER_ELECTION_* pair |
| TestApplyStatusAccessNoOperatorUsers | - | with no configured security.users, Build synthesizes the reserved account as the only user, carrying the literal status password, never entering Model.Secrets |
| TestApplyStatusAccessAppendsAfterExistingUsers | - | operator-configured users get the reserved account appended last; existing users still resolve to secretRef placeholders |
| TestApplyStatusAccessCarriesOperatorRoles | - | an operator's roles reach the model verbatim (only the password is rewritten), the reserved account is appended with none so it stays read-only, and the caller's own Defaults are left unmutated despite sharing the roles backing array |
| TestApplyStatusAccessExposureIsFixed | - | applyStatusAccess always sets Management.Exposure to health,info,metrics,leaderelection,workflows, ignoring whatever spec.Management.Exposure carries |
| TestDurableNameGolden | - | DurableName of fixed inputs equals pinned solmq-3631c883-c0c4-5bc8-985e-ea2842831ad6 |
| TestDurableNameDeterministic | same inputs called twice | DurableName returns identical value both times |
| TestDurableNameDeterministic | different file name g.yaml vs f.yaml | DurableName differs when file name changes |
| TestGeneratedSecretNamesStayOutOfChildEnvDanger | literal credentials | every derived SecretRef.Stable Build() can produce (binder creds, security-user passwords, TLS stores, leader-election) carries spec.GeneratedNamePrefix and matches a fixed-suffix pattern, so adversarial conn-ref/security-user names (e.g. "path", "ld-preload", "LD") can never fold to a bare dangerous docker-compose child-env name like PATH or LD_PRELOAD |
| TestGeneratedSecretNamesStayOutOfChildEnvDanger | -env credentials | Stable equals EnvVar, so the pair injected into the compose child only restates the variable it was read from -- an operator-chosen name like PATH overwrites it with its own value |
| TestGeneratedSecretNamesStayOutOfChildEnvDanger | fixture coverage | the fixture yields at least one of each naming path, so neither half of the guarantee can silently stop being exercised |
| TestStableTokenFolding | mq-conn-1 / svc.a / punctuation runs / leading digit / empty / underscore runs | stableToken folds to upper-snake, collapses non-alphanumeric runs to one `_`, trims edge `_`, prefixes a leading digit with X, and returns X for an unfoldable input |
| TestBinderFieldsCarryStablePlaceholders | - | no credential value reaches a rendered binder field -- only ${NAME} placeholders -- and Model.Secrets records the real literal/env source under each name; an -env credential is keyed by its own variable name, a literal by the derived stableName for its position |
| TestEnvCredentialsShareOneMountName | two binders, one -env variable | two positions naming the same host variable dedup to a single SecretRef and record no conflict |
| TestSecretNameConflictIsRecorded | security.users "ops.1" and "ops-1" | two derived names folding to one via stableToken -- the collision that survives the reserved prefix; Model.SecretConflicts records the key plus both positions, and the first credential still wins the name so the model stays deterministic |
| TestBuildSolaceBinderPassesValidateCertificateThrough | - | a conn-ref Solace binder and its leader-election session, selecting the store MQ uses, carry the store wiring then the connection's own SSL_VALIDATE_CERTIFICATE: false, with no collision warning; the MQ bundle is unchanged and the shared store's password is one secret |
| TestBuildSeparateStoresPerConnection | - | MQ and Solace each wired to the stores they select (Solace trust + client store, MQ trust only, the session its own pair); Model.Stores lists every selected store once in first-use order, Trusted when used as a truststore and Private when used as a keystore anywhere -- mq-ca is both; each store password is its own secret; an MQ binder selecting nothing gets no bundle and no warning; an unselected store is never referenced |
| TestBuildMQKeystoreOnlyBundle | - | an MQ binder selecting only a keystore gets a bundle with the client certificate and alias and no truststore section, and the store is listed Private |
| TestBuildCreatedStores | - | a created store is listed Created and Trusted with no File, and wired at its mount path like every store |

## internal/tls

Resolve TLS store paths -- the host path as written vs the in-container path under each store's name -- and the Solace api-property keys of the stores a connection selects.

Tests: [tls_test.go](../internal/tls/tls_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestMountPathIsPerStore | mq-ca | MountDir/mq-ca, whatever the file is called |
| TestMountPathIsPerStore | solace-client | MountDir/solace-client, so two stores never share a path |
| TestSolacePropsUseStoreMountPaths | - | sets SSL_TRUST_STORE and SSL_KEY_STORE to MountDir/solace-ca and MountDir/solace-client, the selected stores' names, though both files are named store.p12 and written with backslashes |
| TestSolacePropsWireACreatedStoreAtItsMountPath | - | a created store, having no file, is wired at MountDir/<name> with PKCS12 and its password placeholder |
| TestSolacePropsStorePasswordIsPlaceholderNeverLiteral | - | store passwords reach api-properties only as the placeholder secretRef returns; secretRef gets each selected store's own name and credential, and no literal or env-var name leaks into any value |
| TestSolacePropsSkipsSecretRefWhenNothingSelected | - | no selection, or names nothing defines, emits nothing and never calls secretRef |
| TestSolacePropsWritesNoValidationPolicy | none | no stores: no keys |
| TestSolacePropsWritesNoValidationPolicy | truststore | only SSL_TRUST_STORE, _PASSWORD, _FORMAT -- never SSL_VALIDATE_CERTIFICATE |
| TestSolacePropsWritesNoValidationPolicy | keystore without alias | a keystore alone presents its certificate: SSL_KEY_STORE, _PASSWORD, _FORMAT, no alias |
| TestSolacePropsWritesNoValidationPolicy | keystore with alias | SSL_KEY_STORE, _PASSWORD, _FORMAT, SSL_PRIVATE_KEY_ALIAS |
| TestSolacePropsWritesNoValidationPolicy | both with alias | truststore keys then keystore keys, no validation policy |
| TestSolacePropsAlwaysWritePKCS12 | - | both FORMAT keys are always PKCS12 -- every store is one |
| TestSolacePropsSkipsUnconfiguredStore | - | a store with no source, or an existing block without a path, is not wired and its password is never recorded |

## internal/storebuild

Build created PKCS12 stores in memory -- a truststore from certificate files or a JKS's trusted entries, a keystore from one key and its chain -- deterministically, marked for Java, and without secrets in any message.

Tests: [storebuild_test.go](../internal/storebuild/storebuild_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestTruststoreFromCertsList | - | a certs list numbers every certificate of a PEM file and a DER file ca-1..ca-3, warns about nothing, and DecodeTrustStore -- which accepts only certificates carrying Java's trusted-cert attribute -- returns the three in order |
| TestTruststoreNamedAliases | - | a named map aliases each certificate by its name, numbers a second certificate in the same file (roots, roots-2), and reports the alias Java will list, lower-cased (issuing) |
| TestTruststoreAliasCollision | - | two names differing only in case are one Java alias: an error naming both files, not a silently lost certificate |
| TestTruststoreWarnings | a self-signed leaf is pinned | warns that it is trusted as-is, so only that exact certificate is accepted |
| TestTruststoreWarnings | expired | warns that it expired, and still builds |
| TestTruststoreWarnings | expiring | warns that it expires within 30 days |
| TestTruststoreWarnings | repeated | the same certificate twice is kept once, with a warning |
| TestTruststoreWarnings | a key beside the certificates | a private key in a certs file is left out, with a warning |
| TestTruststoreRejects | not a certificate | error naming tls.stores.ca.create.certs[0] (./junk.txt): no certificate |
| TestTruststoreRejects | only a key | a file holding only a private key has no certificate |
| TestTruststoreRejects | a broken certificate | certificate 1 does not parse |
| TestTruststoreRejects | no source | no certificate to trust |
| TestKeystoreFromPEMChain | - | the leaf is found by its key among the cert file's certificates, the chain is ordered issuing then root across files, the one key entry is reported as alias 1, and DecodeChain returns the leaf, [issuing, root] and the key |
| TestKeystoreKeyEncodings | PKCS#8 PEM | builds |
| TestKeystoreKeyEncodings | SEC1 PEM | builds |
| TestKeystoreKeyEncodings | PKCS#1 PEM | builds (RSA) |
| TestKeystoreKeyEncodings | PKCS#8 DER | builds |
| TestKeystoreKeyEncodings | PKCS#1 DER | builds (RSA), after PKCS#8 fails to parse it |
| TestKeystoreKeyEncodings | SEC1 DER | builds, after PKCS#8 and PKCS#1 fail to parse it |
| TestKeystoreKeyEncodings | one file with the certificate and the key | the same file serves as cert and key |
| TestKeystoreRejects | encrypted PKCS#8 | error: an encrypted private key cannot be read, with the decrypt remedy |
| TestKeystoreRejects | legacy encrypted PEM | a Proc-Type: 4,ENCRYPTED block gets the same error |
| TestKeystoreRejects | two keys | more than one private key |
| TestKeystoreRejects | no key | a key file holding only certificates has no private key |
| TestKeystoreRejects | a broken key | the private key does not parse |
| TestKeystoreRejects | a key of another certificate | no certificate matches the key, naming the cert and key positions and files |
| TestKeystoreRejects | an expired certificate | an expired leaf is an error: the peer refuses it |
| TestKeystoreRejects | an X25519 key | a key that cannot sign is not supported, naming its type |
| TestKeystoreWarnings | - | a certificate outside the leaf's chain is left out with a warning, a not-yet-valid leaf is reported but built, and the chain holds only issuing and root |
| TestJKSTrustedEntriesBecomeATruststore | - | a JKS of trusted certificates becomes a truststore keeping its aliases, in alias order |
| TestJKSSingleKeyBecomesAKeystore | - | a JKS holding exactly one key becomes a keystore with that key's leaf and two-certificate chain |
| TestJKSRejects | trusted and a key | error listing every alias by kind: a created store is a truststore or one key |
| TestJKSRejects | two keys | two keys are refused the same way |
| TestJKSRejects | empty | a JKS with no entries |
| TestJKSRejects | another password | the JKS integrity check fails: the store's own password must open it |
| TestJKSRejects | a key with its own password | the key does not open with the store password, with the keytool -keypasswd remedy |
| TestJKSRejects | JCEKS | JCEKS magic: not readable here, with the keytool conversion |
| TestJKSRejects | already PKCS12 | PKCS12 bytes under create.jks: name the file with existing: {pkcs12: <file>} |
| TestJKSRejects | not a keystore | anything else is not a JKS |
| TestJKSRejects | truncated | a JKS cut short is not readable |
| TestJKSRejects | a key that is not its certificate's | a key entry whose key does not match its chain's first certificate |
| TestJKSRejects | a key without a certificate | a key entry with no certificate |
| TestJKSRejects | a non-X.509 entry | an entry of another certificate type is named, with its type |
| TestJKSRejects | a damaged certificate | an entry whose certificate does not parse is named |
| TestCertsWithJKS | - | certs plus a JKS's trusted entries make one truststore (ca-1, legacy); a JKS holding a key cannot join a truststore, nor one the store password does not open |
| TestKeystoreRejectsBrokenFiles | the cert file | a cert file holding no certificate is named by its own position |
| TestKeystoreRejectsBrokenFiles | a chain file | so is a chain file |
| TestKeystoreLeafWithoutItsIssuer | - | a leaf given without its issuer builds with an empty chain -- the peer may already hold the intermediate |
| TestPasswordOutsideBMP | truststore | a password outside the Basic Multilingual Plane cannot be encoded (UCS-2): an error, not an unreadable file |
| TestPasswordOutsideBMP | keystore | the same for a keystore |
| TestReadBackError | no decoder error | a store that does not read back as built is reported as a bug |
| TestReadBackError | a decoder error | the decoder's error is carried in the report |
| TestBuildIsDeterministic | - | the same inputs and password encode to the same bytes, moving a source file changes nothing, and another password or another certificate changes the bytes |
| TestBuildNeedsPasswordAndBoundedSources | - | an empty password is an error, and so is a source over 1 MiB, naming its position and size |
| TestGuardTurnsPanicIntoError | - | a library panic becomes an error naming the store and the panic's type, never its value |
| TestErrorsCarryNoSecrets | - | a key mismatch, a wrong JKS password and a key-only certs file fail without the password or the key's bytes in the message |
| TestSubjectIsPrintable | - | control characters in a certificate subject are replaced before printing |
| TestSniff | JKS / JCEKS / PKCS12 | FEEDFEED, CECECECE and a DER SEQUENCE are told apart |
| TestSniff | PEM text | text is unknown |
| TestSniff | empty | no bytes are unknown |
| TestJavaReadsBuiltStores | truststore | keytool lists the named entry as a trustedCertEntry (skipped when keytool is not on PATH) |
| TestJavaReadsBuiltStores | keystore | keytool lists one PrivateKeyEntry aliased 1 -- the fact behind the key-alias rule for created keystores |

## internal/yamlwriter

The indentation-aware line writer every generated artifact is built from, keeping indentation consistent across every artifact it renders.

Tests: [yamlwriter_test.go](../internal/yamlwriter/yamlwriter_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestWriterLineIndent | - | Line indents by the given level across several nesting depths |
| TestWriterRawPassthrough | - | Raw writes pre-formatted text between Line calls without indenting it |
| TestSplitLines | trailing newline | a terminating newline does not yield a trailing empty element |
| TestSplitLines | no trailing newline | the final line is kept |
| TestSplitLines | empty string | yields no lines |
| TestSplitLines | blank line inside | interior blank lines are preserved |

## internal/render

Render application.yml from the consolidated model via the deterministic ordered emitter.

Tests: [render_test.go](../internal/render/render_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestApplicationRich | - | rendered output contains all expected substrings for rich MQ->Solace workflow (ssl bundle, binders, jms/solace bindings, defaults blocks) |
| TestApplicationRichExact | - | generated application.yml matches richApplicationWant golden fixture byte-for-byte |
| TestRenderManagementDeclaresTheStatusHealthGroup | - | the management block declares the `solmq` health group: every contributor, show-details as the operator set it (`when-authorized` here), and DOWN and OUT_OF_SERVICE mapped to 200; it is the only http-mapping, so `/actuator/health` keeps its 503 |
| TestApplicationRendersWorkflowTransforms | - | each workflow's transform renders verbatim right after enabled -- quoting as written, each list item a bare dash over its indented mapping, a folded expression on one quoted line so its " #" survives -- transform-headers under its own workflow, an empty transform renders nothing, and the document reads back as the connector's expressions list |
| TestApplicationMinimalNoOptionalBlocks | ssl: / logging: | ssl: and logging: blocks stay absent when defaults are empty |
| TestApplicationMinimalNoOptionalBlocks | management: / security: | management: and security: are unconditional now: the fixed exposure list, show-details: always and the reserved solmq-status account render even with empty defaults |
| TestApplicationMinimalNoOptionalBlocks | type: undefined | undefined binder is always emitted even with minimal config |
| TestApplicationLeaderElection | - | leader-election, fail-over, queue and management render for active_standby, and the whole session block matches exactly: binder-shared credential names, solace-defaults between the credentials and api-properties, verbatim passthrough last |
| TestApplicationLeaderElectionSessionMatchesBinderKeySet | - | rendered from one connection, the session key sequence equals the binder solace.java key sequence -- the guard against the two renderers drifting apart again |
| TestApplicationLeaderElectionSessionPlaintext | - | a non-tcps session emits no SSL_ keys but still carries solace-defaults and its own api-properties, and falls back to the LEADER_ELECTION_* names |
| TestApplicationRendersExtraKeys | mq block / solace block / read-back | an MQ binder's other keys render after ssl-bundle and before additional-properties, mq-defaults first and the connection's own after, a nested block whole; a Solace binder's where solace-defaults render, the connection's value in place of the default it overrides; the document reads back with the key as a sibling of queue-manager, never inside additional-properties |
| TestApplicationSessionCarriesExtraKeysLikeTheBinder | - | the anti-drift guard extended to other keys: a session built from the same connection as a binder renders the same key set, client-name included |
| TestApplicationEmitsExactlyTheToolManagedKeys | mq / solace | the keys render writes at the top of a binder's block are exactly spec.ToolManagedKeys, so a key added to one but not the other fails |
| TestApplicationOmitsEmptyCredentials | conn-name: h(1414) | binder identity fields still rendered when credentials absent |
| TestApplicationOmitsEmptyCredentials | queue-manager: QM1 | binder identity fields still rendered when credentials absent |
| TestApplicationOmitsEmptyCredentials | host: tcp://b:55555 | binder identity fields still rendered when credentials absent |
| TestApplicationOmitsEmptyCredentials | msg-vpn: v | binder identity fields still rendered when credentials absent |
| TestApplicationOmitsEmptyCredentials | user: | empty user credential line omitted, no null value emitted |
| TestApplicationOmitsEmptyCredentials | password: | empty password credential line omitted, no null value emitted |
| TestApplicationOmitsEmptyCredentials | client-username: | empty client-username credential line omitted, no null value emitted |
| TestApplicationOmitsEmptyCredentials | client-password: | empty client-password credential line omitted, no null value emitted |
| TestApplicationQuotesRiskyScalars | password: "p@ss #1" | a value containing " #" is double-quoted so it is not read as a comment |
| TestApplicationQuotesRiskyScalars | client-username: "no" | a bool-lookalike is double-quoted so it stays a string |
| TestApplicationQuotesRiskyScalars | client-password: "key: value" | a value containing ": " is double-quoted so it does not open a nested mapping |
| TestApplicationQuotesRiskyScalars | plain values | ordinary hosts, conn-names, users and destinations stay unquoted |
| TestApplicationBlockScalarPassthrough | - | a literal (\|) passthrough value is re-emitted as an indented block scalar, never flattened onto the key line |
| TestApplicationSkipsBundleWithoutStores | - | tls: true selecting no store emits no ssl bundle or ssl-bundle reference, keeps MQTLS set, and warns about nothing: JVM default trust is a legitimate choice |
| TestApplicationMQBinderLoadsSSLAutoConfiguration | - | every MQ binder environment, TLS or not, starts with spring.main.sources: SslAutoConfiguration, so ibm.mq.ssl-bundle resolves in the binder's own context; a Solace binder gets none |
| TestApplicationKeystoreOnlyBundle | - | a binder selecting only a keystore renders a bundle with keystore and key.alias and no truststore section |
| TestApplicationConfigImport | ConfigImport set / empty | Application() leads with spring.config.import when Model.ConfigImport is set, and omits the block entirely when it is empty |
| TestApplicationSecurityUserRoles | - | a roles-bearing user renders a block-style roles sequence under its password; a role-less user and the reserved solmq-status account emit no roles key at all, keeping pre-roles output byte-identical |
| TestApplicationWritesPKCS12Everywhere | - | the MQ bundle's truststore and keystore are both typed PKCS12 and both Solace FORMAT keys say PKCS12, with locations, alias and SSL_TRUST_STORE rendered |

## internal/logback

Render the logback-spring.xml the connector reads for syslog output, and the
in-container path every platform mounts it at.

Tests: [logback_test.go](../internal/logback/logback_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestXMLPicksTheAppenderTheProtocolNames | udp / tcp | udp uses Logback's built-in SyslogAppender and never references the logstash one; tcp uses LogstashTcpSocketAppender and never falls back. The two are not interchangeable: tcp needs a jar udp does not |
| TestXMLDefaultsToUDP | empty / unknown | an unset or unrecognised protocol renders the udp config, the safe choice because it needs nothing on the classpath; validate rejects an unknown value long before this |
| TestBothConfigsReadTheSameThreeProperties | udp / tcp | both bind host, port and appname via springProperty, which is why all three platforms set the same LOGGING_SYSLOG_* env vars rather than templating values into the file, and both keep the console appender (syslog is in addition to stdout, not instead of it) |
| TestContainerPathAndFileNameAgree | - | ContainerPath ends in /FileName, so a platform that writes the file to disk and mounts it cannot land it where the connector does not read |

## internal/statusscript

Render the POSIX status script the generated deploy artifacts embed and `solmq-conn-util status` execs inside each running instance -- a pure renderer with no os/exec, filesystem, or network access. Its tests read the rendered text, except TestHealthParseIsKeyOrderIndependent, which runs the script's own health parse under sh.

Tests: [statusscript_test.go](../internal/statusscript/statusscript_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderSubstitution | defaults | Render substitutes PORT=8090 and USER_NAME=solmq-status; BASE and the leaderelection/workflows endpoints are built from $PORT at run time, not by Render |
| TestRenderSubstitution | non-default port and user | Render substitutes PORT=19090 and USER_NAME=custom-mgmt-user |
| TestRenderIsPureASCIINoCRLF | - | output has no carriage return, no byte over 127, and ends with a trailing newline |
| TestRenderHeaderHasExecOneLiners | - | the header pins the kubectl/docker/podman exec one-liners, the `--health` invocation, each built from ContainerPath, and the `--details` argument a hand run needs for the enrichment lines |
| TestRenderPasswordResolution | - | the password-lookup chain references ContainerPath, SecretsDir and the from_configs account lookup, and no credential is embedded |
| TestRenderAlwaysExitsZero | - | every exit in the report path is `exit 0` and the script's only `exit 1` is the healthcheck verdict, `set -e` is absent while `set -u` stays, the EXIT trap holds the contract, and active/standby are both quiet outcomes |
| TestRenderHealthModeShortCircuits | - | `--health` sits below get(), the password it needs and the health_status it reads its verdict with, but above the exposure check that would exit 0, and makes exactly one actuator call -- no metrics, no /actuator/info, no JVM spawn, so it is cheap enough to run on a timer |
| TestRenderHealthModeExitsOnVerdict | - | `--health` clears the EXIT trap before answering (or its non-zero status would be swallowed), exits 0 only on UP, and echoes the verdict to stdout without a `status:` prefix so the engine's health log carries it |
| TestRenderSendsStatusToStdoutAndProblemsToStderr | - | the mode/state/health/workflow report lines go to stdout unredirected, and every `status:` diagnostic ends in `>&2` |
| TestRenderAlignsWorkflowColumn | - | the workflows block is a bare header plus one indented row per workflow, with the ids right-aligned to the widest id present so every colon sits in the same column |
| TestRenderReportsHealthUptimeAndVersion | endpoints | health, /actuator/metrics/process.uptime and /actuator/info are each read and rendered as their own report line |
| TestRenderReportsHealthUptimeAndVersion | dropped when silent | each enrichment line sits behind a non-empty guard, so an endpoint that answers nothing drops its line instead of printing an empty value |
| TestRenderReportsHealthUptimeAndVersion | own status | the health line reads the instance's status through health_status, the function the healthcheck answers with, so the two cannot disagree |
| TestRenderReportsEveryWorkflowInNumericOrder | ordering | each workflow line carries the id as a leading tab-separated sort key, goes through `sort -n`, and has the key cut off, so the report reads 1..9..10..19 instead of the actuator's map order |
| TestRenderReportsEveryWorkflowInNumericOrder | completeness | the workflows response is fed to the read loop with a terminating newline, and the bare `printf %s "$WF"` form is absent -- without it `read` skips the unterminated final line and the last workflow is dropped from every report |
| TestRenderReportsOnlyConfiguredWorkflows | filter | a chunk is reported only when it carries an id and a state, so nested JSON fragments with an id of their own stay out, and the `${st:-unknown}` padding is gone |
| TestRenderReportsOnlyConfiguredWorkflows | N/A slots | a state of `N/A` is dropped case-insensitively -- the connector marks every unconfigured slot that way, turning one real workflow into twenty report lines |
| TestRenderReportsOnlyConfiguredWorkflows | no allowlist | real states are not matched against a fixed list, so a state the script has never seen still reaches the operator |
| TestRenderReportsOnlyConfiguredWorkflows | nothing left | filtering every entry away is reported on stderr rather than leaving the workflow half of the report silently blank -- on an active instance only |
| TestRenderWarnsOnEmptyWorkflowsOnlyWhenActive | gated | the empty-workflow warning sits behind `elif [ "$STATE" = "active" ]`, so a standby (which runs no workflow) reports nothing rather than looking broken -- at `replicas: 2` that is half of every report |
| TestRenderWarnsOnEmptyWorkflowsOnlyWhenActive | not ungated | the bare `else` form that warned on every standby cannot come back |
| TestRenderVerifiesExposure | - | the has_entry membership check and the leaderelection/workflows exposure gate run before the first actuator request; an unexposed leaderelection stops the run on stderr (still exit 0), an unlocatable config only warns |
| TestRenderSearchesSpringConfigLocations | - | the config search covers SPRING_CONFIG_LOCATION, SPRING_CONFIG_ADDITIONAL_LOCATION and SPRING_CONFIG_NAME, ConfigDir and its wildcard form, the ./ and ./config/ defaults, both YAML extensions, comma splitting with optional:/file: stripping, the classpath: skip, and runs before the exposure check and password lookup |
| TestRenderEscapesUserForSedAddress | 7 names | USER_MATCH is regex-escaped for the sed address (dot, slash, brackets, star, backslash, anchors) while USER_NAME stays raw for the Authorization header |
| TestFilenameAndPathConstants | - | the script's name and directory, that ContainerPath is not nested inside the libs, spring/config or classpath mounts -- the nesting that made the libs mount shadow it -- and the healthcheck contract the three renderers share (HealthArg, HealthShell, and a cadence whose timeout is below the interval and whose start period is above it) |
| TestRenderReportsHealthComponents | - | the per-component health breakdown is read by health_components from the document the health line already fetched -- defined first, parsed after the fetch -- and the block prints only when something parsed |
| TestRenderReportsJavaConfigAndHeap | - | the three details-level lines from outside the report endpoints: `java -version` (stderr redirected, run with JAVA_TOOL_OPTIONS/JDK_JAVA_OPTIONS/_JAVA_OPTIONS unset so the JVM's "Picked up ..." notice is not reported as the version, folded to "openjdk 17.0.9" or passed through raw), the config the report was read from, and heap used/max tagged `area:heap`; each guarded so an absent source drops its line, a negative maximum is left out, and the byte arithmetic is deliberately *not done* here (busybox would read Jackson's 4.32013312E8 as 4) |
| TestRenderGatesDetailsCallsBehindTheArg | - | DETAILS starts empty and is set from `--details` by a loop over every argument after the `--health` branch; the metrics, info and `java -version` calls sit only inside its gate, while leaderelection, health, workflows, the config line and the health components stay outside, so a basic report has four call sites (leaderelection, the health group, its `/actuator/health` fallback, workflows) and makes three calls when the group answers |
| TestRenderHeaderNamesEveryReportedFact | - | the script's own header names what it reports, since it is the first thing someone running the script by hand reads |
| TestHealthParseIsKeyOrderIndependent | spring boot 3 | runs the rendered health_status and health_components under a real sh (this package's only exec, skipped where no sh is found): a connector 2.x document, every status first, reads UP with one row per component -- a composite's sub-components right after it, the ssl certificate chain's validity status not among them |
| TestHealthParseIsKeyOrderIndependent | spring boot 4 | the same document with every key sorted, as connector 3.x writes it (status last), reads the same verdict and the same rows -- reading it from the wrong end is what left 3.x containers unhealthy and pods NotReady |
| TestHealthParseIsKeyOrderIndependent | down, spring boot 3 | a DOWN instance with a healthy component after the failing one reads DOWN, not the last status in the document; the error text's braces, brackets, < > and escaped quotes do not throw the component rows off |
| TestHealthParseIsKeyOrderIndependent | down, spring boot 4 | the same in sorted order, with a healthy component ahead of the failing one: DOWN, not the first status in the document |
| TestHealthParseIsKeyOrderIndependent | pretty-printed | a document with every key on its own line reads the same as a compact one |
| TestHealthParseIsKeyOrderIndependent | no components | a document without components (show-details off) reads its status and prints no rows |
| TestHealthParseIsKeyOrderIndependent | no document | an empty answer prints nothing, which the healthcheck reports as unreachable |
| TestRenderReadsHealthFromTheStatusGroupFirst | - | the report reads the `solmq` health group first, `/actuator/health` only when the group answers nothing, and says on stderr, naming the group, when neither answers; HealthGroup is spec.StatusHealthGroup; the `--health` branch keeps its one call to `/actuator/health` |
| TestRenderRawOutputOnlyUnderTheArg | - | RAW is set from `--raw`; the `health-json:` line, the health document joined onto one line, and the `binding:` lines are printed only inside its block, after the health fetch they reuse, with no actuator call of their own |
| TestBindingBindersPairsEachBindingWithItsBinder | generated config | binding_binders, run under sh against testdata/golden's application.yml, pairs every `input-N`/`output-N` binding with its `binder:` in config order |
| TestBindingBindersPairsEachBindingWithItsBinder | hand-written config | a binding with no `binder:` key prints nothing, `default-binder` and the extended `jms.bindings` block are not read as a binding's binder, and a quoted binder comes through as written |

## internal/storescript

Render the startup script a docker service carrying created stores starts through -- decode every base64 store into the in-memory store directory, then exec the image's own command -- and record that command.

Tests: [storescript_test.go](../internal/storescript/storescript_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderDecodesEveryStoreThenExecs | - | the script stops on any failure, keeps the decoded stores to the connector's identity (umask 077), decodes each file of /app/external/var/stores through a temporary name into /app/external/classpath/truststores under its own name, and execs the command it is handed |
| TestCommandIsTheImagesOwn | - | the script sits at /app/external/.stores-init, and Command is the image's entrypoint then its command, returned as a copy |

## internal/deploy

Render the Kubernetes manifests -- Namespace/ConfigMap/Secret/Deployment/Service, syslog, libs (PVC or download), and multi-instance layout.

Tests: [deploy_test.go](../internal/deploy/deploy_test.go), [imagepull_test.go](../internal/deploy/imagepull_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull | - | rendered output contains all expected fragments: configmap, secrets, deployment env/mounts/probes/resources, service; blank line preserved in block scalar |
| TestRenderFull_ExactDocument | - | Render output matches wantRenderFull golden document byte-for-byte, including the pod securityContext of the default non-root run-as and the connector container's restrictions |
| TestRenderNoSecretsNoServiceNoTLS | - | output omits JAVA_TOOL_OPTIONS, envFrom, Secret, Service, stores volume, logback-spring.xml, syslog host, libs volume, initContainers |
| TestRenderJavaOptionsEnv | - | java-options.tool and .jdk become JAVA_TOOL_OPTIONS / JDK_JAVA_OPTIONS env entries, the MQ TLS flag first, and the operator's options alone open the env: list |
| TestEnvValueQuoteEscapes | - | envValueQuote doubles '$' (kubernetes expands $(VAR) in env values) and escapes quotes and backslashes -- the render-side half of the java-options gate |
| TestRenderSyslogUDP | - | UDP syslog config emits SyslogAppender, appname/host/port env vars, logback mount; no LogstashTcpSocketAppender |
| TestRenderSyslogTCP | - | TCP syslog config emits LogstashTcpSocketAppender and destination tag; no SyslogAppender or AsyncAppender |
| TestRenderLibsPVCExisting | - | existing libs PVC yields claimName and mount, no initContainers or PersistentVolume |
| TestRenderLibsPVCCreate | - | created libs PVC emits PV/PVC docs with NFS server/path and storage size, the claim named `<deployment>-libs` and its PV `<namespace>-<deployment>-libs-pv`, the retired create.name nowhere in the manifest, PV/PVC preceding the Deployment |
| TestCreatedSecretsTakeDerivedNames | - | the credentials and stores Secrets are named and mounted as `<deployment>-credentials` / `<deployment>-stores`, and the retired create.name keys never reach the manifest -- so two instances in one namespace cannot overwrite or delete each other's |
| TestRenderLibsDownload | emptyDir | init container wgets each jar url into /libs, mounted via emptyDir |
| TestRenderLibsDownload | existing PVC download | download target uses claimName dl-pvc instead of emptyDir |
| TestRenderNamespaceAlwaysFirst | - | Namespace doc is first in output and precedes ConfigMap |
| TestRenderExistingSecrets | - | an existing credentials Secret and a store held only in an operator's Secret produce no Secret doc, referencing my-creds and projecting my-tls |
| TestStoresVolumeProjectsEverySource | two host-held stores | one projected volume (defaultMode 0400) over the tool's stores Secret, each store keyed and mounted by its name |
| TestStoresVolumeProjectsEverySource | one file under two names | one file defined under two names is two keys and two paths |
| TestStoresVolumeProjectsEverySource | two stores from one operator Secret | one source for the operator's Secret, each store's key mounted under the store's name |
| TestStoresVolumeProjectsEverySource | a Secret first selected after an operator one | the tool's Secret comes first, then each operator Secret once, in first-use order |
| TestStoresVolumeProjectsEverySource | a file beside a Secret is read from the Secret | a store naming both is projected from the operator's Secret only |
| TestStoresVolumeProjectsEverySource | no store selected | no stores volume at all |
| TestStoresVolumeProjectsEverySource | no model | storeSources returns nil: no store to list |
| TestRenderConfigMapStatusScript | - | the ConfigMap always carries the status script under its own `status` key, alongside application.yml |
| TestRenderDeploymentNodeSelector | - | the pod spec carries nodeSelector after automountServiceAccountToken, keys sorted, a key or value YAML would misread ("2", "true", "") quoted; no selector writes none |
| TestRenderDeploymentTolerations | - | tolerations are written in Kubernetes' shape, each field only when set, a value YAML would misread quoted, before the pod's securityContext |
| TestRenderDeploymentAffinityAndSpread | verbatim + preferred | the affinity block is written as it stands, then spread's preferred anti-affinity on kubernetes.io/hostname against app: <name>, weight 100 |
| TestRenderDeploymentAffinityAndSpread | required | spread: required alone writes a required anti-affinity term |
| TestRenderDeploymentAffinityAndSpread | none | spread none or unset and no affinity write no affinity block |
| TestRenderStatusScriptMountAfterLibs | - | the single-file status mount is declared after the libs directory mount, so it is not shadowed, and carries `subPath: status` |
| TestProbesSplitLivenessFromReadiness | - | liveness stays a tcpSocket check so a slow downstream cannot become a restart loop, while readiness execs `sh <ContainerPath> --health` with an explicit timeoutSeconds (kubernetes defaults it to 1s) against a path the pod actually mounts |
| TestManagementPort | Defaults.Management.Port 9999 | ManagementPort returns 9999, ignoring Kube.Service.Port entirely |
| TestManagementPort | empty Defaults | ManagementPort returns 8090 (the connector default) |
| TestManagementPort | nil Defaults | ManagementPort returns 8090 (the connector default) |
| TestRenderLeaderModeLabels | nil Defaults / standalone | the le-mode label is standalone and role: active is present |
| TestRenderLeaderModeLabels | active_active | le-mode is active_active and role: active is present |
| TestRenderLeaderModeLabels | active_standby | le-mode is active_standby and role: active is withheld (only the actuator knows the live role) |
| TestRenderSelectorMatchLabelsAppOnly | - | spec.selector.matchLabels stays app-only even when the pod template carries le-mode/role, since a selector must stay immutable for the Deployment's life |
| TestQuoteRes | 1 | quoteRes("1") returns quoted "1" |
| TestQuoteRes | 250m | quoteRes("250m") returns unquoted 250m |
| TestQuoteRes | 512Mi | quoteRes("512Mi") returns unquoted 512Mi |
| TestRenderNoResources | - | empty Resources produces no resources: block in output |
| TestRenderServicePort | host:container distinct ports | Render emits port: 8081 / targetPort: 9000 verbatim from the given spec.Port |
| TestRenderServicePort | resolved to the default management port | Render emits port/targetPort 8090 for a Port already resolved to the connector default |
| TestRenderServicePort | resolved to a non-default management port | Render emits port/targetPort 9500 for a Port already resolved to defaults.management.port 9500 |
| TestTeardownReversesTheDocumentOrder | - | with a libs PVC present, apply orders the claim before the Deployment that mounts it; teardown fully reverses the set -- Deployment before the claim, Service before ConfigMap -- with separators intact and as many as apply's (the Namespace dropped, the stores Secret named) |
| TestTeardownNamesSecretsOnly | - | a teardown renders every Secret the tool owns by kind and name alone -- credentials and image-pull when created, the stores Secret always -- carries no Secret payload, and never names an operator's Secret |
| TestOmitNamespaceDropsOnlyTheNamespace | - | Input.OmitNamespace leaves the Namespace document out of an apply -- the ConfigMap comes first with no leading separator and nothing else differs -- which is how a deploy settles the namespace on its own while `generate` keeps it |
| TestWrittenIsWhatTheRenderEmits | no store, no pull secret / store and pull secret / teardown / Service disabled | Written narrows spec.CreatedNames to what one render writes -- the stores Secret only when a store is carried, the image-pull Secret only when built, both always for a teardown, the Service only when enabled -- so the permission probes and the owned objects cannot drift |
| TestLibsPVNameIsNamespaced | - | the same `libs.pvc.create.name` in two different namespaces derives two different PV names, each suffixed `-pv` and carrying its own namespace |
| TestImagePullSecretStates | no block / name alone / created | imagePullSecrets is absent, rendered without a Secret, or rendered with one -- the middle case is what stops an apply overwriting a Secret the operator built |
| TestImagePullSecretPayloadIsOpaqueToDeploy | - | deploy places the already-encoded payload verbatim and the registry password appears nowhere outside it |
| TestEnvBlockOmittedWhenEmpty | nothing to emit | env: is omitted entirely rather than rendered with nothing beneath it, since the timezone is one optional top-level key |
| TestEnvBlockOmittedWhenEmpty | TZ only / MQTLS only | either entry alone opens the block, and no timezone means no TZ entry |

## internal/dockergen

Render the docker compose file from the target model.

Tests: [dockergen_test.go](../internal/dockergen/dockergen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull_WithEverything | - | full golden output with project-name, creds, store, libs, MQTLS, ports, timezone and the default run-as with its posture (user, privileged false, cap_drop ALL, no-new-privileges) matches exactly, the deploy digest label after the role label; the compose project is the document's first line, and every name, bind-mount entry, port, user and config or secret key is a double-quoted scalar (compose.Quote) while the image, numbers and constants stay bare |
| TestRenderFull_Minimal | - | minimal golden output carries the deploy digest label and omits the leading name: line along with restart, ports, environment, env_file, volumes blocks; run-as image drops the user: line and keeps the rest of the posture |
| TestEnvironmentBranches | TZ only, MQTLS false | environment block contains only the quoted TZ: "UTC", no JAVA_TOOL_OPTIONS |
| TestEnvironmentBranches | MQTLS only, no timezone | environment block contains only JAVA_TOOL_OPTIONS line, no TZ |
| TestJavaOptionsEnvironment | - | both JVM options variables land, quoted by compose.Quote, in the service environment, the MQ TLS flag first in JAVA_TOOL_OPTIONS, jdk options alone open the block, and an empty java-options block adds nothing |
| TestSecretsBranches | at least one secret | the per-service secrets list (long syntax, absolute target, quoted source) and the top-level environment-provider secrets block (quoted key, bare variable) are both emitted, through compose.ServiceSecrets and compose.TopSecrets |
| TestSecretsBranches | no secrets | neither block, nor any env_file line, is ever emitted |
| TestLabelsPerMode | empty defaults to standalone | le-mode label defaults to standalone and role: active is present |
| TestLabelsPerMode | standalone / active_active | le-mode matches the mode and role: active is present |
| TestLabelsPerMode | active_standby | le-mode is active_standby and role: active is withheld |
| TestContentIndentationAndBlankLines | nested key indentation | nested line gets extra indent on top of 6-space block indent |
| TestContentIndentationAndBlankLines | blank line preserved | blank line stays empty with no spaces between content lines |
| TestContentIndentationAndBlankLines | no trailing spaces | no rendered line has trailing spaces |
| TestStoresOnlyAndLibsOnly | stores only | volumes block contains only the store mount line |
| TestStoresOnlyAndLibsOnly | libs only | volumes block contains only the libs mount line |
| TestSplitLinesNoTrailingNewline | - | app.yml lacking trailing newline still renders content line with no dropped element |
| TestStatusScriptConfigSourceAndTarget | - | the service references a second quoted `<name>-status` config and mounts it at /app/external/.status-script |
| TestHealthcheckRunsTheStatusScript | - | the service declares a healthcheck built from the statusscript constants, in the exec form (compose.FlowList) rather than CMD-SHELL, against the path the configs block mounts -- without it docker populates no .State.Health and the HEALTH column can only read n/a |
| TestStatusScriptContentIsEscaped | - | the status script body is inlined under the status config's content: block, indented 6 spaces, blank line preserved as truly empty, and its shell `$` doubled |
| TestContentEscapesDollarsForCompose | $VAR / ${VAR} / ${VAR:-default} / $(cmd) / $$ | each shape reaches the content block with every `$` doubled, so compose's interpolation pass delivers it unchanged instead of blanking it or rejecting the document |
| TestContentEscapesDollarsForCompose | no lone `$` | dropping every `$$` pair from the rendered document leaves no `$` behind anywhere |
| TestAppYAMLSecretPlaceholdersAreNotInterpolated | - | application.yml's ${...} credential placeholders render doubled so compose cannot substitute the values the CLI passes it, while the `secrets:` provider entries -- compose's own -- stay unescaped |
| TestRenderSyslogAddsConfigAndEnv | - | compose inlines the logback config as a third configs entry targeted at logback.ContainerPath, and sets the three quoted LOGGING_SYSLOG_* env vars the config reads at runtime |
| TestRenderSyslogTCPUsesTheLogstashAppender | - | the protocol reaches the inlined config: tcp renders the logstash appender, since the two have different classpath requirements and the wrong one fails at runtime rather than at generate time |
| TestRenderWithoutSyslogEmitsNeither | - | no block, no config entry, no env vars |
| TestRenderCreatedStores | - | a created store makes the service start through the storescript with the image's own command handed over, mounts the encoded store at /app/external/var/stores/<name> after the credentials, adds the script as a config (its $ doubled), a 0700 tmpfs owned by the run-as uid:gid and the stores label filled from ${_GEN_TLS_STORES_SHA256}, then the deploy digest label, and declares the store a top-level environment secret after the credentials; an existing store is still bind-mounted |
| TestRenderCreatedStoresRunAsImageAndNone | run-as image | the tmpfs names no uid or gid |
| TestRenderCreatedStoresRunAsImageAndNone | no store selected | no entrypoint, command, tmpfs, script, stores label or encoded mount appears; the deploy digest label (`solace-connector/deploy-sha256`, `${_GEN_DEPLOY_SHA256:-}`) does, since it does not depend on a store |
| TestRenderCreatedStoresRunAsImageAndNone | existing stores alone | the stores digest label, written with a :- default, and none of the created-store machinery |

## internal/podmangen

Render the quadlet unit from the target model.

Tests: [podmangen_test.go](../internal/podmangen/podmangen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderQuadletSecretsCarryNoValues | - | each credential appears only as a Secret= directive naming its store entry and absolute target; the value itself never reaches the unit, a world-readable file in the quadlet directory |
| TestRenderQuadletFull | - | full input yields 1 unit named solmq-connector.container with content matching golden string: User=/Group= of the default run-as, NoNewPrivileges and DropCapability (quadlet.Posture), every Environment= quoted whole (quadlet.Environment), application.yml, the truststore and the status script as Secret= mounts, the libs directory the only Volume= |
| TestRenderQuadletMinimal | - | minimal input yields 1 unit with content matching golden string (no Service section, no restart, no Volume=; run-as image writes no User=/Group= but keeps NoNewPrivileges and DropCapability) |
| TestLeaderLabelsPerMode | empty defaults to standalone / standalone / active_active | the unit carries the le-mode label and role: active |
| TestLeaderLabelsPerMode | active_standby | the unit carries le-mode active_standby and withholds role: active |
| TestQuadletMountsNothingFromTheHostButLibs | - | with credentials, a store, the status script, syslog and libs all set, the libs directory is the one Volume= line and all six Secret= lines mount at an absolute target |
| TestStatusScriptMountOmittedWhenSecretEmpty | - | an empty StatusScriptSecret omits the status mount and the healthcheck that execs it, rather than writing a Secret= line with no name or declaring a check that cannot run |
| TestHealthcheckRunsTheStatusScript | - | the unit declares HealthCmd and its cadence inside [Container], built from the statusscript constants, and leaves HealthOnFailure unset so the check reports without restarting anything |
| TestQuadletSyslogMountsAndSetsEnv | - | podman cannot inline file content, so the unit mounts the logback config from podman's secret store via Secret= and sets the three LOGGING_SYSLOG_* vars via quoted Environment= lines |
| TestQuadletJavaOptionsEnvironment | - | both JVM options variables are set in the unit, the MQ TLS flag first, every assignment quoted whole through quadlet.Environment (a value with spaces would otherwise split), and neither appears when nothing is set |
| TestSyslogAbsentEmitsNoMountOrEnv | - | no block, no mount, no env |

## internal/runner

The helpers over the shared runner -- ParseCommand safe-tokenizing, kubectl/docker/podman deploy and remove argv, the status verb's read-only queries, the caller's context reaching the runner, the logs verb's per-platform argv, the cli verb's shared exec argv builder, quadlet scope resolution, and WriteFile modes.

Tests: [runner_test.go](../internal/runner/runner_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestExecArgvPerPlatform | kubernetes bare / with a namespace / with stdin / with a terminal / docker / podman | the shape each engine's flag parser actually accepts, with the caller's command appended: kubectl takes -n and -c after the pod and needs the `--` terminator, docker and podman (containercli.Exec) stop parsing at the container name so their flags come first and no `--` is added; the container is a constant on every kubernetes row |
| TestExecArgvRefusesATTYWithoutStdin | - | a terminal the child cannot read from can only be a caller mistake, so it is refused rather than quietly produced |
| TestExecArgvRefusesAnEmptyCommand | kubernetes / docker / podman | an exec with nothing to run is a caller mistake on every platform, refused rather than built |
| TestExecArgvUnknownPlatform | - | an unrecognised platform names the three that exist rather than producing a half-built argv |
| TestParseCommand | kubectl | parses to [kubectl], ok |
| TestParseCommand | kubectl.Exe | accepted: the .exe suffix is folded in any case |
| TestParseCommand | kubectl --kubeconfig=C:/RUNNER~1/config | accepted: an embedded tilde is an 8.3 short name |
| TestParseCommand | kubectl --context prod -n solace | parses to [kubectl `--context` prod `-n` solace] tokens, ok |
| TestParseCommand | oc | trims whitespace to [oc], ok |
| TestParseCommand | empty string | returns error, nil result |
| TestParseCommand | whitespace only | returns error, nil result |
| TestParseCommand | kubectl; rm -rf / | semicolon rejected as unsafe, error |
| TestParseCommand | kubectl $(evil) | $ and ( rejected as unsafe, error |
| TestParseCommand | kubectl \`id\` | backtick rejected as unsafe, error |
| TestParseCommand | kubectl --kubeconfig "a b" | quote+space rejected as unsafe, error |
| TestParseCommand | curl | not on the kubernetes allowlist, error |
| TestParseCommand | /tmp/evil | path, not a bare name, error |
| TestParseCommand | every refused case | the error is marked as a usage error, so it is exit 2 wherever it surfaces |
| TestParseCommandExtraAllowed | lima podman without an approval | rejected, error |
| TestParseCommandExtraAllowed | lima podman with lima approved | accepted, argv [lima podman] with the platform binary unmodified |
| TestParseCommandExtraAllowed | sudo podman with lima approved | refused as a privilege-escalation wrapper that is never permitted, marked as a usage error |
| TestParseCommandExtraAllowed | --allow-command sudo | execguard.Approve refuses it, naming `sudo solmq-conn-util` as the supported way |
| TestKubernetesDeployApplyOnStdin | - | deploy issues one call kubectl `--context` prod apply `-f` - with manifest on stdin |
| TestKubernetesRemoveUsesDeleteVerb | - | the remove action issues argv [oc delete -f - --ignore-not-found] |
| TestKubernetesRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestKubernetesRolloutStatusRejectsUnsafeCommand | - | the rollout wait goes through the same guard as the apply: an unsafe command is rejected with zero calls made |
| TestKubernetesUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestDockerUpAndDown | - | deploy runs compose -f - --project-directory <dir> up -d, remove the same with down, each with the compose file on stdin and the credentials in the child env |
| TestDockerRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestDockerDeployDigest | not running / a sibling is not the container / running with the label / label missing / inspect fails / inspect is not json / inspect answers another name / the listing fails / unsafe command | the running probe is a `ps --filter status=running` matched by exact name in Go; a running container is inspected and its deploy digest label read, and a missing label or any inspect that cannot be read gives an empty digest (a change); a listing that fails is an error naming the container; an unsafe command runs nothing |
| TestResolveQuadletScope | the six rootless decisions | scope resolved from injected inputs (euid, rootful, getenv, home, isDir, Linux), never the process, so there is no skip as root and no HOME redirect: non-root rootless, non-root asking for rootful refused, root asking for rootful rootful, root without it refused, and no euid (Windows) rootless or rootful with no session; a refusal is a usage error (exit 2) naming `podman.rootful` |
| TestResolveQuadletScope | XDG_CONFIG_HOME | an absolute value moves the rootless unit directory, a relative one is a usage error naming it, rootful ignores it |
| TestResolveQuadletScope | unresolvable home | the home lookup error is carried in a usage error |
| TestResolveQuadletScope | session | a stripped session is derived (XDG_RUNTIME_DIR and the bus), an inherited one is left alone, a missing runtime directory or a probe that cannot run is a plain error (exit 1), and non-Linux and rootful scopes have no session and probe nothing |
| TestLiveScopeEnv | - | the live inputs read this process (euid, GOOS), and the directory probe answers true, false, and false for a file or a missing path, with an error for a path that cannot be looked at |
| TestQuadletScopeWithEnv | - | the session is attached to a rootless scope's commands as a copy that never aliases the scope's slice, and a system scope's commands are untouched |
| TestEnsureLinger | already lingers / off / enable refused / probe fails | logind is asked with `loginctl show-user <uid> --property=Linger`; `yes` makes no enable call, `no` runs `loginctl enable-linger`, a refused enable names `sudo loginctl enable-linger <user>`, a probe that cannot run names the command and logind; every call carries the session; ProbeLinger reads stdout alone |
| TestPodmanIDMap | uid map / gid map / failure / unsafe command | the id mapping is read from stdout alone through podman.IDMap, a failure carries its cause, and an unsafe command is refused before anything runs |
| TestPodmanDeployReloadThenStart | inactive / active / reloading / failed / unanswered x user / system scope, and unchanged inactive / active / reloading | the state probe (`is-active <unit>`, no `--quiet`, read from stdout) runs first; when something changed the seam's write runs after it, then daemon-reload, then restart for an active or reloading unit and start for anything else, an unanswered probe included; an unchanged running unit is left alone and the output says so and names `solmq-conn-util restart`; an unchanged stopped unit is started with no reload; the question is asked only for a changed, running unit; `--user` and the session env only in user scope |
| TestPodmanDeployDeclinedChangesNothing | - | a changed, running service asks, and a "no" returns cancelled with the write, the reload and the restart never run: only the state probe ran |
| TestPodmanDeployUnaskableQuestionStopsBeforeAnyWrite | - | a question that cannot be asked (no terminal, no `--no-prompt`) ends the deploy with its own error before any write |
| TestPodmanDeployWriteFailureStopsBeforeSystemctl | - | a write that fails stops the deploy where it fails, its output kept, with no reload and no start |
| TestPodmanDeploySystemModeNoUserFlag | - | system mode issues systemctl daemon-reload (call 1, after the state probe) without `--user` flag |
| TestPodmanRemoveStopsRemovesReloads | - | stop then daemon-reload called and unit file removed from disk |
| TestPodmanDeployStartFailureIsReported | daemon-reload / start / restart | a failing daemon-reload (call 1, after the is-active probe) stops there; a failing start or restart (call 2, whose verb the printed state picks: "inactive" with exit 3 starts, "active" restarts) surfaces naming the verb, `systemctl start a.service` or `systemctl restart a.service`; nothing runs after the failure |
| TestDockerUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestDockerImageCommand | an image the engine has | one image inspect call returns the entrypoint and command |
| TestDockerImageCommand | not pulled / one line / not json | an engine that cannot answer, or an answer that does not parse, is found=false with no error, for the caller to pull and ask again |
| TestDockerImageCommand | unsafe command | rejected before anything runs |
| TestDockerImagePull | - | the pull runs through the configured engine command, and an unsafe command is rejected before anything runs |
| TestPodmanRemoveStopFailureIsReported | - | stop failure surfaces error containing 'stop solmq-connector.service' |
| TestPodmanSecretCreateRemovesThenCreatesValueOnStdin | - | PodmanSecretCreate issues rm `--ignore` then create with the value on stdin, never in argv |
| TestPodmanSecretCreateReplacesInOneCall | - | with replace, one secret create --replace carries the value on stdin, and a failure is named |
| TestPodmanSecretMatches | - | the stored value is read back from stdout alone (`secret inspect --showsecret --format {{.SecretData}}`) and matches only when it equals the new one exactly, a warning on stderr notwithstanding; a different value, a missing or extra newline, a failed probe, nothing printed, an unsafe command and an empty name all read as a change, and the value never reaches argv |
| TestPodmanVersion | - | the client version is read from stdout alone; 4.7 and newer (any suffix, a leading v, surrounding space) support --replace through podman.SupportsReplace, older or unparsable versions do not; a failed call and an unsafe command are errors |
| TestPodmanSecretCreateSkipsCreateWhenRmFails | - | a failed rm surfaces an error naming it, and create never runs |
| TestPodmanSecretCreateReportsCreateFailure | - | a failed create surfaces an error naming it, after rm still ran |
| TestPodmanSecretCreateRejectsUnsafeCommand | - | an unsafe command is rejected before anything runs |
| TestPodmanSecretRemoveBatchesNames | - | PodmanSecretRemove issues one batched `secret rm --ignore` call naming every secret |
| TestPodmanSecretRemoveNoNamesIsNoop | - | an empty name list invokes the runner zero times and returns no error |
| TestPodmanSecretRemoveRejectsUnsafeCommand | - | an unsafe command is rejected before anything runs |
| TestPodmanSecretRemoveReportsFailure | - | a failed rm surfaces an error naming the operation |
| TestWriteFileCreatesDirsAndMode | - | creates nested dirs, writes content, sets mode 0600 (non-windows) |
| TestWriteFileDoesNotTightenExistingFileMode | - | content replaced but existing 0644 mode left unchanged (non-windows) |
| TestWriteFileParentIsFileReturnsError | - | parent path is a file: MkdirAll error surfaces naming the blocker path |
| TestWriteFileTargetIsDirectoryReturnsError | - | target path is a directory: write error surfaces naming the target path |
| TestPreflightKubernetesArgvDeployNoNamespace | - | a kubernetes preflight of one create probe issues <argv> auth can-i create deployments with no namespace flag when the probe names none, and no other call |
| TestPreflightKubernetesArgvRemoveWithNamespace | - | a delete probe in a namespace issues <argv> auth can-i delete deployments -n <ns> |
| TestPreflightDockerArgvIsInfo | - | docker preflight issues <argv> info |
| TestPreflightPodmanArgvIsInfo | - | podman preflight issues <argv> info |
| TestPreflightFailureWrapsPlatformHint | kubernetes / docker / rootless podman / rootful podman | a failing kubernetes probe says "preflight failed for kubernetes", keeps the cause and carries the login hint; docker and podman give containercli.InfoError's one message -- engine, `<command> info`, the cause, the engine's own output and the hint for Docker, rootless podman (`systemctl --user start podman.socket`) or rootful podman -- with nothing left to print twice |
| TestPreflightFailureDropsControlCharactersFromEngineOutput | - | the engine's output inside the error loses its control characters (a terminal escape) and keeps its line breaks |
| TestKubernetesReadVerbsAskNothing | status / logs / cli | KubernetesProbes returns nothing for the read verbs and Preflight makes no call on kubernetes -- a view-only operator is never turned away by a question; docker still runs info |
| TestPreflightRefusalNamesThePermission | "no" / "no - reason" / unreachable | a can-i "no" is a *rbac.RefusedError behind "preflight failed for kubernetes" naming the permission, the namespace (or the current one) and the answer, without the login hint and with nothing else to print; an unreachable cluster still gets the login hint |
| TestPreflightRejectsDisallowedBinaryBeforeRunning | - | a command outside the platform allowlist (curl) is rejected before the probe ever runs, zero runner calls |
| TestPreflightExtraAllowedThreadsThrough | - | "lima podman" is rejected without an approval and accepted with lima approved, running argv [lima podman info]; "sudo podman" is refused as an escalation wrapper before any probe runs |
| TestKubernetesProbesUnknownAction | unknown action / unmapped kind | an action other than deploy/remove/start/stop/restart/status/logs/cli, and an object kind with no resource mapped, are errors naming what is wrong, with no probe built |
| TestPreflightKubernetesLifecycleAsksToPatch | - | start, stop and restart ask kubectl auth can-i to patch deployments in the namespace, one call |
| TestPreflightStopsAtTheFirstRefusal | refused second probe / all allowed | every probe of a command runs in order before its first write and the first refusal ends the run -- the later probes are not asked |
| TestPreflightRefusesAnAnswerThatIsNotYes | - | an exit 0 whose answer is not "yes" (a wrapper changed the output) is refused, never taken for a permission nobody confirmed |
| TestKubernetesProbesDeployAndRemoveSets | no secret, no volume / secrets / libs claim with its volume | create (deploy) and delete (remove) of the plural resource of every written object, once each, plus persistentvolumes (asked with --all-namespaces) for the libs claim and never namespaces |
| TestKubernetesNamespaceState | exists / absent / somebody else's / read refused | exactly namespace/<ns> exists, nothing is absent, anything else and any failure is unknown, read from stdout alone with the argv `get namespace <ns> --ignore-not-found -o name` |
| TestKubernetesNamespaceGateArgv | - | the gate's listing is nsgate.List's one query and its delete is `delete namespace <ns> --ignore-not-found` |
| TestCheckProbeReadsStdoutAlone | - | a "yes" on stdout passes even with a warning on stderr after it |
| TestLifecycleArgv | kubernetes stop / start / restart / rollout status | scale to 0, scale to the replicas, rollout restart and rollout status (--timeout=300s), each with `-n <ns>` after the subcommand when one is set |
| TestLifecycleArgv | docker stop / start / restart | containercli.Stop, Start and Restart: the engine command, its flags kept, then the action and the container |
| TestLifecycleArgv | podman start, user unit / stop, system unit | systemctl in the unit's scope |
| TestLifecycleArgv | unknown action / unsafe command | refused before anything runs |
| TestKubernetesPodsJSONArgv | by selector, no namespace / by selector in a namespace / explicit names / every namespace | `get pods ... -o json` in each of its three scopings; `--all-namespaces` outranks a namespace resolved from env.yaml, so a cluster-wide search is never narrowed back down |
| TestKubernetesPodsJSONRunFailureWraps | - | a run failure surfaces as an error naming the "listing pods" operation |
| TestKubernetesGetJSONArgv | deployment in a namespace / no namespace / a referenced object | `get <kind> <name> [-n ns] -o json` for the workload summary and for each object a pod references |
| TestKubernetesGetJSONMissingObjectIsAnError | - | a missing object exits non-zero and the error names it, which is the answer the components check asks for -- reported as MISSING rather than as a failed run |
| TestKubernetesTopArgv | by selector / explicit names / every namespace | `top pod --containers --no-headers`, attributing the sample to one container so it is comparable with that container's limits |
| TestKubernetesTopWithoutMetricsAPIWraps | - | a cluster with no metrics API fails here, naming the operation, so the caller can degrade to a note |
| TestEngineInspectJSONArgv | - | one `container inspect a b c` covers every target (the container subcommand answers for containers alone), and a chained command keeps its own tokens ahead of the subcommand |
| TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing | - | inspecting nothing is an error and starts no process |
| TestEngineInspectJSONFailureNamesTheTargets | - | a failure names the containers it was asked about |
| TestEngineFailuresDropControlCharactersFromTheEnginesOutput | inspect / image inspect / stats / list | the engine's output appended to an error loses its control characters and keeps its line breaks |
| TestEngineImageInspectJSONArgv | - | `image inspect <ref>`, where the registry digest lives (the container's own inspect does not carry it) |
| TestEngineStatsArgv | - | `stats --no-stream --format <template>`: the four tab-separated template fields docker and podman render identically, and `--no-stream` so the call cannot stream forever |
| TestEngineListArgv | - | `ps --all --no-trunc --format <name, image, status>` for `--all` discovery, including stopped containers -- an instance that died is what such a search is for -- and, for the redeploy check, with `--filter status=running` |
| TestSystemctlNRestarts | user scope / system scope / unknown unit answers nothing / negative count is refused / no systemd on this host | `systemctl [--user] show <unit> -p NRestarts --value`, the only truthful restart count under quadlet; an empty, negative or unreadable answer is an error so the caller can fall back to the container's own counter |
| TestInstallScriptArgv | kubernetes no namespace / kubernetes with namespace / docker / podman | InstallScript execs `mkdir -p '<dir>' && cat > '<path>'`; on kubernetes as kube.Upload (`-i`, `-n <namespace>` when given, the connector container), on docker and podman as containercli.Upload, each single-quoting both paths |
| TestInstallScriptPassesScriptOnStdinNotArgv | - | the script body travels on stdin, never appearing in any argv token |
| TestInstallScriptRefusesAnEmptyTargetOnAnEngine | - | an engine upload that names no container is refused by containercli.Upload, zero runner calls |
| TestInstallScriptUnknownPlatform | - | an unknown platform is rejected, zero runner calls |
| TestRunStatusScriptArgv | kubernetes no namespace / kubernetes with namespace / kubernetes with an argument / docker / podman with an argument | RunStatusScript checks for the script and runs it in one exec through each platform's exec form: `sh -c` with a fixed program, the path as `$0` and any argument after it, so neither is spliced into the shell text; `-n <namespace>` only for kubernetes when given |
| TestRunStatusScriptReadsTheAbsentMarker | absent marker / marker among engine chatter / a report is present / marker inside a longer line is not absent | the absent marker on a line of its own means nothing ran and no report is returned; anything else is the script's report, returned whole |
| TestRunStatusScriptReturnsOutputAlongsideNonZeroExit | - | a failed exec with no marker is an error returned alongside the engine's output, and is not mistaken for an absent script |
| TestRunStatusScriptUnknownPlatform | - | an unknown platform is rejected, zero runner calls |
| TestLogsArgvPerPlatform | kubernetes bare / kubernetes without a namespace or container / kubernetes with every option / docker bare / docker with every option it has / podman reads the container, never the journal / tail zero is not tail all | kubectl takes the pod positionally with the namespace and container as flags; docker and podman take options first and the container name last; TailAll adds no flag while an explicit 0 does |
| TestLogsArgvRefusesPreviousOffKubernetes | docker / podman | `--previous` is refused with containercli.NoPrevious's message, naming the option and the engine, rather than dropped, so a caller asking for the previous log is never handed the current one |
| TestLogsArgvUnknownPlatform | - | an unrecognised platform names all three that exist rather than producing a half-built argv |
| TestParsingHelpersIgnoreAWarningOnStderr | KubernetesPodsJSON / KubernetesGetJSON / KubernetesListJSON / KubernetesTop / EngineInspectJSON / EngineImageInspectJSON / EngineStats / EngineList / SystemctlNRestarts / DockerImageCommand | every helper that parses rather than scans its output returns the payload alone when a warning sits on stderr, through `Split` with the two streams kept apart |
| TestParsedFailureCarriesBothStreams | - | a failed command still reports what it said, on whichever stream it said it on, so splitting the streams did not cost the error context |
| TestHelpersPassTheCallersContextToTheRunner | Kubernetes (combined) | a context that is already done starts nothing and the error wraps `context.Canceled`: the helper hands its caller's context to the shared runner rather than minting one |
| TestHelpersPassTheCallersContextToTheRunner | KubernetesGetJSON (split) | the same through the parsed-output path |
| TestHelpersPassTheCallersContextToTheRunner | PodmanDeploy (several commands) | the first of several commands is refused, so none of the sequence starts |
| TestHelpersPassTheCallersContextToTheRunner | Preflight | the read-only probe is refused too, so a cancelled run reaches the platform with nothing |

## internal/validate

Validate the parsed model -- per-side rules, connection refs, leader election, the docker/podman/kubernetes platform sections, ports, container names, TLS/stores wiring, and the safe-token charset.

Tests: [validate_test.go](../internal/validate/validate_test.go), [validate_extra_test.go](../internal/validate/validate_extra_test.go), [validate_deploycommand_test.go](../internal/validate/validate_deploycommand_test.go), [validate_image_test.go](../internal/validate/validate_image_test.go), [validate_javaoptions_test.go](../internal/validate/validate_javaoptions_test.go), [validate_derivednames_test.go](../internal/validate/validate_derivednames_test.go), [validate_transform_test.go](../internal/validate/validate_transform_test.go), [validate_extrakeys_test.go](../internal/validate/validate_extrakeys_test.go), [validate_runas_test.go](../internal/validate/validate_runas_test.go), [validate_apiprops_test.go](../internal/validate/validate_apiprops_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestBinderIdentityUsesTheCredentialPair | different -env usernames | two -env usernames on one host are different binders, so no false key-alias conflict |
| TestBinderIdentityUsesTheCredentialPair | same -env username | one binder with two key-aliases still conflicts |
| TestWorkflowCap | 21 workflows | fatal error naming the count, the 20 cap, and the split-into-separate-folders remedy |
| TestWorkflowCap | 20 workflows | exactly at the cap does not error |
| TestDeployNameTooLong | 57-char name | deployment.name + "-config", the shortest name derived from it, exceeds the 63-char DNS-1123 limit |
| TestDeployNameTooLong | short name | a short deployment name has no length error |
| TestValidGoldenLikeInputPasses | - | valid solace->mq workflow selecting stores produces no errors and no unused-store warning |
| TestMissingSourceTarget | - | workflow with neither side set errors missing 'source' and missing 'target' |
| TestExactlyOneSystem | - | side with empty system errors exactly one of 'solace:' or 'mq:' |
| TestExactlyOneDestination | - | side with empty dest-kind errors exactly one of 'queue:' or 'topic:' |
| TestSolaceTopicSourceIsError | - | solace topic source errors cannot be consumed from, and no longer emits the retired non-durable-subscription advisory |
| TestConnNameFormat | - | bad mq conn-name errors host(port) format message |
| TestKeyAliasNeedsKeystore | - | solace key-alias with no keystore selected errors, naming the alias and how to add keystore: |
| TestStoreSelectionConflicts | key-alias | two sides on one binder with different aliases error conflicting key-alias, naming both values and the other file |
| TestStoreSelectionConflicts | truststore set vs none | a side selecting no truststore conflicts with one selecting ts: the merged binder would trust a store a side never asked for |
| TestStoreSelectionConflicts | keystore differs | different keystores on one binder error; identical selections do not |
| TestLeaderElectionActiveStandbyValid | - | valid active_standby leader election config passes with no errors |
| TestLeaderElectionActiveMissingQueueAndSession | - | active_active mode missing queue/conn-ref errors requires a 'queue' and requires a solace session |
| TestLeaderElectionConnRefMustBeSolace | - | leader-election conn-ref pointing to mq connection errors must be a solace connection |
| TestLeaderElectionInvalidMode | - | unknown leader-election mode 'bogus' errors is invalid |
| TestLeaderElectionSolaceKeyRenamed | active_standby / standalone | a `leader-election.solace` key errors, naming `leader-election.session` as the key to use, in every mode, not only the active ones |
| TestLeaderElectionConnRefAndInlineSession | - | conn-ref set alongside an inline session errors sets both conn-ref |
| TestLeaderElectionSessionRejectsBindingFields | queue / topic / consumer / producer | each binding key inside `session:` errors may not set queue/topic/consumer/producer |
| TestLeaderElectionInlineSessionValid | - | a bare solace tuple under `session:` passes clean |
| TestLeaderSessionPasswordConflict | inline session disagrees | a session sharing a binder tuple with a different password errors reaches the same broker tuple with a different password |
| TestLeaderSessionPasswordConflict | conn-ref session disagrees | same conflict through the conn-ref form |
| TestLeaderSessionPasswordConflict | session agrees | one password across both passes |
| TestLeaderSessionPasswordConflict | session is a different broker | a distinct tuple is a distinct binder, so the passwords may differ |
| TestMQCipherRequiresTLS | - | mq cipher set with tls false errors require 'tls: true' |
| TestDeployKubeChecks | - | a bad deployment name errors with not a valid DNS-1123 label |
| TestCredentialsEnvChecks | - | an -env variable that is unset, or set but empty, warns (never errors), naming the variable and which of the two it is |
| TestCheckSideMQMissingFields | - | mq side missing conn-name/queue-manager/channel errors each; user/password not flagged missing |
| TestCheckSideSolaceMissingAndBadScheme | missing-host-vpn | empty solace side errors missing host and msg-vpn; client creds not flagged |
| TestCheckSideSolaceMissingAndBadScheme | bad-scheme | http scheme host errors must start with tcp:// or tcps:// |
| TestSolaceKeyAliasRequiresTCPSAndKeystore | - | solace key-alias and store selections with a plain tcp host each error, requiring a tcps:// host |
| TestMQKeyAliasRequiresKeystore | - | mq key-alias with no keystore selected errors the same way |
| TestMQStoreSelectionRequiresTLS | - | an mq side selecting a store without tls: true errors truststore/keystore require tls: true |
| TestMQTLSWithoutCipherWarns | inline side without cipher | tls: true with no cipher suite warns once, on the side, never an error |
| TestMQTLSWithoutCipherWarns | cipher: / ssl-cipher-spec of its own / sslCipherSuite from mq-defaults | each counts as a cipher suite: no warning |
| TestMQTLSWithoutCipherWarns | no tls | a plain-text connection is not warned |
| TestMQTLSWithoutCipherWarns | conn-ref | warned once, on connections.<name> in env.yaml, not on the referring side |
| TestStoreSelectionMustBeDefined | - | a truststore or keystore naming no tls.stores entry errors, listing the defined names (none when there are none) |
| TestCheckKubeRequiredAndReplicas | - | kube deployment missing name/namespace and replicas 3 errors each field plus replicas: 1 message |
| TestCheckKubeNodeSelector | valid labels | prefixed, empty-valued and mixed-case label keys and values pass |
| TestCheckKubeNodeSelector | bad keys | a space, a leading '-', two '/', an uppercase or empty prefix, and a 64-character name are each refused as a label key |
| TestCheckKubeNodeSelector | bad values | a leading '-', a space, and 64 characters are each refused as a label value |
| TestCheckKubeTolerations | valid | Equal with key and value, a prefixed key with only an effect, a key-less Exists, and NoExecute with tolerationSeconds all pass |
| TestCheckKubeTolerations | invalid | a key-less Equal, a value with Exists, an unknown operator or effect, a bad label key or value, and tolerationSeconds without NoExecute are each refused |
| TestCheckKubeSpreadAndAffinity | valid | spread none or unset beside an affinity with nodeAffinity and podAntiAffinity passes, and affinity: with no value counts as unset |
| TestCheckKubeSpreadAndAffinity | invalid | an unknown spread, a non-mapping affinity, an unknown affinity kind, podAntiAffinity with a spread, an anchor or alias, and a key that is not plain are each refused |
| TestCheckKubeServicePort | - | kubernetes.service.port is range-checked like docker/podman ports: a scalar or distinct host:container pair both pass, and an out-of-range host or container side each error independently naming the offending side |
| TestCheckKubeCredentialCreateRemovedKeys | source/variables/values-file set | credentials.create carrying `source`, `variables`, and `values-file` errors naming all three and telling the operator to remove them |
| TestCheckKubeCredentialCreateRemovedKeys | source alone | credentials.create carrying only `source` errors naming it alone |
| TestCheckKubeCredentialCreateRemovedKeys | bare name | a bare (retired) create.name trips no removed-keys error |
| TestRetiredStoresBlockTiers | - | kubernetes.secrets.stores create: (any spelling) or an empty block is ignored by generate/deploy and an error under Lint asking for its removal; existing: NAME is an error on every kubernetes path naming each selected store's replacement, a warning on teardown, and offers both ways on for a created store |
| TestNoStoresWiringWarning | - | a workflow selecting a store with a kubernetes section and no stores block warns about nothing |
| TestClusterOnlyStoreNeedsAHostSource | - | docker and podman refuse a selected store held only in a kubernetes Secret, accept it with a file or sources beside the Secret, and kubernetes and an unselected store raise nothing |
| TestCheckCredRules | both literal and env set | errors target password and target password-env are both set |
| TestCheckCredRules | env value is a ${...} reference | errors target password-env holds a variable reference, not a name |
| TestCheckCredRules | env value not a valid identifier | errors is not a valid environment variable name |
| TestCheckCredRules | -env var unset in this environment | warns (not errors) which is not set; export it before running |
| TestCheckCredRules | -env var set but empty | warns (not errors) which is set but empty; export it before running |
| TestCheckCredInvalidNameNotEchoed | a value with a space / a ${...} reference / a leading digit | an invalid -env name is never printed back, in an error or a warning, in case it is the secret itself |
| TestCheckCredRejectsReservedPrefix | -env starting with spec.GeneratedNamePrefix | rejected: the prefix is reserved for derived mount names, which is what keeps an operator's name and a derived name from ever meeting |
| TestCheckCredRejectsReservedPrefix | MY_GEN_PASSWORD / _MY_PASSWORD / SOL_PASSWORD | accepted -- only a name *starting* with the prefix is reserved |
| TestCheckCredLiteralLooksLikeEnvRefWarns | - | a literal credential containing ${ warns naming the -env key to use instead; the value is still used as a literal |
| TestSolaceQueueDestinationWarnsNotErrors | - | mq source to solace queue target allowed, warns point-to-point |
| TestIdiomaticSolaceCombosNoEDAWarn | - | idiomatic solace topic-target/queue-source combos emit no errors and no EDA warnings |
| TestConnRefSolaceTopicSourceIsError | - | conn-ref solace source with a topic errors cannot be consumed from, so the rejection is not inline-only |
| TestConnRefStrictOnlyDestination | - | conn-ref side also setting host errors may set only queue/topic |
| TestExtraKeysQuietForLegitimateProperties | IBM starter keys / Solace direct keys / a list value / a mounted credential reference | the properties the connector reads pass through without a word: starter keys in either spelling, a nested block, a list, Solace's direct keys, and a secret-looking key referencing a mounted credential |
| TestExtraKeysManagedByTheToolIsAnError | ssl-bundle / sslBundle / queueManager / additionalProperties / msgVpn / apiProperties | ssl-bundle, derived from tls: and the truststore, and any other spelling of a key the tool reads are errors naming the key to write |
| TestExtraKeysCredentialSpellingIsAnError | Password / USER / clientPassword / clientUsername | another spelling of a credential key is an error: it would land in application.yml as a literal instead of a mounted secret |
| TestExtraKeysCipherSuiteClashesWithCipher | ssl-cipher-suite beside cipher / sslCipherSuite beside cipher / ssl-cipher-suite alone | the starter's cipher-suite key beside cipher: sets the same thing twice and is an error; alone it is fine |
| TestExtraKeysUnsafeKeyIsAnError | a space / a colon / a hash / empty | a key that is not a plain property name is refused rather than written unquoted into application.yml |
| TestExtraKeysEnvSuffixWarns | -env / _env / environment (contains env, no suffix) | an -env name outside the credential pairs is a plain property, passed through with a warning naming the pairs the tool resolves |
| TestExtraKeysSecretLookingLiteralWarns | a token literal / a nested secret / a mounted reference / an empty value / an ordinary key | a secret-looking key, at the top or inside a nested block, carrying a literal value warns with its dotted path; a ${...} reference, an empty value and an ordinary key do not |
| TestExtraKeysNearMissWarns | queue-managr / chanel / an unrelated key / a very short key / keyAlias | a key within an edit or two of one the tool reads warns "did you mean", an unrelated or very short key does not, and another spelling is the error rather than the question |
| TestExtraKeysTLSInterplayWarns | use-ibm-cipher-mappings true and false / jks / client-name fixed, per-instance, standalone | the JVM cipher-mapping flag the tool sets, the starter's jks stores that ssl-bundle supersedes, and a fixed client-name every replica of an active_* deployment would share each warn, and their safe forms do not |
| TestConnRefSideRejectsExtraKeys | - | another key beside conn-ref is a connection field: the strict conn-ref error |
| TestExtraKeysCheckedOnConnectionsSessionAndDefaults | - | a connection in env.yaml, the inline management session and the mq-defaults block are each checked under their own label on every run, the defaults block with its own wording for a per-connection key |
| TestConnRefUnknownAndSystemMismatch | unknown-ref | conn-ref to undefined connection errors is not defined under connections |
| TestConnRefUnknownAndSystemMismatch | system-mismatch | mq side referencing solace connection errors is a solace connection but referenced under mq |
| TestConnRefValidResolvesNoError | - | valid conn-ref resolution for both sides passes with no errors |
| TestCheckSyslog | missing-host | syslog without host errors logging.syslog.host is required |
| TestCheckSyslog | bad-host-chars | syslog host with semicolon errors may only contain |
| TestCheckSyslog | port-0 | syslog port 0 errors must be 1-65535 |
| TestCheckSyslog | port-70000 | syslog port 70000 errors must be 1-65535 |
| TestCheckSyslog | bad-protocol | syslog protocol xxx errors must be udp or tcp |
| TestCheckSyslog | tcp-valid | syslog tcp protocol has no errors, warns logstash-logback-encoder |
| TestCheckSyslog | udp-valid | valid udp syslog config has no errors |
| TestCheckLibs | empty-libs | no pvc/download set errors exactly one of 'pvc' or 'download' |
| TestCheckLibs | pvc-and-download | both pvc and download set errors exactly one of 'pvc' or 'download' |
| TestCheckLibs | pvc-neither | pvc with neither create nor existing errors exactly one of 'create' or 'existing' |
| TestCheckLibs | pvc-both | pvc with both create and existing errors exactly one of 'create' or 'existing' |
| TestCheckLibs | pvc-create-no-nfs | pvc create without nfs server/path errors requires nfs.server and nfs.path |
| TestCheckLibs | pvc-create-bad-name | the claim is named after deployment.name, so a retired create.name -- even Bad_Name -- is ignored outside validate: no DNS-1123 or retired-key error |
| TestCheckLibs | download-empty-urls | download with empty urls errors non-empty 'urls' list |
| TestCheckLibs | download-ftp-url | non-http(s) download url errors must be http(s) |
| TestCheckLibs | download-injection-quote | download url with quote/semicolon errors no spaces, quotes, or control characters |
| TestCheckLibs | download-injection-dollar | download url with $() errors no spaces, quotes, or control characters |
| TestCheckLibs | download-bad-pvc-name | valid download url with bad pvc name errors DNS-1123 |
| TestCheckLibs | pvc-existing-valid | pvc.existing set alone has no errors |
| TestCheckLibs | download-valid | valid download urls has no errors |
| TestCheckDocker | valid-docker | valid docker section passes with no errors |
| TestCheckDocker | missing-image-empty-cmd-bad-port | errors docker.image required, docker.command is empty, and port 0 must be 1-65535 |
| TestCheckDocker | unsafe-command | docker command with semicolon errors "which is not allowed in a command token" |
| TestCheckDocker | stores-removed | a present docker.stores errors `docker.stores is no longer configured` |
| TestCheckDocker | libs-no-dir | docker libs set without dir errors docker.libs.dir is required |
| TestCheckLibsMountPathRemoved | docker custom / podman fixed / dir alone | libs.mount-path is rejected for both sections whatever its value -- the fixed one included, since the key decides nothing -- while libs with dir alone passes |
| TestCheckDocker | checkdocker-false-gate | docker section not checked when CheckDocker is false |
| TestCheckDockerProjectName | lowercase and hyphens / digits / single char | a DNS-1123 project-name passes |
| TestCheckDockerProjectName | uppercase / leading hyphen / embedded space | rejected, naming docker.project-name and quoting the offending value |
| TestCheckDockerProjectName | underscore / trailing hyphen | rejected even though docker compose would accept both -- one name grammar across the spec |
| TestCheckDockerProjectName | empty | rejected: the check is unconditional, so an empty value means ParseEnv's defaults never ran |
| TestCheckPodmanHasNoProjectName | - | a podman section is never checked for a project name; the key is docker-only |
| TestCheckDockerPodmanSecretsRemoved | docker.secrets set | a docker section with a `.secrets` block errors naming docker.secrets as not a configurable section |
| TestCheckDockerPodmanSecretsRemoved | podman.secrets set | a podman section with a `.secrets` block errors naming podman.secrets as not a configurable section |
| TestCheckDockerPodmanSecretsRemoved | nil secrets | omitting `.secrets` trips no such error |
| TestCheckPodmanModeAndScope | valid-podman | valid podman section passes with no errors |
| TestCheckPodmanModeAndScope | run / quadlet / swarm | every value of podman.mode errors `podman.mode is no longer configured` -- quadlet included, since it is the only artifact and the key decides nothing |
| TestCheckPodmanModeAndScope | quadlet present / omitted | a present podman.quadlet block of any shape errors `podman.quadlet is no longer configured`; omitting it is clean. The unit directory follows the invoking uid -- the only thing that could ever decide it, so there is nothing left for a scope or dir key to configure |
| TestCheckPodmanStoresRemoved | present / nil | a present podman.stores errors `podman.stores is no longer configured`, saying the stores are loaded into podman's secret store; omitting it trips no such error |
| TestPodmanBaseDirIgnoredButNotedByValidate | unset | no error and no note |
| TestPodmanBaseDirIgnoredButNotedByValidate | set, under validate | no error; one Lint warning naming the value and the files earlier deploys left there (`c-application.yml`, `c-status`, `c-logback-spring.xml`), the first holding the solmq-status password |
| TestPodmanBaseDirIgnoredButNotedByValidate | set, under generate or deploy | no error and no warning: only validate notes the ignored key |
| TestPodmanBaseDirIgnoredButNotedByValidate | unsafe value, under validate | a value the old host-path gate refused is no error either, just the same note |
| TestCheckCommandMultiToken | safe-multi-token | docker command with extra safe tokens has no docker.command error |
| TestCheckCommandMultiToken | unsafe-token | docker command with $(evil) token errors docker.command[2] "which is not allowed in a command token" |
| TestCommandRulesPerPlatform | kubernetes / docker / podman | CommandRules names the field `<platform>.command`, scopes the allowlist to the platform, copies AllowedCommands and sets no subword |
| TestCommandRulesPerPlatform | unknown platform | fails as a wiring error naming the field rather than accepting anything |
| TestCommandGuardMessages | path / unlisted binary / end-of-flags / bare word / empty / unsafe character / sudo / sudo after argv[0] | the shared guard's messages reach the operator named by their field, indexed by token |
| TestCommandGuardAccepts | kubectl / oc / flags / docker / podman / kubectl.exe / kubectl.Exe / embedded tilde in a flag value | accepted; a leading tilde in a flag value is still refused |
| TestCheckKubeCommandNowValidated | - | an unsafe kubernetes.command such as "kubectl; rm -rf /" errors; a safe kubectl command produces no such error |
| TestCheckKubeCommandDefaultKubectlUnvalidated | - | the zero-value default (spec.DefaultKubeCommand) validates clean |
| TestContextAllowCommandsHonored | - | Context.AllowCommands threads into checkKube and checkContainerTarget: "lima docker"/"lima podman"/"lima kubectl" reject with the zero Approvals, accept with lima approved; "sudo podman" is refused as an escalation wrapper even with other approvals, and Approve refuses sudo with the run-it-under-sudo advice |
| TestCheckContainerCommandUnlistedBinaryRejected | - | docker.command "curl" and podman.command "/tmp/evil" are rejected by the platform allowlist and the bare-name rule, not merely the charset check |
| TestSafeTokenCharset | plain name / empty | accepted |
| TestSafeTokenCharset | embedded tilde / leading tilde / metacharacter / $ / backslash / space | refused: execguard.UnsafeChars plus '~' |
| TestSafeTokenCharset | unicode space / next line / DEL / C1 as UTF-8 / zero-width space / invalid UTF-8 | refused: whitespace over all of Unicode, control (C0, DEL, C1) and invisible characters, and invalid UTF-8 |
| TestSafeHostPathCharset | unix path / windows path / 8.3 short name | accepted: a host path admits a backslash and a tilde |
| TestSafeHostPathCharset | space / metacharacter / quote / C1 as UTF-8 / zero-width space / invalid UTF-8 | refused, the same as SafeToken |
| TestSafeHostPathAllowsWindowsShortNames | RUNNER~1 / PROGRA~1 / ~/certs | a tilde is legal in a host path: 8.3 short names are real directories an operator cannot rename, and no sink expands one (argv only, no shell; systemd does not expand in a unit directive; ordinary inside a compose scalar) |
| TestSafeHostPathAllowsWindowsShortNames | space / newline / $ / ; / \| / * / () / # / ! / backtick | every other metacharacter is still refused -- the tilde is the only concession, and only for paths |
| TestSafeToken | kubectl | SafeToken returns true |
| TestSafeToken | docker | SafeToken returns true |
| TestSafeToken | --context=prod | SafeToken returns true |
| TestSafeToken | /usr/local/bin/kubectl | SafeToken returns true |
| TestSafeToken | --namespace=solace-connectors | SafeToken returns true |
| TestSafeToken | --server=https://api.k8s.local:6443 | SafeToken returns true |
| TestSafeToken | a b | SafeToken returns false |
| TestSafeToken | a;b | SafeToken returns false |
| TestSafeToken | a$b | SafeToken returns false |
| TestSafeToken | a\`b | SafeToken returns false |
| TestSafeToken | a\b | SafeToken returns false |
| TestSafeToken | a'b | SafeToken returns false |
| TestSafeToken | a"b | SafeToken returns false |
| TestSafeToken | a\|b | SafeToken returns false |
| TestSafeToken | a&b | SafeToken returns false |
| TestSafeToken | a>b | SafeToken returns false |
| TestSafeToken | a<b | SafeToken returns false |
| TestSafeToken | a(b) | SafeToken returns false |
| TestSafeToken | a*b | SafeToken returns false |
| TestSafeToken | a?b | SafeToken returns false |
| TestSafeToken | a#b | SafeToken returns false |
| TestSafeToken | a!b | SafeToken returns false |
| TestSafeToken | a\x00b | SafeToken returns false (null byte) |
| TestSafeToken | a\x7fb | SafeToken returns false (0x7f char) |
| TestSafeToken | a[b | SafeToken returns false |
| TestSafeToken | a]b | SafeToken returns false |
| TestSafeToken | a{b | SafeToken returns false |
| TestSafeToken | a}b | SafeToken returns false |
| TestSafeToken | a~b | SafeToken returns false |
| TestSafeToken | empty string | SafeToken("") returns true, pinned documented exported-API behavior |
| TestSafeActuatorUser | 4 accepted names | letters, digits, '.', '-' and '_' pass, and every accepted name also satisfies SafeToken |
| TestSafeActuatorUser | 6 SafeToken-permitted names | '/', ':', '=', ',', '+' and '@' are rejected here even though SafeToken allows them, since the name reaches a sed address |
| TestSafeActuatorUser | 7 shell-unsafe names, empty | quotes, whitespace, backslash, '$', '*', '[' and the empty string are rejected |
| TestConnectionDefinitionValidation | - | connection with dest set errors must not define queue/topic; incomplete mq connection errors missing 'queue-manager' |
| TestCheckContainerNameRejected | ../evil | rejected for both docker.name and podman.name |
| TestCheckContainerNameRejected | Bad_Name | rejected for both docker.name and podman.name |
| TestCheckContainerNameRejected | valid-default-name | solmq-connector accepted with no docker.name error |
| TestDockerPodmanTLSNeedsNoStoresOptIn | docker / podman | a workflow selecting a store with no stores: block warns about nothing, since every selected store is mounted (bind mounts on docker, the secret store on podman) |
| TestContainerStorePathsGatedPerPlatform | docker | an unsafe character in tls.stores.ts.existing.pkcs12, or a spaced directory, is rejected with no stores: block present, since docker bind-mounts the path itself |
| TestContainerStorePathsGatedPerPlatform | podman | neither the path nor the file name reaches the unit, which mounts each store under its name, so `$(evil)`, a spaced directory, ',', '%' and a space all pass |
| TestContainerStorePathsGatedPerPlatform | kubernetes | the same path is not gated for kubernetes, which embeds the store content in a Secret rather than naming a host path |
| TestRunAsCheckedOnEveryPlatform | kubernetes / docker / podman | the default, unset, image and a non-root id are quiet; a value outside the grammar is an error naming the section and the reason; uid 0 warns it runs as root, adding that it is host root on docker and podman, which use no user namespace |
| TestPodmanRunAsBeyondTheSubuidRangeWarns | 65536 / 65537 / gid 70000 / image / no Lint / docker / kubernetes | under Lint (validate) an id past 65536 warns on podman, which deployed rootless maps container ids 1..65536 into a stock 65536-wide subordinate range; 65536 fits, the gid counts, image is quiet, generate and deploy (no Lint) stay quiet, and docker and kubernetes never warn |
| TestStoresMayShareAFileName | shared file name | ./trust/store.p12 and ./key/store.p12 give no store finding on docker or podman |
| TestStoresMayShareAFileName | one file as both | the same file under two store names gives no store finding on docker or podman |
| TestStoreNamesHeldToGrammar | - | store names are lowercase letters and digits joined by single hyphens, at most 63 characters: mq-ca, solace2, a and 63 characters pass; 64 characters, upper case, an underscore, a double, leading or trailing hyphen and a dot each error |
| TestSelectsStores | solace tcps selecting | a tcps:// solace side selecting a store mounts it |
| TestSelectsStores | solace tcps selecting nothing | TLS without a selection mounts nothing |
| TestSelectsStores | mq tls selecting | an mq tls: true side selecting a store mounts it |
| TestSelectsStores | mq without tls selecting | a selection without TLS (an error elsewhere) mounts nothing |
| TestSelectsStores | plain tcp solace selecting | a selection on a tcp:// host (an error elsewhere) mounts nothing |
| TestSelectsStores | inline leader session selecting | the rendered inline management session selecting a store mounts it |
| TestSelectsStores | conn-ref leader session selecting | so does the connection the management session conn-refs |
| TestCheckContainerRestartUnsafe | newline in restart | docker.restart is rejected; image and timezone are top-level keys, covered by their own per-platform-rejection and charset tests |
| TestCheckContainerRestartUnsafe | realistic value | on-failure:5 is accepted |
| TestRestartIsEachEnginesOwnWord | docker | compose's no, always, unless-stopped, on-failure and on-failure:5 pass; on-abnormal, on-failure:0 and an unknown word are not a compose restart policy |
| TestRestartIsEachEnginesOwnWord | podman | systemd's seven Restart= words pass; unless-stopped is an error naming always, on-failure:5 one naming on-failure, and an unknown word lists systemd's words |
| TestCheckContainerHostPathsUnsafe | newline in tls.stores.ts.existing.pkcs12 | docker's bind-mounted store path rejected |
| TestCheckContainerHostPathsUnsafe | space in libs.dir | podman.libs.dir rejected |
| TestCheckContainerHostPathsUnsafe | windows paths | `C:\certs\...` store paths and `C:\libs` accepted (backslash and colon permitted) |
| TestCheckKubeSecretNames | cred existing bad | non-DNS-1123 credentials existing rejected |
| TestCheckKubeSecretNames | created Secrets | create with no name at all produces no name error -- the name is derived |
| TestCheckKubeSecretsCreateXorExisting | credentials both set | rejected: Render would take the create branch and emit a Secret doc over the object existing names |
| TestCheckKubeSecretsCreateXorExisting | credentials neither set | rejected: a present block must choose, or the SecretsDir mount silently disappears |
| TestCheckKubeSecretsCreateXorExisting | create only / existing only / blocks omitted | all three accepted -- omitting a block stays the way to say "none" |
| TestCheckLibsNFSFields | newline in nfs.server | rejected against the host charset |
| TestCheckLibsNFSFields | newline in nfs.path | rejected against the host-path charset |
| TestCheckLibsNFSFields | valid server and path | nfs1.corp.example and /solace-libs accepted |
| TestKubeQuantitiesAndDownloadImage | resources.cpu / resources.memory / libs.pvc.create.storage | 1, 500m, 0.5, .5, 1., 512Mi, 1Gi, 2Ti, 100k, 1e3 and 5E-1 pass; a newline injection, "1 Gi", "1GB", a sign, a bare unit, a trailing ';' and an unexpanded ${MEMORY} are each refused as not a Kubernetes quantity |
| TestKubeQuantitiesAndDownloadImage | libs.download.image | busybox:1.37 and a registry:port/path@sha256 reference pass; a newline injection, a space and a ';' are refused as unsafe characters |
| TestPasswordConflictOnSameBinder | differing passwords | same MQ tuple with two passwords errors conflicting password for the same binder |
| TestPasswordConflictOnSameBinder | identical passwords | same tuple sharing one password passes |
| TestPasswordConflictOnSameBinder | distinct tuples | different queue-manager means different binders, so passwords may differ |
| TestPasswordConflictSolaceSide | - | the solace branch keys on client-password and errors on a conflict |
| TestRemovedDefaultsKeysRejected | - | a security.enabled value (true or false) errors naming security.enabled as not configurable, a management.exposure value errors naming management.exposure as not configurable, and neither key set validates clean (the third such key, leader-election.solace, is covered by TestLeaderElectionSolaceKeyRenamed) |
| TestStatusUserReservedName | - | a security.users entry named spec.StatusUserName errors reserved, naming security.users[1].name; a differently-named user does not collide |
| TestSecurityUserRoles | admin / unknown-but-well-formed / several / empty / whitespace-only / shell metacharacter / embedded space / no roles | roles are checked for usability, not against an allowlist: a well-formed unrecognized role passes, an empty or whitespace-only entry errors naming both indices, an unsafe-charset entry errors, and omitting roles entirely stays clean. Also pins both error texts verbatim, since the generator page's JS validator mirrors them word for word |
| TestStatusUserPasswordEnvCharset | nil Env / unset / empty / valid value | none trip the SECURITY_USER_SOLMQ_STATUS_PASSWORD charset error |
| TestStatusUserPasswordEnvCharset | space / double quote / single quote / backslash / dollar-brace / control char / non-ASCII byte | each errors the charset check, and the error text never echoes the secret value |
| TestCheckSyslogRunsForEveryPlatform | docker / podman / config only | syslog is a top-level key, so it is validated whichever platform is generated |
| TestKubernetesLoggingIsRetired | - | a `kubernetes.logging` block is rejected by name, and the error names the top-level `logging:` block to use instead; ParseEnv decodes non-strict, so without this check it would be dropped in silence and the instance would come up with no syslog and no diagnostic |
| TestCredentialsNotWiredIsAnError | - | a config referencing credentials with kubernetes.secrets.credentials omitted is an error naming the mount path and the create: true / existing: fix; on a teardown it is a warning |
| TestCredentialsWiredNoError | create / existing | either way of wiring kubernetes.secrets.credentials clears the error |
| TestNoCredentialsNoError | - | a config whose connections need no authentication is not required to carry a Secret it does not need |
| TestCredentialsFoundOutsideAWorkflowSide | a management account / a selected store's password / an unselected store's password / a conn-ref leader session no workflow uses / a session standalone never renders | a management account, a selected store's password and a rendered leader session's account are credentials too and trip the error; an unselected store's password and a session standalone never renders do not |
| TestLibsPVNameLengthIsCapped | - | a 30-char namespace and a 30-char deployment.name derive a PV name (`<namespace>-<name>-libs-pv`) that exceeds the 63-char DNS-1123 limit, which errors naming both fields; shortening both to fit is not flagged, so the check cannot simply always fire |
| TestRetiredCreateNamesReportedOnlyByValidate | - | the three retired name keys (credentials create.name, image-pull.name with create, libs.pvc.create.name) produce no error on a generate/deploy run and exactly one each under Lint, naming the derived object, the ignored value and how to remove the key |
| TestRetiredCreateNamesNeedAKey | - | a created object with no name key, and an image-pull name without create (a reference), are not retired keys |
| TestDerivedNamesMustFitALabel | - | the longest name derived from deployment.name that the config actually builds (here `<name>-credentials`) must fit a DNS-1123 label, naming it; a referenced Secret derives nothing |
| TestTransformHeadersValidBlockPasses | - | a well-formed transform-headers block raises no error and no warning |
| TestTransformHeadersShapeErrors | not a mapping / no expressions / expressions as a list / an expression that is not a string / a header set twice | each shape the connector would not apply is an error on every run |
| TestTransformHeadersWarnings | - | a key beside expressions (a likely typo, passed through) and an empty expressions mapping are warnings, not errors |
| TestMisplacedTransformsAreErrors | - | a misplaced transform or transform-headers is told to move to the top level of its file, the legacy transform-payload section (either spelling) how payload transforms are written now, any other transform key that it is not a key (naming both valid spellings), and a transform in env.yaml that it belongs in a workflow file -- each an error naming its file and path |
| TestTransformLegitimateBlocksStayQuiet | header copy / payload mapping / IBM MQ migration example / content types alone / placeholders / empty payload values | the transform blocks Solace documents -- variables, functions, the dynamic destination header, Solace's own enabled: true (passed through unchecked for now) -- and ${...} placeholders raise nothing, even under Lint, and a top-level transform is not misplaced |
| TestTransformShapeErrors | not a mapping / a list / expressions as a mapping / as a scalar / a scalar item / an item without transform / a later item / a non-scalar expression / an empty expression / an unquoted expression cut at its # / a payload block that is not a mapping / a content-type that is not one value / a key set twice / a content-type set twice | each shape the connector would not apply is an error on every run, saying what was found and how to write it; the transform-headers shape carried over under transform gets the migration hint, and an unquoted expression YAML cut off at " #" is caught by the "=" it is left ending in |
| TestTransformWarnings | an unknown key / in a payload block / in an item / an undocumented content type / xml / empty expressions / an empty block / enabled alone | what is passed through but probably not meant is a warning, never an error, and enabled draws no warning of its own |
| TestTransformAndTransformHeadersCannotBeCombined | - | both blocks in one file is an error on every run quoting Solace's rule, each block's own shape is still checked beside it, and either alone is fine |
| TestTransformHeadersDeprecatedOnlyUnderLint | - | transform-headers draws no deprecation notice on a generate or deploy run and exactly one under Lint, naming the timeline and section 6.7; a transform-only file draws none, and both blocks still draw it |
| TestRetiredPerPlatformImageRejected | kubernetes / docker / podman | each per-platform image key errors, and the message names the top-level image: block to use instead |
| TestImageBlockRequired | absent / no name / no tag / unsafe repo, name, tag | the top-level block is required once a platform is in play, tag included (an untagged image resolves to :latest and pins nothing), and the fields that reach an argv are charset-checked |
| TestImageBlockRequired | bad pass-env name / either credential set both ways | the registry account (`user`/`pass`) goes through the shared checkCred, so it gets the same literal-xor-env rule and variable-name check as every other credential |
| TestImageNotRequiredWithoutAPlatform | - | `generate config` renders application.yml alone and pulls nothing, so no image is demanded |
| TestImagePullSecretChecks | name alone | referencing a Secret requires no registry credentials at all |
| TestImagePullSecretChecks | name required / DNS-1123 | a referenced Secret's name is required and held to the label rule the cluster would apply |
| TestImagePullSecretChecks | create without credentials | create errors unless the registry account is set, in either the literal or the -env form |
| TestImagePullSecretChecks | create, variable unset / set | an unset variable warns rather than errors, so a config can be linted without the deploy secrets |
| TestRetiredPerPlatformTimezoneRejected | kubernetes / docker / podman | each per-platform timezone key errors and names the top-level timezone: key |
| TestJavaOptionsAcceptsJVMSyntax | - | spaces, the '*' and ':' of -Xlog, an agent string, a '%p' error-file path, an @argfile, a '#' inside one option and a literal block's line breaks all pass, on both sub-keys |
| TestJavaOptionsRejectsWhatThePlatformsReadDifferently | leftover variable reference / double quote / single quote / backslash / backtick / control character / a comment inside a block / a block that opens with a comment / a C1 control character / an invisible formatting character (U+200B) / invalid UTF-8 | each is exactly one error naming the key, the variable it feeds and the character, with what to do instead, on both sub-keys |
| TestJavaOptionsCheckedOnlyWhenAPlatformIsInPlay | - | a config-only run does not check java-options; a kubernetes, docker or podman run does |
| TestTopLevelTimezoneUnsafe | unsafe / realistic | the top-level timezone keeps the charset gate the per-platform key had, and an empty value is not an error |
| TestValidatePostureInlineSide | false warns | an inline tcps side's SSL_VALIDATE_CERTIFICATE: false warns, naming the host |
| TestValidatePostureInlineSide | false in any case warns | FALSE is the same literal |
| TestValidatePostureInlineSide | true is the default, silent | an explicit true reports nothing |
| TestValidatePostureInlineSide | non-literal is unknown | a value other than true/false (no) warns that the posture cannot be reported |
| TestValidatePostureInlineSide | another spelling | ssl_validate_certificate warns that the Solace API likely ignores it |
| TestValidatePostureInlineSide | tcp:// host is silent | no TLS, no posture finding |
| TestValidatePostureConnectionWarnsOnceWithReach | - | a connection warns once in env.yaml, never per conn-ref side, naming the workflow sides it reaches and the leader-election session |
| TestValidatePostureBinderConflict | inline sides disagree | two inline sides on one binder stating false and true are an error at the later file |
| TestValidatePostureBinderConflict | conn-ref side is reported by its connection | a quoted 'false' against a connection's false is a conflict, and the earlier value is located as env.yaml connections.edge |
| TestValidatePostureBinderConflict | stated beside absent is not a conflict | no error; the warning counts both sides the false applies to |
| TestValidatePostureSession | standalone session is not a site | an unrendered inline session reports nothing |
| TestValidatePostureSession | session off, sides validate | the session's false warns, and its divergence from the workflow sides on its tuple warns |
| TestValidatePostureSession | sides off, session validates | the reverse divergence warns too |
| TestStoreDefinitions | no block / no source | a store with neither create: nor existing: is an error naming both, and the Secret form |
| TestStoreDefinitions | both | create: and existing.pkcs12 together is an error: one host-side source per store |
| TestStoreDefinitions | existing naming nothing | existing: with neither pkcs12 nor secret is an error |
| TestStoreDefinitions | a secret without a name / a secret name no Secret can have / a key no Secret can hold | existing.secret needs a DNS-1123 name and a valid data key |
| TestStoreDefinitions | created without a password | a created store needs its own password: it is encrypted with it |
| TestStoreDefinitions | created from nothing | a create: naming no source is an error |
| TestStoreDefinitions | certs and a key / a jks and a key | trusted certificates and a key in one store are an error: a truststore or a keystore |
| TestStoreDefinitions | two keys | more than one key entry is an error: one store per identity |
| TestStoreDefinitions | a key entry without its key | a key entry needs cert and key |
| TestStoreDefinitions | an empty chain file / an empty certs entry / an empty named entry | an empty path is an error naming its position |
| TestStoreDefinitions | an empty alias / an alias twice | a named map needs every alias, unique ignoring case |
| TestStoreDefinitions | well-formed | existing, a certs list, a named map, one key entry with its chain, a jks alone, certs with a jks, and a Secret alone, beside a file or beside sources are all accepted |
| TestCreatedStoreSelections | a key store as truststore | a store created from a key selected as truststore is an error |
| TestCreatedStoreSelections | a certificate store as keystore | a store created from certificates selected as keystore is an error |
| TestCreatedStoreSelections | a key-alias on a created keystore | key-alias on a created keystore is an error: its one key needs no alias (Java names it 1) |
| TestCreatedStoreSelections | fitting roles | roles that fit pass, and a jks-only store's role is left to deploy |
| TestUnusedStoreWarns | - | a store nothing selects warns that it is not mounted; one a connection selects (even an unused connection) does not |

## internal/examples

Write the shipped starter files (create/skip/force), prove they generate config, and pin the workflow copies shared with the golden fixtures.

Tests: [examples_test.go](../internal/examples/examples_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestWriteCreatesSkipsForces | first-write | first write creates 5 files (4 workflows + env.yaml), 0 skipped |
| TestWriteCreatesSkipsForces | second-write-no-force | second write with no force skips all previously written files |
| TestWriteCreatesSkipsForces | force-rewrite | force write rewrites all files, restoring workflow-0.yaml to embedded original content over junk |
| TestWriteMkdirError | - | Write returns error when target dir path is under a regular file |
| TestShippedExamplesGenerateConfig | - | embedded example set written to disk generates config via gen.Config with no errors and at least one non-empty rendered application.yml |
| TestEnvExtraKeysExampleValidWhenUncommented | - | uncommenting the other-key examples shipped in env.yaml (two connection lines and the mq-defaults block) still generates with no finding about them, and each key lands under its binder |
| TestEnvCreatedStoreExamplesValidWhenUncommented | - | uncommenting the store examples shipped in env.yaml (a certs list, a certs map, a key entry, a jks, a file beside a kubernetes Secret) still validates with no error (real PKCS12 stores sit where the selected stores are named, their passwords set, since validate reads and opens them as deploy does), each store read and reported only as not selected |
| TestWorkflow0TransformExampleValidWhenUncommented | - | uncommenting the transform example shipped in workflow-0.yaml still generates, with no transform finding and the block rendered under workflow 0, so the example a new user copies cannot rot |
| TestWorkflowExamplesMatchGoldenSpecs | workflow-0..3 | the four workflow files are byte-identical to testdata/golden/specs (go:embed cannot share one copy); env.yaml is excluded, it diverges on purpose |

## internal/gen

Orchestrate parse -> validate -> consolidate -> render, resolve credentials/stores, and assert the byte-for-byte golden fixtures.

Tests: [gen_extra_test.go](../internal/gen/gen_extra_test.go), [gen_runas_test.go](../internal/gen/gen_runas_test.go), [golden_test.go](../internal/gen/golden_test.go), [htmlgolden_test.go](../internal/gen/htmlgolden_test.go), [imagepull_test.go](../internal/gen/imagepull_test.go), [stores_test.go](../internal/gen/stores_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless | - | parse() expands host/msg-vpn from Lookup, leaves client-password-env verbatim, and returns exactly one warning naming TYPO for the unset defaultless conn-name variable |
| TestResolveStores | two selected stores with ReadFile | resolves 2 store files keyed by store name, in Model.Stores order |
| TestResolveStores | no ReadFile provided | missing ReadFile returns error |
| TestResolveStores | ReadFile returns error | read error propagates, naming tls.stores.<name>.existing.pkcs12 and the path |
| TestResolveStores | no stores selected | nil stores yields 0, no error |
| TestResolveStores | over the apply limit | a set over the 256 KiB kubectl apply carries, once base64-encoded, is an error naming each store's size and existing.secret as the way out; a set under it passes |
| TestResolveStores | a cluster-held store | a store kept in an operator's Secret is never read |
| TestToIssues | - | toIssues wraps each string into an Issue carrying it as Msg |
| TestPodmanFileSecretNames | - | the three document secrets an instance can own (application.yml, status-script, logback-spring.xml) and the store secret name <name>-tls-<store>, none of whose suffixes could also be a credential's stable name |
| TestPodmanUnitStoreSecrets | - | the store secrets a previous unit mounts are its own Secret= lines with the container's store prefix -- not a credential, a document, an empty store name or another instance's store; no unit gives none |
| TestResolvePodmanFiles | document / binary store / 511999 bytes | a document passes through as rendered, a store is read byte for byte (NUL, CR and non-UTF-8 included), and 511999 bytes is accepted |
| TestResolvePodmanFiles | missing / no file access / empty / 512000 bytes / empty document | each fails before anything is loaded, naming the field and path (wrapping the read error) and never the content |
| TestBuildStore | - | a created store is read through the resolver and built with the password from the environment; DecodeTrustStore reads it, and the expiring certificate comes back as an env.yaml warning |
| TestBuildStoreReadsEverySourceByPosition | key entry with chain | a keystore built from cert, chain and key decodes with its leaf and chain |
| TestBuildStoreReadsEverySourceByPosition | named certs / key cert / key / chain / jks | a missing source is named by its env.yaml position: create.certs.<alias>, keys[0].cert, keys[0].key, keys[0].chain[0], create.jks |
| TestBuildStoreRejects | an unset / an empty password variable / no password / no environment access | each fails naming the variable or the missing password, never a value |
| TestBuildStoreRejects | an unexpanded variable / no file access | a ${...} left in a path, or no file access, fails before anything is read |
| TestBuildStoreRejects | a source the builder refuses | the builder's error propagates, naming the position and file |
| TestBuildStoreRejects | a JKS key selected as truststore / a JKS truststore selected as keystore | a jks-only store whose content does not fit the role it is selected in fails at build |
| TestBuildStoreRejects | not a created store | an existing store has nothing to build |
| TestCreatedStoresOnEachPlatform | kubernetes | the stores Secret carries the built truststore, and an unchanged store renders the same manifest twice |
| TestCreatedStoresOnEachPlatform | kubernetes, operator Secret | a created store also held in an operator's Secret is not built on kubernetes: nothing is read, no tool stores Secret rendered, the operator's projected |
| TestCreatedStoresOnEachPlatform | podman | the plan carries a trusted build with no source, and ResolvePodmanFiles builds the secret |
| TestCreatedStoresOnEachPlatform | docker | the plan carries the store as a trusted docker store with no bind mount, the compose mounts its secret, and ResolveDockerStores returns its base64 -- the built truststore -- and a 64-hex digest that does not change while the store does not |
| TestCreatedStoresOnEachPlatform | config | generate config references the created store at its mount path and reads nothing |
| TestCreatedStoreWarningsSurface | kubernetes / podman | a created store's expiry warning reaches the operator on both build paths |
| TestResolveDockerStoresLimits | over the limit | a store whose base64 form would not fit one environment variable fails naming the store, its sizes, the limit and the host |
| TestResolveDockerStoresLimits | a build failure | a store that does not build stops the deploy |
| TestResolveDockerStoresLimits | no selected store | no environment, no warning, no error |
| TestResolveDockerStoresDigestsExistingStores | - | an existing store's bytes, read from its mount path, enter the digest: unchanged keeps it, replaced in place changes it; one its password does not open stops the deploy, one this user cannot read warns and stays out, a missing one is the host checks' to report |
| TestEnvVarMax | windows / linux / darwin | 32767 on Windows, 131072 elsewhere |
| TestValidateReadsWhatDeployReads | a readable existing store | kubernetes and podman read it, nothing is reported |
| TestValidateReadsWhatDeployReads | a missing existing store, kubernetes / podman | each reports the unreadable store, naming its key and path |
| TestValidateReadsWhatDeployReads | missing under two platforms | read once, reported once |
| TestValidateReadsWhatDeployReads | two bad stores | both reported in the same run |
| TestValidateReadsWhatDeployReads | a docker existing store, missing | read as deploy reads it for the digest; missing is the host checks' to report, not an error here |
| TestValidateReadsWhatDeployReads | a docker existing store, opened | a JKS named as pkcs12 is refused for docker too |
| TestValidateReadsWhatDeployReads | a created store that does not build | the build's error is reported |
| TestValidateReadsWhatDeployReads | unset password | the store is still built with a stand-in, so a bad source is reported, beside the unset-variable warning |
| TestValidateReadsWhatDeployReads | unset password, a JKS | not opened, but read: a missing JKS file is reported beside the unset-variable warning |
| TestValidateReadsWhatDeployReads | unset password, a JKS that is not one | a PKCS12 named as create.jks is reported without the password |
| TestValidateReadsWhatDeployReads | unset password, a JKS beside certs | the certs are still built, so a bad one is reported |
| TestValidateReadsWhatDeployReads | a password variable set but empty | deploy's set-but-empty error |
| TestValidateReadsWhatDeployReads | a JKS named as a PKCS12 / a store its password does not open | an existing store is opened with its password: each is an error |
| TestValidateReadsWhatDeployReads | a store in an operator's Secret | not read or built on kubernetes |
| TestValidateReadsWhatDeployReads | a store in an operator's Secret, podman | still read for podman, which needs its file |
| TestValidateReadsWhatDeployReads | an expiring certificate | warned once across kubernetes, podman and docker |
| TestValidateReadsWhatDeployReads | over the kubernetes Secret / a podman secret / a windows environment variable | each platform's size error; the same docker store fits on linux |
| TestValidateReadsWhatDeployReads | every case | a store nothing selects is never read |
| TestValidateStoreBytesWithoutAccess | no file access | an existing store is reported unread, never passed |
| TestValidateStoreBytesWithoutAccess | no environment access / a literal password | a created store is built with the stand-in, or with its literal as written |
| TestKubernetesTeardownResolvesNothing | - | a kubernetes teardown reads no credential and opens no store, names the credentials and stores Secrets alone, warns about a retired stores existing: that deploy refuses |
| TestKubernetesSecretProbes | - | the probes are each operator Secret a selected store names, with every key the volume lists, then an existing credentials Secret with every mount name |
| TestKubeConfigMapSizeError | - | the ConfigMap is held to the 256 KiB kubectl apply carries, counting application.yml, the status script and the logback config; a set that fits passes |
| TestCheckExistingStore | a truststore / a keystore / a store used both ways | each opens with its password: no finding |
| TestCheckExistingStore | the wrong password / the wrong literal | an error naming the store, its file and the password field |
| TestCheckExistingStore | a JKS / a JKS, no password / a JCEKS | an error with the conversion, with or without the password |
| TestCheckExistingStore | a PEM / a certificate renamed .p12 | not a PKCS12: an error, with or without the password |
| TestCheckExistingStore | no password to open it / an empty password variable | only the shape is checked: no finding |
| TestCheckExistingStore | a keystore read as a truststore | opens, but its contents are unreadable here: a warning, not an error |
| TestCheckExistingStoreFromOpenSSL | - | where openssl is on PATH: its keystore opens with no finding; its -nokeys truststore warns that it lacks Java's trusted-certificate attribute |
| TestResolvePodmanFilesOpensExistingStores | - | podman deploy opens an existing store before loading it: a good one loads, one its password does not open stops the deploy |
| TestDockerStorePlaceholders | - | remove's placeholders cover each created store's secret, named _GEN_TLS_<STORE>_STORE, and the digest -- never an existing store's -- existing stores alone need only the digest's, and a plan with no store selected needs none |
| TestTargetMounts | selected stores + libs | 2 store mounts at MountDir/<store name> -- a created store, having no host file, is skipped --, each named tls.stores.<name>.existing.pkcs12, the keystore Private (carried into HostMount), and a libs mount whose source is the resolved abs host dir and whose target is the fixed default libs path |
| TestTargetMounts | no stores, no libs | yields nil,nil -- no selected store is the only way to get no store mounts |
| TestResolveCredentials | nil refs | no kvs, no error |
| TestResolveCredentials | literal + -env mix | literal ref passes through, -env ref reads from the resolver's environment |
| TestResolveCredentials | unset -env variable | fails loud naming the stable secret and the variable, never a value |
| TestResolveCredentials | set but empty -env variable | fails loud naming the stable secret and saying the variable is set but empty |
| TestResolveCredentials | no environment access | fails loud rather than silently resolving to empty |
| TestConfigRejectsSecretNameConflict | security.users "ops.1" and "ops-1" | Config renders nothing and errors naming the contested key *and both claiming positions*, rather than emit a config where one credential silently takes the other's password |
| TestConfigRejectsSecretNameConflict | same spec through Validate | the collision is caught while linting too, not only at generate/deploy -- names are assigned in consolidate, so Validate builds to see them |
| TestValidateCleanSpecStillPasses | no collision | the build call Validate now makes adds no errors of its own, and consolidate's warnings do not leak into validate output |
| TestConfigNumbersWorkflowsInLsOrder | - | gen.parse numbers an explicit file list in the same listing order the folder scan uses: 10.yaml takes workflow 0 ahead of 2.yaml, whatever order the files arrived in |
| TestConfigCarriesExtraKeysEndToEnd | - | an inline side's other keys land under its own binder after the mq-defaults entry as siblings of the tool's keys, a connection's other key reaches the binder a conn-ref side builds from it and the management session, and a ${...} inside one reaches application.yml as typed |
| TestConfigRejectsExtraKeysBesideConnRef | - | another key beside conn-ref fails the render with the strict conn-ref error |
| TestConfigWarnsOnANearMissKeyButStillRenders | - | a probable typo warns "did you mean" and still passes through to a rendered config |
| TestConfigWorkflowCap | 21 workflows | Config produces no output and one error naming the count, the 20 cap, and the split-into-folders remedy |
| TestConfigWorkflowCap | 20 workflows | exactly at the cap does not error |
| TestGenerateKubernetesWorkflowCap | 21 workflows | GenerateKubernetes produces no manifest and the same workflow-cap error |
| TestGenerateKubernetesOmitNamespaceAndWritten | - | KubeOpts.OmitNamespace renders a deploy without the Namespace document while the default (generate) keeps it first, and Written receives the kinds the manifest writes for the probes |
| TestConfigCarriesSecurityUserRoles | - | end-to-end: a roles-bearing env.yaml validates clean and its role reaches the rendered application.yml, while the reserved account still renders none |
| TestConfigNoSecretsLeak | - | every rendered password is a ${STABLE} placeholder except the one permitted literal: the reserved spec.StatusUserName account, whose value the override pins for the test |
| TestGenerateDockerBasics | - | generates non-empty compose opening with the defaulted, quoted `name: "solace-ibmmq-connectors"` project line and containing the image; all four credential positions render as top-level environment-provider secrets, never inlined as values, and each ${STABLE} placeholder in application.yml is doubled so compose cannot interpolate the value in |
| TestArtifactsStateTheirPosture | default / 1234:5678 / 0 / image | kubernetes, compose and quadlet each carry the restrictions whatever the run-as (compose and quadlet through artifacttest.AssertComposePosture and AssertQuadletPosture: the block exactly once, quadlet inside [Container]), write its ids (fsGroup = the gid on kubernetes), add runAsNonRoot only for a non-root uid, write no ids for image, and harden the libs-download init container too |
| TestArtifactsCarryNoWideningTokens | default / 1234:5678 / 0 / image | no artifact carries a spelling that would widen the container's privileges: privileged, privilege escalation, added capabilities, unconfined seccomp or SELinux, host devices or sysctls, or the host's namespaces and network -- the shared artifacttest list plus the kubernetes spellings, run over every run-as and over a compose file and a unit with every optional block set |
| TestQuadletKeysAreWithinTheFloor | each run-as / every optional block | every [Container] key of every unit shape is known and within podman 4.5, the floor the tool documents (artifacttest.AssertQuadletKeys) |
| TestSecretTargetsAreAbsolutePaths | each run-as / every optional block | every service secret of every compose file and every Secret= line of every unit shape mounts at an absolute target (artifacttest.AssertSecretTargets) |
| TestRunAsOfRefusesWhatValidateRefuses | - | the renderers' own parse of a run-as refuses what validate refuses, as an issue naming the section, never as the zero identity, which is root |
| TestPlansCarryTheExpandedHostPathsTheyMount | docker stores+libs / podman / docker without mounts | each plan lists the host paths its artifact bind-mounts, named by env.yaml key, after ${VAR} expansion and resolution against env.yaml: docker the selected stores (the keystore marked private) and libs.dir, podman libs.dir alone (its stores come from the secret store), and a docker spec with neither nothing |
| TestGeneratePodmanQuadlet | - | produces the `<name>.container` unit with the application.yml and status-script file secrets, the service name and 4 credentials, each mounted from podman's store by its namespaced name at an absolute target under the secrets mount |
| TestGenerateJavaOptionsReachEveryPlatform | - | a java-options block written as a >- folded block with ${VAR} references reaches the kubernetes manifest, the compose file and the quadlet unit as the same single line, and an unsafe value stops every platform's generation |
| TestGeneratePodmanRejectsModeKey | run / quadlet | podman.mode is rejected at generate for either value |
| TestGeneratePodmanNoModeKeyIsClean | - | an omitted mode: generates cleanly, guarding against applyPodmanDefaults ever defaulting the key, which would trip the rejection for every section |
| TestResolveStatusPasswordFixedRand | - | with nothing to derive from, a fixed Rand hook yields the exact expected hex literal (16 bytes -> 32 lowercase hex chars), never a randomized value that would make the test flaky |
| TestResolveStatusPasswordEnvOverride | - | a set, non-empty spec.StatusUserPasswordEnvVar is used verbatim and Rand is never consulted |
| TestResolveStatusPasswordEmptyEnvFallsBackToRand | - | an empty override is treated as unset, falling back to Rand rather than returning "" |
| TestResolveStatusPasswordRandError | - | a Rand failure surfaces as an actionable error naming the underlying cause, never a predictable fallback password |
| TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName | - | with credentials the password is 32 lowercase hex characters derived by HMAC and Rand is never drawn: the same inputs give the same value, credential order does not matter, and a changed value, name, key, field boundary, added credential or the config render's empty name each give a different one |
| TestResolveStatusPasswordUsesEnvCredentialsAndFallsBackWhenUnresolvable | - | an -env credential's resolved value keys the password (rotating it rotates it); an unset variable, or no environment access, falls back to the random path |
| TestResolveStatusPasswordNoMaterialIsRandomAndOverrideWins | - | no credentials at all is random, a Rand failure still errors, and the SECURITY_USER_SOLMQ_STATUS_PASSWORD override beats the derivation |
| TestStatusPasswordIsStableAcrossRenders | - | through GenerateDocker the same spec renders the same compose file with no random draw, and a changed credential changes it |
| TestStatusPasswordStaysRandomOffDockerAndPodman | - | generate config and kubernetes draw Rand even with credentials to derive from, since a kubernetes ConfigMap may be readable where the credentials are not |
| TestConfigStatusPasswordRandErrorNoOutput | - | with -env credentials that cannot be resolved there is nothing to derive from, so the random path runs and its failure through Config is a hard error with no output |
| TestGenerateKubernetesRefusesAnInjectedQuantity | memory with a newline | generate refuses it as not a Kubernetes quantity and the injected key never reaches the manifest |
| TestGenerateKubernetesCarriesStatusScript | - | the ConfigMap gets a "status: \|" key carrying the rendered script, addressed to spec.StatusUserName on the resolved management port |
| TestGenerateDockerCarriesStatusScript | - | compose gets a second top-level config (`<name>-status`) inlining the rendered script, mounted at statusscript.ContainerPath |
| TestGeneratePodmanCarriesStatusScript | - | the plan carries `<name>-application.yml` and `<name>-status-script` with their rendered content, the unit mounts each from podman's secret store at its fixed path, and no Volume= appears without libs |
| TestGeneratePodmanCarriesLogbackOnlyWithSyslog | no syslog | neither a `<name>-logback-spring.xml` file on the plan nor any logback mount in the unit |
| TestGeneratePodmanCarriesLogbackOnlyWithSyslog | syslog | the plan carries the rendered logback-spring.xml as `<name>-logback-spring.xml` and the unit mounts it from the secret store at the image's logback path |
| TestGeneratePodmanNeverReadsTheStores | - | with two selected stores generate never calls ReadFile: the plan names each store by its file, the unit mounts `<name>-tls-ts` / `<name>-tls-ks` at /app/external/classpath/truststores/ts and .../ks, and neither is a Volume= |
| TestStoresSharingAFileNameMountApart | shared file name | ./trust/store.p12 and ./key/store.p12 generate on docker (two bind targets), podman (two Secret= targets) and kubernetes (keys ts and ks, mapped by items) with no error |
| TestStoresSharingAFileNameMountApart | one file as both | the same file under two store names generates on all three, each store at its own path or key |
| TestGeneratePodmanIgnoresBaseDir | - | a base-dir, even a spaced one holding an unset variable, changes neither the unit nor the plan and adds no error or warning outside validate |
| TestGenerateMissingTargetSection | kubernetes | error contains kubernetes target requires a 'kubernetes:' section in env.yaml |
| TestGenerateMissingTargetSection | docker | error contains docker target requires a 'docker:' section in env.yaml |
| TestGenerateMissingTargetSection | podman | error contains podman target requires a 'podman:' section in env.yaml |
| TestSelectedStoreIsTheRequestOnKubernetes | - | a workflow selecting a store with no stores block validates clean, validate reads the store kubernetes carries, and generate renders the stores Secret and its projected source |
| TestGeneratorPageGoldenInSync | - | the golden embedded in solmq-conn-util-generator.html matches testdata/golden/application.yml (regenerate with -update-html-golden) |
| TestGeneratorPageFindingsGoldenInSync | - | the findings embedded in solmq-conn-util-generator.html (id="golden-findings") match what gen.Validate reports for testdata/golden/specs, CRLF and trailing newlines normalized (regenerate with -update-html-golden) |
| TestGeneratorPageKnownKeysInSync | SOLACE_KEYS / MQ_KEYS / SOLACE_MANAGED / MQ_MANAGED / SOLACE_CREDS / MQ_CREDS | the page's hand-copied key lists equal spec.KnownKeys, ToolManagedKeys and CredentialKeys per system, so the validate port cannot drift from the CLI when a key is added |
| TestGeneratorPageSelfTestNormalizesEmptyFindings | - | the page's Self-test normalizes its own findings output the way it normalizes the golden block (trailing newlines trimmed, exactly one added), pinned by matching the normalization pattern in the page source, so the check needs no JS engine -- without the normalization the zero-findings case diffs a bare newline against an empty string and Self-test fails on the page's own fixture |
| TestGoldenConfig | - | generated config output matches testdata/golden/application.yml byte-for-byte, one instance |
| TestGoldenStoreFixtures | - | the golden spec's PKCS12 fixtures open with the passwords the spec gives them, so validate finds nothing about them; -update-fixtures rewrites them from fresh certificates |
| TestGoldenKubernetesCreate | - | generated kubernetes manifests (namespace, configmap incl. status script, secret, stores keyed by store name holding the committed fixtures and projected by items, pv, pvc, deployment with secrets-volume/stores/syslog/libs mounts and le-mode/role labels, service) match golden fixture byte-for-byte |
| TestConfigIsWhatDeployShips | - | generate config's application.yml is byte for byte what podman mounts from its secret store and what the compose file inlines (indented, every $ doubled); the kubernetes goldens embed the same file |
| TestGoldenKubernetesNoSecrets | - | generated manifests with no Secret the tool creates but the stores, and no syslog or libs (namespace, configmap incl. status script, stores Secret, deployment mounting the operator's credentials Secret and the stores volume, service) match golden fixture byte-for-byte |
| TestDockerConfigJSON | private registry / docker hub fallback | the payload is an auths map keyed by registry carrying the account plus the base64 user:password the engines send |
| TestDockerConfigJSONEscapesAwkwardValues | - | a password carrying quotes, a backslash or JSON of its own round-trips as data rather than reshaping the document -- which is why it is marshalled, not concatenated |
| TestResolvePullSecret | reference only | resolves to the given name alone and never reads the registry password |
| TestResolvePullSecret | create | builds the payload from the environment |
| TestResolvePullSecret | both -env / both literal | all four halves resolve: the -env pair reads each variable, and a literal pair needs no environment access at all |
| TestResolvePullSecret | unset user-env | an unset user variable fails naming that variable, not only the password one |
| TestResolvePullSecret | variable unset / no environment access | both fail loudly, naming the variable |
| TestResolvePullSecret | no image block / partial image block | the guard that exists so a caller who skipped validate gets an error rather than a nil dereference |
| TestGenerateKubernetesImagePull | no block / reference / create / create with a retired name | the wiring from config to rendered manifest: nothing, an imagePullSecrets entry alone, or the entry plus a dockerconfigjson Secret named `<deployment>-image-pull` -- with a retired name alongside create ignored rather than rejected -- the integration point TestResolvePullSecret skips |
| TestGenerateKubernetesImagePull | payload and leak check | the rendered payload decodes to the real account, and the registry password appears nowhere else in the manifest |
| TestGenerateKubernetesImagePull | variable unset | create fails the generate with an issue naming the variable |
| TestRetiredCreateNamesDeployButDoNotValidate | - | an env.yaml still naming its credentials Secret and libs claim, with a retired stores create: block, generates cleanly under the derived names, the old names nowhere in the manifest, while Validate reports each old key and the retired block |
| TestConfigRendersTransformHeadersUnderItsWorkflow | - | the second workflow file's transform-headers renders verbatim under solace.connector.workflows.1 and nowhere under 0, and a misspelt transform-header: fails the render instead of being dropped |
| TestConfigRendersTransformUnderItsWorkflow | - | the second workflow file's transform renders verbatim under solace.connector.workflows.1 and nowhere under 0, a ${...} inside an expression is left as written with no warning, and the transform-headers shape under transform, a transform under a side, and both sections in one file each fail the render |
| TestTransformHeadersDeprecationIsLintOnly | - | Validate warns that transform-headers is deprecated while Config renders the same file with no such warning |
| TestValidateReportsValidationOffOnEveryCommand | - | gen.Validate and gen.Config both report a connection's SSL_VALIDATE_CERTIFICATE: false; generate renders it beside an MQ binder validating with its own store, with no tool-written true and no collision warning |

## internal/libs

Resolve the Maven dependency closure for the IBM MQ / syslog jar sets and download the jars to a local directory -- version resolution (including a pinned `--version`), image-aware omission against a connector image's jar list, and the safe download/redirect machinery.

Tests: [libs_test.go](../internal/libs/libs_test.go), [maven_test.go](../internal/libs/maven_test.go), [image_test.go](../internal/libs/image_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestDownloadURLHappyPath | - | a single `--url` download writes the jar under its basename with no Failed/Skipped |
| TestDownloadRejectsNonHTTPSURL | - | a plain http:// `--url` errors before any request or directory creation, Report stays zero-valued |
| TestDownloadWritesJarsWorldReadable | - | a downloaded jar lands 0644, not the 0600 os.CreateTemp gives it, since the connector reads the libs directory as a non-root uid that does not own the file |
| TestDownloadRedirectToHTTPRejected | - | a redirect Location that downgrades to http:// is rejected as a Failure, nothing written |
| TestDownloadRedirectToHTTPSFollowed | - | an https redirect is followed and the file is written under the originally-requested URL's basename, not the redirect target's |
| TestDownloadSkipsExistingWithoutForce | - | an existing file is reported Skipped and no HTTP request is made at all (skip before request) |
| TestDownloadOverwritesWithForce | - | Force true overwrites an existing file with the newly fetched bytes |
| TestDownloadPerArtifactFailureDoesNotBlockOthers | - | one URL 404ing still lets a second URL in the same call succeed and land in Written |
| TestDownloadUncreatableDirIsSystemic | - | a destination directory that cannot be created (parent is a file) is a systemic error, zero Report |
| TestDownloadUnknownSetIsSystemic | - | an unrecognized Set is a systemic error naming the bad value plus both valid names, and creates no directory |
| TestResolveSeedPerSet | mq / syslog | each set maps to its seed coordinate, and the mq seed is asserted *not* to be the javax build (com.ibm.mq.allclient), which cannot satisfy the image's jakarta.jms binder |
| TestValidateFilenameShapeRejections | empty / . / .. / a slash or backslash segment / absolute / non-.jar / control char | each shape is rejected |
| TestValidateFilenameShapeAccepts | - | a plain versioned jar name passes |
| TestFilenameFromEscapedPathRejectsEscapedTraversal | - | a percent-encoded ../ segment is caught after decoding, not waved through because the raw segment looked clean |
| TestFilenameFromEscapedPathAcceptsPlainName | - | a plain path's final segment is returned as the filename |
| TestDownloadByteCapTripLeavesNoTempFile | - | a body exceeding maxArtifactBytes is a Failure and leaves no temp file behind in the destination directory |
| TestDownloadUserinfoNotInOutput | - | a URL carrying userinfo never leaks the username or password into Failure.Name or Failure.Err |
| TestDownloadTooManyRedirectsFails | - | a chain past maxRedirectHops is a Failure naming "too many redirects", not an infinite follow |
| TestDownloadEmptyBodyLeavesNoTempFile | - | a clean 200 OK with a zero-byte body is a Failure, not a silently-accepted empty jar; no temp file is left behind |
| TestDownloadContentLengthMismatchLeavesNoTempFile | - | a body shorter than its own advertised Content-Length is a Failure naming the mismatch, not a truncated jar reported as written |
| TestDownloadSHA1MatchSucceeds | - | a jar whose downloaded body matches its .sha1 sidecar writes normally and reports no Unverified entries |
| TestDownloadSHA1MismatchFailsAndLeavesNoFile | - | a .sha1 sidecar that does not match the downloaded body is a Failure naming sha1, nothing written, no leftover temp file |
| TestDownloadMalformedSHA1SidecarRejected | too short / non-hex / empty / html body | each malformed .sha1 sidecar body is a Failure and nothing is written |
| TestDownloadURLUnverifiedOn404SidecarStillWrites | - | a `--url` download whose .sha1 sidecar 404s still writes the jar, reported in Unverified rather than Failed |
| TestDownloadMavenResolved404SidecarFailsArtifact | - | unlike `--url`, a Maven-resolved artifact (mq/syslog Set) whose .sha1 sidecar 404s is a Failure, not Unverified -- a resolved closure is never written unverified |
| TestDownloadContentLengthUnknownWithGoodSHA1Succeeds | - | a response with no Content-Length still succeeds when its .sha1 sidecar verifies the body |
| TestDownloadContentLengthUnknownWithNoDigestFails | - | a response with neither a Content-Length nor a verifiable .sha1 (sidecar 404s) is a Failure, leaving no temp file |
| TestDownloadOmitsDependencyImageProvidesAtNewerVersion | - | a dependency the image already has at an equal-or-newer version is never requested over HTTP and lands in Report.Omitted naming the jar and the image's version, while the seed itself still downloads |
| TestDownloadEmptyOmitListFileOmitsNothing | - | a supplied `--omit-lib-file` that parses to no entries at all *replaces* the embedded default rather than merging with it, so the whole closure -- seed and dependency -- downloads and Report.Omitted stays empty; OmitListProvenance still names the supplied file |
| TestDownloadCommentsOnlyOmitListFileOmitsNothing | - | an `--omit-lib-file` containing only comments and blank lines omits nothing, same as an empty file |
| TestDownloadNeverOmitsSeedEvenWhenImageClaimsHugeVersion | - | an omit-list entry naming the seed's own artifact at an absurd version never omits the seed -- Download identifies the seed by Coord equality, not by trusting the omit list |
| TestDownloadDownloadsArtifactImageHasOlderVersion | - | an artifact the image has only at an older version still downloads, with nothing in Report.Omitted |
| TestDownloadDownloadsArtifactAbsentFromImage | - | an artifact absent from the image list entirely downloads |
| TestDownloadIncludeProvidedDownloadsEverything | - | `--include-provided` (IncludeProvided true) downloads an artifact the image list would otherwise have satisfied, and Report.Omitted stays empty |
| TestDownloadURLNeverOmittedEvenWhenImageHasIt | - | an explicit `--url` downloads even when the image list already has that exact jar -- omission never applies to `--url` |
| TestDownloadBadOmitLibFilePathIsSystemic | - | an unreadable `--omit-lib-file` path is a systemic error naming the path, with nothing written |
| TestDownloadEmbeddedDefaultListLoadFailureIsSystemic | - | a corrupted built-in omit list (a line exceeding the scanner's token-size ceiling, swapped in through embeddedListFS) is a systemic error naming the "built-in omit list", zero Report, nothing written |
| TestDownloadReportsOmitListProvenance | - | Report.OmitListProvenance names the `--omit-lib-file` path used; a line that fails to split is skipped silently, and a rejected entry no closure artifact asks about produces *no* warning -- the noise fix |
| TestDownloadWarnsWhenRejectedEntryAffectedThisClosure | - | a rejected entry naming an artifact the closure does resolve warns, names that artifact, and the jar is downloaded rather than omitted |
| TestDownloadVersionPinsSeed | - | `--version` pins the seed to that release, and the resulting closure/filename reflect the pinned version, not latest stable |
| TestDownloadVersionRejectsPathEscape | - | a `--version` value containing a path escape is rejected before any network access |
| TestDownloadSetPathAlwaysResolvesEvenWhenFilesExist | - | unlike `--url` (TestDownloadSkipsExistingWithoutForce), the mq/syslog Set path always attempts Maven resolution first -- even with every plausible target file already on disk -- because it cannot know the target filenames without resolving the closure first |
| TestDownloadMQOmitsAgainstTheDeployedLine | 2.14.1 | judged silently against the 2.13.0 list: BouncyCastle and jakarta.jms-api are omitted, while org.json 20251224 downloads beside the client because that image's 20250517 is older |
| TestDownloadMQOmitsAgainstTheDeployedLine | 3.1.0 | judged silently against the 3.1.0 list, which also ships org.json 20251224: only the IBM MQ client downloads |
| TestDownloadMQOmitsAgainstTheDeployedLine | 3.2.0 | the same against the 3.2.0 list: only the client downloads, and the list in effect is the 3.2.0 one |
| TestDownloadSyslogEncoderFollowsConnectorLine | 2.x connector / 3.x connector / none declared / latest / --version | with no `--version`, a 2.x connector (Jackson 2) gets the newest 8.x encoder -- not the metadata's <release> 9.0, nor an 8.x pre-release -- while a 3.x connector, no declared image, or a tag naming no release gets the newest release; Report.SeedChoice names the release and why (the last two also say how to pin), and a pinned `--version` wins with no SeedChoice |
| TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic | - | metadata with no 8.x release for a 2.x connector is a systemic error naming the 8.x line and `--version`, with nothing written -- never a quiet fall back to a 9.x encoder the connector cannot load |
| TestSetNames | - | SetNames() returns the exact ordered [mq, syslog] list the CLI layer gates against |
| TestCompareVersions | patch / lexical-trap / major / prerelease suffix / date-like / short segments / equal / empty | compareVersions orders numeric segments correctly, ranks a pre-release suffix lower, and is antisymmetric under argument swap |
| TestCompareVersions | Final / RELEASE / GA equal the plain release | Maven treats those words as aliases of the empty qualifier, so 4.1.135.Final and 4.1.135 are the same version rather than one outranking the other |
| TestCompareVersions | release qualifier below a number, above a prerelease; SP above the release | pins the aligned-segment case, where strings.Compare would otherwise make "Final" beat "1" and read 1.0.Final as newer than 1.0.1 |
| TestIsPreRelease | SNAPSHOT / M1 / m2 / alpha / beta / cr / pr / ea / preview / plain release / date-like / empty | isPreRelease recognizes every qualifier convention Maven Central uses and accepts a bare numeric or date-like release |
| TestIsPreRelease | Final / RELEASE / GA / SP | a release qualifier is *not* a pre-release -- counting it as one would skip every netty and hibernate release when picking a latest stable version |
| TestValidateCoordPart | valid group/artifact/version/qualifier forms / empty / traversal / doubled dot / slash / backslash / NUL | validateCoordPart accepts safe coordinate segments and rejects every unsafe one before it can reach a URL |
| TestLatestStablePrefersRelease | - | latestStable returns metadata's <release> when it is itself a stable version |
| TestLatestStableSkipsPreReleaseCandidateInRelease | - | the verified jackson-annotations case: <release> names a candidate, so the highest surviving stable <version> is used instead |
| TestLatestStableAllPreReleaseVersionsIsError | - | every listed version being a pre-release is an error, never a silent pre-release pick |
| TestLatestStableUnreachableMetadataIsError | - | an unreachable maven-metadata.xml is an error |
| TestLatestStableMajor | - | latestStableMajor picks the highest stable version on one major line by numeric comparison (8.10 over 8.1), skipping pre-releases and versions naming no number and never taking the <release> of another line; a line with no stable release, or unreachable metadata, is an error naming the line |
| TestResolveClosureMQJakarta | - | the verified com.ibm.mq.jakarta.client closure resolves to seed + BC trio + jakarta.jms-api + org.json:json at the versions the seed's POM declares |
| TestResolveClosureSyslogResolvesParentProperties | - | the verified logstash-logback-encoder:9.0 -> jackson-databind:3.0.1 case: jackson-databind's own version-less dependencies are resolved through its parent jackson-base's <properties>, none marked Fallback |
| TestResolveClosureAppliesScopeOptionalTypeFilter | - | test/provided/system/import scope, optional=true, and type=pom dependencies are all excluded from the closure; plain compile/runtime deps survive |
| TestResolveClosureDependencyVersionFromDependencyManagement | - | a version-less dependency resolves from its parent's <dependencyManagement> |
| TestResolveClosurePropertyDefinedTwoParentsUp | - | a `${property}` version defined only on the grandparent POM still resolves by walking the full parent chain |
| TestResolveClosureUndefinedPropertyFallsBackToLatestStable | - | a `${property}` with no definition anywhere in the parent chain falls back to that dependency's own latest stable release, marked Fallback |
| TestResolveClosureDependencyCycleTerminates | - | a dependency cycle (x -> y -> x) terminates and both artifacts appear exactly once |
| TestResolveClosureArtifactCountCap | - | a closure past maxArtifacts is truncated to exactly maxArtifacts rather than growing unbounded |
| TestResolveClosureHostileGroupIDIsDropped | - | a dependency with a path-traversal groupId is dropped from the closure and never reaches the HTTP layer, while a sibling good dependency still resolves |
| TestResolveClosureUnreachableSeedMetadataIsError | - | the seed's own maven-metadata.xml being unreachable is an error |
| TestResolveClosureUnreachableDependencyPomKeepsArtifact | - | a non-seed dependency whose POM 404s stays in the closure at its declared version rather than being dropped or erroring; libs.Download's own per-artifact Failure is where a real fetch problem surfaces |
| TestResolveClosureDependencyVersionUsesProjectVersionProperty | - | resolveProperty's `${project.version}` case: a dependency version referring back to its declaring POM's own version resolves correctly |
| TestResolveClosureDependencyVersionUsesProjectGroupIdProperty | - | resolveProperty's `${project.groupId}` case: the substitution fires and its result reaches the closure unchanged |
| TestResolveClosureDependencyUnresolvableVersionAndUnreachableMetadataIsDropped | - | resolveDependencyVersion's ok=false path: a dependency with no version anywhere in the chain, whose own latest-stable fallback also 404s, is silently dropped from the closure like any other malformed item |
| TestResolveClosureAtPinnedVersionNeverFetchesMetadata | - | resolveClosureAt with a pinned version never consults maven-metadata.xml at all, and the seed artifact is not marked Fallback |
| TestResolveClosureAtPinnedVersionResolvesDependencyClosure | - | the verified com.ibm.mq.jakarta.client:9.4.2.0 pin still resolves its jakarta.jms-api:3.0.0 dependency through the normal parent/dependency chain |
| TestResolveClosureAtPinnedVersionNotFoundIsActionableError | - | a pinned version that does not exist on Maven Central is an error naming both the version and the artifact |
| TestResolveClosureAtPinnedVersionInvalidCharsetIsError | - | a pinned version outside the safe coordinate charset is rejected before it can reach the HTTP layer |
| TestResolveParentChainDetectsCycle | - | a parent-POM cycle (a -> b -> a) is detected and errors rather than looping forever |
| TestResolveParentChainExceedsMaxDepth | - | a parent chain past maxParentDepth (with no cycle) is capped with an error |
| TestJarURL | - | jarURL assembles the Maven Central path from group/artifact/version exactly |
| TestSplitJarBasename | hyphenated / dotted / short date-like / hyphen-in-version / Final qualifier / alpha qualifier / underscore in version / no digit-led hyphen / plain | splitJarBasename recovers (artifact, version) from a real jar filename across every naming convention lib-list actually contains, and refuses a name with no digit-led hyphen to split on |
| TestSplitJarBasename | classifier stripped (netty native, sources) | a trailing classifier is not part of the version and is dropped, so the entry parses instead of being rejected whole |
| TestValidateImageVersionQualifiers | accepted / rejected | Final/RELEASE/GA/SP join numerics and pre-release qualifiers as orderable, while genuine garbage (9zzzzzzzzzzz, a bare classifier, an unknown word) stays rejected -- the gate exists so a stale entry cannot compare as newer than a real release |
| TestImageNameTag | hub / no namespace / registry with a port / no tag / digest / empty | an image reference splits into name and tag; the tag separator is found in the last path element so a registry port is not mistaken for it |
| TestBuiltinList | silent: none / captured tag / newer tag / the floor itself / registry mirror / 3.1.0 / a 3.1 point release / 3.2.0 / past the newest capture. warns: below the floor / first release past the range / a 3.0.x release / different image / digest / no tag / latest | each built-in list describes a *range*, so a release inside it is silent and gets that list -- a differing tag is not itself a mismatch, 3.1.x stays on the 3.1.0 list, 3.2.0 gets its own and a later release stays on the newest; below the floor, past the 2.x list's ceiling, in the uncaptured 3.0.x gap, a different image, or a reference naming no release (`latest` used to sort past every number and pass) warns, names the reference, what it was judged against and `--omit-lib-file`, and gets the nearest list |
| TestBuiltinListPicksTheNearestLine | 2.14.1 / 3.1.0 / 4.0.0 / 3.0.5 / 2.9.0 / none | against a two-line table (2.x and 3.x): each release gets its own line, one past the last line gets it silently, one in the gap between lines gets the line before it with a note, one older than every capture gets the oldest, and nothing declared gets the newest |
| TestEmbeddedListsTable | - | the rows run oldest first with non-overlapping ranges, each covers the release it was captured from, every row's file is really embedded, every embedded list has a row, and no list is identical to the one before it (that release belongs in the earlier list's range) -- a capture without a row, a row without its capture, or a duplicate capture fails the build |
| TestEmbeddedListsMatchTheirConnectorLine | one subtest per built-in list | each list ships the generation its capture tag's line runs on -- 2.x Spring Boot 3 and Jackson 2, 3.x Spring Boot 4 and Jackson 3 -- and a list for a line with no recorded generation fails until one is added |
| TestEmbeddedListRange | - | covers() includes a range's from and excludes its before (open-ended when before is unset), and describes()/file() render "2.10.0 and later, before 3.0.0", "3.1.0 and later" and the imagelibs path |
| TestConnectorRelease | 2.x / 3.x / two-digit major / suffixed tag / registry mirror / latest / digest / different image / overflowing major / none | the connector release -- tag and leading major -- is read off the connector image's reference only, and nothing is read from a reference naming no release |
| TestDownloadImageMismatchReported | uncovered (2.9.0, first release past the range) / covered (2.13.0, 2.14.1) / none / --omit-lib-file | the check end to end: only an image no built-in list can speak for warns -- below the floor or past the ceiling alike -- both ends of the covered range stay silent and report the list's range in OmitListRange, and a named omit list suppresses the check and states no range -- the operator declared that list, as with an explicit `--url`. Every suppression case uses the uncovered reference, so it cannot pass vacuously |
| TestEmbeddedOmitListFullyParses | one subtest per built-in list | every line of every shipped image list either is not a jar reference or parses to an orderable version, so a future capture that reintroduces an unparseable shape fails the build instead of printing warnings on every run; each list's provenance is its own name |
| TestSplitJarBasenameRejectsNonJar | - | a non-.jar name is rejected outright |
| TestLoadImageLibsSkipsCommentsAndBlankLines | - | a `#`-commented header line and blank/whitespace-only lines are skipped, leaving only the real jar entry |
| TestLoadImageLibsSkipsUnsplittableLineWithoutFailing | - | a line that does not split into artifact+version (e.g. jrt-fs.jar) is skipped rather than failing the whole load |
| TestLoadImageLibsToleratesSurroundingWhitespace | - | leading/trailing whitespace and a trailing \r around a jar name do not stop it from loading |
| TestLoadImageLibsBadPathIsError | - | a nonexistent `--omit-lib-file` path is an error |
| TestLoadImageLibsEmbeddedDefault | - | an empty path loads the built-in list it is handed -- the 2.x one, captured from solace/solace-pubsub-connector-ibmmq:2.13.0 -- which carries the BC/jakarta.jms-api/json/logstash-logback-encoder versions and deliberately omits com.ibm.mq.jakarta.client (the licensing carve-out) |
| TestImageSatisfies | image has a newer version / equal version / older version / does not have it at all | imageSatisfies reports provided=true and the image's version for an equal-or-newer match, and provided=false (with the image's version, or none) otherwise |

## internal/statusreport

The CLI-side half of the status verb: the typed model both status views are built into, the parsers that fill it from what the engines and the in-container script emit, and the two renderings (human tables/blocks, and the --output json document). A pure package -- no os/exec, filesystem, network or globals -- so the whole report is testable from captured fixtures.

Tests: [statusreport_test.go](../internal/statusreport/statusreport_test.go), [parse_test.go](../internal/statusreport/parse_test.go), [render_test.go](../internal/statusreport/render_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestAge | seconds / minutes / hours and minutes / days and hours / past a week / fractional seconds / docker zero time / empty / unparseable / clock skew ahead | the compact age a table column wants, against an injected clock; an unparseable or zero stamp is no age at all rather than an age of zero, and a host clock behind the engine's never renders negative |
| TestParseQuantity | 12 values | kubernetes quantities (`120m`, `512Mi`), engine sizes (`512MiB`, `1.5GiB`, `1kB`), plain byte counts, and Jackson's scientific notation (`4.32013312E8`) -- the form a large heap actually arrives in; longest suffix wins, so `Mi` is never read as the decimal `M` |
| TestPercent | 7 pairs | a whole-number percentage only when both sides read and the limit is non-zero; docker's unlimited container (limit 0) yields no percentage rather than a division |
| TestBytesAndCores | 7 byte values / 3 core values | limits rendered in the binary units env.yaml writes them in; 0 (no docker limit) and -1 (no JVM maximum) both render as nothing |
| TestBanners | section, all names / no names / gap in the middle / instance / podman | the two banner levels; an unset name is dropped so a separator always sits between two real names, and the instance form is unchanged from the report this verb printed before the container view existed |
| TestTableAlignsColumnsAndNeverPadsTheLast | - | every column padded to its widest cell with a two-space gutter, an empty cell rendered `-` so a column never collapses, a short row padded rather than panicking, and no trailing whitespace on any line |
| TestTableEmptyReportsNoRows | - | Empty lets a caller skip printing a heading with nothing under it |
| TestKVAlignsValuesOnTheWidestKey | - | every value starts in the one column the widest key decides |
| TestKVBuilderDropsEmptyValues | - | the report's noise rule in one place: a fact that was not collected prints nothing rather than a line saying it is unknown |
| TestResourceLineDropsWhatIsMissing | nil / no usage / usage alone / usage and limit / all three / engine-formatted | a resource renders only what was collected, including docker's own both-sides-in-one-string form |
| TestImageMismatch | identical / kubernetes-normalised / library namespace / implied latest / different tag / different repository / no running image / no expectation / digest pin matches / digest pin differs / digest pin with no reported digest | the failed-rollout check: tolerant of the spellings the engines use for one reference (so a correct kubernetes instance never reads as a mismatch), digest-pinned env.yaml answered from the digest, and never a claim when either side was not collected |
| TestSortInstancesGroupsByNamespaceThenName | - | a repeated run prints the same order however the engine listed them, grouped by namespace for a cluster-wide report |
| TestExitCodeText | - | a container that has not terminated carries no exit code |
| TestParsePodsReadsBothHalvesOfARealPair | - | one `get pods -o json` yields identity, state, readiness, restarts, digest, node and limits; the container's own running-since stamp wins over the pod's startTime, and a CrashLoopBackOff reads as restarting with the last termination's exit code (137, OOMKilled) |
| TestParsePodsComponentsComeFromWhatThePodReferences | - | components are read from the pod spec (volumes, each source of a projected volume, envFrom, imagePullSecrets) rather than from env.yaml, an emptyDir is not a component, a mounted volume carries its path, and parsing claims no status -- that is a separate probe |
| TestParsePodsSingleObjectDocument | - | `get pod <name> -o json` answers the object itself with no items array; both shapes parse, since status uses both |
| TestParsePodsPendingAndTerminatedStates | pending, no container status / waiting on an image pull / terminated | a pod with no container status yet still reports a state from its phase, and each waiting/terminated reason reaches the report |
| TestParsePodsReadinessIsOnlyAVerdictWhenAProbeExists | - | without a readiness probe the column reads n/a, since kubernetes would otherwise report ready as soon as the container runs |
| TestParsePodsPicksTheConnectorContainer | - | a sidecar's restarts and limits are never reported as the connector's |
| TestParsePodsSeveralContainersNoneNamedConnector | - | no container-level fact is claimed when the connector cannot be identified; the pod phase still gives a state |
| TestParsePodsImageFilterIsWhatAllSearchesBy | - | the `--all` image filter keeps only connector pods, and an unfiltered parse keeps everything -- the filter is `--all`'s, not the parser's opinion |
| TestParsePodsErrorsAndSkips | empty / undecodable / one bad item in a list | an unreadable response is an error, but one unreadable item is skipped so the rest of the report survives |
| TestParseDeploymentAndService | - | replica counts, `replicas` omitted defaulting to the API's 1, service ports with the omitted protocol defaulting to TCP, a service merged without losing the deployment counts, a service-only workload, and an undecodable service |
| TestObjectExists | secret / bound claim / pending claim / not a document / empty | a live object reports "present", a volume claim reports its own phase (the only status here that can be bad while the object exists) |
| TestSecretDataKeys | - | the data keys of a Secret document are read without their values; a Secret with no data has none, and anything else is not a Secret |
| TestApplyTop | - | the connector's row wins over a sidecar's, a percentage appears only where a limit was read, and a pod the metrics API said nothing about keeps no usage |
| TestParseInspectDocker | - | docker's leading slash stripped from the name, the compose project read off the container's own label in the same call, the configured image reference rather than the local id, the nanocpu/memory ceilings, the age, and mounts/networks as attached components |
| TestParseInspectStatesAndHealthSpellings | exited / oom killed / restarting / paused / created / podman stopped cleanly / unknown status / podman Healthcheck key / podman Health key with a log / unhealthy / empty status / no healthcheck | every engine status normalised, a clean stop reporting no exit code (a zero exit adds nothing to the state), both spellings of podman's healthcheck block (including the current one, nesting a Log beside the status, which is what a quadlet HealthCmd produces), an unhealthy verdict passed through as the engine spells it, and n/a both where the block is absent and where it carries an empty status |
| TestParseInspectFilterAndErrors | - | the `--all` image filter, an empty response, and an undecodable one |
| TestParseInspectDropsControlCharacters | - | every string the engine supplied (name, compose project, image, error, health, mount source, destination and name, network and secret names) loses its control characters at the parse boundary, C1 controls included, so a terminal escape is never printed |
| TestParseImageDigest | - | the first RepoDigest is the registry digest; an image never pushed has none, which is not an error |
| TestApplyStats | - | the engine's own percentages are taken as given (docker's memory string already carries both sides), and a container with no sample keeps no usage |
| TestParseApplication | - | every line of the script's report: leader election, health, health components, uptime, version, java, config, raw heap bytes rendered to a percentage, numerically-ordered workflows, and the script's own stderr note arriving on the same combined stream; with no health document (an older script) no connection or verdict is invented |
| TestParseApplicationJudgesEachRunningWorkflowByItsConnections | - | with the MQ binder reconnecting, only the workflows whose bindings ride on it are marked, binders are listed in name order with the first line of their error, each workflow carries the connector's own health, and the raw lines never surface as notes |
| TestParseApplicationConnectionVerdicts | every connection up / the connector's own workflow health down / a binding down under a binder that is up / a stopped workflow is never judged / an unmapped binding is judged against every binder / a binder the document does not mention is unknown, not down | each source of the verdict marks the right workflow, and only a status that is present and not UP fails the instance |
| TestParseApplicationUnreadableHealthDocumentIsANote | - | a health document that does not parse becomes one note, nothing is judged on its account, and it is not kept as the raw document |
| TestParseApplicationKeepsTheHealthDocument | - | the raw document is the health-json line verbatim, or an older script's health-detail line, never a value that is not JSON, and absent from an older script's healthy report |
| TestParseApplicationKeepsWhatItDoesNotRecognise | - | an unknown line is kept as a note, so an instance carrying a newer script still reports everything it printed |
| TestParseApplicationHealthDetailAndBareHeap | - | health-detail is not swallowed by the health prefix that starts the same way, and an unbounded heap reports no maximum and no percentage |
| TestParseApplicationCRLFAndBlankLines | - | CRLF output and blank lines parse identically |
| TestParseApplicationIndentedLineWithNoBlockIsANote | - | an indented line with no block header above it is a note, never a workflow |
| TestRenderContainerViewBasic | - | the section banner, the column set, and the workload summary; the basic level carries neither the NODE column nor any detail block, and kubernetes reports READY rather than a HEALTH column |
| TestRenderContainerViewDetails | - | the details block: digest, resource lines, the components table, and the image-expected line whose presence is itself the finding; an instance with no sample carries no resource lines at all |
| TestRenderContainerViewDockerUsesHealthColumn | - | docker reports the engine's healthcheck verdict where kubernetes reports readiness, and has no NODE column |
| TestRenderContainerViewPodmanReportsTheHealthVerdict | - | the podman half of that pair: a quadlet healthcheck's verdict reaches the table, a container declaring none still falls back to n/a rather than an empty cell, and podman gets no READY column |
| TestRenderContainerViewAllNamespacesLeadsWithNamespace | - | instances spanning namespaces cannot share one banner, so each row leads with its own; one shared namespace rides in the banner instead |
| TestRenderApplicationViewBasicAndDetails | - | the unchanged instance banner, the aligned basic lines, right-aligned workflow ids, enrichment only at the details level, and no container table in this view |
| TestRenderHealthDocumentInFull | - | `--raw`'s document is printed indented as JSON under a `health document:` header after the workflows, in place of the one-line `health-detail` |
| TestRenderConnectionsAndDegradedWorkflows | - | a running workflow that is not connected names what holds it in parentheses; the basic level shows a connections block only for binders that are not UP and moves `health-detail` to `-d`; details lists every binder; a healthy report carries no block; an older script's report keeps `health-detail` at the basic level; JSON carries connections, degraded and the workflow's own health |
| TestRenderFailedInstanceKeepsItsBlock | - | an instance whose script could not run still gets a banner with the failure as a body line, and the container table that explains it comes first |
| TestRenderNotesAndScriptNotesShareOneIdiom | - | a note the CLI made and a note the script made read the same, and run-level notes come after the facts they qualify |
| TestRenderEmptyReport | - | a report with no instances renders nothing |
| TestRenderInstanceWithNoContainerFactsStillReportsItsApplication | - | a failed engine query leaves no table but does not cost the application half |
| TestJSONIsTheSameModelTheTablesRender | - | the document round-trips, carries schemaVersion and the field spellings a consumer keys off (the compatibility contract), and omits an unset field entirely rather than emitting null |
| TestJSONEmptyRunIsAnEmptyList | - | an empty run is `[]`, so a consumer can iterate without a nil check |

## cmd/solmq-conn-util

The CLI shell -- flag parsing, the exit-code contract, the generate/validate/examples/auto-complete commands, verb aliases, and the deploy/remove/start/stop/restart/status/logs/cli seams for all three engines. The progress tests cover the other half of a status run -- the stderr spinner and the `--verbose` step lines -- including the guarantee that stdout does not change when either is drawn. The completion tests also gate the four generated shell scripts against the command model, and the doc tests gate the two generated markdown references against it.

Tests: [main_test.go](../cmd/solmq-conn-util/main_test.go), [lifecycle_test.go](../cmd/solmq-conn-util/lifecycle_test.go), [commands_doc_test.go](../cmd/solmq-conn-util/commands_doc_test.go), [abbreviation_doc_test.go](../cmd/solmq-conn-util/abbreviation_doc_test.go), [completion_test.go](../cmd/solmq-conn-util/completion_test.go), [testcatalog_test.go](../cmd/solmq-conn-util/testcatalog_test.go), [support_test.go](../cmd/solmq-conn-util/support_test.go), [runas_test.go](../cmd/solmq-conn-util/runas_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestHostReaderFollowsTheKernelsRule | 0600 / 0640 / 0604 / 0644 / 0044 x stranger / owner / group | a host file is judged as the kernel does: the owner bits for the owning uid, else the group bits for the group, else the other bits -- an owner with no owner bits cannot read however open the rest is |
| TestMountChecksLabelEachPathByItsKey | store / keystore store / libs.dir / none | a plan's host mounts become checks labelled `<section> mounts <key>` at the plan's path, a store used as a keystore judged as a private key, libs.dir as a directory and any other store as a public file; no mounts give no checks |
| TestMissingStoreFiles | a present store / libs.dir | a store file that is there passes, and libs.dir is left to the engine |
| TestMissingStoreFiles | missing | a missing store is refused, naming the key and path: docker would create an empty directory in its place |
| TestMissingStoreFiles | a directory | a directory where the store belongs is refused too: an earlier deploy may have made it |
| TestMissingStoreFiles | both in one plan | every bad source is reported, in mount order |
| TestPodmanReader | root / rootless | podman's bind mounts are judged with the host's own ids as root, through the subordinate ranges rootless |
| TestUnreadableMountsNamesWhatTheConnectorCannotRead | - | the deploy warning names a 0600 truststore, a 0600 keystore, a 0600 jar inside a readable libs dir and a 0700 libs dir, each with its mode, the run-as identity and a remedy for its kind -- o+r or the group for a truststore or jar, o+rx or the group for a directory, the group rather than everyone for a private key; readable files, non-jar files and missing paths are not named |
| TestUnreadableMountsJudgesRootByTheModeBits | image / uid 0 | root -- the image's own under run-as image, or uid 0 -- gets no pass with every capability dropped: a 0600 file another uid owns is named, a 0644 one is not; the platform gate checks on linux only |
| TestRootlessReaderSeesTheInvokersGroup | gid 0 / gid 1000 / uid 0 / image | rootless podman's non-zero uid owns no host file and its gid 0 is the invoker's group, so a 0640 jar is readable and a 0600 one is not; a non-zero gid maps to no host group; container root -- uid 0 or image -- is the invoking user and reads its own 0600 file |
| TestCurrentUserNameIsTheEffectiveUser | - | the name resolves the effective uid, falling back to the uid in decimal, and is never empty |
| TestDeployPodmanRootlessLinger | already lingers / off / enable refused / probe fails | a rootless Linux deploy asks logind after the preflight and before any write: `yes` makes no enable call, `no` enables and prints one `note:` line, a refused enable gives `sudo loginctl enable-linger <user>`, a failed probe names the command (exit 1 and nothing stored or written for the last two); loginctl and systemctl calls carry the repaired session and podman calls do not |
| TestValidatePodmanRefusesWithoutLinger | linger off / probe fails | validate reports a user that does not linger, and a probe that cannot run, as host errors (exit 1) and never enables linger; a lingering user validates |
| TestPodmanRunAsAgainstTheSubIDRanges | inside / uid outside / gid outside / no allocation / id 0 / image / unreadable files / probe fails, deploy and validate | podman is asked for the uid map then the gid map and the run-as id is checked against it: a refusal names the range and the usermod fix naming both ranges computed from the subid files (`not computed` when they cannot be read), then the run-as alternative; deploy exits 1 before storing or writing, validate reports a host error (exit 1); id 0 and run-as image run no probe |
| TestPodmanRootfulRule | non-root asking for rootful / root without it x deploy / remove / start / validate; root asking for rootful | the refusal is exit 2 naming `podman.rootful` with no command run, in every verb; root asking for rootful drives the system manager with no session |
| TestPodmanScopeRefusals | relative XDG_CONFIG_HOME / no user runtime directory | a relative XDG_CONFIG_HOME is exit 2 with nothing run, deploy and validate; a missing runtime directory is exit 1 before the preflight, and a host error in validate |
| TestDispatchHandlersMatchModel | verbs / generate targets / deploy platforms / completion shells | the dispatch handler sets and cliVerbs agree in *both* directions, so a command added to one cannot drift from the other |
| TestPlatformMapsCoverThreeNames | - | platformNames, platformGenerators and actTargets agree in both directions -- a modeled platform with no handler, or a handler with no modeled entry, fails |
| TestExitCodeContract | nil args | run(nil) returns exit code 2 |
| TestExitCodeContract | unknown command | run([bogus]) returns exit code 2 |
| TestExitCodeContract | help short -h | run([-h]) returns exit code 0 |
| TestExitCodeContract | help long --help | run([`--help`]) returns exit code 0 |
| TestExitCodeContract | help word | run([help]) returns exit code 0 |
| TestExitCodeContract | 6 requested-help spellings | `status -h`, `cli -h`, `deploy --help`, `-h` among a verb's flags (routed as flag.ErrHelp), `help status`, and `help sts` each exit 0 -- requested help is never an error, wherever it is asked |
| TestExitCodeContract | help with an unknown command | run([help, bogus]) returns exit code 2 |
| TestExitCodeContract | unknown flag -nope | run returns exit code 2 for unrecognized flag |
| TestExitCodeContract | missing env file | run returns exit code 2 when env file does not exist: the env file is what the tool was given |
| TestExitCodeContract | invalid spec | run returns exit code 2 for structurally invalid workflow: refused before acting |
| TestExitCodeContract | auto-complete no shell | run([auto-complete]) returns exit code 2 |
| TestExitCodeContract | auto-complete bogus shell | run([auto-complete, bogus]) returns exit code 2 |
| TestExitCodeContract | completion is no longer a command | run([completion, bash]) returns exit code 2: `completion` is rejected as an unknown command, with no alias to fall back on |
| TestExitCodeContract | 8 near misses | d / v / s / g / comp / h / hlp / stat each exit 2: none was picked as an alias, so none may resolve |
| TestExitCodeContract | dep is no longer deploy | run([dep]) returns exit code 2: `dep` is not a recognized command; `deploy`'s alias is `dp`, matching solace-util's convention, and there is no `dep` alias |
| TestVerbAliasesDispatchLikeCanonical | gen / dp / rm / sts / ver / vld / eg / dl | each alias reaches the same handler as its canonical verb, so the alias table and the dispatch map cannot drift apart |
| TestGenerateConfigStdoutAndFileMatch | stdout run | exit 0 and stdout contains 'spring:' |
| TestGenerateConfigStdoutAndFileMatch | file run | exit 0 and written file content equals prior stdout exactly |
| TestGenerateConfigTargetAliasResolves | - | `cfg` resolves to the `config` target and `gen cfg` emits application.yml (exit 0, stdout has `spring:`) -- pins that the positional goes through resolveTarget, so a modeled target alias is not documented-but-rejected |
| TestGenerateFlagsBeforeAndAfterPositional | flags before positional target | exit 0 and output file written |
| TestGenerateFlagsBeforeAndAfterPositional | flags after positional target | exit 0 and output file written |
| TestGenerateConfigWorkflowCapExceeded | - | a folder over validate.MaxWorkflows (20) is a fatal error (exit 2) naming the count and the cap, and writes no `-o` output file |
| TestGenerateConfigEmitWriteError | - | emit to path with missing parent dir returns exit code 1 |
| TestLoadEnvWorkflowsDirRelativeToEnvFile | - | workflows.dir resolved relative to env file not cwd; exit 0 and stdout has spring: |
| TestLoadEnvExcludesEnvFileFromWorkflowSet | - | env.yaml excluded from its own workflow scan; exit code 0 |
| TestDeployKubernetesSeamHappyPath | - | exit 0 and 6 runner calls: a create probe each for deployments, configmaps and secrets, the namespace read (it exists, so nothing is created), the apply with argv [kubectl apply -f -] and kind: Deployment on stdin, and the rollout wait |
| TestRemoveKubernetesSeamHappyPath | - | exit 0, a delete probe per kind (deployments, configmaps, secrets), the manifest delete [kubectl delete -f - --ignore-not-found], the occupancy listing, the delete-namespaces probe and the namespace delete |
| TestDeployKubernetesProbesOperatorSecrets | missing Secret / missing key | after the permission probes, a get secret per operator Secret; a missing Secret or key stops the deploy before apply, naming the Secret |
| TestDeployKubernetesProbesOperatorSecrets | not allowed to read it / complete | a Secret the identity may not read only warns, and a complete one lets the apply through |
| TestLifecycleVerbsActInPlace | stop / start / restart x kubernetes / docker / podman | the platform's preflight, then one call acting on the instance env.yaml names -- scale, rollout restart, the engine's own verb, or systemctl -- -- on kubernetes start and restart then wait for the rollout, stop does not -- with no -env variable exported and no store file present |
| TestLifecycleRolloutWait | start and restart x refused / timed out / rollout failed; stop | a wait kubectl refuses is a warning and exit 0 with the ok line, a timeout or failed rollout is "<verb> kubernetes failed" and exit 1, and stop makes no wait |
| TestLifecycleVerbsRefuseBeforeActing | - | a failed preflight stops before acting, a name that is not a DNS-1123 label is refused before anything runs (exit 2), a positional argument is a usage error, and no compose file is written |
| TestDeployKubernetesSeamRejectsUnsafeCommand | - | unsafe kubernetes.command yields exit 2 and zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | path value | `--allow-command` /usr/bin/lima exits 2 naming the bare-name rule, zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | unsafe character | `--allow-command` lima;rm exits 2 with the shared charset message, zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | escalator, including another case with .exe | `--allow-command` sudo and Doas.EXE exit 2 with "never permitted" and "elevate the tool instead", zero runner calls |
| TestAllowCommandSudoNamesTheSupportedAlternative | deploy / remove / start / status / logs / cli | `--allow-command sudo` exits 2 on every verb that takes the flag, naming `sudo solmq-conn-util`, zero runner calls |
| TestAllowCommandFlagRejectedOnGenerateAndValidate | generate config / validate | `--allow-command` is undefined on generate/validate; exit 2 as an unknown flag |
| TestAllowCommandFlagRepeatableThreadsToRunner | - | "lima podman" rejects (exit 2) with zero runner calls without the flag; repeating `--allow-command` lima twice threads through to preflight (argv [lima podman info]) and to the podman secret calls |
| TestDeployRefusesSudoCommandWhateverIsApproved | no approval / an unrelated approval | podman.command `sudo podman` exits 2, naming the escalation wrapper and "run the tool itself under sudo instead", zero runner calls |
| TestDeployKubernetesPreflightFailureStopsBeforeApply | - | a failing kubernetes preflight (auth can-i create deployments -n <ns>) stops with exit 1 and exactly 1 runner call |
| TestDeployDockerPreflightFailureStopsBeforeWrite | - | a failing docker preflight (argv [docker info]) stops before the compose file is written, exit 1, exactly 1 runner call |
| TestDeployPodmanPreflightFailureStopsBeforeWrite | - | a failing podman preflight (argv [podman info]) stops before the unit is written, exit 1, exactly 1 runner call |
| TestValidateOKAndErrors | valid spec | validate exits 0 |
| TestValidateOKAndErrors | invalid spec | validate exits 2: an error in env.yaml or its workflows is a usage error (a host-check-only finding stays 1, see TestValidateChecksDockerHost) |
| TestValidateChecksDockerHost | - | validate reports a missing docker store source and a directory in its place together, exits 1, and runs nothing |
| TestValidateWarnsAboutHostMountsTheConnectorCannotRead | - | validate warns, as deploy does, about a docker truststore and a rootless podman libs.dir the connector could not read; warnings alone exit 0 and nothing runs (Linux only) |
| TestExamplesWriteSkipForceThenGenerate | first write | examples command exits 0 creating env.yaml |
| TestExamplesWriteSkipForceThenGenerate | re-run without -f | exits 0 and skips existing file, content stays 'touched' |
| TestExamplesWriteSkipForceThenGenerate | re-run with -f before dir | exits 0 and overwrites existing file |
| TestExamplesWriteSkipForceThenGenerate | generate on shipped examples | generate config on generated env.yaml exits 0 |
| TestExamplesDefaultDir | - | examples with no dir arg exits 0 and creates ./examples/workflow-0.yaml |
| TestDownloadMissingAndUnknownWordsRejected | missing target / unknown target / missing set / unknown set | each is a usage error (exit 2) naming the offending word, with the downloader and the runner both left uncalled |
| TestDownloadDirDefaultAndPositionalOverride | default dir / explicit dir | the trailing [dir] positional defaults to ./libs and threads an explicit value through to Input.Dir unchanged |
| TestDownloadSetReachesInput | mq / syslog | both modeled sets thread through to libs.Input.Set unchanged |
| TestDownloadURLFlagRepeatable | - | repeated `--url` collects every occurrence, in order, into Input.URLs |
| TestDownloadJMSFlagIsGone | mq / syslog, either value | `--jms` is an unknown flag: exit 2 and the downloader never runs, so a script still passing it fails loudly instead of being silently ignored |
| TestDownloadForceFlagReachesInput | default false / -f short / --force long | both -f and `--force` spellings reach Input.Force, defaulting to false |
| TestDownloadVersionFlagReachesInput | default empty / explicit pin | `--version` defaults to "" (latest stable) and an explicit pin reaches Input.Version unchanged |
| TestDownloadOmitLibFileFlagReachesInput | default empty / explicit path | `--omit-lib-file` defaults to "" (a built-in list) and an explicit path reaches Input.OmitLibFile unchanged |
| TestDownloadIncludeProvidedFlagReachesInput | default false / --include-provided | `--include-provided` defaults to false and reaches Input.IncludeProvided true when given |
| TestDownloadReportExitCode | clean write / skip only / omitted only / partial failure / total failure / systemic error | a clean write, a skip-only run, and an omitted-only run all exit 0; a non-empty Report.Failed -- whether partial or total -- and a systemic downloadFn error both exit 1 identically, pinning the deliberate choice not to mint a distinct code for partial vs total failure |
| TestDownloadReportPrintsWrittenSkippedFailedAndFallback | - | reportDownload prints "wrote:"/"exists (use -f to overwrite):"/"failed:" lines and a Fallback note labelled "guessed version:", mirroring runExamples' line shapes |
| TestDownloadReportPrintsOmittedBlockDistinctFromFallback | - | Report.Omitted prints its own "omitted:"-prefixed lines, distinct from "failed:" and "guessed version:", and the counts footer names the omitted count; exit stays 0 with no Failed entries |
| TestDownloadReportExitZeroWhenEverythingOmitted | - | a Report where every artifact was omitted (nothing written, skipped, or failed) exits 0, and the counts footer reads "0 written, 0 skipped, 4 omitted, 0 failed" |
| TestDownloadReportNextHint | no omissions / with omissions | the "next:" hint always points at wiring libs.dir/libs: config key, gaining an extra clause about the omitted jars already being on the image only when Report.Omitted is non-empty |
| TestDownloadReportPrintsOmitListProvenance | built in / explicit file | the "omit list:" line names Report.OmitListProvenance, annotated "(built in; describes <range>)" from Report.OmitListRange for a built-in list; an explicit path prints bare, and each Report.OmitListWarnings entry gets its own "omit list warning:" line |
| TestDownloadReportPrintsSeedChoice | - | Report.SeedChoice prints as its own "seed:" line, and a report without one (a pinned `--version`, the mq set) prints no seed line |
| TestDownloadSetMapMatchesModel | - | downloadSets (main.go's set-name dispatch table), the model's download/jar Sets, and internal/libs.SetNames() all name exactly the same sets |
| TestDownloadReadsDeployedImageFromEnv | default present / no env.yaml / explicit -e unreadable / defaulted malformed / no image block | the advisory config read: the image reaches libs.Input, an absent default is silent and still downloads, and only a file the operator named is systemic (exit 2) |
| TestGenerateKubernetesStdout | - | exit 0 and stdout contains kind: Deployment |
| TestGenerateDockerToFile | - | exit 0 and compose file opens with the defaulted, quoted project line, then contains services: and image: img:1 |
| TestGeneratePodmanQuadletStdout | - | exit 0 and stdout contains unit banner '# === solmq-conn-util.container ===' |
| TestGeneratePodmanOnlyLibsIsAHostPath | -e env.yaml from the file's own dir | the libs Volume= source is absolute even when -e is spelled relatively, and it is the only Volume=: the store selected in the same spec is a Secret= mount. A relative source is not a near-miss in a quadlet: systemd starts the unit with no useful cwd, and podman reads a source with no ./ or / prefix as a named volume, so `Volume=libs:...` would silently mount an empty volume over the jars |
| TestDeployDockerSeamPipesComposeAndRuns | - | exit 0, no compose file on disk, 3 runner calls (preflight, the running probe, up) argv [docker compose -f - --project-directory <env.yaml's dir> up -d] with the compose file on stdin -- no -p, since the project is declared by the file's own name: key; the file carries the deploy digest label, and the child env ends with `_GEN_DEPLOY_SHA256` set to containercli.ComposeDigest of the file and the rest of the env |
| TestDeployDockerSeamFailedRunLeavesNoComposeFile | - | preflight and the running probe succeed but the real `up` call (index 2) fails; exit 1, the up still had its file on stdin, and no compose file is left on disk |
| TestDeployDockerSeamChildEnvCarriesCredentials | - | preflight call carries no env; the real `up` call (index 2) carries the resolved literal and -env credentials as STABLE=value pairs, then the deploy digest |
| TestDeployDockerWarnsAboutMountsTheConnectorCannotRead | - | linux only: a 0600 truststore and a 0600 jar under libs.dir, both written with ${VAR}s and owned by another uid and gid than the run-as ones, are each named on stderr at their expanded path with their mode, and the deploy still runs (preflight, the running probe, then up) |
| TestRemoveDockerSeam | - | exit 0, 2 runner calls (preflight then down) argv [docker compose -f <compose> down] |
| TestDeployDockerCarriesCreatedStores | - | a docker deploy runs info, image inspect (the image starts as the storescript hands over to), the running probe and compose up, whose child environment -- never argv -- carries the built store base64-encoded under _GEN_TLS_TS_STORE and its digest |
| TestDeployDockerPullsAnImageItCannotInspect | pulled, then it starts as expected | an image the engine cannot describe is pulled and inspected again, then the running probe and compose up run |
| TestDeployDockerPullsAnImageItCannotInspect | pulled, still not described / the pull fails | the deploy stops before compose runs |
| TestDeployDockerRefusesAnImageStartingOtherwise | - | an image the engine reports starting another way fails the deploy naming both commands, after only info and image inspect, with no compose file written |
| TestRemoveDockerNeedsNoCreatedStoreSources | - | remove succeeds with the store's source gone -- its password resolved like every declared credential -- handing compose placeholders for the store secret and the digest |
| TestDeployDockerRefusesAMissingStoreFile | - | a docker deploy whose existing store is missing exits 1 naming it, before any runner call; remove still tears down |
| TestDeployPodmanSeamWritesUnitsAndStarts | - | exit 0; a leading `podman info` preflight; the `podman version` probe before anything is compared (the fake's "out" is no version, so no value is read back and every secret counts as changed) and the unit's `is-active` probe, both before anything is written; then `secret rm --ignore`/`create` per secret (the five credentials, application.yml, the selected store, the status script; no label, the store's bytes on stdin unchanged), one batched `secret rm --ignore` of the unmounted logback secret and of the store secret the previous unit mounted but this spec no longer selects, then `systemctl daemon-reload` and `start`; the `.container` unit (0644) is the only file written, and a leftover base-dir is not even created |
| TestDeployPodmanRootlessWarnsAboutAnUnlistableLibsDir | - | linux only, rootless: the default 1000:0 maps to no host uid and to the invoker's group, so a 0700 libs.dir written as ${VAR} is named on stderr at its expanded path with the listable remedy, and the deploy still runs |
| TestDeployPodmanUnresolvedInputFailsBeforeAnyWrite | unreadable store | a selected store's file that cannot be read fails the deploy (exit 2) naming tls.stores.ts.existing.pkcs12 and its path, after only the read-only preflight probe and before any secret is created or the unit written: the steps that follow have side effects outside the process, so a late failure would leave a half-built deployment |
| TestDeployPodmanUnresolvedInputFailsBeforeAnyWrite | unset credential variable | a password-env naming an unset variable fails the same way (exit 2), naming the variable, before any secret is created |
| TestDeployPodmanSetButEmptyCredentialFailsBeforeAnyWrite | set but empty password-env variable | exit 2 saying the variable is set but empty, only the preflight probe run, nothing written; validate warns the same |
| TestDeployPodmanFailureStopsBeforeSystemctl | credential | a failing `podman secret rm` for the first credential (call 3, after the preflight, the version and the state probe) stops the deploy there (exit 1, the failed call named), with no unit written and no daemon-reload or start |
| TestDeployPodmanFailureStopsBeforeSystemctl | document | a failing secret call for application.yml stops the deploy the same way |
| TestDeployPodmanFailureStopsBeforeSystemctl | unmounted file secret | a failing batched `secret rm` of the file secrets the spec no longer mounts stops the deploy before the unit, which would otherwise run beside a stale secret |
| TestDeployPodmanFailureStopsBeforeSystemctl | unit write | a quadlet directory that cannot be created (a file where ~/.config should be) fails the deploy naming it, after every secret call and before any daemon-reload |
| TestRemovePodmanSeamStopsRemovesReloads | - | exit 0, a leading `podman info` preflight, `systemctl stop` then unit removal and `daemon-reload`, then one `podman secret rm --ignore` for the credentials, the three document secrets and the store secret the deployed unit mounted (read before the unit is removed), only after the unit is gone; files an older deploy left under base-dir are left alone |
| TestRemovePodmanDeletesACreatedStoreSecretWithoutTheUnit | - | with no unit left to read, remove still deletes a created store's secret, needing no store source |
| TestPlatformFlagHitOverridesInference | - | an explicit `--platform` is used even when another section is also present in env.yaml |
| TestPlatformFlagMissingSectionIsLoudError | - | a `--platform` value with no matching section fails loud (exit 2), naming both the requested and the present sections, before the runner is invoked |
| TestPlatformAliasesResolveToCanonical | kube / dk / pm | each short `--platform` spelling reaches the same platform binary as its canonical name; the podman case redirects HOME, so its deploy never writes into the real quadlet directory |
| TestPlatformAliasMissingSectionNamesCanonicalSection | - | exit 2: an alias is resolved before the section check, so the error names the `kubernetes:` section to add rather than echoing `kube` |
| TestPlatformUnknownValueListsEverySpelling | - | a bogus value (k8s) is rejected (exit 2) with every accepted spelling listed, canonical and short |
| TestPlatformSpellingsAreDeterministic | - | platformSpellings is built from an ordered slice, not map iteration, so the rejection message cannot vary between runs; canonical names lead |
| TestPlatformAliasesCoverEveryPlatformExactlyOnce | - | every alias maps to a real platform, no alias is declared twice or collides with a canonical name, and the platform abbrev.Set holds exactly the declared short forms |
| TestAliasDeclarationsFollowTheSharedRules | - | the real abbrev sets -- the verbs, each verb's targets and the platforms -- are built by the same builders the binary uses (abbrev.New panics on a declaration that breaks the shared rule: lower-case letters and digits, shorter than the word, unique, never a canonical name), and every declared alias expands to its canonical word; only verbs that model targets have a target set |
| TestPlatformSingleSectionInferred | - | with no `--platform` and exactly one section present, that section is used and echoed to stderr |
| TestPlatformMenuOnMultipleSections | - | with no `--platform` and more than one section, the interactive menu (via the injected prompter) picks the platform |
| TestPlatformMenuNonTTYRefusesWithPlatformHint | - | the menu refuses to block when the prompter reports no terminal, failing (exit 2, platform selection) with an error naming `--platform` instead of hanging |
| TestPlatformZeroSectionsIsLoudError | - | with no `--platform` and no section present at all, exit 2 and the error names all three section keys |
| TestOldPositionalFormsRejectedWithPlatformHint | deploy kubernetes / remove docker / generate podman | passing the platform as a second positional argument is a usage error (exit 2) that points at `--platform`, not resolved as a target |
| TestStatusTargetWordIsRequired | - | a bare `status` prints the target words and the verb's own help page, exits 2, and runs nothing, since neither view is a safe default; the short spellings (cnt, app) are deliberately absent, since aliases are documented only in the markdown docs |
| TestStatusUnknownAndExtraTargetWords | unknown word / a second word | each is a usage error (exit 2) naming the problem, with nothing run |
| TestStatusTargetsMatchModel | - | the drift gate between the modeled target words, the constants the views switch on, and statusTargetArgBracket (which cannot be built from the model, since cliVerbs' own initialiser uses it) |
| TestStatusTargetAliasesResolve | cnt / app / container / all / unknown | resolveTarget maps each alias to its canonical word and passes an unknown one through; `sts cnt` really drives the container view, which costs one get on kubernetes (no preflight) |
| TestStatusRejectsImpossibleFlagCombinations | unknown --output / json with watch / --all with --pod / --all with --container / --install, --user, --management-port and --raw on the container view | every combination that cannot mean anything is refused (exit 2) before a single query runs, rather than being silently ignored |
| TestWatchFlagAcceptsBareAndInterval | bare / interval / off / 3 rejected values | the flag is boolean in every documented sense (IsBoolFlag) but also takes the deliberately undocumented `-w=<seconds>` form, bounded |
| TestWatchStatusRedrawsUnderTheHeader | - | on a terminal each refresh opens with the shared clear-screen sequence and the `Every 5s: solmq-conn-util status all   <time>   Ctrl-C to stop` header, then the report |
| TestWatchStatusAppendsSectionsWhenNotATerminal | - | off a terminal each refresh is one `=== <time> ===` section line followed by the report, no escape byte and no redraw header |
| TestWatchStatusSleepsTheIntervalAfterEachRefresh | - | the pause is the flag's interval, taken once after each refresh, and each refresh collects once |
| TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn | - | a refused collection goes back to the shared loop, which warns once per refresh on stderr with `warning: <err> -- trying again in 5s` on one line (the error's trailing newline trimmed) and no `error:` line, and the next refresh still runs; exit 0 |
| TestWatchStatusCancelledContextExitsZero | - | Ctrl-C (a done context) ends the watch with exit 0 |
| TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected | - | nothing is drawn before the first collection finishes, the second collection runs with exactly the first frame on screen, and each frame is the clear, the header and the report written together |
| TestWatchStatusDrawsNothingForARefreshCutShort | - | Ctrl-C during a collection draws nothing, not even the header, and the watch exits 0 after that one collection |
| TestWatchStatusRejectsAnIntervalTheLoopCannotUse | - | an interval the shared loop refuses is reported through errExit (non-zero) before anything is collected |
| TestWatchTitleNamesTheViewAndItsFlags | container / application / all / details / all instances / both | the header names the view and appends `--details` and `--all` when set |
| TestStatusContainerViewReadsEngineFactsWithoutExecing | - | one read-only `get pods -o json` answers discovery and the whole table together (1 call in all, no permission probe first), nothing is exec'd into, and every column reaches the output |
| TestStatusContainerDetailsSamplesAndChecksComponents | - | `--details` adds one sampling call for the run and one presence check per distinct referenced object (deduplicated across pods), plus the NODE column, digest and resource lines |
| TestStatusContainerDetailsWithoutMetricsServerDegradesToANote | - | a cluster with no metrics API costs the resource lines and nothing else: a note naming what to install, the table still printed, exit 0 |
| TestStatusDockerContainerViewIsOneInspect | - | one inspect answers every docker target and carries the compose project too; docker reports HEALTH where kubernetes reports READY |
| TestStatusPodmanRestartCountComesFromSystemd | - | the quadlet truth: the count in the table comes from `systemctl show ... NRestarts`, not from podman's own counter, and the inspect fixture's current-shape Health block reaches the HEALTH column |
| TestStatusPodmanRestartCountFallsBackWhenSystemdCannotAnswer | - | a container systemd knows nothing about keeps the container's own counter, nothing fails, and an inspect carrying no healthcheck still reads n/a |
| TestStatusAllSearchesByImage | kubernetes searches every namespace / docker lists then inspects the matches | `--all` finds instances by image reference: `--all-namespaces` plus a client-side filter on kubernetes (with a NAMESPACE column), `ps --all` then an inspect of only the matches on docker |
| TestStatusAllWithNoMatchIsActionable | - | an empty search names the image it looked for, since there is no env.yaml in play to point at |
| TestStatusApplicationViewRunsTheScriptAndRendersItsFacts | - | the exact application block: the unchanged banner, values aligned in one column, right-aligned workflow ids, and no container table |
| TestStatusApplicationDetailsAddsTheEnrichmentLines | - | both levels run the script with `--raw`; `--details` adds `--details` and renders uptime/version/java/config/heap (raw bytes rendered to 412Mi of 1Gi) and the health components; the basic level runs it without the argument and renders none of them, even from a script that still prints them |
| TestStatusFailedScriptRunStillGetsItsOwnBlock | - | an instance whose script could not run keeps its banner with the failure as a body line, the container table above it explains why, a reachable instance in the same run still reports, and the exit code is 1 (one script worker, so the ordered answers land in report order) |
| TestStatusInstallPaths | --install installs without asking / prompt answered yes installs / prompt declined skips the instance and exits 1 | a run that finds the script absent, then the install and a second run, with the declined case reporting the reason in the instance's own block rather than on stderr |
| TestStatusInstallPromptNonTTYRefusesWithInstallHint | - | the install confirmation refuses to block when the prompter reports no terminal, pointing at `--install` and installing/running nothing |
| TestStatusScriptStillMissingAfterInstallFails | - | a run that still finds the script absent after a successful install reports that in the instance's block and exits 1, with no second install |
| TestStatusStandbyIsAnAnswerNotAFailure | - | standby prints like any other answer (the script always exits 0) and the run still exits 0 |
| TestStatusRawPrintsTheHealthDocumentInFull | table / json / without raw / an older script sends no document | `--raw` prints the instance's document in full under its block in place of `health-detail`, and is the `healthDocument` object in JSON; without it neither appears; an instance that sent no document gets a `status:` line and the run still exits 0 |
| TestStatusRunningWorkflowNotConnectedFailsTheRun | active / stopped workflow on a standby | a running workflow whose MQ binder is reconnecting is marked, the binder and its error show at the basic level, and the run exits 1; the same binder beside a stopped workflow marks nothing and exits 0 |
| TestStatusRunsTheScriptsConcurrently | - | over five instances with two workers, the first two execs overlap (a serial run would time out at the barrier), no more than two are ever in flight, and each instance's JSON block carries its own answer |
| TestStatusJSONOutputIsOneDocument | - | `--output` json emits one parseable document carrying schemaVersion and both halves of each instance |
| TestStatusDockerProjectMismatchIsReported | a different project is a note | the container's compose-project label disagrees with docker.project-name, so a status: note names both projects and the way out; exit stays 0 |
| TestStatusDockerProjectMismatchIsReported | the configured project says nothing at all | a matching label is silent |
| TestStatusDockerProjectMismatchIsReported | no compose label is not a mismatch | a container compose never created carries no label, which is not drift and must stay silent |
| TestStatusImageMismatchIsReportedAtBothLevels | basic reports it as a note / details reports it per instance / a matching image says nothing at all | the failed-rollout finding surfaces at both levels -- a run-level note where the per-instance detail block is not printed, the image-expected line where it is, and nothing at all when the running image is the configured one |
| TestStatusDockerDetailsAddsDigestAndStats | - | on docker/podman the digest lives on the image, so `--details` costs an `image inspect` plus the `stats --no-stream` sample (4 calls in all), and both reach the report |
| TestStatusRejectsUnsafeUserBeforeAnyExec | 4 names | a `--user` carrying '/', '$', a space or a quote is rejected (exit 2) via validate.SafeActuatorUser before any exec, since the name reaches a sed address in the script |
| TestStatusTargetValidationRejectsBadPodAndNamespace | bad pod name / bad namespace | an unsafe `--pod` or `--namespace` value is rejected (exit 2) via validate.SafeToken before any exec |
| TestStatusManagementPortBounds | -1 / 65536 | an out-of-range `--management-port` is rejected (exit 2) before any exec |
| TestStatusNoPodsFoundNamesTheSelector | - | discovery with nothing matching names the selector, the namespace, and `--pod` -- the things an operator would fix |
| TestNewProgressPicksOneRendering | watch beats verbose and a terminal | `--watch` owns the screen, so its redraw wins over both `--verbose` and a terminal, and the run reports no steps at all |
| TestNewProgressPicksOneRendering | verbose prints without a terminal | `--verbose` is an explicit request, so the step lines are printed into a pipe too -- a `2>steps.log` capture is the point of the flag |
| TestNewProgressPicksOneRendering | a terminal gets the spinner | the default rendering, chosen only from stderr being a character device, since an in-place rewrite means nothing in a file |
| TestNewProgressPicksOneRendering | no terminal and no verbose: nothing | a redirected run with no flag writes no progress at all, so a captured stderr carries only diagnostics |
| TestStderrIsTerminalReportsAPipe | - | the real probe body every other case injects around: the pipe captureStderr installs is not a character device, so the seam answers false rather than being assumed |
| TestProgressOffAndNilWriteNothing | - | a nil receiver and progressOff are silent on both streams, which is what lets every call site be one unconditional line; `pause` still runs its fn in both, so suppressed progress cannot suppress the install confirmation |
| TestProgressSpinnerRewritesOneLineAndErases | - | the exact bytes of the spinner: one carriage-return-prefixed line per step, the erase blanking the widest line written before the next label, nothing on stdout, `stop` idempotent, and the frame goroutine joined with its channels cleared |
| TestProgressSpinnerGoroutineAdvancesAndCountsSeconds | - | the ticker advances the frame in place and appends the whole-second counter once a step passes 1s; after `stop` the next write starts on a blanked line, so the report cannot land on spinner residue |
| TestProgressPauseHandsBackTheStream | - | the live line (the script batch) is erased before the install question is asked on the same stream, and the next step (the install) draws afterwards, so the prompt is never overwritten by a frame |
| TestProgressStepsPauseFinishesTheStepFirst | - | under `--verbose` the step in flight is closed out before the question, so its elapsed time is the time the call took and not the time the operator took to answer |
| TestProgressDoneLinesCarryTheirOwnTime | - | under `--verbose` each call run alongside others prints its own line with its own time at once, while the batch step stays in flight and is timed as a whole; the spinner, the off mode and a nil progress print nothing for it |
| TestProgressElapsedShapes | -1s / 0 / 300ms / 4.1s / 59.5s / 1m / 2m05s | tenths below a minute and `NmSSs` above it, with a negative clock reading clamped to 0.0s rather than printed |
| TestProgressTrimElidesTheMiddle | shorter than the line / exactly the line / middle elided / odd remainder / a real label / no room for the elision / one column / no columns / negative | a label wider than the row is cut in the middle so the call at the front and the `i/N` at the end both survive, and the result never exceeds the columns asked for -- narrower than `...` the label is cut hard, and at zero or less nothing is drawn |
| TestProgressSpinnerNeverExceedsOneRow | - | a 120-character label still leaves every `\r`-delimited segment -- the draws and the blanking alike -- inside the 79-column cap, with the elision and the `i/N` both still present, because a line that wrapped would leave a row the carriage return cannot reach |
| TestStatusVerboseNamesEverySlowCall | - | `status application -v` over two instances prints exactly one `step:` line per slow call, in order (list pods -- kubernetes has no preflight step -- then one status script line per instance with i/N as each finishes, then the batch line around them), each carrying its elapsed time, and none of it on stdout -- and no `resolve instance names` line, since both `--pod` values are names rather than indexes so no enumeration is made |
| TestStatusVerboseUnderWatchSaysWhichWins | -v / --verbose | both spellings reach the field and print the precedence note on stderr rather than a usage error -- the flags do not conflict, one just wins; the note is the last thing the flag checks do, so what ends this run instead is a failing docker preflight (exit 1, that probe the only call made) |
| TestStatusStdoutIsIdenticalWithAndWithoutTheSpinner | - | the stream contract the feature rests on: `status application --output json` yields byte-identical stdout with and without a terminal (and still parses), while only the terminal run writes a spinner to stderr |
| TestVersionOutputShape | - | `version` prints `solmq-conn-util <version> <go version> <GOOS>/<GOARCH>` as its unchanged first line, then the two-line support notice (notice.Short()), exit 0; the package-level version var defaults to "dev" in an un-injected test build |
| TestSupportNoticeInHandWrittenDocs | README.md and every hand-written doc under docs/ (getting-started, configuration, operations, troubleshooting, DESIGN, DEVELOPMENT) | each carries supportNotice (the shared notice package's Markdown block) verbatim, with a blank line either side (a line directly after a blockquote folds into it), ahead of its first `## ` heading, checked with noticetest.AssertDocs; both forms of the notice are plain ASCII |
| TestSupportNoticeInGeneratedDocs | docs/commands.md / docs/abbreviation.md | renderCommandsDoc and renderAbbreviationDoc both emit the notice block ahead of their first `## ` heading; the committed files are held to them by the two DocInSync gates |
| TestSupportNoticeInGeneratorPage | - | solmq-conn-util-generator.html's id="support-notice" element reads exactly as supportNotice.Plain(), sits above the page body, and the page no longer sizes its body with `calc(100vh - 60px)`, which left no room for the strip |
| TestParityDocMatchesTheSharedBody | - | docs/parity.md passes paritydoc.Check: the page keeps the shared module's H1, the support notice and the canonical body (paritydoc/body.md) byte for byte |
| TestAbsPath | absolute input | absPath returns input unchanged when already absolute |
| TestAbsPath | relative input | absPath joins relative path onto base dir |
| TestCommandsDocInSync | - | docs/commands.md equals what the command model renders; -update rewrites it instead of asserting |
| TestInvocationTargetArgs | - | invocation(status, "") still contains statusTargetArgBracket, while invocation(status, tg.Name) for every modeled status target omits it and starts with the resolved target word instead -- a chosen target's own placeholder is never re-appended from Args |
| TestCommandsModelMatchesUsage | - | the summary page is one line per modeled command carrying its description, points at `help <verb>`, shows no alias anywhere (md-only, by decision), and no line exceeds the 100-column budget the page is designed never to wrap in |
| TestAbbreviationDocInSync | - | docs/abbreviation.md equals what the command model renders; -update rewrites it instead of asserting (same `-update` flag as TestCommandsDocInSync -- one registration per package) |
| TestAbbreviationDocCoversModel | - | every verb alias, target alias, platformAliasList short spelling and short flag form has a row on the page, and the page renders no more rows than the model declares -- the check the byte comparison cannot make, since a regenerated file agrees with a renderer that forgot a whole class of abbreviation |
| TestAbbreviationDocTableShape | - | every table row has its header's cell count, counted honouring the `\|` escape tableCell writes, so an unescaped delimiter in a flag Meaning fails instead of silently rendering a broken table |
| TestVerbUsagePages | one subtest per verb | every per-command page carries its Synopsis, description, every target word (and set) with its summary, every modeled flag the verb takes -- each spelling plus its terse Usage text, wrap-tolerantly asserted -- and its example; no alias appears and no line exceeds the width budget |
| TestVerbUsagePages | orphans and platform shorts | every modeled flag is listed by at least one verb (a flag no verb lists would appear on no help page at all), and `--platform`'s Usage text names each short spelling from platformAliasList |
| TestAutoCompleteDispatchPrintsScript | bash / zsh / fish / powershell | `auto-complete <shell>` exits 0, writes the script to stdout, and never reaches the runner |
| TestCompletionGoldenInSync | bash / zsh / fish / powershell | each rendered script equals its snapshot under cmd/solmq-conn-util/testdata/completions; -update rewrites them |
| TestCompletionCoversModel | bash / zsh / fish / powershell | every modeled verb, target, flag spelling and verb alias reaches every shell (fish exempts a verb with no targets/posarg/flags, e.g. version, which has nothing beyond word 1 to normalize), with descriptions in the three shells that show them |
| TestCompletionOnlyDownloadJarHasSets | - | pins the third command level to exactly where the model puts it: no target other than download/jar carries a non-empty Sets list |
| TestCompletionThirdLevelOffersSets | bash / zsh / fish / powershell | once "download jar" (or alias "dl jar") is typed, every renderer offers the mq/syslog sets by name and description |
| TestCompletionThirdLevelUnlocksPosArg | bash / zsh / fish / powershell | the trailing [dir] positional is offered only after all three words (verb, target, set) are typed, never after just "download jar" |
| TestCompletionRecognizesFlagAliases | bash / zsh / powershell | every spelling flag.Parse accepts (-e, --e, -env, `--env`) is in the value-skipping table, so a value is never mistaken for a positional |
| TestCompletionDownloadFlagsDescribed | - | all four download flags (`--url`, `--version`, `--omit-lib-file`, `--include-provided`) are modeled by exact Long spelling, with the description reaching every shell that carries one (bash compgen word lists carry none) |
| TestCompletionOmitLibFileCompletesFiles | bash / zsh / fish / powershell | `--omit-lib-file` completes file paths in every shell, the same value kind `-e`, `--env` already gets |
| TestCompletionShellStructure | bash / zsh / fish / powershell | each script keeps the registration line that makes it load, and the zsh script opens with #compdef |
| TestCompletionVerbAliasesResolveToCanonical | bash / zsh / fish / powershell | each shell's own alias-normalization construct ($verb= case arm, __fish_seen_subcommand_from, $verbAlias[...]) maps every verb alias to its canonical verb name (same fish exemption as TestCompletionCoversModel) |
| TestCompletionVerbAliasesNotOfferedAtWordOne | bash / zsh / fish / powershell | no verb alias appears in the position-1 candidate list (compgen -W, the zsh verbs array, the __fish_use_subcommand lines, the powershell $verbs array) -- recognized everywhere, but never offered on TAB |
| TestCompletionValueKindsReachScripts | bash / zsh / fish / powershell | a path flag completes files and `examples` completes directories in every shell |
| TestCompletionOutputIsPlainASCIILF | bash / zsh / fish / powershell | generated scripts are plain ASCII, LF only, newline-terminated |
| TestCompletionModelMetadataComplete | - | every verb has a description and a known PosArg, every flag a known Arg and a non-empty Meaning, every modeled shell a renderer and a snapshot; verb/target names and verb/target aliases stay [a-z0-9-] for unquoted case patterns, no alias collides with another verb, another alias, a target under the same verb, or -h/`--help`, and no description carries an apostrophe -- fish escapes it, powershell doubles it and zsh passes it through bashQuote, so the raw text would never appear in those scripts |
| TestPlainText | code spans stripped / newline folded / tab and CR folded / whitespace runs collapse / trimmed / control chars dropped / empty / only backticks / punctuation preserved | model text is reduced to a single-line tooltip that cannot break the enclosing shell statement |
| TestShellQuoting | plain / empty / apostrophe / backslash / dollar and backtick / double quote / semicolon and pipe | bashQuote, fishQuote and psQuote each neutralize their shell's escape rules |
| TestZshEntry | plain / colon in the value escaped / colon in the description left alone / apostrophe quoted / empty description | _describe entries split on the intended colon only |
| TestFlagAliasesAndOffered | short and long pair / long only | flagOffered suggests the documented spellings, flagAliases lists all four dash forms, fishFlagSpec renders -s/-l correctly |
| TestTestCatalogSnapshotInSync | - | this doc's own Snapshot line (test functions, case rows, packages) matches three facts computed independently: a repo-wide walk counting `^func Test` in every *_test.go file (skipping .git/testdata/graphify-out), the doc's own case rows (pipe-prefixed lines minus header+separator per table), and its package sections (`##` headings starting with internal/ or cmd/); each mismatch names which count drifted and how to recompute it |
| TestParseTestCatalogSnapshot | well-formed line among other prose | extracts 741/1009/18 from a snapshot sentence embedded in surrounding text |
| TestParseTestCatalogSnapshot | no snapshot line present | returns an error rather than zeros |
| TestIsTableSeparatorRow | separator row | an all-dash/colon/pipe line between the bounding pipes reports true |
| TestIsTableSeparatorRow | header row | a header row of real cell text reports false |
| TestIsTableSeparatorRow | data row | an ordinary data row reports false |
| TestIsTableSeparatorRow | data row containing a dash | a dash inside real cell text does not misread a data row as a separator |
| TestIsTableSeparatorRow | not pipe-prefixed | a line with no leading pipe reports false |
| TestIsTableSeparatorRow | bare pipe | a single unclosed pipe reports false |
| TestCountDocShapeCountsCaseRowsAndPackageSections | - | a hand-built 2-table/2-package fixture doc yields caseRows=3 and packageSections=2, with "## Contents" excluded as front matter -- a fixture that never needs updating when this doc grows |
| TestCountTestFuncsWalksTreeSkippingDataDirs | - | a fixture tree proves a real top-level func Test counts, the same text under a non-`_test.go` suffix or indented inside a comment does not, and files under testdata/, graphify-out/ and .git/ are never visited |

### logs

The `logs` verb shares status's platform resolution and instance discovery (instances.go), so these cases concentrate on what is its own: the per-platform argv, the combinations it refuses, and the fact that every operator-supplied name is rejected before a process starts. Like the status cases they pin argv and call count, since discovery is one query for every instance.

| Test | Case | Verifies |
|------|------|----------|
| TestLogsKubernetesArgvShape | - | a named pod costs no discovery call at all; the log is read with the namespace and the connector container both named, so a pod with a sidecar cannot have the wrong half read |
| TestLogsDockerArgvShape | - | docker takes its options first and the container name last, and `--since` reaches the argv in the canonical duration form (10m -> 10m0s) rather than as typed; a single instance prints no heading so the output pipes cleanly |
| TestLogsPodmanReadsTheContainerNotTheJournal | - | the quadlet path needs no new binary: `podman logs` reads the container, and neither journalctl nor systemctl appears in any argv |
| TestLogsTailAllAndZero | default / all / zero | `--tail` all is the default and adds no flag, while an explicit 0 is a real request and does |
| TestLogsRejectsImpossibleFlagCombinations | follow with previous / follow with all / all with pod / all with container | each refusal exits 2, names both flags, and runs nothing |
| TestLogsPreviousIsKubernetesOnly | docker / podman | the one refusal that needs the resolved platform: `--previous` is refused with containercli.NoPrevious's message, before the preflight probe, so a flag that cannot work does not first make the operator wait on a daemon |
| TestLogsPreviousReachesTheKubernetesArgv | - | where the concept exists, `--previous` arrives as kubectl's own -p |
| TestLogsFollowReadsTheOneInstance | - | the accepted case: -f reaches the argv and a clean end is exit 0 |
| TestLogsFollowCancelledByTheOperatorEndsClean | - | the operator's Ctrl-C ends a follow with exit 0 and nothing on stderr: the shared runner reports the cancel as an error wrapping `context.Canceled`, and logs reads it as the clean end (skipped on windows, where a process cannot be sent an interrupt) |
| TestLogsFollowFailureIsNotACleanEnd | follow, engine failure | a follow that fails without being cancelled exits 1 and names the instance |
| TestLogsFollowFailureIsNotACleanEnd | follow, canceled error but a live context | an error that merely wraps `context.Canceled` while the follow's own context is live is a failure, not a cancel |
| TestLogsFollowFailureIsNotACleanEnd | one-shot read, canceled error | only `--follow` has a clean end: a one-shot read that wraps `context.Canceled` is exit 1 |
| TestLogsRejectsUnsafeNamesBeforeAnyCall | pod / container / namespace | an unsafe name exits 2 saying why, with zero calls -- the preflight probe included, so a rejected name is never even observed by the platform |
| TestLogsSinceAndTailAreValidatedAtParse | since not a duration / since not positive / since with a metacharacter / tail not a number / tail above the ceiling / tail negative | the two flags carrying a value into an argv are validated at parse, exit 2, nothing runs |
| TestLogsUnexpectedPositionalArgument | - | logs has no target word, so a bare word exits 2 naming `--pod`, `--container` rather than being guessed at |
| TestLogsPlatformMenuWhenSeveralSectionsArePresent | - | an env.yaml with two platform sections cannot resolve itself, so the menu decides; the answer picks the binary and the deployment selector/namespace discovery uses |
| TestLogsPlatformFlagSkipsTheMenu | - | `--platform` is the first step of the resolution order, so it wins before the prompter is consulted, and the instance still comes from that section |
| TestLogsWithoutEnvFileNeedsAnExplicitPlatform | - | the explicit-target exception: instance plus `--platform` needs no env.yaml, while without `--platform` the missing file is reported by name (exit 2) and nothing runs |
| TestLogsNoInstancesFoundNamesTheFix | - | an empty discovery result carries the selector and namespace it used, plus `--pod`, since those are what the operator would change |
| TestLogsWithoutASectionToDiscoverFrom | kubernetes / docker / podman | exit 2: the discovery branch with nothing to work from, reached by naming an instance with the other platform's flag (`--container` on kubernetes, `--pod` on docker/podman); the error names the missing section and the way forward, and must *not* name `--all`, which logs does not have |
| TestLogsPlatformWithNoSectionAtAllIsLoud | - | an env.yaml that parses but describes no platform cannot answer `--platform` either (exit 2), and says so naming all three sections before discovery is attempted |
| TestLogsNamedPodIsNotPreChecked | - | a name is taken at its word: no `get pods <name>` precedes the read (and kubernetes asks no permission first: the read is the only call), and a pod that is not there is reported by the platform on the read itself |
| TestReadVerbsAskKubernetesNoPermissionFirst | status / logs / cli | on kubernetes the first call is the read itself (get pods, logs, exec) and no can-i is ever asked, so a view-only operator is not turned away by a question; a refused read exits non-zero with the platform's own words |
| TestLogsUsesTheSectionCommandOverride | - | logs reaches for the binary env.yaml names (oc, not kubectl) on every call, through the shared instanceCommand resolution |
| TestLogsCommandFlagOverridesTheSection | - | `--command` wins over the section, and a binary outside the per-platform allowlist is refused by name before anything runs, with exit 2 |
| TestLogsAllIsNotALogsFlag | - | `--all` searches by image and returns many, the one thing logs cannot do, so it is an unknown flag here rather than one accepted and then refused |
| TestLogsPickerListsPasteableCommands | - | several discovered and none named: nothing is read, and the matching instances are listed on stdout in sorted order as commands carrying the flags already typed plus the resolved `--platform`, with the index range spelled out; exit 0 |
| TestLogsOneDiscoveredInstanceIsJustRead | - | the picker is for ambiguity only, so a single match is read directly with no listing |
| TestLogsIndexSelectsFromTheSortedList | 0 / 1 / 2 | the pod list arrives unsorted and index 0/1/2 still selects pod-a/pod-b/pod-c, so a dropped sort fails the test |
| TestLogsNameWinsOverIndex | - | a pod genuinely named 0 is reached by name; the index reading applies only when nothing is actually named that |
| TestLogsIndexOutOfRangeNamesTheRange | - | an index past the end says how many matched and what the valid range is, and reads nothing |
| TestLogsIndexWithNothingToIndexInto | - | an index given where discovery has no list says that is the problem, rather than reporting a missing pod named 0, with exit 2 |
| TestLogsPickerCarriesEveryFlagBack | - | every flag already typed comes back in the suggested command (`--platform`, -e, `--namespace`, `--command`, `--allow-command`, `--previous`, `--timestamps`, `--tail`, `--since` in canonical duration form), or pasting a line would silently drop what was asked for |

### cli

The `cli` verb reaches its instance through the same resolution `status` and `logs` use (instances.go), so these cases concentrate on what is its own: the attaching seam and its refusal when a Runner lacks it, the per-platform session argv, the split at `--` that keeps the in-container command out of the flag parser, the non-TTY refusal, and the exit status coming back from the session rather than from this tool. Every refusal asserts zero calls: a mistake in the invocation must cost no process.

| Test | Case | Verifies |
|------|------|----------|
| TestCliRefusesAShellWithoutATerminal | - | the same non-TTY contract the platform menu and the remove confirmation keep, with the one-shot form named as the next step |
| TestCliKubernetesArgvShape | - | `-i -t`, the namespace, `-c connector` and the `--` terminator, in that order; a named pod costs no discovery call, and kubernetes asks nothing first, so the run is the attach alone |
| TestCliEngineArgvShape | docker / podman | engine flags precede the container name, and neither `-c` nor `--` is added -- both would reach the container as arguments |
| TestCliOneShotRunsTheCommandWithNoTerminal | - | everything after `--` replaces the shell, with no `-t` and no `-i` while stdin is a terminal |
| TestCliOneShotAttachesStdinWhenSomethingIsPiped | - | a stdin that is not a terminal is one something is piped into, so `-i` is added; also the case proving the one-shot form needs no terminal |
| TestCliPropagatesTheSessionExitStatus | - | `exit 3` in the session exits 3, the one departure from the 0/1/2 contract |
| TestCliSessionThatCouldNotStartIsAnError | - | a session that never began is exit 1 and names the instance, keeping it apart from a session that ran and failed |
| TestCliRejectsWhatCannotMeanAnything | bare word / separator with no command / two pods / two containers / unsafe command token / unsafe instance name | each refused before any process starts (exit 2, including an unsafe command token or instance name), with a message naming what to do instead; the glob case says where a shell metacharacter can be written instead |
| TestCliPickerCarriesTheCommandBehindTheSeparator | - | the picker keeps the in-container command after its `--` and therefore after the `--pod` being suggested, so a pasted line parses correctly |
| TestCliIndexSelectsFromTheSortedList | - | an index selects out of the same sorted order status and logs print, so a number copied off one verb means the same instance in another |
| TestCliPreflightFailureOpensNothing | - | an unreachable engine is reported once, up front, and no session is attempted |

### remove / instance resolution

The `remove` verb's namespace teardown safety -- the occupancy probe, the two
confirmation questions, and what counts as this release's own object -- and the
three-step instance/namespace/binary resolution that `status`, `logs` and `cli` all
share through `instances.go`.

Tests: [main_test.go](../cmd/solmq-conn-util/main_test.go) (most cases),
[deploy_test.go](../internal/deploy/deploy_test.go) (the namespace-omission and
manifest-parity cases)

| Test | Case | Verifies |
|------|------|----------|
| TestInstanceCommandResolution | override / section / default x kubernetes, docker, podman / nil env / unknown platform | the three-step binary resolution both verbs share; the defaults matter because a run can happen with no section at all (explicit targets), where parse-time defaulting never ran |
| TestInstanceNamespaceResolution | override / section / absent / nil env / docker / podman | the same three steps for the namespace, and that it stays empty on docker/podman, which have no such concept |
| TestConfiguredInstanceName | docker / podman / each section absent / nil env | the single container name each engine section names, and that a missing-section error names the section that is absent rather than just reporting nothing found |
| TestNoPodsFoundExplainsTheBranchItCameFrom | image search / selector / named pods | an empty result is explained in the terms of whichever discovery produced it, so a selector is never quoted back at someone who named pods directly |
| TestRemoveTeardownManifestOmitsTheNamespace | - | the invariant this pins: the manifest piped to `kubectl delete -f -` carries no Namespace, since deleting one cascades to every object inside it including workloads this tool never deployed; the namespace goes by its own `delete namespace <ns> --ignore-not-found` with no stdin, only after the occupancy check and the permission probe passed |
| TestDeployManifestOmitsTheNamespace | - | the manifest a deploy applies carries no Namespace document, while `generate --platform kubernetes` still emits one first |
| TestDeployNamespaceFlow | exists / absent allowed / absent refused / absent unanswered / unknown allowed / unknown refused / unknown unanswered / Namespace apply fails | an existing namespace asks no cluster permission and applies nothing extra; an absent or unreadable one asks `create namespaces` and applies the Namespace on its own before the manifest; absent and refused is MissingNamespace (exit 1, nothing applied); unreadable and refused warns and carries on; an unanswered probe or a failed Namespace apply is an error |
| TestDeployWaitsForTheRollout | rolled out / refused / timed out; failed apply | the deploy waits for the Deployment after the apply and keeps one ok: line; a refused wait warns and exits 0, a timeout exits 1, and a failed apply is not waited on |
| TestKubernetesProbeSets | credentials secret / service and libs claim with volume / no created secret | a deploy asks create and a remove asks delete of each kind the manifest writes -- once each, the Secret only when written, the Service only when enabled, persistentvolumes with --all-namespaces for libs.pvc.create -- and never a namespaces probe other than the gate's delete |
| TestKubernetesRefusedProbeStopsBeforeAnyWrite | - | every probe runs before the first write and the first refusal stops the deploy, naming the permission; the later probes and the apply never run |
| TestRemoveNamespaceOccupiedLeavesItAlone | - | anything living in the namespace that the release does not own is listed as kind/name and the namespace is kept, with no call after the listing |
| TestRemoveNamespaceEmptyPromptsSeparately | y / n / blank | an empty namespace asks its own question, after the delete-namespaces probe, because saying yes to removing a deployment is not saying yes to removing the namespace around it; declining leaves it |
| TestRemoveSharesOneReaderAcrossBothQuestions | - | one input `y\ny\n` answers both the teardown and the namespace question, because the run holds a single buffered reader (the shared confirm package) and the first read does not swallow the second answer |
| TestRemoveNoPromptNeverRemovesAnOccupiedNamespace | - | the invariant: no flag removes a namespace still holding someone else's work. `--no-prompt` approves the questions, never a cascade -- the occupancy check runs first and an occupant of any kind ends it, silently or not |
| TestRemovePlainRunChecksTheNamespace | - | a plain remove with no flags lists the namespace, probes the delete permission and asks both questions -- teardown, then namespace |
| TestRemoveNamespaceRefusesClusterNamespaces | default / kube-system / kube-public / kube-node-lease | refused before the occupancy listing even runs, so no emptiness result can authorise deleting one |
| TestRemoveNamespaceProbeFailureLeavesItAlone | - | a failed listing leaves the namespace in place and says so, exit 0; a namespace is not worth deleting on a guess |
| TestRemoveFailedTeardownSkipsTheNamespace | - | a failed teardown never probes or deletes the namespace: whatever did not come down is still in there |
| TestTeardownDropsTheNamespaceDocument | - | apply keeps the Namespace document, a teardown drops it, dropping the first document leaves no leading separator, and the ConfigMap and Deployment still render |
| TestNamespaceManifestMatchesWhatRenderEmits | - | the manifest `generate` writes starts with exactly the shared kube.NamespaceManifest text a deploy applies on its own, then a separator, so the two cannot drift |
| TestRenderRefusesANamespaceThatIsNotALabel | - | Render writes the Namespace document through kube.NamespaceManifest, so a namespace that is not a DNS-1123 label (validate refuses it first) gets no manifest and an error naming it |
| TestOwnedObjectsCoverEverythingThisReleaseCreates | created secrets/PVC / referenced ones | the safety net behind the occupancy check: CreatedNames carries everything this release creates, with its kind, under the name it was actually created with (derived from deployment.name, never a retired name key), and objects merely referenced (existing:) are *not* ours -- the operator manages them and their presence is a real reason to keep the namespace |
| TestRemoveNamespaceWithoutANamespaceSaysSo | - | a kubernetes section with no namespace has nothing to remove, and says so rather than acting on an empty name |
| TestRemoveKubernetesSeamRejectsUnsafeCommand | - | a remove whose command the guard refuses exits 2 with nothing run: the namespace gate takes the argv the command was guarded into, so there is no second path around the allowlist |
| TestRemoveNamespaceUnreadableProbeLeavesItAlone | not JSON / no items (`{}`) / empty | a listing that cannot be read, has no items, or came back empty is an error from nsgate.Occupants and the namespace is kept with a warning naming why -- it must never read as "empty" and authorise the delete |
| TestRemoveNamespaceNonTTYKeepsItNamingTheFlag | - | the namespace question with no terminal keeps the namespace with a warning that names `--no-prompt` and exits 0 -- the gate never fails a remove that succeeded; nothing is deleted |
| TestRemoveNamespaceProbeArgvIsOneQuery | - | one nsgate.List call listing every kind whose loss would matter, in the namespace being considered |
| TestRemoveNamespaceDiscountsWhatTheReleaseOwnsByKind | - | the release's own objects are passed with their kinds: the Deployment, its ConfigMap and the ReplicaSet and Pod it owns do not count, but a ConfigMap that merely shares the Deployment's name does |
| TestRemoveNamespaceAsksThePermissionBeforeTheQuestion | refused / unanswered | `auth can-i delete namespaces` runs after the listing and before the prompt, so nobody is asked a question that would be refused; a refusal or no answer keeps the namespace with a warning and exit 0 |
| TestRemoveNamespaceDeleteOutcome | deleted / delete fails | a confirmed delete reports "ok: remove namespace kubernetes"; a failed delete is a warning that keeps the namespace, exit 0, with neither an ok nor an error line |
| TestRemoveKubernetesSeamPromptsBeforeTearingDown | y / yes / YES / Y | an accepted confirmation runs the full teardown: preflight, delete, occupancy probe and, since the same answer approves the namespace question, the namespace delete |
| TestRemoveDeclinedTouchesNothing | n / no / blank / whitespace / anything else | a declined teardown exits 0 having run nothing at all, not even the read-only preflight probe, and says "cancelled"; a blank line declines, so the safe answer is the one Enter gives |
| TestRemoveNoPromptSkipsThePromptEntirely | - | `--no-prompt` writes no question at all (asserted on the injected prompter's output), and covers both questions: the empty namespace is removed without asking |
| TestRemoveNonTTYFailsFastNamingTheFlag | - | a non-TTY refusal exits 1 naming `--no-prompt` rather than blocking on a read that will never return, the same shape the platform menu and the status install confirmation use; nothing runs |
| TestRemovePromptNamesWhatItWillDestroy | kubernetes / docker / podman | the question carries the identifier that tells one target from another -- deployment name and namespace on kubernetes, the container on docker, the container and its .service unit on podman |
| TestDeployKubernetesNeverPromptsAndRefusesNoPrompt | - | a kubernetes deploy never writes a prompt, and `--no-prompt` there is a usage error (exit 2) saying kubernetes deploy asks nothing, with no command run, as solace-util refuses it |
| TestDeployDockerUnchangedRunningAsksNothingAndRunsUp | - | a running container whose deploy digest label equals this deploy's asks nothing, even with no terminal, and compose is told a plain `up -d` (no `--force-recreate`) with the digest in the child env: calls info, ps, container inspect, compose |
| TestDeployDockerNotRunningAsksNothing | absent / another container / a sibling | a container that is not running (the name matched exactly, so `solmq-conn-2` is not `solmq-conn`) is never inspected or asked about |
| TestDeployDockerChangedAsksBeforeCompose | a different digest / no label / the inspect fails / the inspect is not json | each counts as changed: the deploy asks, naming the container and that compose will recreate it, and "y" lets compose run |
| TestDeployDockerNoPromptApproves | - | `--no-prompt` approves with no question and no terminal, and compose runs |
| TestDeployDockerNonTTYFailsNamingTheFlag | - | no terminal and no flag exits 1 naming `--no-prompt` and the redeploy, with nothing written to the prompt and compose not run |
| TestDeployDockerDeclinedRunsNothing | n / blank / maybe | anything but y or yes prints "deploy: cancelled; nothing changed" and exits 0, and compose does not run |
| TestDeployDockerFailedRunningProbeIsAnError | - | a listing that cannot be read is exit 1 naming the container and carrying the engine's text, with no question and no compose |
| TestDeployPodmanUnchangedRunningIsLeftAlone | - | with the unit unchanged, every secret reading back as this value and the unit active, nothing is stored, removed, reloaded or restarted, no terminal is needed, the unit file is not rewritten, and the output says it is unchanged and names `solmq-conn-util restart` |
| TestDeployPodmanUnchangedStoppedIsStarted | - | nothing changed and the unit is inactive: it is started, with no secret written and no reload |
| TestDeployPodmanChangedSecretAsksThenWritesOnlyIt | a different value / no such secret | one secret whose stored value differs or is missing asks, naming it and never a value, then only that secret is stored, the unchanged unit and the matching secrets are left as they are, and the unit is reloaded and restarted; the value never appears in the output or an argv |
| TestDeployPodmanOlderPodmanCountsEverySecretChangedAndProbesNone | 4.6.2 / not a version | below 4.7, or a version that cannot be read, no `--showsecret` read is made, every secret is stored again with remove then create (no `--replace`) |
| TestDeployPodmanReadsTheVersionThenEachSecretBack | - | on 5.8.2 the version is read first, then one `secret inspect --showsecret --format {{.SecretData}}` per secret |
| TestDeployPodmanChangedUnitAsksAndWritesOnlyIt | - | a changed unit alone asks (a `reloading` unit counts as running), is the one thing written, and no secret is touched |
| TestDeployPodmanDeclinedWritesNothing | - | a "no" exits 0 with the cancelled message, the unit file as it was, and no secret created or removed, no version probe, no reload, no start or restart |
| TestDeployPodmanNonTTYFailsNamingTheFlagAndWritesNothing | - | no terminal and no flag exits 1 naming `--no-prompt` and the redeploy, with the unit as it was and nothing stored or restarted |
| TestDeployPodmanNoPromptApproves | - | `--no-prompt` approves with no question: the changed unit is written and the running service restarted |
| TestDeployPodmanChangedButStoppedAsksNothing | - | a changed deploy of a service that is not running asks nothing, stores every secret and starts it |
| TestSessionRootless | podman default / no env.yaml / no podman section / rootful by a non-root user / docker / kubernetes | the rootless answer that picks a failed preflight's hint: the deploy rule without the session repair, false off podman and where the rule refuses |
| TestPodmanChangeSummary | unit / secrets / both | the question names the unit and the changed secrets by name, nothing else |
| TestRemoveTargetDescriptions | kube with/without namespace, section missing, env nil, docker, podman | removeTarget names the identifier that tells one target from another for every platform, including the defensive branches the end-to-end tests cannot reach (a section-less env, and a namespace-less deployment that validate rejects moments later) |
| TestStatusAcceptsAnIndexToo | - | status resolves `--pod` 1 to a real name before querying, so the two verbs share one vocabulary and no index ever reaches an argv |
| TestStatusKeepsPodRepeatable | - | the one-entry limit is a logs rule: status reports many instances by design and still accepts `--pod` twice |
| TestStatusAllSkipsAnUnsafeNameOutLoud | - | under `--all` the names come from ps and are checked with containercli.ValidName, so one that could not go into an argv (a `;`, or a terminal escape) is skipped with a note naming it, quoted so an escape prints as text, never reaches any argv, and the instances around it are still reported |
| TestStatusAllSortsRowsAlphabetically | - | docker ps returns creation order, so the rows are sorted by name; without it an index would select a different instance than the row an operator counted to |
| TestIsIndex | 0 / 12 / empty / 0-abc / abc / -1 / 1.0 / leading space | only a bare run of digits is an index, so an instance named 0-abc is still reachable by name |
| TestResolveOneIndexRules | plain name / name that is digits / in-range index / unknown name / past the end | an exact name always wins over the index reading, a digit run in range selects positionally, an unknown name passes through for discovery to judge, and one past the end names how many matched and the valid range |
| TestSortIsByNameThenNamespace | - | the order an index selects into, for both instanceRef and statusreport.Instance: by name, tie-broken on namespace for a cross-namespace status `--all` where two deployments run pods of the same name |
| TestNamingHintIsVerbAware | status/logs/cli x kubernetes/docker/podman | a hint never names a flag the verb it came from does not accept: status offers `--all`, logs and cli have no such flag |

