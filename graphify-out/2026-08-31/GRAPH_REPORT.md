# Graph Report - solace-ibmmq-connector-helper  (2026-08-31)

## Corpus Check
- 91 files · ~269,179 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1818 nodes · 5515 edges · 106 communities (90 shown, 16 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1014 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dae77310`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- consolidate_extra_test.go
- validate.go
- .Run
- testing.T
- PodmanDeploy
- gen.go
- SolaceProps
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- Kubernetes
- Scan
- completion.go
- Runner
- DurableName
- Expand
- solmq-conn-util -- User Guide
- Solace PubSub+ Connector for IBM MQ — Configuration Guide
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- YAML & Spring Boot Essentials (For Non-Spring Users)
- Verification Checklist
- Podman
- 12. Spring Actuator / Management Endpoint
- 2. IBM MQ Connection
- 16. Spring Profiles & Config Locations
- main.go
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
- Side
- runAction
- solmq-conn-util -- Development Guide
- Download
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- spec_test.go
- render.go
- names.go
- consolidate.go
- Build
- consolidate_test.go
- maven_test.go
- write
- Render
- parse_test.go
- ScriptInstalled
- libs/image_test.go
- parse.go
- Defaults
- runner.go
- TestDownloadSetMapMatchesModel
- TestShippedExamplesGenerateConfig
- statusreport/render.go
- render
- main_test.go
- 10. Status: the container, the connector, or both
- solmq-conn-util command reference
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- Defaults
- 12. `download jar`
- 5. Workflow file
- auto-complete
- spec.go
- 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)
- 8. Secrets model
- maven.go
- ParseEnv
- solmq-conn-util abbreviations
- ParseCommand
- libs.go
- Env
- Image
- 10. Solace Connector — Management & Leader Election
- ApplyTop
- .Run
- 1. Solace Event Broker Connection
- 13. Logs: the lines behind the state
- allowCommandValue
- 7. Solace Binding-Level Options (Consumer & Producer)
- Solace Message Headers Reference
- net/http.Response
- WriteFile
- streamRunner
- EngineList
- EngineInspectJSON
- Workload
- sinceFlag
- tailFlag

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 113 edges
2. `hasErr()` - 70 edges
3. `write()` - 52 edges
4. `Download()` - 52 edges
5. `Build()` - 49 edges
6. `captureStderr()` - 44 edges
7. `captureStdout()` - 38 edges
8. `Runner` - 37 edges
9. `wfOK()` - 36 edges
10. `wf()` - 36 edges

## Surprising Connections (you probably didn't know these)
- `renderedCompletions()` --calls--> `render()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/statusreport/render_test.go
- `TestCompletionCoversModel()` --calls--> `contains()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/runner/runner_test.go
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `readLog()` --references--> `Streamer`  [EXTRACTED]
  cmd/solmq-conn-util/logs.go → internal/runner/runner.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (106 total, 16 thin omitted)

### Community 0 - "consolidate_extra_test.go"
Cohesion: 0.13
Nodes (22): displayName(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning() (+14 more)

### Community 1 - "validate.go"
Cohesion: 0.13
Nodes (42): Workflow, checkConnections(), checkContainerTarget(), CheckDeployCommand(), checkDocker(), checkDuplicateSources(), checkImage(), checkImagePull() (+34 more)

### Community 2 - ".Run"
Cohesion: 0.10
Nodes (90): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+82 more)

### Community 3 - "testing.T"
Cohesion: 0.08
Nodes (52): run(), captureStdout(), containsToken(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError() (+44 more)

### Community 4 - "PodmanDeploy"
Cohesion: 0.19
Nodes (12): QuadletScope, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported(), TestPodmanDeploySystemModeNoUserFlag() (+4 more)

### Community 5 - "gen.go"
Cohesion: 0.05
Nodes (90): built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef, b64() (+82 more)

### Community 6 - "SolaceProps"
Cohesion: 0.29
Nodes (10): MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): cmd/solmq-conn-util, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "Kubernetes"
Cohesion: 0.13
Nodes (21): TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, ImagePullSecret, Kubernetes, Resources (+13 more)

### Community 11 - "Scan"
Cohesion: 0.22
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.05
Nodes (89): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync() (+81 more)

### Community 13 - "Runner"
Cohesion: 0.20
Nodes (27): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namingHint() (+19 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.24
Nodes (18): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+10 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.14
Nodes (14): 11. `examples`, 14. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery (+6 more)

### Community 18 - "Solace PubSub+ Connector for IBM MQ — Configuration Guide"
Cohesion: 0.13
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.17
Nodes (21): Input, Instance, composeEscape(), Mount, yw, Render(), renderContentConfig(), renderSecrets() (+13 more)

### Community 21 - "Render"
Cohesion: 0.09
Nodes (47): Input, Instance, KV, strings.Builder, PullSecret, StoreFile, yw, leaderMode() (+39 more)

### Community 23 - "YAML & Spring Boot Essentials (For Non-Spring Users)"
Cohesion: 0.25
Nodes (8): Converting YAML Keys to Environment Variables, Property Placeholders (`${}`), Property Precedence (Highest → Lowest), Relaxed Binding (kebab-case vs camelCase), Single Binder vs Multi-Binder Syntax, The `undefined` Binder, YAML Indentation, YAML & Spring Boot Essentials (For Non-Spring Users)

### Community 24 - "Verification Checklist"
Cohesion: 0.33
Nodes (6): 1. Mechanical diff against the bundled property catalog (recommended), 2. Live actuator introspection (verifies the JSON samples in [Section 12](#12-spring-actuator--management-endpoint)), 3. GitHub source spot-check, 4. Transform-expression smoke test (for [Section 8](#8-solace-connector--workflow-configuration)), Verification Checklist, What to update after verifying

### Community 25 - "Podman"
Cohesion: 0.22
Nodes (15): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets, applyDockerDefaults(), applyMountDefaults(), applyPodmanDefaults(), Docker (+7 more)

### Community 26 - "12. Spring Actuator / Management Endpoint"
Cohesion: 0.50
Nodes (4): 12. Spring Actuator / Management Endpoint, Sample Responses, Solace Binder Health Statuses, Solace Binder Metrics

### Community 27 - "2. IBM MQ Connection"
Cohesion: 0.50
Nodes (4): 2. IBM MQ Connection, Additional Properties (`additional-properties`), Core Connection Properties, JNDI Configuration (Alternative to Manual)

### Community 28 - "16. Spring Profiles & Config Locations"
Cohesion: 0.67
Nodes (3): 16. Spring Profiles & Config Locations, Config File Locations, Spring Profiles

### Community 29 - "main.go"
Cohesion: 0.18
Nodes (28): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+20 more)

### Community 30 - "6. JMS Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 6. JMS Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (33): Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Instance, Workflow (+25 more)

### Community 32 - "statusCollector"
Cohesion: 0.10
Nodes (21): sortInstances(), actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), markMissing(), watchStatus() (+13 more)

### Community 43 - "Side"
Cohesion: 0.14
Nodes (8): Cred, Side, placeholderSecretRef(), checkCred(), checkDefaultsCredentials(), checkSide(), checkSideCredentials(), edaAdvisory()

### Community 44 - "runAction"
Cohesion: 0.14
Nodes (34): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage() (+26 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.11
Nodes (50): net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+42 more)

### Community 48 - "runner_test.go"
Cohesion: 0.13
Nodes (24): call, canIVerb(), Preflight(), helperProcessArgv(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput() (+16 more)

### Community 49 - "Command details"
Cohesion: 0.12
Nodes (17): Command details, deploy, download, examples, generate, help, logs, remove (+9 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "spec_test.go"
Cohesion: 0.09
Nodes (31): ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe() (+23 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "names.go"
Cohesion: 0.22
Nodes (9): leaderNameFn, TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestBinderFieldsCarryStablePlaceholders(), TestGeneratedSecretNamesStayOutOfChildEnvDanger() (+1 more)

### Community 56 - "consolidate.go"
Cohesion: 0.14
Nodes (20): Opts, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), buildLeaderElection(), TestAppendPassthroughCollision(), TestFormatScalarQuoting() (+12 more)

### Community 57 - "Build"
Cohesion: 0.26
Nodes (20): Build(), Model, Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough() (+12 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "maven_test.go"
Cohesion: 0.25
Nodes (33): metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures(), pomXMLBody() (+25 more)

### Community 60 - "write"
Cohesion: 0.14
Nodes (31): podmanEnv(), podmanEnvSudo(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun(), TestDeployDockerSeamWritesComposeAndRuns(), TestDeployKubernetesPreflightFailureStopsBeforeApply() (+23 more)

### Community 61 - "Render"
Cohesion: 0.11
Nodes (36): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), itoa(), lineDiff(), loadSpecs(), mustRead() (+28 more)

### Community 62 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV() (+22 more)

### Community 63 - "ScriptInstalled"
Cohesion: 0.14
Nodes (14): execArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv(), TestInstallScriptUnknownPlatform(), TestRunStatusScriptArgv() (+6 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.15
Nodes (24): imageMismatchNote(), imageNameTag(), imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestEmbeddedOmitListFullyParses(), TestImageMismatchNote() (+16 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+17 more)

### Community 67 - "runner.go"
Cohesion: 0.17
Nodes (14): EngineImageInspectJSON(), LogsOpts, Streamer, Kubernetes(), KubernetesPodsJSON(), kubeVerb(), LogsArgv(), logsCommonFlags() (+6 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "TestShippedExamplesGenerateConfig"
Cohesion: 0.31
Nodes (8): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "render"
Cohesion: 0.28
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "main_test.go"
Cohesion: 0.08
Nodes (68): dispatch(), captureStderr(), downloadEnvWithImage(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAutoCompleteDispatchPrintsScript(), TestDownloadDirDefaultAndPositionalOverride() (+60 more)

### Community 73 - "10. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 10.1 `status container` -- the engine's view, 10.2 `status application` -- the connector's view, 10.3 `-d` / `--details`, 10.4 `--all`: find every instance by image, 10.5 `-w` / `--watch`, 10.6 `--output json`, 10.7 What the exit code means, and what each view costs, 10.8 The manual alternative (+3 more)

### Community 74 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 75 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 76 - "Defaults"
Cohesion: 0.26
Nodes (15): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, LeaderElection, Management, rawDefaults (+7 more)

### Community 77 - "12. `download jar`"
Cohesion: 0.22
Nodes (9): 12. `download jar`, Flags and defaults, Image-aware omission, Integrity verification (sha1), `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, The two sets, `--url` overrides all resolution (+1 more)

### Community 78 - "5. Workflow file"
Cohesion: 0.29
Nodes (7): 5.1 Top-level, 5.2 `solace:` options, 5.3 `mq:` options, 5.4 Destinations, durable names, passthrough, 5.5 Event-driven guidance (warnings), 5.6 Reusable connections (`conn-ref`), 5. Workflow file

### Community 79 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 80 - "spec.go"
Cohesion: 0.31
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 81 - "7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 7.0 image and timezone (shared by every platform), 7.1 kubernetes, 7.2 docker, 7.3 podman, 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)

### Community 82 - "8. Secrets model"
Cohesion: 0.40
Nodes (5): 8.1 Declaring a credential, 8.2 Stable names, 8.3 How each platform delivers them, 8.4 Registry credentials (pulling the image), 8. Secrets model

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (32): encoding/xml.Name, acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+24 more)

### Community 84 - "ParseEnv"
Cohesion: 0.12
Nodes (24): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+16 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Command abbreviations, Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations

### Community 86 - "ParseCommand"
Cohesion: 0.15
Nodes (13): ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure() (+5 more)

### Community 87 - "libs.go"
Cohesion: 0.12
Nodes (28): net/http.Client, net/url.URL, defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), filenameFromEscapedPath(), Input (+20 more)

### Community 88 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), Defaults, Env (+1 more)

### Community 89 - "Image"
Cohesion: 0.24
Nodes (5): workflowsFromRaw(), Image, rawEnv, rawWorkflows, Workflows

### Community 90 - "10. Solace Connector — Management & Leader Election"
Cohesion: 0.50
Nodes (4): 10. Solace Connector — Management & Leader Election, Active-Standby Configuration, Failover Configuration, Leader Election Mode

### Community 91 - "ApplyTop"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 92 - ".Run"
Cohesion: 0.21
Nodes (10): context.Context, io.Writer, os/exec.Cmd, applyCmdInput(), Cmd, KubernetesTop(), resolveArgv0(), TestKubernetesTopArgv() (+2 more)

### Community 93 - "1. Solace Event Broker Connection"
Cohesion: 0.67
Nodes (3): 1. Solace Event Broker Connection, Core Connection Properties, Solace API Properties (`api-properties`)

### Community 94 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 96 - "7. Solace Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 7. Solace Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 97 - "Solace Message Headers Reference"
Cohesion: 0.67
Nodes (3): Solace Binder Headers (`solace_scst_*`), Solace Message Headers Reference, Solace Message Headers (`solace_*`)

### Community 99 - "WriteFile"
Cohesion: 0.33
Nodes (6): os.FileMode, TestWriteFileCreatesDirsAndMode(), TestWriteFileDoesNotTightenExistingFileMode(), TestWriteFileParentIsFileReturnsError(), TestWriteFileTargetIsDirectoryReturnsError(), WriteFile()

### Community 100 - "streamRunner"
Cohesion: 0.40
Nodes (6): fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 101 - "EngineList"
Cohesion: 0.40
Nodes (5): EngineList(), EngineStats(), contains(), TestEngineListArgv(), TestEngineStatsArgv()

### Community 102 - "EngineInspectJSON"
Cohesion: 0.50
Nodes (4): EngineInspectJSON(), TestEngineInspectJSONArgv(), TestEngineInspectJSONFailureNamesTheTargets(), TestEngineInspectJSONNoNamesIsAnErrorAndRunsNothing()

### Community 103 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

## Knowledge Gaps
- **164 isolated node(s):** `logTarget`, `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `call` (+159 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `dispatch()` connect `main_test.go` to `testing.T`, `runAction`, `Runner`, `write`, `main.go`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes` to `Image`, `validate.go`, `.Run`, `spec_test.go`, `Render`, `Env`, `Podman`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `render()` connect `render` to `testing.T`, `completion.go`, `Render`, `statusreport/render.go`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 108 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 108 INFERRED edges - model-reasoned connections that need verification._
- **Are the 47 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 47 INFERRED edges - model-reasoned connections that need verification._
- **Are the 43 inferred relationships involving `Download()` (e.g. with `imageMismatchNote()` and `imageSatisfies()`) actually correct?**
  _`Download()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **What connects `logTarget`, `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn` to the rest of the system?**
  _164 weakly-connected nodes found - possible documentation gaps or missing edges._