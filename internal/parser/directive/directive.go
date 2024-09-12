package directive

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go/ast"
	"strings"
)

var ErrInvalidDirective = errors.New("invalid directive")

// var ErrInvalidOption = errors.New("invalid option")

type Directive struct {
	Tool    string
	Name    string
	Options *OptionList
}

type OptionList struct {
	om map[string][]string
}

func NewOptionList() *OptionList {
	return &OptionList{
		om: make(map[string][]string),
	}
}

func NewOptionListWithDefaults(defaults map[string][]string) *OptionList {
	ol := NewOptionList()
	for k, v := range defaults {
		ol.Set(k, v)
	}
	return ol
}

// Set sets a single option value by its key
func (ol *OptionList) Set(key string, val []string) {
	ol.om[key] = val
}

// GetOne returns a single option value by its key
// If the option does not exist an empty string is returned
func (ol *OptionList) GetOne(key, def string) (val string, ok bool) {
	v, ok := ol.om[key]
	if len(v) == 0 {
		return def, false
	}
	return v[0], true
}

// GetAll returns a list of option values by key
func (ol *OptionList) GetAll(key string, def []string) (val []string, ok bool) {
	if val, ok = ol.om[key]; !ok {
		val = def
	}
	return
}

// Has checks if an option is present by its key
// it is possible for a key to be present with no value
func (ol *OptionList) Has(key string) bool {
	_, ok := ol.om[key]
	return ok
}

func (ol *OptionList) HasOneOf(keys ...string) bool {
	for _, key := range keys {
		if ol.Has(key) {
			return true
		}
	}
	return false
}

type List []Directive
type FilterFunc func(d Directive) bool

func (l List) Filter(filter FilterFunc) List {
	return lo.Filter(l, func(d Directive, _ int) bool {
		return filter(d)
	})
}

func (l List) Some(some FilterFunc) bool {
	return lo.SomeBy(l, some)
}

func (l List) Find(predicate FilterFunc) (Directive, bool) {
	return lo.Find(l, predicate)
}

func OneOf(filters ...FilterFunc) FilterFunc {
	return func(d Directive) bool {
		for _, filter := range filters {
			if filter(d) {
				return true
			}
		}
		return false
	}
}

func HasKey(tool, name string) FilterFunc {
	return func(d Directive) bool {
		return d.Tool == tool && d.Name == name
	}
}

// FromCommentGroup returns a list of directives by parsing an *ast.CommentGroup.
func FromCommentGroup(d *ast.CommentGroup) (result List, err error) {
	for _, comment := range d.List {
		if comment.Text[:2] == "//" {
			txt := comment.Text[2:]
			if IsDirective(txt) {
				var dir Directive
				dir, err = Parse(txt)
				if err != nil {
					return
				}
				result = append(result, dir)
			}
		}
	}
	return
}

// IsDirective reports whether c is a comment directive.
// This code is also in go/printer.
// Copied from private go/ast/ast.go IsDirective
func IsDirective(c string) bool {
	// "//line " is a line directive.
	// "//extern " is for gccgo.
	// "//export " is for cgo.
	// (The // has been removed.)
	if strings.HasPrefix(c, "line ") || strings.HasPrefix(c, "extern ") || strings.HasPrefix(c, "export ") {
		return true
	}

	// "//[a-z0-9]+:[a-z0-9]"
	// (The // has been removed.)
	colon := strings.Index(c, ":")
	if colon <= 0 || colon+1 >= len(c) {
		return false
	}
	for i := 0; i <= colon+1; i++ {
		if i == colon {
			continue
		}
		b := c[i]
		if !('a' <= b && b <= 'z' || '0' <= b && b <= '9') {
			return false
		}
	}
	return true
}

// Parse extracts data from a directive string.
//
// Example: kibue:endpoint method=GET path=/api/v1/users
//
// Path is an unquoted string literal (kibue:endpoint)
// Path contains two parts tool (kibue) and name (endpoint) separated by a colon
// Path is required
//
// value is an unquoted string literal (method=GET path=/api/v1/users)
// value is optional
// A value is a list of Options separated by a space
// An Option is a key value pair separated by an equals sign
// An Option key is an unquoted string literal (method)
// An Option value is an unquoted string literal (GET)
func Parse(d string) (dir Directive, err error) {
	if !IsDirective(d) {
		err = errors.Wrapf(ErrInvalidDirective, "%s", d)
		return
	}

	parts := strings.Split(d, " ")
	dir.Tool, dir.Name, err = parseKey(parts[0])
	if err != nil {
		return
	}

	dir.Options, err = parseOptions(parts[1:])
	if err != nil {
		return
	}

	return
}

func parseKey(s string) (string, string, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", errors.Wrapf(ErrInvalidDirective,
			"failed to parse key expected form at (tool:name) got %s", s)
	}
	return parts[0], parts[1], nil
}

func parseOptions(opts []string) (result *OptionList, err error) {
	result = NewOptionList()
	if len(opts) == 0 {
		return
	}

	for _, opt := range opts {
		// clean up any leading or trailing spaces
		opt = strings.TrimSpace(opt)

		// ignore spaces between options (e.g. "key1=value1     key2=value2")
		if opt == "" {
			continue
		}

		pair := strings.Split(opt, "=")
		existing, _ := result.GetAll(pair[0], nil)
		result.Set(pair[0], append(existing, tryIndex(pair, 1)...))
	}
	return result, nil
}

func tryIndex(pair []string, i int) []string {
	if len(pair) > i {
		return strings.Split(pair[i], ",")
	}
	return nil
}

// FromDecls returns a list of directives cached by *ast.Ident
func FromDecls(decls []ast.Decl) (result map[*ast.Ident]List, err error) {
	result = make(map[*ast.Ident]List)

	for _, decl := range decls {
		if err = applyFromDecl(decl, result); err != nil {
			return
		}
	}

	return
}

func applyFromDecl(decl ast.Decl, result map[*ast.Ident]List) (err error) {
	var comments *ast.CommentGroup

	switch decl := decl.(type) {
	case *ast.GenDecl:
		comments = decl.Doc
	case *ast.FuncDecl:
		comments = decl.Doc
	}

	if comments == nil {
		return
	}

	dirs, err := FromCommentGroup(comments)
	if err != nil {
		return
	}

	switch decl := decl.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				result[spec.Name] = dirs
			case *ast.ValueSpec:
				for _, name := range spec.Names {
					result[name] = dirs
				}
			}
		}
	case *ast.FuncDecl:
		result[decl.Name] = dirs
	}
	return
}
