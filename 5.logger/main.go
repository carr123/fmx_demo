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
	router.Use(fmx.Logger(false, false))

	router.GET("/api/profile", APIGetProfile)
	router.POST("/api/profile", APIPostProfile)

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
