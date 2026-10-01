# Graph Report - solace-ibmmq-connector-helper  (2026-10-01)

## Corpus Check
- 107 files · ~336,976 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .jks 2, .graphify-bak 1)

## Summary
- 2150 nodes · 8004 edges · 127 communities (89 shown, 38 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 1156 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `148155dc`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- GenerateDocker
- validate.go
- validate_extra_test.go
- dispatch
- GeneratePodman
- Defaults
- statusreport_test.go
- solmq-conn-util test catalogue
- dev.sh
- dev.ps1
- main_test.go
- scan_test.go
- completion.go
- go_pkg_os
- builtinList
- completion_test.go
- solmq-conn-util user guide
- libs/image_test.go
- dockergen_test.go
- deploy_test.go
- github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn
- gen_extra_test.go
- podmangen_test.go
- main.go
- progress
- withProgressClock
- testcatalog_test.go
- go_pkg_path_filepath
- tls_test.go
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
- render/render.go
- commands_doc_test.go
- spec_test.go
- instances.go
- consolidate_extra_test.go
- abbreviation_doc_test.go
- commands.go
- maven_test.go
- write
- Docker
- logs.go
- parse_test.go
- libs/image.go
- instanceFromInspect
- Defaults
- uuid.go
- TestDownloadSetMapMatchesModel
- Issue
- statusreport/render.go
- statusreport/render_test.go
- nsList
- 12. Status: the container, the connector, or both
- 14. cli: a shell inside the instance
- 10. `download jar`
- Writer
- 13. Logs: the lines behind the state
- 6. Workflow file
- resolvePullSecret
- resolveStores
- YAMLKind
- targets_test.go
- maven.go
- Libs
- solmq-conn-util abbreviations
- auto-complete
- libs.go
- SafeToken
- 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)
- Workload
- namespaceOccupants
- runner.go
- fakeRunner
- Expand
- ParseCommand
- runner_test.go
- repeatableName
- testing.T
- kubernetes_test.go
- 9. Secrets model
- misplacedEnvTransforms
- solmq-conn-util command reference
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- gen.go
- status
- htmlgolden_test.go
- cmd/solmq-conn-util
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

## Communities (127 total, 38 thin omitted)

### Community 0 - "GenerateDocker"
Cohesion: 0.15
Nodes (18): DockerPlan, File, SecretRef, testResolver(), TestShippedExamplesGenerateConfig(), build(), TestConfigRejectsSecretNameConflict(), TestGenValidateStoresWarning() (+10 more)

### Community 1 - "validate.go"
Cohesion: 0.15
Nodes (40): Workflow, checkConnections(), checkContainerTarget(), checkCred(), checkDefaultsCredentials(), checkDocker(), checkDuplicateSources(), checkImage() (+32 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.07
Nodes (114): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey() (+106 more)

### Community 3 - "dispatch"
Cohesion: 0.10
Nodes (41): dispatch(), captureStdout(), containsToken(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance(), TestLogsIndexSelectsFromTheSortedList() (+33 more)

### Community 4 - "GeneratePodman"
Cohesion: 0.13
Nodes (19): mount, NamedDoc, Mount, TestNamesAndPaths(), TestTargetMounts(), GeneratePodman(), PodmanPlan, pathIn() (+11 more)

### Community 5 - "Defaults"
Cohesion: 0.09
Nodes (25): defaultsFromRaw(), Defaults, Security, TLSConfig, yaml.Node, applyDest(), yaml.Node, nodePtr() (+17 more)

### Community 6 - "statusreport_test.go"
Cohesion: 0.13
Nodes (13): go_pkg_time, ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues() (+5 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.10
Nodes (20): Contents, How the suite is built, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs (+12 more)

### Community 8 - "dev.sh"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "dev.ps1"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "main_test.go"
Cohesion: 0.07
Nodes (47): run(), manyWorkflowsDir(), TestAbsPath(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestAutoCompleteDispatchPrintsScript(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner() (+39 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.21
Nodes (26): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+18 more)

### Community 13 - "go_pkg_os"
Cohesion: 0.22
Nodes (8): newProgress(), clearScreen(), enableVirtualTerminal(), enableVirtualTerminal(), go_pkg_os, go_pkg_syscall, go_pkg_unsafe, os.File

### Community 14 - "builtinList"
Cohesion: 0.22
Nodes (10): builtinList(), listRanges(), omitListProvenance(), releaseTag(), TestBuiltinList(), TestBuiltinListPicksTheNearestLine(), TestEmbeddedListsTable(), compareVersions() (+2 more)

### Community 15 - "completion_test.go"
Cohesion: 0.13
Nodes (26): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+18 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.14
Nodes (14): 11. What gets generated, 15. Notes and gotchas, 1.1 Shell completion, 1. Running solmq-conn-util, 2.1 The spec generator (no editor required), 2. Quick start, 3. Commands, 4. `examples` (+6 more)

### Community 18 - "libs/image_test.go"
Cohesion: 0.28
Nodes (12): loadImageLibs(), splitJarBasename(), TestEmbeddedListRange(), TestEmbeddedOmitListFullyParses(), TestLoadImageLibsBadPathIsError(), TestLoadImageLibsEmbeddedDefault(), TestLoadImageLibsSkipsCommentsAndBlankLines(), TestLoadImageLibsSkipsUnsplittableLineWithoutFailing() (+4 more)

### Community 19 - "dockergen_test.go"
Cohesion: 0.06
Nodes (59): Input, Instance, go_pkg_os_exec, composeEscape(), composeQuote(), yw, Render(), renderContentConfig() (+51 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.09
Nodes (56): Input, Instance, KV, envValueQuote(), PullSecret, StoreFile, yw, leaderMode() (+48 more)

### Community 23 - "gen_extra_test.go"
Cohesion: 0.14
Nodes (30): KubeOpts, Config(), issuesContain(), synthWorkflowFiles(), synthWorkflows(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder() (+22 more)

### Community 24 - "podmangen_test.go"
Cohesion: 0.22
Nodes (19): leaderLabels(), RenderQuadlet(), seconds(), systemdEnv(), fullInput(), Input, minimalInput(), syslogInput() (+11 more)

### Community 25 - "main.go"
Cohesion: 0.12
Nodes (44): resolveTarget(), runLogs(), absPath(), absResolver(), allowCommandFlag(), collectFlagsAndDirs(), confirmRemove(), contains() (+36 more)

### Community 26 - "progress"
Cohesion: 0.25
Nodes (5): progressTrim(), go_pkg_sync, sync.Mutex, progress, progressMode

### Community 27 - "withProgressClock"
Cohesion: 0.16
Nodes (17): newPipeReader(), progressErase(), TestNewProgressPicksOneRendering(), TestProgressElapsedShapes(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+9 more)

### Community 28 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 29 - "go_pkg_path_filepath"
Cohesion: 0.24
Nodes (9): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs(), go_pkg_embed, go_pkg_io_fs, go_pkg_path_filepath (+1 more)

### Community 30 - "tls_test.go"
Cohesion: 0.27
Nodes (10): MountPath(), SolaceProps(), StorePath(), TestMountPathSeparatorAgnostic(), TestSolacePropsRawPathWhenNotMounted(), TestSolacePropsSkipsSecretRefWhenStoreMissing(), TestSolacePropsStorePasswordIsStablePlaceholderNeverLiteral(), TestSolacePropsUseMountedBaseName() (+2 more)

### Community 31 - "statusreport.go"
Cohesion: 0.11
Nodes (27): heapValue(), parseHeap(), Banner(), banner(), Bytes(), canonicalRef(), Cores(), Instance (+19 more)

### Community 32 - "statusCollector"
Cohesion: 0.15
Nodes (13): sortInstances(), actStatus(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), watchStatus(), PodmanServiceName() (+5 more)

### Community 43 - "Side"
Cohesion: 0.15
Nodes (5): fixedLeaderNames(), Image, Cred, Side, placeholderSecretRef()

### Community 44 - "errExit"
Cohesion: 0.25
Nodes (23): actDocker(), actKubernetes(), actPodman(), emit(), envPairs(), errExit(), failFast(), genConfig() (+15 more)

### Community 45 - "solmq-conn-util -- Development Guide"
Cohesion: 0.29
Nodes (7): Build, Design notes, Release (CI), Shell completion, solmq-conn-util -- development guide, Testing, The spec generator (`solmq-conn-util-generator.html`)

### Community 46 - "libs_test.go"
Cohesion: 0.10
Nodes (56): go_pkg_testing_fstest, net/http.Header, Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic() (+48 more)

### Community 48 - "Kubernetes"
Cohesion: 0.13
Nodes (15): ownedNames(), TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), applyKubeDefaults(), Deployment, Kubernetes (+7 more)

### Community 49 - "Command details"
Cohesion: 0.14
Nodes (14): cli, Command details, deploy, download, examples, generate, help, logs (+6 more)

### Community 50 - "solmq-conn-util.bash"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "render/render.go"
Cohesion: 0.15
Nodes (33): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+25 more)

### Community 52 - "commands_doc_test.go"
Cohesion: 0.31
Nodes (8): TestAbbreviationDocInSync(), assertHelpWidth(), assertNoAliases(), normLF(), TestCommandsDocInSync(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), go_pkg_flag

### Community 54 - "spec_test.go"
Cohesion: 0.10
Nodes (40): go_pkg_slices, Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport() (+32 more)

### Community 55 - "instances.go"
Cohesion: 0.15
Nodes (31): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+23 more)

### Community 56 - "consolidate_extra_test.go"
Cohesion: 0.06
Nodes (68): leaderNameFn, Opts, secretFn, appendPassthrough(), applyStatusAccess(), binderOwner(), Build(), buildBundle() (+60 more)

### Community 57 - "abbreviation_doc_test.go"
Cohesion: 0.29
Nodes (10): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+2 more)

### Community 58 - "commands.go"
Cohesion: 0.22
Nodes (22): TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan(), invocation(), pad(), renderCommandsDoc(), tableCell() (+14 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (37): TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep() (+29 more)

### Community 60 - "write"
Cohesion: 0.08
Nodes (47): downloadEnvWithImage(), podmanEnv(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerSeamChildEnvCarriesCredentials() (+39 more)

### Community 61 - "Docker"
Cohesion: 0.14
Nodes (18): TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), Secrets, Service, applyDockerDefaults(), applyPodmanDefaults(), Docker (+10 more)

### Community 62 - "logs.go"
Cohesion: 0.11
Nodes (23): namedInstance(), actLogs(), checkLogsFlags(), logsInvocation(), readLog(), actShell(), attachShell(), checkShellFlags() (+15 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (33): ApplyStats(), ApplyTop(), EngineNamesByImage(), Instance, ObjectExists(), ParseApplication(), ParseInspect(), ParsePods() (+25 more)

### Community 64 - "libs/image.go"
Cohesion: 0.18
Nodes (14): go_pkg_bufio, connectorRelease(), imageNameTag(), imageSatisfies(), TestConnectorRelease(), TestImageNameTag(), TestImageSatisfies(), TestValidateImageVersionQualifiers() (+6 more)

### Community 65 - "instanceFromInspect"
Cohesion: 0.13
Nodes (23): encoding/json.RawMessage, time.Time, connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect() (+15 more)

### Community 67 - "uuid.go"
Cohesion: 0.31
Nodes (7): go_pkg_crypto_sha1, go_pkg_encoding_hex, DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5()

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
Cohesion: 0.23
Nodes (12): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone(), TestRemoveNamespaceProbeArgvIsOneQuery() (+4 more)

### Community 73 - "12. Status: the container, the connector, or both"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "14. cli: a shell inside the instance"
Cohesion: 0.33
Nodes (6): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance

### Community 75 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.25
Nodes (8): 6.1 Top-level, 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Header transforms (`transform-headers`), 6. Workflow file

### Community 79 - "resolvePullSecret"
Cohesion: 0.47
Nodes (6): dockerConfigJSON(), resolvePullSecret(), decodeAuths(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestResolvePullSecret()

### Community 80 - "resolveStores"
Cohesion: 0.40
Nodes (5): b64(), TestResolveStores(), resolveStores(), TestBaseName(), BaseName()

### Community 81 - "YAMLKind"
Cohesion: 0.18
Nodes (10): yaml.Node, YAMLKind(), decodeCreate(), yaml.Node, checkTransformHeaders(), yaml.Node, CredCreate, CredentialsSecret (+2 more)

### Community 82 - "targets_test.go"
Cohesion: 0.10
Nodes (28): ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors(), TestWorkflowsFromRawDefaultWhenAbsent(), TestWorkflowsFromRawDirOverride(), TestWorkflowsFromRawFilePatternOverride(), TestImagePullSecretCreateDefaultsFalse() (+20 more)

### Community 83 - "maven.go"
Cohesion: 0.14
Nodes (35): go_pkg_encoding_xml, encoding/xml.Name, acceptDependency(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM() (+27 more)

### Community 84 - "Libs"
Cohesion: 0.40
Nodes (5): Libs, LibsDownload, LibsPVC, NFS, PVCCreate

### Community 85 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "auto-complete"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.11
Nodes (26): go_pkg_io, go_pkg_net_http, go_pkg_net_url, net/http.Client, net/http.Request, net/http.Response, net/url.URL, defaultClient() (+18 more)

### Community 88 - "SafeToken"
Cohesion: 0.20
Nodes (8): checkSecurityUserRoles(), TestSafeActuatorUser(), TestSafeToken(), SafeActuatorUser(), safeLibsURL(), safeShellChars(), SafeToken(), allowCommandValue

### Community 89 - "8. Platform sections (`kubernetes:`, `docker:`, `podman:`)"
Cohesion: 0.40
Nodes (5): 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8. Platform sections (`kubernetes:`, `docker:`, `podman:`)

### Community 90 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

### Community 91 - "namespaceOccupants"
Cohesion: 0.67
Nodes (4): isClusterDefault(), isOurs(), namespaceOccupants(), nsItem

### Community 92 - "runner.go"
Cohesion: 0.08
Nodes (46): podmanRemove(), context.Context, io.Writer, os/exec.Cmd, PodmanSecretStoreName(), applyCmdEnv(), applyCmdInput(), Docker() (+38 more)

### Community 94 - "Expand"
Cohesion: 0.18
Nodes (19): reflect.Value, Expand(), expandMap(), expandString(), expandValue(), Workflow, lookupOf(), TestExpandBareDollarVarUntouched() (+11 more)

### Community 95 - "ParseCommand"
Cohesion: 0.17
Nodes (13): ParseCommand(), PodmanSecretCreate(), TestParseCommand(), TestParseCommandExtraAllowed(), TestPodmanSecretCreateRejectsUnsafeCommand(), TestPodmanSecretCreateRemovesThenCreatesValueOnStdin(), TestPodmanSecretCreateReportsCreateFailure(), TestPodmanSecretCreateSkipsCreateWhenRmFails() (+5 more)

### Community 96 - "runner_test.go"
Cohesion: 0.04
Nodes (79): go_pkg_runtime, os.FileMode, TestWriteMkdirError(), canIVerb(), ExecArgv(), InstallScript(), Preflight(), RunStatusScript() (+71 more)

### Community 98 - "testing.T"
Cohesion: 0.06
Nodes (73): captureStderr(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride(), TestDownloadForceFlagReachesInput(), TestDownloadIncludeProvidedFlagReachesInput(), TestDownloadJMSFlagIsGone(), TestDownloadMissingAndUnknownWordsRejected(), TestDownloadOmitLibFileFlagReachesInput() (+65 more)

### Community 99 - "kubernetes_test.go"
Cohesion: 0.31
Nodes (9): go_pkg_reflect, first(), parseKube(), second(), TestCreatedNames(), TestDerivedObjectNames(), TestSecretsCreateFollowsAnAlias(), TestSecretsCreateRejectsOtherShapes() (+1 more)

### Community 100 - "9. Secrets model"
Cohesion: 0.40
Nodes (5): 9.1 Declaring a credential, 9.2 Mount names, 9.3 How each platform delivers them, 9.4 Registry credentials (pulling the image), 9. Secrets model

### Community 101 - "misplacedEnvTransforms"
Cohesion: 0.57
Nodes (7): envTransforms(), appendTransformKeys(), documentMapping(), yaml.Node, mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 102 - "solmq-conn-util command reference"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "gen.go"
Cohesion: 0.37
Nodes (13): go_pkg_bytes, go_pkg_crypto_rand, go_pkg_encoding_base64, go_pkg_encoding_json, go_pkg_fmt, go_pkg_gopkg_in_yaml_v3, go_pkg_regexp, go_pkg_sort (+5 more)

### Community 105 - "status"
Cohesion: 0.50
Nodes (4): `solmq-conn-util status all [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status application [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, `solmq-conn-util status container [--details] [--watch] [--verbose] [--all] [--output table|json] [--install] [--platform kubernetes|docker|podman] [-e env.yaml] [--pod name] [--container name] [--namespace ns] [--management-port port] [--user name] [--command name] [--allow-command name]`, status

### Community 106 - "htmlgolden_test.go"
Cohesion: 0.15
Nodes (25): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+17 more)

### Community 107 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

## Knowledge Gaps
- **131 isolated node(s):** `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem`, `Defaults`, `kubeDeployment` (+126 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 219 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **38 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Kubernetes` connect `Kubernetes` to `validate.go`, `validate_extra_test.go`, `kubernetes_test.go`, `Defaults`, `gen.go`, `errExit`, `Libs`, `deploy_test.go`, `instances.go`, `Docker`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Why does `Side` connect `Side` to `validate.go`, `validate_extra_test.go`, `Defaults`, `gen.go`, `consolidate_extra_test.go`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 135 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 135 INFERRED edges - model-reasoned connections that need verification._
- **Are the 64 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 64 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solmq-conn-util.bash script`, `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `downloadItem` to the rest of the system?**
  _131 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `GenerateDocker` be split into smaller, more focused modules?**
  _Cohesion score 0.14619883040935672 - nodes in this community are weakly interconnected._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.07106446776611694 - nodes in this community are weakly interconnected._