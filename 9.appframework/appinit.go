package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/carr123/fmx"

	"github.com/carr123/easyconf/textconf"
	"github.com/carr123/easylog"
)

var (
	conf     *textconf.TextConfig
	httplog  *easylog.EasyLog
	debuglog *easylog.EasyLog
	errorlog *easylog.EasyLog
	paniclog *easylog.EasyLog
	Recover  func()
)

func appInit() error {
	var err error

	rand.Seed(time.Now().UnixNano())

	if err = confInit(); err != nil {
		return err
	}

	if err = logInit(); err != nil {
		return err
	}

	return nil
}

//初始化程序配置信息
func confInit() error {
	var err error

	defcfg := "backup.conf"
	// if runtime.GOOS == "windows" {
	// 	defcfg = "backup.conf"
	// }

	_, path := getAppPath()
	fullPath := filepath.Join(path, defcfg)

	conf, err = textconf.FromCmdLine()
	if err != nil {
		conf, err = textconf.FromFile(fullPath)
		if err != nil {
			return err
		}
	}

	return nil
}

//初始化程序日志信息
func logInit() error {
	logdir := conf.GetConfString("logdir")

	httplog = easylog.NewLog(1000, time.Second)
	httplog.SetDir(logdir, "http.txt")
	httplog.SetMaxFileSize(1024 * 1024 * 4)
	httplog.SetMaxFileCount(5)

	debuglog = easylog.NewLog(1000, time.Second)
	debuglog.SetDir(logdir, "debug.txt")
	debuglog.SetMaxFileSize(1024 * 1024 * 4)
	debuglog.SetMaxFileCount(5)

	errorlog = easylog.NewLog(1000, time.Second)
	errorlog.SetDir(logdir, "error.txt")
	errorlog.SetMaxFileSize(1024 * 1024 * 4)
	errorlog.SetMaxFileCount(5)

	paniclog = easylog.NewLog(1000, time.Second)
	paniclog.SetDir(logdir, "panic.txt")
	paniclog.SetMaxFileSize(1024 * 1024 * 4)
	paniclog.SetMaxFileCount(5)

	Recover = fmx.RecoverFn(func(szLog string) {
		paniclog.Write([]byte(szLog))
	})

	return nil
}

func WriteErrLog(err error) {
	var errstr string
	if e, ok := err.(fmx.ErrorWithPos); ok {
		errstr = e.String()
	} else {
		errstr = err.Error()
	}

	szInfo := fmt.Sprintf("%s %s\n\n", time.Now().Format("2006-01-02 15:04:05"), errstr)
	errorlog.Write([]byte(szInfo))
}

func getAppPath() (string, string) {
	file, _ := exec.LookPath(os.Args[0])
	apppath, _ := filepath.Abs(file)
	dir := filepath.Dir(apppath)
	return apppath, dir
}

func HTTPLogger(bShowReqBody bool, bShowRespBody bool) fmx.HandlerFunc {
	return fmx.LoggerWithFunc(bShowReqBody, bShowRespBody, func(c *fmx.Context, szLog []byte) {
		httplog.Write(szLog)
	})
}
