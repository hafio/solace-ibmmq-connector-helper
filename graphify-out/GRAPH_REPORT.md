# Graph Report - solace-ibmmq-connector-helper  (2026-11-10)

## Corpus Check
- 145 files · ~461,662 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2893 nodes · 12442 edges · 131 communities (111 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1867 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6c706478`
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
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- withPrompter
- deployDocker
- statusreport.go
- buildLeaderElection
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
- 12. Runner seams
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- deref
- renderDeployment
- Command Reference Details
- Golden Bash Completion
- ParseWorkflow
- validate.go
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
- CA
- Resolver
- runner_test.go
- rootless_test.go
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- context.Context
- solmq-conn-util configuration reference
- watchStatus
- .step
- Writer
- support_test.go
- libs/image_test.go
- Model
- builtinList
- testing.T
- resolveStatusPassword
- maven.go
- Security
- solmq-conn-util operations reference
- libs.go
- Secrets
- solaceSide
- gen_runas_test.go
- solmq-conn-util troubleshooting
- TestSelectsStores
- Expand
- FormatScalar
- repeatableName
- SolaceProps
- tailFlag
- TestRenderLibsPVCCreate
- applyStatusAccess
- solmq-conn-util design notes
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- Image
- GenerateKubernetes
- renderWithPull
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- targetMounts
- gen_extra_test.go
- validate_test.go
- Side
- hostChecks
- Build
- solmq-conn-util abbreviations
- 7. status
- TestDownloadSetMapMatchesModel
- go_pkg_io_fs
- solmq-conn-util command reference
- javaCtx
- 4. Workflow files
- TestStoresVolumeProjectsEverySource
- auto-complete
- solmq-conn-util -- development guide
- TestSanitizeAndIsTCPS
- tlsSelectingWF
- SafeToken

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 158 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 91 edges
4. `write()` - 76 edges
5. `bg()` - 75 edges
6. `Build()` - 73 edges
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
- `TLS` --references--> `existing()`  [INFERRED]
  docs/troubleshooting.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (131 total, 20 thin omitted)

### Community 0 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.15
Nodes (12): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), MergeService(), ParseDeployment(), TestParseDeploymentAndService() (+4 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (69): Deployment, TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey(), baseKubeDeploy() (+61 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (34): Input, Instance, Render(), renderContentConfig(), renderHealthcheck(), renderService(), seconds(), secretList() (+26 more)

### Community 4 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 5 - "deploy_test.go"
Cohesion: 0.18
Nodes (34): baseKube(), before(), fullFixtureInput(), lineDiff(), mustRender(), one(), TestCreatedSecretsTakeDerivedNames(), TestNamespaceManifestMatchesWhatRenderEmits() (+26 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.16
Nodes (19): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+11 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (26): cli, cmd/solmq-conn-util, Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen (+18 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.13
Nodes (46): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+38 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.19
Nodes (29): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+21 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.07
Nodes (52): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+44 more)

### Community 14 - ".run"
Cohesion: 0.17
Nodes (23): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateDockerToFile(), TestGenerateFlagsBeforeAndAfterPositional(), TestLoadEnvExcludesEnvFileFromWorkflowSet() (+15 more)

### Community 15 - "completion_test.go"
Cohesion: 0.16
Nodes (22): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+14 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.11
Nodes (47): waitForRollout(), actKubernetes(), bg(), checkImageCommand(), podmanDeploy(), podmanRemove(), preflight(), preflightRead() (+39 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (35): 4.2 `solace:` side, 4.3 `mq:` side, 4.5 Binders and durable names, readJKS(), SniffJKS(), build(), certAlias(), day() (+27 more)

### Community 19 - "runAction"
Cohesion: 0.08
Nodes (74): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), absPath(), absResolver(), actDocker(), actPodman() (+66 more)

### Community 21 - "jks.go"
Cohesion: 0.20
Nodes (6): jksCert(), joinAliases(), sniff(), TestSniff(), format, jksKey

### Community 23 - "runner_test.go"
Cohesion: 0.14
Nodes (38): currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, DurableName(), mustParseUUID(), TestDurableNameDeterministic() (+30 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.05
Nodes (123): TestLifecycleRolloutWait(), TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStderr(), captureStdout(), containsToken(), downloadEnvWithImage(), newTargetRunner() (+115 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.25
Nodes (13): newPipeReader(), progressErase(), TestProgressDoneLinesCarryTheirOwnTime(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+5 more)

### Community 28 - "Defaults"
Cohesion: 0.12
Nodes (35): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+27 more)

### Community 29 - "withPrompter"
Cohesion: 0.09
Nodes (48): canIAsked(), gateKube(), kubeFixture(), nsItemJSON(), nsList(), removeResp(), runKube(), TestCliEngineArgvShape() (+40 more)

### Community 30 - "deployDocker"
Cohesion: 0.22
Nodes (19): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources(), deployDocker() (+11 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (40): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), stateCell(), Banner(), banner() (+32 more)

### Community 32 - "buildLeaderElection"
Cohesion: 0.14
Nodes (18): leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+10 more)

### Community 43 - "12. Runner seams"
Cohesion: 0.31
Nodes (9): init(), liveHost(), 12. Runner seams, ScopeEnv, LiveScopeEnv(), ResolveQuadletScope(), scopeEnv(), TestLiveScopeEnv() (+1 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (30): KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.09
Nodes (57): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+49 more)

### Community 47 - "deref"
Cohesion: 0.12
Nodes (27): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), envTransforms(), isNull(), deref() (+19 more)

### Community 48 - "renderDeployment"
Cohesion: 0.09
Nodes (29): Input, Instance, KV, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), ManagementPort() (+21 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.10
Nodes (21): cli, Command details, deploy, download, examples, generate, help, logs (+13 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "ParseWorkflow"
Cohesion: 0.11
Nodes (20): extraKeyNames(), extraValue(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(), TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), TestResolveCarriesTheConnectionsExtraKeys() (+12 more)

### Community 52 - "validate.go"
Cohesion: 0.10
Nodes (32): 3. Deploy execution, Toleration, checkStoreDefinition(), checkMisplacedTransform(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker() (+24 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.08
Nodes (45): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+37 more)

### Community 56 - "certfixture.go"
Cohesion: 0.15
Nodes (16): Key, Option, params, Trusted, AsCA(), Cert, JKS(), KeyEC() (+8 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (18): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+10 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.12
Nodes (39): TestLifecycleVerbsActInPlace(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+31 more)

### Community 61 - "Kubernetes"
Cohesion: 0.10
Nodes (21): KubeOpts, Written(), SecretProbe, secretProbes(), Kubernetes, Secrets, Service, applyDockerDefaults() (+13 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV(), keys() (+22 more)

### Community 64 - "CA"
Cohesion: 0.21
Nodes (22): CA(), DER(), Issue(), KeyStorePKCS12(), TrustStorePKCS12(), Validity(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestCheckExistingStore() (+14 more)

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (35): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, StoreFile, b64(), TestResolvePodmanFiles() (+27 more)

### Community 67 - "runner_test.go"
Cohesion: 0.04
Nodes (92): Contract, ExecArgv(), InstallScript(), KubernetesProbes(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript() (+84 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.15
Nodes (25): indexOfCall(), testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb(), TestDeployPodmanRootlessLinger(), TestPodmanRootfulRule() (+17 more)

### Community 69 - "parse.go"
Cohesion: 0.12
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), kubeItems(), ParseImageDigest() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.14
Nodes (20): appendField(), deriveStatusPassword(), groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication() (+12 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "context.Context"
Cohesion: 0.12
Nodes (27): defaultAnswer(), scriptOut(), DockerDeployDigest(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesDeleteNamespace() (+19 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "watchStatus"
Cohesion: 0.33
Nodes (11): sleepAfter(), TestWatchStatusAppendsSectionsWhenNotATerminal(), TestWatchStatusCancelledContextExitsZero(), TestWatchStatusDrawsNothingForARefreshCutShort(), TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected(), TestWatchStatusRedrawsUnderTheHeader(), TestWatchStatusRejectsAnIntervalTheLoopCannotUse(), TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn() (+3 more)

### Community 75 - ".step"
Cohesion: 0.09
Nodes (11): newProgress(), progressElapsed(), progressTrim(), watchTitle(), enableVirtualTerminal(), View, progress, progressMode (+3 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.12
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 79 - "Model"
Cohesion: 0.15
Nodes (18): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+10 more)

### Community 80 - "builtinList"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 81 - "testing.T"
Cohesion: 0.05
Nodes (67): manyWorkflowsDir(), TestAbsPath(), TestGenerateConfigWorkflowCapExceeded(), TestIsIndex(), TestOwnedObjectsCoverEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes() (+59 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.20
Nodes (14): SecretRef, derivedPW(), kvRefs(), TestResolveCredentials(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName() (+6 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 85 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.15
Nodes (13): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 89 - "Secrets"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 90 - "solaceSide"
Cohesion: 0.28
Nodes (15): propsNode(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildSolaceBinderPassesValidateCertificateThrough(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder() (+7 more)

### Community 91 - "gen_runas_test.go"
Cohesion: 0.35
Nodes (9): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestRunAsOfRefusesWhatValidateRefuses(), TestSecretTargetsAreAbsolutePaths() (+1 more)

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.22
Nodes (9): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release (+1 more)

### Community 93 - "TestSelectsStores"
Cohesion: 0.67
Nodes (4): kubeNoSecrets(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.24
Nodes (8): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), nodeToProps(), sameNode()

### Community 98 - "SolaceProps"
Cohesion: 0.32
Nodes (12): SolaceProps(), defaults(), existing(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(), TestSolacePropsUseStoreMountPaths() (+4 more)

### Community 100 - "TestRenderLibsPVCCreate"
Cohesion: 0.67
Nodes (3): TestLibsPVNameIsNamespaced(), TestRenderLibsPVCCreate(), LibsPVName()

### Community 101 - "applyStatusAccess"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 102 - "solmq-conn-util design notes"
Cohesion: 0.20
Nodes (10): 10. Status, logs and cli, 11. download jar, 1. Architecture, 2. CLI, 4. Kubernetes, 5. Docker, 8. TLS, 9. Secrets (+2 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.18
Nodes (15): podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestMissingStoreFiles(), TestPodmanReader(), TestRootlessReaderSeesTheInvokersGroup() (+7 more)

### Community 106 - ".add"
Cohesion: 0.19
Nodes (18): JavaOptions, YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkPayloadBlock() (+10 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.28
Nodes (14): Management, blockIndicator(), Node(), q(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "Image"
Cohesion: 0.16
Nodes (10): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+2 more)

### Community 110 - "GenerateKubernetes"
Cohesion: 0.12
Nodes (28): File, synthTLSWorkflowFiles(), TestGenerateKubernetesRefusesAnInjectedQuantity(), TestGeneratePodmanNeverReadsTheStores(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate(), TestSelectedStoreIsTheRequestOnKubernetes(), TestStoresSharingAFileNameMountApart() (+20 more)

### Community 111 - "renderWithPull"
Cohesion: 0.67
Nodes (3): renderWithPull(), TestImagePullSecretPayloadIsOpaqueToDeploy(), TestImagePullSecretStates()

### Community 114 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 115 - "targetMounts"
Cohesion: 0.10
Nodes (31): mount, CreatedStore, Mount, build(), TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestTargetMounts() (+23 more)

### Community 116 - "gen_extra_test.go"
Cohesion: 0.15
Nodes (31): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+23 more)

### Community 117 - "validate_test.go"
Cohesion: 0.21
Nodes (17): registryImage(), retiredNamesKube(), TestRetiredCreateNamesReportedOnlyByValidate(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), defsWithStores(), leWorkflows(), mgmtSide(), TestDeployNameTooLong() (+9 more)

### Community 118 - "Side"
Cohesion: 0.10
Nodes (21): fixedLeaderNames(), TLSConfig, TestLeaderElectionEffectiveMode(), TestTLSConfigStoreLookup(), applyDest(), Side, TestSideUsernameSecretBothSystems(), checkCreatedSelection() (+13 more)

### Community 119 - "hostChecks"
Cohesion: 0.28
Nodes (9): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), TestMountChecksLabelEachPathByItsKey(), DockerPlan, HostMount, DockerStorePlaceholders() (+1 more)

### Community 120 - "Build"
Cohesion: 0.10
Nodes (29): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed() (+21 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 123 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 124 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 125 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 126 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 127 - "4. Workflow files"
Cohesion: 0.33
Nodes (6): 4.1 Top-level keys, 4.4 Reusing a connection (`conn-ref`), 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms, 4. Workflow files

### Community 129 - "TestStoresVolumeProjectsEverySource"
Cohesion: 0.40
Nodes (4): storeItem, storeSource, storeSources(), TestStoresVolumeProjectsEverySource()

### Community 130 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 133 - "solmq-conn-util -- development guide"
Cohesion: 0.50
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 134 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 135 - "tlsSelectingWF"
Cohesion: 0.67
Nodes (4): TestClusterOnlyStoreNeedsAHostSource(), TestRetiredStoresBlockTiers(), tlsSelectingWF(), certsOf()

### Community 136 - "SafeToken"
Cohesion: 0.50
Nodes (4): TestSafeActuatorUser(), TestSafeToken(), TestSafeTokenCharset(), SafeToken()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 227 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `runner_test.go`, `rootless_test.go`, `watchStatus`, `GenerateKubernetes`, `renderDeployment`, `github.com/hafio/go-deploy-utility/runner.Runner`, `getting-started.md`, `main_test.go`, `Defaults`, `withPrompter`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 153 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`dispatch()` has 153 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10890269151138716 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util operations reference` connect `solmq-conn-util operations reference` to `7. status`, `getting-started.md`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `dockergen_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08771929824561403 - nodes in this community are weakly interconnected._