// Package scan walks the workflows folder (workflows.dir from env.yaml) and
// returns the *.yaml/*.yml files whose base name matches file_pattern, sorted
// in natural numeric order by base name (spec.WorkflowFileLess: 2.yaml before
// 10.yaml) so workflow IDs are assigned deterministically and in the order an
// operator numbered the files. The env.yaml file itself is always excluded,
// regardless of dir/pattern.
package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
)

// Result is the outcome of scanning the workflows folder.
type Result struct {
	Dir           string
	WorkflowFiles []string // full paths, sorted in natural order by base name
}

// Scan reads dir (non-recursive) and returns its *.yaml/*.yml files whose base
// name matches filePattern (only the '*' wildcard is honoured; leading, middle,
// and trailing). envAbsPath, when non-empty, is the absolute path of the env
// file and is always excluded from the workflow set even if it matches.
func Scan(dir, filePattern, envAbsPath string) (*Result, error) {
	if filePattern == "" {
		filePattern = "*"
	}
	if err := validatePattern(filePattern); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading workflows folder: %w", err)
	}

	res := &Result{Dir: dir}
	for _, e := range entries {
		if full, ok := workflowPath(dir, e.Name(), e.IsDir(), filePattern, envAbsPath); ok {
			res.WorkflowFiles = append(res.WorkflowFiles, full)
		}
	}
	sort.Slice(res.WorkflowFiles, func(i, j int) bool {
		return spec.WorkflowFileLess(filepath.Base(res.WorkflowFiles[i]), filepath.Base(res.WorkflowFiles[j]))
	})
	return res, nil
}

// workflowPath decides whether one directory entry is a workflow file and, if
// so, returns its full path. It filters out directories, non-YAML names, the
// env file, and names that do not match the pattern.
func workflowPath(dir, name string, isDir bool, filePattern, envAbsPath string) (string, bool) {
	if isDir || !isYAML(name) || !matchStar(filePattern, name) {
		return "", false
	}
	full := filepath.Join(dir, name)
	if envAbsPath != "" {
		if abs, err := filepath.Abs(full); err == nil && sameFile(abs, envAbsPath) {
			return "", false // never treat the config file as a workflow
		}
	}
	return full, true
}

// validatePattern enforces the "leading/middle/trailing '*' only" rule: every
// other glob metacharacter is rejected up front with an actionable error.
func validatePattern(p string) error {
	for _, r := range p {
		switch r {
		case '?', '[', ']', '\\':
			return fmt.Errorf("invalid file_pattern %q: only the '*' wildcard is supported (found %q)", p, string(r))
		}
	}
	return nil
}

// matchStar matches name against a pattern whose only wildcard is '*' (matching
// any run of characters, including empty). Splitting on '*' turns the match into
// an ordered prefix/infix/suffix walk, which naturally supports leading, middle,
// trailing, and repeated stars.
func matchStar(pattern, name string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == name // no wildcard -> exact match
	}
	if !strings.HasPrefix(name, parts[0]) {
		return false
	}
	name = name[len(parts[0]):]
	for _, mid := range parts[1 : len(parts)-1] {
		i := strings.Index(name, mid)
		if i < 0 {
			return false
		}
		name = name[i+len(mid):]
	}
	return strings.HasSuffix(name, parts[len(parts)-1])
}

// sameFile compares two absolute paths for identity, case-insensitively on
// Windows where the filesystem is case-preserving but case-insensitive.
func sameFile(a, b string) bool {
	if filepath.Separator == '\\' {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func isYAML(name string) bool {
	l := strings.ToLower(name)
	return strings.HasSuffix(l, ".yaml") || strings.HasSuffix(l, ".yml")
}
