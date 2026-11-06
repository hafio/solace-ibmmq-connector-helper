# Graph Report - solace-ibmmq-connector-helper  (2026-11-06)

## Corpus Check
- 142 files · ~429,393 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2732 nodes · 11566 edges · 120 communities (100 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1758 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fe31b03e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Store
- .step
- validate_extra_test.go
- dockergen_test.go
- Config
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- testing.T
- run
- completion_test.go
- withPrompter
- crypto/x509.Certificate
- runAction
- os.File
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- builtinList
- ParseDefaults
- statusreport.go
- TestGoldenKubernetesNoSecrets
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
- podmanDeploy
- storebuild_test.go
- testcatalog_test.go
- libs_test.go
- rawDefaults
- gen_runas_test.go
- Command Reference Details
- Golden Bash Completion
- Model
- .add
- Application
- instanceSession
- certfixture.go
- renderCommandsDoc
- verbUsage
- maven_test.go
- write
- Kubernetes
- sinceFlag
- parse_test.go
- decodeCreate
- Resolver
- runner_test.go
- TestGeneratorPageKnownKeysInSync
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- Runner
- solmq-conn-util configuration reference
- Side
- statusreport_test.go
- Writer
- support_test.go
- libs/image_test.go
- buildLeaderElection
- go_pkg_io_fs
- createdStoreDockerSpec
- certfixture.go
- maven.go
- getting-started.md
- Cred
- solmq-conn-util operations reference
- libs.go
- ParseWorkflow
- deref
- solaceSide
- misplacedEnvTransforms
- SolaceProps
- TestSelectsStores
- Expand
- FormatScalar
- solmq-conn-util design notes
- repeatableName
- dispatch
- tailFlag
- flattenBlock
- extraKeyNames
- CommandRules
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- solmq-conn-util troubleshooting
- javaCtx
- allowCommandValue
- GenerateDocker
- GenerateKubernetes
- validate_test.go
- binderNames
- solmq-conn-util abbreviations
- Security
- Workload

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 158 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 86 edges
4. `write()` - 80 edges
5. `Build()` - 73 edges
6. `bg()` - 63 edges
7. `wfOK()` - 63 edges
8. `Download()` - 58 edges
9. `imageOK()` - 51 edges
10. `Defaults` - 50 edges

## Surprising Connections (you probably didn't know these)
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestCompletionModelMetadataComplete()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go
- `Testing` --references--> `TestTestCatalogSnapshotInSync()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/testcatalog_test.go

## Import Cycles
- None detected.

## Communities (120 total, 20 thin omitted)

### Community 0 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.13
Nodes (15): sortInstances(), bg(), newProgress(), progressTrim(), actStatus(), confirmInstall(), instanceNames(), markMissing() (+7 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "dockergen_test.go"
Cohesion: 0.07
Nodes (40): Input, Instance, renderConfigMap(), CreatedStore, Render(), renderContentConfig(), renderHealthcheck(), renderService() (+32 more)

### Community 4 - "Config"
Cohesion: 0.10
Nodes (37): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain(), synthWorkflowFiles() (+29 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.07
Nodes (67): Input, KV, storeItem, storeSource, envValueQuote(), PullSecret, StoreFile, leaderMode() (+59 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.29
Nodes (11): CanonicalKey(), TestCanonicalKey(), canonicalMatch(), checkExtraKeys(), editDistance(), hasCipherKey(), isFalse(), isLiteral() (+3 more)

### Community 7 - "solmq-conn-util test catalogue"
Cohesion: 0.08
Nodes (25): cli, cmd/solmq-conn-util, Contents, internal/consolidate, internal/deploy, internal/dockergen, internal/examples, internal/gen (+17 more)

### Community 8 - "Bash Dev Script Tasks"
Cohesion: 0.17
Nodes (22): c(), expand(), finish(), host_arch(), host_os(), log_begin(), NO_COLOR, run() (+14 more)

### Community 9 - "PowerShell Dev Script Tasks"
Cohesion: 0.20
Nodes (13): Get-Log(), Get-Now(), Invoke-Logged(), Task-build(), Task-cov(), Task-graphify(), Task-regen(), Task-scan() (+5 more)

### Community 10 - "wf"
Cohesion: 0.14
Nodes (39): postureConn(), postureIssues(), postureSolace(), TestValidatePostureBinderConflict(), TestValidatePostureConnectionWarnsOnceWithReach(), TestValidatePostureInlineSide(), TestValidatePostureSession(), connDefaults() (+31 more)

### Community 11 - "scan_test.go"
Cohesion: 0.18
Nodes (23): isYAML(), matchStar(), sameFile(), Scan(), bases(), TestIsYAML(), TestMatchStar(), TestScanEmptyPatternDefaultsToStar() (+15 more)

### Community 12 - "completion.go"
Cohesion: 0.20
Nodes (28): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+20 more)

### Community 13 - "testing.T"
Cohesion: 0.05
Nodes (81): TestWorkflowExamplesMatchGoldenSpecs(), TestGeneratorPageSelfTestNormalizesEmptyFindings(), TestContainerPathAndFileNameAgree(), TestStoreConfigured(), ParseEnv(), TestParseEnvEmpty(), TestParseEnvUnknownKeyIgnored(), TestParseEnvWrongScalarTypeErrors() (+73 more)

### Community 14 - "run"
Cohesion: 0.13
Nodes (20): run(), manyWorkflowsDir(), TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch() (+12 more)

### Community 15 - "completion_test.go"
Cohesion: 0.14
Nodes (24): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+16 more)

### Community 17 - "withPrompter"
Cohesion: 0.09
Nodes (31): nsItemJSON(), nsList(), TestCliEngineArgvShape(), TestCliKubernetesArgvShape(), TestCliOneShotAttachesStdinWhenSomethingIsPiped(), TestCliOneShotRunsTheCommandWithNoTerminal(), TestCliPreflightFailureOpensNothing(), TestCliPropagatesTheSessionExitStatus() (+23 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.08
Nodes (41): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+33 more)

### Community 19 - "runAction"
Cohesion: 0.09
Nodes (71): resolveTarget(), verbUsage(), runLifecycle(), runLogs(), absPath(), absResolver(), actDocker(), actKubernetes() (+63 more)

### Community 21 - "os.File"
Cohesion: 0.11
Nodes (15): scriptOut(), clearScreen(), enableVirtualTerminal(), LogsOpts, LogsArgv(), logsCommonFlags(), emit(), TestHelpersPassTheCallersContextToTheRunner() (+7 more)

### Community 23 - "runner_test.go"
Cohesion: 0.15
Nodes (43): isClusterDefault(), isOurs(), currentUserName(), subIDCount(), subIDsFromFiles(), subIDWarnings(), fakeHostFiles(), TestSubIDWarningsNeedARangeWideEnough() (+35 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.06
Nodes (65): assertSameNameSet(), downloadEnvWithImage(), keySet(), nameSet(), TestAbsPath(), TestConfiguredInstanceName(), TestDispatchHandlersMatchModel(), TestDownloadDirDefaultAndPositionalOverride() (+57 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.14
Nodes (16): newPipeReader(), progressErase(), TestProgressElapsedShapes(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+8 more)

### Community 28 - "Defaults"
Cohesion: 0.18
Nodes (25): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+17 more)

### Community 29 - "builtinList"
Cohesion: 0.19
Nodes (13): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, builtinList(), listRanges() (+5 more)

### Community 30 - "ParseDefaults"
Cohesion: 0.12
Nodes (17): buildRich(), lineDiff(), TestApplicationLeaderElectionSessionPlaintext(), TestApplicationRich(), TestApplicationRichExact(), ParseDefaults(), TestParseCreateBlock(), TestParseTLSIsStrict() (+9 more)

### Community 31 - "statusreport.go"
Cohesion: 0.10
Nodes (29): ApplyTop(), heapValue(), parseHeap(), TestApplyTop(), withUsed(), Banner(), banner(), Bytes() (+21 more)

### Community 32 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (23): The spec generator (`solmq-conn-util-generator.html`), configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff() (+15 more)

### Community 43 - "podmanDeploy"
Cohesion: 0.19
Nodes (15): podmanDeploy(), podmanRemove(), previousStoreSecrets(), TestPodmanFileSecretNames(), PodmanFileSecretNames(), QuadletScope, PodmanDeploy(), PodmanRemove() (+7 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (32): KeyPKCS8(), NotCA(), PEM(), Build(), guard(), aliases(), chain(), derOf() (+24 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.09
Nodes (57): Download(), filenameFromEscapedPath(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing() (+49 more)

### Community 47 - "rawDefaults"
Cohesion: 0.20
Nodes (13): defaultsFromRaw(), applyDest(), nodePtr(), presentBlock(), rawDefaults, rawLeader, rawLogging, rawManagement (+5 more)

### Community 48 - "gen_runas_test.go"
Cohesion: 0.35
Nodes (9): 7. Container posture, renderAll(), renderShapes(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPosture(), TestQuadletKeysAreWithinTheFloor(), TestRunAsOfRefusesWhatValidateRefuses(), TestSecretTargetsAreAbsolutePaths() (+1 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "Model"
Cohesion: 0.11
Nodes (24): acc, Binder, Binding, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding (+16 more)

### Community 52 - ".add"
Cohesion: 0.10
Nodes (46): checkStoreDefinition(), badAffinityNode(), checkContainerStores(), checkContainerTarget(), checkDocker(), checkDuplicateSources(), checkImage(), checkImagePull() (+38 more)

### Community 54 - "Application"
Cohesion: 0.10
Nodes (37): Opts, Build(), containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildCreatedStores(), TestBuildLeaderElectionSharesBinderSecretNames() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.12
Nodes (32): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceCommand(), instanceNamespace(), instanceNoun(), isIndex() (+24 more)

### Community 56 - "certfixture.go"
Cohesion: 0.14
Nodes (35): Key, Trusted, CA(), Cert, Issue(), JKS(), KeyEC(), KeyPKCS1() (+27 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.26
Nodes (11): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+3 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (23): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), flagsLine(), flagSpan() (+15 more)

### Community 59 - "maven_test.go"
Cohesion: 0.23
Nodes (36): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), metadataURL(), pomURL(), resolveClosure(), assertClosure(), dep(), metaXML() (+28 more)

### Community 60 - "write"
Cohesion: 0.07
Nodes (62): TestLifecycleVerbsActInPlace(), TestLifecycleVerbsRefuseBeforeActing(), captureStderr(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagBadValueExitsUsageError() (+54 more)

### Community 61 - "Kubernetes"
Cohesion: 0.08
Nodes (27): ownedNames(), Instance, Mount, SecretRef, JavaOptions, Kubernetes, Secrets, Service (+19 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (31): ApplyStats(), EngineNamesByImage(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+23 more)

### Community 64 - "decodeCreate"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (36): DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, TestResolvePodmanFiles(), TestTargetMounts(), Resolver (+28 more)

### Community 67 - "runner_test.go"
Cohesion: 0.05
Nodes (79): Contract, canIDenied(), canIProbe(), EngineInspectJSON(), ExecArgv(), InstallScript(), PodmanSecretCreate(), PodmanSecretRemove() (+71 more)

### Community 68 - "TestGeneratorPageKnownKeysInSync"
Cohesion: 0.38
Nodes (7): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CredentialKeys(), KnownKeys(), TestToolManagedKeysAreSchemaKeys(), ToolManagedKeys()

### Community 69 - "parse.go"
Cohesion: 0.13
Nodes (21): connectorIndex(), digestFrom(), engineComponents(), exitCode(), healthStatus(), instanceFromInspect(), instanceFromPod(), kubeItems() (+13 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.23
Nodes (16): View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication(), renderApplications() (+8 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "Runner"
Cohesion: 0.15
Nodes (33): actLogs(), checkLogsFlags(), logsInvocation(), readLog(), checkImageCommand(), confirmNamespaceRemoval(), removeNamespace(), firstLine() (+25 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Side"
Cohesion: 0.09
Nodes (22): fixedLeaderNames(), TLSConfig, TestLeaderElectionEffectiveMode(), TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+14 more)

### Community 75 - "statusreport_test.go"
Cohesion: 0.14
Nodes (12): ExitCodeText(), joinCells(), NewTable(), TestAge(), TestExitCodeText(), TestKVAlignsValuesOnTheWidestKey(), TestKVBuilderDropsEmptyValues(), TestResourceLineDropsWhatIsMissing() (+4 more)

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.13
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 79 - "buildLeaderElection"
Cohesion: 0.13
Nodes (19): leaderNameFn, secretFn, applyStatusAccess(), buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection() (+11 more)

### Community 80 - "go_pkg_io_fs"
Cohesion: 0.13
Nodes (7): fileOwner(), fileOwner(), enableVirtualTerminal(), mustWrite(), TestWriteCreatesSkipsForces(), TestWriteMkdirError(), Write()

### Community 81 - "createdStoreDockerSpec"
Cohesion: 0.38
Nodes (7): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources()

### Community 82 - "certfixture.go"
Cohesion: 0.10
Nodes (9): Option, params, AsCA(), DER(), RSAKey(), jksCert(), sniff(), TestSniff() (+1 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "getting-started.md"
Cohesion: 0.22
Nodes (8): Build, Release checklist, Release (CI), solmq-conn-util -- development guide, Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 85 - "Cred"
Cohesion: 0.13
Nodes (13): SecretRef, dockerConfigJSON(), TestResolveCredentials(), ResolveCredentials(), resolvePullSecret(), stableNames(), decodeAuths(), kubeEnvWithPull() (+5 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.10
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "ParseWorkflow"
Cohesion: 0.20
Nodes (11): TestParseWorkflowTypeErrorInsideASideStillNamesTheLine(), ParseWorkflow(), TestParseWorkflowAmbiguousSystemAndDest(), TestParseWorkflowConnRef(), TestParseWorkflowEnabledDefaultsTrue(), TestParseWorkflowSolaceAndMQ(), TestParseWorkflowSyntaxError(), TestMisplacedWorkflowTransforms() (+3 more)

### Community 89 - "deref"
Cohesion: 0.61
Nodes (6): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), isNull(), deref()

### Community 90 - "solaceSide"
Cohesion: 0.17
Nodes (22): propsNode(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildSolaceBinderPassesValidateCertificateThrough(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders() (+14 more)

### Community 91 - "misplacedEnvTransforms"
Cohesion: 0.57
Nodes (6): envTransforms(), appendTransformKeys(), documentMapping(), mappingChild(), misplacedEnvTransforms(), misplacedWorkflowTransforms()

### Community 92 - "SolaceProps"
Cohesion: 0.25
Nodes (14): TLS, SolaceProps(), defaults(), existing(), keys(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore() (+6 more)

### Community 93 - "TestSelectsStores"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.13
Nodes (15): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), needsQuote(), nodeToProps() (+7 more)

### Community 96 - "solmq-conn-util design notes"
Cohesion: 0.18
Nodes (11): 10. Status, logs and cli, 11. download jar, 1. Architecture, 2. CLI, 4. Kubernetes, 5. Docker, 6. Podman, 8. TLS (+3 more)

### Community 98 - "dispatch"
Cohesion: 0.09
Nodes (46): dispatch(), captureStdout(), containsToken(), TestAutoCompleteDispatchPrintsScript(), TestCliIndexSelectsFromTheSortedList(), TestCliPickerCarriesTheCommandBehindTheSeparator(), TestLogsDockerArgvShape(), TestLogsFollowReadsTheOneInstance() (+38 more)

### Community 100 - "flattenBlock"
Cohesion: 0.52
Nodes (4): extraKeys(), flattenBlock(), isMergeKey(), keyValue

### Community 101 - "extraKeyNames"
Cohesion: 0.43
Nodes (6): extraKeyNames(), extraValue(), TestMisplacedEnvTransformsCoversLeaderSession(), TestParseDefaultsCapturesExtraKeys(), TestParseWorkflowCapturesExtraKeys(), TestParseWorkflowExtraKeysFollowAliasesAndMergeKeys()

### Community 102 - "CommandRules"
Cohesion: 0.33
Nodes (5): 3. Deploy execution, CommandRules(), TestCommandGuardAccepts(), TestCommandGuardMessages(), TestCommandRulesPerPlatform()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.11
Nodes (26): dockerHostChecks(), missingStoreFiles(), mountChecks(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule() (+18 more)

### Community 106 - ".add"
Cohesion: 0.31
Nodes (12): YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions(), checkTransformHeaders(), checkTransformItem() (+4 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Platforms, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, Upgrading from an earlier release, Workflows and numbering

### Community 110 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

### Community 115 - "GenerateDocker"
Cohesion: 0.16
Nodes (14): mount, Mount, TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand(), TestResolveStatusPasswordRandError() (+6 more)

### Community 116 - "GenerateKubernetes"
Cohesion: 0.10
Nodes (33): File, KubeOpts, b64(), build(), synthTLSWorkflowFiles(), TestConfigRejectsSecretNameConflict(), TestGenerateKubernetesCarriesStatusScript(), TestGenerateKubernetesRefusesAnInjectedQuantity() (+25 more)

### Community 117 - "validate_test.go"
Cohesion: 0.15
Nodes (29): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF() (+21 more)

### Community 120 - "binderNames"
Cohesion: 0.39
Nodes (8): binderNames(), binderOf(), eqStrs(), TestConnRefDedupCollapsesToOneBinder(), TestConnRefNamingAndClashSuffix(), TestMQToMQSingleBinder(), TestSolaceToSolaceSingleBinder(), TestBinderFieldsCarryStablePlaceholders()

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 122 - "Security"
Cohesion: 0.40
Nodes (4): Security, TestUserSecretLiteralAndEnv(), rawSecurity, User

### Community 124 - "Workload"
Cohesion: 0.67
Nodes (4): MergeService(), ParseDeployment(), TestParseDeploymentAndService(), Workload

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 231 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util command reference` connect `Command Reference Details` to `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Are the 153 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleVerbsActInPlace()` and `TestLifecycleVerbsRefuseBeforeActing()`) actually correct?**
  _`dispatch()` has 153 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `.step` be split into smaller, more focused modules?**
  _Cohesion score 0.12660028449502134 - nodes in this community are weakly interconnected._
- **Why does `existing()` connect `SolaceProps` to `certfixture.go`, `solmq-conn-util configuration reference`, `Store`, `runner_test.go`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11971153846153847 - nodes in this community are weakly interconnected._