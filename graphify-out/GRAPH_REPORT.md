# Graph Report - solace-ibmmq-connector-helper  (2026-11-05)

## Corpus Check
- 142 files · ~429,401 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2731 nodes · 11558 edges · 120 communities (101 shown, 19 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1754 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d3ffd848`
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
- statusscript_test.go
- completion_test.go
- withPrompter
- crypto/x509.Certificate
- runAction
- os.File
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- RunAsID
- .truststore
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
- attachRunner
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- deref
- checkTargets
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
- decodeCreate
- Resolver
- runner_test.go
- TestGeneratorPageKnownKeysInSync
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- Runner
- solmq-conn-util configuration reference
- Cred
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
- writeMode
- JavaEnv
- solaceSide
- SolaceProps
- validate_extra_test.go
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- dockerHostChecks
- CommandRules
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- hostReader
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- GeneratePodman
- GenerateKubernetes
- defsWithStores
- filenameFromEscapedPath
- CA
- Build
- solmq-conn-util abbreviations
- Security
- Workload
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 158 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 86 edges
4. `write()` - 80 edges
5. `Build()` - 73 edges
6. `bg()` - 63 edges
7. `wfOK()` - 63 edges
8. `Download()` - 58 edges
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
- `12. Runner seams` --references--> `readLog()`  [INFERRED]
  docs/DESIGN.md → cmd/solmq-conn-util/logs.go

## Import Cycles
- None detected.

## Communities (120 total, 19 thin omitted)

### Community 0 - "Store"
Cohesion: 0.12
Nodes (17): Store, StoreSecret, StoreSources, TLSConfig, TestStoreSecretParses(), TestTLSConfigStoreLookup(), checkCreatedSelection(), createdRole() (+9 more)

### Community 1 - ".step"
Cohesion: 0.11
Nodes (16): sortInstances(), bg(), newProgress(), progressTrim(), actStatus(), confirmInstall(), instanceNames(), markMissing() (+8 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.10
Nodes (63): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+55 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.08
Nodes (38): Input, Instance, composeEscape(), composeQuote(), flowList(), CreatedStore, Render(), renderContentConfig() (+30 more)

### Community 4 - "Config"
Cohesion: 0.17
Nodes (23): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain(), synthWorkflowFiles() (+15 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.07
Nodes (68): Input, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret (+60 more)

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
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (77): TestAbsPath(), TestIsIndex(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestSortIsByNameThenNamespace(), TestWatchFlagAcceptsBareAndInterval(), mustWrite() (+69 more)

### Community 14 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "withPrompter"
Cohesion: 0.10
Nodes (28): nsItemJSON(), nsList(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+20 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.11
Nodes (29): build(), day(), dedupe(), encryptedKey(), entry(), fingerprint(), Input, guard() (+21 more)

### Community 19 - "runAction"
Cohesion: 0.09
Nodes (70): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), absPath(), absResolver(), actDocker(), actKubernetes() (+62 more)

### Community 21 - "os.File"
Cohesion: 0.18
Nodes (5): scriptOut(), clearScreen(), enableVirtualTerminal(), emit(), fakeRunner

### Community 23 - "runner_test.go"
Cohesion: 0.15
Nodes (41): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+33 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (72): TestLifecycleVerbsRefuseBeforeActing(), dispatch(), captureStderr(), downloadEnvWithImage(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandSudoNamesTheSupportedAlternative(), TestConfiguredInstanceName(), TestDeployKubernetesProbesOperatorSecrets() (+64 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.17
Nodes (17): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow() (+9 more)

### Community 28 - "Defaults"
Cohesion: 0.11
Nodes (34): fixedLeaderNames(), leaderSession(), synthWorkflows(), Defaults, TestLeaderElectionEffectiveMode(), TestResolveCarriesTheConnectionsExtraKeys(), Side, Workflow (+26 more)

### Community 29 - "RunAsID"
Cohesion: 0.25
Nodes (10): podmanReader(), rootlessIDs(), sharedIDs(), TestPodmanReader(), Instance, JavaOptions, RunAsID, FileSecret (+2 more)

### Community 30 - ".truststore"
Cohesion: 0.22
Nodes (10): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+2 more)

### Community 31 - "statusreport.go"
Cohesion: 0.10
Nodes (29): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Banner(), banner(), Bytes() (+21 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.18
Nodes (24): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+16 more)

### Community 43 - "attachRunner"
Cohesion: 0.33
Nodes (10): How the suite is built, attached, attachRunner, fakeCall, fakeRunner, noAttach, queuedResp, queueRunner (+2 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (34): CA(), KeyPKCS8(), NotCA(), PEM(), TestCheckExistingStoreFromOpenSSL(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases() (+26 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "deref"
Cohesion: 0.11
Nodes (31): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), envTransforms(), isNull(), deref() (+23 more)

### Community 48 - "checkTargets"
Cohesion: 0.29
Nodes (8): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkRootlessRange(), checkRunAs(), checkTargets(), credentialsWired()

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.13
Nodes (21): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+13 more)

### Community 52 - ".add"
Cohesion: 0.11
Nodes (43): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources() (+35 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.11
Nodes (37): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+29 more)

### Community 56 - "certfixture.go"
Cohesion: 0.14
Nodes (23): Key, Option, params, Trusted, AsCA(), DER(), Cert, Issue() (+15 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.26
Nodes (11): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+3 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (23): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+15 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.12
Nodes (39): TestLifecycleVerbsActInPlace(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+31 more)

### Community 61 - "Kubernetes"
Cohesion: 0.08
Nodes (34): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), podmanDeploy(), podmanRemove(), previousStoreSecrets(), removeTarget() (+26 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (31): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+23 more)

### Community 64 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (37): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestTargetMounts() (+29 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (84): Contract, canIDenied(), ExecArgv(), InstallScript(), PodmanDeploy(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight() (+76 more)

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
Cohesion: 0.16
Nodes (33): checkImageCommand(), confirmNamespaceRemoval(), ownedNames(), removeNamespace(), firstLine(), probeOperatorSecrets(), 12. Runner seams, Docker() (+25 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Cred"
Cohesion: 0.13
Nodes (6): storePasswordValue(), Image, Cred, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), placeholderSecretRef()

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
Nodes (34): 9.4 podman, builtinList(), connectorRelease(), imageNameTag(), listRanges(), loadImageLibs(), omitListProvenance(), releaseTag() (+26 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.14
Nodes (18): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection(), TestSanitizeAndIsTCPS() (+10 more)

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
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "Cred"
Cohesion: 0.20
Nodes (12): SecretRef, dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), stableNames(), decodeAuths(), kubeEnvWithPull() (+4 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "writeMode"
Cohesion: 0.53
Nodes (6): skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits(), TestUnreadableMountsNamesWhatTheConnectorCannotRead(), writeMode()

### Community 89 - "JavaEnv"
Cohesion: 0.40
Nodes (5): JavaEnv(), NormalizeJavaOptions(), TestJavaEnvMergesTheTLSFlag(), TestParseEnvJavaOptionsAbsent(), TestParseEnvJavaOptionsSpellings()

### Community 90 - "solaceSide"
Cohesion: 0.26
Nodes (16): propsNode(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames(), TestBuildSolaceBinderPassesValidateCertificateThrough(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages() (+8 more)

### Community 92 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

### Community 93 - "validate_extra_test.go"
Cohesion: 0.14
Nodes (23): connDefaults(), kubeNoSecrets(), kubeWith(), TestCheckKubeSpreadAndAffinity(), TestCheckKubeTolerations(), TestClusterOnlyStoreNeedsAHostSource(), TestConnRefSolaceTopicSourceIsError(), TestConnRefStrictOnlyDestination() (+15 more)

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.15
Nodes (13): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), nodeToProps(), sameNode() (+5 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.13
Nodes (17): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.5 Run-as and hardening, 9. Platform sections, 10. Status, logs and cli, 11. download jar, 1. Architecture (+9 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (60): run(), captureStdout(), containsToken(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator() (+52 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "dockerHostChecks"
Cohesion: 0.40
Nodes (6): dockerHostChecks(), missingStoreFiles(), mountChecks(), TestMissingStoreFiles(), TestMountChecksLabelEachPathByItsKey(), HostMount

### Community 102 - "CommandRules"
Cohesion: 0.33
Nodes (5): 3. Deploy execution, CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "hostReader"
Cohesion: 0.39
Nodes (4): unreadableMounts(), hostReader, mountCheck, mountKind

### Community 106 - ".add"
Cohesion: 0.31
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "solmq-conn-util troubleshooting"
Cohesion: 0.29
Nodes (7): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 115 - "GeneratePodman"
Cohesion: 0.07
Nodes (41): 7. Container posture, mount, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), TestGeneratePodmanCarriesLogbackOnlyWithSyslog(), TestGeneratePodmanCarriesStatusScript() (+33 more)

### Community 116 - "GenerateKubernetes"
Cohesion: 0.09
Nodes (29): File, KubeOpts, synthTLSWorkflowFiles(), TestConfigRejectsSecretNameConflict(), TestGenerateKubernetesCarriesStatusScript(), TestGeneratePodmanNeverReadsTheStores(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestPlansCarryTheExpandedHostPathsTheyMount() (+21 more)

### Community 117 - "defsWithStores"
Cohesion: 0.18
Nodes (19): TestConnectionDefinitionValidation(), TestRemovedDefaultsKeysRejected(), TestSecurityUserRoles(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), TestStatusUserPasswordEnvCharset(), TestStatusUserReservedName(), defsWithStores(), leWorkflows() (+11 more)

### Community 118 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 119 - "CA"
Cohesion: 0.20
Nodes (15): Validity(), TestResolvePodmanFilesOpensExistingStores(), TestGenerateKubernetesRefusesAnInjectedQuantity(), createdCA(), createdEnv(), hasIssue(), storeFiles(), TestBuildStore() (+7 more)

### Community 120 - "Build"
Cohesion: 0.08
Nodes (35): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers() (+27 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 124 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 128 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 228 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util command reference` connect `Command Reference Details` to `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Are the 153 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 153 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Store` be split into smaller, more focused modules?**
  _Cohesion score 0.11692307692307692 - nodes in this community are weakly interconnected._
- **Why does `cmd/solmq-conn-util` connect `solmq-conn-util test catalogue` to `.stop`, `maven_test.go`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.10801393728222997 - nodes in this community are weakly interconnected._