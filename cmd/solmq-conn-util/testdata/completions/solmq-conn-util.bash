# solmq-conn-util bash completion -- GENERATED, do not edit by hand.
#
# Rendered from the command model in cmd/solmq-conn-util/commands.go by
# `solmq-conn-util completion bash`, so it matches the binary that printed it.
# Re-run that command after upgrading solmq-conn-util.
#
# Install:
#   Add this to ~/.bashrc. It depends on nothing but bash itself:
#   source <(solmq-conn-util completion bash)
#   
#   System-wide instead -- but only where the bash-completion package is
#   installed and sourced from the profile, since that is what reads the
#   directory. Without it the file is never loaded:
#   solmq-conn-util completion bash > /etc/bash_completion.d/solmq-conn-util
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
    -pod|--pod) printf 'name' ;;
    -container|--container) printf 'name' ;;
    -namespace|--namespace) printf 'name' ;;
    -management-port|--management-port) printf 'name' ;;
    -user|--user) printf 'name' ;;
    -command|--command) printf 'name' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_targets <verb> prints the verb's target names, or nothing.
_solmq_conn_util_targets() {
  case "$1" in
    generate) printf 'config' ;;
    completion) printf 'bash zsh fish powershell' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_flags <verb> prints the flag spellings valid under the verb.
_solmq_conn_util_flags() {
  case "$1" in
    generate) printf '--platform -e --env -o --out' ;;
    deploy) printf '--platform -e --env --allow-command' ;;
    delete) printf '--platform -e --env --allow-command' ;;
    status) printf '--install --platform -e --env --pod --container --namespace --management-port --user --command --allow-command' ;;
    validate) printf '-e --env' ;;
    examples) printf '-f --force' ;;
    *) printf '' ;;
  esac
}

# _solmq_conn_util_posarg <verb> prints the completion kind for the verb's own
# positional argument, or nothing when it takes none.
_solmq_conn_util_posarg() {
  case "$1" in
    examples) printf 'dir' ;;
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
  local cur prev word kind targets verb=''
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
  # verb and how many positional arguments already follow it.
  i=1
  while [ "$i" -lt "$COMP_CWORD" ]; do
    word="${COMP_WORDS[$i]}"
    case "$word" in
      -*=*) ;;   # --env=path carries its own value; nothing to skip
      -*)
        if [ -n "$(_solmq_conn_util_flag_arg "$word")" ]; then i=$((i + 1)); fi
        ;;
      *)
        if [ -z "$verb" ]; then verb="$word"; else positional=$((positional + 1)); fi
        ;;
    esac
    i=$((i + 1))
  done

  if [ -z "$verb" ]; then
    COMPREPLY=( $(compgen -W 'generate deploy delete status version validate examples completion help -h --help' -- "$cur") )
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
  fi

  COMPREPLY=()
}

# .exe as well, for a Windows binary driven from git-bash.
complete -F _solmq_conn_util solmq-conn-util solmq-conn-util.exe
