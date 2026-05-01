// Package main, repodaki tum .go dosyalarinin syntax olarak parse edilebildigini
// dogrulamak icin CI'da kosulan kucuk bir tooldur.
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	errs := 0
	_ = filepath.Walk(".", func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		if strings.HasPrefix(p, "vendor/") || strings.HasPrefix(p, "_ci/") {
			return nil
		}
		if _, e := parser.ParseFile(fset, p, nil, parser.AllErrors); e != nil {
			fmt.Fprintln(os.Stderr, p, e)
			errs++
		}
		return nil
	})
	if errs > 0 {
		os.Exit(1)
	}
	fmt.Println("all .go files parse OK")
}
