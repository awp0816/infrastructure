package valid

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	englishTranslations "github.com/go-playground/validator/v10/translations/en"
	hanTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	han        = "zh"
	english    = "en"
	Language   = han
	Translator ut.Translator
)

// SetLanguage 设置翻译器语言,默认中文
func SetLanguage() gin.HandlerFunc {
	return func(context *gin.Context) {
		language := context.Request.Header.Get("Accept-Language")
		if strings.Contains(language, english) {
			Language = english
		} else if strings.Contains(language, han) {
			Language = han
		} else {
			Language = han
		}
	}
}

// RegisterTranslate 注册翻译器,默认获取jsonTag返回,若无,则获取formTag,若无,则返回字段名称
func RegisterTranslate(locale string) (err error) {
	// 修改gin框架中的Validator引擎属性，实现自定制
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器
		uni := ut.New(zhT, zhT, enT)
		var ok bool
		Translator, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			label := field.Tag.Get("json")
			if label == "" {
				if label = field.Tag.Get("form"); label == "" {
					return field.Name
				}
			}
			return label
		})
		// 注册翻译器
		switch locale {
		case english:
			return englishTranslations.RegisterDefaultTranslations(v, Translator)
		default:
			return hanTranslations.RegisterDefaultTranslations(v, Translator)
		}
	}
	return
}

// _BindAndValid 这里因错误码个人定义不同,暂时不提供使用,只提供具体示例
func _BindAndValid(c *gin.Context, obj interface{}) (err error) {
	if err = c.ShouldBind(obj); err != nil {
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"code":    "000001",
				"message": "参数错误",
				"data":    err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    "000001",
			"message": "参数错误",
			"data":    errs.Translate(Translator),
		})
		return
	}
	return
}

func example() {
	engine := gin.Default()
	//设置语言
	engine.Use(SetLanguage())
	//注册翻译器
	if e := RegisterTranslate(Language); e != nil {
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
		//更多参数验证参考: https://blog.csdn.net/zhaozuoyou/article/details/127812519
		if err = _BindAndValid(ctx, &request); err != nil {
			//记录日志
			return
		}
		ctx.JSON(http.StatusOK, nil)
	})
	_ = engine.Run(":8000")
}

type _Children struct {
	//设置为必传,若未传,则校验不通过
	Name string `json:"name" binding:"required"`
	//未设置,可传可不传
	Age int `json:"age"`
}
