package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The completion scripts are not shipped as files -- `solmq-conn-util
// auto-complete <shell>` renders them from the compiled-in model -- so these snapshots under
// testdata/ are fixtures, not artifacts. They exist so that a change to
// cliVerbs/cliFlags shows up as a reviewable diff in the generated shell code
// rather than only as a behaviour change nobody sees until they press TAB.
//
// Regenerate with the same -update flag docs/commands.md uses:
//
//	go generate ./cmd/solmq-conn-util
const completionGoldenDir = "testdata/completions"

// completionGolden names each modeled shell's snapshot file, using that shell's
// own convention (zsh loads a completion from a file named _<command>).
var completionGolden = map[string]string{
	"bash":       "solmq-conn-util.bash",
	"zsh":        "_solmq-conn-util",
	"fish":       "solmq-conn-util.fish",
	"powershell": "solmq-conn-util.ps1",
}

// renderedCompletions renders every modeled shell, failing if the model and the
// renderer map have drifted apart (TestDispatchHandlersMatchModel gates that
// too, but this test must not silently skip a shell when it happens).
func renderedCompletions(t *testing.T) map[string]string {
	t.Helper()
	out := make(map[string]string, len(completionShellRenderers))
	for _, shell := range completionShells() {
		render, ok := completionShellRenderers[shell]
		if !ok {
			t.Fatalf("shell %q is modeled but has no renderer", shell)
		}
		out[shell] = render()
	}
	return out
}

// TestCompletionGoldenInSync is the drift gate: each shell's script must equal
// its committed snapshot. It runs under `dev.sh test` -> `all` -> CI, so a
// command added to the model without regenerating turns the build red.
func TestCompletionGoldenInSync(t *testing.T) {
	for shell, got := range renderedCompletions(t) {
		name, ok := completionGolden[shell]
		if !ok {
			t.Fatalf("shell %q is modeled but has no snapshot filename in completionGolden", shell)
		}
		t.Run(shell, func(t *testing.T) {
			path := filepath.Join(filepath.FromSlash(completionGoldenDir), name)
			if *updateDoc {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatalf("update %s: %v", path, err)
				}
				t.Logf("regenerated %s", path)
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v\nregenerate: go generate ./cmd/solmq-conn-util", path, err)
			}
			if normLF(string(want)) != normLF(got) {
				t.Errorf("%s is out of sync with the command model (cmd/solmq-conn-util/commands.go).\n"+
					"regenerate: go generate ./cmd/solmq-conn-util", path)
			}
		})
	}
}

// TestCompletionCoversModel is the semantic half of the gate the snapshots
// cannot give: a snapshot only proves the output did not change, so on its own a
// new verb that renders into none of the scripts would pass once regenerated.
// This asserts every modeled name actually reaches every shell.
func TestCompletionCoversModel(t *testing.T) {
	// Per shell, the literal a flag is expected to appear as. fish names a flag
	// by its letters rather than its spelling ("-s e -l env"), so it needs the
	// production renderer rather than the offered spellings.
	flagLiterals := map[string]func(completionFlag) []string{
		"bash":       func(f completionFlag) []string { return f.Offer },
		"zsh":        func(f completionFlag) []string { return f.Offer },
		"powershell": func(f completionFlag) []string { return f.Offer },
		"fish":       func(f completionFlag) []string { return []string{fishFlagSpec(f)} },
	}

	// bash's compgen word lists carry no descriptions; the other three do.
	describes := map[string]bool{"zsh": true, "fish": true, "powershell": true}

	// The -h/--help pair is spelled the same way flags are, per shell.
	helpLiterals := map[string][]string{
		"bash":       {"-h", "--help"},
		"zsh":        {"-h", "--help"},
		"powershell": {"-h", "--help"},
		"fish":       {"-s h -l help"},
	}

	for shell, script := range renderedCompletions(t) {
		t.Run(shell, func(t *testing.T) {
			spell, ok := flagLiterals[shell]
			if !ok {
				t.Fatalf("shell %q has no expected flag spelling; add one when adding a shell", shell)
			}
			if _, ok := helpLiterals[shell]; !ok {
				t.Fatalf("shell %q has no expected help spelling; add one when adding a shell", shell)
			}
			contains := func(what, want string) {
				t.Helper()
				if !strings.Contains(script, want) {
					t.Errorf("%s completion is missing %s %q", shell, what, want)
				}
			}
			for _, v := range completionVerbs() {
				contains("verb", v.Name)
				if describes[shell] {
					contains("verb description", v.Desc)
				}
				// fish has no single $verb variable to normalize (unlike bash/zsh/
				// powershell): it recognizes an alias only by widening a condition
				// scoped to that verb's targets/posarg/flags. A verb with none of
				// those (version) has nothing in the fish script to widen, but also
				// nothing beyond word 1 to complete either way, so there is no
				// normalization to forget and the alias is exempt there.
				if shell != "fish" || len(v.Targets) > 0 || v.PosArg == posDir || len(v.Flags) > 0 {
					for _, a := range v.Aliases {
						contains("verb alias", a)
					}
				}
				for _, tg := range v.Targets {
					contains("target", tg.Name)
					if describes[shell] {
						contains("target description", tg.Desc)
					}
					for _, s := range tg.Sets {
						contains("set", s.Name)
						if describes[shell] {
							contains("set description", s.Desc)
						}
					}
				}
				for _, f := range v.Flags {
					for _, lit := range spell(f) {
						contains("flag", lit)
					}
					if describes[shell] {
						contains("flag description", f.Desc)
					}
				}
			}
			for _, lit := range helpLiterals[shell] {
				contains("help alias", lit)
			}
		})
	}
}

// downloadJarTarget returns download's "jar" target (the one modeled target
// with a third level today), failing the test if the model does not have it.
func downloadJarTarget(t *testing.T) completionItem {
	t.Helper()
	for _, v := range completionVerbs() {
		if v.Name != "download" {
			continue
		}
		for _, tg := range v.Targets {
			if tg.Name == "jar" {
				return tg
			}
		}
	}
	t.Fatal("model has no download/jar target")
	return completionItem{}
}

// TestCompletionOnlyDownloadJarHasSets pins the third command level to exactly
// where the model puts it today. A future target silently growing a Sets list
// would otherwise only surface as a completion behavior nobody notices --
// this makes it a reviewable model diff instead, and documents the assumption
// every renderer's position-3/4 dispatch logic relies on.
func TestCompletionOnlyDownloadJarHasSets(t *testing.T) {
	for _, v := range completionVerbs() {
		for _, tg := range v.Targets {
			if v.Name == "download" && tg.Name == "jar" {
				continue
			}
			if len(tg.Sets) != 0 {
				t.Errorf("target %q under verb %q unexpectedly has Sets; only download/jar is expected to", tg.Name, v.Name)
			}
		}
	}
}

// TestCompletionThirdLevelOffersSets checks the new third command level end to
// end: download's only target (jar) fans out into its two sets, and every
// renderer offers them once "download jar" (or the alias "dl jar") has been
// typed. bash/zsh/powershell key their lookup construct by the CANONICAL verb
// name only, because $verb is normalized from an alias to its canonical
// spelling once, before any position-based lookup runs (the same
// normalization TestCompletionVerbAliasesResolveToCanonical checks elsewhere)
// -- so "download" appearing here already covers "dl". fish has no single
// $verb to normalize, so its own condition lists both names directly, checked
// below.
func TestCompletionThirdLevelOffersSets(t *testing.T) {
	scripts := renderedCompletions(t)
	jar := downloadJarTarget(t)
	if len(jar.Sets) == 0 {
		t.Fatal("download/jar has no Sets; nothing to test")
	}

	cases := map[string]string{
		"bash": "    download)\n" +
			`      case "$2" in` + "\n" +
			`        jar) printf 'mq syslog' ;;`,
		"zsh": "    download)\n" +
			`      case "$2" in` + "\n" +
			`        jar) s=(`,
		"powershell": "$sets['download']['jar'] = @(",
		"fish": "__fish_seen_subcommand_from download dl" +
			"; and __fish_seen_subcommand_from jar" +
			"; and not __fish_seen_subcommand_from mq syslog",
	}
	for shell, want := range cases {
		t.Run(shell, func(t *testing.T) {
			if !strings.Contains(scripts[shell], want) {
				t.Errorf("%s completion is missing the third-level sets construct %q", shell, want)
			}
		})
	}

	// The set names and descriptions themselves reach every shell verbatim --
	// the same guarantee TestCompletionCoversModel gives the first two levels
	// (and now, via the loop added there, this level too); repeated here
	// scoped to just download/jar so a failure names the exact pair.
	for _, s := range jar.Sets {
		for shell, script := range scripts {
			if !strings.Contains(script, s.Name) {
				t.Errorf("%s completion is missing set %q", shell, s.Name)
			}
			if shell != "bash" && !strings.Contains(script, s.Desc) {
				t.Errorf("%s completion is missing set description %q", shell, s.Desc)
			}
		}
	}
}

// TestCompletionThirdLevelUnlocksPosArg checks the word after the set: since
// download declares PosArg posDir, every renderer must offer a directory once
// "download jar mq" (verb, target, set) has all been typed -- but must NOT
// offer one after just "download jar" (verb, target only), which is exactly
// the regression the position-3/4 split guards against.
func TestCompletionThirdLevelUnlocksPosArg(t *testing.T) {
	cases := map[string]string{
		"bash": `  elif [ "$positional" -eq 2 ]; then` + "\n" +
			`    if [ -n "$(_solmq_conn_util_sets "$verb" "$arg1")" ]; then`,
		"zsh": "  elif (( positional == 2 )); then\n" +
			`    if _solmq_conn_util_has_sets "$verb" "$arg1"; then`,
		"powershell": "elseif ($positional -eq 2) {\n" +
			"        if ($sets.ContainsKey($verb) -and $sets[$verb].ContainsKey($arg1)) {",
		"fish": "__fish_seen_subcommand_from download dl" +
			"; and __fish_seen_subcommand_from jar" +
			"; and __fish_seen_subcommand_from mq syslog",
	}
	scripts := renderedCompletions(t)
	for shell, want := range cases {
		t.Run(shell, func(t *testing.T) {
			if !strings.Contains(scripts[shell], want) {
				t.Errorf("%s completion does not unlock the directory positional after a set is typed (want %q)", shell, want)
			}
		})
	}
}

// TestCompletionRecognizesFlagAliases covers the value-skipping table: Go's flag
// package accepts -env for --env, so a script that only knows the documented
// spelling would treat the value after -env as a positional and offer targets
// where a filename belongs. fish is excluded -- its -s/-l model has no alias
// table and covers the two documented spellings only.
func TestCompletionRecognizesFlagAliases(t *testing.T) {
	scripts := renderedCompletions(t)
	for _, shell := range []string{"bash", "zsh", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			for _, f := range completionFlags() {
				if f.Arg == argNone {
					continue // consumes no value, so there is nothing to skip
				}
				for _, alias := range f.Aliases {
					if !strings.Contains(scripts[shell], alias) {
						t.Errorf("%s completion does not recognize flag spelling %q", shell, alias)
					}
				}
			}
		})
	}
}

// TestCompletionDownloadFlagsDescribed pins every flag download introduces
// (--url, --version, --omit-lib-file, --include-provided) by literal
// name, on top of the generic per-flag coverage TestCompletionCoversModel
// already gives every modeled flag -- so a typo in a Long spelling, or a
// Meaning that never made it into cliFlags, fails here by name instead of
// only as an unnamed gap in the generic loop.
func TestCompletionDownloadFlagsDescribed(t *testing.T) {
	// bash's compgen word lists carry no descriptions; the other three do.
	describes := map[string]bool{"zsh": true, "fish": true, "powershell": true}

	scripts := renderedCompletions(t)
	want := map[string]bool{
		"--url":              false,
		"--version":          false,
		"--omit-lib-file":    false,
		"--include-provided": false,
	}
	for _, f := range completionFlags() {
		if _, ok := want[f.Long]; !ok {
			continue
		}
		want[f.Long] = true
		for shell, script := range scripts {
			if !describes[shell] {
				continue
			}
			if !strings.Contains(script, f.Desc) {
				t.Errorf("%s completion is missing %s description %q", shell, f.Long, f.Desc)
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected flag %q is not modeled in cliFlags", name)
		}
	}
}

// TestCompletionOmitLibFileCompletesFiles checks that --omit-lib-file is wired
// to the same file-completion path -e/--env already gets (TestCompletionValueKindsReachScripts
// above proves that path exists at all; this proves --omit-lib-file, specifically,
// is on it): the flag has no short form, so its own alias/spec construct is
// pinned by its Long spelling alone in every shell.
func TestCompletionOmitLibFileCompletesFiles(t *testing.T) {
	scripts := renderedCompletions(t)
	cases := map[string]string{
		"bash":       "-omit-lib-file|--omit-lib-file) printf 'file' ;;",
		"zsh":        "-omit-lib-file|--omit-lib-file) print -n 'file' ;;",
		"fish":       "-l omit-lib-file -r -F",
		"powershell": "$flagArg['--omit-lib-file'] = 'file'",
	}
	for shell, want := range cases {
		t.Run(shell, func(t *testing.T) {
			if !strings.Contains(scripts[shell], want) {
				t.Errorf("%s completion does not complete files after --omit-lib-file (want %q)", shell, want)
			}
		})
	}
}

// TestCompletionShellStructure pins the one line per shell that makes the script
// load at all. Losing it fails no other assertion: every name would still be
// present and the file would simply do nothing.
func TestCompletionShellStructure(t *testing.T) {
	cases := map[string]string{
		"bash":       "complete -F _solmq_conn_util solmq-conn-util solmq-conn-util.exe",
		"zsh":        "compdef _solmq_conn_util solmq-conn-util",
		"fish":       "complete -c solmq-conn-util -f",
		"powershell": "Register-ArgumentCompleter -Native -CommandName @('solmq-conn-util', 'solmq-conn-util.exe')",
	}
	scripts := renderedCompletions(t)
	// Driven off the model, not the table, so a shell added without a case here
	// fails rather than quietly going unchecked.
	for shell, script := range scripts {
		t.Run(shell, func(t *testing.T) {
			want, ok := cases[shell]
			if !ok {
				t.Fatalf("shell %q has no expected registration line; add one when adding a shell", shell)
			}
			if !strings.Contains(script, want) {
				t.Errorf("%s completion is missing its registration line %q", shell, want)
			}
		})
	}
	// zsh only autoloads a completion whose file starts with the #compdef tag.
	if first, _, _ := strings.Cut(scripts["zsh"], "\n"); first != "#compdef solmq-conn-util" {
		t.Errorf("zsh completion must open with #compdef solmq-conn-util, got %q", first)
	}
}

// TestCompletionValueKindsReachScripts checks the two value kinds that change
// what a shell offers mid-command: a path after -e/-o, and a directory after
// examples. Both are what makes the completion useful rather than decorative.
func TestCompletionValueKindsReachScripts(t *testing.T) {
	cases := map[string]struct{ file, dir string }{
		"bash":       {file: "printf 'file'", dir: "printf 'dir'"},
		"zsh":        {file: "print -n 'file'", dir: "print -n 'dir'"},
		"fish":       {file: "-r -F", dir: "__fish_complete_directories"},
		"powershell": {file: "$flagArg['-e'] = 'file'", dir: "$posArg['examples'] = 'dir'"},
	}
	for shell, script := range renderedCompletions(t) {
		t.Run(shell, func(t *testing.T) {
			want, ok := cases[shell]
			if !ok {
				t.Fatalf("shell %q has no expected value-kind markers; add them when adding a shell", shell)
			}
			if !strings.Contains(script, want.file) {
				t.Errorf("%s completion offers no file completion for a path flag (want %q)", shell, want.file)
			}
			if !strings.Contains(script, want.dir) {
				t.Errorf("%s completion offers no directory completion for examples (want %q)", shell, want.dir)
			}
		})
	}
}

// TestCompletionOutputIsPlainASCIILF holds the generated scripts to the same
// output contract as every other file this repo emits: plain ASCII, LF only, and
// a trailing newline. A stray CR or a smart quote breaks a shell that sources it.
func TestCompletionOutputIsPlainASCIILF(t *testing.T) {
	for shell, script := range renderedCompletions(t) {
		t.Run(shell, func(t *testing.T) {
			for i := 0; i < len(script); i++ {
				if script[i] >= 0x80 {
					t.Fatalf("%s completion has a non-ASCII byte %#x at offset %d", shell, script[i], i)
				}
				if script[i] == '\r' {
					t.Fatalf("%s completion has a CR at offset %d; output must be LF only", shell, i)
				}
			}
			if !strings.HasSuffix(script, "\n") {
				t.Errorf("%s completion does not end with a newline", shell)
			}
		})
	}
}

// TestCompletionModelMetadataComplete fails loudly (rule 4a: a systemic
// misconfiguration stops the run at setup) when a verb or flag is added without
// the metadata the renderers need. Without it a new verb would render with an
// empty tooltip, or with a value kind no shell knows how to honour, and only the
// snapshot diff would hint at it.
func TestCompletionModelMetadataComplete(t *testing.T) {
	knownPosArgs := map[string]bool{posNone: true, posDir: true}
	knownArgs := map[string]bool{argNone: true, argFile: true, argName: true}

	// Collides with a verb name, another alias, or -h/--help would make dispatch's
	// alias resolution (or a shell's own $verb-normalization case) ambiguous.
	seen := map[string]string{
		"-h":     "the -h help literal",
		"--help": "the --help help literal",
	}
	for _, v := range cliVerbs {
		seen[v.Name] = "verb " + v.Name
	}

	for _, v := range cliVerbs {
		if verbBlurb(v) == "" {
			t.Errorf("verb %q has neither Blurb nor Summary, so the completion scripts would describe it as nothing", v.Name)
		}
		if !knownPosArgs[v.PosArg] {
			t.Errorf("verb %q declares PosArg %q, which no renderer handles", v.Name, v.PosArg)
		}
		assertShellSafeName(t, "verb", v.Name)
		for _, a := range v.Aliases {
			assertShellSafeName(t, "verb alias", a)
			if prior, ok := seen[a]; ok {
				t.Errorf("verb %q alias %q collides with %s", v.Name, a, prior)
				continue
			}
			seen[a] = "verb " + v.Name + " alias " + a
		}
		for _, tg := range v.Targets {
			if plainText(tg.Summary) == "" {
				t.Errorf("target %q under %q has an empty Summary", tg.Name, v.Name)
			}
			assertShellSafeName(t, "target", tg.Name)
			// A target alias is spliced into shell case patterns and fish
			// conditions exactly as a verb alias is, so it needs the same charset
			// gate -- and it must not collide with the canonical word of any
			// target under the same verb, which would make the pattern ambiguous.
			for _, a := range tg.Aliases {
				assertShellSafeName(t, "target alias", a)
				for _, other := range v.Targets {
					if other.Name == a {
						t.Errorf("%q target alias %q collides with target %q", v.Name, a, other.Name)
					}
				}
			}
			for _, s := range tg.Sets {
				if plainText(s.Summary) == "" {
					t.Errorf("set %q under %q %q has an empty Summary", s.Name, v.Name, tg.Name)
				}
				assertShellSafeName(t, "set", s.Name)
			}
		}
	}

	// Every description reaching a completion script must avoid the apostrophe.
	// fish escapes it, powershell doubles it, and zsh's entries pass through
	// bashQuote, so the raw text never appears verbatim in those three scripts --
	// which is what TestCompletionCoversModel checks for. Failing here names the
	// one string to reword instead of three generated scripts that look wrong.
	for _, v := range completionVerbs() {
		if strings.Contains(v.Desc, "'") {
			t.Errorf("verb %q description carries an apostrophe, which three shells quote away: %q", v.Name, v.Desc)
		}
		for _, tg := range v.Targets {
			if strings.Contains(tg.Desc, "'") {
				t.Errorf("target %q description carries an apostrophe, which three shells quote away: %q", tg.Name, tg.Desc)
			}
			for _, st := range tg.Sets {
				if strings.Contains(st.Desc, "'") {
					t.Errorf("set %q description carries an apostrophe, which three shells quote away: %q", st.Name, st.Desc)
				}
			}
		}
	}

	for _, f := range cliFlags {
		if !knownArgs[f.Arg] {
			t.Errorf("flag %q declares Arg %q, which no renderer handles", f.Long, f.Arg)
		}
		if plainText(f.Meaning) == "" {
			t.Errorf("flag %q has an empty Meaning", f.Long)
		}
	}

	// Every modeled shell needs both a renderer and a snapshot filename; a shell
	// added to the model with neither would silently drop out of the gate.
	for _, shell := range completionShells() {
		if _, ok := completionShellRenderers[shell]; !ok {
			t.Errorf("shell %q is modeled but has no renderer", shell)
		}
		if _, ok := completionGolden[shell]; !ok {
			t.Errorf("shell %q is modeled but has no snapshot filename", shell)
		}
	}
}

// assertShellSafeName rejects a verb or target name that would not survive being
// spliced into a shell `case` pattern or a fish condition unquoted. Every name
// today is plain lowercase; this is the guard that keeps it that way.
func assertShellSafeName(t *testing.T, kind, name string) {
	t.Helper()
	if name == "" {
		t.Errorf("%s name is empty", kind)
		return
	}
	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			t.Errorf("%s name %q contains %q; names are spliced into shell case patterns unquoted and must be [a-z0-9-]", kind, name, r)
			return
		}
	}
}

// TestCompletionVerbAliasesResolveToCanonical checks the recognize half of
// recognize-but-do-not-suggest: every verb alias appears paired with its
// canonical verb in the construct that shell's own $verb (or, for fish, its
// per-verb condition) normalizes through -- so typing the alias completes
// exactly like the canonical verb. A renderer that forgot the normalization
// step would never emit this pairing even though the alias name alone might
// still show up elsewhere (TestCompletionCoversModel checks presence; this
// checks the alias resolves to the right verb).
func TestCompletionVerbAliasesResolveToCanonical(t *testing.T) {
	scripts := renderedCompletions(t)

	pairing := map[string]func(alias, canonical string) string{
		"bash":       func(alias, canonical string) string { return alias + `) verb="` + canonical + `" ;;` },
		"zsh":        func(alias, canonical string) string { return alias + `) verb="` + canonical + `" ;;` },
		"fish":       func(alias, canonical string) string { return "__fish_seen_subcommand_from " + canonical + " " + alias },
		"powershell": func(alias, canonical string) string { return "$verbAlias['" + alias + "'] = '" + canonical + "'" },
	}

	for shell, script := range scripts {
		t.Run(shell, func(t *testing.T) {
			build, ok := pairing[shell]
			if !ok {
				t.Fatalf("shell %q has no expected alias-pairing construct; add one when adding a shell", shell)
			}
			for _, v := range cliVerbs {
				for _, a := range v.Aliases {
					// fish only ever widens a condition scoped to a verb's targets/
					// posarg/flags, so a verb with none of those (version) has no
					// per-verb construct to pair the alias in -- but also nothing
					// beyond word 1 to complete either way, canonical or alias, so
					// there is no normalization to have forgotten. See the matching
					// exemption in TestCompletionCoversModel.
					if shell == "fish" && len(v.Targets) == 0 && v.PosArg != posDir && len(v.Flags) == 0 {
						continue
					}
					want := build(a, v.Name)
					if !strings.Contains(script, want) {
						t.Errorf("%s completion does not map alias %q to canonical verb %q (want %q)", shell, a, v.Name, want)
					}
				}
			}
		})
	}
}

// TestCompletionVerbAliasesNotOfferedAtWordOne is the other half of recognize-
// but-do-not-suggest: no verb alias may appear in the position-1 candidate
// list any shell's TAB menu draws from -- verbNamesAndHelp for bash, the zsh
// verbs array, the fish __fish_use_subcommand lines, and the powershell
// $verbs array. Checked against just that region, with word-boundary matching,
// because an alias legitimately appears elsewhere in the script (the
// normalization construct TestCompletionVerbAliasesResolveToCanonical checks)
// and a plain substring search would also false-positive on a canonical verb
// that happens to start with its own alias (e.g. "generate" contains "gen").
func TestCompletionVerbAliasesNotOfferedAtWordOne(t *testing.T) {
	scripts := renderedCompletions(t)

	region := map[string]func(t *testing.T, script string) string{
		"bash": func(t *testing.T, script string) string { return firstLineContaining(t, script, "compgen -W ") },
		"zsh": func(t *testing.T, script string) string {
			return between(t, script, "verbs=(", "_describe -t commands")
		},
		"fish": func(t *testing.T, script string) string { return linesContaining(script, "__fish_use_subcommand") },
		"powershell": func(t *testing.T, script string) string {
			return between(t, script, "$verbs = @(", "$verbAlias = @{}")
		},
	}

	for shell, script := range scripts {
		t.Run(shell, func(t *testing.T) {
			extract, ok := region[shell]
			if !ok {
				t.Fatalf("shell %q has no word-1 region extractor; add one when adding a shell", shell)
			}
			r := extract(t, script)
			for _, v := range cliVerbs {
				for _, a := range v.Aliases {
					if hasWord(r, a) {
						t.Errorf("%s word-1 candidate list unexpectedly offers alias %q", shell, a)
					}
				}
			}
		})
	}
}

// firstLineContaining returns the first line of script containing marker.
func firstLineContaining(t *testing.T, script, marker string) string {
	t.Helper()
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, marker) {
			return line
		}
	}
	t.Fatalf("no line contains %q", marker)
	return ""
}

// linesContaining returns every line of script containing marker, joined.
func linesContaining(script, marker string) string {
	var b strings.Builder
	for _, line := range strings.Split(script, "\n") {
		if strings.Contains(line, marker) {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// between returns the slice of script from start up to (not including) the
// first end found after it.
func between(t *testing.T, script, start, end string) string {
	t.Helper()
	i := strings.Index(script, start)
	if i < 0 {
		t.Fatalf("marker %q not found", start)
	}
	j := strings.Index(script[i:], end)
	if j < 0 {
		t.Fatalf("marker %q not found after %q", end, start)
	}
	return script[i : i+j]
}

// hasWord reports whether word appears in s at a regexp word boundary on both
// sides, not merely as a substring -- "gen" must not match inside "generate".
// Safe for every current alias (assertShellSafeName already limits a verb
// name to [a-z0-9-], and none of the aliases contain a hyphen).
func hasWord(s, word string) bool {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`)
	return re.MatchString(s)
}

// ---- text helpers -------------------------------------------------------------

func TestPlainText(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"code spans stripped", "default: `env.yaml`", "default: env.yaml"},
		{"newline folded to a space", "one\ntwo", "one two"},
		{"tab and CR folded", "a\tb\r\nc", "a b c"},
		{"whitespace runs collapse", "a    b", "a b"},
		{"leading and trailing trimmed", "  spaced  ", "spaced"},
		{"control chars dropped", "a\x00\x1bb", "a b"},
		{"empty stays empty", "", ""},
		{"only backticks becomes empty", "``", ""},
		{"punctuation preserved", "kubectl/oc apply -f - (stdin); repeatable", "kubectl/oc apply -f - (stdin); repeatable"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := plainText(c.in); got != c.want {
				t.Errorf("plainText(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestShellQuoting(t *testing.T) {
	cases := []struct {
		name, in, bash, fish, ps string
	}{
		{"plain", "abc", `'abc'`, `'abc'`, `'abc'`},
		{"empty", "", `''`, `''`, `''`},
		{"apostrophe", "don't", `'don'\''t'`, `'don\'t'`, `'don''t'`},
		{"backslash", `a\b`, `'a\b'`, `'a\\b'`, `'a\b'`},
		{"dollar and backtick", "$x `y`", "'$x `y`'", "'$x `y`'", "'$x `y`'"},
		{"double quote", `say "hi"`, `'say "hi"'`, `'say "hi"'`, `'say "hi"'`},
		{"semicolon and pipe", "a; b | c", `'a; b | c'`, `'a; b | c'`, `'a; b | c'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := bashQuote(c.in); got != c.bash {
				t.Errorf("bashQuote(%q) = %q, want %q", c.in, got, c.bash)
			}
			if got := fishQuote(c.in); got != c.fish {
				t.Errorf("fishQuote(%q) = %q, want %q", c.in, got, c.fish)
			}
			if got := psQuote(c.in); got != c.ps {
				t.Errorf("psQuote(%q) = %q, want %q", c.in, got, c.ps)
			}
		})
	}
}

func TestZshEntry(t *testing.T) {
	cases := []struct {
		name, value, desc, want string
	}{
		{"plain", "config", "Emit application.yml", `'config:Emit application.yml'`},
		{"colon in the value is escaped", "a:b", "desc", `'a\:b:desc'`},
		{"colon in the description is left alone", "gen", "a:b", `'gen:a:b'`},
		{"apostrophe quoted for zsh", "gen", "don't", `'gen:don'\''t'`},
		{"empty description", "gen", "", `'gen:'`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := zshEntry(c.value, c.desc); got != c.want {
				t.Errorf("zshEntry(%q, %q) = %q, want %q", c.value, c.desc, got, c.want)
			}
		})
	}
}

func TestFlagAliasesAndOffered(t *testing.T) {
	cases := []struct {
		name          string
		flag          cliFlag
		offer, alias  []string
		fishFlagsSpec string
	}{
		{
			name:  "short and long pair",
			flag:  cliFlag{Short: "-e", Long: "--env"},
			offer: []string{"-e", "--env"},
			// Both dash counts, both names: flag.Parse accepts all four.
			alias:         []string{"-e", "--e", "-env", "--env"},
			fishFlagsSpec: "-s e -l env",
		},
		{
			name:          "long only, no short alias",
			flag:          cliFlag{Short: "--allow-command", Long: "--allow-command"},
			offer:         []string{"--allow-command"},
			alias:         []string{"-allow-command", "--allow-command"},
			fishFlagsSpec: "-l allow-command",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertStrings(t, "flagOffered", flagOffered(c.flag), c.offer)
			assertStrings(t, "flagAliases", flagAliases(c.flag), c.alias)
			got := fishFlagSpec(completionFlag{Short: c.flag.Short, Long: c.flag.Long})
			if got != c.fishFlagsSpec {
				t.Errorf("fishFlagSpec = %q, want %q", got, c.fishFlagsSpec)
			}
		})
	}
}

func assertStrings(t *testing.T, label string, got, want []string) {
	t.Helper()
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}
