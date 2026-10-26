# Graph Report - solace-ibmmq-connector-helper  (2026-10-26)

## Corpus Check
- 132 files · ~420,574 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2609 nodes · 10757 edges · 125 communities (105 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1619 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b3f61ccf`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ExecArgv
- .step
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
- withProgressClock
- Defaults
- jks.go
- storebuild.go
- statusreport.go
- attachRunner
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
- Store
- Application
- instanceSession
- certfixture.go
- abbreviation_doc_test.go
- verbUsage
- maven_test.go
- write
- Podman
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
- Preflight
- ResolveCredentials
- Writer
- 13. Logs: the lines behind the state
- .keystore
- Build
- Design notes
- rawDefaults
- ApplyTop
- maven.go
- targetMounts
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- statusscript_test.go
- builtinList
- solaceSide
- .Write
- SolaceProps
- JavaEnv
- Expand
- FormatScalar
- TestGoldenKubernetesNoSecrets
- repeatableName
- dispatch
- tailFlag
- WriteFile
- Env
- Command Reference Overview
- Project Overview And Quick Start
- cmd/solmq-conn-util
- RunAsID
- .add
- support_test.go
- yw
- 10. `download jar`
- TestCreatedStoreWarningsSurface
- TestCredentialsNotWiredWarning
- podmanDeploy
- GeneratePodman
- buildLeaderElection
- defsWithStores
- statusreport_test.go
- javaCtx
- Component
- status
- binderNames
- Context
- XML

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 148 edges
2. `hasErr()` - 98 edges
3. `captureStderr()` - 77 edges
4. `Build()` - 72 edges
5. `write()` - 71 edges
6. `wfOK()` - 59 edges
7. `Download()` - 58 edges
8. `wf()` - 49 edges
9. `imageOK()` - 47 edges
10. `Defaults` - 46 edges

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

## Communities (125 total, 20 thin omitted)

### Community 0 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 1 - ".step"
Cohesion: 0.08
Nodes (26): sortInstances(), newProgress(), progressElapsed(), progressTrim(), actStatus(), clearScreen(), confirmInstall(), instanceNames() (+18 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (61): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+53 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (36): Input, Instance, composeEscape(), composeQuote(), flowList(), Render(), renderContentConfig(), renderHealthcheck() (+28 more)

### Community 4 - "Config"
Cohesion: 0.10
Nodes (31): shippedExamplesRequest(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), build(), Config() (+23 more)

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
Nodes (47): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+39 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.18
Nodes (30): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+22 more)

### Community 13 - "testing.T"
Cohesion: 0.05
Nodes (68): TestWorkflowExamplesMatchGoldenSpecs(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), TestOSStreamRejectsEmptyAndUnresolvableArgv(), TestStoreConfigured(), ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored() (+60 more)

### Community 14 - "deref"
Cohesion: 0.19
Nodes (16): checkCreateBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), deref(), extraKeys(), flattenBlock() (+8 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (24): completionShells(), assertShellSafeName(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining(), renderedCompletions() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.10
Nodes (21): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas, 1.1 Shell completion (+13 more)

### Community 18 - "helperProcessArgv"
Cohesion: 0.11
Nodes (25): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+17 more)

### Community 19 - "runAction"
Cohesion: 0.15
Nodes (34): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+26 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.08
Nodes (54): storeItem, envValueQuote(), leaderMode(), ManagementPort(), NamespaceManifest(), quoteRes(), Render(), renderConfigMap() (+46 more)

### Community 23 - "runner_test.go"
Cohesion: 0.17
Nodes (37): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+29 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (72): captureStderr(), downloadEnvWithImage(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput() (+64 more)

### Community 26 - "issueWith"
Cohesion: 0.13
Nodes (28): fixedLeaderNames(), Side, extraMQ(), extraSolace(), runExtra(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - "withProgressClock"
Cohesion: 0.14
Nodes (16): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Defaults"
Cohesion: 0.19
Nodes (17): Defaults, TestLeaderElectionEffectiveMode(), checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost(), leaderConnRef() (+9 more)

### Community 29 - "jks.go"
Cohesion: 0.18
Nodes (6): jksCert(), joinAliases(), sniff(), TestSniff(), format, jksKey

### Community 30 - "storebuild.go"
Cohesion: 0.11
Nodes (24): params, day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard(), matches() (+16 more)

### Community 31 - "statusreport.go"
Cohesion: 0.14
Nodes (20): Banner(), banner(), Bytes(), canonicalRef(), Cores(), Instance, ImageMismatch(), InstanceBanner() (+12 more)

### Community 32 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 43 - "Runner"
Cohesion: 0.12
Nodes (37): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), checkImageCommand(), emit(), envPairs() (+29 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.24
Nodes (30): CA(), KeyPKCS8(), PEM(), TestBuildStoreReadsEverySourceByPosition(), Build(), aliases(), chain(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.23
Nodes (15): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot() (+7 more)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (55): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+47 more)

### Community 47 - "Project Documentation Files"
Cohesion: 0.44
Nodes (3): Build, Release (CI), solmq-conn-util -- development guide

### Community 48 - "Kubernetes"
Cohesion: 0.09
Nodes (23): ownedNames(), Input, Instance, KV, 11. What gets generated, StoreFile, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand() (+15 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.14
Nodes (19): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+11 more)

### Community 52 - "Store"
Cohesion: 0.16
Nodes (12): Store, StoreSources, TLSConfig, TestTLSConfigStoreLookup(), checkCreatedSelection(), createdRole(), checkConnections(), checkLeaderElection() (+4 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (36): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+28 more)

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
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.11
Nodes (44): createdStoreDockerSpec(), envValue(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPreflightFailureStopsBeforeWrite() (+36 more)

### Community 61 - "Podman"
Cohesion: 0.09
Nodes (20): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), decodeCreate(), Secrets, TestCredCreateRemovedKeys(), applyDockerDefaults() (+12 more)

### Community 63 - "parse_test.go"
Cohesion: 0.13
Nodes (23): EngineNamesByImage(), ObjectExists(), ParseApplication(), ParsePods(), splitKV(), keys(), TestEngineNamesByImage(), TestObjectExists() (+15 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.12
Nodes (26): connectorRelease(), imageNameTag(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable() (+18 more)

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (32): DockerStore, PodmanFile, StoreBuild, StoreUse, CreatedStore, TestResolvePodmanFiles(), TestResolveStores(), DockerPlan (+24 more)

### Community 67 - "run"
Cohesion: 0.12
Nodes (21): main(), run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError() (+13 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.15
Nodes (17): ApplyStats(), connectorIndex(), digestFrom(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), ParseImageDigest() (+9 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.18
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.29
Nodes (13): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+5 more)

### Community 72 - "removeNamespace"
Cohesion: 0.21
Nodes (13): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceProbeArgvIsOneQuery() (+5 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Preflight"
Cohesion: 0.20
Nodes (10): canIVerb(), Preflight(), TestPreflightDockerArgvIsInfo(), TestPreflightExtraAllowedThreadsThrough(), TestPreflightFailureWrapsPlatformHint(), TestPreflightKubernetesArgvDeployNoNamespace(), TestPreflightKubernetesArgvRemoveWithNamespace(), TestPreflightPodmanArgvIsInfo() (+2 more)

### Community 75 - "ResolveCredentials"
Cohesion: 0.11
Nodes (20): SecretRef, PullSecret, dockerConfigJSON(), TestResolveCredentials(), podmanSecretRefs(), ResolveCredentials(), resolvePullSecret(), stableNames() (+12 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - ".keystore"
Cohesion: 0.17
Nodes (18): 5.1 Variable expansion (`${VAR}`), 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`) (+10 more)

### Community 79 - "Build"
Cohesion: 0.08
Nodes (30): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+22 more)

### Community 80 - "Design notes"
Cohesion: 0.06
Nodes (47): Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), applyCmdEnv(), applyCmdInput(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList() (+39 more)

### Community 81 - "rawDefaults"
Cohesion: 0.22
Nodes (12): defaultsFromRaw(), applyDest(), nodePtr(), presentBlock(), rawDefaults, rawLeader, rawManagement, rawMQ (+4 more)

### Community 82 - "ApplyTop"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (36): imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+28 more)

### Community 84 - "targetMounts"
Cohesion: 0.19
Nodes (15): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), TestMissingStoreFiles(), TestMountChecksLabelEachPathByItsKey(), mount, Mount (+7 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (20): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+12 more)

### Community 88 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 89 - "builtinList"
Cohesion: 0.15
Nodes (17): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+9 more)

### Community 90 - "solaceSide"
Cohesion: 0.37
Nodes (13): TestBuildCarriesEachWorkflowsTransformBlocks(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder(), TestBuildMQDefaultsOverriddenByConnection(), TestBuildNestedExtraKeyThroughSharedConnRefIsQuiet() (+5 more)

### Community 91 - ".Write"
Cohesion: 0.25
Nodes (8): How the suite is built, mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 92 - "SolaceProps"
Cohesion: 0.28
Nodes (13): SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral() (+5 more)

### Community 93 - "JavaEnv"
Cohesion: 0.29
Nodes (7): JavaEnv(), NormalizeJavaOptions(), TestJavaEnvMergesTheTLSFlag(), TestParseEnvJavaOptionsAbsent(), TestParseEnvJavaOptionsSpellings(), checkJavaOptions(), javaOptionsProblem()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.20
Nodes (10): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), needsQuote(), nodeToProps() (+2 more)

### Community 96 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (22): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+14 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (55): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+47 more)

### Community 100 - "WriteFile"
Cohesion: 0.33
Nodes (5): TestWriteFileCreatesDirsAndMode(), TestWriteFileDoesNotTightenExistingFileMode(), TestWriteFileParentIsFileReturnsError(), TestWriteFileTargetIsDirectoryReturnsError(), WriteFile()

### Community 101 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), PodmanServiceName(), Env (+1 more)

### Community 102 - "Command Reference Overview"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "Project Overview And Quick Start"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 105 - "RunAsID"
Cohesion: 0.11
Nodes (18): fileOwner(), fileOwner(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestPodmanReader() (+10 more)

### Community 106 - ".add"
Cohesion: 0.13
Nodes (35): synthWorkflows(), Workflow, checkStoreDefinition(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions() (+27 more)

### Community 107 - "support_test.go"
Cohesion: 0.53
Nodes (5): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 110 - "TestCreatedStoreWarningsSurface"
Cohesion: 0.36
Nodes (9): Validity(), createdCA(), storeFiles(), TestBuildStore(), TestBuildStoreRejects(), TestCreatedStoreWarningsSurface(), TestResolveDockerStoresLimits(), TestValidateReadsWhatDeployReads() (+1 more)

### Community 113 - "TestCredentialsNotWiredWarning"
Cohesion: 0.40
Nodes (6): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredWarning(), TestCredentialsWiredNoWarning(), TestNoCredentialsNoWarning(), vSolacePlainTCP()

### Community 114 - "podmanDeploy"
Cohesion: 0.11
Nodes (30): podmanDeploy(), podmanRemove(), previousStoreSecrets(), podmanHostChecks(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), PodmanPlan, PodmanFileSecretNames() (+22 more)

### Community 115 - "GeneratePodman"
Cohesion: 0.10
Nodes (39): File, KubeOpts, b64(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform() (+31 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.13
Nodes (20): leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), propsNode(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision() (+12 more)

### Community 117 - "defsWithStores"
Cohesion: 0.21
Nodes (17): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), defsWithStores(), leWorkflows(), mgmtSide() (+9 more)

### Community 118 - "statusreport_test.go"
Cohesion: 0.15
Nodes (11): stateCell(), ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVBuilderDropsEmptyValues(), TestTableAlignsColumnsAndNeverPadsTheLast() (+3 more)

### Community 119 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 120 - "Component"
Cohesion: 0.22
Nodes (10): engineComponents(), kubeItems(), podComponents(), podMatchesImage(), sortComponents(), Component, kubeContainer, kubeContainerState (+2 more)

### Community 121 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 123 - "binderNames"
Cohesion: 0.39
Nodes (8): binderNames(), binderOf(), eqStrs(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder(), TestBinderFieldsCarryStablePlaceholders()

### Community 129 - "Context"
Cohesion: 0.13
Nodes (24): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkImagePull() (+16 more)

### Community 130 - "XML"
Cohesion: 0.50
Nodes (4): TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 205 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `Status Command User Guide`, `13. Logs: the lines behind the state`, `10. `download jar``, `Project Documentation Files`, `Kubernetes`, `.keystore`, `builtinList`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 143 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 143 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ExecArgv` be split into smaller, more focused modules?**
  _Cohesion score 0.1111111111111111 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util -- development guide` connect `Project Documentation Files` to `Design notes`, `TestGoldenKubernetesNoSecrets`, `testcatalog_test.go`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 71 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 71 INFERRED edges - model-reasoned connections that need verification._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.0783744557329463 - nodes in this community are weakly interconnected._