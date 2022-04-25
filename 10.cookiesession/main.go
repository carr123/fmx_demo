package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/carr123/ssdbsession"

	"github.com/carr123/fmx"
)

var (
	sessMgr *ssdbsession.SessionMgr
)

func main() {
	var err error
	//浏览器cookie 600 秒过期， 服务器session 1000秒过期
	szAddr := "192.168.0.103:6379"
	sessMgr, err = ssdbsession.New(szAddr, "", "yoocco", 7200, 7200, false, "")
	if err != nil {
		fmt.Println("redis session:", err)
		return
	}

	router := fmx.NewServeMux()
	router.Use(fmx.Logger(true, true))
	router.GET("/api/login", handleLogin)
	router.GET("/api/readdata", checkCookie, handleReadData)
	router.GET("/api/loginout", handleLoginOut)
	router.GET("/api/loginoutbyusrid", handleLoginOut)

	mux := http.NewServeMux()
	mux.Handle("/api/", router)

	fmt.Println("listen:", "http://127.0.0.1/api/login")
	err = http.ListenAndServe("127.0.0.1:80", mux)
	if err != nil {
		fmt.Println(err)
	}
}

func handleLogin(c *fmx.Context) {
	//删除旧的会话
	szCookie := sessMgr.GetCookie(c.Request)
	sessMgr.DelSession(szCookie)

	//把其他同账号登录用户踢下线
	sessMgr.DelAllSessionByUserID("user001")

	//创建新的会话
	sid, err := sessMgr.NewSession("user001")
	if err != nil {
		log.Println(err)
		return
	}

	sessMgr.SetCookie(c.Writer, c.Request, sid)
	sessMgr.SetSessionValue(sid, "name", "jack", "age", 22, "score", 11.56, "time", time.Now().Unix(), "img", []byte{1, 2, 3, 4})

	c.JSON(200, fmx.H{"status": true, "desc": "登录成功"})
}

func handleLoginOut(c *fmx.Context) {
	szCookie := sessMgr.GetCookie(c.Request)

	bValid, err := sessMgr.IsSessionValid(szCookie)
	if err != nil {
		log.Println(err)
		return
	}

	if bValid {
		if err := sessMgr.DelSession(szCookie); err != nil {
			log.Println(err)
			return
		}

		sessMgr.DelCookie(c.Writer, c.Request)
		c.JSON(200, fmx.H{"status": true, "desc": "退出成功"})
	} else {
		c.JSON(200, fmx.H{"status": true, "desc": "session 不存在"})
	}
}

func checkCookie(c *fmx.Context) {
	szCookie := sessMgr.GetCookie(c.Request)

	bValid, err := sessMgr.IsSessionValid(szCookie)
	if err != nil {
		log.Println(err)
		c.String(500, "server error")
		c.Abort()
		return
	}

	if !bValid {
		sessMgr.DelCookie(c.Writer, c.Request)
		c.String(401, "Not Login")
		c.Abort()
		return
	}

	c.Set("session", szCookie)

	c.Next()

	//刷新会话最后访问时间，使得会话不过期
	sessMgr.SessionKeepAlive(szCookie, time.Now().Unix())
}

func handleReadData(c *fmx.Context) {
	session := c.MustGet("session").(string)

	name, err := sessMgr.GetSessionString(session, "name")
	if err != nil {
		log.Println(err)
		return
	}

	age, err := sessMgr.GetSessionInt64(session, "age")
	if err != nil {
		log.Println(err)
		return
	}

	score, err := sessMgr.GetSessionFloat64(session, "score")
	if err != nil {
		log.Println(err)
		return
	}

	nTime, err := sessMgr.GetSessionInt64(session, "time")
	if err != nil {
		log.Println(err)
		return
	}

	img, err := sessMgr.GetSessionBytes(session, "img")
	if err != nil {
		log.Println(err)
		return
	}

	s := fmt.Sprintf("name=%s\nage=%d\nscore=%f\ntime=%d\nimg=%v \n\n", name, age, score, nTime, img)
	c.String(200, s)
	//sessMgr.SessionKeepAlive(szCookie)
}
