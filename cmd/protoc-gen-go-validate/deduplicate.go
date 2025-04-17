package main

// func FindRedeclaredInPkg(packagePath string, exclude string) (result []string, err error) {
// 	if _, err = os.Stat(packagePath); os.IsNotExist(err) {
// 		err = os.MkdirAll(packagePath, 0755)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	// Get all '*_validate.pb.go' files under the package
// 	filePaths, err := getGoFilesInPackage(packagePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, filePath := range filePaths {
// 		if strings.Contains(filePath, exclude) {
// 			continue
// 		}
// 		fset := token.NewFileSet()
//
// 		f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
// 		if err != nil {
// 			println("err", err.Error())
// 			return nil, err
// 		}
//
// 		ast.Inspect(f, func(n ast.Node) bool {
// 			switch decl := n.(type) {
// 			case *ast.GenDecl:
// 				// process declaration statements (variables or constants)
// 				for _, spec := range decl.Specs {
// 					if vSpec, ok := spec.(*ast.ValueSpec); ok {
// 						for _, ident := range vSpec.Names {
// 							name := ident.Name
// 							switch name {
// 							case "_uuidPattern":
// 								result = append(result, "_uuidPattern")
// 							}
// 						}
// 					}
// 				}
// 			case *ast.FuncDecl:
// 				// process function declaration
// 				name := decl.Name.Name
// 				switch name {
// 				case "_validateUuid":
// 					result = append(result, "_validateUuid")
// 				case "_validateHostname":
// 					result = append(result, "_validateHostname")
// 				case "_validateEmail":
// 					result = append(result, "_validateEmail")
// 				}
// 			case *ast.TypeSpec:
// 				// process type declaration
// 				// check if it is a structure declaration
// 				if _, ok := decl.Type.(*ast.StructType); ok {
// 					// Get the name of the structure
// 					structName := decl.Name.Name
// 					if structName == "ValidationError" {
// 						result = append(result, "ValidationError")
// 						return false
// 					}
// 				}
//
// 			}
// 			return true
// 		})
// 	}
//
// 	return result, nil
// }

// func getGoFilesInPackage(packagePath string) ([]string, error) {
// 	var filePaths []string
// 	absPath, err := filepath.Abs(packagePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var curDir string
// 	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		dir := filepath.Dir(path)
// 		if curDir != "" && curDir != dir {
// 			return nil
// 		}
// 		if strings.HasSuffix(path, "_validate.pb.go") {
// 			curDir = filepath.Dir(path)
// 			filePaths = append(filePaths, path)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return filePaths, nil
// }

// func isRelativePath(parameter string) bool {
// 	var isRelative bool
// 	for _, param := range strings.Split(parameter, ",") {
// 		var value string
// 		if i := strings.Index(param, "="); i >= 0 {
// 			value = param[i+1:]
// 			param = param[0:i]
// 		}
// 		if param == "paths" && value == "source_relative" {
// 			isRelative = true
// 			return isRelative
// 		}
// 	}
// 	return isRelative
// }
