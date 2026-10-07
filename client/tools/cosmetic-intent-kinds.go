// Verification only: inspect the production Go AST without compiling or importing
// the service. This is not an independently maintained command registry.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func cosmeticKinds(authority, decoder string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "cosmetic_intent.go", authority, 0)
	if err != nil {
		return nil, fmt.Errorf("cosmetic authority cannot parse: %w", err)
	}
	constants := map[string]string{}
	var predicate *ast.FuncDecl
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.GenDecl:
			if declaration.Tok != token.CONST {
				continue
			}
			for _, specification := range declaration.Specs {
				value := specification.(*ast.ValueSpec)
				for i, name := range value.Names {
					if _, exists := constants[name.Name]; exists || i >= len(value.Values) {
						return nil, fmt.Errorf("duplicate or implicit cosmetic constant %s", name.Name)
					}
					literal, ok := value.Values[i].(*ast.BasicLit)
					if !ok || literal.Kind != token.STRING {
						return nil, fmt.Errorf("cosmetic constant %s is not a string literal", name.Name)
					}
					text, err := strconv.Unquote(literal.Value)
					if err != nil || text == "" {
						return nil, fmt.Errorf("cosmetic constant %s is empty or invalid", name.Name)
					}
					constants[name.Name] = text
				}
			}
		case *ast.FuncDecl:
			if declaration.Name.Name == "isCosmeticIntent" {
				if predicate != nil || declaration.Recv != nil {
					return nil, fmt.Errorf("duplicate or method cosmetic predicate")
				}
				predicate = declaration
			}
		}
	}
	if predicate == nil || predicate.Body == nil || len(predicate.Body.List) != 1 ||
		predicate.Type.Params == nil || len(predicate.Type.Params.List) != 1 ||
		len(predicate.Type.Params.List[0].Names) != 1 || predicate.Type.Results == nil || len(predicate.Type.Results.List) != 1 {
		return nil, fmt.Errorf("unsupported cosmetic predicate signature/body")
	}
	parameter := predicate.Type.Params.List[0]
	inputType, inputOK := parameter.Type.(*ast.Ident)
	outputType, outputOK := predicate.Type.Results.List[0].Type.(*ast.Ident)
	returned, returnOK := predicate.Body.List[0].(*ast.ReturnStmt)
	if !inputOK || inputType.Name != "string" || !outputOK || outputType.Name != "bool" || !returnOK || len(returned.Results) != 1 {
		return nil, fmt.Errorf("unsupported cosmetic predicate types/return")
	}
	selected := map[string]string{}
	values := map[string]bool{}
	var inspect func(ast.Expr) error
	inspect = func(expression ast.Expr) error {
		if parens, ok := expression.(*ast.ParenExpr); ok {
			return inspect(parens.X)
		}
		binary, ok := expression.(*ast.BinaryExpr)
		if !ok {
			return fmt.Errorf("unsupported cosmetic predicate expression")
		}
		if binary.Op == token.LOR {
			if err := inspect(binary.X); err != nil {
				return err
			}
			return inspect(binary.Y)
		}
		left, leftOK := binary.X.(*ast.Ident)
		right, rightOK := binary.Y.(*ast.Ident)
		if binary.Op != token.EQL || !leftOK || left.Name != parameter.Names[0].Name || !rightOK {
			return fmt.Errorf("unsupported cosmetic predicate comparison")
		}
		value, exists := constants[right.Name]
		if !exists || values[value] || selected[right.Name] != "" {
			return fmt.Errorf("missing or duplicate selected cosmetic constant %s", right.Name)
		}
		selected[right.Name], values[value] = value, true
		return nil
	}
	if err := inspect(returned.Results[0]); err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("cosmetic command population is empty")
	}
	parsedDecoder, err := parser.ParseFile(token.NewFileSet(), "intents.go", decoder, 0)
	if err != nil {
		return nil, fmt.Errorf("production intent decoder cannot parse: %w", err)
	}
	var decode *ast.FuncDecl
	for _, declaration := range parsedDecoder.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == "ParseIntent" {
			if decode != nil || function.Recv != nil {
				return nil, fmt.Errorf("duplicate or method production decoder")
			}
			decode = function
		}
	}
	if decode == nil || decode.Body == nil {
		return nil, fmt.Errorf("production ParseIntent decoder missing")
	}
	cases := map[string]int{}
	ast.Inspect(decode.Body, func(node ast.Node) bool {
		switchNode, ok := node.(*ast.SwitchStmt)
		if !ok {
			return true
		}
		selector, ok := switchNode.Tag.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Kind" {
			return true
		}
		base, ok := selector.X.(*ast.Ident)
		if !ok || base.Name != "request" {
			return true
		}
		for _, statement := range switchNode.Body.List {
			for _, expression := range statement.(*ast.CaseClause).List {
				if name, ok := expression.(*ast.Ident); ok {
					cases[name.Name]++
				}
			}
		}
		return true
	})
	kinds := make([]string, 0, len(selected))
	for name, value := range selected {
		if cases[name] != 1 {
			return nil, fmt.Errorf("cosmetic command %s requires one production decoder case, got %d", name, cases[name])
		}
		kinds = append(kinds, value)
	}
	sort.Strings(kinds)
	return kinds, nil
}

func verifyFixtures() (int, error) {
	authority := `package production
const ( A = "allowed_a"; B = "allowed_b" )
func isCosmeticIntent(kind string) bool { return kind == A || kind == B }`
	decoder := `package production
func ParseIntent() { switch request.Kind { case A, B: } }`
	positive := authority + "\n// const Fake = \"fake\"\nvar lookalike = `func isCosmeticIntent(kind string) bool { return kind == Fake }`"
	for _, source := range []string{authority, positive} {
		kinds, err := cosmeticKinds(source, decoder)
		if err != nil || strings.Join(kinds, ",") != "allowed_a,allowed_b" {
			return 0, fmt.Errorf("cosmetic authority positive fixture failed: %v", err)
		}
	}
	negative := [][2]string{
		{authority + "\nfunc", decoder},
		{strings.ReplaceAll(authority, "isCosmeticIntent", "unrelated"), decoder},
		{strings.ReplaceAll(authority, `A = "allowed_a"`, `A = unknown`), decoder},
		{strings.ReplaceAll(authority, "kind == A", "kind == Unknown"), decoder},
		{strings.ReplaceAll(authority, `"allowed_b"`, `"allowed_a"`), decoder},
		{strings.ReplaceAll(authority, "kind == A || kind == B", "true"), decoder},
		{authority, strings.ReplaceAll(decoder, "case A, B:", "case A:")},
		{authority, strings.ReplaceAll(decoder, "ParseIntent", "unrelated")},
		{authority, strings.ReplaceAll(decoder, "case A, B:", "case A, A, B:")},
		{authority + "\nfunc isCosmeticIntent(kind string) bool { return kind == A }", decoder},
	}
	for i, pair := range negative {
		if _, err := cosmeticKinds(pair[0], pair[1]); err == nil {
			return 0, fmt.Errorf("cosmetic authority negative fixture %d survived", i)
		}
	}
	return len(negative), nil
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: cosmetic-intent-kinds.go production-directory")
	}
	negative, err := verifyFixtures()
	if err != nil {
		return err
	}
	authority, err := os.ReadFile(filepath.Join(os.Args[1], "cosmetic_intent.go"))
	if err != nil {
		return err
	}
	decoder, err := os.ReadFile(filepath.Join(os.Args[1], "intents.go"))
	if err != nil {
		return err
	}
	kinds, err := cosmeticKinds(string(authority), string(decoder))
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Kinds            []string `json:"kinds"`
		NegativeFixtures int      `json:"negative_fixtures"`
		PositiveFixtures int      `json:"positive_fixtures"`
	}{kinds, negative, 2})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
