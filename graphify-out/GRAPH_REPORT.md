# Graph Report - solace-ibmmq-connector-helper  (2026-08-18)

## Corpus Check
- 66 files · ~144,215 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1104 nodes · 3117 edges · 67 communities (55 shown, 12 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 495 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `125a2417`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- .Run
- main.go
- Render
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
- Project Instructions
- Render
- ParseEnv
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
- testing.T
- solmq-conn-util command reference
- solmq-conn-util -- Development Guide
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- completion
- render.go
- buildLeaderElection
- consolidate.go
- Application
- consolidate_test.go
- runner.go
- PodmanDeploy
- InstallScript
- ParseCommand
- WriteFile
- PodmanSecretRemove
- ScriptInstalled
- Defaults

## God Nodes (most connected - your core abstractions)
1. `hasErr()` - 57 edges
2. `Build()` - 44 edges
3. `dispatch()` - 37 edges
4. `wf()` - 36 edges
5. `vMQ()` - 33 edges
6. `write()` - 29 edges
7. `vSolace()` - 29 edges
8. `wfOK()` - 27 edges
9. `Defaults` - 26 edges
10. `ParseEnv()` - 26 edges

## Surprising Connections (you probably didn't know these)
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckSyslog()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestStatusUserPasswordEnvCharset()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `dispatch()` --references--> `Runner`  [EXTRACTED]
  cmd/solmq-conn-util/main.go → internal/runner/runner.go

## Import Cycles
- None detected.

## Communities (67 total, 12 thin omitted)

### Community 0 - "Build"
Cohesion: 0.14
Nodes (21): appendPassthrough(), Build(), displayName(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessExposure(), TestApplyStatusAccessSecurityAbsent(), TestApplyStatusAccessSecurityDisabled() (+13 more)

### Community 1 - "validate.go"
Cohesion: 0.05
Nodes (68): defaultsFromRaw(), Defaults, Management, Security, TLSConfig, yaml.Node, applyDest(), Cred (+60 more)

### Community 2 - ".Run"
Cohesion: 0.12
Nodes (72): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), connDefaults(), dockerOK(), podmanOK() (+64 more)

### Community 3 - "main.go"
Cohesion: 0.10
Nodes (59): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), actStatus(), allowCommandFlag(), collectFlagsAndDirs() (+51 more)

### Community 4 - "Render"
Cohesion: 0.08
Nodes (58): Input, Instance, KV, Instance, StoreFile, yw, leaderMode(), LogbackXML() (+50 more)

### Community 5 - "gen.go"
Cohesion: 0.09
Nodes (57): built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef, b64() (+49 more)

### Community 6 - "SolaceProps"
Cohesion: 0.29
Nodes (10): MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseMountedBaseName() (+2 more)

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
Nodes (25): mustWrite(), testResolver(), TestShippedExamplesGenerateConfig(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write(), isYAML(), matchStar() (+17 more)

### Community 12 - "completion.go"
Cohesion: 0.09
Nodes (49): normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), flagsLine(), flagSpan(), invocation(), renderCommandsDoc(), tableCell() (+41 more)

### Community 13 - "RenderRunScript"
Cohesion: 0.19
Nodes (23): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs() (+15 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.27
Nodes (16): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar() (+8 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.06
Nodes (33): 10. Status: which instance is active, 11. `examples`, 11. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`) (+25 more)

### Community 18 - "Project Instructions"
Cohesion: 0.12
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.11
Nodes (25): Input, Instance, strings.Builder, Mount, yw, Render(), renderConfig(), renderSecrets() (+17 more)

### Community 21 - "ParseEnv"
Cohesion: 0.09
Nodes (35): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent() (+27 more)

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

### Community 43 - "testing.T"
Cohesion: 0.05
Nodes (112): dispatch(), run(), assertSameNameSet(), captureStderr(), captureStdout(), keySet(), manyWorkflowsDir(), nameSet() (+104 more)

### Community 44 - "solmq-conn-util command reference"
Cohesion: 0.40
Nodes (5): All commands, Command tree, Exit codes, Flags, solmq-conn-util command reference

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "runner_test.go"
Cohesion: 0.16
Nodes (20): Preflight(), helperProcessArgv(), TestHelperProcess(), TestOSRunAcceptsAbsolutePathArgv0(), TestOSRunCombinesStdoutAndStderr(), TestOSRunEchoesResolvedPathToStderr(), TestOSRunEnvReachesChildAndAmbientInherited(), TestOSRunNonZeroExitReturnsErrorWithOutput() (+12 more)

### Community 49 - "Command details"
Cohesion: 0.20
Nodes (10): Command details, delete, deploy, examples, generate, help, `solmq-conn-util generate config [--platform kubernetes|docker|podman] [-e env.yaml] [-o out]`, status (+2 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.29
Nodes (3): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_paths()

### Community 51 - "Model"
Cohesion: 0.28
Nodes (15): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+7 more)

### Community 52 - "completion"
Cohesion: 0.40
Nodes (5): completion, `solmq-conn-util completion bash`, `solmq-conn-util completion fish`, `solmq-conn-util completion powershell`, `solmq-conn-util completion zsh`

### Community 54 - "render.go"
Cohesion: 0.37
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 55 - "buildLeaderElection"
Cohesion: 0.19
Nodes (14): secretFn, buildLeaderElection(), TestBuildLeaderElection(), TestSanitizeAndIsTCPS(), testSecretRef(), trustStoreVal(), isTCPS(), sanitize() (+6 more)

### Community 56 - "consolidate.go"
Cohesion: 0.20
Nodes (12): Opts, applyStatusAccess(), TestFormatScalarQuoting(), TestHasExposureEntry(), TestNodeToProps(), FormatScalar(), Model, yaml.Node (+4 more)

### Community 57 - "Application"
Cohesion: 0.35
Nodes (13): Application(), buildRich(), lineDiff(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection(), TestApplicationMinimalNoOptionalBlocks(), TestApplicationOmitsEmptyCredentials() (+5 more)

### Community 58 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 59 - "runner.go"
Cohesion: 0.21
Nodes (10): canIVerb(), Cmd, Kubernetes(), KubernetesPodNames(), kubeVerb(), TestKubernetesPodNamesArgv(), TestKubernetesPodNamesEmptyResultIsNotError(), TestKubernetesPodNamesRunFailureWraps() (+2 more)

### Community 60 - "PodmanDeploy"
Cohesion: 0.22
Nodes (10): QuadletScope, PodmanDelete(), PodmanDeploy(), ResolveQuadletScope(), TestPodmanDeleteStopFailureIsReported(), TestPodmanDeleteStopsRemovesReloads(), TestPodmanDeployReloadThenStart(), TestPodmanDeployStartFailureIsReported() (+2 more)

### Community 61 - "InstallScript"
Cohesion: 0.22
Nodes (9): execArgv(), InstallScript(), RunStatusScript(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv(), TestInstallScriptUnknownPlatform(), TestRunStatusScriptArgv(), TestRunStatusScriptReturnsOutputAlongsideNonZeroExit() (+1 more)

### Community 62 - "ParseCommand"
Cohesion: 0.25
Nodes (8): ParseCommand(), PodmanSecretCreate(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure(), TestPodmanSecretCreateSkipsCreateWhenRmFails()

### Community 63 - "WriteFile"
Cohesion: 0.33
Nodes (6): os.FileMode, TestWriteFileCreatesDirsAndMode(), TestWriteFileDoesNotTightenExistingFileMode(), TestWriteFileParentIsFileReturnsError(), TestWriteFileTargetIsDirectoryReturnsError(), WriteFile()

### Community 64 - "PodmanSecretRemove"
Cohesion: 0.40
Nodes (5): PodmanSecretRemove(), TestPodmanSecretRemoveBatchesNames(), TestPodmanSecretRemoveNoNamesIsNoop(), TestPodmanSecretRemoveRejectsUnsafeCommand(), TestPodmanSecretRemoveReportsFailure()

### Community 65 - "ScriptInstalled"
Cohesion: 0.40
Nodes (5): ScriptInstalled(), TestScriptInstalledArgv(), TestScriptInstalledReadsMarkers(), TestScriptInstalledUnknownPlatform(), TestScriptInstalledUnreachableTargetIsError()

## Knowledge Gaps
- **123 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `Defaults`, `NO_COLOR`, `graphify` (+118 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Side` connect `validate.go` to `consolidate.go`, `consolidate_test.go`, `.Run`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Why does `Defaults` connect `validate.go` to `Build`, `.Run`, `Render`, `gen.go`, `SolaceProps`, `testing.T`, `Model`, `buildLeaderElection`, `consolidate.go`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Why does `Application()` connect `Application` to `Model`, `gen.go`, `render.go`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 38 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 38 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `Build()` (e.g. with `assignBinderNames()` and `stableName()`) actually correct?**
  _`Build()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 33 inferred relationships involving `dispatch()` (e.g. with `TestAllowCommandFlagBadValueExitsUsageError()` and `TestAllowCommandFlagRepeatableThreadsToRunner()`) actually correct?**
  _`dispatch()` has 33 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `wf()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckSideMQMissingFields()`) actually correct?**
  _`wf()` has 19 INFERRED edges - model-reasoned connections that need verification._