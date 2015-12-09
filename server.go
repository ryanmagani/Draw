package main

import (
	"./lib/go.net/websocket"
	"fmt"
	"net/http"
	"sync"
	"math/rand"
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
	score int
	output chan Packet
}

type Game struct {
	// most recently assigned client ID
	maxId uint64
	// current word to guess
	word string
	clients []*Client
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
		0,
		[BOARD_SIZE][BOARD_SIZE]int{},
		&sync.Mutex{}}

	fmt.Println("Game Started on port 7777")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/socket", websocket.Handler(handleSocketIn))
	http.ListenAndServe(":7777", nil)
}

// requires that game is locked
func isDrawer(c *Client) bool {
	return c == game.clients[game.drawerIndex]
}

func nextWord() string {
	return "newWord"
}

// requires that game is locked
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

// requires that game is locked
// sends the packet to all channels, modifying the
// packet to set IsDrawer to true for the drawer
func updateAllChan(packet Packet) {
	updateNonDrawer(packet)
	packet.IsDrawer = true
	game.clients[game.drawerIndex].output <- packet
}

// requires that game is locked
// sends the packet to all channels,
// except that of the drawer
func updateNonDrawer(packet Packet) {
	for i := 0; i < len(game.clients); i++ {
		if (i != game.drawerIndex) {
			game.clients[i].output <- packet
		}
	}
}

func handleSocketIn(ws *websocket.Conn) {
	// setup connection with new user
	// store their information in the game
	// return a piece of information regarding whether or not they are drawing
	currClient := join(ws)
	handleSocket(currClient)
}

func join(ws *websocket.Conn) *Client {
	game.Lock()
	defer game.Unlock()

	isDrawer := false
	if (len(game.clients) == 0) {
		isDrawer = true
	}

	fmt.Println("Debug: client joined, isDrawer:", isDrawer)

	pkt := Packet{"init",
		getBoard(),
		"",
		isDrawer,
		""}

	websocket.JSON.Send(ws, pkt)

	newClient := &Client{game.maxId, ws, 0, make(chan Packet, 1)}

	// increment maxId
	game.maxId++

	game.clients = append(game.clients, newClient)
	return newClient;
}

func handleSocket(currClient * Client) {
	input := make(chan Packet, 1)
	go func() {
		var packet Packet
		for {
			err := websocket.JSON.Receive(currClient.ws, &packet)

			if err != nil {
				fmt.Println("Debug: websocket is closed")
				return
			}

			input<-packet
		}
	}()

	for {
		select {
		case packet := <-currClient.output:
			websocket.JSON.Send(currClient.ws, packet)
		case packet := <-input:
			switch packet.Ptype {
			case "ack":
				handleAck(currClient, packet)
			case "guess":
				handleGuess(currClient, packet)
			case "draw":
				handleDraw(currClient, packet)
			case "clear":
				handleClear(currClient)
			case "quit":
				handleQuit(currClient)
				return
			}
		}
	}
}

func handleAck(currClient * Client, packetIn Packet) {
	fmt.Println("ack recv")
}

// If the guess is correct, update the guesser and
// alert all clients of a change in word, otherwise,
// do nothing
func handleGuess(currClient * Client, packetIn Packet) {
	game.Lock()
	defer game.Unlock()

	if isDrawer(currClient) {
		fmt.Println("Debug: a drawer tried to guess")
		return
	}

	fmt.Println("Debug: guesser guessing", packetIn.Data, "actual", game.word)

	if game.word == packetIn.Data {
		game.word = nextWord()
		game.canvas = [BOARD_SIZE][BOARD_SIZE]int{}
		packetOut := Packet{Ptype: "next",
					Board: nil,
					Color: "",
					IsDrawer: false,
					Data: ""} // TODO: set data to the new drawer's name

		// TODO: potentially call updateAllChan, though this
		// is more efficient
		for i := 0; i < len(game.clients); i++ {
			if game.clients[i] == currClient {
				// tell the guesser that s/he has correctly
				// guessed the word
				game.drawerIndex = i
				packetOut.IsDrawer = true
				websocket.JSON.Send(game.clients[i].ws, packetOut)
				packetOut.IsDrawer = false
			} else {
				// delegate each client to send a packet on their
				// own so that if that 'send' fails, it does not
				// affect other clients
				game.clients[i].output <- packetOut
			}
		}
	}
}

// Send the drawing to all the clients and update
// our internal representation of the game board
func handleDraw(currClient * Client, packetIn Packet) {
	game.Lock()
	defer game.Unlock()

	if !isDrawer(currClient) {
		fmt.Println("Debug: a guesser tried to draw")
		return
	}

	fmt.Println("Debug: drawer drawing")

	packetOut := Packet{Ptype: "draw",
						Board: packetIn.Board,
						Color: packetIn.Color,
						IsDrawer: false,
						Data: ""}

	updateNonDrawer(packetOut)

	// TODO: update internal reprsentation, no point
	// doing this right now if we're going to change it
}

func handleClear(currClient * Client) {
	game.Lock()
	defer game.Unlock()

	if !isDrawer(currClient) {
		fmt.Println("Debug: a guesser tried to clear")
		return
	}

	fmt.Println("drawer clearing")

	packetOut := Packet{Ptype: "clear",
						Board: nil,
						Color: "",
						IsDrawer: false,
						Data: ""}

	updateNonDrawer(packetOut)
}

// Remove the client from the list of clients and close
// his/her websocket
// If this client was the drawer, assign some random
// guesser to be the drawer and start a new round
// If the last drawer quit, do not assign a new drawer
func handleQuit(currClient * Client) {
	game.Lock()
	defer game.Unlock()

	currClient.ws.Close()

	isDrawer := isDrawer(currClient)

	fmt.Println("Debug: client quitting, isDrawer:", isDrawer)

	// increment i until the we find the index of
	// the quitting client
	var i int
	for i = 0; game.clients[i] != currClient; i++ {
	}

	game.clients = append(game.clients[:i], game.clients[i+1:]...)

	var packetOut Packet
	if len(game.clients) == 0 {
		game.drawerIndex = 0
		game.word = nextWord()
		game.canvas = [BOARD_SIZE][BOARD_SIZE]int{}
		game.clients = make([]*Client, 0)
		return
	} else if isDrawer {
		// the drawer just quit, clear current game state
		game.word = nextWord()
		game.canvas = [BOARD_SIZE][BOARD_SIZE]int{}

		// otherwise, randomly assign a new drawer and
		// set up a new round
		game.drawerIndex = rand.Intn(len(game.clients))

		packetOut = Packet{Ptype: "drawerQuit",
					Board: nil,
					Color: "",
					IsDrawer: false,
					Data: ""} // TODO: set data to the new drawer's name
	} else {
		// otherwise, tell everyone about the quit anyways so
		// any leaderboards, etc. can be updated
		if i < game.drawerIndex {
			// if the quitter's index was lower than the drawers,
			// adjust accordingly
			game.drawerIndex--
		}

		packetOut = Packet{Ptype: "otherQuit",
					Board: nil, // TODO: this does NOT imply that 
								// the board should be cleared,
								// what should we do here?
					Color: "",
					IsDrawer: false,
					Data: ""} // TODO: set data to the quitter's username
	}

	updateAllChan(packetOut)
}
