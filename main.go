package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	indexFile, err := os.Open("templates/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = indexFile.Close()

	jsFile, err := os.Open("templates/index.js")
	if err != nil {
		fmt.Println(err)
		return
	}
	js, err := ioutil.ReadAll(jsFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = jsFile.Close()

	srv := newServer()

	http.HandleFunc("/websocket", srv.handle)
	http.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(js))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})

	go srv.listen()

	http.ListenAndServe(":3000", nil)
}
