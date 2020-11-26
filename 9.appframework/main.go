package main

import (
	"fmt"
	"net/http"

	"github.com/carr123/fmx"
)

func main() {
	var err error

	if err = appInit(); err != nil {
		fmt.Println(err)
		return
	}

	mux := fmx.NewServeMux()
	mux.Use(HTTPLogger(false, false))
	//mux.GET("/api/profile", APIGetProfile)
	//mux.POST("/api/profile", APIPostProfile)

	err = http.ListenAndServe("127.0.0.1:8080", mux)
	if err != nil {
		fmt.Println(err)
	}
}
