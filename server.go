package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
)

func main() {
	fmt.Printf("Hello World\n")
	http.HandleFunc("/", game)
	http.ListenAndServe(":7777", nil)
}

func game(w http.ResponseWriter, r * http.Request) {
	fmt.Fprintf(w, "Welcome to the page")
}