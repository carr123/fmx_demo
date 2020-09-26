package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/carr123/fmx"
)

//this example shows how to POST form data to server.
//open your web browser, navigate to http://127.0.0.1:8080/static/index.html
func main() {
	router := fmx.NewServeMux()

	//post image to server through form data
	router.POST("/api/avatar", func(c *fmx.Context) {
		r := c.Request
		r.ParseMultipartForm(32 << 20)

		username := r.Form.Get("name")
		fimg, handler, err := r.FormFile("avatar")
		if err != nil {
			c.String(400, err.Error())
			return
		}

		defer fimg.Close()

		//save avatar to file
		fullPath := filepath.Join(getAppDir(), "www", username+filepath.Ext(handler.Filename))
		os.MkdirAll(filepath.Dir(fullPath), 0777)

		f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			c.String(500, err.Error())
			return
		}

		defer f.Close()

		io.Copy(f, fimg)

		fmt.Println("you post name=" + username + ",file has been saved to " + fullPath)
		c.File(fullPath)
	})

	router.ServeDir("/static", filepath.Join(getAppDir(), "www"))

	fmt.Println("open your web browser, navigate to http://127.0.0.1:8080/static/index.html")

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}

func getAppDir() string {
	file, _ := exec.LookPath(os.Args[0])
	apppath, _ := filepath.Abs(file)
	dir := filepath.Dir(apppath)
	return dir
}
