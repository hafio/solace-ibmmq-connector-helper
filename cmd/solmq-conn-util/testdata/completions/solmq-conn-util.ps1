# solmq-conn-util PowerShell completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util auto-complete PowerShell`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   Register-ArgumentCompleter below is per-session, so appending to the
#   profile is what makes it stick:
#   solmq-conn-util auto-complete powershell >> $PROFILE
#   
#   This session only, without touching the profile:
#   solmq-conn-util auto-complete powershell | Out-String | Invoke-Expression
#
# Windows PowerShell 5.1 compatible: no &&/||, no ternary, no null-coalescing.
# Both command names are registered because the Windows binary is solmq-conn-util.exe.
Register-ArgumentCompleter -Native -CommandName @('solmq-conn-util', 'solmq-conn-util.exe') -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    # Value kind each flag consumes. Both the one- and two-dash spellings are
    # present because Go's flag package accepts either.
    $flagArg = @{}
    $flagArg['-e'] = 'file'
    $flagArg['--e'] = 'file'
    $flagArg['-env'] = 'file'
    $flagArg['--env'] = 'file'
    $flagArg['-o'] = 'file'
    $flagArg['--o'] = 'file'
    $flagArg['-out'] = 'file'
    $flagArg['--out'] = 'file'
    $flagArg['-allow-command'] = 'name'
    $flagArg['--allow-command'] = 'name'
    $flagArg['-platform'] = 'name'
    $flagArg['--platform'] = 'name'
    $flagArg['-url'] = 'name'
    $flagArg['--url'] = 'name'
    $flagArg['-version'] = 'name'
    $flagArg['--version'] = 'name'
    $flagArg['-omit-lib-file'] = 'file'
    $flagArg['--omit-lib-file'] = 'file'
    $flagArg['-output'] = 'name'
    $flagArg['--output'] = 'name'
    $flagArg['-pod'] = 'name'
    $flagArg['--pod'] = 'name'
    $flagArg['-container'] = 'name'
    $flagArg['--container'] = 'name'
    $flagArg['-namespace'] = 'name'
    $flagArg['--namespace'] = 'name'
    $flagArg['-management-port'] = 'name'
    $flagArg['--management-port'] = 'name'
    $flagArg['-user'] = 'name'
    $flagArg['--user'] = 'name'
    $flagArg['-command'] = 'name'
    $flagArg['--command'] = 'name'
    $flagArg['-tail'] = 'name'
    $flagArg['--tail'] = 'name'
    $flagArg['-since'] = 'name'
    $flagArg['--since'] = 'name'

    # Aliases are deliberately absent from $verbs -- the TAB menu for word 1
    # keeps showing only canonical verbs.
    $verbs = @(
        @{ Name = 'generate'; Desc = 'Render application.yml, or the deploy artifacts for the resolved platform' }
        @{ Name = 'deploy'; Desc = 'Generate for a platform, then apply it' }
        @{ Name = 'remove'; Desc = 'Tear down what deploy created for a platform' }
        @{ Name = 'status'; Desc = 'Report each instance: container (engine), application (connector), or all' }
        @{ Name = 'logs'; Desc = 'Print one instance log, where status says what but not why' }
        @{ Name = 'cli'; Desc = 'Open a shell inside one instance, or run one command in it' }
        @{ Name = 'version'; Desc = 'Print the utility name, version, Go version and OS/arch' }
        @{ Name = 'validate'; Desc = 'Lint the whole env.yaml + workflows' }
        @{ Name = 'examples'; Desc = 'Write a starter env.yaml + workflows' }
        @{ Name = 'download'; Desc = 'Download IBM MQ or syslog encoder jars and their dependencies' }
        @{ Name = 'auto-complete'; Desc = 'Print a shell completion script' }
        @{ Name = 'help'; Desc = 'Print this summary, or the help page of one verb' }
        @{ Name = '-h'; Desc = 'Print the usage summary' }
        @{ Name = '--help'; Desc = 'Print the usage summary' }
    )

    # Every alias resolves to its canonical verb below, once, so $targets/
    # $flags/$posArg stay keyed by canonical verb names only.
    $verbAlias = @{}
    $verbAlias['gen'] = 'generate'
    $verbAlias['dp'] = 'deploy'
    $verbAlias['rm'] = 'remove'
    $verbAlias['sts'] = 'status'
    $verbAlias['lg'] = 'logs'
    $verbAlias['ver'] = 'version'
    $verbAlias['vld'] = 'validate'
    $verbAlias['eg'] = 'examples'
    $verbAlias['dl'] = 'download'

    $targets = @{}
    $targets['generate'] = @(@{ Name = 'config'; Desc = 'Emit application.yml' })
    $targets['status'] = @(@{ Name = 'container'; Desc = 'Report what the engine knows: state, restarts, age and image per instance' }, @{ Name = 'application'; Desc = 'Report what the connector knows: leader-election state, health and workflows' }, @{ Name = 'all'; Desc = 'Report both halves: the container table, then the application block per instance' })
    $targets['download'] = @(@{ Name = 'jar'; Desc = 'Download a set of jars and their dependencies into a directory' })
    $targets['auto-complete'] = @(@{ Name = 'bash'; Desc = 'Print the bash completion script' }, @{ Name = 'zsh'; Desc = 'Print the zsh completion script' }, @{ Name = 'fish'; Desc = 'Print the fish completion script' }, @{ Name = 'powershell'; Desc = 'Print the PowerShell completion script' })

    # Target aliases, resolved to their canonical word before the lookups
    # below, exactly as $verbAlias is for verbs. Aliases are absent from
    # $targets above, so the TAB menu keeps showing one spelling per target.
    $targetAlias = @{}
    $targetAlias['cfg'] = 'config'
    $targetAlias['cnt'] = 'container'
    $targetAlias['app'] = 'application'

    # A target's own further words -- the third command level -- keyed by
    # verb then by target name; only download/jar has any today.
    $sets = @{}
    if (-not $sets.ContainsKey('download')) { $sets['download'] = @{} }
    $sets['download']['jar'] = @(@{ Name = 'mq'; Desc = 'Download the IBM MQ client jar and its dependencies' }, @{ Name = 'syslog'; Desc = 'Download the logstash syslog encoder jar and its dependencies' })

    $flags = @{}
    $flags['generate'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '-o'; Desc = 'write output to a file (default: stdout)' }, @{ Name = '--out'; Desc = 'write output to a file (default: stdout)' })
    $flags['deploy'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['remove'] = @(@{ Name = '--no-prompt'; Desc = 'tear down without asking anything -- what a script or CI job passes, since the prompts refuse to read a non-TTY rather than hang. It covers both questions: the teardown confirmation, and whether to remove a namespace that turned out to be empty. It cannot authorise more than that: a namespace holding anything this release does not own is never removed, with or without it' }, @{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['status'] = @(@{ Name = '-d'; Desc = 'add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap' }, @{ Name = '--details'; Desc = 'add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap' }, @{ Name = '-w'; Desc = 're-render the report every 5s until interrupted (Ctrl-C)' }, @{ Name = '--watch'; Desc = 're-render the report every 5s until interrupted (Ctrl-C)' }, @{ Name = '--all'; Desc = 'reach every connector instance found by image name (solace-pubsub-connector-ibmmq) instead of the ones env.yaml describes -- every namespace on kubernetes, every container on docker/podman; cannot be combined with --pod/--container' }, @{ Name = '--output'; Desc = 'output format: table (default) or json, one machine-readable document per run; json cannot be combined with --watch' }, @{ Name = '--install'; Desc = 'install the status script on every instance without prompting' }, @{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--pod'; Desc = 'the kubernetes pod to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running pod on status, and on logs/cli the matching instances are listed instead. No effect on docker/podman' }, @{ Name = '--container'; Desc = 'the docker/podman container to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running container on status, and on logs/cli the one the section in env.yaml names. No effect on kubernetes' }, @{ Name = '--namespace'; Desc = 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman' }, @{ Name = '--management-port'; Desc = 'actuator management port to reach inside each instance (default: the configured management port)' }, @{ Name = '--user'; Desc = 'actuator account the status script authenticates as (default solmq-status)' }, @{ Name = '--command'; Desc = 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each instance, instead of the command: in that section' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['logs'] = @(@{ Name = '--follow'; Desc = 'keep the log open and print new lines as they arrive, until interrupted (Ctrl-C); reads one instance, so it cannot be combined with --all or --previous' }, @{ Name = '--previous'; Desc = 'read the log of the previous container instead of the running one -- what a pod that is restarting printed before it died; kubernetes only, since neither docker nor podman keeps a prior run under the same name' }, @{ Name = '--tail'; Desc = 'read only the last N lines, or all for the whole log (default: all)' }, @{ Name = '--since'; Desc = 'read only lines newer than this duration, spelled as a Go duration (30s, 10m, 2h)' }, @{ Name = '--timestamps'; Desc = 'prefix every line with the time the platform recorded for it' }, @{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--pod'; Desc = 'the kubernetes pod to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running pod on status, and on logs/cli the matching instances are listed instead. No effect on docker/podman' }, @{ Name = '--container'; Desc = 'the docker/podman container to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running container on status, and on logs/cli the one the section in env.yaml names. No effect on kubernetes' }, @{ Name = '--namespace'; Desc = 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman' }, @{ Name = '--command'; Desc = 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each instance, instead of the command: in that section' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['cli'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see [Platform resolution](commands.md#platform-resolution))' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--pod'; Desc = 'the kubernetes pod to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running pod on status, and on logs/cli the matching instances are listed instead. No effect on docker/podman' }, @{ Name = '--container'; Desc = 'the docker/podman container to reach, by name or by index into the listed order (alphabetical, the order status prints); a name always wins over the index reading. Repeatable on status; on logs and cli, which reach one instance, it may be given once. Default: every running container on status, and on logs/cli the one the section in env.yaml names. No effect on kubernetes' }, @{ Name = '--namespace'; Desc = 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman' }, @{ Name = '--command'; Desc = 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each instance, instead of the command: in that section' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['validate'] = @(@{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' })
    $flags['examples'] = @(@{ Name = '-f'; Desc = 'overwrite existing files' }, @{ Name = '--force'; Desc = 'overwrite existing files' })
    $flags['download'] = @(@{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--url'; Desc = 'exact URL to download instead of Maven resolution; repeatable; when given, no resolution happens at all' }, @{ Name = '--version'; Desc = 'pin the seed release (the IBM MQ client jar, or the syslog encoder jar) instead of resolving latest stable; empty means latest stable' }, @{ Name = '--omit-lib-file'; Desc = 'a jar list that replaces (never merges with) the embedded default the omission rule compares against; an empty file omits nothing' }, @{ Name = '--include-provided'; Desc = 'download the whole closure even where the connector image already provides a jar, instead of omitting it' }, @{ Name = '-f'; Desc = 'overwrite existing files' }, @{ Name = '--force'; Desc = 'overwrite existing files' })

    $posArg = @{}
    $posArg['examples'] = 'dir'
    $posArg['download'] = 'dir'

    $emit = {
        param($items, $word)
        foreach ($it in $items) {
            if ($it.Name.StartsWith($word, [System.StringComparison]::OrdinalIgnoreCase)) {
                [System.Management.Automation.CompletionResult]::new(
                    $it.Name, $it.Name, 'ParameterValue', $it.Desc)
            }
        }
    }

    # Path completion, keeping whatever directory prefix the user already typed.
    $emitPaths = {
        param($word, $dirsOnly)
        $prefix = ''
        $leaf = $word
        $cut = $word.LastIndexOfAny([char[]]@('\', '/'))
        if ($cut -ge 0) {
            $prefix = $word.Substring(0, $cut + 1)
            $leaf = $word.Substring($cut + 1)
        }
        $dir = '.'
        if ($prefix -ne '') { $dir = $prefix }
        $items = Get-ChildItem -LiteralPath $dir -ErrorAction SilentlyContinue
        foreach ($it in $items) {
            if ($dirsOnly -and (-not $it.PSIsContainer)) { continue }
            if (-not $it.Name.StartsWith($leaf, [System.StringComparison]::OrdinalIgnoreCase)) { continue }
            $path = $prefix + $it.Name
            $text = $path
            if ($text.Contains(' ')) { $text = "'" + $text + "'" }
            $kind = 'ProviderItem'
            if ($it.PSIsContainer) { $kind = 'ProviderContainer' }
            [System.Management.Automation.CompletionResult]::new($text, $path, $kind, $path)
        }
    }

    # The words already typed, excluding the command and the word under the cursor.
    $words = @()
    $elements = $commandAst.CommandElements
    for ($i = 1; $i -lt $elements.Count; $i++) { $words += [string]$elements[$i].Extent.Text }
    if ($wordToComplete -ne '' -and $words.Count -gt 0) {
        if ($words[$words.Count - 1] -eq $wordToComplete) {
            if ($words.Count -eq 1) { $words = @() }
            else { $words = $words[0..($words.Count - 2)] }
        }
    }

    # A value for the flag just typed wins over everything else.
    $prev = ''
    if ($words.Count -gt 0) { $prev = $words[$words.Count - 1] }
    if ($flagArg.ContainsKey($prev)) {
        if ($flagArg[$prev] -eq 'file') { & $emitPaths $wordToComplete $false }
        return
    }

    # Walk what is typed, skipping flags and their values, to find the verb,
    # its first two positional arguments (target and, if the target itself
    # fans out into a further word, that word), and how many positional
    # arguments already follow the verb in total.
    $verb = ''
    $arg1 = ''
    $arg2 = ''
    $positional = 0
    $i = 0
    while ($i -lt $words.Count) {
        $w = $words[$i]
        if ($w.StartsWith('-')) {
            if (-not $w.Contains('=')) {
                if ($flagArg.ContainsKey($w)) { $i++ }
            }
        }
        else {
            if ($verb -eq '') { $verb = $w }
            elseif ($arg1 -eq '') { $arg1 = $w; $positional++ }
            elseif ($arg2 -eq '') { $arg2 = $w; $positional++ }
            else { $positional++ }
        }
        $i++
    }

    if ($verbAlias.ContainsKey($verb)) { $verb = $verbAlias[$verb] }
    if ($targetAlias.ContainsKey($arg1)) { $arg1 = $targetAlias[$arg1] }

    if ($verb -eq '') {
        & $emit $verbs $wordToComplete
        return
    }

    if ($wordToComplete.StartsWith('-')) {
        if ($flags.ContainsKey($verb)) { & $emit $flags[$verb] $wordToComplete }
        return
    }

    if ($positional -eq 0) {
        if ($targets.ContainsKey($verb)) {
            & $emit $targets[$verb] $wordToComplete
            return
        }
        if ($posArg.ContainsKey($verb)) {
            & $emitPaths $wordToComplete ($posArg[$verb] -eq 'dir')
            return
        }
    }
    elseif ($positional -eq 1) {
        if ($sets.ContainsKey($verb) -and $sets[$verb].ContainsKey($arg1)) {
            & $emit $sets[$verb][$arg1] $wordToComplete
            return
        }
    }
    elseif ($positional -eq 2) {
        if ($sets.ContainsKey($verb) -and $sets[$verb].ContainsKey($arg1)) {
            if ($posArg.ContainsKey($verb)) {
                & $emitPaths $wordToComplete ($posArg[$verb] -eq 'dir')
            }
        }
    }
}
