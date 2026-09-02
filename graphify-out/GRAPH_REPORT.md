# Graph Report - solace-ibmmq-connector-helper  (2026-09-02)

## Corpus Check
- 98 files · ~302,086 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1972 nodes · 6001 edges · 105 communities (92 shown, 13 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 1050 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e07d3290`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- hasErr
- main_test.go
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
- main.go
- 12. Spring Actuator / Management Endpoint
- 2. IBM MQ Connection
- dispatch
- golden_test.go
- 6. JMS Binding-Level Options (Consumer & Producer)
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
- errExit
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
- Pending doc changes -- secrets mount path, stderr split, credentials guard
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
- ApplyTop
- attachRunner
- 10. Solace Connector — Management & Leader Election
- rawEnv
- Cmd
- 1. Solace Event Broker Connection
- 13. Logs: the lines behind the state
- runner.go
- removeNamespace
- testing.T
- runner_test.go
- 14. cli: a shell inside the instance
- EngineInspectJSON
- Workload
- 7. Solace Binding-Level Options (Consumer & Producer)
- Solace Message Headers Reference
- SafeToken

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 134 edges
2. `hasErr()` - 80 edges
3. `write()` - 62 edges
4. `captureStderr()` - 59 edges
5. `Download()` - 52 edges
6. `Build()` - 51 edges
7. `wfOK()` - 42 edges
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

## Communities (105 total, 13 thin omitted)

### Community 0 - "Build"
Cohesion: 0.11
Nodes (31): Build(), displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles() (+23 more)

### Community 1 - "validate.go"
Cohesion: 0.16
Nodes (39): Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources(), checkImage() (+31 more)

### Community 2 - "hasErr"
Cohesion: 0.08
Nodes (101): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+93 more)

### Community 3 - "main_test.go"
Cohesion: 0.09
Nodes (48): run(), captureStdout(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract() (+40 more)

### Community 4 - "Runner"
Cohesion: 0.09
Nodes (30): EngineImageInspectJSON(), EngineList(), EngineStats(), QuadletScope, Runner, KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON() (+22 more)

### Community 5 - "gen.go"
Cohesion: 0.07
Nodes (70): built, DockerPlan, File, KubeOpts, mount, NamedDoc, PodmanOpts, SecretRef (+62 more)

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

### Community 25 - "main.go"
Cohesion: 0.13
Nodes (39): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+31 more)

### Community 26 - "12. Spring Actuator / Management Endpoint"
Cohesion: 0.50
Nodes (4): 12. Spring Actuator / Management Endpoint, Sample Responses, Solace Binder Health Statuses, Solace Binder Metrics

### Community 27 - "2. IBM MQ Connection"
Cohesion: 0.50
Nodes (4): 2. IBM MQ Connection, Additional Properties (`additional-properties`), Core Connection Properties, JNDI Configuration (Alternative to Manual)

### Community 28 - "dispatch"
Cohesion: 0.09
Nodes (49): dispatch(), captureStderr(), TestAllowCommandFlagBadValueExitsUsageError(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+41 more)

### Community 29 - "golden_test.go"
Cohesion: 0.18
Nodes (22): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+14 more)

### Community 30 - "6. JMS Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 6. JMS Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (33): Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance, Workflow (+25 more)

### Community 32 - "statusCollector"
Cohesion: 0.10
Nodes (19): sortInstances(), actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), markMissing(), quoteAll() (+11 more)

### Community 44 - "errExit"
Cohesion: 0.21
Nodes (26): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+18 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 48 - "helperProcessArgv"
Cohesion: 0.11
Nodes (25): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+17 more)

### Community 49 - "Command details"
Cohesion: 0.07
Nodes (29): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+21 more)

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
Cohesion: 0.20
Nodes (26): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+18 more)

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
Nodes (30): ApplyStats(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV() (+22 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (24): imageMismatchNote(), imageNameTag(), imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote() (+16 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+17 more)

### Community 67 - "Pending doc changes -- secrets mount path, stderr split, credentials guard"
Cohesion: 0.18
Nodes (10): Also worth a sentence somewhere, Change A -- credentials moved off `/run/secrets`, on every platform, Change B -- stderr no longer corrupts parsed output, Change C -- validate warns when credentials are referenced but not wired, Counter, Files to update, Files to update, Files to update (+2 more)

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

### Community 73 - "10. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 10.1 `status container` -- the engine's view, 10.2 `status application` -- the connector's view, 10.3 `-d` / `--details`, 10.4 `--all`: find every instance by image, 10.5 `-w` / `--watch`, 10.6 `--output json`, 10.7 What the exit code means, and what each view costs, 10.8 The manual alternative (+3 more)

### Community 74 - "Podman"
Cohesion: 0.22
Nodes (15): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets, applyDockerDefaults(), applyMountDefaults(), applyPodmanDefaults(), Docker (+7 more)

### Community 75 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 76 - "Defaults"
Cohesion: 0.21
Nodes (17): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Syslog, LeaderElection, Logging (+9 more)

### Community 77 - "12. `download jar`"
Cohesion: 0.22
Nodes (9): 12. `download jar`, Flags and defaults, Image-aware omission, Integrity verification (sha1), `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, The two sets, `--url` overrides all resolution (+1 more)

### Community 78 - "5. Workflow file"
Cohesion: 0.29
Nodes (7): 5.1 Top-level, 5.2 `solace:` options, 5.3 `mq:` options, 5.4 Destinations, durable names, passthrough, 5.5 Event-driven guidance (warnings), 5.6 Reusable connections (`conn-ref`), 5. Workflow file

### Community 79 - "ExecArgv"
Cohesion: 0.15
Nodes (13): ExecArgv(), InstallScript(), RunStatusScript(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv() (+5 more)

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
Nodes (9): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), runShell(), shellInvocation(), splitAtSeparator(), Attacher (+1 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "TestShippedExamplesGenerateConfig"
Cohesion: 0.27
Nodes (9): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd() (+1 more)

### Community 87 - "libs.go"
Cohesion: 0.09
Nodes (30): reportDownload(), net/http.Client, net/http.Request, net/http.Response, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification() (+22 more)

### Community 88 - "ApplyTop"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 89 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 90 - "10. Solace Connector — Management & Leader Election"
Cohesion: 0.50
Nodes (4): 10. Solace Connector — Management & Leader Election, Active-Standby Configuration, Failover Configuration, Leader Election Mode

### Community 91 - "rawEnv"
Cohesion: 0.70
Nodes (4): workflowsFromRaw(), rawEnv, rawWorkflows, Workflows

### Community 92 - "Cmd"
Cohesion: 0.26
Nodes (11): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, resolveArgv0() (+3 more)

### Community 93 - "1. Solace Event Broker Connection"
Cohesion: 0.67
Nodes (3): 1. Solace Event Broker Connection, Core Connection Properties, Solace API Properties (`api-properties`)

### Community 94 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 96 - "runner.go"
Cohesion: 0.16
Nodes (16): canIVerb(), Docker(), LogsOpts, Kubernetes(), kubeVerb(), LogsArgv(), logsCommonFlags(), ParseCommand() (+8 more)

### Community 97 - "removeNamespace"
Cohesion: 0.50
Nodes (7): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), ownedNames(), removeNamespace(), nsItem

### Community 98 - "testing.T"
Cohesion: 0.09
Nodes (34): containsToken(), TestAbsPath(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal() (+26 more)

### Community 99 - "runner_test.go"
Cohesion: 0.10
Nodes (31): os.FileMode, PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), ScriptInstalled(), TestOSStreamRejectsEmptyAndUnresolvableArgv(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin() (+23 more)

### Community 100 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 102 - "EngineInspectJSON"
Cohesion: 0.50
Nodes (4): EngineInspectJSON(), TestEngineInspectJSONArgv(), TestEngineInspectJSONFailureNamesTheTargets(), TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing()

### Community 103 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 104 - "7. Solace Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 7. Solace Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 105 - "Solace Message Headers Reference"
Cohesion: 0.67
Nodes (3): Solace Binder Headers (`solace_scst_*`), Solace Message Headers Reference, Solace Message Headers (`solace_*`)

### Community 108 - "SafeToken"
Cohesion: 0.14
Nodes (12): CheckDeployCommand(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL() (+4 more)

## Knowledge Gaps
- **177 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Splitter`, `Defaults` (+172 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `dispatch` to `testing.T`, `main_test.go`, `Runner`, `nsList`, `main.go`, `write`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes` to `removeNamespace`, `validate.go`, `hasErr`, `Podman`, `Defaults`, `spec_test.go`, `Render`, `instances.go`, `rawEnv`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `resolveParentChain()` connect `maven.go` to `maven_test.go`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Are the 129 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 129 INFERRED edges - model-reasoned connections that need verification._
- **Are the 57 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 57 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _177 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.10984848484848485 - nodes in this community are weakly interconnected._