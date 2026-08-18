# solmq-conn-util fish completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util completion fish`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   fish autoloads this path, so writing the file is the whole install --
#   no rc edit, and new shells pick it up:
#   mkdir -p ~/.config/fish/completions
#   solmq-conn-util completion fish > ~/.config/fish/completions/solmq-conn-util.fish
#   
#   This session only, without installing anything:
#   solmq-conn-util completion fish | source
#
# No bare filename completion: every position below opts back in explicitly.
complete -c solmq-conn-util -f

# Verbs, offered only while no verb has been given yet.
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'generate' -d 'Render application.yml, or the artifacts for the resolved platform, to stdout or a file'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'deploy' -d 'Generate for a platform, then apply it'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'delete' -d 'Tear down what deploy created for a platform'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'status' -d 'Ensure and run the status script, printing per-target leader-election and workflow state'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'version' -d 'Print the utility name, version, Go version and OS/arch'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'validate' -d 'Lint the whole env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'examples' -d 'Write a starter env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'completion' -d 'Print a shell completion script'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'help' -d 'Print the usage summary (also -h, --help)'
complete -c solmq-conn-util -n '__fish_use_subcommand' -s h -l help -d 'Print the usage summary'

# Targets, offered once their verb is seen and until one is chosen.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate; and not __fish_seen_subcommand_from config' -a 'config' -d 'Emit application.yml'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'bash' -d 'Print the bash completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'zsh' -d 'Print the zsh completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'fish' -d 'Print the fish completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'powershell' -d 'Print the PowerShell completion script'

# Positional arguments that are not targets.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples' -a '(__fish_complete_directories)' -d 'directory'

# Flags, scoped to the verbs that accept them. -r means the flag takes a value;
# -F completes a file for it. --allow-command takes a value but not a filename,
# so it gets -r without -F and offers nothing.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate' -l platform -r -d 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate' -s o -l out -r -F -d 'write output to a file (default: stdout)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy' -l platform -r -d 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete' -l platform -r -d 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l install -d 'install the status script on every target without prompting'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l platform -r -d 'the platform: kubernetes, docker, or podman (default: resolved from env.yaml, or an interactive menu -- see Command details)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l pod -r -d 'limit checks to this kubernetes pod name; repeatable (default: every running pod); no effect on docker/podman'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l container -r -d 'limit checks to this docker/podman container name; repeatable (default: every running container); no effect on kubernetes'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l namespace -r -d 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l management-port -r -d 'actuator management port to reach inside each target (default: the management port configured for the target)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l user -r -d 'actuator account the status script authenticates as (default solmq-status)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l command -r -d 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each target, instead of the command: in that section'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from validate' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples' -s f -l force -d 'overwrite existing files'
