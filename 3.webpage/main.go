package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/carr123/fmx"
)

//this example shows how to server static web pages.
//open your web browser, navigate to http://127.0.0.1:8080/static/index.html
func main() {
	router := fmx.NewServeMux()
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
