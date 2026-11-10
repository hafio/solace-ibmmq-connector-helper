# Graph Report - solace-ibmmq-connector-helper  (2026-11-10)

## Corpus Check
- 145 files · ~469,094 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2922 nodes · 12577 edges · 121 communities (100 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1888 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `53747e5a`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Store
- .step
- validate_extra_test.go
- dockergen_test.go
- shippedExamplesRequest
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- statusscript_test.go
- .run
- completion_test.go
- github.com/hafio/go-deploy-utility/runner.Runner
- crypto/x509.Certificate
- runAction
- jks.go
- Go Module Definition
- runner_test.go
- hostReader
- main_test.go
- issueWith
- .stop
- Defaults
- withPrompter
- deployDocker
- statusreport.go
- buildBundle
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
- QuadletScope
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- deref
- TestGoldenKubernetesNoSecrets
- Command Reference Details
- Golden Bash Completion
- Side
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
- storeFiles
- Resolver
- runner_test.go
- rootless_test.go
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- context.Context
- solmq-conn-util configuration reference
- watchStatus
- ParseDefaults
- Writer
- support_test.go
- libs/image_test.go
- Model
- builtinList
- testing.T
- resolveStatusPassword
- maven.go
- getting-started.md
- rawDefaults
- solmq-conn-util operations reference
- libs.go
- RenderQuadlet
- libs/image_test.go
- solaceSide
- os.File
- solmq-conn-util troubleshooting
- TestSelectsStores
- Expand
- FormatScalar
- buildLeaderElection
- repeatableName
- SolaceProps
- tailFlag
- cmd/solmq-conn-util
- solmq-conn-util design notes
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- applyConnections
- .add
- Getting started with solmq-conn-util
- yw
- Image
- GenerateKubernetes
- Kubernetes
- GeneratePodman
- gen_extra_test.go
- validate_test.go
- filenameFromEscapedPath
- solmq-conn-util abbreviations
- TestDownloadSetMapMatchesModel
- solmq-conn-util command reference
- allowCommandValue
- TestSanitizeAndIsTCPS

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 160 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 91 edges
4. `write()` - 76 edges
5. `bg()` - 75 edges
6. `Build()` - 74 edges
7. `wfOK()` - 63 edges
8. `Download()` - 58 edges
9. `Defaults` - 51 edges
10. `imageOK()` - 51 edges

## Surprising Connections (you probably didn't know these)
- `9. Secrets` --references--> `statusPasswordFor()`  [INFERRED]
  docs/DESIGN.md → internal/gen/gen.go
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

## Communities (121 total, 21 thin omitted)

### Community 0 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.12
Nodes (17): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), MergeService(), ParseDeployment(), TestParseDeploymentAndService() (+9 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (68): approved(), TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService() (+60 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (34): Input, Instance, CreatedStore, Render(), renderContentConfig(), renderHealthcheck(), renderService(), seconds() (+26 more)

### Community 4 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 5 - "deploy_test.go"
Cohesion: 0.05
Nodes (76): storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), ManagementPort(), nameOnlySecret() (+68 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.16
Nodes (19): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+11 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (27): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+19 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.13
Nodes (47): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+39 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.12
Nodes (33): breEscape(), Render(), runBindingBinders(), runScriptFunction(), scriptFunction(), splitDetailsBlock(), splitHealthBlock(), TestBindingBindersPairsEachBindingWithItsBinder() (+25 more)

### Community 14 - ".run"
Cohesion: 0.12
Nodes (29): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateDockerToFile(), TestGenerateFlagsBeforeAndAfterPositional() (+21 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (23): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+15 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.12
Nodes (45): waitForRollout(), actDocker(), actKubernetes(), bg(), checkImageCommand(), confirmDockerRedeploy(), confirmRedeploy(), envPairs() (+37 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (25): params, certAlias(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard() (+17 more)

### Community 19 - "runAction"
Cohesion: 0.10
Nodes (63): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), absPath(), absResolver(), actPodman(), allowCommandFlag() (+55 more)

### Community 21 - "jks.go"
Cohesion: 0.12
Nodes (23): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+15 more)

### Community 23 - "runner_test.go"
Cohesion: 0.13
Nodes (40): currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, DurableName(), mustParseUUID(), TestDurableNameDeterministic() (+32 more)

### Community 24 - "hostReader"
Cohesion: 0.11
Nodes (18): fileOwner(), fileOwner(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestMissingStoreFiles() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.05
Nodes (122): TestLifecycleRolloutWait(), TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStderr(), captureStdout(), containsToken(), podListJSON(), reconnectingReport() (+114 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.17
Nodes (15): progressErase(), TestProgressDoneLinesCarryTheirOwnTime(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+7 more)

### Community 28 - "Defaults"
Cohesion: 0.13
Nodes (33): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+25 more)

### Community 29 - "withPrompter"
Cohesion: 0.09
Nodes (48): canIAsked(), gateKube(), kubeFixture(), nsItemJSON(), nsList(), removeResp(), runKube(), TestCliKubernetesArgvShape() (+40 more)

### Community 30 - "deployDocker"
Cohesion: 0.22
Nodes (19): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources(), deployDocker() (+11 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (40): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), stateCell(), Banner(), banner() (+32 more)

### Community 32 - "buildBundle"
Cohesion: 0.29
Nodes (7): applyStatusAccess(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 43 - "QuadletScope"
Cohesion: 0.12
Nodes (29): hostChecks(), init(), checkLinger(), checkSubIDs(), ensureLinger(), liveHost(), rootlessHostErrors(), subIDRefusal() (+21 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.18
Nodes (39): Option, AsCA(), CA(), KeyPKCS8(), NotCA(), PEM(), RSAKey(), Validity() (+31 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 48 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.18
Nodes (24): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+16 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Side"
Cohesion: 0.10
Nodes (21): fixedLeaderNames(), TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef(), TestSideUsernameSecretBothSystems() (+13 more)

### Community 52 - ".add"
Cohesion: 0.15
Nodes (26): checkStoreDefinition(), checkMisplacedTransform(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkImagePull(), checkJavaOptions() (+18 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (42): Opts, Build(), containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildCreatedStores(), TestBuildLeaderElectionSharesBinderSecretNames() (+34 more)

### Community 55 - "instanceSession"
Cohesion: 0.09
Nodes (44): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+36 more)

### Community 56 - "certfixture.go"
Cohesion: 0.16
Nodes (22): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+14 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (18): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+10 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.12
Nodes (39): TestLifecycleVerbsActInPlace(), downloadEnvWithImage(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+31 more)

### Community 61 - "Podman"
Cohesion: 0.08
Nodes (27): renderService(), toPodmanMount(), Mount, JavaOptions, JavaEnv(), NormalizeJavaOptions(), TestJavaEnvMergesTheTLSFlag(), TestParseEnvJavaOptionsAbsent() (+19 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (31): bindingLine(), ObjectExists(), ParseApplication(), ParsePods(), SecretDataKeys(), splitKV(), degradedOf(), keys() (+23 more)

### Community 64 - "storeFiles"
Cohesion: 0.36
Nodes (8): TestResolvePodmanFilesOpensExistingStores(), createdCA(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestResolveDockerStoresDigestsExistingStores(), TestResolveDockerStoresLimits(), TestValidateStoreBytesWithoutAccess()

### Community 65 - "Resolver"
Cohesion: 0.10
Nodes (40): TestMountChecksLabelEachPathByItsKey(), DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestResolvePodmanFiles() (+32 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (87): Contract, CheckProbe(), InstallScript(), KubernetesProbes(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript() (+79 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.15
Nodes (25): indexOfCall(), newTargetRunner(), testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb(), TestDeployPodmanRootlessLinger() (+17 more)

### Community 69 - "parse.go"
Cohesion: 0.11
Nodes (24): ApplyStats(), connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), ParseImageDigest() (+16 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.15
Nodes (20): appendField(), deriveStatusPassword(), groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication() (+12 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.23
Nodes (16): connectionsReport(), render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderConnectionsAndDegradedWorkflows(), TestRenderContainerViewAllNamespacesLeadsWithNamespace() (+8 more)

### Community 72 - "context.Context"
Cohesion: 0.12
Nodes (26): defaultAnswer(), scriptOut(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesDeleteNamespace(), KubernetesGetJSON() (+18 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "watchStatus"
Cohesion: 0.18
Nodes (15): sleepAfter(), TestWatchStatusAppendsSectionsWhenNotATerminal(), TestWatchStatusCancelledContextExitsZero(), TestWatchStatusDrawsNothingForARefreshCutShort(), TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected(), TestWatchStatusRedrawsUnderTheHeader(), TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn(), TestWatchStatusSleepsTheIntervalAfterEachRefresh() (+7 more)

### Community 75 - "ParseDefaults"
Cohesion: 0.15
Nodes (13): TestApplicationLeaderElectionSessionPlaintext(), ParseDefaults(), TestParseCreateBlock(), TestParseTLSIsStrict(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError(), TestParseDefaultsFull() (+5 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.12
Nodes (14): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+6 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.16
Nodes (14): connectorRelease(), imageNameTag(), imageSatisfies(), releaseTag(), TestConnectorRelease(), TestImageNameTag(), TestImageSatisfies(), TestValidateImageVersionQualifiers() (+6 more)

### Community 79 - "Model"
Cohesion: 0.11
Nodes (22): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+14 more)

### Community 80 - "builtinList"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 81 - "testing.T"
Cohesion: 0.04
Nodes (81): manyWorkflowsDir(), newPipeReader(), TestAbsPath(), TestGenerateConfigWorkflowCapExceeded(), TestIsIndex(), TestOwnedObjectsCoverEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic() (+73 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.17
Nodes (16): SecretRef, derivedPW(), kvRefs(), TestResolveCredentials(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName() (+8 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 85 - "rawDefaults"
Cohesion: 0.15
Nodes (16): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults, rawLeader (+8 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 89 - "libs/image_test.go"
Cohesion: 0.27
Nodes (13): loadImageLibs(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines() (+5 more)

### Community 90 - "solaceSide"
Cohesion: 0.21
Nodes (21): TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+13 more)

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 93 - "TestSelectsStores"
Cohesion: 0.40
Nodes (6): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (9): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+1 more)

### Community 96 - "buildLeaderElection"
Cohesion: 0.15
Nodes (18): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+10 more)

### Community 98 - "SolaceProps"
Cohesion: 0.21
Nodes (16): TLS, MountPath(), SolaceProps(), defaults(), existing(), keys(), TestMountPathIsPerStore(), TestSolacePropsAlwaysWritePKCS12() (+8 more)

### Community 100 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 102 - "solmq-conn-util design notes"
Cohesion: 0.12
Nodes (15): 10. Status, logs and cli, 11. download jar, 1. Architecture, 2. CLI, 3. Deploy execution, 4. Kubernetes, 5. Docker, 8. TLS (+7 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "applyConnections"
Cohesion: 0.25
Nodes (8): applyConnections(), detailError(), kubeItems(), walkBindings(), renderHealthDocument(), IsUp(), healthNode, kubeList

### Community 106 - ".add"
Cohesion: 0.23
Nodes (17): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkPayloadBlock(), checkTransform() (+9 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "Image"
Cohesion: 0.16
Nodes (10): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+2 more)

### Community 110 - "GenerateKubernetes"
Cohesion: 0.12
Nodes (25): File, b64(), synthTLSWorkflowFiles(), TestGeneratePodmanNeverReadsTheStores(), TestPlansCarryTheExpandedHostPathsTheyMount(), TestRetiredCreateNamesDeployButDoNotValidate(), TestSelectedStoreIsTheRequestOnKubernetes(), TestStoresSharingAFileNameMountApart() (+17 more)

### Community 111 - "Kubernetes"
Cohesion: 0.14
Nodes (14): Input, Instance, KV, KubeOpts, StoreFile, Written(), SecretProbe, secretProbes() (+6 more)

### Community 115 - "GeneratePodman"
Cohesion: 0.17
Nodes (22): 7. Container posture, mount, Mount, build(), TestTargetMounts(), TestToIssues(), GenerateDocker(), GeneratePodman() (+14 more)

### Community 116 - "gen_extra_test.go"
Cohesion: 0.15
Nodes (32): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+24 more)

### Community 117 - "validate_test.go"
Cohesion: 0.16
Nodes (22): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF() (+14 more)

### Community 119 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 123 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 125 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 134 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 227 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util command reference` connect `solmq-conn-util command reference` to `Command Reference Details`, `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 155 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`dispatch()` has 155 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.11561561561561562 - nodes in this community are weakly interconnected._
- **Why does `Command details` connect `Command Reference Details` to `solmq-conn-util command reference`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10869565217391304 - nodes in this community are weakly interconnected._