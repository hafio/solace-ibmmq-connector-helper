# Graph Report - solace-ibmmq-connector-helper  (2026-11-07)

## Corpus Check
- 145 files · ~448,510 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2835 nodes · 12115 edges · 120 communities (100 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1846 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6cda87cb`
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
- Runner
- solmq-conn-util configuration reference
- TLSConfig
- KV
- Writer
- support_test.go
- libs/image_test.go
- Model
- builtinList
- testing.T
- resolveStatusPassword
- maven.go
- getting-started.md
- Cred
- solmq-conn-util operations reference
- libs.go
- TestDownloadSetMapMatchesModel
- filenameFromEscapedPath
- solaceSide
- Side
- solmq-conn-util troubleshooting
- TestSelectsStores
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- Resource
- binderNames
- Port
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- checkTargets
- shippedExamplesRequest
- allowCommandValue
- PodmanStoreSecretName
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- GeneratePodman
- GenerateKubernetes
- defsWithStores
- Build
- solmq-conn-util abbreviations

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 163 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 89 edges
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
- `TLS` --references--> `existing()`  [INFERRED]
  docs/troubleshooting.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (120 total, 20 thin omitted)

### Community 0 - "Store"
Cohesion: 0.12
Nodes (20): Store, StoreSecret, StoreSources, TestStoreSecretParses(), SolaceProps(), defaults(), existing(), keys() (+12 more)

### Community 1 - ".step"
Cohesion: 0.11
Nodes (16): noPodsFound(), sortInstances(), newProgress(), progressTrim(), actStatus(), confirmInstall(), instanceNames(), markMissing() (+8 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.10
Nodes (59): approved(), TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), dockerOK() (+51 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (32): Input, Instance, Render(), renderContentConfig(), renderHealthcheck(), renderService(), seconds(), secretList() (+24 more)

### Community 4 - "Config"
Cohesion: 0.15
Nodes (31): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+23 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.05
Nodes (75): storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), ManagementPort(), nameOnlySecret() (+67 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.17
Nodes (18): pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys() (+10 more)

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
Cohesion: 0.14
Nodes (53): connDefaults(), TestBinderIdentityUsesTheCredentialPair(), TestCheckSideMQMissingFields(), TestCheckSideSolaceMissingAndBadScheme(), TestConnRefSolaceTopicSourceIsError(), TestConnRefStrictOnlyDestination(), TestConnRefUnknownAndSystemMismatch(), TestConnRefValidResolvesNoError() (+45 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 14 - ".run"
Cohesion: 0.12
Nodes (30): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateDockerToFile() (+22 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.12
Nodes (41): actDocker(), actKubernetes(), bg(), checkImageCommand(), confirmDockerRedeploy(), confirmRedeploy(), envPairs(), failAll() (+33 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.14
Nodes (19): day(), dedupe(), entry(), fingerprint(), guard(), matchesKey(), orderChain(), readBackError() (+11 more)

### Community 19 - "runAction"
Cohesion: 0.10
Nodes (63): resolveTarget(), runLifecycle(), runLogs(), absPath(), absResolver(), actPodman(), allowCommandFlag(), approvalsFrom() (+55 more)

### Community 21 - "jks.go"
Cohesion: 0.10
Nodes (28): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+20 more)

### Community 23 - "runner_test.go"
Cohesion: 0.13
Nodes (45): isClusterDefault(), isOurs(), currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, DurableName() (+37 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (68): captureStderr(), downloadEnvWithImage(), manyWorkflowsDir(), podmanEnvSudo(), TestAbsPath(), TestConfiguredInstanceName(), TestDeployRefusesSudoCommandWhateverIsApproved(), TestDownloadDirDefaultAndPositionalOverride() (+60 more)

### Community 26 - "issueWith"
Cohesion: 0.15
Nodes (27): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError(), TestExtraKeysEnvSuffixWarns() (+19 more)

### Community 27 - ".stop"
Cohesion: 0.14
Nodes (16): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Defaults"
Cohesion: 0.14
Nodes (29): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+21 more)

### Community 29 - "withPrompter"
Cohesion: 0.09
Nodes (32): nsItemJSON(), nsList(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+24 more)

### Community 30 - "deployDocker"
Cohesion: 0.19
Nodes (22): deployDocker(), dockerInspectWithLabel(), dockerSpec(), TestDeployDockerChangedAsksBeforeCompose(), TestDeployDockerDeclinedRunsNothing(), TestDeployDockerFailedRunningProbeIsAnError(), TestDeployDockerNonTTYFailsNamingTheFlag(), TestDeployDockerNoPromptApproves() (+14 more)

### Community 31 - "statusreport.go"
Cohesion: 0.09
Nodes (29): stateCell(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance (+21 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.12
Nodes (32): The spec generator (`solmq-conn-util-generator.html`), TestValidateCleanSpecStillPasses(), syslogXML(), Validate(), configMapDoc(), deploymentDoc(), dirReader(), envWithKube() (+24 more)

### Community 43 - "podmanDeploy"
Cohesion: 0.15
Nodes (23): hostChecks(), init(), checkLinger(), ensureLinger(), liveHost(), rootlessHostErrors(), subIDRefusal(), podmanHostChecks() (+15 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.22
Nodes (31): KeyPKCS8(), NotCA(), PEM(), TestCheckExistingStoreFromOpenSSL(), Build(), aliases(), chain(), derOf() (+23 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "deref"
Cohesion: 0.06
Nodes (50): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode() (+42 more)

### Community 48 - "gen_runas_test.go"
Cohesion: 0.47
Nodes (7): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestSecretTargetsAreAbsolutePaths()

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Kubernetes"
Cohesion: 0.09
Nodes (24): ownedNames(), Input, Instance, KV, StoreFile, workflowsFromRaw(), JavaOptions, JavaEnv() (+16 more)

### Community 52 - ".add"
Cohesion: 0.08
Nodes (40): checkStoreDefinition(), checkMisplacedTransform(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkImage(), checkImagePull() (+32 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (37): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+29 more)

### Community 56 - "certfixture.go"
Cohesion: 0.15
Nodes (20): Key, params, Trusted, DER(), Cert, Issue(), JKS(), KeyEC() (+12 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.16
Nodes (18): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+10 more)

### Community 58 - "verbUsage"
Cohesion: 0.20
Nodes (19): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+11 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.10
Nodes (43): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), podmanEnv(), podmanEnvChained(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestAllowCommandSudoNamesTheSupportedAlternative() (+35 more)

### Community 61 - "Env"
Cohesion: 0.23
Nodes (12): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), Env, Docker, LibsMount, Podman (+4 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV(), keys() (+22 more)

### Community 64 - "CA"
Cohesion: 0.23
Nodes (16): Option, AsCA(), CA(), RSAKey(), Validity(), TestResolvePodmanFilesOpensExistingStores(), createdCA(), storeFiles() (+8 more)

### Community 65 - "Resolver"
Cohesion: 0.12
Nodes (35): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestTargetMounts() (+27 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (89): Contract, canIDenied(), ExecArgv(), InstallScript(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript() (+81 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.32
Nodes (13): testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb(), TestDeployPodmanRootlessLinger(), TestPodmanRootfulRule(), TestPodmanRunAsAgainstTheSubIDRanges() (+5 more)

### Community 69 - "parse.go"
Cohesion: 0.13
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), kubeItems(), ParseImageDigest() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.20
Nodes (14): groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications(), renderContainerDetail() (+6 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "Runner"
Cohesion: 0.12
Nodes (23): scriptOut(), clearScreen(), enableVirtualTerminal(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON() (+15 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "TLSConfig"
Cohesion: 0.22
Nodes (13): TLSConfig, TestTLSConfigStoreLookup(), checkCreatedSelection(), checkConnections(), checkCred(), checkDefaultsCredentials(), checkLeaderElection(), checkSide() (+5 more)

### Community 75 - "KV"
Cohesion: 0.25
Nodes (9): appendField(), deriveStatusPassword(), MergeService(), ParseDeployment(), TestParseDeploymentAndService(), renderWorkload(), KV(), TestKVAlignsValuesOnTheWidestKey() (+1 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.12
Nodes (26): connectorRelease(), imageNameTag(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable() (+18 more)

### Community 79 - "Model"
Cohesion: 0.12
Nodes (22): Binding, JMSBinding, leaderNameFn, Session, SolaceBinding, WorkflowEnable, built, buildLeaderElection() (+14 more)

### Community 80 - "builtinList"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 81 - "testing.T"
Cohesion: 0.06
Nodes (60): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestDeployDockerSeamPipesComposeAndRuns(), TestRemoveDockerNeedsNoCreatedStoreSources() (+52 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.17
Nodes (16): SecretRef, derivedPW(), kvRefs(), TestResolveCredentials(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName() (+8 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "Cred"
Cohesion: 0.11
Nodes (9): Security, Image, Cred, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), TestUserSecretLiteralAndEnv(), placeholderSecretRef(), rawSecurity (+1 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 89 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "Side"
Cohesion: 0.27
Nodes (9): fixedLeaderNames(), Side, postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide() (+1 more)

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.22
Nodes (9): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release (+1 more)

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

### Community 98 - "dispatch"
Cohesion: 0.09
Nodes (46): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance() (+38 more)

### Community 100 - "Resource"
Cohesion: 0.24
Nodes (10): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+2 more)

### Community 101 - "binderNames"
Cohesion: 0.36
Nodes (9): binderNames(), binderOf(), eqStrs(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder() (+1 more)

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
Cohesion: 0.08
Nodes (27): fileOwner(), fileOwner(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+19 more)

### Community 106 - ".add"
Cohesion: 0.34
Nodes (11): YAMLKind(), checkContentType(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem(), checkTransforms() (+3 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.28
Nodes (14): Management, blockIndicator(), Node(), q(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "checkTargets"
Cohesion: 0.29
Nodes (8): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkRootlessRange(), checkRunAs(), checkTargets(), credentialsWired()

### Community 110 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 113 - "PodmanStoreSecretName"
Cohesion: 0.50
Nodes (4): TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), PodmanStoreSecretName(), PodmanUnitStoreSecrets()

### Community 114 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 115 - "GeneratePodman"
Cohesion: 0.11
Nodes (31): File, mount, CreatedStore, Mount, build(), synthTLSWorkflowFiles(), TestGeneratePodmanNeverReadsTheStores(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless() (+23 more)

### Community 116 - "GenerateKubernetes"
Cohesion: 0.11
Nodes (22): KubeOpts, PullSecret, dockerConfigJSON(), TestResolveStores(), TestRetiredCreateNamesDeployButDoNotValidate(), TestSelectedStoreIsTheRequestOnKubernetes(), GenerateKubernetes(), SecretProbe (+14 more)

### Community 117 - "defsWithStores"
Cohesion: 0.14
Nodes (18): registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestCheckKubeSecretNames(), TestCheckKubeSecretsCreateXorExisting(), TestClusterOnlyStoreNeedsAHostSource() (+10 more)

### Community 120 - "Build"
Cohesion: 0.07
Nodes (38): acc, Opts, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), buildBundle() (+30 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 231 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `existing()` connect `Store` to `certfixture.go`, `solmq-conn-util configuration reference`, `solmq-conn-util troubleshooting`, `runner_test.go`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 158 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 158 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Store` be split into smaller, more focused modules?**
  _Cohesion score 0.1168091168091168 - nodes in this community are weakly interconnected._
- **Why does `6. TLS stores` connect `solmq-conn-util configuration reference` to `Store`, `builtinList`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.11153846153846154 - nodes in this community are weakly interconnected._