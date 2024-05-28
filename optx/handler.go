package optx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ml444/gkit/errorx"
)

var (
	errGreaterThan  = errors.New("start value is greater than end value")
	createErrorFunc = func(typ string, value interface{}, es ...error) error {
		err := errorx.CreateErrorf(
			errorx.DefaultStatusCode,
			errorx.ErrCodeInvalidParamSys,
			fmt.Sprintf("ths %s type handler, invalid value: %v", typ, value),
		)
		if len(es) == 1 {
			return err.WithCause(es[0])
		}
		return err
	}
)

type handler interface {
	Apply(v interface{}) error
}

func NewNone(fn func() error) *noneHandler {
	return &noneHandler{fn: fn}
}

type noneHandler struct {
	fn func() error
}

func (h *noneHandler) Apply(_ interface{}) error {
	return h.fn()
}

func NewBool(fn func(bool) error) *boolHandler {
	return &boolHandler{fn: fn}
}

type boolHandler struct {
	fn func(bool) error
}

func (h *boolHandler) exactValue(v interface{}) (bool, error) {
	var value bool
	switch vv := v.(type) {
	case bool:
		value = vv
	case *bool:
		value = *vv
	case string:
		if inSliceStr(vv, []string{"1", "true"}) {
			value = true
		} else if inSliceStr(vv, []string{"0", "false"}) {
			value = false
		} else {
			return false, createErrorFunc("bool", v)
		}
	default:
		return false, createErrorFunc("bool", v)
	}
	return value, nil
}

func (h *boolHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	return h.fn(value)
}

func NewString(fn func(string) error, ignoreZero bool) *stringHandler {
	return &stringHandler{fn: fn, ignoreZero: ignoreZero}
}

type stringHandler struct {
	fn         func(string) error
	ignoreZero bool
}

func (h *stringHandler) exactValue(v interface{}) (string, error) {
	var value string
	switch vv := v.(type) {
	case *string:
		value = *vv
	case string:
		value = vv
	default:
		return "", createErrorFunc("string", v)
	}
	return value, nil
}

func (h *stringHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if value == "" && h.ignoreZero {
		return nil
	}
	return h.fn(value)
}

func NewStringList(fn func([]string) error, ignoreZero bool) *stringListHandler {
	return &stringListHandler{fn: fn, ignoreZero: ignoreZero}
}

type stringListHandler struct {
	fn         func([]string) error
	ignoreZero bool
}

func (h *stringListHandler) exactValue(v interface{}) ([]string, error) {
	var value []string
	switch vv := v.(type) {
	case []string:
		value = vv
	case *[]string:
		value = *vv
	case string:
		list := strings.Split(vv, ",")
		for _, s := range list {
			if s != "" {
				value = append(value, s)
			}
		}
	default:
		return nil, createErrorFunc("[]string", v)
	}
	return value, nil
}

func (h *stringListHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && len(value) == 0 {
		return nil
	}
	return h.fn(value)
}

func NewInt32(fn func(int32) error, ignoreZero bool) *int32Handler {
	return &int32Handler{fn: fn, ignoreZero: ignoreZero}
}

type int32Handler struct {
	fn         func(int32) error
	ignoreZero bool
}

func (h *int32Handler) exactValue(v interface{}) (int32, error) {
	var value int32
	switch vv := v.(type) {
	case int32:
		value = vv
	case *int32:
		value = *vv
	case string:
		vv = strings.TrimSpace(vv)
		if vv == "" {
			break
		}
		x, err := strconv.ParseInt(vv, 10, 32)
		if err != nil {
			return 0, createErrorFunc("int32", v)
		}
		value = int32(x)
	}
	return value, nil
}

func (h *int32Handler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && value == 0 {
		return nil
	}
	return h.fn(value)
}

func NewInt32List(fn func([]int32) error, ignoreZero bool) *int32ListHandler {
	return &int32ListHandler{fn: fn, ignoreZero: ignoreZero}
}

type int32ListHandler struct {
	fn         func([]int32) error
	ignoreZero bool
}

func (h *int32ListHandler) exactValue(v interface{}) ([]int32, error) {
	var value []int32
	switch vv := v.(type) {
	case []int32:
		value = vv
	case *[]int32:
		value = *vv
	case string:
		if vv == "" {
			return value, nil
		}
		list := strings.Split(vv, ",")
		for _, item := range list {
			x, err := strconv.ParseInt(strings.TrimSpace(item), 10, 32)
			if err != nil {
				return nil, createErrorFunc("[]int32", v)
			}
			value = append(value, int32(x))
		}
	}
	return value, nil
}

func (h *int32ListHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && len(value) == 0 {
		return nil
	}
	return h.fn(value)
}

func NewInt32Range(fn func(begin, end int32) error, ignoreZero bool) *int32RangeHandler {
	return &int32RangeHandler{fn: fn, ignoreZero: ignoreZero}
}

type int32RangeHandler struct {
	fn         func(begin, end int32) error
	ignoreZero bool
}

func (h *int32RangeHandler) exactValue(v interface{}) (begin, end int32, err error) {
	switch vv := v.(type) {
	case [2]int32:
		begin = vv[0]
		end = vv[1]
	case *[2]int32:
		begin = vv[0]
		end = vv[1]
	case []int32:
		if vvLen := len(vv); vvLen == 0 {
			break
		} else if vvLen == 1 {
			begin = vv[0]
		} else {
			begin = vv[0]
			end = vv[1]
		}
	case string:
		t1, t2, err := split2Int64(vv, "[2]int32", 32)
		return int32(t1), int32(t2), err
	}

	return
}

func (h *int32RangeHandler) Apply(v interface{}) error {
	begin, end, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && begin == 0 && end == 0 {
		return nil
	}
	return h.fn(begin, end)
}

func NewInt64(fn func(int64) error, ignoreZero bool) *int64Handler {
	return &int64Handler{fn: fn, ignoreZero: ignoreZero}
}

type int64Handler struct {
	fn         func(int64) error
	ignoreZero bool
}

func (h *int64Handler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && value == 0 {
		return nil
	}
	return h.fn(value)
}

func (h *int64Handler) exactValue(v interface{}) (value int64, err error) {
	switch vv := v.(type) {
	case int64:
		value = vv
	case *int64:
		value = *vv
	case string:
		vv = strings.TrimSpace(vv)
		if vv == "" {
			break
		}
		value, err = strconv.ParseInt(vv, 10, 64)
		if err != nil {
			return 0, createErrorFunc("int64", v)
		}
	}
	return value, nil
}

func NewInt64List(fn func([]int64) error, ignoreZero bool) *int64ListHandler {
	return &int64ListHandler{fn: fn, ignoreZero: ignoreZero}
}

type int64ListHandler struct {
	fn         func([]int64) error
	ignoreZero bool
}

func (h *int64ListHandler) exactValue(v interface{}) ([]int64, error) {
	var value []int64
	switch vv := v.(type) {
	case []int64:
		value = vv
	case *[]int64:
		value = *vv
	case string:
		if vv == "" {
			return value, nil
		}
		list := strings.Split(vv, ",")
		for _, item := range list {
			x, err := strconv.ParseInt(strings.TrimSpace(item), 10, 64)
			if err != nil {
				return nil, err
			}
			value = append(value, x)
		}
	default:
		return nil, createErrorFunc("[]int64", v)
	}
	return value, nil
}

func (h *int64ListHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && len(value) == 0 {
		return nil
	}
	return h.fn(value)
}

func NewInt64Range(fn func(begin, end int64) error, ignoreZero bool) *int64RangeHandler {
	return &int64RangeHandler{fn: fn, ignoreZero: ignoreZero}
}

type int64RangeHandler struct {
	fn         func(begin, end int64) error
	ignoreZero bool
}

func (h *int64RangeHandler) exactValue(v interface{}) (begin, end int64, err error) {
	switch vv := v.(type) {
	case [2]int64:
		begin = vv[0]
		end = vv[1]
	case *[2]int64:
		begin = vv[0]
		end = vv[1]
	case []int64:
		if vvLen := len(vv); vvLen == 0 {
			break
		} else if vvLen == 1 {
			begin = vv[0]
		} else {
			begin = vv[0]
			end = vv[1]
		}
	case string:
		return split2Int64(vv, "[2]int64", 64)
	}

	return
}

func (h *int64RangeHandler) Apply(v interface{}) error {
	begin, end, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && begin == 0 && end == 0 {
		return nil
	}
	return h.fn(begin, end)
}

func NewUint32(fn func(uint32) error, ignoreZero bool) *uint32Handler {
	return &uint32Handler{fn: fn, ignoreZero: ignoreZero}
}

type uint32Handler struct {
	fn         func(uint32) error
	ignoreZero bool
}

func (h *uint32Handler) exactValue(v interface{}) (uint32, error) {
	var value uint32
	switch vv := v.(type) {
	case uint32:
		value = vv
	case *uint32:
		value = *vv
	case string:
		vv = strings.TrimSpace(vv)
		if vv == "" {
			break
		}
		x, err := strconv.ParseUint(vv, 10, 32)
		if err != nil {
			return 0, createErrorFunc("uint32", v)
		}
		value = uint32(x)
	}
	return value, nil
}

func (h *uint32Handler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && value == 0 {
		return nil
	}
	return h.fn(value)
}

func NewUint32List(fn func([]uint32) error, ignoreZero bool) *uint32ListHandler {
	return &uint32ListHandler{fn: fn, ignoreZero: ignoreZero}
}

type uint32ListHandler struct {
	fn         func([]uint32) error
	ignoreZero bool
}

func (h *uint32ListHandler) exactValue(v interface{}) ([]uint32, error) {
	var value []uint32
	switch vv := v.(type) {
	case []uint32:
		value = vv
	case *[]uint32:
		value = *vv
	case string:
		if vv == "" {
			return value, nil
		}
		list := strings.Split(vv, ",")
		for _, item := range list {
			item = strings.TrimSpace(item)
			x, err := strconv.ParseUint(item, 10, 32)
			if err != nil {
				return nil, createErrorFunc("[]uint32", v)
			}
			value = append(value, uint32(x))
		}
	}
	return value, nil
}

func (h *uint32ListHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && len(value) == 0 {
		return nil
	}
	return h.fn(value)
}

func NewUint32Range(fn func(begin, end uint32) error, ignoreZero bool) *uint32RangeHandler {
	return &uint32RangeHandler{fn: fn, ignoreZero: ignoreZero}
}

type uint32RangeHandler struct {
	fn         func(begin, end uint32) error
	ignoreZero bool
}

func (h *uint32RangeHandler) exactValue(v interface{}) (begin, end uint32, err error) {
	switch vv := v.(type) {
	case [2]uint32:
		begin = vv[0]
		end = vv[1]
	case *[2]uint32:
		begin = vv[0]
		end = vv[1]
	case []uint32:
		if vvLen := len(vv); vvLen == 0 {
			break
		} else if vvLen == 1 {
			begin = vv[0]
		} else {
			begin = vv[0]
			end = vv[1]
		}
	case string:
		t1, t2, err := split2Uint64(vv, "[2]uint32", 32)
		return uint32(t1), uint32(t2), err
	}

	return
}

func (h *uint32RangeHandler) Apply(v interface{}) error {
	begin, end, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && begin == 0 && end == 0 {
		return nil
	}
	return h.fn(begin, end)
}

func NewUint64(fn func(uint64) error, ignoreZero bool) *uint64Handler {
	return &uint64Handler{fn: fn, ignoreZero: ignoreZero}
}

type uint64Handler struct {
	fn         func(uint64) error
	ignoreZero bool
}

func (h *uint64Handler) exactValue(v interface{}) (value uint64, err error) {
	switch vv := v.(type) {
	case uint64:
		value = vv
	case *uint64:
		value = *vv
	case string:
		vv = strings.TrimSpace(vv)
		if vv == "" {
			break
		}
		value, err = strconv.ParseUint(vv, 10, 64)
		if err != nil {
			return 0, createErrorFunc("uint64", v)
		}
	}
	return value, nil
}

func (h *uint64Handler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && value == 0 {
		return nil
	}
	return h.fn(value)
}

func NewUint64List(fn func([]uint64) error, ignoreZero bool) *uint64ListHandler {
	return &uint64ListHandler{fn: fn, ignoreZero: ignoreZero}
}

type uint64ListHandler struct {
	fn         func([]uint64) error
	ignoreZero bool
}

func (h *uint64ListHandler) exactValue(v interface{}) ([]uint64, error) {
	var value []uint64
	switch vv := v.(type) {
	case []uint64:
		value = vv
	case *[]uint64:
		value = *vv
	case string:
		if vv == "" {
			return value, nil
		}
		list := strings.Split(vv, ",")
		for _, item := range list {
			item = strings.TrimSpace(item)
			x, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				return nil, createErrorFunc("[]uint64", v)
			}
			value = append(value, x)
		}
	}
	return value, nil
}

func (h *uint64ListHandler) Apply(v interface{}) error {
	value, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && len(value) == 0 {
		return nil
	}
	return h.fn(value)
}

func NewUint64Range(fn func(begin, end uint64) error, ignoreZero bool) *uint64RangeHandler {
	return &uint64RangeHandler{fn: fn, ignoreZero: ignoreZero}
}

type uint64RangeHandler struct {
	fn         func(begin, end uint64) error
	ignoreZero bool
}

func (h *uint64RangeHandler) exactValue(v interface{}) (begin, end uint64, err error) {
	switch vv := v.(type) {
	case [2]uint64:
		begin = vv[0]
		end = vv[1]
	case *[2]uint64:
		begin = vv[0]
		end = vv[1]
	case []uint64:
		if vvLen := len(vv); vvLen == 0 {
			break
		} else if vvLen == 1 {
			begin = vv[0]
		} else {
			begin = vv[0]
			end = vv[1]
		}
	case string:
		return split2Uint64(vv, "[2]uint64", 64)
	}

	return
}

func (h *uint64RangeHandler) Apply(v interface{}) error {
	begin, end, err := h.exactValue(v)
	if err != nil {
		return err
	}
	if h.ignoreZero && begin == 0 && end == 0 {
		return nil
	}
	return h.fn(begin, end)
}

func split2Int64(vv string, typStr string, bitSize int) (begin, end int64, err error) {
	sList := strings.Split(vv, ",")
	if sLen := len(sList); sLen == 0 {
		return 0, 0, createErrorFunc(typStr, vv)
	} else if sLen == 1 {
		item := strings.TrimSpace(sList[0])
		if item == "" {
			return 0, 0, nil
		}
		begin, err = strconv.ParseInt(item, 10, bitSize)
		if err != nil {
			return 0, 0, createErrorFunc(typStr, vv, err)
		}
		if begin > 0 {
			return 0, 0, createErrorFunc(typStr, vv, errGreaterThan)
		}
	} else {
		item1 := strings.TrimSpace(sList[0])
		if item1 != "" {
			begin, err = strconv.ParseInt(item1, 10, bitSize)
			if err != nil {
				return 0, 0, createErrorFunc(typStr, vv, err)
			}
		}
		item2 := strings.TrimSpace(sList[1])
		if item2 != "" {
			end, err = strconv.ParseInt(item2, 10, bitSize)
			if err != nil {
				return 0, 0, createErrorFunc(typStr, vv, err)
			}
		}
		if begin > end {
			return 0, 0, createErrorFunc(typStr, vv, errGreaterThan)
		}
	}
	return
}

func split2Uint64(vv string, typStr string, bitSize int) (begin, end uint64, err error) {
	sList := strings.Split(vv, ",")
	if sLen := len(sList); sLen <= 1 {
		return 0, 0, createErrorFunc(typStr, vv)
	} else {
		item1 := strings.TrimSpace(sList[0])
		if item1 != "" {
			begin, err = strconv.ParseUint(item1, 10, bitSize)
			if err != nil {
				return 0, 0, createErrorFunc(typStr, vv, err)
			}
		}
		item2 := strings.TrimSpace(sList[1])
		if item2 != "" {
			end, err = strconv.ParseUint(item2, 10, bitSize)
			if err != nil {
				return 0, 0, createErrorFunc(typStr, vv, err)
			}
		}
		if begin > end {
			return 0, 0, createErrorFunc(typStr, vv, errGreaterThan)
		}
	}
	return
}
