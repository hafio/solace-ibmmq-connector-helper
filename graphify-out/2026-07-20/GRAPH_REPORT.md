# Graph Report - solace-ibmmq-connector-helper  (2026-07-20)

## Corpus Check
- 38 files · ~42,744 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 543 nodes · 1435 edges · 42 communities (30 shown, 12 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 226 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `a9ae98eb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Binding Consolidation
- Spec Model & Validation
- Validation Test Suite
- CLI Entry & Commands
- Kubernetes Deploy Rendering
- Generation Orchestration
- application.yml Rendering
- Spec Parsing Tests
- Dev Script (Bash)
- Dev Script (PowerShell)
- Golden File Tests
- Input Directory Scanning
- TLS Store Paths
- Package Structure
- Durable UUID Naming
- Examples Writer
- Docs & Golden Output
- Project Instructions
- External IBM MQ Package
- consolidate_test.go
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
1. `Run()` - 49 edges
2. `hasErr()` - 34 edges
3. `Build()` - 32 edges
4. `run()` - 29 edges
5. `wf()` - 28 edges
6. `vMQ()` - 25 edges
7. `Solace PubSub+ Connector for IBM MQ — Configuration Guide` - 24 edges
8. `vSolace()` - 23 edges
9. `Defaults` - 22 edges
10. `Render()` - 21 edges

## Surprising Connections (you probably didn't know these)
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-gen/main.go
- `TestCheckSyslog()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-gen/main.go
- `runConfig()` --calls--> `Config()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go
- `runDeploy()` --calls--> `Deploy()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go
- `runValidate()` --calls--> `Validate()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (42 total, 12 thin omitted)

### Community 0 - "Binding Consolidation"
Cohesion: 0.09
Nodes (55): acc, Binder, Binding, Bundle, JMSBinding, LeaderElectionModel, MQBinder, Prop (+47 more)

### Community 1 - "Spec Model & Validation"
Cohesion: 0.09
Nodes (42): Node, applyDest(), Node, nodePtr(), checkConnections(), checkDuplicateSources(), checkKeyAliasConflicts(), checkKube() (+34 more)

### Community 2 - "Validation Test Suite"
Cohesion: 0.20
Nodes (48): baseKubeDeploy(), connDefaults(), T, TestCheckKubeCredentialSources(), TestCheckKubeRequiredAndReplicas(), TestCheckKubeStoresRequireTruststore(), TestCheckLibs(), TestCheckSideMQMissingFields() (+40 more)

### Community 3 - "CLI Entry & Commands"
Cohesion: 0.12
Nodes (45): T, TestRunExamplesDefaultDir(), TestRunExamplesGeneratesAndConfigs(), TestRunExamplesSkipsThenForces(), collectFlagsAndDirs(), emit(), emitConfigs(), failFast() (+37 more)

### Community 4 - "Kubernetes Deploy Rendering"
Cohesion: 0.15
Nodes (35): Input, Instance, KV, StoreFile, yw, baseName(), Builder, Model (+27 more)

### Community 5 - "Generation Orchestration"
Cohesion: 0.14
Nodes (36): File, Request, Resolver, shard, b64(), baseName(), buildShards(), Config() (+28 more)

### Community 6 - "application.yml Rendering"
Cohesion: 0.24
Nodes (20): Application(), Builder, Model, Node, renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+12 more)

### Community 7 - "Spec Parsing Tests"
Cohesion: 0.21
Nodes (21): ParseDefaults(), ParseKubernetes(), ParseWorkflow(), T, TestHasSystem(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError() (+13 more)

### Community 8 - "Dev Script (Bash)"
Cohesion: 0.46
Nodes (16): dev.sh script, die(), ok(), run(), step(), task_all(), task_build(), task_cov() (+8 more)

### Community 9 - "Dev Script (PowerShell)"
Cohesion: 0.42
Nodes (14): Die(), Ok(), Run(), Step(), Task-All(), Task-Build(), Task-Cov(), Task-Dist() (+6 more)

### Community 10 - "Golden File Tests"
Cohesion: 0.43
Nodes (13): configMapDoc(), deploymentDoc(), dirReader(), T, itoa(), lineDiff(), loadSpecs(), mustRead() (+5 more)

### Community 11 - "Input Directory Scanning"
Cohesion: 0.36
Nodes (12): isYAML(), Scan(), bases(), T, TestIsYAML(), TestScanClassifies(), TestScanCustomKubeNoDefaults(), TestScanEmptyKubeDefaultsToKubernetesYAML() (+4 more)

### Community 12 - "TLS Store Paths"
Cohesion: 0.32
Nodes (10): base(), MountPath(), SolaceProps(), StorePath(), T, TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 14 - "Durable UUID Naming"
Cohesion: 0.39
Nodes (6): DurableName(), mustParseUUID(), T, TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Examples Writer"
Cohesion: 0.47
Nodes (4): T, TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write()

### Community 17 - "Docs & Golden Output"
Cohesion: 0.07
Nodes (28): Build, Design notes, Release (CI), solmq-gen — Development Guide, Testing, Commands, Documentation, Minimal working example (+20 more)

### Community 18 - "Project Instructions"
Cohesion: 0.12
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 21 - "consolidate_test.go"
Cohesion: 0.20
Nodes (18): syslogOf(), storesWired(), CredCreate, CredentialsSecret, Deployment, Kubernetes, Libs, LibsDownload (+10 more)

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
- **81 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen`, `Model`, `graphify`, `Quick start`, `Minimal working example` (+76 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **12 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Defaults` connect `Spec Model & Validation` to `Binding Consolidation`, `Validation Test Suite`, `Kubernetes Deploy Rendering`, `Generation Orchestration`, `Spec Parsing Tests`, `TLS Store Paths`?**
  _High betweenness centrality (0.128) - this node is a cross-community bridge._
- **Why does `Build()` connect `Binding Consolidation` to `Spec Model & Validation`, `TLS Store Paths`, `Generation Orchestration`, `application.yml Rendering`?**
  _High betweenness centrality (0.117) - this node is a cross-community bridge._
- **Why does `Deploy()` connect `Generation Orchestration` to `Golden File Tests`, `Validation Test Suite`, `CLI Entry & Commands`, `Kubernetes Deploy Rendering`?**
  _High betweenness centrality (0.110) - this node is a cross-community bridge._
- **Are the 38 inferred relationships involving `Run()` (e.g. with `Config()` and `Deploy()`) actually correct?**
  _`Run()` has 38 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `hasErr()` (e.g. with `TestCheckKubeCredentialSources()` and `TestCheckKubeRequiredAndReplicas()`) actually correct?**
  _`hasErr()` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 18 inferred relationships involving `Build()` (e.g. with `SolaceProps()` and `TestBuildCipherConflictWarning()`) actually correct?**
  _`Build()` has 18 INFERRED edges - model-reasoned connections that need verification._
- **Are the 22 inferred relationships involving `run()` (e.g. with `TestRunExamplesDefaultDir()` and `TestRunExamplesGeneratesAndConfigs()`) actually correct?**
  _`run()` has 22 INFERRED edges - model-reasoned connections that need verification._