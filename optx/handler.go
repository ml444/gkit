package optx

import (
	"fmt"

	"github.com/ml444/gkit/errorx"
)

var createError = func(key, value interface{}) error {
	return errorx.CreateErrorf(errorx.DefaultStatusCode, errorx.ErrCodeInvalidParamSys,
		fmt.Sprintf("key[%v] invalid value: %v", key, value))
}

type handler interface {
	apply() error
	setValue(v interface{})
}

type handlerNone struct {
	fn func() error
}

func (h *handlerNone) apply() error {
	return h.fn()
}
func (h *handlerNone) setValue(v interface{}) {
	return
}
func NewNone(fn func() error) *handlerNone {
	return &handlerNone{fn: fn}
}

type handlerBool struct {
	v  bool
	fn func(bool) error
}

func (h *handlerBool) apply() error {
	return h.fn(h.v)
}
func (h *handlerBool) setValue(v interface{}) {
	h.v = v.(bool)
}
func NewBool(fn func(bool) error) *handlerBool {
	return &handlerBool{fn: fn}
}

type handlerString struct {
	v  string
	fn func(string) error
}

func (h *handlerString) apply() error {
	return h.fn(h.v)
}
func (h *handlerString) setValue(v interface{}) {
	h.v = v.(string)
}
func NewString(fn func(string) error) *handlerString {
	return &handlerString{fn: fn}
}

type handlerStringList struct {
	v  []string
	fn func([]string) error
}

func (h *handlerStringList) apply() error {
	return h.fn(h.v)
}
func (h *handlerStringList) setValue(v interface{}) {
	h.v = v.([]string)
}
func NewStringList(fn func([]string) error) *handlerStringList {
	return &handlerStringList{fn: fn}
}

type handlerInt32 struct {
	v  int32
	fn func(int32) error
}

func (h *handlerInt32) apply() error {
	return h.fn(h.v)
}
func (h *handlerInt32) setValue(v interface{}) {
	h.v = v.(int32)
}
func NewInt32(fn func(int32) error) *handlerInt32 {
	return &handlerInt32{fn: fn}
}

type handlerInt32List struct {
	v  []int32
	fn func([]int32) error
}

func (h *handlerInt32List) apply() error {
	return h.fn(h.v)
}
func (h *handlerInt32List) setValue(v interface{}) {
	h.v = v.([]int32)
}
func NewInt32List(fn func([]int32) error) *handlerInt32List {
	return &handlerInt32List{fn: fn}
}

type handlerInt32Range struct {
	beginValue int32
	endValue   int32
	fn         func(begin, end int32) error
	e          error
}

func (h *handlerInt32Range) apply() error {
	return h.fn(h.beginValue, h.endValue)
}
func (h *handlerInt32Range) setValue(v interface{}) {
	vv := v.([]int32)
	if len(vv) != 2 {
		h.e = createError("int32Range", v)
		return
	}
	h.beginValue = v.([]int32)[0]
	h.endValue = v.([]int32)[1]
}
func NewInt32Range(fn func(begin, end int32) error) *handlerInt32Range {
	return &handlerInt32Range{fn: fn}
}

type handlerInt64 struct {
	v  int64
	fn func(int64) error
}

func (h *handlerInt64) apply() error {
	return h.fn(h.v)
}
func (h *handlerInt64) setValue(v interface{}) {
	h.v = v.(int64)
}
func NewInt64(fn func(int64) error) *handlerInt64 {
	return &handlerInt64{fn: fn}
}

type handlerInt64List struct {
	v  []int64
	fn func([]int64) error
}

func (h *handlerInt64List) apply() error {
	return h.fn(h.v)
}
func (h *handlerInt64List) setValue(v interface{}) {
	h.v = v.([]int64)
}
func NewInt64List(fn func([]int64) error) *handlerInt64List {
	return &handlerInt64List{fn: fn}
}

type handlerInt64Range struct {
	beginValue int64
	endValue   int64
	fn         func(begin, end int64) error
	e          error
}

func (h *handlerInt64Range) apply() error {
	return h.fn(h.beginValue, h.endValue)
}
func (h *handlerInt64Range) setValue(v interface{}) {
	vv := v.([]int64)
	if len(vv) != 2 {
		h.e = createError("int32Range", v)
		return
	}
	h.beginValue = v.([]int64)[0]
	h.endValue = v.([]int64)[1]
}
func NewInt64Range(fn func(begin, end int64) error) *handlerInt64Range {
	return &handlerInt64Range{fn: fn}
}

type handlerUint32 struct {
	v  uint32
	fn func(uint32) error
}

func (h *handlerUint32) apply() error {
	return h.fn(h.v)
}
func (h *handlerUint32) setValue(v interface{}) {
	h.v = v.(uint32)
}
func NewUint32(fn func(uint32) error) *handlerUint32 {
	return &handlerUint32{fn: fn}
}

type handlerUint32List struct {
	v  []uint32
	fn func([]uint32) error
}

func (h *handlerUint32List) apply() error {
	return h.fn(h.v)
}
func (h *handlerUint32List) setValue(v interface{}) {
	h.v = v.([]uint32)
}
func NewUint32List(fn func([]uint32) error) *handlerUint32List {
	return &handlerUint32List{fn: fn}
}

type handlerUint32Range struct {
	beginValue uint32
	endValue   uint32
	fn         func(begin, end uint32) error
	e          error
}

func (h *handlerUint32Range) apply() error {
	return h.fn(h.beginValue, h.endValue)
}
func (h *handlerUint32Range) setValue(v interface{}) {
	vv := v.([]uint32)
	if len(vv) != 2 {
		h.e = createError("int32Range", v)
		return
	}
	h.beginValue = v.([]uint32)[0]
	h.endValue = v.([]uint32)[1]
}
func NewUint32Range(fn func(begin, end uint32) error) *handlerUint32Range {
	return &handlerUint32Range{fn: fn}
}

type handlerUint64 struct {
	v  uint64
	fn func(uint64) error
}

func (h *handlerUint64) apply() error {
	return h.fn(h.v)
}
func (h *handlerUint64) setValue(v interface{}) {
	h.v = v.(uint64)
}
func NewUint64(fn func(uint64) error) *handlerUint64 {
	return &handlerUint64{fn: fn}
}

type handlerUint64List struct {
	v  []uint64
	fn func([]uint64) error
}

func (h *handlerUint64List) apply() error {
	return h.fn(h.v)
}
func (h *handlerUint64List) setValue(v interface{}) {
	h.v = v.([]uint64)
}
func NewUint64List(fn func([]uint64) error) *handlerUint64List {
	return &handlerUint64List{fn: fn}
}

type handlerUint64Range struct {
	beginValue uint64
	endValue   uint64
	fn         func(begin, end uint64) error
	e          error
}

func (h *handlerUint64Range) apply() error {
	return h.fn(h.beginValue, h.endValue)
}
func (h *handlerUint64Range) setValue(v interface{}) {
	vv := v.([]uint64)
	if len(vv) != 2 {
		h.e = createError("uint64Range", v)
		return
	}
	h.beginValue = v.([]uint64)[0]
	h.endValue = v.([]uint64)[1]
}
func NewUint64Range(fn func(begin, end uint64) error) *handlerUint64Range {
	return &handlerUint64Range{fn: fn}
}
