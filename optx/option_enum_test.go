package optx

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const (
	noneKey int32 = iota
	boolKey
	int32Key
	int32ListKey
	int32RangeKey
	int64Key
	int64ListKey
	int64RangeKey
	uint32Key
	uint32ListKey
	uint32RangeKey
	uint64Key
	uint64ListKey
	uint64RangeKey
)

type testData struct {
	noneV       string
	boolV       bool
	int32V      int32
	int32List   []int32
	int32Range  [2]int32
	int64V      int64
	int64List   []int64
	int64Range  [2]int64
	uint32V     uint32
	uint32List  []uint32
	uint32Range [2]uint32
	uint64V     uint64
	uint64List  []uint64
	uint64Range [2]uint64
}

func getDefaultTestData() *testData {
	return &testData{
		noneV:       "None",
		boolV:       true,
		int32V:      math.MaxInt32,
		int32List:   []int32{math.MinInt32, 0, math.MaxInt32},
		int32Range:  [2]int32{math.MinInt32, math.MaxInt32},
		int64V:      math.MaxInt64,
		int64List:   []int64{math.MinInt64, 0, math.MaxInt64},
		int64Range:  [2]int64{math.MinInt64, math.MaxInt64},
		uint32V:     math.MaxUint32,
		uint32List:  []uint32{0, 99, math.MaxUint32},
		uint32Range: [2]uint32{0, math.MaxUint32},
		uint64V:     math.MaxUint64,
		uint64List:  []uint64{0, 456, math.MaxUint64},
		uint64Range: [2]uint64{0, math.MaxUint64},
	}
}

func getTestOptions() *Options {
	opts := NewEnumOptions(
		noneKey, "None",
		boolKey, true,
		int32Key, math.MaxInt32,
		int32ListKey, []int32{math.MinInt32, 0, math.MaxInt32},
		int32RangeKey, [2]int32{math.MinInt32, math.MaxInt32},
		int64Key, fmt.Sprintf("%d", math.MaxInt64),
		int64ListKey, fmt.Sprintf("%d , %d,%d", math.MinInt64, 0, math.MaxInt64), // check space
		int64RangeKey, fmt.Sprintf("%d,%d,%d", math.MinInt64, math.MaxInt64, 0), // test multiple values
	).
		AddOpt(uint32Key, math.MaxUint32).
		AddOpt(uint32ListKey, []uint32{0, 99, math.MaxUint32}).
		AddOpt(uint32RangeKey, [2]uint32{0, math.MaxUint32}).
		AddOpt(uint64Key, fmt.Sprintf("%d", uint64(math.MaxUint64))).
		AddOpt(uint64ListKey, fmt.Sprintf("%d , %d,%d", 0, 456, uint64(math.MaxUint64))).
		AddOpt(uint64RangeKey, fmt.Sprintf("%d,%d,%d", 0, uint64(math.MaxUint64), 456))

	return opts
}

func getTestProcessor(data *testData) *Processor {
	p := NewEnumProcessor().
		AddNone(noneKey, func() error {
			data.noneV = "None"
			return nil
		}).
		AddBool(boolKey, func(val bool) error {
			data.boolV = val
			return nil
		}).
		AddInt32(int32Key, func(val int32) error {
			data.int32V = val
			return nil
		}).
		AddInt32List(int32ListKey, func(val []int32) error {
			data.int32List = val
			return nil
		}).
		AddInt32Range(int32RangeKey, func(begin, end int32) error {
			data.int32Range = [2]int32{begin, end}
			return nil
		}).
		AddInt64(int64Key, func(val int64) error {
			data.int64V = val
			return nil
		}).
		AddInt64List(int64ListKey, func(val []int64) error {
			data.int64List = val
			return nil
		}).
		AddInt64Range(int64RangeKey, func(begin, end int64) error {
			data.int64Range = [2]int64{begin, end}
			return nil
		}).
		AddUint32(uint32Key, func(val uint32) error {
			data.uint32V = val
			return nil
		}).
		AddUint32List(uint32ListKey, func(val []uint32) error {
			data.uint32List = val
			return nil
		}).
		AddUint32Range(uint32RangeKey, func(begin, end uint32) error {
			data.uint32Range = [2]uint32{begin, end}
			return nil
		}).
		AddUint64(uint64Key, func(val uint64) error {
			data.uint64V = val
			return nil
		}).
		AddUint64List(uint64ListKey, func(val []uint64) error {
			data.uint64List = val
			return nil
		}).
		AddUint64Range(uint64RangeKey, func(begin, end uint64) error {
			data.uint64Range = [2]uint64{begin, end}
			return nil
		})

	return p
}

func TestEnumProcessor(t *testing.T) {
	data := &testData{}
	p := getTestProcessor(data)
	err := p.Process(getTestOptions())
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(data, getDefaultTestData()) {
		t.Errorf("Data not equal: %+v", data)
	}
}
