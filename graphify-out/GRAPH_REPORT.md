# Graph Report - solace-ibmmq-connector-helper  (2026-08-20)

## Corpus Check
- 71 files · ~167,925 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1197 nodes · 3457 edges · 77 communities (64 shown, 13 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 597 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a4d15b72`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- consolidate_extra_test.go
- validate.go
- .Run
- main.go
- runner.go
- gen.go
- SolaceProps
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- golden_test.go
- Scan
- completion.go
- RenderRunScript
- DurableName
- Expand
- solmq-conn-util -- User Guide
- Solace PubSub+ Connector for IBM MQ — Configuration Guide
- Render
- Render
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- YAML & Spring Boot Essentials (For Non-Spring Users)
- Verification Checklist
- 10. Solace Connector — Management & Leader Election
- 12. Spring Actuator / Management Endpoint
- 2. IBM MQ Connection
- 16. Spring Profiles & Config Locations
- 1. Solace Event Broker Connection
- 6. JMS Binding-Level Options (Consumer & Producer)
- 7. Solace Binding-Level Options (Consumer & Producer)
- Solace Message Headers Reference
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
- spec_test.go
- solmq-conn-util command reference
- solmq-conn-util -- Development Guide
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- testing.T
- render.go
- buildLeaderElection
- consolidate.go
- Build
- consolidate_test.go
- Runner
- run
- Render
- main_test.go
- runStatusOnTargets
- repeatableName
- Defaults
- Defaults
- Cred
- TestDispatchHandlersMatchModel
- SafeToken
- Kubernetes
- spec.go
- Side
- 5. Workflow file
- 10. Status: which instance is active
- 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)
- 8. Secrets model

## God Nodes (most connected - your core abstractions)
1. `hasErr()` - 66 edges
2. `Build()` - 45 edges
3. `dispatch()` - 44 edges
4. `wfOK()` - 36 edges
5. `wf()` - 36 edges
6. `vMQ()` - 33 edges
7. `write()` - 32 edges
8. `ParseEnv()` - 32 edges
9. `vSolace()` - 29 edges
10. `imageOK()` - 29 edges

## Surprising Connections (you probably didn't know these)
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckKubeServicePort()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (77 total, 13 thin omitted)

### Community 0 - "consolidate_extra_test.go"
Cohesion: 0.17
Nodes (14): containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildLeaderElection(), TestBuildMessageLoopWarning(), TestBuildMQmTLSBundle(), TestBuildSolaceTopicSourceEmitsConsumerTopic() (+6 more)

### Community 1 - "validate.go"
Cohesion: 0.19
Nodes (26): Workflow, checkContainerTarget(), checkDocker(), checkDuplicateSources(), checkImagePull(), checkKeyAliasConflicts(), checkKube(), checkLibs() (+18 more)

### Community 2 - ".Run"
Cohesion: 0.11
Nodes (84): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), connDefaults(), dockerOK() (+76 more)

### Community 3 - "main.go"
Cohesion: 0.13
Nodes (32): actStatus(), allowCommandFlag(), collectFlagsAndDirs(), contains(), envFlag(), loadEnvFile(), loadStatusEnv(), main() (+24 more)

### Community 4 - "runner.go"
Cohesion: 0.14
Nodes (18): canIVerb(), DockerComposeProject(), Cmd, QuadletScope, Kubernetes(), kubeVerb(), PodmanDeploy(), PodmanRemove() (+10 more)

### Community 5 - "gen.go"
Cohesion: 0.08
Nodes (66): built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef, b64() (+58 more)

### Community 6 - "SolaceProps"
Cohesion: 0.23
Nodes (12): TestBaseName(), BaseName(), MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing() (+4 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.12
Nodes (17): cmd/solmq-conn-util, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/podmangen (+9 more)

### Community 8 - "dev.sh"
Cohesion: 0.16
Nodes (18): c(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die(), ok() (+10 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 10 - "golden_test.go"
Cohesion: 0.31
Nodes (16): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), itoa(), lineDiff(), loadSpecs(), mustRead() (+8 more)

### Community 11 - "Scan"
Cohesion: 0.15
Nodes (27): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write(), isYAML(), matchStar() (+19 more)

### Community 12 - "completion.go"
Cohesion: 0.08
Nodes (57): normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), flagsLine(), flagSpan(), invocation(), renderCommandsDoc(), tableCell() (+49 more)

### Community 13 - "RenderRunScript"
Cohesion: 0.19
Nodes (23): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs() (+15 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.26
Nodes (17): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar() (+9 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.14
Nodes (14): 11. `examples`, 12. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery (+6 more)

### Community 18 - "Solace PubSub+ Connector for IBM MQ — Configuration Guide"
Cohesion: 0.13
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.12
Nodes (24): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+16 more)

### Community 21 - "Render"
Cohesion: 0.11
Nodes (45): Input, Instance, KV, Instance, PullSecret, StoreFile, yw, leaderMode() (+37 more)

### Community 23 - "YAML & Spring Boot Essentials (For Non-Spring Users)"
Cohesion: 0.25
Nodes (8): Converting YAML Keys to Environment Variables, Property Placeholders (`${}`), Property Precedence (Highest → Lowest), Relaxed Binding (kebab-case vs camelCase), Single Binder vs Multi-Binder Syntax, The `undefined` Binder, YAML Indentation, YAML & Spring Boot Essentials (For Non-Spring Users)

### Community 24 - "Verification Checklist"
Cohesion: 0.33
Nodes (6): 1. Mechanical diff against the bundled property catalog (recommended), 2. Live actuator introspection (verifies the JSON samples in [Section 12](#12-spring-actuator--management-endpoint)), 3. GitHub source spot-check, 4. Transform-expression smoke test (for [Section 8](#8-solace-connector--workflow-configuration)), Verification Checklist, What to update after verifying

### Community 25 - "10. Solace Connector — Management & Leader Election"
Cohesion: 0.50
Nodes (4): 10. Solace Connector — Management & Leader Election, Active-Standby Configuration, Failover Configuration, Leader Election Mode

### Community 26 - "12. Spring Actuator / Management Endpoint"
Cohesion: 0.50
Nodes (4): 12. Spring Actuator / Management Endpoint, Sample Responses, Solace Binder Health Statuses, Solace Binder Metrics

### Community 27 - "2. IBM MQ Connection"
Cohesion: 0.50
Nodes (4): 2. IBM MQ Connection, Additional Properties (`additional-properties`), Core Connection Properties, JNDI Configuration (Alternative to Manual)

### Community 28 - "16. Spring Profiles & Config Locations"
Cohesion: 0.67
Nodes (3): 16. Spring Profiles & Config Locations, Config File Locations, Spring Profiles

### Community 29 - "1. Solace Event Broker Connection"
Cohesion: 0.67
Nodes (3): 1. Solace Event Broker Connection, Core Connection Properties, Solace API Properties (`api-properties`)

### Community 30 - "6. JMS Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 6. JMS Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 31 - "7. Solace Binding-Level Options (Consumer & Producer)"
Cohesion: 0.67
Nodes (3): 7. Solace Binding-Level Options (Consumer & Producer), Consumer Options, Producer Options

### Community 32 - "Solace Message Headers Reference"
Cohesion: 0.67
Nodes (3): Solace Binder Headers (`solace_scst_*`), Solace Message Headers Reference, Solace Message Headers (`solace_*`)

### Community 43 - "spec_test.go"
Cohesion: 0.11
Nodes (26): ParseDefaults(), ParseKubernetes(), ParseWorkflow(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty() (+18 more)

### Community 44 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "runner_test.go"
Cohesion: 0.07
Nodes (44): os.FileMode, KubernetesPodNames(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), helperProcessArgv(), TestHelperProcess() (+36 more)

### Community 49 - "Command details"
Cohesion: 0.13
Nodes (15): auto-complete, Command details, deploy, examples, generate, help, remove, `solmq-conn-util auto-complete bash` (+7 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.29
Nodes (3): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_paths()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "testing.T"
Cohesion: 0.13
Nodes (29): testing.T, TestLeaderElectionEffectiveMode(), ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride() (+21 more)

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "buildLeaderElection"
Cohesion: 0.15
Nodes (15): secretFn, buildLeaderElection(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames() (+7 more)

### Community 56 - "consolidate.go"
Cohesion: 0.16
Nodes (14): Opts, appendPassthrough(), applyStatusAccess(), displayName(), TestAppendPassthroughCollision(), TestDisplayName(), TestFormatScalarQuoting(), TestNodeToProps() (+6 more)

### Community 57 - "Build"
Cohesion: 0.33
Nodes (16): Build(), Model, Application(), buildRich(), lineDiff(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+8 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "Runner"
Cohesion: 0.22
Nodes (25): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+17 more)

### Community 60 - "run"
Cohesion: 0.14
Nodes (20): run(), captureStdout(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateDockerToFile() (+12 more)

### Community 61 - "Render"
Cohesion: 0.19
Nodes (17): breEscape(), Render(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderIsPureASCIINoCRLF() (+9 more)

### Community 62 - "main_test.go"
Cohesion: 0.11
Nodes (53): dispatch(), captureStderr(), manyWorkflowsDir(), podmanEnv(), podmanEnvSudo(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner() (+45 more)

### Community 63 - "runStatusOnTargets"
Cohesion: 0.12
Nodes (18): confirmInstall(), printTargetReport(), runStatusOnTargets(), statusBanner(), execArgv(), InstallScript(), RunStatusScript(), ScriptInstalled() (+10 more)

### Community 65 - "Defaults"
Cohesion: 0.17
Nodes (17): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, checkLeaderElection(), checkRemovedDefaultsKeys(), LeaderElection (+9 more)

### Community 67 - "Cred"
Cohesion: 0.16
Nodes (3): Image, Cred, placeholderSecretRef()

### Community 68 - "TestDispatchHandlersMatchModel"
Cohesion: 0.47
Nodes (6): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "SafeToken"
Cohesion: 0.14
Nodes (12): CheckDeployCommand(), checkImage(), checkSecurityUserRoles(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser(), TestSafeToken() (+4 more)

### Community 70 - "Kubernetes"
Cohesion: 0.08
Nodes (39): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), workflowsFromRaw() (+31 more)

### Community 71 - "spec.go"
Cohesion: 0.31
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 72 - "Side"
Cohesion: 0.33
Nodes (9): Side, checkConnections(), checkCred(), checkDefaultsCredentials(), checkSide(), checkSideCredentials(), checkTuple(), checkWorkflowSides() (+1 more)

### Community 73 - "5. Workflow file"
Cohesion: 0.29
Nodes (7): 5.1 Top-level, 5.2 `solace:` options, 5.3 `mq:` options, 5.4 Destinations, durable names, passthrough, 5.5 Event-driven guidance (warnings), 5.6 Reusable connections (`conn-ref`), 5. Workflow file

### Community 74 - "10. Status: which instance is active"
Cohesion: 0.40
Nodes (5): 10. Status: which instance is active, First run: installing the script, Instances this tool did not deploy, Sample output, The manual alternative

### Community 75 - "7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 7.0 image and timezone (shared by every platform), 7.1 kubernetes, 7.2 docker, 7.3 podman, 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)

### Community 76 - "8. Secrets model"
Cohesion: 0.40
Nodes (5): 8.1 Declaring a credential, 8.2 Stable names, 8.3 How each platform delivers them, 8.4 Registry credentials (pulling the image), 8. Secrets model

## Knowledge Gaps
- **127 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `Defaults`, `NO_COLOR`, `graphify` (+122 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Defaults` connect `Defaults` to `validate.go`, `.Run`, `gen.go`, `SolaceProps`, `SafeToken`, `Side`, `spec_test.go`, `Model`, `Render`, `buildLeaderElection`, `consolidate.go`, `Build`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Why does `Build()` connect `Build` to `consolidate_extra_test.go`, `Defaults`, `validate.go`, `gen.go`, `SolaceProps`, `Model`, `buildLeaderElection`, `consolidate.go`, `consolidate_test.go`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `Model` connect `Model` to `Defaults`, `gen.go`, `Render`, `render.go`, `Build`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Are the 47 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 47 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `Build()` (e.g. with `assignBinderNames()` and `stableName()`) actually correct?**
  _`Build()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 39 inferred relationships involving `dispatch()` (e.g. with `TestAllowCommandFlagBadValueExitsUsageError()` and `TestAllowCommandFlagRepeatableThreadsToRunner()`) actually correct?**
  _`dispatch()` has 39 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `wfOK()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`wfOK()` has 13 INFERRED edges - model-reasoned connections that need verification._