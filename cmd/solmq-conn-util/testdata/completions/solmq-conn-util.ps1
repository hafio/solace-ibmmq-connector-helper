# solmq-conn-util PowerShell completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util completion PowerShell`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   Register-ArgumentCompleter below is per-session, so appending to the
#   profile is what makes it stick:
#   solmq-conn-util completion powershell >> $PROFILE
#   
#   This session only, without touching the profile:
#   solmq-conn-util completion powershell | Out-String | Invoke-Expression
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

    $verbs = @(
        @{ Name = 'generate'; Desc = 'Render application.yml, or the artifacts for the resolved platform, to stdout or a file' }
        @{ Name = 'deploy'; Desc = 'Generate for a platform, then apply it' }
        @{ Name = 'delete'; Desc = 'Tear down what deploy created for a platform' }
        @{ Name = 'status'; Desc = 'Ensure and run the status script, printing per-target leader-election and workflow state' }
        @{ Name = 'version'; Desc = 'Print the utility name, version, Go version and OS/arch' }
        @{ Name = 'validate'; Desc = 'Lint the whole env.yaml + workflows' }
        @{ Name = 'examples'; Desc = 'Write a starter env.yaml + workflows' }
        @{ Name = 'completion'; Desc = 'Print a shell completion script' }
        @{ Name = 'help'; Desc = 'Print the usage summary (also -h, --help)' }
        @{ Name = '-h'; Desc = 'Print the usage summary' }
        @{ Name = '--help'; Desc = 'Print the usage summary' }
    )

    $targets = @{}
    $targets['generate'] = @(@{ Name = 'config'; Desc = 'Emit application.yml' })
    $targets['completion'] = @(@{ Name = 'bash'; Desc = 'Print the bash completion script' }, @{ Name = 'zsh'; Desc = 'Print the zsh completion script' }, @{ Name = 'fish'; Desc = 'Print the fish completion script' }, @{ Name = 'powershell'; Desc = 'Print the PowerShell completion script' })

    $flags = @{}
    $flags['generate'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '-o'; Desc = 'write output to a file (default: stdout)' }, @{ Name = '--out'; Desc = 'write output to a file (default: stdout)' })
    $flags['deploy'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['delete'] = @(@{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['status'] = @(@{ Name = '--install'; Desc = 'install the status script on every target without prompting' }, @{ Name = '--platform'; Desc = 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)' }, @{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--pod'; Desc = 'limit checks to this kubernetes pod name; repeatable (default: every running pod); no effect on docker/podman' }, @{ Name = '--container'; Desc = 'limit checks to this docker/podman container name; repeatable (default: every running container); no effect on kubernetes' }, @{ Name = '--namespace'; Desc = 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman' }, @{ Name = '--management-port'; Desc = 'actuator management port to reach inside each target (default: the management port configured for the target)' }, @{ Name = '--user'; Desc = 'actuator account the status script authenticates as (default solmq-status)' }, @{ Name = '--command'; Desc = 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each target, instead of the command: in that section' }, @{ Name = '--allow-command'; Desc = 'approve an extra command binary beyond the command: allowlist; repeatable' })
    $flags['validate'] = @(@{ Name = '-e'; Desc = 'config file, relative or absolute path (default: env.yaml)' }, @{ Name = '--env'; Desc = 'config file, relative or absolute path (default: env.yaml)' })
    $flags['examples'] = @(@{ Name = '-f'; Desc = 'overwrite existing files' }, @{ Name = '--force'; Desc = 'overwrite existing files' })

    $posArg = @{}
    $posArg['examples'] = 'dir'

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

    # Walk what is typed, skipping flags and their values, to find the verb and
    # how many positional arguments already follow it.
    $verb = ''
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
            else { $positional++ }
        }
        $i++
    }

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
}
