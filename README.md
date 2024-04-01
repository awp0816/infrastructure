# infrastructure
example:
```
    import(
        "github.com/gin-gonic/gin"
        
        "github.com/awp0816/infrastructure/gm/exception"
        "github.com/awp0816/infrastructure/gm/valid"
        "github.com/awp0816/infrastructure/logger"
    )
    
    func main() {
        defer func() {
            if e := recover(); e != nil {
                fmt.Println(fmt.Sprintf("Server Start ERROR:%v", e))
            }
        }()
        if err := logger.SetupLogger(&logger.LogConfig{
            ...
        });err != nil{
            return
        }
        defer logger.Logger.Sync()
        
        engine := gin.New()
        engine.Use(
            //设置语言
            valid.SetLanguage(),
            //异常处理记录日志
            exception.RecoveryPanic(logger.Logger,true)
        )
        //注册翻译器
        if e := valid.RegisterTranslate(valid.Language); e != nil {
            logger.Logger.Error("message......")
            return
        }
        engine.Handle(http.MethodPost, "/api/v1/example", func(ctx *gin.Context) {
		var (
			request struct {
				//必传参数
				Name string `json:"name" binding:"required"`
				//大于等于18小于等于120
				Age int `json:"age" binding:"required,gte=18,lte=120"`
				//ipv4验证
				Ip string `json:"ip" binding:"required,ipv4"`
				//数组验证,数组长度必须大于0,数组内元素长度最小为1,最大为100
				Uniques []string `json:"uniques" binding:"required,dive,gt=0,min=1,max=100"`
				//日期格式验证
				Date string `json:"date" binding:"required,datetime=2006-01-02 15:04:05"`
				//设置为必传,若内部元素未设置,可以校验通过
				Children _Children `json:"children" binding:"required"`
			}
			err error
		)
		//这里需自己实现
		if err = _BindAndValid(ctx, &request); err != nil {
			//记录日志
			return
		}
		ctx.JSON(http.StatusOK, nil)
	})
	_ = engine.Run(":8000")
    }
```