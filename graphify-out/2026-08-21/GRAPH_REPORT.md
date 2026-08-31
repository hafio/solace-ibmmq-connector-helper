# Graph Report - solace-ibmmq-connector-helper  (2026-08-21)

## Corpus Check
- 86 files · ~232,475 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1626 nodes · 4814 edges · 83 communities (69 shown, 14 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 887 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `affff7d7`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- consolidate_extra_test.go
- validate.go
- .Run
- main.go
- Runner
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
- statusreport.go
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
- Env
- solmq-conn-util -- Development Guide
- Download
- runner_test.go
- Command details
- solmq-conn-util.bash
- Model
- ParseEnv
- render.go
- buildLeaderElection
- consolidate.go
- Build
- consolidate_test.go
- maven_test.go
- main_test.go
- Render
- parse_test.go
- ScriptInstalled
- loadImageLibs
- parse.go
- Defaults
- runner.go
- TestDownloadSetMapMatchesModel
- captureStdout
- statusreport/render.go
- render
- dispatch
- 10. Status: the container, the connector, or both
- ApplyTop
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- repeatableName
- 12. `download jar`
- 5. Workflow file
- watchFlag
- PodmanSecretCreate
- 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)
- 8. Secrets model

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 72 edges
2. `hasErr()` - 66 edges
3. `Download()` - 49 edges
4. `Build()` - 45 edges
5. `wfOK()` - 36 edges
6. `wf()` - 36 edges
7. `vMQ()` - 33 edges
8. `write()` - 32 edges
9. `ParseEnv()` - 32 edges
10. `Runner` - 30 edges

## Surprising Connections (you probably didn't know these)
- `renderedCompletions()` --calls--> `render()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/statusreport/render_test.go
- `TestCompletionCoversModel()` --calls--> `contains()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/runner/runner_test.go
- `TestCompletionVerbAliasesResolveToCanonical()` --calls--> `build()`  [INFERRED]
  cmd/solmq-conn-util/completion_test.go → internal/gen/gen.go
- `TestGenerateKubernetesImagePull()` --calls--> `run()`  [INFERRED]
  internal/gen/imagepull_test.go → cmd/solmq-conn-util/main.go
- `TestCheckKubeCredentialCreateRemovedKeys()` --calls--> `run()`  [INFERRED]
  internal/validate/validate_extra_test.go → cmd/solmq-conn-util/main.go

## Import Cycles
- None detected.

## Communities (83 total, 14 thin omitted)

### Community 0 - "consolidate_extra_test.go"
Cohesion: 0.17
Nodes (14): containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildLeaderElection(), TestBuildMessageLoopWarning(), TestBuildMQmTLSBundle(), TestBuildSolaceTopicSourceEmitsConsumerTopic() (+6 more)

### Community 1 - "validate.go"
Cohesion: 0.06
Nodes (73): synthWorkflows(), defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, Image, applyDest() (+65 more)

### Community 2 - ".Run"
Cohesion: 0.10
Nodes (88): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy() (+80 more)

### Community 3 - "main.go"
Cohesion: 0.10
Nodes (52): verbUsage(), absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), allowCommandFlag(), collectFlagsAndDirs() (+44 more)

### Community 4 - "Runner"
Cohesion: 0.20
Nodes (14): podmanRemove(), QuadletScope, Runner, PodmanDeploy(), PodmanRemove(), ResolveQuadletScope(), SystemctlNRestarts(), TestPodmanDeployReloadThenStart() (+6 more)

### Community 5 - "gen.go"
Cohesion: 0.08
Nodes (67): podmanDeploy(), built, DockerPlan, File, mount, NamedDoc, PodmanOpts, SecretRef (+59 more)

### Community 6 - "SolaceProps"
Cohesion: 0.26
Nodes (11): MountPath(), SolaceProps(), StorePath(), placeholderSecretRef(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.11
Nodes (19): cmd/solmq-conn-util, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+11 more)

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
Cohesion: 0.07
Nodes (69): normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), flagsLine(), flagSpan(), invocation(), renderCommandsDoc(), resolveTarget() (+61 more)

### Community 13 - "RenderRunScript"
Cohesion: 0.19
Nodes (23): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), RenderRunScript(), renderSecretPreamble(), runArgs() (+15 more)

### Community 14 - "DurableName"
Cohesion: 0.43
Nodes (5): DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 15 - "Expand"
Cohesion: 0.24
Nodes (18): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+10 more)

### Community 17 - "solmq-conn-util -- User Guide"
Cohesion: 0.14
Nodes (14): 11. `examples`, 13. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2. Quick start, 3. Commands, 4.1 Variable expansion (`${VAR}`), 4. The config file and workflow discovery (+6 more)

### Community 18 - "Solace PubSub+ Connector for IBM MQ — Configuration Guide"
Cohesion: 0.11
Nodes (18): 11. Spring SSL Bundles, 13. Logging, 14. JVM System Properties, 15. Environment Variable Overrides, 3. Spring Cloud Stream — Binders, 4. Spring Cloud Stream — Bindings (Workflows), 5. JMS Binder Options, 7. Solace Binding-Level Options (Consumer & Producer) (+10 more)

### Community 19 - "Render"
Cohesion: 0.10
Nodes (27): Input, Instance, strings.Builder, composeEscape(), Mount, yw, Render(), renderContentConfig() (+19 more)

### Community 21 - "Render"
Cohesion: 0.05
Nodes (83): Input, Instance, KV, PullSecret, StoreFile, yw, leaderMode(), LogbackXML() (+75 more)

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

### Community 31 - "statusreport.go"
Cohesion: 0.08
Nodes (34): Banner(), banner(), Bytes(), canonicalRef(), Cores(), ExitCodeText(), Workflow, ImageMismatch() (+26 more)

### Community 32 - "Solace Message Headers Reference"
Cohesion: 0.67
Nodes (3): Solace Binder Headers (`solace_scst_*`), Solace Message Headers Reference, Solace Message Headers (`solace_*`)

### Community 43 - "testing.T"
Cohesion: 0.11
Nodes (35): TestAbsPath(), TestPlatformAliasesCoverEveryPlatformExactlyOnce(), TestPlatformSpellingsAreDeterministic(), TestStatusRejectsUnsafeUserBeforeAnyExec(), TestWatchFlagAcceptsBareAndInterval(), testing.T, ParseDefaults(), TestLeaderElectionEffectiveMode() (+27 more)

### Community 44 - "Env"
Cohesion: 0.13
Nodes (19): actStatus(), checkStatusFlags(), clearScreen(), confirmInstall(), instanceNames(), loadStatusEnv(), markMissing(), statusCommand() (+11 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- Development Guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "Download"
Cohesion: 0.06
Nodes (76): reportDownload(), net/http.Client, net/http.Header, net/http.Request, net/http.Response, net/url.URL, defaultClient(), Download() (+68 more)

### Community 48 - "runner_test.go"
Cohesion: 0.08
Nodes (37): call, os.FileMode, canIVerb(), EngineInspectJSON(), PodmanSecretRemove(), Preflight(), helperProcessArgv(), TestEngineInspectJSONArgv() (+29 more)

### Community 49 - "Command details"
Cohesion: 0.07
Nodes (27): All commands, auto-complete, Command details, Command tree, deploy, download, examples, Exit codes (+19 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.25
Nodes (3): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_paths()

### Community 51 - "Model"
Cohesion: 0.26
Nodes (16): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+8 more)

### Community 52 - "ParseEnv"
Cohesion: 0.12
Nodes (24): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+16 more)

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

### Community 59 - "maven_test.go"
Cohesion: 0.11
Nodes (66): encoding/xml.Name, syslogFixturesWithDependency(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML() (+58 more)

### Community 60 - "main_test.go"
Cohesion: 0.12
Nodes (43): run(), manyWorkflowsDir(), podmanEnv(), podmanEnvSudo(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials() (+35 more)

### Community 61 - "Render"
Cohesion: 0.16
Nodes (20): breEscape(), Render(), TestFilenameAndPathConstants(), TestRenderAlignsWorkflowColumn(), TestRenderAlwaysExitsZero(), TestRenderEscapesUserForSedAddress(), TestRenderHeaderHasExecOneLiners(), TestRenderHeaderNamesEveryReportedFact() (+12 more)

### Community 62 - "parse_test.go"
Cohesion: 0.10
Nodes (30): ApplyStats(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), splitKV() (+22 more)

### Community 63 - "ScriptInstalled"
Cohesion: 0.14
Nodes (14): execArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestInstallScriptArgv(), TestInstallScriptPassesScriptOnStdinNotArgv(), TestInstallScriptUnknownPlatform(), TestRunStatusScriptArgv() (+6 more)

### Community 64 - "loadImageLibs"
Cohesion: 0.22
Nodes (16): imageSatisfies(), loadImageLibs(), omitListProvenance(), splitJarBasename(), TestImageSatisfies(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines() (+8 more)

### Community 65 - "parse.go"
Cohesion: 0.13
Nodes (29): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+21 more)

### Community 67 - "runner.go"
Cohesion: 0.11
Nodes (24): Docker(), EngineImageInspectJSON(), EngineList(), EngineStats(), Cmd, Kubernetes(), KubernetesGetJSON(), KubernetesPodsJSON() (+16 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "captureStdout"
Cohesion: 0.12
Nodes (20): captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestStatusAllSearchesByImage(), TestStatusApplicationDetailsAddsTheEnrichmentLines(), TestStatusApplicationViewRunsTheScriptAndRendersItsFacts(), TestStatusContainerDetailsSamplesAndChecksComponents(), TestStatusContainerDetailsWithoutMetricsServerDegradesToANote() (+12 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.25
Nodes (18): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+10 more)

### Community 71 - "render"
Cohesion: 0.28
Nodes (14): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+6 more)

### Community 72 - "dispatch"
Cohesion: 0.13
Nodes (35): dispatch(), captureStderr(), TestAllowCommandFlagBadValueExitsUsageError(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone(), TestDownloadMissingAndUnknownWordsRejected() (+27 more)

### Community 73 - "10. Status: the container, the connector, or both"
Cohesion: 0.18
Nodes (11): 10.1 `status container` -- the engine's view, 10.2 `status application` -- the connector's view, 10.3 `-d` / `--details`, 10.4 `--all`: find every instance by image, 10.5 `-w` / `--watch`, 10.6 `--output json`, 10.7 What the exit code means, and what each view costs, 10.8 The manual alternative (+3 more)

### Community 74 - "ApplyTop"
Cohesion: 0.25
Nodes (9): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), ParseQuantity(), Percent(), TestParseQuantity() (+1 more)

### Community 75 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 77 - "12. `download jar`"
Cohesion: 0.22
Nodes (9): 12. `download jar`, Flags and defaults, Image-aware omission, Integrity verification (sha1), `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, The two sets, `--url` overrides all resolution (+1 more)

### Community 78 - "5. Workflow file"
Cohesion: 0.29
Nodes (7): 5.1 Top-level, 5.2 `solace:` options, 5.3 `mq:` options, 5.4 Destinations, durable names, passthrough, 5.5 Event-driven guidance (warnings), 5.6 Reusable connections (`conn-ref`), 5. Workflow file

### Community 80 - "PodmanSecretCreate"
Cohesion: 0.40
Nodes (5): PodmanSecretCreate(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure(), TestPodmanSecretCreateSkipsCreateWhenRmFails()

### Community 81 - "7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 7.0 image and timezone (shared by every platform), 7.1 kubernetes, 7.2 docker, 7.3 podman, 7. Deploy targets (`kubernetes:`, `docker:`, `podman:`)

### Community 82 - "8. Secrets model"
Cohesion: 0.40
Nodes (5): 8.1 Declaring a credential, 8.2 Stable names, 8.3 How each platform delivers them, 8.4 Registry credentials (pulling the image), 8. Secrets model

## Knowledge Gaps
- **151 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `call`, `Defaults` (+146 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Scan()` connect `Scan` to `validate.go`, `golden_test.go`, `main.go`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Download()` connect `Download` to `loadImageLibs`, `maven_test.go`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Application()` connect `Build` to `Model`, `gen.go`, `render.go`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Are the 67 inferred relationships involving `dispatch()` (e.g. with `TestAllowCommandFlagBadValueExitsUsageError()` and `TestAllowCommandFlagRepeatableThreadsToRunner()`) actually correct?**
  _`dispatch()` has 67 INFERRED edges - model-reasoned connections that need verification._
- **Are the 47 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 47 INFERRED edges - model-reasoned connections that need verification._
- **Are the 40 inferred relationships involving `Download()` (e.g. with `imageSatisfies()` and `loadImageLibs()`) actually correct?**
  _`Download()` has 40 INFERRED edges - model-reasoned connections that need verification._
- **Are the 20 inferred relationships involving `Build()` (e.g. with `assignBinderNames()` and `stableName()`) actually correct?**
  _`Build()` has 20 INFERRED edges - model-reasoned connections that need verification._