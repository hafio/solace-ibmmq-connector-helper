# Graph Report - solace-ibmmq-connector-helper  (2026-10-26)

## Corpus Check
- 132 files · ~426,797 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2633 nodes · 10867 edges · 123 communities (102 shown, 21 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1642 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a1b9a83c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- allowCommandValue
- statusCollector
- validate_extra_test.go
- dockergen_test.go
- Config
- Cred
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- deref
- completion_test.go
- solmq-conn-util user guide
- helperProcessArgv
- runAction
- deploy_test.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .step
- Defaults
- jks.go
- storebuild.go
- statusreport.go
- podmanDeploy
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
- Project Documentation Files
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- ExecArgv
- Application
- instanceSession
- certfixture.go
- abbreviation_doc_test.go
- verbUsage
- maven_test.go
- write
- Docker
- sinceFlag
- parse_test.go
- libs/image_test.go
- Resolver
- run
- Dispatch Model Consistency Tests
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- removeNamespace
- Status Command User Guide
- Side
- ParseWorkflow
- Writer
- 13. Logs: the lines behind the state
- 6. Workflow file
- Build
- Runner
- .Write
- ApplyTop
- maven.go
- targetMounts
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- go_pkg_io_fs
- builtinList
- solaceSide
- actStatus
- SolaceProps
- decodeCreate
- Expand
- FormatScalar
- TestGoldenKubernetesNoSecrets
- repeatableName
- dispatch
- tailFlag
- 9. Secrets model
- Env
- Command Reference Overview
- Project Overview And Quick Start
- status
- RunAsID
- .add
- support_test.go
- yw
- 10. `download jar`
- javaCtx
- TestCredentialsNotWiredWarning
- Preflight
- GeneratePodman
- buildLeaderElection
- defsWithStores
- statusreport_test.go
- filenameFromEscapedPath
- binderNames
- Context
- ResolveCredentials

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 148 edges
2. `hasErr()` - 102 edges
3. `captureStderr()` - 77 edges
4. `Build()` - 73 edges
5. `write()` - 71 edges
6. `wfOK()` - 61 edges
7. `Download()` - 58 edges
8. `wf()` - 50 edges
9. `imageOK()` - 49 edges
10. `Defaults` - 47 edges

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

## Communities (123 total, 21 thin omitted)

### Community 1 - "statusCollector"
Cohesion: 0.24
Nodes (7): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), Report, statusCollector

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.07
Nodes (41): renderConfigMap(), composeEscape(), composeQuote(), flowList(), Render(), renderContentConfig(), renderHealthcheck(), renderSecrets() (+33 more)

### Community 4 - "Config"
Cohesion: 0.14
Nodes (24): shippedExamplesRequest(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain() (+16 more)

### Community 5 - "Cred"
Cohesion: 0.11
Nodes (9): Security, Image, Cred, TestSideUsernameSecretBothSystems(), TestStoreSecretNilSafe(), TestUserSecretLiteralAndEnv(), placeholderSecretRef(), rawSecurity (+1 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.16
Nodes (19): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+11 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.10
Nodes (21): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+13 more)

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
Cohesion: 0.18
Nodes (30): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+22 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (97): newPipeReader(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestRootlessReaderSeesTheInvokersGroup(), TestUnreadableMountsJudgesRootByTheModeBits(), TestUnreadableMountsNamesWhatTheConnectorCannotRead(), writeMode(), TestWorkflowExamplesMatchGoldenSpecs() (+89 more)

### Community 14 - "deref"
Cohesion: 0.09
Nodes (32): checkCreateBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+24 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (24): completionShells(), assertShellSafeName(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining(), renderedCompletions() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.12
Nodes (16): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas, 1.1 Shell completion (+8 more)

### Community 18 - "helperProcessArgv"
Cohesion: 0.13
Nodes (23): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+15 more)

### Community 19 - "runAction"
Cohesion: 0.15
Nodes (34): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+26 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.07
Nodes (62): Input, KV, storeItem, needsQuote(), QuoteScalar(), envValueQuote(), PullSecret, StoreFile (+54 more)

### Community 23 - "runner_test.go"
Cohesion: 0.17
Nodes (37): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+29 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.05
Nodes (85): captureStderr(), downloadEnvWithImage(), podmanEnvSudo(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput() (+77 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".step"
Cohesion: 0.14
Nodes (17): progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst(), withProgressClock() (+9 more)

### Community 28 - "Defaults"
Cohesion: 0.21
Nodes (19): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+11 more)

### Community 29 - "jks.go"
Cohesion: 0.18
Nodes (6): jksCert(), joinAliases(), sniff(), TestSniff(), format, jksKey

### Community 30 - "storebuild.go"
Cohesion: 0.09
Nodes (37): params, 5.1 Variable expansion (`${VAR}`), 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, readJKS(), build(), certAlias() (+29 more)

### Community 31 - "statusreport.go"
Cohesion: 0.11
Nodes (25): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Age(), Banner(), banner(), Bytes(), canonicalRef() (+17 more)

### Community 32 - "podmanDeploy"
Cohesion: 0.14
Nodes (24): hostChecks(), podmanDeploy(), podmanRemove(), previousStoreSecrets(), podmanHostChecks(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), PodmanPlan (+16 more)

### Community 43 - "Runner"
Cohesion: 0.16
Nodes (29): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+21 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.19
Nodes (38): CA(), KeyPKCS8(), PEM(), Validity(), countIssues(), createdCA(), storeFiles(), TestBuildStore() (+30 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.23
Nodes (15): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot() (+7 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (48): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+40 more)

### Community 47 - "Project Documentation Files"
Cohesion: 0.44
Nodes (3): Build, Release (CI), solmq-conn-util -- development guide

### Community 48 - "Kubernetes"
Cohesion: 0.09
Nodes (19): ownedNames(), Instance, 11. What gets generated, renderService(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction() (+11 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.15
Nodes (19): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+11 more)

### Community 52 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.11
Nodes (37): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+29 more)

### Community 56 - "certfixture.go"
Cohesion: 0.15
Nodes (19): Key, Option, Trusted, AsCA(), DER(), Cert, Issue(), JKS() (+11 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.18
Nodes (16): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+8 more)

### Community 58 - "verbUsage"
Cohesion: 0.15
Nodes (21): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+13 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.11
Nodes (43): createdStoreDockerSpec(), envValue(), podmanEnv(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+35 more)

### Community 61 - "Docker"
Cohesion: 0.16
Nodes (14): Input, Instance, CreatedStore, dockerCreated(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets (+6 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (28): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+20 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.12
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 65 - "Resolver"
Cohesion: 0.13
Nodes (28): DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestResolvePodmanFiles(), TestResolveStores() (+20 more)

### Community 67 - "run"
Cohesion: 0.10
Nodes (26): main(), run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError() (+18 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "removeNamespace"
Cohesion: 0.20
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+6 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Side"
Cohesion: 0.09
Nodes (22): fixedLeaderNames(), leaderSession(), Store, StoreSources, TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side (+14 more)

### Community 75 - "ParseWorkflow"
Cohesion: 0.15
Nodes (16): extraKeyNames(), extraValue(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys(), TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), ParseWorkflow(), TestParseWorkflowAmbiguousSystemAndDest() (+8 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.33
Nodes (6): 6.1 Top-level, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file, Moving from `transform-headers` to `transform`

### Community 79 - "Build"
Cohesion: 0.08
Nodes (30): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers() (+22 more)

### Community 80 - "Runner"
Cohesion: 0.07
Nodes (43): checkImageCommand(), Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), applyCmdEnv(), applyCmdInput(), DockerImageCommand(), EngineImageInspectJSON() (+35 more)

### Community 81 - ".Write"
Cohesion: 0.25
Nodes (8): How the suite is built, mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 82 - "ApplyTop"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.11
Nodes (32): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchPOM(), groupPath(), highestStable() (+24 more)

### Community 84 - "targetMounts"
Cohesion: 0.18
Nodes (16): dockerHostChecks(), missingStoreFiles(), mountChecks(), TestMissingStoreFiles(), TestMountChecksLabelEachPathByItsKey(), mount, Mount, TestTargetMounts() (+8 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (26): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+18 more)

### Community 88 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 89 - "builtinList"
Cohesion: 0.14
Nodes (18): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+10 more)

### Community 90 - "solaceSide"
Cohesion: 0.26
Nodes (16): propsNode(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildSolaceBinderPassesValidateCertificateThrough(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages() (+8 more)

### Community 91 - "actStatus"
Cohesion: 0.17
Nodes (8): newProgress(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), Attacher, statusOpts, watchFlag

### Community 92 - "SolaceProps"
Cohesion: 0.28
Nodes (13): SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral() (+5 more)

### Community 93 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.24
Nodes (8): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), nodeToProps(), sameNode()

### Community 96 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (22): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+14 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (56): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+48 more)

### Community 100 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), PodmanServiceName(), Env (+1 more)

### Community 102 - "Command Reference Overview"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "Project Overview And Quick Start"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 105 - "RunAsID"
Cohesion: 0.25
Nodes (9): podmanReader(), rootlessIDs(), sharedIDs(), TestPodmanReader(), unreadableMounts(), RunAsID, hostReader, mountCheck (+1 more)

### Community 106 - ".add"
Cohesion: 0.19
Nodes (18): YAMLKind(), ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock() (+10 more)

### Community 107 - "support_test.go"
Cohesion: 0.53
Nodes (5): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 108 - "yw"
Cohesion: 0.24
Nodes (16): Bundle, Management, blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector() (+8 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 110 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 113 - "TestCredentialsNotWiredWarning"
Cohesion: 0.40
Nodes (6): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredWarning(), TestCredentialsWiredNoWarning(), TestNoCredentialsNoWarning(), vSolacePlainTCP()

### Community 114 - "Preflight"
Cohesion: 0.08
Nodes (24): canIVerb(), Docker(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), TestParseCommand(), TestParseCommandExtraAllowed() (+16 more)

### Community 115 - "GeneratePodman"
Cohesion: 0.10
Nodes (41): File, KubeOpts, synthTLSWorkflowFiles(), synthWorkflowFiles(), TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), TestGenerateKubernetesCarriesStatusScript() (+33 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.19
Nodes (14): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection(), testSecretRef() (+6 more)

### Community 117 - "defsWithStores"
Cohesion: 0.16
Nodes (21): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), TestStoresNotWiredWarning(), TestStoresWiredExistingNoWarning(), tlsSelectingWF() (+13 more)

### Community 118 - "statusreport_test.go"
Cohesion: 0.13
Nodes (13): ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestImageMismatch(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues() (+5 more)

### Community 122 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 123 - "binderNames"
Cohesion: 0.36
Nodes (9): binderNames(), binderOf(), eqStrs(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder() (+1 more)

### Community 129 - "Context"
Cohesion: 0.10
Nodes (53): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), CheckDeployCommand(), checkDocker() (+45 more)

### Community 130 - "ResolveCredentials"
Cohesion: 0.17
Nodes (12): SecretRef, TestResolveCredentials(), podmanSecretRefs(), ResolveCredentials(), resolvePullSecret(), stableNames(), toPodmanMount(), Mount (+4 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 206 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **21 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `9. Secrets model`, `Status Command User Guide`, `13. Logs: the lines behind the state`, `10. `download jar``, `Project Documentation Files`, `Kubernetes`, `6. Workflow file`, `builtinList`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 143 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 143 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11586538461538462 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util -- development guide` connect `Project Documentation Files` to `Runner`, `TestGoldenKubernetesNoSecrets`, `testcatalog_test.go`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 75 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 75 INFERRED edges - model-reasoned connections that need verification._
- **Should `dockergen_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07399577167019028 - nodes in this community are weakly interconnected._