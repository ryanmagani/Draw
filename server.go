package main

import (
	"./lib/go.net/websocket"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const BOARD_SIZE = 400



type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Golang will NOT send out data whose
// var name is uncapitalized...
type Packet struct {
	Ptype string `json:"type"`
	Board []Point `json:"points"`
	Color string `json:"color"`
	IsDrawer bool `json:"isDrawer"`
}

type Client struct {
	_id  uint64
	ws *websocket.Conn
	isDrawer bool
	score int
}

type Game struct {
	maxId uint64
	word string
	clients []*Client
	guessCorrect chan bool
	drawerIndex int
	canvas [BOARD_SIZE][BOARD_SIZE]int
	// num clients is len(g.clients) where g is some Game
	// initialize with g := Game{"stringlit", []int{int, lits, 3}}
	*sync.Mutex // apparently this is initialized with "&sync.Mutex{}"
				// which gives a unique lock each time...
}

var game Game;

func main() {
	game = Game {0,
		"newGame",
		make([]*Client, 0),
		make(chan bool, 1),
		0,
		[BOARD_SIZE][BOARD_SIZE]int{},
		&sync.Mutex{}}

	fmt.Printf("Hello World\n")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	// http.HandleFunc("/draw", draw)
	// http.HandleFunc("/guess", guess)
	http.Handle("/socket", websocket.Handler(handleSocketIn))
	// http.HandleFunc("/quit", quit)
	http.ListenAndServe(":7777", nil)
}

func nextWord() {
	game.Lock()
	defer game.Unlock()
	game.drawerIndex++
	game.drawerIndex = game.drawerIndex % len(game.clients)
	game.word = "new"
}

func handleSocketIn(ws *websocket.Conn) {
	// setup connection with new user
	// store their information in the game
	// return a piece of information regarding whether or not they are drawing
	currClient := join(ws)
	if currClient.isDrawer {
		handleDrawer(currClient)
	} else {
		handleGuesser(currClient)
	}
}

func handleDrawer(currClient *Client) {
//	var pkt Packet
	input := make(chan Point, 1)
	go func() {
		var point Point
		for {
			websocket.JSON.Receive(currClient.ws, &point)
			input<-point
		}
	}()

	for {
		select {

		case <-game.guessCorrect:
			game.Lock()
			fmt.Println("I'm hard")
			pkt := Packet{"nextWord",
				nil,
				"",
				false}

			for i := 0; i < len(game.clients); i++ {
				if (i == game.drawerIndex) {
					currClient = game.clients[i]
					pkt.IsDrawer = true
				} else {
					pkt.IsDrawer = false
				}
				websocket.JSON.Send(game.clients[i].ws, pkt)

			}
			game.Unlock()

		case pnt := <-input:
		// case websocket.JSON.Receive(currClient.ws, &point):
			//			drawnPoints := Packet{}
			
			
			fmt.Println("x: ", pnt.X, "y: ", pnt.Y)
			game.Lock()
			game.canvas[pnt.X][pnt.Y] = 1
			game.Unlock()
		}
	}
}

func handleGuesser(currClient *Client) {
	var guess string
	for {
		websocket.JSON.Receive(currClient.ws, &guess)
		if game.word == guess {
			// guessed correctly, switch ourselves with drawer
			game.Lock()
			currDrawer := game.clients[game.drawerIndex]

			// find our current index
			i := 0

			for ; game.clients[i]._id != currClient._id; i++ { }

			// set drawer index to our index
			game.drawerIndex = i
			game.guessCorrect <- true

			// set ourselves to old drawer
			currClient = currDrawer

			game.canvas = [BOARD_SIZE][BOARD_SIZE]int{}

			game.Unlock()
		} else {
			// client guessed wrong
		}
	}

}

func draw(w http.ResponseWriter, r * http.Request) {
	// parse the request looking for:
		// which game it belongs to
		// what user is drawing
		// what the user drew
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf(string(body) + "\n")
}

func guess(w http.ResponseWriter, r * http.Request) {
	fmt.Printf("guess rcvd\n")
}

func getBoard() []Point {
	drawnPoints := make([]Point, 0)
	for i := 0; i < BOARD_SIZE; i++ {
		for j := 0; j < BOARD_SIZE; j++ {
			if game.canvas[i][j] == 1 {
				drawnPoints = append(drawnPoints, Point{i,j})
			}
		}
	}
	return drawnPoints
}

func join(ws *websocket.Conn) *Client {
	// parse the request looking for:
		// which game the user wants to join
		// what user is trying to join
	//c := null

	game.Lock()
	defer game.Unlock()

	isDrawer := false
	if (len(game.clients) == 0) {
		isDrawer = true
	}


	pkt := Packet{"init",
		getBoard(),
		"",
		isDrawer}

/*	drawnPoints := make([]Point, 2)
	drawnPoints[0] = Point{1,1}
	drawnPoints[1] = Point{2,2}*/
	websocket.JSON.Send(ws, pkt)

	newClient := &Client{game.maxId, ws, isDrawer, 0}

	// increment maxId
	game.maxId++

	game.clients = append(game.clients, newClient)
	return newClient;
}

func quit(w http.ResponseWriter, r * http.Request) {
	// getguid
	// GM.games[].c
}
