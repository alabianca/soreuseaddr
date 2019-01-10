package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func startListen(listen net.Listener, id int) {
	fmt.Println("Listening ...", id)
	go func() {
		_, err := listen.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Got a connection: ", id)
		//listen.Close()
		return
	}()
}

// func controlSetup(network string, address string, c syscall.RawConn) error {
// 	var operr error

// 	fn := func(fd uintptr) {
// 		i := int(fd)
// 		operr = syscall.SetsockoptInt(i, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
// 	}

// 	if err := c.Control(fn); err != nil {
// 		return err
// 	}

// 	if operr != nil {
// 		return operr
// 	}

// 	return nil

// }

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
		relayIP:      "159.89.152.225",
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
	holepunch.Connect()

}
