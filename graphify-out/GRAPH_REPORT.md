# Graph Report - solace-ibmmq-connector-helper  (2026-10-26)

## Corpus Check
- 132 files · ~423,601 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2618 nodes · 10802 edges · 129 communities (109 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1632 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `90c90201`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- main.go
- statusCollector
- validate_extra_test.go
- dockergen_test.go
- Validate
- Cred
- checkExtraKeys
- Internal Package Test Catalog
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- deref
- completion_test.go
- solmq-conn-util user guide
- runner_test.go
- runAction
- deploy_test.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .step
- Defaults
- jks.go
- storebuild.go
- statusreport.go
- Cmd
- Graphify Project Guidance
- Consolidate Package
- Deploy Package
- Gen Package
- Render Package
- Scan Package
- Spec Package
- TLS Package
- Validate Package
- Golden Application YAML
- Runner
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- Project Documentation Files
- Kubernetes
- Command Reference Details
- Golden Bash Completion
- Model
- Store
- Application
- instanceSession
- certfixture.go
- renderCommandsDoc
- verbUsage
- maven_test.go
- write
- Podman
- sinceFlag
- parse_test.go
- libs/image_test.go
- Resolver
- run
- Dispatch Model Consistency Tests
- instanceFromInspect
- statusreport/render.go
- statusreport/render_test.go
- removeNamespace
- Status Command User Guide
- Side
- resolvePullSecret
- Writer
- 13. Logs: the lines behind the state
- 6. Workflow file
- Build
- Design notes
- rawDefaults
- renderDeployment
- maven.go
- GenerateDocker
- Command Abbreviation Reference
- Auto Complete Command Reference
- libs.go
- go_pkg_io_fs
- builtinList
- solaceSide
- actStatus
- SolaceProps
- decodeCreate
- Expand
- FormatScalar
- TestGoldenKubernetesNoSecrets
- repeatableName
- dispatch
- tailFlag
- abbreviation_doc_test.go
- Env
- Command Reference Overview
- Project Overview And Quick Start
- cmd/solmq-conn-util
- RunAsID
- .add
- support_test.go
- yw
- 10. `download jar`
- TestCreatedStoreWarningsSurface
- TestCredentialsNotWiredWarning
- Runner
- gen_extra_test.go
- buildLeaderElection
- defsWithStores
- statusreport_test.go
- postureConn
- subIDWarnings
- TestGeneratorPageKnownKeysInSync
- filenameFromEscapedPath
- binderNames
- LogsArgv
- TestStoresVolumeMapsKeysToStoreNames
- renderWithPull
- Context
- GeneratePodman

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 148 edges
2. `hasErr()` - 100 edges
3. `captureStderr()` - 77 edges
4. `Build()` - 73 edges
5. `write()` - 71 edges
6. `wfOK()` - 60 edges
7. `Download()` - 58 edges
8. `wf()` - 50 edges
9. `imageOK()` - 48 edges
10. `Defaults` - 47 edges

## Surprising Connections (you probably didn't know these)
- `9.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/userguide.md → internal/tls/tls_test.go
- `Design notes` --references--> `resolveOneInstance()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/instances.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Build()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/consolidate/consolidate.go
- `The spec generator (`solmq-conn-util-generator.html`)` --references--> `Validate()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go
- `Design notes` --references--> `PodmanFileSecretNames()`  [INFERRED]
  docs/DEVELOPMENT.md → internal/gen/gen.go

## Import Cycles
- None detected.

## Communities (129 total, 20 thin omitted)

### Community 0 - "main.go"
Cohesion: 0.09
Nodes (17): absPath(), absResolver(), confirmRemove(), envPairs(), failAll(), fileReader(), main(), platformSpellings() (+9 more)

### Community 1 - "statusCollector"
Cohesion: 0.14
Nodes (16): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), KubernetesGetJSON(), TestKubernetesGetJSONArgv(), TestKubernetesGetJSONMissingObjectIsAnError() (+8 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.11
Nodes (57): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), baseKubeDeploy(), baseKubeService(), dockerOK(), podmanOK() (+49 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.08
Nodes (38): Input, Instance, composeEscape(), composeQuote(), flowList(), CreatedStore, Render(), renderContentConfig() (+30 more)

### Community 4 - "Validate"
Cohesion: 0.16
Nodes (17): File, shippedExamplesRequest(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), build() (+9 more)

### Community 5 - "Cred"
Cohesion: 0.12
Nodes (8): Security, Image, Cred, TestSideUsernameSecretBothSystems(), TestUserSecretLiteralAndEnv(), placeholderSecretRef(), rawSecurity, User

### Community 6 - "checkExtraKeys"
Cohesion: 0.29
Nodes (11): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+3 more)

### Community 7 - "Internal Package Test Catalog"
Cohesion: 0.10
Nodes (21): Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen, internal/libs, internal/logback (+13 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.13
Nodes (53): connDefaults(), TestBinderIdentityUsesTheCredentialPair(), TestCheckSideMQMissingFields(), TestCheckSideSolaceMissingAndBadScheme(), TestConnRefSolaceTopicSourceIsError(), TestConnRefStrictOnlyDestination(), TestConnRefUnknownAndSystemMismatch(), TestConnRefValidResolvesNoError() (+45 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.22
Nodes (25): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flattenTarget() (+17 more)

### Community 13 - "testing.T"
Cohesion: 0.04
Nodes (91): mustWrite(), TestWorkflowExamplesMatchGoldenSpecs(), TestWriteCreatesSkipsForces(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), TestStoreConfigured(), ParseEnv(), TestParseEnvEmpty() (+83 more)

### Community 14 - "deref"
Cohesion: 0.17
Nodes (17): checkCreateBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref(), extraKeys() (+9 more)

### Community 15 - "completion_test.go"
Cohesion: 0.15
Nodes (24): completionShells(), assertShellSafeName(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining(), renderedCompletions() (+16 more)

### Community 17 - "solmq-conn-util user guide"
Cohesion: 0.10
Nodes (21): 14.1 One instance per run, 14.2 The shell is `sh`, 14.3 The one-shot form, and when it is the only form, 14.4 Exit status, 14.5 Which container it enters, 14. cli: a shell inside the instance, 15. Notes and gotchas, 1.1 Shell completion (+13 more)

### Community 18 - "runner_test.go"
Cohesion: 0.06
Nodes (58): canIVerb(), ExecArgv(), InstallScript(), Preflight(), RunStatusScript(), ScriptInstalled(), attachFiles(), helperProcessArgv() (+50 more)

### Community 19 - "runAction"
Cohesion: 0.20
Nodes (28): resolveTarget(), verbUsage(), runLogs(), allowCommandFlag(), collectFlagsAndDirs(), contains(), downloadDeployedImage(), envFlag() (+20 more)

### Community 21 - "deploy_test.go"
Cohesion: 0.17
Nodes (33): NamespaceManifest(), Render(), baseKube(), before(), fullFixtureInput(), lineDiff(), one(), TestCreatedSecretsTakeDerivedNames() (+25 more)

### Community 23 - "runner_test.go"
Cohesion: 0.18
Nodes (35): isClusterDefault(), isOurs(), DurableName(), mustParseUUID(), TestDurableNameDeterministic(), TestDurableNameGolden(), uuidv5(), sortedKeys() (+27 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.13
Nodes (20): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), systemdEnv(), fullInput(), minimalInput() (+12 more)

### Community 25 - "main_test.go"
Cohesion: 0.05
Nodes (78): captureStderr(), downloadEnvWithImage(), podmanEnvSudo(), TestAbsPath(), TestAllowCommandFlagBadValueExitsUsageError(), TestAutoCompleteDispatchPrintsScript(), TestConfiguredInstanceName(), TestDownloadDirDefaultAndPositionalOverride() (+70 more)

### Community 26 - "issueWith"
Cohesion: 0.15
Nodes (27): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError(), TestExtraKeysEnvSuffixWarns() (+19 more)

### Community 27 - ".step"
Cohesion: 0.13
Nodes (18): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+10 more)

### Community 28 - "Defaults"
Cohesion: 0.15
Nodes (28): Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost(), leaderConnRef() (+20 more)

### Community 29 - "jks.go"
Cohesion: 0.25
Nodes (3): sniff(), TestSniff(), format

### Community 30 - "storebuild.go"
Cohesion: 0.09
Nodes (38): 5.1 Variable expansion (`${VAR}`), 6.2 `solace:` options, 6.3 `mq:` options, 6.4 Destinations, durable names, passthrough, jksCert(), joinAliases(), readJKS(), build() (+30 more)

### Community 31 - "statusreport.go"
Cohesion: 0.10
Nodes (29): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Banner(), banner(), Bytes() (+21 more)

### Community 32 - "Cmd"
Cohesion: 0.12
Nodes (16): applyCmdEnv(), applyCmdInput(), Cmd, Streamer, resolveArgv0(), runParsed(), fakeRunner, OS (+8 more)

### Community 43 - "Runner"
Cohesion: 0.28
Nodes (21): actDocker(), actKubernetes(), actPodman(), emit(), errExit(), failFast(), genConfig(), genDocker() (+13 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.23
Nodes (30): KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf(), hasWarning() (+22 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.23
Nodes (15): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot() (+7 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (51): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+43 more)

### Community 47 - "Project Documentation Files"
Cohesion: 0.44
Nodes (3): Build, Release (CI), solmq-conn-util -- development guide

### Community 48 - "Kubernetes"
Cohesion: 0.11
Nodes (18): ownedNames(), 11. What gets generated, TestKubernetesDeployApplyOnStdin(), TestKubernetesRejectsUnsafeCommand(), TestKubernetesRemoveUsesDeleteVerb(), TestKubernetesUnknownAction(), applyKubeDefaults(), Deployment (+10 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.11
Nodes (18): cli, Command details, deploy, download, examples, generate, help, logs (+10 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.17
Nodes (15): Binding, JMSBinding, Session, SolaceBinding, WorkflowEnable, built, scalarProps(), LeaderElectionModel (+7 more)

### Community 52 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSources, TestStoreSecretNilSafe(), checkStoreDefinition(), createdRole(), checkStoreFiles(), CertSources, StoreExisting

### Community 54 - "Application"
Cohesion: 0.09
Nodes (37): Application(), blockKeys(), buildRich(), lineDiff(), renderLeaderFixture(), TestApplicationBlockScalarPassthrough(), TestApplicationConfigImport(), TestApplicationEmitsExactlyTheToolManagedKeys() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.12
Nodes (33): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+25 more)

### Community 56 - "certfixture.go"
Cohesion: 0.16
Nodes (17): Key, Option, params, Trusted, AsCA(), DER(), Cert, Issue() (+9 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.22
Nodes (13): abbrevFlagByShort(), abbrevTable(), renderAbbreviationDoc(), flagsLine(), flagSpan(), renderCommandsDoc(), tableCell(), flagAliases() (+5 more)

### Community 58 - "verbUsage"
Cohesion: 0.17
Nodes (18): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+10 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (35): TestDownloadMQOmitsAgainstTheDeployedLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML(), mqFixtures() (+27 more)

### Community 60 - "write"
Cohesion: 0.11
Nodes (42): createdStoreDockerSpec(), envValue(), podmanEnv(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+34 more)

### Community 61 - "Podman"
Cohesion: 0.12
Nodes (19): renderService(), TestDockerRejectsUnsafeCommand(), TestDockerUnknownAction(), TestDockerUpAndDown(), workflowsFromRaw(), Secrets, applyDockerDefaults(), applyPodmanDefaults() (+11 more)

### Community 63 - "parse_test.go"
Cohesion: 0.12
Nodes (25): ApplyStats(), ParseApplication(), ParseInspect(), ParsePods(), splitKV(), keys(), TestApplyStats(), TestParseApplication() (+17 more)

### Community 64 - "libs/image_test.go"
Cohesion: 0.14
Nodes (26): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+18 more)

### Community 65 - "Resolver"
Cohesion: 0.13
Nodes (28): DockerStore, PodmanFile, StoreBuild, StoreUse, TestParseExpandsNonCredentialAndWarnsOnUnsetDefaultless(), TestResolvePodmanFiles(), TestResolveStores(), Resolver (+20 more)

### Community 67 - "run"
Cohesion: 0.14
Nodes (18): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+10 more)

### Community 68 - "Dispatch Model Consistency Tests"
Cohesion: 0.48
Nodes (7): targetSetNames(), assertSameNameSet(), keySet(), nameSet(), TestDispatchHandlersMatchModel(), TestDownloadSetMapMatchesModel(), TestPlatformMapsCoverThreeNames()

### Community 69 - "instanceFromInspect"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "removeNamespace"
Cohesion: 0.20
Nodes (14): nsItemJSON(), nsList(), TestNamespaceOccupantsAreSortedAndLabelled(), TestNamespaceOccupantsRejectsUnreadableOutput(), TestNamespaceOccupantsRules(), TestRemoveNamespaceEmptyPromptsSeparately(), TestRemoveNamespaceNonTTYFailsFastNamingTheFlag(), TestRemoveNamespaceOccupiedLeavesItAlone() (+6 more)

### Community 73 - "Status Command User Guide"
Cohesion: 0.17
Nodes (12): 12.10 The manual alternative, 12.11 Instances this tool did not deploy, 12.1 `status container` -- the engine's view, 12.2 `status application` -- the connector's view, 12.3 First run: installing the script, 12.4 `-d` / `--details`, 12.5 `--all`: find every instance by image, 12.6 `-w` / `--watch` (+4 more)

### Community 74 - "Side"
Cohesion: 0.13
Nodes (20): fixedLeaderNames(), leaderSession(), TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+12 more)

### Community 75 - "resolvePullSecret"
Cohesion: 0.28
Nodes (9): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+1 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "13. Logs: the lines behind the state"
Cohesion: 0.29
Nodes (7): 13.1 `--previous` -- why a restarting instance died, 13.2 `--follow` -- keeping one open, 13.3 How much to read, 13.4 Choosing the instance, 13.5 Output shape and exit code, 13.6 The manual alternative, 13. Logs: the lines behind the state

### Community 78 - "6. Workflow file"
Cohesion: 0.33
Nodes (6): 6.1 Top-level, 6.5 Event-driven guidance (errors and warnings), 6.6 Reusable connections (`conn-ref`), 6.7 Transforms (`transform`), 6. Workflow file, Moving from `transform-headers` to `transform`

### Community 79 - "Build"
Cohesion: 0.11
Nodes (23): Opts, appendPassthrough(), binderOwner(), Build(), containsSub(), TestAppendPassthroughCollision(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers() (+15 more)

### Community 80 - "Design notes"
Cohesion: 0.09
Nodes (26): Design notes, TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), EngineImageInspectJSON(), EngineInspectJSON(), EngineList(), EngineStats(), KubernetesListJSON() (+18 more)

### Community 81 - "rawDefaults"
Cohesion: 0.17
Nodes (15): defaultsFromRaw(), applyDest(), nodePtr(), presentBlock(), Syslog, Logging, rawDefaults, rawLeader (+7 more)

### Community 82 - "renderDeployment"
Cohesion: 0.11
Nodes (20): Input, Instance, KV, envValueQuote(), StoreFile, leaderMode(), ManagementPort(), quoteRes() (+12 more)

### Community 83 - "maven.go"
Cohesion: 0.11
Nodes (32): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchPOM(), groupPath(), highestStable() (+24 more)

### Community 84 - "GenerateDocker"
Cohesion: 0.15
Nodes (21): KubeOpts, mount, Mount, b64(), TestGenerateJavaOptionsReachEveryPlatform(), TestStoresSharingAFileNameMountApart(), TestTargetMounts(), TestToIssues() (+13 more)

### Community 85 - "Command Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 86 - "Auto Complete Command Reference"
Cohesion: 0.40
Nodes (5): auto-complete, `solmq-conn-util auto-complete bash`, `solmq-conn-util auto-complete fish`, `solmq-conn-util auto-complete powershell`, `solmq-conn-util auto-complete zsh`

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (25): reportDownload(), defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus() (+17 more)

### Community 88 - "go_pkg_io_fs"
Cohesion: 0.22
Nodes (3): fileOwner(), fileOwner(), enableVirtualTerminal()

### Community 89 - "builtinList"
Cohesion: 0.15
Nodes (17): 7.1 The reserved status account (`solmq-status`), 7.2 Created stores (`create:`), 7. Connector defaults (`env.yaml` top level), 8.0 Image, timezone and JVM options (shared by every platform), 8.1 kubernetes, 8.2 docker, 8.3 podman, 8.4 Run-as and container hardening (+9 more)

### Community 90 - "solaceSide"
Cohesion: 0.18
Nodes (21): Binder, MQBinder, SolaceBinder, propsNode(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildLeaderElectionSharesBinderSecretNames() (+13 more)

### Community 91 - "actStatus"
Cohesion: 0.12
Nodes (14): newProgress(), actStatus(), clearScreen(), watchStatus(), enableVirtualTerminal(), How the suite is built, TestWriteMkdirError(), Attacher (+6 more)

### Community 92 - "SolaceProps"
Cohesion: 0.28
Nodes (13): SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral() (+5 more)

### Community 93 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.18
Nodes (11): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), TestSameNode(), FormatScalar(), mergeProp(), needsQuote() (+3 more)

### Community 96 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (22): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+14 more)

### Community 98 - "dispatch"
Cohesion: 0.08
Nodes (53): dispatch(), captureStdout(), containsToken(), TestCliEngineArgvShape(), TestCliIndexSelectsFromTheSortedList(), TestCliKubernetesArgvShape(), TestCliNeedsAnAttachingRunner(), TestCliOneShotAttachesStdinWhenSomethingIsPiped() (+45 more)

### Community 100 - "abbreviation_doc_test.go"
Cohesion: 0.33
Nodes (8): addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape(), targetSuffix(), cliTarget

### Community 101 - "Env"
Cohesion: 0.29
Nodes (9): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), presentPlatforms(), removeTarget(), PodmanServiceName(), Env (+1 more)

### Community 102 - "Command Reference Overview"
Cohesion: 0.33
Nodes (6): All commands, Command tree, Exit codes, Flags, Platform resolution, solmq-conn-util command reference

### Community 103 - "Project Overview And Quick Start"
Cohesion: 0.40
Nodes (5): Commands, Documentation, Minimal working example, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 104 - "cmd/solmq-conn-util"
Cohesion: 0.50
Nodes (4): cli, cmd/solmq-conn-util, logs, remove / instance resolution

### Community 105 - "RunAsID"
Cohesion: 0.10
Nodes (33): podmanDeploy(), podmanRemove(), previousStoreSecrets(), dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanHostChecks(), podmanReader() (+25 more)

### Community 106 - ".add"
Cohesion: 0.24
Nodes (16): JavaOptions, YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders() (+8 more)

### Community 107 - "support_test.go"
Cohesion: 0.53
Nodes (5): assertNoticeAtTop(), supportNoticePlain(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage(), TestSupportNoticeInHandWrittenDocs()

### Community 108 - "yw"
Cohesion: 0.30
Nodes (13): blockIndicator(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection(), renderLogging() (+5 more)

### Community 109 - "10. `download jar`"
Cohesion: 0.22
Nodes (9): 10.1 The two sets, 10.2 Version resolution, 10.3 Image-aware omission, 10.4 The image jar list: built-in, `--omit-lib-file`, and `--include-provided`, 10.5 `logstash-logback-encoder` and Jackson: the encoder follows the connector line, 10.6 `--url` overrides all resolution, 10.7 Flags and defaults, 10.8 Integrity verification (sha1) (+1 more)

### Community 110 - "TestCreatedStoreWarningsSurface"
Cohesion: 0.23
Nodes (16): CA(), Validity(), countIssues(), createdCA(), createdEnv(), hasIssue(), storeFiles(), TestBuildStore() (+8 more)

### Community 113 - "TestCredentialsNotWiredWarning"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredWarning(), TestCredentialsWiredNoWarning(), TestNoCredentialsNoWarning(), TestSelectsStores(), vSolacePlainTCP()

### Community 114 - "Runner"
Cohesion: 0.08
Nodes (29): checkImageCommand(), Docker(), DockerImageCommand(), Runner, Kubernetes(), kubeVerb(), ParseCommand(), PodmanDeploy() (+21 more)

### Community 115 - "gen_extra_test.go"
Cohesion: 0.14
Nodes (32): Config(), issuesContain(), synthWorkflowFiles(), synthWorkflows(), TestConfigCarriesExtraKeysEndToEnd(), TestConfigCarriesSecurityUserRoles(), TestConfigNoSecretsLeak(), TestConfigNumbersWorkflowsInLsOrder() (+24 more)

### Community 116 - "buildLeaderElection"
Cohesion: 0.11
Nodes (23): acc, leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), displayName(), dropManaged() (+15 more)

### Community 117 - "defsWithStores"
Cohesion: 0.16
Nodes (14): registryImage(), retiredNamesKube(), TestDerivedNamesMustFitALabel(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestConnectionDefinitionValidation(), TestRemovedDefaultsKeysRejected(), TestSecurityUserRoles() (+6 more)

### Community 118 - "statusreport_test.go"
Cohesion: 0.14
Nodes (12): ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing() (+4 more)

### Community 119 - "postureConn"
Cohesion: 0.48
Nodes (7): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession()

### Community 120 - "subIDWarnings"
Cohesion: 0.33
Nodes (6): currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough()

### Community 121 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.47
Nodes (6): pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 122 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 123 - "binderNames"
Cohesion: 0.36
Nodes (9): binderNames(), binderOf(), eqStrs(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder() (+1 more)

### Community 124 - "LogsArgv"
Cohesion: 0.40
Nodes (6): LogsOpts, LogsArgv(), logsCommonFlags(), TestLogsArgvPerPlatform(), TestLogsArgvRefusesPreviousOffKubernetes(), TestLogsArgvUnknownPlatform()

### Community 125 - "TestStoresVolumeMapsKeysToStoreNames"
Cohesion: 0.50
Nodes (3): storeItem, storeItems(), TestStoresVolumeMapsKeysToStoreNames()

### Community 126 - "renderWithPull"
Cohesion: 0.67
Nodes (3): renderWithPull(), TestImagePullSecretPayloadIsOpaqueToDeploy(), TestImagePullSecretStates()

### Community 129 - "Context"
Cohesion: 0.13
Nodes (36): checkContainerStores(), checkContainerTarget(), CheckDeployCommand(), checkDocker(), checkImage(), checkImagePull(), checkJavaOptions(), checkKube() (+28 more)

### Community 130 - "GeneratePodman"
Cohesion: 0.15
Nodes (17): TestGeneratePodmanQuadlet(), TestPodmanFileSecretNames(), TestPodmanUnitStoreSecrets(), GeneratePodman(), podmanSecretRefs(), PodmanSecretStoreName(), PodmanStoreSecretName(), PodmanUnitStoreSecrets() (+9 more)

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 205 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util -- development guide` connect `Project Documentation Files` to `Design notes`, `TestGoldenKubernetesNoSecrets`, `testcatalog_test.go`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Are the 143 inferred relationships involving `dispatch()` (e.g. with `verbUsage()` and `TestAllowCommandFlagBadValueExitsUsageError()`) actually correct?**
  _`dispatch()` has 143 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `main.go` be split into smaller, more focused modules?**
  _Cohesion score 0.09401709401709402 - nodes in this community are weakly interconnected._
- **Why does `solmq-conn-util user guide` connect `solmq-conn-util user guide` to `Status Command User Guide`, `13. Logs: the lines behind the state`, `10. `download jar``, `Project Documentation Files`, `Kubernetes`, `6. Workflow file`, `builtinList`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 73 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 73 INFERRED edges - model-reasoned connections that need verification._
- **Should `statusCollector` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._