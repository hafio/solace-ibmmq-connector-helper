# solmq-conn-util test catalogue

Every test in the suite, grouped by package and expanded to individual cases, so you can
see what behavior is covered and jump to the test that covers it. This is a living
document: when a test or case is added, removed, or renamed, update the matching row in
the same change (S5/S6). For build and release details see the
[development guide](../docs/DEVELOPMENT.md); for using the tool, the
[user guide](../userguide.md).

Run the suite with `./scripts/dev.sh test` (`.\scripts\dev.ps1 test` on Windows);
measure coverage with the `cov` task.

## How the suite is built

- **Table-driven tests** iterate a list of cases; each case is its own row below.
- **Golden-file tests** assert generated output byte-for-byte against fixtures under
  [`testdata/golden/`](../testdata/golden) (driven by `internal/gen/golden_test.go`);
  the deterministic ordered emitters make that output stable.
- **The exec seam** (`internal/runner`) is faked by `fakeRunner`, which records the argv
  and stdin crossing the boundary instead of starting a process; the real `os/exec` path
  is exercised through the `TestHelperProcess` child-process pattern.
- Tests are cross-referenced by file and test name only -- no line numbers (they rot as
  tests move).

_Snapshot: 270 test functions, 599 cases across 14 packages._

## internal/spec

Parse env.yaml into the typed model -- workflows, defaults, named connections, the kubernetes/docker/podman target sections, and ports -- and apply section defaults.

Tests: [spec_test.go](../internal/spec/spec_test.go), [env_test.go](../internal/spec/env_test.go), [targets_test.go](../internal/spec/targets_test.go), [expand_test.go](../internal/spec/expand_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestParseWorkflowSolaceAndMQ | - | parses full solace source and mq target with dest kind, tls, key alias, props |
| TestParseWorkflowConnRef | - | mq source with conn-ref resolves ConnRef, DestKind queue, Dest A.IN; SetsConnFields false |
| TestBaseName | url / posix / windows / bare / empty | one shared BaseName splits on both separators, so a Windows-authored path resolves identically on Linux |
| TestConnRefSideMayTuneBinding | - | consumer block parses on a conn-ref side, SetsConnFields ignores it, and Resolve keeps it alongside the referenced tuple and destination |
| TestParseDefaultsConnectionsAndLeaderElection | - | parses 2 named connections and leader-election active_standby with fail-over |
| TestResolveConnRef | known ref edge | resolves host/msg-vpn/key-alias from connections map, keeps dest |
| TestResolveConnRef | unknown ref nope | returned unchanged with ConnRef nope and empty Host |
| TestParseWorkflowEnabledDefaultsTrue | - | enabled defaults true and target stays unset when absent |
| TestParseWorkflowAmbiguousSystemAndDest | solace and mq both set | HasSystem returns false when both systems present |
| TestParseWorkflowAmbiguousSystemAndDest | queue and topic both set | DestKind is empty string when queue+topic ambiguous |
| TestParseWorkflowSyntaxError | - | malformed yaml returns non-nil error |
| TestParseDefaultsFull | - | parses tls stores, management port 8090, security disabled with 1 user, leader-election standalone, logging/solace-defaults nodes captured |
| TestParseDefaultsSecurityDefaultsEnabled | - | security present with empty users defaults Enabled true |
| TestParseDefaultsEmpty | - | empty input yields zero-valued Management/Security/TLS (Present false, Truststore nil) |
| TestParseDefaultsError | - | malformed tls yaml returns non-nil error |
| TestParseKubernetesReplicasDefault | - | deployment without replicas defaults Replicas to 1 |
| TestParseKubernetesFull | - | parses replicas 2, service enabled port 8090, secrets credentials source env, secrets stores create present |
| TestParseKubernetesError | - | deployment as sequence instead of map returns non-nil error |
| TestParseKubernetesResources | - | parses deployment resources CPU '1' and Memory 1Gi |
| TestParseKubernetesLoggingLibsDefaults | syslog and libs download present | syslog Protocol defaults to udp and libs download Image defaults to busybox:1.37 |
| TestParseKubernetesLoggingLibsDefaults | libs pvc create without storage | pvc create Storage defaults to 1Gi |
| TestParseKubernetesLoggingLibsDefaults | no logging or libs block | Logging and Libs stay nil when absent |
| TestParseEnvEmpty | - | empty file yields Workflows dir '.' pattern '*', Kubernetes/Docker/Podman nil, defaults zero-valued |
| TestWorkflowsFromRawDefaultWhenAbsent | - | workflows section absent defaults dir '.' and file pattern '*' |
| TestWorkflowsFromRawDirOverride | - | dir override /custom/dir applied, file pattern stays default '*' |
| TestWorkflowsFromRawFilePatternOverride | - | file_pattern override *.yaml applied, dir stays default '.' |
| TestParseEnvUnknownKeyIgnored | - | unknown top-level key is silently ignored, no error, docker section still parses |
| TestParseEnvWrongScalarTypeErrors | - | non-integer management.port errors, message contains 'cannot unmarshal' |
| TestParseEnvPortsValid | bare int 8090 | parses Host=8090 Container=8090 String()='8090:8090' |
| TestParseEnvPortsValid | host:container 8080:8090 | parses Host=8080 Container=8090 String()='8080:8090' |
| TestParseEnvPortsValid | padded host:container with spaces | trims spaces to Host=8080 Container=8090 String()='8080:8090' |
| TestParseEnvPortsInvalid | non-integer 'abc' | error 'env.yaml: ports entry "abc" must be an integer or "host:container"' |
| TestParseEnvPortsInvalid | more than one colon '1:2:3' | error 'env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)' |
| TestParseEnvPortsInvalid | non-integer host and container 'a:b' | error 'env.yaml: ports entry "a:b" must be "host:container" with integer ports' |
| TestParseEnvPortsInvalid | mapping node {a: 1} | error 'env.yaml: ports entry must be an integer or "host:container", got a !!map' |
| TestApplyDockerDefaultsFillsMissing | - | docker defaults command/name/restart applied, ports default to mgmt port pair, stores/libs stay nil |
| TestApplyDockerDefaultsOverrideWins | - | explicit command/name/restart/ports override defaults exactly as given |
| TestApplyPodmanDefaultsFillsMissing | - | podman defaults command/mode run/name/restart/ports applied and Quadlet non-nil with scope auto and empty dir |
| TestApplyPodmanDefaultsOverrideWins | - | explicit command/mode quadlet/name/restart/ports/quadlet scope+dir override defaults exactly |
| TestApplyMountDefaultsFillsMissing | - | stores mount-path defaults, libs dir kept and mount-path defaulted |
| TestApplyMountDefaultsOverrideWins | - | explicit stores and libs mount-path overrides retained |
| TestExpandBracedVar | - | `${HOST}` in Side.Host expands from Lookup |
| TestExpandDefaultVarSetUsesValue | - | `${VPN:fallback}` uses the looked-up value when VPN is set |
| TestExpandDefaultVarUnsetUsesDefault | - | `${VPN:fallback}` falls back to the default when VPN is unset |
| TestExpandUnsetNoDefaultPassesThroughWithWarning | - | unset defaultless `${TYPO}` passes through verbatim and Warn is called exactly once naming TYPO |
| TestExpandBareDollarVarUntouched | - | bare `$VPN` (no braces) is left untouched even though VPN is set |
| TestExpandCredentialFieldLeftAlone | - | Side.Password/PasswordEnv (`expand:"no"`) never expand |
| TestExpandYAMLNodePassthroughLeftAlone | - | a `*yaml.Node` field (APIProps) is never walked or rewritten |
| TestExpandDefaultsConnectionsMapEntry | - | a `${HOST}` value inside Defaults.Connections (map[string]Side) expands via read-modify-write |
| TestExpandNilLookupDisablesEverything | - | nil Lookup makes Expand a no-op, leaving `${HOST}` untouched |

## internal/validate

Validate the parsed model -- per-side rules, connection refs, leader election, the docker/podman/kubernetes target sections, ports, container names, TLS/stores wiring, and the safe-token charset.

Tests: [validate_test.go](../internal/validate/validate_test.go), [validate_extra_test.go](../internal/validate/validate_extra_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestBinderIdentityUsesTheCredentialPair | different -env usernames | two -env usernames on one host are different binders, so no false key-alias conflict |
| TestBinderIdentityUsesTheCredentialPair | same -env username | one binder with two key-aliases still conflicts |
| TestWorkflowCap | 21 workflows | fatal error naming the count, the 20 cap, and the split-into-separate-folders remedy |
| TestWorkflowCap | 20 workflows | exactly at the cap does not error |
| TestDeployNameTooLong | 57-char name | deployment.name + "-config" exceeds the 63-char DNS-1123 limit |
| TestDeployNameTooLong | short name | a short deployment name has no length error |
| TestValidGoldenLikeInputPasses | - | valid solace->mq queue workflow produces no errors |
| TestMissingSourceTarget | - | workflow with neither side set errors missing 'source' and missing 'target' |
| TestExactlyOneSystem | - | side with empty system errors exactly one of 'solace:' or 'mq:' |
| TestExactlyOneDestination | - | side with empty dest-kind errors exactly one of 'queue:' or 'topic:' |
| TestSolaceTopicSourceWarnsNotErrors | - | solace topic source allowed with no errors, warns non-durable subscription |
| TestConnNameFormat | - | bad mq conn-name errors host(port) format message |
| TestKeyAliasNeedsKeystore | - | solace key-alias without keystore errors no keystore defined |
| TestKeyAliasConflict | - | same solace tuple with different key-alias errors conflicting key-alias |
| TestLeaderElectionActiveStandbyValid | - | valid active_standby leader election config passes with no errors |
| TestLeaderElectionActiveMissingQueueAndSession | - | active_active mode missing queue/conn-ref errors requires a 'queue' and requires a solace session |
| TestLeaderElectionConnRefMustBeSolace | - | leader-election conn-ref pointing to mq connection errors must be a solace connection |
| TestLeaderElectionInvalidMode | - | unknown leader-election mode 'bogus' errors is invalid |
| TestMQCipherRequiresTLS | - | mq cipher set with tls false errors require 'tls: true' |
| TestDeployKubeChecks | - | bad name and empty image error not a valid DNS-1123 label and deployment.image is required |
| TestCredentialsEnvChecks | - | unset env var for credentials errors variable MISSING_VAR is not set |
| TestCheckSideMQMissingFields | - | mq side missing conn-name/queue-manager/channel errors each; user/password not flagged missing |
| TestCheckSideSolaceMissingAndBadScheme | missing-host-vpn | empty solace side errors missing host and msg-vpn; client creds not flagged |
| TestCheckSideSolaceMissingAndBadScheme | bad-scheme | http scheme host errors must start with tcp:// or tcps:// |
| TestSolaceKeyAliasRequiresTCPSAndKeystore | - | solace key-alias with plain tcp host errors requires a tcps:// host |
| TestMQKeyAliasRequiresKeystore | - | mq key-alias without keystore errors no keystore defined |
| TestCheckKubeRequiredAndReplicas | - | kube deployment missing name/namespace and replicas 3 errors each field plus replicas: 1 message |
| TestCheckKubeCredentialSources | source-env-empty-vars | credentials source env with no variables errors non-empty 'variables' |
| TestCheckKubeCredentialSources | source-file-no-values | credentials source file without values-file errors requires 'values-file' |
| TestCheckKubeCredentialSources | source-weird | unknown credentials source errors source must be |
| TestCheckKubeStoresRequireTruststore | - | kube stores create without tls.truststore errors requires tls.truststore |
| TestStoresNotWiredWarning | - | TLS workflow with kube deploy and no stores wiring warns secrets.stores is omitted |
| TestStoresWiredExistingNoWarning | - | stores wired via existing secret produces no stores-omitted warning |
| TestUnsuppliedVarsWarning | - | unsupplied ${T} and ${HC} vars each warn is used but not supplied |
| TestSolaceQueueDestinationWarnsNotErrors | - | mq source to solace queue target allowed, warns point-to-point |
| TestIdiomaticSolaceCombosNoEDAWarn | - | idiomatic solace topic-target/queue-source combos emit no EDA warnings |
| TestConnRefStrictOnlyDestination | - | conn-ref side also setting host errors may set only queue/topic |
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
| TestCheckLibs | pvc-create-bad-name | pvc create with Bad_Name errors DNS-1123 |
| TestCheckLibs | download-empty-urls | download with empty urls errors non-empty 'urls' list |
| TestCheckLibs | download-ftp-url | non-http(s) download url errors must be http(s) |
| TestCheckLibs | download-injection-quote | download url with quote/semicolon errors no spaces, quotes, or control characters |
| TestCheckLibs | download-injection-dollar | download url with $() errors no spaces, quotes, or control characters |
| TestCheckLibs | download-bad-pvc-name | valid download url with bad pvc name errors DNS-1123 |
| TestCheckLibs | pvc-existing-valid | pvc.existing set alone has no errors |
| TestCheckLibs | download-valid | valid download urls has no errors |
| TestCheckDocker | valid-docker | valid docker section passes with no errors |
| TestCheckDocker | missing-image-empty-cmd-bad-port | errors docker.image required, command must not be empty, and port 0 must be 1-65535 |
| TestCheckDocker | unsafe-command | docker command with semicolon errors unsafe character |
| TestCheckDocker | stores-no-truststore | docker stores set without truststore errors docker.stores requires tls.truststore |
| TestCheckDocker | libs-no-dir | docker libs set without dir errors docker.libs.dir is required |
| TestCheckDocker | checkdocker-false-gate | docker section not checked when CheckDocker is false |
| TestCheckDockerCredentials | - | docker credentials source env with no variables errors non-empty 'variables' |
| TestCheckPodmanModeAndScope | valid-podman | valid podman section passes with no errors |
| TestCheckPodmanModeAndScope | bad-mode | podman mode 'swarm' errors podman.mode must be |
| TestCheckPodmanModeAndScope | bad-scope | podman quadlet scope 'root' errors scope must be auto, user, or system |
| TestCheckCommandMultiToken | safe-multi-token | docker command with extra safe tokens has no unsafe-character error |
| TestCheckCommandMultiToken | unsafe-token | docker command with $(evil) token errors unsafe character |
| TestCheckDeployCommandAcceptReject | kubectl / oc / kubectl with flags / docker with flag / podman / kubectl.exe / sudo podman with extraAllowed | accept matrix: bare allowlisted argv[0], flag-shaped args, .exe-stripped comparison, and a chained binary approved via extraAllowed all pass |
| TestCheckDeployCommandAcceptReject | curl / absolute path / relative path / bare positional arg / sudo podman without extraAllowed / bare "--" / empty command | reject matrix: unlisted binary, path argv[0], a bare positional argument, an unapproved chained binary, a bare end-of-flags marker, and an empty command all error |
| TestCheckDeployCommandEndOfFlagsMarkerMidCommand | - | "kubectl --" errors token "--": end-of-flags marker is not accepted, distinct from the argv[0] allowlist rejection |
| TestCheckDeployCommandErrorTexts | - | pins the canonical wording verbatim for the path, allowlist, end-of-flags, flag-shape, and empty-command errors |
| TestCheckKubeCommandNowValidated | - | kubernetes.command "kubectl; rm -rf /" now errors (previously skipped entirely for k8s); a safe kubectl command produces no such error |
| TestCheckKubeCommandDefaultKubectlUnvalidated | - | the zero-value default (spec.DefaultKubeCommand) validates clean |
| TestContextAllowCommandsHonored | - | Context.AllowCommands threads into checkKube and checkContainerTarget: "sudo docker"/"sudo podman"/"sudo kubectl" reject with AllowCommands nil, accept with AllowCommands=[sudo] |
| TestCheckContainerCommandUnlistedBinaryRejected | - | docker.command "curl" and podman.command "/tmp/evil" are rejected by the platform allowlist, not merely the charset check |
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
| TestConnectionDefinitionValidation | - | connection with dest set errors must not define queue/topic; incomplete mq connection errors missing 'queue-manager' |
| TestCheckContainerNameRejected | ../evil | rejected for both docker.name and podman.name |
| TestCheckContainerNameRejected | Bad_Name | rejected for both docker.name and podman.name |
| TestCheckContainerNameRejected | valid-default-name | solmq-connector accepted with no docker.name error |
| TestCheckStoresMountPathRejected | custom-mount-path | non-default stores.mount-path errors mount-path is not supported |
| TestCheckStoresMountPathRejected | default-mount-path | fixed default mount-path accepted with no mount-path error |
| TestDockerPodmanTLSWithoutStoresWarning | docker-tls-no-stores | TLS workflow with docker and no stores warns docker.stores is omitted |
| TestDockerPodmanTLSWithoutStoresWarning | docker-tls-stores-wired | docker stores wired suppresses docker.stores omitted warning |
| TestDockerPodmanTLSWithoutStoresWarning | podman-tls-no-stores | TLS workflow with podman and no stores warns podman.stores is omitted |
| TestDockerPodmanTLSWithoutStoresWarning | podman-tls-stores-wired | podman stores wired suppresses podman.stores omitted warning |
| TestUsesTLS | solace-tcps-host | solace side with tcps host returns usesTLS true |
| TestUsesTLS | mq-tls-true-no-tcps | no solace side, mq tls true returns usesTLS true |
| TestUsesTLS | plain-tcp-mq-false | plain tcp solace and mq tls false returns usesTLS false |
| TestMQOnlyTLSStoresOmittedWarning | - | mq-only TLS workflow with docker stores omitted still warns store files missing at runtime |
| TestPlainTCPStoresOmittedNoWarning | - | plain-tcp workflow with stores omitted has no store-files-missing warning |
| TestCheckContainerImageRestartTimezoneUnsafe | newline in image | rejected for both docker.image and podman.image |
| TestCheckContainerImageRestartTimezoneUnsafe | metacharacter in image | `img; rm -rf /` and `img $(evil)` rejected for both targets |
| TestCheckContainerImageRestartTimezoneUnsafe | space in image | rejected for both targets |
| TestCheckContainerImageRestartTimezoneUnsafe | newline in restart | docker.restart rejected |
| TestCheckContainerImageRestartTimezoneUnsafe | backtick in timezone | podman.timezone rejected |
| TestCheckContainerImageRestartTimezoneUnsafe | realistic values | tagged registry image, Asia/Singapore and on-failure:5 all accepted |
| TestCheckContainerHostPathsUnsafe | newline in tls.truststore.file | bind-mounted store path rejected once docker.stores opts in |
| TestCheckContainerHostPathsUnsafe | space in libs.dir | podman.libs.dir rejected |
| TestCheckContainerHostPathsUnsafe | windows paths | `C:\certs\...` store paths and `C:\libs` accepted (backslash and colon permitted) |
| TestCheckKubeSecretNames | cred create bad | non-DNS-1123 credentials create.name rejected |
| TestCheckKubeSecretNames | cred create empty | missing credentials create.name reported as required |
| TestCheckKubeSecretNames | cred existing bad | non-DNS-1123 credentials existing rejected |
| TestCheckKubeSecretNames | stores create bad | non-DNS-1123 stores create.name rejected |
| TestCheckKubeSecretNames | stores existing bad | non-DNS-1123 stores existing rejected |
| TestCheckKubeSecretNames | valid names | solmq-credentials and solmq-tls produce no name error |
| TestCheckLibsNFSFields | newline in nfs.server | rejected against the host charset |
| TestCheckLibsNFSFields | newline in nfs.path | rejected against the host-path charset |
| TestCheckLibsNFSFields | valid server and path | nfs1.corp.example and /solace-libs accepted |
| TestPasswordConflictOnSameBinder | differing passwords | same MQ tuple with two passwords errors conflicting password for the same binder |
| TestPasswordConflictOnSameBinder | identical passwords | same tuple sharing one password passes |
| TestPasswordConflictOnSameBinder | distinct tuples | different queue-manager means different binders, so passwords may differ |
| TestPasswordConflictSolaceSide | - | the solace branch keys on client-password and errors on a conflict |

## internal/consolidate

Build the consolidated binder model from the workflows -- dedup connections, TLS bundles, destination roles, store-path rewriting, leader election, and durable-name UUIDs.

Tests: [consolidate_test.go](../internal/consolidate/consolidate_test.go), [consolidate_extra_test.go](../internal/consolidate/consolidate_extra_test.go), [uuid_test.go](../internal/consolidate/uuid_test.go)

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
| TestFormatScalarQuoting | plain passthrough | FormatScalar returns plain |
| TestFormatScalarQuoting | double-quoted requoting | FormatScalar returns "dq" |
| TestFormatScalarQuoting | single-quote-doubling escape | FormatScalar returns 'a''b' |
| TestFormatScalarQuoting | depth>=2 nested quoted passthrough from parsed YAML | FormatScalar preserves quoting, returns "R1" |
| TestSanitizeAndIsTCPS | sanitize A_b.2-x/y | returns A-b-2-x-y |
| TestSanitizeAndIsTCPS | isTCPS tcps:// and tcp:// | tcps:// true, tcp:// false |
| TestDisplayName | acc with name set | returns the name |
| TestDisplayName | acc with only connName | falls back to connection name |
| TestDisplayName | acc solace kind with vpn | returns solace:myvpn |
| TestDisplayName | acc mq kind with qm | returns mq:QM1 |
| TestMergeProp | append new key b | list grows to 2 entries, no warnings |
| TestMergeProp | overwrite existing key a with new value | value updated to 9, one warning emitted |
| TestMergeProp | overwrite key a with same value again | no additional warning |
| TestAppendPassthroughCollision | passthrough T collides with existing T | output has 2 props and 1 collision warning |
| TestNodeToProps | mapping with scalar and nested keys | 2 props, first key k1/val v1, second has Sub set |
| TestNodeToProps | nil node | nodeToProps returns nil |
| TestBuildMQmTLSBundle | mq TLS side with cipher and keyAlias plus solace target | MQTLS true, 1 bundle, HasKeystore true, KeyAlias mc, KeystoreTyp PKCS12, TruststoreTyp JKS |
| TestBuildCipherConflictWarning | two mq sources with different ciphers C1/C2 | warnings contain conflicting cipher |
| TestBuildMessageLoopWarning | same side used as source and target dest SAME | warnings contain message loop |
| TestBuildSolaceTopicSourceEmitsConsumerTopic | solace topic source -> mq queue target | input-0 solace binding is consumer with DestType topic |
| TestBuildStorePathsRawVsMount | mount=false (config) | TruststoreLoc reflects env.yaml path verbatim ./certs/t.jks |
| TestBuildStorePathsRawVsMount | mount=true (deploy) | TruststoreLoc rewritten to /app/external/classpath/truststores/t.jks |
| TestBuildLeaderElection | dangling conn-ref missing | le non-nil with Mode/Queue set but Session nil, no guard warning |
| TestBuildLeaderElection | conn-ref happy path tcps host | Session host/vpn set and APIProps has SSL_TRUST_STORE, SSL_KEY_STORE, SSL_PRIVATE_KEY_ALIAS mounted paths and alias sc |
| TestBuildLeaderElection | inline session non-tcps host | Session set from inline fields with zero TLS APIProps |
| TestBuildLeaderElection | mount rewrite raw vs mnt for leader election truststore | raw keeps ./certs/truststore.jks verbatim, mnt rewrites to /app/external/classpath/truststores/truststore.jks |
| TestDurableNameGolden | - | DurableName of fixed inputs equals pinned solmq-3631c883-c0c4-5bc8-985e-ea2842831ad6 |
| TestDurableNameDeterministic | same inputs called twice | DurableName returns identical value both times |
| TestDurableNameDeterministic | different file name g.yaml vs f.yaml | DurableName differs when file name changes |
| TestGeneratedSecretNamesStayOutOfChildEnvDanger | - | every SecretRef.Stable Build() can produce (binder creds, security-user passwords, TLS stores, leader-election) matches a fixed-suffix pattern, so adversarial conn-ref/security-user names (e.g. "path", "ld-preload", "LD") can never fold to a bare dangerous docker-compose child-env name like PATH or LD_PRELOAD |

## internal/tls

Resolve TLS store paths -- host source vs the fixed in-container mount dir, separator-agnostic base names, and the Solace api-property store keys.

Tests: [tls_test.go](../internal/tls/tls_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestMountPathSeparatorAgnostic | truststore.jks | unrooted plain name -> MountDir/truststore.jks |
| TestMountPathSeparatorAgnostic | ./certs/truststore.jks | unix relative path -> MountDir/truststore.jks |
| TestMountPathSeparatorAgnostic | .\certs\truststore.jks | windows relative backslash path -> MountDir/truststore.jks |
| TestMountPathSeparatorAgnostic | C:\app\certs\keystore.jks | windows absolute path -> MountDir/keystore.jks |
| TestMountPathSeparatorAgnostic | /abs/unix/path/keystore.jks | unix absolute path -> MountDir/keystore.jks |
| TestMountPathSeparatorAgnostic | mixed/dir\store.jks | mixed slash and backslash path -> MountDir/store.jks |
| TestSolacePropsUseMountedBaseName | - | mount=true rewrites SSL_TRUST_STORE and SSL_KEY_STORE to MountDir base names, not raw backslash paths |
| TestStorePathConfigVsDeploy | mount=false | StorePath returns raw defaults path ./certs/t.jks unchanged |
| TestStorePathConfigVsDeploy | mount=true | StorePath returns MountDir/t.jks base name from backslash input |
| TestSolacePropsRawPathWhenNotMounted | - | mount=false keeps SSL_TRUST_STORE as raw ./certs/truststore.jks path |

## internal/yamlwriter

The indentation-aware line writer every generated artifact is built from. Four packages carried private copies that drifted; one definition keeps their indentation identical.

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
| TestApplicationMinimalNoOptionalBlocks | ssl: | ssl: block absent when defaults empty |
| TestApplicationMinimalNoOptionalBlocks | management: | management: block absent when defaults empty |
| TestApplicationMinimalNoOptionalBlocks | logging: | logging: block absent when defaults empty |
| TestApplicationMinimalNoOptionalBlocks | security: | security: block absent when defaults empty |
| TestApplicationMinimalNoOptionalBlocks | type: undefined | undefined binder is always emitted even with minimal config |
| TestApplicationLeaderElection | - | output contains leader-election, fail-over, management, session and TLS fields for active_standby mode |
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
| TestApplicationSkipsBundleWithoutTruststore | - | tls: true with no tls.truststore emits no ssl bundle or ssl-bundle reference, keeps MQTLS set, and warns |

## internal/deploy

Render the Kubernetes manifests -- Namespace/ConfigMap/Secret/Deployment/Service, syslog, libs (PVC or download), and multi-instance layout.

Tests: [deploy_test.go](../internal/deploy/deploy_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull | - | rendered output contains all expected fragments: configmap, secrets, deployment env/mounts/probes/resources, service; blank line preserved in block scalar |
| TestRenderFull_ExactDocument | - | Render output matches wantRenderFull golden document byte-for-byte |
| TestRenderNoSecretsNoServiceNoTLS | - | output omits JAVA_TOOL_OPTIONS, envFrom, Secret, Service, stores volume, logback-spring.xml, syslog host, libs volume, initContainers |
| TestRenderSyslogUDP | - | UDP syslog config emits SyslogAppender, appname/host/port env vars, logback mount; no LogstashTcpSocketAppender |
| TestRenderSyslogTCP | - | TCP syslog config emits LogstashTcpSocketAppender and destination tag; no SyslogAppender or AsyncAppender |
| TestRenderLibsPVCExisting | - | existing libs PVC yields claimName and mount, no initContainers or PersistentVolume |
| TestRenderLibsPVCCreate | - | created libs PVC emits PV/PVC docs with NFS server/path, storage size, claim name, PV/PVC precede Deployment |
| TestRenderLibsDownload | emptyDir | init container wgets each jar url into /libs, mounted via emptyDir |
| TestRenderLibsDownload | existing PVC download | download target uses claimName dl-pvc instead of emptyDir |
| TestRenderNamespaceAlwaysFirst | - | Namespace doc is first in output and precedes ConfigMap |
| TestRenderExistingSecrets | - | existing creds/tls secrets produce no Secret doc, referencing my-creds and secretName my-tls |
| TestManagementPortFallback | defaults.Management.Present true Port 9999 | managementPort returns 9999 from defaults |
| TestManagementPortFallback | empty Defaults | managementPort returns 8090 from service port |
| TestManagementPortFallback | Service.Port 0, Defaults nil | managementPort returns 8090 default fallback |
| TestQuoteRes | 1 | quoteRes("1") returns quoted "1" |
| TestQuoteRes | 250m | quoteRes("250m") returns unquoted 250m |
| TestQuoteRes | 512Mi | quoteRes("512Mi") returns unquoted 512Mi |
| TestRenderNoResources | - | empty Resources produces no resources: block in output |
| TestServicePortDefaultsToManagementPort | unset follows the management port | service.port omitted renders port/targetPort 9000 from defaults.management.port |
| TestServicePortDefaultsToManagementPort | unset with no management port | falls back to the connector default 8090, never port: 0 |
| TestServicePortDefaultsToManagementPort | explicit port is kept verbatim | service.port 8081 rendered as given |

## internal/dockergen

Render the docker compose file from the target model.

Tests: [dockergen_test.go](../internal/dockergen/dockergen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull_WithEverything | - | full golden output with creds, store, libs, MQTLS, ports, timezone matches exactly |
| TestRenderFull_Minimal | - | minimal golden output omits restart, ports, environment, env_file, volumes blocks |
| TestEnvironmentBranches | TZ only, MQTLS false | environment block contains only TZ: UTC, no JAVA_TOOL_OPTIONS |
| TestEnvironmentBranches | MQTLS only, no timezone | environment block contains only JAVA_TOOL_OPTIONS line, no TZ |
| TestContentIndentationAndBlankLines | nested key indentation | nested line gets extra indent on top of 6-space block indent |
| TestContentIndentationAndBlankLines | blank line preserved | blank line stays empty with no spaces between content lines |
| TestContentIndentationAndBlankLines | no trailing spaces | no rendered line has trailing spaces |
| TestStoresOnlyAndLibsOnly | stores only | volumes block contains only the store mount line |
| TestStoresOnlyAndLibsOnly | libs only | volumes block contains only the libs mount line |
| TestSplitLinesNoTrailingNewline | - | app.yml lacking trailing newline still renders content line with no dropped element |

## internal/podmangen

Render the podman run script and the quadlet units from the target model.

Tests: [podmangen_test.go](../internal/podmangen/podmangen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderRunScriptFull | - | full input renders run script matching exact golden string byte-for-byte |
| TestRenderRunScriptMinimal | - | minimal input renders run script matching exact golden string byte-for-byte |
| TestRenderQuadletFull | - | full input yields 1 unit named solmq-connector.container with content matching golden string |
| TestRenderQuadletMinimal | - | minimal input yields 1 unit with content matching golden string (no Service section, no restart) |

## internal/gen

Orchestrate parse -> validate -> consolidate -> render, resolve credentials/stores, and assert the byte-for-byte golden fixtures.

Tests: [gen_extra_test.go](../internal/gen/gen_extra_test.go), [golden_test.go](../internal/gen/golden_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestLooksDotenv | A=1\nB=2\n | recognized as dotenv format |
| TestLooksDotenv | a: 1\n | yaml mapping is not dotenv |
| TestLooksDotenv | # c\n\n | comments/blank-only lines are not dotenv |
| TestLooksDotenv | URL=http://x:1\n | value containing colon after = is still dotenv |
| TestLooksDotenv | KEY: v=1\n | colon before = means yaml, not dotenv |
| TestParseValuesDotenv | - | parses dotenv into 2 KVs, trims whitespace, skips comments/blanks |
| TestParseValuesYAML | A: 1\nB: two\n | parses yaml mapping into 2 KVs |
| TestParseValuesYAML | - just\n- a\n | yaml sequence (not mapping) returns error |
| TestResolveCredEnv | env vars present | resolves 2 KVs from env source |
| TestResolveCredEnv | env var missing | missing env var returns error |
| TestResolveCredEnv | nil Env resolver | nil env func yields 2 KVs with empty values, no error |
| TestResolveCredFileAndBadSource | file source with ReadFile | resolves 1 KV (A=1) from file |
| TestResolveCredFileAndBadSource | no ReadFile provided | missing ReadFile returns error |
| TestResolveCredFileAndBadSource | ReadFile returns error | read error propagates |
| TestResolveCredFileAndBadSource | source=weird | unknown credential source returns error |
| TestResolveStores | truststore+keystore with ReadFile | resolves 2 store files, first named t.jks |
| TestResolveStores | no ReadFile provided | missing ReadFile returns error |
| TestResolveStores | ReadFile returns error | read error propagates |
| TestResolveStores | no stores configured | empty Defaults yields 0 stores, no error |
| TestNamesAndPaths | credEnvFileName nil creds | returns empty string |
| TestNamesAndPaths | credEnvFileName Existing set | existing filename wins, returns e.env |
| TestNamesAndPaths | credEnvFileName Create set | returns <name>.env |
| TestNamesAndPaths | pathIn base variants | empty base returns bare path; trailing/no-trailing slash both join to /base/a |
| TestNamesAndPaths | instanceName variants | single instance unsuffixed; multi-instance suffixed -1/-2 |
| TestTargetMounts | tls+stores+libs configured | 2 store mounts at fixed default mount path (ignores custom MountPath), libs mount uses resolved abs source and given target |
| TestTargetMounts | nil stores and libs | opt-in mounts yield nil,nil when sections absent |
| TestResolveCredentialsAndEnvFileContent | nil CredentialsSecret | returns nil,nil |
| TestResolveCredentialsAndEnvFileContent | Existing set | returns nil,nil (no creation needed) |
| TestResolveCredentialsAndEnvFileContent | Create with env source | resolves 2 KVs and EnvFileContent renders "A=1\nB=2\n" |
| TestGenerateDockerBasics | - | generates non-empty compose containing image, uses existing env-file name solmq.env |
| TestGeneratePodmanRunAndQuadlet | mode: run | produces run script, no units, env-file solmq-connector.env, 1 app yaml, 1 service |
| TestGeneratePodmanRunAndQuadlet | ForceQuadlet true | produces quadlet units, no run script |
| TestGenerateMissingTargetSection | kubernetes | error contains kubernetes target requires a 'kubernetes:' section in env.yaml |
| TestGenerateMissingTargetSection | docker | error contains docker target requires a 'docker:' section in env.yaml |
| TestGenerateMissingTargetSection | podman | error contains podman target requires a 'podman:' section in env.yaml |
| TestGenValidateAndValuesFileKeys | Validate with TLS but no secrets.stores | no errors, exactly one warning about missing store files at runtime |
| TestGenValidateAndValuesFileKeys | valuesFileKeys with kubernetes create/file source | returns keys map with SOL=true, no issues |
| TestGenValidateAndValuesFileKeys | valuesFileKeys nil kubernetes | returns nil,nil |
| TestGeneratorPageGoldenInSync | - | the golden embedded in solmq-conn-util-generator.html matches testdata/golden/application.yml (regenerate with -update-html-golden) |
| TestGoldenConfig | - | generated config output matches testdata/golden/application.yml byte-for-byte, one instance |
| TestGoldenKubernetesCreate | - | generated kubernetes manifests (namespace, configmap, secret, stores, pv, pvc, deployment with envFrom/stores/syslog/libs, service) match golden fixture byte-for-byte |
| TestGoldenKubernetesNoSecrets | - | generated manifests without secrets/syslog/libs (namespace, configmap, deployment, service) match golden fixture byte-for-byte |
| TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless | - | parse() expands host/msg-vpn from Lookup, leaves client-password-env verbatim, and returns exactly one warning naming TYPO for the unset defaultless conn-name variable |

## internal/runner

The os/exec seam -- ParseCommand safe-tokenizing, kubectl/docker/podman deploy and delete argv, quadlet scope resolution, and WriteFile modes.

Tests: [runner_test.go](../internal/runner/runner_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestHelperProcess | - | helper child-process entry point guarded by GO_WANT_HELPER_PROCESS; dispatches stdin/both/fail/unknown modes |
| TestOSRunWiresStdinToChild | - | OS.Run passes stdin through to child, output equals hello-stdin |
| TestOSRunCombinesStdoutAndStderr | - | combined output contains both stdout-line and stderr-line |
| TestOSRunNonZeroExitReturnsErrorWithOutput | - | non-zero exit returns non-nil error and output still contains before-exit |
| TestOSRunAcceptsAbsolutePathArgv0 | - | absolute path as argv0 runs successfully, output equals abs-argv0-ok |
| TestParseCommand | kubectl | parses to [kubectl], ok |
| TestParseCommand | kubectl --context prod -n solace | parses to [kubectl --context prod -n solace] tokens, ok |
| TestParseCommand | oc | trims whitespace to [oc], ok |
| TestParseCommand | empty string | returns error, nil result |
| TestParseCommand | whitespace only | returns error, nil result |
| TestParseCommand | kubectl; rm -rf / | semicolon rejected as unsafe, error |
| TestParseCommand | kubectl $(evil) | $ and ( rejected as unsafe, error |
| TestParseCommand | kubectl \`id\` | backtick rejected as unsafe, error |
| TestParseCommand | kubectl --kubeconfig "a b" | quote+space rejected as unsafe, error |
| TestParseCommand | curl | not on the kubernetes allowlist, error |
| TestParseCommand | /tmp/evil | path, not a bare name, error |
| TestParseCommandExtraAllowed | sudo podman without extraAllowed | rejected, error |
| TestParseCommandExtraAllowed | sudo podman with extraAllowed=[sudo] | accepted, argv [sudo podman] with the platform binary unmodified |
| TestKubernetesDeployApplyOnStdin | - | deploy issues one call kubectl --context prod apply -f - with manifest on stdin |
| TestKubernetesDeleteUsesDeleteVerb | - | delete issues argv [oc delete -f -] |
| TestKubernetesRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestKubernetesUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestDockerUpAndDown | - | deploy runs compose up -d, delete runs compose down, correct argv each |
| TestDockerRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestResolveQuadletScope | system scope | UserMode false and Dir equals quadletSystem |
| TestResolveQuadletScope | user scope | UserMode true and Dir has quadletUserSub suffix |
| TestResolveQuadletScope | user scope with dir override | Dir equals override /custom/dir |
| TestResolveQuadletScope | bogus scope | unknown scope returns error |
| TestPodmanDeployReloadThenStart | - | user mode issues daemon-reload then start with --user flag in order |
| TestPodmanDeploySystemModeNoUserFlag | - | system mode issues systemctl daemon-reload without --user flag |
| TestPodmanDeleteStopsRemovesReloads | - | stop then daemon-reload called and unit file removed from disk |
| TestPodmanDeployStartFailureIsReported | - | start failure on call 1 surfaces error containing 'start a.service', 2 calls made |
| TestDockerUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestPodmanDeleteStopFailureIsReported | - | stop failure surfaces error containing 'stop solmq-connector.service' |
| TestWriteFileCreatesDirsAndMode | - | creates nested dirs, writes content, sets mode 0600 (non-windows) |
| TestWriteFileDoesNotTightenExistingFileMode | - | content replaced but existing 0644 mode left unchanged (non-windows) |
| TestWriteFileParentIsFileReturnsError | - | parent path is a file: MkdirAll error surfaces naming the blocker path |
| TestWriteFileTargetIsDirectoryReturnsError | - | target path is a directory: write error surfaces naming the target path |
| TestOSRunEchoesResolvedPathToStderr | - | OS.Run prints "exec: <resolved-path> <remaining args>" to stderr before exec'ing, using the resolved binary path |
| TestOSRunRejectsUnresolvableArgv0 | - | a binary LookPath cannot find on PATH is a Run error, not a deferred exec.Start failure |
| TestPreflightKubernetesArgvDeployNoNamespace | - | kubernetes deploy preflight issues <argv> auth can-i create deployment with no --namespace when namespace is empty |
| TestPreflightKubernetesArgvDeleteWithNamespace | - | kubernetes delete preflight issues <argv> auth can-i delete deployment --namespace <ns> |
| TestPreflightDockerArgvIsInfo | - | docker preflight issues <argv> info |
| TestPreflightPodmanArgvIsInfo | - | podman preflight issues <argv> info |
| TestPreflightFailureWrapsPlatformHint | kubernetes / docker / podman | a failing probe's error contains "preflight failed for <platform>", preserves the underlying cause, and carries the platform's login/daemon hint |
| TestPreflightRejectsDisallowedBinaryBeforeRunning | - | a command outside the platform allowlist (curl) is rejected before the probe ever runs, zero runner calls |
| TestPreflightExtraAllowedThreadsThrough | - | "sudo podman" is rejected without extraAllowed and accepted with it, running argv [sudo podman info] |
| TestPreflightUnknownAction | - | an action other than deploy/delete is rejected, zero runner calls |

## internal/scan

Discover workflow files -- YAML-only, env-file exclusion, wildcard matching, and metacharacter rejection.

Tests: [scan_test.go](../internal/scan/scan_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestScanSortsYAMLOnly | - | result sorted to [10.yml, 20.yaml], non-yaml and dirs ignored, Dir set to input dir |
| TestScanExcludesEnvFile | - | env.yaml excluded even with pattern '*', only workflow-0.yaml returned |
| TestScanEnvFileExcludedRegardlessOfPattern | - | pattern 'env*' still excludes env.yaml, only envoy.yaml remains |
| TestScanEmptyPatternDefaultsToStar | - | empty pattern behaves as '*', matches a.yaml |
| TestScanErrorMissingDir | - | scanning nonexistent directory returns an error |
| TestScanPatternWildcards | workflow-* | trailing star matches workflow-0.yaml and workflow-1.yaml only |
| TestScanPatternWildcards | *hoc* | mid-string star matches only adhoc.yaml |
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
| TestMatchStar | a*b*c,axxbyyc | multiple stars match interspersed segments -> true |
| TestMatchStar | a*b*c,axxc | multiple stars but missing required segment -> false |
| TestMatchStar | **,anything | consecutive stars still match any name -> true |
| TestIsYAML | a.yaml | isYAML true for .yaml extension |
| TestIsYAML | a.yml | isYAML true for .yml extension |
| TestIsYAML | A.YAML | isYAML true case-insensitively |
| TestIsYAML | a.txt | isYAML false for .txt extension |
| TestIsYAML | yaml | isYAML false when no extension present |
| TestIsYAML | a.yamlx | isYAML false for non-exact extension match |

## internal/examples

Write the shipped starter files (create/skip/force) and prove they generate config.

Tests: [examples_test.go](../internal/examples/examples_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestWriteCreatesSkipsForces | first-write | first write creates 5 files (4 workflows + env.yaml), 0 skipped |
| TestWriteCreatesSkipsForces | second-write-no-force | second write with no force skips all previously written files |
| TestWriteCreatesSkipsForces | force-rewrite | force write rewrites all files, restoring workflow-0.yaml to embedded original content over junk |
| TestWriteMkdirError | - | Write returns error when target dir path is under a regular file |
| TestShippedExamplesGenerateConfig | - | embedded example set written to disk generates config via gen.Config with no errors and at least one non-empty rendered application.yml |

## cmd/solmq-conn-util

The CLI shell -- flag parsing, the exit-code contract, the generate/validate/examples/completion commands, and the deploy/delete seams for all three engines. The completion tests also gate the four generated shell scripts against the command model.

Tests: [main_test.go](../cmd/solmq-conn-util/main_test.go), [commands_doc_test.go](../cmd/solmq-conn-util/commands_doc_test.go), [completion_test.go](../cmd/solmq-conn-util/completion_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestDispatchHandlersMatchModel | verbs / generate targets / deploy platforms / completion shells | the dispatch handler sets and cliVerbs agree in BOTH directions, so a command added to one cannot drift from the other |
| TestExitCodeContract | nil args | run(nil) returns exit code 2 |
| TestExitCodeContract | unknown command | run([bogus]) returns exit code 2 |
| TestExitCodeContract | help short -h | run([-h]) returns exit code 0 |
| TestExitCodeContract | help long --help | run([--help]) returns exit code 0 |
| TestExitCodeContract | help word | run([help]) returns exit code 0 |
| TestExitCodeContract | unknown flag -nope | run returns exit code 2 for unrecognized flag |
| TestExitCodeContract | missing env file | run returns exit code 1 when env file does not exist |
| TestExitCodeContract | invalid spec | run returns exit code 1 for structurally invalid workflow |
| TestGenerateConfigStdoutAndFileMatch | stdout run | exit 0 and stdout contains 'spring:' |
| TestGenerateConfigStdoutAndFileMatch | file run | exit 0 and written file content equals prior stdout exactly |
| TestGenerateFlagsBeforeAndAfterPositional | flags before positional target | exit 0 and output file written |
| TestGenerateFlagsBeforeAndAfterPositional | flags after positional target | exit 0 and output file written |
| TestWriteCredEnvFile | empty name | no error and no file written when name is empty |
| TestWriteCredEnvFile | resolver error | error returned and no creds.env file written for unset variable |
| TestWriteCredEnvFile | nil kvs existing env-file | no error and existing.env content PRESERVE=1 left untouched |
| TestWriteCredEnvFile | happy path ordered KVs | writes A=1\nB=2\n and file mode 0600 (non-windows) |
| TestGenerateConfigEmitWriteError | - | emit to path with missing parent dir returns exit code 1 |
| TestLoadEnvWorkflowsDirRelativeToEnvFile | - | workflows.dir resolved relative to env file not cwd; exit 0 and stdout has spring: |
| TestLoadEnvExcludesEnvFileFromWorkflowSet | - | env.yaml excluded from its own workflow scan; exit code 0 |
| TestDeployKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then apply) with argv [kubectl apply -f -], stdin contains kind: Deployment |
| TestDeleteKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then delete) with argv [kubectl delete -f -] |
| TestDeployKubernetesSeamRejectsUnsafeCommand | - | unsafe kubernetes.command yields exit 1 and zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | path value | --allow-command /usr/bin/sudo exits 2, zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | unsafe character | --allow-command sudo;rm exits 2, zero runner calls |
| TestAllowCommandFlagRejectedOnGenerateAndValidate | generate config / validate | --allow-command is undefined on generate/validate; exit 2 as an unknown flag |
| TestAllowCommandFlagRepeatableThreadsToRunner | - | "sudo podman" rejects with zero runner calls without the flag; repeating --allow-command sudo twice threads through to preflight (argv [sudo podman info]) and to the podman secret calls |
| TestDeployKubernetesPreflightFailureStopsBeforeApply | - | a failing kubernetes preflight (auth can-i argv incl. --namespace) stops with exit 1 and exactly 1 runner call |
| TestDeployDockerPreflightFailureStopsBeforeWrite | - | a failing docker preflight (argv [docker info]) stops before the compose file is written, exit 1, exactly 1 runner call |
| TestDeployPodmanPreflightFailureStopsBeforeWrite | - | a failing podman preflight (argv [podman info]) stops before the unit/app-yaml files are written, exit 1, exactly 1 runner call |
| TestValidateOKAndErrors | valid spec | validate exits 0 |
| TestValidateOKAndErrors | invalid spec | validate exits 1 |
| TestExamplesWriteSkipForceThenGenerate | first write | examples command exits 0 creating env.yaml |
| TestExamplesWriteSkipForceThenGenerate | re-run without -f | exits 0 and skips existing file, content stays 'touched' |
| TestExamplesWriteSkipForceThenGenerate | re-run with -f before dir | exits 0 and overwrites existing file |
| TestExamplesWriteSkipForceThenGenerate | generate on shipped examples | generate config on generated env.yaml exits 0 |
| TestExamplesDefaultDir | - | examples with no dir arg exits 0 and creates ./examples/workflow-0.yaml |
| TestGenerateKubernetesStdout | - | exit 0 and stdout contains kind: Deployment |
| TestGenerateDockerToFile | - | exit 0 and compose file contains services: and image: img:1 |
| TestGeneratePodmanQuadletStdout | - | exit 0 and stdout contains unit banner '# === solmq-conn-util.container ===' |
| TestDeployDockerSeamWritesComposeAndRuns | - | exit 0, compose file written, 2 runner calls (preflight then up) argv [docker compose -f <compose> up -d] |
| TestDeployDockerSeamComposeFileSurvivesFailedRun | - | preflight succeeds but the real `up` call fails; compose file still exists on disk afterward, exit 1 |
| TestDeployDockerSeamChildEnvCarriesCredentials | - | preflight call carries no env; the real `up` call (index 1) carries the resolved literal and -env credentials as STABLE=value pairs |
| TestDeleteDockerSeam | - | exit 0, 2 runner calls (preflight then down) argv [docker compose -f <compose> down] |
| TestDeployPodmanSeamWritesUnitsAndStarts | - | exit 0, a leading podman info preflight call, then app yaml and container unit written to quadlet dir, systemctl daemon-reload then start calls |
| TestDeletePodmanSeamStopsRemovesReloads | - | exit 0, a leading podman info preflight call, then systemctl stop then daemon-reload calls, unit and app yaml files removed |
| TestAbsPath | absolute input | absPath returns input unchanged when already absolute |
| TestAbsPath | relative input | absPath joins relative path onto base dir |
| TestCommandsDocInSync | - | docs/commands.md equals what the command model renders; -update rewrites it instead of asserting |
| TestCommandsModelMatchesUsage | - | every InUsage command and every flag in the model appears in usage(), and usage() lists no command the model omits |
| TestCompletionDispatchPrintsScript | bash / zsh / fish / powershell | `completion <shell>` exits 0, writes the script to stdout, and never reaches the runner |
| TestCompletionGoldenInSync | bash / zsh / fish / powershell | each rendered script equals its snapshot under cmd/solmq-conn-util/testdata/completions; -update rewrites them |
| TestCompletionCoversModel | bash / zsh / fish / powershell | every modeled verb, target and flag spelling reaches every shell, with descriptions in the three shells that show them |
| TestCompletionRecognizesFlagAliases | bash / zsh / powershell | every spelling flag.Parse accepts (-e, --e, -env, --env) is in the value-skipping table, so a value is never mistaken for a positional |
| TestCompletionShellStructure | bash / zsh / fish / powershell | each script keeps the registration line that makes it load, and the zsh script opens with #compdef |
| TestCompletionValueKindsReachScripts | bash / zsh / fish / powershell | a path flag completes files and `examples` completes directories in every shell |
| TestCompletionOutputIsPlainASCIILF | bash / zsh / fish / powershell | generated scripts are plain ASCII, LF only, newline-terminated |
| TestCompletionModelMetadataComplete | - | every verb has a description and a known PosArg, every flag a known Arg and a non-empty Meaning, every modeled shell a renderer and a snapshot; verb/target names stay [a-z0-9-] for unquoted case patterns |
| TestPlainText | code spans stripped / newline folded / tab and CR folded / whitespace runs collapse / trimmed / control chars dropped / empty / only backticks / punctuation preserved | model text is reduced to a single-line tooltip that cannot break the enclosing shell statement |
| TestShellQuoting | plain / empty / apostrophe / backslash / dollar and backtick / double quote / semicolon and pipe | bashQuote, fishQuote and psQuote each neutralize their shell's escape rules |
| TestZshEntry | plain / colon in the value escaped / colon in the description left alone / apostrophe quoted / empty description | _describe entries split on the intended colon only |
| TestFlagAliasesAndOffered | short and long pair / long only | flagOffered suggests the documented spellings, flagAliases lists all four dash forms, fishFlagSpec renders -s/-l correctly |
