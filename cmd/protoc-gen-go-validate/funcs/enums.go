package funcs

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
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
		names := enumNamesMap(enum.Values)
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
		return enumNamesMap(enum.Values)[val]
	}
	return fmt.Sprint(val)
}
