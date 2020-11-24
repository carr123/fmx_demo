package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/carr123/fmx_demo/9.appframework/easyconf/textconf"
)

var (
	conf *textconf.TextConfig
)

func InitAppConfig() error {
	defcfg := "gomq.conf"

	_, path := getAppPath()
	fullPath := filepath.Join(path, defcfg)

	var err error
	conf, err = textconf.FromCmdLine()
	if err != nil {
		conf, err = textconf.FromFile(fullPath)
		if err != nil {
			return err
		}
	}

	fmt.Println("成功读取配置文件:", conf.GetConfFile())
	return nil
}

func getAppPath() (string, string) {
	file, _ := exec.LookPath(os.Args[0])
	apppath, _ := filepath.Abs(file)
	dir := filepath.Dir(apppath)
	return apppath, dir
}

func GetString(key string) string {
	return conf.GetConfString(key)
}

func GetInt(key string) int64 {
	return conf.GetConfInt(key)
}
