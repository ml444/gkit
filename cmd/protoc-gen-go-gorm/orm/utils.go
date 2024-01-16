package orm

import (
	"errors"
	"strings"
)

func (x *IndexClause) ToString() string {
	if x == nil {
		return ""
	}
	var buf strings.Builder
	buf.WriteString("hints.")
	switch x.Op {
	case IndexOpKind_FORCE:
		buf.WriteString("ForceIndex")
	case IndexOpKind_IGNORE:
		buf.WriteString("IgnoreIndex")
	default:
		buf.WriteString("UseIndex")
	}
	buf.WriteByte('(')

	for i, key := range x.Indexs {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		buf.WriteString(key)
		buf.WriteByte('"')
	}
	buf.WriteByte(')')

	switch x.For {
	case IndexForKind_JOIN:
		buf.WriteString(".ForJoin()")
	case IndexForKind_GROUP_BY:
		buf.WriteString(".ForGroupBy()")
	case IndexForKind_ORDER_BY:
		buf.WriteString(".ForOrderBy()")
	}
	return buf.String()
}

func (x *IndexClause) ToFuncName() (string, error) {
	if len(x.Indexs) == 0 {
		return "", errors.New("must provider index's name")
	}
	buf := strings.Builder{}
	switch x.Op {
	case IndexOpKind_USE:
		buf.WriteString("UseIndex")
	case IndexOpKind_FORCE:
		buf.WriteString("ForceIndex")
	case IndexOpKind_IGNORE:
		buf.WriteString("IgnoreIndex")
	default:
		buf.WriteString("UseIndex")
	}
	switch x.For {
	case IndexForKind_JOIN:
		buf.WriteString("ForJoin")
	case IndexForKind_GROUP_BY:
		buf.WriteString("ForGroupBy")
	case IndexForKind_ORDER_BY:
		buf.WriteString("ForOrderBy")
	}

	suffix := JoinStringsByCamel(x.Indexs)
	if suffix != "" {
		buf.WriteString("2")
		buf.WriteString(suffix)
	}
	return buf.String(), nil
}

func SnakeToCamel(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if !k && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || !k) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func JoinStringsByCamel(ss []string) string {
	res := ""
	for _, s := range ss {
		res += SnakeToCamel(s)
	}
	return res
}
