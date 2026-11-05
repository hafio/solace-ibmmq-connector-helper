# Graph Report - solace-ibmmq-connector-helper  (2026-11-05)

## Corpus Check
- 142 files · ~433,762 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2767 nodes · 11573 edges · 128 communities (107 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1742 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `98e95768`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Side
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
- podmanRemove
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
- ParseWorkflow
- Resolver
- runner_test.go
- progress
- parse.go
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
- go_pkg_io_fs
- createdStoreDockerSpec
- jks.go
- maven.go
- getting-started.md
- Cred
- solmq-conn-util operations reference
- libs.go
- flattenBlock
- .name
- solaceSide
- libs/image_test.go
- SolaceProps
- TestSelectsStores
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- GenerateDocker
- CommandRules
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- misplacedEnvTransforms
- resolveStatusPassword
- javaCtx
- 7. status
- GeneratePodman
- XML
- defsWithStores
- filenameFromEscapedPath
- CA
- Build
- solmq-conn-util abbreviations
- Security
- StoreSources
- Workload
- cmd/solmq-conn-util
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- allowCommandValue

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 158 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 86 edges
4. `write()` - 80 edges
5. `Build()` - 73 edges
6. `wfOK()` - 63 edges
7. `Download()` - 58 edges
8. `Runner` - 53 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `Runner seams` --references--> `Runner`  [INFERRED]
  docs/parity.md → internal/runner/runner.go
- `How the suite is built` --references--> `TestHelperProcess()`  [INFERRED]
  docs/test.md → internal/runner/runner_test.go
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (128 total, 21 thin omitted)

### Community 0 - "Side"
Cohesion: 0.09
Nodes (22): fixedLeaderNames(), TLSConfig, TestLeaderElectionEffectiveMode(), TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+14 more)

### Community 1 - ".step"
Cohesion: 0.13
Nodes (17): anyIndex(), resolveIndexes(), sortInstances(), actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames() (+9 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.08
Nodes (38): Input, Instance, composeEscape(), composeQuote(), flowList(), CreatedStore, Render(), renderContentConfig() (+30 more)

### Community 4 - "Config"
Cohesion: 0.13
Nodes (25): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), build(), Config(), issuesContain() (+17 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.05
Nodes (78): Input, Instance, KV, storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote() (+70 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.18
Nodes (17): pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys() (+9 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.09
Nodes (22): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+14 more)

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

### Community 13 - "testing.T"
Cohesion: 0.07
Nodes (57): TestAbsPath(), TestIsIndex(), TestLogsPodmanReadsTheContainerNotTheJournal(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes(), TestSortIsByNameThenNamespace() (+49 more)

### Community 14 - "deref"
Cohesion: 0.61
Nodes (6): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), isNull(), deref()

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "withPrompter"
Cohesion: 0.09
Nodes (30): nsItemJSON(), nsList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+22 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.10
Nodes (22): params, certAlias(), day(), dedupe(), entry(), fingerprint(), guard(), matchesKey() (+14 more)

### Community 19 - "runAction"
Cohesion: 0.10
Nodes (69): resolveTarget(), runLifecycle(), runLogs(), absPath(), absResolver(), actDocker(), actKubernetes(), actPodman() (+61 more)

### Community 21 - "Cmd"
Cohesion: 0.16
Nodes (10): 10. Status, logs and cli, applyCmdEnv(), applyCmdInput(), Attacher, Cmd, Streamer, resolveArgv0(), fakeRunner (+2 more)

### Community 23 - "runner_test.go"
Cohesion: 0.11
Nodes (75): loadInstanceEnv(), isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles() (+67 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (21): leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput(), syslogInput() (+13 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (70): captureStderr(), downloadEnvWithImage(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandSudoNamesTheSupportedAlternative(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliRejectsWhatCannotMeanAnything(), TestConfiguredInstanceName() (+62 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.14
Nodes (14): newPipeReader(), progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+6 more)

### Community 28 - "Defaults"
Cohesion: 0.18
Nodes (25): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+17 more)

### Community 29 - "run"
Cohesion: 0.13
Nodes (19): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestDeployPodmanSetButEmptyCredentialFailsBeforeAnyWrite(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+11 more)

### Community 30 - ".keystore"
Cohesion: 0.14
Nodes (23): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+15 more)

### Community 31 - "statusreport.go"
Cohesion: 0.10
Nodes (30): heapValue(), parseHeap(), withUsed(), stateCell(), Banner(), banner(), Bytes(), canonicalRef() (+22 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (23): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+15 more)

### Community 43 - "podmanRemove"
Cohesion: 0.23
Nodes (12): podmanRemove(), previousStoreSecrets(), QuadletScope, PodmanDeploy(), PodmanLifecycle(), PodmanRemove(), ResolveQuadletScope(), TestPodmanDeployReloadThenStart() (+4 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (30): KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "rawDefaults"
Cohesion: 0.22
Nodes (12): defaultsFromRaw(), applyDest(), nodePtr(), presentBlock(), rawDefaults, rawLeader, rawManagement, rawMQ (+4 more)

### Community 48 - "Kubernetes"
Cohesion: 0.16
Nodes (11): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes, Syslog, checkSyslog() (+3 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.20
Nodes (13): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+5 more)

### Community 52 - ".add"
Cohesion: 0.11
Nodes (41): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkDuplicateSources(), checkImage(), checkImagePull() (+33 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (36): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+28 more)

### Community 55 - "instanceSession"
Cohesion: 0.09
Nodes (42): configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex(), kubeDiscovery() (+34 more)

### Community 56 - "certfixture.go"
Cohesion: 0.15
Nodes (19): Key, Option, Trusted, AsCA(), DER(), Cert, JKS(), KeyEC() (+11 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "verbUsage"
Cohesion: 0.16
Nodes (27): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+19 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.09
Nodes (46): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+38 more)

### Community 61 - "Podman"
Cohesion: 0.11
Nodes (22): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), JavaOptions, Secrets, Service (+14 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ApplyTop(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+22 more)

### Community 64 - "ParseWorkflow"
Cohesion: 0.15
Nodes (16): extraKeyNames(), extraValue(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(), TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), ParseWorkflow(), TestParseWorkflowAmbiguousSystemAndDest() (+8 more)

### Community 65 - "Resolver"
Cohesion: 0.12
Nodes (33): DockerExisting, DockerStore, StoreBuild, StoreUse, TestResolvePodmanFiles(), DockerPlan, Resolver, kubeStoresSizeError() (+25 more)

### Community 67 - "runner_test.go"
Cohesion: 0.04
Nodes (88): TestWriteMkdirError(), canIDenied(), canIProbe(), ExecArgv(), InstallScript(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight() (+80 more)

### Community 68 - "progress"
Cohesion: 0.22
Nodes (5): newProgress(), progressElapsed(), progressTrim(), progress, progressMode

### Community 69 - "parse.go"
Cohesion: 0.14
Nodes (24): connectorIndex(), digestFrom(), engineComponents(), EngineNamesByImage(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod() (+16 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.25
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "Runner"
Cohesion: 0.10
Nodes (42): checkImageCommand(), firstLine(), probeOperatorSecrets(), 12. Runner seams, Docker(), DockerImageCommand(), DockerImagePull(), DockerLifecycle() (+34 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Store"
Cohesion: 0.23
Nodes (6): Store, StoreSecret, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), StoreExisting

### Community 75 - "statusreport_test.go"
Cohesion: 0.15
Nodes (12): Age(), joinCells(), NewTable(), parseTime(), TestAge(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing() (+4 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image.go"
Cohesion: 0.18
Nodes (13): connectorRelease(), imageNameTag(), releaseTag(), TestConnectorRelease(), TestEmbeddedListsTable(), TestImageNameTag(), TestValidateImageVersionQualifiers(), trimJarClassifier() (+5 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.10
Nodes (24): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), dropManaged() (+16 more)

### Community 80 - "go_pkg_io_fs"
Cohesion: 0.17
Nodes (4): fileOwner(), fileOwner(), enableVirtualTerminal(), Write()

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "jks.go"
Cohesion: 0.20
Nodes (4): jksCert(), sniff(), TestSniff(), format

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "Cred"
Cohesion: 0.10
Nodes (16): SecretRef, dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), stableNames(), decodeAuths(), kubeEnvWithPull() (+8 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.12
Nodes (19): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference, 10. download jar, 11. What gets generated (+11 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "flattenBlock"
Cohesion: 0.52
Nodes (4): extraKeys(), flattenBlock(), isMergeKey(), keyValue

### Community 89 - ".name"
Cohesion: 0.20
Nodes (12): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+4 more)

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "libs/image_test.go"
Cohesion: 0.24
Nodes (14): loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedOmitListFullyParses(), TestImageSatisfies(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault() (+6 more)

### Community 92 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

### Community 93 - "TestSelectsStores"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.14
Nodes (14): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), nodeToProps() (+6 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.20
Nodes (10): 11. download jar, 1. Architecture, 2. CLI, 4. Kubernetes, 5. Docker, 6. Podman, 8. TLS, 9. Secrets (+2 more)

### Community 98 - "dispatch"
Cohesion: 0.10
Nodes (43): dispatch(), captureStdout(), containsToken(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList() (+35 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "GenerateDocker"
Cohesion: 0.20
Nodes (12): mount, Mount, TestTargetMounts(), HostMount, hostMounts(), targetMounts(), toDockerMount(), toDockerMounts() (+4 more)

### Community 102 - "CommandRules"
Cohesion: 0.33
Nodes (5): 3. Deploy execution, CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.33
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.12
Nodes (24): dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix() (+16 more)

### Community 106 - ".add"
Cohesion: 0.19
Nodes (18): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock() (+10 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 110 - "misplacedEnvTransforms"
Cohesion: 0.73
Nodes (5): appendTransformKeys(), documentMapping(), mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 111 - "resolveStatusPassword"
Cohesion: 0.40
Nodes (5): TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError(), resolveStatusPassword()

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 114 - "7. status"
Cohesion: 0.29
Nodes (7): 7.1 Container view, 7.2 Application view, 7.3 Installing the status script, 7.4 Flags, 7.5 Exit code and cost, 7.6 Instances this tool did not deploy, 7. status

### Community 115 - "GeneratePodman"
Cohesion: 0.08
Nodes (49): 7. Container posture, File, KubeOpts, b64(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript() (+41 more)

### Community 116 - "XML"
Cohesion: 0.40
Nodes (5): syslogXML(), TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

### Community 117 - "defsWithStores"
Cohesion: 0.15
Nodes (29): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF() (+21 more)

### Community 118 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 119 - "CA"
Cohesion: 0.20
Nodes (23): CA(), Issue(), TrustStorePKCS12(), Validity(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestCheckExistingStore(), TestCheckExistingStoreFromOpenSSL(), TestResolvePodmanFilesOpensExistingStores() (+15 more)

### Community 120 - "Build"
Cohesion: 0.09
Nodes (32): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers() (+24 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 124 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 127 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 128 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 224 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `TestHelperProcess()` connect `runner_test.go` to `testing.T`, `solmq-conn-util test catalogue`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 153 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 153 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Side` be split into smaller, more focused modules?**
  _Cohesion score 0.09269162210338681 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util operations reference` connect `solmq-conn-util operations reference` to `7. status`, `getting-started.md`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.12878787878787878 - nodes in this community are weakly interconnected._