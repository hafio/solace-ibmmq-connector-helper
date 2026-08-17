package main

import "strings"

// This file renders the shell completion scripts printed by `solmq-conn-util
// completion <shell>`. Every verb, target, flag and description in them comes
// from the cliVerbs/cliFlags model in commands.go, so a command added there
// reaches all four shells without a second edit -- and TestCompletionGoldenInSync
// gates the rendered output against a committed snapshot, so a model change the
// renderers do not handle shows up as a reviewable diff instead of silently
// completing nothing.
//
// The completion logic is the same in all four shells:
//
//	word 1                          the verb names, plus -h/--help
//	word 2 after a verb             that verb's targets, or a directory (PosArg)
//	a value after -e/--env, -o/--out    a filesystem path
//	a value after --allow-command   nothing: it is a free-form binary name, and
//	                                allowCommandValue rejects anything path-shaped
//	a word starting with -          that verb's flags
//
// Positions are counted only after skipping flag words and their values, because
// collectFlagsAndDirs accepts flags before, after, or between the positional
// arguments -- `generate -e env.yaml <TAB>` must still offer targets.

// completionItem is one completable word plus the description those shells that
// support one will show beside it.
type completionItem struct{ Name, Desc string }

// completionFlag is one cliFlag flattened for the renderers: the spellings to
// offer, every spelling to recognize when skipping a value, the value kind, and
// a plain-text description.
type completionFlag struct {
	Short, Long string   // documented spellings; equal when the flag has no short form
	Offer       []string // the distinct spellings to offer ("-e", "--env")
	Aliases     []string // every spelling flag.Parse accepts ("-e", "--e", "-env", "--env")
	Arg         string   // argNone/argFile/argName
	Desc        string
}

// completionVerb is one cliVerb flattened for the renderers.
type completionVerb struct {
	Name    string
	Desc    string
	PosArg  string
	Targets []completionItem
	Flags   []completionFlag
}

// ---- model flattening ---------------------------------------------------------

// completionFlags returns every modeled flag in declaration order, flattened.
func completionFlags() []completionFlag {
	out := make([]completionFlag, 0, len(cliFlags))
	for _, f := range cliFlags {
		out = append(out, completionFlag{
			Short: f.Short, Long: f.Long,
			Offer: flagOffered(f), Aliases: flagAliases(f),
			Arg: f.Arg, Desc: plainText(f.Meaning),
		})
	}
	return out
}

// completionVerbs returns every modeled verb in declaration order, flattened.
func completionVerbs() []completionVerb {
	byShort := make(map[string]cliFlag, len(cliFlags))
	for _, f := range cliFlags {
		byShort[f.Short] = f
	}
	out := make([]completionVerb, 0, len(cliVerbs))
	for _, v := range cliVerbs {
		cv := completionVerb{Name: v.Name, Desc: plainText(verbBlurb(v)), PosArg: v.PosArg}
		for _, tg := range v.Targets {
			cv.Targets = append(cv.Targets, completionItem{Name: tg.Name, Desc: plainText(tg.Summary)})
		}
		for _, sh := range v.Flags {
			f, ok := byShort[sh]
			if !ok { // an unmodeled flag key; TestCompletionModelMetadataComplete catches it
				continue
			}
			cv.Flags = append(cv.Flags, completionFlag{
				Short: f.Short, Long: f.Long,
				Offer: flagOffered(f), Aliases: flagAliases(f),
				Arg: f.Arg, Desc: plainText(f.Meaning),
			})
		}
		out = append(out, cv)
	}
	return out
}

// flagOffered returns the spellings a script suggests: the short/long pair, or
// just the one spelling when the flag has no short form (Short == Long).
func flagOffered(f cliFlag) []string {
	if f.Short == f.Long {
		return []string{f.Long}
	}
	return []string{f.Short, f.Long}
}

// flagAliases returns every spelling flag.Parse accepts for f. The flag package
// treats one and two leading dashes as the same flag, so -env and --env are
// interchangeable at the command line; a script must skip the value after any of
// them even though it only ever offers the documented pair.
func flagAliases(f cliFlag) []string {
	seen := map[string]bool{}
	var out []string
	add := func(name string) {
		for _, dash := range []string{"-", "--"} {
			s := dash + name
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	add(strings.TrimLeft(f.Short, "-"))
	add(strings.TrimLeft(f.Long, "-"))
	return out
}

// completionShells returns the modeled shell names, in the order cliVerbs
// declares them under the completion verb.
func completionShells() []string { return targetNames("completion") }

// ---- text helpers -------------------------------------------------------------

// plainText renders a model string for a shell tooltip: the markdown code-span
// backticks the model carries for docs/commands.md are dropped, control
// characters (a newline above all -- it would terminate the enclosing statement
// in every one of the four formats) collapse to spaces, and runs of whitespace
// fold to one. Quoting happens after this, per shell.
func plainText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '`':
		case r < 0x20 || r == 0x7f:
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// bashQuote single-quotes s for bash and zsh, where the single quote is the only
// character that cannot appear literally inside single quotes.
func bashQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// fishQuote single-quotes s for fish, which -- unlike bash -- also treats the
// backslash as an escape inside single quotes, so it is doubled first.
func fishQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "'", `\'`)
	return "'" + s + "'"
}

// psQuote single-quotes s for PowerShell, where an embedded single quote is
// written by doubling it. A single-quoted PowerShell string is literal, so `$`
// and the backtick escape character need no further handling.
func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// zshEntry builds one "value:description" entry for zsh's _describe, which
// splits each entry on the first unescaped colon -- so a colon in the value half
// is escaped, while the description half may carry colons freely.
func zshEntry(name, desc string) string {
	return bashQuote(strings.ReplaceAll(name, ":", `\:`) + ":" + desc)
}

// ---- shared header ------------------------------------------------------------

// scriptHeader is the comment block every generated script opens with, using the
// given line-comment marker. install is the shell's own install recipe.
func scriptHeader(shell, marker string, install []string) []string {
	l := []string{
		marker + " solmq-conn-util " + shell + " completion -- GENERATED, do not edit by hand.",
		marker + "",
		marker + " Rendered from the command model in cmd/solmq-conn-util/commands.go by",
		marker + " `solmq-conn-util completion " + shell + "`, so it matches the binary that printed it.",
		marker + " Re-run that command after upgrading solmq-conn-util.",
		marker + "",
		marker + " Install:",
	}
	for _, s := range install {
		l = append(l, marker+"   "+s)
	}
	return append(l, marker+"")
}

// ---- bash ---------------------------------------------------------------------

func renderBashCompletion() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	l = append(l, scriptHeader("bash", "#", []string{
		"Add this to ~/.bashrc. It depends on nothing but bash itself:",
		"source <(solmq-conn-util completion bash)",
		"",
		"System-wide instead -- but only where the bash-completion package is",
		"installed and sourced from the profile, since that is what reads the",
		"directory. Without it the file is never loaded:",
		"solmq-conn-util completion bash > /etc/bash_completion.d/solmq-conn-util",
	})...)

	add("# _solmq_conn_util_flag_arg <word> prints the value kind the flag consumes:")
	add("# 'file', 'name', or nothing. Boolean flags consume no value and so are")
	add("# absent, falling through to the catch-all. Both the one- and two-dash")
	add("# spellings are listed because Go's flag package accepts either.")
	add("_solmq_conn_util_flag_arg() {")
	add(`  case "$1" in`)
	for _, f := range completionFlags() {
		if f.Arg == argNone {
			continue
		}
		add("    " + strings.Join(f.Aliases, "|") + ") printf '" + f.Arg + "' ;;")
	}
	add("    *) printf '' ;;")
	add("  esac")
	add("}")
	add("")

	add("# _solmq_conn_util_targets <verb> prints the verb's target names, or nothing.")
	add("_solmq_conn_util_targets() {")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if len(v.Targets) == 0 {
			continue
		}
		add("    " + v.Name + ") printf " + bashQuote(itemNames(v.Targets)) + " ;;")
	}
	add("    *) printf '' ;;")
	add("  esac")
	add("}")
	add("")

	add("# _solmq_conn_util_flags <verb> prints the flag spellings valid under the verb.")
	add("_solmq_conn_util_flags() {")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if len(v.Flags) == 0 {
			continue
		}
		add("    " + v.Name + ") printf " + bashQuote(offeredSpellings(v.Flags)) + " ;;")
	}
	add("    *) printf '' ;;")
	add("  esac")
	add("}")
	add("")

	add("# _solmq_conn_util_posarg <verb> prints the completion kind for the verb's own")
	add("# positional argument, or nothing when it takes none.")
	add("_solmq_conn_util_posarg() {")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if v.PosArg == posNone {
			continue
		}
		add("    " + v.Name + ") printf '" + v.PosArg + "' ;;")
	}
	add("    *) printf '' ;;")
	add("  esac")
	add("}")
	add("")

	add("# _solmq_conn_util_paths <word> <kind> fills COMPREPLY with matching paths.")
	add("# compopt is guarded: bash 3.2, still the system bash on macOS, lacks it.")
	add("_solmq_conn_util_paths() {")
	add(`  local IFS=$'\n'`)
	add(`  if [ "$2" = 'dir' ]; then`)
	add(`    COMPREPLY=( $(compgen -d -- "$1") )`)
	add("  else")
	add(`    COMPREPLY=( $(compgen -f -- "$1") )`)
	add("  fi")
	add("  if type compopt >/dev/null 2>&1; then compopt -o filenames; fi")
	add("}")
	add("")

	add("_solmq_conn_util() {")
	add("  local cur prev word kind targets verb=''")
	add("  local i positional=0")
	add("")
	add(`  cur="${COMP_WORDS[COMP_CWORD]}"`)
	add("  prev=''")
	add(`  if [ "$COMP_CWORD" -gt 0 ]; then prev="${COMP_WORDS[$((COMP_CWORD - 1))]}"; fi`)
	add("")
	add("  # A value for the flag just typed wins over everything else.")
	add(`  kind="$(_solmq_conn_util_flag_arg "$prev")"`)
	add(`  case "$kind" in`)
	add(`    file) _solmq_conn_util_paths "$cur" file; return ;;`)
	add("    name) COMPREPLY=(); return ;;")
	add("  esac")
	add("")
	add("  # Walk what is already typed, skipping flags and their values, to find the")
	add("  # verb and how many positional arguments already follow it.")
	add("  i=1")
	add(`  while [ "$i" -lt "$COMP_CWORD" ]; do`)
	add(`    word="${COMP_WORDS[$i]}"`)
	add(`    case "$word" in`)
	add("      -*=*) ;;   # --env=path carries its own value; nothing to skip")
	add("      -*)")
	add(`        if [ -n "$(_solmq_conn_util_flag_arg "$word")" ]; then i=$((i + 1)); fi`)
	add("        ;;")
	add("      *)")
	add(`        if [ -z "$verb" ]; then verb="$word"; else positional=$((positional + 1)); fi`)
	add("        ;;")
	add("    esac")
	add("    i=$((i + 1))")
	add("  done")
	add("")
	add(`  if [ -z "$verb" ]; then`)
	add("    COMPREPLY=( $(compgen -W " + bashQuote(verbNamesAndHelp()) + ` -- "$cur") )`)
	add("    return")
	add("  fi")
	add("")
	add(`  case "$cur" in`)
	add("    -*)")
	add(`      COMPREPLY=( $(compgen -W "$(_solmq_conn_util_flags "$verb")" -- "$cur") )`)
	add("      return")
	add("      ;;")
	add("  esac")
	add("")
	add(`  if [ "$positional" -eq 0 ]; then`)
	add(`    targets="$(_solmq_conn_util_targets "$verb")"`)
	add(`    if [ -n "$targets" ]; then`)
	add(`      COMPREPLY=( $(compgen -W "$targets" -- "$cur") )`)
	add("      return")
	add("    fi")
	add(`    kind="$(_solmq_conn_util_posarg "$verb")"`)
	add(`    if [ -n "$kind" ]; then _solmq_conn_util_paths "$cur" "$kind"; return; fi`)
	add("  fi")
	add("")
	add("  COMPREPLY=()")
	add("}")
	add("")
	add("# .exe as well, for a Windows binary driven from git-bash.")
	add("complete -F _solmq_conn_util solmq-conn-util solmq-conn-util.exe")

	return strings.Join(l, "\n") + "\n"
}

// ---- zsh ----------------------------------------------------------------------

func renderZshCompletion() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	add("#compdef solmq-conn-util")
	l = append(l, scriptHeader("zsh", "#", []string{
		"mkdir -p ~/.zsh/completions",
		"solmq-conn-util completion zsh > ~/.zsh/completions/_solmq-conn-util",
		"",
		"The #compdef line above is what zsh autoloads on, so the directory must",
		"be on $fpath BEFORE compinit runs -- in ~/.zshrc:",
		"fpath=(~/.zsh/completions $fpath)",
		"autoload -Uz compinit && compinit",
		"",
		"Nothing completing? compinit is serving a stale cache:",
		"rm -f ~/.zcompdump* && exec zsh",
	})...)

	add("# _solmq_conn_util_flag_arg <word> prints the value kind the flag consumes:")
	add("# 'file', 'name', or nothing. Boolean flags fall through to the catch-all.")
	add("_solmq_conn_util_flag_arg() {")
	add(`  case "$1" in`)
	for _, f := range completionFlags() {
		if f.Arg == argNone {
			continue
		}
		add("    " + strings.Join(f.Aliases, "|") + ") print -n '" + f.Arg + "' ;;")
	}
	add("    *) print -n '' ;;")
	add("  esac")
	add("}")
	add("")

	add("# _solmq_conn_util_targets <verb> offers the verb's targets; 1 when it has none.")
	add("_solmq_conn_util_targets() {")
	add("  local -a t")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if len(v.Targets) == 0 {
			continue
		}
		entries := make([]string, 0, len(v.Targets))
		for _, tg := range v.Targets {
			entries = append(entries, zshEntry(tg.Name, tg.Desc))
		}
		add("    " + v.Name + ") t=(" + strings.Join(entries, " ") + ") ;;")
	}
	add("    *) return 1 ;;")
	add("  esac")
	add("  _describe -t targets 'target' t")
	add("  return 0")
	add("}")
	add("")

	add("# _solmq_conn_util_flags <verb> offers the verb's flags; 1 when it has none.")
	add("_solmq_conn_util_flags() {")
	add("  local -a f")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if len(v.Flags) == 0 {
			continue
		}
		entries := make([]string, 0, 2*len(v.Flags))
		for _, fl := range v.Flags {
			for _, sp := range fl.Offer {
				entries = append(entries, zshEntry(sp, fl.Desc))
			}
		}
		add("    " + v.Name + ") f=(" + strings.Join(entries, " ") + ") ;;")
	}
	add("    *) return 1 ;;")
	add("  esac")
	add("  _describe -t options 'option' f")
	add("  return 0")
	add("}")
	add("")

	add("# _solmq_conn_util_posarg <verb> prints the completion kind for the verb's own")
	add("# positional argument, or nothing when it takes none.")
	add("_solmq_conn_util_posarg() {")
	add(`  case "$1" in`)
	for _, v := range completionVerbs() {
		if v.PosArg == posNone {
			continue
		}
		add("    " + v.Name + ") print -n '" + v.PosArg + "' ;;")
	}
	add("    *) print -n '' ;;")
	add("  esac")
	add("}")
	add("")

	add("_solmq_conn_util() {")
	add("  local -a verbs")
	add("  local verb='' cur prev kind")
	add("  local -i i positional=0")
	add("")
	add(`  cur="${words[CURRENT]}"`)
	add("  prev=''")
	add(`  (( CURRENT > 1 )) && prev="${words[CURRENT-1]}"`)
	add("")
	add("  # A value for the flag just typed wins over everything else.")
	add(`  kind="$(_solmq_conn_util_flag_arg "$prev")"`)
	add(`  case "$kind" in`)
	add("    file) _files; return ;;")
	add("    name) return ;;")
	add("  esac")
	add("")
	add("  # words is 1-indexed and words[1] is the command, so the already-typed")
	add("  # arguments are 2..CURRENT-1. Flags and their values are skipped, because")
	add("  # they may appear before, after, or between the positional arguments.")
	add("  for (( i = 2; i < CURRENT; i++ )); do")
	add(`    case "${words[i]}" in`)
	add("      -*=*) ;;")
	add(`      -*) [[ -n "$(_solmq_conn_util_flag_arg "${words[i]}")" ]] && (( i++ )) ;;`)
	add(`      *) if [[ -z "$verb" ]]; then verb="${words[i]}"; else (( positional++ )); fi ;;`)
	add("    esac")
	add("  done")
	add("")
	add(`  if [[ -z "$verb" ]]; then`)
	add("    verbs=(")
	for _, v := range completionVerbs() {
		add("      " + zshEntry(v.Name, v.Desc))
	}
	add("      " + zshEntry("-h", "Print the usage summary"))
	add("      " + zshEntry("--help", "Print the usage summary"))
	add("    )")
	add("    _describe -t commands 'command' verbs")
	add("    return")
	add("  fi")
	add("")
	add(`  if [[ "$cur" == -* ]]; then`)
	add(`    _solmq_conn_util_flags "$verb"`)
	add("    return")
	add("  fi")
	add("")
	add("  if (( positional == 0 )); then")
	add(`    _solmq_conn_util_targets "$verb" && return`)
	add(`    case "$(_solmq_conn_util_posarg "$verb")" in`)
	add("      dir) _files -/; return ;;")
	add("      file) _files; return ;;")
	add("    esac")
	add("  fi")
	add("}")
	add("")
	add("# Works both ways: autoloaded from $fpath, or sourced directly from a shell rc.")
	add(`if [ "$funcstack[1]" = '_solmq_conn_util' ]; then`)
	add("  _solmq_conn_util")
	add("else")
	add("  compdef _solmq_conn_util solmq-conn-util")
	add("fi")

	return strings.Join(l, "\n") + "\n"
}

// ---- fish ---------------------------------------------------------------------

func renderFishCompletion() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	l = append(l, scriptHeader("fish", "#", []string{
		"fish autoloads this path, so writing the file is the whole install --",
		"no rc edit, and new shells pick it up:",
		"mkdir -p ~/.config/fish/completions",
		"solmq-conn-util completion fish > ~/.config/fish/completions/solmq-conn-util.fish",
		"",
		"This session only, without installing anything:",
		"solmq-conn-util completion fish | source",
	})...)

	add("# No bare filename completion: every position below opts back in explicitly.")
	add("complete -c solmq-conn-util -f")
	add("")

	verbs := completionVerbs()

	add("# Verbs, offered only while no verb has been given yet.")
	for _, v := range verbs {
		add("complete -c solmq-conn-util -n '__fish_use_subcommand' -a " + fishQuote(v.Name) + " -d " + fishQuote(v.Desc))
	}
	add("complete -c solmq-conn-util -n '__fish_use_subcommand' -s h -l help -d " + fishQuote("Print the usage summary"))
	add("")

	add("# Targets, offered once their verb is seen and until one is chosen.")
	for _, v := range verbs {
		if len(v.Targets) == 0 {
			continue
		}
		cond := "__fish_seen_subcommand_from " + v.Name +
			"; and not __fish_seen_subcommand_from " + itemNames(v.Targets)
		for _, tg := range v.Targets {
			add("complete -c solmq-conn-util -n " + fishQuote(cond) + " -a " + fishQuote(tg.Name) + " -d " + fishQuote(tg.Desc))
		}
	}
	add("")

	add("# Positional arguments that are not targets.")
	for _, v := range verbs {
		if v.PosArg == posNone {
			continue
		}
		// posDir is the only non-none kind; TestCompletionModelMetadataComplete
		// fails loudly if a verb ever declares one the renderers do not handle.
		if v.PosArg == posDir {
			cond := "__fish_seen_subcommand_from " + v.Name
			add("complete -c solmq-conn-util -n " + fishQuote(cond) + " -a '(__fish_complete_directories)' -d " + fishQuote("directory"))
		}
	}
	add("")

	add("# Flags, scoped to the verbs that accept them. -r means the flag takes a value;")
	add("# -F completes a file for it. --allow-command takes a value but not a filename,")
	add("# so it gets -r without -F and offers nothing.")
	for _, v := range verbs {
		for _, f := range v.Flags {
			spec := "complete -c solmq-conn-util -n " + fishQuote("__fish_seen_subcommand_from "+v.Name) +
				" " + fishFlagSpec(f)
			switch f.Arg {
			case argFile:
				spec += " -r -F"
			case argName:
				spec += " -r"
			}
			add(spec + " -d " + fishQuote(f.Desc))
		}
	}

	return strings.Join(l, "\n") + "\n"
}

// fishFlagSpec renders a flag as fish's selectors: a short+long pair becomes
// "-s e -l env", a flag with no short form just "-l allow-command".
func fishFlagSpec(f completionFlag) string {
	long := strings.TrimLeft(f.Long, "-")
	short := strings.TrimLeft(f.Short, "-")
	if f.Short == f.Long || len(short) != 1 {
		return "-l " + long
	}
	return "-s " + short + " -l " + long
}

// ---- powershell ---------------------------------------------------------------

func renderPowerShellCompletion() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	l = append(l, scriptHeader("PowerShell", "#", []string{
		"Register-ArgumentCompleter below is per-session, so appending to the",
		"profile is what makes it stick:",
		"solmq-conn-util completion powershell >> $PROFILE",
		"",
		"This session only, without touching the profile:",
		"solmq-conn-util completion powershell | Out-String | Invoke-Expression",
	})...)

	add("# Windows PowerShell 5.1 compatible: no &&/||, no ternary, no null-coalescing.")
	add("# Both command names are registered because the Windows binary is solmq-conn-util.exe.")
	add("Register-ArgumentCompleter -Native -CommandName @('solmq-conn-util', 'solmq-conn-util.exe') -ScriptBlock {")
	add("    param($wordToComplete, $commandAst, $cursorPosition)")
	add("")

	add("    # Value kind each flag consumes. Both the one- and two-dash spellings are")
	add("    # present because Go's flag package accepts either.")
	add("    $flagArg = @{}")
	for _, f := range completionFlags() {
		if f.Arg == argNone {
			continue
		}
		for _, a := range f.Aliases {
			add("    $flagArg[" + psQuote(a) + "] = " + psQuote(f.Arg))
		}
	}
	add("")

	add("    $verbs = @(")
	for _, v := range completionVerbs() {
		add("        @{ Name = " + psQuote(v.Name) + "; Desc = " + psQuote(v.Desc) + " }")
	}
	add("        @{ Name = '-h'; Desc = 'Print the usage summary' }")
	add("        @{ Name = '--help'; Desc = 'Print the usage summary' }")
	add("    )")
	add("")

	add("    $targets = @{}")
	for _, v := range completionVerbs() {
		if len(v.Targets) == 0 {
			continue
		}
		items := make([]string, 0, len(v.Targets))
		for _, tg := range v.Targets {
			items = append(items, "@{ Name = "+psQuote(tg.Name)+"; Desc = "+psQuote(tg.Desc)+" }")
		}
		add("    $targets[" + psQuote(v.Name) + "] = @(" + strings.Join(items, ", ") + ")")
	}
	add("")

	add("    $flags = @{}")
	for _, v := range completionVerbs() {
		if len(v.Flags) == 0 {
			continue
		}
		items := make([]string, 0, 2*len(v.Flags))
		for _, f := range v.Flags {
			for _, sp := range f.Offer {
				items = append(items, "@{ Name = "+psQuote(sp)+"; Desc = "+psQuote(f.Desc)+" }")
			}
		}
		add("    $flags[" + psQuote(v.Name) + "] = @(" + strings.Join(items, ", ") + ")")
	}
	add("")

	add("    $posArg = @{}")
	for _, v := range completionVerbs() {
		if v.PosArg == posNone {
			continue
		}
		add("    $posArg[" + psQuote(v.Name) + "] = " + psQuote(v.PosArg))
	}
	add("")

	add("    $emit = {")
	add("        param($items, $word)")
	add("        foreach ($it in $items) {")
	add("            if ($it.Name.StartsWith($word, [System.StringComparison]::OrdinalIgnoreCase)) {")
	add("                [System.Management.Automation.CompletionResult]::new(")
	add("                    $it.Name, $it.Name, 'ParameterValue', $it.Desc)")
	add("            }")
	add("        }")
	add("    }")
	add("")

	add("    # Path completion, keeping whatever directory prefix the user already typed.")
	add("    $emitPaths = {")
	add("        param($word, $dirsOnly)")
	add("        $prefix = ''")
	add("        $leaf = $word")
	add("        $cut = $word.LastIndexOfAny([char[]]@('\\', '/'))")
	add("        if ($cut -ge 0) {")
	add("            $prefix = $word.Substring(0, $cut + 1)")
	add("            $leaf = $word.Substring($cut + 1)")
	add("        }")
	add("        $dir = '.'")
	add("        if ($prefix -ne '') { $dir = $prefix }")
	add("        $items = Get-ChildItem -LiteralPath $dir -ErrorAction SilentlyContinue")
	add("        foreach ($it in $items) {")
	add("            if ($dirsOnly -and (-not $it.PSIsContainer)) { continue }")
	add("            if (-not $it.Name.StartsWith($leaf, [System.StringComparison]::OrdinalIgnoreCase)) { continue }")
	add("            $path = $prefix + $it.Name")
	add("            $text = $path")
	add("            if ($text.Contains(' ')) { $text = \"'\" + $text + \"'\" }")
	add("            $kind = 'ProviderItem'")
	add("            if ($it.PSIsContainer) { $kind = 'ProviderContainer' }")
	add("            [System.Management.Automation.CompletionResult]::new($text, $path, $kind, $path)")
	add("        }")
	add("    }")
	add("")

	add("    # The words already typed, excluding the command and the word under the cursor.")
	add("    $words = @()")
	add("    $elements = $commandAst.CommandElements")
	add("    for ($i = 1; $i -lt $elements.Count; $i++) { $words += [string]$elements[$i].Extent.Text }")
	add("    if ($wordToComplete -ne '' -and $words.Count -gt 0) {")
	add("        if ($words[$words.Count - 1] -eq $wordToComplete) {")
	add("            if ($words.Count -eq 1) { $words = @() }")
	add("            else { $words = $words[0..($words.Count - 2)] }")
	add("        }")
	add("    }")
	add("")

	add("    # A value for the flag just typed wins over everything else.")
	add("    $prev = ''")
	add("    if ($words.Count -gt 0) { $prev = $words[$words.Count - 1] }")
	add("    if ($flagArg.ContainsKey($prev)) {")
	add("        if ($flagArg[$prev] -eq 'file') { & $emitPaths $wordToComplete $false }")
	add("        return")
	add("    }")
	add("")

	add("    # Walk what is typed, skipping flags and their values, to find the verb and")
	add("    # how many positional arguments already follow it.")
	add("    $verb = ''")
	add("    $positional = 0")
	add("    $i = 0")
	add("    while ($i -lt $words.Count) {")
	add("        $w = $words[$i]")
	add("        if ($w.StartsWith('-')) {")
	add("            if (-not $w.Contains('=')) {")
	add("                if ($flagArg.ContainsKey($w)) { $i++ }")
	add("            }")
	add("        }")
	add("        else {")
	add("            if ($verb -eq '') { $verb = $w }")
	add("            else { $positional++ }")
	add("        }")
	add("        $i++")
	add("    }")
	add("")

	add("    if ($verb -eq '') {")
	add("        & $emit $verbs $wordToComplete")
	add("        return")
	add("    }")
	add("")
	add("    if ($wordToComplete.StartsWith('-')) {")
	add("        if ($flags.ContainsKey($verb)) { & $emit $flags[$verb] $wordToComplete }")
	add("        return")
	add("    }")
	add("")
	add("    if ($positional -eq 0) {")
	add("        if ($targets.ContainsKey($verb)) {")
	add("            & $emit $targets[$verb] $wordToComplete")
	add("            return")
	add("        }")
	add("        if ($posArg.ContainsKey($verb)) {")
	add("            & $emitPaths $wordToComplete ($posArg[$verb] -eq 'dir')")
	add("            return")
	add("        }")
	add("    }")
	add("}")

	return strings.Join(l, "\n") + "\n"
}

// ---- small shared formatters --------------------------------------------------

// itemNames joins completion item names with spaces, for the word lists bash's
// compgen -W and fish's __fish_seen_subcommand_from both take.
func itemNames(items []completionItem) string {
	names := make([]string, 0, len(items))
	for _, it := range items {
		names = append(names, it.Name)
	}
	return strings.Join(names, " ")
}

// offeredSpellings joins every spelling the given flags offer, space separated.
func offeredSpellings(flags []completionFlag) string {
	var out []string
	for _, f := range flags {
		out = append(out, f.Offer...)
	}
	return strings.Join(out, " ")
}

// verbNamesAndHelp is the word-1 list: every verb, plus the help aliases that
// dispatch accepts as args[0] but the model cannot spell as a verb name.
func verbNamesAndHelp() string {
	names := make([]string, 0, len(cliVerbs)+2)
	for _, v := range cliVerbs {
		names = append(names, v.Name)
	}
	return strings.Join(append(names, "-h", "--help"), " ")
}
