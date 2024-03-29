package valid

import "fmt"

type ErrCode string

const (
	SuccessCode = "000000"
	ErrParams   = "000001"
	ErrInternal = "999999"
)

var TextErr = map[ErrCode]string{
	SuccessCode: "SUCCESS",
	ErrParams:   "参数错误",
	ErrInternal: "系统异常，请联系管理员",
}

func (e ErrCode) Error() string {
	if v, ok := TextErr[e]; ok {
		return v
	}
	return fmt.Sprintf("错误码未定义:%s", e)
}

func SetTextErr(code ErrCode, desc string) {
	TextErr[code] = desc
}
