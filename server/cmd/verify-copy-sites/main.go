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

func serializedBindingVariable(statement ast.Stmt, item binding) (string, bool) {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || len(assignment.Lhs) < 1 || len(assignment.Rhs) != 1 {
		return "", false
	}
	variable, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok || variable.Name == "_" {
		return "", false
	}
	call, ok := assignment.Rhs[0].(*ast.CallExpr)
	if !ok {
		return "", false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Marshal" || len(call.Args) != 1 {
		return "", false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || packageName.Name != "json" {
		return "", false
	}
	literal, ok := call.Args[0].(*ast.CompositeLit)
	if !ok {
		return "", false
	}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		field, fieldOK := literalString(pair.Key)
		value, valueOK := literalString(pair.Value)
		if fieldOK && valueOK && field == item.JSONField && value == item.Key {
			return variable.Name, true
		}
	}
	return "", false
}

func authoritativeSinkUses(statement ast.Stmt, variable string) int {
	conditional, ok := statement.(*ast.IfStmt)
	if !ok || conditional.Init == nil {
		return 0
	}
	assignment, ok := conditional.Init.(*ast.AssignStmt)
	if !ok || len(assignment.Lhs) < 1 {
		return 0
	}
	errorVariable, ok := assignment.Lhs[len(assignment.Lhs)-1].(*ast.Ident)
	if !ok || errorVariable.Name == "_" {
		return 0
	}
	condition, ok := conditional.Cond.(*ast.BinaryExpr)
	if !ok || condition.Op != token.NEQ {
		return 0
	}
	left, leftOK := condition.X.(*ast.Ident)
	right, rightOK := condition.Y.(*ast.Ident)
	if !leftOK || !rightOK || left.Name != errorVariable.Name || right.Name != "nil" {
		return 0
	}
	uses := 0
	ast.Inspect(conditional.Init, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || (selector.Sel.Name != "ExecContext" && selector.Sel.Name != "QueryRowContext") {
			return true
		}
		for _, argument := range call.Args {
			identifier, ok := argument.(*ast.Ident)
			if ok && identifier.Name == variable {
				uses++
			}
		}
		return true
	})
	return uses
}

func assignsIdentifier(statement ast.Stmt, name string) bool {
	assigned := false
	ast.Inspect(statement, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, expression := range assignment.Lhs {
			identifier, ok := expression.(*ast.Ident)
			if ok && identifier.Name == name {
				assigned = true
			}
		}
		return !assigned
	})
	return assigned
}

func serializedBindings(body *ast.BlockStmt, item binding) int {
	matches := 0
	for index, statement := range body.List {
		variable, ok := serializedBindingVariable(statement, item)
		if !ok {
			continue
		}
		sinkUses := 0
		for _, later := range body.List[index+1:] {
			uses := authoritativeSinkUses(later, variable)
			if uses > 0 {
				sinkUses += uses
				continue
			}
			if assignsIdentifier(later, variable) {
				sinkUses = 0
				break
			}
		}
		if sinkUses == 1 {
			matches++
		}
	}
	return matches
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
		matches += serializedBindings(function.Body, item)
	}
	if functions != 1 || matches != 1 {
		return fmt.Errorf("%w: %s requires exactly one top-level json.Marshal payload in %s with [%q]=%q used once by a checked database sink; functions=%d matches=%d", errInvalidRegistry, filename, item.GoFunction, item.JSONField, item.Key, functions, matches)
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
