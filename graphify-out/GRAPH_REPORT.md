# Graph Report - solace-ibmmq-connector-helper  (2026-10-23)

## Corpus Check
- 123 files · ~391,803 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 11 file(s) not represented in the graph (top: (none) 4, .list 3, .jks 2)

## Summary
- 2381 nodes · 9659 edges · 117 communities (97 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1481 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f55bfbb5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- runner_test.go
- flattenBlock
- validate_extra_test.go
- hostReader
- synthWorkflowFiles
- Cred
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- SolaceProps
- completion_test.go
- User Guide Core Chapters
- validate_test.go
- main_test.go
- deploy_test.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- runAction
- issueWith
- withProgressClock
- Defaults
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
- The spec generator (`solmq-conn-util-generator.html`)
- Runner
- testcatalog_test.go
- libs_test.go
- Project Documentation Files
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- Side
- Application
- instanceSession
- binderNames
- abbreviation_doc_test.go
- verbUsage
- maven_test.go
- write
- Docker
- sinceFlag
- parse_test.go
- libs/image_test.go
- GenerateKubernetes
- resolveStatusPassword
- Dispatch Model Consistency Tests
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- nsList
- Status Command User Guide
- Runner
- Download Jar User Guide
- Writer
- ExecArgv
- Workflow File User Guide
- solaceSide
- runner.go
- main.go
- go_pkg_io_fs
- maven.go
- resolvePullSecret
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- attachRunner
- builtinList
- buildLeaderElection
- .step
- filenameFromEscapedPath
- Cmd
- Expand
- FormatScalar
- targetMounts
- repeatableName
- dispatch
- tailFlag
- Secrets Model User Guide
- Known And Managed Keys
- Command Reference Overview
- Project Overview And Quick Start
- cmd/solmq-conn-util
- TestSanitizeAndIsTCPS
- validate.go
- support_test.go
- yw
- TestCredentialsNotWiredWarning
- javaCtx
- .Write
- GeneratePodman
- GenerateDocker
- Env

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 142 edges
2. `hasErr()` - 97 edges
3. `captureStderr()` - 73 edges
4. `Build()` - 70 edges
5. `write()` - 67 edges
6. `wfOK()` - 59 edges
7. `Download()` - 58 edges
8. `wf()` - 48 edges
9. `imageOK()` - 47 edges
10. `Defaults` - 46 edges

## Surprising Connections (you probably didn't know these)
- `Build` --references--> `build()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go
- `Design notes` --references--> `resolveOneInstance()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/instances.go
- `Testing` --references--> `TestTestCatalogSnapshotInSync()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/testcatalog_test.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Build()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/consolidate/consolidate.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Validate()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (117 total, 20 thin omitted)

### Community 0 - "runner_test.go"
Cohesion: 0.07
Nodes (53): canIVerb(), PodmanSecretCreate(), Preflight(), ScriptInstalled(), attachFiles(), helperProcessArgv(), readFile(), TestLogsArgvPerPlatform() (+45 more)

### Community 1 - "flattenBlock"
Cohesion: 0.30
Nodes (11): envTransforms(), deref(), extraKeys(), flattenBlock(), isMergeKey(), appendTransformKeys(), documentMapping(), mappingChild() (+3 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (61): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+53 more)

### Community 3 - "hostReader"
Cohesion: 0.19
Nodes (12): rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits(), TestUnreadableMountsNamesWhatTheConnectorCannotRead(), writeMode() (+4 more)

### Community 4 - "synthWorkflowFiles"
Cohesion: 0.20
Nodes (21): File, issuesContain(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder() (+13 more)

### Community 5 - "Cred"
Cohesion: 0.11
Nodes (9): Security, Image, Cred, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), TestUserSecretLiteralAndEnv(), placeholderSecretRef(), rawSecurity (+1 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.31
Nodes (10): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), isFalse(), isLiteral(), nearestKnownKey() (+2 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.11
Nodes (19): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+11 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.13
Nodes (44): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+36 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.19
Nodes (29): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+21 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (90): Render(), minimalDockerInput(), TestAppYAMLSecretPlaceholdersAreNotInterpolated(), TestContentEscapesDollarsForCompose(), TestContentIndentationAndBlankLines(), TestEnvironmentBranches(), TestHealthcheckRunsTheStatusScript(), TestJavaOptionsEnvironment() (+82 more)

### Community 14 - "SolaceProps"
Cohesion: 0.10
Nodes (36): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), goldenMounted(), itoa() (+28 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "User Guide Core Chapters"
Cohesion: 0.09
Nodes (23): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state, 14.1 One instance per run (+15 more)

### Community 18 - "validate_test.go"
Cohesion: 0.23
Nodes (19): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), defsWithStores(), leWorkflows(), mgmtSide(), TestDeployNameTooLong() (+11 more)

### Community 19 - "main_test.go"
Cohesion: 0.07
Nodes (56): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+48 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.05
Nodes (76): Input, Instance, KV, storeItem, Input, Instance, envValueQuote(), PullSecret (+68 more)

### Community 23 - "runner_test.go"
Cohesion: 0.14
Nodes (54): isClusterDefault(), isOurs(), namespaceOccupants(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles() (+46 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "runAction"
Cohesion: 0.20
Nodes (28): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), downloadDeployedImage(), envFlag() (+20 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - "withProgressClock"
Cohesion: 0.24
Nodes (12): newPipeReader(), progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+4 more)

### Community 28 - "Defaults"
Cohesion: 0.19
Nodes (20): leaderSession(), synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession() (+12 more)

### Community 29 - "JavaOptions"
Cohesion: 0.29
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 30 - "podmanDeploy"
Cohesion: 0.24
Nodes (11): podmanRemove(), QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeploySystemModeNoUserFlag() (+3 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (40): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+32 more)

### Community 32 - "statusCollector"
Cohesion: 0.16
Nodes (14): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), PodmanServiceName(), MergeService(), ObjectExists() (+6 more)

### Community 43 - "The spec generator (`solmq-conn-util-generator.html`)"
Cohesion: 0.15
Nodes (15): defaultsFromRaw(), TestLeaderElectionEffectiveMode(), applyDest(), nodePtr(), presentBlock(), LeaderElection, rawDefaults, rawLeader (+7 more)

### Community 44 - "Runner"
Cohesion: 0.31
Nodes (19): actDocker(), actKubernetes(), actPodman(), emit(), errExit(), failFast(), genConfig(), genDocker() (+11 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (51): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+43 more)

### Community 47 - "Project Documentation Files"
Cohesion: 0.44
Nodes (3): Build, Release (CI), solmq-conn-util -- development guide

### Community 48 - "Kubernetes"
Cohesion: 0.17
Nodes (11): ownedNames(), 11. What gets generated, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, Kubernetes (+3 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.11
Nodes (18): cli, Command details, deploy, download, examples, generate, help, logs (+10 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.12
Nodes (21): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+13 more)

### Community 52 - "Side"
Cohesion: 0.15
Nodes (14): fixedLeaderNames(), Store, TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+6 more)

### Community 54 - "Application"
Cohesion: 0.10
Nodes (35): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+27 more)

### Community 55 - "instanceSession"
Cohesion: 0.13
Nodes (31): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+23 more)

### Community 56 - "binderNames"
Cohesion: 0.36
Nodes (9): binderNames(), binderOf(), eqStrs(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder() (+1 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.17
Nodes (18): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+10 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.07
Nodes (60): captureStderr(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials() (+52 more)

### Community 61 - "Docker"
Cohesion: 0.09
Nodes (23): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), JavaOptions, decodeCreate(), Secrets, TestCredCreateRemovedKeys(), applyDockerDefaults() (+15 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (29): ApplyStats(), ApplyTop(), EngineNamesByImage(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+21 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.12
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 65 - "GenerateKubernetes"
Cohesion: 0.17
Nodes (19): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), build(), Config(), TestConfigRejectsSecretNameConflict() (+11 more)

### Community 67 - "resolveStatusPassword"
Cohesion: 0.40
Nodes (5): TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError(), resolveStatusPassword()

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.25
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.19
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+6 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Runner"
Cohesion: 0.14
Nodes (17): confirmNamespaceRemoval(), removeNamespace(), Docker(), Runner, Kubernetes(), KubernetesListJSON(), kubeVerb(), ParseCommand() (+9 more)

### Community 75 - "Download Jar User Guide"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "ExecArgv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

### Community 78 - "Workflow File User Guide"
Cohesion: 0.22
Nodes (9): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file (+1 more)

### Community 79 - "solaceSide"
Cohesion: 0.11
Nodes (35): Opts, applyStatusAccess(), Build(), containsSub(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestApplyStatusAccessNoOperatorUsers() (+27 more)

### Community 80 - "runner.go"
Cohesion: 0.09
Nodes (31): Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Attacher (+23 more)

### Community 81 - "main.go"
Cohesion: 0.09
Nodes (18): absPath(), absResolver(), confirmRemove(), envPairs(), fileReader(), main(), platformSpellings(), promptPlatformMenu() (+10 more)

### Community 82 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "resolvePullSecret"
Cohesion: 0.17
Nodes (15): KubeOpts, dockerConfigJSON(), TestGenerateJavaOptionsReachEveryPlatform(), TestGenerateKubernetesCarriesStatusScript(), TestResolveCredentials(), TestRetiredCreateNamesDeployButDoNotValidate(), GenerateKubernetes(), ResolveCredentials() (+7 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.08
Nodes (24): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5(), defaultClient(), downloadOne(), downloadWithVerification() (+16 more)

### Community 88 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 89 - "builtinList"
Cohesion: 0.15
Nodes (17): 15. Notes and gotchas, 7.1 The reserved status account (`solmq-status`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+9 more)

### Community 90 - "buildLeaderElection"
Cohesion: 0.15
Nodes (18): leaderNameFn, secretFn, appendPassthrough(), binderOwner(), buildBundle(), buildLeaderElection(), propsNode(), TestAppendPassthroughCollision() (+10 more)

### Community 91 - ".step"
Cohesion: 0.13
Nodes (11): newProgress(), progressElapsed(), progressTrim(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), progress (+3 more)

### Community 92 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 93 - "Cmd"
Cohesion: 0.23
Nodes (8): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (18): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+10 more)

### Community 95 - "FormatScalar"
Cohesion: 0.18
Nodes (11): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), needsQuote() (+3 more)

### Community 96 - "targetMounts"
Cohesion: 0.27
Nodes (10): StoreUse, b64(), TestResolveStores(), TestStoresSharingAFileNameMountApart(), TestTargetMounts(), resolveStores(), StoreField(), targetMounts() (+2 more)

### Community 98 - "dispatch"
Cohesion: 0.06
Nodes (68): run(), downloadEnvWithImage(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput() (+60 more)

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

### Community 104 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 105 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 106 - "validate.go"
Cohesion: 0.10
Nodes (53): checkConnections(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker(), checkDuplicateSources() (+45 more)

### Community 107 - "support_test.go"
Cohesion: 0.31
Nodes (9): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs() (+1 more)

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "TestCredentialsNotWiredWarning"
Cohesion: 0.67
Nodes (4): kubeNoSecrets(), TestCredentialsNotWiredWarning(), TestCredentialsWiredNoWarning(), vSolacePlainTCP()

### Community 110 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 131 - ".Write"
Cohesion: 0.33
Nodes (6): How the suite is built, TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 132 - "GeneratePodman"
Cohesion: 0.13
Nodes (23): podmanDeploy(), previousStoreSecrets(), PodmanFile, TestGeneratePodmanCarriesLogbackOnlyWithSyslog(), TestGeneratePodmanCarriesStatusScript(), TestGeneratePodmanNoModeKeyIsClean(), TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames() (+15 more)

### Community 133 - "GenerateDocker"
Cohesion: 0.13
Nodes (20): DockerPlan, mount, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), GenerateDocker(), HostMount (+12 more)

### Community 135 - "Env"
Cohesion: 0.33
Nodes (8): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), Env, instanceRequest

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 191 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `User Guide Core Chapters` to `Secrets Model User Guide`, `Status Command User Guide`, `Download Jar User Guide`, `Workflow File User Guide`, `Project Documentation Files`, `Kubernetes`, `builtinList`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Are the 137 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 137 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `runner_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07017543859649122 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util -- development guide` connect `Project Documentation Files` to `runner.go`, `support_test.go`, `SolaceProps`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 71 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 71 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.12215758857747223 - nodes in this community are weakly interconnected._