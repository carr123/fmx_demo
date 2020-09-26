package main

import (
	"fmt"
	"net/http"

	"github.com/carr123/fmx"
)

//middleware is actually a fmx.HandlerFunc, fmx itself supports a set of fmx.HandlerFunc combined to
//serve a http request. each function does its own stuff, when a http request comes, those funtions are
//called one by one.

/*
c.Next()
call next middlewares, only when all middleware functions in the back return, c.Next() return
*/

/*
c.Abort()
ignore all middlewares calls in the back
*/

func Middleware1(c *fmx.Context) {
	//c.Next()
	fmt.Println("in middle 1")
}

func Middleware2(c *fmx.Context) {
	//c.Abort()
	fmt.Println("in middle 2")
}

func Middleware3(c *fmx.Context) {
	fmt.Println("in middle 3")
}

func APIGetProfile(c *fmx.Context) {
	c.JSON(200, fmx.H{"name": "jack", "age": 20})
}

func APIGetUser(c *fmx.Context) {
	c.String(200, "hello")
}

func main() {
	mwLogger := fmx.LoggerWithFunc(true, true, fmx.DefaultLoggerFunc())
	router := fmx.NewServeMux()

	//mwLogger, Middleware1 are added to router
	//so afterwards, each http request will apply these two funtions.
	router.Use(mwLogger, Middleware1)

	router.GET("/api/profile", Middleware2, Middleware3, APIGetProfile) //2+3=5 middlewares
	router.GET("/api/user", Middleware3, APIGetUser)                    //2+2=4 middlewares

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
