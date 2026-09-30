# Graph Report - solace-ibmmq-connector-helper  (2026-09-30)

## Corpus Check
- 107 files · ~336,800 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2151 nodes · 8006 edges · 129 communities (93 shown, 36 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1155 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `dc3253b1`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- consolidate_test.go
- validate.go
- validate_extra_test.go
- main_test.go
- deploy.go
- Defaults
- .add
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- testing.T
- scan_test.go
- completion.go
- os.File
- render/render.go
- completion_test.go
- solmq-conn-util user guide
- ExecArgv
- dockergen_test.go
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- gen.go
- podmangen_test.go
- runAction
- progress
- withProgressClock
- attachRunner
- runner.go
- go_pkg_testing
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
- Kubernetes
- Command details
- solmq-conn-util.bash
- model.go
- Build
- spec_test.go
- instances.go
- consolidate_extra_test.go
- go_pkg_path_filepath
- commands.go
- maven_test.go
- write
- kubernetes.go
- logs.go
- parse_test.go
- go_pkg_os
- parse.go
- Defaults
- statusscript_test.go
- TestDownloadSetMapMatchesModel
- transform_test.go
- statusreport/render.go
- statusreport/render_test.go
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- spec.go
- 13. Logs: the lines behind the state
- 6. Workflow file
- gen/imagepull_test.go
- filenameFromEscapedPath
- status.go
- targets_test.go
- maven.go
- consolidate.go
- solmq-conn-util abbreviations
- auto-complete
- libs.go
- SafeToken
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- Runner
- namespace.go
- Cmd
- renderCommandsDoc
- expand_test.go
- runner_test.go
- helperProcessArgv
- shell.go
- dispatch
- kubernetes_test.go
- 9. Secrets model
- transform.go
- solmq-conn-util command reference
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- go_pkg_strings
- status
- env.go
- cmd/solmq-conn-util
- main.go
- watchFlag
- go_pkg_path
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
1. `dispatch()` - 140 edges
2. `hasErr()` - 87 edges
3. `captureStderr()` - 69 edges
4. `write()` - 65 edges
5. `Download()` - 56 edges
6. `wfOK()` - 53 edges
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

## Communities (129 total, 36 thin omitted)

### Community 0 - "consolidate_test.go"
Cohesion: 0.40
Nodes (12): binderNames(), binderOf(), eqStrs(), Model, mqSide(), solaceSide(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+4 more)

### Community 1 - "validate.go"
Cohesion: 0.21
Nodes (20): Workflow, checkDuplicateSources(), checkJavaOptions(), checkKeyAliasConflicts(), checkLeaderSessionPassword(), checkPasswordConflicts(), checkStatusUser(), checkSyslog() (+12 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.07
Nodes (110): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+102 more)

### Community 3 - "main_test.go"
Cohesion: 0.10
Nodes (42): captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestGenerateKubernetesStdout(), TestLoadEnvWorkflowsDirRelativeToEnvFile(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance() (+34 more)

### Community 4 - "deploy.go"
Cohesion: 0.14
Nodes (21): Input, Instance, KV, envValueQuote(), PullSecret, StoreFile, yw, leaderMode() (+13 more)

### Community 5 - "Defaults"
Cohesion: 0.17
Nodes (13): defaultsFromRaw(), Defaults, Security, yaml.Node, checkLeaderElection(), checkRemovedDefaultsKeys(), LeaderElection, rawDefaults (+5 more)

### Community 6 - ".add"
Cohesion: 0.31
Nodes (14): checkContainerTarget(), checkDocker(), checkImagePull(), checkKube(), checkLibs(), checkPodman(), checkPortRange(), checkPorts() (+6 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "testing.T"
Cohesion: 0.09
Nodes (38): run(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+30 more)

### Community 11 - "scan_test.go"
Cohesion: 0.25
Nodes (19): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+11 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "os.File"
Cohesion: 0.15
Nodes (11): newProgress(), clearScreen(), enableVirtualTerminal(), enableVirtualTerminal(), os.File, TestWriteMkdirError(), Attacher, regularFile() (+3 more)

### Community 14 - "render/render.go"
Cohesion: 0.38
Nodes (15): blockIndicator(), yaml.Node, yw, q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer() (+7 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "ExecArgv"
Cohesion: 0.11
Nodes (18): ExecArgv(), InstallScript(), RunStatusScript(), ScriptInstalled(), TestExecArgvPerPlatform(), TestExecArgvRefusesATTYWithoutStdin(), TestExecArgvUnknownPlatform(), TestInstallScriptArgv() (+10 more)

### Community 19 - "dockergen_test.go"
Cohesion: 0.13
Nodes (31): Input, Instance, composeEscape(), composeQuote(), yw, Render(), renderContentConfig(), renderHealthcheck() (+23 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.12
Nodes (43): go_pkg_crypto_sha1, go_pkg_encoding_hex, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5(), ManagementPort() (+35 more)

### Community 23 - "gen.go"
Cohesion: 0.06
Nodes (92): built, DockerPlan, File, KubeOpts, mount, NamedDoc, go_pkg_crypto_rand, SecretRef (+84 more)

### Community 24 - "podmangen_test.go"
Cohesion: 0.17
Nodes (24): Mount, SecretRef, Unit, leaderLabels(), RenderQuadlet(), seconds(), systemdEnv(), fullInput() (+16 more)

### Community 25 - "runAction"
Cohesion: 0.20
Nodes (28): resolveTarget(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), downloadDeployedImage(), envFlag(), flagExit() (+20 more)

### Community 26 - "progress"
Cohesion: 0.32
Nodes (3): sync.Mutex, progress, progressMode

### Community 27 - "withProgressClock"
Cohesion: 0.16
Nodes (17): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+9 more)

### Community 28 - "attachRunner"
Cohesion: 0.20
Nodes (8): attached, attachRunner, fakeCall, fakeRunner, queuedResp, queueRunner, streamed, streamRunner

### Community 29 - "runner.go"
Cohesion: 0.13
Nodes (20): canIVerb(), LogsOpts, QuadletScope, LogsArgv(), logsCommonFlags(), PodmanDeploy(), PodmanRemove(), ResolveQuadletScope() (+12 more)

### Community 30 - "go_pkg_testing"
Cohesion: 0.11
Nodes (24): leaderNameFn, go_pkg_testing, stableName(), stableToken(), TestBinderFieldsCarryStablePlaceholders(), TestEnvCredentialsShareOneMountName(), TestGeneratedSecretNamesStayOutOfChildEnvDanger(), TestSecretNameConflictIsRecorded() (+16 more)

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (43): go_pkg_time, time.Time, heapValue(), parseHeap(), Age(), Banner(), banner(), Bytes() (+35 more)

### Community 32 - "statusCollector"
Cohesion: 0.22
Nodes (10): sortInstances(), MergeService(), ObjectExists(), ParseDeployment(), TestObjectExists(), TestParseDeploymentAndService(), Instance, Report (+2 more)

### Community 43 - "Side"
Cohesion: 0.13
Nodes (11): Cred, Side, placeholderSecretRef(), checkConnections(), checkCred(), checkDefaultsCredentials(), checkSide(), checkSideCredentials() (+3 more)

### Community 44 - "errExit"
Cohesion: 0.23
Nodes (25): actDocker(), actKubernetes(), actPodman(), emit(), errExit(), failFast(), genConfig(), genDocker() (+17 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.12
Nodes (49): go_pkg_net_http, go_pkg_testing_fstest, net/http.Header, Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic() (+41 more)

### Community 48 - "Kubernetes"
Cohesion: 0.23
Nodes (7): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), Kubernetes, Service

### Community 49 - "Command details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "model.go"
Cohesion: 0.25
Nodes (17): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+9 more)

### Community 52 - "Build"
Cohesion: 0.26
Nodes (20): Build(), Model, Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough() (+12 more)

### Community 54 - "spec_test.go"
Cohesion: 0.08
Nodes (34): go_pkg_slices, ParseDefaults(), applyKubeDefaults(), ParseKubernetes(), ParseWorkflow(), TestBaseName(), TestConnRefSideMayTuneBinding(), TestCredCreateRemovedKeys() (+26 more)

### Community 55 - "instances.go"
Cohesion: 0.17
Nodes (29): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+21 more)

### Community 56 - "consolidate_extra_test.go"
Cohesion: 0.11
Nodes (26): secretFn, go_pkg_reflect, buildLeaderElection(), containsSub(), fixedLeaderNames(), yaml.Node, propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers() (+18 more)

### Community 57 - "go_pkg_path_filepath"
Cohesion: 0.17
Nodes (16): addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocInSync(), TestAbbreviationDocTableShape(), normLF() (+8 more)

### Community 58 - "commands.go"
Cohesion: 0.20
Nodes (22): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+14 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (37): TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep() (+29 more)

### Community 60 - "write"
Cohesion: 0.09
Nodes (40): downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials() (+32 more)

### Community 61 - "kubernetes.go"
Cohesion: 0.10
Nodes (28): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), decodeCreate(), Deployment, Resources, Secrets, yaml.Node (+20 more)

### Community 62 - "logs.go"
Cohesion: 0.17
Nodes (11): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), go_pkg_context, go_pkg_errors, go_pkg_os_signal (+3 more)

### Community 63 - "parse_test.go"
Cohesion: 0.11
Nodes (28): EngineNamesByImage(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys(), TestApplyStats(), TestApplyTop() (+20 more)

### Community 64 - "go_pkg_os"
Cohesion: 0.06
Nodes (52): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+44 more)

### Community 65 - "parse.go"
Cohesion: 0.16
Nodes (25): encoding/json.RawMessage, ApplyStats(), ApplyTop(), connectorIndex(), digestFrom(), engineComponents(), exitCode(), Instance (+17 more)

### Community 67 - "statusscript_test.go"
Cohesion: 0.14
Nodes (27): go_pkg_os_exec, Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 68 - "TestDownloadSetMapMatchesModel"
Cohesion: 0.48
Nodes (7): assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames(), V

### Community 69 - "transform_test.go"
Cohesion: 0.24
Nodes (12): go_pkg_gopkg_in_yaml_v3, TestMisplacedEnvTransforms(), TestParseWorkflowTransformHeaders(), transformFile(), yaml.Node, issueWith(), runTransform(), TestMisplacedTransformsAreErrors() (+4 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.22
Nodes (20): Instance, Report, View, Workflow, groupOf(), JSON(), namespaceScope(), noteLine() (+12 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (15): Report, render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic() (+7 more)

### Community 72 - "nsList"
Cohesion: 0.12
Nodes (22): nsItemJSON(), nsList(), TestLogsPlatformFlagSkipsTheMenu(), TestLogsPlatformMenuWhenSeveralSectionsArePresent(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestPlatformMenuOnMultipleSections(), TestRemoveDeclinedTouchesNothing() (+14 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 76 - "spec.go"
Cohesion: 0.29
Nodes (11): applyDest(), digitRun(), yaml.Node, isDigit(), nodePtr(), TestWorkflowFileLess(), WorkflowFileLess(), rawMQ (+3 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.25
Nodes (8): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Header transforms (`transform-headers`), 6. Workflow file

### Community 79 - "gen/imagepull_test.go"
Cohesion: 0.46
Nodes (7): dockerConfigJSON(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull(), TestResolvePullSecret()

### Community 80 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 81 - "status.go"
Cohesion: 0.11
Nodes (20): confirmInstall(), instanceNames(), markMissing(), quoteAll(), go_pkg_strconv, JavaOptions, yaml.Node, JavaEnv() (+12 more)

### Community 82 - "targets_test.go"
Cohesion: 0.16
Nodes (20): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestApplyDockerDefaultsFillsMissing() (+12 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (38): go_pkg_encoding_xml, encoding/xml.Name, imageSatisfies(), acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties() (+30 more)

### Community 84 - "consolidate.go"
Cohesion: 0.11
Nodes (21): Opts, appendPassthrough(), applyStatusAccess(), binderOwner(), displayName(), TestAppendPassthroughCollision(), TestDisplayName(), TestFormatScalarQuoting() (+13 more)

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (27): go_pkg_io, go_pkg_net_url, net/http.Client, net/http.Request, net/http.Response, net/url.URL, defaultClient(), downloadOne() (+19 more)

### Community 88 - "SafeToken"
Cohesion: 0.20
Nodes (10): CheckDeployCommand(), checkImage(), checkSecurityUserRoles(), TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL(), safeShellChars() (+2 more)

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "Runner"
Cohesion: 0.10
Nodes (25): EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), Runner, KubernetesGetJSON(), KubernetesListJSON(), KubernetesPodsJSON() (+17 more)

### Community 91 - "namespace.go"
Cohesion: 0.27
Nodes (10): confirmNamespaceRemoval(), isClusterDefault(), isOurs(), namespaceOccupants(), go_pkg_encoding_base64, go_pkg_encoding_json, renderWithPull(), TestImagePullSecretPayloadIsOpaqueToDeploy() (+2 more)

### Community 92 - "Cmd"
Cohesion: 0.23
Nodes (12): call, context.Context, io.Writer, os/exec.Cmd, applyCmdEnv(), applyCmdInput(), Cmd, Streamer (+4 more)

### Community 93 - "renderCommandsDoc"
Cohesion: 0.33
Nodes (9): abbrevFlagByShort(), abbrevTable(), renderAbbreviationDoc(), flagsLine(), flagSpan(), renderCommandsDoc(), tableCell(), abbrevRow (+1 more)

### Community 94 - "expand_test.go"
Cohesion: 0.22
Nodes (19): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+11 more)

### Community 95 - "runner_test.go"
Cohesion: 0.10
Nodes (31): go_pkg_runtime, os.FileMode, Docker(), ParseCommand(), PodmanSecretCreate(), PodmanSecretRemove(), Preflight(), TestOSStreamRejectsEmptyAndUnresolvableArgv() (+23 more)

### Community 96 - "helperProcessArgv"
Cohesion: 0.12
Nodes (24): attachFiles(), helperProcessArgv(), readFile(), TestOSAttachEnvReachesChild(), TestOSAttachHandsTheChildTheCallersFilesNotPipes(), TestOSAttachRefusesACmdCarryingStdinText(), TestOSAttachRefusesANilFile(), TestOSAttachRejectsEmptyAndUnresolvableArgv() (+16 more)

### Community 97 - "shell.go"
Cohesion: 0.50
Nodes (7): actShell(), attachShell(), checkShellFlags(), ignoreInterruptWhileAttached(), shellInvocation(), splitAtSeparator(), shellOpts

### Community 98 - "dispatch"
Cohesion: 0.08
Nodes (53): dispatch(), captureStderr(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone(), TestDownloadMissingAndUnknownWordsRejected() (+45 more)

### Community 99 - "kubernetes_test.go"
Cohesion: 0.36
Nodes (8): first(), parseKube(), second(), TestCreatedNames(), TestDerivedObjectNames(), TestSecretsCreateFollowsAnAlias(), TestSecretsCreateRejectsOtherShapes(), TestSecretsCreateSpellings()

### Community 100 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "transform.go"
Cohesion: 0.76
Nodes (6): appendTransformKeys(), documentMapping(), yaml.Node, mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 102 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "go_pkg_strings"
Cohesion: 0.10
Nodes (19): progressTrim(), go_pkg_fmt, go_pkg_strings, go_pkg_sync, strings.Builder, mustWrite(), TestWorkflowExamplesMatchGoldenSpecs(), TestWriteCreatesSkipsForces() (+11 more)

### Community 105 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 106 - "env.go"
Cohesion: 0.19
Nodes (10): TestBothConfigsReadTheSameThreeProperties(), TestContainerPathAndFileNameAgree(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML(), workflowsFromRaw(), Image, rawEnv (+2 more)

### Community 107 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 108 - "main.go"
Cohesion: 0.11
Nodes (17): absPath(), absResolver(), confirmRemove(), envPairs(), fileReader(), main(), platformSpellings(), promptPlatformMenu() (+9 more)

## Knowledge Gaps
- **131 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Defaults`, `kubeDeployment` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 219 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **36 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Kubernetes` connect `Kubernetes` to `validate.go`, `validate_extra_test.go`, `kubernetes_test.go`, `deploy.go`, `.add`, `env.go`, `errExit`, `deploy_test.go`, `spec_test.go`, `instances.go`, `kubernetes.go`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Why does `Side` connect `Side` to `consolidate_test.go`, `validate.go`, `validate_extra_test.go`, `Defaults`, `spec.go`, `consolidate.go`, `consolidate_extra_test.go`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 135 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 135 INFERRED edges - model-reasoned connections that need verification._
- **Are the 64 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 64 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07452258965999069 - nodes in this community are weakly interconnected._
- **Should `main_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09523809523809523 - nodes in this community are weakly interconnected._