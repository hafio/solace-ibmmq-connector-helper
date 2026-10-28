# Graph Report - solace-ibmmq-connector-helper  (2026-10-28)

## Corpus Check
- 137 files · ~445,298 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2707 nodes · 11375 edges · 119 communities (99 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1724 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fb581a72`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- statusscript_test.go
- statusCollector
- validate_extra_test.go
- dockergen_test.go
- GenerateKubernetes
- Cred
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- ParseWorkflow
- completion_test.go
- solmq-conn-util user guide
- crypto/x509.Certificate
- runAction
- Design notes
- Go Module Definition
- runner_test.go
- RenderQuadlet
- TestDownloadSetMapMatchesModel
- issueWith
- .stop
- Defaults
- main.go
- storebuild.go
- statusreport.go
- runner.go
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
- libs/image_test.go
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- validate.go
- Application
- instanceSession
- certfixture.go
- ExecArgv
- verbUsage
- maven_test.go
- write
- Podman
- sinceFlag
- parse_test.go
- libs/image_test.go
- Resolver
- main_test.go
- CA
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- removeNamespace
- Status Command User Guide
- Store
- QuadletScope
- Writer
- 13. Logs: the lines behind the state
- abbreviation_doc_test.go
- Build
- runner_test.go
- createdStoreDockerSpec
- actStatus
- maven.go
- TestGeneratorPageKnownKeysInSync
- jks.go
- Env
- libs.go
- go_pkg_io_fs
- builtinList
- solaceSide
- Cmd
- SolaceProps
- shellContentInfo
- Expand
- FormatScalar
- TestGoldenKubernetesNoSecrets
- repeatableName
- dispatch
- tailFlag
- TestDownloadSetMapMatchesModel
- WriteFile
- TestApplyStatusAccessAppendsAfterExistingUsers
- filenameFromEscapedPath
- shippedExamplesRequest
- RunAsID
- .add
- PodmanSecretRemove
- yw
- 10. `download jar`
- tlsSelectingWF
- javaCtx
- GeneratePodman
- buildLeaderElection
- defsWithStores
- .add
- PodmanSecretStoreName

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 153 edges
2. `hasErr()` - 102 edges
3. `captureStderr()` - 81 edges
4. `write()` - 75 edges
5. `Build()` - 73 edges
6. `wfOK()` - 61 edges
7. `Download()` - 58 edges
8. `Runner` - 52 edges
9. `Defaults` - 50 edges
10. `wf()` - 49 edges

## Surprising Connections (you probably didn't know these)
- `How the suite is built` --references--> `TestHelperProcess()`  [INFERRED]
  docs/test.md → internal/runner/runner_test.go
- `9.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/userguide.md → internal/tls/tls_test.go
- `Design notes` --references--> `resolveOneInstance()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/instances.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Build()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/consolidate/consolidate.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Validate()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (119 total, 20 thin omitted)

### Community 0 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 1 - "statusCollector"
Cohesion: 0.19
Nodes (11): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), MergeService(), ParseDeployment(), TestParseDeploymentAndService() (+3 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (69): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+61 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.09
Nodes (37): Input, Instance, composeEscape(), composeQuote(), flowList(), CreatedStore, Render(), renderContentConfig() (+29 more)

### Community 4 - "GenerateKubernetes"
Cohesion: 0.09
Nodes (33): KubeOpts, mount, Mount, build(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestResolveCredentials(), TestValidateCleanSpecStillPasses(), GenerateDocker() (+25 more)

### Community 5 - "Cred"
Cohesion: 0.12
Nodes (8): Security, Image, Cred, TestSideUsernameSecretBothSystems(), TestUserSecretLiteralAndEnv(), placeholderSecretRef(), rawSecurity, User

### Community 6 - "checkExtraKeys"
Cohesion: 0.29
Nodes (11): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.06
Nodes (37): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations, Build, Release checklist (+29 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.13
Nodes (45): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+37 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (76): newPipeReader(), TestAbsPath(), TestIsIndex(), TestOwnedNamesCoversEverythingThisReleaseCreates(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestProgressElapsedShapes(), TestSortIsByNameThenNamespace() (+68 more)

### Community 14 - "ParseWorkflow"
Cohesion: 0.06
Nodes (47): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), defaultsFromRaw(), envTransforms(), StoreSources, isNull() (+39 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (25): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+17 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.10
Nodes (21): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas, 1.1 Shell completion (+13 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.12
Nodes (20): certAlias(), day(), dedupe(), entry(), fingerprint(), guard(), matchesKey(), orderChain() (+12 more)

### Community 19 - "runAction"
Cohesion: 0.21
Nodes (29): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), downloadDeployedImage() (+21 more)

### Community 21 - "Design notes"
Cohesion: 0.08
Nodes (28): Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Attacher (+20 more)

### Community 23 - "runner_test.go"
Cohesion: 0.08
Nodes (112): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+104 more)

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
Cohesion: 0.11
Nodes (23): progressErase(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst(), withProgressClock() (+15 more)

### Community 28 - "Defaults"
Cohesion: 0.18
Nodes (24): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+16 more)

### Community 29 - "main.go"
Cohesion: 0.10
Nodes (17): absPath(), absResolver(), confirmRemove(), envPairs(), fileReader(), main(), platformSpellings(), promptPlatformMenu() (+9 more)

### Community 30 - "storebuild.go"
Cohesion: 0.14
Nodes (24): 5.1 Variable expansion (`${VAR}`), 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`) (+16 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (43): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Age(), Banner(), banner() (+35 more)

### Community 32 - "runner.go"
Cohesion: 0.12
Nodes (28): checkImageCommand(), firstLine(), probeOperatorSecrets(), Docker(), DockerImageCommand(), DockerImagePull(), DockerLifecycle(), LogsOpts (+20 more)

### Community 43 - "Runner"
Cohesion: 0.28
Nodes (21): actDocker(), actKubernetes(), actPodman(), emit(), errExit(), failAll(), failFast(), genConfig() (+13 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (30): KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.16
Nodes (20): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs() (+12 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (51): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+43 more)

### Community 47 - "libs/image_test.go"
Cohesion: 0.27
Nodes (13): loadImageLibs(), splitJarBasename(), TestEmbeddedListsMatchTheirConnectorLine(), TestEmbeddedListsTable(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines() (+5 more)

### Community 48 - "Kubernetes"
Cohesion: 0.17
Nodes (11): confirmNamespaceRemoval(), ownedNames(), removeNamespace(), KubernetesListJSON(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction() (+3 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.08
Nodes (26): auto-complete, cli, Command details, deploy, download, examples, generate, help (+18 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.20
Nodes (13): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+5 more)

### Community 52 - "validate.go"
Cohesion: 0.09
Nodes (53): ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget() (+45 more)

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.12
Nodes (33): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+25 more)

### Community 56 - "certfixture.go"
Cohesion: 0.15
Nodes (17): Key, Option, params, Trusted, AsCA(), DER(), Cert, JKS() (+9 more)

### Community 57 - "ExecArgv"
Cohesion: 0.14
Nodes (14): ExecArgv(), InstallScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+6 more)

### Community 58 - "verbUsage"
Cohesion: 0.17
Nodes (25): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+17 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.08
Nodes (52): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner() (+44 more)

### Community 61 - "Podman"
Cohesion: 0.08
Nodes (28): Instance, TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), JavaOptions, decodeCreate(), Secrets (+20 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (28): ApplyStats(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV(), keys() (+20 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.17
Nodes (14): connectorRelease(), imageNameTag(), imageSatisfies(), releaseTag(), TestConnectorRelease(), TestImageNameTag(), TestImageSatisfies(), TestValidateImageVersionQualifiers() (+6 more)

### Community 65 - "Resolver"
Cohesion: 0.10
Nodes (42): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, b64(), TestResolvePodmanFiles(), TestResolveStores() (+34 more)

### Community 67 - "main_test.go"
Cohesion: 0.06
Nodes (67): run(), captureStdout(), containsToken(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator() (+59 more)

### Community 68 - "CA"
Cohesion: 0.21
Nodes (23): CA(), Issue(), KeyStorePKCS12(), TrustStorePKCS12(), Validity(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestCheckExistingStore(), TestCheckExistingStoreFromOpenSSL() (+15 more)

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (19): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+11 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.24
Nodes (15): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+7 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "removeNamespace"
Cohesion: 0.22
Nodes (13): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+5 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Store"
Cohesion: 0.23
Nodes (6): Store, StoreSecret, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), StoreExisting

### Community 75 - "QuadletScope"
Cohesion: 0.20
Nodes (13): QuadletScope, PodmanDeploy(), PodmanLifecycle(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported() (+5 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "abbreviation_doc_test.go"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 79 - "Build"
Cohesion: 0.12
Nodes (26): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestBuildCipherConflictWarning(), TestBuildCreatedStores() (+18 more)

### Community 80 - "runner_test.go"
Cohesion: 0.08
Nodes (47): TestWriteMkdirError(), canIVerb(), PodmanSecretCreate(), Preflight(), attachFiles(), helperProcessArgv(), readFile(), regularFile() (+39 more)

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "actStatus"
Cohesion: 0.22
Nodes (7): newProgress(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), statusOpts, watchFlag

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (33): acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML(), highestStable() (+25 more)

### Community 84 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 85 - "jks.go"
Cohesion: 0.20
Nodes (4): jksCert(), sniff(), TestSniff(), format

### Community 86 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), PodmanServiceName(), Env (+1 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "go_pkg_io_fs"
Cohesion: 0.17
Nodes (4): fileOwner(), fileOwner(), enableVirtualTerminal(), Write()

### Community 89 - "builtinList"
Cohesion: 0.13
Nodes (20): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+12 more)

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "Cmd"
Cohesion: 0.23
Nodes (8): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS

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
Cohesion: 0.19
Nodes (23): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+15 more)

### Community 98 - "dispatch"
Cohesion: 0.07
Nodes (67): dispatch(), captureStderr(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing() (+59 more)

### Community 100 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.60
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 101 - "WriteFile"
Cohesion: 0.33
Nodes (5): TestWriteFileCreatesDirsAndMode(), TestWriteFileDoesNotTightenExistingFileMode(), TestWriteFileParentIsFileReturnsError(), TestWriteFileTargetIsDirectoryReturnsError(), WriteFile()

### Community 102 - "TestApplyStatusAccessAppendsAfterExistingUsers"
Cohesion: 0.33
Nodes (6): TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 103 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 104 - "shippedExamplesRequest"
Cohesion: 0.60
Nodes (5): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 105 - "RunAsID"
Cohesion: 0.13
Nodes (24): hostChecks(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader(), rootlessIDs(), sharedIDs() (+16 more)

### Community 106 - ".add"
Cohesion: 0.31
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 107 - "PodmanSecretRemove"
Cohesion: 0.40
Nodes (5): PodmanSecretRemove(), TestPodmanSecretRemoveBatchesNames(), TestPodmanSecretRemoveNoNamesIsNoop(), TestPodmanSecretRemoveRejectsUnsafeCommand(), TestPodmanSecretRemoveReportsFailure()

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 110 - "tlsSelectingWF"
Cohesion: 0.67
Nodes (4): TestClusterOnlyStoreNeedsAHostSource(), TestRetiredStoresBlockTiers(), tlsSelectingWF(), certsOf()

### Community 113 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 115 - "GeneratePodman"
Cohesion: 0.10
Nodes (46): File, Config(), issuesContain(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak() (+38 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.11
Nodes (23): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), TestBuildLeaderElection() (+15 more)

### Community 117 - "defsWithStores"
Cohesion: 0.20
Nodes (18): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestSolaceKeyAliasRequiresTCPSAndKeystore(), defsWithStores(), leWorkflows(), mgmtSide() (+10 more)

### Community 129 - ".add"
Cohesion: 0.12
Nodes (21): fixedLeaderNames(), TLSConfig, TestLeaderElectionEffectiveMode(), TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+13 more)

### Community 130 - "PodmanSecretStoreName"
Cohesion: 0.14
Nodes (15): SecretRef, TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), PodmanFileSecretNames(), podmanSecretRefs(), PodmanSecretStoreName(), PodmanStoreSecretName(), PodmanUnitStoreSecrets() (+7 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 207 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `solmq-conn-util test catalogue`, `Status Command User Guide`, `13. Logs: the lines behind the state`, `10. `download jar``, `runner_test.go`, `builtinList`, `storebuild.go`?**
  _High betweenness centrality (0.045) - this node is a cross-community bridge._
- **Are the 148 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 148 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `statusscript_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.13054187192118227 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util command reference` connect `.stop` to `Command Reference Details`, `solmq-conn-util test catalogue`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Are the 75 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 75 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.10807453416149068 - nodes in this community are weakly interconnected._