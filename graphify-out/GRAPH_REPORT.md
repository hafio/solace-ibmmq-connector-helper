# Graph Report - solace-ibmmq-connector-helper  (2026-11-06)

## Corpus Check
- 144 files · ~435,522 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2767 nodes · 11770 edges · 115 communities (94 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1794 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ac5da834`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Store
- .step
- validate_extra_test.go
- dockergen_test.go
- Config
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- run
- completion_test.go
- actDocker
- crypto/x509.Certificate
- runAction
- .keystore
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- hostChecks
- attachRunner
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
- podmanDeploy
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- deref
- gen_runas_test.go
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
- Kubernetes
- sinceFlag
- parse_test.go
- storeFiles
- Resolver
- runner_test.go
- TestGeneratorPageKnownKeysInSync
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- Runner
- solmq-conn-util configuration reference
- Side
- statusreport_test.go
- Writer
- support_test.go
- libs/image_test.go
- buildLeaderElection
- go_pkg_io_fs
- createdStoreDockerSpec
- jks.go
- maven.go
- getting-started.md
- Cred
- solmq-conn-util operations reference
- libs.go
- TestDownloadSetMapMatchesModel
- filenameFromEscapedPath
- solaceSide
- shellContentInfo
- SolaceProps
- validate_extra_test.go
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- canIProbe
- CommandRules
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- allowCommandValue
- GeneratePodman
- GenerateKubernetes
- defsWithStores
- Build
- solmq-conn-util abbreviations

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 159 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 87 edges
4. `write()` - 81 edges
5. `Build()` - 73 edges
6. `bg()` - 65 edges
7. `wfOK()` - 63 edges
8. `Download()` - 58 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `6. Podman` --references--> `ResolveQuadletScope()`  [INFERRED]
  docs/DESIGN.md → internal/runner/runner.go
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestCompletionModelMetadataComplete()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (115 total, 21 thin omitted)

### Community 0 - "Store"
Cohesion: 0.19
Nodes (7): Store, StoreSecret, StoreSources, TestStoreSecretParses(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.07
Nodes (25): sortInstances(), bg(), newProgress(), progressElapsed(), progressTrim(), actStatus(), clearScreen(), confirmInstall() (+17 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.10
Nodes (63): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+55 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.07
Nodes (40): Input, Instance, renderConfigMap(), CreatedStore, Render(), renderContentConfig(), renderHealthcheck(), renderService() (+32 more)

### Community 4 - "Config"
Cohesion: 0.17
Nodes (23): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain(), synthWorkflowFiles() (+15 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.05
Nodes (75): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret (+67 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.29
Nodes (11): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (25): cli, cmd/solmq-conn-util, Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen (+17 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.15
Nodes (41): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), TestBinderIdentityUsesTheCredentialPair() (+33 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.19
Nodes (29): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+21 more)

### Community 13 - "testing.T"
Cohesion: 0.05
Nodes (76): TestOwnedNamesCoversEverythingThisReleaseCreates(), TestProgressElapsedShapes(), TestWatchFlagAcceptsBareAndInterval(), mustWrite(), TestWorkflowExamplesMatchGoldenSpecs(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write() (+68 more)

### Community 14 - "run"
Cohesion: 0.11
Nodes (22): resolveTarget(), main(), run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestDeployPodmanSetButEmptyCredentialFailsBeforeAnyWrite(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate() (+14 more)

### Community 15 - "completion_test.go"
Cohesion: 0.16
Nodes (22): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+14 more)

### Community 17 - "actDocker"
Cohesion: 0.19
Nodes (23): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), failAll() (+15 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (25): params, certAlias(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard() (+17 more)

### Community 19 - "runAction"
Cohesion: 0.14
Nodes (40): runLifecycle(), runLogs(), allowCommandFlag(), approvalsFrom(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+32 more)

### Community 21 - ".keystore"
Cohesion: 0.20
Nodes (15): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+7 more)

### Community 23 - "runner_test.go"
Cohesion: 0.11
Nodes (72): isClusterDefault(), isOurs(), currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), envTransforms(), sortedKeys(), StoreCreateKeys(), StoreExistingKeys() (+64 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.11
Nodes (23): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput() (+15 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (72): captureStderr(), downloadEnvWithImage(), nsItemJSON(), TestAbsPath(), TestAllowCommandSudoNamesTheSupportedAlternative(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDeployKubernetesProbesOperatorSecrets() (+64 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.26
Nodes (11): newPipeReader(), progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+3 more)

### Community 28 - "Defaults"
Cohesion: 0.21
Nodes (22): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+14 more)

### Community 29 - "hostChecks"
Cohesion: 0.24
Nodes (10): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), TestMissingStoreFiles(), TestMountChecksLabelEachPathByItsKey(), DockerPlan, HostMount (+2 more)

### Community 30 - "attachRunner"
Cohesion: 0.33
Nodes (10): How the suite is built, attached, attachRunner, fakeCall, fakeRunner, noAttach, queuedResp, queueRunner (+2 more)

### Community 31 - "statusreport.go"
Cohesion: 0.10
Nodes (29): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Banner(), banner(), Bytes() (+21 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.18
Nodes (24): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+16 more)

### Community 43 - "podmanDeploy"
Cohesion: 0.11
Nodes (32): podmanDeploy(), podmanRemove(), previousStoreSecrets(), init(), checkLinger(), checkSubIDs(), ensureLinger(), liveHost() (+24 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (34): CA(), KeyPKCS8(), NotCA(), PEM(), TestCheckExistingStoreFromOpenSSL(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases() (+26 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (48): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+40 more)

### Community 47 - "deref"
Cohesion: 0.10
Nodes (33): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), Security, isNull(), deref() (+25 more)

### Community 48 - "gen_runas_test.go"
Cohesion: 0.47
Nodes (7): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestSecretTargetsAreAbsolutePaths()

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.13
Nodes (17): acc, Binder, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding, displayName() (+9 more)

### Community 52 - ".add"
Cohesion: 0.11
Nodes (45): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources() (+37 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (36): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+28 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (39): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+31 more)

### Community 56 - "certfixture.go"
Cohesion: 0.14
Nodes (26): Key, Option, Trusted, AsCA(), DER(), Cert, Issue(), JKS() (+18 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "verbUsage"
Cohesion: 0.15
Nodes (29): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+21 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.07
Nodes (60): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner() (+52 more)

### Community 61 - "Kubernetes"
Cohesion: 0.11
Nodes (20): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), ownedNames(), Env, Kubernetes, Secrets (+12 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (27): ApplyStats(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV(), keys(), TestApplyStats() (+19 more)

### Community 64 - "storeFiles"
Cohesion: 0.43
Nodes (7): TestResolvePodmanFilesOpensExistingStores(), createdCA(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestResolveDockerStoresLimits(), TestValidateStoreBytesWithoutAccess()

### Community 65 - "Resolver"
Cohesion: 0.12
Nodes (33): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, TestResolvePodmanFiles(), Resolver, kubeStoresSizeError() (+25 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (84): Contract, canIDenied(), ExecArgv(), InstallScript(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript() (+76 more)

### Community 68 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 69 - "parse.go"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "Runner"
Cohesion: 0.13
Nodes (33): checkImageCommand(), scriptOut(), firstLine(), probeOperatorSecrets(), Docker(), DockerImageCommand(), DockerImagePull(), DockerLifecycle() (+25 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Side"
Cohesion: 0.10
Nodes (22): fixedLeaderNames(), leaderSession(), TLSConfig, TestLeaderElectionEffectiveMode(), TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding() (+14 more)

### Community 75 - "statusreport_test.go"
Cohesion: 0.14
Nodes (12): ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing() (+4 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.10
Nodes (35): 9.4 podman, builtinList(), connectorRelease(), imageNameTag(), imageSatisfies(), listRanges(), loadImageLibs(), omitListProvenance() (+27 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.18
Nodes (15): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+7 more)

### Community 80 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "jks.go"
Cohesion: 0.19
Nodes (8): jksCert(), joinAliases(), readJKS(), sniff(), SniffJKS(), TestSniff(), format, jksKey

### Community 83 - "maven.go"
Cohesion: 0.11
Nodes (32): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchPOM(), groupPath(), highestStable() (+24 more)

### Community 84 - "getting-started.md"
Cohesion: 0.22
Nodes (8): Build, Release checklist, Release (CI), solmq-conn-util -- development guide, Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 85 - "Cred"
Cohesion: 0.13
Nodes (6): Image, Cred, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), TestUserSecretLiteralAndEnv(), placeholderSecretRef()

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (26): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+18 more)

### Community 88 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 89 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 90 - "solaceSide"
Cohesion: 0.41
Nodes (12): TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+4 more)

### Community 91 - "shellContentInfo"
Cohesion: 0.50
Nodes (3): existingPKCS12, existingPKCS12Shell, shellContentInfo

### Community 92 - "SolaceProps"
Cohesion: 0.14
Nodes (21): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release, Workflows and numbering (+13 more)

### Community 93 - "validate_extra_test.go"
Cohesion: 0.14
Nodes (23): connDefaults(), kubeNoSecrets(), kubeWith(), TestCheckKubeSpreadAndAffinity(), TestCheckKubeTolerations(), TestClusterOnlyStoreNeedsAHostSource(), TestConnRefSolaceTopicSourceIsError(), TestConnRefStrictOnlyDestination() (+15 more)

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.14
Nodes (14): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+6 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.13
Nodes (17): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.5 Run-as and hardening, 9. Platform sections, 10. Status, logs and cli, 11. download jar, 1. Architecture (+9 more)

### Community 98 - "dispatch"
Cohesion: 0.06
Nodes (68): dispatch(), captureStdout(), containsToken(), nsList(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+60 more)

### Community 102 - "CommandRules"
Cohesion: 0.33
Nodes (5): 3. Deploy execution, CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.19
Nodes (15): podmanHostChecks(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestPodmanReader(), TestRootlessReaderSeesTheInvokersGroup() (+7 more)

### Community 106 - ".add"
Cohesion: 0.17
Nodes (20): Instance, JavaOptions, YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType() (+12 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.16
Nodes (22): Binding, WorkflowEnable, built, Bundle, Management, Model, SecretConflict, secretConflictError() (+14 more)

### Community 115 - "GeneratePodman"
Cohesion: 0.09
Nodes (32): mount, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), TestGeneratePodmanCarriesLogbackOnlyWithSyslog(), TestGeneratePodmanCarriesStatusScript() (+24 more)

### Community 116 - "GenerateKubernetes"
Cohesion: 0.09
Nodes (35): File, KubeOpts, b64(), build(), synthTLSWorkflowFiles(), TestConfigRejectsSecretNameConflict(), TestGenerateKubernetesCarriesStatusScript(), TestGenerateKubernetesRefusesAnInjectedQuantity() (+27 more)

### Community 117 - "defsWithStores"
Cohesion: 0.18
Nodes (19): TestConnectionDefinitionValidation(), TestRemovedDefaultsKeysRejected(), TestSecurityUserRoles(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), TestStatusUserPasswordEnvCharset(), TestStatusUserReservedName(), defsWithStores(), leWorkflows() (+11 more)

### Community 120 - "Build"
Cohesion: 0.08
Nodes (36): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers() (+28 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 231 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `getting-started.md`, `attachRunner`, `.add`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 154 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 154 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.07268170426065163 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util command reference` connect `Command Reference Details` to `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10291858678955453 - nodes in this community are weakly interconnected._