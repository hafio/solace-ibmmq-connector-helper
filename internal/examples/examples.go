// Package examples embeds a ready-to-edit starter set of spec files (sample
// workflows + defaults.yaml + kubernetes.yaml) that the `solmq-gen examples`
// command writes to disk. They are valid inputs on their own: `solmq-gen config
// <dir>` on a freshly written set generates an application.yml with no errors.
package examples

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

//go:embed files/*.yaml
var files embed.FS

// Write copies every embedded example into dir (created if needed). Existing
// files are left untouched and returned in skipped, unless force is true, in
// which case they are overwritten. Returned paths are sorted.
func Write(dir string, force bool) (written, skipped []string, err error) {
	if err = os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	entries, err := fs.ReadDir(files, "files")
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		dst := filepath.Join(dir, e.Name())
		if !force {
			if _, statErr := os.Stat(dst); statErr == nil {
				skipped = append(skipped, dst)
				continue
			}
		}
		data, rerr := files.ReadFile("files/" + e.Name()) // embed.FS always uses '/'
		if rerr != nil {
			return written, skipped, rerr
		}
		if werr := os.WriteFile(dst, data, 0o644); werr != nil {
			return written, skipped, werr
		}
		written = append(written, dst)
	}
	sort.Strings(written)
	sort.Strings(skipped)
	return written, skipped, nil
}
