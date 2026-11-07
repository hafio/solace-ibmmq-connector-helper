# Graph Report - solace-ibmmq-connector-helper  (2026-11-07)

## Corpus Check
- 145 files · ~450,007 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2849 nodes · 12176 edges · 127 communities (108 shown, 19 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1854 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7c1f3709`
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
- ResolveQuadletScope
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- deref
- gen_runas_test.go
- Command Reference Details
- Golden Bash Completion
- Kubernetes
- .add
- Application
- instanceSession
- certfixture.go
- renderCommandsDoc
- verbUsage
- maven_test.go
- write
- Env
- sinceFlag
- parse_test.go
- CA
- Resolver
- runner_test.go
- rootless_test.go
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- github.com/hafio/go-deploy-utility/runner.Cmd
- solmq-conn-util configuration reference
- TLSConfig
- Workload
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
- TestDownloadSetMapMatchesModel
- ParseWorkflow
- solaceSide
- watchStatus
- solmq-conn-util troubleshooting
- TestSelectsStores
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- SolaceProps
- tailFlag
- Resource
- libs/image_test.go
- Port
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- Image
- shippedExamplesRequest
- 7. status
- TestGeneratorPageKnownKeysInSync
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- GeneratePodman
- GenerateKubernetes
- validate_test.go
- SafeToken
- solmq-conn-util command reference
- Build
- solmq-conn-util abbreviations
- TestApplyStatusAccessAppendsAfterExistingUsers
- auto-complete
- 6. TLS stores
- cmd/solmq-conn-util
- javaCtx

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 163 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 91 edges
4. `write()` - 83 edges
5. `Build()` - 73 edges
6. `bg()` - 73 edges
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

## Communities (127 total, 19 thin omitted)

### Community 0 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.13
Nodes (13): sortInstances(), bg(), newProgress(), progressElapsed(), progressTrim(), confirmInstall(), instanceNames(), markMissing() (+5 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.08
Nodes (39): Input, Instance, renderConfigMap(), Render(), renderContentConfig(), renderHealthcheck(), renderService(), seconds() (+31 more)

### Community 4 - "Config"
Cohesion: 0.16
Nodes (29): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+21 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.05
Nodes (74): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret (+66 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.29
Nodes (11): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+3 more)

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
Cohesion: 0.14
Nodes (39): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+31 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.08
Nodes (49): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+41 more)

### Community 14 - ".run"
Cohesion: 0.12
Nodes (29): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateDockerToFile(), TestGenerateFlagsBeforeAndAfterPositional() (+21 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.11
Nodes (52): checkImageCommand(), confirmRedeploy(), podmanDeploy(), podmanRemove(), previousStoreSecrets(), checkLinger(), checkSubIDs(), ensureLinger() (+44 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.11
Nodes (24): params, certAlias(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard() (+16 more)

### Community 19 - "runAction"
Cohesion: 0.07
Nodes (79): resolveTarget(), runLifecycle(), runLogs(), absPath(), absResolver(), actDocker(), actKubernetes(), actPodman() (+71 more)

### Community 21 - "jks.go"
Cohesion: 0.10
Nodes (24): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+16 more)

### Community 23 - "runner_test.go"
Cohesion: 0.13
Nodes (46): isClusterDefault(), isOurs(), namespaceOccupants(), currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo (+38 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.05
Nodes (123): TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStderr(), captureStdout(), containsToken(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandSudoNamesTheSupportedAlternative() (+115 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.20
Nodes (15): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+7 more)

### Community 28 - "Defaults"
Cohesion: 0.13
Nodes (33): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+25 more)

### Community 29 - "withPrompter"
Cohesion: 0.11
Nodes (24): nsItemJSON(), nsList(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+16 more)

### Community 30 - "deployDocker"
Cohesion: 0.22
Nodes (19): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources(), deployDocker() (+11 more)

### Community 31 - "statusreport.go"
Cohesion: 0.09
Nodes (30): stateCell(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance (+22 more)

### Community 32 - "buildLeaderElection"
Cohesion: 0.11
Nodes (23): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), TestBuildLeaderElection() (+15 more)

### Community 43 - "ResolveQuadletScope"
Cohesion: 0.29
Nodes (8): init(), liveHost(), ScopeEnv, LiveScopeEnv(), ResolveQuadletScope(), scopeEnv(), TestLiveScopeEnv(), TestResolveQuadletScope()

### Community 44 - "storebuild_test.go"
Cohesion: 0.19
Nodes (37): Option, AsCA(), CA(), KeyPKCS8(), NotCA(), PEM(), RSAKey(), Validity() (+29 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.09
Nodes (57): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+49 more)

### Community 47 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 48 - "gen_runas_test.go"
Cohesion: 0.47
Nodes (7): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestSecretTargetsAreAbsolutePaths()

### Community 49 - "Command Reference Details"
Cohesion: 0.10
Nodes (21): cli, Command details, deploy, download, examples, generate, help, logs (+13 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Kubernetes"
Cohesion: 0.11
Nodes (18): ownedNames(), Instance, Mount, SecretRef, workflowsFromRaw(), JavaOptions, Kubernetes, Syslog (+10 more)

### Community 52 - ".add"
Cohesion: 0.11
Nodes (27): badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkImagePull(), checkJavaOptions(), checkKube(), checkLibs() (+19 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (38): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+30 more)

### Community 56 - "certfixture.go"
Cohesion: 0.16
Nodes (22): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+14 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "verbUsage"
Cohesion: 0.16
Nodes (27): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+19 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.12
Nodes (39): TestLifecycleVerbsActInPlace(), downloadEnvWithImage(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+31 more)

### Community 61 - "Env"
Cohesion: 0.20
Nodes (15): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), Env, Secrets, Docker (+7 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV(), keys() (+22 more)

### Community 64 - "CA"
Cohesion: 0.31
Nodes (10): TestResolvePodmanFilesOpensExistingStores(), BuildStore(), createdCA(), storeFiles(), TestBuildStore(), TestBuildStoreReadsEverySourceByPosition(), TestBuildStoreRejects(), TestResolveDockerStoresDigestsExistingStores() (+2 more)

### Community 65 - "Resolver"
Cohesion: 0.10
Nodes (38): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestResolveStores() (+30 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (93): Contract, canIDenied(), ExecArgv(), InstallScript(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), ScriptInstalled() (+85 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.13
Nodes (26): scriptOut(), testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb(), TestDeployPodmanRootlessLinger(), TestPodmanRootfulRule() (+18 more)

### Community 69 - "parse.go"
Cohesion: 0.12
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), kubeItems(), ParseImageDigest() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.15
Nodes (19): appendField(), deriveStatusPassword(), groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication() (+11 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "github.com/hafio/go-deploy-utility/runner.Cmd"
Cohesion: 0.14
Nodes (6): enableVirtualTerminal(), TestParseCommandExtraAllowed(), call, deployFns, fakeRunner, splitRunner

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.13
Nodes (15): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+7 more)

### Community 74 - "TLSConfig"
Cohesion: 0.13
Nodes (17): fixedLeaderNames(), TLSConfig, TestTLSConfigStoreLookup(), Side, TestSideUsernameSecretBothSystems(), checkCreatedSelection(), checkConnections(), checkCred() (+9 more)

### Community 75 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.16
Nodes (14): connectorRelease(), imageNameTag(), imageSatisfies(), releaseTag(), TestConnectorRelease(), TestImageNameTag(), TestImageSatisfies(), TestValidateImageVersionQualifiers() (+6 more)

### Community 79 - "Model"
Cohesion: 0.20
Nodes (13): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+5 more)

### Community 80 - "builtinList"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 81 - "testing.T"
Cohesion: 0.06
Nodes (58): manyWorkflowsDir(), TestGenerateConfigWorkflowCapExceeded(), TestWorkflowExamplesMatchGoldenSpecs(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), TestStoreConfigured(), ParseEnv(), TestParseEnvEmpty() (+50 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.20
Nodes (14): SecretRef, derivedPW(), kvRefs(), TestResolveCredentials(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName() (+6 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "rawDefaults"
Cohesion: 0.15
Nodes (16): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults, rawLeader (+8 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.15
Nodes (13): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 89 - "ParseWorkflow"
Cohesion: 0.11
Nodes (20): extraKeyNames(), extraValue(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(), TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), TestResolveCarriesTheConnectionsExtraKeys() (+12 more)

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "watchStatus"
Cohesion: 0.16
Nodes (16): sleepAfter(), TestWatchFlagAcceptsBareAndInterval(), TestWatchStatusAppendsSectionsWhenNotATerminal(), TestWatchStatusCancelledContextExitsZero(), TestWatchStatusRedrawsUnderTheHeader(), TestWatchStatusRejectsAnIntervalTheLoopCannotUse(), TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn(), TestWatchStatusSleepsTheIntervalAfterEachRefresh() (+8 more)

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 93 - "TestSelectsStores"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (9): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+1 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.20
Nodes (10): 10. Status, logs and cli, 11. download jar, 1. Architecture, 2. CLI, 4. Kubernetes, 5. Docker, 8. TLS, 9. Secrets (+2 more)

### Community 98 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

### Community 100 - "Resource"
Cohesion: 0.24
Nodes (10): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+2 more)

### Community 101 - "libs/image_test.go"
Cohesion: 0.27
Nodes (13): loadImageLibs(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines() (+5 more)

### Community 102 - "Port"
Cohesion: 0.29
Nodes (5): renderService(), Service, Port, checkPortRange(), checkPorts()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.09
Nodes (23): fileOwner(), fileOwner(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+15 more)

### Community 106 - ".add"
Cohesion: 0.18
Nodes (21): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkStoreDefinition(), checkContentType(), checkMisplacedTransform() (+13 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "Image"
Cohesion: 0.18
Nodes (9): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret() (+1 more)

### Community 110 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 111 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 113 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 114 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 115 - "GeneratePodman"
Cohesion: 0.12
Nodes (29): mount, CreatedStore, Mount, build(), TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestStoresSharingAFileNameMountApart() (+21 more)

### Community 116 - "GenerateKubernetes"
Cohesion: 0.13
Nodes (26): File, KubeOpts, synthTLSWorkflowFiles(), TestGenerateKubernetesRefusesAnInjectedQuantity(), TestGeneratePodmanNeverReadsTheStores(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestPlansCarryTheExpandedHostPathsTheyMount(), TestRetiredCreateNamesDeployButDoNotValidate() (+18 more)

### Community 117 - "validate_test.go"
Cohesion: 0.15
Nodes (29): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF() (+21 more)

### Community 118 - "SafeToken"
Cohesion: 0.29
Nodes (7): checkImage(), checkSecurityUserRoles(), TestSafeActuatorUser(), TestSafeToken(), TestSafeTokenCharset(), SafeActuatorUser(), SafeToken()

### Community 119 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 120 - "Build"
Cohesion: 0.12
Nodes (26): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestBuildCipherConflictWarning(), TestBuildCreatedStores() (+18 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "TestApplyStatusAccessAppendsAfterExistingUsers"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 123 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 124 - "6. TLS stores"
Cohesion: 0.40
Nodes (5): 6.1 Existing stores, 6.2 Created stores, 6.3 Certificate validation, 6.4 Rotating a store, 6. TLS stores

### Community 125 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 126 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 232 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util command reference` connect `solmq-conn-util command reference` to `Command Reference Details`, `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 158 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 158 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.12857142857142856 - nodes in this community are weakly interconnected._
- **Why does `Command details` connect `Command Reference Details` to `auto-complete`, `solmq-conn-util command reference`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11971153846153847 - nodes in this community are weakly interconnected._