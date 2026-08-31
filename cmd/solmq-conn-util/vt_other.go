//go:build !windows

package main

import "os"

// enableVirtualTerminal is a no-op everywhere but Windows: every terminal this
// tool runs in elsewhere processes ANSI escapes without being asked. The
// function exists so the watch loop has one unconditional call site rather than
// a build-tagged branch of its own.
func enableVirtualTerminal(*os.File) {}
