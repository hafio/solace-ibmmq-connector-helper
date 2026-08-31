package main

import (
	"sort"
	"strings"
)

// This file renders docs/abbreviation.md: every short spelling the CLI accepts,
// keyed by the abbreviation. The facts all live in the command model
// (cliVerb.Aliases, cliTarget.Aliases, cliFlag.Short) and in platformAliasList
// (main.go), and docs/commands.md already shows them scattered through the tree,
// the flag table and the per-verb detail blocks -- this page exists because the
// lookup a reader actually performs goes the other way: from "dl" to download.
// TestAbbreviationDocInSync gates the committed file against this renderer, and
// TestAbbreviationDocCoversModel gates the renderer against the model, so a new
// alias cannot reach the binary without reaching this page.

// abbrevRow is one lookup-table row: the short spelling plus the already
// rendered cells that follow it. Keeping the abbreviation separate from the rest
// is what lets every table -- two columns or four -- share one sort.
type abbrevRow struct {
	Short string
	Cells []string
}

// abbrevTable emits one markdown table, rows sorted by the abbreviation. Sorted
// rather than left in model order on purpose: model order is what commands.md
// renders, and a lookup page is only useful keyed by the word being looked up.
// Aliases are unique across the model, so the ordering is total and the output
// byte-stable.
func abbrevTable(add func(string), header []string, rows []abbrevRow) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].Short < rows[j].Short })

	seps := make([]string, 0, len(header))
	for _, h := range header {
		seps = append(seps, strings.Repeat("-", len(h)+2))
	}
	add("| " + strings.Join(header, " | ") + " |")
	add("|" + strings.Join(seps, "|") + "|")
	for _, r := range rows {
		cells := make([]string, 0, len(r.Cells)+1)
		cells = append(cells, bt+r.Short+bt)
		for _, c := range r.Cells {
			cells = append(cells, tableCell(c))
		}
		add("| " + strings.Join(cells, " | ") + " |")
	}
	add("")
}

// abbrevFlagByShort finds the modeled flag with the given short name. Used for
// the one cross-reference this page needs (--platform's AppliesTo), so the verbs
// that take the flag are read from the model instead of spelled out again here.
func abbrevFlagByShort(short string) (cliFlag, bool) {
	for _, f := range cliFlags {
		if f.Short == short {
			return f, true
		}
	}
	return cliFlag{}, false
}

// renderAbbreviationDoc produces docs/abbreviation.md from the model.
// Deterministic and byte-stable so TestAbbreviationDocInSync can gate the
// committed file against it, exactly as renderCommandsDoc is gated.
func renderAbbreviationDoc() string {
	var l []string
	add := func(s string) { l = append(l, s) }

	add("# solmq-conn-util abbreviations")
	add("")
	add("<!-- GENERATED -- do not edit by hand.")
	add("Source of truth: cmd/solmq-conn-util/commands.go (the cliSpec model) and")
	add("cmd/solmq-conn-util/main.go (platformAliasList).")
	add("Regenerate: go test ./cmd/solmq-conn-util -run TestAbbreviationDocInSync -update")
	add("TestAbbreviationDocInSync fails the build if this file drifts from the model. -->")
	add("")
	add("Every short spelling " + bt + "solmq-conn-util" + bt + " accepts, keyed by the abbreviation.")
	add("Each one is recognised wherever its canonical word is -- both when the command")
	add("runs and in shell completion -- but only the canonical word is ever printed by")
	add("terminal help or offered by the TAB menu, so the short forms are documented here")
	add("and in [commands.md](commands.md) rather than in the binary. Generated from the")
	add("command model in")
	add("[" + bt + "cmd/solmq-conn-util/commands.go" + bt + "](../cmd/solmq-conn-util/commands.go); see")
	add("[DEVELOPMENT.md](DEVELOPMENT.md#testing) to regenerate.")
	add("")

	add("## Command abbreviations")
	add("")
	rows := make([]abbrevRow, 0, len(cliVerbs))
	for _, v := range cliVerbs {
		for _, a := range v.Aliases {
			rows = append(rows, abbrevRow{Short: a, Cells: []string{bt + v.Name + bt, verbBlurb(v)}})
		}
	}
	abbrevTable(add, []string{"Short", "Stands for", "What it does"}, rows)

	add("## Target abbreviations")
	add("")
	add("The second (or third) word of a command, after the verb.")
	add("")
	rows = nil
	for _, v := range cliVerbs {
		for _, tg := range v.Targets {
			for _, a := range tg.Aliases {
				rows = append(rows, abbrevRow{Short: a, Cells: []string{bt + tg.Name + bt, bt + v.Name + bt, tg.Summary}})
			}
			// Sets are the third command level (download jar mq|syslog). None of
			// them carries an alias today; walking them anyway means one that is
			// added later lands on this page without a second edit here.
			for _, s := range tg.Sets {
				for _, a := range s.Aliases {
					rows = append(rows, abbrevRow{Short: a, Cells: []string{bt + s.Name + bt, bt + v.Name + " " + tg.Name + bt, s.Summary}})
				}
			}
		}
	}
	abbrevTable(add, []string{"Short", "Stands for", "Under", "Summary"}, rows)

	add("## Platform abbreviations")
	add("")
	if f, ok := abbrevFlagByShort(platformFlagName); ok {
		add("Accepted as a " + platformSpan + " value by " + f.AppliesTo + ", alongside the")
		add("canonical names.")
		add("")
	}
	rows = nil
	for _, e := range platformAliasList {
		rows = append(rows, abbrevRow{Short: e.Alias, Cells: []string{bt + e.Canonical + bt}})
	}
	abbrevTable(add, []string{"Short", "Stands for"}, rows)

	add("## Flag abbreviations")
	add("")
	add("Only flags with a short form appear here; a flag spelled one way only (for")
	add("example " + platformSpan + ") is in the [commands.md flag table](commands.md#flags).")
	add("")
	rows = nil
	for _, f := range cliFlags {
		if f.Short == f.Long {
			continue
		}
		rows = append(rows, abbrevRow{Short: f.Short, Cells: []string{bt + f.Long + bt, f.AppliesTo, f.Meaning}})
	}
	abbrevTable(add, []string{"Short", "Stands for", "Applies to", "Meaning"}, rows)

	add("## Notes")
	add("")
	add("- An abbreviation is accepted wherever its canonical word is, but is never")
	add("  offered as a completion candidate and never shown in terminal help, so each")
	add("  menu and help page keeps exactly one spelling per command.")
	add("- " + bt + "help" + bt + " also answers to " + bt + "-h" + bt + " and " + bt + "--help" + bt + ". Those are flag spellings of the verb,")
	add("  not model aliases, which is why they are not in the table above.")
	add("- The platform short spellings are curated, not a prefix rule: only " + bt + "kubernetes" + bt)
	add("  has a widely recognized short form, and a prefix scheme would silently change")
	add("  meaning the day a platform is added.")

	return strings.TrimRight(strings.Join(l, "\n"), "\n") + "\n"
}
