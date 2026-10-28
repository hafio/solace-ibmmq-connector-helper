# Graph Report - solace-ibmmq-connector-helper  (2026-10-28)

## Corpus Check
- 137 files · ~448,851 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2718 nodes · 11417 edges · 116 communities (95 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1738 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `59df1c88`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- statusscript_test.go
- statusCollector
- validate_extra_test.go
- dockergen_test.go
- GenerateKubernetes
- Side
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
- solmq-conn-util user guide
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
- storebuild.go
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
- main_test.go
- Image
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- JavaOptions
- Status Command User Guide
- Store
- podmanDeploy
- Writer
- support_test.go
- abbreviation_doc_test.go
- Build
- runner_test.go
- createdStoreDockerSpec
- actStatus
- maven.go
- 14. cli: a shell inside the instance
- jks.go
- 9. Secrets model
- libs.go
- TestSelectsStores
- builtinList
- solaceSide
- Deployment
- SolaceProps
- shellContentInfo
- Expand
- FormatScalar
- TestGoldenKubernetesNoSecrets
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- checkSpreadAndAffinity
- TestApplyStatusAccessAppendsAfterExistingUsers
- shippedExamplesRequest
- RunAsID
- .add
- yw
- 10. `download jar`
- javaCtx
- GeneratePodman
- buildLeaderElection
- defsWithStores
- .add
- GeneratePodman

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

## Communities (116 total, 21 thin omitted)

### Community 0 - "statusscript_test.go"
Cohesion: 0.14
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 1 - "statusCollector"
Cohesion: 0.17
Nodes (13): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), ResolveQuadletScope(), TestResolveQuadletScope(), MergeService() (+5 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (69): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey(), baseKubeDeploy(), baseKubeService() (+61 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (36): Input, Instance, composeEscape(), composeQuote(), flowList(), Render(), renderContentConfig(), renderHealthcheck() (+28 more)

### Community 4 - "GenerateKubernetes"
Cohesion: 0.11
Nodes (29): File, KubeOpts, build(), synthTLSWorkflowFiles(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate(), TestSelectedStoreIsTheRequestOnKubernetes(), TestStoresSharingAFileNameMountApart() (+21 more)

### Community 5 - "Side"
Cohesion: 0.25
Nodes (3): fixedLeaderNames(), Side, TestSideUsernameSecretBothSystems()

### Community 6 - "checkExtraKeys"
Cohesion: 0.18
Nodes (18): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+10 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.06
Nodes (36): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations, Build, Release checklist (+28 more)

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
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.22
Nodes (25): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flattenTarget() (+17 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (87): TestAbsPath(), TestIsIndex(), TestLogsPodmanReadsTheContainerNotTheJournal(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes(), TestSortIsByNameThenNamespace() (+79 more)

### Community 14 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (24): completionShells(), assertShellSafeName(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining(), renderedCompletions() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.12
Nodes (16): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state, 15. Notes and gotchas (+8 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.09
Nodes (27): params, jksCert(), readJKS(), certAlias(), day(), dedupe(), encryptedKey(), entry() (+19 more)

### Community 19 - "runAction"
Cohesion: 0.14
Nodes (37): resolveTarget(), runLifecycle(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+29 more)

### Community 21 - "Runner"
Cohesion: 0.06
Nodes (55): checkImageCommand(), firstLine(), probeOperatorSecrets(), Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), applyCmdEnv(), applyCmdInput() (+47 more)

### Community 23 - "runner_test.go"
Cohesion: 0.16
Nodes (38): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+30 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.10
Nodes (24): newPipeReader(), progressErase(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst(), withProgressClock() (+16 more)

### Community 28 - "Defaults"
Cohesion: 0.17
Nodes (26): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+18 more)

### Community 29 - "run"
Cohesion: 0.13
Nodes (20): main(), run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError() (+12 more)

### Community 30 - "storebuild.go"
Cohesion: 0.17
Nodes (18): 5.1 Variable expansion (`${VAR}`), 5. The config file and workflow discovery, 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`) (+10 more)

### Community 31 - "statusreport.go"
Cohesion: 0.09
Nodes (29): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+21 more)

### Community 32 - "deploy_test.go"
Cohesion: 0.07
Nodes (68): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), StoreFile (+60 more)

### Community 43 - "Runner"
Cohesion: 0.18
Nodes (27): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+19 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.15
Nodes (46): Option, AsCA(), CA(), KeyPKCS8(), NotCA(), PEM(), RSAKey(), Validity() (+38 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.23
Nodes (15): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot() (+7 more)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (55): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+47 more)

### Community 47 - "rawDefaults"
Cohesion: 0.15
Nodes (16): defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults, rawLeader (+8 more)

### Community 48 - "Kubernetes"
Cohesion: 0.12
Nodes (17): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb() (+9 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.10
Nodes (25): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+17 more)

### Community 52 - "Context"
Cohesion: 0.10
Nodes (30): Toleration, checkContainerStores(), checkContainerTarget(), CheckDeployCommand(), checkDocker(), checkImagePull(), checkKube(), checkLibs() (+22 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.12
Nodes (33): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+25 more)

### Community 56 - "certfixture.go"
Cohesion: 0.17
Nodes (21): Key, Trusted, DER(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+13 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.22
Nodes (13): abbrevFlagByShort(), abbrevTable(), renderAbbreviationDoc(), flagsLine(), flagSpan(), renderCommandsDoc(), tableCell(), flagAliases() (+5 more)

### Community 58 - "verbUsage"
Cohesion: 0.17
Nodes (22): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+14 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.06
Nodes (67): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), nsItemJSON(), nsList(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner() (+59 more)

### Community 61 - "Podman"
Cohesion: 0.09
Nodes (20): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), decodeCreate(), Secrets, TestCredCreateRemovedKeys(), applyDockerDefaults() (+12 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ApplyTop(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+22 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.12
Nodes (26): connectorRelease(), imageNameTag(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable() (+18 more)

### Community 65 - "Resolver"
Cohesion: 0.10
Nodes (41): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, CreatedStore, b64(), TestTargetMounts() (+33 more)

### Community 67 - "main_test.go"
Cohesion: 0.09
Nodes (44): dispatch(), captureStdout(), containsToken(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList() (+36 more)

### Community 68 - "Image"
Cohesion: 0.16
Nodes (10): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+2 more)

### Community 69 - "instanceFromInspect"
Cohesion: 0.12
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.13
Nodes (26): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+18 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "JavaOptions"
Cohesion: 0.22
Nodes (9): Instance, SecretRef, JavaOptions, Syslog, FileSecret, Input, Instance, Logging (+1 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Store"
Cohesion: 0.13
Nodes (9): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestTLSConfigStoreLookup(), TestStoreSecretNilSafe(), createdRole(), CertSources (+1 more)

### Community 75 - "podmanDeploy"
Cohesion: 0.14
Nodes (23): podmanDeploy(), podmanRemove(), previousStoreSecrets(), podmanHostChecks(), SecretRef, TestResolveCredentials(), PodmanPlan, PodmanFileSecretNames() (+15 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.53
Nodes (5): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 78 - "abbreviation_doc_test.go"
Cohesion: 0.33
Nodes (8): addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape(), targetSuffix(), cliTarget

### Community 79 - "Build"
Cohesion: 0.10
Nodes (30): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed() (+22 more)

### Community 80 - "runner_test.go"
Cohesion: 0.04
Nodes (86): How the suite is built, TestWriteMkdirError(), canIDenied(), canIProbe(), ExecArgv(), LogsOpts, InstallScript(), LogsArgv() (+78 more)

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "actStatus"
Cohesion: 0.18
Nodes (9): newProgress(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), TestSafeActuatorUser(), SafeActuatorUser(), statusOpts (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 85 - "jks.go"
Cohesion: 0.22
Nodes (5): joinAliases(), sniff(), SniffJKS(), TestSniff(), format

### Community 86 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (20): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+12 more)

### Community 88 - "TestSelectsStores"
Cohesion: 0.50
Nodes (5): kubeNoSecrets(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 89 - "builtinList"
Cohesion: 0.14
Nodes (18): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+10 more)

### Community 90 - "solaceSide"
Cohesion: 0.32
Nodes (14): propsNode(), TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildLeaderElectionSessionCarriesExtraKeys(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder() (+6 more)

### Community 91 - "Deployment"
Cohesion: 0.67
Nodes (3): 11. What gets generated, Deployment, Service

### Community 92 - "SolaceProps"
Cohesion: 0.28
Nodes (13): SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral() (+5 more)

### Community 93 - "shellContentInfo"
Cohesion: 0.50
Nodes (3): existingPKCS12, existingPKCS12Shell, shellContentInfo

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.22
Nodes (9): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+1 more)

### Community 96 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.18
Nodes (24): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+16 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (69): captureStderr(), downloadEnvWithImage(), TestAllowCommandFlagBadValueExitsUsageError(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+61 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 102 - "TestApplyStatusAccessAppendsAfterExistingUsers"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 104 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 105 - "RunAsID"
Cohesion: 0.09
Nodes (24): fileOwner(), fileOwner(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+16 more)

### Community 106 - ".add"
Cohesion: 0.18
Nodes (19): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock() (+11 more)

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 115 - "GeneratePodman"
Cohesion: 0.12
Nodes (36): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+28 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.18
Nodes (14): leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+6 more)

### Community 117 - "defsWithStores"
Cohesion: 0.21
Nodes (21): registryImage(), retiredNamesKube(), TestRetiredCreateNamesReportedOnlyByValidate(), defsWithStores(), leWorkflows(), mgmtSide(), TestDeployNameTooLong(), TestKeyAliasNeedsKeystore() (+13 more)

### Community 129 - ".add"
Cohesion: 0.16
Nodes (26): TLSConfig, checkCreatedSelection(), checkStoreDefinition(), checkConnections(), checkCred(), checkDefaultsCredentials(), checkImage(), checkJavaOptions() (+18 more)

### Community 130 - "GeneratePodman"
Cohesion: 0.12
Nodes (25): hostChecks(), mount, Mount, TestGeneratePodmanNeverReadsTheStores(), TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), TestToIssues() (+17 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 208 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `solmq-conn-util test catalogue`, `Status Command User Guide`, `10. `download jar``, `14. cli: a shell inside the instance`, `9. Secrets model`, `builtinList`, `Deployment`, `storebuild.go`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 149 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 149 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `statusscript_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13756613756613756 - nodes in this community are weakly interconnected._
- **Why does `12. Status: the container, the connector, or both` connect `Status Command User Guide` to `solmq-conn-util user guide`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10973084886128365 - nodes in this community are weakly interconnected._