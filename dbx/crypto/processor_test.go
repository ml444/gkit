package crypto

import (
	"errors"
	"reflect"
	"testing"
)

type userRow struct {
	Phone string `json:"phone"`
	Name  string `json:"name"`
}

func TestProcessorStructCopyEncryptDecrypt(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields: []FieldSpec{
			{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
		},
		Mutate: MutateCopy,
	}
	p, err := NewProcessor(cfg, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	if p.MutatesInPlace() {
		t.Fatal("copy processor reported in-place mutation")
	}
	in := &userRow{Phone: "138", Name: "n"}
	outAny, err := p.EncryptValue(in)
	if err != nil {
		t.Fatal(err)
	}
	out := outAny.(*userRow)
	if in.Phone != "138" {
		t.Fatal("MutateCopy must not change input")
	}
	if out.Phone == "138" || out.Name != "n" {
		t.Fatalf("out = %#v", out)
	}
	plainAny, err := p.DecryptValue(out)
	if err != nil {
		t.Fatal(err)
	}
	plain := plainAny.(*userRow)
	if plain.Phone != "138" {
		t.Fatalf("plain = %#v", plain)
	}
}

func TestProcessorMapAndQuery(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields: []FieldSpec{
			{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
			{Column: "id_card", Struct: "IDCard", Mode: ModeStorage},
		},
		Mutate: MutateCopy,
	}
	p, err := NewProcessor(cfg, &struct {
		Phone  string `json:"phone"`
		IDCard string `json:"id_card"`
	}{})
	if err != nil {
		t.Fatal(err)
	}
	q, err := p.EncryptQueryMap(map[string]any{"phone": "138"})
	if err != nil {
		t.Fatal(err)
	}
	if q["phone"] == "138" {
		t.Fatal("phone should be encrypted")
	}
	if _, err := p.EncryptQueryMap(map[string]any{"id_card": "x"}); !errors.Is(err, ErrNotSearchable) {
		t.Fatalf("want ErrNotSearchable, got %v", err)
	}
}

func TestProcessorInPlace(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields:  []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}},
		Mutate:  MutateInPlace,
	}
	p, _ := NewProcessor(cfg, &userRow{})
	if !p.MutatesInPlace() {
		t.Fatal("in-place processor did not report its mutation policy")
	}
	in := &userRow{Phone: "138"}
	out, err := p.EncryptValue(in)
	if err != nil {
		t.Fatal(err)
	}
	if out != in || in.Phone == "138" {
		t.Fatalf("inplace failed: %#v", in)
	}
}

func TestProcessorLegacyDecrypt(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields:  []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}},
		LegacyDecrypt: func(mode Mode, ciphertext string) (string, error) {
			if ciphertext == "legacy:138" {
				return "138", nil
			}
			return "", ErrDecryptFailed
		},
	}
	p, _ := NewProcessor(cfg, &userRow{})
	row := &userRow{Phone: "legacy:138"}
	out, err := p.DecryptValue(row)
	if err != nil {
		t.Fatal(err)
	}
	if out.(*userRow).Phone != "138" {
		t.Fatalf("%#v", out)
	}
}

func TestNewProcessorResolveFieldBranches(t *testing.T) {
	cfg := Config{KeyRing: testRing(t), Fields: []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}}}
	if _, err := NewProcessor(Config{}, &userRow{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("invalid cfg: %v", err)
	}
	if _, err := NewProcessor(cfg, 1); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("scalar model: %v", err)
	}
	missingStruct := cfg
	missingStruct.Fields = []FieldSpec{{Struct: "Missing", Mode: ModeSearchable}}
	if _, err := NewProcessor(missingStruct, &userRow{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("missing struct: %v", err)
	}
	missingColumn := cfg
	missingColumn.Fields = []FieldSpec{{Column: "missing", Mode: ModeSearchable}}
	if _, err := NewProcessor(missingColumn, &userRow{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("missing column: %v", err)
	}
	emptySpec := cfg
	emptySpec.Fields = []FieldSpec{{Mode: ModeSearchable}}
	if _, err := NewProcessor(emptySpec, &userRow{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("empty spec: %v", err)
	}
	type nonStringRow struct {
		Phone int `json:"phone"`
	}
	if _, err := NewProcessor(cfg, &nonStringRow{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("non string: %v", err)
	}
	type ptrStringRow struct {
		Phone *string `json:"phone"`
	}
	if _, err := NewProcessor(cfg, &ptrStringRow{}); err != nil {
		t.Fatalf("pointer string should be accepted: %v", err)
	}
	type inferredRow struct{ Phone string }
	inferred := cfg
	inferred.Fields = []FieldSpec{{Struct: "Phone", Mode: ModeSearchable}}
	p, err := NewProcessor(inferred, &inferredRow{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.byColumn["phone"]; !ok {
		t.Fatalf("inferred column missing: %#v", p.byColumn)
	}
	jsonInferred := cfg
	jsonInferred.Fields = []FieldSpec{{Struct: "Phone", Mode: ModeSearchable}}
	p, err = NewProcessor(jsonInferred, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.byColumn["phone"]; !ok {
		t.Fatalf("json inferred column missing: %#v", p.byColumn)
	}
	columnOnly := cfg
	columnOnly.Fields = []FieldSpec{{Column: "Phone", Mode: ModeSearchable}}
	if _, err := NewProcessor(columnOnly, &inferredRow{}); err != nil {
		t.Fatalf("column by field name: %v", err)
	}
	type privateFirst struct {
		hidden string
		Phone  string
	}
	columnOnly.Fields = []FieldSpec{{Column: "phone", Mode: ModeSearchable}}
	if _, err := NewProcessor(columnOnly, &privateFirst{}); err != nil {
		t.Fatalf("column should skip private fields: %v", err)
	}
	if _, _, err := resolveField(reflect.TypeOf(userRow{}), FieldSpec{}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("direct empty resolve: %v", err)
	}
}

func TestProcessorNilEmptyAndDisabledPaths(t *testing.T) {
	cfg := Config{KeyRing: testRing(t), Fields: []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}}}
	p, err := NewProcessor(cfg, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := p.EncryptValue(nil); err != nil || got != nil {
		t.Fatalf("nil encrypt = %#v, %v", got, err)
	}
	noFields := &Processor{}
	row := &userRow{Phone: "138"}
	if got, err := noFields.EncryptValue(row); err != nil || got != row {
		t.Fatalf("no-field encrypt = %#v, %v", got, err)
	}
	if got, err := noFields.DecryptValue(row); err != nil || got != row {
		t.Fatalf("no-field decrypt = %#v, %v", got, err)
	}
	disabled := *p
	disabled.disableDecrypt = true
	if got, err := disabled.DecryptValue(row); err != nil || got != row {
		t.Fatalf("disabled decrypt = %#v, %v", got, err)
	}
	if got, err := p.transform(nil, true, false); err != nil || got != nil {
		t.Fatalf("nil transform = %#v, %v", got, err)
	}
	if got, err := p.transform((*userRow)(nil), true, false); err != nil || got != (*userRow)(nil) {
		t.Fatalf("nil ptr transform = %#v, %v", got, err)
	}
	n := 12
	if got, err := p.transform(&n, true, false); err != nil || got != &n {
		t.Fatalf("ptr scalar transform = %#v, %v", got, err)
	}
	if got, err := p.transform(12, true, false); err != nil || got != 12 {
		t.Fatalf("scalar transform = %#v, %v", got, err)
	}
	if got, err := p.transform(&[]int{1, 2}, true, false); err != nil || !reflect.DeepEqual(got, &[]int{1, 2}) {
		t.Fatalf("non-struct slice ptr = %#v, %v", got, err)
	}
}

func TestProcessorTransformStructSliceMapAndQueries(t *testing.T) {
	cfg := Config{
		KeyRing: testRing(t),
		Fields: []FieldSpec{
			{Column: "phone", Struct: "Phone", Mode: ModeSearchable},
			{Column: "name", Struct: "Name", Mode: ModeStorage},
		},
		Mutate: MutateCopy,
	}
	p, err := NewProcessor(cfg, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	valueOut, err := p.EncryptValue(userRow{Phone: "138", Name: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if valueOut.(userRow).Phone == "138" || valueOut.(userRow).Name == "alice" {
		t.Fatalf("struct value was not encrypted: %#v", valueOut)
	}
	plainValue, err := p.DecryptValue(valueOut)
	if err != nil {
		t.Fatal(err)
	}
	if plainValue.(userRow).Phone != "138" || plainValue.(userRow).Name != "alice" {
		t.Fatalf("struct value decrypt = %#v", plainValue)
	}

	rows := []userRow{{Phone: "138", Name: "a"}, {Phone: "", Name: ""}}
	encRowsAny, err := p.EncryptValue(rows)
	if err != nil {
		t.Fatal(err)
	}
	encRows := encRowsAny.([]userRow)
	if encRows[0].Phone == "138" || rows[0].Phone != "138" || encRows[1].Phone != "" {
		t.Fatalf("slice encrypt = %#v, original=%#v", encRows, rows)
	}
	decRowsAny, err := p.DecryptValue(encRows)
	if err != nil {
		t.Fatal(err)
	}
	if decRowsAny.([]userRow)[0].Phone != "138" {
		t.Fatalf("slice decrypt = %#v", decRowsAny)
	}

	ptrRows := []*userRow{{Phone: "139", Name: "b"}}
	encPtrRowsAny, err := p.EncryptValue(&ptrRows)
	if err != nil {
		t.Fatal(err)
	}
	encPtrRows := encPtrRowsAny.(*[]*userRow)
	if (*encPtrRows)[0].Phone == "139" || ptrRows[0].Phone != "139" {
		t.Fatalf("ptr slice copy = %#v original=%#v", encPtrRows, ptrRows)
	}

	inPlace, err := NewProcessor(Config{KeyRing: testRing(t), Fields: cfg.Fields, Mutate: MutateInPlace}, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	inPlaceRows := []userRow{{Phone: "140", Name: "c"}}
	got, err := inPlace.EncryptValue(&inPlaceRows)
	if err != nil {
		t.Fatal(err)
	}
	if got != &inPlaceRows || inPlaceRows[0].Phone == "140" {
		t.Fatalf("in-place slice = %#v", inPlaceRows)
	}

	m := map[string]any{"phone": "141", "name": "d", "other": "x"}
	encMapAny, err := p.EncryptValue(m)
	if err != nil {
		t.Fatal(err)
	}
	encMap := encMapAny.(map[string]any)
	if encMap["phone"] == "141" || encMap["name"] == "d" || m["phone"] != "141" {
		t.Fatalf("map encrypt = %#v original=%#v", encMap, m)
	}
	decMapAny, err := p.DecryptValue(encMap)
	if err != nil {
		t.Fatal(err)
	}
	decMap := decMapAny.(map[string]any)
	if decMap["phone"] != "141" || decMap["name"] != "d" {
		t.Fatalf("map decrypt = %#v", decMap)
	}
	nonStringKey := map[int]string{1: "x"}
	if got, err := p.EncryptValue(nonStringKey); err != nil || !reflect.DeepEqual(got, nonStringKey) {
		t.Fatalf("non-string map = %#v, %v", got, err)
	}
	skippedMap, err := p.EncryptValue(map[string]any{"phone": "", "name": 123})
	if err != nil {
		t.Fatal(err)
	}
	if skippedMap.(map[string]any)["phone"] != "" || skippedMap.(map[string]any)["name"] != 123 {
		t.Fatalf("map skip values = %#v", skippedMap)
	}

	q, err := p.EncryptQueryMap(nil)
	if err != nil || q != nil {
		t.Fatalf("nil query = %#v, %v", q, err)
	}
	query := map[string]any{"phone": "142", "none": nil, "other": 3}
	encQuery, err := p.EncryptQueryMap(query)
	if err != nil {
		t.Fatal(err)
	}
	if encQuery["phone"] == "142" || query["phone"] != "142" {
		t.Fatalf("query copy = %#v original=%#v", encQuery, query)
	}
	str := "143"
	encQuery, err = p.EncryptQueryMap(map[string]any{"phone": &str})
	if err != nil || encQuery["phone"] == &str {
		t.Fatalf("query pointer string = %#v, %v", encQuery, err)
	}
	var nilString *string
	encQuery, err = p.EncryptQueryMap(map[string]any{"phone": nilString})
	if err != nil || encQuery["phone"] != nilString {
		t.Fatalf("query nil string pointer = %#v, %v", encQuery, err)
	}
	encQuery, err = p.EncryptQueryMap(map[string]any{"phone": 123})
	if err != nil || encQuery["phone"] != 123 {
		t.Fatalf("query non-string value = %#v, %v", encQuery, err)
	}
	_, err = p.EncryptQueryMap(map[string]any{"name": "storage-field"})
	if !errors.Is(err, ErrNotSearchable) {
		t.Fatalf("storage query: %v", err)
	}
	badRing := *p
	badRing.ring = KeyRing{}
	if _, err := badRing.EncryptQueryMap(map[string]any{"phone": "144"}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("bad active query: %v", err)
	}
	badMaterial := *p
	badMaterial.ring = KeyRing{ActiveID: "bad", Keys: map[string][]byte{"bad": []byte("short")}}
	if _, err := badMaterial.EncryptQueryMap(map[string]any{"phone": "145"}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad query material: %v", err)
	}
}

func TestProcessorErrorBranches(t *testing.T) {
	p, err := NewProcessor(Config{KeyRing: testRing(t), Fields: []FieldSpec{{Column: "phone", Struct: "Phone", Mode: ModeSearchable}}}, &userRow{})
	if err != nil {
		t.Fatal(err)
	}
	badActive := *p
	badActive.ring = KeyRing{}
	if _, err := badActive.EncryptValue(&userRow{Phone: "138"}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("bad active struct: %v", err)
	}
	if _, err := badActive.EncryptValue(userRow{Phone: "138"}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("bad active struct value: %v", err)
	}
	if _, err := badActive.EncryptValue([]userRow{{Phone: "138"}}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("bad active slice: %v", err)
	}
	if _, err := badActive.EncryptValue(map[string]any{"phone": "138"}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("bad active map: %v", err)
	}
	badMode := *p
	badMode.byIndex = append([]fieldBinding(nil), p.byIndex...)
	badMode.byIndex[0].mode = Mode(99)
	if _, err := badMode.EncryptValue(&userRow{Phone: "138"}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad mode encrypt: %v", err)
	}
	enc, err := p.EncryptValue(&userRow{Phone: "138"})
	if err != nil {
		t.Fatal(err)
	}
	badModeDecrypt := *p
	badModeDecrypt.byIndex = append([]fieldBinding(nil), p.byIndex...)
	badModeDecrypt.byIndex[0].mode = Mode(99)
	if _, err := badModeDecrypt.DecryptValue(enc); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad mode decrypt: %v", err)
	}
	manual := *p
	manual.byIndex = []fieldBinding{{index: 1, column: "hidden", mode: ModeSearchable}}
	type hiddenRow struct {
		Phone  string
		hidden string
	}
	if got, err := manual.EncryptValue(&hiddenRow{hidden: "x"}); err != nil || got.(*hiddenRow).hidden != "x" {
		t.Fatalf("hidden field should be skipped: %#v, %v", got, err)
	}
	manual.byIndex = []fieldBinding{{index: 0, column: "count", mode: ModeSearchable}}
	type intRow struct{ Count int }
	if _, err := manual.EncryptValue(&intRow{Count: 1}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("manual non-string field: %v", err)
	}
	if _, err := p.DecryptValue(&userRow{Phone: "not-envelope"}); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("decrypt failure: %v", err)
	}
	legacyEnvelope := *p
	legacyEnvelope.legacy = func(Mode, string) (string, error) { return "legacy", nil }
	if _, err := legacyEnvelope.DecryptValue(&userRow{Phone: "v1:k9:abc"}); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("legacy should not handle parseable envelope: %v", err)
	}
	badMapMode := *p
	badMapMode.byColumn = map[string]fieldBinding{"phone": {column: "phone", mode: Mode(99)}}
	if _, err := badMapMode.EncryptValue(map[string]any{"phone": "138"}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad map encrypt mode: %v", err)
	}
	if _, err := badMapMode.DecryptValue(map[string]any{"phone": "v1:k1:abc"}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad map decrypt mode: %v", err)
	}
	badMapMaterial := *p
	badMapMaterial.ring = KeyRing{ActiveID: "bad", Keys: map[string][]byte{"bad": []byte("short")}}
	if _, err := badMapMaterial.EncryptValue(map[string]any{"phone": "138"}); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad map material: %v", err)
	}
}
