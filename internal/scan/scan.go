// Package scan walks a spec folder and separates the reserved files
// (defaults.yaml, the kubernetes settings file, and the -o output file if it
// lives inside the folder) from the workflow files, which are returned sorted
// by base name so workflow IDs are assigned deterministically.
package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultsName is the reserved shared-globals file.
const DefaultsName = "defaults.yaml"

// Result is the outcome of scanning a spec folder.
type Result struct {
	Dir            string
	WorkflowFiles  []string // full paths, sorted by base name
	DefaultsPath   string   // "" if absent
	KubernetesPath string   // "" if absent
}

// Scan reads dir (non-recursive) and classifies its *.yaml/*.yml files.
// kubeFile is the base name of the Kubernetes settings file (default
// "kubernetes.yaml"); outFile is the -o target (may be "" for stdout) and is
// excluded from the workflow scan when it resolves inside dir.
func Scan(dir, kubeFile, outFile string) (*Result, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading spec folder: %w", err)
	}
	if kubeFile == "" {
		kubeFile = "kubernetes.yaml"
	}
	kubeBase := filepath.Base(kubeFile)

	// Determine the -o base name to exclude, only if it sits inside dir.
	outBase := ""
	if outFile != "" {
		absDir, _ := filepath.Abs(dir)
		absOut, _ := filepath.Abs(outFile)
		if filepath.Dir(absOut) == absDir {
			outBase = filepath.Base(absOut)
		}
	}

	res := &Result{Dir: dir}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !isYAML(name) {
			continue
		}
		full := filepath.Join(dir, name)
		switch name {
		case DefaultsName:
			res.DefaultsPath = full
			continue
		case kubeBase:
			res.KubernetesPath = full
			continue
		}
		if outBase != "" && name == outBase {
			continue
		}
		res.WorkflowFiles = append(res.WorkflowFiles, full)
	}
	sort.Slice(res.WorkflowFiles, func(i, j int) bool {
		return filepath.Base(res.WorkflowFiles[i]) < filepath.Base(res.WorkflowFiles[j])
	})
	return res, nil
}

func isYAML(name string) bool {
	l := strings.ToLower(name)
	return strings.HasSuffix(l, ".yaml") || strings.HasSuffix(l, ".yml")
}
