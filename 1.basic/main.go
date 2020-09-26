package main

import (
	"fmt"
	"net/http"

	"github.com/carr123/fmx"
)

//http restful api demo
//basic GET and POST demo

func APIGetProfile(c *fmx.Context) {
	//fmx.H is actually a map
	c.JSON(200, fmx.H{"name": "jack", "age": 20}) //send json
	//c.String(200, "hello world") //send plain text
	//c.File("e:\\111.jpg") //send a file
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

func main() {
	router := fmx.NewServeMux()

	router.GET("/api/profile", APIGetProfile)
	router.POST("/api/profile", APIPostProfile)

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
