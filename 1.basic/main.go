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

	fmt.Println("version:", fmx.Version)

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}

func splitURLpath(path string) (parts []string, names map[string]int) {

	var (
		nameidx      int = -1
		partidx      int
		paramCounter int
	)

	for i := 0; i < len(path); i++ {
		// recording name
		if nameidx != -1 {
			//found /
			if path[i] == '/' {

				if names == nil {
					names = make(map[string]int)
				}

				names[path[nameidx:i]] = paramCounter
				paramCounter++

				nameidx = -1 // switch to normal recording
				partidx = i
			}
		} else {
			if path[i] == ':' || path[i] == '*' {
				if path[i-1] != '/' {
					panic(fmt.Errorf("Inválid parameter : or * comes anwais after / - %q", path))
				}
				nameidx = i + 1
				if partidx != i {
					parts = append(parts, path[partidx:i])
				}
				parts = append(parts, path[i:nameidx])
			}
		}
	}

	if nameidx != -1 {
		if names == nil {
			names = make(map[string]int)
		}
		names[path[nameidx:]] = paramCounter
		paramCounter++
	} else if partidx < len(path) {
		parts = append(parts, path[partidx:])
	}
	return
}
