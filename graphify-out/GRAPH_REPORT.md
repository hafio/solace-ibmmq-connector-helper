# Graph Report - solace-ibmmq-connector-helper  (2026-10-10)

## Corpus Check
- 113 files · ~365,454 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 11 file(s) not represented in the graph (top: (none) 4, .list 3, .jks 2)

## Summary
- 2284 nodes · 8786 edges · 130 communities (91 shown, 39 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1274 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1cd2b728`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- statusscript_test.go
- testing.T
- validate_extra_test.go
- dispatch
- gen_extra_test.go
- Defaults
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- GenerateDocker
- scan_test.go
- completion.go
- TestGoldenKubernetesNoSecrets
- attachRunner
- completion_test.go
- User Guide Core Chapters
- os.File
- run
- deploy_test.go
- Go Module Definition
- gen.go
- RenderQuadlet
- runAction
- progress
- withProgressClock
- validate.go
- JavaOptions
- podmanDeploy
- statusreport.go
- statusCollector
- Graphify Project Guidance
- Consolidate Package
- Deploy Package
- Gen Package
- Render Package
- Scan Package
- Spec Package
- TLS Package
- Validate Package
- Golden Application YAML
- Side
- main.go
- Development Guide
- libs_test.go
- Kubernetes Object Names And Deploy
- Command Reference Details
- Golden Bash Completion
- Model
- support_test.go
- Application
- instanceSession
- Build
- abbreviation_doc_test.go
- commands.go
- maven_test.go
- write
- Env
- sinceFlag
- parse_test.go
- libs/image_test.go
- GenerateKubernetes
- Management Endpoint Defaults
- Resolver
- Dispatch Model Consistency Tests
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- nsList
- Status Command User Guide
- CLI Shell User Guide
- Download Jar User Guide
- Writer
- Logs Command User Guide
- Workflow File User Guide
- yaml.Node
- Runner
- decodeCreate
- flattenBlock
- Maven Version And POM Resolution
- ResolveCredentials
- Command Abbreviation Reference
- Auto Complete Command Reference
- Jar Download And Verification
- binderNames
- Platform Sections User Guide
- buildLeaderElection
- runner_test.go
- filenameFromEscapedPath
- Cmd
- Expand
- LogsArgv
- TestSanitizeAndIsTCPS
- repeatableName
- main_test.go
- tailFlag
- Secrets Model User Guide
- Known And Managed Keys
- Command Reference Overview
- Project Overview And Quick Start
- dockergen_test.go
- Status Command Reference
- CLI Package Test Catalog
- yw
- FormatScalar
- shippedExamplesRequest
- XML

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 140 edges
2. `hasErr()` - 89 edges
3. `captureStderr()` - 71 edges
4. `write()` - 65 edges
5. `Build()` - 63 edges
6. `Download()` - 57 edges
7. `wfOK()` - 56 edges
8. `imageOK()` - 46 edges
9. `captureStdout()` - 44 edges
10. `ParseEnv()` - 43 edges

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

## Communities (130 total, 39 thin omitted)

### Community 0 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 1 - "testing.T"
Cohesion: 0.04
Nodes (88): TestIsIndex(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes(), TestSortIsByNameThenNamespace(), TestStatusRejectsUnsafeUserBeforeAnyExec(), TestWatchFlagAcceptsBareAndInterval() (+80 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.05
Nodes (143): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+135 more)

### Community 3 - "dispatch"
Cohesion: 0.10
Nodes (39): dispatch(), captureStdout(), containsToken(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList(), TestLogsKubernetesArgvShape(), TestLogsNamedPodIsNotPreChecked() (+31 more)

### Community 4 - "gen_extra_test.go"
Cohesion: 0.18
Nodes (28): Config(), issuesContain(), synthWorkflowFiles(), synthWorkflows(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder() (+20 more)

### Community 5 - "Defaults"
Cohesion: 0.12
Nodes (17): Defaults, Security, TLSConfig, Syslog, checkRemovedDefaultsKeys(), checkSyslog(), LeaderElection, Logging (+9 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.31
Nodes (10): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), isFalse(), isLiteral(), nearestKnownKey() (+2 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "GenerateDocker"
Cohesion: 0.13
Nodes (19): DockerPlan, mount, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestTargetMounts(), GenerateDocker() (+11 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.22
Nodes (19): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+11 more)

### Community 14 - "attachRunner"
Cohesion: 0.32
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (25): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+17 more)

### Community 17 - "User Guide Core Chapters"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "os.File"
Cohesion: 0.13
Nodes (9): newProgress(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), enableVirtualTerminal(), Attacher, statusOpts (+1 more)

### Community 19 - "run"
Cohesion: 0.14
Nodes (18): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+10 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.09
Nodes (53): Input, Instance, KV, envValueQuote(), PullSecret, leaderMode(), ManagementPort(), NamespaceManifest() (+45 more)

### Community 23 - "gen.go"
Cohesion: 0.16
Nodes (37): isClusterDefault(), isOurs(), namespaceOccupants(), buildBundle(), Bundle, storeOrigin(), DurableName(), mustParseUUID() (+29 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "runAction"
Cohesion: 0.13
Nodes (36): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage(), envFlag() (+28 more)

### Community 26 - "progress"
Cohesion: 0.23
Nodes (4): progressElapsed(), progressTrim(), progress, progressMode

### Community 27 - "withProgressClock"
Cohesion: 0.20
Nodes (13): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+5 more)

### Community 28 - "validate.go"
Cohesion: 0.11
Nodes (48): Workflow, checkAllExtraKeys(), propPrefix(), checkConnections(), checkContainerStores(), checkContainerTarget(), CheckDeployCommand(), checkDocker() (+40 more)

### Community 29 - "JavaOptions"
Cohesion: 0.26
Nodes (13): JavaOptions, YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders() (+5 more)

### Community 30 - "podmanDeploy"
Cohesion: 0.13
Nodes (22): podmanDeploy(), podmanRemove(), PodmanPlan, PodmanFileSecretNames(), PodmanSecretStoreName(), QuadletScope, PodmanDeploy(), PodmanRemove() (+14 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (42): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), stateCell(), Banner(), banner() (+34 more)

### Community 32 - "statusCollector"
Cohesion: 0.17
Nodes (13): noPodsFound(), sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), PodmanServiceName(), MergeService() (+5 more)

### Community 43 - "Side"
Cohesion: 0.12
Nodes (10): fixedLeaderNames(), Image, Cred, Side, placeholderSecretRef(), checkCred(), checkDefaultsCredentials(), checkSide() (+2 more)

### Community 44 - "main.go"
Cohesion: 0.14
Nodes (30): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+22 more)

### Community 45 - "Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 48 - "Kubernetes Object Names And Deploy"
Cohesion: 0.25
Nodes (6): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes

### Community 49 - "Command Reference Details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.14
Nodes (19): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+11 more)

### Community 52 - "support_test.go"
Cohesion: 0.33
Nodes (8): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 54 - "Application"
Cohesion: 0.10
Nodes (32): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+24 more)

### Community 55 - "instanceSession"
Cohesion: 0.13
Nodes (32): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+24 more)

### Community 56 - "Build"
Cohesion: 0.11
Nodes (34): Opts, Build(), containsSub(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestApplyStatusAccessNoOperatorUsers(), TestBuildCarriesEachWorkflowsTransformBlocks() (+26 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "commands.go"
Cohesion: 0.16
Nodes (27): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+19 more)

### Community 59 - "maven_test.go"
Cohesion: 0.22
Nodes (37): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep() (+29 more)

### Community 60 - "write"
Cohesion: 0.13
Nodes (36): downloadEnvWithImage(), podmanEnv(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+28 more)

### Community 61 - "Env"
Cohesion: 0.09
Nodes (28): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown() (+20 more)

### Community 63 - "parse_test.go"
Cohesion: 0.15
Nodes (21): ObjectExists(), ParseApplication(), ParsePods(), splitKV(), keys(), TestObjectExists(), TestParseApplication(), TestParseApplicationCRLFAndBlankLines() (+13 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.07
Nodes (48): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+40 more)

### Community 65 - "GenerateKubernetes"
Cohesion: 0.18
Nodes (15): File, KubeOpts, build(), TestGenerateKubernetesCarriesStatusScript(), TestGenValidateStoresWarning(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate(), TestToIssues() (+7 more)

### Community 67 - "Resolver"
Cohesion: 0.12
Nodes (15): PodmanFile, StoreFile, b64(), TestResolvePodmanFiles(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError() (+7 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.10
Nodes (27): ApplyStats(), connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod() (+19 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.22
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "nsList"
Cohesion: 0.25
Nodes (11): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone(), TestRemoveNamespaceProbeArgvIsOneQuery() (+3 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "CLI Shell User Guide"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "Download Jar User Guide"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 77 - "Logs Command User Guide"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "Workflow File User Guide"
Cohesion: 0.22
Nodes (9): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file (+1 more)

### Community 79 - "yaml.Node"
Cohesion: 0.20
Nodes (8): defaultsFromRaw(), applyDest(), rawMQ, rawSolace, nodePtr(), presentBlock(), rawSide, rawWorkflow

### Community 80 - "Runner"
Cohesion: 0.13
Nodes (26): EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Runner, KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON() (+18 more)

### Community 81 - "decodeCreate"
Cohesion: 0.28
Nodes (5): decodeCreate(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 82 - "flattenBlock"
Cohesion: 0.24
Nodes (13): envTransforms(), deref(), extraKeys(), flattenBlock(), rawMQ, rawSolace, isMergeKey(), appendTransformKeys() (+5 more)

### Community 83 - "Maven Version And POM Resolution"
Cohesion: 0.14
Nodes (33): acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML(), highestStable() (+25 more)

### Community 84 - "ResolveCredentials"
Cohesion: 0.24
Nodes (10): dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues() (+2 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "Jar Download And Verification"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "binderNames"
Cohesion: 0.36
Nodes (8): binderNames(), binderOf(), eqStrs(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder(), TestBinderFieldsCarryStablePlaceholders()

### Community 89 - "Platform Sections User Guide"
Cohesion: 0.33
Nodes (6): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`), Connector 2.x and 3.x

### Community 90 - "buildLeaderElection"
Cohesion: 0.17
Nodes (15): leaderNameFn, secretFn, appendPassthrough(), binderOwner(), buildLeaderElection(), propsNode(), TestAppendPassthroughCollision(), TestBuildLeaderElection() (+7 more)

### Community 91 - "runner_test.go"
Cohesion: 0.06
Nodes (63): TestWriteMkdirError(), canIVerb(), ExecArgv(), InstallScript(), Preflight(), RunStatusScript(), ScriptInstalled(), attachFiles() (+55 more)

### Community 92 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 93 - "Cmd"
Cohesion: 0.09
Nodes (19): applyCmdEnv(), applyCmdInput(), Docker(), Cmd, Streamer, Kubernetes(), kubeVerb(), ParseCommand() (+11 more)

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (18): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+10 more)

### Community 95 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 96 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 98 - "main_test.go"
Cohesion: 0.05
Nodes (83): captureStderr(), podmanEnvSudo(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape() (+75 more)

### Community 100 - "Secrets Model User Guide"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "Known And Managed Keys"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 102 - "Command Reference Overview"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "Project Overview And Quick Start"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "dockergen_test.go"
Cohesion: 0.11
Nodes (29): Input, Instance, composeEscape(), composeQuote(), Render(), renderContentConfig(), renderHealthcheck(), renderSecrets() (+21 more)

### Community 105 - "Status Command Reference"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 107 - "CLI Package Test Catalog"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 112 - "FormatScalar"
Cohesion: 0.15
Nodes (13): applyStatusAccess(), TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), Model (+5 more)

### Community 132 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 133 - "XML"
Cohesion: 0.50
Nodes (4): TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

## Knowledge Gaps
- **135 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `renderedTransform`, `solaceBlock` (+130 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 230 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **39 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Side` connect `Side` to `validate_extra_test.go`, `Defaults`, `checkExtraKeys`, `yaml.Node`, `FormatScalar`, `gen.go`, `Build`, `validate.go`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes Object Names And Deploy` to `testing.T`, `validate_extra_test.go`, `Defaults`, `main.go`, `deploy_test.go`, `gen.go`, `validate.go`, `Env`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `sinceFlag` connect `sinceFlag` to `gen.go`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 135 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 135 INFERRED edges - model-reasoned connections that need verification._
- **Are the 66 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 66 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _135 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `statusscript_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13054187192118227 - nodes in this community are weakly interconnected._