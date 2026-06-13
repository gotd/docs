# docgen

Generates the Docusaurus API reference under [`docs/reference`](../../docs/reference)
from the [`github.com/gotd/td/tg`](https://github.com/gotd/td) source — one page per
MTProto RPC method, grouped by namespace.

## How it works

`docgen` parses the `tg` package source with `go/parser` (no compilation): it reads each
`func (c *Client) X(...)` method, its doc comment (summary, documented errors, the
`name#hash` TL reference and the `core.telegram.org` link) and the matching `XRequest`
struct (parameter names, Go types, descriptions, and whether a field is optional). It
then renders one MDX page per method plus per-namespace category pages.

The tool is **stdlib-only** and does not import `td`; it only reads its source as data.

## Regenerating

```bash
npm run gen:reference
```

This runs [`generate.sh`](./generate.sh), which uses `github.com/gotd/td` at the pinned
`TD_REF` (clone cached under `.cache/`, gitignored). To generate from a local checkout:

```bash
TD_DIR=/path/to/td npm run gen:reference
```

The generated `docs/reference/**` is committed. A CI job regenerates and runs
`git diff --exit-code` to ensure the committed output stays in sync; bump `TD_REF` in
`generate.sh` (and the workflow) when updating to a newer `td`.

## Layout

| File | Responsibility |
| ---- | -------------- |
| `parse.go` | AST pass: collect methods and request structs |
| `doccomment.go` | Parse doc comments; clean/escape descriptions |
| `mdx.go` | MDX- and table-safe escaping |
| `render.go` | Group by namespace, render templates, write files |
| `model.go` | `Method`, `Param`, `ErrorRow`, `Namespace` types |
| `templates/` | `method.mdx.tmpl`, `index.md.tmpl` |

## Scope

Methods only. TL types/constructors and cross-links between them are a future phase; for
the full Go reference see [ref.gotd.dev](https://ref.gotd.dev/pkg/github.com/gotd/td/tg.html).
