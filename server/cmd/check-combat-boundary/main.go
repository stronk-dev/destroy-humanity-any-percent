// Command check-combat-boundary enforces C3's no-direct-division boundary in
// every Go combat source, including tests and future nested engine packages.
package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := flag.String("root", "..", "repository root")
	flag.Parse()
	if err := checkDirectory(filepath.Join(*root, "server", "combat")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Go combat division boundary ok")
}

func checkDirectory(root string) error {
	count := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("combat boundary refuses symlink: %s", path)
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		count++
		return checkFile(path)
	})
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("combat boundary found no Go sources")
	}
	return nil
}

func checkFile(path string) error {
	positions := token.NewFileSet()
	source, err := parser.ParseFile(positions, path, nil, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("combat boundary cannot parse %s: %w", path, err)
	}
	var violation error
	ast.Inspect(source, func(node ast.Node) bool {
		if violation != nil {
			return false
		}
		var location token.Pos
		switch node := node.(type) {
		case *ast.BinaryExpr:
			if node.Op == token.QUO {
				location = node.OpPos
			}
		case *ast.AssignStmt:
			if node.Tok == token.QUO_ASSIGN {
				location = node.TokPos
			}
		}
		if location.IsValid() {
			violation = fmt.Errorf("%s: native combat division is forbidden; use the shared integer helper", positions.Position(location))
		}
		return violation == nil
	})
	return violation
}
