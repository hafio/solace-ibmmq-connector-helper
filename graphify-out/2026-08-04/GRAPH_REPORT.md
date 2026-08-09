# Graph Report - solace-ibmmq-connector-helper  (2026-08-04)

## Corpus Check
- 44 files · ~55,770 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 685 nodes · 1811 edges · 43 communities (32 shown, 11 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 278 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `021326af`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Binding Consolidation
- validate.go
- .Run
- runner_test.go
- Render
- gen.go
- application.yml Rendering
- Spec Parsing Tests
- dev.sh
- dev.ps1
- golden_test.go
- Scan
- TLS Store Paths
- podmangen.go
- Durable UUID Naming
- main.go
- solmq-conn -- User Guide
- Project Instructions
- Render
- Kubernetes
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

## God Nodes (most connected - your core abstractions)
1. `hasErr()` - 38 edges
2. `Build()` - 32 edges
3. `wf()` - 28 edges
4. `vMQ()` - 25 edges
5. `Solace PubSub+ Connector for IBM MQ — Configuration Guide` - 24 edges
6. `Defaults` - 23 edges
7. `vSolace()` - 23 edges
8. `Kubernetes` - 22 edges
9. `Side` - 22 edges
10. `Render()` - 20 edges

## Surprising Connections (you probably didn't know these)
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn/main.go
- `TestCheckSyslog()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn/main.go
- `genConfig()` --calls--> `Config()`  [INFERRED]
  cmd/solmq-conn/main.go → internal/gen/gen.go
- `genKubernetes()` --calls--> `GenerateKubernetes()`  [INFERRED]
  cmd/solmq-conn/main.go → internal/gen/gen.go
- `genDocker()` --calls--> `GenerateDocker()`  [INFERRED]
  cmd/solmq-conn/main.go → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (43 total, 11 thin omitted)

### Community 0 - "Binding Consolidation"
Cohesion: 0.09
Nodes (55): acc, Binder, Binding, Bundle, JMSBinding, LeaderElectionModel, MQBinder, Prop (+47 more)

### Community 1 - "validate.go"
Cohesion: 0.08
Nodes (54): defaultsFromRaw(), Node, applyDest(), Node, nodePtr(), checkCommand(), checkConnections(), checkContainerTarget() (+46 more)

### Community 2 - ".Run"
Cohesion: 0.18
Nodes (53): baseKubeDeploy(), connDefaults(), dockerOK(), T, TestCheckCommandMultiToken(), TestCheckDocker(), TestCheckDockerCredentials(), TestCheckKubeCredentialSources() (+45 more)

### Community 3 - "runner_test.go"
Cohesion: 0.14
Nodes (26): Docker(), Kubernetes(), kubeVerb(), ParseCommand(), PodmanDelete(), PodmanDeploy(), ResolveQuadletScope(), T (+18 more)

### Community 4 - "Render"
Cohesion: 0.14
Nodes (36): Input, Instance, KV, StoreFile, yw, Instance, baseName(), Builder (+28 more)

### Community 5 - "gen.go"
Cohesion: 0.10
Nodes (58): DockerPlan, File, mount, NamedDoc, PodmanOpts, PodmanPlan, Request, Resolver (+50 more)

### Community 6 - "application.yml Rendering"
Cohesion: 0.24
Nodes (20): Application(), Builder, Model, Node, renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+12 more)

### Community 7 - "Spec Parsing Tests"
Cohesion: 0.21
Nodes (21): ParseDefaults(), ParseKubernetes(), ParseWorkflow(), T, TestHasSystem(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError() (+13 more)

### Community 8 - "dev.sh"
Cohesion: 0.16
Nodes (18): c(), finish(), log_begin(), NO_COLOR, run(), dev.sh script, die(), ok() (+10 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (12): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-scan(), Task-test() (+4 more)

### Community 10 - "golden_test.go"
Cohesion: 0.41
Nodes (14): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), T, itoa(), lineDiff(), loadSpecs() (+6 more)

### Community 11 - "Scan"
Cohesion: 0.26
Nodes (20): isYAML(), matchStar(), sameFile(), Scan(), bases(), T, TestIsYAML(), TestMatchStar() (+12 more)

### Community 12 - "TLS Store Paths"
Cohesion: 0.32
Nodes (10): base(), MountPath(), SolaceProps(), StorePath(), T, TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 13 - "podmangen.go"
Cohesion: 0.18
Nodes (20): Input, Builder, portPair(), RenderQuadlet(), RenderRunScript(), runArgs(), fullInput(), T (+12 more)

### Community 14 - "Durable UUID Naming"
Cohesion: 0.39
Nodes (6): DurableName(), mustParseUUID(), T, TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "main.go"
Cohesion: 0.14
Nodes (42): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), collectFlagsAndDirs(), emit(), emitConfigs() (+34 more)

### Community 17 - "solmq-conn -- User Guide"
Cohesion: 0.06
Nodes (31): Build, Design notes, Release (CI), solmq-conn -- Development Guide, Testing, Commands, Documentation, Minimal working example (+23 more)

### Community 18 - "Project Instructions"
Cohesion: 0.12
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.20
Nodes (18): Input, Instance, Mount, yw, Builder, Render(), renderConfig(), renderService() (+10 more)

### Community 21 - "Kubernetes"
Cohesion: 0.13
Nodes (31): syslogOf(), ParseEnv(), workflowsFromRaw(), applyKubeDefaults(), applyDockerDefaults(), applyMountDefaults(), applyPodmanDefaults(), CredCreate (+23 more)

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

## Knowledge Gaps
- **83 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `Model`, `NO_COLOR`, `graphify`, `Quick start` (+78 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Defaults` connect `validate.go` to `Binding Consolidation`, `.Run`, `Render`, `gen.go`, `Spec Parsing Tests`, `TLS Store Paths`, `Kubernetes`?**
  _High betweenness centrality (0.146) - this node is a cross-community bridge._
- **Why does `Build()` connect `Binding Consolidation` to `validate.go`, `TLS Store Paths`, `gen.go`, `application.yml Rendering`?**
  _High betweenness centrality (0.104) - this node is a cross-community bridge._
- **Why does `buildShards()` connect `gen.go` to `Binding Consolidation`, `validate.go`, `application.yml Rendering`?**
  _High betweenness centrality (0.082) - this node is a cross-community bridge._
- **Are the 21 inferred relationships involving `hasErr()` (e.g. with `TestCheckCommandMultiToken()` and `TestCheckDocker()`) actually correct?**
  _`hasErr()` has 21 INFERRED edges - model-reasoned connections that need verification._
- **Are the 18 inferred relationships involving `Build()` (e.g. with `SolaceProps()` and `TestBuildCipherConflictWarning()`) actually correct?**
  _`Build()` has 18 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `wf()` (e.g. with `TestCheckSideMQMissingFields()` and `TestCheckSideSolaceMissingAndBadScheme()`) actually correct?**
  _`wf()` has 13 INFERRED edges - model-reasoned connections that need verification._
- **Are the 11 inferred relationships involving `vMQ()` (e.g. with `TestCheckSideSolaceMissingAndBadScheme()` and `TestConnRefStrictOnlyDestination()`) actually correct?**
  _`vMQ()` has 11 INFERRED edges - model-reasoned connections that need verification._