// Command docgen generates Docusaurus reference pages for every MTProto API
// method exposed by github.com/gotd/td/tg.Client.
//
// It parses the tg package source (doc comments + request structs) and renders
// one MDX page per method, grouped by namespace, into the output directory.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	tgDir := flag.String("tg", "", "path to the tg package source (default: resolve github.com/gotd/td via `go list -m`)")
	out := flag.String("out", "docs/reference", "output directory for generated pages")
	clean := flag.Bool("clean", true, "remove the methods output directory before generating")
	only := flag.String("only", "", "generate only this namespace (for validation)")
	flag.Parse()

	if err := run(*tgDir, *out, *clean, *only); err != nil {
		fmt.Fprintln(os.Stderr, "docgen:", err)
		os.Exit(1)
	}
}

func run(tgDir, out string, clean bool, only string) error {
	if tgDir == "" {
		resolved, err := resolveTGDir()
		if err != nil {
			return fmt.Errorf("locate tg source (pass --tg): %w", err)
		}
		tgDir = resolved
	}
	if _, err := os.Stat(tgDir); err != nil {
		return fmt.Errorf("tg source %q: %w", tgDir, err)
	}

	model, err := Parse(tgDir)
	if err != nil {
		return err
	}
	if len(model.Methods) == 0 {
		return fmt.Errorf("no methods found in %q", tgDir)
	}
	if only != "" {
		filterNamespace(model, only)
	}

	if clean {
		for _, sub := range []string{"methods", "types", "constructors"} {
			if err := os.RemoveAll(filepath.Join(out, sub)); err != nil {
				return err
			}
		}
	}
	return render(model, out)
}

// filterNamespace restricts the model to a single namespace (for validation).
func filterNamespace(m *Result, ns string) {
	methods := m.Methods[:0]
	for _, x := range m.Methods {
		if x.Namespace == ns {
			methods = append(methods, x)
		}
	}
	m.Methods = methods
	types := m.Types[:0]
	for _, x := range m.Types {
		if x.Namespace == ns {
			types = append(types, x)
		}
	}
	m.Types = types
	ctors := m.Constructors[:0]
	for _, x := range m.Constructors {
		if x.Namespace == ns {
			ctors = append(ctors, x)
		}
	}
	m.Constructors = ctors
}

// resolveTGDir locates the tg package via the Go module cache.
func resolveTGDir() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/gotd/td")
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(string(b))
	if dir == "" {
		return "", fmt.Errorf("empty module dir")
	}
	return filepath.Join(dir, "tg"), nil
}
