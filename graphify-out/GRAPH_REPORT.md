# Graph Report - solace-ibmmq-connector-helper  (2026-10-24)

## Corpus Check
- 132 files · ~417,411 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2594 nodes · 10643 edges · 122 communities (102 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1611 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `1ec5b16c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- runner_test.go
- .step
- validate_extra_test.go
- dockergen_test.go
- gen_extra_test.go
- Cred
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- deref
- completion_test.go
- solmq-conn-util user guide
- helperProcessArgv
- runAction
- deploy_test.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main.go
- issueWith
- withProgressClock
- Defaults
- JavaOptions
- storebuild.go
- statusreport.go
- Cmd
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
- actDocker
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- Project Documentation Files
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- .add
- Application
- instanceSession
- certfixture.go
- abbreviation_doc_test.go
- verbUsage
- maven_test.go
- write
- Podman
- sinceFlag
- parse_test.go
- libs/image_test.go
- Issue
- run
- Dispatch Model Consistency Tests
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- removeNamespace
- Status Command User Guide
- ParseCommand
- resolvePullSecret
- Writer
- 13. Logs: the lines behind the state
- .keystore
- Build
- Runner
- rawDefaults
- go_pkg_io_fs
- maven.go
- GenerateDocker
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- jks.go
- builtinList
- solaceSide
- progress
- SolaceProps
- PodmanDeploy
- Expand
- FormatScalar
- attachRunner
- repeatableName
- dispatch
- tailFlag
- Secrets Model User Guide
- Env
- Command Reference Overview
- Project Overview And Quick Start
- cmd/solmq-conn-util
- hostReader
- validate.go
- support_test.go
- yw
- 10. `download jar`
- TestCreatedStoreWarningsSurface
- createdStoreDockerSpec
- podmanDeploy
- GenerateKubernetes
- buildLeaderElection
- TestRetiredCreateNamesNeedAKey
- Validate
- javaCtx
- LogsArgv
- .Write

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 146 edges
2. `hasErr()` - 98 edges
3. `captureStderr()` - 75 edges
4. `Build()` - 72 edges
5. `write()` - 69 edges
6. `wfOK()` - 59 edges
7. `Download()` - 58 edges
8. `wf()` - 49 edges
9. `imageOK()` - 47 edges
10. `Defaults` - 46 edges

## Surprising Connections (you probably didn't know these)
- `9.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/userguide.md → internal/tls/tls_test.go
- `Design notes` --references--> `resolveOneInstance()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/instances.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Build()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/consolidate/consolidate.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Validate()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go
- `Design notes` --references--> `PodmanFileSecretNames()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (122 total, 20 thin omitted)

### Community 0 - "runner_test.go"
Cohesion: 0.08
Nodes (39): canIVerb(), ExecArgv(), InstallScript(), PodmanSecretRemove(), Preflight(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform() (+31 more)

### Community 1 - ".step"
Cohesion: 0.17
Nodes (13): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), MergeService(), ObjectExists(), ParseDeployment() (+5 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.08
Nodes (38): Input, Instance, composeEscape(), composeQuote(), flowList(), CreatedStore, Render(), renderContentConfig() (+30 more)

### Community 4 - "gen_extra_test.go"
Cohesion: 0.14
Nodes (32): Config(), issuesContain(), synthWorkflowFiles(), synthWorkflows(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder() (+24 more)

### Community 5 - "Cred"
Cohesion: 0.09
Nodes (17): fixedLeaderNames(), Image, Cred, Side, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), TestUserSecretLiteralAndEnv(), placeholderSecretRef() (+9 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.31
Nodes (10): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), isFalse(), isLiteral(), nearestKnownKey() (+2 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.10
Nodes (21): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+13 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.12
Nodes (53): connDefaults(), TestBinderIdentityUsesTheCredentialPair(), TestCheckSideMQMissingFields(), TestCheckSideSolaceMissingAndBadScheme(), TestConnRefSolaceTopicSourceIsError(), TestConnRefStrictOnlyDestination(), TestConnRefUnknownAndSystemMismatch(), TestConnRefValidResolvesNoError() (+45 more)

### Community 11 - "scan_test.go"
Cohesion: 0.07
Nodes (55): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), goldenMounted(), itoa() (+47 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (108): skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestMissingStoreFiles(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits(), TestUnreadableMountsNamesWhatTheConnectorCannotRead(), writeMode(), mustWrite() (+100 more)

### Community 14 - "deref"
Cohesion: 0.14
Nodes (19): checkCreateBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), StoreSources, isNull(), TestLeaderElectionEffectiveMode(), deref() (+11 more)

### Community 15 - "completion_test.go"
Cohesion: 0.13
Nodes (27): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+19 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.13
Nodes (15): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas, 1.1 Shell completion (+7 more)

### Community 18 - "helperProcessArgv"
Cohesion: 0.11
Nodes (26): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+18 more)

### Community 19 - "runAction"
Cohesion: 0.23
Nodes (24): runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), envFlag(), flagExit(), noPromptFlag(), outFlag() (+16 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.08
Nodes (59): Input, Instance, KV, storeItem, envValueQuote(), PullSecret, StoreFile, leaderMode() (+51 more)

### Community 23 - "runner_test.go"
Cohesion: 0.16
Nodes (41): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+33 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.11
Nodes (24): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv() (+16 more)

### Community 25 - "main.go"
Cohesion: 0.09
Nodes (19): absPath(), absResolver(), confirmRemove(), downloadDeployedImage(), envPairs(), envVarMax(), fileReader(), loadEnvFile() (+11 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - "withProgressClock"
Cohesion: 0.17
Nodes (17): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow() (+9 more)

### Community 28 - "Defaults"
Cohesion: 0.14
Nodes (28): Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost(), leaderConnRef() (+20 more)

### Community 29 - "JavaOptions"
Cohesion: 0.34
Nodes (11): YAMLKind(), checkContentType(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem(), checkTransforms() (+3 more)

### Community 30 - "storebuild.go"
Cohesion: 0.10
Nodes (26): params, certAlias(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard() (+18 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (39): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+31 more)

### Community 32 - "Cmd"
Cohesion: 0.17
Nodes (10): applyCmdEnv(), applyCmdInput(), Attacher, Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner (+2 more)

### Community 43 - "actDocker"
Cohesion: 0.26
Nodes (21): actDocker(), actKubernetes(), actPodman(), emit(), errExit(), failFast(), genConfig(), genDocker() (+13 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (32): CA(), KeyPKCS8(), NotCA(), PEM(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases(), chain() (+24 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.23
Nodes (15): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot() (+7 more)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (55): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+47 more)

### Community 47 - "Project Documentation Files"
Cohesion: 0.44
Nodes (3): Build, Release (CI), solmq-conn-util -- development guide

### Community 48 - "Kubernetes"
Cohesion: 0.14
Nodes (13): 11. What gets generated, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, Kubernetes, Resources (+5 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.11
Nodes (18): cli, Command details, deploy, download, examples, generate, help, logs (+10 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.13
Nodes (21): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+13 more)

### Community 52 - ".add"
Cohesion: 0.17
Nodes (16): Store, TLSConfig, TestTLSConfigStoreLookup(), checkCreatedSelection(), checkStoreDefinition(), createdRole(), checkMisplacedTransform(), checkConnections() (+8 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (36): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+28 more)

### Community 55 - "instanceSession"
Cohesion: 0.12
Nodes (33): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+25 more)

### Community 56 - "certfixture.go"
Cohesion: 0.18
Nodes (16): Key, Option, Trusted, AsCA(), DER(), Cert, Issue(), JKS() (+8 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.19
Nodes (19): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+11 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.11
Nodes (43): nsList(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile(), TestDeployDockerSeamChildEnvCarriesCredentials() (+35 more)

### Community 61 - "Podman"
Cohesion: 0.08
Nodes (27): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), JavaOptions, decodeCreate(), Secrets (+19 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (27): ApplyStats(), ApplyTop(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys(), TestApplyStats() (+19 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.13
Nodes (26): connectorRelease(), imageNameTag(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable() (+18 more)

### Community 65 - "Issue"
Cohesion: 0.21
Nodes (14): PodmanFile, StoreBuild, b64(), TestResolvePodmanFiles(), TestResolveStores(), Resolver, ResolvePodmanFiles(), resolveStores() (+6 more)

### Community 67 - "run"
Cohesion: 0.13
Nodes (19): resolveTarget(), run(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+11 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "removeNamespace"
Cohesion: 0.16
Nodes (14): nsItemJSON(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceProbeArgvIsOneQuery(), TestRemoveNamespaceRejectsAnUnsafeCommand() (+6 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "ParseCommand"
Cohesion: 0.18
Nodes (11): Docker(), Kubernetes(), kubeVerb(), ParseCommand(), PodmanSecretCreate(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand() (+3 more)

### Community 75 - "resolvePullSecret"
Cohesion: 0.32
Nodes (8): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret()

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - ".keystore"
Cohesion: 0.18
Nodes (17): 5.1 Variable expansion (`${VAR}`), 5. The config file and workflow discovery, 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`) (+9 more)

### Community 79 - "Build"
Cohesion: 0.08
Nodes (31): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers() (+23 more)

### Community 80 - "Runner"
Cohesion: 0.10
Nodes (31): checkImageCommand(), Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), DockerImageCommand(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList() (+23 more)

### Community 81 - "rawDefaults"
Cohesion: 0.17
Nodes (15): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), rawDefaults, rawLeader, rawManagement (+7 more)

### Community 82 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "GenerateDocker"
Cohesion: 0.12
Nodes (24): DockerPlan, DockerStore, mount, StoreUse, Mount, TestTargetMounts(), GenerateDocker(), HostMount (+16 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (20): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+12 more)

### Community 88 - "jks.go"
Cohesion: 0.22
Nodes (7): jksCert(), joinAliases(), readJKS(), sniff(), TestSniff(), format, jksKey

### Community 89 - "builtinList"
Cohesion: 0.15
Nodes (17): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+9 more)

### Community 90 - "solaceSide"
Cohesion: 0.21
Nodes (21): TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+13 more)

### Community 91 - "progress"
Cohesion: 0.13
Nodes (10): newProgress(), progressTrim(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), progress, progressMode (+2 more)

### Community 92 - "SolaceProps"
Cohesion: 0.23
Nodes (16): SolaceProps(), StorePath(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenNothingSelected() (+8 more)

### Community 93 - "PodmanDeploy"
Cohesion: 0.25
Nodes (10): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeploySystemModeNoUserFlag(), TestPodmanRemoveStopFailureIsReported() (+2 more)

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (10): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), needsQuote(), nodeToProps() (+2 more)

### Community 96 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 98 - "dispatch"
Cohesion: 0.05
Nodes (125): dispatch(), captureStderr(), captureStdout(), containsToken(), downloadEnvWithImage(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError() (+117 more)

### Community 100 - "Secrets Model User Guide"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), PodmanServiceName(), Env (+1 more)

### Community 102 - "Command Reference Overview"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "Project Overview And Quick Start"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 105 - "hostReader"
Cohesion: 0.33
Nodes (5): rootlessIDs(), unreadableMounts(), hostReader, mountCheck, mountKind

### Community 106 - "validate.go"
Cohesion: 0.11
Nodes (40): checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker(), checkImage(), checkImagePull() (+32 more)

### Community 107 - "support_test.go"
Cohesion: 0.53
Nodes (5): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 110 - "TestCreatedStoreWarningsSurface"
Cohesion: 0.36
Nodes (8): Validity(), createdCA(), hasIssue(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestCreatedStoreWarningsSurface(), TestResolveDockerStoresLimits()

### Community 113 - "createdStoreDockerSpec"
Cohesion: 0.50
Nodes (5): createdStoreDockerSpec(), envValue(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 114 - "podmanDeploy"
Cohesion: 0.20
Nodes (16): podmanDeploy(), podmanRemove(), previousStoreSecrets(), SecretRef, TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestResolveCredentials() (+8 more)

### Community 115 - "GenerateKubernetes"
Cohesion: 0.15
Nodes (23): File, KubeOpts, build(), synthTLSWorkflowFiles(), TestGenerateJavaOptionsReachEveryPlatform(), TestGeneratePodmanNeverReadsTheStores(), TestGenValidateStoresWarning(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless() (+15 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.14
Nodes (18): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), dropManaged(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision() (+10 more)

### Community 117 - "TestRetiredCreateNamesNeedAKey"
Cohesion: 0.50
Nodes (4): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate()

### Community 118 - "Validate"
Cohesion: 0.53
Nodes (6): shippedExamplesRequest(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 119 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 121 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 131 - ".Write"
Cohesion: 0.33
Nodes (6): How the suite is built, TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 205 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `Secrets Model User Guide`, `Status Command User Guide`, `13. Logs: the lines behind the state`, `10. `download jar``, `Project Documentation Files`, `Kubernetes`, `.keystore`, `builtinList`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 141 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 141 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `runner_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07665505226480836 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util test catalogue` connect `Internal Package Test Catalog` to `cmd/solmq-conn-util`, `.Write`, `Project Documentation Files`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 71 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 71 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11634615384615385 - nodes in this community are weakly interconnected._