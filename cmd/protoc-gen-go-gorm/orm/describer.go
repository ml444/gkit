package orm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 预编译正则，检查是否包含破坏 GORM Tag 的非法字符
var invalidTagChars = regexp.MustCompile(`[";\n]`)

func validateOrmOption(fieldName string, optValue string) error {
	if invalidTagChars.MatchString(optValue) {
		// 使用 protogen.Plugin 的 Error 方法报告错误并中断编译
		return fmt.Errorf("protoc-gen-go-gorm: field %q contains invalid characters (';', '\"', or newline) in gorm option: %s", fieldName, optValue)
	}
	return nil
}

type FileDesc struct {
	PackageName string
	Imports     []string
	Commons     []string
	Messages    []*MessageDesc
}

type MessageDesc struct {
	Name              string
	Opts              *MessageOpts
	Fields            []*ORMField
	SerializeFields   []*SerializeDesc
	UtilMap           map[string]string
	Imports           []string
	ForceORM          bool
	NeedGenFuncFields []*ORMField
}

type MessageOpts struct {
	TableName    string
	IndexClauses []*IndexClause
	// ForceIndex   string
	// IgnoreIndex string
}

type ORMField struct {
	FieldName string
	OldType   string
	NewType   string
	// InjectTags *orm.ORMTags
	ORMTag string
}

type SerializeDesc struct {
	IsIgnore           bool
	SerializerName     string
	SerializerTypeName string
	FieldType          string
	// Imports            []string
	Tmpl string
}

func JoinTags(jsonName string, args ...string) string {
	if jsonName != "" {
		args = append(args, fmt.Sprintf(`json:"%s"`, jsonName))
	}
	return strings.Join(args, " ")
}

func JoinORMTags(tags *ORMTags) (bool, string, error) {
	var result []string
	var forceORM bool
	if tags.IgnoreRw != nil && *tags.IgnoreRw {
		forceORM = true
		result = append(result, "-")
	}
	if tags.IgnoreMigration != nil && *tags.IgnoreMigration {
		forceORM = true
		result = append(result, "-:migration")
	}
	if tags.IgnoreAll != nil && *tags.IgnoreAll {
		forceORM = true
		result = append(result, "-:all")
	}
	if tags.OnlyCreate != nil && *tags.OnlyCreate {
		forceORM = true
		result = append(result, "<-:create")
	}
	if tags.OnlyUpdate != nil && *tags.OnlyUpdate {
		forceORM = true
		result = append(result, "<-:update")
	}
	if tags.DisableWrite != nil && *tags.DisableWrite {
		forceORM = true
		result = append(result, "<-:false")
	}
	if tags.DisableRead != nil && *tags.DisableRead {
		forceORM = true
		result = append(result, "->:false")
	}

	if tags.NotNull != nil && *tags.NotNull {
		result = append(result, "not null")
	}
	if tags.Column != nil {
		if err := validateOrmOption("column", *tags.Column); err != nil {
			return false, "", err
		}
		result = append(result, "column:"+*tags.Column)
	}
	if tags.Type != nil {
		if err := validateOrmOption("type", *tags.Type); err != nil {
			return false, "", err
		}
		result = append(result, "type:"+*tags.Type)
	}
	if tags.Default != nil {
		v := *tags.Default
		if v == "" {
			result = append(result, "default:''")
		} else {
			if err := validateOrmOption("default", v); err != nil {
				return false, "", err
			}
			result = append(result, "default:"+v)
		}
	}

	if tags.Comment != nil {
		if err := validateOrmOption("comment", *tags.Comment); err != nil {
			return false, "", err
		}
		result = append(result, "comment:"+*tags.Comment)
	}
	if tags.PrimaryKey != nil && *tags.PrimaryKey {
		result = append(result, "primaryKey")
	}
	if len(tags.Index) > 0 {
		for _, index := range tags.Index {
			if err := validateOrmOption("index", index); err != nil {
				return false, "", err
			}
			result = append(result, "index:"+index)
		}
	}
	if len(tags.UniqueIndex) > 0 {
		for _, index := range tags.UniqueIndex {
			if err := validateOrmOption("uniqueIndex", index); err != nil {
				return false, "", err
			}
			result = append(result, "uniqueIndex:"+index)
		}
	}
	if tags.Size != nil {
		result = append(result, fmt.Sprintf("size:%d", *tags.Size))
	}
	if tags.Precision != nil {
		result = append(result, fmt.Sprintf("precision:%d", *tags.Precision))
	}
	if tags.Scale != nil {
		result = append(result, fmt.Sprintf("scale:%d", *tags.Scale))
	}
	if tags.Embedded != nil && *tags.Embedded {
		forceORM = true
		result = append(result, "embedded")
	}
	if tags.EmbeddedPrefix != nil {
		if err := validateOrmOption("embeddedPrefix", *tags.EmbeddedPrefix); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "embeddedPrefix:"+*tags.EmbeddedPrefix)
	}
	if tags.AutoIncrement != nil && *tags.AutoIncrement {
		result = append(result, "autoIncrement")
	}
	if tags.AutoIncrementIncrement != nil {
		result = append(result, fmt.Sprintf("autoIncrementIncrement:%d", *tags.AutoIncrementIncrement))
	}
	if tags.AutoCreateTime != nil {
		forceORM = true
		s := TimeKindToString(*tags.AutoCreateTime)
		if s == "" {
			result = append(result, "autoCreateTime")
		} else {
			result = append(result, "autoCreateTime:"+s)
		}
	}
	if tags.AutoUpdateTime != nil {
		forceORM = true
		s := TimeKindToString(*tags.AutoUpdateTime)
		if s == "" {
			result = append(result, "autoUpdateTime")
		} else {
			result = append(result, "autoUpdateTime:"+s)
		}
	}
	if tags.Check != nil {
		if err := validateOrmOption("check", *tags.Check); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "check:"+*tags.Check)
	}
	if tags.Encrypt != nil {
		result = append(result, "encrypt:"+strconv.FormatBool(*tags.Encrypt))
	}

	if tags.Serializer != nil {
		if err := validateOrmOption("serializer", *tags.Serializer); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "serializer:"+*tags.Serializer)
	}

	if tags.ForeignKey != nil {
		if err := validateOrmOption("foreignKey", *tags.ForeignKey); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "foreignKey:"+*tags.ForeignKey)
	}
	if tags.References != nil {
		if err := validateOrmOption("references", *tags.References); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "references:"+*tags.References)
	}

	// polymorphic
	if tags.Polymorphic != nil {
		if err := validateOrmOption("polymorphic", *tags.Polymorphic); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "polymorphic:"+*tags.Polymorphic)
	}
	if tags.PolymorphicType != nil {
		if err := validateOrmOption("polymorphicType", *tags.PolymorphicType); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "polymorphicType:"+*tags.PolymorphicType)
	}
	if tags.PolymorphicValue != nil {
		if err := validateOrmOption("polymorphicValue", *tags.PolymorphicValue); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "polymorphicValue:"+*tags.PolymorphicValue)
	}
	if tags.PolymorphicId != nil {
		if err := validateOrmOption("polymorphicId", *tags.PolymorphicId); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "polymorphicId:"+*tags.PolymorphicId)
	}

	if tags.Many2Many != nil {
		if err := validateOrmOption("many2many", *tags.Many2Many); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "many2many:"+*tags.Many2Many)
	}
	if tags.JoinForeignKey != nil {
		if err := validateOrmOption("joinForeignKey", *tags.JoinForeignKey); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "joinForeignKey:"+*tags.JoinForeignKey)
	}
	if tags.JoinReferences != nil {
		if err := validateOrmOption("joinReferences", *tags.JoinReferences); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "joinReferences:"+*tags.JoinReferences)
	}

	if tags.Constraint != nil {
		if err := validateOrmOption("constraint", *tags.Constraint); err != nil {
			return false, "", err
		}
		forceORM = true
		result = append(result, "constraint:"+*tags.Constraint)
	}
	return forceORM, fmt.Sprintf(`gorm:"%s"`, strings.Join(result, ";")), nil
}

func TimeKindToString(kind TimestampKind) string {
	switch kind {
	case TimestampKind_NANO:
		return "nano"
	case TimestampKind_MILLI:
		return "milli"
	default:
		return ""
	}
}
