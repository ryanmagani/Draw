package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
)

func main() {
	fmt.Printf("Hello World\n")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/draw", draw)
	http.ListenAndServe(":7777", nil)
}

func draw(w http.ResponseWriter, r * http.Request) {
	fmt.Printf("draw recieved")
}

// func guess(w http.ResponseWriter, r * http.Request) {
// }