package form

import (
	"net/url"
	"reflect"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/typepb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestBytesAndUnknownEnumRoundTrip(t *testing.T) {
	for _, encoded := range []string{"+/8=", "-_8="} {
		out := new(wrapperspb.BytesValue)
		if err := DecodeValues(out, url.Values{"value": {encoded}}); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out.Value, []byte{0xfb, 0xff}) {
			t.Fatalf("bytes=%v", out.Value)
		}
	}
	in := &typepb.Type{Syntax: typepb.Syntax(777)}
	values, err := EncodeValues(in)
	if err != nil {
		t.Fatal(err)
	}
	out := new(typepb.Type)
	if err = DecodeValues(out, values); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(in, out) {
		t.Fatalf("enum lost: %v", out)
	}
}

func TestNullTemporalFieldsClearWithoutPanic(t *testing.T) {
	fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{Name: proto.String("temporal.proto"), Package: proto.String("formtest"), Syntax: proto.String("proto3"), Dependency: []string{"google/protobuf/timestamp.proto", "google/protobuf/duration.proto"}, MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Request"), Field: []*descriptorpb.FieldDescriptorProto{
		{Name: proto.String("at"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".google.protobuf.Timestamp")},
		{Name: proto.String("duration"), Number: proto.Int32(2), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".google.protobuf.Duration")},
	}}}}, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	m := dynamicpb.NewMessage(fd.Messages().Get(0))
	at := m.Descriptor().Fields().Get(0)
	duration := m.Descriptor().Fields().Get(1)
	m.Set(at, protoreflect.ValueOfMessage((&timestamppb.Timestamp{Seconds: 1}).ProtoReflect()))
	m.Set(duration, protoreflect.ValueOfMessage((&durationpb.Duration{Seconds: 1}).ProtoReflect()))
	if err = DecodeValues(m, url.Values{"at": {"null"}, "duration": {"null"}}); err != nil {
		t.Fatal(err)
	}
	if m.Has(at) || m.Has(duration) {
		t.Fatal("null did not clear fields")
	}
	if err = DecodeValues(m, url.Values{"at": {"2026-09-10T00:00:00Z"}, "duration": {"3s"}}); err != nil {
		t.Fatal(err)
	}
	if !m.Has(at) || !m.Has(duration) {
		t.Fatal("valid temporal fields not populated")
	}
	for _, key := range []string{"at", "duration"} {
		if err = DecodeValues(m, url.Values{key: {"invalid"}}); err == nil {
			t.Fatalf("invalid %s accepted", key)
		}
	}
}

func TestMapValueEncodingErrorIsReturned(t *testing.T) {
	m, err := structpb.NewStruct(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = EncodeValues(m); err == nil {
		t.Fatal("unsupported map value silently dropped")
	}
}

func TestFieldMaskConcurrentEncodingPreservesInput(t *testing.T) {
	m := newDynamicFieldMaskMessage(t)
	mask := m.Get(m.Descriptor().Fields().ByName("mask")).Message().Interface().(*fieldmaskpb.FieldMask)
	original := append([]string(nil), mask.Paths...)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := EncodeValues(m.Interface()); err != nil {
				t.Error(err)
			}
			if got := EncodeFieldMask(m); got != "mask=displayName%2Cage" {
				t.Errorf("mask=%s", got)
			}
		}()
	}
	wg.Wait()
	if !reflect.DeepEqual(mask.Paths, original) {
		t.Fatalf("input mutated: %v", mask.Paths)
	}
}
