package code

import "fmt"

type ErrCode string

const (
	SUCCESS     ErrCode = "000000"
	ErrParams   ErrCode = "000001"
	ErrInternal ErrCode = "999999"
)

var TextErr = map[ErrCode]string{
	SUCCESS:     "SUCCESS",
	ErrParams:   "参数错误",
	ErrInternal: "系统异常,请联系管理员",
}

func (e ErrCode) Error() string {
	if v, ok := TextErr[e]; ok {
		return v
	}
	return fmt.Sprintf("错误码未定义:%s", e)
}
