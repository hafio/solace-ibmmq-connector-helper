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
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'generate' -d 'Render artifacts for one target to stdout or a file'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'deploy' -d 'Generate for a platform, then apply it'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'delete' -d 'Tear down what deploy created for a platform'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'validate' -d 'Lint the whole env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'examples' -d 'Write a starter env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'completion' -d 'Print a shell completion script'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'help' -d 'Print the usage summary (also -h, --help)'
complete -c solmq-conn-util -n '__fish_use_subcommand' -s h -l help -d 'Print the usage summary'

# Targets, offered once their verb is seen and until one is chosen.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate; and not __fish_seen_subcommand_from config kubernetes docker podman' -a 'config' -d 'Emit application.yml'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate; and not __fish_seen_subcommand_from config kubernetes docker podman' -a 'kubernetes' -d 'Emit ConfigMap+Deployment+Service (+Secrets)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate; and not __fish_seen_subcommand_from config kubernetes docker podman' -a 'docker' -d 'Emit docker-compose.yml (application.yml inlined)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate; and not __fish_seen_subcommand_from config kubernetes docker podman' -a 'podman' -d 'Emit a podman run script or quadlet unit'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'kubernetes' -d 'kubectl/oc apply -f - (manifest on stdin)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'docker' -d 'docker compose up -d'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'podman' -d 'write the quadlet unit; systemctl start'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'kubernetes' -d 'kubectl/oc delete -f -'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'docker' -d 'docker compose down'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete; and not __fish_seen_subcommand_from kubernetes docker podman' -a 'podman' -d 'systemctl stop; remove the unit'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'bash' -d 'Print the bash completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'zsh' -d 'Print the zsh completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'fish' -d 'Print the fish completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from completion; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'powershell' -d 'Print the PowerShell completion script'

# Positional arguments that are not targets.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples' -a '(__fish_complete_directories)' -d 'directory'

# Flags, scoped to the verbs that accept them. -r means the flag takes a value;
# -F completes a file for it. --allow-command takes a value but not a filename,
# so it gets -r without -F and offers nothing.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate' -s o -l out -r -F -d 'write output to a file (default: stdout)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from delete' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from validate' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples' -s f -l force -d 'overwrite existing files'
