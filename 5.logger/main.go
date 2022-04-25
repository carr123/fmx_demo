package main

import (
	"fmt"
	"net/http"

	"github.com/carr123/fmx"
)

//this example shows how to capture goroutine panic logs and http request logs.

func APIGetProfile(c *fmx.Context) {
	c.JSON(200, fmx.H{"name": "jack", "age": 20})
}

//client post a json string to server
func APIPostProfile(c *fmx.Context) {
	var User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	if err := c.ReadReqBodyJson(&User); err != nil {
		fmt.Println(err)
		return
	}

	c.JSON(200, fmx.H{"success": true})
}

var (
	Recover func()
)

func main() {
	Recover = fmx.RecoverFn(func(szLog string) {
		fmt.Println("panic:", szLog)
	})

	go func() {
		defer Recover() //if this goroutine panics, detailed panic logs will be printed out.
		fmt.Println("hello")
	}()

	//create a middleware to print each http request
	// mwLogger := fmx.LoggerWithFunc(true, true, func(c *fmx.Context, szLog []byte) {
	// 	fmt.Println("LOG:", string(szLog))
	// })

	router := fmx.NewServeMux()
	//router.Use(mwLogger)
	router.Use(fmx.Logger(true, true), XOrigin())

	//router.OPTIONS("/api/profile", APIOption)
	router.GET("/api/profile", APIGetProfile)
	router.POST("/api/profile", APIPostProfile)

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}

func APIOption(c *fmx.Context) {
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(200)
		return
	}
	c.Next()
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
