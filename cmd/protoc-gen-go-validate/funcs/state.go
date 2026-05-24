package funcs

// ResetFileState initializes per-file mutable state. Must be called at the
// start of each generated file to avoid cross-file import alias collisions.
func ResetFileState(fileAlias, pkgName string) {
	FileAlias = fileAlias
	CurrentProtoPkgName = pkgName
	ExtraPkg = map[string]string{}
	ExtraPkgPath = map[string]string{}
	StdPkg = map[string]int{
		"bytes":   0,
		"errors":  0,
		"fmt":     0,
		"net":     0,
		"mail":    0,
		"url":     0,
		"regexp":  0,
		"sort":    0,
		"strings": 0,
		"time":    0,
		"utf8":    0,
		"anypb":   0,
	}
	enumNameCache = map[string]map[int32]string{}
}
