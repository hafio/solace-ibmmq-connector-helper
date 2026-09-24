# Graph Report - solace-ibmmq-connector-helper  (2026-09-24)

## Corpus Check
- 102 files · ~322,570 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2072 nodes · 6739 edges · 108 communities (89 shown, 19 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1102 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `cfa04581`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- hasErr
- main_test.go
- golden_test.go
- gen.go
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- go_pkg_fmt
- Scan
- completion_test.go
- runner_test.go
- uuid.go
- Runner
- solmq-conn-util user guide
- ExecArgv
- Render
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- gen_extra_test.go
- podmangen_test.go
- runAction
- net/http.Response
- withProgressClock
- testing.T
- podmanRemove
- runner.go
- statusreport.go
- status.go
- CLAUDE.md
- internal/consolidate
- internal/deploy
- internal/gen
- internal/render
- internal/scan
- internal/spec
- internal/tls
- internal/validate
- application.yml (Golden)
- Side
- main.go
- solmq-conn-util -- Development Guide
- libs_test.go
- Kubernetes
- Command details
- solmq-conn-util.bash
- Model
- spec_test.go
- render.go
- instances.go
- consolidate.go
- render/render_test.go
- consolidate_test.go
- maven_test.go
- write
- go_pkg_gopkg_in_yaml_v3
- logs.go
- parse_test.go
- go_pkg_strconv
- parse.go
- Defaults
- Render
- TestDownloadSetMapMatchesModel
- go_pkg_strings
- statusreport/render.go
- render
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- Defaults
- 13. Logs: the lines behind the state
- 6. Workflow file
- gen/imagepull_test.go
- attachRunner
- podmanDeploy
- ParseEnv
- maven.go
- shell.go
- solmq-conn-util abbreviations
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- libs.go
- cmd/solmq-conn-util
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- TestParsingHelpersIgnoreAWarningOnStderr
- testcatalog_test.go
- Cmd
- go_pkg_os
- Expand
- Env
- helperProcessArgv
- namespace.go
- dispatch
- examples_test.go
- vt_windows.go
- solmq-conn-util command reference
- auto-complete
- 9. Secrets model
- status
- allowCommandValue
- Workload

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 139 edges
2. `hasErr()` - 84 edges
3. `captureStderr()` - 68 edges
4. `write()` - 65 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 48 edges
8. `captureStdout()` - 44 edges
9. `Runner` - 42 edges
10. `imageOK()` - 42 edges

## Surprising Connections (you probably didn't know these)
- `renderedCompletions()` --calls--> `render()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/statusreport/render_test.go
- `TestCompletionCoversModel()` --calls--> `contains()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/runner/runner_test.go
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go
- `TestDownloadImageMismatchReported()` --calls--> `run()`  [INFERRED]
  internal/libs/libs_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (108 total, 19 thin omitted)

### Community 0 - "Build"
Cohesion: 0.10
Nodes (34): Build(), displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+26 more)

### Community 1 - "validate.go"
Cohesion: 0.11
Nodes (50): synthWorkflows(), Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker() (+42 more)

### Community 2 - "hasErr"
Cohesion: 0.07
Nodes (108): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+100 more)

### Community 3 - "main_test.go"
Cohesion: 0.08
Nodes (54): run(), captureStdout(), containsToken(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestExamplesDefaultDir() (+46 more)

### Community 4 - "golden_test.go"
Cohesion: 0.21
Nodes (23): go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_gen, go_pkg_regexp, configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa() (+15 more)

### Community 5 - "gen.go"
Cohesion: 0.11
Nodes (25): built, File, mount, go_pkg_crypto_rand, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_consolidate, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_dockergen, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_podmangen, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_render (+17 more)

### Community 6 - "go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec"
Cohesion: 0.20
Nodes (13): go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec, breEscape(), MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted() (+5 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "go_pkg_fmt"
Cohesion: 0.21
Nodes (7): progressTrim(), go_pkg_fmt, go_pkg_sync, go_pkg_time, sync.Mutex, progress, progressMode

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion_test.go"
Cohesion: 0.05
Nodes (89): abbrevFlagByShort(), abbrevTable(), TestAbbreviationDocInSync(), renderAbbreviationDoc(), assertHelpWidth(), assertNoAliases(), normLF(), TestCommandsDocInSync() (+81 more)

### Community 13 - "runner_test.go"
Cohesion: 0.10
Nodes (31): go_pkg_runtime, os.FileMode, TestWriteMkdirError(), EngineInspectJSON(), Preflight(), ScriptInstalled(), regularFile(), TestEngineInspectJSONArgv() (+23 more)

### Community 14 - "uuid.go"
Cohesion: 0.31
Nodes (7): go_pkg_crypto_sha1, go_pkg_encoding_hex, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Runner"
Cohesion: 0.12
Nodes (21): removeNamespace(), Docker(), Runner, Kubernetes(), KubernetesListJSON(), kubeVerb(), ParseCommand(), PodmanSecretCreate() (+13 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "ExecArgv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

### Community 19 - "Render"
Cohesion: 0.13
Nodes (31): Input, Instance, composeEscape(), composeQuote(), yw, Render(), renderContentConfig(), renderHealthcheck() (+23 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.09
Nodes (55): Input, Instance, KV, go_pkg_encoding_base64, go_pkg_encoding_json, envValueQuote(), PullSecret, StoreFile (+47 more)

### Community 23 - "gen_extra_test.go"
Cohesion: 0.14
Nodes (36): KubeOpts, Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigRejectsSecretNameConflict(), TestConfigStatusPasswordRandErrorNoOutput() (+28 more)

### Community 24 - "podmangen_test.go"
Cohesion: 0.16
Nodes (26): go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_logback, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_statusscript, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_yamlwriter, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds() (+18 more)

### Community 25 - "runAction"
Cohesion: 0.18
Nodes (30): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), downloadDeployedImage(), envFlag(), flagExit() (+22 more)

### Community 27 - "withProgressClock"
Cohesion: 0.18
Nodes (16): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+8 more)

### Community 28 - "testing.T"
Cohesion: 0.09
Nodes (39): TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestCliEngineArgvShape(), TestConfiguredInstanceName(), TestExitCodeContract(), TestInstanceCommandResolution(), TestInstanceNamespaceResolution(), TestIsIndex() (+31 more)

### Community 29 - "podmanRemove"
Cohesion: 0.19
Nodes (13): podmanRemove(), QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported() (+5 more)

### Community 30 - "runner.go"
Cohesion: 0.16
Nodes (11): go_pkg_os_exec, canIVerb(), Attacher, LogsOpts, Streamer, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform() (+3 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (41): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+33 more)

### Community 32 - "status.go"
Cohesion: 0.12
Nodes (18): sortInstances(), newProgress(), actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), markMissing() (+10 more)

### Community 43 - "Side"
Cohesion: 0.11
Nodes (14): Image, applyDest(), digitRun(), Cred, Side, yaml.Node, isDigit(), nodePtr() (+6 more)

### Community 44 - "main.go"
Cohesion: 0.15
Nodes (32): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), confirmRemove(), emit(), envPairs() (+24 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.12
Nodes (49): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+41 more)

### Community 48 - "Kubernetes"
Cohesion: 0.14
Nodes (21): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, ImagePullSecret, Kubernetes (+13 more)

### Community 49 - "Command details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (32): go_pkg_slices, ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys() (+24 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.31
Nodes (19): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+11 more)

### Community 56 - "consolidate.go"
Cohesion: 0.13
Nodes (24): leaderNameFn, Opts, secretFn, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_tls, appendPassthrough(), applyStatusAccess(), binderOwner(), buildBundle() (+16 more)

### Community 57 - "render/render_test.go"
Cohesion: 0.25
Nodes (18): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+10 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "maven_test.go"
Cohesion: 0.24
Nodes (34): groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures() (+26 more)

### Community 60 - "write"
Cohesion: 0.11
Nodes (37): downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun() (+29 more)

### Community 61 - "go_pkg_gopkg_in_yaml_v3"
Cohesion: 0.14
Nodes (21): go_pkg_gopkg_in_yaml_v3, TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Service, applyDockerDefaults(), applyPodmanDefaults() (+13 more)

### Community 62 - "logs.go"
Cohesion: 0.22
Nodes (8): actLogs(), checkLogsFlags(), logsInvocation(), readLog(), go_pkg_context, logsOpts, sinceFlag, tailFlag

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (32): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods() (+24 more)

### Community 64 - "go_pkg_strconv"
Cohesion: 0.13
Nodes (26): go_pkg_bufio, go_pkg_path, go_pkg_strconv, imageMismatchNote(), imageNameTag(), loadImageLibs(), omitListProvenance(), splitJarBasename() (+18 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+17 more)

### Community 67 - "Render"
Cohesion: 0.17
Nodes (22): Render(), splitHealthBlock(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderHeaderNamesEveryReportedFact() (+14 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "go_pkg_strings"
Cohesion: 0.10
Nodes (14): go_pkg_strings, go_pkg_testing, strings.Builder, TestBothConfigsReadTheSameThreeProperties(), TestContainerPathAndFileNameAgree(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML() (+6 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "render"
Cohesion: 0.29
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.21
Nodes (12): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestRemoveKubernetesSeamHappyPath(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone(), TestRemoveNamespaceProbeArgvIsOneQuery() (+4 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "Defaults"
Cohesion: 0.18
Nodes (17): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, LeaderElection, Logging (+9 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.29
Nodes (7): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6. Workflow file

### Community 79 - "gen/imagepull_test.go"
Cohesion: 0.42
Nodes (8): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret()

### Community 80 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 81 - "podmanDeploy"
Cohesion: 0.22
Nodes (11): podmanDeploy(), DockerPlan, NamedDoc, SecretRef, TestGeneratePodmanQuadlet(), TestResolveCredentials(), PodmanPlan, podmanSecretRefs() (+3 more)

### Community 82 - "ParseEnv"
Cohesion: 0.08
Nodes (35): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+27 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): go_pkg_encoding_xml, encoding/xml.Name, imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties() (+27 more)

### Community 84 - "shell.go"
Cohesion: 0.32
Nodes (10): namedInstance(), actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), go_pkg_errors (+2 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (30): go_pkg_io, go_pkg_net_http, go_pkg_net_url, net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification() (+22 more)

### Community 88 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "TestParsingHelpersIgnoreAWarningOnStderr"
Cohesion: 0.12
Nodes (17): EngineImageInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON(), KubernetesPodsJSON(), KubernetesTop(), contains(), TestEngineImageInspectJSONArgv() (+9 more)

### Community 91 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 93 - "go_pkg_os"
Cohesion: 0.16
Nodes (16): addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape(), assertNoticeAtTop(), supportNoticePlain() (+8 more)

### Community 94 - "Expand"
Cohesion: 0.21
Nodes (20): go_pkg_reflect, reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf() (+12 more)

### Community 95 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), Defaults, Env (+1 more)

### Community 96 - "helperProcessArgv"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 97 - "namespace.go"
Cohesion: 0.33
Nodes (8): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_deploy, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_runner, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_validate, nsItem

### Community 98 - "dispatch"
Cohesion: 0.08
Nodes (53): dispatch(), captureStderr(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestCliPreflightFailureOpensNothing() (+45 more)

### Community 99 - "examples_test.go"
Cohesion: 0.32
Nodes (7): go_pkg_bytes, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_scan, mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflowExamplesMatchGoldenSpecs(), TestWriteCreatesSkipsForces()

### Community 100 - "vt_windows.go"
Cohesion: 0.50
Nodes (3): enableVirtualTerminal(), go_pkg_syscall, go_pkg_unsafe

### Community 102 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 105 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 106 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 107 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 110 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

## Knowledge Gaps
- **130 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Defaults`, `kubeDeployment` (+125 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 206 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `progress` connect `go_pkg_fmt` to `status.go`, `parse.go`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes` to `validate.go`, `hasErr`, `Defaults`, `Runner`, `spec_test.go`, `deploy_test.go`, `go_pkg_gopkg_in_yaml_v3`, `Env`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `dispatch()` connect `dispatch` to `main_test.go`, `nsList`, `main.go`, `completion_test.go`, `write`, `Runner`, `withProgressClock`, `testing.T`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 134 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 134 INFERRED edges - model-reasoned connections that need verification._
- **Are the 61 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 61 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _130 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._