package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templates embed.FS

func render(model *Result, outDir string) error {
	tmpl := template.New("").Funcs(template.FuncMap{
		"mdxText":    mdxText,
		"mdxCell":    mdxCell,
		"yamlString": yamlString,
	})
	tmpl, err := tmpl.ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	namespaces := groupMethods(model.Methods)

	// Overview index (doubles as the "API Reference" category landing page).
	if err := writeTemplate(tmpl, "index.md.tmpl", filepath.Join(outDir, "index.md"), map[string]any{
		"Namespaces":   namespaces,
		"Methods":      len(model.Methods),
		"Types":        len(model.Types),
		"Constructors": len(model.Constructors),
	}); err != nil {
		return err
	}
	if err := writeCategoryDoc(filepath.Join(outDir, "_category_.json"), "API Reference", 7, "reference/index"); err != nil {
		return err
	}

	if err := renderMethods(tmpl, namespaces, len(model.Methods), outDir); err != nil {
		return err
	}
	if err := renderTypes(tmpl, model.Types, outDir); err != nil {
		return err
	}
	if err := renderConstructors(tmpl, model.Constructors, outDir); err != nil {
		return err
	}

	fmt.Printf("generated %d methods, %d types, %d constructors into %s\n",
		len(model.Methods), len(model.Types), len(model.Constructors), outDir)
	return nil
}

func groupMethods(methods []*Method) []*Namespace {
	byNS := map[string]*Namespace{}
	for _, m := range methods {
		ns := byNS[m.Namespace]
		if ns == nil {
			ns = &Namespace{Name: m.Namespace, Description: namespaceDescriptions[m.Namespace]}
			byNS[m.Namespace] = ns
		}
		ns.Methods = append(ns.Methods, m)
	}
	var out []*Namespace
	for _, ns := range byNS {
		sort.Slice(ns.Methods, func(i, j int) bool { return ns.Methods[i].Short < ns.Methods[j].Short })
		out = append(out, ns)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func renderMethods(tmpl *template.Template, namespaces []*Namespace, total int, outDir string) error {
	dir := filepath.Join(outDir, "methods")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeCategory(filepath.Join(dir, "_category_.json"), "Methods", 2,
		"/reference/methods", fmt.Sprintf("All %d RPC methods, grouped by namespace.", total)); err != nil {
		return err
	}
	for i, ns := range namespaces {
		nsDir := filepath.Join(dir, ns.Name)
		if err := os.MkdirAll(nsDir, 0o755); err != nil {
			return err
		}
		if err := writeCategory(filepath.Join(nsDir, "_category_.json"), ns.Name, i+1,
			"/reference/methods/"+ns.Name,
			fmt.Sprintf("%d methods in the %s namespace.", len(ns.Methods), ns.Name)); err != nil {
			return err
		}
		for _, m := range ns.Methods {
			if err := writeTemplate(tmpl, "method.mdx.tmpl", filepath.Join(nsDir, m.Slug+".mdx"), m); err != nil {
				return fmt.Errorf("render method %s: %w", m.TLName, err)
			}
		}
	}
	return nil
}

func renderTypes(tmpl *template.Template, types []*Type, outDir string) error {
	dir := filepath.Join(outDir, "types")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeCategory(filepath.Join(dir, "_category_.json"), "Types", 3,
		"/reference/types", fmt.Sprintf("All %d TL classes (types), grouped by namespace.", len(types))); err != nil {
		return err
	}
	byNS := map[string][]*Type{}
	for _, t := range types {
		byNS[t.Namespace] = append(byNS[t.Namespace], t)
	}
	return renderNamespaced(tmpl, "type.mdx.tmpl", dir, "/reference/types", "types", byNS, func(t *Type) (string, string) {
		return t.Slug, t.TLName
	})
}

func renderConstructors(tmpl *template.Template, ctors []*Constructor, outDir string) error {
	dir := filepath.Join(outDir, "constructors")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeCategory(filepath.Join(dir, "_category_.json"), "Constructors", 4,
		"/reference/constructors", fmt.Sprintf("All %d TL constructors, grouped by namespace.", len(ctors))); err != nil {
		return err
	}
	byNS := map[string][]*Constructor{}
	for _, c := range ctors {
		byNS[c.Namespace] = append(byNS[c.Namespace], c)
	}
	return renderNamespaced(tmpl, "constructor.mdx.tmpl", dir, "/reference/constructors", "constructors", byNS, func(c *Constructor) (string, string) {
		return c.Slug, c.TLName
	})
}

// renderNamespaced renders a map of namespace -> items into per-namespace dirs.
func renderNamespaced[T any](tmpl *template.Template, page, dir, slugBase, kind string, byNS map[string][]T, meta func(T) (slug, tlName string)) error {
	var names []string
	for ns := range byNS {
		names = append(names, ns)
	}
	sort.Strings(names)
	for i, ns := range names {
		nsDir := filepath.Join(dir, ns)
		if err := os.MkdirAll(nsDir, 0o755); err != nil {
			return err
		}
		items := byNS[ns]
		if err := writeCategory(filepath.Join(nsDir, "_category_.json"), ns, i+1,
			slugBase+"/"+ns, fmt.Sprintf("%d %s in the %s namespace.", len(items), kind, ns)); err != nil {
			return err
		}
		for _, item := range items {
			slug, tlName := meta(item)
			if err := writeTemplate(tmpl, page, filepath.Join(nsDir, slug+".mdx"), item); err != nil {
				return fmt.Errorf("render %s %s: %w", kind, tlName, err)
			}
		}
	}
	return nil
}

func writeTemplate(tmpl *template.Template, name, path string, data any) error {
	var b strings.Builder
	if err := tmpl.ExecuteTemplate(&b, name, data); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeCategory(path, label string, position int, slug, description string) error {
	type link struct {
		Type        string `json:"type"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	cat := struct {
		Label    string `json:"label"`
		Position int    `json:"position"`
		Link     link   `json:"link"`
	}{Label: label, Position: position, Link: link{Type: "generated-index", Slug: slug, Description: description}}
	return writeJSON(path, cat)
}

func writeCategoryDoc(path, label string, position int, docID string) error {
	type link struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	}
	cat := struct {
		Label    string `json:"label"`
		Position int    `json:"position"`
		Link     link   `json:"link"`
	}{Label: label, Position: position, Link: link{Type: "doc", ID: docID}}
	return writeJSON(path, cat)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
