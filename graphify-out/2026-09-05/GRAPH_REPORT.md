# Graph Report - solace-ibmmq-connector-helper  (2026-09-04)

## Corpus Check
- 97 files · ~302,367 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1927 nodes · 6008 edges · 101 communities (85 shown, 16 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1055 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d5cb810b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- hasErr
- dispatch
- golden_test.go
- gen.go
- SolaceProps
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- commands.go
- Scan
- completion.go
- runner.go
- DurableName
- Expand
- solmq-conn-util user guide
- ExecArgv
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- os.File
- ParseEnv
- main.go
- net/http.Response
- completion_test.go
- main_test.go
- PodmanDeploy
- run
- statusreport.go
- renderAbbreviationDoc
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
- Cred
- Runner
- solmq-conn-util -- Development Guide
- Download
- DEVELOPMENT.md
- helperProcessArgv
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
- Render
- actLogs
- parse_test.go
- libs/image_test.go
- parse.go
- Defaults
- buildLeaderElection
- TestDownloadSetMapMatchesModel
- ParseCommand
- statusreport/render.go
- render
- nsList
- 12. Status: the container, the connector, or both
- testcatalog_test.go
- 10. `download jar`
- Defaults
- 13. Logs: the lines behind the state
- 6. Workflow file
- LogsArgv
- spec.go
- solmq-conn-util command reference
- Env
- maven.go
- actShell
- solmq-conn-util abbreviations
- Syslog
- libs.go
- auto-complete
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- 9. Secrets model
- absPath
- Cmd
- attachRunner
- Write
- runner_test.go
- removeNamespace
- testing.T
- Workload
- statusCollector
- allowCommandValue

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 136 edges
2. `hasErr()` - 82 edges
3. `write()` - 65 edges
4. `captureStderr()` - 59 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 46 edges
8. `captureStdout()` - 41 edges
9. `Runner` - 41 edges
10. `imageOK()` - 40 edges

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

## Communities (101 total, 16 thin omitted)

### Community 0 - "Build"
Cohesion: 0.13
Nodes (25): Build(), displayName(), containsSub(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed() (+17 more)

### Community 1 - "validate.go"
Cohesion: 0.09
Nodes (52): fixedLeaderNames(), synthWorkflows(), Side, Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials() (+44 more)

### Community 2 - "hasErr"
Cohesion: 0.08
Nodes (103): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+95 more)

### Community 3 - "dispatch"
Cohesion: 0.10
Nodes (42): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance() (+34 more)

### Community 4 - "golden_test.go"
Cohesion: 0.16
Nodes (26): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+18 more)

### Community 5 - "gen.go"
Cohesion: 0.05
Nodes (91): built, DockerPlan, File, KubeOpts, mount, NamedDoc, SecretRef, b64() (+83 more)

### Community 6 - "SolaceProps"
Cohesion: 0.26
Nodes (11): MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (24): cli, cmd/solmq-conn-util, Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples (+16 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "commands.go"
Cohesion: 0.18
Nodes (26): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+18 more)

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "runner.go"
Cohesion: 0.09
Nodes (27): canIVerb(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON() (+19 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.24
Nodes (18): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+10 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.10
Nodes (20): 11. What gets generated, 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas (+12 more)

### Community 18 - "ExecArgv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

### Community 19 - "Render"
Cohesion: 0.09
Nodes (32): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+24 more)

### Community 21 - "Render"
Cohesion: 0.11
Nodes (49): Input, Instance, KV, PullSecret, StoreFile, yw, leaderMode(), ManagementPort() (+41 more)

### Community 23 - "os.File"
Cohesion: 0.17
Nodes (12): enableVirtualTerminal(), enableVirtualTerminal(), os.File, mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError() (+4 more)

### Community 24 - "ParseEnv"
Cohesion: 0.05
Nodes (61): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), ParseEnv() (+53 more)

### Community 25 - "main.go"
Cohesion: 0.13
Nodes (39): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage(), envFlag() (+31 more)

### Community 27 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 28 - "main_test.go"
Cohesion: 0.08
Nodes (48): captureStderr(), manyWorkflowsDir(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing() (+40 more)

### Community 29 - "PodmanDeploy"
Cohesion: 0.19
Nodes (12): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported(), TestPodmanDeploySystemModeNoUserFlag() (+4 more)

### Community 30 - "run"
Cohesion: 0.17
Nodes (16): run(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves() (+8 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (42): time.Time, heapValue(), parseHeap(), Age(), Banner(), banner(), Bytes(), canonicalRef() (+34 more)

### Community 32 - "renderAbbreviationDoc"
Cohesion: 0.22
Nodes (13): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+5 more)

### Community 44 - "Runner"
Cohesion: 0.24
Nodes (25): actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit(), failFast(), genConfig() (+17 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 47 - "DEVELOPMENT.md"
Cohesion: 0.33
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "helperProcessArgv"
Cohesion: 0.17
Nodes (15): helperProcessArgv(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput(), TestOSRunRejectsUnresolvableArgv0(), TestOSRunSplitKeepsTheStreamsApart(), TestOSRunSplitRejectsEmptyAndUnresolvableArgv() (+7 more)

### Community 49 - "Command details"
Cohesion: 0.11
Nodes (18): cli, Command details, deploy, download, examples, generate, help, logs (+10 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (31): ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe() (+23 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.31
Nodes (19): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+11 more)

### Community 56 - "consolidate.go"
Cohesion: 0.14
Nodes (17): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), TestAppendPassthroughCollision(), TestFormatScalarQuoting(), TestNodeToProps(), TestSanitizeAndIsTCPS() (+9 more)

### Community 57 - "Application"
Cohesion: 0.25
Nodes (18): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+10 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "maven_test.go"
Cohesion: 0.25
Nodes (33): metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures(), pomXMLBody() (+25 more)

### Community 60 - "write"
Cohesion: 0.13
Nodes (34): podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+26 more)

### Community 61 - "Render"
Cohesion: 0.16
Nodes (20): breEscape(), Render(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderHeaderNamesEveryReportedFact() (+12 more)

### Community 62 - "actLogs"
Cohesion: 0.20
Nodes (9): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), Streamer, logsOpts, sinceFlag (+1 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys(), TestApplyStats() (+22 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (24): imageMismatchNote(), imageNameTag(), imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote() (+16 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, ApplyStats(), ApplyTop(), connectorIndex(), digestFrom(), engineComponents(), exitCode(), Instance (+17 more)

### Community 67 - "buildLeaderElection"
Cohesion: 0.22
Nodes (14): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), testSecretRef(), trustStoreVal() (+6 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "ParseCommand"
Cohesion: 0.15
Nodes (13): ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure() (+5 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "render"
Cohesion: 0.28
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.33
Nodes (9): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveNamespaceOccupiedLeavesItAlone(), TestRemoveNamespaceProbeArgvIsOneQuery(), TestRemoveNoPromptNeverRemovesAnOccupiedNamespace(), TestRemoveNoPromptSkipsThePromptEntirely() (+1 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 12.10 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch`, 12.7 `--output json` (+3 more)

### Community 74 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "Defaults"
Cohesion: 0.26
Nodes (15): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, LeaderElection, Management, rawDefaults (+7 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.29
Nodes (7): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (warnings), 6.6 Reusable connections (`conn-ref`), 6. Workflow file

### Community 79 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 80 - "spec.go"
Cohesion: 0.31
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 81 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 82 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), Defaults, Env (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (32): encoding/xml.Name, acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+24 more)

### Community 84 - "actShell"
Cohesion: 0.42
Nodes (8): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), Attacher, shellOpts

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "Syslog"
Cohesion: 0.67
Nodes (3): Syslog, checkSyslog(), Logging

### Community 87 - "libs.go"
Cohesion: 0.12
Nodes (28): net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), filenameFromEscapedPath(), Input (+20 more)

### Community 88 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image and timezone (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 91 - "absPath"
Cohesion: 0.67
Nodes (3): absPath(), absResolver(), fileReader()

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 93 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 96 - "runner_test.go"
Cohesion: 0.12
Nodes (30): os.FileMode, Preflight(), ScriptInstalled(), attachFiles(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText() (+22 more)

### Community 97 - "removeNamespace"
Cohesion: 0.50
Nodes (7): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), ownedNames(), removeNamespace(), nsItem

### Community 98 - "testing.T"
Cohesion: 0.09
Nodes (41): downloadEnvWithImage(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+33 more)

### Community 99 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 103 - "statusCollector"
Cohesion: 0.13
Nodes (15): sortInstances(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), watchStatus() (+7 more)

## Knowledge Gaps
- **130 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Splitter`, `Defaults` (+125 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `dispatch` to `testing.T`, `nsList`, `commands.go`, `Runner`, `main_test.go`, `main.go`, `write`, `run`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Runner` connect `Runner` to `runner_test.go`, `removeNamespace`, `dispatch`, `ParseCommand`, `statusCollector`, `runner.go`, `ExecArgv`, `actShell`, `instances.go`, `main.go`, `Cmd`, `PodmanDeploy`, `actLogs`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `ParseEnv` to `removeNamespace`, `validate.go`, `hasErr`, `Env`, `spec_test.go`, `Render`, `Syslog`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 131 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 131 INFERRED edges - model-reasoned connections that need verification._
- **Are the 59 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 59 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _130 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.13105413105413105 - nodes in this community are weakly interconnected._