package crypto

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

type MutatePolicy int

const (
	MutateCopy MutatePolicy = iota
	MutateInPlace
)

type FieldSpec struct {
	Column string
	Struct string
	Mode   Mode
}

type Config struct {
	KeyRing        KeyRing
	Fields         []FieldSpec
	Mutate         MutatePolicy
	DisableDecrypt bool
	// LegacyDecrypt is an opt-in hook for pre-envelope ciphertext. Default nil (off).
	LegacyDecrypt func(mode Mode, ciphertext string) (string, error)
}

func (c Config) Validate() error {
	if err := c.KeyRing.Validate(); err != nil {
		return err
	}
	seenCol := make(map[string]struct{})
	for _, f := range c.Fields {
		if !f.Mode.Valid() {
			return fmt.Errorf("%w: invalid mode", ErrInvalidConfig)
		}
		if f.Column == "" && f.Struct == "" {
			return fmt.Errorf("%w: field needs Column or Struct", ErrInvalidConfig)
		}
		if f.Column != "" {
			if _, ok := seenCol[f.Column]; ok {
				return fmt.Errorf("%w: duplicate column %q", ErrInvalidConfig, f.Column)
			}
			seenCol[f.Column] = struct{}{}
		}
	}
	return nil
}

// DiscoverFields reads gorm tags encrypt:storage|searchable. v1 encrypt:true is ignored.
func DiscoverFields(ormModel any) ([]FieldSpec, error) {
	if ormModel == nil {
		return nil, fmt.Errorf("%w: nil model", ErrInvalidConfig)
	}
	t := reflect.TypeOf(ormModel)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: model must be struct", ErrInvalidConfig)
	}
	var fields []FieldSpec
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			continue
		}
		mode, ok := parseEncryptMode(gormTag)
		if !ok {
			continue
		}
		col := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			col = strings.Split(jsonTag, ",")[0]
		} else {
			col = camelToSnake(field.Name)
		}
		fields = append(fields, FieldSpec{
			Column: col,
			Struct: field.Name,
			Mode:   mode,
		})
	}
	return fields, nil
}

func parseEncryptMode(gormTag string) (Mode, bool) {
	for _, s := range strings.Split(gormTag, ";") {
		ss := strings.SplitN(strings.TrimSpace(s), ":", 2)
		if len(ss) != 2 || strings.TrimSpace(ss[0]) != "encrypt" {
			continue
		}
		switch strings.TrimSpace(ss[1]) {
		case "storage":
			return ModeStorage, true
		case "searchable":
			return ModeSearchable, true
		default:
			// encrypt:true and other v1 values are ignored
			return 0, false
		}
	}
	return 0, false
}

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
