# Graph Report - solace-ibmmq-connector-helper  (2026-10-11)

## Corpus Check
- 121 files · ~379,317 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 11 file(s) not represented in the graph (top: (none) 4, .list 3, .jks 2)

## Summary
- 2354 nodes · 9229 edges · 136 communities (95 shown, 41 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1321 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dea01f9a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ExecArgv
- testing.T
- validate_extra_test.go
- hostReader
- synthWorkflowFiles
- Cred
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- GenerateKubernetes
- scan_test.go
- completion.go
- statusscript_test.go
- Workload
- completion_test.go
- User Guide Core Chapters
- watchFlag
- main_test.go
- deploy_test.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- runAction
- progress
- withProgressClock
- Run
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
- Defaults
- Runner
- Development Guide
- libs_test.go
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- actLogs
- Application
- instanceSession
- Build
- abbreviation_doc_test.go
- verbUsage
- maven_test.go
- write
- Docker
- sinceFlag
- parse_test.go
- libs/image_test.go
- parse
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
- solaceSide
- TestParsingHelpersIgnoreAWarningOnStderr
- allowCommandValue
- flattenBlock
- maven.go
- resolvePullSecret
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- binderNames
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- buildLeaderElection
- helperProcessArgv
- filenameFromEscapedPath
- Cmd
- Expand
- GeneratePodman
- MountPath
- repeatableName
- dispatch
- tailFlag
- Secrets Model User Guide
- Known And Managed Keys
- Command Reference Overview
- Project Overview And Quick Start
- dockergen_test.go
- Status Command Reference
- validate.go
- CLI Package Test Catalog
- yw
- statusreport_test.go
- FormatScalar
- Preflight
- .Write
- Input
- GenerateDocker
- shippedExamplesRequest
- Env

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 142 edges
2. `hasErr()` - 92 edges
3. `captureStderr()` - 73 edges
4. `write()` - 67 edges
5. `Build()` - 63 edges
6. `Download()` - 58 edges
7. `wfOK()` - 58 edges
8. `imageOK()` - 48 edges
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

## Communities (136 total, 41 thin omitted)

### Community 0 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 1 - "testing.T"
Cohesion: 0.04
Nodes (86): TestAbsPath(), TestIsIndex(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestSortIsByNameThenNamespace(), TestWatchFlagAcceptsBareAndInterval(), TestWorkflowExamplesMatchGoldenSpecs() (+78 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.05
Nodes (147): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+139 more)

### Community 3 - "hostReader"
Cohesion: 0.09
Nodes (20): fileOwner(), fileOwner(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits() (+12 more)

### Community 4 - "synthWorkflowFiles"
Cohesion: 0.24
Nodes (17): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+9 more)

### Community 5 - "Cred"
Cohesion: 0.11
Nodes (9): Image, Cred, SolaceProps(), placeholderSecretRef(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseFixedMountPaths() (+1 more)

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

### Community 10 - "GenerateKubernetes"
Cohesion: 0.18
Nodes (12): KubeOpts, StoreFile, b64(), TestGenerateJavaOptionsReachEveryPlatform(), TestGenerateKubernetesCarriesStatusScript(), TestResolveStores(), TestStoresSharingAFileNameMountApart(), GenerateKubernetes() (+4 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.19
Nodes (29): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+21 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.08
Nodes (47): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+39 more)

### Community 14 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "User Guide Core Chapters"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 19 - "main_test.go"
Cohesion: 0.06
Nodes (67): main(), run(), captureStdout(), containsToken(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList() (+59 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.08
Nodes (58): Input, Instance, KV, storeItem, envValueQuote(), PullSecret, leaderMode(), ManagementPort() (+50 more)

### Community 23 - "runner_test.go"
Cohesion: 0.17
Nodes (32): currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough(), countDocShape(), countTestFuncs() (+24 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.14
Nodes (19): leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput(), syslogInput() (+11 more)

### Community 25 - "runAction"
Cohesion: 0.14
Nodes (36): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage(), envFlag() (+28 more)

### Community 26 - "progress"
Cohesion: 0.23
Nodes (4): newProgress(), progressTrim(), progress, progressMode

### Community 27 - "withProgressClock"
Cohesion: 0.16
Nodes (16): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Run"
Cohesion: 0.14
Nodes (21): synthWorkflows(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), Workflow, checkAllExtraKeys(), propPrefix() (+13 more)

### Community 29 - "JavaOptions"
Cohesion: 0.26
Nodes (13): JavaOptions, YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders() (+5 more)

### Community 30 - "podmanDeploy"
Cohesion: 0.10
Nodes (27): podmanDeploy(), podmanRemove(), TestPodmanFileSecretNames(), PodmanFileSecretNames(), PodmanSecretStoreName(), QuadletScope, PodmanDeploy(), PodmanRemove() (+19 more)

### Community 31 - "statusreport.go"
Cohesion: 0.11
Nodes (28): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Banner(), banner(), Bytes() (+20 more)

### Community 32 - "statusCollector"
Cohesion: 0.19
Nodes (10): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), KubernetesGetJSON(), TestKubernetesGetJSONArgv(), TestKubernetesGetJSONMissingObjectIsAnError() (+2 more)

### Community 43 - "Defaults"
Cohesion: 0.07
Nodes (27): fixedLeaderNames(), defaultsFromRaw(), Defaults, Security, TLSConfig, applyDest(), rawMQ, rawSolace (+19 more)

### Community 44 - "Runner"
Cohesion: 0.15
Nodes (32): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+24 more)

### Community 45 - "Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 48 - "Kubernetes"
Cohesion: 0.25
Nodes (6): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes

### Community 49 - "Command Reference Details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.16
Nodes (17): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+9 more)

### Community 52 - "actLogs"
Cohesion: 0.22
Nodes (12): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), LogsOpts, LogsArgv(), logsCommonFlags() (+4 more)

### Community 54 - "Application"
Cohesion: 0.10
Nodes (32): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+24 more)

### Community 55 - "instanceSession"
Cohesion: 0.11
Nodes (30): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+22 more)

### Community 56 - "Build"
Cohesion: 0.09
Nodes (26): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers() (+18 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.12
Nodes (24): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+16 more)

### Community 58 - "verbUsage"
Cohesion: 0.17
Nodes (22): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+14 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.13
Nodes (37): podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+29 more)

### Community 61 - "Docker"
Cohesion: 0.08
Nodes (25): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), decodeCreate(), Secrets, Service (+17 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (29): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+21 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.10
Nodes (34): builtinList(), connectorRelease(), imageNameTag(), imageSatisfies(), listRanges(), loadImageLibs(), omitListProvenance(), releaseTag() (+26 more)

### Community 65 - "parse"
Cohesion: 0.18
Nodes (12): File, SecretConflict, build(), TestConfigRejectsSecretNameConflict(), TestGenValidateStoresWarning(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate(), TestValidateCleanSpecStillPasses() (+4 more)

### Community 67 - "Resolver"
Cohesion: 0.20
Nodes (9): PodmanFile, TestResolvePodmanFiles(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError(), Resolver, ResolvePodmanFiles() (+1 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.22
Nodes (17): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+9 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.16
Nodes (17): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+9 more)

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

### Community 79 - "solaceSide"
Cohesion: 0.19
Nodes (19): propsNode(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), testSecretRef(), trustStoreVal(), binderOfKind(), extraWF() (+11 more)

### Community 80 - "TestParsingHelpersIgnoreAWarningOnStderr"
Cohesion: 0.10
Nodes (21): EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesListJSON(), KubernetesPodsJSON(), KubernetesTop(), contains() (+13 more)

### Community 82 - "flattenBlock"
Cohesion: 0.24
Nodes (13): envTransforms(), deref(), extraKeys(), flattenBlock(), rawMQ, rawSolace, isMergeKey(), appendTransformKeys() (+5 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "resolvePullSecret"
Cohesion: 0.32
Nodes (8): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret()

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.08
Nodes (25): reportDownload(), DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5(), defaultClient(), downloadOne() (+17 more)

### Community 88 - "binderNames"
Cohesion: 0.36
Nodes (8): binderNames(), binderOf(), eqStrs(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder(), TestBinderFieldsCarryStablePlaceholders()

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.29
Nodes (7): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`), Connector 2.x and 3.x

### Community 90 - "buildLeaderElection"
Cohesion: 0.18
Nodes (13): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), TestSanitizeAndIsTCPS(), TestOverlayExtras(), isTCPS(), overlayExtras() (+5 more)

### Community 91 - "helperProcessArgv"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 92 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 93 - "Cmd"
Cohesion: 0.15
Nodes (9): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS (+1 more)

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (18): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+10 more)

### Community 95 - "GeneratePodman"
Cohesion: 0.18
Nodes (12): TestGeneratePodmanCarriesLogbackOnlyWithSyslog(), TestGeneratePodmanCarriesStatusScript(), TestGeneratePodmanIgnoresBaseDir(), TestGeneratePodmanNeverReadsTheStores(), TestGeneratePodmanNoModeKeyIsClean(), TestGeneratePodmanQuadlet(), TestToIssues(), GeneratePodman() (+4 more)

### Community 96 - "MountPath"
Cohesion: 0.50
Nodes (5): MountPath(), StorePath(), TestMountPathIsFixedPerRole(), TestStorePathConfigVsDeploy(), Role

### Community 98 - "dispatch"
Cohesion: 0.05
Nodes (80): dispatch(), captureStderr(), downloadEnvWithImage(), TestAllowCommandFlagBadValueExitsUsageError(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+72 more)

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
Cohesion: 0.09
Nodes (36): Input, Instance, renderConfigMap(), composeEscape(), composeQuote(), Render(), renderContentConfig(), renderHealthcheck() (+28 more)

### Community 105 - "Status Command Reference"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 106 - "validate.go"
Cohesion: 0.12
Nodes (42): checkConnections(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker(), checkImage() (+34 more)

### Community 107 - "CLI Package Test Catalog"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "statusreport_test.go"
Cohesion: 0.13
Nodes (13): ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestImageMismatch(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues() (+5 more)

### Community 112 - "FormatScalar"
Cohesion: 0.20
Nodes (11): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), needsQuote() (+3 more)

### Community 113 - "Preflight"
Cohesion: 0.20
Nodes (10): canIVerb(), Preflight(), TestPreflightDockerArgvIsInfo(), TestPreflightExtraAllowedThreadsThrough(), TestPreflightFailureWrapsPlatformHint(), TestPreflightKubernetesArgvDeployNoNamespace(), TestPreflightKubernetesArgvRemoveWithNamespace(), TestPreflightPodmanArgvIsInfo() (+2 more)

### Community 131 - ".Write"
Cohesion: 0.29
Nodes (7): mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 132 - "Input"
Cohesion: 0.29
Nodes (7): podmanSecretRefs(), toPodmanMount(), Mount, SecretRef, FileSecret, Input, Instance

### Community 133 - "GenerateDocker"
Cohesion: 0.15
Nodes (19): DockerPlan, mount, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestPlansCarryTheExpandedHostPathsTheyMount(), TestResolveCredentials() (+11 more)

### Community 134 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 135 - "Env"
Cohesion: 0.27
Nodes (8): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), removeTarget(), PodmanServiceName(), Env, instanceRequest

## Knowledge Gaps
- **136 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `renderedTransform`, `solaceBlock` (+131 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 233 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **41 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Cred` connect `Cred` to `validate.go`, `Defaults`, `Model`, `runner_test.go`, `buildLeaderElection`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Why does `progress` connect `progress` to `statusCollector`, `instanceFromInspect`, `instanceSession`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **Why does `sinceFlag` connect `sinceFlag` to `runner_test.go`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 137 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 137 INFERRED edges - model-reasoned connections that need verification._
- **Are the 69 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 69 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _136 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ExecArgv` be split into smaller, more focused modules?**
  _Cohesion score 0.1111111111111111 - nodes in this community are weakly interconnected._