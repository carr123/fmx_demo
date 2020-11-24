package textconf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//仿nginx配置文件格式, key=value, 注释从#开始至一行结束

type TextConfig struct {
	szFile  string
	content []byte
	mp      map[string]string
	lock    *sync.Mutex
}

//命令行参数为配置文件路径
func FromCmdLine() (*TextConfig, error) {
	if len(os.Args) < 2 {
		return nil, errors.New("请输入启动命令行配置参数")
	}

	v := &TextConfig{
		szFile: os.Args[1],
		mp:     make(map[string]string),
		lock:   new(sync.Mutex),
	}

	err := v.reloadConf()
	if err != nil {
		return nil, err
	}

	return v, nil
}

//直接加载配置文件
func FromFile(szCfgFile string) (*TextConfig, error) {
	v := &TextConfig{
		szFile: szCfgFile,
		mp:     make(map[string]string),
		lock:   new(sync.Mutex),
	}

	err := v.reloadConf()
	if err != nil {
		return nil, err
	}

	return v, nil
}

//环境变量值为配置文件路径
func FromEnv(env string) (*TextConfig, error) {
	cfg := os.Getenv(env)
	if cfg == "" {
		s := fmt.Sprintf("environment variable %s not set", env)
		return nil, errors.New(s)
	}

	v := &TextConfig{
		szFile: cfg,
		mp:     make(map[string]string),
		lock:   new(sync.Mutex),
	}
	err := v.reloadConf()
	if err != nil {
		return nil, err
	}
	/*
		go func() {
			time.Sleep(5)
			v.reloadConf()
		}()
	*/
	return v, nil
}

func FromString(szContent []byte) (*TextConfig, error) {
	v := &TextConfig{
		szFile:  "",
		content: szContent,
		mp:      make(map[string]string),
		lock:    new(sync.Mutex),
	}

	if err := v.parseConfig(); err != nil {
		return nil, err
	}

	return v, nil
}

///////////////////////////////////////////////////////////////
func (this *TextConfig) reloadConf() error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	if err := this.readFile(this.szFile); err != nil {
		return err
	}

	return nil
}

func (this *TextConfig) readFile(szFile string) error {
	var err error

	if len(szFile) == 0 {
		return nil
	}

	this.content, err = ioutil.ReadFile(szFile)
	if err != nil {
		return err
	}

	if err = this.parseConfig(); err != nil {
		return err
	}

	return nil
}

func (this *TextConfig) parseConfig() error {
	r := bufio.NewReader(bytes.NewReader(this.content))

	newmp := make(map[string]string)

	re, _ := regexp.Compile("#.*")

	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		content := re.ReplaceAllString(string(b), "")
		index := strings.Index(content, "=")
		if index < 1 {
			continue
		}

		first := strings.TrimSpace(content[:index])
		if len(first) == 0 {
			continue
		}

		second := strings.TrimSpace(content[index+1:])

		if _, ok := newmp[first]; ok {
			return errors.New("配置项 " + first + " 重复!")
		}

		newmp[first] = second
	}

	this.lock.Lock()
	this.mp = newmp
	this.lock.Unlock()

	return nil
}

func (this *TextConfig) GetConfString(key string) string {
	this.lock.Lock()
	defer this.lock.Unlock()
	if s, ok := this.mp[key]; ok {
		return s
	}

	panic(fmt.Sprintf("配置项读取失败, key = %s", key))
	return ""
}

func (this *TextConfig) SetConfString(key string, value string) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.mp[key] = value
	return
}

func (this *TextConfig) MustGetConfString(key string, def string) string {
	this.lock.Lock()
	defer this.lock.Unlock()
	if s, ok := this.mp[key]; ok {
		return s
	}

	return def
}

func (this *TextConfig) GetConfInt(key string) int64 {
	this.lock.Lock()
	defer this.lock.Unlock()
	if s, ok := this.mp[key]; ok {
		i, _ := strconv.ParseInt(s, 0, 0)
		return i
	}

	panic("配置项读取失败")
	return 0
}

func (this *TextConfig) SetConfInt(key string, value int64) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.mp[key] = fmt.Sprintf("%d", value)
	return
}

func (this *TextConfig) GetConfBool(key string) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if s, ok := this.mp[key]; ok {
		if s == "true" {
			return true
		} else if s == "false" {
			return false
		}
	}

	panic("配置项读取失败")

	return false
}

func (this *TextConfig) SetConfBool(key string, value bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	this.mp[key] = fmt.Sprintf("%v", value)
	return
}

func (this *TextConfig) GetConfFile() string {
	return this.szFile
}

func (this *TextConfig) SaveTo(w io.Writer) error {
	mp := make(map[string]string)

	this.lock.Lock()
	for k, v := range this.mp {
		mp[k] = v
	}
	this.lock.Unlock()

	r := bufio.NewReader(bytes.NewReader(this.content))
	newContent := new(bytes.Buffer)

	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var cfg string
		var comments string
		cfg, comments = split_cfg_comments(string(b))
		key, space, ok := parsecfg(cfg)
		if !ok {
			newContent.Write(b)
			newContent.WriteString("\r\n")
			continue
		} else {
			value := this.GetConfString(key)
			newLine := fmt.Sprintf("%s=%s%s%s\r\n", key, value, space, comments)
			newContent.WriteString(newLine)
			delete(mp, key)
			continue
		}
	}

	if len(mp) > 0 {
		newContent.WriteString("\r\n")
		for k, v := range mp {
			newContent.WriteString(fmt.Sprintf("%s=%s\r\n", k, v))
		}
	}

	io.Copy(w, newContent)

	return nil
}

func (this *TextConfig) SaveToFile(fname string) error {
	err := os.MkdirAll(filepath.Dir(fname), 0777)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	return this.SaveTo(f)
}

func split_cfg_comments(s string) (string, string) {
	var cfg string
	var comments string
	ab := strings.Split(s, "#")
	if len(ab) == 1 {
		cfg = ab[0]
	} else {
		cfg = ab[0]
		comments = s[len(cfg):]
	}

	return cfg, comments
}

func parsecfg(s string) (string, string, bool) {
	re, _ := regexp.Compile(`^(.*?)=(.*?)(\s*)$`)
	res := re.FindStringSubmatch(s)
	if len(res) == 4 {
		return strings.TrimSpace(res[1]), res[3], true
	}

	return "", "", false
}
