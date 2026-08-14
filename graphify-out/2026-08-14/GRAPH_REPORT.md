# Graph Report - solace-ibmmq-connector-helper  (2026-08-13)

## Corpus Check
- 58 files · ~102,489 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 889 nodes · 2430 edges · 52 communities (41 shown, 11 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 354 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `94e438ca`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- testing.T
- runner_test.go
- Render
- gen.go
- spec_test.go
- solmq-conn test catalogue
- dev.sh
- dev.ps1
- golden_test.go
- Scan
- SolaceProps
- podmangen.go
- DurableName
- Expand
- solmq-conn -- User Guide
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
- main_test.go
- solmq-conn command reference
- solmq-conn -- Development Guide
- solmq-conn -- Solace IBM MQ Connector config generator and deployer
- generate
- Command details
- commands.go
- deploy

## God Nodes (most connected - your core abstractions)
1. `hasErr()` - 51 edges
2. `Build()` - 39 edges
3. `wf()` - 36 edges
4. `vMQ()` - 33 edges
5. `vSolace()` - 29 edges
6. `Side` - 25 edges
7. `Defaults` - 24 edges
8. `ParseEnv()` - 24 edges
9. `Solace PubSub+ Connector for IBM MQ — Configuration Guide` - 24 edges
10. `GeneratePodman()` - 21 edges

## Surprising Connections (you probably didn't know these)
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn/main.go
- `TestCheckLibs()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn/main.go
- `TestCheckSyslog()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn/main.go
- `dispatch()` --references--> `Runner`  [EXTRACTED]
  cmd/solmq-conn/main.go → internal/runner/runner.go
- `genConfig()` --calls--> `Config()`  [EXTRACTED]
  cmd/solmq-conn/main.go → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (52 total, 11 thin omitted)

### Community 0 - "Build"
Cohesion: 0.05
Nodes (94): acc, Binder, Binding, JMSBinding, MQBinder, Opts, secretFn, Session (+86 more)

### Community 1 - "validate.go"
Cohesion: 0.07
Nodes (59): defaultsFromRaw(), Defaults, Management, Security, TLSConfig, yaml.Node, applyDest(), Cred (+51 more)

### Community 2 - "testing.T"
Cohesion: 0.15
Nodes (68): testing.T, baseKubeDeploy(), connDefaults(), dockerOK(), podmanOK(), TestBinderIdentityUsesTheCredentialPair(), TestCheckCommandMultiToken(), TestCheckContainerHostPathsUnsafe() (+60 more)

### Community 3 - "runner_test.go"
Cohesion: 0.06
Nodes (81): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), collectFlagsAndDirs(), emit(), envFlag() (+73 more)

### Community 4 - "Render"
Cohesion: 0.14
Nodes (35): Input, Instance, KV, Instance, StoreFile, yw, LogbackXML(), managementPort() (+27 more)

### Community 5 - "gen.go"
Cohesion: 0.10
Nodes (48): built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef, b64() (+40 more)

### Community 6 - "spec_test.go"
Cohesion: 0.11
Nodes (27): ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys(), TestCredEmptyBothKeyDescribe() (+19 more)

### Community 7 - "solmq-conn test catalogue"
Cohesion: 0.12
Nodes (16): cmd/solmq-conn, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/podmangen (+8 more)

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

### Community 12 - "SolaceProps"
Cohesion: 0.29
Nodes (10): MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 13 - "podmangen.go"
Cohesion: 0.20
Nodes (19): Mount, SecretRef, Unit, RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs(), fullInput() (+11 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.27
Nodes (16): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar() (+8 more)

### Community 17 - "solmq-conn -- User Guide"
Cohesion: 0.08
Nodes (26): 10. `examples`, 11. Notes and gotchas, 1. Running solmq-conn, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery, 5.1 Top-level (+18 more)

### Community 18 - "Project Instructions"
Cohesion: 0.12
Nodes (15): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 8. Solace Connector — Workflow Configuration (+7 more)

### Community 19 - "Render"
Cohesion: 0.12
Nodes (21): Input, Instance, strings.Builder, Mount, yw, Render(), renderConfig(), renderSecrets() (+13 more)

### Community 21 - "ParseEnv"
Cohesion: 0.07
Nodes (50): TestKubernetesDeleteUsesDeleteVerb(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesUnknownAction(), Env, ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored() (+42 more)

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

### Community 43 - "main_test.go"
Cohesion: 0.12
Nodes (40): dispatch(), main(), run(), assertSameNameSet(), captureStderr(), captureStdout(), keySet(), manyWorkflowsDir() (+32 more)

### Community 44 - "solmq-conn command reference"
Cohesion: 0.40
Nodes (5): All commands, Command tree, Exit codes, Flags, solmq-conn command reference

### Community 45 - "solmq-conn -- Development Guide"
Cohesion: 0.33
Nodes (6): Build, Design notes, Release (CI), solmq-conn -- Development Guide, Testing, The spec generator (`solmq-conn-generator.html`)

### Community 46 - "solmq-conn -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "generate"
Cohesion: 0.40
Nodes (5): generate, `solmq-conn generate config [-e env.yaml] [-o out]`, `solmq-conn generate docker [-e env.yaml] [-o out]`, `solmq-conn generate kubernetes [-e env.yaml] [-o out]`, `solmq-conn generate podman [-e env.yaml] [-o out]`

### Community 49 - "Command details"
Cohesion: 0.25
Nodes (8): Command details, delete, examples, help, `solmq-conn delete docker [-e env.yaml]`, `solmq-conn delete kubernetes [-e env.yaml]`, `solmq-conn delete podman [-e env.yaml]`, validate

### Community 50 - "commands.go"
Cohesion: 0.27
Nodes (10): normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), flagsLine(), invocation(), renderCommandsDoc(), usageText(), cliFlag (+2 more)

### Community 51 - "deploy"
Cohesion: 0.50
Nodes (4): deploy, `solmq-conn deploy docker [-e env.yaml]`, `solmq-conn deploy kubernetes [-e env.yaml]`, `solmq-conn deploy podman [-e env.yaml]`

## Knowledge Gaps
- **119 isolated node(s):** `cliFlag`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `Model`, `NO_COLOR`, `graphify` (+114 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Defaults` connect `validate.go` to `Build`, `testing.T`, `Render`, `gen.go`, `spec_test.go`, `SolaceProps`, `ParseEnv`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Why does `Build()` connect `Build` to `validate.go`, `SolaceProps`, `gen.go`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Why does `Kubernetes` connect `ParseEnv` to `validate.go`, `Render`, `spec_test.go`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 32 inferred relationships involving `hasErr()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckCommandMultiToken()`) actually correct?**
  _`hasErr()` has 32 INFERRED edges - model-reasoned connections that need verification._
- **Are the 16 inferred relationships involving `Build()` (e.g. with `assignBinderNames()` and `securityUserPasswordName()`) actually correct?**
  _`Build()` has 16 INFERRED edges - model-reasoned connections that need verification._
- **Are the 19 inferred relationships involving `wf()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckSideMQMissingFields()`) actually correct?**
  _`wf()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **Are the 17 inferred relationships involving `vMQ()` (e.g. with `TestBinderIdentityUsesTheCredentialPair()` and `TestCheckSideSolaceMissingAndBadScheme()`) actually correct?**
  _`vMQ()` has 17 INFERRED edges - model-reasoned connections that need verification._