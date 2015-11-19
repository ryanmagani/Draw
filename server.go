package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
)

func main() {
	fmt.Printf("Hello World\n")
	// http.Handle("/", http.FileServer(http.Dir("/site")))
	// http.HandleFunc("/", game)
	// http.HandleFunc("/draw", draw)
	// http.HandleFunc("/guess", guess)
	http.ListenAndServe(":7777", http.FileServer(http.Dir("./public")))
}

// func game(w http.ResponseWriter, r * http.Request) {
// 	// fmt.Fprintf(w, "Welcome to the page")
// 	// http.ServeFile(w, r, r.URL.Path)
	// http.FileServer(http.Dir("/site"))
	// http.ServeFile(w, r, "css/compressed.css")
// }

// func draw(w http.ResponseWriter, r * http.Request) {
// }

// func guess(w http.ResponseWriter, r * http.Request) {
// }