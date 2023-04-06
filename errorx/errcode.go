package errorx

const (
	DefaultStatusCode = 400
	UnknownStatusCode = 500
)
const (
	ErrCodeUnknown           int32 = -1
	ErrCodeInvalidReqSys     int32 = 40000
	ErrCodeInvalidParamSys   int32 = 40001
	ErrCodeInvalidHeaderSys  int32 = 40002
	ErrCodeInvalidBodySys    int32 = 40003
	ErrCodeRecordNotFoundSys int32 = 40004
)
