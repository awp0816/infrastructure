package exception

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
)

func RecoveryPanic(logger *zap.Logger, isStack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				//检查是否发生了断开的连接,因为它不是真正需要 panic 堆栈跟踪的情况。
				var brokenPipe bool
				if ne, ok := e.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				//记录请求头到map
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				var (
					allHeaders   = make(map[string]string)
					headerString = ""
					_Methods     = []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
				)
				for _, header := range headers {
					if len(strings.TrimSpace(header)) == 0 {
						continue
					}
					flag := false
					for _, method := range _Methods {
						if strings.Contains(header, method) {
							flag = true
							break
						}
					}
					if flag {
						continue
					}
					arrayHeaders := strings.Split(header, ":")
					if strings.ToLower(arrayHeaders[0]) == "authorization" {
						allHeaders[arrayHeaders[0]] = "*"
						continue
					}
					if len(arrayHeaders) > 2 {
						var original string
						for i, arrayParam := range arrayHeaders {
							if i == 0 {
								continue
							}
							if i == len(arrayHeaders)-1 {
								original += fmt.Sprintf("%s", arrayParam)
							} else {
								original += fmt.Sprintf("%s:", arrayParam)
							}
						}
						allHeaders[arrayHeaders[0]] = original
					} else {
						allHeaders[arrayHeaders[0]] = arrayHeaders[1]
					}
				}
				if len(allHeaders) > 0 {
					headerBody, _ := json.Marshal(allHeaders)
					headerString = string(headerBody)
				}
				if brokenPipe {
					logger.DPanic("连接被断开,服务异常",
						zap.Any("error", e),
						zap.String("headers", headerString),
					)
					_ = c.Error(e.(error)) // 如果连接已断开,我们就无法向其写入状态
					c.Abort()
					return
				}

				debugStack := ""
				for _, v := range strings.Split(string(debug.Stack()), "\n") {
					debugStack += v + "\n"
				}
				if isStack {
					logger.DPanic("运行时错误,服务异常",
						zap.Any("error", e),
						zap.String("headers", headerString),
						zap.String("debug_stack", debugStack),
					)
				} else {
					logger.DPanic("运行时错误,服务异常",
						zap.Any("error", e),
						zap.String("headers", headerString),
					)
				}
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    1000,
					"message": "服务异常,请联系管理员!",
				})
			}
		}()
		c.Next()
	}
}
