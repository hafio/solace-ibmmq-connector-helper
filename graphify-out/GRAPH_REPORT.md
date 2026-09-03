# Graph Report - solace-ibmmq-connector-helper  (2026-09-03)

## Corpus Check
- 96 files · ~295,347 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1905 nodes · 5928 edges · 95 communities (80 shown, 15 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1043 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0d9eca4d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- hasErr
- main_test.go
- TestParsingHelpersIgnoreAWarningOnStderr
- gen.go
- SolaceProps
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- Kubernetes
- Scan
- completion.go
- EngineInspectJSON
- DurableName
- Expand
- solmq-conn-util -- User Guide
- ParseCommand
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- attachRunner
- ParseEnv
- main.go
- net/http.Response
- Docker
- dispatch
- golden_test.go
- ParseKubernetes
- statusreport.go
- status.go
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
- errExit
- solmq-conn-util -- Development Guide
- Download
- DEVELOPMENT.md
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- spec_test.go
- render.go
- instances.go
- consolidate.go
- Application
- consolidate_test.go
- maven_test.go
- write
- Render
- actLogs
- parse_test.go
- libs/image_test.go
- parse.go
- Defaults
- rawEnv
- TestDownloadSetMapMatchesModel
- spec/image_test.go
- statusreport/render.go
- render
- nsList
- 12. Status: the container, the connector, or both
- Defaults
- 10. `download jar`
- 6. Workflow file
- runner.go
- Side
- solmq-conn-util command reference
- maven.go
- actShell
- solmq-conn-util abbreviations
- TestShippedExamplesGenerateConfig
- libs.go
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- Cmd
- 13. Logs: the lines behind the state
- Runner
- removeNamespace
- testing.T
- 14. cli: a shell inside the instance
- statusCollector
- Write
- SafeToken

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 134 edges
2. `hasErr()` - 81 edges
3. `write()` - 62 edges
4. `captureStderr()` - 59 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 45 edges
8. `Runner` - 41 edges
9. `captureStdout()` - 40 edges
10. `imageOK()` - 39 edges

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

## Communities (95 total, 15 thin omitted)

### Community 0 - "Build"
Cohesion: 0.11
Nodes (31): Build(), displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+23 more)

### Community 1 - "validate.go"
Cohesion: 0.16
Nodes (39): Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources(), checkImage() (+31 more)

### Community 2 - "hasErr"
Cohesion: 0.08
Nodes (102): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+94 more)

### Community 3 - "main_test.go"
Cohesion: 0.09
Nodes (48): run(), captureStdout(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract() (+40 more)

### Community 4 - "TestParsingHelpersIgnoreAWarningOnStderr"
Cohesion: 0.11
Nodes (20): EngineImageInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON(), KubernetesTop(), contains() (+12 more)

### Community 5 - "gen.go"
Cohesion: 0.05
Nodes (93): podmanDeploy(), built, DockerPlan, File, KubeOpts, mount, NamedDoc, PodmanOpts (+85 more)

### Community 6 - "SolaceProps"
Cohesion: 0.26
Nodes (11): MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (24): cli, cmd/solmq-conn-util, Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples (+16 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "Kubernetes"
Cohesion: 0.16
Nodes (19): TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, ImagePullSecret, Kubernetes, Resources (+11 more)

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.05
Nodes (89): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+81 more)

### Community 13 - "EngineInspectJSON"
Cohesion: 0.50
Nodes (4): EngineInspectJSON(), TestEngineInspectJSONArgv(), TestEngineInspectJSONFailureNamesTheTargets(), TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing()

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.24
Nodes (18): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+10 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.11
Nodes (19): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+11 more)

### Community 18 - "ParseCommand"
Cohesion: 0.14
Nodes (14): Docker(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin() (+6 more)

### Community 19 - "Render"
Cohesion: 0.09
Nodes (32): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+24 more)

### Community 21 - "Render"
Cohesion: 0.11
Nodes (49): Input, Instance, KV, PullSecret, StoreFile, yw, leaderMode(), ManagementPort() (+41 more)

### Community 23 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 24 - "ParseEnv"
Cohesion: 0.16
Nodes (20): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestApplyDockerDefaultsFillsMissing() (+12 more)

### Community 25 - "main.go"
Cohesion: 0.14
Nodes (38): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+30 more)

### Community 27 - "Docker"
Cohesion: 0.22
Nodes (13): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), applyDockerDefaults(), applyPodmanDefaults(), Docker, LibsMount, Podman (+5 more)

### Community 28 - "dispatch"
Cohesion: 0.09
Nodes (49): dispatch(), captureStderr(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+41 more)

### Community 29 - "golden_test.go"
Cohesion: 0.18
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 30 - "ParseKubernetes"
Cohesion: 0.29
Nodes (7): applyKubeDefaults(), ParseKubernetes(), TestParseKubernetesError(), TestParseKubernetesFull(), TestParseKubernetesLoggingLibsDefaults(), TestParseKubernetesReplicasDefault(), TestParseKubernetesResources()

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (40): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+32 more)

### Community 32 - "status.go"
Cohesion: 0.13
Nodes (13): actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), quoteAll(), watchStatus(), enableVirtualTerminal(), enableVirtualTerminal() (+5 more)

### Community 44 - "errExit"
Cohesion: 0.22
Nodes (24): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+16 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 47 - "DEVELOPMENT.md"
Cohesion: 0.31
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "runner_test.go"
Cohesion: 0.08
Nodes (46): os.FileMode, Preflight(), ScriptInstalled(), attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes() (+38 more)

### Community 49 - "Command details"
Cohesion: 0.09
Nodes (23): auto-complete, cli, Command details, deploy, download, examples, generate, help (+15 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.11
Nodes (25): ParseDefaults(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty() (+17 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.19
Nodes (27): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+19 more)

### Community 56 - "consolidate.go"
Cohesion: 0.11
Nodes (26): leaderNameFn, Opts, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), buildBundle(), buildLeaderElection() (+18 more)

### Community 57 - "Application"
Cohesion: 0.25
Nodes (18): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+10 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "maven_test.go"
Cohesion: 0.25
Nodes (33): metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures(), pomXMLBody() (+25 more)

### Community 60 - "write"
Cohesion: 0.10
Nodes (34): downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+26 more)

### Community 61 - "Render"
Cohesion: 0.16
Nodes (20): breEscape(), Render(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderHeaderNamesEveryReportedFact() (+12 more)

### Community 62 - "actLogs"
Cohesion: 0.20
Nodes (9): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), Streamer, logsOpts, sinceFlag (+1 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ParseApplication(), ParseInspect(), ParsePods(), splitKV() (+22 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (24): imageMismatchNote(), imageNameTag(), imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote() (+16 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+17 more)

### Community 67 - "rawEnv"
Cohesion: 0.70
Nodes (4): workflowsFromRaw(), rawEnv, rawWorkflows, Workflows

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "spec/image_test.go"
Cohesion: 0.40
Nodes (4): TestImagePullSecretCreateDefaultsFalse(), TestImageRef(), TestImageRegistry(), TestRetiredPerPlatformImageStillParses()

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "render"
Cohesion: 0.28
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "nsList"
Cohesion: 0.13
Nodes (21): nsItemJSON(), nsList(), TestLogsPlatformFlagSkipsTheMenu(), TestLogsPlatformMenuWhenSeveralSectionsArePresent(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestPlatformMenuOnMultipleSections(), TestRemoveDeclinedTouchesNothing() (+13 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 12.10 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch`, 12.7 `--output json` (+3 more)

### Community 76 - "Defaults"
Cohesion: 0.21
Nodes (17): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, LeaderElection, Logging (+9 more)

### Community 77 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 78 - "6. Workflow file"
Cohesion: 0.29
Nodes (7): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (warnings), 6.6 Reusable connections (`conn-ref`), 6. Workflow file

### Community 79 - "runner.go"
Cohesion: 0.11
Nodes (23): canIVerb(), ExecArgv(), LogsOpts, InstallScript(), Kubernetes(), kubeVerb(), LogsArgv(), logsCommonFlags() (+15 more)

### Community 80 - "Side"
Cohesion: 0.23
Nodes (12): applyDest(), digitRun(), Side, yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess() (+4 more)

### Community 81 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (32): encoding/xml.Name, acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+24 more)

### Community 84 - "actShell"
Cohesion: 0.42
Nodes (9): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), runShell(), shellInvocation(), splitAtSeparator(), Attacher (+1 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "TestShippedExamplesGenerateConfig"
Cohesion: 0.27
Nodes (9): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd() (+1 more)

### Community 87 - "libs.go"
Cohesion: 0.12
Nodes (28): net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), filenameFromEscapedPath(), Input (+20 more)

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image and timezone (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 94 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 96 - "Runner"
Cohesion: 0.22
Nodes (13): podmanRemove(), QuadletScope, Runner, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart() (+5 more)

### Community 97 - "removeNamespace"
Cohesion: 0.50
Nodes (7): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), ownedNames(), removeNamespace(), nsItem

### Community 98 - "testing.T"
Cohesion: 0.09
Nodes (34): containsToken(), TestAbsPath(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal() (+26 more)

### Community 100 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 103 - "statusCollector"
Cohesion: 0.19
Nodes (12): sortInstances(), instanceNames(), markMissing(), MergeService(), ObjectExists(), ParseDeployment(), TestObjectExists(), TestParseDeploymentAndService() (+4 more)

### Community 108 - "SafeToken"
Cohesion: 0.14
Nodes (12): CheckDeployCommand(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL() (+4 more)

## Knowledge Gaps
- **130 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Splitter`, `Defaults` (+125 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `dispatch` to `Runner`, `testing.T`, `main_test.go`, `nsList`, `main.go`, `write`?**
  _High betweenness centrality (0.034) - this node is a cross-community bridge._
- **Why does `ParseApplication()` connect `parse_test.go` to `parse.go`, `statusCollector`, `statusreport.go`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes` to `removeNamespace`, `validate.go`, `hasErr`, `rawEnv`, `Defaults`, `Render`, `instances.go`, `ParseKubernetes`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Are the 129 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 129 INFERRED edges - model-reasoned connections that need verification._
- **Are the 58 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 58 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _130 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.10984848484848485 - nodes in this community are weakly interconnected._