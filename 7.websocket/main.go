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
		s := websocket.Server{websocket.Config{}, nil, func(conn *websocket.Conn) {
			defer conn.Close()

			fmt.Printf("ws new client:%s\r\n", c.ClientIP())

			conn.MaxPayloadBytes = 32 << 20 // 32MB

			var szNewMsg string

			for {

				if err := websocket.Message.Receive(conn, &szNewMsg); err != nil {
					break
				}

				fmt.Printf("websocket read from %s, data:%s\r\n", c.ClientIP(), szNewMsg)

				if err := websocket.Message.Send(conn, szNewMsg); err != nil {
					break
				}
			}
		}}

		s.ServeHTTP(c.Writer, c.Request)
	}
}
