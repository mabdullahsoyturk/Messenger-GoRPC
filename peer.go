package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var peers []string
var peerClients []*rpc.Client
var myID string
var myPort string
var myIP string
var senderSequenceNumber int = 0

type Message struct {
	Transcript string // The text that the user of the sender process has entered.
	SID        string // Identifier of the original sender process
	TSM        int    // Sender sequence number of the message
}

type MessengerAPI int

func (m *MessengerAPI) MessagePost(message *Message, reply *int) error {
	fmt.Println((*message).SID + ":\t" + (*message).Transcript + "\t(" + strconv.Itoa((*message).TSM) + ")")
	*reply = 1
	return nil
}

func connectMessenger() {
	i := 0
	for {
		client, err := rpc.Dial("tcp", peers[i]) // Try to connect to the ith peer
		if err == nil {
			fmt.Printf("Connected to %s\n", peers[i])
			peers = append(peers[:i], peers[i+1:]...) // Remove the connected peer from peers
			peerClients = append(peerClients, client)
		} else {
			fmt.Println("Couldn't connect" + peers[i] + " trying next peer")
		}

		if len(peers) > 0 {
			i = (i + 1) % len(peers) // if there is a next peer, try to connect it
		} else {
			break // We are now connected to all pairs. 
		}

		time.Sleep(1000 * time.Millisecond) // Try every second
	}
	fmt.Printf("Connected to all peers\n")

	scanner := bufio.NewScanner(os.Stdin)
	var reply int
	for {
		scanner.Scan()
		input := scanner.Text()

		// Create the message
		senderSequenceNumber++
		msg := Message{input, myID, senderSequenceNumber}
		fmt.Println(msg.SID + ":\t" + msg.Transcript + "\t(" + strconv.Itoa(msg.TSM) + ")") // print my own message

		// Multicast the message
		for _, client := range peerClients {
			err := client.Call("MessengerAPI.MessagePost", msg, &reply)
			checkError(err)
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "myId")
		os.Exit(1)
	}
	givenId := os.Args[1]

	readPeers("group.txt", givenId)

	// Create MessengerAPI
	api := new(MessengerAPI)
	rpc.Register(api)

	// Start the server
	listener, err := net.Listen("tcp", myID)
	checkError(err)
	fmt.Println("> Service is running!")

	// Connect to Peers
	go connectMessenger()

	for {
		conn, err := listener.Accept() // Wait for connections
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}

}

// Modified version of https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go/16615559#16615559
func readPeers(filename string, givenId string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerId := scanner.Text()
		if peerId == givenId {
			myIP = strings.Split(peerId, "/")[0]
			myPort = strings.Split(peerId, "/")[1]
			myID = myIP + ":" + myPort
		} else {
			peerIP := strings.Split(peerId, "/")[0]
			peerPort := strings.Split(peerId, "/")[1]
			peers = append(peers, peerIP+":"+peerPort)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
