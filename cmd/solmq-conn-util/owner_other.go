//go:build !unix

package main

import "io/fs"

// fileOwner reports no owner off Unix: Windows file modes carry no uid or gid,
// and unreadableMounts checks nothing there anyway.
func fileOwner(fs.FileInfo) (uid, gid int, ok bool) { return 0, 0, false }
