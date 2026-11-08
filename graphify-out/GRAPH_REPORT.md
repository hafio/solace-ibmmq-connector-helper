# Graph Report - solace-ibmmq-connector-helper  (2026-11-08)

## Corpus Check
- 145 files · ~458,565 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 9 file(s) not represented in the graph (top: (none) 4, .list 3, .graphify-bak 1)

## Summary
- 2877 nodes · 12379 edges · 122 communities (102 shown, 20 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 1859 edges (avg confidence: 0.86)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `01b049c1`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Store
- .step
- validate_extra_test.go
- Input
- Config
- deploy_test.go
- checkExtraKeys
- solmq-conn-util test catalogue
- Bash Dev Script Tasks
- PowerShell Dev Script Tasks
- wf
- scan_test.go
- completion.go
- statusscript_test.go
- .run
- completion_test.go
- github.com/hafio/go-deploy-utility/runner.Runner
- crypto/x509.Certificate
- runAction
- jks.go
- Go Module Definition
- runner_test.go
- RenderQuadlet
- main_test.go
- issueWith
- .stop
- Defaults
- withPrompter
- createdStoreDockerSpec
- statusreport.go
- buildLeaderElection
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
- deref
- renderDeployment
- Command Reference Details
- Golden Bash Completion
- TestGoldenKubernetesNoSecrets
- validate.go
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
- CA
- Resolver
- runner_test.go
- deployDocker
- parse.go
- statusreport/render.go
- statusreport/render_test.go
- context.Context
- solmq-conn-util configuration reference
- Cred
- progress
- Writer
- support_test.go
- libs/image_test.go
- Model
- builtinList
- testing.T
- resolveStatusPassword
- maven.go
- getting-started.md
- rawDefaults
- solmq-conn-util operations reference
- libs.go
- ParseDefaults
- Secrets
- solaceSide
- runPodmanVerb
- solmq-conn-util troubleshooting
- TestSelectsStores
- Expand
- FormatScalar
- Port
- repeatableName
- SolaceProps
- tailFlag
- Input
- applyStatusAccess
- filenameFromEscapedPath
- Platform parity: solmq-conn-util and solace-util
- TestAliasDeclarationsFollowTheSharedRules
- RunAsID
- .add
- Getting started with solmq-conn-util
- yw
- Image
- XML
- renderWithPull
- allowCommandValue
- solmq-conn-util -- Solace IBM MQ Connector config generator and deployer
- targetMounts
- GeneratePodman
- validate_test.go
- Side
- Model
- solmq-conn-util abbreviations
- javaCtx

## God Nodes (most connected - your core abstractions)
1. `dispatch()` - 156 edges
2. `hasErr()` - 105 edges
3. `captureStderr()` - 90 edges
4. `bg()` - 78 edges
5. `write()` - 76 edges
6. `Build()` - 73 edges
7. `wfOK()` - 63 edges
8. `Download()` - 58 edges
9. `Defaults` - 51 edges
10. `imageOK()` - 51 edges

## Surprising Connections (you probably didn't know these)
- `9. Secrets` --references--> `statusPasswordFor()`  [INFERRED]
  docs/DESIGN.md → internal/gen/gen.go
- `1. Architecture` --references--> `render()`  [INFERRED]
  docs/DESIGN.md → internal/statusreport/render_test.go
- `10.3 How each platform delivers them` --references--> `existing()`  [INFERRED]
  docs/configuration.md → internal/tls/tls_test.go
- `TLS` --references--> `existing()`  [INFERRED]
  docs/troubleshooting.md → internal/tls/tls_test.go
- `Testing` --references--> `TestCompletionCoversModel()`  [INFERRED]
  docs/DEVELOPMENT.md → cmd/solmq-conn-util/completion_test.go

## Import Cycles
- None detected.

## Communities (122 total, 20 thin omitted)

### Community 0 - "Store"
Cohesion: 0.16
Nodes (8): Store, StoreSecret, StoreSources, TestStoreSecretParses(), TestStoreSecretNilSafe(), createdRole(), CertSources, StoreExisting

### Community 1 - ".step"
Cohesion: 0.18
Nodes (11): sortInstances(), confirmInstall(), instanceNames(), markMissing(), quoteAll(), MergeService(), ParseDeployment(), TestParseDeploymentAndService() (+3 more)

### Community 2 - "validate_extra_test.go"
Cohesion: 0.12
Nodes (64): TestCheckContainerCommandUnlistedBinaryRejected(), TestCheckKubeCommandDefaultKubectlUnvalidated(), TestCheckKubeCommandNowValidated(), TestContextAllowCommandsHonored(), TestDerivedNamesMustFitALabel(), baseKubeDeploy(), baseKubeService(), dockerOK() (+56 more)

### Community 3 - "Input"
Cohesion: 0.11
Nodes (20): Input, Instance, renderConfigMap(), renderContentConfig(), renderHealthcheck(), renderService(), seconds(), secretList() (+12 more)

### Community 4 - "Config"
Cohesion: 0.13
Nodes (24): shippedExamplesRequest(), TestEnvExtraKeysExampleValidWhenUncommented(), testResolver(), TestShippedExamplesGenerateConfig(), TestWorkflow0TransformExampleValidWhenUncommented(), Config(), issuesContain(), TestConfigCarriesExtraKeysEndToEnd() (+16 more)

### Community 5 - "deploy_test.go"
Cohesion: 0.17
Nodes (36): baseKube(), before(), fullFixtureInput(), lineDiff(), mustRender(), one(), TestCreatedSecretsTakeDerivedNames(), TestNamespaceManifestMatchesWhatRenderEmits() (+28 more)

### Community 6 - "checkExtraKeys"
Cohesion: 0.18
Nodes (18): dropManaged(), pageKeyList(), TestGeneratorPageKnownKeysInSync(), CanonicalKey(), CredentialKeys(), KnownKeys(), TestCanonicalKey(), TestToolManagedKeysAreSchemaKeys() (+10 more)

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
Cohesion: 0.19
Nodes (29): aliasPattern(), bashQuote(), completionFlags(), completionVerbs(), fishFlagSpec(), fishQuote(), fishSeenNames(), flagAliases() (+21 more)

### Community 13 - "statusscript_test.go"
Cohesion: 0.13
Nodes (27): breEscape(), Render(), runScriptFunction(), scriptFunction(), splitHealthBlock(), TestFilenameAndPathConstants(), TestHealthParseIsKeyOrderIndependent(), TestRenderAlignsWorkflowColumn() (+19 more)

### Community 14 - ".run"
Cohesion: 0.12
Nodes (30): TestAllowCommandFlagRejectedOnGenerateAndValidate(), TestExamplesDefaultDir(), TestExamplesWriteSkipForceThenGenerate(), TestExitCodeContract(), TestGenerateConfigEmitWriteError(), TestGenerateConfigStdoutAndFileMatch(), TestGenerateConfigTargetAliasResolves(), TestGenerateDockerToFile() (+22 more)

### Community 15 - "completion_test.go"
Cohesion: 0.16
Nodes (22): completionShells(), assertShellSafeName(), assertStrings(), between(), downloadJarTarget(), firstLineContaining(), hasWord(), linesContaining() (+14 more)

### Community 17 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.14
Nodes (42): checkImageCommand(), ensureNamespace(), firstLine(), probeOperatorSecrets(), 12. Runner seams, CheckProbe(), Docker(), DockerDeployDigest() (+34 more)

### Community 18 - "crypto/x509.Certificate"
Cohesion: 0.12
Nodes (23): day(), dedupe(), encryptedKey(), entry(), fingerprint(), guard(), matches(), matchesKey() (+15 more)

### Community 19 - "runAction"
Cohesion: 0.08
Nodes (75): resolveTarget(), verbUsage(), runLifecycle(), waitForRollout(), runLogs(), absPath(), absResolver(), actDocker() (+67 more)

### Community 21 - "jks.go"
Cohesion: 0.10
Nodes (24): 4.1 Top-level keys, 4.2 `solace:` side, 4.3 `mq:` side, 4.4 Reusing a connection (`conn-ref`), 4.5 Binders and durable names, 4.6 Passthrough keys, 4.7 Destination rules, 4.8 Transforms (+16 more)

### Community 23 - "runner_test.go"
Cohesion: 0.14
Nodes (38): currentUserName(), TestCurrentUserNameIsTheEffectiveUser(), existingPKCS12, existingPKCS12Shell, shellContentInfo, DurableName(), mustParseUUID(), TestDurableNameDeterministic() (+30 more)

### Community 24 - "RenderQuadlet"
Cohesion: 0.15
Nodes (18): Unit, leaderLabels(), RenderQuadlet(), seconds(), secretMount(), fullInput(), minimalInput(), syslogInput() (+10 more)

### Community 25 - "main_test.go"
Cohesion: 0.04
Nodes (141): TestLifecycleRolloutWait(), TestLifecycleVerbsRefuseBeforeActing(), dispatch(), assertSameNameSet(), canIAsked(), captureStderr(), captureStdout(), containsToken() (+133 more)

### Community 26 - "issueWith"
Cohesion: 0.14
Nodes (28): extraMQ(), extraSolace(), runExtra(), siteIssues(), TestConnRefSideRejectsExtraKeys(), TestExtraKeysCheckedOnConnectionsSessionAndDefaults(), TestExtraKeysCipherSuiteClashesWithCipher(), TestExtraKeysCredentialSpellingIsAnError() (+20 more)

### Community 27 - ".stop"
Cohesion: 0.14
Nodes (22): newPipeReader(), progressErase(), sleepAfter(), TestProgressOffAndNilWriteNothing(), TestProgressPauseHandsBackTheStream(), TestProgressSpinnerGoroutineAdvancesAndCountsSeconds(), TestProgressSpinnerNeverExceedsOneRow(), TestProgressSpinnerRewritesOneLineAndErases() (+14 more)

### Community 28 - "Defaults"
Cohesion: 0.18
Nodes (25): synthWorkflows(), Defaults, Workflow, checkSessionPosture(), checkValidateConflicts(), checkValidatePosture(), inlineSession(), isTCPSHost() (+17 more)

### Community 29 - "withPrompter"
Cohesion: 0.14
Nodes (34): gateKube(), kubeFixture(), nsItemJSON(), nsList(), removeResp(), runKube(), TestDeployManifestOmitsTheNamespace(), TestDeployNamespaceFlow() (+26 more)

### Community 30 - "createdStoreDockerSpec"
Cohesion: 0.32
Nodes (8): createdStoreDockerSpec(), envValue(), imageStartReply(), TestDeployDockerCarriesCreatedStores(), TestDeployDockerPullsAnImageItCannotInspect(), TestDeployDockerRefusesAnImageStartingOtherwise(), TestRemoveDockerNeedsNoCreatedStoreSources(), dockerDigestOf()

### Community 31 - "statusreport.go"
Cohesion: 0.07
Nodes (41): heapValue(), parseHeap(), withUsed(), stateCell(), Age(), Banner(), banner(), Bytes() (+33 more)

### Community 32 - "buildLeaderElection"
Cohesion: 0.20
Nodes (14): leaderNameFn, secretFn, buildBundle(), buildLeaderElection(), TestBuildLeaderElection(), TestBuildLeaderElectionSessionPassthroughCollision(), TestBuildSeparateStoresPerConnection(), testSecretRef() (+6 more)

### Community 43 - "podmanDeploy"
Cohesion: 0.08
Nodes (43): actPodman(), hostChecks(), podmanChangeSummary(), podmanDeploy(), podmanRemove(), previousStoreSecrets(), init(), TestPodmanChangeSummary() (+35 more)

### Community 44 - "storebuild_test.go"
Cohesion: 0.21
Nodes (31): DER(), KeyPKCS8(), NotCA(), PEM(), Build(), aliases(), chain(), derOf() (+23 more)

### Community 45 - "testcatalog_test.go"
Cohesion: 0.32
Nodes (11): countDocShape(), countTestFuncs(), isPackageHeading(), isTableSeparatorRow(), parseTestCatalogSnapshot(), TestCountDocShapeCountsCaseRowsAndPackageSections(), TestCountTestFuncsWalksTreeSkippingDataDirs(), TestIsTableSeparatorRow() (+3 more)

### Community 46 - "libs_test.go"
Cohesion: 0.11
Nodes (49): Download(), sha1Hex(), syslogFixtures(), syslogFixturesWithDependency(), TestDownloadBadOmitLibFilePathIsSystemic(), TestDownloadByteCapTripLeavesNoTempFile(), TestDownloadCommentsOnlyOmitListFileOmitsNothing(), TestDownloadContentLengthMismatchLeavesNoTempFile() (+41 more)

### Community 47 - "deref"
Cohesion: 0.17
Nodes (18): checkCreateBlock(), checkExistingBlock(), checkKeyEntry(), checkStoreBlock(), envTransforms(), isNull(), TestLeaderElectionEffectiveMode(), deref() (+10 more)

### Community 48 - "renderDeployment"
Cohesion: 0.11
Nodes (21): storeItem, storeSource, needsQuote(), QuoteScalar(), envValueQuote(), leaderMode(), quoteRes(), renderAffinity() (+13 more)

### Community 49 - "Command Reference Details"
Cohesion: 0.06
Nodes (32): All commands, auto-complete, cli, Command details, Command tree, deploy, download, examples (+24 more)

### Community 50 - "Golden Bash Completion"
Cohesion: 0.39
Nodes (8): solmq-conn-util.bash script, _solmq_conn_util(), _solmq_conn_util_flag_arg(), _solmq_conn_util_flags(), _solmq_conn_util_paths(), _solmq_conn_util_posarg(), _solmq_conn_util_sets(), _solmq_conn_util_targets()

### Community 51 - "TestGoldenKubernetesNoSecrets"
Cohesion: 0.19
Nodes (23): configMapDoc(), deploymentDoc(), dirReader(), envWithKube(), envWithKubeNoSyslog(), itoa(), lineDiff(), loadSpecs() (+15 more)

### Community 52 - "validate.go"
Cohesion: 0.08
Nodes (55): 3. Deploy execution, ParseRunAs(), parseRunAsID(), TestDefaultRunAsIsNonRootAndParses(), TestParseRunAs(), checkStoreDefinition(), badAffinityNode(), checkContainerStores() (+47 more)

### Community 54 - "Application"
Cohesion: 0.10
Nodes (37): Opts, Build(), containsSub(), TestApplyStatusAccessCarriesOperatorRoles(), TestApplyStatusAccessExposureIsFixed(), TestBuildCipherConflictWarning(), TestBuildCreatedStores(), TestBuildLeaderElectionSharesBinderSecretNames() (+29 more)

### Community 55 - "instanceSession"
Cohesion: 0.10
Nodes (39): anyIndex(), configuredInstanceName(), engineNames(), instanceCandidates(), instanceNoun(), isIndex(), kubeDiscovery(), namedInstance() (+31 more)

### Community 56 - "certfixture.go"
Cohesion: 0.14
Nodes (18): Key, Option, params, Trusted, AsCA(), Cert, JKS(), KeyEC() (+10 more)

### Community 57 - "renderCommandsDoc"
Cohesion: 0.18
Nodes (17): abbrevFlagByShort(), abbrevTable(), addTargetAbbreviations(), countCells(), modeledAbbreviations(), renderedAbbreviations(), TestAbbreviationDocCoversModel(), TestAbbreviationDocTableShape() (+9 more)

### Community 58 - "verbUsage"
Cohesion: 0.18
Nodes (18): assertHelpWidth(), assertNoAliases(), TestCommandsModelMatchesUsage(), TestInvocationTargetArgs(), TestVerbUsagePages(), flagEntries(), invocation(), pad() (+10 more)

### Community 59 - "maven_test.go"
Cohesion: 0.21
Nodes (38): TestDownloadMQOmitsAgainstTheDeployedLine(), TestDownloadSyslogEncoderFollowsConnectorLine(), TestDownloadSyslogWithNoReleaseOnTheLineIsSystemic(), groupPath(), metadataURL(), pomURL(), resolveClosure(), assertClosure() (+30 more)

### Community 60 - "write"
Cohesion: 0.13
Nodes (37): TestLifecycleVerbsActInPlace(), podmanEnv(), podmanEnvChained(), podmanEnvSudo(), podmanQuadletHome(), TestAllowCommandFlagRepeatableThreadsToRunner(), TestDeployDockerPreflightFailureStopsBeforeWrite(), TestDeployDockerRefusesAMissingStoreFile() (+29 more)

### Community 61 - "Kubernetes"
Cohesion: 0.09
Nodes (18): instanceCommand(), instanceNamespace(), loadInstanceEnv(), resolveInstanceSession(), KubeOpts, Written(), SecretProbe, secretProbes() (+10 more)

### Community 63 - "parse_test.go"
Cohesion: 0.09
Nodes (32): ApplyStats(), ApplyTop(), ObjectExists(), ParseApplication(), ParseInspect(), ParsePods(), SecretDataKeys(), splitKV() (+24 more)

### Community 64 - "CA"
Cohesion: 0.21
Nodes (22): CA(), Issue(), TrustStorePKCS12(), Validity(), TestEnvCreatedStoreExamplesValidWhenUncommented(), TestCheckExistingStore(), TestCheckExistingStoreFromOpenSSL(), TestResolvePodmanFilesOpensExistingStores() (+14 more)

### Community 65 - "Resolver"
Cohesion: 0.11
Nodes (35): TestMountChecksLabelEachPathByItsKey(), DockerExisting, DockerStore, PodmanFile, StoreBuild, StoreUse, CreatedStore, b64() (+27 more)

### Community 67 - "runner_test.go"
Cohesion: 0.04
Nodes (92): Contract, InstallScript(), KubernetesProbes(), PodmanDeploy(), PodmanSecretCreate(), Preflight(), bg(), contains() (+84 more)

### Community 68 - "deployDocker"
Cohesion: 0.15
Nodes (25): indexOfCall(), deployDocker(), dockerInspectWithLabel(), dockerSpec(), TestDeployDockerChangedAsksBeforeCompose(), TestDeployDockerDeclinedRunsNothing(), TestDeployDockerFailedRunningProbeIsAnError(), TestDeployDockerNonTTYFailsNamingTheFlag() (+17 more)

### Community 69 - "parse.go"
Cohesion: 0.13
Nodes (17): connectorIndex(), digestFrom(), engineComponents(), exitCode(), instanceFromInspect(), instanceFromPod(), kubeItems(), ParseImageDigest() (+9 more)

### Community 70 - "statusreport/render.go"
Cohesion: 0.12
Nodes (20): watchTitle(), View, groupOf(), JSON(), namespaceScope(), noteLine(), Render(), renderApplication() (+12 more)

### Community 71 - "statusreport/render_test.go"
Cohesion: 0.26
Nodes (14): render(), sample(), TestJSONEmptyRunIsAnEmptyList(), TestJSONIsTheSameModelTheTablesRender(), TestRenderApplicationViewBasicAndDetails(), TestRenderContainerViewAllNamespacesLeadsWithNamespace(), TestRenderContainerViewBasic(), TestRenderContainerViewDetails() (+6 more)

### Community 72 - "context.Context"
Cohesion: 0.11
Nodes (13): defaultAnswer(), scriptOut(), enableVirtualTerminal(), ExecArgv(), emit(), isCanI(), TestExecArgvPerPlatform(), TestExecArgvRefusesAnEmptyCommand() (+5 more)

### Community 73 - "solmq-conn-util configuration reference"
Cohesion: 0.10
Nodes (20): 10.1 Declaring a credential, 10.2 Mount names, 10.3 How each platform delivers them, 10.4 Registry credentials, 10. Secrets, 1. env.yaml layout, 2. Workflow discovery, 3. Variable expansion (+12 more)

### Community 74 - "Cred"
Cohesion: 0.23
Nodes (4): storePasswordValue(), Cred, TestSideUsernameSecretBothSystems(), placeholderSecretRef()

### Community 75 - "progress"
Cohesion: 0.22
Nodes (5): newProgress(), progressElapsed(), progressTrim(), progress, progressMode

### Community 76 - "Writer"
Cohesion: 0.33
Nodes (3): TestWriterLineIndent(), TestWriterRawPassthrough(), Writer

### Community 77 - "support_test.go"
Cohesion: 0.16
Nodes (10): TestAbbreviationDocInSync(), normLF(), TestCommandsDocInSync(), TestCompletionGoldenInSync(), mustNotice(), TestParityDocMatchesTheSharedBody(), TestSupportNoticeInGeneratedDocs(), TestSupportNoticeInGeneratorPage() (+2 more)

### Community 78 - "libs/image_test.go"
Cohesion: 0.12
Nodes (27): connectorRelease(), imageNameTag(), imageSatisfies(), loadImageLibs(), releaseTag(), splitJarBasename(), TestConnectorRelease(), TestEmbeddedListsMatchTheirConnectorLine() (+19 more)

### Community 79 - "Model"
Cohesion: 0.11
Nodes (20): acc, Binder, JMSBinding, MQBinder, Session, SolaceBinder, SolaceBinding, appendPassthrough() (+12 more)

### Community 80 - "builtinList"
Cohesion: 0.10
Nodes (24): 9.1 Keys common to every platform, 9.2 kubernetes, 9.3 docker, 9.4 podman, 9.5 Run-as and hardening, 9. Platform sections, 10. Status, logs and cli, 11. download jar (+16 more)

### Community 81 - "testing.T"
Cohesion: 0.04
Nodes (91): Render(), minimalDockerInput(), TestAppYAMLSecretPlaceholdersAreNotInterpolated(), TestContentEscapesDollarsForCompose(), TestContentIndentationAndBlankLines(), TestEnvironmentBranches(), TestHealthcheckRunsTheStatusScript(), TestJavaOptionsEnvironment() (+83 more)

### Community 82 - "resolveStatusPassword"
Cohesion: 0.19
Nodes (14): SecretRef, appendField(), deriveStatusPassword(), derivedPW(), kvRefs(), TestResolveStatusPasswordEmptyEnvFallsBackToRand(), TestResolveStatusPasswordEnvOverride(), TestResolveStatusPasswordFixedRand() (+6 more)

### Community 83 - "maven.go"
Cohesion: 0.13
Nodes (35): acceptDependency(), compareVersions(), compareVersionSegment(), coordKey(), extractProperties(), fetchMetadataXML(), fetchPOM(), fetchXML() (+27 more)

### Community 84 - "getting-started.md"
Cohesion: 0.31
Nodes (4): Build, Release checklist, Release (CI), solmq-conn-util -- development guide

### Community 85 - "rawDefaults"
Cohesion: 0.14
Nodes (17): The spec generator (`solmq-conn-util-generator.html`), defaultsFromRaw(), Security, applyDest(), nodePtr(), presentBlock(), TestUserSecretLiteralAndEnv(), rawDefaults (+9 more)

### Community 86 - "solmq-conn-util operations reference"
Cohesion: 0.10
Nodes (20): 10. download jar, 11. What gets generated, 12. Shell completion, 1. Commands, 2. validate, 3. generate, 4. deploy, 5. remove (+12 more)

### Community 87 - "libs.go"
Cohesion: 0.11
Nodes (19): defaultClient(), downloadOne(), downloadWithVerification(), fetchSHA1Sidecar(), Input, Report, isRedirectStatus(), parseSHA1Sidecar() (+11 more)

### Community 88 - "ParseDefaults"
Cohesion: 0.12
Nodes (17): buildRich(), lineDiff(), TestApplicationLeaderElectionSessionPlaintext(), TestApplicationRich(), TestApplicationRichExact(), ParseDefaults(), TestParseCreateBlock(), TestParseTLSIsStrict() (+9 more)

### Community 89 - "Secrets"
Cohesion: 0.24
Nodes (6): decodeCreate(), TestCredCreateRemovedKeys(), CredCreate, CredentialsSecret, StoreCreate, StoresSecret

### Community 90 - "solaceSide"
Cohesion: 0.28
Nodes (15): propsNode(), TestBuildCarriesEachWorkflowsTransformBlocks(), TestBuildSolaceBinderPassesValidateCertificateThrough(), binderOfKind(), extraWF(), TestBuildCarriesExtraKeysOntoBothBinders(), TestBuildDropsExtraKeysTheToolManages(), TestBuildMergesExtraKeysAcrossSidesOfOneBinder() (+7 more)

### Community 91 - "runPodmanVerb"
Cohesion: 0.34
Nodes (12): testHost(), useHost(), TestSessionRootless(), linuxHost(), runPodmanVerb(), TestDeployPodmanRootlessLinger(), TestPodmanRootfulRule(), TestPodmanRunAsAgainstTheSubIDRanges() (+4 more)

### Community 92 - "solmq-conn-util troubleshooting"
Cohesion: 0.25
Nodes (8): Contents, Passthrough keys, Secrets and credentials, solmq-conn-util troubleshooting, Status, logs and cli, TLS, Upgrading from an earlier release, Workflows and numbering

### Community 93 - "TestSelectsStores"
Cohesion: 0.33
Nodes (7): kubeNoSecrets(), TestCredentialsFoundOutsideAWorkflowSide(), TestCredentialsNotWiredIsAnError(), TestCredentialsWiredNoError(), TestNoCredentialsNoError(), TestSelectsStores(), vSolacePlainTCP()

### Community 94 - "Expand"
Cohesion: 0.17
Nodes (19): Expand(), expandMap(), expandString(), expandValue(), lookupOf(), TestExpandBareDollarVarUntouched(), TestExpandBracedVar(), TestExpandCredentialFieldLeftAlone() (+11 more)

### Community 95 - "FormatScalar"
Cohesion: 0.24
Nodes (8): TestFormatScalarQuoting(), TestMergeProp(), TestNodeToProps(), TestMergePropNestedValues(), FormatScalar(), mergeProp(), nodeToProps(), sameNode()

### Community 96 - "Port"
Cohesion: 0.20
Nodes (8): renderService(), Service, LibsMount, Port, StoresMount, checkPortRange(), checkPorts(), containerTarget

### Community 98 - "SolaceProps"
Cohesion: 0.32
Nodes (12): SolaceProps(), defaults(), existing(), TestSolacePropsAlwaysWritePKCS12(), TestSolacePropsSkipsSecretRefWhenNothingSelected(), TestSolacePropsSkipsUnconfiguredStore(), TestSolacePropsStorePasswordIsPlaceholderNeverLiteral(), TestSolacePropsUseStoreMountPaths() (+4 more)

### Community 100 - "Input"
Cohesion: 0.22
Nodes (10): Input, KV, StoreFile, ManagementPort(), nameOnlySecret(), Render(), TestLibsPVNameIsNamespaced(), TestManagementPort() (+2 more)

### Community 101 - "applyStatusAccess"
Cohesion: 0.29
Nodes (7): applyStatusAccess(), TestApplyStatusAccessAppendsAfterExistingUsers(), TestApplyStatusAccessNoOperatorUsers(), securityUserPasswordName(), stableName(), stableToken(), TestStableTokenFolding()

### Community 102 - "filenameFromEscapedPath"
Cohesion: 0.33
Nodes (6): filenameFromEscapedPath(), TestFilenameFromEscapedPathAcceptsPlainName(), TestFilenameFromEscapedPathRejectsEscapedTraversal(), TestValidateFilenameShapeAccepts(), TestValidateFilenameShapeRejections(), validateFilenameShape()

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "TestAliasDeclarationsFollowTheSharedRules"
Cohesion: 0.60
Nodes (4): newTargetAbbrev(), newCommandAbbrev(), newPlatformAbbrev(), TestAliasDeclarationsFollowTheSharedRules()

### Community 105 - "RunAsID"
Cohesion: 0.11
Nodes (18): fileOwner(), fileOwner(), podmanReader(), rootlessIDs(), sharedIDs(), skipOffUnix(), TestHostReaderFollowsTheKernelsRule(), TestMissingStoreFiles() (+10 more)

### Community 106 - ".add"
Cohesion: 0.25
Nodes (14): Instance, JavaOptions, YAMLKind(), checkContentType(), checkMisplacedTransform(), checkPayloadBlock(), checkTransform(), checkTransformExpressions() (+6 more)

### Community 107 - "Getting started with solmq-conn-util"
Cohesion: 0.25
Nodes (8): 1. Install, 2. Write a spec, 3. Validate and preview, 4. Add a platform section, 5. Deploy, 6. Check, read logs, tear down, Getting started with solmq-conn-util, Where next

### Community 108 - "yw"
Cohesion: 0.29
Nodes (14): blockIndicator(), Node(), q(), renderBundles(), renderCloudStream(), renderConnector(), renderContainer(), renderLeaderElection() (+6 more)

### Community 109 - "Image"
Cohesion: 0.28
Nodes (9): PullSecret, dockerConfigJSON(), resolvePullSecret(), decodeAuths(), kubeEnvWithPull(), TestDockerConfigJSON(), TestDockerConfigJSONEscapesAwkwardValues(), TestGenerateKubernetesImagePull() (+1 more)

### Community 110 - "XML"
Cohesion: 0.40
Nodes (5): syslogXML(), TestBothConfigsReadTheSameThreeProperties(), TestXMLDefaultsToUDP(), TestXMLPicksTheAppenderTheProtocolNames(), XML()

### Community 111 - "renderWithPull"
Cohesion: 0.67
Nodes (3): renderWithPull(), TestImagePullSecretPayloadIsOpaqueToDeploy(), TestImagePullSecretStates()

### Community 114 - "solmq-conn-util -- Solace IBM MQ Connector config generator and deployer"
Cohesion: 0.50
Nodes (4): Commands, Documentation, Quick start, solmq-conn-util -- Solace IBM MQ Connector config generator and deployer

### Community 115 - "targetMounts"
Cohesion: 0.13
Nodes (18): mount, Mount, TestPodmanFileSecretNames(), TestTargetMounts(), hostMounts(), PodmanFileSecretNames(), podmanSecretRefs(), PodmanSecretStoreName() (+10 more)

### Community 116 - "GeneratePodman"
Cohesion: 0.09
Nodes (49): 7. Container posture, File, build(), synthTLSWorkflowFiles(), synthWorkflowFiles(), TestGenerateDockerBasics(), TestGenerateDockerCarriesStatusScript(), TestGenerateJavaOptionsReachEveryPlatform() (+41 more)

### Community 117 - "validate_test.go"
Cohesion: 0.15
Nodes (29): registryImage(), retiredNamesKube(), TestRetiredCreateNamesNeedAKey(), TestRetiredCreateNamesReportedOnlyByValidate(), TestClusterOnlyStoreNeedsAHostSource(), TestNoStoresWiringWarning(), TestRetiredStoresBlockTiers(), tlsSelectingWF() (+21 more)

### Community 118 - "Side"
Cohesion: 0.13
Nodes (20): fixedLeaderNames(), leaderSession(), TLSConfig, TestTLSConfigStoreLookup(), TestResolveCarriesTheConnectionsExtraKeys(), Side, TestConnRefSideMayTuneBinding(), TestResolveConnRef() (+12 more)

### Community 120 - "Model"
Cohesion: 0.17
Nodes (16): Binding, WorkflowEnable, built, binderNames(), binderOf(), eqStrs(), TestBinderDedupAcrossWorkflows(), TestConnRefDedupCollapsesToOneBinder() (+8 more)

### Community 121 - "solmq-conn-util abbreviations"
Cohesion: 0.33
Nodes (6): Flag abbreviations, Notes, Platform abbreviations, solmq-conn-util abbreviations, Target abbreviations, Verb abbreviations

### Community 126 - "javaCtx"
Cohesion: 0.67
Nodes (3): javaCtx(), TestJavaOptionsAcceptsJVMSyntax(), TestJavaOptionsRejectsWhatThePlatformsReadDifferently()

## Knowledge Gaps
- **11 isolated node(s):** `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate`, `internal/deploy`, `internal/gen` (+6 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 229 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `solmq-conn-util command reference` connect `Command Reference Details` to `.stop`, `getting-started.md`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 151 inferred relationships involving `dispatch()` (e.g. with `TestLifecycleRolloutWait()` and `TestLifecycleVerbsActInPlace()`) actually correct?**
  _`dispatch()` has 151 INFERRED edges - model-reasoned connections that need verification._
- **What connects `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`, `NO_COLOR`, `internal/consolidate` to the rest of the system?**
  _11 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `validate_extra_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.11971153846153847 - nodes in this community are weakly interconnected._
- **Why does `existing()` connect `SolaceProps` to `CA`, `Store`, `solmq-conn-util configuration reference`, `runner_test.go`, `solmq-conn-util troubleshooting`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 77 inferred relationships involving `hasErr()` (e.g. with `TestCheckContainerCommandUnlistedBinaryRejected()` and `TestCheckKubeCommandDefaultKubectlUnvalidated()`) actually correct?**
  _`hasErr()` has 77 INFERRED edges - model-reasoned connections that need verification._
- **Should `Input` be split into smaller, more focused modules?**
  _Cohesion score 0.10822510822510822 - nodes in this community are weakly interconnected._