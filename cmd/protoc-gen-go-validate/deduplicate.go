package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func existValidationErrorInPackage(packagePath string, exclude string) (exist bool, err error) {
	if _, err = os.Stat(packagePath); os.IsNotExist(err) {
		err = os.MkdirAll(packagePath, 0755)
		if err != nil {
			return false, err
		}
	}
	// Get all '*_validate.pb.go' files under the package
	filePaths, err := getGoFilesInPackage(packagePath)
	if err != nil {
		return false, err
	}
	for _, filePath := range filePaths {
		if exist {
			return exist, nil
		}
		if strings.Contains(filePath, exclude) {
			continue
		}
		fset := token.NewFileSet()

		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			println("err", err.Error())
			return false, err
		}

		ast.Inspect(f, func(n ast.Node) bool {
			// Check if it is a structure declaration
			if typeSpec, ok := n.(*ast.TypeSpec); ok {
				if _, ok := typeSpec.Type.(*ast.StructType); ok {
					// Get the name of the structure
					structName := typeSpec.Name.Name
					if structName == "ValidationError" {
						exist = true
						return false
					}
				}
			}
			return true
		})
	}

	return exist, nil
}

func getGoFilesInPackage(packagePath string) ([]string, error) {
	var filePaths []string
	absPath, err := filepath.Abs(packagePath)
	if err != nil {
		return nil, err
	}
	var curDir string
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dir := filepath.Dir(path)
		if curDir != "" && curDir != dir {
			return nil
		}
		if strings.HasSuffix(path, "_validate.pb.go") {
			curDir = filepath.Dir(path)
			filePaths = append(filePaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

func isRelativePath(parameter string) bool {
	var isRelative bool
	for _, param := range strings.Split(parameter, ",") {
		var value string
		if i := strings.Index(param, "="); i >= 0 {
			value = param[i+1:]
			param = param[0:i]
		}
		if param == "paths" && value == "source_relative" {
			isRelative = true
			return isRelative
		}
	}
	return isRelative
}
