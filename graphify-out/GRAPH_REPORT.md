# Graph Report - solace-ibmmq-connector-helper  (2026-09-17)

## Corpus Check
- 97 files · ~314,377 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2029 nodes · 6604 edges · 98 communities (80 shown, 18 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1073 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `89460685`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- hasErr
- testing.T
- go_pkg_testing
- gen.go
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- progress
- Scan
- completion.go
- runner_test.go
- uuid.go
- completion_test.go
- solmq-conn-util user guide
- ExecArgv
- Render
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- gen_extra_test.go
- podmangen_test.go
- main.go
- net/http.Response
- withProgressClock
- main_test.go
- PodmanDeploy
- commands.go
- statusreport.go
- statusCollector
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
- Runner
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
- Application
- consolidate_test.go
- maven_test.go
- write
- Docker
- logs.go
- parse_test.go
- go_pkg_fmt
- parse.go
- Defaults
- Render
- TestDownloadSetMapMatchesModel
- go_pkg_strings
- statusreport/render.go
- statusreport/render_test.go
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- Defaults
- 13. Logs: the lines behind the state
- 6. Workflow file
- GenerateKubernetes
- spec.go
- GeneratePodman
- LogsArgv
- maven.go
- shell.go
- solmq-conn-util abbreviations
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- libs.go
- cmd/solmq-conn-util
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- runner.go
- testcatalog_test.go
- Cmd
- commands_doc_test.go
- helperProcessArgv
- status.go
- dispatch
- allowCommandValue

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 139 edges
2. `hasErr()` - 83 edges
3. `captureStderr()` - 68 edges
4. `write()` - 65 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 46 edges
8. `captureStdout()` - 44 edges
9. `Runner` - 41 edges
10. `imageOK()` - 40 edges

## Surprising Connections (you probably didn't know these)
- `renderedCompletions()` --calls--> `render()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/statusreport/render_test.go
- `TestCompletionCoversModel()` --calls--> `contains()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/runner/runner_test.go
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `readLog()` --references--> `Streamer`  [EXTRACTED]
  cmd/solmq-conn-util/logs.go → internal/runner/runner.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (98 total, 18 thin omitted)

### Community 0 - "Build"
Cohesion: 0.10
Nodes (34): Build(), displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+26 more)

### Community 1 - "validate.go"
Cohesion: 0.11
Nodes (50): synthWorkflows(), Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker() (+42 more)

### Community 2 - "hasErr"
Cohesion: 0.08
Nodes (104): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+96 more)

### Community 3 - "testing.T"
Cohesion: 0.10
Nodes (45): run(), captureStdout(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+37 more)

### Community 4 - "go_pkg_testing"
Cohesion: 0.06
Nodes (59): go_pkg_bytes, go_pkg_flag, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_gen, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_scan, go_pkg_testing, mustWrite(), testResolver(), TestShippedExamplesGenerateConfig() (+51 more)

### Community 5 - "gen.go"
Cohesion: 0.12
Nodes (27): built, DockerPlan, File, NamedDoc, go_pkg_crypto_rand, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_consolidate, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_dockergen, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_podmangen (+19 more)

### Community 6 - "go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec"
Cohesion: 0.16
Nodes (16): go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec, TestBothConfigsReadTheSameThreeProperties(), TestContainerPathAndFileNameAgree(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML(), MountPath(), SolaceProps() (+8 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "progress"
Cohesion: 0.23
Nodes (6): progressTrim(), go_pkg_sync, go_pkg_time, sync.Mutex, progress, progressMode

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "runner_test.go"
Cohesion: 0.08
Nodes (42): go_pkg_runtime, os.FileMode, canIVerb(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), attachFiles() (+34 more)

### Community 14 - "uuid.go"
Cohesion: 0.31
Nodes (7): go_pkg_crypto_sha1, go_pkg_encoding_hex, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.11
Nodes (19): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+11 more)

### Community 18 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 19 - "Render"
Cohesion: 0.10
Nodes (33): Input, Instance, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_yamlwriter, composeEscape(), yw, Render(), renderContentConfig(), renderHealthcheck() (+25 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.10
Nodes (52): Input, Instance, KV, go_pkg_encoding_base64, go_pkg_encoding_json, PullSecret, StoreFile, yw (+44 more)

### Community 23 - "gen_extra_test.go"
Cohesion: 0.18
Nodes (22): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigRejectsSecretNameConflict(), TestConfigStatusPasswordRandErrorNoOutput(), TestConfigWorkflowCap() (+14 more)

### Community 24 - "podmangen_test.go"
Cohesion: 0.19
Nodes (21): go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_logback, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_statusscript, SecretRef, leaderLabels(), RenderQuadlet(), seconds(), fullInput(), Input (+13 more)

### Community 25 - "main.go"
Cohesion: 0.12
Nodes (42): resolveTarget(), runLogs(), absPath(), absResolver(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains() (+34 more)

### Community 27 - "withProgressClock"
Cohesion: 0.16
Nodes (17): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+9 more)

### Community 28 - "main_test.go"
Cohesion: 0.07
Nodes (42): containsToken(), manyWorkflowsDir(), TestAbsPath(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+34 more)

### Community 29 - "PodmanDeploy"
Cohesion: 0.19
Nodes (12): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported(), TestPodmanDeploySystemModeNoUserFlag() (+4 more)

### Community 30 - "commands.go"
Cohesion: 0.22
Nodes (20): flagEntries(), flagsLine(), flagSpan(), invocation(), pad(), renderCommandsDoc(), tableCell(), targetEntries() (+12 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (43): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), stateCell(), Banner(), banner() (+35 more)

### Community 32 - "statusCollector"
Cohesion: 0.18
Nodes (13): sortInstances(), confirmInstall(), instanceNames(), markMissing(), MergeService(), ObjectExists(), ParseDeployment(), TestObjectExists() (+5 more)

### Community 43 - "Side"
Cohesion: 0.11
Nodes (5): Image, Cred, Side, placeholderSecretRef(), LeaderElection

### Community 44 - "Runner"
Cohesion: 0.23
Nodes (27): actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit(), failFast(), genConfig() (+19 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.12
Nodes (49): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+41 more)

### Community 48 - "Kubernetes"
Cohesion: 0.16
Nodes (19): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, ImagePullSecret, Kubernetes (+11 more)

### Community 49 - "Command details"
Cohesion: 0.07
Nodes (29): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+21 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (33): go_pkg_slices, ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys() (+25 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.19
Nodes (27): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+19 more)

### Community 56 - "consolidate.go"
Cohesion: 0.12
Nodes (25): leaderNameFn, Opts, secretFn, go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_tls, go_pkg_regexp, appendPassthrough(), applyStatusAccess(), binderOwner() (+17 more)

### Community 57 - "Application"
Cohesion: 0.25
Nodes (18): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+10 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "maven_test.go"
Cohesion: 0.24
Nodes (34): groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures() (+26 more)

### Community 60 - "write"
Cohesion: 0.09
Nodes (39): downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials() (+31 more)

### Community 61 - "Docker"
Cohesion: 0.14
Nodes (21): go_pkg_gopkg_in_yaml_v3, TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Service, applyDockerDefaults(), applyPodmanDefaults() (+13 more)

### Community 62 - "logs.go"
Cohesion: 0.22
Nodes (8): actLogs(), checkLogsFlags(), logsInvocation(), readLog(), go_pkg_context, logsOpts, sinceFlag, tailFlag

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (28): ApplyStats(), EngineNamesByImage(), Instance, ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+20 more)

### Community 64 - "go_pkg_fmt"
Cohesion: 0.08
Nodes (45): go_pkg_bufio, go_pkg_fmt, go_pkg_path, go_pkg_reflect, reflect.Value, imageMismatchNote(), imageNameTag(), loadImageLibs() (+37 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+17 more)

### Community 67 - "Render"
Cohesion: 0.15
Nodes (24): go_pkg_strconv, breEscape(), Render(), splitHealthBlock(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress() (+16 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "go_pkg_strings"
Cohesion: 0.16
Nodes (13): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+5 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.22
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.12
Nodes (22): nsItemJSON(), nsList(), TestLogsPlatformFlagSkipsTheMenu(), TestLogsPlatformMenuWhenSeveralSectionsArePresent(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestPlatformMenuOnMultipleSections(), TestRemoveDeclinedTouchesNothing() (+14 more)

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
Cohesion: 0.22
Nodes (16): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, Logging, Management (+8 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.29
Nodes (7): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6. Workflow file

### Community 79 - "GenerateKubernetes"
Cohesion: 0.19
Nodes (16): KubeOpts, b64(), dockerConfigJSON(), TestResolveCredentials(), TestResolveStores(), GenerateKubernetes(), Resolver, ResolveCredentials() (+8 more)

### Community 80 - "spec.go"
Cohesion: 0.31
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 81 - "GeneratePodman"
Cohesion: 0.19
Nodes (14): mount, Mount, TestNamesAndPaths(), TestTargetMounts(), GeneratePodman(), pathIn(), PodmanServiceName(), targetMounts() (+6 more)

### Community 82 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): go_pkg_encoding_xml, go_pkg_io, encoding/xml.Name, imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey() (+28 more)

### Community 84 - "shell.go"
Cohesion: 0.28
Nodes (11): namedInstance(), actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), go_pkg_errors (+3 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (29): go_pkg_net_http, go_pkg_net_url, net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar() (+21 more)

### Community 88 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image and timezone (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "runner.go"
Cohesion: 0.09
Nodes (28): go_pkg_os_exec, EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Streamer, KubernetesGetJSON(), KubernetesListJSON() (+20 more)

### Community 91 - "testcatalog_test.go"
Cohesion: 0.18
Nodes (16): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+8 more)

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 93 - "commands_doc_test.go"
Cohesion: 0.29
Nodes (10): TestAbbreviationDocInSync(), assertHelpWidth(), assertNoAliases(), normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages() (+2 more)

### Community 96 - "helperProcessArgv"
Cohesion: 0.17
Nodes (15): helperProcessArgv(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput(), TestOSRunRejectsUnresolvableArgv0(), TestOSRunSplitKeepsTheStreamsApart(), TestOSRunSplitRejectsEmptyAndUnresolvableArgv() (+7 more)

### Community 97 - "status.go"
Cohesion: 0.09
Nodes (24): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), newProgress(), actStatus(), checkStatusFlags(), clearScreen() (+16 more)

### Community 98 - "dispatch"
Cohesion: 0.08
Nodes (54): dispatch(), captureStderr(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone(), TestDownloadMissingAndUnknownWordsRejected() (+46 more)

## Knowledge Gaps
- **131 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Splitter`, `Defaults` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 202 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Kubernetes` connect `Kubernetes` to `validate.go`, `hasErr`, `Runner`, `Defaults`, `spec_test.go`, `deploy_test.go`, `instances.go`, `Docker`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `Model` connect `Model` to `gen.go`, `Defaults`, `deploy_test.go`, `render.go`, `Application`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Side` connect `Side` to `Build`, `validate.go`, `hasErr`, `Defaults`, `spec.go`, `consolidate.go`, `consolidate_test.go`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Are the 134 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 134 INFERRED edges - model-reasoned connections that need verification._
- **Are the 60 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 60 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._