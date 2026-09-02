# Graph Report - solace-ibmmq-connector-helper  (2026-09-02)

## Corpus Check
- 97 files · ~298,129 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1942 nodes · 5985 edges · 111 communities (97 shown, 14 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1103 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dae77310`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- consolidate_extra_test.go
- validate.go
- .Run
- dispatch
- Runner
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
- Solace PubSub+ Connector for IBM MQ — Configuration Guide
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- YAML & Spring Boot Essentials (For Non-Spring Users)
- Verification Checklist
- runAction
- 12. Spring Actuator / Management Endpoint
- 2. IBM MQ Connection
- main_test.go
- golden_test.go
- 6. JMS Binding-Level Options (Consumer & Producer)
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
- main.go
- solmq-conn-util -- Development Guide
- Download
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- spec_test.go
- render.go
- instances.go
- Build
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
- PodmanSecretCreate
- TestDownloadSetMapMatchesModel
- Kubernetes
- statusreport/render.go
- render
- nsList
- 10. Status: the container, the connector, or both
- Podman
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- Defaults
- 12. `download jar`
- 5. Workflow file
- ExecArgv
- Side
- 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)
- 8. Secrets model
- maven.go
- actShell
- solmq-conn-util abbreviations
- TestShippedExamplesGenerateConfig
- libs.go
- resolveInstanceSession
- attachRunner
- 10. Solace Connector — Management & Leader Election
- Env
- .Run
- 1. Solace Event Broker Connection
- 13. Logs: the lines behind the state
- allowCommandValue
- runner.go
- removeNamespace
- testing.T
- Preflight
- 14. cli: a shell inside the instance
- spec/image_test.go
- EngineInspectJSON
- statusCollector
- 7. Solace Binding-Level Options (Consumer & Producer)
- Solace Message Headers Reference
- checkSide
- EngineList
- CheckDeployCommand
- KubernetesGetJSON
- KubernetesTop

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 134 edges
2. `hasErr()` - 76 edges
3. `write()` - 62 edges
4. `captureStderr()` - 59 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 42 edges
8. `captureStdout()` - 40 edges
9. `Runner` - 40 edges
10. `wf()` - 36 edges

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

## Communities (111 total, 14 thin omitted)

### Community 0 - "consolidate_extra_test.go"
Cohesion: 0.16
Nodes (18): containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildLeaderElection() (+10 more)

### Community 1 - "validate.go"
Cohesion: 0.17
Nodes (34): Workflow, checkContainerTarget(), checkDocker(), checkDuplicateSources(), checkImage(), checkImagePull(), checkKeyAliasConflicts(), checkKube() (+26 more)

### Community 2 - ".Run"
Cohesion: 0.10
Nodes (96): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+88 more)

### Community 3 - "dispatch"
Cohesion: 0.10
Nodes (40): dispatch(), captureStdout(), containsToken(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList() (+32 more)

### Community 4 - "Runner"
Cohesion: 0.14
Nodes (19): podmanRemove(), QuadletScope, Runner, PodmanDeploy(), PodmanRemove(), PodmanSecretRemove(), ResolveQuadletScope(), SystemctlNRestarts() (+11 more)

### Community 5 - "gen.go"
Cohesion: 0.07
Nodes (67): podmanDeploy(), DockerPlan, File, KubeOpts, mount, NamedDoc, PodmanOpts, SecretRef (+59 more)

### Community 6 - "SolaceProps"
Cohesion: 0.26
Nodes (11): MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.09
Nodes (22): cli, cmd/solmq-conn-util, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen (+14 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "ParseEnv"
Cohesion: 0.16
Nodes (20): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestApplyDockerDefaultsFillsMissing() (+12 more)

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.05
Nodes (90): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+82 more)

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
Nodes (14): 11. `examples`, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery (+6 more)

### Community 18 - "Solace PubSub+ Connector for IBM MQ — Configuration Guide"
Cohesion: 0.11
Nodes (18): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 16. Spring Profiles & Config Locations, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options (+10 more)

### Community 19 - "Render"
Cohesion: 0.09
Nodes (32): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+24 more)

### Community 21 - "Render"
Cohesion: 0.12
Nodes (45): Input, Instance, KV, PullSecret, StoreFile, yw, leaderMode(), ManagementPort() (+37 more)

### Community 23 - "YAML & Spring Boot Essentials (For Non-Spring Users)"
Cohesion: 0.25
Nodes (8): Converting YAML Keys to Environment Variables, Property Placeholders (`${}`), Property Precedence (Highest → Lowest), Relaxed Binding (kebab-case vs camelCase), Single Binder vs Multi-Binder Syntax, The `undefined` Binder, YAML Indentation, YAML & Spring Boot Essentials (For Non-Spring Users)

### Community 24 - "Verification Checklist"
Cohesion: 0.33
Nodes (6): 1. Mechanical diff against the bundled property catalog (recommended), 2. Live actuator introspection (verifies the JSON samples in [Section 12](#12-spring-actuator--management-endpoint)), 3. GitHub source spot-check, 4. Transform-expression smoke test (for [Section 8](#8-solace-connector--workflow-configuration)), Verification Checklist, What to update after verifying

### Community 25 - "runAction"
Cohesion: 0.18
Nodes (29): runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage(), envFlag(), flagExit() (+21 more)

### Community 26 - "12. Spring Actuator / Management Endpoint"
Cohesion: 0.50
Nodes (4): 12. Spring Actuator / Management Endpoint, Sample Responses, Solace Binder Health Statuses, Solace Binder Metrics

### Community 27 - "2. IBM MQ Connection"
Cohesion: 0.50
Nodes (4): 2. IBM MQ Connection, Additional Properties (`additional-properties`), Core Connection Properties, JNDI Configuration (Alternative to Manual)

### Community 28 - "main_test.go"
Cohesion: 0.09
Nodes (51): captureStderr(), downloadEnvWithImage(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+43 more)

### Community 29 - "golden_test.go"
Cohesion: 0.18
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 30 - "6. JMS Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 6. JMS Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (40): heapValue(), parseHeap(), withUsed(), Banner(), banner(), Bytes(), canonicalRef(), Cores() (+32 more)

### Community 32 - "status.go"
Cohesion: 0.13
Nodes (13): actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), quoteAll(), watchStatus(), enableVirtualTerminal(), enableVirtualTerminal() (+5 more)

### Community 44 - "main.go"
Cohesion: 0.19
Nodes (27): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+19 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 48 - "runner_test.go"
Cohesion: 0.13
Nodes (29): call, os.FileMode, attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText() (+21 more)

### Community 49 - "Command details"
Cohesion: 0.07
Nodes (29): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+21 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.22
Nodes (18): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+10 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (32): ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe() (+24 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "instances.go"
Cohesion: 0.31
Nodes (19): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+11 more)

### Community 56 - "Build"
Cohesion: 0.09
Nodes (39): leaderNameFn, Opts, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), buildBundle() (+31 more)

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
Cohesion: 0.09
Nodes (40): podmanEnv(), podmanEnvSudo(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns() (+32 more)

### Community 61 - "Render"
Cohesion: 0.16
Nodes (20): breEscape(), Render(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderHeaderNamesEveryReportedFact() (+12 more)

### Community 62 - "actLogs"
Cohesion: 0.20
Nodes (9): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), Streamer, logsOpts, sinceFlag (+1 more)

### Community 63 - "parse_test.go"
Cohesion: 0.10
Nodes (31): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ParseApplication(), ParseInspect(), ParsePods(), splitKV() (+23 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (24): imageMismatchNote(), imageNameTag(), imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote() (+16 more)

### Community 65 - "parse.go"
Cohesion: 0.17
Nodes (24): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+16 more)

### Community 67 - "PodmanSecretCreate"
Cohesion: 0.40
Nodes (5): PodmanSecretCreate(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure(), TestPodmanSecretCreateSkipsCreateWhenRmFails()

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
Cohesion: 0.19
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+6 more)

### Community 73 - "10. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 10.1 `status container` -- the engine's view, 10.2 `status application` -- the connector's view, 10.3 `-d` / `--details`, 10.4 `--all`: find every instance by image, 10.5 `-w` / `--watch`, 10.6 `--output json`, 10.7 What the exit code means, and what each view costs, 10.8 The manual alternative (+3 more)

### Community 74 - "Podman"
Cohesion: 0.19
Nodes (18): TestTargetMounts(), targetMounts(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets, applyDockerDefaults(), applyMountDefaults() (+10 more)

### Community 75 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 76 - "Defaults"
Cohesion: 0.18
Nodes (18): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, checkSyslog(), LeaderElection (+10 more)

### Community 77 - "12. `download jar`"
Cohesion: 0.22
Nodes (9): 12. `download jar`, Flags and defaults, Image-aware omission, Integrity verification (sha1), `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, The two sets, `--url` overrides all resolution (+1 more)

### Community 78 - "5. Workflow file"
Cohesion: 0.29
Nodes (7): 5.1 Top-level, 5.2 `solace:` options, 5.3 `mq:` options, 5.4 Destinations, durable names, passthrough, 5.5 Event-driven guidance (warnings), 5.6 Reusable connections (`conn-ref`), 5. Workflow file

### Community 79 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 80 - "Side"
Cohesion: 0.23
Nodes (12): applyDest(), digitRun(), Side, yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess() (+4 more)

### Community 81 - "7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 7.0 image and timezone (shared by every platform), 7.1 kubernetes, 7.2 docker, 7.3 podman, 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)

### Community 82 - "8. Secrets model"
Cohesion: 0.40
Nodes (5): 8.1 Declaring a credential, 8.2 Mount names, 8.3 How each platform delivers them, 8.4 Registry credentials (pulling the image), 8. Secrets model

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (33): encoding/xml.Name, acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+25 more)

### Community 84 - "actShell"
Cohesion: 0.42
Nodes (8): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), Attacher, shellOpts

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "TestShippedExamplesGenerateConfig"
Cohesion: 0.27
Nodes (9): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd() (+1 more)

### Community 87 - "libs.go"
Cohesion: 0.09
Nodes (30): reportDownload(), net/http.Client, net/http.Request, net/http.Response, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification() (+22 more)

### Community 88 - "resolveInstanceSession"
Cohesion: 0.25
Nodes (8): instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), platformSpellings(), promptPlatformMenu(), resolvePlatform(), resolvePlatformAlias(), instanceRequest

### Community 89 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 90 - "10. Solace Connector — Management & Leader Election"
Cohesion: 0.50
Nodes (4): 10. Solace Connector — Management & Leader Election, Active-Standby Configuration, Failover Configuration, Leader Election Mode

### Community 91 - "Env"
Cohesion: 0.33
Nodes (8): instanceCommand(), removeTarget(), Defaults, Env, workflowsFromRaw(), rawEnv, rawWorkflows, Workflows

### Community 92 - ".Run"
Cohesion: 0.30
Nodes (10): context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), EngineImageInspectJSON(), Cmd, resolveArgv0() (+2 more)

### Community 93 - "1. Solace Event Broker Connection"
Cohesion: 0.67
Nodes (3): 1. Solace Event Broker Connection, Core Connection Properties, Solace API Properties (`api-properties`)

### Community 94 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 96 - "runner.go"
Cohesion: 0.16
Nodes (16): Docker(), LogsOpts, Kubernetes(), KubernetesListJSON(), KubernetesPodsJSON(), kubeVerb(), LogsArgv(), logsCommonFlags() (+8 more)

### Community 97 - "removeNamespace"
Cohesion: 0.50
Nodes (7): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), ownedNames(), removeNamespace(), nsItem

### Community 98 - "testing.T"
Cohesion: 0.08
Nodes (42): resolveTarget(), run(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+34 more)

### Community 99 - "Preflight"
Cohesion: 0.20
Nodes (10): canIVerb(), Preflight(), TestPreflightDockerArgvIsInfo(), TestPreflightExtraAllowedThreadsThrough(), TestPreflightFailureWrapsPlatformHint(), TestPreflightKubernetesArgvDeployNoNamespace(), TestPreflightKubernetesArgvRemoveWithNamespace(), TestPreflightPodmanArgvIsInfo() (+2 more)

### Community 100 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 101 - "spec/image_test.go"
Cohesion: 0.40
Nodes (4): TestImagePullSecretCreateDefaultsFalse(), TestImageRef(), TestImageRegistry(), TestRetiredPerPlatformImageStillParses()

### Community 102 - "EngineInspectJSON"
Cohesion: 0.50
Nodes (4): EngineInspectJSON(), TestEngineInspectJSONArgv(), TestEngineInspectJSONFailureNamesTheTargets(), TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing()

### Community 103 - "statusCollector"
Cohesion: 0.19
Nodes (12): sortInstances(), instanceNames(), markMissing(), MergeService(), ObjectExists(), ParseDeployment(), TestObjectExists(), TestParseDeploymentAndService() (+4 more)

### Community 104 - "7. Solace Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 7. Solace Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 105 - "Solace Message Headers Reference"
Cohesion: 0.67
Nodes (3): Solace Binder Headers (`solace_scst_*`), Solace Message Headers Reference, Solace Message Headers (`solace_*`)

### Community 106 - "checkSide"
Cohesion: 0.32
Nodes (8): checkConnections(), checkCred(), checkDefaultsCredentials(), checkLeaderElection(), checkSide(), checkSideCredentials(), checkTuple(), edaAdvisory()

### Community 107 - "EngineList"
Cohesion: 0.40
Nodes (5): EngineList(), EngineStats(), contains(), TestEngineListArgv(), TestEngineStatsArgv()

### Community 108 - "CheckDeployCommand"
Cohesion: 0.40
Nodes (5): CheckDeployCommand(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), Tokenize()

### Community 109 - "KubernetesGetJSON"
Cohesion: 0.67
Nodes (3): KubernetesGetJSON(), TestKubernetesGetJSONArgv(), TestKubernetesGetJSONMissingObjectIsAnError()

### Community 110 - "KubernetesTop"
Cohesion: 0.67
Nodes (3): KubernetesTop(), TestKubernetesTopArgv(), TestKubernetesTopWithoutMetricsAPIWraps()

## Knowledge Gaps
- **171 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `call`, `Defaults` (+166 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `dispatch` to `testing.T`, `Runner`, `nsList`, `main.go`, `completion.go`, `main_test.go`, `write`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Why does `Side` connect `Side` to `consolidate_extra_test.go`, `validate.go`, `.Run`, `checkSide`, `Cred`, `Defaults`, `Build`, `consolidate_test.go`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `resolveParentChain()` connect `maven.go` to `maven_test.go`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Are the 129 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 129 INFERRED edges - model-reasoned connections that need verification._
- **Are the 53 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 53 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _171 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.Run` be split into smaller, more focused modules?**
  _Cohesion score 0.09900990099009901 - nodes in this community are weakly interconnected._