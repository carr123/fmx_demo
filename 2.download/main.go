package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/carr123/fmx"
)

//this example shows how can a web browser download a file from server,
//especially when the file content in memory
//and how to use gzip in http response
//open your web browser, navigate to http://127.0.0.1:8080/api/export
func main() {
	router := fmx.NewServeMux()

	//export file (web browser will download this file)
	router.GET("/api/export", func(c *fmx.Context) {
		var output io.Writer
		if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Writer.Header().Set("Content-Encoding", "gzip")
			zipw := gzip.NewWriter(c.Writer)
			defer zipw.Close()
			output = zipw
		} else {
			output = c.Writer
		}

		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.txt", "doc"))
		c.Writer.Header().Set("Content-Type", "application/octet-stream")

		filecontent := fmt.Sprintf("name=%s age=%d", "jack", 20)

		c.Writer.WriteHeader(200)
		io.Copy(output, strings.NewReader(filecontent))
	})

	fmt.Println("open your web browser, navigate to http://127.0.0.1:8080/api/export")

	err := http.ListenAndServe("127.0.0.1:8080", router)
	if err != nil {
		fmt.Println(err)
	}
}
