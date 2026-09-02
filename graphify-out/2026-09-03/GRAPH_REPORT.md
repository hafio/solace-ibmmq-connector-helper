# Graph Report - solace-ibmmq-connector-helper  (2026-09-02)

## Corpus Check
- 96 files · ~294,837 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1912 nodes · 5958 edges · 101 communities (86 shown, 15 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1053 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `291247c8`
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
- ParseEnv
- Scan
- completion.go
- podmangen_test.go
- DurableName
- Expand
- solmq-conn-util -- User Guide
- runner_test.go
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- attachRunner
- Env
- main.go
- net/http.Response
- absPath
- dispatch
- golden_test.go
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
- Side
- Runner
- solmq-conn-util -- Development Guide
- Download
- helperProcessArgv
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
- .add
- TestDownloadSetMapMatchesModel
- Kubernetes
- statusreport/render.go
- render
- nsList
- 12. Status: the container, the connector, or both
- Podman
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- Defaults
- 10. `download jar`
- 6. Workflow file
- ExecArgv
- spec.go
- solmq-conn-util command reference
- maven.go
- actShell
- solmq-conn-util abbreviations
- TestShippedExamplesGenerateConfig
- libs.go
- auto-complete
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- rawEnv
- Cmd
- 13. Logs: the lines behind the state
- 9. Secrets model
- runner.go
- removeNamespace
- testing.T
- status
- 14. cli: a shell inside the instance
- cmd/solmq-conn-util
- Workload
- Write
- allowCommandValue

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 134 edges
2. `hasErr()` - 81 edges
3. `write()` - 62 edges
4. `captureStderr()` - 59 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 43 edges
8. `Runner` - 41 edges
9. `captureStdout()` - 40 edges
10. `wf()` - 40 edges

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

## Communities (101 total, 15 thin omitted)

### Community 0 - "Build"
Cohesion: 0.11
Nodes (31): Build(), displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+23 more)

### Community 1 - "validate.go"
Cohesion: 0.15
Nodes (31): checkContainerTarget(), CheckDeployCommand(), checkDocker(), checkImage(), checkImagePull(), checkKube(), checkLibs(), checkPodman() (+23 more)

### Community 2 - "hasErr"
Cohesion: 0.08
Nodes (102): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+94 more)

### Community 3 - "main_test.go"
Cohesion: 0.09
Nodes (48): run(), captureStdout(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract() (+40 more)

### Community 4 - "TestParsingHelpersIgnoreAWarningOnStderr"
Cohesion: 0.09
Nodes (24): EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON(), KubernetesTop() (+16 more)

### Community 5 - "gen.go"
Cohesion: 0.07
Nodes (70): built, DockerPlan, File, KubeOpts, mount, NamedDoc, PodmanOpts, SecretRef (+62 more)

### Community 6 - "SolaceProps"
Cohesion: 0.29
Nodes (10): MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "ParseEnv"
Cohesion: 0.12
Nodes (24): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+16 more)

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.05
Nodes (89): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+81 more)

### Community 13 - "podmangen_test.go"
Cohesion: 0.17
Nodes (27): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs() (+19 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.24
Nodes (18): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+10 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "runner_test.go"
Cohesion: 0.11
Nodes (29): os.FileMode, ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), TestOSStreamRejectsEmptyAndUnresolvableArgv(), TestParseCommand(), TestParseCommandExtraAllowed() (+21 more)

### Community 19 - "Render"
Cohesion: 0.09
Nodes (32): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+24 more)

### Community 21 - "Render"
Cohesion: 0.11
Nodes (49): Input, Instance, KV, PullSecret, StoreFile, yw, leaderMode(), ManagementPort() (+41 more)

### Community 23 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 24 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), Defaults, Env (+1 more)

### Community 25 - "main.go"
Cohesion: 0.13
Nodes (39): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+31 more)

### Community 27 - "absPath"
Cohesion: 0.67
Nodes (3): absPath(), absResolver(), fileReader()

### Community 28 - "dispatch"
Cohesion: 0.09
Nodes (49): dispatch(), captureStderr(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+41 more)

### Community 29 - "golden_test.go"
Cohesion: 0.18
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (38): withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance (+30 more)

### Community 32 - "statusCollector"
Cohesion: 0.10
Nodes (18): sortInstances(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), watchStatus() (+10 more)

### Community 43 - "Side"
Cohesion: 0.11
Nodes (5): Image, Cred, Side, placeholderSecretRef(), edaAdvisory()

### Community 44 - "Runner"
Cohesion: 0.24
Nodes (25): actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit(), failFast(), genConfig() (+17 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 48 - "helperProcessArgv"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 49 - "Command details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (32): ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe() (+24 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.31
Nodes (19): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+11 more)

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
Cohesion: 0.13
Nodes (24): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ParseInspect(), ParsePods(), keys(), TestApplyStats() (+16 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (23): imageMismatchNote(), imageNameTag(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote(), TestImageNameTag() (+15 more)

### Community 65 - "parse.go"
Cohesion: 0.10
Nodes (35): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), heapValue() (+27 more)

### Community 67 - ".add"
Cohesion: 0.21
Nodes (19): synthWorkflows(), Workflow, checkConnections(), checkCred(), checkDefaultsCredentials(), checkDuplicateSources(), checkKeyAliasConflicts(), checkLeaderElection() (+11 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "Kubernetes"
Cohesion: 0.15
Nodes (18): TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, ImagePullSecret, Kubernetes, Resources (+10 more)

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

### Community 74 - "Podman"
Cohesion: 0.22
Nodes (15): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets, applyDockerDefaults(), applyMountDefaults(), applyPodmanDefaults(), Docker (+7 more)

### Community 75 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 76 - "Defaults"
Cohesion: 0.21
Nodes (17): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, LeaderElection, Logging (+9 more)

### Community 77 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 78 - "6. Workflow file"
Cohesion: 0.29
Nodes (7): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (warnings), 6.6 Reusable connections (`conn-ref`), 6. Workflow file

### Community 79 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 80 - "spec.go"
Cohesion: 0.31
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 81 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (33): encoding/xml.Name, imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML() (+25 more)

### Community 84 - "actShell"
Cohesion: 0.50
Nodes (7): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), shellOpts

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "TestShippedExamplesGenerateConfig"
Cohesion: 0.27
Nodes (9): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd() (+1 more)

### Community 87 - "libs.go"
Cohesion: 0.12
Nodes (28): net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), filenameFromEscapedPath(), Input (+20 more)

### Community 88 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image and timezone (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 91 - "rawEnv"
Cohesion: 0.70
Nodes (4): workflowsFromRaw(), rawEnv, rawWorkflows, Workflows

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 94 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 95 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 96 - "runner.go"
Cohesion: 0.12
Nodes (22): canIVerb(), Attacher, LogsOpts, QuadletScope, kubeVerb(), LogsArgv(), logsCommonFlags(), PodmanDeploy() (+14 more)

### Community 97 - "removeNamespace"
Cohesion: 0.42
Nodes (8): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), ownedNames(), removeNamespace(), Kubernetes(), nsItem

### Community 98 - "testing.T"
Cohesion: 0.09
Nodes (34): containsToken(), TestAbsPath(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal() (+26 more)

### Community 99 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all <container|application|all> [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application <container|application|all> [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container <container|application|all> [--details] [--watch] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 100 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 101 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 103 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

## Knowledge Gaps
- **130 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Splitter`, `Defaults` (+125 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `dispatch` to `testing.T`, `main_test.go`, `nsList`, `Runner`, `main.go`, `write`?**
  _High betweenness centrality (0.035) - this node is a cross-community bridge._
- **Why does `ParsePods()` connect `parse_test.go` to `statusCollector`, `parse.go`, `instances.go`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `ParseEnv()` connect `ParseEnv` to `gen.go`, `Podman`, `Runner`, `Defaults`, `spec_test.go`, `TestShippedExamplesGenerateConfig`, `Env`, `main.go`, `rawEnv`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Are the 129 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 129 INFERRED edges - model-reasoned connections that need verification._
- **Are the 58 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 58 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _130 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.10984848484848485 - nodes in this community are weakly interconnected._