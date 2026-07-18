# Graph Report - .  (2026-07-10)

## Corpus Check
- Corpus is ~27,805 words - fits in a single context window. You may not need a graph.

## Summary
- 404 nodes · 1172 edges · 21 communities (19 shown, 2 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 211 edges (avg confidence: 0.8)
- Token cost: 17,692 input · 1,241 output

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
- External IBM MQ Package

## God Nodes (most connected - your core abstractions)
1. `Run()` - 46 edges
2. `Build()` - 32 edges
3. `hasErr()` - 31 edges
4. `wf()` - 28 edges
5. `vMQ()` - 25 edges
6. `run()` - 24 edges
7. `vSolace()` - 23 edges
8. `Defaults` - 21 edges
9. `Side` - 21 edges
10. `Deploy()` - 16 edges

## Surprising Connections (you probably didn't know these)
- `runConfig()` --calls--> `Config()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go
- `runDeploy()` --calls--> `Deploy()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go
- `runValidate()` --calls--> `Validate()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/gen/gen.go
- `runExamples()` --calls--> `Write()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/examples/examples.go
- `loadRequest()` --calls--> `Scan()`  [INFERRED]
  cmd/solmq-gen/main.go → internal/scan/scan.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Core Generation Flow** — internal_scan, internal_spec, internal_consolidate, internal_validate, internal_render [EXTRACTED 0.90]
- **Example Specification Set** — internal_examples_files_workflow_0_yaml, internal_examples_files_workflow_1_yaml, internal_examples_files_defaults_yaml, internal_examples_files_kubernetes_yaml [EXTRACTED 1.00]

## Communities (21 total, 2 thin omitted)

### Community 0 - "Binding Consolidation"
Cohesion: 0.09
Nodes (54): acc, Binder, Binding, Bundle, JMSBinding, LeaderElectionModel, MQBinder, Prop (+46 more)

### Community 1 - "Spec Model & Validation"
Cohesion: 0.10
Nodes (38): Node, applyDest(), Node, nodePtr(), checkConnections(), checkDuplicateSources(), checkKeyAliasConflicts(), checkKube() (+30 more)

### Community 2 - "Validation Test Suite"
Cohesion: 0.23
Nodes (43): connDefaults(), T, TestCheckKubeCredentialSources(), TestCheckKubeRequiredAndReplicas(), TestCheckKubeStoresRequireTruststore(), TestCheckSideMQMissingFields(), TestCheckSideSolaceMissingAndBadScheme(), TestConnectionDefinitionValidation() (+35 more)

### Community 3 - "CLI Entry & Commands"
Cohesion: 0.15
Nodes (37): T, TestRunExamplesDefaultDir(), TestRunExamplesGeneratesAndConfigs(), TestRunExamplesSkipsThenForces(), collectFlagsAndDirs(), emit(), failFast(), fileReader() (+29 more)

### Community 4 - "Kubernetes Deploy Rendering"
Cohesion: 0.12
Nodes (31): Input, KV, StoreFile, yw, Builder, Model, managementPort(), quoteRes() (+23 more)

### Community 5 - "Generation Orchestration"
Cohesion: 0.19
Nodes (26): File, Request, Resolver, b64(), baseName(), Config(), Deploy(), T (+18 more)

### Community 6 - "application.yml Rendering"
Cohesion: 0.25
Nodes (19): Application(), Builder, Model, Node, renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+11 more)

### Community 7 - "Spec Parsing Tests"
Cohesion: 0.22
Nodes (20): ParseDefaults(), ParseKubernetes(), ParseWorkflow(), T, TestHasSystem(), TestParseDefaultsConnectionsAndLeaderElection(), TestParseDefaultsEmpty(), TestParseDefaultsError() (+12 more)

### Community 8 - "Dev Script (Bash)"
Cohesion: 0.47
Nodes (15): dev.sh script, die(), ok(), run(), step(), task_all(), task_build(), task_cov() (+7 more)

### Community 9 - "Dev Script (PowerShell)"
Cohesion: 0.43
Nodes (13): Die(), Ok(), Run(), Step(), Task-All(), Task-Build(), Task-Cov(), Task-Dist() (+5 more)

### Community 10 - "Golden File Tests"
Cohesion: 0.43
Nodes (13): configMapDoc(), deploymentDoc(), dirReader(), T, itoa(), lineDiff(), loadSpecs(), mustRead() (+5 more)

### Community 11 - "Input Directory Scanning"
Cohesion: 0.36
Nodes (12): isYAML(), Scan(), bases(), T, TestIsYAML(), TestScanClassifies(), TestScanCustomKubeNoDefaults(), TestScanEmptyKubeDefaultsToKubernetesYAML() (+4 more)

### Community 12 - "TLS Store Paths"
Cohesion: 0.28
Nodes (11): buildBundle(), base(), MountPath(), SolaceProps(), StorePath(), T, TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted() (+3 more)

### Community 13 - "Package Structure"
Cohesion: 0.20
Nodes (10): solmq-gen release, internal/consolidate, internal/deploy, internal/gen, internal/render, internal/scan, internal/spec, internal/tls (+2 more)

### Community 14 - "Durable UUID Naming"
Cohesion: 0.39
Nodes (6): DurableName(), mustParseUUID(), T, TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Examples Writer"
Cohesion: 0.47
Nodes (4): T, TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write()

## Knowledge Gaps
- **12 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-gen`, `Model`, `User Guide`, `internal/scan`, `internal/spec` (+7 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **2 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Build()` connect `Binding Consolidation` to `Spec Model & Validation`, `TLS Store Paths`, `Generation Orchestration`, `application.yml Rendering`?**
  _High betweenness centrality (0.220) - this node is a cross-community bridge._
- **Why does `Deploy()` connect `Generation Orchestration` to `Binding Consolidation`, `Validation Test Suite`, `CLI Entry & Commands`, `Kubernetes Deploy Rendering`, `application.yml Rendering`, `Golden File Tests`?**
  _High betweenness centrality (0.219) - this node is a cross-community bridge._
- **Why does `Defaults` connect `Spec Model & Validation` to `Binding Consolidation`, `Validation Test Suite`, `Kubernetes Deploy Rendering`, `Generation Orchestration`, `Spec Parsing Tests`, `TLS Store Paths`?**
  _High betweenness centrality (0.139) - this node is a cross-community bridge._
- **Are the 35 inferred relationships involving `Run()` (e.g. with `Config()` and `Deploy()`) actually correct?**
  _`Run()` has 35 INFERRED edges - model-reasoned connections that need verification._
- **Are the 18 inferred relationships involving `Build()` (e.g. with `SolaceProps()` and `TestBuildCipherConflictWarning()`) actually correct?**
  _`Build()` has 18 INFERRED edges - model-reasoned connections that need verification._
- **Are the 15 inferred relationships involving `hasErr()` (e.g. with `TestCheckKubeCredentialSources()` and `TestCheckKubeRequiredAndReplicas()`) actually correct?**
  _`hasErr()` has 15 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `wf()` (e.g. with `TestCheckSideMQMissingFields()` and `TestCheckSideSolaceMissingAndBadScheme()`) actually correct?**
  _`wf()` has 13 INFERRED edges - model-reasoned connections that need verification._