package orm

import (
	"fmt"
	"strconv"
	"strings"
)

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

func JoinORMTags(tags *ORMTags) (bool, string) {
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
		result = append(result, "column:"+*tags.Column)
	}
	if tags.Type != nil {
		result = append(result, "type:"+*tags.Type)
	}
	if tags.Default != nil {
		result = append(result, "default:"+*tags.Default)
	}

	if tags.Comment != nil {
		result = append(result, "comment:"+*tags.Comment)
	}
	if tags.PrimaryKey != nil && *tags.PrimaryKey {
		result = append(result, "primaryKey")
	}
	if len(tags.Index) > 0 {
		for _, index := range tags.Index {
			result = append(result, "index:"+index)
		}
	}
	if len(tags.UniqueIndex) > 0 {
		for _, index := range tags.UniqueIndex {
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
		forceORM = true
		result = append(result, "check:"+*tags.Check)
	}
	if tags.Encrypt != nil {
		result = append(result, "encrypt:"+strconv.FormatBool(*tags.Encrypt))
	}

	if tags.Serializer != nil {
		forceORM = true
		result = append(result, "serializer:"+*tags.Serializer)
	}

	if tags.ForeignKey != nil {
		forceORM = true
		result = append(result, "foreignKey:"+*tags.ForeignKey)
	}
	if tags.References != nil {
		forceORM = true
		result = append(result, "references:"+*tags.References)
	}

	// polymorphic
	if tags.Polymorphic != nil {
		forceORM = true
		result = append(result, "polymorphic:"+*tags.Polymorphic)
	}
	if tags.PolymorphicType != nil {
		forceORM = true
		result = append(result, "polymorphicType:"+*tags.PolymorphicType)
	}
	if tags.PolymorphicValue != nil {
		forceORM = true
		result = append(result, "polymorphicValue:"+*tags.PolymorphicValue)
	}
	if tags.PolymorphicId != nil {
		forceORM = true
		result = append(result, "polymorphicId:"+*tags.PolymorphicId)
	}

	if tags.Many2Many != nil {
		forceORM = true
		result = append(result, "many2many:"+*tags.Many2Many)
	}
	if tags.JoinForeignKey != nil {
		forceORM = true
		result = append(result, "joinForeignKey:"+*tags.JoinForeignKey)
	}
	if tags.JoinReferences != nil {
		forceORM = true
		result = append(result, "joinReferences:"+*tags.JoinReferences)
	}

	if tags.Constraint != nil {
		forceORM = true
		result = append(result, "constraint:"+*tags.Constraint)
	}
	return forceORM, fmt.Sprintf(`gorm:"%s"`, strings.Join(result, ";"))
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
