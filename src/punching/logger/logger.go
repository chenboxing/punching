package logger

import (
	"fmt"

	"github.com/cihub/seelog"
	"os"
	"strings"
)

// init 初始化包
func init() {
	// 解析服务配置文件
	logFileName := "./log/punching.log"
	if os.Getenv("Log_FILE") != "" {
		logFileName = os.Getenv("Log_FILE")
	}
	xml := `
	<seelog>
    <outputs formatid="main">
        <filter levels="info,critical,error,debug">
            <console formatid="main" />
            <rollingfile formatid="main" type="date" filename="#LOG_FILE_NAME#" datepattern="2006.01.02" />
        </filter>
    </outputs>

    <formats>
        <format id="main" format="%Date %Time [%LEV] %Msg [%File][%FuncShort][%Line]%n"/>
    </formats>
	</seelog>
	`

	xml = strings.Replace(xml, "#LOG_FILE_NAME", logFileName, 1)
	// 解析日志配置（从默认配置）
	logg, err := seelog.LoggerFromConfigAsBytes([]byte(xml))
	if err != nil {
		panic(fmt.Errorf("log configuration parse error: %s", err.Error()))
	}
	seelog.ReplaceLogger(logg)

}

var (
	// Tracef 写一条格式化的日志信息。级别等于 Trace
	Tracef = seelog.Tracef
	// Trace 写一条日志信息。级别等于 Trace
	Trace = seelog.Trace

	// Debugf 写一条格式化的日志信息。级别等于 Debug
	Debugf = seelog.Debugf

	// Debug 写一条日志信息。级别等于 Debug
	Debug = seelog.Debug

	// Infof 写一条格式化的日志信息。级别等于 Info
	Infof = seelog.Infof

	// Info 写一条日志信息。级别等于 Info
	Info = seelog.Info

	// Warnf 写一条格式化的日志信息。级别等于 Warn
	Warnf = seelog.Warnf

	// Warn 写一条日志信息。级别等于 Warn
	Warn = seelog.Warn

	// Errorf 写一条格式化的日志信息。级别等于 Error
	Errorf = seelog.Errorf

	// Error 写一条日志信息。级别等于 Error
	Error = seelog.Error

	// Criticalf 写一条格式化的日志信息。级别等于 Critical
	Criticalf = seelog.Criticalf

	// Critical 写一条日志信息。级别等于 Critical
	Critical = seelog.Critical
)

// Flush 将所有日志信息写入缓存
func Flush() {
	seelog.Flush()
}
