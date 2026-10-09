# Graph Report - solace-ibmmq-connector-helper  (2026-10-09)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 2282 nodes · 8772 edges · 132 communities (95 shown, 37 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1265 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `82cc6785`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Status Script And Generate Tests
- Env And Platform Defaults Tests
- Validation Rule Tests
- Status Logs CLI Command Tests
- Config And Platform Generation
- Connector Defaults Model
- Extra Key Validation
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- Defaults And Credential Parsing Tests
- Workflow File Scan Tests
- Shell Completion Rendering
- Generated Output Golden Tests
- Fake Runner Test Doubles
- Shell Completion Script Tests
- User Guide Core Chapters
- Status Command Action And Watch
- Extra Key Capture Tests
- Kubernetes Manifest Rendering
- Go Module Definition
- TLS Store Paths And Secrets
- Podman Quadlet Rendering
- CLI Command Handlers And Flags
- Status Progress Reporter
- Progress Spinner Tests
- Env And Workflow Validation
- Workflow Transform Validation
- Podman Quadlet Lifecycle
- Status Report Model And Formatting
- Status Report Collection
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
- Workflow Side And Credentials
- Generate And Deploy Actions
- Development Guide
- Jar Download Tests
- Kubernetes Object Names And Deploy
- Command Reference Details
- Golden Bash Completion
- Connector Binder And Binding Model
- Documentation Sync Tests
- Application YAML Output Tests
- Instance Discovery And Picker
- Binder Consolidation Tests
- Command And Abbreviation Doc Rendering
- Command Usage And Help Rendering
- Maven Dependency Closure Tests
- Deploy Remove And Platform Tests
- Docker And Podman Section Model
- Logs Command Flags And Reading
- Status Output Parsing Tests
- Image Libs And Test Catalog
- Env Image And Workflow Model
- Management Endpoint Defaults
- Platform Resolution And Env Loading
- Dispatch Model Consistency Tests
- Kubernetes And Engine JSON Parsing
- Status Report Rendering
- Status Table And JSON Tests
- Kubernetes Namespace Removal Tests
- Status Command User Guide
- CLI Shell User Guide
- Download Jar User Guide
- Java Options And YAML Writer
- Logs Command User Guide
- Workflow File User Guide
- CLI Command Shell Attach
- Runner Engine And Kubernetes Queries
- Kubernetes Section Model
- Extra Keys And Misplaced Transforms
- Maven Version And POM Resolution
- Stable Names And Status Access
- Command Abbreviation Reference
- Auto Complete Command Reference
- Jar Download And Verification
- Allow Command Flag
- Platform Sections User Guide
- Leader Election Build Tests
- Preflight And Secret Removal Tests
- Workflow Parsing Tests
- External Process Runner
- Environment Variable Expansion Tests
- Logs Argv And Secret Creation
- OS Process Execution Tests
- Kubernetes Create And Naming Tests
- Download And Flag Validation Tests
- Exec And Status Script Argv
- Secrets Model User Guide
- Known And Managed Keys
- Command Reference Overview
- Project Overview And Quick Start
- Docker Compose Rendering
- Status Command Reference
- Kubernetes Section Parsing Tests
- CLI Package Test Catalog
- Application YAML Rendering
- Secrets Create Value Parsing
- Passthrough Property Merging
- Shipped Examples And Scan Patterns
- Image Pull Syslog And Namespace

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 140 edges
2. `hasErr()` - 89 edges
3. `captureStderr()` - 71 edges
4. `write()` - 65 edges
5. `Build()` - 63 edges
6. `wfOK()` - 56 edges
7. `Download()` - 56 edges
8. `imageOK()` - 46 edges
9. `captureStdout()` - 44 edges
10. `ParseEnv()` - 43 edges

## Surprising Connections (you probably didn't know these)
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go
- `TestDownloadImageMismatchReported()` --calls--> `run()`  [INFERRED]
  internal/libs/libs_test.go → cmd/solmq-conn-util/main.go
- `TestDerivedNamesMustFitALabel()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_derivednames_test.go → cmd/solmq-conn-util/main.go
- `TestCheckCredRejectsReservedPrefix()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (132 total, 37 thin omitted)

### Community 0 - "Status Script And Generate Tests"
Cohesion: 0.10
Nodes (46): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+38 more)

### Community 1 - "Env And Platform Defaults Tests"
Cohesion: 0.15
Nodes (21): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestApplyDockerDefaultsFillsMissing() (+13 more)

### Community 2 - "Validation Rule Tests"
Cohesion: 0.06
Nodes (137): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey(), baseKubeDeploy(), baseKubeService() (+129 more)

### Community 3 - "Status Logs CLI Command Tests"
Cohesion: 0.08
Nodes (51): dispatch(), captureStdout(), containsToken(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+43 more)

### Community 4 - "Config And Platform Generation"
Cohesion: 0.06
Nodes (83): podmanDeploy(), podmanRemove(), DockerPlan, File, KubeOpts, mount, PodmanFile, SecretRef (+75 more)

### Community 5 - "Connector Defaults Model"
Cohesion: 0.24
Nodes (14): defaultsFromRaw(), Defaults, Security, TLSConfig, LeaderElection, Management, rawDefaults, rawLeader (+6 more)

### Community 6 - "Extra Key Validation"
Cohesion: 0.36
Nodes (11): CanonicalKey(), canonicalMatch(), checkAllExtraKeys(), checkExtraKeys(), editDistance(), isFalse(), isLiteral(), nearestKnownKey() (+3 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "Defaults And Credential Parsing Tests"
Cohesion: 0.14
Nodes (19): envTransforms(), ParseDefaults(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError(), TestParseDefaultsFull() (+11 more)

### Community 11 - "Workflow File Scan Tests"
Cohesion: 0.34
Nodes (15): Scan(), bases(), TestScanEmptyPatternDefaultsToStar(), TestScanEnvFileExcludedRegardlessOfPattern(), TestScanErrorMissingDir(), TestScanExcludesEnvFile(), TestScanPatternNoMatchIsEmptyNotError(), TestScanPatternWildcards() (+7 more)

### Community 12 - "Shell Completion Rendering"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "Generated Output Golden Tests"
Cohesion: 0.24
Nodes (21): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+13 more)

### Community 14 - "Fake Runner Test Doubles"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 15 - "Shell Completion Script Tests"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "User Guide Core Chapters"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "Status Command Action And Watch"
Cohesion: 0.14
Nodes (10): newProgress(), actStatus(), clearScreen(), confirmInstall(), quoteAll(), watchStatus(), enableVirtualTerminal(), repeatableName (+2 more)

### Community 19 - "Extra Key Capture Tests"
Cohesion: 0.23
Nodes (11): knownKeys(), extraKeyNames(), extraValue(), TestCanonicalKey(), TestKnownKeysReadsYAMLTags(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys() (+3 more)

### Community 21 - "Kubernetes Manifest Rendering"
Cohesion: 0.09
Nodes (53): Input, Instance, KV, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5() (+45 more)

### Community 23 - "TLS Store Paths And Secrets"
Cohesion: 0.16
Nodes (17): secretFn, buildBundle(), storeOrigin(), storeSecret(), TestParseWorkflowTransform(), TestParseWorkflowTransformHeaders(), transformFile(), MountPath() (+9 more)

### Community 24 - "Podman Quadlet Rendering"
Cohesion: 0.15
Nodes (25): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv() (+17 more)

### Community 25 - "CLI Command Handlers And Flags"
Cohesion: 0.24
Nodes (22): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), downloadDeployedImage(), envFlag(), flagExit(), outFlag() (+14 more)

### Community 26 - "Status Progress Reporter"
Cohesion: 0.21
Nodes (3): progressTrim(), progress, progressMode

### Community 27 - "Progress Spinner Tests"
Cohesion: 0.16
Nodes (16): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Env And Workflow Validation"
Cohesion: 0.11
Nodes (52): Workflow, checkConnections(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker() (+44 more)

### Community 29 - "Workflow Transform Validation"
Cohesion: 0.34
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 30 - "Podman Quadlet Lifecycle"
Cohesion: 0.19
Nodes (12): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported(), TestPodmanDeploySystemModeNoUserFlag() (+4 more)

### Community 31 - "Status Report Model And Formatting"
Cohesion: 0.07
Nodes (44): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Age(), Banner(), banner() (+36 more)

### Community 32 - "Status Report Collection"
Cohesion: 0.20
Nodes (10): sortInstances(), instanceNames(), markMissing(), PodmanServiceName(), MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Report (+2 more)

### Community 43 - "Workflow Side And Credentials"
Cohesion: 0.11
Nodes (10): applyDest(), Cred, rawMQ, rawSolace, Side, nodePtr(), presentBlock(), placeholderSecretRef() (+2 more)

### Community 44 - "Generate And Deploy Actions"
Cohesion: 0.16
Nodes (29): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+21 more)

### Community 45 - "Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Jar Download Tests"
Cohesion: 0.10
Nodes (54): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+46 more)

### Community 48 - "Kubernetes Object Names And Deploy"
Cohesion: 0.25
Nodes (6): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes

### Community 49 - "Command Reference Details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Connector Binder And Binding Model"
Cohesion: 0.20
Nodes (19): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+11 more)

### Community 52 - "Documentation Sync Tests"
Cohesion: 0.17
Nodes (14): addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync(), TestAbbreviationDocTableShape(), normLF() (+6 more)

### Community 54 - "Application YAML Output Tests"
Cohesion: 0.22
Nodes (24): Build(), Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport() (+16 more)

### Community 55 - "Instance Discovery And Picker"
Cohesion: 0.31
Nodes (19): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+11 more)

### Community 56 - "Binder Consolidation Tests"
Cohesion: 0.24
Nodes (20): binderOfKind(), extraWF(), scalarProps(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+12 more)

### Community 57 - "Command And Abbreviation Doc Rendering"
Cohesion: 0.33
Nodes (9): abbrevFlagByShort(), abbrevTable(), renderAbbreviationDoc(), flagsLine(), flagSpan(), renderCommandsDoc(), tableCell(), abbrevRow (+1 more)

### Community 58 - "Command Usage And Help Rendering"
Cohesion: 0.20
Nodes (21): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+13 more)

### Community 59 - "Maven Dependency Closure Tests"
Cohesion: 0.21
Nodes (37): TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep() (+29 more)

### Community 60 - "Deploy Remove And Platform Tests"
Cohesion: 0.08
Nodes (46): podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun() (+38 more)

### Community 61 - "Docker And Podman Section Model"
Cohesion: 0.18
Nodes (13): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), applyDockerDefaults(), applyPodmanDefaults(), Docker, LibsMount, Podman (+5 more)

### Community 62 - "Logs Command Flags And Reading"
Cohesion: 0.22
Nodes (7): actLogs(), checkLogsFlags(), logsInvocation(), readLog(), logsOpts, sinceFlag, tailFlag

### Community 63 - "Status Output Parsing Tests"
Cohesion: 0.11
Nodes (28): ApplyStats(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys(), TestApplyStats() (+20 more)

### Community 64 - "Image Libs And Test Catalog"
Cohesion: 0.06
Nodes (48): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+40 more)

### Community 65 - "Env Image And Workflow Model"
Cohesion: 0.26
Nodes (8): workflowsFromRaw(), Image, registryImage(), retiredNamesKube(), TestRetiredCreateNamesReportedOnlyByValidate(), rawEnv, rawWorkflows, Workflows

### Community 67 - "Platform Resolution And Env Loading"
Cohesion: 0.15
Nodes (17): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), confirmRemove(), contains(), loadEnvFile(), noPromptFlag() (+9 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "Kubernetes And Engine JSON Parsing"
Cohesion: 0.16
Nodes (22): connectorIndex(), digestFrom(), engineComponents(), EngineNamesByImage(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod() (+14 more)

### Community 70 - "Status Report Rendering"
Cohesion: 0.25
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "Status Table And JSON Tests"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "Kubernetes Namespace Removal Tests"
Cohesion: 0.19
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+6 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "CLI Shell User Guide"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "Download Jar User Guide"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "Java Options And YAML Writer"
Cohesion: 0.10
Nodes (19): JavaOptions, JavaEnv(), NormalizeJavaOptions(), TestJavaEnvMergesTheTLSFlag(), TestParseEnvJavaOptionsAbsent(), TestParseEnvJavaOptionsAlias(), TestParseEnvJavaOptionsShapeErrors(), TestParseEnvJavaOptionsSpellings() (+11 more)

### Community 77 - "Logs Command User Guide"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "Workflow File User Guide"
Cohesion: 0.22
Nodes (9): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file (+1 more)

### Community 79 - "CLI Command Shell Attach"
Cohesion: 0.32
Nodes (8): namedInstance(), actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), shellOpts

### Community 80 - "Runner Engine And Kubernetes Queries"
Cohesion: 0.09
Nodes (26): EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Attacher, Runner, KubernetesGetJSON(), KubernetesListJSON() (+18 more)

### Community 81 - "Kubernetes Section Model"
Cohesion: 0.21
Nodes (14): Deployment, Resources, Secrets, Service, CredCreate, CredentialsSecret, ImagePullSecret, Libs (+6 more)

### Community 82 - "Extra Keys And Misplaced Transforms"
Cohesion: 0.20
Nodes (14): deref(), extraKeys(), flattenBlock(), rawMQ, rawSolace, isMergeKey(), appendTransformKeys(), documentMapping() (+6 more)

### Community 83 - "Maven Version And POM Resolution"
Cohesion: 0.14
Nodes (33): acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML(), highestStable() (+25 more)

### Community 84 - "Stable Names And Status Access"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "Jar Download And Verification"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 89 - "Platform Sections User Guide"
Cohesion: 0.33
Nodes (6): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`), Connector 2.x and 3.x

### Community 90 - "Leader Election Build Tests"
Cohesion: 0.13
Nodes (22): leaderNameFn, buildLeaderElection(), containsSub(), fixedLeaderNames(), propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks() (+14 more)

### Community 91 - "Preflight And Secret Removal Tests"
Cohesion: 0.09
Nodes (30): TestWriteMkdirError(), PodmanSecretRemove(), Preflight(), ScriptInstalled(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), TestOSStreamRejectsEmptyAndUnresolvableArgv() (+22 more)

### Community 92 - "Workflow Parsing Tests"
Cohesion: 0.22
Nodes (9): TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), ParseWorkflow(), TestConnRefSideMayTuneBinding(), TestParseWorkflowAmbiguousSystemAndDest(), TestParseWorkflowConnRef(), TestParseWorkflowEnabledDefaultsTrue(), TestParseWorkflowSolaceAndMQ(), TestParseWorkflowSyntaxError() (+1 more)

### Community 93 - "External Process Runner"
Cohesion: 0.23
Nodes (8): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS

### Community 94 - "Environment Variable Expansion Tests"
Cohesion: 0.21
Nodes (18): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+10 more)

### Community 95 - "Logs Argv And Secret Creation"
Cohesion: 0.13
Nodes (19): canIVerb(), Docker(), LogsOpts, Kubernetes(), kubeVerb(), LogsArgv(), logsCommonFlags(), ParseCommand() (+11 more)

### Community 96 - "OS Process Execution Tests"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 97 - "Kubernetes Create And Naming Tests"
Cohesion: 0.36
Nodes (8): first(), parseKube(), second(), TestCreatedNames(), TestDerivedObjectNames(), TestSecretsCreateFollowsAnAlias(), TestSecretsCreateRejectsOtherShapes(), TestSecretsCreateSpellings()

### Community 98 - "Download And Flag Validation Tests"
Cohesion: 0.07
Nodes (64): captureStderr(), downloadEnvWithImage(), TestAbsPath(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput() (+56 more)

### Community 99 - "Exec And Status Script Argv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

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

### Community 104 - "Docker Compose Rendering"
Cohesion: 0.12
Nodes (30): Input, Instance, composeEscape(), composeQuote(), Mount, Render(), renderContentConfig(), renderHealthcheck() (+22 more)

### Community 105 - "Status Command Reference"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 106 - "Kubernetes Section Parsing Tests"
Cohesion: 0.29
Nodes (7): applyKubeDefaults(), ParseKubernetes(), TestParseKubernetesError(), TestParseKubernetesFull(), TestParseKubernetesLoggingLibsDefaults(), TestParseKubernetesReplicasDefault(), TestParseKubernetesResources()

### Community 107 - "CLI Package Test Catalog"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 108 - "Application YAML Rendering"
Cohesion: 0.38
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 112 - "Passthrough Property Merging"
Cohesion: 0.12
Nodes (21): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), displayName(), TestAppendPassthroughCollision(), TestDisplayName(), TestFormatScalarQuoting() (+13 more)

### Community 132 - "Shipped Examples And Scan Patterns"
Cohesion: 0.20
Nodes (14): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), TestWorkflowExamplesMatchGoldenSpecs(), isYAML(), matchStar() (+6 more)

### Community 133 - "Image Pull Syslog And Namespace"
Cohesion: 0.10
Nodes (26): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), TestBinderFieldsCarryStablePlaceholders(), TestEnvCredentialsShareOneMountName(), TestGeneratedSecretNamesStayOutOfChildEnvDanger(), TestSecretNameConflictIsRecorded() (+18 more)

## Knowledge Gaps
- **135 isolated node(s):** `9.1 Declaring a credential`, `9.2 Mount names`, `9.3 How each platform delivers them`, `9.4 Registry credentials (pulling the image)`, `All commands` (+130 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 230 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **37 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `progress` connect `Status Progress Reporter` to `Status Report Collection`, `Status Command Action And Watch`, `Status Report Model And Formatting`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **Why does `watchFlag` connect `Status Command Action And Watch` to `Progress Spinner Tests`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 135 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 135 INFERRED edges - model-reasoned connections that need verification._
- **Are the 66 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 66 INFERRED edges - model-reasoned connections that need verification._
- **What connects `9.1 Declaring a credential`, `9.2 Mount names`, `9.3 How each platform delivers them` to the rest of the system?**
  _135 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Status Script And Generate Tests` be split into smaller, more focused modules?**
  _Cohesion score 0.10119047619047619 - nodes in this community are weakly interconnected._
- **Should `Validation Rule Tests` be split into smaller, more focused modules?**
  _Cohesion score 0.05732295873140943 - nodes in this community are weakly interconnected._