# Graph Report - solace-ibmmq-connector-helper  (2026-11-05)

## Corpus Check
- 142 files · ~431,758 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2752 nodes · 11466 edges · 129 communities (106 shown, 23 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1727 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f0f2088f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- attachFiles
- .step
- validate_extra_test.go
- dockergen_test.go
- GenerateKubernetes
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- deref
- completion_test.go
- removeNamespace
- crypto/x509.Certificate
- runAction
- Cmd
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- run
- .keystore
- statusreport.go
- TestGoldenKubernetesNoSecrets
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
- Runner
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- rawDefaults
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- .add
- Application
- instanceSession
- certfixture.go
- renderCommandsDoc
- verbUsage
- maven_test.go
- write
- Podman
- sinceFlag
- parse_test.go
- attachRunner
- Resolver
- ScriptInstalled
- Image
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- Runner
- solmq-conn-util configuration reference
- Store
- decodeCreate
- Writer
- support_test.go
- libs/image.go
- buildLeaderElection
- helperProcessArgv
- createdStoreDockerSpec
- Resource
- maven.go
- .Username
- solmq-conn-util operations reference
- libs.go
- tlsSelectingWF
- .name
- solaceSide
- TestApplyStatusAccessAppendsAfterExistingUsers
- SolaceProps
- shellContentInfo
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- GenerateDocker
- .RunSplit
- .Write
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- examples.go
- LeaderElection
- javaCtx
- Env
- GeneratePodman
- defsWithStores
- CA
- Build
- solmq-conn-util abbreviations
- misplacedEnvTransforms
- LogsArgv
- flattenBlock
- cmd/solmq-conn-util
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- TLSConfig
- TestGeneratorPageKnownKeysInSync
- Security
- SafeToken

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 154 edges
2. `hasErr()` - 104 edges
3. `captureStderr()` - 82 edges
4. `write()` - 76 edges
5. `Build()` - 73 edges
6. `wfOK()` - 63 edges
7. `Download()` - 58 edges
8. `Runner` - 53 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `TLS` --references--> `existing()`  [INFERRED]
  docs/troubleshooting.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestCompletionModelMetadataComplete()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (129 total, 23 thin omitted)

### Community 0 - "attachFiles"
Cohesion: 0.31
Nodes (9): attachFiles(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv(), TestOSAttachReportsTheChildExitStatus() (+1 more)

### Community 1 - ".step"
Cohesion: 0.10
Nodes (20): sortInstances(), newProgress(), progressTrim(), actStatus(), clearScreen(), confirmInstall(), instanceNames(), markMissing() (+12 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (65): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+57 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (36): Input, Instance, composeEscape(), composeQuote(), flowList(), Render(), renderContentConfig(), renderHealthcheck() (+28 more)

### Community 4 - "GenerateKubernetes"
Cohesion: 0.10
Nodes (32): File, KubeOpts, b64(), build(), synthTLSWorkflowFiles(), TestGeneratePodmanNeverReadsTheStores(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate() (+24 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.07
Nodes (67): storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), ManagementPort(), nameOnlySecret() (+59 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.26
Nodes (12): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+4 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (21): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+13 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.11
Nodes (45): fixedLeaderNames(), Side, postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide() (+37 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (93): TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestWorkflowExamplesMatchGoldenSpecs(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), canIDenied(), Preflight(), contains(), TestEngineListArgv() (+85 more)

### Community 14 - "deref"
Cohesion: 0.61
Nodes (6): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), isNull(), deref()

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "removeNamespace"
Cohesion: 0.15
Nodes (17): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveKubernetesSeamHappyPath(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceEmptyPromptsSeparately() (+9 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (25): params, day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard(), matches() (+17 more)

### Community 19 - "runAction"
Cohesion: 0.15
Nodes (37): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains() (+29 more)

### Community 21 - "Cmd"
Cohesion: 0.07
Nodes (34): 10. Status, logs and cli, 12. Runner seams, applyCmdEnv(), applyCmdInput(), DockerImageCommand(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList() (+26 more)

### Community 23 - "runner_test.go"
Cohesion: 0.12
Nodes (66): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+58 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.07
Nodes (64): captureStderr(), downloadEnvWithImage(), manyWorkflowsDir(), TestAbsPath(), TestConfiguredInstanceName(), TestDeployKubernetesProbesOperatorSecrets(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput() (+56 more)

### Community 26 - "issueWith"
Cohesion: 0.15
Nodes (27): extraMQ(), extraSolace(), runExtra(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError(), TestExtraKeysEnvSuffixWarns() (+19 more)

### Community 27 - ".stop"
Cohesion: 0.14
Nodes (16): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Defaults"
Cohesion: 0.13
Nodes (32): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+24 more)

### Community 29 - "run"
Cohesion: 0.14
Nodes (18): run(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves() (+10 more)

### Community 30 - ".keystore"
Cohesion: 0.10
Nodes (24): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+16 more)

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (32): stateCell(), Age(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText() (+24 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 43 - "Runner"
Cohesion: 0.15
Nodes (34): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), checkImageCommand(), emit(), envPairs() (+26 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (30): KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.09
Nodes (57): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+49 more)

### Community 47 - "rawDefaults"
Cohesion: 0.15
Nodes (17): Build, Release checklist, Release (CI), solmq-conn-util -- development guide, The spec generator (`solmq-conn-util-generator.html`), defaultsFromRaw(), applyDest(), nodePtr() (+9 more)

### Community 48 - "Kubernetes"
Cohesion: 0.10
Nodes (21): ownedNames(), Input, Instance, KV, StoreFile, Mount, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand() (+13 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.20
Nodes (13): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+5 more)

### Community 52 - ".add"
Cohesion: 0.14
Nodes (28): checkStoreDefinition(), checkMisplacedTransform(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkImagePull(), checkJavaOptions() (+20 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.13
Nodes (32): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+24 more)

### Community 56 - "certfixture.go"
Cohesion: 0.18
Nodes (20): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+12 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.26
Nodes (11): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+3 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (23): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+15 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.09
Nodes (49): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+41 more)

### Community 61 - "Podman"
Cohesion: 0.11
Nodes (20): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Secrets, Service, applyDockerDefaults() (+12 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (31): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+23 more)

### Community 64 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 65 - "Resolver"
Cohesion: 0.09
Nodes (41): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, CreatedStore, TestResolvePodmanFiles(), TestResolveStores() (+33 more)

### Community 67 - "ScriptInstalled"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 68 - "Image"
Cohesion: 0.14
Nodes (12): PullSecret, dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON() (+4 more)

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.13
Nodes (21): DockerStorePlaceholders(), TestDockerStorePlaceholders(), View, groupOf(), JSON(), namespaceScope(), noteLine(), Render() (+13 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "Runner"
Cohesion: 0.07
Nodes (46): podmanDeploy(), podmanRemove(), previousStoreSecrets(), firstLine(), probeOperatorSecrets(), Contract, Decisions, Divergences (+38 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 75 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image.go"
Cohesion: 0.13
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.11
Nodes (23): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), TestBuildLeaderElection() (+15 more)

### Community 80 - "helperProcessArgv"
Cohesion: 0.22
Nodes (12): helperProcessArgv(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput(), TestOSRunRejectsUnresolvableArgv0(), TestOSRunWiresStdinToChild(), TestOSRunWritesNothingToStderr() (+4 more)

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "Resource"
Cohesion: 0.24
Nodes (10): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+2 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "tlsSelectingWF"
Cohesion: 0.33
Nodes (7): TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF(), certsOf(), TestCreatedStoreSelections(), TestStoreDefinitions()

### Community 89 - ".name"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "TestApplyStatusAccessAppendsAfterExistingUsers"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 92 - "SolaceProps"
Cohesion: 0.32
Nodes (12): SolaceProps(), defaults(), existing(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(), TestSolacePropsUseStoreMountPaths() (+4 more)

### Community 93 - "shellContentInfo"
Cohesion: 0.50
Nodes (3): existingPKCS12, existingPKCS12Shell, shellContentInfo

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (9): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+1 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.20
Nodes (10): 11. download jar, 1. Architecture, 2. CLI, 4. Kubernetes, 5. Docker, 6. Podman, 8. TLS, 9. Secrets (+2 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (56): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+48 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "GenerateDocker"
Cohesion: 0.12
Nodes (25): 7. Container posture, mount, Mount, TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestToIssues(), GenerateDocker() (+17 more)

### Community 102 - ".RunSplit"
Cohesion: 0.33
Nodes (5): TestOSRunSplitKeepsTheStreamsApart(), TestOSRunSplitRejectsEmptyAndUnresolvableArgv(), TestOSRunSplitWiresStdinAndEnv(), call, splitRunner

### Community 103 - ".Write"
Cohesion: 0.25
Nodes (8): How the suite is built, mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.07
Nodes (34): fileOwner(), fileOwner(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader(), rootlessIDs() (+26 more)

### Community 106 - ".add"
Cohesion: 0.20
Nodes (17): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkPayloadBlock(), checkTransform() (+9 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "solmq-conn-util troubleshooting"
Cohesion: 0.22
Nodes (9): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release (+1 more)

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 114 - "Env"
Cohesion: 0.31
Nodes (8): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), removeTarget(), PodmanServiceName(), Env, instanceRequest

### Community 115 - "GeneratePodman"
Cohesion: 0.11
Nodes (39): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain(), synthWorkflowFiles() (+31 more)

### Community 117 - "defsWithStores"
Cohesion: 0.14
Nodes (23): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestMQStoreSelectionRequiresTLS(), TestPasswordConflictOnSameBinder(), TestPasswordConflictSolaceSide(), TestSolaceKeyAliasRequiresTCPSAndKeystore() (+15 more)

### Community 119 - "CA"
Cohesion: 0.22
Nodes (17): Option, AsCA(), CA(), RSAKey(), Validity(), TestCheckExistingStoreFromOpenSSL(), TestResolvePodmanFilesOpensExistingStores(), createdCA() (+9 more)

### Community 120 - "Build"
Cohesion: 0.12
Nodes (26): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestBuildCipherConflictWarning(), TestBuildCreatedStores() (+18 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 123 - "misplacedEnvTransforms"
Cohesion: 0.57
Nodes (6): envTransforms(), appendTransformKeys(), documentMapping(), mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 124 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 126 - "flattenBlock"
Cohesion: 0.52
Nodes (4): extraKeys(), flattenBlock(), isMergeKey(), keyValue

### Community 127 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 128 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 129 - "TLSConfig"
Cohesion: 0.22
Nodes (13): TLSConfig, TestTLSConfigStoreLookup(), checkCreatedSelection(), checkConnections(), checkCred(), checkDefaultsCredentials(), checkLeaderElection(), checkSide() (+5 more)

### Community 132 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 133 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 136 - "SafeToken"
Cohesion: 0.12
Nodes (15): 3. Deploy execution, CheckDeployCommand(), checkImage(), checkSecurityUserRoles(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser() (+7 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 220 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **23 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `getting-started.md`, `.Write`, `cmd/solmq-conn-util`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 149 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 149 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.09797979797979799 - nodes in this community are weakly interconnected._
- **Why does `TestHelperProcess()` connect `.Write` to `testing.T`, `runner_test.go`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11655011655011654 - nodes in this community are weakly interconnected._