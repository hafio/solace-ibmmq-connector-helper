# Graph Report - solace-ibmmq-connector-helper  (2026-11-05)

## Corpus Check
- 142 files · ~431,496 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2752 nodes · 11481 edges · 130 communities (110 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1728 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ab88cd61`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- attachFiles
- .step
- validate_extra_test.go
- dockergen_test.go
- GenerateKubernetes
- deploy_test.go
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
- withPrompter
- crypto/x509.Certificate
- runAction
- Cmd
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- run
- .keystore
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
- rawDefaults
- Kubernetes
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
- Podman
- sinceFlag
- parse_test.go
- time.Time
- Resolver
- ScriptInstalled
- Image
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- Runner
- solmq-conn-util configuration reference
- Store
- statusreport_test.go
- Writer
- support_test.go
- libs/image.go
- buildLeaderElection
- helperProcessArgv
- createdStoreDockerSpec
- Resource
- maven.go
- Cred
- solmq-conn-util operations reference
- libs.go
- tlsSelectingWF
- .name
- solaceSide
- libs/image_test.go
- SolaceProps
- shellContentInfo
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- GenerateDocker
- .RunSplit
- .Write
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- PodmanSecretStoreName
- checkTargets
- javaCtx
- 7. status
- GeneratePodman
- solmq-conn-util command reference
- defsWithStores
- filenameFromEscapedPath
- CA
- Build
- solmq-conn-util abbreviations
- auto-complete
- 6. TLS stores
- LogsArgv
- status
- solmq-conn-util -- development guide
- cmd/solmq-conn-util
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- SafeToken

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 155 edges
2. `hasErr()` - 104 edges
3. `captureStderr()` - 82 edges
4. `write()` - 77 edges
5. `Build()` - 73 edges
6. `wfOK()` - 63 edges
7. `Download()` - 58 edges
8. `Runner` - 53 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `TLS` --references--> `existing()`  [INFERRED]
  docs/troubleshooting.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestCompletionModelMetadataComplete()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (130 total, 20 thin omitted)

### Community 0 - "attachFiles"
Cohesion: 0.31
Nodes (9): attachFiles(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv(), TestOSAttachReportsTheChildExitStatus() (+1 more)

### Community 1 - ".step"
Cohesion: 0.09
Nodes (21): sortInstances(), newProgress(), progressElapsed(), progressTrim(), clearScreen(), confirmInstall(), instanceNames(), markMissing() (+13 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (65): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+57 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (36): Input, Instance, composeEscape(), composeQuote(), flowList(), Render(), renderContentConfig(), renderHealthcheck() (+28 more)

### Community 4 - "GenerateKubernetes"
Cohesion: 0.09
Nodes (31): File, KubeOpts, shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), build() (+23 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.07
Nodes (67): storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), ManagementPort(), nameOnlySecret() (+59 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.17
Nodes (18): pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys() (+10 more)

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
Cohesion: 0.11
Nodes (45): fixedLeaderNames(), Side, postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide() (+37 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (93): TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestWorkflowExamplesMatchGoldenSpecs(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), canIDenied(), Preflight(), contains(), TestEngineListArgv() (+85 more)

### Community 14 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "withPrompter"
Cohesion: 0.09
Nodes (32): nsItemJSON(), nsList(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing() (+24 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.12
Nodes (23): day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard(), matches(), matchesKey() (+15 more)

### Community 19 - "runAction"
Cohesion: 0.11
Nodes (58): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), absPath(), absResolver(), actPodman(), allowCommandFlag() (+50 more)

### Community 21 - "Cmd"
Cohesion: 0.08
Nodes (30): 10. Status, logs and cli, 12. Runner seams, applyCmdEnv(), applyCmdInput(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats() (+22 more)

### Community 23 - "runner_test.go"
Cohesion: 0.12
Nodes (65): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+57 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (72): captureStderr(), downloadEnvWithImage(), manyWorkflowsDir(), TestAbsPath(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput() (+64 more)

### Community 26 - "issueWith"
Cohesion: 0.15
Nodes (27): extraMQ(), extraSolace(), runExtra(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError(), TestExtraKeysEnvSuffixWarns() (+19 more)

### Community 27 - ".stop"
Cohesion: 0.16
Nodes (14): newPipeReader(), progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+6 more)

### Community 28 - "Defaults"
Cohesion: 0.17
Nodes (25): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+17 more)

### Community 29 - "run"
Cohesion: 0.14
Nodes (18): run(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves() (+10 more)

### Community 30 - ".keystore"
Cohesion: 0.11
Nodes (24): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+16 more)

### Community 31 - "statusreport.go"
Cohesion: 0.13
Nodes (22): stateCell(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance (+14 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 43 - "podmanDeploy"
Cohesion: 0.10
Nodes (31): hostChecks(), podmanDeploy(), podmanRemove(), previousStoreSecrets(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks() (+23 more)

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
Cohesion: 0.14
Nodes (17): The spec generator (`solmq-conn-util-generator.html`), defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults (+9 more)

### Community 48 - "Kubernetes"
Cohesion: 0.11
Nodes (18): ownedNames(), Input, Instance, KV, StoreFile, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb() (+10 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.12
Nodes (17): cli, Command details, deploy, download, examples, generate, help, logs (+9 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.20
Nodes (13): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+5 more)

### Community 52 - ".add"
Cohesion: 0.10
Nodes (44): TLSConfig, checkCreatedSelection(), checkStoreDefinition(), badAffinityNode(), checkConnections(), checkContainerStores(), checkContainerTarget(), checkCred() (+36 more)

### Community 54 - "Application"
Cohesion: 0.11
Nodes (38): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+30 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (40): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+32 more)

### Community 56 - "certfixture.go"
Cohesion: 0.11
Nodes (27): Key, Option, params, Trusted, AsCA(), DER(), Cert, Issue() (+19 more)

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
Cohesion: 0.09
Nodes (44): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+36 more)

### Community 61 - "Podman"
Cohesion: 0.08
Nodes (28): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), decodeCreate(), Secrets, Service (+20 more)

### Community 63 - "parse_test.go"
Cohesion: 0.12
Nodes (25): EngineNamesByImage(), ObjectExists(), ParseApplication(), ParsePods(), SecretDataKeys(), splitKV(), keys(), TestEngineNamesByImage() (+17 more)

### Community 64 - "time.Time"
Cohesion: 0.15
Nodes (17): ApplyStats(), connectorIndex(), digestFrom(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), ParseImageDigest() (+9 more)

### Community 65 - "Resolver"
Cohesion: 0.10
Nodes (39): TestMountChecksLabelEachPathByItsKey(), DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless() (+31 more)

### Community 67 - "ScriptInstalled"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 68 - "Image"
Cohesion: 0.28
Nodes (9): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+1 more)

### Community 69 - "instanceFromInspect"
Cohesion: 0.22
Nodes (10): engineComponents(), kubeItems(), podComponents(), podMatchesImage(), sortComponents(), Component, kubeContainer, kubeContainerState (+2 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.17
Nodes (17): groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications(), renderContainerDetail() (+9 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "Runner"
Cohesion: 0.07
Nodes (42): actDocker(), actKubernetes(), checkImageCommand(), envPairs(), failAll(), preflight(), report(), warnIfPVUnchanged() (+34 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.13
Nodes (15): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+7 more)

### Community 74 - "Store"
Cohesion: 0.13
Nodes (9): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestTLSConfigStoreLookup(), TestStoreSecretNilSafe(), createdRole(), CertSources (+1 more)

### Community 75 - "statusreport_test.go"
Cohesion: 0.16
Nodes (10): joinCells(), NewTable(), TestAge(), TestImageMismatch(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing(), TestTableAlignsColumnsAndNeverPadsTheLast(), TestTableEmptyReportsNoRows() (+2 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image.go"
Cohesion: 0.15
Nodes (15): connectorRelease(), imageNameTag(), imageSatisfies(), releaseTag(), TestConnectorRelease(), TestEmbeddedListsTable(), TestImageNameTag(), TestImageSatisfies() (+7 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.10
Nodes (24): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), dropManaged() (+16 more)

### Community 80 - "helperProcessArgv"
Cohesion: 0.22
Nodes (12): helperProcessArgv(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput(), TestOSRunRejectsUnresolvableArgv0(), TestOSRunWiresStdinToChild(), TestOSRunWritesNothingToStderr() (+4 more)

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "Resource"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 85 - "Cred"
Cohesion: 0.15
Nodes (5): storePasswordValue(), Image, Cred, TestSideUsernameSecretBothSystems(), placeholderSecretRef()

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.15
Nodes (13): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "tlsSelectingWF"
Cohesion: 0.33
Nodes (7): TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF(), certsOf(), TestCreatedStoreSelections(), TestStoreDefinitions()

### Community 89 - ".name"
Cohesion: 0.25
Nodes (8): 9.4 podman, builtinList(), listRanges(), omitListProvenance(), TestBuiltinList(), TestBuiltinListPicksTheNearestLine(), TestEmbeddedListRange(), embeddedList

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "libs/image_test.go"
Cohesion: 0.29
Nodes (12): loadImageLibs(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines(), TestLoadImageLibsSkipsUnsplittableLineWithoutFailing() (+4 more)

### Community 92 - "SolaceProps"
Cohesion: 0.32
Nodes (12): SolaceProps(), defaults(), existing(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(), TestSolacePropsUseStoreMountPaths() (+4 more)

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
Cohesion: 0.14
Nodes (16): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.5 Run-as and hardening, 9. Platform sections, 11. download jar, 1. Architecture, 2. CLI (+8 more)

### Community 98 - "dispatch"
Cohesion: 0.09
Nodes (46): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance() (+38 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "GenerateDocker"
Cohesion: 0.14
Nodes (20): 7. Container posture, mount, CreatedStore, Mount, TestStoresSharingAFileNameMountApart(), TestTargetMounts(), GenerateDocker(), hostMounts() (+12 more)

### Community 102 - ".RunSplit"
Cohesion: 0.33
Nodes (5): TestOSRunSplitKeepsTheStreamsApart(), TestOSRunSplitRejectsEmptyAndUnresolvableArgv(), TestOSRunSplitWiresStdinAndEnv(), call, splitRunner

### Community 103 - ".Write"
Cohesion: 0.25
Nodes (8): How the suite is built, mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.08
Nodes (24): fileOwner(), fileOwner(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestMissingStoreFiles() (+16 more)

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
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release, Workflows and numbering

### Community 110 - "PodmanSecretStoreName"
Cohesion: 0.29
Nodes (8): TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), PodmanFileSecretNames(), podmanSecretRefs(), PodmanSecretStoreName(), PodmanStoreSecretName(), PodmanUnitStoreSecrets(), SecretRef

### Community 111 - "checkTargets"
Cohesion: 0.29
Nodes (8): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkRootlessRange(), checkRunAs(), checkTargets(), credentialsWired()

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 114 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 115 - "GeneratePodman"
Cohesion: 0.12
Nodes (40): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRejectsExtraKeysBesideConnRef() (+32 more)

### Community 116 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 117 - "defsWithStores"
Cohesion: 0.14
Nodes (23): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestMQStoreSelectionRequiresTLS(), TestPasswordConflictOnSameBinder(), TestPasswordConflictSolaceSide(), TestSolaceKeyAliasRequiresTCPSAndKeystore() (+15 more)

### Community 118 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 119 - "CA"
Cohesion: 0.27
Nodes (11): Validity(), TestResolvePodmanFilesOpensExistingStores(), createdCA(), hasIssue(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestCreatedStoreWarningsSurface() (+3 more)

### Community 120 - "Build"
Cohesion: 0.09
Nodes (32): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers() (+24 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 123 - "6. TLS stores"
Cohesion: 0.40
Nodes (5): 6.1 Existing stores, 6.2 Created stores, 6.3 Certificate validation, 6.4 Rotating a store, 6. TLS stores

### Community 124 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 125 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 126 - "solmq-conn-util -- development guide"
Cohesion: 0.50
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 127 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 128 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 136 - "SafeToken"
Cohesion: 0.12
Nodes (15): 3. Deploy execution, CheckDeployCommand(), checkImage(), checkSecurityUserRoles(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser() (+7 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 221 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util test catalogue` connect `solmq-conn-util test catalogue` to `getting-started.md`, `.Write`, `cmd/solmq-conn-util`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 150 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 150 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.08880666049953746 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util command reference` connect `solmq-conn-util command reference` to `Command Reference Details`, `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11655011655011654 - nodes in this community are weakly interconnected._