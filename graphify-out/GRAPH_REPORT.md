# Graph Report - solace-ibmmq-connector-helper  (2026-10-01)

## Corpus Check
- 108 files · ~344,658 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2179 nodes · 8150 edges · 133 communities (94 shown, 39 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1172 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `94e73e13`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- GenerateKubernetes
- Context
- validate_extra_test.go
- dispatch
- GeneratePodman
- Defaults
- statusreport_test.go
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- run
- scan_test.go
- completion.go
- helperProcessArgv
- attachRunner
- completion_test.go
- solmq-conn-util user guide
- ExecArgv
- dockergen_test.go
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- Config
- RenderQuadlet
- main.go
- progress
- withProgressClock
- validate.go
- JavaOptions
- Resource
- statusreport.go
- statusCollector
- CLAUDE.md
- internal/consolidate
- internal/deploy
- internal/gen
- internal/render
- internal/scan
- internal/spec
- internal/tls
- internal/validate
- application.yml (Golden)
- Cred
- Runner
- solmq-conn-util -- Development Guide
- libs_test.go
- Kubernetes
- Command details
- solmq-conn-util.bash
- Model
- support_test.go
- Application
- instanceSession
- Build
- abbreviation_doc_test.go
- commands.go
- maven_test.go
- write
- Docker
- sinceFlag
- parse_test.go
- libs/image.go
- instanceFromInspect
- Defaults
- Env
- TestDownloadSetMapMatchesModel
- Issue
- Render
- statusreport/render_test.go
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- Writer
- 13. Logs: the lines behind the state
- 6. Workflow file
- resolvePullSecret
- TestParsingHelpersIgnoreAWarningOnStderr
- decodeCreate
- testing.T
- maven.go
- resolveStores
- solmq-conn-util abbreviations
- auto-complete
- libs.go
- SafeToken
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- buildLeaderElection
- .Write
- PodmanDeploy
- Cmd
- Expand
- LogsArgv
- runner_test.go
- repeatableName
- main_test.go
- resolveStatusPassword
- 9. Secrets model
- ParseWorkflow
- solmq-conn-util command reference
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- gen.go
- status
- TestGoldenKubernetesNoSecrets
- cmd/solmq-conn-util
- yw
- shippedExamplesRequest
- nodeToProps
- TestSanitizeAndIsTCPS
- XML
- tailFlag

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 140 edges
2. `hasErr()` - 87 edges
3. `captureStderr()` - 69 edges
4. `write()` - 65 edges
5. `Download()` - 56 edges
6. `wfOK()` - 54 edges
7. `Build()` - 53 edges
8. `imageOK()` - 45 edges
9. `captureStdout()` - 44 edges
10. `Runner` - 42 edges

## Surprising Connections (you probably didn't know these)
- `renderedCompletions()` --calls--> `render()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/statusreport/render_test.go
- `TestCompletionCoversModel()` --calls--> `contains()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/runner/runner_test.go
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go
- `TestDownloadImageMismatchReported()` --calls--> `run()`  [INFERRED]
  internal/libs/libs_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (133 total, 39 thin omitted)

### Community 0 - "GenerateKubernetes"
Cohesion: 0.17
Nodes (16): File, KubeOpts, build(), TestConfigRejectsSecretNameConflict(), TestGenerateKubernetesCarriesStatusScript(), TestGenValidateStoresWarning(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestResolveCredentials() (+8 more)

### Community 1 - "Context"
Cohesion: 0.22
Nodes (16): checkContainerTarget(), CheckDeployCommand(), checkDocker(), checkImagePull(), checkKube(), checkLibs(), checkPodman(), checkSecretName() (+8 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.07
Nodes (114): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+106 more)

### Community 3 - "dispatch"
Cohesion: 0.08
Nodes (51): dispatch(), captureStdout(), containsToken(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+43 more)

### Community 4 - "GeneratePodman"
Cohesion: 0.09
Nodes (31): DockerPlan, mount, NamedDoc, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform() (+23 more)

### Community 5 - "Defaults"
Cohesion: 0.11
Nodes (20): fixedLeaderNames(), defaultsFromRaw(), Defaults, TLSConfig, applyDest(), Side, nodePtr(), presentBlock() (+12 more)

### Community 6 - "statusreport_test.go"
Cohesion: 0.18
Nodes (9): joinCells(), NewTable(), TestAge(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing(), TestTableAlignsColumnsAndNeverPadsTheLast(), TestTableEmptyReportsNoRows(), ts() (+1 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "run"
Cohesion: 0.14
Nodes (18): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+10 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "helperProcessArgv"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 14 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (25): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+17 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "ExecArgv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

### Community 19 - "dockergen_test.go"
Cohesion: 0.11
Nodes (29): Input, Instance, composeEscape(), composeQuote(), Render(), renderContentConfig(), renderHealthcheck(), renderSecrets() (+21 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.09
Nodes (52): Input, Instance, KV, envValueQuote(), leaderMode(), ManagementPort(), NamespaceManifest(), quoteRes() (+44 more)

### Community 23 - "Config"
Cohesion: 0.30
Nodes (14): Config(), issuesContain(), synthWorkflowFiles(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder(), TestConfigRendersTransformHeadersUnderItsWorkflow(), TestConfigRendersTransformUnderItsWorkflow() (+6 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.16
Nodes (18): leaderLabels(), RenderQuadlet(), seconds(), systemdEnv(), fullInput(), minimalInput(), syslogInput(), TestHealthcheckRunsTheStatusScript() (+10 more)

### Community 25 - "main.go"
Cohesion: 0.11
Nodes (42): resolveTarget(), runLogs(), absPath(), absResolver(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains() (+34 more)

### Community 26 - "progress"
Cohesion: 0.12
Nodes (9): newProgress(), progressElapsed(), progressTrim(), clearScreen(), enableVirtualTerminal(), enableVirtualTerminal(), Attacher, progress (+1 more)

### Community 27 - "withProgressClock"
Cohesion: 0.19
Nodes (14): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+6 more)

### Community 28 - "validate.go"
Cohesion: 0.16
Nodes (33): synthWorkflows(), Workflow, checkMisplacedTransform(), checkTransforms(), checkConnections(), checkCred(), checkDefaultsCredentials(), checkDuplicateSources() (+25 more)

### Community 29 - "JavaOptions"
Cohesion: 0.31
Nodes (11): JavaOptions, YAMLKind(), checkContentType(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+3 more)

### Community 30 - "Resource"
Cohesion: 0.24
Nodes (10): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+2 more)

### Community 31 - "statusreport.go"
Cohesion: 0.12
Nodes (22): stateCell(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance (+14 more)

### Community 32 - "statusCollector"
Cohesion: 0.11
Nodes (18): noPodsFound(), sortInstances(), removeTarget(), actStatus(), confirmInstall(), instanceNames(), markMissing(), quoteAll() (+10 more)

### Community 43 - "Cred"
Cohesion: 0.13
Nodes (6): Security, Image, Cred, placeholderSecretRef(), rawSecurity, User

### Community 44 - "Runner"
Cohesion: 0.20
Nodes (27): actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit(), failFast(), genConfig() (+19 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (54): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+46 more)

### Community 48 - "Kubernetes"
Cohesion: 0.21
Nodes (8): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes, Syslog, Logging

### Community 49 - "Command details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.13
Nodes (20): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+12 more)

### Community 52 - "support_test.go"
Cohesion: 0.33
Nodes (8): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 54 - "Application"
Cohesion: 0.10
Nodes (29): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+21 more)

### Community 55 - "instanceSession"
Cohesion: 0.14
Nodes (30): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+22 more)

### Community 56 - "Build"
Cohesion: 0.08
Nodes (37): Opts, applyStatusAccess(), Build(), containsSub(), propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed() (+29 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "commands.go"
Cohesion: 0.16
Nodes (27): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+19 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (37): TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep() (+29 more)

### Community 60 - "write"
Cohesion: 0.09
Nodes (39): podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+31 more)

### Community 61 - "Docker"
Cohesion: 0.11
Nodes (22): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Secrets, Service, applyDockerDefaults() (+14 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (29): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys() (+21 more)

### Community 64 - "libs/image.go"
Cohesion: 0.07
Nodes (47): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+39 more)

### Community 65 - "instanceFromInspect"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 67 - "Env"
Cohesion: 0.33
Nodes (7): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), Env, instanceRequest

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "Issue"
Cohesion: 0.20
Nodes (14): issueWith(), runParsed(), runTransform(), TestTransformAndTransformHeadersCannotBeCombined(), TestTransformHeadersDeprecatedOnlyUnderLint(), TestTransformHeadersShapeErrors(), TestTransformHeadersValidBlockPasses(), TestTransformHeadersWarnings() (+6 more)

### Community 70 - "Render"
Cohesion: 0.18
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.14
Nodes (20): nsItemJSON(), nsList(), TestLogsPlatformMenuWhenSeveralSectionsArePresent(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestPlatformMenuOnMultipleSections(), TestRemoveDeclinedTouchesNothing(), TestRemoveDockerSeam() (+12 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.22
Nodes (9): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file (+1 more)

### Community 79 - "resolvePullSecret"
Cohesion: 0.28
Nodes (9): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+1 more)

### Community 80 - "TestParsingHelpersIgnoreAWarningOnStderr"
Cohesion: 0.11
Nodes (20): EngineImageInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON(), KubernetesTop(), contains() (+12 more)

### Community 81 - "decodeCreate"
Cohesion: 0.28
Nodes (5): decodeCreate(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 82 - "testing.T"
Cohesion: 0.06
Nodes (51): mustWrite(), TestWorkflowExamplesMatchGoldenSpecs(), TestWriteCreatesSkipsForces(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), TestLeaderElectionEffectiveMode(), ParseEnv(), TestParseEnvEmpty() (+43 more)

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (33): acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML(), highestStable() (+25 more)

### Community 84 - "resolveStores"
Cohesion: 0.33
Nodes (6): StoreFile, b64(), TestResolveStores(), resolveStores(), TestBaseName(), BaseName()

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "SafeToken"
Cohesion: 0.22
Nodes (7): TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL(), safeShellChars(), SafeToken(), allowCommandValue

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "buildLeaderElection"
Cohesion: 0.18
Nodes (14): leaderNameFn, secretFn, appendPassthrough(), binderOwner(), buildBundle(), buildLeaderElection(), TestAppendPassthroughCollision(), TestBuildLeaderElection() (+6 more)

### Community 91 - ".Write"
Cohesion: 0.40
Nodes (5): TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 92 - "PodmanDeploy"
Cohesion: 0.19
Nodes (12): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported(), TestPodmanDeploySystemModeNoUserFlag() (+4 more)

### Community 93 - "Cmd"
Cohesion: 0.19
Nodes (9): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS (+1 more)

### Community 94 - "Expand"
Cohesion: 0.18
Nodes (17): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+9 more)

### Community 95 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 96 - "runner_test.go"
Cohesion: 0.08
Nodes (38): canIVerb(), EngineInspectJSON(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), ScriptInstalled(), TestEngineInspectJSONArgv() (+30 more)

### Community 98 - "main_test.go"
Cohesion: 0.06
Nodes (66): captureStderr(), downloadEnvWithImage(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput() (+58 more)

### Community 99 - "resolveStatusPassword"
Cohesion: 0.40
Nodes (5): TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError(), resolveStatusPassword()

### Community 100 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "ParseWorkflow"
Cohesion: 0.16
Nodes (17): envTransforms(), ParseWorkflow(), TestConnRefSideMayTuneBinding(), TestParseWorkflowAmbiguousSystemAndDest(), TestParseWorkflowConnRef(), TestParseWorkflowEnabledDefaultsTrue(), TestParseWorkflowSolaceAndMQ(), TestParseWorkflowSyntaxError() (+9 more)

### Community 102 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "gen.go"
Cohesion: 0.10
Nodes (66): isClusterDefault(), isOurs(), namespaceOccupants(), DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5() (+58 more)

### Community 105 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 106 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.21
Nodes (20): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+12 more)

### Community 107 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "shippedExamplesRequest"
Cohesion: 0.67
Nodes (4): shippedExamplesRequest(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented()

### Community 112 - "nodeToProps"
Cohesion: 0.33
Nodes (6): TestFormatScalarQuoting(), TestNodeToProps(), FormatScalar(), needsQuote(), nodeToProps(), QuoteScalar()

### Community 113 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 131 - "XML"
Cohesion: 0.50
Nodes (4): TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

## Knowledge Gaps
- **131 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `renderedTransform`, `kubeDeployment` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 219 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **39 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Runner` connect `Runner` to `statusCollector`, `runner_test.go`, `dispatch`, `gen.go`, `TestParsingHelpersIgnoreAWarningOnStderr`, `ExecArgv`, `instanceSession`, `main.go`, `PodmanDeploy`, `Cmd`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **Why does `sinceFlag` connect `sinceFlag` to `gen.go`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Why does `allowCommandValue` connect `SafeToken` to `main.go`?**
  _High betweenness centrality (0.009) - this node is a cross-community bridge._
- **Are the 135 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 135 INFERRED edges - model-reasoned connections that need verification._
- **Are the 64 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 64 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.06727688787185354 - nodes in this community are weakly interconnected._