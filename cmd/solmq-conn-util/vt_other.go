//go:build !windows

package main

import "os"

// enableVirtualTerminal is a no-op everywhere but Windows: every terminal this
// tool runs in elsewhere processes ANSI escapes without being asked. The
// function exists so its callers -- the status watch loop, and cli before it
// hands over the terminal -- have one unconditional call site rather than a
// build-tagged branch each.
func enableVirtualTerminal(*os.File) {}
