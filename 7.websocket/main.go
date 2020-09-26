package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"net/http"

	"github.com/carr123/fmx"
	"golang.org/x/net/websocket"
)

func main() {
	router := fmx.NewServeMux()
	router.GET("/ws/long", HandleWS())
	router.ServeDir("/", filepath.Join(getAppDir(), "www"))

	fmt.Println("open your web browser, navigate to http://127.0.0.1:8080")

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

func HandleWS() func(c *fmx.Context) {
	return func(c *fmx.Context) {
		s := websocket.Server{websocket.Config{}, nil, func(ws *websocket.Conn) {
			defer ws.Close()
			var msgRead = make([]byte, 1024*1024)
			var n int
			var err error

			fmt.Printf("ws new client:%s\r\n", c.ClientIP())

			for {
				if n, err = ws.Read(msgRead); err != nil || n < 0 {
					break
				}
				fmt.Printf("websocket read from %s, data:%s\r\n", c.ClientIP(), string(msgRead[:n]))

				if err := websocket.Message.Send(ws, string(msgRead[:n])); err != nil {
					break
				}
			}
		}}

		s.ServeHTTP(c.Writer, c.Request)
	}
}
