# Graph Report - solace-ibmmq-connector-helper  (2026-11-11)

## Corpus Check
- 145 files · ~470,187 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2937 nodes · 12881 edges · 130 communities (106 shown, 24 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1901 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `298cccf9`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- StoreSources
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
- main_test.go
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
- runKube
- deployDocker
- statusreport.go
- .withRunner
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
- actPodman
- Command Reference Details
- Golden Bash Completion
- Store
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
- rootless_test.go
- parse.go
- statusreport/render.go
- Env
- context.Context
- solmq-conn-util configuration reference
- watchStatus
- ParseWorkflow
- Writer
- support_test.go
- libs/image_test.go
- Model
- Input
- testing.T
- resolveStatusPassword
- maven.go
- rawDefaults
- .stop
- libs.go
- RenderQuadlet
- Side
- Build
- decodeCreate
- solmq-conn-util troubleshooting
- vt_windows.go
- Expand
- FormatScalar
- buildLeaderElection
- repeatableName
- SolaceProps
- tailFlag
- 7. status
- misplacedEnvTransforms
- solmq-conn-util design notes
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- applyConnections
- .add
- Getting started with solmq-conn-util
- yw
- Image
- flattenBlock
- renderDeployment
- extraKeyNames
- 4. Workflow files
- GenerateDocker
- gen_extra_test.go
- validate_test.go
- CommandRules
- Security
- ParseRunAs
- solmq-conn-util abbreviations
- Table
- TestDownloadSetMapMatchesModel
- secretProbes
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- LeaderElection
- javaCtx
- allowCommandValue
- TestSanitizeAndIsTCPS

## God Nodes (most connected - your core abstractions)
1. `testApp()` - 190 edges
2. `dispatch()` - 161 edges
3. `hasErr()` - 105 edges
4. `captureStderr()` - 91 edges
5. `write()` - 76 edges
6. `bg()` - 75 edges
7. `Build()` - 74 edges
8. `wfOK()` - 63 edges
9. `Download()` - 58 edges
10. `Defaults` - 51 edges

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

## Communities (130 total, 24 thin omitted)

### Community 1 - ".step"
Cohesion: 0.15
Nodes (12): sortInstances(), instanceNames(), markMissing(), quoteAll(), MergeService(), ParseDeployment(), TestParseDeploymentAndService(), renderWorkload() (+4 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.10
Nodes (74): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+66 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.07
Nodes (32): Input, Instance, CreatedStore, Render(), renderContentConfig(), renderHealthcheck(), renderService(), seconds() (+24 more)

### Community 4 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 5 - "deploy_test.go"
Cohesion: 0.06
Nodes (64): storeItem, storeSource, PullSecret, storeSources(), baseKube(), before(), fullFixtureInput(), lineDiff() (+56 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.17
Nodes (18): pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys() (+10 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.09
Nodes (22): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+14 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.12
Nodes (48): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+40 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.07
Nodes (59): Build, Release checklist, Release (CI), solmq-conn-util -- development guide, The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader() (+51 more)

### Community 14 - "main_test.go"
Cohesion: 0.06
Nodes (74): captureStderr(), createdStoreDockerSpec(), gateKube(), imageStartReply(), manyWorkflowsDir(), newPipeReader(), nsItemJSON(), podListJSON() (+66 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.10
Nodes (51): waitForRollout(), actDocker(), actKubernetes(), bg(), checkImageCommand(), confirmDockerRedeploy(), confirmRedeploy(), envPairs() (+43 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.09
Nodes (38): params, 4.2 `solace:` side, 4.3 `mq:` side, 4.5 Binders and durable names, joinAliases(), readJKS(), SniffJKS(), build() (+30 more)

### Community 19 - "runAction"
Cohesion: 0.11
Nodes (48): resolveTarget(), runLifecycle(), runLogs(), allowCommandFlag(), approvalsFrom(), collectFlagsAndDirs(), confirmRemove(), contains() (+40 more)

### Community 21 - "jks.go"
Cohesion: 0.22
Nodes (4): jksCert(), sniff(), TestSniff(), format

### Community 23 - "runner_test.go"
Cohesion: 0.13
Nodes (37): podmanChangeSummary(), TestPodmanChangeSummary(), currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, mustWrite() (+29 more)

### Community 24 - "hostReader"
Cohesion: 0.12
Nodes (25): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+17 more)

### Community 25 - "main_test.go"
Cohesion: 0.08
Nodes (69): TestLifecycleRolloutWait(), TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStdout(), containsToken(), reconnectingReport(), testApp(), TestAutoCompleteDispatchPrintsScript() (+61 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.15
Nodes (6): newProgress(), progressElapsed(), progressTrim(), progress, progressMode, watchFlag

### Community 28 - "Defaults"
Cohesion: 0.13
Nodes (31): built, SecretConflict, build(), synthWorkflows(), secretConflictError(), statusPasswordFor(), Defaults, Workflow (+23 more)

### Community 29 - "runKube"
Cohesion: 0.21
Nodes (24): kubeFixture(), nsList(), removeResp(), runKube(), TestDeployManifestOmitsTheNamespace(), TestDeployNamespaceFlow(), TestDeployWaitsForTheRollout(), TestKubernetesProbeSets() (+16 more)

### Community 30 - "deployDocker"
Cohesion: 0.33
Nodes (14): envValue(), TestRemoveDockerNeedsNoCreatedStoreSources(), deployDocker(), dockerDigestOf(), dockerInspectWithLabel(), dockerSpec(), TestDeployDockerChangedAsksBeforeCompose(), TestDeployDockerDeclinedRunsNothing() (+6 more)

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (39): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), stateCell(), Banner(), banner() (+31 more)

### Community 32 - ".withRunner"
Cohesion: 0.11
Nodes (31): downloadEnvWithImage(), app, TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+23 more)

### Community 43 - "QuadletScope"
Cohesion: 0.14
Nodes (23): liveApp(), main(), run(), ordinaryHost(), ensureLinger(), liveHost(), subIDRefusal(), 12. Runner seams (+15 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (34): CA(), KeyPKCS8(), NotCA(), PEM(), TestCheckExistingStoreFromOpenSSL(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases() (+26 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (55): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+47 more)

### Community 47 - "deref"
Cohesion: 0.61
Nodes (6): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), isNull(), deref()

### Community 48 - "actPodman"
Cohesion: 0.30
Nodes (14): absPath(), absResolver(), actPodman(), emit(), failFast(), fileReader(), genConfig(), genDocker() (+6 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Store"
Cohesion: 0.12
Nodes (19): Store, StoreSecret, TLSConfig, TestStoreSecretParses(), TestTLSConfigStoreLookup(), TestStoreSecretNilSafe(), checkCreatedSelection(), createdRole() (+11 more)

### Community 52 - ".add"
Cohesion: 0.11
Nodes (37): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkDuplicateSources(), checkImage(), checkImagePull() (+29 more)

### Community 54 - "Application"
Cohesion: 0.11
Nodes (39): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+31 more)

### Community 55 - "instanceSession"
Cohesion: 0.09
Nodes (40): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+32 more)

### Community 56 - "certfixture.go"
Cohesion: 0.12
Nodes (29): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+21 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.16
Nodes (23): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+15 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.14
Nodes (36): TestLifecycleVerbsActInPlace(), podmanEnv(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamFailedRunLeavesNoComposeFile() (+28 more)

### Community 61 - "Kubernetes"
Cohesion: 0.14
Nodes (10): workflowsFromRaw(), Kubernetes, Service, Port, checkPortRange(), checkPorts(), credentialsWired(), rawEnv (+2 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (37): ApplyStats(), bindingLine(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+29 more)

### Community 64 - "storeFiles"
Cohesion: 0.17
Nodes (16): Option, AsCA(), RSAKey(), Validity(), TestResolvePodmanFilesOpensExistingStores(), createdCA(), createdEnv(), hasIssue() (+8 more)

### Community 65 - "Resolver"
Cohesion: 0.13
Nodes (32): DockerExisting, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestTargetMounts(), Resolver (+24 more)

### Community 67 - "runner_test.go"
Cohesion: 0.04
Nodes (94): Contract, ExecArgv(), InstallScript(), KubernetesProbes(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript() (+86 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.12
Nodes (29): canIAsked(), indexOfCall(), newTargetRunner(), testHost(), TestReadVerbsAskKubernetesNoPermissionFirst(), useHost(), TestSessionRootless(), linuxHost() (+21 more)

### Community 69 - "parse.go"
Cohesion: 0.16
Nodes (17): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), ParseImageDigest(), podComponents() (+9 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.18
Nodes (16): groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications(), renderConnections() (+8 more)

### Community 71 - "Env"
Cohesion: 0.19
Nodes (14): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), Env, Secrets, Docker, LibsMount (+6 more)

### Community 72 - "context.Context"
Cohesion: 0.11
Nodes (24): defaultAnswer(), scriptOut(), enableVirtualTerminal(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON() (+16 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "watchStatus"
Cohesion: 0.29
Nodes (13): sleepAfter(), TestWatchStatusAppendsSectionsWhenNotATerminal(), TestWatchStatusCancelledContextExitsZero(), TestWatchStatusDrawsNothingForARefreshCutShort(), TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected(), TestWatchStatusRedrawsUnderTheHeader(), TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn(), TestWatchStatusSleepsTheIntervalAfterEachRefresh() (+5 more)

### Community 75 - "ParseWorkflow"
Cohesion: 0.14
Nodes (14): TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), TestResolveCarriesTheConnectionsExtraKeys(), ParseWorkflow(), TestConnRefSideMayTuneBinding(), TestParseWorkflowAmbiguousSystemAndDest(), TestParseWorkflowConnRef(), TestParseWorkflowEnabledDefaultsTrue(), TestParseWorkflowSolaceAndMQ() (+6 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.10
Nodes (36): builtinList(), connectorRelease(), imageNameTag(), imageSatisfies(), listRanges(), loadImageLibs(), omitListProvenance(), releaseTag() (+28 more)

### Community 79 - "Model"
Cohesion: 0.14
Nodes (19): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+11 more)

### Community 80 - "Input"
Cohesion: 0.20
Nodes (11): mount, Mount, podmanSecretRefs(), toDockerMount(), toDockerMounts(), toPodmanMount(), Mount, SecretRef (+3 more)

### Community 81 - "testing.T"
Cohesion: 0.05
Nodes (79): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateDockerToFile() (+71 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.16
Nodes (17): SecretRef, appendField(), deriveStatusPassword(), derivedPW(), kvRefs(), TestResolveCredentials(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride() (+9 more)

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (33): acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML(), highestStable() (+25 more)

### Community 85 - "rawDefaults"
Cohesion: 0.20
Nodes (13): defaultsFromRaw(), applyDest(), nodePtr(), presentBlock(), rawDefaults, rawLeader, rawLogging, rawManagement (+5 more)

### Community 86 - ".stop"
Cohesion: 0.09
Nodes (23): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference, 10. download jar, 11. What gets generated (+15 more)

### Community 87 - "libs.go"
Cohesion: 0.13
Nodes (15): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), isRedirectStatus(), parseSHA1Sidecar(), requireHTTPS(), resolveSeed() (+7 more)

### Community 88 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 89 - "Side"
Cohesion: 0.25
Nodes (3): fixedLeaderNames(), Side, TestSideUsernameSecretBothSystems()

### Community 90 - "Build"
Cohesion: 0.09
Nodes (45): Opts, applyStatusAccess(), Build(), containsSub(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestApplyStatusAccessNoOperatorUsers() (+37 more)

### Community 91 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.29
Nodes (7): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 93 - "vt_windows.go"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.14
Nodes (14): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+6 more)

### Community 96 - "buildLeaderElection"
Cohesion: 0.13
Nodes (19): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), dropManaged(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision() (+11 more)

### Community 98 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

### Community 100 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 101 - "misplacedEnvTransforms"
Cohesion: 0.57
Nodes (6): envTransforms(), appendTransformKeys(), documentMapping(), mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 102 - "solmq-conn-util design notes"
Cohesion: 0.13
Nodes (17): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, 10. Status, logs and cli, 11. download jar (+9 more)

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
Cohesion: 0.31
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "Image"
Cohesion: 0.16
Nodes (9): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestResolvePullSecret(), Image, Cred (+1 more)

### Community 110 - "flattenBlock"
Cohesion: 0.52
Nodes (4): extraKeys(), flattenBlock(), isMergeKey(), keyValue

### Community 111 - "renderDeployment"
Cohesion: 0.07
Nodes (40): Input, Instance, KV, needsQuote(), QuoteScalar(), envValueQuote(), StoreFile, leaderMode() (+32 more)

### Community 113 - "extraKeyNames"
Cohesion: 0.43
Nodes (6): extraKeyNames(), extraValue(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys()

### Community 114 - "4. Workflow files"
Cohesion: 0.33
Nodes (6): 4.1 Top-level keys, 4.4 Reusing a connection (`conn-ref`), 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms, 4. Workflow files

### Community 115 - "GenerateDocker"
Cohesion: 0.14
Nodes (20): 7. Container posture, DockerStore, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), GenerateDocker(), DockerPlan, renderAll() (+12 more)

### Community 116 - "gen_extra_test.go"
Cohesion: 0.08
Nodes (52): File, Config(), issuesContain(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak() (+44 more)

### Community 117 - "validate_test.go"
Cohesion: 0.21
Nodes (17): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), defsWithStores(), leWorkflows(), mgmtSide(), TestDeployNameTooLong() (+9 more)

### Community 118 - "CommandRules"
Cohesion: 0.40
Nodes (4): CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 119 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 120 - "ParseRunAs"
Cohesion: 0.40
Nodes (5): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkRootlessRange()

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 123 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 124 - "secretProbes"
Cohesion: 0.50
Nodes (3): KubeOpts, SecretProbe, secretProbes()

### Community 125 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 127 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 134 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 232 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **24 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `existing()` connect `SolaceProps` to `certfixture.go`, `solmq-conn-util configuration reference`, `Store`, `runner_test.go`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 15 inferred relationships involving `testApp()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`testApp()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.14772727272727273 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `runner_test.go`, `rootless_test.go`, `watchStatus`, `main_test.go`, `renderDeployment`, `github.com/hafio/go-deploy-utility/runner.Runner`, `getting-started.md`, `gen_extra_test.go`, `.stop`, `.add`, `main_test.go`, `runKube`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Are the 156 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`dispatch()` has 156 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10234234234234234 - nodes in this community are weakly interconnected._