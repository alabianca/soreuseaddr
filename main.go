package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error ", err)
		}

		fmt.Printf("[ %s ]: %s", time.Now().Format("03:04PM"), msg)
	}
}

func main() {
	var port = flag.String("port", "4000", "The local port used to holepunch")
	var peer = flag.String("peer", "", "The peer you want to establish a p2p connnection with")
	var uuid = flag.String("uuid", "", "your unique id (could be anything) for other peers to ask for you")

	flag.Parse()

	if *uuid == "" {
		fmt.Println("uuid is required")
		flag.Usage()
		os.Exit(1)
	}

	myIp, err := getMyIpv4Addr()

	if err != nil {
		fmt.Println("Could not find my ipv4 address")
		os.Exit(1)
	}

	config := Config{
		relayIP: "127.0.0.1",
		// relayIP:      "159.89.152.225",
		relayPort:    "8080",
		listenAddr:   myIp.String(),
		localPort:    *port,
		localRelayIP: myIp.String(),
		uID:          *uuid,
	}

	holepunch, err := NewHolepunch(config)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *peer != "" {
		holepunch.connRequestChan <- *peer
	}

	fmt.Println("Connecting ...")
	p2pConn, err := holepunch.Connect()

	if err != nil {
		fmt.Println("could not establish p2p connection")
		os.Exit(1)
	}

	go handleConnection(p2pConn)

	fmt.Printf("P2P Connection is established.... You can start chatting now without a central relay server\n\n")

	writer := bufio.NewWriter(p2pConn)
	in := bufio.NewReader(os.Stdin)
	for {

		msg, err := in.ReadString('\n')
		if err != nil {
			fmt.Println("err ", err)
		} else {
			writer.Write([]byte(msg))
			writer.Flush()
		}
	}

}
