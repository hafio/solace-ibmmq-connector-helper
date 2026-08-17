# Graph Report - solace-ibmmq-connector-helper  (2026-08-17)

## Corpus Check
- 63 files · ~119,258 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 982 nodes · 2705 edges · 54 communities (43 shown, 11 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 412 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `979a019d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- .Run
- runner_test.go
- Render
- gen.go
- SolaceProps
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- golden_test.go
- Scan
- completion.go
- podmangen.go
- DurableName
- Expand
- solmq-conn-util -- User Guide
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
- testing.T
- solmq-conn-util command reference
- solmq-conn-util -- Development Guide
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- generate
- Command details
- solmq-conn-util.bash
- deploy
- completion

## God Nodes (most connected - your core abstractions)
1. `hasErr()` - 55 edges
2. `Build()` - 40 edges
3. `wf()` - 36 edges
4. `vMQ()` - 33 edges
5. `vSolace()` - 29 edges
6. `Side` - 25 edges
7. `wfOK()` - 25 edges
8. `Defaults` - 24 edges
9. `ParseEnv()` - 24 edges
10. `Solace PubSub+ Connector for IBM MQ — Configuration Guide` - 24 edges

## Surprising Connections (you probably didn't know these)
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `TestCheckSyslog()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go
- `dispatch()` --references--> `Runner`  [EXTRACTED]
  cmd/solmq-conn-util/main.go → internal/runner/runner.go
- `genConfig()` --calls--> `Config()`  [EXTRACTED]
  cmd/solmq-conn-util/main.go → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (54 total, 11 thin omitted)

### Community 0 - "Build"
Cohesion: 0.05
Nodes (95): acc, Binder, Binding, JMSBinding, MQBinder, Opts, secretFn, Session (+87 more)

### Community 1 - "validate.go"
Cohesion: 0.06
Nodes (63): defaultsFromRaw(), Defaults, Management, Security, TLSConfig, yaml.Node, applyDest(), Cred (+55 more)

### Community 2 - ".Run"
Cohesion: 0.13
Nodes (70): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), connDefaults(), dockerOK(), podmanOK() (+62 more)

### Community 3 - "runner_test.go"
Cohesion: 0.05
Nodes (96): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), allowCommandFlag(), collectFlagsAndDirs(), emit() (+88 more)

### Community 4 - "Render"
Cohesion: 0.15
Nodes (34): Input, Instance, KV, Instance, StoreFile, yw, LogbackXML(), managementPort() (+26 more)

### Community 5 - "gen.go"
Cohesion: 0.10
Nodes (48): built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef, b64() (+40 more)

### Community 6 - "SolaceProps"
Cohesion: 0.21
Nodes (13): TestBaseName(), BaseName(), MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted() (+5 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.12
Nodes (16): cmd/solmq-conn-util, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/podmangen (+8 more)

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
Nodes (48): normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), flagsLine(), flagSpan(), invocation(), renderCommandsDoc(), usageText() (+40 more)

### Community 13 - "podmangen.go"
Cohesion: 0.20
Nodes (19): Mount, SecretRef, Unit, RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs(), fullInput() (+11 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.27
Nodes (16): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar() (+8 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.07
Nodes (27): 10. `examples`, 11. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery (+19 more)

### Community 18 - "Project Instructions"
Cohesion: 0.12
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.12
Nodes (21): Input, Instance, strings.Builder, Mount, yw, Render(), renderConfig(), renderSecrets() (+13 more)

### Community 21 - "Kubernetes"
Cohesion: 0.09
Nodes (37): TestDockerRejectsUnsafeCommand(), TestKubernetesDeleteUsesDeleteVerb(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesUnknownAction(), Env, workflowsFromRaw(), Deployment (+29 more)

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
Cohesion: 0.06
Nodes (90): dispatch(), run(), assertSameNameSet(), captureStderr(), captureStdout(), keySet(), manyWorkflowsDir(), nameSet() (+82 more)

### Community 44 - "solmq-conn-util command reference"
Cohesion: 0.40
Nodes (5): All commands, Command tree, Exit codes, Flags, solmq-conn-util command reference

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "generate"
Cohesion: 0.40
Nodes (5): generate, `solmq-conn-util generate config [-e env.yaml] [-o out]`, `solmq-conn-util generate docker [-e env.yaml] [-o out]`, `solmq-conn-util generate kubernetes [-e env.yaml] [-o out]`, `solmq-conn-util generate podman [-e env.yaml] [-o out]`

### Community 49 - "Command details"
Cohesion: 0.25
Nodes (8): Command details, delete, examples, help, `solmq-conn-util delete docker [-e env.yaml] [--allow-command name]`, `solmq-conn-util delete kubernetes [-e env.yaml] [--allow-command name]`, `solmq-conn-util delete podman [-e env.yaml] [--allow-command name]`, validate

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.29
Nodes (3): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_paths()

### Community 51 - "deploy"
Cohesion: 0.50
Nodes (4): deploy, `solmq-conn-util deploy docker [-e env.yaml] [--allow-command name]`, `solmq-conn-util deploy kubernetes [-e env.yaml] [--allow-command name]`, `solmq-conn-util deploy podman [-e env.yaml] [--allow-command name]`

### Community 52 - "completion"
Cohesion: 0.40
Nodes (5): completion, `solmq-conn-util completion bash`, `solmq-conn-util completion fish`, `solmq-conn-util completion powershell`, `solmq-conn-util completion zsh`

## Knowledge Gaps
- **123 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `Model`, `NO_COLOR`, `graphify` (+118 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Defaults` connect `validate.go` to `Build`, `.Run`, `Render`, `gen.go`, `SolaceProps`, `testing.T`, `Kubernetes`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Build()` connect `Build` to `validate.go`, `gen.go`, `SolaceProps`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `Kubernetes` to `validate.go`, `testing.T`, `Render`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 36 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 36 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `Build()` (e.g. with `assignBinderNames()` and `securityUserPasswordName()`) actually correct?**
  _`Build()` has 17 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `wf()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckSideMQMissingFields()`) actually correct?**
  _`wf()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `vMQ()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckSideSolaceMissingAndBadScheme()`) actually correct?**
  _`vMQ()` has 17 INFERRED edges - model-reasoned connections that need verification._