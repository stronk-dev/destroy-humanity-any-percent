// Command verify-copy-sites proves that each declared code-owned copy key is
// the value of its declared JSON field inside its declared Go producer function.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	errInvalidRegistry = errors.New("invalid copy code-reference registry")
	keyPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	functionPattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	fieldPattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

type registry struct {
	SchemaVersion int       `json:"schema_version"`
	References    []binding `json:"references"`
}

type binding struct {
	GoFunction string `json:"go_function"`
	JSONField  string `json:"json_field"`
	Key        string `json:"key"`
	SourceFile string `json:"source_file"`
}

func decodeRegistry(data []byte) (registry, error) {
	var value registry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return registry{}, errors.Join(errInvalidRegistry, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return registry{}, fmt.Errorf("%w: exactly one JSON value required", errInvalidRegistry)
	}
	if value.SchemaVersion != 1 || len(value.References) == 0 {
		return registry{}, errInvalidRegistry
	}
	previous := ""
	for _, item := range value.References {
		if !keyPattern.MatchString(item.Key) || item.Key <= previous ||
			!functionPattern.MatchString(item.GoFunction) || !fieldPattern.MatchString(item.JSONField) ||
			!strings.HasPrefix(item.SourceFile, "server/") || filepath.Clean(item.SourceFile) != item.SourceFile || filepath.Ext(item.SourceFile) != ".go" {
			return registry{}, fmt.Errorf("%w: invalid binding for %q", errInvalidRegistry, item.Key)
		}
		previous = item.Key
	}
	return value, nil
}

func literalString(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func verifySource(filename string, source []byte, item binding) error {
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, source, 0)
	if err != nil {
		return err
	}
	functions, matches := 0, 0
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != item.GoFunction {
			continue
		}
		functions++
		ast.Inspect(function.Body, func(node ast.Node) bool {
			pair, ok := node.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			field, fieldOK := literalString(pair.Key)
			value, valueOK := literalString(pair.Value)
			if fieldOK && valueOK && field == item.JSONField && value == item.Key {
				matches++
			}
			return true
		})
	}
	if functions != 1 || matches != 1 {
		return fmt.Errorf("%w: %s requires exactly one %s[%q]=%q binding; functions=%d matches=%d", errInvalidRegistry, filename, item.GoFunction, item.JSONField, item.Key, functions, matches)
	}
	return nil
}

func run(root string) error {
	if root == "" {
		return errInvalidRegistry
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(root, "copy", "code-reference-sites.v1.json"))
	if err != nil {
		return err
	}
	value, err := decodeRegistry(data)
	if err != nil {
		return err
	}
	for _, item := range value.References {
		filename := filepath.Join(root, filepath.FromSlash(item.SourceFile))
		if !strings.HasPrefix(filename, root+string(filepath.Separator)) {
			return errInvalidRegistry
		}
		source, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := verifySource(filename, source, item); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	if err := run(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("copy code-reference producer sites ok")
}
