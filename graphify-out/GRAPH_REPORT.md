# Graph Report - solace-ibmmq-connector-helper  (2026-09-30)

## Corpus Check
- 107 files · ~332,554 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2130 nodes · 7931 edges · 129 communities (92 shown, 37 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1136 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7e81959c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Build
- validate.go
- validate_extra_test.go
- dispatch
- htmlgolden_test.go
- GenerateDocker
- statusreport_test.go
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- testing.T
- scan_test.go
- completion.go
- .Write
- yw
- completion_test.go
- solmq-conn-util user guide
- ExecArgv
- dockergen_test.go
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- gen_extra_test.go
- podmangen_test.go
- main.go
- examples.go
- withProgressClock
- attachRunner
- podmanDeploy
- gen.go
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
- errExit
- solmq-conn-util -- Development Guide
- libs_test.go
- DEVELOPMENT.md
- Kubernetes
- Command details
- solmq-conn-util.bash
- Model
- Application
- spec_test.go
- instances.go
- propsNode
- support_test.go
- commands.go
- maven_test.go
- write
- Docker
- logs.go
- parse_test.go
- go_pkg_os
- instanceFromInspect
- Defaults
- statusscript_test.go
- TestDownloadSetMapMatchesModel
- Issue
- statusreport/render.go
- statusreport/render_test.go
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- rawDefaults
- 13. Logs: the lines behind the state
- 6. Workflow file
- ResolveCredentials
- buildLeaderElection
- JavaOptions
- targets_test.go
- maven.go
- nodeToProps
- solmq-conn-util abbreviations
- GenerateKubernetes
- libs.go
- SafeToken
- TestSanitizeAndIsTCPS
- runner.go
- kubernetes.go
- Cmd
- abbreviation_doc_test.go
- Expand
- ParseCommand
- runner_test.go
- shell.go
- main_test.go
- kubernetes_test.go
- uuid.go
- misplacedEnvTransforms
- solmq-conn-util command reference
- LogsArgv
- Writer
- Workload
- XML
- Resolver
- allowCommandValue
- repeatableName
- QuoteScalar
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_consolidate
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_deploy
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_dockergen
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_examples
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_gen
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_libs
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_logback
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_podmangen
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_render
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_runner
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_scan
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_spec
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_statusreport
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_statusscript
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_tls
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_validate
- go_pkg_github_com_solacecommunity_hafio_solace_connectors_ibmmq_solmq_conn_internal_yamlwriter

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 139 edges
2. `hasErr()` - 87 edges
3. `captureStderr()` - 68 edges
4. `write()` - 65 edges
5. `wfOK()` - 53 edges
6. `Download()` - 52 edges
7. `Build()` - 51 edges
8. `imageOK()` - 45 edges
9. `captureStdout()` - 44 edges
10. `Runner` - 42 edges

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

## Communities (129 total, 37 thin omitted)

### Community 0 - "Build"
Cohesion: 0.11
Nodes (31): Opts, Build(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestApplyStatusAccessNoOperatorUsers(), TestBuildLeaderElectionSharesBinderSecretNames(), TestBuildMQmTLSBundle() (+23 more)

### Community 1 - "validate.go"
Cohesion: 0.15
Nodes (43): Defaults, Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources() (+35 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.07
Nodes (114): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+106 more)

### Community 3 - "dispatch"
Cohesion: 0.10
Nodes (40): dispatch(), captureStdout(), containsToken(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList() (+32 more)

### Community 4 - "htmlgolden_test.go"
Cohesion: 0.19
Nodes (21): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+13 more)

### Community 5 - "GenerateDocker"
Cohesion: 0.17
Nodes (15): DockerPlan, mount, SecretRef, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform(), GenerateDocker() (+7 more)

### Community 6 - "statusreport_test.go"
Cohesion: 0.13
Nodes (13): go_pkg_time, ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues() (+5 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (24): cli, cmd/solmq-conn-util, Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples (+16 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "testing.T"
Cohesion: 0.07
Nodes (47): resolveTarget(), run(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+39 more)

### Community 11 - "scan_test.go"
Cohesion: 0.21
Nodes (21): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+13 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - ".Write"
Cohesion: 0.40
Nodes (5): TestWriteMkdirError(), regularFile(), TestHelperProcess(), TestOSStreamDeliversOutputBeforeExitAndCancelIsCleanEnd(), writerFunc

### Community 14 - "yw"
Cohesion: 0.30
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (25): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+17 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.08
Nodes (24): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+16 more)

### Community 18 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 19 - "dockergen_test.go"
Cohesion: 0.10
Nodes (35): Input, Instance, composeEscape(), composeQuote(), yw, Render(), renderContentConfig(), renderHealthcheck() (+27 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.09
Nodes (54): Input, Instance, KV, envValueQuote(), PullSecret, StoreFile, yw, leaderMode() (+46 more)

### Community 23 - "gen_extra_test.go"
Cohesion: 0.18
Nodes (24): Config(), issuesContain(), synthWorkflowFiles(), synthWorkflows(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigRejectsSecretNameConflict(), TestConfigRendersTransformHeadersUnderItsWorkflow() (+16 more)

### Community 24 - "podmangen_test.go"
Cohesion: 0.21
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), systemdEnv(), fullInput(), Input, minimalInput() (+12 more)

### Community 25 - "main.go"
Cohesion: 0.14
Nodes (36): runLogs(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains(), downloadDeployedImage(), envFlag(), flagExit() (+28 more)

### Community 26 - "examples.go"
Cohesion: 0.50
Nodes (3): go_pkg_embed, go_pkg_io_fs, Write()

### Community 27 - "withProgressClock"
Cohesion: 0.19
Nodes (15): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases(), TestProgressStepsPauseFinishesTheStepFirst() (+7 more)

### Community 28 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 29 - "podmanDeploy"
Cohesion: 0.17
Nodes (15): podmanDeploy(), podmanRemove(), NamedDoc, PodmanPlan, PodmanSecretStoreName(), QuadletScope, PodmanDeploy(), PodmanRemove() (+7 more)

### Community 30 - "gen.go"
Cohesion: 0.37
Nodes (14): go_pkg_bytes, go_pkg_crypto_rand, go_pkg_encoding_base64, go_pkg_fmt, go_pkg_gopkg_in_yaml_v3, go_pkg_path_filepath, go_pkg_regexp, go_pkg_strings (+6 more)

### Community 31 - "statusreport.go"
Cohesion: 0.11
Nodes (27): heapValue(), parseHeap(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), Instance (+19 more)

### Community 32 - "statusCollector"
Cohesion: 0.15
Nodes (12): sortInstances(), actStatus(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), watchStatus(), Instance (+4 more)

### Community 43 - "Side"
Cohesion: 0.08
Nodes (13): fixedLeaderNames(), Security, TLSConfig, Image, Cred, Side, placeholderSecretRef(), edaAdvisory() (+5 more)

### Community 44 - "errExit"
Cohesion: 0.21
Nodes (26): absPath(), absResolver(), actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit() (+18 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (54): net/http.Header, Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile() (+46 more)

### Community 47 - "DEVELOPMENT.md"
Cohesion: 0.33
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 48 - "Kubernetes"
Cohesion: 0.16
Nodes (11): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Deployment, Kubernetes, Resources (+3 more)

### Community 49 - "Command details"
Cohesion: 0.09
Nodes (23): auto-complete, cli, Command details, deploy, download, examples, generate, help (+15 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.13
Nodes (21): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+13 more)

### Community 52 - "Application"
Cohesion: 0.17
Nodes (18): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationLeaderElection() (+10 more)

### Community 54 - "spec_test.go"
Cohesion: 0.08
Nodes (37): go_pkg_slices, ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), digitRun(), isDigit(), ParseWorkflow(), TestConnRefSideMayTuneBinding() (+29 more)

### Community 55 - "instances.go"
Cohesion: 0.18
Nodes (28): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+20 more)

### Community 56 - "propsNode"
Cohesion: 0.22
Nodes (11): containsSub(), yaml.Node, propsNode(), TestBuildCipherConflictWarning(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildLeaderElectionWarningsReachBuild(), TestBuildMessageLoopWarning() (+3 more)

### Community 57 - "support_test.go"
Cohesion: 0.33
Nodes (8): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 58 - "commands.go"
Cohesion: 0.16
Nodes (28): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+20 more)

### Community 59 - "maven_test.go"
Cohesion: 0.25
Nodes (33): metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures(), pomXMLBody() (+25 more)

### Community 60 - "write"
Cohesion: 0.08
Nodes (45): podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials(), TestDeployDockerSeamComposeFileSurvivesFailedRun() (+37 more)

### Community 61 - "Docker"
Cohesion: 0.11
Nodes (23): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Secrets, Service, applyDockerDefaults() (+15 more)

### Community 62 - "logs.go"
Cohesion: 0.17
Nodes (11): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), go_pkg_context, go_pkg_errors, go_pkg_os_signal (+3 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (33): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods() (+25 more)

### Community 64 - "go_pkg_os"
Cohesion: 0.05
Nodes (51): newProgress(), progressElapsed(), progressTrim(), clearScreen(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow() (+43 more)

### Community 65 - "instanceFromInspect"
Cohesion: 0.13
Nodes (23): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+15 more)

### Community 67 - "statusscript_test.go"
Cohesion: 0.13
Nodes (28): go_pkg_os_exec, breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent() (+20 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "Issue"
Cohesion: 0.24
Nodes (10): TestToIssues(), toIssues(), yaml.Node, issueWith(), runTransform(), TestTransformHeadersShapeErrors(), TestTransformHeadersValidBlockPasses(), TestTransformHeadersWarnings() (+2 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (19): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+11 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (15): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+7 more)

### Community 72 - "nsList"
Cohesion: 0.19
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveKubernetesSeamHappyPath(), TestRemoveKubernetesSeamPromptsBeforeTearingDown(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag() (+6 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: verify before relying on tcp syslog, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "rawDefaults"
Cohesion: 0.22
Nodes (13): defaultsFromRaw(), yaml.Node, applyDest(), yaml.Node, nodePtr(), rawDefaults, rawLeader, rawLogging (+5 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.25
Nodes (8): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Header transforms (`transform-headers`), 6. Workflow file

### Community 79 - "ResolveCredentials"
Cohesion: 0.24
Nodes (10): dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues() (+2 more)

### Community 80 - "buildLeaderElection"
Cohesion: 0.12
Nodes (19): leaderNameFn, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), buildBundle(), buildLeaderElection(), TestAppendPassthroughCollision() (+11 more)

### Community 81 - "JavaOptions"
Cohesion: 0.15
Nodes (12): SecretRef, JavaOptions, yaml.Node, YAMLKind(), decodeCreate(), yaml.Node, Input, Instance (+4 more)

### Community 82 - "targets_test.go"
Cohesion: 0.10
Nodes (29): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+21 more)

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (34): go_pkg_encoding_xml, encoding/xml.Name, acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML() (+26 more)

### Community 84 - "nodeToProps"
Cohesion: 0.33
Nodes (6): TestFormatScalarQuoting(), TestNodeToProps(), FormatScalar(), Model, yaml.Node, nodeToProps()

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "GenerateKubernetes"
Cohesion: 0.23
Nodes (12): File, KubeOpts, build(), TestGenerateKubernetesCarriesStatusScript(), TestGenValidateStoresWarning(), TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestRetiredCreateNamesDeployButDoNotValidate(), TestValidateCleanSpecStillPasses() (+4 more)

### Community 87 - "libs.go"
Cohesion: 0.09
Nodes (30): reportDownload(), go_pkg_crypto_sha1, go_pkg_io, go_pkg_net_http, go_pkg_net_url, net/http.Client, net/http.Request, net/http.Response (+22 more)

### Community 88 - "SafeToken"
Cohesion: 0.18
Nodes (11): CheckDeployCommand(), TestCheckDeployCommandAcceptReject(), TestCheckDeployCommandEndOfFlagsMarkerMidCommand(), TestCheckDeployCommandErrorTexts(), TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL() (+3 more)

### Community 89 - "TestSanitizeAndIsTCPS"
Cohesion: 0.50
Nodes (4): TestSanitizeAndIsTCPS(), isTCPS(), sanitize(), assignBinderNames()

### Community 90 - "runner.go"
Cohesion: 0.28
Nodes (18): Docker(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Runner, Kubernetes(), KubernetesGetJSON() (+10 more)

### Community 91 - "kubernetes.go"
Cohesion: 0.19
Nodes (14): isClusterDefault(), isOurs(), namespaceOccupants(), go_pkg_encoding_json, go_pkg_sort, go_pkg_strconv, nsItem, Libs (+6 more)

### Community 92 - "Cmd"
Cohesion: 0.17
Nodes (13): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Attacher, Cmd (+5 more)

### Community 93 - "abbreviation_doc_test.go"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 94 - "Expand"
Cohesion: 0.18
Nodes (19): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+11 more)

### Community 95 - "ParseCommand"
Cohesion: 0.15
Nodes (13): ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure() (+5 more)

### Community 96 - "runner_test.go"
Cohesion: 0.06
Nodes (55): go_pkg_runtime, os.FileMode, canIVerb(), Preflight(), attachFiles(), contains(), helperProcessArgv(), readFile() (+47 more)

### Community 97 - "shell.go"
Cohesion: 0.49
Nodes (8): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), runShell(), shellInvocation(), splitAtSeparator(), shellOpts

### Community 98 - "main_test.go"
Cohesion: 0.08
Nodes (55): captureStderr(), downloadEnvWithImage(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone() (+47 more)

### Community 99 - "kubernetes_test.go"
Cohesion: 0.31
Nodes (9): go_pkg_reflect, first(), parseKube(), second(), TestCreatedNames(), TestDerivedObjectNames(), TestSecretsCreateFollowsAnAlias(), TestSecretsCreateRejectsOtherShapes() (+1 more)

### Community 100 - "uuid.go"
Cohesion: 0.36
Nodes (6): go_pkg_encoding_hex, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

### Community 101 - "misplacedEnvTransforms"
Cohesion: 0.57
Nodes (7): envTransforms(), appendTransformKeys(), documentMapping(), yaml.Node, mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 102 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 105 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 106 - "XML"
Cohesion: 0.50
Nodes (4): TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

### Community 107 - "Resolver"
Cohesion: 0.17
Nodes (12): testResolver(), TestShippedExamplesGenerateConfig(), b64(), TestResolveStores(), TestTargetMounts(), Resolver, resolveStores(), targetMounts() (+4 more)

## Knowledge Gaps
- **131 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Defaults`, `kubeDeployment` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 218 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **37 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `progress` connect `go_pkg_os` to `statusCollector`, `instanceFromInspect`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `Cred` connect `Side` to `buildLeaderElection`, `validate.go`, `Model`, `gen.go`?**
  _High betweenness centrality (0.011) - this node is a cross-community bridge._
- **Why does `dispatch()` connect `dispatch` to `main_test.go`, `runner.go`, `nsList`, `testing.T`, `main.go`, `commands.go`, `withProgressClock`, `write`?**
  _High betweenness centrality (0.010) - this node is a cross-community bridge._
- **Are the 134 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 134 INFERRED edges - model-reasoned connections that need verification._
- **Are the 64 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 64 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Build` be split into smaller, more focused modules?**
  _Cohesion score 0.11182795698924732 - nodes in this community are weakly interconnected._