package main

import (
	"fmt"
	"net/http"

	"github.com/carr123/fmx"
)

//http basic auth
func BasicAuth() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		var bCheckAuth bool = false
		username, password, ok := c.Request.BasicAuth()
		if ok && username == "root" && password == "123456" {
			bCheckAuth = true
		}

		if !bCheckAuth {
			c.Writer.Header().Add("WWW-Authenticate", `Basic realm=""`)
			c.String(401, "Unauthorized")
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

//cross origin
func XOrigin() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		if origin := c.Request.Header.Get("Origin"); origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,HEAD,PUT")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,Accept,Origin,User-Agent,Cache-Control,X-Data-Type,X-Requested-With")
		}

		//preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

func APIGetProfile(c *fmx.Context) {
	username := c.MustGet("username")
	c.String(200, fmt.Sprintf("%s login success", username))
}

func main() {
	mwLogger := fmx.LoggerWithFunc(true, true, fmx.DefaultLoggerFunc())
	router := fmx.NewServeMux()

	router.Use(mwLogger, BasicAuth(), XOrigin())
	router.GET("/api/profile", APIGetProfile)

	fmt.Println("open your web browser, navigate to http://127.0.0.1:8080/api/profile")

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
