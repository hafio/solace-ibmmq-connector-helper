# solmq-conn-util bash completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util auto-complete bash`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   Add this to ~/.bashrc. It depends on nothing but bash itself:
#   source <(solmq-conn-util auto-complete bash)
#   
#   System-wide instead -- but only where the bash-completion package is
#   installed and sourced from the profile, since that is what reads the
#   directory. Without it the file is never loaded:
#   solmq-conn-util auto-complete bash > /etc/bash_completion.d/solmq-conn-util
#
# _solmq_conn_util_flag_arg <word> prints the value kind the flag consumes:
# 'file', 'name', or nothing. Boolean flags consume no value and so are
# absent, falling through to the catch-all. Both the one- and two-dash
# spellings are listed because Go's flag package accepts either.
_solmq_conn_util_flag_arg() {
  case "$1" in
    -e|--e|-env|--env) printf 'file' ;;
    -o|--o|-out|--out) printf 'file' ;;
    -allow-command|--allow-command) printf 'name' ;;
    -platform|--platform) printf 'name' ;;
    -url|--url) printf 'name' ;;
    -version|--version) printf 'name' ;;
    -omit-lib-file|--omit-lib-file) printf 'file' ;;
    -output|--output) printf 'name' ;;
    -pod|--pod) printf 'name' ;;
    -container|--container) printf 'name' ;;
    -namespace|--namespace) printf 'name' ;;
    -management-port|--management-port) printf 'name' ;;
    -user|--user) printf 'name' ;;
    -command|--command) printf 'name' ;;
    -tail|--tail) printf 'name' ;;
    -since|--since) printf 'name' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_targets <verb> prints the verb's target names, or nothing.
_solmq_conn_util_targets() {
  case "$1" in
    generate) printf 'config' ;;
    status) printf 'container application all' ;;
    download) printf 'jar' ;;
    auto-complete) printf 'bash zsh fish powershell' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_sets <verb> <target> prints the target's own further
# words -- the third command level -- or nothing when it has none.
_solmq_conn_util_sets() {
  case "$1" in
    download)
      case "$2" in
        jar) printf 'mq syslog' ;;
        *) printf '' ;;
      esac
      ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_flags <verb> prints the flag spellings valid under the verb.
_solmq_conn_util_flags() {
  case "$1" in
    generate) printf '--platform -e --env -o --out' ;;
    deploy) printf '--platform -e --env --allow-command' ;;
    remove) printf '--platform -e --env --allow-command' ;;
    status) printf '-d --details -w --watch --all --output --install --platform -e --env --pod --container --namespace --management-port --user --command --allow-command' ;;
    logs) printf '--follow --previous --tail --since --timestamps --all --platform -e --env --pod --container --namespace --command --allow-command' ;;
    validate) printf '-e --env' ;;
    examples) printf '-f --force' ;;
    download) printf '-e --env --url --version --omit-lib-file --include-provided -f --force' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_posarg <verb> prints the completion kind for the verb's own
# positional argument, or nothing when it takes none.
_solmq_conn_util_posarg() {
  case "$1" in
    examples) printf 'dir' ;;
    download) printf 'dir' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_paths <word> <kind> fills COMPREPLY with matching paths.
# compopt is guarded: bash 3.2, still the system bash on macOS, lacks it.
_solmq_conn_util_paths() {
  local IFS=$'\n'
  if [ "$2" = 'dir' ]; then
    COMPREPLY=( $(compgen -d -- "$1") )
  else
    COMPREPLY=( $(compgen -f -- "$1") )
  fi
  if type compopt >/dev/null 2>&1; then compopt -o filenames; fi
}

_solmq_conn_util() {
  local cur prev word kind targets sets verb='' arg1='' arg2=''
  local i positional=0

  cur="${COMP_WORDS[COMP_CWORD]}"
  prev=''
  if [ "$COMP_CWORD" -gt 0 ]; then prev="${COMP_WORDS[$((COMP_CWORD - 1))]}"; fi

  # A value for the flag just typed wins over everything else.
  kind="$(_solmq_conn_util_flag_arg "$prev")"
  case "$kind" in
    file) _solmq_conn_util_paths "$cur" file; return ;;
    name) COMPREPLY=(); return ;;
  esac

  # Walk what is already typed, skipping flags and their values, to find the
  # verb, its first two positional arguments (target and, if the target
  # itself fans out into a further word, that word), and how many
  # positional arguments already follow the verb in total.
  i=1
  while [ "$i" -lt "$COMP_CWORD" ]; do
    word="${COMP_WORDS[$i]}"
    case "$word" in
      -*=*) ;;   # --env=path carries its own value; nothing to skip
      -*)
        if [ -n "$(_solmq_conn_util_flag_arg "$word")" ]; then i=$((i + 1)); fi
        ;;
      *)
        if [ -z "$verb" ]; then
          verb="$word"
        elif [ -z "$arg1" ]; then
          arg1="$word"
          positional=$((positional + 1))
        elif [ -z "$arg2" ]; then
          arg2="$word"
          positional=$((positional + 1))
        else
          positional=$((positional + 1))
        fi
        ;;
    esac
    i=$((i + 1))
  done

  # Aliases resolve to their canonical verb here, once, so the lookup
  # blocks above stay keyed by canonical verb names only.
  case "$verb" in
    gen) verb="generate" ;;
    dp) verb="deploy" ;;
    rm) verb="remove" ;;
    sts) verb="status" ;;
    lg) verb="logs" ;;
    ver) verb="version" ;;
    vld) verb="validate" ;;
    eg) verb="examples" ;;
    dl) verb="download" ;;
  esac

  # Aliases are deliberately absent from this word list -- the TAB menu for
  # word 1 keeps showing only canonical verbs.
  if [ -z "$verb" ]; then
    COMPREPLY=( $(compgen -W 'generate deploy remove status logs version validate examples download auto-complete help -h --help' -- "$cur") )
    return
  fi

  case "$cur" in
    -*)
      COMPREPLY=( $(compgen -W "$(_solmq_conn_util_flags "$verb")" -- "$cur") )
      return
      ;;
  esac

  if [ "$positional" -eq 0 ]; then
    targets="$(_solmq_conn_util_targets "$verb")"
    if [ -n "$targets" ]; then
      COMPREPLY=( $(compgen -W "$targets" -- "$cur") )
      return
    fi
    kind="$(_solmq_conn_util_posarg "$verb")"
    if [ -n "$kind" ]; then _solmq_conn_util_paths "$cur" "$kind"; return; fi
  elif [ "$positional" -eq 1 ]; then
    sets="$(_solmq_conn_util_sets "$verb" "$arg1")"
    if [ -n "$sets" ]; then
      COMPREPLY=( $(compgen -W "$sets" -- "$cur") )
      return
    fi
  elif [ "$positional" -eq 2 ]; then
    if [ -n "$(_solmq_conn_util_sets "$verb" "$arg1")" ]; then
      kind="$(_solmq_conn_util_posarg "$verb")"
      if [ -n "$kind" ]; then _solmq_conn_util_paths "$cur" "$kind"; return; fi
    fi
  fi

  COMPREPLY=()
}

# .exe as well, for a Windows binary driven from git-bash.
complete -F _solmq_conn_util solmq-conn-util solmq-conn-util.exe
