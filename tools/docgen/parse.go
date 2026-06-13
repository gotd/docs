package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// refBase is the golds-generated reference for the tg package. Symbol
// declarations are anchored as #name-<GoSymbol>.
const refBase = "https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html"

// structInfo holds a parsed generated struct.
type structInfo struct {
	name     string
	doc      string
	hasFlags bool
	fields   []*ast.Field
}

// Parse reads the tg package source at dir and returns its methods, types and
// constructors.
func Parse(dir string) (*Result, error) {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, "_gen.go") {
			continue
		}
		// Slice helper files contain no methods, request structs or constructors.
		if strings.HasSuffix(name, "_slices_gen.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	structs := map[string]*structInfo{} // Go name -> struct
	interfaces := map[string]string{}   // Go name (XClass) -> doc text
	ctorClass := map[string]string{}    // ctor Go name -> class Go name
	var clientFns []*ast.FuncDecl

	for _, f := range files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name == "construct" {
					if ctor, class := constructMapping(d); ctor != "" {
						ctorClass[ctor] = class
					}
					continue
				}
				if d.Doc != nil && isClientMethod(d) {
					clientFns = append(clientFns, d)
				}
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					switch t := ts.Type.(type) {
					case *ast.StructType:
						structs[ts.Name.Name] = &structInfo{
							name:   ts.Name.Name,
							doc:    docText(ts, d),
							fields: nonFlagFields(t),
							hasFlags: hasFlagsField(t),
						}
					case *ast.InterfaceType:
						if strings.HasSuffix(ts.Name.Name, "Class") {
							interfaces[ts.Name.Name] = docText(ts, d)
						}
					}
				}
			}
		}
	}

	res := &Result{}

	// Methods.
	for _, fn := range clientFns {
		if m := parseMethod(fset, fn, structs); m != nil {
			res.Methods = append(res.Methods, m)
		}
	}
	sort.Slice(res.Methods, func(i, j int) bool { return res.Methods[i].TLName < res.Methods[j].TLName })

	// Constructors: structs documented as "represents TL type" that are not
	// method requests (which are documented on their method pages).
	for _, si := range structs {
		if strings.HasSuffix(si.name, "Request") {
			continue
		}
		d := parseTypeDoc(si.doc)
		if d.tlName == "" {
			continue
		}
		c := &Constructor{
			GoName:  si.name,
			TLName:  d.tlName,
			Hash:    d.hash,
			Summary: d.summary,
			DocURL:  d.url,
			RefURL:  refBase + "#name-" + si.name,
			Fields:  params(fset, si),
		}
		if c.DocURL == "" {
			c.DocURL = "https://core.telegram.org/constructor/" + c.TLName
		}
		c.Namespace, c.Short = splitTLName(c.TLName)
		c.Slug = kebab(c.Short)
		res.Constructors = append(res.Constructors, c)
	}

	// Types (classes), with their constructors resolved from ctorClass.
	classCtors := map[string][]*Constructor{}
	ctorByGoName := map[string]*Constructor{}
	for _, c := range res.Constructors {
		ctorByGoName[c.GoName] = c
		if class := ctorClass[c.GoName]; class != "" {
			classCtors[class] = append(classCtors[class], c)
		}
	}
	for goName, doc := range interfaces {
		d := parseClassDoc(doc)
		if d.tlName == "" {
			continue
		}
		t := &Type{GoName: goName, TLName: d.tlName, Summary: d.summary, DocURL: d.url, RefURL: refBase + "#name-" + goName}
		if t.DocURL == "" {
			t.DocURL = "https://core.telegram.org/type/" + t.TLName
		}
		t.Namespace, t.Short = splitTLName(t.TLName)
		t.Slug = kebab(t.Short)
		ctors := classCtors[goName]
		sort.Slice(ctors, func(i, j int) bool { return ctors[i].TLName < ctors[j].TLName })
		for _, c := range ctors {
			t.Constructors = append(t.Constructors, &Ref{GoName: c.GoName, TLName: c.TLName, Namespace: c.Namespace, Slug: c.Slug})
		}
		res.Types = append(res.Types, t)
	}
	sort.Slice(res.Types, func(i, j int) bool { return res.Types[i].TLName < res.Types[j].TLName })

	// Resolve each constructor's Implements ref now that types exist.
	typeByGoName := map[string]*Type{}
	for _, t := range res.Types {
		typeByGoName[t.GoName] = t
	}
	for _, c := range res.Constructors {
		if class := ctorClass[c.GoName]; class != "" {
			if t := typeByGoName[class]; t != nil {
				c.Implements = &Ref{GoName: t.GoName, TLName: t.TLName, Namespace: t.Namespace, Slug: t.Slug}
			}
		}
	}
	sort.Slice(res.Constructors, func(i, j int) bool { return res.Constructors[i].TLName < res.Constructors[j].TLName })

	linkTypes(res)
	return res, nil
}

// linkTypes renders Markdown for every parameter, field and return type
// (linking Go types that have a reference page) and builds the reverse
// "returned by" index on types and constructors.
func linkTypes(res *Result) {
	reg := map[string]*Ref{}
	types := map[string]*Type{}
	ctors := map[string]*Constructor{}
	for _, t := range res.Types {
		reg[t.GoName] = &Ref{Kind: "types", Namespace: t.Namespace, Slug: t.Slug}
		types[t.GoName] = t
	}
	for _, c := range res.Constructors {
		reg[c.GoName] = &Ref{Kind: "constructors", Namespace: c.Namespace, Slug: c.Slug}
		ctors[c.GoName] = c
	}

	for _, m := range res.Methods {
		for i := range m.Params {
			m.Params[i].TypeMD = typeMarkdown(m.Params[i].GoType, reg)
		}
		m.ReturnTypeMD = typeMarkdown(m.ReturnType, reg)

		// Reverse index: record the method on the type/constructor it returns.
		ref := &Ref{Kind: "methods", TLName: m.TLName, Namespace: m.Namespace, Slug: m.Slug}
		switch base := baseType(m.ReturnType); {
		case types[base] != nil:
			types[base].ReturnedBy = append(types[base].ReturnedBy, ref)
		case ctors[base] != nil:
			ctors[base].ReturnedBy = append(ctors[base].ReturnedBy, ref)
		}
	}
	for _, c := range res.Constructors {
		for i := range c.Fields {
			c.Fields[i].TypeMD = typeMarkdown(c.Fields[i].GoType, reg)
		}
	}
}

// baseType strips leading slice and pointer markers from a Go type.
func baseType(goType string) string {
	for {
		switch {
		case strings.HasPrefix(goType, "[]"):
			goType = goType[2:]
		case strings.HasPrefix(goType, "*"):
			goType = goType[1:]
		default:
			return goType
		}
	}
}

// typeMarkdown renders a Go type as Markdown, linking the base identifier to
// its reference page when one exists (slices/pointers are unwrapped).
func typeMarkdown(goType string, reg map[string]*Ref) string {
	if goType == "" {
		return ""
	}
	if ref, ok := reg[baseType(goType)]; ok {
		return "[`" + goType + "`](/docs/reference/" + ref.Kind + "/" + ref.Namespace + "/" + ref.Slug + ")"
	}
	return "`" + goType + "`"
}

// docText returns the doc comment attached to a type spec (either on the spec
// itself or, for single-spec declarations, on the GenDecl).
func docText(ts *ast.TypeSpec, gd *ast.GenDecl) string {
	if ts.Doc != nil {
		return ts.Doc.Text()
	}
	if gd.Doc != nil {
		return gd.Doc.Text()
	}
	return ""
}

func nonFlagFields(st *ast.StructType) []*ast.Field {
	var out []*ast.Field
	for _, field := range st.Fields.List {
		if len(field.Names) == 1 && field.Names[0].Name == "Flags" {
			continue
		}
		out = append(out, field)
	}
	return out
}

func hasFlagsField(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if len(field.Names) == 1 && field.Names[0].Name == "Flags" {
			return true
		}
	}
	return false
}

// constructMapping returns (ctorGoName, classGoName) for a `construct() XClass` method.
func constructMapping(fn *ast.FuncDecl) (ctor, class string) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return "", ""
	}
	ctor = recvTypeName(fn.Recv.List[0].Type)
	if id, ok := fn.Type.Results.List[0].Type.(*ast.Ident); ok && strings.HasSuffix(id.Name, "Class") {
		class = id.Name
	}
	return ctor, class
}

func recvTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// isClientMethod reports whether fn is a method on *Client.
func isClientMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false
	}
	star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	id, ok := star.X.(*ast.Ident)
	return ok && id.Name == "Client"
}

func parseMethod(fset *token.FileSet, fn *ast.FuncDecl, structs map[string]*structInfo) *Method {
	doc := parseDoc(fn.Doc.Text())
	if doc.tlName == "" {
		return nil // not an RPC method (e.g. Invoker())
	}

	m := &Method{
		GoName:    fn.Name.Name,
		TLName:    doc.tlName,
		Hash:      doc.hash,
		Summary:   doc.summary,
		DocURL:    doc.url,
		Errors:    doc.errors,
		Signature: renderSignature(fset, fn),
		// Methods have no own anchor (folded under Client); link the request
		// struct, which carries the parameters and TL documentation.
		RefURL: refBase + "#name-" + fn.Name.Name + "Request",
	}
	if m.DocURL == "" {
		m.DocURL = "https://core.telegram.org/method/" + m.TLName
	}
	m.Namespace, m.Short = splitTLName(m.TLName)
	m.Slug = kebab(m.Short)
	m.ReturnType = renderReturn(fset, fn)
	if si := structs[fn.Name.Name+"Request"]; si != nil {
		m.Params = params(fset, si)
	}
	m.Usage = buildUsage(fn, m.ReturnType)
	return m
}

// params extracts the documented fields from a request or constructor struct.
func params(fset *token.FileSet, si *structInfo) []Param {
	var out []Param
	for _, field := range si.fields {
		if len(field.Names) == 0 {
			continue
		}
		goType := renderExpr(fset, field.Type)
		var docText string
		if field.Doc != nil {
			docText = field.Doc.Text()
		}
		optional := hasHelperNote(docText) || (si.hasFlags && goType == "bool")
		for _, name := range field.Names {
			out = append(out, Param{
				Name:        name.Name,
				GoType:      goType,
				Optional:    optional,
				Description: cleanDescription(docText),
			})
		}
	}
	return out
}

// buildUsage renders a runnable snippet showing how to call the method.
func buildUsage(fn *ast.FuncDecl, returnType string) string {
	// Collect call arguments (everything after ctx).
	var args []string
	if fn.Type.Params != nil {
		for i, p := range fn.Type.Params.List {
			if i == 0 {
				continue // ctx
			}
			// Request-shaped methods take a single *XRequest.
			if star, ok := p.Type.(*ast.StarExpr); ok {
				if id, ok := star.X.(*ast.Ident); ok && strings.HasSuffix(id.Name, "Request") {
					args = append(args, "&tg."+id.Name+"{\n\t// see Parameters\n}")
					continue
				}
			}
			for _, n := range p.Names {
				args = append(args, n.Name)
			}
		}
	}

	call := "api." + fn.Name.Name + "(ctx"
	if len(args) > 0 {
		call += ", " + strings.Join(args, ", ")
	}
	call += ")"

	var lines []string
	lines = append(lines, "api := client.API()", "")
	if returnType != "" {
		lines = append(lines, "res, err := "+call, "if err != nil {", "\treturn err", "}", "_ = res // "+returnType)
	} else {
		lines = append(lines, "if _, err := "+call+"; err != nil {", "\treturn err", "}")
	}

	// Indent every line by one tab so the block nests inside the Run closure.
	out := strings.Join(lines, "\n")
	return strings.ReplaceAll(out, "\n", "\n\t")
}

// renderSignature renders the user-facing Go signature of an RPC method.
func renderSignature(fset *token.FileSet, fn *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString("func (c *Client) ")
	b.WriteString(fn.Name.Name)
	b.WriteString("(")
	if fn.Type.Params != nil {
		var parts []string
		for _, p := range fn.Type.Params.List {
			t := renderExpr(fset, p.Type)
			if len(p.Names) == 0 {
				parts = append(parts, t)
				continue
			}
			var names []string
			for _, n := range p.Names {
				names = append(names, n.Name)
			}
			parts = append(parts, strings.Join(names, ", ")+" "+t)
		}
		b.WriteString(strings.Join(parts, ", "))
	}
	b.WriteString(") ")
	if fn.Type.Results != nil {
		var res []string
		for _, r := range fn.Type.Results.List {
			res = append(res, renderExpr(fset, r.Type))
		}
		if len(res) == 1 {
			b.WriteString(res[0])
		} else {
			b.WriteString("(" + strings.Join(res, ", ") + ")")
		}
	}
	return b.String()
}

// renderReturn returns the first (non-error) result type, if any.
func renderReturn(fset *token.FileSet, fn *ast.FuncDecl) string {
	if fn.Type.Results == nil {
		return ""
	}
	var res []string
	for _, r := range fn.Type.Results.List {
		res = append(res, renderExpr(fset, r.Type))
	}
	if n := len(res); n > 0 && res[n-1] == "error" {
		res = res[:n-1]
	}
	if len(res) == 0 {
		return ""
	}
	return strings.Join(res, ", ")
}

func renderExpr(fset *token.FileSet, e ast.Expr) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, e)
	return buf.String()
}

func splitTLName(tl string) (namespace, short string) {
	if i := strings.IndexByte(tl, '.'); i >= 0 {
		return tl[:i], tl[i+1:]
	}
	return "general", tl
}
