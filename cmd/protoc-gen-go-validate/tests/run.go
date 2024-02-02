package main

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests/cases"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests/cases/other_package"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests/cases/sort"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests/cases/yet_another_package"
)

type validator interface {
	Validate() error
	ValidateAll() error
}

func validateData(correctVal, errorVal interface{}, errCount int) {
	m, ok := correctVal.(validator)
	if ok {
		if err := m.Validate(); err != nil {
			panic(err)
		}
	}

	errM, ok1 := errorVal.(validator)
	if ok1 {
		multiErr := errM.ValidateAll()
		if multiErr == nil && errCount > 0 {
			panic(fmt.Sprintf("%T expected error, got nil", errorVal))
		}
		errs, ok2 := multiErr.(cases.MultiError)
		if ok2 && len(errs.AllErrors()) != errCount {
			panic(fmt.Sprintf("expected %d errors, got %d", errCount, len(errs.AllErrors())))
		}
	}

	return
}

func main() {
	validateData(&cases.BoolConstTrue{Val: true}, &cases.BoolConstTrue{Val: false}, 1)
	validateData(&cases.BoolConstFalse{Val: false}, &cases.BoolConstFalse{Val: true}, 1)

	validateData(&cases.BytesConst{Val: []byte("foo")}, &cases.BytesConst{Val: []byte("bar")}, 1)
	validateData(&cases.BytesIn{Val: []byte("bar")}, &cases.BytesIn{Val: []byte("foo")}, 1)
	validateData(&cases.BytesNotIn{Val: []byte("foo")}, &cases.BytesNotIn{Val: []byte("fizz")}, 1)
	validateData(&cases.BytesLen{Val: []byte("foo")}, &cases.BytesLen{Val: []byte("foobar")}, 1)
	validateData(&cases.BytesMinLen{Val: []byte("foo")}, &cases.BytesMinLen{Val: []byte("fo")}, 1)
	validateData(&cases.BytesMaxLen{Val: []byte("foo")}, &cases.BytesMaxLen{Val: []byte("foobar")}, 1)
	validateData(&cases.BytesMinMaxLen{Val: []byte("foo")}, &cases.BytesMinMaxLen{Val: []byte("fo")}, 1)
	validateData(&cases.BytesEqualMinMaxLen{Val: []byte("foo11")}, &cases.BytesEqualMinMaxLen{Val: []byte("bar")}, 1)
	validateData(&cases.BytesPattern{Val: []byte("\x00\x7F")}, &cases.BytesPattern{Val: []byte("\xFF")}, 1)
	validateData(&cases.BytesPrefix{Val: []byte("\x99\x00")}, &cases.BytesPrefix{Val: []byte("\x00")}, 1)
	validateData(&cases.BytesContains{Val: []byte("\x00bar\x99")}, &cases.BytesContains{Val: []byte("\x00\x99")}, 1)
	validateData(&cases.BytesSuffix{Val: []byte("\x00\x99buz\x7a")}, &cases.BytesSuffix{Val: []byte("\x00buz")}, 1)
	//validateData(&cases.BytesIP{Val: []byte{0x20, 0x01, 0x48, 0x60, 0, 0, 0x20, 0x01, 0, 0, 0, 0, 0, 0, 0x00, 0x68}}, &cases.BytesIP{Val: []byte("foo")}, 1)
	//validateData(&cases.BytesIPv4{Val: []byte("127.0.0.1")}, &cases.BytesIPv4{Val: []byte("foo")}, 1)
	//validateData(&cases.BytesIPv6{Val: []byte("::1")}, &cases.BytesIPv6{Val: []byte("foo")}, 1)
	//validateData(&cases.BytesIPv6Ignore{Val: []byte("::1")}, &cases.BytesIPv6Ignore{Val: []byte("foo")}, 1)
	validateData(&cases.BytesIPv6Ignore{Val: []byte{}}, &cases.BytesIPv6Ignore{Val: []byte("foo")}, 1)

	// enums.proto
	validateData(&cases.EnumConst{Val: cases.TestEnum_TWO}, &cases.EnumConst{Val: cases.TestEnum_ZERO}, 1)
	validateData(&cases.EnumAliasConst{Val: cases.TestEnumAlias_C}, &cases.EnumAliasConst{Val: cases.TestEnumAlias_A}, 1)
	validateData(&cases.EnumDefined{Val: cases.TestEnum_ZERO}, &cases.EnumDefined{Val: cases.TestEnum(99)}, 1)
	validateData(&cases.EnumAliasDefined{Val: cases.TestEnumAlias_A}, &cases.EnumAliasDefined{Val: cases.TestEnumAlias(99)}, 1)
	validateData(&cases.EnumIn{Val: cases.TestEnum_TWO}, &cases.EnumIn{Val: cases.TestEnum_ONE}, 1)
	validateData(&cases.EnumAliasIn{Val: cases.TestEnumAlias_C}, &cases.EnumAliasIn{Val: cases.TestEnumAlias_B}, 1)
	validateData(&cases.EnumNotIn{Val: cases.TestEnum_TWO}, &cases.EnumNotIn{Val: cases.TestEnum_ONE}, 1)
	validateData(&cases.EnumAliasNotIn{Val: cases.TestEnumAlias_GAMMA}, &cases.EnumAliasNotIn{Val: cases.TestEnumAlias_BETA}, 1)
	validateData(&cases.EnumExternal{Val: other_package.Embed_VALUE}, &cases.EnumExternal{Val: other_package.Embed_Enumerated(999)}, 1)
	validateData(&cases.EnumExternal2{Val: other_package.Embed_DoubleEmbed_VALUE}, &cases.EnumExternal2{Val: other_package.Embed_DoubleEmbed_DoubleEnumerated(999)}, 1)
	validateData(
		&cases.EnumExternal3{Foo: other_package.Embed_TWO, Bar: yet_another_package.Embed_ZERO},
		&cases.EnumExternal3{Foo: other_package.Embed_ONE, Bar: yet_another_package.Embed_ONE},
		2,
	)
	validateData(&cases.EnumExternal4{SortDirection: sort.Direction_ASC}, &cases.EnumExternal4{SortDirection: sort.Direction_DESC}, 1)
	validateData(&cases.RepeatedEnumDefined{Val: []cases.TestEnum{cases.TestEnum_ONE, cases.TestEnum_TWO}}, &cases.RepeatedEnumDefined{Val: []cases.TestEnum{cases.TestEnum(99), cases.TestEnum(999)}}, 2)
	validateData(&cases.RepeatedExternalEnumDefined{Val: []other_package.Embed_Enumerated{other_package.Embed_VALUE}}, &cases.RepeatedExternalEnumDefined{Val: []other_package.Embed_Enumerated{other_package.Embed_Enumerated(999)}}, 1)
	validateData(&cases.RepeatedYetAnotherExternalEnumDefined{Val: []yet_another_package.Embed_Enumerated{yet_another_package.Embed_VALUE}}, &cases.RepeatedYetAnotherExternalEnumDefined{Val: []yet_another_package.Embed_Enumerated{yet_another_package.Embed_Enumerated(999)}}, 1)

	validateData(
		&cases.RepeatedEnumExternal{Foo: []other_package.Embed_FooNumber{other_package.Embed_ZERO, other_package.Embed_TWO}, Bar: []yet_another_package.Embed_BarNumber{yet_another_package.Embed_ZERO, yet_another_package.Embed_TWO}},
		&cases.RepeatedEnumExternal{Foo: []other_package.Embed_FooNumber{other_package.Embed_ZERO, other_package.Embed_ONE}, Bar: []yet_another_package.Embed_BarNumber{yet_another_package.Embed_ZERO, yet_another_package.Embed_ONE}},
		2,
	)
	validateData(&cases.MapEnumDefined{Val: map[string]cases.TestEnum{"foo": cases.TestEnum_ONE, "bar": cases.TestEnum_TWO}}, &cases.MapEnumDefined{Val: map[string]cases.TestEnum{"foo": cases.TestEnum(99), "bar": cases.TestEnum(999)}}, 2)
	validateData(&cases.MapExternalEnumDefined{Val: map[string]other_package.Embed_Enumerated{"foo": other_package.Embed_VALUE}}, &cases.MapExternalEnumDefined{Val: map[string]other_package.Embed_Enumerated{"foo": other_package.Embed_Enumerated(999)}}, 1)
	validateData(
		&cases.EnumInsideOneOf{Foo: &cases.EnumInsideOneOf_Val{Val: cases.TestEnum_ZERO}, Bar: &cases.EnumInsideOneOf_Val21{Val21: cases.TestEnum_TWO}},
		&cases.EnumInsideOneOf{Foo: &cases.EnumInsideOneOf_Val{Val: cases.TestEnum(99)}, Bar: &cases.EnumInsideOneOf_Val21{Val21: cases.TestEnum(999)}},
		2,
	)
	validateData(
		&cases.EnumInsideOneOf{Foo: &cases.EnumInsideOneOf_Val{Val: cases.TestEnum_ZERO}, Bar: &cases.EnumInsideOneOf_Val21{Val21: cases.TestEnum_TWO}},
		&cases.EnumInsideOneOf{Foo: &cases.EnumInsideOneOf_Val{Val: cases.TestEnum(99)}, Bar: &cases.EnumInsideOneOf_Val21{Val21: cases.TestEnum_ZERO}},
		2,
	)

	// kitchen_sink.proto
	complexTestMsgCorretVal := &cases.ComplexTestMsg{
		Const: "abcd",
		Nested: &cases.ComplexTestMsg{
			Const:     "abcd",
			IntConst:  5,
			DurVal:    &durationpb.Duration{Seconds: 16},
			DoubleIn:  123,
			EnumConst: cases.ComplexTestEnum_ComplexTWO,
			BytesVal:  []byte("\x00\x99"),
			O:         &cases.ComplexTestMsg_Y{Y: 123},
		},
		IntConst:  5,
		BoolConst: false,
		FloatVal:  wrapperspb.Float(1.23),
		DurVal: &durationpb.Duration{
			Seconds: 16,
			Nanos:   0,
		},
		TsVal: &timestamppb.Timestamp{
			Seconds: 8,
			Nanos:   0,
		},
		Another:    nil,
		FloatConst: 7,
		DoubleIn:   456.789,
		EnumConst:  cases.ComplexTestEnum_ComplexTWO,
		AnyVal: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Duration",
			Value:   nil,
		},
		RepTsVal: []*timestamppb.Timestamp{{
			Seconds: 123,
			Nanos:   1000000,
		}},
		MapVal:   map[int32]string{-1: "value1"},
		BytesVal: []byte("\x00\x99"),
		O:        &cases.ComplexTestMsg_X{X: ""},
	}
	complexTestMsgErrVal := &cases.ComplexTestMsg{
		Const:     "abc",
		Nested:    &cases.ComplexTestMsg{},
		IntConst:  4,
		BoolConst: true,
		FloatVal:  wrapperspb.Float(0),
		DurVal: &durationpb.Duration{
			Seconds: 17,
			Nanos:   0,
		},
		TsVal: &timestamppb.Timestamp{
			Seconds: 7,
			Nanos:   0,
		},
		Another:    &cases.ComplexTestMsg{},
		FloatConst: 8,
		DoubleIn:   3456.789,
		EnumConst:  cases.ComplexTestEnum_ComplexONE,
		AnyVal: &anypb.Any{
			TypeUrl: "",
			Value:   nil,
		},
		RepTsVal: []*timestamppb.Timestamp{{
			Seconds: 0,
			Nanos:   100000,
		}},
		MapVal:   map[int32]string{0: "0"},
		BytesVal: []byte("\x01\x99"),
		//O:        &cases.ComplexTestMsg_X{X: ""},
	}
	validateData(complexTestMsgCorretVal, complexTestMsgErrVal, 16)

	// maps.proto
	validateData(&cases.MapMin{Val: map[int32]float32{123: 456.789, 456: 789.123}}, &cases.MapMin{Val: map[int32]float32{123: 456.789}}, 1)
	validateData(&cases.MapMax{Val: map[int64]float64{123: 456.789, 456: 789.123}}, &cases.MapMax{Val: map[int64]float64{123: 456.789, 123456: 78.9, 123456789: 123.456789, 1: 456}}, 1)
	validateData(&cases.MapMinMax{Val: map[string]bool{"foo": true, "bar": false}}, &cases.MapMinMax{Val: map[string]bool{"foo": true}}, 1)
	validateData(&cases.MapMinMax{Val: map[string]bool{"foo": true, "bar": false, "zoo": false, "ya": true}}, &cases.MapMinMax{Val: map[string]bool{"foo": true, "bar": false, "zoo": false, "ya": true, "x": false}}, 1)
	validateData(&cases.MapExact{Val: map[uint64]string{123: "foo", 456: "bar", 789: "zoo"}}, &cases.MapExact{Val: map[uint64]string{123: "foo"}}, 1)
	validateData(&cases.MapExact{Val: map[uint64]string{123: "foo", 456: "bar", 789: "zoo"}}, &cases.MapExact{Val: map[uint64]string{123: "foo", 123456: "bar", 123456789: "zoo", 1: "ya"}}, 1)
	validateData(&cases.MapNoSparse{Val: map[uint32]*cases.MapNoSparse_Msg{}}, &cases.MapNoSparse{Val: map[uint32]*cases.MapNoSparse_Msg{123: nil}}, 1)
	validateData(&cases.MapKeys{Val: map[int64]string{-123: "bar"}}, &cases.MapKeys{Val: map[int64]string{0: "baz"}}, 1)
	validateData(&cases.MapValues{Val: map[string]string{"123": "bar"}}, &cases.MapValues{Val: map[string]string{"0": "ya"}}, 1)
	validateData(&cases.MapKeysPattern{Val: map[string]string{"foo": "bar"}}, &cases.MapKeysPattern{Val: map[string]string{"Foo": "baz"}}, 1)
	validateData(&cases.MapValuesPattern{Val: map[string]string{"foo": "bar"}}, &cases.MapValuesPattern{Val: map[string]string{"foo": "#Baz"}}, 1)
	validateData(&cases.MapRecursive{Val: map[uint32]*cases.MapRecursive_Msg{123: {Val: "foo"}}}, &cases.MapRecursive{Val: map[uint32]*cases.MapRecursive_Msg{123: {Val: "ya"}}}, 1)
	validateData(&cases.MapExactIgnore{Val: map[uint64]string{123: "foo", 456: "bar", 789: "zoo"}}, &cases.MapExactIgnore{Val: map[uint64]string{123: "foo", 123456: "bar", 123456789: "zoo", 1: "ya"}}, 1)
	validateData(&cases.MapExactIgnore{Val: map[uint64]string{}}, &cases.MapExactIgnore{Val: map[uint64]string{123: "foo", 123456: "bar"}}, 1)
	validateData(&cases.MapKeysIn{Val: map[string]string{"foo": "123", "bar": "456"}}, &cases.MapKeysIn{Val: map[string]string{"foo1": "baz"}}, 1)
	validateData(&cases.MapKeysNotIn{Val: map[string]string{"foo1": "123", "bar1": "456"}}, &cases.MapKeysNotIn{Val: map[string]string{"foo": "123", "bar": "456"}}, 2)
	validateData(
		&cases.MultipleMaps{
			First:  map[uint32]string{1: "foo", 123: "bar"},
			Second: map[int32]bool{-1: true, -123: false},
			Third:  map[int32]bool{1: true, 123: false},
		},
		&cases.MultipleMaps{
			First:  map[uint32]string{0: "foo"},
			Second: map[int32]bool{0: true},
			Third:  map[int32]bool{0: false},
		},
		3,
	)

	// messages.proto
	validateData(&cases.TestMsg{Const: "foo", Nested: &cases.TestMsg{Const: "foo"}}, &cases.TestMsg{Const: "bar", Nested: &cases.TestMsg{Const: "bar"}}, 2)
	validateData(&cases.MessageDisabled{Val: 124}, &cases.MessageDisabled{Val: 122}, 0)
	validateData(&cases.MessageIgnored{Val: 124}, &cases.MessageIgnored{Val: 122}, 0)
	validateData(&cases.Message{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, &cases.Message{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, 1)
	validateData(&cases.MessageCrossPackage{Val: &other_package.Embed{Val: 12}}, &cases.MessageCrossPackage{Val: &other_package.Embed{Val: 0}}, 0)
	validateData(&cases.MessageSkip{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, &cases.MessageSkip{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, 0)
	validateData(&cases.MessageRequired{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, &cases.MessageRequired{Val: nil}, 1)
	validateData(&cases.MessageRequiredButOptional{Val: &cases.TestMsg{Const: "foo", Nested: nil}}, &cases.MessageRequiredButOptional{Val: &cases.TestMsg{}}, 1)
	validateData(&cases.MessageRequiredOneof{One: &cases.MessageRequiredOneof_Val{Val: &cases.TestMsg{Const: "foo", Nested: nil}}}, &cases.MessageRequiredOneof{}, 1)

	// numbers.proto

	validateData(&cases.FloatConst{Val: 1.23}, &cases.FloatConst{Val: 2.0}, 1)
	validateData(&cases.FloatIn{Val: 4.56}, &cases.FloatIn{Val: 2.0}, 1)
	validateData(&cases.FloatNotIn{Val: 1.0}, &cases.FloatNotIn{Val: 0}, 1)
	validateData(&cases.FloatLT{Val: -1.0}, &cases.FloatLT{Val: 2.0}, 1)
	validateData(&cases.FloatLTE{Val: 64.0}, &cases.FloatLTE{Val: 65.0}, 1)
	validateData(&cases.FloatGT{Val: 16.1}, &cases.FloatGT{Val: 2.0}, 1)
	validateData(&cases.FloatGTE{Val: 8.0}, &cases.FloatGTE{Val: 2.0}, 1)
	validateData(&cases.FloatGTLT{Val: 1.0}, &cases.FloatGTLT{Val: -1.0}, 1)
	validateData(&cases.FloatExLTGT{Val: -1.0}, &cases.FloatExLTGT{Val: 2.0}, 1)
	validateData(&cases.FloatGTELTE{Val: 128.0}, &cases.FloatGTELTE{Val: 127.9}, 1)
	validateData(&cases.FloatGTELTE{Val: 256.0}, &cases.FloatGTELTE{Val: 256.1}, 1)
	validateData(&cases.FloatExGTELTE{Val: -128.0}, &cases.FloatExGTELTE{Val: -127.9}, 1)
	validateData(&cases.FloatExGTELTE{Val: 256.0}, &cases.FloatExGTELTE{Val: 255.9}, 1)
	validateData(&cases.FloatIgnore{Val: 0}, &cases.FloatIgnore{}, 0)
	validateData(&cases.FloatIgnore{Val: 128.0}, &cases.FloatIgnore{Val: 1.0}, 1)

	validateData(&cases.DoubleConst{Val: 1.23}, &cases.DoubleConst{Val: 2.0}, 1)
	validateData(&cases.DoubleIn{Val: 4.56}, &cases.DoubleIn{Val: 2.0}, 1)
	validateData(&cases.DoubleNotIn{Val: 1.0}, &cases.DoubleNotIn{Val: 0}, 1)
	validateData(&cases.DoubleLT{Val: -1.0}, &cases.DoubleLT{Val: 2.0}, 1)
	validateData(&cases.DoubleLTE{Val: 64.0}, &cases.DoubleLTE{Val: 65.0}, 1)
	validateData(&cases.DoubleGT{Val: 16.1}, &cases.DoubleGT{Val: 2.0}, 1)
	validateData(&cases.DoubleGTE{Val: 8.0}, &cases.DoubleGTE{Val: 2.0}, 1)
	validateData(&cases.DoubleGTLT{Val: 1.0}, &cases.DoubleGTLT{Val: -1.0}, 1)
	validateData(&cases.DoubleExLTGT{Val: -1.0}, &cases.DoubleExLTGT{Val: 2.0}, 1)
	validateData(&cases.DoubleGTELTE{Val: 128.0}, &cases.DoubleGTELTE{Val: 127.9}, 1)
	validateData(&cases.DoubleGTELTE{Val: 256.0}, &cases.DoubleGTELTE{Val: 256.1}, 1)
	validateData(&cases.DoubleExGTELTE{Val: -128.0}, &cases.DoubleExGTELTE{Val: -127.9}, 1)
	validateData(&cases.DoubleExGTELTE{Val: 256.0}, &cases.DoubleExGTELTE{Val: 255.9}, 1)
	validateData(&cases.DoubleIgnore{Val: 0}, &cases.DoubleIgnore{}, 0)
	validateData(&cases.DoubleIgnore{Val: 128.0}, &cases.DoubleIgnore{Val: 129}, 1)

	validateData(&cases.Int32Const{Val: 1}, &cases.Int32Const{Val: 2}, 1)
	validateData(&cases.Int32In{Val: 2}, &cases.Int32In{Val: 1}, 1)
	validateData(&cases.Int32NotIn{Val: 1}, &cases.Int32NotIn{Val: 0}, 1)
	validateData(&cases.Int32LT{Val: -1}, &cases.Int32LT{Val: 2}, 1)
	validateData(&cases.Int32LTE{Val: 1}, &cases.Int32LTE{Val: 65}, 1)
	validateData(&cases.Int32GT{Val: 17}, &cases.Int32GT{Val: 16}, 1)
	validateData(&cases.Int32GTE{Val: 8}, &cases.Int32GTE{Val: 7}, 1)
	validateData(&cases.Int32GTLT{Val: 1}, &cases.Int32GTLT{Val: 0}, 1)
	validateData(&cases.Int32GTLT{Val: 9}, &cases.Int32GTLT{Val: 10}, 1)
	validateData(&cases.Int32ExLTGT{Val: -1}, &cases.Int32ExLTGT{Val: 0}, 1)
	validateData(&cases.Int32ExLTGT{Val: 11}, &cases.Int32ExLTGT{Val: 10}, 1)
	validateData(&cases.Int32GTELTE{Val: 128}, &cases.Int32GTELTE{Val: 127}, 1)
	validateData(&cases.Int32GTELTE{Val: 256}, &cases.Int32GTELTE{Val: 257}, 1)
	validateData(&cases.Int32ExGTELTE{Val: 128}, &cases.Int32ExGTELTE{Val: 129}, 1)
	validateData(&cases.Int32ExGTELTE{Val: 256}, &cases.Int32ExGTELTE{Val: 255}, 1)
	validateData(&cases.Int32Ignore{Val: 0}, &cases.Int32Ignore{Val: 1}, 1)
	validateData(&cases.Int32Ignore{Val: -129}, &cases.Int32Ignore{Val: 1}, 1)

	validateData(&cases.Int64Const{Val: 1}, &cases.Int64Const{Val: 2}, 1)
	validateData(&cases.Int64In{Val: 2}, &cases.Int64In{Val: 1}, 1)
	validateData(&cases.Int64NotIn{Val: 1}, &cases.Int64NotIn{Val: 0}, 1)
	validateData(&cases.Int64LT{Val: -1}, &cases.Int64LT{Val: 2}, 1)
	validateData(&cases.Int64LTE{Val: 1}, &cases.Int64LTE{Val: 65}, 1)
	validateData(&cases.Int64GT{Val: 17}, &cases.Int64GT{Val: 16}, 1)
	validateData(&cases.Int64GTE{Val: 8}, &cases.Int64GTE{Val: 7}, 1)
	validateData(&cases.Int64GTLT{Val: 1}, &cases.Int64GTLT{Val: 0}, 1)
	validateData(&cases.Int64GTLT{Val: 9}, &cases.Int64GTLT{Val: 10}, 1)
	validateData(&cases.Int64ExLTGT{Val: -1}, &cases.Int64ExLTGT{Val: 0}, 1)
	validateData(&cases.Int64ExLTGT{Val: 11}, &cases.Int64ExLTGT{Val: 10}, 1)
	validateData(&cases.Int64GTELTE{Val: 128}, &cases.Int64GTELTE{Val: 127}, 1)
	validateData(&cases.Int64GTELTE{Val: 256}, &cases.Int64GTELTE{Val: 257}, 1)
	validateData(&cases.Int64ExGTELTE{Val: 128}, &cases.Int64ExGTELTE{Val: 129}, 1)
	validateData(&cases.Int64ExGTELTE{Val: 256}, &cases.Int64ExGTELTE{Val: 255}, 1)
	validateData(&cases.Int64Ignore{Val: 0}, &cases.Int64Ignore{Val: 1}, 1)
	validateData(&cases.Int64Ignore{Val: -129}, &cases.Int64Ignore{Val: 1}, 1)

	validateData(&cases.UInt32Const{Val: 1}, &cases.UInt32Const{Val: 2}, 1)
	validateData(&cases.UInt32In{Val: 2}, &cases.UInt32In{Val: 1}, 1)
	validateData(&cases.UInt32NotIn{Val: 1}, &cases.UInt32NotIn{Val: 0}, 1)
	validateData(&cases.UInt32LT{Val: 4}, &cases.UInt32LT{Val: 5}, 1)
	validateData(&cases.UInt32LTE{Val: 1}, &cases.UInt32LTE{Val: 65}, 1)
	validateData(&cases.UInt32GT{Val: 17}, &cases.UInt32GT{Val: 16}, 1)
	validateData(&cases.UInt32GTE{Val: 8}, &cases.UInt32GTE{Val: 7}, 1)
	validateData(&cases.UInt32GTLT{Val: 6}, &cases.UInt32GTLT{Val: 5}, 1)
	validateData(&cases.UInt32GTLT{Val: 9}, &cases.UInt32GTLT{Val: 10}, 1)
	validateData(&cases.UInt32ExLTGT{Val: 4}, &cases.UInt32ExLTGT{Val: 5}, 1)
	validateData(&cases.UInt32ExLTGT{Val: 11}, &cases.UInt32ExLTGT{Val: 10}, 1)
	validateData(&cases.UInt32GTELTE{Val: 128}, &cases.UInt32GTELTE{Val: 127}, 1)
	validateData(&cases.UInt32GTELTE{Val: 256}, &cases.UInt32GTELTE{Val: 257}, 1)
	validateData(&cases.UInt32ExGTELTE{Val: 128}, &cases.UInt32ExGTELTE{Val: 129}, 1)
	validateData(&cases.UInt32ExGTELTE{Val: 256}, &cases.UInt32ExGTELTE{Val: 255}, 1)
	validateData(&cases.UInt32Ignore{Val: 0}, &cases.UInt32Ignore{Val: 129}, 1)
	validateData(&cases.UInt32Ignore{Val: 128}, &cases.UInt32Ignore{Val: 129}, 1)
	validateData(&cases.UInt32Ignore{Val: 256}, &cases.UInt32Ignore{Val: 255}, 1)

	validateData(&cases.UInt64Const{Val: 1}, &cases.UInt64Const{Val: 2}, 1)
	validateData(&cases.UInt64In{Val: 2}, &cases.UInt64In{Val: 1}, 1)
	validateData(&cases.UInt64NotIn{Val: 1}, &cases.UInt64NotIn{Val: 0}, 1)
	validateData(&cases.UInt64LT{Val: 4}, &cases.UInt64LT{Val: 5}, 1)
	validateData(&cases.UInt64LTE{Val: 1}, &cases.UInt64LTE{Val: 65}, 1)
	validateData(&cases.UInt64GT{Val: 17}, &cases.UInt64GT{Val: 16}, 1)
	validateData(&cases.UInt64GTE{Val: 8}, &cases.UInt64GTE{Val: 7}, 1)
	validateData(&cases.UInt64GTLT{Val: 6}, &cases.UInt64GTLT{Val: 5}, 1)
	validateData(&cases.UInt64GTLT{Val: 9}, &cases.UInt64GTLT{Val: 10}, 1)
	validateData(&cases.UInt64ExLTGT{Val: 4}, &cases.UInt64ExLTGT{Val: 5}, 1)
	validateData(&cases.UInt64ExLTGT{Val: 11}, &cases.UInt64ExLTGT{Val: 10}, 1)
	validateData(&cases.UInt64GTELTE{Val: 128}, &cases.UInt64GTELTE{Val: 127}, 1)
	validateData(&cases.UInt64GTELTE{Val: 256}, &cases.UInt64GTELTE{Val: 257}, 1)
	validateData(&cases.UInt64ExGTELTE{Val: 128}, &cases.UInt64ExGTELTE{Val: 129}, 1)
	validateData(&cases.UInt64ExGTELTE{Val: 256}, &cases.UInt64ExGTELTE{Val: 255}, 1)
	validateData(&cases.UInt64Ignore{Val: 0}, &cases.UInt64Ignore{Val: 129}, 1)
	validateData(&cases.UInt64Ignore{Val: 128}, &cases.UInt64Ignore{Val: 129}, 1)
	validateData(&cases.UInt64Ignore{Val: 256}, &cases.UInt64Ignore{Val: 255}, 1)

	validateData(&cases.SInt32Const{Val: 1}, &cases.SInt32Const{Val: 2}, 1)
	validateData(&cases.SInt32In{Val: 2}, &cases.SInt32In{Val: 1}, 1)
	validateData(&cases.SInt32NotIn{Val: 1}, &cases.SInt32NotIn{Val: 0}, 1)
	validateData(&cases.SInt32LT{Val: -1}, &cases.SInt32LT{Val: 2}, 1)
	validateData(&cases.SInt32LTE{Val: 1}, &cases.SInt32LTE{Val: 65}, 1)
	validateData(&cases.SInt32GT{Val: 17}, &cases.SInt32GT{Val: 16}, 1)
	validateData(&cases.SInt32GTE{Val: 8}, &cases.SInt32GTE{Val: 7}, 1)
	validateData(&cases.SInt32GTLT{Val: 1}, &cases.SInt32GTLT{Val: 0}, 1)
	validateData(&cases.SInt32GTLT{Val: 9}, &cases.SInt32GTLT{Val: 10}, 1)
	validateData(&cases.SInt32ExLTGT{Val: -1}, &cases.SInt32ExLTGT{Val: 0}, 1)
	validateData(&cases.SInt32ExLTGT{Val: 11}, &cases.SInt32ExLTGT{Val: 10}, 1)
	validateData(&cases.SInt32GTELTE{Val: 128}, &cases.SInt32GTELTE{Val: 127}, 1)
	validateData(&cases.SInt32GTELTE{Val: 256}, &cases.SInt32GTELTE{Val: 257}, 1)
	validateData(&cases.SInt32ExGTELTE{Val: 128}, &cases.SInt32ExGTELTE{Val: 129}, 1)
	validateData(&cases.SInt32ExGTELTE{Val: 256}, &cases.SInt32ExGTELTE{Val: 255}, 1)
	validateData(&cases.SInt32Ignore{Val: 0}, &cases.SInt32Ignore{Val: 1}, 1)
	validateData(&cases.SInt32Ignore{Val: -129}, &cases.SInt32Ignore{Val: 1}, 1)

	validateData(&cases.SInt64Const{Val: 1}, &cases.SInt64Const{Val: 2}, 1)
	validateData(&cases.SInt64In{Val: 2}, &cases.SInt64In{Val: 1}, 1)
	validateData(&cases.SInt64NotIn{Val: 1}, &cases.SInt64NotIn{Val: 0}, 1)
	validateData(&cases.SInt64LT{Val: -1}, &cases.SInt64LT{Val: 2}, 1)
	validateData(&cases.SInt64LTE{Val: 1}, &cases.SInt64LTE{Val: 65}, 1)
	validateData(&cases.SInt64GT{Val: 17}, &cases.SInt64GT{Val: 16}, 1)
	validateData(&cases.SInt64GTE{Val: 8}, &cases.SInt64GTE{Val: 7}, 1)
	validateData(&cases.SInt64GTLT{Val: 1}, &cases.SInt64GTLT{Val: 0}, 1)
	validateData(&cases.SInt64GTLT{Val: 9}, &cases.SInt64GTLT{Val: 10}, 1)
	validateData(&cases.SInt64ExLTGT{Val: -1}, &cases.SInt64ExLTGT{Val: 0}, 1)
	validateData(&cases.SInt64ExLTGT{Val: 11}, &cases.SInt64ExLTGT{Val: 10}, 1)
	validateData(&cases.SInt64GTELTE{Val: 128}, &cases.SInt64GTELTE{Val: 127}, 1)
	validateData(&cases.SInt64GTELTE{Val: 256}, &cases.SInt64GTELTE{Val: 257}, 1)
	validateData(&cases.SInt64ExGTELTE{Val: 128}, &cases.SInt64ExGTELTE{Val: 129}, 1)
	validateData(&cases.SInt64ExGTELTE{Val: 256}, &cases.SInt64ExGTELTE{Val: 255}, 1)
	validateData(&cases.SInt64Ignore{Val: 0}, &cases.SInt64Ignore{Val: 1}, 1)
	validateData(&cases.SInt64Ignore{Val: -129}, &cases.SInt64Ignore{Val: 1}, 1)

	validateData(&cases.Fixed32Const{Val: 1}, &cases.Fixed32Const{Val: 2}, 1)
	validateData(&cases.Fixed32In{Val: 2}, &cases.Fixed32In{Val: 1}, 1)
	validateData(&cases.Fixed32NotIn{Val: 1}, &cases.Fixed32NotIn{Val: 0}, 1)
	validateData(&cases.Fixed32LT{Val: 4}, &cases.Fixed32LT{Val: 5}, 1)
	validateData(&cases.Fixed32LTE{Val: 1}, &cases.Fixed32LTE{Val: 65}, 1)
	validateData(&cases.Fixed32GT{Val: 17}, &cases.Fixed32GT{Val: 16}, 1)
	validateData(&cases.Fixed32GTE{Val: 8}, &cases.Fixed32GTE{Val: 7}, 1)
	validateData(&cases.Fixed32GTLT{Val: 6}, &cases.Fixed32GTLT{Val: 5}, 1)
	validateData(&cases.Fixed32GTLT{Val: 9}, &cases.Fixed32GTLT{Val: 10}, 1)
	validateData(&cases.Fixed32ExLTGT{Val: 4}, &cases.Fixed32ExLTGT{Val: 5}, 1)
	validateData(&cases.Fixed32ExLTGT{Val: 11}, &cases.Fixed32ExLTGT{Val: 10}, 1)
	validateData(&cases.Fixed32GTELTE{Val: 128}, &cases.Fixed32GTELTE{Val: 127}, 1)
	validateData(&cases.Fixed32GTELTE{Val: 256}, &cases.Fixed32GTELTE{Val: 257}, 1)
	validateData(&cases.Fixed32ExGTELTE{Val: 128}, &cases.Fixed32ExGTELTE{Val: 129}, 1)
	validateData(&cases.Fixed32ExGTELTE{Val: 256}, &cases.Fixed32ExGTELTE{Val: 255}, 1)
	validateData(&cases.Fixed32Ignore{Val: 0}, &cases.Fixed32Ignore{Val: 129}, 1)
	validateData(&cases.Fixed32Ignore{Val: 128}, &cases.Fixed32Ignore{Val: 129}, 1)
	validateData(&cases.Fixed32Ignore{Val: 256}, &cases.Fixed32Ignore{Val: 255}, 1)

	validateData(&cases.Fixed64Const{Val: 1}, &cases.Fixed64Const{Val: 2}, 1)
	validateData(&cases.Fixed64In{Val: 2}, &cases.Fixed64In{Val: 1}, 1)
	validateData(&cases.Fixed64NotIn{Val: 1}, &cases.Fixed64NotIn{Val: 0}, 1)
	validateData(&cases.Fixed64LT{Val: 4}, &cases.Fixed64LT{Val: 5}, 1)
	validateData(&cases.Fixed64LTE{Val: 1}, &cases.Fixed64LTE{Val: 65}, 1)
	validateData(&cases.Fixed64GT{Val: 17}, &cases.Fixed64GT{Val: 16}, 1)
	validateData(&cases.Fixed64GTE{Val: 8}, &cases.Fixed64GTE{Val: 7}, 1)
	validateData(&cases.Fixed64GTLT{Val: 6}, &cases.Fixed64GTLT{Val: 5}, 1)
	validateData(&cases.Fixed64GTLT{Val: 9}, &cases.Fixed64GTLT{Val: 10}, 1)
	validateData(&cases.Fixed64ExLTGT{Val: 4}, &cases.Fixed64ExLTGT{Val: 5}, 1)
	validateData(&cases.Fixed64ExLTGT{Val: 11}, &cases.Fixed64ExLTGT{Val: 10}, 1)
	validateData(&cases.Fixed64GTELTE{Val: 128}, &cases.Fixed64GTELTE{Val: 127}, 1)
	validateData(&cases.Fixed64GTELTE{Val: 256}, &cases.Fixed64GTELTE{Val: 257}, 1)
	validateData(&cases.Fixed64ExGTELTE{Val: 128}, &cases.Fixed64ExGTELTE{Val: 129}, 1)
	validateData(&cases.Fixed64ExGTELTE{Val: 256}, &cases.Fixed64ExGTELTE{Val: 255}, 1)
	validateData(&cases.Fixed64Ignore{Val: 0}, &cases.Fixed64Ignore{Val: 129}, 1)
	validateData(&cases.Fixed64Ignore{Val: 128}, &cases.Fixed64Ignore{Val: 129}, 1)
	validateData(&cases.Fixed64Ignore{Val: 256}, &cases.Fixed64Ignore{Val: 255}, 1)

	validateData(&cases.SFixed32Const{Val: 1}, &cases.SFixed32Const{Val: 2}, 1)
	validateData(&cases.SFixed32In{Val: 2}, &cases.SFixed32In{Val: 1}, 1)
	validateData(&cases.SFixed32NotIn{Val: 1}, &cases.SFixed32NotIn{Val: 0}, 1)
	validateData(&cases.SFixed32LT{Val: -1}, &cases.SFixed32LT{Val: 2}, 1)
	validateData(&cases.SFixed32LTE{Val: 1}, &cases.SFixed32LTE{Val: 65}, 1)
	validateData(&cases.SFixed32GT{Val: 17}, &cases.SFixed32GT{Val: 16}, 1)
	validateData(&cases.SFixed32GTE{Val: 8}, &cases.SFixed32GTE{Val: 7}, 1)
	validateData(&cases.SFixed32GTLT{Val: 1}, &cases.SFixed32GTLT{Val: 0}, 1)
	validateData(&cases.SFixed32GTLT{Val: 9}, &cases.SFixed32GTLT{Val: 10}, 1)
	validateData(&cases.SFixed32ExLTGT{Val: -1}, &cases.SFixed32ExLTGT{Val: 0}, 1)
	validateData(&cases.SFixed32ExLTGT{Val: 11}, &cases.SFixed32ExLTGT{Val: 10}, 1)
	validateData(&cases.SFixed32GTELTE{Val: 128}, &cases.SFixed32GTELTE{Val: 127}, 1)
	validateData(&cases.SFixed32GTELTE{Val: 256}, &cases.SFixed32GTELTE{Val: 257}, 1)
	validateData(&cases.SFixed32ExGTELTE{Val: 128}, &cases.SFixed32ExGTELTE{Val: 129}, 1)
	validateData(&cases.SFixed32ExGTELTE{Val: 256}, &cases.SFixed32ExGTELTE{Val: 255}, 1)
	validateData(&cases.SFixed32Ignore{Val: 0}, &cases.SFixed32Ignore{Val: 1}, 1)
	validateData(&cases.SFixed32Ignore{Val: -129}, &cases.SFixed32Ignore{Val: 1}, 1)

	validateData(&cases.SFixed64Const{Val: 1}, &cases.SFixed64Const{Val: 2}, 1)
	validateData(&cases.SFixed64In{Val: 2}, &cases.SFixed64In{Val: 1}, 1)
	validateData(&cases.SFixed64NotIn{Val: 1}, &cases.SFixed64NotIn{Val: 0}, 1)
	validateData(&cases.SFixed64LT{Val: -1}, &cases.SFixed64LT{Val: 2}, 1)
	validateData(&cases.SFixed64LTE{Val: 1}, &cases.SFixed64LTE{Val: 65}, 1)
	validateData(&cases.SFixed64GT{Val: 17}, &cases.SFixed64GT{Val: 16}, 1)
	validateData(&cases.SFixed64GTE{Val: 8}, &cases.SFixed64GTE{Val: 7}, 1)
	validateData(&cases.SFixed64GTLT{Val: 1}, &cases.SFixed64GTLT{Val: 0}, 1)
	validateData(&cases.SFixed64GTLT{Val: 9}, &cases.SFixed64GTLT{Val: 10}, 1)
	validateData(&cases.SFixed64ExLTGT{Val: -1}, &cases.SFixed64ExLTGT{Val: 0}, 1)
	validateData(&cases.SFixed64ExLTGT{Val: 11}, &cases.SFixed64ExLTGT{Val: 10}, 1)
	validateData(&cases.SFixed64GTELTE{Val: 128}, &cases.SFixed64GTELTE{Val: 127}, 1)
	validateData(&cases.SFixed64GTELTE{Val: 256}, &cases.SFixed64GTELTE{Val: 257}, 1)
	validateData(&cases.SFixed64ExGTELTE{Val: 128}, &cases.SFixed64ExGTELTE{Val: 129}, 1)
	validateData(&cases.SFixed64ExGTELTE{Val: 256}, &cases.SFixed64ExGTELTE{Val: 255}, 1)
	validateData(&cases.SFixed64Ignore{Val: 0}, &cases.SFixed64Ignore{Val: 1}, 1)
	validateData(&cases.SFixed64Ignore{Val: -129}, &cases.SFixed64Ignore{Val: 1}, 1)

	v := int64(64)
	vv := int64(65)
	validateData(&cases.Int64LTEOptional{Val: &v}, &cases.Int64LTEOptional{Val: &vv}, 1)
	validateData(&cases.Int64LTEOptional{}, &cases.Int64LTEOptional{}, 0)

	// oneofs.proto
	validateData(&cases.TestOneOfMsg{Val: true}, &cases.TestOneOfMsg{Val: false}, 1)
	validateData(&cases.OneOfNone{}, &cases.OneOfNone{}, 0)
	validateData(&cases.OneOf{O: &cases.OneOf_X{X: "foobar"}}, &cases.OneOf{O: &cases.OneOf_X{X: "barfoo"}}, 1)
	validateData(&cases.OneOf{O: &cases.OneOf_Y{Y: 1}}, &cases.OneOf{O: &cases.OneOf_Y{Y: 0}}, 1)
	validateData(&cases.OneOf{O: &cases.OneOf_Z{Z: nil}}, &cases.OneOf{}, 0)
	validateData(&cases.OneOf{O: &cases.OneOf_Z{Z: &cases.TestOneOfMsg{Val: true}}}, &cases.OneOf{O: &cases.OneOf_Z{Z: &cases.TestOneOfMsg{Val: false}}}, 1)
	validateData(&cases.OneOfRequired{O: &cases.OneOfRequired_X{X: "foobar"}}, &cases.OneOfRequired{O: nil}, 1)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_X{X: ""}}, &cases.OneOfIgnoreEmpty{O: nil}, 0)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_X{X: "foo"}}, &cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_X{X: "foobar"}}, 1)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Y{Y: nil}}, &cases.OneOfIgnoreEmpty{O: nil}, 0)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Y{Y: []byte("foo")}}, &cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Y{Y: []byte("foobar")}}, 1)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Z{Z: 0}}, &cases.OneOfIgnoreEmpty{O: nil}, 0)
	validateData(&cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Z{Z: -128}}, &cases.OneOfIgnoreEmpty{O: &cases.OneOfIgnoreEmpty_Z{Z: 129}}, 1)

	// repeated.proto
	validateData(&cases.Embed{Val: 1}, &cases.Embed{Val: 0}, 1)
	validateData(&cases.RepeatedEmbedNone{Val: nil}, &cases.RepeatedEmbedNone{Val: []*cases.Embed{{Val: 0}}}, 1)
	validateData(&cases.RepeatedEmbedNone{Val: []*cases.Embed{{Val: 1}}}, &cases.RepeatedEmbedNone{Val: []*cases.Embed{{Val: 0}}}, 1)

	// strings.proto
	validateData(&cases.StringConst{Val: "foo"}, &cases.StringConst{Val: "bar"}, 1)
	validateData(&cases.StringIn{Val: "bar"}, &cases.StringIn{Val: "foo"}, 1)
	validateData(&cases.StringNotIn{Val: "foo"}, &cases.StringNotIn{Val: "fizz"}, 1)
	validateData(&cases.StringLen{Val: "foo"}, &cases.StringLen{Val: "foobar"}, 1)
	validateData(&cases.StringMinLen{Val: "foo"}, &cases.StringMinLen{Val: "fo"}, 1)
	validateData(&cases.StringMaxLen{Val: "foo"}, &cases.StringMaxLen{Val: "foobar"}, 1)
	validateData(&cases.StringMinMaxLen{Val: "foo"}, &cases.StringMinMaxLen{Val: "fo"}, 1)
	validateData(&cases.StringEqualMinMaxLen{Val: "foo11"}, &cases.StringEqualMinMaxLen{Val: "bar"}, 1)

	validateData(&cases.StringLenBytes{Val: "fizz"}, &cases.StringLenBytes{Val: "foobar"}, 1)
	validateData(&cases.StringMinBytes{Val: "fizz"}, &cases.StringMinBytes{Val: "fo"}, 1)
	validateData(&cases.StringMaxBytes{Val: "foo"}, &cases.StringMaxBytes{Val: "foobarfizz"}, 1)
	validateData(&cases.StringMinMaxBytes{Val: "fizz"}, &cases.StringMinMaxBytes{Val: "fo"}, 1)
	validateData(&cases.StringEqualMinMaxBytes{Val: "fizz"}, &cases.StringEqualMinMaxBytes{Val: "fizzbuzzbar"}, 1)

	validateData(&cases.StringPattern{Val: "foo"}, &cases.StringPattern{Val: "#foo"}, 1)
	//validateData(&cases.StringPatternEscapes{Val: "foo \\ bar"}, &cases.StringPatternEscapes{Val: "foo \\\\ 0"}, 1)
	validateData(&cases.StringPrefix{Val: "foo bar"}, &cases.StringPrefix{Val: "bar foo"}, 1)
	validateData(&cases.StringContains{Val: "\x00bar\x99"}, &cases.StringContains{Val: "\x00\x99"}, 1)
	validateData(&cases.StringNotContains{Val: "\x00\x99"}, &cases.StringNotContains{Val: "\x00bar\x99"}, 1)
	validateData(&cases.StringSuffix{Val: "\x00\x99baz"}, &cases.StringSuffix{Val: "\x00buz"}, 1)
	validateData(&cases.StringEmail{Val: "xxx@email.com"}, &cases.StringEmail{Val: "foo"}, 1)
	validateData(&cases.StringAddress{Val: "127.0.0.1"}, &cases.StringAddress{Val: "-.foo"}, 1)
	validateData(&cases.StringHostname{Val: "127.0.0.1"}, &cases.StringHostname{Val: "-.foo"}, 1)
	validateData(&cases.StringIP{Val: "127.0.0.1"}, &cases.StringIP{Val: "foo"}, 1)
	validateData(&cases.StringIPv4{Val: "127.0.0.1"}, &cases.StringIPv4{Val: "foo"}, 1)
	validateData(&cases.StringIPv6{Val: "::1"}, &cases.StringIPv6{Val: "foo"}, 1)
	validateData(&cases.StringURI{Val: "http://localhost:8080/v1/user"}, &cases.StringURI{Val: "foo"}, 1)
	validateData(&cases.StringURIRef{Val: "http://localhost:8080/v1/user"}, &cases.StringURIRef{Val: "\x7f#"}, 1)
	validateData(&cases.StringUUID{Val: "bb6529fc-5ddf-4c45-b372-78b311500d4b"}, &cases.StringUUID{Val: "bb6529fc-5ddf-4c45-b372--78b311500d4b"}, 1)
	validateData(&cases.StringUUIDIgnore{Val: ""}, &cases.StringUUIDIgnore{Val: "foo"}, 1)
	//validateData(&cases.StringHttpHeaderName{Val: "user"}, &cases.StringHttpHeaderName{}, 1)
	//validateData(&cases.StringHttpHeaderValue{Val: "123"}, &cases.StringHttpHeaderValue{Val: "foo"}, 1)
	//validateData(&cases.StringValidHeader{Val: "user"}, &cases.StringValidHeader{Val: "foo"}, 1)

	// wrappers.proto
	//validateData(&cases.WrapperDouble{Val: 1.0}, &cases.WrapperDouble{Val: 2.0}, 1)
	//validateData(&cases.WrapperFloat{Val: 1.0}, &cases.WrapperFloat{Val: 2.0}, 1)
	//validateData(&cases.WrapperInt64{Val: 1}, &cases.WrapperInt64{Val: 2}, 1)
	//validateData(&cases.WrapperUint64{Val: 1}, &cases.WrapperUint64{Val: 2}, 1)
	//validateData(&cases.WrapperInt32{Val: 1}, &cases.WrapperInt32{Val: 2}, 1)
	//validateData(&cases.WrapperUint32{Val: 1}, &cases.WrapperUint32{Val: 2}, 1)
	//validateData(&cases.WrapperBool{Val: true}, &cases.WrapperBool{Val: false}, 1)
	//validateData(&cases.WrapperString{Val: "foo"}, &cases.WrapperString{Val: "bar"}, 1)
	//validateData(&cases.WrapperBytes{Val: []byte("foo")}, &cases.WrapperBytes{Val: []byte("bar")}, 1)
	//validateData(&cases.WrapperEnum{Val: cases.TestEnum_ONE}, &cases.WrapperEnum{Val: cases.TestEnum_TWO}, 1)
	//validateData(&cases.WrapperEnumAlias{Val: cases.TestEnumAlias_A}, &cases.WrapperEnumAlias{Val: cases.TestEnumAlias_B}, 1)
	//validateData(&cases.WrapperEnumDefined{Val: cases.TestEnum_ONE}, &cases.WrapperEnumDefined{Val: cases.TestEnum(99)}, 1)

}
