//go:build unix

package main

import (
	"io/fs"
	"syscall"
)

// fileOwner returns the uid and gid that own a file, where the platform keeps
// them in the stat result.
func fileOwner(info fs.FileInfo) (uid, gid int, ok bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return int(st.Uid), int(st.Gid), true
}
