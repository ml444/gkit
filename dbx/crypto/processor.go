package crypto

import (
	"fmt"
	"reflect"
	"strings"
)

type fieldBinding struct {
	index  int
	column string
	mode   Mode
}

type Processor struct {
	ring           KeyRing
	prim           *Primitive
	mutate         MutatePolicy
	disableDecrypt bool
	legacy         func(mode Mode, ciphertext string) (string, error)
	byIndex        []fieldBinding
	byColumn       map[string]fieldBinding
}

func NewProcessor(cfg Config, ormModel any) (*Processor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	t := reflect.TypeOf(ormModel)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: orm model must be struct", ErrInvalidConfig)
	}

	p := &Processor{
		ring:           cfg.KeyRing,
		prim:           NewPrimitive(),
		mutate:         cfg.Mutate,
		disableDecrypt: cfg.DisableDecrypt,
		legacy:         cfg.LegacyDecrypt,
		byColumn:       make(map[string]fieldBinding),
	}

	for _, spec := range cfg.Fields {
		idx, col, err := resolveField(t, spec)
		if err != nil {
			return nil, err
		}
		ft := t.Field(idx).Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.String {
			return nil, fmt.Errorf("%w: field %s must be string", ErrInvalidConfig, t.Field(idx).Name)
		}
		b := fieldBinding{index: idx, column: col, mode: spec.Mode}
		p.byIndex = append(p.byIndex, b)
		p.byColumn[col] = b
	}
	return p, nil
}

func resolveField(t reflect.Type, spec FieldSpec) (int, string, error) {
	if spec.Struct != "" {
		f, ok := t.FieldByName(spec.Struct)
		if !ok {
			return 0, "", fmt.Errorf("%w: struct field %q not found", ErrInvalidConfig, spec.Struct)
		}
		col := spec.Column
		if col == "" {
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				col = strings.Split(jsonTag, ",")[0]
			} else {
				col = camelToSnake(f.Name)
			}
		}
		// FieldByName doesn't give index for embedded easily; scan
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Name == spec.Struct {
				return i, col, nil
			}
		}
	}
	if spec.Column != "" {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			col := f.Name
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				col = strings.Split(jsonTag, ",")[0]
			} else {
				col = camelToSnake(f.Name)
			}
			if col == spec.Column || f.Name == spec.Column {
				return i, col, nil
			}
		}
		return 0, "", fmt.Errorf("%w: column %q not found", ErrInvalidConfig, spec.Column)
	}
	return 0, "", fmt.Errorf("%w: empty field spec", ErrInvalidConfig)
}

func (p *Processor) EncryptValue(m any) (any, error) {
	if m == nil || len(p.byIndex) == 0 && len(p.byColumn) == 0 {
		return m, nil
	}
	return p.transform(m, true, false)
}

func (p *Processor) DecryptValue(m any) (any, error) {
	if m == nil || p.disableDecrypt {
		return m, nil
	}
	if len(p.byIndex) == 0 && len(p.byColumn) == 0 {
		return m, nil
	}
	return p.transform(m, false, false)
}

// EncryptQueryMap encrypts searchable columns only. Storage columns error.
func (p *Processor) EncryptQueryMap(m map[string]any) (map[string]any, error) {
	if m == nil {
		return nil, nil
	}
	out := m
	if p.mutate != MutateInPlace {
		out = make(map[string]any, len(m))
		for k, v := range m {
			out[k] = v
		}
	}
	active, err := p.ring.Active()
	if err != nil {
		return nil, err
	}
	for col, v := range out {
		b, ok := p.byColumn[col]
		if !ok {
			continue
		}
		if b.mode != ModeSearchable {
			return nil, fmt.Errorf("%w: %s", ErrNotSearchable, col)
		}
		s, ok := stringify(v)
		if !ok || s == "" {
			continue
		}
		enc, err := p.prim.Encrypt(b.mode, active, s)
		if err != nil {
			return nil, err
		}
		out[col] = enc
	}
	return out, nil
}

func (p *Processor) transform(m any, encrypt, _ bool) (any, error) {
	v := reflect.ValueOf(m)
	if !v.IsValid() {
		return m, nil
	}
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return m, nil
		}
		elem := v.Elem()
		switch elem.Kind() {
		case reflect.Struct:
			return p.transformStructPtr(v, encrypt)
		case reflect.Slice, reflect.Array:
			return p.transformSlice(v, encrypt)
		default:
			return m, nil
		}
	case reflect.Struct:
		// need addressable copy
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		out, err := p.transformStructPtr(ptr, encrypt)
		if err != nil {
			return nil, err
		}
		return reflect.ValueOf(out).Elem().Interface(), nil
	case reflect.Slice, reflect.Array:
		return p.transformSlice(v, encrypt)
	case reflect.Map:
		return p.transformMap(v, encrypt)
	default:
		return m, nil
	}
}

func (p *Processor) transformStructPtr(ptr reflect.Value, encrypt bool) (any, error) {
	src := ptr.Elem()
	dstPtr := ptr
	if p.mutate != MutateInPlace {
		dstPtr = reflect.New(src.Type())
		dstPtr.Elem().Set(src)
	}
	dst := dstPtr.Elem()
	active, err := p.ring.Active()
	if err != nil {
		return nil, err
	}
	for _, b := range p.byIndex {
		fv := dst.Field(b.index)
		if !fv.IsValid() || !fv.CanSet() {
			continue
		}
		if fv.Kind() != reflect.String {
			return nil, fmt.Errorf("%w: non-string field", ErrInvalidConfig)
		}
		s := fv.String()
		if s == "" {
			continue
		}
		var next string
		if encrypt {
			next, err = p.prim.Encrypt(b.mode, active, s)
		} else {
			next, err = p.decryptString(b.mode, s)
		}
		if err != nil {
			return nil, err
		}
		fv.SetString(next)
	}
	return dstPtr.Interface(), nil
}

func (p *Processor) transformSlice(v reflect.Value, encrypt bool) (any, error) {
	wasPtr := v.Kind() == reflect.Ptr
	slice := v
	if wasPtr {
		slice = v.Elem()
	}
	n := slice.Len()
	var out reflect.Value
	if p.mutate == MutateInPlace && wasPtr {
		out = slice
	} else {
		out = reflect.MakeSlice(slice.Type(), n, n)
		for i := 0; i < n; i++ {
			out.Index(i).Set(slice.Index(i))
		}
	}
	for i := 0; i < n; i++ {
		elem := out.Index(i)
		transformed, err := p.transform(elem.Interface(), encrypt, false)
		if err != nil {
			return nil, err
		}
		tv := reflect.ValueOf(transformed)
		if elem.Kind() == reflect.Ptr && tv.Kind() != reflect.Ptr {
			// shouldn't happen for our struct path
		}
		if tv.IsValid() {
			out.Index(i).Set(tv)
		}
	}
	if wasPtr {
		if p.mutate == MutateInPlace {
			return v.Interface(), nil
		}
		ptr := reflect.New(out.Type())
		ptr.Elem().Set(out)
		return ptr.Interface(), nil
	}
	return out.Interface(), nil
}

func (p *Processor) transformMap(v reflect.Value, encrypt bool) (any, error) {
	if v.Type().Key().Kind() != reflect.String {
		return v.Interface(), nil
	}
	out := v
	if p.mutate != MutateInPlace {
		out = reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			out.SetMapIndex(iter.Key(), iter.Value())
		}
	}
	active, err := p.ring.Active()
	if err != nil {
		return nil, err
	}
	iter := out.MapRange()
	for iter.Next() {
		col := iter.Key().String()
		b, ok := p.byColumn[col]
		if !ok {
			continue
		}
		s, ok := stringify(iter.Value().Interface())
		if !ok || s == "" {
			continue
		}
		var next string
		if encrypt {
			next, err = p.prim.Encrypt(b.mode, active, s)
		} else {
			next, err = p.decryptString(b.mode, s)
		}
		if err != nil {
			return nil, err
		}
		out.SetMapIndex(iter.Key(), reflect.ValueOf(next))
	}
	return out.Interface(), nil
}

func (p *Processor) decryptString(mode Mode, ciphertext string) (string, error) {
	plain, err := p.prim.Decrypt(mode, p.ring, ciphertext)
	if err == nil {
		return plain, nil
	}
	if p.legacy != nil {
		if _, parseErr := ParseEnvelope(ciphertext); parseErr != nil {
			return p.legacy(mode, ciphertext)
		}
	}
	return "", err
}

func stringify(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case *string:
		if x == nil {
			return "", true
		}
		return *x, true
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Interface && !rv.IsNil() {
			return stringify(rv.Elem().Interface())
		}
		return "", false
	}
}
