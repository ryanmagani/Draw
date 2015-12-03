package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
	"sync"
)

type Game struct {
	word string;
	clients []int;
	drawer int;
	// num clients is len(g.clients) where g is some Game
	// initialize with g := Game{"stringlit", []int{int, lits, 3}}
	*sync.Mutex // apparently this is initialized with "&sync.Mutex{}"
				// which gives a unique lock each time...
}

type GameManager struct {
	games []Game;
	*sync.Mutex
}

var GM GameManager;

func main() {
	GM = GameManager{[]Game{}, &sync.Mutex{}}
	fmt.Printf("Hello World\n")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/draw", draw)
	http.HandleFunc("/guess", guess)
	http.HandleFunc("/join", join)
	http.HandleFunc("/quit", quit)
	http.ListenAndServe(":7777", nil)
}

func NewGame() {
	GM.Lock()
	// defer statements called after function finishes
	defer GM.Unlock()
	g := Game{"newGame", []int{}, 0, &sync.Mutex{}}
	GM.games = append(GM.games, g)
	// somehow spin off a thread for this game
}

func nextWord(g Game) {
	g.Lock()
	defer g.Unlock()
	g.drawer++
	g.drawer = g.drawer % len(g.clients)
	g.word = "new"
}

func draw(w http.ResponseWriter, r * http.Request) {
	// parse the request looking for:
		// which game it belongs to
		// what user is drawing
		// what the user drew
	fmt.Printf("draw recieved\n")
}

func guess(w http.ResponseWriter, r * http.Request) {
	fmt.Printf("guess rcvd\n")
}

func join(w http.ResponseWriter, r * http.Request) {
	// parse the request looking for:
		// which game the user wants to join
		// what user is trying to join
	c := 0
	GM.games[0].Lock() // CHANGE THIS FROM 0
	defer GM.games[0].Unlock()
	GM.games[0].clients = append(GM.games[0].clients, c)
}

func quit(w http.ResponseWriter, r * http.Request) {
	// getguid
	// GM.games[].c
}

// func guess(w http.ResponseWriter, r * http.Request) {
// }
