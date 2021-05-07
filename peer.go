package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

var peers []string
var myId string

type Message struct {
	Transcript string // The text that the user of the sender process has entered.
	SID        string // Identifier of the original sender process
	TSM        string // Sender sequence number of the message
}

type MessengerAPI int

//MessagePost : Exported message post function for the messenger
func (m *MessengerAPI) MessagePost(message *Message, reply *int) error {
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "myId")
		os.Exit(1)
	}
	myId = os.Args[1]

	readPeers("group.txt", myId)
	fmt.Println(peers)
}

// Modified version of https://stackoverflow.com/questions/8757389/reading-a-file-line-by-line-in-go/16615559#16615559
func readPeers(filename string, myId string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerId := scanner.Text()
		if peerId != myId {
			peers = append(peers, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
