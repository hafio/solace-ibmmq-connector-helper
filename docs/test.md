# solmq-conn-util test catalogue

Every test in the suite, grouped by package and expanded to individual cases, so you can
see what behavior is covered and jump to the test that covers it. This is a living
document: when a test or case is added, removed, or renamed, update the matching row in
the same change. For build and release details see [DEVELOPMENT.md](DEVELOPMENT.md);
for using the tool, see [userguide.md](userguide.md).

Run the suite with `./scripts/dev.sh test` (`.\scripts\dev.ps1 test` on Windows);
measure coverage with the `cov` task.

## Contents

- [internal/scan](#internalscan)
- [internal/spec](#internalspec)
- [internal/consolidate](#internalconsolidate)
- [internal/tls](#internaltls)
- [internal/yamlwriter](#internalyamlwriter)
- [internal/render](#internalrender)
- [internal/logback](#internallogback)
- [internal/statusscript](#internalstatusscript)
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
- **The exec seam** (`internal/runner`) is faked by `fakeRunner`, which records the argv
  and stdin crossing the boundary instead of starting a process; the real `os/exec` path
  is exercised through the `TestHelperProcess` child-process pattern.
- **Columns**: a `-` in the Case column means the test runs once; otherwise Case carries
  the subtest name or the case's label/input value, in source order where the test defines
  one.
- Tests are cross-referenced by file and test name only -- no line numbers (they rot as
  tests move).

_Snapshot: 737 test functions, 1005 case rows across 18 packages. (Functions counted from `func Test` in the source; case rows are the data rows of the tables below, not a suite run -- human, please confirm against `./scripts/dev.sh test` / `cov` output.)_

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
| TestParseDefaultsLeaderSession | - | an inline `session:` block parses the full solace tuple, api-properties included, and leaves the SolaceKey marker unset |
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
| TestParseDefaultsFull | - | parses tls stores, management port 8090 with exposure health (not a configurable key, but parsed anyway so validate can reject it), security enabled false with 1 user, leader-election standalone, logging/solace-defaults nodes captured |
| TestParseSecurityUserRoles | absent / one / several | security.users[].roles parses to no roles (the connector's read-only default), a single role, and several in authored order |
| TestParseDefaultsSecurityEnabledKeyOmittedStaysNil | - | security.enabled is not a configurable key: an omitted key parses to Security.Enabled nil rather than being defaulted |
| TestParseDefaultsEmpty | - | empty input yields a zero-valued Management (Management{}), Security.Enabled nil with no users, and TLS.Truststore nil |
| TestParseDefaultsError | - | malformed tls yaml returns non-nil error |
| TestParseKubernetesReplicasDefault | - | deployment without replicas defaults Replicas to 1 |
| TestParseKubernetesFull | - | parses replicas 2, service enabled port 8090, credentials create name, stores create present; source and variables set on credentials.create parse structurally regardless, so RemovedKeys can report them |
| TestParseKubernetesError | - | deployment as sequence instead of map returns non-nil error |
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
| TestApplyDockerDefaultsFillsMissing | - | docker defaults command/name/project-name/restart applied, ports stay empty (publishing is opt-in), stores/libs stay nil |
| TestApplyDockerDefaultsOverrideWins | - | explicit command/name/project-name/restart/ports override defaults exactly as given; a custom name does not drag project-name with it |
| TestApplyPodmanDefaultsFillsMissing | - | podman defaults command/name/restart applied, ports stay empty (publishing is opt-in), and both removed keys left alone: mode empty and Quadlet nil. Defaulting either would trip validate's rejection for something the operator never wrote, and nothing dereferences Quadlet any more -- the unit dir comes from the invoking uid |
| TestApplyPodmanDefaultsOverrideWins | - | explicit command/name/restart/ports override defaults exactly, and a present quadlet: block still decodes non-nil so validate can reject it by name |
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
| TestCredCreateRemovedKeys | - | RemovedKeys reports each of source, variables, values-file alone and all three in order; a nil receiver and a bare create.name report none |
| TestImageRef | hub / private registry / digest / trailing slash / no tag / no name / nil | Ref() assembles repo/name:tag, drops an unset repo, and joins a sha256: tag with `@` so a digest pin is a reference an engine accepts |
| TestImageRegistry | hub fallback / private / nil / trailing slash | the auths key is the registry host, falling back to Docker Hub's v1 URL -- a Hub namespace lives in name and never reaches the lookup |
| TestRetiredPerPlatformImageStillParses | - | kubernetes.deployment.image, docker.image and podman.image parse into their fields, which is what lets validate reject them instead of yaml dropping them silently |
| TestImagePullSecretCreateDefaultsFalse | absent / explicit true | create defaults to false so naming a Secret only references it; an explicit true is honoured |
| TestParseEnvTopLevelSyslog | present / absent | logging.syslog parses beside logging.level at the top level, protocol defaults to udp, and an absent block stays nil (presence is what turns syslog on) |

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
| TestAppendPassthroughCollision | passthrough T collides with existing T | output has 2 props and 1 collision warning reading `binder "bndr": passthrough overrides tool-managed key "T"; tool value kept` byte for byte, since the owner label is caller-supplied now |
| TestNodeToProps | mapping with scalar and nested keys | 2 props, first key k1/val v1, second has Sub set |
| TestNodeToProps | nil node | nodeToProps returns nil |
| TestBuildMQmTLSBundle | mq TLS side with cipher and keyAlias plus solace target | MQTLS true, 1 bundle, HasKeystore true, KeyAlias mc, KeystoreTyp PKCS12, TruststoreTyp JKS |
| TestBuildCipherConflictWarning | two mq sources with different ciphers C1/C2 | warnings contain conflicting cipher |
| TestBuildMessageLoopWarning | same side used as source and target dest `SAME` | warnings contain message loop |
| TestBuildSolaceTopicSourceEmitsConsumerTopic | solace topic source -> mq queue target | input-0 solace binding is consumer with DestType topic |
| TestBuildStorePathsRawVsMount | mount=false (config) | TruststoreLoc reflects env.yaml path verbatim ./certs/t.jks |
| TestBuildStorePathsRawVsMount | mount=true (deploy) | TruststoreLoc rewritten to /app/external/classpath/truststores/t.jks |
| TestBuildLeaderElection | dangling conn-ref missing | le non-nil with Mode/Queue set but Session nil, no guard warning |
| TestBuildLeaderElection | conn-ref happy path tcps host | Session host/vpn set and APIProps has SSL_TRUST_STORE, SSL_KEY_STORE, SSL_PRIVATE_KEY_ALIAS mounted paths and alias sc; Extras carry solace-defaults in authored order and the connection api-properties land last, after the tool TLS keys |
| TestBuildLeaderElection | inline session non-tcps host | Session set from inline fields with no TLS APIProps, but the inline api-properties and solace-defaults still come through |
| TestBuildLeaderElection | mount rewrite raw vs mnt for leader election truststore | raw keeps ./certs/truststore.jks verbatim, mnt rewrites to /app/external/classpath/truststores/truststore.jks |
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
| TestApplicationMinimalNoOptionalBlocks | ssl: / logging: | ssl: and logging: blocks stay absent when defaults are empty |
| TestApplicationMinimalNoOptionalBlocks | management: / security: | management: and security: are unconditional now: the fixed exposure list and the reserved solmq-status account render even with empty defaults |
| TestApplicationMinimalNoOptionalBlocks | type: undefined | undefined binder is always emitted even with minimal config |
| TestApplicationLeaderElection | - | leader-election, fail-over, queue and management render for active_standby, and the whole session block matches exactly: binder-shared credential names, solace-defaults between the credentials and api-properties, verbatim passthrough last |
| TestApplicationLeaderElectionSessionMatchesBinderKeySet | - | rendered from one connection, the session key sequence equals the binder solace.java key sequence -- the guard against the two renderers drifting apart again |
| TestApplicationLeaderElectionSessionPlaintext | - | a non-tcps session emits no SSL_ keys but still carries solace-defaults and its own api-properties, and falls back to the LEADER_ELECTION_* names |
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
| TestRetiredPerPlatformImageRejected | kubernetes / docker / podman | each per-platform image key errors, and the message names the top-level image: block to use instead |
| TestImageBlockRequired | absent / no name / no tag / unsafe repo, name, tag | the top-level block is required once a platform is in play, tag included (an untagged image resolves to :latest and pins nothing), and the fields that reach an argv are charset-checked |
| TestImageBlockRequired | bad pass-env name / either credential set both ways | the registry account (`user`/`pass`) goes through the shared checkCred, so it gets the same literal-xor-env rule and variable-name check as every other credential |
| TestImageNotRequiredWithoutAPlatform | - | `generate config` renders application.yml alone and pulls nothing, so no image is demanded |
| TestImagePullSecretChecks | name alone | referencing a Secret requires no registry credentials at all |
| TestImagePullSecretChecks | name required / DNS-1123 | the Secret name is required and held to the label rule the cluster would apply |
| TestImagePullSecretChecks | create without credentials | create errors unless the registry account is set, in either the literal or the -env form |
| TestImagePullSecretChecks | create, variable unset / set | an unset variable warns rather than errors, so a config can be linted without the deploy secrets |
| TestRetiredPerPlatformTimezoneRejected | kubernetes / docker / podman | each per-platform timezone key errors and names the top-level timezone: key |
| TestTopLevelTimezoneUnsafe | unsafe / realistic | the top-level timezone keeps the charset gate the per-platform key had, and an empty value is not an error |

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

Render the POSIX status script the generated deploy artifacts embed and `solmq-conn-util status` execs inside each running instance -- a pure renderer with no os/exec, filesystem, or network access.

Tests: [statusscript_test.go](../internal/statusscript/statusscript_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderSubstitution | defaults | Render substitutes PORT=8090 and USER_NAME=solmq-status; BASE and the leaderelection/workflows endpoints are built from $PORT at run time, not by Render |
| TestRenderSubstitution | non-default port and user | Render substitutes PORT=19090 and USER_NAME=custom-mgmt-user |
| TestRenderIsPureASCIINoCRLF | - | output has no carriage return, no byte over 127, and ends with a trailing newline |
| TestRenderHeaderHasExecOneLiners | - | the header pins the kubectl/docker/podman exec one-liners, each built from ContainerPath |
| TestRenderPasswordResolution | - | the password-lookup chain references ContainerPath, SecretsDir and the from_configs account lookup, and no credential is embedded |
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
| TestFilenameAndPathConstants | - | the script's name and directory, and that ContainerPath is not nested inside the libs, spring/config or classpath mounts -- the nesting that made the libs mount shadow it |
| TestRenderReportsHealthComponents | - | the per-component health breakdown: a newline before every `{"status"` puts each component's status at the start of a line and its name at the end of the line above, so the name is carried forward in $pending (guarded with `${pending:-}` for set -u); the block prints only when something parsed |
| TestRenderReportsJavaConfigAndHeap | - | the three details-level lines from outside the report endpoints: `java -version` (stderr redirected, folded to "openjdk 17.0.9" or passed through raw), the config the report was read from, and heap used/max tagged `area:heap`; each guarded so an absent source drops its line, a negative maximum is left out, and the byte arithmetic is deliberately *not done* here (busybox would read Jackson's 4.32013312E8 as 4) |
| TestRenderHeaderNamesEveryReportedFact | - | the script's own header names what it reports, since it is the first thing someone running the script by hand reads |
| TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd | - | the child prints one line then blocks far longer than the test waits, so seeing that line proves output is not buffered until exit; cancelling then ends the run and reports nil, because a follow the operator stopped did not fail |
| TestOSStreamKeepsStdoutAndStderrApart | - | Stream's two writers stay separate where Run merges into one, so `logs > app.log` captures the log and leaves the platform's diagnostics on the terminal |
| TestOSStreamReportsAFailureThatWasNotCancelled | - | an uncancelled non-zero exit is still an error, and output written before it still reaches the writer |
| TestOSStreamRejectsEmptyAndUnresolvableArgv | - | Stream refuses exactly what Run refuses (both go through resolveArgv0), naming the binary it could not resolve |
| TestLogsArgvPerPlatform | kubernetes bare / kubernetes without a namespace or container / kubernetes with every option / docker bare / docker with every option it has / podman reads the container, never the journal / tail zero is not tail all | kubectl takes the pod positionally with the namespace and container as flags; docker and podman take options first and the container name last; TailAll adds no flag while an explicit 0 does |
| TestLogsArgvRefusesPreviousOffKubernetes | docker / podman | `--previous` is refused by name rather than dropped, so a caller asking for the previous log is never handed the current one |
| TestLogsArgvUnknownPlatform | - | an unrecognised platform names all three that exist rather than producing a half-built argv |

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
| TestTeardownReversesTheDocumentOrder | - | with a libs PVC present, apply orders the claim before the Deployment that mounts it; teardown fully reverses the set -- Deployment before the claim, Service before ConfigMap -- with separators intact and one fewer than apply's (the dropped Namespace) |
| TestLibsPVNameIsNamespaced | - | the same `libs.pvc.create.name` in two different namespaces derives two different PV names, each suffixed `-pv` and carrying its own namespace |

## internal/dockergen

Render the docker compose file from the target model.

Tests: [dockergen_test.go](../internal/dockergen/dockergen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderFull_WithEverything | - | full golden output with project-name, creds, store, libs, MQTLS, ports, timezone matches exactly; the compose project is the document's first line |
| TestRenderFull_Minimal | - | minimal golden output omits the leading name: line along with restart, ports, environment, env_file, volumes blocks |
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
| TestStatusScriptConfigSourceAndTarget | - | the service references a second `<name>-status` config and mounts it at /app/external/.status-script |
| TestStatusScriptContentIsEscaped | - | the status script body is inlined under the status config's content: block, indented 6 spaces, blank line preserved as truly empty, and its shell `$` doubled |
| TestContentEscapesDollarsForCompose | $VAR / ${VAR} / ${VAR:-default} / $(cmd) / $$ | each shape reaches the content block with every `$` doubled, so compose's interpolation pass delivers it unchanged instead of blanking it or rejecting the document |
| TestContentEscapesDollarsForCompose | no lone `$` | dropping every `$$` pair from the rendered document leaves no `$` behind anywhere |
| TestAppYAMLSecretPlaceholdersAreNotInterpolated | - | application.yml's ${...} credential placeholders render doubled so compose cannot substitute the values the CLI passes it, while the `secrets:` provider entries -- compose's own -- stay unescaped |
| TestRenderSyslogAddsConfigAndEnv | - | compose inlines the logback config as a third configs entry targeted at logback.ContainerPath, and sets the three LOGGING_SYSLOG_* env vars the config reads at runtime |
| TestRenderSyslogTCPUsesTheLogstashAppender | - | the protocol reaches the inlined config: tcp renders the logstash appender, since the two have different classpath requirements and the wrong one fails at runtime rather than at generate time |
| TestRenderWithoutSyslogEmitsNeither | - | no block, no config entry, no env vars |

## internal/podmangen

Render the quadlet unit from the target model.

Tests: [podmangen_test.go](../internal/podmangen/podmangen_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestRenderQuadletSecretsCarryNoValues | - | each credential appears only as a Secret= directive naming its store entry and absolute target; the value itself never reaches the unit, which sits on disk beside application.yml |
| TestRenderQuadletFull | - | full input yields 1 unit named solmq-connector.container with content matching golden string |
| TestRenderQuadletMinimal | - | minimal input yields 1 unit with content matching golden string (no Service section, no restart) |
| TestLeaderLabelsPerMode | empty defaults to standalone / standalone / active_active | the unit carries the le-mode label and role: active |
| TestLeaderLabelsPerMode | active_standby | the unit carries le-mode active_standby and withholds role: active |
| TestStatusScriptMountNestsAfterLibs | - | the status script volume is declared after the libs volume, so it nests rather than being shadowed |
| TestStatusScriptMountOmittedWhenPathEmpty | - | an empty StatusScriptPath omits the status volume entirely, rather than mounting an empty source |
| TestImagePullSecretStates | no block / name alone / created | imagePullSecrets is absent, rendered without a Secret, or rendered with one -- the middle case is what stops an apply overwriting a Secret the operator built |
| TestImagePullSecretPayloadIsOpaqueToDeploy | - | deploy places the already-encoded payload verbatim and the registry password appears nowhere outside it |
| TestEnvBlockOmittedWhenEmpty | nothing to emit | env: is omitted entirely rather than rendered with nothing beneath it, since the timezone is one optional top-level key |
| TestEnvBlockOmittedWhenEmpty | TZ only / MQTLS only | either entry alone opens the block, and no timezone means no TZ entry |
| TestQuadletSyslogMountsAndSetsEnv | - | podman cannot inline file content, so the unit bind-mounts the logback file read-only via Volume= and sets the three LOGGING_SYSLOG_* vars via Environment= |
| TestSyslogAbsentEmitsNoMountOrEnv | - | no block, no mount, no env |

## internal/runner

The os/exec seam -- ParseCommand safe-tokenizing, kubectl/docker/podman deploy and remove argv, the status verb's read-only queries, the logs verb's streaming seam and per-platform argv, the cli verb's terminal-attach seam and the shared exec argv builder, quadlet scope resolution, and WriteFile modes.

Tests: [runner_test.go](../internal/runner/runner_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestHelperProcess | - | helper child-process entry point guarded by GO_WANT_HELPER_PROCESS; dispatches stdin/both/fail/drip/env/stdiokind/unknown modes |
| TestExecArgvPerPlatform | kubernetes bare / with a namespace / with stdin / with a terminal / docker / podman | the shape each engine's flag parser actually accepts: kubectl takes -n and -c after the pod and needs the `--` terminator, docker and podman stop parsing at the container name so their flags come first and no `--` is added; the container is a constant on every kubernetes row |
| TestExecArgvRefusesATTYWithoutStdin | - | a terminal the child cannot read from can only be a caller mistake, so it is refused rather than quietly produced |
| TestExecArgvUnknownPlatform | - | an unrecognised platform names the three that exist rather than producing a half-built argv |
| TestOSAttachHandsTheChildTheCallersFilesNotPipes | - | the child reports that all three standard files are the caller's own regular files; os/exec substitutes a pipe for any writer that is not an *os.File, and a pipe would make the engine refuse the tty |
| TestOSAttachStdinIsTheHandedFile | - | what the caller passes as stdin is what the child reads, with nothing copied through this process |
| TestOSAttachReportsTheChildExitStatus | - | a session that ran and exited 3 is reported as 3 with no error, which is what the cli verb's exit-code contract rests on |
| TestOSAttachEnvReachesChild | - | the third call site of the applyCmdEnv split: supplied values win over ambient ones here exactly as under Run |
| TestOSAttachRefusesACmdCarryingStdinText | - | "write this string to the child" and "give the child the terminal" are contradictory, so the Cmd is refused by name rather than one of them silently honoured |
| TestOSAttachRefusesANilFile | no stdin / no stdout / no stderr | refused here rather than at the child's first write, where os/exec panics on a typed-nil *os.File |
| TestOSAttachRejectsEmptyAndUnresolvableArgv | - | mirrors the Run and Stream refusals; all three go through resolveArgv0 so the rules cannot drift apart |
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
| TestResolveQuadletScope | follows euid | UserMode tracks `os.Geteuid() != 0`, and the directory pairs with it: user mode under the home dir, system mode at quadletSystem. Asserts the pairing rather than one fixed answer, since the answer legitimately differs for a root run |
| TestResolveQuadletScope | home redirect | in user mode the dir tracks HOME/USERPROFILE, which is what lets a test (or a relocated home) move it without a config key -- there is no scope or dir key any more |
| TestPodmanDeployReloadThenStart | - | user mode issues daemon-reload then start with `--user` flag in order |
| TestPodmanDeploySystemModeNoUserFlag | - | system mode issues systemctl daemon-reload without `--user` flag |
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
| TestPreflightKubernetesArgvDeployNoNamespace | - | kubernetes deploy preflight issues <argv> auth can-i create deployment with no `--namespace` when namespace is empty |
| TestPreflightKubernetesArgvRemoveWithNamespace | - | kubernetes remove preflight issues <argv> auth can-i delete deployment `--namespace` <ns> |
| TestPreflightDockerArgvIsInfo | - | docker preflight issues <argv> info |
| TestPreflightPodmanArgvIsInfo | - | podman preflight issues <argv> info |
| TestPreflightFailureWrapsPlatformHint | kubernetes / docker / podman | a failing probe's error contains "preflight failed for <platform>", preserves the underlying cause, and carries the platform's login/daemon hint |
| TestPreflightRejectsDisallowedBinaryBeforeRunning | - | a command outside the platform allowlist (curl) is rejected before the probe ever runs, zero runner calls |
| TestPreflightExtraAllowedThreadsThrough | - | "sudo podman" is rejected without extraAllowed and accepted with it, running argv [sudo podman info] |
| TestPreflightUnknownAction | - | an action other than deploy/remove is rejected, zero runner calls |
| TestKubernetesPodsJSONArgv | by selector, no namespace / by selector in a namespace / explicit names / every namespace | `get pods ... -o json` in each of its three scopings; `--all-namespaces` outranks a namespace resolved from env.yaml, so a cluster-wide search is never narrowed back down |
| TestKubernetesPodsJSONRunFailureWraps | - | a run failure surfaces as an error naming the "listing pods" operation |
| TestKubernetesGetJSONArgv | deployment in a namespace / no namespace / a referenced object | `get <kind> <name> [-n ns] -o json` for the workload summary and for each object a pod references |
| TestKubernetesGetJSONMissingObjectIsAnError | - | a missing object exits non-zero and the error names it, which is the answer the components check asks for -- reported as MISSING rather than as a failed run |
| TestKubernetesTopArgv | by selector / explicit names / every namespace | `top pod --containers --no-headers`, attributing the sample to one container so it is comparable with that container's limits |
| TestKubernetesTopWithoutMetricsAPIWraps | - | a cluster with no metrics API fails here, naming the operation, so the caller can degrade to a note |
| TestEngineInspectJSONArgv | - | one `inspect a b c` covers every target, and a chained command keeps its own tokens ahead of the subcommand |
| TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing | - | inspecting nothing is an error and starts no process |
| TestEngineInspectJSONFailureNamesTheTargets | - | a failure names the containers it was asked about |
| TestEngineImageInspectJSONArgv | - | `image inspect <ref>`, where the registry digest lives (the container's own inspect does not carry it) |
| TestEngineStatsArgv | - | `stats --no-stream --format <template>`: the four tab-separated template fields docker and podman render identically, and --no-stream so the call cannot stream forever |
| TestEngineListArgv | - | `ps --all --no-trunc --format <template>` for `--all` discovery, including stopped containers -- an instance that died is what such a search is for |
| TestSystemctlNRestarts | user scope / system scope / unknown unit answers nothing / no systemd on this host | `systemctl [--user] show <unit> -p NRestarts --value`, the only truthful restart count under quadlet; an empty or unreadable answer is an error so the caller can fall back to the container's own counter |
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
| TestOSRunSplitKeepsTheStreamsApart | - | RunSplit keeps stdout and stderr separate where Run merges them, so a stderr deprecation warning from `oc get -o json` cannot land ahead of the JSON and break the parse |
| TestOSRunSplitWiresStdinAndEnv | - | the split path wires stdin and env exactly as Run does, so a helper's choice between them is invisible to its caller |
| TestOSRunSplitRejectsEmptyAndUnresolvableArgv | - | RunSplit refuses an empty or unresolvable argv[0] through the same resolveArgv0 seam as Run, Stream and Attach |
| TestParsingHelpersIgnoreAWarningOnStderr | KubernetesPodsJSON / KubernetesGetJSON / KubernetesListJSON / KubernetesTop / EngineInspectJSON / EngineImageInspectJSON / EngineStats / EngineList / SystemctlNRestarts | every helper that parses rather than scans its output returns the payload alone when a warning sits on stderr, via the Splitter path |
| TestParsingHelpersFallBackToRun | - | a Runner with no Splitter still works, on exactly the combined output Run has always returned |
| TestParsedFailureCarriesBothStreams | - | a failed command still reports what it said, on whichever stream it said it on, so splitting the streams did not cost the error context |

## internal/validate

Validate the parsed model -- per-side rules, connection refs, leader election, the docker/podman/kubernetes platform sections, ports, container names, TLS/stores wiring, and the safe-token charset.

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
| TestCredentialsEnvChecks | - | unset env var for credentials errors variable MISSING_VAR is not set |
| TestCheckSideMQMissingFields | - | mq side missing conn-name/queue-manager/channel errors each; user/password not flagged missing |
| TestCheckSideSolaceMissingAndBadScheme | missing-host-vpn | empty solace side errors missing host and msg-vpn; client creds not flagged |
| TestCheckSideSolaceMissingAndBadScheme | bad-scheme | http scheme host errors must start with tcp:// or tcps:// |
| TestSolaceKeyAliasRequiresTCPSAndKeystore | - | solace key-alias with plain tcp host errors requires a tcps:// host |
| TestMQKeyAliasRequiresKeystore | - | mq key-alias without keystore errors no keystore defined |
| TestCheckKubeRequiredAndReplicas | - | kube deployment missing name/namespace and replicas 3 errors each field plus replicas: 1 message |
| TestCheckKubeServicePort | - | kubernetes.service.port is range-checked like docker/podman ports: a scalar or distinct host:container pair both pass, and an out-of-range host or container side each error independently naming the offending side |
| TestCheckKubeCredentialCreateRemovedKeys | source/variables/values-file set | credentials.create carrying `source`, `variables`, and `values-file` errors naming all three and telling the operator to remove them |
| TestCheckKubeCredentialCreateRemovedKeys | source alone | credentials.create carrying only `source` errors naming it alone |
| TestCheckKubeCredentialCreateRemovedKeys | bare name | a bare create.name trips no removed-keys error |
| TestCheckKubeStoresRequireTruststore | - | kube stores create without tls.truststore errors requires tls.truststore |
| TestStoresNotWiredWarning | - | TLS workflow with kube deploy and no stores wiring warns secrets.stores is omitted |
| TestStoresWiredExistingNoWarning | - | stores wired via existing secret produces no stores-omitted warning |
| TestCheckCredRules | both literal and env set | errors sets both a literal value and target password-env |
| TestCheckCredRules | env value is a ${...} reference | errors must be a bare variable name, not a ${...} reference |
| TestCheckCredRules | env value not a valid identifier | errors is not a valid environment variable name |
| TestCheckCredRules | -env var unset in this environment | warns (not errors) which is not set in this environment |
| TestCheckCredRejectsReservedPrefix | -env starting with spec.GeneratedNamePrefix | rejected: the prefix is reserved for derived mount names, which is what keeps an operator's name and a derived name from ever meeting |
| TestCheckCredRejectsReservedPrefix | MY_GEN_PASSWORD / _MY_PASSWORD / SOL_PASSWORD | accepted -- only a name *starting* with the prefix is reserved |
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
| TestCheckPodmanModeAndScope | quadlet present / omitted | a present podman.quadlet block of any shape errors `podman.quadlet is no longer configured`; omitting it is clean. Both keys are gone: scope only ever worked when it agreed with the invoking uid, and dir could only move the unit somewhere systemd does not scan |
| TestCheckPodmanStoresRemoved | present / nil | a present podman.stores errors `podman.stores is no longer configured`; omitting it trips no such error |
| TestCheckPodmanBaseDirRequired | omitted | `podman.base-dir is required` -- it is baked into the unit's Volume= lines, so there is no safe default to guess |
| TestCheckPodmanBaseDirRequired | unsafe / relative | a whitespace-bearing base-dir is rejected by the same host-path gate as libs.dir; a relative one is accepted and resolves against env.yaml at render time |
| TestCheckCommandMultiToken | safe-multi-token | docker command with extra safe tokens has no unsafe-character error |
| TestCheckCommandMultiToken | unsafe-token | docker command with $(evil) token errors unsafe character |
| TestCheckDeployCommandAcceptReject | kubectl / oc / kubectl with flags / docker with flag / podman / kubectl.exe / sudo podman with extraAllowed | accept matrix: bare allowlisted argv[0], flag-shaped args, .exe-stripped comparison, and a chained binary approved via extraAllowed all pass |
| TestCheckDeployCommandAcceptReject | curl / absolute path / relative path / bare positional arg / sudo podman without extraAllowed / bare "--" / empty command | reject matrix: unlisted binary, path argv[0], a bare positional argument, an unapproved chained binary, a bare end-of-flags marker, and an empty command all error |
| TestCheckDeployCommandEndOfFlagsMarkerMidCommand | - | "kubectl --" errors token "--": end-of-flags marker is not accepted, distinct from the argv[0] allowlist rejection |
| TestCheckDeployCommandErrorTexts | - | pins the canonical wording verbatim for the path, allowlist, end-of-flags, flag-shape, and empty-command errors |
| TestCheckKubeCommandNowValidated | - | an unsafe kubernetes.command such as "kubectl; rm -rf /" errors; a safe kubectl command produces no such error |
| TestCheckKubeCommandDefaultKubectlUnvalidated | - | the zero-value default (spec.DefaultKubeCommand) validates clean |
| TestContextAllowCommandsHonored | - | Context.AllowCommands threads into checkKube and checkContainerTarget: "sudo docker"/"sudo podman"/"sudo kubectl" reject with AllowCommands nil, accept with AllowCommands=[sudo] |
| TestCheckContainerCommandUnlistedBinaryRejected | - | docker.command "curl" and podman.command "/tmp/evil" are rejected by the platform allowlist, not merely the charset check |
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
| TestDockerPodmanTLSNeedsNoStoresOptIn | docker / podman | a TLS workflow with no stores: block warns about nothing -- the store files are bind-mounted whenever tls.*.file is set, so the old "will be missing at runtime" case cannot arise |
| TestDockerPodmanStorePathAlwaysGated | docker / podman | an unsafe character in tls.truststore.file is rejected with no stores: block present, since those paths are always bind-mount sources |
| TestDockerPodmanStorePathAlwaysGated | kubernetes | the same path is not gated for kubernetes, which embeds the store content in a Secret rather than naming a host path |
| TestUsesTLS | solace-tcps-host | solace side with tcps host returns usesTLS true |
| TestUsesTLS | mq-tls-true-no-tcps | no solace side, mq tls true returns usesTLS true |
| TestUsesTLS | plain-tcp-mq-false | plain tcp solace and mq tls false returns usesTLS false |
| TestCheckContainerRestartUnsafe | newline in restart | docker.restart is rejected; image and timezone are top-level keys, covered by their own per-platform-rejection and charset tests |
| TestCheckContainerRestartUnsafe | realistic value | on-failure:5 is accepted |
| TestCheckContainerHostPathsUnsafe | newline in tls.truststore.file | bind-mounted store path rejected |
| TestCheckContainerHostPathsUnsafe | space in libs.dir | podman.libs.dir rejected |
| TestCheckContainerHostPathsUnsafe | windows paths | `C:\certs\...` store paths and `C:\libs` accepted (backslash and colon permitted) |
| TestCheckKubeSecretNames | cred create bad | non-DNS-1123 credentials create.name rejected |
| TestCheckKubeSecretNames | cred create empty | missing credentials create.name reported as required |
| TestCheckKubeSecretNames | cred existing bad | non-DNS-1123 credentials existing rejected |
| TestCheckKubeSecretNames | stores create bad | non-DNS-1123 stores create.name rejected |
| TestCheckKubeSecretNames | stores existing bad | non-DNS-1123 stores existing rejected |
| TestCheckKubeSecretNames | valid names | solmq-credentials and solmq-tls produce no name error |
| TestCheckKubeSecretsCreateXorExisting | credentials both set | rejected: Render would take the create branch and emit a Secret doc over the object existing names |
| TestCheckKubeSecretsCreateXorExisting | credentials neither set | rejected: a present block must choose, or the SecretsDir mount silently disappears |
| TestCheckKubeSecretsCreateXorExisting | stores both / neither set | same rule enforced for the stores Secret |
| TestCheckKubeSecretsCreateXorExisting | create only / existing only / blocks omitted | all three accepted -- omitting a block stays the way to say "none" |
| TestCheckLibsNFSFields | newline in nfs.server | rejected against the host charset |
| TestCheckLibsNFSFields | newline in nfs.path | rejected against the host-path charset |
| TestCheckLibsNFSFields | valid server and path | nfs1.corp.example and /solace-libs accepted |
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
| TestCredentialsNotWiredWarning | - | a config referencing credentials with kubernetes.secrets.credentials omitted warns kubernetes.secrets.credentials is omitted, naming the mount path they would have used |
| TestCredentialsWiredNoWarning | create / existing | either way of wiring kubernetes.secrets.credentials suppresses the warning |
| TestNoCredentialsNoWarning | - | a config whose connections need no authentication is not warned about a Secret it does not need |
| TestCredentialsFoundOutsideAWorkflowSide | a management account / a truststore password | a management account password and a store password are credentials too, and each trips the same warning as a missing binder credential |
| TestLibsPVNameLengthIsCapped | - | a 40-char namespace and a 40-char `libs.pvc.create.name` derive a PV name that exceeds the 63-char DNS-1123 limit, which errors; shortening both to fit is not flagged, so the check cannot simply always fire |

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
| TestTargetMounts | tls+libs configured | 2 store mounts at the fixed default store path, and a libs mount whose source is the resolved abs host dir and whose target is the fixed default libs path -- both container-side paths come from constants, not from the spec |
| TestTargetMounts | no tls, no libs | yields nil,nil -- with stores derived, an absent tls block is the only way to get no store mounts |
| TestTargetMounts | store with no file | a tls.*.store present but with an empty file is skipped rather than mounted from an empty source |
| TestResolveCredentials | nil refs | no kvs, no error |
| TestResolveCredentials | literal + -env mix | literal ref passes through, -env ref reads from the resolver's environment |
| TestResolveCredentials | unset -env variable | fails loud naming the stable secret and the variable, never a value |
| TestResolveCredentials | no environment access | fails loud rather than silently resolving to empty |
| TestConfigRejectsSecretNameConflict | security.users "ops.1" and "ops-1" | Config renders nothing and errors naming the contested key *and both claiming positions*, rather than emit a config where one credential silently takes the other's password |
| TestConfigRejectsSecretNameConflict | same spec through Validate | the collision is caught while linting too, not only at generate/deploy -- names are assigned in consolidate, so Validate builds to see them |
| TestValidateCleanSpecStillPasses | no collision | the build call Validate now makes adds no errors of its own, and consolidate's warnings do not leak into validate output |
| TestConfigWorkflowCap | 21 workflows | Config produces no output and one error naming the count, the 20 cap, and the split-into-folders remedy |
| TestConfigWorkflowCap | 20 workflows | exactly at the cap does not error |
| TestGenerateKubernetesWorkflowCap | 21 workflows | GenerateKubernetes produces no manifest and the same workflow-cap error |
| TestConfigCarriesSecurityUserRoles | - | end-to-end: a roles-bearing env.yaml validates clean and its role reaches the rendered application.yml, while the reserved account still renders none |
| TestConfigNoSecretsLeak | - | every rendered password is a ${STABLE} placeholder except the one permitted literal: the reserved spec.StatusUserName account |
| TestGenerateDockerBasics | - | generates non-empty compose opening with the defaulted `name: solace-ibmmq-connectors` project line and containing the image; all four credential positions render as top-level environment-provider secrets, never inlined as values, and each ${STABLE} placeholder in application.yml is doubled so compose cannot interpolate the value in |
| TestGeneratePodmanQuadlet | - | produces the `<name>.container` unit with the app yaml name, service name and 4 secrets, each mounted from podman's store by its namespaced name at an absolute target under the secrets mount |
| TestGeneratePodmanRejectsModeKey | run / quadlet | the removed podman.mode is rejected at generate for either former value |
| TestGeneratePodmanNoModeKeyIsClean | - | an omitted mode: generates cleanly, guarding the removed applyPodmanDefaults default that would otherwise trip the rejection for every section |
| TestResolveStatusPasswordFixedRand | - | a fixed Rand hook yields the exact 32-lowercase-hex-char literal (16 bytes hex-encoded) |
| TestResolveStatusPasswordEnvOverride | - | a set, non-empty spec.StatusUserPasswordEnvVar is used verbatim and Rand is never consulted |
| TestResolveStatusPasswordEmptyEnvFallsBackToRand | - | an empty override is treated as unset, falling back to Rand rather than returning "" |
| TestResolveStatusPasswordRandError | - | a Rand failure surfaces as an actionable error naming the underlying cause, never a predictable fallback password |
| TestConfigStatusPasswordRandErrorNoOutput | - | the same Rand failure through Config is a hard error with no output |
| TestGenerateKubernetesCarriesStatusScript | - | the ConfigMap gets a "status: \|" key carrying the rendered script, addressed to spec.StatusUserName on the resolved management port |
| TestGenerateDockerCarriesStatusScript | - | compose gets a second top-level config (`<name>-status`) inlining the rendered script, mounted at statusscript.ContainerPath |
| TestGeneratePodmanCarriesStatusScript | - | PodmanPlan.StatusScript names `<name>-status` and the unit's Volume= for it is BaseDir-resolved exactly like AppYAML, since systemd starts the unit with no useful cwd |
| TestGenerateMissingTargetSection | kubernetes | error contains kubernetes target requires a 'kubernetes:' section in env.yaml |
| TestGenerateMissingTargetSection | docker | error contains docker target requires a 'docker:' section in env.yaml |
| TestGenerateMissingTargetSection | podman | error contains podman target requires a 'podman:' section in env.yaml |
| TestGenValidateStoresWarning | - | kubernetes credentials.create (name-only) plus a TLS-without-stores config yields no errors and exactly one stores-omitted advisory warning |
| TestGeneratorPageGoldenInSync | - | the golden embedded in solmq-conn-util-generator.html matches testdata/golden/application.yml (regenerate with -update-html-golden) |
| TestGoldenConfig | - | generated config output matches testdata/golden/application.yml byte-for-byte, one instance |
| TestGoldenKubernetesCreate | - | generated kubernetes manifests (namespace, configmap incl. status script, secret, stores, pv, pvc, deployment with secrets-volume/stores/syslog/libs mounts and le-mode/role labels, service) match golden fixture byte-for-byte |
| TestGoldenKubernetesNoSecrets | - | generated manifests without secrets/syslog/libs (namespace, configmap incl. status script, deployment, service) match golden fixture byte-for-byte |
| TestDockerConfigJSON | private registry / docker hub fallback | the payload is an auths map keyed by registry carrying the account plus the base64 user:password the engines send |
| TestDockerConfigJSONEscapesAwkwardValues | - | a password carrying quotes, a backslash or JSON of its own round-trips as data rather than reshaping the document -- which is why it is marshalled, not concatenated |
| TestResolvePullSecret | reference only | resolves to the name alone and never reads the registry password |
| TestResolvePullSecret | create | builds the payload from the environment |
| TestResolvePullSecret | both -env / both literal | all four halves resolve: the -env pair reads each variable, and a literal pair needs no environment access at all |
| TestResolvePullSecret | unset user-env | an unset user variable fails naming that variable, not only the password one |
| TestResolvePullSecret | variable unset / no environment access | both fail loudly, naming the variable |
| TestResolvePullSecret | no image block / partial image block | the guard that exists so a caller who skipped validate gets an error rather than a nil dereference |
| TestGenerateKubernetesImagePull | no block / reference / create | the wiring from config to rendered manifest: nothing, an imagePullSecrets entry alone, or the entry plus the dockerconfigjson Secret -- the integration point TestResolvePullSecret skips |
| TestGenerateKubernetesImagePull | payload and leak check | the rendered payload decodes to the real account, and the registry password appears nowhere else in the manifest |
| TestGenerateKubernetesImagePull | variable unset | create fails the generate with an issue naming the variable |

## internal/libs

Resolve the Maven dependency closure for the IBM MQ / syslog jar sets and download the jars to a local directory -- version resolution (including a pinned `--version`), image-aware omission against a connector image's jar list, and the safe download/redirect machinery.

Tests: [libs_test.go](../internal/libs/libs_test.go), [maven_test.go](../internal/libs/maven_test.go), [image_test.go](../internal/libs/image_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestDownloadURLHappyPath | - | a single `--url` download writes the jar under its basename with no Failed/Skipped |
| TestDownloadRejectsNonHTTPSURL | - | a plain http:// `--url` errors before any request or directory creation, Report stays zero-valued |
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
| TestDownloadEmbeddedDefaultListLoadFailureIsSystemic | - | a corrupted embedded default omit list (a line exceeding the scanner's token-size ceiling) is a systemic error naming "embedded default omit list", zero Report, nothing written |
| TestDownloadReadsDeployedImageFromEnv | default present / no env.yaml / explicit -e unreadable / defaulted malformed / no image block | the advisory config read: the image reaches libs.Input, an absent default is silent and still downloads, and only a file the operator named is systemic |
| TestDownloadReportsOmitListProvenance | - | Report.OmitListProvenance names the `--omit-lib-file` path used; a line that fails to split is skipped silently, and a rejected entry no closure artifact asks about produces *no* warning -- the noise fix |
| TestDownloadWarnsWhenRejectedEntryAffectedThisClosure | - | a rejected entry naming an artifact the closure does resolve warns, names that artifact, and the jar is downloaded rather than omitted |
| TestDownloadVersionPinsSeed | - | `--version` pins the seed to that release, and the resulting closure/filename reflect the pinned version, not latest stable |
| TestDownloadVersionRejectsPathEscape | - | a `--version` value containing a path escape is rejected before any network access |
| TestDownloadSetPathAlwaysResolvesEvenWhenFilesExist | - | unlike `--url` (TestDownloadSkipsExistingWithoutForce), the mq/syslog Set path always attempts Maven resolution first -- even with every plausible target file already on disk -- because it cannot know the target filenames without resolving the closure first |
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
| TestImageMismatchNote | silent: none / captured tag / newer tag / the floor itself / past every capture / registry mirror. warns: below the floor / different image / digest / no tag | the embedded list describes a *range*, so any release at or above `EmbeddedListMinVersion` is silent -- a differing tag is not itself a mismatch; below the floor, a different image, or a reference with no comparable tag warns and names both the reference and what it was judged against |
| TestDownloadImageMismatchReported | uncovered (2.9.0) / covered (2.13.0, 2.14.1) / none / --omit-lib-file | the check end to end: only an image the embedded list cannot speak for warns, both ends of the covered range stay silent, and a named omit list suppresses it -- the operator declared that list, as with an explicit `--url`. Every suppression case uses the uncovered reference, so it cannot pass vacuously |
| TestEmbeddedOmitListFullyParses | - | every line of the shipped image list either is not a jar reference or parses to an orderable version, so a future capture that reintroduces an unparseable shape fails the build instead of printing warnings on every run |
| TestSplitJarBasenameRejectsNonJar | - | a non-.jar name is rejected outright |
| TestLoadImageLibsSkipsCommentsAndBlankLines | - | a `#`-commented header line and blank/whitespace-only lines are skipped, leaving only the real jar entry |
| TestLoadImageLibsSkipsUnsplittableLineWithoutFailing | - | a line that does not split into artifact+version (e.g. jrt-fs.jar) is skipped rather than failing the whole load |
| TestLoadImageLibsToleratesSurroundingWhitespace | - | leading/trailing whitespace and a trailing \r around a jar name do not stop it from loading |
| TestLoadImageLibsBadPathIsError | - | a nonexistent `--omit-lib-file` path is an error |
| TestLoadImageLibsEmbeddedDefault | - | an empty path loads the binary's embedded default (captured from solace/solace-pubsub-connector-ibmmq:2.13.0), which carries the BC/jakarta.jms-api/json/logstash-logback-encoder versions and deliberately omits com.ibm.mq.jakarta.client (the licensing carve-out) |
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
| TestParsePodsComponentsComeFromWhatThePodReferences | - | components are read from the pod spec (volumes, envFrom, imagePullSecrets) rather than from env.yaml, an emptyDir is not a component, a mounted volume carries its path, and parsing claims no status -- that is a separate probe |
| TestParsePodsSingleObjectDocument | - | `get pod <name> -o json` answers the object itself with no items array; both shapes parse, since status uses both |
| TestParsePodsPendingAndTerminatedStates | pending, no container status / waiting on an image pull / terminated | a pod with no container status yet still reports a state from its phase, and each waiting/terminated reason reaches the report |
| TestParsePodsReadinessIsOnlyAVerdictWhenAProbeExists | - | without a readiness probe the column reads n/a, since kubernetes would otherwise report ready as soon as the container runs |
| TestParsePodsPicksTheConnectorContainer | - | a sidecar's restarts and limits are never reported as the connector's |
| TestParsePodsSeveralContainersNoneNamedConnector | - | no container-level fact is claimed when the connector cannot be identified; the pod phase still gives a state |
| TestParsePodsImageFilterIsWhatAllSearchesBy | - | the `--all` image filter keeps only connector pods, and an unfiltered parse keeps everything -- the filter is `--all`'s, not the parser's opinion |
| TestParsePodsErrorsAndSkips | empty / undecodable / one bad item in a list | an unreadable response is an error, but one unreadable item is skipped so the rest of the report survives |
| TestParseDeploymentAndService | - | replica counts, `replicas` omitted defaulting to the API's 1, service ports with the omitted protocol defaulting to TCP, a service merged without losing the deployment counts, a service-only workload, and an undecodable service |
| TestObjectExists | secret / bound claim / pending claim / not a document / empty | a live object reports "present", a volume claim reports its own phase (the only status here that can be bad while the object exists) |
| TestApplyTop | - | the connector's row wins over a sidecar's, a percentage appears only where a limit was read, and a pod the metrics API said nothing about keeps no usage |
| TestParseInspectDocker | - | docker's leading slash stripped from the name, the compose project read off the container's own label in the same call, the configured image reference rather than the local id, the nanocpu/memory ceilings, the age, and mounts/networks as attached components |
| TestParseInspectStatesAndHealthSpellings | exited / oom killed / restarting / paused / created / podman stopped cleanly / unknown status / podman Healthcheck key / no healthcheck | every engine status normalised, a clean stop reporting no exit code (a zero exit adds nothing to the state), both spellings of podman's healthcheck block, and n/a where no healthcheck is defined -- the usual case, since the generated compose and quadlet artifacts declare none |
| TestParseInspectFilterAndErrors | - | the `--all` image filter, an empty response, and an undecodable one |
| TestParseImageDigest | - | the first RepoDigest is the registry digest; an image never pushed has none, which is not an error |
| TestApplyStats | - | the engine's own percentages are taken as given (docker's memory string already carries both sides), and a container with no sample keeps no usage |
| TestEngineNamesByImage | - | `--all` discovery keeps the containers whose image matches, skipping malformed rows |
| TestParseApplication | - | every line of the script's report: leader election, health, health components, uptime, version, java, config, raw heap bytes rendered to a percentage, numerically-ordered workflows, and the script's own stderr note arriving on the same combined stream |
| TestParseApplicationKeepsWhatItDoesNotRecognise | - | an unknown line is kept as a note, so an instance carrying a newer script still reports everything it printed |
| TestParseApplicationHealthDetailAndBareHeap | - | health-detail is not swallowed by the health prefix that starts the same way, and an unbounded heap reports no maximum and no percentage |
| TestParseApplicationCRLFAndBlankLines | - | CRLF output and blank lines parse identically |
| TestParseApplicationIndentedLineWithNoBlockIsANote | - | an indented line with no block header above it is a note, never a workflow |
| TestRenderContainerViewBasic | - | the section banner, the column set, and the workload summary; the basic level carries neither the NODE column nor any detail block, and kubernetes reports READY rather than a HEALTH column |
| TestRenderContainerViewDetails | - | the details block: digest, resource lines, the components table, and the image-expected line whose presence is itself the finding; an instance with no sample carries no resource lines at all |
| TestRenderContainerViewDockerUsesHealthColumn | - | docker reports the engine's healthcheck verdict where kubernetes reports readiness, and has no NODE column |
| TestRenderContainerViewAllNamespacesLeadsWithNamespace | - | instances spanning namespaces cannot share one banner, so each row leads with its own; one shared namespace rides in the banner instead |
| TestRenderApplicationViewBasicAndDetails | - | the unchanged instance banner, the aligned basic lines, right-aligned workflow ids, enrichment only at the details level, and no container table in this view |
| TestRenderFailedInstanceKeepsItsBlock | - | an instance whose script could not run still gets a banner with the failure as a body line, and the container table that explains it comes first |
| TestRenderNotesAndScriptNotesShareOneIdiom | - | a note the CLI made and a note the script made read the same, and run-level notes come after the facts they qualify |
| TestRenderEmptyReport | - | a report with no instances renders nothing |
| TestRenderInstanceWithNoContainerFactsStillReportsItsApplication | - | a failed engine query leaves no table but does not cost the application half |
| TestJSONIsTheSameModelTheTablesRender | - | the document round-trips, carries schemaVersion and the field spellings a consumer keys off (the compatibility contract), and omits an unset field entirely rather than emitting null |
| TestJSONEmptyRunIsAnEmptyList | - | an empty run is `[]`, so a consumer can iterate without a nil check |

## cmd/solmq-conn-util

The CLI shell -- flag parsing, the exit-code contract, the generate/validate/examples/auto-complete commands, verb aliases, and the deploy/remove/status/logs/cli seams for all three engines. The completion tests also gate the four generated shell scripts against the command model, and the doc tests gate the two generated markdown references against it.

Tests: [main_test.go](../cmd/solmq-conn-util/main_test.go), [commands_doc_test.go](../cmd/solmq-conn-util/commands_doc_test.go), [abbreviation_doc_test.go](../cmd/solmq-conn-util/abbreviation_doc_test.go), [completion_test.go](../cmd/solmq-conn-util/completion_test.go)

| Test | Case | Verifies |
|------|------|----------|
| TestDispatchHandlersMatchModel | verbs / generate targets / deploy platforms / completion shells | the dispatch handler sets and cliVerbs agree in *both* directions, so a command added to one cannot drift from the other |
| TestPlatformMapsCoverThreeNames | - | platformNames, platformGenerators and actTargets agree in both directions -- a modeled platform with no handler, or a handler with no modeled entry, fails |
| TestExitCodeContract | nil args | run(nil) returns exit code 2 |
| TestExitCodeContract | unknown command | run([bogus]) returns exit code 2 |
| TestExitCodeContract | help short -h | run([-h]) returns exit code 0 |
| TestExitCodeContract | help long --help | run([`--help`]) returns exit code 0 |
| TestExitCodeContract | help word | run([help]) returns exit code 0 |
| TestExitCodeContract | 5 requested-help spellings | `status -h`, `deploy --help`, `-h` among a verb's flags (routed as flag.ErrHelp), `help status`, and `help sts` each exit 0 -- requested help is never an error, wherever it is asked |
| TestExitCodeContract | help with an unknown command | run([help, bogus]) returns exit code 2 |
| TestExitCodeContract | unknown flag -nope | run returns exit code 2 for unrecognized flag |
| TestExitCodeContract | missing env file | run returns exit code 1 when env file does not exist |
| TestExitCodeContract | invalid spec | run returns exit code 1 for structurally invalid workflow |
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
| TestGenerateConfigWorkflowCapExceeded | - | a folder over validate.MaxWorkflows (20) is a fatal error (exit 1) naming the count and the cap, and writes no `-o` output file |
| TestGenerateConfigEmitWriteError | - | emit to path with missing parent dir returns exit code 1 |
| TestLoadEnvWorkflowsDirRelativeToEnvFile | - | workflows.dir resolved relative to env file not cwd; exit 0 and stdout has spring: |
| TestLoadEnvExcludesEnvFileFromWorkflowSet | - | env.yaml excluded from its own workflow scan; exit code 0 |
| TestDeployKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then apply) with argv [kubectl apply -f -], stdin contains kind: Deployment |
| TestRemoveKubernetesSeamHappyPath | - | exit 0, 2 runner calls (preflight then delete) with argv [kubectl delete -f -] |
| TestDeployKubernetesSeamRejectsUnsafeCommand | - | unsafe kubernetes.command yields exit 1 and zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | path value | `--allow-command` /usr/bin/sudo exits 2, zero runner calls |
| TestAllowCommandFlagBadValueExitsUsageError | unsafe character | `--allow-command` sudo;rm exits 2, zero runner calls |
| TestAllowCommandFlagRejectedOnGenerateAndValidate | generate config / validate | `--allow-command` is undefined on generate/validate; exit 2 as an unknown flag |
| TestAllowCommandFlagRepeatableThreadsToRunner | - | "sudo podman" rejects with zero runner calls without the flag; repeating `--allow-command` sudo twice threads through to preflight (argv [sudo podman info]) and to the podman secret calls |
| TestDeployKubernetesPreflightFailureStopsBeforeApply | - | a failing kubernetes preflight (auth can-i argv incl. `--namespace`) stops with exit 1 and exactly 1 runner call |
| TestDeployDockerPreflightFailureStopsBeforeWrite | - | a failing docker preflight (argv [docker info]) stops before the compose file is written, exit 1, exactly 1 runner call |
| TestDeployPodmanPreflightFailureStopsBeforeWrite | - | a failing podman preflight (argv [podman info]) stops before the unit/app-yaml files are written, exit 1, exactly 1 runner call |
| TestValidateOKAndErrors | valid spec | validate exits 0 |
| TestValidateOKAndErrors | invalid spec | validate exits 1 |
| TestExamplesWriteSkipForceThenGenerate | first write | examples command exits 0 creating env.yaml |
| TestExamplesWriteSkipForceThenGenerate | re-run without -f | exits 0 and skips existing file, content stays 'touched' |
| TestExamplesWriteSkipForceThenGenerate | re-run with -f before dir | exits 0 and overwrites existing file |
| TestExamplesWriteSkipForceThenGenerate | generate on shipped examples | generate config on generated env.yaml exits 0 |
| TestExamplesDefaultDir | - | examples with no dir arg exits 0 and creates ./examples/workflow-0.yaml |
| TestDownloadMissingAndUnknownWordsRejected | missing target / unknown target / missing set / unknown set | each is a usage error (exit 2) naming the offending word, with downloadFn and the runner both left uncalled |
| TestDownloadDirDefaultAndPositionalOverride | default dir / explicit dir | the trailing [dir] positional defaults to ./libs and threads an explicit value through to Input.Dir unchanged |
| TestDownloadSetReachesInput | mq / syslog | both modeled sets thread through to libs.Input.Set unchanged |
| TestDownloadURLFlagRepeatable | - | repeated `--url` collects every occurrence, in order, into Input.URLs |
| TestDownloadJMSFlagIsGone | mq / syslog, either value | `--jms` is an unknown flag: exit 2 and downloadFn never runs, so a script still passing it fails loudly instead of being silently ignored |
| TestDownloadForceFlagReachesInput | default false / -f short / --force long | both -f and `--force` spellings reach Input.Force, defaulting to false |
| TestDownloadVersionFlagReachesInput | default empty / explicit pin | `--version` defaults to "" (latest stable) and an explicit pin reaches Input.Version unchanged |
| TestDownloadOmitLibFileFlagReachesInput | default empty / explicit path | `--omit-lib-file` defaults to "" (the embedded list) and an explicit path reaches Input.OmitLibFile unchanged |
| TestDownloadIncludeProvidedFlagReachesInput | default false / --include-provided | `--include-provided` defaults to false and reaches Input.IncludeProvided true when given |
| TestDownloadReportExitCode | clean write / skip only / omitted only / partial failure / total failure / systemic error | a clean write, a skip-only run, and an omitted-only run all exit 0; a non-empty Report.Failed -- whether partial or total -- and a systemic downloadFn error both exit 1 identically, pinning the deliberate choice not to mint a distinct code for partial vs total failure |
| TestDownloadReportPrintsWrittenSkippedFailedAndFallback | - | reportDownload prints "wrote:"/"exists (use -f to overwrite):"/"failed:" lines and a Fallback note labelled "guessed version:", mirroring runExamples' line shapes |
| TestDownloadReportPrintsOmittedBlockDistinctFromFallback | - | Report.Omitted prints its own "omitted:"-prefixed lines, distinct from "failed:" and "guessed version:", and the counts footer names the omitted count; exit stays 0 with no Failed entries |
| TestDownloadReportExitZeroWhenEverythingOmitted | - | a Report where every artifact was omitted (nothing written, skipped, or failed) exits 0, and the counts footer reads "0 written, 0 skipped, 4 omitted, 0 failed" |
| TestDownloadReportNextHint | no omissions / with omissions | the "next:" hint always points at wiring libs.dir/libs: config key, gaining an extra clause about the omitted jars already being on the image only when Report.Omitted is non-empty |
| TestDownloadReportPrintsOmitListProvenance | built in / explicit file | the "omit list:" line names Report.OmitListProvenance, annotated "(built in)" only when `--omit-lib-file` was left empty; an explicit path prints bare, and each Report.OmitListWarnings entry gets its own "omit list warning:" line |
| TestDownloadSetMapMatchesModel | - | downloadSets (main.go's set-name dispatch table), the model's download/jar Sets, and internal/libs.SetNames() all name exactly the same sets |
| TestGenerateKubernetesStdout | - | exit 0 and stdout contains kind: Deployment |
| TestGenerateDockerToFile | - | exit 0 and compose file opens with the defaulted project line, then contains services: and image: img:1 |
| TestGeneratePodmanQuadletStdout | - | exit 0 and stdout contains unit banner '# === solmq-conn-util.container ===' |
| TestGeneratePodmanVolumeSourcesAreAbsolute | -e env.yaml from the file's own dir | the truststore and libs Volume= sources are absolute even when -e is spelled relatively. A relative source is not a near-miss in a quadlet: systemd starts the unit with no useful cwd, and podman reads a source with no ./ or / prefix as a named volume, so `Volume=libs:...` would silently mount an empty volume over the jars |
| TestDeployDockerSeamWritesComposeAndRuns | - | exit 0, compose file written, 2 runner calls (preflight then up) argv [docker compose -f <compose> up -d] -- no -p, since the project is declared by the file's own name: key |
| TestDeployDockerSeamComposeFileSurvivesFailedRun | - | preflight succeeds but the real `up` call fails; compose file still exists on disk afterward, exit 1 |
| TestDeployDockerSeamChildEnvCarriesCredentials | - | preflight call carries no env; the real `up` call (index 1) carries the resolved literal and -env credentials as STABLE=value pairs |
| TestRemoveDockerSeam | - | exit 0, 2 runner calls (preflight then down) argv [docker compose -f <compose> down] |
| TestDeployPodmanSeamWritesUnitsAndStarts | - | exit 0, a leading podman info preflight call, then app yaml and container unit written to quadlet dir, systemctl daemon-reload then start calls |
| TestDeployPodmanMissingBaseDirFailsBeforeAnyWrite | - | a podman section with no base-dir exits non-zero, makes ZERO runner calls (not even the preflight probe) and writes no file. Being rejected is not enough on its own: the steps after validation have side effects outside the process -- secrets in podman's store, files on disk, systemctl -- so a late failure would leave a half-built deployment |
| TestDeployPodmanSplitsBaseDirFromQuadletDir | distinct dirs | the mounted application.yml and status script go to podman.base-dir (created on demand) and NOT to the quadlet dir; only the .container unit goes to the quadlet dir -- which the spec cannot name, so the test redirects HOME to reach it -- and the unit's Volume= names the base-dir path |
| TestRemovePodmanSeamStopsRemovesReloads | - | exit 0, a leading podman info preflight call, then systemctl stop then daemon-reload calls, unit and app yaml files removed |
| TestPlatformFlagHitOverridesInference | - | an explicit `--platform` is used even when another section is also present in env.yaml |
| TestPlatformFlagMissingSectionIsLoudError | - | a `--platform` value with no matching section fails loud, naming both the requested and the present sections, before the runner is invoked |
| TestPlatformAliasesResolveToCanonical | kube / dk / pm | each short `--platform` spelling reaches the same platform binary as its canonical name |
| TestPlatformAliasMissingSectionNamesCanonicalSection | - | an alias is resolved before the section check, so the error names the `kubernetes:` section to add rather than echoing `kube` |
| TestPlatformUnknownValueListsEverySpelling | - | a bogus value (k8s) is rejected with every accepted spelling listed, canonical and short |
| TestPlatformSpellingsAreDeterministic | - | platformSpellings is built from an ordered slice, not map iteration, so the rejection message cannot vary between runs; canonical names lead |
| TestPlatformAliasesCoverEveryPlatformExactlyOnce | - | every alias maps to a real platform, no alias is declared twice or collides with a canonical name, and the lookup map matches the declared list |
| TestPlatformSingleSectionInferred | - | with no `--platform` and exactly one section present, that section is used and echoed to stderr |
| TestPlatformMenuOnMultipleSections | - | with no `--platform` and more than one section, the interactive menu (via the injected promptLine seam) picks the platform |
| TestPlatformMenuNonTTYRefusesWithPlatformHint | - | the menu refuses to block when stdin is not a TTY, failing with an error naming `--platform` instead of hanging |
| TestPlatformZeroSectionsIsLoudError | - | with no `--platform` and no section present at all, the error names all three section keys |
| TestOldPositionalFormsRejectedWithPlatformHint | deploy kubernetes / remove docker / generate podman | passing the platform as a second positional argument is a usage error (exit 2) that points at `--platform`, not resolved as a target |
| TestStatusTargetWordIsRequired | - | a bare `status` prints the target words and the verb's own help page, exits 2, and runs nothing, since neither view is a safe default; the short spellings (cnt, app) are deliberately absent, since aliases are documented only in the markdown docs |
| TestStatusUnknownAndExtraTargetWords | unknown word / a second word | each is a usage error (exit 2) naming the problem, with nothing run |
| TestStatusTargetsMatchModel | - | the drift gate between the modeled target words, the constants the views switch on, and statusTargetArgBracket (which cannot be built from the model, since cliVerbs' own initialiser uses it) |
| TestStatusTargetAliasesResolve | cnt / app / container / all / unknown | resolveTarget maps each alias to its canonical word and passes an unknown one through; `sts cnt` really drives the container view, which costs one preflight and one get |
| TestStatusRejectsImpossibleFlagCombinations | unknown --output / json with watch / --all with --pod / --all with --container / --install, --user and --management-port on the container view | every combination that cannot mean anything is refused (exit 2) before a single query runs, rather than being silently ignored |
| TestWatchFlagAcceptsBareAndInterval | bare / interval / off / 3 rejected values | the flag is boolean in every documented sense (IsBoolFlag) but also takes the deliberately undocumented `-w=<seconds>` form, bounded |
| TestStatusContainerViewReadsEngineFactsWithoutExecing | - | one read-only `get pods -o json` answers discovery and the whole table together (2 calls in all), nothing is exec'd into, and every column reaches the output |
| TestStatusContainerDetailsSamplesAndChecksComponents | - | `--details` adds one sampling call for the run and one presence check per distinct referenced object (deduplicated across pods), plus the NODE column, digest and resource lines |
| TestStatusContainerDetailsWithoutMetricsServerDegradesToANote | - | a cluster with no metrics API costs the resource lines and nothing else: a note naming what to install, the table still printed, exit 0 |
| TestStatusDockerContainerViewIsOneInspect | - | one inspect answers every docker target and carries the compose project too; docker reports HEALTH where kubernetes reports READY |
| TestStatusPodmanRestartCountComesFromSystemd | - | the quadlet truth: the count in the table comes from `systemctl show ... NRestarts`, not from podman's own counter |
| TestStatusPodmanRestartCountFallsBackWhenSystemdCannotAnswer | - | a container systemd knows nothing about keeps the container's own counter, and nothing fails |
| TestStatusAllSearchesByImage | kubernetes searches every namespace / docker lists then inspects the matches | `--all` finds instances by image reference: `--all-namespaces` plus a client-side filter on kubernetes (with a NAMESPACE column), `ps --all` then an inspect of only the matches on docker |
| TestStatusAllWithNoMatchIsActionable | - | an empty search names the image it looked for, since there is no env.yaml in play to point at |
| TestStatusApplicationViewRunsTheScriptAndRendersItsFacts | - | the exact application block: the unchanged banner, values aligned in one column, right-aligned workflow ids, and no container table |
| TestStatusApplicationDetailsAddsTheEnrichmentLines | - | one script run, two levels of report: `--details` renders uptime/version/java/config/heap (raw bytes rendered to 412Mi of 1Gi) and the health components, and the basic level renders none of them |
| TestStatusFailedScriptRunStillGetsItsOwnBlock | - | an instance whose script could not run keeps its banner with the failure as a body line, the container table above it explains why, a reachable instance in the same run still reports, and the exit code is 1 |
| TestStatusInstallPaths | --install installs without asking / prompt answered yes installs / prompt declined skips the instance and exits 1 | the probe/install/run dance, with the declined case reporting the reason in the instance's own block rather than on stderr |
| TestStatusInstallPromptNonTTYRefusesWithInstallHint | - | the install confirmation refuses to block when stdin is not a TTY, pointing at `--install` and installing/running nothing |
| TestStatusStandbyIsAnAnswerNotAFailure | - | standby prints like any other answer (the script always exits 0) and the run still exits 0 |
| TestStatusJSONOutputIsOneDocument | - | `--output` json emits one parseable document carrying schemaVersion and both halves of each instance |
| TestStatusDockerProjectMismatchIsReported | a different project is a note | the container's compose-project label disagrees with docker.project-name, so a status: note names both projects and the way out; exit stays 0 |
| TestStatusDockerProjectMismatchIsReported | the configured project says nothing at all | a matching label is silent |
| TestStatusDockerProjectMismatchIsReported | no compose label is not a mismatch | a container compose never created carries no label, which is not drift and must stay silent |
| TestStatusImageMismatchIsReportedAtBothLevels | basic reports it as a note / details reports it per instance / a matching image says nothing at all | the failed-rollout finding surfaces at both levels -- a run-level note where the per-instance detail block is not printed, the image-expected line where it is, and nothing at all when the running image is the configured one |
| TestStatusDockerDetailsAddsDigestAndStats | - | on docker/podman the digest lives on the image, so `--details` costs an `image inspect` plus the `stats --no-stream` sample (4 calls in all), and both reach the report |
| TestStatusRejectsUnsafeUserBeforeAnyExec | 4 names | a `--user` carrying '/', '$', a space or a quote is rejected via validate.SafeActuatorUser before any exec, since the name reaches a sed address in the script |
| TestStatusTargetValidationRejectsBadPodAndNamespace | bad pod name / bad namespace | an unsafe `--pod` or `--namespace` value is rejected via validate.SafeToken before any exec |
| TestStatusManagementPortBounds | -1 / 65536 | an out-of-range `--management-port` is rejected before any exec |
| TestStatusNoPodsFoundNamesTheSelector | - | discovery with nothing matching names the selector, the namespace, and `--pod` -- the things an operator would fix |
| TestVersionOutputShape | - | `version` prints `solmq-conn-util <version> <go version> <GOOS>/<GOARCH>`, exit 0; the package-level version var defaults to "dev" in an un-injected test build |
| TestAbsPath | absolute input | absPath returns input unchanged when already absolute |
| TestAbsPath | relative input | absPath joins relative path onto base dir |
| TestCommandsDocInSync | - | docs/commands.md equals what the command model renders; -update rewrites it instead of asserting |
| TestCommandsModelMatchesUsage | - | the summary page is one line per modeled command carrying its description, points at `help <command>`, shows no alias anywhere (md-only, by decision), and no line exceeds the 100-column budget the page is designed never to wrap in |
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

### logs

The `logs` verb shares status's platform resolution and instance discovery (instances.go), so these cases concentrate on what is its own: the per-platform argv, the combinations it refuses, and the fact that every operator-supplied name is rejected before a process starts. Like the status cases they pin argv and call count, since discovery is one query for every instance.

| Test | Case | Verifies |
|------|------|----------|
| TestLogsKubernetesArgvShape | - | a named pod costs no discovery call at all; the log is read with the namespace and the connector container both named, so a pod with a sidecar cannot have the wrong half read |
| TestLogsDockerArgvShape | - | docker takes its options first and the container name last, and `--since` reaches the argv in the canonical duration form (10m -> 10m0s) rather than as typed; a single instance prints no heading so the output pipes cleanly |
| TestLogsPodmanReadsTheContainerNotTheJournal | - | the quadlet path needs no new binary: `podman logs` reads the container, and neither journalctl nor systemctl appears in any argv |
| TestLogsTailAllAndZero | default / all / zero | `--tail` all is the default and adds no flag, while an explicit 0 is a real request and does |
| TestLogsRejectsImpossibleFlagCombinations | follow with previous / follow with all / all with pod / all with container | each refusal exits 2, names both flags, and runs nothing |
| TestLogsPreviousIsKubernetesOnly | docker / podman | the one refusal that needs the resolved platform: `--previous` is refused by name, before the preflight probe, so a flag that cannot work does not first make the operator wait on a daemon |
| TestLogsPreviousReachesTheKubernetesArgv | - | where the concept exists, `--previous` arrives as kubectl's own -p |
| TestLogsFollowReadsTheOneInstance | - | the accepted case: -f reaches the argv and a clean end is exit 0 |
| TestLogsRejectsUnsafeNamesBeforeAnyCall | pod / container / namespace | an unsafe name exits 1 saying why, with zero calls -- the preflight probe included, so a rejected name is never even observed by the platform |
| TestLogsSinceAndTailAreValidatedAtParse | since not a duration / since not positive / since with a metacharacter / tail not a number / tail above the ceiling / tail negative | the two flags carrying a value into an argv are validated at parse, exit 2, nothing runs |
| TestLogsUnexpectedPositionalArgument | - | logs has no target word, so a bare word exits 2 naming `--pod`, `--container` rather than being guessed at |
| TestLogsPlatformMenuWhenSeveralSectionsArePresent | - | an env.yaml with two platform sections cannot resolve itself, so the menu decides; the answer picks the binary and the deployment selector/namespace discovery uses |
| TestLogsPlatformFlagSkipsTheMenu | - | `--platform` is the first step of the resolution order, so it wins before promptLine is consulted, and the instance still comes from that section |
| TestLogsWithoutEnvFileNeedsAnExplicitPlatform | - | the explicit-target exception: instance plus `--platform` needs no env.yaml, while without `--platform` the missing file is reported by name and nothing runs |
| TestLogsNoInstancesFoundNamesTheFix | - | an empty discovery result carries the selector and namespace it used, plus `--pod`, since those are what the operator would change |
| TestLogsNeedsAStreamingRunner | - | logs reads through runner.Streamer and a Runner without it fails loudly, rather than silently falling back to a buffered read that would merge diagnostics into the log |
| TestLogsWithoutASectionToDiscoverFrom | kubernetes / docker / podman | the discovery branch with nothing to work from, reached by naming an instance with the other platform's flag (`--container` on kubernetes, `--pod` on docker/podman); the error names the missing section and the way forward, and must *not* name `--all`, which logs does not have |
| TestLogsPlatformWithNoSectionAtAllIsLoud | - | an env.yaml that parses but describes no platform cannot answer `--platform` either, and says so naming all three sections before discovery is attempted |
| TestLogsNamedPodIsNotPreChecked | - | a name is taken at its word: no `get pods <name>` precedes the read, and a pod that is not there is reported by the platform on the read itself |
| TestLogsUsesTheSectionCommandOverride | - | logs reaches for the binary env.yaml names (oc, not kubectl) on every call, through the shared instanceCommand resolution |
| TestLogsCommandFlagOverridesTheSection | - | `--command` wins over the section, and a binary outside the per-platform allowlist is refused by name before anything runs |
| TestRemoveKubernetesSeamPromptsBeforeTearingDown | y / yes / YES / Y | an accepted confirmation runs the full teardown: preflight, delete, occupancy probe and, since the same answer approves the namespace question, the namespace delete |
| TestRemoveDeclinedTouchesNothing | n / no / blank / whitespace / anything else | a declined teardown exits 0 having run nothing at all, not even the read-only preflight probe, and says "cancelled"; a blank line declines, so the safe answer is the one Enter gives |
| TestRemoveNoPromptSkipsThePromptEntirely | - | `--no-prompt` never reaches promptLine at all (asserted by a seam that fails the test if called), and covers both questions: the empty namespace is removed without asking |
| TestRemoveNonTTYFailsFastNamingTheFlag | - | a non-TTY refusal exits 1 naming `--no-prompt` rather than blocking on a read that will never return, the same shape the platform menu and the status install confirmation use; nothing runs |
| TestRemovePromptNamesWhatItWillDestroy | kubernetes / docker / podman | the question carries the identifier that tells one target from another -- deployment name and namespace on kubernetes, the container on docker, the container and its .service unit on podman |
| TestDeployNeverPromptsAndRejectsNoPrompt | - | deploy is additive and re-runnable so it never prompts, and `--no-prompt` there is an unknown flag (exit 2) rather than a flag that silently does nothing |
| TestRemoveTargetDescriptions | kube with/without namespace, section missing, env nil, docker, podman | removeTarget names the identifier that tells one target from another for every platform, including the defensive branches the end-to-end tests cannot reach (a section-less env, and a namespace-less deployment that validate rejects moments later) |
| TestLogsAllIsNotALogsFlag | - | `--all` searches by image and returns many, the one thing logs cannot do, so it is an unknown flag here rather than one accepted and then refused |
| TestLogsPickerListsPasteableCommands | - | several discovered and none named: nothing is read, and the matching instances are listed on stdout in sorted order as commands carrying the flags already typed plus the resolved `--platform`, with the index range spelled out; exit 0 |
| TestLogsOneDiscoveredInstanceIsJustRead | - | the picker is for ambiguity only, so a single match is read directly with no listing |
| TestLogsIndexSelectsFromTheSortedList | 0 / 1 / 2 | the pod list arrives unsorted and index 0/1/2 still selects pod-a/pod-b/pod-c, so a dropped sort fails the test |
| TestLogsNameWinsOverIndex | - | a pod genuinely named 0 is reached by name; the index reading applies only when nothing is actually named that |
| TestLogsIndexOutOfRangeNamesTheRange | - | an index past the end says how many matched and what the valid range is, and reads nothing |
| TestLogsIndexWithNothingToIndexInto | - | an index given where discovery has no list says that is the problem, rather than reporting a missing pod named 0 |
| TestStatusAcceptsAnIndexToo | - | status resolves `--pod` 1 to a real name before querying, so the two verbs share one vocabulary and no index ever reaches an argv |
| TestStatusKeepsPodRepeatable | - | the one-entry limit is a logs rule: status reports many instances by design and still accepts `--pod` twice |
| TestStatusAllSkipsAnUnsafeNameOutLoud | - | under `--all` the names come from ps, so one that could not go into an argv is skipped with a note naming it, never reaches any argv, and the instances around it are still reported |
| TestStatusAllSortsRowsAlphabetically | - | docker ps returns creation order, so the rows are sorted by name; without it an index would select a different instance than the row an operator counted to |
| TestIsIndex | 0 / 12 / empty / 0-abc / abc / -1 / 1.0 / leading space | only a bare run of digits is an index, so an instance named 0-abc is still reachable by name |
| TestResolveOneIndexRules | plain name / name that is digits / in-range index / unknown name / past the end | an exact name always wins over the index reading, a digit run in range selects positionally, an unknown name passes through for discovery to judge, and one past the end names how many matched and the valid range |
| TestSortIsByNameThenNamespace | - | the order an index selects into, for both instanceRef and statusreport.Instance: by name, tie-broken on namespace for a cross-namespace status `--all` where two deployments run pods of the same name |
| TestNamingHintIsVerbAware | status/logs/cli x kubernetes/docker/podman | a hint never names a flag the verb it came from does not accept: status offers `--all`, logs and cli have no such flag |
| TestLogsPickerCarriesEveryFlagBack | - | every flag already typed comes back in the suggested command (`--platform`, -e, `--namespace`, `--command`, `--allow-command`, `--previous`, `--timestamps`, `--tail`, `--since` in canonical duration form), or pasting a line would silently drop what was asked for |

### cli

The `cli` verb reaches its instance through the same resolution `status` and `logs` use (instances.go), so these cases concentrate on what is its own: the attaching seam and its refusal when a Runner lacks it, the per-platform session argv, the split at `--` that keeps the in-container command out of the flag parser, the non-TTY refusal, and the exit status coming back from the session rather than from this tool. Every refusal asserts zero calls: a mistake in the invocation must cost no process.

| Test | Case | Verifies |
|------|------|----------|
| TestCliNeedsAnAttachingRunner | - | a Runner that cannot hand over the terminal fails loudly before anything runs, rather than degrading to a session with no prompt |
| TestCliRefusesAShellWithoutATerminal | - | the same non-TTY contract the platform menu and the remove confirmation keep, with the one-shot form named as the next step |
| TestCliKubernetesArgvShape | - | `-i -t`, the namespace, `-c connector` and the `--` terminator, in that order; a named pod costs no discovery call, so the run is preflight plus attach |
| TestCliEngineArgvShape | docker / podman | engine flags precede the container name, and neither `-c` nor `--` is added -- both would reach the container as arguments |
| TestCliOneShotRunsTheCommandWithNoTerminal | - | everything after `--` replaces the shell, with no `-t` and no `-i` while stdin is a terminal |
| TestCliOneShotAttachesStdinWhenSomethingIsPiped | - | a stdin that is not a terminal is one something is piped into, so `-i` is added; also the case proving the one-shot form needs no terminal |
| TestCliPropagatesTheSessionExitStatus | - | `exit 3` in the session exits 3, the one departure from the 0/1/2 contract |
| TestCliSessionThatCouldNotStartIsAnError | - | a session that never began is exit 1 and names the instance, keeping it apart from a session that ran and failed |
| TestCliRejectsWhatCannotMeanAnything | bare word / separator with no command / two pods / two containers / unsafe command token / unsafe instance name | each refused before any process starts, with a message naming what to do instead; the glob case says where a shell metacharacter can be written instead |
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
| TestRemoveTeardownManifestOmitsTheNamespace | - | the bug this exists for: the manifest piped to `kubectl delete -f -` carries no Namespace, since deleting one cascades to every object inside it including workloads this tool never deployed; the separate namespace step is the only place a Namespace document is piped, and deploy still emits it in the manifest or the objects have nowhere to land |
| TestRemoveNamespaceOccupiedLeavesItAlone | - | anything living in the namespace that the release does not own is listed as kind/name and the namespace is kept, with no fourth call |
| TestRemoveNamespaceEmptyPromptsSeparately | y / n / blank | an empty namespace asks its own question, because saying yes to removing a deployment is not saying yes to removing the namespace around it; declining leaves it |
| TestRemoveNoPromptNeverRemovesAnOccupiedNamespace | - | the invariant: no flag removes a namespace still holding someone else's work. `--no-prompt` approves the questions, never a cascade -- the occupancy check runs first and an occupant of any kind ends it, silently or not |
| TestRemovePlainRunChecksTheNamespace | - | a plain remove with no flags probes the namespace and asks both questions -- teardown, then namespace |
| TestRemoveNamespaceRefusesClusterNamespaces | default / kube-system / kube-public / kube-node-lease | refused before the occupancy probe even runs, so no emptiness result can authorise deleting one |
| TestRemoveNamespaceProbeFailureLeavesItAlone | - | a failed occupancy query leaves the namespace in place and says so; a namespace is not worth deleting on a guess |
| TestRemoveFailedTeardownSkipsTheNamespace | - | a failed teardown never probes or deletes the namespace: whatever did not come down is still in there |
| TestNamespaceOccupantsRules | foreign deployment/statefulset/pvc/service; our own; owned pod; terminating; the three cluster defaults | the exclusion table -- each non-occupant would otherwise keep a namespace alive forever, the likeliest being the release deleted moments ago and still terminating |
| TestNamespaceOccupantsAreSortedAndLabelled | - | kind/name, lower-cased and sorted, so the list is stable between runs |
| TestNamespaceOccupantsRejectsUnreadableOutput | - | a parse failure is an error, never an empty list: unreadable output must not read as "empty" and authorise the delete |
| TestTeardownDropsTheNamespaceDocument | - | apply keeps the Namespace document, a teardown drops it, dropping the first document leaves no leading separator, and the ConfigMap and Deployment still render |
| TestNamespaceManifestMatchesWhatRenderEmits | - | the standalone document the namespace delete pipes is exactly the one Render emits, so the two cannot drift |
| TestOwnedNamesCoversEverythingThisReleaseCreates | created secrets/PVC / referenced ones / nil section | the safety net behind the occupancy check: everything this release creates counts as ours, an empty name never does, and objects merely referenced (existing:) are *not* ours -- the operator manages them and their presence is a real reason to keep the namespace |
| TestRemoveNamespaceWithoutANamespaceSaysSo | - | a kubernetes section with no namespace has nothing to remove, and says so rather than acting on an empty name |
| TestRemoveNamespaceRejectsAnUnsafeCommand | - | the namespace delete goes through the same binary allowlist as everything else, rather than being a second path around it |
| TestRemoveNamespaceUnreadableProbeLeavesItAlone | - | output that cannot be parsed must not read as "empty" and authorise the delete |
| TestRemoveNamespaceNonTTYFailsFastNamingTheFlag | - | the namespace question refuses a non-TTY naming `--no-prompt`, the same shape every other prompt uses; the namespace is not deleted |
| TestRemoveNamespaceProbeArgvIsOneQuery | - | one call listing every kind whose loss would matter, in the namespace being considered |

