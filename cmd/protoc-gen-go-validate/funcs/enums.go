package funcs

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func isEnum(f protogen.Field) bool {
	return f.Desc.Kind() == protoreflect.EnumKind
}

func enumNamesMap(values []*protogen.EnumValue) (m map[int32]string) {
	m = make(map[int32]string)
	for _, v := range values {
		if _, exists := m[int32(v.Desc.Number())]; !exists {
			m[int32(v.Desc.Number())] = string(v.Desc.Name())
		}
	}
	return m
}

// enumList - if type is ENUM, enum values are returned
func enumList(f protogen.Field, list []int32) string {
	stringList := make([]string, 0, len(list))
	if enum := f.Enum; enum != nil {
		names := cachedEnumNamesMap(string(enum.Desc.FullName()), enum.Values)
		for _, n := range list {
			stringList = append(stringList, names[n])
		}
	} else {
		for _, n := range list {
			stringList = append(stringList, fmt.Sprint(n))
		}
	}
	return "[" + strings.Join(stringList, " ") + "]"
}

// enumVal - if type is ENUM, enum value is returned
func enumVal(f protogen.Field, val int32) string {
	if enum := f.Enum; enum != nil {
		return cachedEnumNamesMap(string(enum.Desc.FullName()), enum.Values)[val]
	}
	return fmt.Sprint(val)
}

func EnumName(enum protoreflect.EnumDescriptor) string {
	if len(ExtraPkg) == 0 {
		GetImports(FileDescriptor(enum))
	}
	out := string(enum.Name())
	parent := enum.Parent()
	for {
		message, ok := parent.(protoreflect.MessageDescriptor)
		if ok {
			out = string(message.Name()) + "_" + out
			parent = message.Parent()
		} else {
			if pkgName, ok := ExtraPkg[string(parent.FullName())]; ok {
				return pkgName + "." + out
			}
			return out
		}
	}
}

var StdPkg = map[string]int{}

var CurrentProtoPkgName string

// map[pkgPath]pkgName
var ExtraPkg = map[string]string{}
var ExtraPkgPath = map[string]string{}

// enumNameCache maps enum full name to value->name map (per file).
var enumNameCache map[string]map[int32]string

func cachedEnumNamesMap(enumFullName string, values []*protogen.EnumValue) map[int32]string {
	if m, ok := enumNameCache[enumFullName]; ok {
		return m
	}
	m := enumNamesMap(values)
	enumNameCache[enumFullName] = m
	return m
}

// AllocPkgAlias returns a unique Go package alias, renaming on collision with stdlib names.
func AllocPkgAlias(pkgName string) string {
	if cnt, ok := StdPkg[pkgName]; ok {
		alias := fmt.Sprintf("%s%d", pkgName, cnt+1)
		StdPkg[alias] = 0
		StdPkg[pkgName] = cnt + 1
		return alias
	}
	StdPkg[pkgName] = 0
	return pkgName
}

func GetImports(fileDesc protoreflect.FileDescriptor) map[string]string {
	ExtraPkg["_"] = "v"
	imports := fileDesc.Imports()
	if imports.Len() == 0 {
		return map[string]string{}
	}

	for i := 0; i < imports.Len(); i++ {
		imp := imports.Get(i)
		fp, ok := imp.Options().(*descriptorpb.FileOptions)
		if !ok {
			continue
		}
		pkgName := string(imp.Package().Name())
		if pkgName == "v" || pkgName == "protobuf" || pkgName == CurrentProtoPkgName{
			continue
		}

		if fp.GoPackage != nil {
			pkgName = AllocPkgAlias(pkgName)
			pkgPath := strings.SplitN(*fp.GoPackage, ";", 2)[0]
			ExtraPkgPath[pkgName] = pkgPath
			ExtraPkg[string(imp.FullName())] = pkgName
		}
	}
	return ExtraPkgPath
}

func FileDescriptor(enum protoreflect.Descriptor) protoreflect.FileDescriptor {
	parent := enum.Parent()
	for {
		file, ok := parent.(protoreflect.FileDescriptor)
		if !ok {
			parent = parent.Parent()
			continue
		}
		return file
	}
}

//	type NormalizedEnum struct {
//		PkgFullname string
//		Name        string
//	}
//
//	func enumPackages(enums []protoreflect.EnumDescriptor) map[string]NormalizedEnum {
//		out := make(map[string]NormalizedEnum, len(enums))
//
//		nameCollision := map[string]int{
//			"bytes":   0,
//			"errors":  0,
//			"fmt":     0,
//			"net":     0,
//			"mail":    0,
//			"url":     0,
//			"regexp":  0,
//			"sort":    0,
//			"strings": 0,
//			"time":    0,
//			"utf8":    0,
//			"anypb":   0,
//		}
//		nameNormalized := make(map[string]struct{})
//
//		for _, en := range enums {
//			// TODO
//			pkgName := _packageName(en)
//			enImportPath := string(en.Parent().FullName())
//			if _, ok := nameNormalized[pkgName]; ok {
//				continue
//			}
//
//			if collision, ok := nameCollision[pkgName]; ok {
//				nameCollision[pkgName] = collision + 1
//				pkgName = pkgName + string(strconv.Itoa(nameCollision[pkgName]))
//			}
//
//			nameNormalized[enImportPath] = struct{}{}
//			out[pkgName] = NormalizedEnum{
//				Name:        EnumName(en, false),
//				PkgFullname: enImportPath,
//			}
//
//		}
//
//		return out
//	}
func _packageName(enum protoreflect.EnumDescriptor) string {
	parent := enum.Parent()
	for {
		file, ok := parent.(protoreflect.FileDescriptor)
		if !ok {
			parent = parent.Parent()
			continue
		}
		return string(file.Name())
	}
}
