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

_Snapshot: 381 test functions, 594 case rows across 15 packages. (Functions counted from `func Test` in the source; case rows are the data rows of the tables below, not a suite run -- human, please confirm against `./scripts/dev.sh test` / `cov` output.)_

## internal/spec

Parse env.yaml into the typed model -- workflows, defaults, named connections, the kubernetes/docker/podman target sections, and ports -- and apply section defaults.

Tests: [spec_test.go](../internal/spec/spec_test.go), [env_test.go](../internal/spec/env_test.go), [targets_test.go](../internal/spec/targets_test.go), [expand_test.go](../internal/spec/expand_test.go), [defaults_test.go](../internal/spec/defaults_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestParseWorkflowSolaceAndMQ | - | parses full solace source and mq target with dest kind, tls, key alias, props |
| TestParseWorkflowConnRef | - | mq source with conn-ref resolves ConnRef, DestKind queue, Dest A.IN; SetsConnFields false |
| TestBaseName | url / posix / windows / bare / empty | one shared BaseName splits on both separators, so a Windows-authored path resolves identically on Linux |
| TestWorkflowFileLess | 2 vs 10 / 9 vs 10 / workflow-2 vs workflow-10 | digit runs compare as numbers, so workflow ids follow the order the files were numbered in rather than lexical order |
| TestWorkflowFileLess | 10.yaml vs 10.yml / a vs b | equal digit runs fall through to byte order, and names without digits compare byte-wise throughout |
| TestWorkflowFileLess | 7 vs 007 / x vs x | equal value with different zero padding is ordered by the raw text, and no name is less than itself -- the strict order sort.Slice requires |
| TestConnRefSideMayTuneBinding | - | consumer block parses on a conn-ref side, SetsConnFields ignores it, and Resolve keeps it alongside the referenced tuple and destination |
| TestParseDefaultsConnectionsAndLeaderElection | - | parses 2 named connections and leader-election active_standby with fail-over |
| TestResolveConnRef | known ref edge | resolves host/msg-vpn/key-alias from connections map, keeps dest |
| TestResolveConnRef | unknown ref nope | returned unchanged with ConnRef nope and empty Host |
| TestParseWorkflowEnabledDefaultsTrue | - | enabled defaults true and target stays unset when absent |
| TestParseWorkflowAmbiguousSystemAndDest | solace and mq both set | HasSystem returns false when both systems present |
| TestParseWorkflowAmbiguousSystemAndDest | queue and topic both set | DestKind is empty string when queue+topic ambiguous |
| TestParseWorkflowSyntaxError | - | malformed yaml returns non-nil error |
| TestParseDefaultsFull | - | parses tls stores, management port 8090 with exposure health (a removed key, still parsed), security enabled false with 1 user, leader-election standalone, logging/solace-defaults nodes captured |
| TestParseSecurityUserRoles | absent / one / several | security.users[].roles parses to no roles (the connector's read-only default), a single role, and several in authored order |
| TestParseDefaultsSecurityEnabledKeyOmittedStaysNil | - | security.enabled is a removed key: an omitted key parses to Security.Enabled nil rather than being defaulted |
| TestParseDefaultsEmpty | - | empty input yields a zero-valued Management (Management{}), Security.Enabled nil with no users, and TLS.Truststore nil |
| TestParseDefaultsError | - | malformed tls yaml returns non-nil error |
| TestParseKubernetesReplicasDefault | - | deployment without replicas defaults Replicas to 1 |
| TestParseKubernetesFull | - | parses replicas 2, service enabled port 8090, credentials create name, stores create present; the removed source/variables keys still parse so RemovedKeys can report them |
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
| TestApplyDockerDefaultsFillsMissing | - | docker defaults command/name/restart applied, ports stay empty (publishing is opt-in), stores/libs stay nil |
| TestApplyDockerDefaultsOverrideWins | - | explicit command/name/restart/ports override defaults exactly as given |
| TestApplyPodmanDefaultsFillsMissing | - | podman defaults command/mode run/name/restart applied, ports stay empty (publishing is opt-in), and Quadlet non-nil with scope auto and empty dir |
| TestApplyPodmanDefaultsOverrideWins | - | explicit command/mode quadlet/name/restart/ports/quadlet scope+dir override defaults exactly |
| TestApplyMountDefaultsFillsMissing | - | stores mount-path defaults, libs dir kept and mount-path defaulted |
| TestPortDefaultsFollowManagementPort | - | management.port 9091 with docker/podman/kubernetes present: kubernetes Service.Port defaults to {9091,9091}; docker/podman publish nothing with ports: omitted |
| TestPortDefaultsFallBackWhenManagementPortUnset | - | no management.port set: kubernetes service.port falls back to {DefaultMgmtPort,DefaultMgmtPort} (8090); docker/podman still publish nothing |
| TestKubernetesServicePortAcceptsBareAndHostContainerForms | bare int 8090 | kubernetes service.port parses to Host=8090 Container=8090 |
| TestKubernetesServicePortAcceptsBareAndHostContainerForms | host:container 8080:8090 | kubernetes service.port parses to Host=8080 Container=8090 |
| TestKubernetesServicePortRejectsInvalidForms | multi-colon 1:2:3 | error 'env.yaml: ports entry "1:2:3" must be "host:container" (exactly one colon)' |
| TestKubernetesServicePortRejectsInvalidForms | mapping node {a: 1} | error 'env.yaml: ports entry must be an integer or "host:container", got a !!map' |
| TestApplyMountDefaultsOverrideWins | - | explicit stores and libs mount-path overrides retained |
| TestExpandBracedVar | - | `${HOST}` in Side.Host expands from Lookup |
| TestExpandDefaultVarSetUsesValue | - | `${VPN:fallback}` uses the looked-up value when VPN is set |
| TestExpandDefaultVarUnsetUsesDefault | - | `${VPN:fallback}` falls back to the default when VPN is unset |
| TestExpandUnsetNoDefaultPassesThroughWithWarning | - | unset defaultless `${TYPO}` passes through verbatim and Warn is called exactly once naming TYPO |
| TestExpandBareDollarVarUntouched | - | bare `$VPN` (no braces) is left untouched even though VPN is set |
| TestExpandCredentialFieldLeftAlone | - | Side.Password/PasswordEnv (`expand:"no"`) never expand |
| TestExpandYAMLNodePassthroughLeftAlone | - | a `*yaml.Node` field (APIProps) is never walked or rewritten |
| TestExpandDefaultsConnectionsMapEntry | - | a `${HOST}` value inside Defaults.Connections (map[string]Side) expands via read-modify-write |
| TestExpandSecurityUserRole | - | a `${VAR}` in security.users[].roles expands (a role is an identity, not a credential), proving Expand reaches a []string inside a slice of structs under Defaults |
| TestExpandNilLookupDisablesEverything | - | nil Lookup makes Expand a no-op, leaving `${HOST}` untouched |
| TestLeaderElectionEffectiveMode | empty | an empty Mode defaults EffectiveMode to standalone |
| TestLeaderElectionEffectiveMode | standalone / active_active / active_standby | an explicit Mode passes through EffectiveMode unchanged |
| TestEffectiveManagementPort | nil receiver / unset port / set port | EffectiveManagementPort falls back to DefaultMgmtPort (8090) for a nil Defaults and an unset port, and returns a configured port unchanged |
| TestCredEmptyBothKeyDescribe | unset / literal only / env only / both set | Cred.Empty/Both/Key/Describe resolve deterministically for every shape; both-set resolves to the env side (validate rejects it separately) rather than panicking |
| TestSideUsernameSecretBothSystems | - | Side.Username/Secret dispatch by System, not by whichever credential pair is non-empty: solace returns client-user/-pass, mq returns user/password |
| TestStoreSecretNilSafe | - | nil *Store yields an empty Cred; literal and -env stores yield the matching Cred side |
| TestUserSecretLiteralAndEnv | - | a security user's Secret() carries the literal or the -env variable, matching what was set |
| TestCredCreateRemovedKeys | - | RemovedKeys reports each removed key (source, variables, values-file) alone and all three in order; a nil receiver and a bare create.name report none |

## internal/validate

Validate the parsed model -- per-side rules, connection refs, leader election, the docker/podman/kubernetes target sections, ports, container names, TLS/stores wiring, and the safe-token charset.

Tests: [validate_test.go](../internal/validate/validate_test.go), [validate_extra_test.go](../internal/validate/validate_extra_test.go), [validate_deploycommand_test.go](../internal/validate/validate_deploycommand_test.go)

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
| TestCheckKubeServicePort | - | kubernetes.service.port is range-checked like docker/podman ports: a scalar or distinct host:container pair both pass, and an out-of-range host or container side each error independently naming the offending side |
| TestCheckKubeCredentialCreateRemovedKeys | source/variables/values-file set | credentials.create still carrying the removed keys errors naming all three and telling the operator to remove them |
| TestCheckKubeCredentialCreateRemovedKeys | source alone | credentials.create carrying only the removed `source` key errors naming it alone |
| TestCheckKubeCredentialCreateRemovedKeys | bare name | a bare create.name (the new shape) trips no removed-keys error |
| TestCheckKubeStoresRequireTruststore | - | kube stores create without tls.truststore errors requires tls.truststore |
| TestStoresNotWiredWarning | - | TLS workflow with kube deploy and no stores wiring warns secrets.stores is omitted |
| TestStoresWiredExistingNoWarning | - | stores wired via existing secret produces no stores-omitted warning |
| TestCheckCredRules | both literal and env set | errors sets both a literal value and target password-env |
| TestCheckCredRules | env value is a ${...} reference | errors must be a bare variable name, not a ${...} reference |
| TestCheckCredRules | env value not a valid identifier | errors is not a valid environment variable name |
| TestCheckCredRules | -env var unset in this environment | warns (not errors) which is not set in this environment |
| TestCheckCredLiteralLooksLikeEnvRefWarns | - | a literal credential containing ${ warns naming the -env key to use instead; the value is still used as a literal |
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
| TestCheckDockerPodmanSecretsRemoved | docker.secrets set | a docker section still carrying a `.secrets` block errors docker.secrets is no longer configured |
| TestCheckDockerPodmanSecretsRemoved | podman.secrets set | a podman section still carrying a `.secrets` block errors podman.secrets is no longer configured |
| TestCheckDockerPodmanSecretsRemoved | nil secrets | the new default (no `.secrets` block) trips no such error |
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
| TestSafeActuatorUser | 4 accepted names | letters, digits, '.', '-' and '_' pass, and every accepted name also satisfies SafeToken |
| TestSafeActuatorUser | 6 SafeToken-permitted names | '/', ':', '=', ',', '+' and '@' are rejected here even though SafeToken allows them, since the name reaches a sed address |
| TestSafeActuatorUser | 7 shell-unsafe names, empty | quotes, whitespace, backslash, '$', '*', '[' and the empty string are rejected |
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
| TestRemovedDefaultsKeysRejected | - | a security.enabled value (true or false) errors security.enabled is no longer configurable, a management.exposure value errors management.exposure is no longer configurable, and neither key set validates clean |
| TestStatusUserReservedName | - | a security.users entry named spec.StatusUserName errors reserved, naming security.users[1].name; a differently-named user does not collide |
| TestSecurityUserRoles | admin / unknown-but-well-formed / several / empty / whitespace-only / shell metacharacter / embedded space / no roles | roles are checked for usability, not against an allowlist: a well-formed unrecognized role passes, an empty or whitespace-only entry errors naming both indices, an unsafe-charset entry errors, and omitting roles entirely stays clean. Also pins both error texts verbatim, since the generator page's JS validator mirrors them word for word |
| TestStatusUserPasswordEnvCharset | nil Env / unset / empty / valid value | none trip the SECURITY_USER_SOLMQ_STATUS_PASSWORD charset error |
| TestStatusUserPasswordEnvCharset | space / double quote / single quote / backslash / dollar-brace / control char / non-ASCII byte | each errors the charset check, and the error text never echoes the secret value |

## internal/consolidate

Build the consolidated binder model from the workflows -- dedup connections, TLS bundles, destination roles, store-path rewriting, leader election, and durable-name UUIDs.

Tests: [consolidate_test.go](../internal/consolidate/consolidate_test.go), [consolidate_extra_test.go](../internal/consolidate/consolidate_extra_test.go), [names_test.go](../internal/consolidate/names_test.go), [uuid_test.go](../internal/consolidate/uuid_test.go)

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
| TestApplyStatusAccessNoOperatorUsers | - | with no configured security.users, Build synthesizes the reserved account as the only user, carrying the literal status password, never entering Model.Secrets |
| TestApplyStatusAccessAppendsAfterExistingUsers | - | operator-configured users get the reserved account appended last; existing users still resolve to secretRef placeholders |
| TestApplyStatusAccessCarriesOperatorRoles | - | an operator's roles reach the model verbatim (only the password is rewritten), the reserved account is appended with none so it stays read-only, and the caller's own Defaults are left unmutated despite sharing the roles backing array |
| TestApplyStatusAccessExposureIsFixed | - | applyStatusAccess always sets Management.Exposure to health,info,metrics,leaderelection,workflows, ignoring whatever spec.Management.Exposure carries |
| TestDurableNameGolden | - | DurableName of fixed inputs equals pinned solmq-3631c883-c0c4-5bc8-985e-ea2842831ad6 |
| TestDurableNameDeterministic | same inputs called twice | DurableName returns identical value both times |
| TestDurableNameDeterministic | different file name g.yaml vs f.yaml | DurableName differs when file name changes |
| TestGeneratedSecretNamesStayOutOfChildEnvDanger | - | every SecretRef.Stable Build() can produce (binder creds, security-user passwords, TLS stores, leader-election) matches a fixed-suffix pattern, so adversarial conn-ref/security-user names (e.g. "path", "ld-preload", "LD") can never fold to a bare dangerous docker-compose child-env name like PATH or LD_PRELOAD |
| TestStableTokenFolding | mq-conn-1 / svc.a / punctuation runs / leading digit / empty / underscore runs | stableToken folds to upper-snake, collapses non-alphanumeric runs to one `_`, trims edge `_`, prefixes a leading digit with X, and returns X for an unfoldable input |
| TestBinderFieldsCarryStablePlaceholders | - | no credential value or host env-var name reaches a rendered binder field -- only ${STABLE_NAME} placeholders -- and Model.Secrets records the real literal/env source under each stable name |

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
| TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral | - | store passwords reach api-properties only as ${TRUSTSTORE_PASSWORD}/${KEYSTORE_PASSWORD} placeholders; secretRef gets each store's own credential and no literal or env-var name leaks into any value |
| TestSolacePropsSkipsSecretRefWhenStoreMissing | - | with no truststore/keystore SolaceProps emits nothing and never calls secretRef |

## internal/statusscript

Render the POSIX status script the generated deploy artifacts embed and `solmq-conn-util status` execs inside each running instance -- a pure renderer with no os/exec, filesystem, or network access.

Tests: [statusscript_test.go](../internal/statusscript/statusscript_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderSubstitution | defaults | Render substitutes PORT=8090 and USER_NAME=solmq-status; BASE and the leaderelection/workflows endpoints are built from $PORT at run time, not by Render |
| TestRenderSubstitution | non-default port and user | Render substitutes PORT=19090 and USER_NAME=custom-mgmt-user |
| TestRenderIsPureASCIINoCRLF | - | output has no carriage return, no byte over 127, and ends with a trailing newline |
| TestRenderHeaderHasExecOneLiners | - | the header pins the kubectl/docker/podman exec one-liners, each built from ContainerPath |
| TestRenderPasswordResolution | - | the password-lookup chain references ContainerPath, /run/secrets and the from_configs account lookup, and no credential is embedded |
| TestRenderAlwaysExitsZero | - | every exit in the script is `exit 0`, `set -e` is absent while `set -u` stays, the EXIT trap holds the contract, and active/standby are both quiet outcomes |
| TestRenderSendsStatusToStdoutAndProblemsToStderr | - | the mode/state/health/workflow report lines go to stdout unredirected, and every `status:` diagnostic ends in `>&2` |
| TestRenderAlignsWorkflowColumn | - | the workflows block is a bare header plus one indented row per workflow, with the ids right-aligned to the widest id present so every colon sits in the same column |
| TestRenderReportsHealthUptimeAndVersion | endpoints | health, /actuator/metrics/process.uptime and /actuator/info are each read and rendered as their own report line |
| TestRenderReportsHealthUptimeAndVersion | dropped when silent | each enrichment line sits behind a non-empty guard, so an endpoint that answers nothing drops its line instead of printing an empty value |
| TestRenderReportsHealthUptimeAndVersion | first status wins | the health match is anchored at the opening brace, so a component's status is never reported as the whole instance's |
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
| TestFilenameAndPathConstants | - | Filename, ContainerDir, ContainerPath and ConfigPath equal their pinned values |

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
| TestApplicationMinimalNoOptionalBlocks | ssl: / logging: | ssl: and logging: blocks stay absent when defaults are empty |
| TestApplicationMinimalNoOptionalBlocks | management: / security: | management: and security: are unconditional now: the fixed exposure list and the reserved solmq-status account render even with empty defaults |
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
| TestApplicationConfigImport | ConfigImport set / empty | Application() leads with spring.config.import when Model.ConfigImport is set, and omits the block entirely when it is empty |
| TestApplicationSecurityUserRoles | - | a roles-bearing user renders a block-style roles sequence under its password; a role-less user and the reserved solmq-status account emit no roles key at all, keeping pre-roles output byte-identical |

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
| TestRenderConfigMapStatusScript | - | the ConfigMap always carries the status script under its own `status` key, alongside application.yml |
| TestRenderStatusScriptMountAfterLibs | - | the single-file status mount is declared after the libs directory mount, so it is not shadowed, and carries `subPath: status` |
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

## internal/dockergen

Render the docker compose file from the target model.

Tests: [dockergen_test.go](../internal/dockergen/dockergen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull_WithEverything | - | full golden output with creds, store, libs, MQTLS, ports, timezone matches exactly |
| TestRenderFull_Minimal | - | minimal golden output omits restart, ports, environment, env_file, volumes blocks |
| TestEnvironmentBranches | TZ only, MQTLS false | environment block contains only TZ: UTC, no JAVA_TOOL_OPTIONS |
| TestEnvironmentBranches | MQTLS only, no timezone | environment block contains only JAVA_TOOL_OPTIONS line, no TZ |
| TestSecretsBranches | at least one secret | the per-service secrets list and the top-level environment-provider secrets block are both emitted |
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
| TestStatusScriptConfigSourceAndTarget | - | the service references a second `<name>-status` config and mounts it at /app/external/libs/status |
| TestStatusScriptContentIsEscaped | - | the status script body is inlined under the status config's content: block, indented 6 spaces, blank line preserved as truly empty, and its shell `$` doubled |
| TestContentEscapesDollarsForCompose | $VAR / ${VAR} / ${VAR:-default} / $(cmd) / $$ | each shape reaches the content block with every `$` doubled, so compose's interpolation pass delivers it unchanged instead of blanking it or rejecting the document |
| TestContentEscapesDollarsForCompose | no lone `$` | dropping every `$$` pair from the rendered document leaves no `$` behind anywhere |
| TestAppYAMLSecretPlaceholdersAreNotInterpolated | - | application.yml's ${...} credential placeholders render doubled so compose cannot substitute the values the CLI passes it, while the `secrets:` provider entries -- compose's own -- stay unescaped |

## internal/podmangen

Render the podman run script and the quadlet units from the target model.

Tests: [podmangen_test.go](../internal/podmangen/podmangen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderRunScriptFull | - | full input renders run script matching exact golden string byte-for-byte |
| TestRenderRunScriptSecretPreamble | - | the secret-loading preamble carries no credential value, only ${NAME:?...} referencing the stable name |
| TestRenderRunScriptNoSecretsOmitsPreamble | - | no Secrets in Input emits no `podman secret` lines at all |
| TestRenderRunScriptMinimal | - | minimal input renders run script matching exact golden string byte-for-byte |
| TestRenderQuadletFull | - | full input yields 1 unit named solmq-connector.container with content matching golden string |
| TestRenderQuadletMinimal | - | minimal input yields 1 unit with content matching golden string (no Service section, no restart) |
| TestLeaderLabelsPerMode | empty defaults to standalone / standalone / active_active | RenderRunScript and RenderQuadlet both carry the le-mode label and role: active |
| TestLeaderLabelsPerMode | active_standby | both renderers carry le-mode active_standby and withhold role: active |
| TestStatusScriptMountNestsAfterLibs | - | in both renderers, the status script mount/volume is declared after the libs mount, so it nests rather than being shadowed |
| TestStatusScriptMountOmittedWhenPathEmpty | - | an empty StatusScriptPath omits the status mount entirely in both renderers, rather than mounting an empty source |

## internal/gen

Orchestrate parse -> validate -> consolidate -> render, resolve credentials/stores, and assert the byte-for-byte golden fixtures.

Tests: [gen_extra_test.go](../internal/gen/gen_extra_test.go), [golden_test.go](../internal/gen/golden_test.go), [htmlgolden_test.go](../internal/gen/htmlgolden_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless | - | parse() expands host/msg-vpn from Lookup, leaves client-password-env verbatim, and returns exactly one warning naming TYPO for the unset defaultless conn-name variable |
| TestResolveStores | truststore+keystore with ReadFile | resolves 2 store files, first named t.jks |
| TestResolveStores | no ReadFile provided | missing ReadFile returns error |
| TestResolveStores | ReadFile returns error | read error propagates |
| TestResolveStores | no stores configured | empty Defaults yields 0 stores, no error |
| TestToIssues | - | toIssues wraps each string into an Issue carrying it as Msg |
| TestNamesAndPaths | pathIn base variants | empty base returns bare path; trailing/no-trailing slash both join to /base/a |
| TestTargetMounts | tls+stores+libs configured | 2 store mounts at fixed default mount path (ignores custom MountPath), libs mount uses resolved abs source and given target |
| TestTargetMounts | nil stores and libs | opt-in mounts yield nil,nil when sections absent |
| TestResolveCredentials | nil refs | no kvs, no error |
| TestResolveCredentials | literal + -env mix | literal ref passes through, -env ref reads from the resolver's environment |
| TestResolveCredentials | unset -env variable | fails loud naming the stable secret and the variable, never a value |
| TestResolveCredentials | no environment access | fails loud rather than silently resolving to empty |
| TestConfigWorkflowCap | 21 workflows | Config produces no output and one error naming the count, the 20 cap, and the split-into-folders remedy |
| TestConfigWorkflowCap | 20 workflows | exactly at the cap does not error |
| TestGenerateKubernetesWorkflowCap | 21 workflows | GenerateKubernetes produces no manifest and the same workflow-cap error |
| TestConfigCarriesSecurityUserRoles | - | end-to-end: a roles-bearing env.yaml validates clean and its role reaches the rendered application.yml, while the reserved account still renders none |
| TestConfigNoSecretsLeak | - | every rendered password is a ${STABLE} placeholder except the one permitted literal: the reserved spec.StatusUserName account |
| TestGenerateDockerBasics | - | generates non-empty compose containing the image; all four credential positions render as top-level environment-provider secrets, never inlined as values, and each ${STABLE} placeholder in application.yml is doubled so compose cannot interpolate the value in |
| TestGeneratePodmanRunAndQuadlet | mode: run | produces a run script (no unit) that loads each secret into podman's store before `podman run`, never carrying the value itself |
| TestGeneratePodmanRunAndQuadlet | ForceQuadlet true | produces quadlet units, no run script |
| TestResolveStatusPasswordFixedRand | - | a fixed Rand hook yields the exact 32-lowercase-hex-char literal (16 bytes hex-encoded) |
| TestResolveStatusPasswordEnvOverride | - | a set, non-empty spec.StatusUserPasswordEnvVar is used verbatim and Rand is never consulted |
| TestResolveStatusPasswordEmptyEnvFallsBackToRand | - | an empty override is treated as unset, falling back to Rand rather than returning "" |
| TestResolveStatusPasswordRandError | - | a Rand failure surfaces as an actionable error naming the underlying cause, never a predictable fallback password |
| TestConfigStatusPasswordRandErrorNoOutput | - | the same Rand failure through Config is a hard error with no output |
| TestGenerateKubernetesCarriesStatusScript | - | the ConfigMap gets a "status: \|" key carrying the rendered script, addressed to spec.StatusUserName on the resolved management port |
| TestGenerateDockerCarriesStatusScript | - | compose gets a second top-level config (`<name>-status`) inlining the rendered script, mounted at statusscript.ContainerPath |
| TestGeneratePodmanCarriesStatusScript | - | PodmanPlan.StatusScript names `<name>-status` and its on-disk mount path is BaseDir-resolved exactly like AppYAML |
| TestGenerateMissingTargetSection | kubernetes | error contains kubernetes target requires a 'kubernetes:' section in env.yaml |
| TestGenerateMissingTargetSection | docker | error contains docker target requires a 'docker:' section in env.yaml |
| TestGenerateMissingTargetSection | podman | error contains podman target requires a 'podman:' section in env.yaml |
| TestGenValidateStoresWarning | - | kubernetes credentials.create (name-only) plus a TLS-without-stores config yields no errors and exactly one stores-omitted advisory warning |
| TestGeneratorPageGoldenInSync | - | the golden embedded in solmq-conn-util-generator.html matches testdata/golden/application.yml (regenerate with -update-html-golden) |
| TestGoldenConfig | - | generated config output matches testdata/golden/application.yml byte-for-byte, one instance |
| TestGoldenKubernetesCreate | - | generated kubernetes manifests (namespace, configmap incl. status script, secret, stores, pv, pvc, deployment with secrets-volume/stores/syslog/libs mounts and le-mode/role labels, service) match golden fixture byte-for-byte |
| TestGoldenKubernetesNoSecrets | - | generated manifests without secrets/syslog/libs (namespace, configmap incl. status script, deployment, service) match golden fixture byte-for-byte |

## internal/runner

The os/exec seam -- ParseCommand safe-tokenizing, kubectl/docker/podman deploy and remove argv, quadlet scope resolution, and WriteFile modes.

Tests: [runner_test.go](../internal/runner/runner_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestHelperProcess | - | helper child-process entry point guarded by GO_WANT_HELPER_PROCESS; dispatches stdin/both/fail/unknown modes |
| TestOSRunWiresStdinToChild | - | OS.Run passes stdin through to child, output equals hello-stdin |
| TestOSRunCombinesStdoutAndStderr | - | combined output contains both stdout-line and stderr-line |
| TestOSRunNonZeroExitReturnsErrorWithOutput | - | non-zero exit returns non-nil error and output still contains before-exit |
| TestOSRunAcceptsAbsolutePathArgv0 | - | absolute path as argv0 runs successfully, output equals abs-argv0-ok |
| TestOSRunEnvReachesChildAndAmbientInherited | - | Cmd.Env entries reach the child and override an ambient var of the same name, while an ambient-only var still passes through |
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
| TestKubernetesRemoveUsesDeleteVerb | - | the remove action issues argv [oc delete -f -] |
| TestKubernetesRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestKubernetesUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestDockerUpAndDown | - | deploy runs compose up -d, remove runs compose down, correct argv each |
| TestDockerRejectsUnsafeCommand | - | unsafe command rejected, error returned, zero calls made |
| TestResolveQuadletScope | system scope | UserMode false and Dir equals quadletSystem |
| TestResolveQuadletScope | user scope | UserMode true and Dir has quadletUserSub suffix |
| TestResolveQuadletScope | user scope with dir override | Dir equals override /custom/dir |
| TestResolveQuadletScope | bogus scope | unknown scope returns error |
| TestPodmanDeployReloadThenStart | - | user mode issues daemon-reload then start with --user flag in order |
| TestPodmanDeploySystemModeNoUserFlag | - | system mode issues systemctl daemon-reload without --user flag |
| TestPodmanRemoveStopsRemovesReloads | - | stop then daemon-reload called and unit file removed from disk |
| TestPodmanDeployStartFailureIsReported | - | start failure on call 1 surfaces error containing 'start a.service', 2 calls made |
| TestDockerUnknownAction | - | unknown action rejected, error returned, zero calls made |
| TestPodmanRemoveStopFailureIsReported | - | stop failure surfaces error containing 'stop solmq-connector.service' |
| TestPodmanSecretCreateRemovesThenCreatesValueOnStdin | - | PodmanSecretCreate issues rm --ignore then create with the value on stdin, never in argv |
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
| TestOSRunWritesNothingToStderr | - | OS.Run writes nothing to stderr of its own; the child's output reaches the caller only through the returned combined output |
| TestOSRunRejectsUnresolvableArgv0 | - | a binary LookPath cannot find on PATH is a Run error, not a deferred exec.Start failure |
| TestPreflightKubernetesArgvDeployNoNamespace | - | kubernetes deploy preflight issues <argv> auth can-i create deployment with no --namespace when namespace is empty |
| TestPreflightKubernetesArgvRemoveWithNamespace | - | kubernetes remove preflight issues <argv> auth can-i delete deployment --namespace <ns> |
| TestPreflightDockerArgvIsInfo | - | docker preflight issues <argv> info |
| TestPreflightPodmanArgvIsInfo | - | podman preflight issues <argv> info |
| TestPreflightFailureWrapsPlatformHint | kubernetes / docker / podman | a failing probe's error contains "preflight failed for <platform>", preserves the underlying cause, and carries the platform's login/daemon hint |
| TestPreflightRejectsDisallowedBinaryBeforeRunning | - | a command outside the platform allowlist (curl) is rejected before the probe ever runs, zero runner calls |
| TestPreflightExtraAllowedThreadsThrough | - | "sudo podman" is rejected without extraAllowed and accepted with it, running argv [sudo podman info] |
| TestPreflightUnknownAction | - | an action other than deploy/remove is rejected, zero runner calls |
| TestKubernetesPodNamesArgv | no namespace / with namespace | KubernetesPodNames issues `get pods -l <selector> -o name`, adding `-n <namespace>` when given |
| TestDockerComposeProjectArgvAndValue | - | the compose project is read from the container's own com.docker.compose.project label via `docker inspect --format`, not derived from the compose file's directory |
| TestDockerComposeProjectDegradesToEmpty | inspect failed / no-value / absent label | every way the lookup can come up empty yields "" rather than an error, so the banner drops the segment instead of failing an otherwise fine report |
| TestKubernetesPodNamesStripsPrefixAndDropsBlankLines | - | each `pod/` prefix is stripped and blank/whitespace-only lines are dropped |
| TestKubernetesPodNamesEmptyResultIsNotError | - | no matching pods returns an empty slice, not an error |
| TestKubernetesPodNamesRunFailureWraps | - | a run failure surfaces as an error naming the "listing pods" operation |
| TestScriptInstalledArgv | kubernetes no namespace / kubernetes with namespace / docker / podman | ScriptInstalled execs the marker-echoing probe through each platform's exec form, `-n <namespace>` only for kubernetes when given |
| TestScriptInstalledReadsMarkers | present / absent / marker among engine chatter / present despite a non-zero exit | the answer comes from the marker on stdout, so a marker is believed even when the engine also reported a non-zero exit |
| TestScriptInstalledUnreachableTargetIsError | engine error with no marker / clean exit with no marker | no marker at all means the probe never ran, so it errors rather than silently reporting absent |
| TestScriptInstalledUnknownPlatform | - | an unknown platform is rejected, zero runner calls |
| TestInstallScriptArgv | kubernetes no namespace / kubernetes with namespace / docker / podman | InstallScript execs `mkdir -p <dir> && cat > <path>` through each platform's exec form (`-i` for stdin, `-n <namespace>` only for kubernetes when given) |
| TestInstallScriptPassesScriptOnStdinNotArgv | - | the script body travels on stdin, never appearing in any argv token |
| TestInstallScriptUnknownPlatform | - | an unknown platform is rejected, zero runner calls |
| TestRunStatusScriptArgv | kubernetes no namespace / kubernetes with namespace / docker / podman | RunStatusScript execs `sh <path>` through each platform's exec form |
| TestRunStatusScriptReturnsOutputAlongsideNonZeroExit | - | a non-zero script exit (the status script's own 1/2 convention) is returned alongside its output, never swallowed |
| TestRunStatusScriptUnknownPlatform | - | an unknown platform is rejected, zero runner calls |

## internal/scan

Discover workflow files -- YAML-only, env-file exclusion, wildcard matching, and metacharacter rejection.

Tests: [scan_test.go](../internal/scan/scan_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestScanSortsYAMLOnly | - | result sorted to [10.yml, 20.yaml], non-yaml and dirs ignored, Dir set to input dir |
| TestScanSortsNumericallyNotLexically | - | 19/2/10/9.yaml come back as [2, 9, 10, 19], the numbering order, not the lexical [10, 19, 2, 9] |
| TestScanSortsPrefixedNamesNumerically | - | the same ordering when the digits follow a shared prefix, as in the workflow-N.yaml names the examples verb writes |
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

The CLI shell -- flag parsing, the exit-code contract, the generate/validate/examples/auto-complete commands, verb aliases, and the deploy/remove seams for all three engines. The completion tests also gate the four generated shell scripts against the command model.

Tests: [main_test.go](../cmd/solmq-conn-util/main_test.go), [commands_doc_test.go](../cmd/solmq-conn-util/commands_doc_test.go), [completion_test.go](../cmd/solmq-conn-util/completion_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestDispatchHandlersMatchModel | verbs / generate targets / deploy platforms / completion shells | the dispatch handler sets and cliVerbs agree in BOTH directions, so a command added to one cannot drift from the other |
| TestPlatformMapsCoverThreeNames | - | platformNames, platformGenerators and actTargets agree in both directions -- a modeled platform with no handler, or a handler with no modeled entry, fails |
| TestExitCodeContract | nil args | run(nil) returns exit code 2 |
| TestExitCodeContract | unknown command | run([bogus]) returns exit code 2 |
| TestExitCodeContract | help short -h | run([-h]) returns exit code 0 |
| TestExitCodeContract | help long --help | run([--help]) returns exit code 0 |
| TestExitCodeContract | help word | run([help]) returns exit code 0 |
| TestExitCodeContract | unknown flag -nope | run returns exit code 2 for unrecognized flag |
| TestExitCodeContract | missing env file | run returns exit code 1 when env file does not exist |
| TestExitCodeContract | invalid spec | run returns exit code 1 for structurally invalid workflow |
| TestExitCodeContract | auto-complete no shell | run([auto-complete]) returns exit code 2 |
| TestExitCodeContract | auto-complete bogus shell | run([auto-complete, bogus]) returns exit code 2 |
| TestExitCodeContract | completion is no longer a command | run([completion, bash]) returns exit code 2 -- the rename to `auto-complete` is a clean break with no compatibility alias, pinned so it cannot silently regress |
| TestExitCodeContract | 8 near misses | d / v / s / g / comp / h / hlp / stat each exit 2: none was picked as an alias, so none may resolve |
| TestVerbAliasesDispatchLikeCanonical | gen / dep / del / sts / ver / vld / eg | each alias reaches the same handler as its canonical verb, so the alias table and the dispatch map cannot drift apart |
| TestGenerateConfigStdoutAndFileMatch | stdout run | exit 0 and stdout contains 'spring:' |
| TestGenerateConfigStdoutAndFileMatch | file run | exit 0 and written file content equals prior stdout exactly |
| TestGenerateFlagsBeforeAndAfterPositional | flags before positional target | exit 0 and output file written |
| TestGenerateFlagsBeforeAndAfterPositional | flags after positional target | exit 0 and output file written |
| TestGenerateConfigWorkflowCapExceeded | - | a folder over validate.MaxWorkflows (20) is a fatal error (exit 1) naming the count and the cap, and writes no `-o` output file |
| TestGenerateConfigEmitWriteError | - | emit to path with missing parent dir returns exit code 1 |
| TestLoadEnvWorkflowsDirRelativeToEnvFile | - | workflows.dir resolved relative to env file not cwd; exit 0 and stdout has spring: |
| TestLoadEnvExcludesEnvFileFromWorkflowSet | - | env.yaml excluded from its own workflow scan; exit code 0 |
| TestDeployKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then apply) with argv [kubectl apply -f -], stdin contains kind: Deployment |
| TestRemoveKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then delete) with argv [kubectl delete -f -] |
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
| TestRemoveDockerSeam | - | exit 0, 2 runner calls (preflight then down) argv [docker compose -f <compose> down] |
| TestDeployPodmanSeamWritesUnitsAndStarts | - | exit 0, a leading podman info preflight call, then app yaml and container unit written to quadlet dir, systemctl daemon-reload then start calls |
| TestRemovePodmanSeamStopsRemovesReloads | - | exit 0, a leading podman info preflight call, then systemctl stop then daemon-reload calls, unit and app yaml files removed |
| TestPlatformFlagHitOverridesInference | - | an explicit --platform is used even when another section is also present in env.yaml |
| TestPlatformFlagMissingSectionIsLoudError | - | a --platform value with no matching section fails loud, naming both the requested and the present sections, before the runner is invoked |
| TestPlatformAliasesResolveToCanonical | kube / dk / pm | each short --platform spelling reaches the same platform binary as its canonical name |
| TestPlatformAliasMissingSectionNamesCanonicalSection | - | an alias is resolved before the section check, so the error names the `kubernetes:` section to add rather than echoing `kube` |
| TestPlatformUnknownValueListsEverySpelling | - | a bogus value (k8s) is rejected with every accepted spelling listed, canonical and short |
| TestPlatformSpellingsAreDeterministic | - | platformSpellings is built from an ordered slice, not map iteration, so the rejection message cannot vary between runs; canonical names lead |
| TestPlatformAliasesCoverEveryPlatformExactlyOnce | - | every alias maps to a real platform, no alias is declared twice or collides with a canonical name, and the lookup map matches the declared list |
| TestPlatformSingleSectionInferred | - | with no --platform and exactly one section present, that section is used and echoed to stderr |
| TestPlatformMenuOnMultipleSections | - | with no --platform and more than one section, the interactive menu (via the injected promptLine seam) picks the platform |
| TestPlatformMenuNonTTYRefusesWithPlatformHint | - | the menu refuses to block when stdin is not a TTY, failing with an error naming --platform instead of hanging |
| TestPlatformZeroSectionsIsLoudError | - | with no --platform and no section present at all, the error names all three section keys |
| TestOldPositionalFormsRejectedWithPlatformHint | deploy kubernetes / remove docker / generate podman | the pre-rework positional grammar is a usage error (exit 2) that points at --platform, not resolved as a target |
| TestStatusBannerNamesTheInstancePerPlatform | k8s / docker / podman | the banner carries namespace/deployment/pod, compose-project/container, or the container alone, and drops any name that is not set rather than leaving an empty segment |
| TestStatusDockerBannerCarriesComposeProject | - | a docker target's banner carries its compose project, read by one extra read-only inspect issued after the report is run |
| TestStatusDockerBannerWithoutComposeProject | - | a container compose did not create drops the group segment, leaving no dangling separator |
| TestStatusScriptPresentRunsAndReportsOutput | - | an already-installed script is just run; stdout prints its output under the target's `=== ... ===` banner, every line indented by two so the report stays left-aligned however long the pod name is |
| TestStatusAbsentPlusInstallFlagInstallsThenRuns | - | --install installs a missing script on the target with no prompt, then runs it |
| TestStatusAbsentPromptYesInstallsThenRuns | - | without --install, a "y" answer to the install prompt (via the injected promptLine seam) installs then runs |
| TestStatusAbsentPromptNoSkipsAndExitsOne | - | declining the install prompt skips the target (no install, no run) and the overall exit code is 1 |
| TestStatusStandbyReportedWithoutAbortingOtherTargets | - | standby prints like any other answer (the script always exits 0), the loop still reaches the next target, and the overall exit code stays 0 |
| TestStatusRunFailureReportedAndExitsOne | - | a non-zero exit from the run can only be a failed exec, so that target is reported on stderr and the exit code is 1, while a reachable target in the same run still reports |
| TestStatusInstallPromptNonTTYRefusesWithInstallHint | - | the install confirmation refuses to block when stdin is not a TTY, failing with an error naming --install and installing/running nothing |
| TestStatusTargetValidationRejectsBadPodAndNamespace | bad pod name / bad namespace | an unsafe --pod or --namespace value is rejected via validate.SafeToken before any exec |
| TestStatusRejectsUnsafeUserBeforeAnyExec | 4 names | a --user carrying '/', '$', a space or a quote is rejected via validate.SafeActuatorUser before any exec, since the name reaches a sed address in the script |
| TestStatusManagementPortBounds | -1 / 65536 | an out-of-range --management-port is rejected before any exec |
| TestVersionOutputShape | - | `version` prints `solmq-conn-util <version> <go version> <GOOS>/<GOARCH>`, exit 0; the package-level version var defaults to "dev" in an un-injected test build |
| TestAbsPath | absolute input | absPath returns input unchanged when already absolute |
| TestAbsPath | relative input | absPath joins relative path onto base dir |
| TestCommandsDocInSync | - | docs/commands.md equals what the command model renders; -update rewrites it instead of asserting |
| TestCommandsModelMatchesUsage | - | every InUsage command, every flag, and every verb alias in the model appears in usage(), and usage() lists no command the model omits |
| TestCommandsModelMatchesUsage | platform short spellings | usage() carries the `--platform` shorts as the pipe-joined list built from platformAliasList, so `-h` cannot fall behind the docs when that table changes (joined, not per-alias: `kube` alone is a substring of `kubernetes`) |
| TestAutoCompleteDispatchPrintsScript | bash / zsh / fish / powershell | `auto-complete <shell>` exits 0, writes the script to stdout, and never reaches the runner |
| TestCompletionGoldenInSync | bash / zsh / fish / powershell | each rendered script equals its snapshot under cmd/solmq-conn-util/testdata/completions; -update rewrites them |
| TestCompletionCoversModel | bash / zsh / fish / powershell | every modeled verb, target, flag spelling and verb alias reaches every shell (fish exempts a verb with no targets/posarg/flags, e.g. version, which has nothing beyond word 1 to normalize), with descriptions in the three shells that show them |
| TestCompletionRecognizesFlagAliases | bash / zsh / powershell | every spelling flag.Parse accepts (-e, --e, -env, --env) is in the value-skipping table, so a value is never mistaken for a positional |
| TestCompletionShellStructure | bash / zsh / fish / powershell | each script keeps the registration line that makes it load, and the zsh script opens with #compdef |
| TestCompletionVerbAliasesResolveToCanonical | bash / zsh / fish / powershell | each shell's own alias-normalization construct ($verb= case arm, __fish_seen_subcommand_from, $verbAlias[...]) maps every verb alias to its canonical verb name (same fish exemption as TestCompletionCoversModel) |
| TestCompletionVerbAliasesNotOfferedAtWordOne | bash / zsh / fish / powershell | no verb alias appears in the position-1 candidate list (compgen -W, the zsh verbs array, the __fish_use_subcommand lines, the powershell $verbs array) -- recognized everywhere, but never offered on TAB |
| TestCompletionValueKindsReachScripts | bash / zsh / fish / powershell | a path flag completes files and `examples` completes directories in every shell |
| TestCompletionOutputIsPlainASCIILF | bash / zsh / fish / powershell | generated scripts are plain ASCII, LF only, newline-terminated |
| TestCompletionModelMetadataComplete | - | every verb has a description and a known PosArg, every flag a known Arg and a non-empty Meaning, every modeled shell a renderer and a snapshot; verb/target names and verb aliases stay [a-z0-9-] for unquoted case patterns, and no alias collides with another verb, another alias, or -h/--help |
| TestPlainText | code spans stripped / newline folded / tab and CR folded / whitespace runs collapse / trimmed / control chars dropped / empty / only backticks / punctuation preserved | model text is reduced to a single-line tooltip that cannot break the enclosing shell statement |
| TestShellQuoting | plain / empty / apostrophe / backslash / dollar and backtick / double quote / semicolon and pipe | bashQuote, fishQuote and psQuote each neutralize their shell's escape rules |
| TestZshEntry | plain / colon in the value escaped / colon in the description left alone / apostrophe quoted / empty description | _describe entries split on the intended colon only |
| TestFlagAliasesAndOffered | short and long pair / long only | flagOffered suggests the documented spellings, flagAliases lists all four dash forms, fishFlagSpec renders -s/-l correctly |
