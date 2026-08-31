# solmq-conn-util fish completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util auto-complete fish`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   fish autoloads this path, so writing the file is the whole install --
#   no rc edit, and new shells pick it up:
#   mkdir -p ~/.config/fish/completions
#   solmq-conn-util auto-complete fish > ~/.config/fish/completions/solmq-conn-util.fish
#   
#   This session only, without installing anything:
#   solmq-conn-util auto-complete fish | source
#
# No bare filename completion: every position below opts back in explicitly.
complete -c solmq-conn-util -f

# Verbs, offered only while no verb has been given yet. Aliases are
# deliberately absent here -- the TAB menu for word 1 keeps showing only
# canonical verbs.
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'generate' -d 'Render application.yml, or the deploy artifacts for the resolved platform'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'deploy' -d 'Generate for a platform, then apply it'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'remove' -d 'Tear down what deploy created for a platform'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'status' -d 'Report each instance: container (engine), application (connector), or all'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'version' -d 'Print the utility name, version, Go version and OS/arch'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'validate' -d 'Lint the whole env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'examples' -d 'Write a starter env.yaml + workflows'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'download' -d 'Download IBM MQ or syslog encoder jars and their dependencies'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'auto-complete' -d 'Print a shell completion script'
complete -c solmq-conn-util -n '__fish_use_subcommand' -a 'help' -d 'Print this summary, or the help page of one command'
complete -c solmq-conn-util -n '__fish_use_subcommand' -s h -l help -d 'Print the usage summary'

# Targets, offered once their verb (or an alias of it) is seen and until one
# is chosen.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate gen; and not __fish_seen_subcommand_from config cfg' -a 'config' -d 'Emit application.yml'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts; and not __fish_seen_subcommand_from container cnt application app all' -a 'container' -d 'Report what the engine knows: state, restarts, age and image per instance'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts; and not __fish_seen_subcommand_from container cnt application app all' -a 'application' -d 'Report what the connector knows: leader-election state, health and workflows'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts; and not __fish_seen_subcommand_from container cnt application app all' -a 'all' -d 'Report both halves: the container table, then the application block per instance'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl; and not __fish_seen_subcommand_from jar' -a 'jar' -d 'Download a set of jars and their dependencies into a directory'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from auto-complete; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'bash' -d 'Print the bash completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from auto-complete; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'zsh' -d 'Print the zsh completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from auto-complete; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'fish' -d 'Print the fish completion script'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from auto-complete; and not __fish_seen_subcommand_from bash zsh fish powershell' -a 'powershell' -d 'Print the PowerShell completion script'

# Sets, the third command level: a target's own further words, offered
# once the verb and that specific target are both seen and until one is
# chosen. Fish has no position counter of its own -- __fish_seen_subcommand_from
# matches anywhere on the command line -- so nesting this a level deeper
# than Targets above is just one more clause, not a special case.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl; and __fish_seen_subcommand_from jar; and not __fish_seen_subcommand_from mq syslog' -a 'mq' -d 'Download the IBM MQ client jar and its dependencies'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl; and __fish_seen_subcommand_from jar; and not __fish_seen_subcommand_from mq syslog' -a 'syslog' -d 'Download the logstash syslog encoder jar and its dependencies'

# Positional arguments that are not targets or sets. A verb whose target
# itself fans out into sets only offers this once one of those sets has
# been seen too.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples eg' -a '(__fish_complete_directories)' -d 'directory'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl; and __fish_seen_subcommand_from jar; and __fish_seen_subcommand_from mq syslog' -a '(__fish_complete_directories)' -d 'directory'

# Flags, scoped to the verbs that accept them (recognizing any alias too).
# -r means the flag takes a value; -F completes a file for it.
# --allow-command takes a value but not a filename, so it gets -r without -F
# and offers nothing.
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate gen' -l platform -r -d 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate gen' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from generate gen' -s o -l out -r -F -d 'write output to a file (default: stdout)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy dp' -l platform -r -d 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy dp' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from deploy dp' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from remove rm' -l platform -r -d 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from remove rm' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from remove rm' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -s d -l details -d 'add the enrichment lines each view can report: worker node, CPU/memory use against allocation, image digest and referenced components; app version, java version, config path and heap'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -s w -l watch -d 're-render the report every 5s until interrupted (Ctrl-C)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l all -d 'report every connector instance found by image name (solace-pubsub-connector-ibmmq) instead of the ones env.yaml describes -- every namespace on kubernetes, every container on docker/podman; cannot be combined with --pod/--container'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l output -r -d 'output format: table (default) or json, one machine-readable document per run; json cannot be combined with --watch'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l install -d 'install the status script on every instance without prompting'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l platform -r -d 'the platform: kubernetes, docker, or podman (short: kube, dk, pm; default: resolved from env.yaml, or an interactive menu -- see Platform resolution)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l pod -r -d 'limit checks to this kubernetes pod name; repeatable (default: every running pod); no effect on docker/podman'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l container -r -d 'limit checks to this docker/podman container name; repeatable (default: every running container); no effect on kubernetes'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l namespace -r -d 'kubernetes namespace to query (default: the namespace of the deployment in env.yaml); no effect on docker/podman'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l management-port -r -d 'actuator management port to reach inside each instance (default: the configured management port)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l user -r -d 'actuator account the status script authenticates as (default solmq-status)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l command -r -d 'override the platform CLI binary (kubectl/oc, docker, or podman) used to reach each instance, instead of the command: in that section'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from status sts' -l allow-command -r -d 'approve an extra command binary beyond the command: allowlist; repeatable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from validate vld' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from examples eg' -s f -l force -d 'overwrite existing files'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -s e -l env -r -F -d 'config file, relative or absolute path (default: env.yaml)'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -l url -r -d 'exact URL to download instead of Maven resolution; repeatable; when given, no resolution happens at all'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -l version -r -d 'pin the seed release (the IBM MQ client jar, or the syslog encoder jar) instead of resolving latest stable; empty means latest stable'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -l omit-lib-file -r -F -d 'a jar list that replaces (never merges with) the embedded default the omission rule compares against; an empty file omits nothing'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -l include-provided -d 'download the whole closure even where the connector image already provides a jar, instead of omitting it'
complete -c solmq-conn-util -n '__fish_seen_subcommand_from download dl' -s f -l force -d 'overwrite existing files'
