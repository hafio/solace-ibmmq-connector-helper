# Graph Report - solace-ibmmq-connector-helper  (2026-11-14)

## Corpus Check
- 148 files · ~487,648 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 3032 nodes · 12983 edges · 140 communities (118 shown, 22 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1909 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `84336bc2`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- .run
- .step
- validate_extra_test.go
- JavaOptions
- statusreport/render_test.go
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- TestGoldenKubernetesNoSecrets
- progressOn
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
- correctness
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
- Findings by area
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
- libs/image.go
- Model
- conventions
- testing.T
- resolveStatusPassword
- maven.go
- rawDefaults
- solmq-conn-util operations reference
- libs.go
- RenderQuadlet
- Side
- solaceSide
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
- Critical review - solmq-conn-util
- solmq-conn-util design notes
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- ParseApplication
- .add
- Getting started with solmq-conn-util
- yw
- libs/image_test.go
- ParseDefaults
- Decisions on the 2026-11-11 critical review
- .stop
- builtinList
- GenerateDocker
- gen_extra_test.go
- validate_test.go
- CommandRules
- cicd
- renderShapes
- solmq-conn-util abbreviations
- applyStatusAccess
- TestDownloadSetMapMatchesModel
- GenerateKubernetes
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- filenameFromEscapedPath
- javaCtx
- 6. TLS stores
- build
- design
- allowCommandValue
- quality
- expandValue
- TestSanitizeAndIsTCPS
- previousStoreSecrets
- solmq-conn-util -- development guide
- TestCredentialsNotWiredIsAnError
- tlsSelectingWF
- watchFlag

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
- `CR-0054 [minor] [new] docs/test.md carries no support notice and the list that enforces it is explicit, so a new doc is exempt` --references--> `TestSupportNoticeInGeneratorPage()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → cmd/solmq-conn-util/support_test.go
- `9. Secrets` --references--> `statusPasswordFor()`  [INFERRED]
  docs/DESIGN.md → internal/gen/gen.go
- `CR-0006 [major] [new] Engine calls use the unexpanded env.yaml while validate and the rendered artifacts use the expanded one` --references--> `Expand()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → internal/spec/expand.go
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go

## Import Cycles
- None detected.

## Communities (140 total, 22 thin omitted)

### Community 0 - ".run"
Cohesion: 0.11
Nodes (31): manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateConfigWorkflowCapExceeded() (+23 more)

### Community 1 - ".step"
Cohesion: 0.14
Nodes (13): sortInstances(), errExit(), newProgress(), actStatus(), instanceNames(), markMissing(), quoteAll(), watchStatus() (+5 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (68): Deployment, TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService() (+60 more)

### Community 3 - "JavaOptions"
Cohesion: 0.10
Nodes (22): Instance, Input, Instance, renderHealthcheck(), renderService(), seconds(), secretList(), workflowsFromRaw() (+14 more)

### Community 4 - "statusreport/render_test.go"
Cohesion: 0.23
Nodes (16): connectionsReport(), render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderConnectionsAndDegradedWorkflows(), TestRenderContainerViewAllNamespacesLeadsWithNamespace() (+8 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.06
Nodes (77): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret (+69 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.16
Nodes (19): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+11 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (26): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+18 more)

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
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (23): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+15 more)

### Community 14 - "progressOn"
Cohesion: 0.24
Nodes (15): app, newPipeReader(), progressErase(), progressOn(), TestProgressDoneLinesCarryTheirOwnTime(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds() (+7 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.10
Nodes (62): waitForRollout(), actDocker(), actKubernetes(), bg(), checkImageCommand(), confirmDockerRedeploy(), confirmRedeploy(), envPairs() (+54 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.08
Nodes (44): params, 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules (+36 more)

### Community 19 - "runAction"
Cohesion: 0.13
Nodes (42): resolveTarget(), runLifecycle(), runLogs(), allowCommandFlag(), approvalsFrom(), collectFlagsAndDirs(), confirmRemove(), contains() (+34 more)

### Community 21 - "jks.go"
Cohesion: 0.20
Nodes (4): jksCert(), sniff(), TestSniff(), format

### Community 23 - "runner_test.go"
Cohesion: 0.13
Nodes (43): podmanChangeSummary(), TestPodmanChangeSummary(), currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, DurableName() (+35 more)

### Community 24 - "hostReader"
Cohesion: 0.12
Nodes (25): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+17 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (125): TestLifecycleRolloutWait(), TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStderr(), captureStdout(), containsToken(), downloadEnvWithImage(), ordinaryHost() (+117 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.20
Nodes (5): progressElapsed(), progressTrim(), progress, progressMode, watchIO

### Community 28 - "Defaults"
Cohesion: 0.17
Nodes (26): build(), synthWorkflows(), statusPasswordFor(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture() (+18 more)

### Community 29 - "runKube"
Cohesion: 0.11
Nodes (42): gateKube(), kubeFixture(), nsItemJSON(), nsList(), removeResp(), runKube(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+34 more)

### Community 30 - "deployDocker"
Cohesion: 0.22
Nodes (19): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources(), deployDocker() (+11 more)

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (36): heapValue(), parseHeap(), stateCell(), Age(), Banner(), banner(), canonicalRef(), ExitCodeText() (+28 more)

### Community 32 - "correctness"
Cohesion: 0.11
Nodes (18): correctness, CR-0005 [blocker] [new] Passthrough map keys are written into application.yml unquoted and unchecked; a key holding a newline adds lines, CR-0006 [major] [new] Engine calls use the unexpanded env.yaml while validate and the rendered artifacts use the expanded one, CR-0007 [major] [new] remove can delete a namespace that still holds kinds its listing never asks for, also under --no-prompt, CR-0008 [major] [new] docker remove resolves every -env credential and refuses when one is not exported, CR-0009 [minor] [new] cipher and key-alias reach application.yml unquoted, CR-0010 [minor] [new] An aliased or non-mapping passthrough block is dropped without a message, CR-0011 [minor] [new] solace-defaults is passed through unchecked while mq-defaults rejects credentials and tool-managed keys (+10 more)

### Community 43 - "QuadletScope"
Cohesion: 0.18
Nodes (16): liveApp(), main(), run(), checkSubIDs(), liveHost(), rootlessHostErrors(), subIDRefusal(), 12. Runner seams (+8 more)

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
Cohesion: 0.21
Nodes (16): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), deref(), extraKeys() (+8 more)

### Community 48 - "actPodman"
Cohesion: 0.33
Nodes (13): absPath(), absResolver(), actPodman(), emit(), failFast(), fileReader(), genConfig(), genDocker() (+5 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Store"
Cohesion: 0.13
Nodes (9): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestTLSConfigStoreLookup(), TestStoreSecretNilSafe(), createdRole(), CertSources (+1 more)

### Community 52 - ".add"
Cohesion: 0.07
Nodes (57): TLSConfig, Toleration, checkCreatedSelection(), checkStoreDefinition(), badAffinityNode(), checkConnections(), checkContainerStores(), checkContainerTarget() (+49 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (42): Opts, Build(), containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildCreatedStores(), TestBuildLeaderElectionSharesBinderSecretNames() (+34 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (36): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+28 more)

### Community 56 - "certfixture.go"
Cohesion: 0.13
Nodes (27): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+19 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.16
Nodes (23): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+15 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.13
Nodes (37): TestLifecycleVerbsActInPlace(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+29 more)

### Community 61 - "Findings by area"
Cohesion: 0.12
Nodes (17): CR-0022 [major] [new] A folder with no workflow files validates, generates and deploys, CR-0023 [major] [new] download jar drops a dependency's subtree when its POM cannot be fetched and still reports success, CR-0024 [minor] [new] podman deploy removes stale secrets before the new unit is written, and writes the unit in place, CR-0025 [minor] [new] No engine call carries a deadline, CR-0030 [major] [new] No gate executes the generator page's JavaScript port, CR-0031 [minor] [new] The docs/test.md gate compares three counts and never a test name, CR-0032 [minor] [new] main_test.go holds the tests for six source files, so the test tree does not mirror the source tree, CR-0033 [nit] [new] Tests that need openssl, keytool or sh skip when the tool is absent, in CI too (+9 more)

### Community 63 - "parse_test.go"
Cohesion: 0.12
Nodes (25): ApplyStats(), ApplyTop(), ObjectExists(), ParseInspect(), ParsePods(), SecretDataKeys(), keys(), TestApplyStats() (+17 more)

### Community 64 - "storeFiles"
Cohesion: 0.19
Nodes (12): Option, AsCA(), RSAKey(), Validity(), TestResolvePodmanFilesOpensExistingStores(), createdCA(), storeFiles(), TestBuildStore() (+4 more)

### Community 65 - "Resolver"
Cohesion: 0.14
Nodes (30): DockerExisting, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestTargetMounts(), Resolver (+22 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (96): Contract, InstallScript(), KubernetesProbes(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), RunStatusScript(), approve() (+88 more)

### Community 68 - "rootless_test.go"
Cohesion: 0.12
Nodes (29): canIAsked(), indexOfCall(), newTargetRunner(), testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb() (+21 more)

### Community 69 - "parse.go"
Cohesion: 0.13
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), ParseImageDigest(), podComponents() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.11
Nodes (26): appendField(), deriveStatusPassword(), MergeService(), ParseDeployment(), TestParseDeploymentAndService(), View, groupOf(), JSON() (+18 more)

### Community 71 - "Env"
Cohesion: 0.11
Nodes (17): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), renderService(), Env, Kubernetes, Service (+9 more)

### Community 72 - "context.Context"
Cohesion: 0.13
Nodes (8): defaultAnswer(), scriptOut(), enableVirtualTerminal(), emit(), isCanI(), TestParseCommandExtraAllowed(), deployFns, fakeRunner

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.13
Nodes (15): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+7 more)

### Community 74 - "watchStatus"
Cohesion: 0.25
Nodes (10): sleepAfter(), TestWatchStatusAppendsSectionsWhenNotATerminal(), TestWatchStatusCancelledContextExitsZero(), TestWatchStatusDrawsNothingForARefreshCutShort(), TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected(), TestWatchStatusRedrawsUnderTheHeader(), TestWatchStatusRejectsAnIntervalTheLoopCannotUse(), TestWatchStatusReportsAFailedCollectionOnceAndCarriesOn() (+2 more)

### Community 75 - "ParseWorkflow"
Cohesion: 0.11
Nodes (20): extraKeyNames(), extraValue(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(), TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), TestResolveCarriesTheConnectionsExtraKeys() (+12 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image.go"
Cohesion: 0.15
Nodes (15): connectorRelease(), imageNameTag(), imageSatisfies(), releaseTag(), TestConnectorRelease(), TestEmbeddedListsTable(), TestImageNameTag(), TestImageSatisfies() (+7 more)

### Community 79 - "Model"
Cohesion: 0.11
Nodes (22): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+14 more)

### Community 80 - "conventions"
Cohesion: 0.14
Nodes (14): conventions, CR-0049 [minor] [new] The README quick start validates and deploys a sample that ships none of the four PKCS12 stores its env.yaml selects, CR-0050 [minor] [new] Messages, the starter files and the generator page point to a user guide that does not exist, CR-0051 [minor] [new] The shipped env.yaml and configuration.md describe behaviour that was removed, CR-0052 [minor] [new] Tests and docs cite the deleted review ledger, stage labels and numbered rules that no file defines, CR-0053 [minor] [new] Two validate messages and the four starter workflow files carry non-ASCII characters, CR-0054 [minor] [new] docs/test.md carries no support notice and the list that enforces it is explicit, so a new doc is exempt, CR-0055 [minor] [new] The README says every artifact runs non-root; run-as image runs the image's own user (+6 more)

### Community 81 - "testing.T"
Cohesion: 0.03
Nodes (114): TestAbsPath(), TestIsIndex(), TestOwnedObjectsCoverEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes(), TestSortIsByNameThenNamespace(), TestWatchFlagAcceptsBareAndInterval() (+106 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.13
Nodes (19): SecretRef, derivedPW(), kvRefs(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordIsDerivedFromTheCredentialsAndName(), TestResolveStatusPasswordNoMaterialIsRandomAndOverrideWins() (+11 more)

### Community 83 - "maven.go"
Cohesion: 0.11
Nodes (32): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchPOM(), groupPath(), highestStable() (+24 more)

### Community 85 - "rawDefaults"
Cohesion: 0.15
Nodes (16): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults, rawLeader (+8 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.15
Nodes (13): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (26): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+18 more)

### Community 88 - "RenderQuadlet"
Cohesion: 0.16
Nodes (17): leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput(), TestHealthcheckRunsTheStatusScript() (+9 more)

### Community 89 - "Side"
Cohesion: 0.11
Nodes (9): fixedLeaderNames(), TestLeaderElectionEffectiveMode(), Image, Cred, Side, TestSideUsernameSecretBothSystems(), placeholderSecretRef(), edaAdvisory() (+1 more)

### Community 90 - "solaceSide"
Cohesion: 0.21
Nodes (21): TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+13 more)

### Community 91 - "decodeCreate"
Cohesion: 0.21
Nodes (8): decodeCreate(), Secrets, TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, ImagePullSecret, StoreCreate, StoresSecret

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release, Workflows and numbering

### Community 93 - "vt_windows.go"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 94 - "Expand"
Cohesion: 0.22
Nodes (15): Expand(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone(), TestExpandDefaultsConnectionsMapEntry(), TestExpandDefaultVarSetUsesValue(), TestExpandDefaultVarUnsetUsesDefault() (+7 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (9): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+1 more)

### Community 96 - "buildLeaderElection"
Cohesion: 0.15
Nodes (18): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+10 more)

### Community 98 - "SolaceProps"
Cohesion: 0.32
Nodes (12): SolaceProps(), defaults(), existing(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(), TestSolacePropsUseStoreMountPaths() (+4 more)

### Community 100 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 101 - "Critical review - solmq-conn-util"
Cohesion: 0.15
Nodes (13): Accepted (not re-litigated), Coverage, CR-0062 [minor] [new] kubernetes restart after stop reports ok while no replica runs, CR-0063 [minor] [new] status --all --details reads kubectl top output as four columns, but --all-namespaces adds a fifth, CR-0064 [minor] [new] kubernetes status strings skip the control-character filter the docker and podman path applies, CR-0065 [minor] [new] The init-container download names its file after the raw URL, query string included, CR-0066 [minor] [new] A repeat podman remove fails at systemctl stop once the unit is gone, Critical review - solmq-conn-util (+5 more)

### Community 102 - "solmq-conn-util design notes"
Cohesion: 0.14
Nodes (16): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.5 Run-as and hardening, 9. Platform sections, 10. Status, logs and cli, 11. download jar, 1. Architecture (+8 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "ParseApplication"
Cohesion: 0.11
Nodes (23): applyConnections(), bindingLine(), detailError(), kubeItems(), ParseApplication(), splitKV(), degradedOf(), rawScriptOutput() (+15 more)

### Community 106 - ".add"
Cohesion: 0.19
Nodes (18): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock() (+10 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "libs/image_test.go"
Cohesion: 0.29
Nodes (12): loadImageLibs(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines(), TestLoadImageLibsSkipsUnsplittableLineWithoutFailing() (+4 more)

### Community 110 - "ParseDefaults"
Cohesion: 0.15
Nodes (13): TestApplicationLeaderElectionSessionPlaintext(), ParseDefaults(), TestParseCreateBlock(), TestParseTLSIsStrict(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError(), TestParseDefaultsFull() (+5 more)

### Community 111 - "Decisions on the 2026-11-11 critical review"
Cohesion: 0.18
Nodes (9): A new generator page, Consequences to check before the work starts, Decisions, Decisions on the 2026-11-11 critical review, Planned work the report does not hold, Shared module, Steps the owner runs, TestBaseName() (+1 more)

### Community 113 - ".stop"
Cohesion: 0.21
Nodes (10): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference, cli, cmd/solmq-conn-util (+2 more)

### Community 114 - "builtinList"
Cohesion: 0.25
Nodes (8): 9.4 podman, builtinList(), listRanges(), omitListProvenance(), TestBuiltinList(), TestBuiltinListPicksTheNearestLine(), TestEmbeddedListRange(), embeddedList

### Community 115 - "GenerateDocker"
Cohesion: 0.13
Nodes (19): DockerStore, mount, CreatedStore, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), TestStatusPasswordIsStableAcrossRenders() (+11 more)

### Community 116 - "gen_extra_test.go"
Cohesion: 0.08
Nodes (49): File, shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain() (+41 more)

### Community 117 - "validate_test.go"
Cohesion: 0.21
Nodes (17): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), defsWithStores(), leWorkflows(), mgmtSide(), TestDeployNameTooLong() (+9 more)

### Community 118 - "CommandRules"
Cohesion: 0.33
Nodes (5): 3. Deploy execution, CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 119 - "cicd"
Cohesion: 0.25
Nodes (8): cicd, CR-0034 [minor] [new] tag.yml lacks the template's tag-format check and latest handling, so any v* tag publishes as Latest, CR-0035 [minor] [new] CI runs all scan where the contract and the template say all cov scan, and tag.yml's header still lists cov, CR-0036 [minor] [new] The dev scripts are behind their contract: no race detector, no cov-queued skip, gofmt taken from PATH, CR-0037 [minor] [new] Local gates build the sibling checkout through go.work while CI builds the pin, and this repo's docs do not say so, CR-0038 [minor] [new] Dependabot covers GitHub Actions only and no scan is scheduled, so module advisories surface at tag time, CR-0039 [minor] [new] The release version stamp comes from git describe, not from the tag being released, CR-0040 [nit] [new] dev.sh takes the host OS from uname, which is not a GOOS under Git Bash

### Community 120 - "renderShapes"
Cohesion: 0.48
Nodes (7): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestSecretTargetsAreAbsolutePaths()

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "applyStatusAccess"
Cohesion: 0.29
Nodes (7): applyStatusAccess(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 123 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 124 - "GenerateKubernetes"
Cohesion: 0.11
Nodes (22): KubeOpts, Written(), TestGenerateKubernetesCarriesStatusScript(), TestGenerateKubernetesOmitNamespaceAndWritten(), TestGenerateKubernetesRefusesAnInjectedQuantity(), GenerateKubernetes(), SecretProbe, kubeConfigMapSizeError() (+14 more)

### Community 125 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 126 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 127 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 128 - "6. TLS stores"
Cohesion: 0.40
Nodes (5): 6.1 Existing stores, 6.2 Created stores, 6.3 Certificate validation, 6.4 Rotating a store, 6. TLS stores

### Community 129 - "build"
Cohesion: 0.40
Nodes (5): build, CR-0041 [major] [new] The kubernetes libs.download init container fetches jars with wget, with no checksum and with http accepted, CR-0042 [minor] [new] No artifact states a read-only root or a network mode, and DESIGN does not say the root must stay writable, CR-0043 [minor] [new] Quadlet values are written without percent escaping, CR-0044 [minor] [new] The status script passes the actuator Basic credential on wget's command line

### Community 130 - "design"
Cohesion: 0.40
Nodes (5): CR-0001 [minor] [new] remove on docker and podman needs a fully valid spec, so a broken workflow file blocks teardown, CR-0002 [minor] [new] Platform orchestration is three hand-sequenced pipelines in cmd, and their step order differs, CR-0003 [nit] [new] actTargets doc comment sits above the actionOpts type, CR-0004 [nit] [new] status takes its clock from time.Now() instead of the injected app clock, design

### Community 132 - "quality"
Cohesion: 0.40
Nodes (5): CR-0026 [minor] [new] Output goes to process-wide os.Stdout and os.Stderr; tests swap them without defer and none runs in parallel, CR-0027 [nit] [new] Small helpers are defined more than once, CR-0028 [nit] [new] checkSecretName's doc comment sits above checkJavaOptions, CR-0029 [nit] [new] Two package-level variables in libs are test seams written at run time, quality

### Community 133 - "expandValue"
Cohesion: 0.70
Nodes (4): expandMap(), expandString(), expandValue(), Expander

### Community 134 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 135 - "previousStoreSecrets"
Cohesion: 0.67
Nodes (4): previousStoreSecrets(), Stale objects after a config change, TestPodmanUnitStoreSecrets(), PodmanUnitStoreSecrets()

### Community 136 - "solmq-conn-util -- development guide"
Cohesion: 0.50
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 137 - "TestCredentialsNotWiredIsAnError"
Cohesion: 0.67
Nodes (4): kubeNoSecrets(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), vSolacePlainTCP()

### Community 138 - "tlsSelectingWF"
Cohesion: 0.67
Nodes (4): TestClusterOnlyStoreNeedsAHostSource(), TestRetiredStoresBlockTiers(), tlsSelectingWF(), certsOf()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 304 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **22 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `.step`, `runner_test.go`, `rootless_test.go`, `deploy_test.go`, `github.com/hafio/go-deploy-utility/runner.Runner`, `.stop`, `getting-started.md`, `.add`, `main_test.go`, `GenerateKubernetes`, `runKube`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 15 inferred relationships involving `testApp()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`testApp()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.run` be split into smaller, more focused modules?**
  _Cohesion score 0.1140819964349376 - nodes in this community are weakly interconnected._
- **Why does `Findings by area` connect `Findings by area` to `correctness`, `build`, `design`, `quality`, `Critical review - solmq-conn-util`, `conventions`, `cicd`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 156 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`dispatch()` has 156 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.14126984126984127 - nodes in this community are weakly interconnected._