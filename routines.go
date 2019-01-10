package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-reuseport"
)

func listenLoop(wg *sync.WaitGroup, addr *net.TCPAddr) {

	l, err := reuseport.Listen("tcp", addr.IP.String()+":"+strconv.Itoa(addr.Port))
	if err != nil {
		fmt.Println("Could not listen\n", err)
		return
	}
	fmt.Println("Listening for Peers at: ", addr)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting....\nExiting listen loop")
			return
		}

		fmt.Println("We have a connection. Exit listen loop")
	}()

}

func readLoop(wg *sync.WaitGroup, reader *bufio.Reader, stopRequestChan chan int, initHolepunchChan chan peerInfo) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			msg, err := reader.ReadString(ETX)

			if err != nil {
				fmt.Println("Error reading")
				return
			}

			packet := []byte(msg)
			opType := packet[0]

			switch opType {
			case CONN_REQUEST_RESPONSE:
				//check if the relay server knows about the peer
				if _, ok := parseConnRequestResponse(packet); ok {
					stopRequestChan <- 1
				}
			case INIT_HOLEPUNCH:
				initHolepunchChan <- parseInitHolePunchMessage(packet)
			}
		}
	}()
}

func connRequestLoop(wg *sync.WaitGroup, writer *bufio.Writer, connRequestChan chan string, stopRequestChan chan int) {
	wg.Add(1)

	ticker := time.NewTicker(time.Duration(3) * time.Second)
	var req string

	go func() {
		defer wg.Done()

		for {
			select {
			case <-stopRequestChan:
				fmt.Println("Stopping: connectRequestLoop")
				ticker.Stop()
				return
			case request := <-connRequestChan:
				req = request
				writer.Write(createConnRequestPacket(req))
				writer.Flush()

			case <-ticker.C:
				if req != "" {
					writer.Write(createConnRequestPacket(req))
					writer.Flush()
				}
			}
		}
	}()
}

func initHolepunch(wg *sync.WaitGroup, laddr string, initHolepunchChan chan peerInfo) {
	wg.Add(1)

	ticker := time.NewTicker(time.Duration(2) * time.Second)
	var peer *peerInfo

	go func() {
		defer wg.Done()

		for {
			select {
			case peerInf := <-initHolepunchChan:
				peer = &peerInf
				conn, err := reuseport.Dial("tcp", laddr, peer.String())

				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("we got a connection ", conn)
				}

			case <-ticker.C:

				if peer != nil {
					conn, err := reuseport.Dial("tcp", laddr, peer.String())

					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println("we got a connection ", conn)
					}
				}

			}
		}

	}()
}
