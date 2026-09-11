package json

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConfiguredCoderCopiesOptionsAndGlobals(t *testing.T) {
	original := MarshalOptions
	t.Cleanup(func() { MarshalOptions = original })
	opts := DefaultOptions()
	opts.Marshal.UseProtoNames = true
	opts.Marshal.UseEnumNumbers = true
	c := NewCoder(opts)
	opts.Marshal.UseProtoNames = false
	opts.Marshal.UseEnumNumbers = false
	MarshalOptions.UseProtoNames = false
	MarshalOptions.UseEnumNumbers = false

	data, err := c.Marshal(&descriptorpb.FieldDescriptorProto{
		JsonName: proto.String("displayName"),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if fields["json_name"] != "displayName" || fields["type"] != float64(9) {
		t.Fatalf("instance options were changed: %s", data)
	}
}

func TestConfiguredCoderUnmarshalIsolation(t *testing.T) {
	looseOptions := DefaultOptions()
	looseOptions.Unmarshal.DiscardUnknown = true
	strictOptions := looseOptions
	strictOptions.Unmarshal.DiscardUnknown = false
	loose, strict := NewCoder(looseOptions), NewCoder(strictOptions)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				for _, tc := range []struct {
					c       *ConfiguredCoder
					wantErr bool
				}{{loose, false}, {strict, true}} {
					var out *descriptorpb.FieldDescriptorProto
					err := tc.c.Unmarshal([]byte(`{"name":"field","unknown":true}`), &out)
					if (err != nil) != tc.wantErr {
						t.Errorf("DiscardUnknown isolation: err=%v, wantErr=%v", err, tc.wantErr)
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestConfiguredCoderGoValuesAndInvalidTargets(t *testing.T) {
	c := NewCoder(DefaultOptions())
	if c.Name() != Name {
		t.Fatal(c.Name())
	}
	data, err := c.Marshal(customJSON{value: "custom"})
	if err != nil || string(data) != `"custom"` {
		t.Fatalf("custom marshal: %s, %v", data, err)
	}
	var custom customJSON
	if err := c.Unmarshal(data, &custom); err != nil || custom.value != "custom" {
		t.Fatalf("custom unmarshal: %+v, %v", custom, err)
	}
	var out struct{ Value string }
	if err := c.Unmarshal([]byte(`{"Value":"ok","extra":1}`), &out); err != nil || out.Value != "ok" {
		t.Fatalf("ordinary Go values: %+v, %v", out, err)
	}
	if err := c.Unmarshal([]byte(`{} {}`), &out); err == nil {
		t.Fatal("accepted trailing JSON value")
	}
	for _, target := range []any{nil, 1, struct{}{}, (*wrapperspb.StringValue)(nil)} {
		var invalid *json.InvalidUnmarshalError
		if err := c.Unmarshal([]byte(`{}`), target); !errors.As(err, &invalid) {
			t.Errorf("target %T: expected InvalidUnmarshalError, got %v", target, err)
		}
	}
}

func TestConfiguredCoderZeroOptionsDoNotInheritGlobals(t *testing.T) {
	c := NewCoder(Options{})
	data, err := c.Marshal(&descriptorpb.FieldDescriptorProto{})
	if err != nil || string(data) != "{}" {
		t.Fatalf("zero marshal options: %s, %v", data, err)
	}
	if err := c.Unmarshal([]byte(`{"unknown":true}`), &descriptorpb.FieldDescriptorProto{}); err == nil {
		t.Fatal("zero unmarshal options should not discard unknown protobuf fields")
	}
}
