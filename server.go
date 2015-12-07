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
	Ptype string `json:"Type"`
	Board []Point `json:"Board"`
	Color string `json:"Color"`
	IsDrawer bool `json:"IsDrawer"`
	Data string `json:"Data"`
}

type Client struct {
	_id  uint64
	ws *websocket.Conn
	isDrawer bool
	score int
}

type Game struct {
	// most recently assigned client ID
	maxId uint64
	// current word to guess
	word string
	clients []*Client
	guessCorrect chan bool
	drawerIndex int
	canvas [BOARD_SIZE][BOARD_SIZE]int
	*sync.Mutex
}

// Single game per server
var game Game;

// Setup the game and file serving
func main() {
	game = Game {0,
		"newGame",
		make([]*Client, 0),
		make(chan bool, 1),
		0,
		[BOARD_SIZE][BOARD_SIZE]int{},
		&sync.Mutex{}}

	fmt.Println("Game Started on port 7777")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/socket", websocket.Handler(handleSocketIn))
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

func assignDrawer() *Client{
	game.Lock()
	fmt.Println("I'm hard")
	pkt := Packet{"nextWord",
		nil,
		"",
		false,
		""}

	var returnClient *Client

	for i := 0; i < len(game.clients); i++ {
		if (i == game.drawerIndex) {
			returnClient = game.clients[i]
			pkt.IsDrawer = true
		} else {
			pkt.IsDrawer = false
		}
		websocket.JSON.Send(game.clients[i].ws, pkt)

	}
	game.Unlock()
	return returnClient
}

func handleDrawer(currClient *Client) {
	input := make(chan Packet, 1)
	go func() {
		var packet Packet
		for {
			websocket.JSON.Receive(currClient.ws, &packet)
			input<-packet
		}
	}()

	for {
		select {

		case <-game.guessCorrect:
			fmt.Println("wtf")
			currClient = assignDrawer()

		case packet := <-input:
			fmt.Println("clr: ", packet.Color)
			fmt.Println("arr: ", packet.Board)

			game.Lock()

			if packet.Ptype == "quit" {
				quit(currClient)
				return
			} else {


				colorVal := 1
				if packet.Color == "white" {
					colorVal = 0
				}

				for i := 0; i < len(packet.Board); i++ {
					game.canvas[packet.Board[i].X][packet.Board[i].Y] = colorVal
				}

				for i := 0; i < len(game.clients); i++ {
					if i != game.drawerIndex {
						websocket.JSON.Send(game.clients[i].ws, packet)
					}
				}

				game.Unlock()
			}
		}
	}
}

func handleGuesser(currClient *Client) {
	var packet Packet
	for {
		websocket.JSON.Receive(currClient.ws, &packet)
		fmt.Println(packet.Ptype)
		switch {
		case packet.Ptype == "guess":
			fmt.Println(packet.Data)
			if game.word == packet.Data {
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

		case packet.Ptype == "quit":
			fmt.Println("quitting...")
			quit(currClient)
			return
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
		isDrawer,
		""}

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

func quit(currClient *Client) {
	currClient.ws.Close()
	// getguid
	// GM.games[].c
	// if we are the drawer, assign new drawer
	if (currClient == game.clients[game.drawerIndex]) {
		game.clients = append(game.clients[:game.drawerIndex], game.clients[game.drawerIndex+1:]...);
		if (game.drawerIndex >= len(game.clients)) {
			game.drawerIndex = 0;
		}
	} else {
		i := 0
		for ; game.clients[i] != currClient; i++ {}
		game.clients = append(game.clients[:i], game.clients[i+1:]...);
	}

}
