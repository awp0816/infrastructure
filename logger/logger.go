package logger

import (
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type LogParams struct {
	Type     string `yaml:"type" json:"-"`             //日志输出格式,console|file
	Filename string `yaml:"filename" json:"filename"`  //保存的文件名
	MaxLines int    `yaml:"max_lines" json:"maxlines"` //每个文件保存的最大行数，默认值 1000000
	MaxSize  int    `yaml:"max_size" json:"maxsize"`   //每个文件保存的最大尺寸，默认值是 1 << 28, 256 MB
	Daily    bool   `yaml:"daily" json:"daily"`        //是否按照每天 logrotate，默认是 true
	MaxDays  int    `yaml:"max_days" json:"maxdays"`   //文件最多保存多少天，默认保存 7 天
	Rotate   bool   `yaml:"rotate" json:"rotate"`      //是否开启 logrotate，默认是 true
	Level    int    `json:"level"`                     //日志保存的时候的级别，默认是 Trace 级别
	Desc     string `yaml:"desc" json:"-"`             //日志级别
	Perm     string `yaml:"perm" json:"perm"`          //日志文件权限
	Async    bool   `yaml:"async" json:"-"`            //是否异步输出
}

var (
	Logger *logs.BeeLogger
)

func SetupLogger(in LogParams) (err error) {

	//校验日志名称
	if in.Filename != "" {
		if filepath.Ext(in.Filename) != ".log" {
			err = errors.New("文件后缀必须以.log结束")
			return
		}
		dirPath := filepath.Dir(in.Filename)
		if _, err = os.Stat(dirPath); os.IsNotExist(err) {
			if err = os.MkdirAll(dirPath, 0755); err != nil {
				return
			}
		}
	} else {
		in.Type = logs.AdapterConsole
	}

	//设置日志级别
	in.Level = _Level[in.Desc]

	//初始化日志对象
	Logger = logs.NewLogger(10000)
	if in.Async {
		Logger.Async()
	}

	// 输出log时能显示输出文件名和行号
	Logger.EnableFuncCallDepth(true)
	switch in.Type {
	case logs.AdapterFile:
		var bytes []byte
		if bytes, err = json.Marshal(in); err != nil {
			return
		}
		return Logger.SetLogger(logs.AdapterFile, string(bytes))
	case logs.AdapterConsole:
		return Logger.SetLogger(logs.AdapterConsole)
	}
	return
}

var _Level = map[string]int{
	"DEBUG":   logs.LevelDebug,
	"ERROR":   logs.LevelError,
	"WARNING": logs.LevelWarning,
	"INFO":    logs.LevelInformational,
}
