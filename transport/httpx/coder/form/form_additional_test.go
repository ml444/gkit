package form

import (
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/errorx"
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
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCoderFormPlainStruct(t *testing.T) {
	type query struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	c := GetCoder()
	data, err := c.Marshal(query{Name: "neo", Age: 7})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != "age=7&name=neo" && string(data) != "name=neo&age=7" {
		t.Fatalf("data = %q", data)
	}
	var out query
	if err := c.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out != (query{Name: "neo", Age: 7}) {
		t.Fatalf("out = %#v", out)
	}
	if got := c.Name(); got != Name {
		t.Fatalf("name = %q", got)
	}
	var bad struct{}
	if err := c.Unmarshal([]byte("%zz"), &bad); err == nil {
		t.Fatal("expected query parse error")
	}
	protoData, err := c.Marshal(&errorx.ErrorInfo{Status: 400, Message: "bad"})
	if err != nil || len(protoData) == 0 {
		t.Fatalf("proto marshal = %q %v", protoData, err)
	}
	var protoOut errorx.ErrorInfo
	if err := c.Unmarshal(protoData, &protoOut); err != nil || protoOut.Status != 400 {
		t.Fatalf("proto unmarshal = %#v %v", protoOut, err)
	}
}

func TestEncodeDecodeProtoValues(t *testing.T) {
	msg := &errorx.ErrorInfo{
		Status:   400,
		Code:     1001,
		Message:  "bad",
		Metadata: map[string]string{"trace": "abc"},
	}
	values, err := EncodeValues(msg)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if values.Get("status") != "400" || values.Get("metadata[trace]") != "abc" {
		t.Fatalf("values = %#v", values)
	}
	var out errorx.ErrorInfo
	if err := DecodeValues(&out, values); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Status != 400 || out.Code != 1001 || out.Message != "bad" || out.Metadata["trace"] != "abc" {
		t.Fatalf("out = %#v", out)
	}
}

func TestDecodePrimitiveProtoFields(t *testing.T) {
	var out pagination.Pagination
	err := DecodeValues(&out, url.Values{
		"page":      {"3"},
		"size":      {"20"},
		"total":     {"99"},
		"skipCount": {"true"},
	})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Page != 3 || out.Size != 20 || out.Total != 99 || !out.SkipCount {
		t.Fatalf("out = %#v", out)
	}
}

func TestParseMessageWellKnownTypes(t *testing.T) {
	tests := []struct {
		name  string
		msg   protoreflect.ProtoMessage
		value string
	}{
		{"timestamp", &timestamppb.Timestamp{}, "2022-01-02T03:04:05Z"},
		{"duration", &durationpb.Duration{}, "2s"},
		{"double", &wrapperspb.DoubleValue{}, "1.5"},
		{"float", &wrapperspb.FloatValue{}, "1.25"},
		{"int64", &wrapperspb.Int64Value{}, "-9"},
		{"int32", &wrapperspb.Int32Value{}, "-8"},
		{"uint64", &wrapperspb.UInt64Value{}, "9"},
		{"uint32", &wrapperspb.UInt32Value{}, "8"},
		{"bool", &wrapperspb.BoolValue{}, "true"},
		{"string", &wrapperspb.StringValue{}, "hello"},
		{"bytes", &wrapperspb.BytesValue{}, base64.StdEncoding.EncodeToString([]byte("abc"))},
		{"fieldmask", &fieldmaskpb.FieldMask{}, "displayName,age"},
		{"value", &structpb.Value{}, "text"},
		{"struct", &structpb.Struct{}, `{"name":"neo"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := tt.msg.ProtoReflect().Descriptor()
			if _, err := parseMessage(md, tt.value); err != nil {
				t.Fatalf("parseMessage: %v", err)
			}
		})
	}
	if _, err := parseMessage((&timestamppb.Timestamp{}).ProtoReflect().Descriptor(), "bad-time"); err == nil {
		t.Fatal("expected timestamp error")
	}
	if _, err := parseMessage((&durationpb.Duration{}).ProtoReflect().Descriptor(), "bad-duration"); err == nil {
		t.Fatal("expected duration error")
	}
	if _, err := parseMessage((&wrapperspb.BytesValue{}).ProtoReflect().Descriptor(), "bad-bytes"); err == nil {
		t.Fatal("expected bytes error")
	}
	for _, msg := range []proto.Message{
		wrapperspb.Double(0),
		wrapperspb.Float(0),
		wrapperspb.Int64(0),
		wrapperspb.Int32(0),
		wrapperspb.UInt64(0),
		wrapperspb.UInt32(0),
		wrapperspb.Bool(false),
	} {
		if _, err := parseMessage(msg.ProtoReflect().Descriptor(), "bad"); err == nil {
			t.Fatalf("expected parseMessage error for %s", msg.ProtoReflect().Descriptor().FullName())
		}
	}
	if _, err := parseMessage((&structpb.Struct{}).ProtoReflect().Descriptor(), "bad-json"); err == nil {
		t.Fatal("expected struct error")
	}
}

func TestEncodeMessageWellKnownTypes(t *testing.T) {
	tests := []struct {
		name string
		msg  proto.Message
	}{
		{"timestamp", timestamppb.New(time.Date(2022, 1, 2, 3, 4, 5, 6, time.UTC))},
		{"duration", durationpb.New(2 * time.Second)},
		{"bytes", wrapperspb.Bytes([]byte("abc"))},
		{"double", wrapperspb.Double(1.5)},
		{"float", wrapperspb.Float(1.25)},
		{"int64", wrapperspb.Int64(-9)},
		{"int32", wrapperspb.Int32(-8)},
		{"uint64", wrapperspb.UInt64(9)},
		{"uint32", wrapperspb.UInt32(8)},
		{"bool", wrapperspb.Bool(true)},
		{"string", wrapperspb.String("hello")},
		{"fieldmask", &fieldmaskpb.FieldMask{Paths: []string{"display_name", "age"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.msg.ProtoReflect()
			if _, err := encodeMessage(msg.Descriptor(), protoreflect.ValueOfMessage(msg)); err != nil {
				t.Fatalf("encodeMessage: %v", err)
			}
		})
	}
	if _, err := encodeMessage((&structpb.Struct{}).ProtoReflect().Descriptor(), protoreflect.ValueOfMessage((&structpb.Struct{}).ProtoReflect())); err == nil {
		t.Fatal("expected unsupported message error")
	}
}

func TestEncodeFieldBranches(t *testing.T) {
	p := (&pagination.Pagination{Page: 3, Total: -7, SkipCount: true}).ProtoReflect()
	fields := p.Descriptor().Fields()
	for _, name := range []protoreflect.Name{"page", "total", "skip_count"} {
		fd := fields.ByName(name)
		if got, err := EncodeField(fd, p.Get(fd)); err != nil || got == "" {
			t.Fatalf("EncodeField(%s) = %q, %v", name, got, err)
		}
	}

	e := (&errorx.ErrorInfo{Message: "bad"}).ProtoReflect()
	fd := e.Descriptor().Fields().ByName("message")
	if got, err := EncodeField(fd, e.Get(fd)); err != nil || got != "bad" {
		t.Fatalf("string field = %q, %v", got, err)
	}

	b := wrapperspb.Bytes([]byte("abc")).ProtoReflect()
	fd = b.Descriptor().Fields().ByName("value")
	if got, err := EncodeField(fd, b.Get(fd)); err != nil || got == "" {
		t.Fatalf("bytes field = %q, %v", got, err)
	}

	for _, msg := range []proto.Message{
		wrapperspb.Double(1.5),
		wrapperspb.Float(1.25),
		wrapperspb.Int64(-9),
		wrapperspb.Int32(-8),
		wrapperspb.UInt64(9),
		wrapperspb.UInt32(8),
		wrapperspb.Bool(true),
		wrapperspb.String("hello"),
	} {
		m := msg.ProtoReflect()
		fd := m.Descriptor().Fields().ByName("value")
		if got, err := EncodeField(fd, m.Get(fd)); err != nil || got == "" {
			t.Fatalf("EncodeField(%s) = %q, %v", fd.FullName(), got, err)
		}
	}
	value := structpb.NewNullValue().ProtoReflect()
	fd = value.Descriptor().Fields().ByName("null_value")
	if got, err := EncodeField(fd, value.Get(fd)); err != nil || got != nullStr {
		t.Fatalf("enum field = %q, %v", got, err)
	}
	if _, err := parseField(fd, "0"); err != nil {
		t.Fatalf("numeric enum: %v", err)
	}
	if _, err := parseField(fd, "NO_SUCH_ENUM"); err == nil {
		t.Fatal("expected enum parse error")
	}

	fm := &fieldmaskpb.FieldMask{Paths: []string{"display_name", "age"}}
	fd = fm.ProtoReflect().Descriptor().Fields().ByName("paths")
	values, err := encodeRepeatedField(fd, fm.ProtoReflect().Get(fd).List())
	if err != nil || len(values) != 2 {
		t.Fatalf("encode repeated = %#v %v", values, err)
	}
}

func TestPopulateRepeatedAndParseFieldBranches(t *testing.T) {
	lv := &structpb.ListValue{}
	fd := lv.ProtoReflect().Descriptor().Fields().ByName("values")
	if err := populateRepeatedField(fd, lv.ProtoReflect().Mutable(fd).List(), []string{"a", "b"}); err != nil {
		t.Fatalf("populate repeated: %v", err)
	}
	if len(lv.Values) != 2 {
		t.Fatalf("list len = %d", len(lv.Values))
	}

	b := &wrapperspb.BytesValue{}
	fd = b.ProtoReflect().Descriptor().Fields().ByName("value")
	if _, err := parseField(fd, base64.StdEncoding.EncodeToString([]byte("abc"))); err != nil {
		t.Fatalf("parse bytes: %v", err)
	}
	if _, err := parseField(fd, "not-base64"); err == nil {
		t.Fatal("expected bytes parse error")
	}
	value := &structpb.Value{}
	fd = value.ProtoReflect().Descriptor().Fields().ByName("null_value")
	if _, err := parseField(fd, "NULL_VALUE"); err != nil {
		t.Fatalf("parse enum: %v", err)
	}

	for _, tt := range []struct {
		msg   proto.Message
		value string
	}{
		{wrapperspb.Double(0), "1.5"},
		{wrapperspb.Float(0), "1.25"},
		{wrapperspb.Int64(0), "-9"},
		{wrapperspb.Int32(0), "-8"},
		{wrapperspb.UInt64(0), "9"},
		{wrapperspb.UInt32(0), "8"},
		{wrapperspb.Bool(false), "true"},
		{wrapperspb.String(""), "hello"},
	} {
		m := tt.msg.ProtoReflect()
		fd := m.Descriptor().Fields().ByName("value")
		if _, err := parseField(fd, tt.value); err != nil {
			t.Fatalf("parseField(%s): %v", fd.FullName(), err)
		}
	}
	for _, tt := range []struct {
		msg   proto.Message
		value string
	}{
		{wrapperspb.Double(0), "bad"},
		{wrapperspb.Float(0), "bad"},
		{wrapperspb.Int64(0), "bad"},
		{wrapperspb.Int32(0), "bad"},
		{wrapperspb.UInt64(0), "bad"},
		{wrapperspb.UInt32(0), "bad"},
		{wrapperspb.Bool(false), "bad"},
	} {
		m := tt.msg.ProtoReflect()
		fd := m.Descriptor().Fields().ByName("value")
		if _, err := parseField(fd, tt.value); err == nil {
			t.Fatalf("expected parse error for %s", fd.FullName())
		}
	}
}

func TestEncodeValuesAndFieldMask(t *testing.T) {
	if values, err := EncodeValues(nil); err != nil || len(values) != 0 {
		t.Fatalf("nil values = %#v %v", values, err)
	}
	var nilMsg *errorx.ErrorInfo
	if values, err := EncodeValues(nilMsg); err != nil || len(values) != 0 {
		t.Fatalf("nil proto values = %#v %v", values, err)
	}
	type plain struct {
		Name string `json:"name"`
	}
	if values, err := EncodeValues(plain{Name: "neo"}); err != nil || values.Get("name") != "neo" {
		t.Fatalf("plain values = %#v %v", values, err)
	}
	fm := &fieldmaskpb.FieldMask{Paths: []string{"display_name", "age"}}
	if values, err := EncodeValues(fm); err != nil || values.Get("paths") == "" {
		t.Fatalf("field mask values = %#v %v", values, err)
	}

	msg := newDynamicFieldMaskMessage(t)
	if query := EncodeFieldMask(msg); query != "mask=displayName,age" {
		t.Fatalf("field mask query = %q", query)
	}
	maskFD := msg.Descriptor().Fields().ByName("mask")
	if got, err := EncodeField(maskFD, msg.Get(maskFD)); err != nil || got != "displayName,age" {
		t.Fatalf("message field = %q %v", got, err)
	}
	values, err := EncodeValues(msg.Interface())
	if err != nil {
		t.Fatalf("dynamic encode: %v", err)
	}
	if values.Get("tags") != "a" || values.Get("mask") != "displayName,age" {
		t.Fatalf("dynamic values = %#v", values)
	}
	if err := DecodeValues(msg.Interface(), url.Values{"mask": {"displayName,age"}}); err != nil {
		t.Fatalf("dynamic decode mask: %v", err)
	}
	if err := DecodeValues(msg.Interface(), url.Values{"tags": {"a", "b"}}); err != nil {
		t.Fatalf("dynamic decode tags: %v", err)
	}
	if err := DecodeValues(msg.Interface(), url.Values{"tags.name": {"bad"}}); err == nil {
		t.Fatal("expected invalid nested repeated path")
	}
}

func TestFormErrorPathsAndCaseHelpers(t *testing.T) {
	if err := populateFieldValues((&pagination.Pagination{}).ProtoReflect(), nil, []string{"1"}); err == nil {
		t.Fatal("expected empty field path error")
	}
	if _, _, err := parseURLQueryMapKey("bad"); err == nil {
		t.Fatal("expected invalid map key")
	}
	if field, key, err := parseURLQueryMapKey("metadata.trace"); err != nil || field != "metadata" || key != "trace" {
		t.Fatalf("dot map key = %q %q %v", field, key, err)
	}
	if got := jsonCamelCase("foo_bar_baz"); got != "fooBarBaz" {
		t.Fatalf("camel = %q", got)
	}
	if got := jsonSnakeCase("fooBarBaz"); got != "foo_bar_baz" {
		t.Fatalf("snake = %q", got)
	}
	if err := DecodeValues(&pagination.Pagination{}, url.Values{"page": {}}); err == nil {
		t.Fatal("expected empty value error")
	}
	if err := DecodeValues(&pagination.Pagination{}, url.Values{"page": {"bad"}}); err == nil {
		t.Fatal("expected parse error")
	}
	if err := DecodeValues(&errorx.ErrorInfo{}, url.Values{"metadata[trace]": {"ok"}}); err != nil {
		t.Fatalf("map decode: %v", err)
	}
	md := (&errorx.ErrorInfo{}).ProtoReflect()
	mapFD := md.Descriptor().Fields().ByName("metadata")
	if err := populateMapField(mapFD, md.Mutable(mapFD).Map(), []string{"metadata[bad"}, []string{"bad"}); err == nil {
		t.Fatal("expected map key error")
	}
	if _, err := marshalTimestamp((&timestamppb.Timestamp{Seconds: maxTimestampSeconds + 1}).ProtoReflect()); err == nil {
		t.Fatal("expected timestamp range error")
	}
	if _, err := marshalTimestamp((&timestamppb.Timestamp{Nanos: secondsInNanos + 1}).ProtoReflect()); err == nil {
		t.Fatal("expected timestamp nanos error")
	}
	if got, err := marshalDuration((&durationpb.Duration{Seconds: 1 << 62, Nanos: secondsInNanos}).ProtoReflect()); err != nil || got == "" {
		t.Fatalf("overflow duration = %q %v", got, err)
	}
}

func newDynamicFieldMaskMessage(t *testing.T) protoreflect.Message {
	t.Helper()
	labelOptional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	labelRepeated := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	typeMessage := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	typeString := descriptorpb.FieldDescriptorProto_TYPE_STRING
	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Syntax:     proto.String("proto3"),
		Name:       proto.String("gkit_transport_form_test.proto"),
		Package:    proto.String("gkit.transport.formtest"),
		Dependency: []string{"google/protobuf/field_mask.proto"},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("Query"),
			Field: []*descriptorpb.FieldDescriptorProto{
				{
					Name:   proto.String("tags"),
					Number: proto.Int32(1),
					Label:  &labelRepeated,
					Type:   &typeString,
				},
				{
					Name:     proto.String("mask"),
					JsonName: proto.String("mask"),
					Number:   proto.Int32(2),
					Label:    &labelOptional,
					Type:     &typeMessage,
					TypeName: proto.String(".google.protobuf.FieldMask"),
				},
			},
		}},
	}, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}
	md := file.Messages().ByName("Query")
	msg := dynamicpb.NewMessage(md)
	tags := msg.Mutable(md.Fields().ByName("tags")).List()
	tags.Append(protoreflect.ValueOfString("a"))
	tags.Append(protoreflect.ValueOfString("b"))
	mask := &fieldmaskpb.FieldMask{Paths: []string{"display_name", "age"}}
	msg.Set(md.Fields().ByName("mask"), protoreflect.ValueOfMessage(mask.ProtoReflect()))
	return msg
}
