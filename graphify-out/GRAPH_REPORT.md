# Graph Report - solace-ibmmq-connector-helper  (2026-10-29)

## Corpus Check
- 141 files · ~429,157 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2735 nodes · 11422 edges · 130 communities (109 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1716 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `43e96d1f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- statusscript_test.go
- .step
- validate_extra_test.go
- dockergen_test.go
- GenerateKubernetes
- Runner
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
- Runner
- Go Module Definition
- runner_test.go
- RenderQuadlet
- TestDownloadSetMapMatchesModel
- issueWith
- .stop
- Defaults
- run
- .keystore
- statusreport.go
- deploy_test.go
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
- Context
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
- libs/image_test.go
- Resolver
- ScriptInstalled
- Image
- time.Time
- statusreport/render.go
- statusreport/render_test.go
- JavaOptions
- solmq-conn-util configuration reference
- Store
- podmanDeploy
- Writer
- support_test.go
- libs/image.go
- Build
- helperProcessArgv
- createdStoreDockerSpec
- watchFlag
- maven.go
- getting-started.md
- jks.go
- solmq-conn-util operations reference
- libs.go
- TestSelectsStores
- .name
- solaceSide
- ParseDefaults
- SolaceProps
- shellContentInfo
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- Resource
- vt_windows.go
- .Write
- shippedExamplesRequest
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- kubePod
- javaCtx
- Env
- GeneratePodman
- buildLeaderElection
- defsWithStores
- 7. status
- storeFiles
- Component
- solmq-conn-util abbreviations
- solmq-conn-util command reference
- filenameFromEscapedPath
- LogsArgv
- auto-complete
- 6. TLS stores
- cmd/solmq-conn-util
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- Side

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 154 edges
2. `hasErr()` - 104 edges
3. `captureStderr()` - 82 edges
4. `write()` - 76 edges
5. `Build()` - 73 edges
6. `wfOK()` - 63 edges
7. `Download()` - 58 edges
8. `Runner` - 52 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestCompletionModelMetadataComplete()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestTestCatalogSnapshotInSync()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/testcatalog_test.go

## Import Cycles
- None detected.

## Communities (130 total, 21 thin omitted)

### Community 0 - "statusscript_test.go"
Cohesion: 0.14
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 1 - ".step"
Cohesion: 0.07
Nodes (28): sortInstances(), newProgress(), progressTrim(), actStatus(), clearScreen(), confirmInstall(), instanceNames(), markMissing() (+20 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.07
Nodes (44): Input, Instance, renderConfigMap(), composeEscape(), composeQuote(), flowList(), CreatedStore, Render() (+36 more)

### Community 4 - "GenerateKubernetes"
Cohesion: 0.10
Nodes (29): Binding, Instance, 7. Container posture, built, KubeOpts, Model, SecretConflict, build() (+21 more)

### Community 5 - "Runner"
Cohesion: 0.07
Nodes (41): checkImageCommand(), firstLine(), probeOperatorSecrets(), 3. Deploy execution, Docker(), DockerImageCommand(), DockerImagePull(), DockerLifecycle() (+33 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.16
Nodes (19): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+11 more)

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
Cohesion: 0.12
Nodes (43): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+35 more)

### Community 11 - "scan_test.go"
Cohesion: 0.09
Nodes (47): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+39 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (92): TestDeployDockerWarnsAboutMountsTheConnectorCannotRead(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits(), TestUnreadableMountsNamesWhatTheConnectorCannotRead(), writeMode(), TestWorkflowExamplesMatchGoldenSpecs() (+84 more)

### Community 14 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (23): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+15 more)

### Community 17 - "removeNamespace"
Cohesion: 0.12
Nodes (22): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestRemoveKubernetesSeamHappyPath(), TestRemoveKubernetesSeamPromptsBeforeTearingDown() (+14 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (25): params, certAlias(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard() (+17 more)

### Community 19 - "runAction"
Cohesion: 0.14
Nodes (38): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains() (+30 more)

### Community 21 - "Runner"
Cohesion: 0.07
Nodes (34): 10. Status, logs and cli, 12. Runner seams, applyCmdEnv(), applyCmdInput(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats() (+26 more)

### Community 23 - "runner_test.go"
Cohesion: 0.17
Nodes (33): isClusterDefault(), isOurs(), DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5(), canIProbe() (+25 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.11
Nodes (24): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv() (+16 more)

### Community 25 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.20
Nodes (14): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+6 more)

### Community 28 - "Defaults"
Cohesion: 0.13
Nodes (33): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+25 more)

### Community 29 - "run"
Cohesion: 0.13
Nodes (20): main(), run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError() (+12 more)

### Community 30 - ".keystore"
Cohesion: 0.20
Nodes (15): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+7 more)

### Community 31 - "statusreport.go"
Cohesion: 0.09
Nodes (27): Banner(), banner(), Bytes(), canonicalRef(), Cores(), Instance, ImageMismatch(), InstanceBanner() (+19 more)

### Community 32 - "deploy_test.go"
Cohesion: 0.06
Nodes (69): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret (+61 more)

### Community 43 - "Runner"
Cohesion: 0.12
Nodes (34): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+26 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (34): CA(), KeyPKCS8(), NotCA(), PEM(), TestCheckExistingStoreFromOpenSSL(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases() (+26 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "rawDefaults"
Cohesion: 0.15
Nodes (16): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults, rawLeader (+8 more)

### Community 48 - "Kubernetes"
Cohesion: 0.17
Nodes (10): TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes, Syslog, checkSyslog(), credentialsWired() (+2 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.10
Nodes (21): cli, Command details, deploy, download, examples, generate, help, logs (+13 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.12
Nodes (19): acc, Binder, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding, WorkflowEnable (+11 more)

### Community 52 - "Context"
Cohesion: 0.12
Nodes (31): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget() (+23 more)

### Community 54 - "Application"
Cohesion: 0.19
Nodes (26): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+18 more)

### Community 55 - "instanceSession"
Cohesion: 0.13
Nodes (31): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+23 more)

### Community 56 - "certfixture.go"
Cohesion: 0.14
Nodes (26): Key, Option, Trusted, AsCA(), DER(), Cert, Issue(), JKS() (+18 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.15
Nodes (21): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+13 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.08
Nodes (61): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), captureStderr(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner() (+53 more)

### Community 61 - "Podman"
Cohesion: 0.08
Nodes (27): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), decodeCreate(), Secrets, Service, TestCredCreateRemovedKeys() (+19 more)

### Community 63 - "parse_test.go"
Cohesion: 0.12
Nodes (25): EngineNamesByImage(), ObjectExists(), ParseApplication(), ParsePods(), SecretDataKeys(), splitKV(), keys(), TestEngineNamesByImage() (+17 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.18
Nodes (18): loadImageLibs(), splitJarBasename(), TestBuiltinList(), TestBuiltinListPicksTheNearestLine(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable(), TestEmbeddedOmitListFullyParses() (+10 more)

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (38): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestTargetMounts(), DockerPlan (+30 more)

### Community 67 - "ScriptInstalled"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 68 - "Image"
Cohesion: 0.18
Nodes (9): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret() (+1 more)

### Community 69 - "time.Time"
Cohesion: 0.15
Nodes (17): ApplyStats(), connectorIndex(), digestFrom(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), ParseImageDigest() (+9 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.18
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "JavaOptions"
Cohesion: 0.29
Nodes (7): JavaOptions, JavaEnv(), NormalizeJavaOptions(), TestJavaEnvMergesTheTLSFlag(), TestParseEnvJavaOptionsAbsent(), TestParseEnvJavaOptionsSpellings(), Instance

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.13
Nodes (15): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+7 more)

### Community 74 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 75 - "podmanDeploy"
Cohesion: 0.21
Nodes (17): podmanDeploy(), podmanRemove(), previousStoreSecrets(), podmanHostChecks(), SecretRef, TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestResolveCredentials() (+9 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.33
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image.go"
Cohesion: 0.21
Nodes (12): builtinList(), connectorRelease(), imageNameTag(), listRanges(), releaseTag(), TestValidateImageVersionQualifiers(), trimJarClassifier(), validateImageVersion() (+4 more)

### Community 79 - "Build"
Cohesion: 0.10
Nodes (29): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed() (+21 more)

### Community 80 - "helperProcessArgv"
Cohesion: 0.09
Nodes (32): PodmanDeploy(), PodmanRemove(), SystemctlNRestarts(), attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes() (+24 more)

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "getting-started.md"
Cohesion: 0.34
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "jks.go"
Cohesion: 0.19
Nodes (8): jksCert(), joinAliases(), readJKS(), sniff(), SniffJKS(), TestSniff(), format, jksKey

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.15
Nodes (13): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (20): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+12 more)

### Community 88 - "TestSelectsStores"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 89 - ".name"
Cohesion: 0.24
Nodes (9): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, omitListProvenance(), TestEmbeddedListRange() (+1 more)

### Community 90 - "solaceSide"
Cohesion: 0.20
Nodes (19): applyStatusAccess(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages() (+11 more)

### Community 91 - "ParseDefaults"
Cohesion: 0.17
Nodes (12): ParseDefaults(), TestParseCreateBlock(), TestParseTLSIsStrict(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError(), TestParseDefaultsFull(), TestParseDefaultsLeaderSession() (+4 more)

### Community 92 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

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
Cohesion: 0.05
Nodes (106): dispatch(), captureStdout(), containsToken(), downloadEnvWithImage(), TestAbsPath(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList() (+98 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "Resource"
Cohesion: 0.24
Nodes (10): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+2 more)

### Community 102 - "vt_windows.go"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 103 - ".Write"
Cohesion: 0.25
Nodes (8): How the suite is built, mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 104 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 105 - "RunAsID"
Cohesion: 0.11
Nodes (23): dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanReader(), rootlessIDs(), sharedIDs(), TestMissingStoreFiles(), TestMountChecksLabelEachPathByItsKey() (+15 more)

### Community 106 - ".add"
Cohesion: 0.31
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.26
Nodes (15): Management, blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 109 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 110 - "kubePod"
Cohesion: 0.29
Nodes (7): kubeItems(), podComponents(), podMatchesImage(), kubeContainer, kubeContainerState, kubeList, kubePod

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 114 - "Env"
Cohesion: 0.43
Nodes (6): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), Env, instanceRequest

### Community 115 - "GeneratePodman"
Cohesion: 0.10
Nodes (51): File, Config(), issuesContain(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak() (+43 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.18
Nodes (15): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+7 more)

### Community 117 - "defsWithStores"
Cohesion: 0.18
Nodes (24): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), defsWithStores(), leWorkflows(), manyWorkflows(), mgmtSide() (+16 more)

### Community 118 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 119 - "storeFiles"
Cohesion: 0.43
Nodes (7): TestResolvePodmanFilesOpensExistingStores(), createdCA(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestResolveDockerStoresLimits(), TestValidateStoreBytesWithoutAccess()

### Community 120 - "Component"
Cohesion: 0.33
Nodes (7): engineComponents(), sortComponents(), stateCell(), ExitCodeText(), TestExitCodeText(), Component, Container

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 123 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 124 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 125 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 126 - "6. TLS stores"
Cohesion: 0.40
Nodes (5): 6.1 Existing stores, 6.2 Created stores, 6.3 Certificate validation, 6.4 Rotating a store, 6. TLS stores

### Community 127 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 128 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 129 - "Side"
Cohesion: 0.10
Nodes (22): fixedLeaderNames(), leaderSession(), TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+14 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 217 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `TestHelperProcess()` connect `.Write` to `testing.T`, `runner_test.go`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 149 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 149 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `statusscript_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13756613756613756 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `getting-started.md`, `.Write`, `cmd/solmq-conn-util`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.0734006734006734 - nodes in this community are weakly interconnected._