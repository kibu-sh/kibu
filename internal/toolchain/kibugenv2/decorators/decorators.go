package decorators

import (
	"encoding/gob"
	"encoding/json"
	"go/ast"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var ErrInvalidDirective = errors.New("invalid directive")

type Line struct {
	Tool      string
	Name      string
	Qualifier string
	Options   *OptionList
}

func (d Line) String() string {
	parts := lineParts(d)
	return strings.Join(parts, ":")
}

func lineParts(d Line) []string {
	if d.Qualifier != "" {
		return []string{d.Tool, d.Name, d.Qualifier}
	}
	return []string{d.Tool, d.Name}
}

var _ gob.GobEncoder = (*OptionList)(nil)
var _ gob.GobDecoder = (*OptionList)(nil)

type OptionList struct {
	om map[string][]string
}

// OptionValue holds the result of a single-value option lookup.
type OptionValue struct {
	value string
	found bool
}

// Or returns the looked-up value if found, otherwise the fallback.
func (v OptionValue) Or(fallback string) string {
	if !v.found {
		return fallback
	}
	return v.value
}

// Found reports whether the option key existed.
func (v OptionValue) Found() bool {
	return v.found
}

func (ol *OptionList) GobDecode(bytes []byte) error {
	ol.om = make(map[string][]string)
	return json.Unmarshal(bytes, &ol.om)
}

func (ol *OptionList) GobEncode() ([]byte, error) {
	return json.Marshal(ol.om)
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

// Lookup returns an OptionValue for the given key.
// Safe to call on a nil receiver.
func (ol *OptionList) Lookup(key string) OptionValue {
	if ol == nil {
		return OptionValue{}
	}
	v, ok := ol.om[key]
	if !ok || len(v) == 0 {
		return OptionValue{}
	}
	return OptionValue{value: v[0], found: true}
}

// ListValues returns all option values for the given key.
// If the key is absent, def is returned.
func (ol *OptionList) ListValues(key string, def []string) (val []string, ok bool) {
	if ol == nil {
		return def, false
	}
	if val, ok = ol.om[key]; !ok {
		val = def
	}
	return
}

// Has checks if an option is present by its key.
// It is possible for a key to be present with no value.
func (ol *OptionList) Has(key string) bool {
	if ol == nil {
		return false
	}
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

type List []Line
type FilterFunc func(d Line) bool

func (l List) Filter(filter FilterFunc) List {
	return lo.Filter(l, func(d Line, _ int) bool {
		return filter(d)
	})
}

func (l List) Some(some FilterFunc) bool {
	return lo.SomeBy(l, some)
}

func (l List) Find(predicate FilterFunc) (Line, bool) {
	return lo.Find(l, predicate)
}

func OneOf(filters ...FilterFunc) FilterFunc {
	return func(d Line) bool {
		for _, filter := range filters {
			if filter(d) {
				return true
			}
		}
		return false
	}
}

func HasKey(parts ...string) FilterFunc {
	return Exactly(strings.Join(parts, ":"))
}

func Matches(pattern string) FilterFunc {
	return func(d Line) bool {
		gl, err := glob.Compile(pattern)
		if err != nil {
			return false
		}

		return gl.Match(d.String())
	}
}

func HasQualifier(qualifier string) FilterFunc {
	return func(d Line) bool {
		return d.Qualifier == qualifier
	}
}

func HasPrefix(prefix string) FilterFunc {
	return func(d Line) bool {
		return strings.HasPrefix(d.String(), prefix)
	}
}

func HasSuffix(suffix string) FilterFunc {
	return func(d Line) bool {
		return strings.HasSuffix(d.String(), suffix)
	}
}

func Exactly(s string) FilterFunc {
	return func(d Line) bool {
		return d.String() == s
	}
}

func HasTool(tool string) FilterFunc {
	return func(d Line) bool {
		return d.Tool == tool
	}
}

// FromCommentGroup returns a list of directives by parsing an *ast.CommentGroup.
func FromCommentGroup(d *ast.CommentGroup) (result List, err error) {
	if d == nil {
		return
	}

	for _, comment := range d.List {
		dir, parseErr := parseComment(comment)
		if parseErr != nil {
			return result, parseErr
		}
		if dir != nil {
			result = append(result, *dir)
		}
	}
	return
}

func parseComment(comment *ast.Comment) (*Line, error) {
	if comment.Text[:2] != "//" {
		return nil, nil
	}
	txt := comment.Text[2:]
	if !IsDirective(txt) {
		return nil, nil
	}
	dir, err := Parse(txt)
	if err != nil {
		return nil, err
	}
	return &dir, nil
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
// Example: kibu:endpoint method=GET path=/api/v1/users
//
// Path is an unquoted string literal (kibu:endpoint)
// Path contains two parts tool (kibu) and name (endpoint) separated by a colon
// Path is required
//
// value is an unquoted string literal (method=GET path=/api/v1/users)
// value is optional
// A value is a list of Options separated by a space
// An Option is a key value pair separated by an equals sign
// An Option key is an unquoted string literal (method)
// An Option value is an unquoted string literal (GET)
func Parse(d string) (dir Line, err error) {
	if !IsDirective(d) {
		err = errors.Wrapf(ErrInvalidDirective, "%s", d)
		return
	}

	parts := strings.Split(d, " ")
	dir.Tool, dir.Name, dir.Qualifier, err = parseKey(parts[0])
	if err != nil {
		return
	}

	dir.Options, err = parseOptions(parts[1:])
	if err != nil {
		return
	}

	return
}

func parseKey(s string) (string, string, string, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return "", "", "", errors.Wrapf(ErrInvalidDirective,
			"failed to parse key expected form at (tool:name) got %s", s)
	}
	return parts[0], parts[1], qualifierFromParts(parts), nil
}

func qualifierFromParts(parts []string) string {
	if len(parts) == 3 {
		return parts[2]
	}
	return ""
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
		existing, _ := result.ListValues(pair[0], nil)
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

type Map = orderedmap.OrderedMap[*ast.Ident, List]

func NewMap() *Map {
	return orderedmap.New[*ast.Ident, List]()
}

// FromDecls returns a list of directives cached by *ast.Ident
func FromDecls(decls []ast.Decl) (result *Map, err error) {
	result = orderedmap.New[*ast.Ident, List]()

	for _, decl := range decls {
		if err = ApplyFromDecl(decl, result); err != nil {
			return
		}
	}

	return
}

// ApplyFromDecl applies the comments from a declaration to the result map
func ApplyFromDecl(decl ast.Decl, result *Map) (err error) {
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

	applyDirsToDecl(decl, dirs, result)
	return
}

func applyDirsToDecl(decl ast.Decl, dirs List, result *Map) {
	switch decl := decl.(type) {
	case *ast.GenDecl:
		applyDirsToGenDecl(decl, dirs, result)
	case *ast.FuncDecl:
		result.Set(decl.Name, dirs)
	}
}

func applyDirsToGenDecl(decl *ast.GenDecl, dirs List, result *Map) {
	for _, spec := range decl.Specs {
		applyDirsToSpec(spec, dirs, result)
	}
}

func applyDirsToSpec(spec ast.Spec, dirs List, result *Map) {
	switch spec := spec.(type) {
	case *ast.TypeSpec:
		result.Set(spec.Name, dirs)
	case *ast.ValueSpec:
		for _, name := range spec.Names {
			result.Set(name, dirs)
		}
	}
}
