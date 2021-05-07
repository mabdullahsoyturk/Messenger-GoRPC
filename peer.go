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

var peerIDs []string
var peerConnections []*rpc.Client
var myID, myPort, myIP string
var senderSequenceNumber int = 0

type Message struct {
	Transcript string // The text that the user of the sender process has entered.
	SID        string // Identifier of the original sender process
	TSM        int    // Sender sequence number of the message
}

type MessengerAPI int

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// Modified version of https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go/16615559#16615559
func readPeers(filename string, givenId string) {
	file, err := os.Open(filename)
	checkError(err)
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
			peerIDs = append(peerIDs, peerIP+":"+peerPort)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (m *MessengerAPI) MessagePost(message *Message, reply *int) error {
	fmt.Println((*message).SID + ":\t" + (*message).Transcript + "\t(" + strconv.Itoa((*message).TSM) + ")") // Print received message
	*reply = 1
	return nil
}

func connectMessenger() {
	i := 0
	for {
		client, err := rpc.Dial("tcp", peerIDs[i]) // Try to connect to the ith peer
		if err == nil {
			fmt.Println("Connected to peer:", peerIDs[i])
			peerIDs = append(peerIDs[:i], peerIDs[i+1:]...) // Remove the connected peer from peers (https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang)
			peerConnections = append(peerConnections, client)
		} else {
			fmt.Println("Couldn't connect" + peerIDs[i] + " trying next peer")
		}

		if len(peerIDs) > 0 {
			i = (i + 1) % len(peerIDs) // if there is a next peer, try to connect it
		} else {
			break // We are now connected to all pairs.
		}

		time.Sleep(1000 * time.Millisecond) // Try every second
	}
	fmt.Println("Connected to all peers")

	scanner := bufio.NewScanner(os.Stdin)
	var reply int
	for {
		scanner.Scan()
		input := scanner.Text()

		senderSequenceNumber++
		msg := Message{input, myID, senderSequenceNumber}                                   // Create the message
		fmt.Println(msg.SID + ":\t" + msg.Transcript + "\t(" + strconv.Itoa(msg.TSM) + ")") // Print my message to the console

		for _, client := range peerConnections { // Multicast my message
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

	readPeers("group.txt", givenId) // Read Peer Info

	api := new(MessengerAPI) // Create rpc API
	rpc.Register(api)

	listener, err := net.Listen("tcp", myID) // Start the server
	checkError(err)
	fmt.Println("> Service is running!")

	go connectMessenger() // // Connect to Peers

	for {
		conn, err := listener.Accept() // Wait for connections
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
