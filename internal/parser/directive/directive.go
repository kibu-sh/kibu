package directive

import (
	"github.com/pkg/errors"
	"go/ast"
	"strings"
)

var ErrInvalidDirective = errors.New("invalid directive")

// var ErrInvalidOption = errors.New("invalid option")

type Directive struct {
	Tool    string
	Name    string
	Options OptionList
}

type OptionList map[string]string

type List []Directive

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

// Parse parses a directive string.
// A directive is a string literal (because it has the string type)
// Example: devx:endpoint method=GET path=/api/v1/users
//
// Key is an unquoted string literal (devx:endpoint)
// Key contains two parts tool (devx) and name (endpoint) separated by a colon
// Key is required
//
// Value is an unquoted string literal (method=GET path=/api/v1/users)
// Value is optional
// A Value is a list of Options separated by a space
// An Option is a key value pair separated by an equals sign
// An Option key is an unquoted string literal (method)
// An Option value is an unquoted string literal (GET)
func Parse(d string) (dir Directive, err error) {
	if !IsDirective(d) {
		err = errors.Wrapf(ErrInvalidDirective, "%s", d)
		return
	}
	parts := strings.Split(d, " ")
	dir.Tool, dir.Name = parseKey(parts[0])
	dir.Options, err = parseOptions(parts[1:])
	if err != nil {
		return
	}

	return
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

func parseKey(s string) (string, string) {
	parts := strings.Split(s, ":")
	return parts[0], parts[1]
}

func parseOptions(opts []string) (result OptionList, err error) {
	if len(opts) == 0 {
		return
	}

	result = make(OptionList)
	for _, opt := range opts {
		opt = strings.TrimSpace(opt)
		if opt == "" {
			continue
		}

		pair := strings.Split(opt, "=")
		result[pair[0]] = tryIndex(pair, 1)
	}
	return result, nil
}

func tryIndex(pair []string, i int) string {
	if len(pair) > i {
		return pair[i]
	}
	return ""
}
