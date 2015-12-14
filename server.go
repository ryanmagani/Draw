package main

import (
	"./lib/go.net/websocket"
	"fmt"
	"net/http"
	"sync"
	"math/rand"
	"time"
)

const BOARD_SIZE = 400
const DELAY_HISTORY_MAX = 5
const LATENCY_COUNTER_MAX = 10
const CHAN_SIZE = 10

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
	PrevX int `json:"prevX"`
	PrevY int `json:"prevY"`
}

// Golang will NOT send out data whose
// var name is uncapitalized...
type Packet struct {
	Ptype string `json:"Type"`
	Board []Point `json:"Board"`
	Leaderboard map[string]int `json:"Leaderboard",omitempty`
	Color string `json:"Color",omitempty`
	IsDrawer bool `json:"IsDrawer",omitempty`
	Data string `json:"Data",omitempty`
	Date int64 `json:"Date",omitempty`
	Delay int64 `json:"Delay",omitempty`
}

type Client struct {
	_id  uint64
	ws *websocket.Conn
	name string
	output chan Packet
	delayHistory []int64
	delay int64
}

type Game struct {
	// most recently assigned client ID
	maxId uint64
	// current word to guess
	word string
	wordValid bool
	clients []*Client
	drawerIndex int
	//	canvas [BOARD_SIZE][BOARD_SIZE]int
	board []Point
	*sync.Mutex
}

type Latency struct {
	maxDelay int64
	counter int
	*sync.Mutex
}

// Single game per server
var game Game;
var leaderboard map[string]int;
var latency Latency;

// Setup the game and file serving
func main() {
	game = Game{0,
		"",
		false,
		make([]*Client, 0),
		0,
		make([]Point, 0),
		&sync.Mutex{}}

	leaderboard = make(map[string]int)

	latency = Latency{0, 0, &sync.Mutex{}}

	fmt.Println("Game Started on port 7777")
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.Handle("/socket", websocket.Handler(handleSocketIn))
	http.ListenAndServe(":7777", nil)
}

// requires that game is locked
func isDrawer(c *Client) bool {
	return c == game.clients[game.drawerIndex]
}

// requires that game is locked
func getBoard() []Point {
	return game.board;
}

func getLeaderboard() map[string]int {
	return leaderboard
}

// requires that game is locked
// sends the packet to all channels, modifying the
// packet to set IsDrawer to true for the drawer
func updateAllChan(packet Packet) {
	updateNonDrawer(packet)
	packet.IsDrawer = true
	packet.Delay = latency.maxDelay - game.clients[game.drawerIndex].delay
	game.clients[game.drawerIndex].output <- packet
}

// requires that game is locked
// sends the packet to all channels,
// except that of the drawer
func updateNonDrawer(packet Packet) {
	for i := 0; i < len(game.clients); i++ {
		if (i != game.drawerIndex) {
			packet.Delay = latency.maxDelay - game.clients[i].delay
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

	pkt := Packet{Ptype: "init",
		Board: getBoard(),
		Leaderboard: getLeaderboard(),
		IsDrawer: isDrawer}

	if !isDrawer {
		for game.clients[game.drawerIndex].name == "" {
		}
		pkt.Data = game.clients[game.drawerIndex].name
	}

	websocket.JSON.Send(ws, pkt)

	newClient := &Client{_id: game.maxId,
		ws: ws,
		name: "",
		output: make(chan Packet, CHAN_SIZE),
		delayHistory: make([]int64, 0),
		delay: 0}

	// increment maxId
	game.maxId++

	game.clients = append(game.clients, newClient)
	return newClient;
}

func handleSocket(currClient * Client) {
	input := make(chan Packet, CHAN_SIZE)
	go func() {
		var packet Packet
		for {
			err := websocket.JSON.Receive(currClient.ws, &packet)

			if err != nil {
				fmt.Println("Debug: websocket is closed, err:", err)
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
			case "name":
				handleName(currClient, packet)
			case "ack":
				handleAck(currClient, packet)
			case "guess":
				handleGuess(currClient, packet)
			case "word":
				handleWordChange(currClient, packet)
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

func handleName(currClient * Client, packetIn Packet) {
	currClient.name = packetIn.Data
	_, inMap := leaderboard[currClient.name]

	if !inMap {
		leaderboard[currClient.name] = 0
	}

	packetOut := Packet{Ptype: "leaderboard",
		Board: nil,
		Leaderboard: getLeaderboard(),
		IsDrawer: isDrawer(currClient)}

	websocket.JSON.Send(currClient.ws, packetOut)
	// TODO: safetey checks
}

func handleAck(currClient * Client, packetIn Packet) {
	// calculate this packet's delay
	var delay int64
	delay = time.Now().UnixNano() / int64(time.Millisecond) - packetIn.Date

	// find this client's new average delay
	currClient.delayHistory = append(currClient.delayHistory, delay)
	// remove old delay information
	if (len(currClient.delayHistory) > DELAY_HISTORY_MAX) {
		currClient.delayHistory = append(currClient.delayHistory[1:])
	}
	var totalDelay int64
	totalDelay = 0
	for i := 0; i < len(currClient.delayHistory); i++ {
		totalDelay += currClient.delayHistory[i]
	}

	currClient.delay = totalDelay / int64(len(currClient.delayHistory))

	latency.Lock()
	defer latency.Unlock()

	// every LATENCY_COUNTER_MAX acks recieved, reset the latency
	// counter incase a highly delayed player quit
	latency.counter = (latency.counter + 1) % LATENCY_COUNTER_MAX
	if latency.counter == 0 {
		latency.maxDelay = 0
	}

	// if we're the slowest client, update the latency record
	if latency.maxDelay < currClient.delay {
		latency.maxDelay = currClient.delay
	}

	fmt.Println("Debug: ack with delay:", delay,
		"my avg delay:", currClient.delay,
		"worst delay", latency.maxDelay)
}

// If the guess is correct, update the guesser and
// alert all clients of a change in word, otherwise,
// do nothing
func handleGuess(currClient * Client, packetIn Packet) {
	game.Lock()
	defer game.Unlock()

	for !game.wordValid {
	}

	if isDrawer(currClient) {
		fmt.Println("Debug: a drawer tried to guess")
		return
	}

	fmt.Println("Debug: guesser guessing", packetIn.Data, "actual", game.word)

	if game.word == packetIn.Data {
		game.wordValid = false;
		game.board = make([]Point, 0)
		leaderboard[currClient.name] = leaderboard[currClient.name] + 1
		packetOut := Packet{Ptype: "next",
					Board: nil,
					IsDrawer: false,
					Leaderboard: getLeaderboard(),
					Data: currClient.name}

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

func handleWordChange(currClient * Client, packetIn Packet) {
	if !isDrawer(currClient) {
		return
	}

	game.word = packetIn.Data
	game.wordValid = true
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
		IsDrawer: false}

	updateNonDrawer(packetOut)

	for i := 0; i < len(packetIn.Board); i++ {
		game.board = append(game.board, packetIn.Board[i])
	}
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
		IsDrawer: false}

	game.board = make([]Point, 0)

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

	delete(leaderboard, currClient.name)

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
		game.wordValid = false
		game.board = make([]Point, 0)
		game.clients = make([]*Client, 0)
		return
	} else if isDrawer {
		// the drawer just quit, clear current game state
		game.wordValid = false
		game.board = make([]Point, 0)

		// otherwise, randomly assign a new drawer and
		// set up a new round
		game.drawerIndex = rand.Intn(len(game.clients))

		packetOut = Packet{Ptype: "drawerQuit",
					Board: nil,
					IsDrawer: false,
					Leaderboard: getLeaderboard(),
					Data: game.clients[game.drawerIndex].name}
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
					IsDrawer: false,
					Leaderboard: getLeaderboard(),
					Data: currClient.name}
	}

	updateAllChan(packetOut)
}
