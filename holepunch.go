package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/libp2p/go-reuseport"
)

type Holepunch struct {
	raddr             *net.TCPAddr //remote relay server address
	laddr             *net.TCPAddr //listen address for peers
	wg                *sync.WaitGroup
	uID               string
	peer              string
	connRequestChan   chan string
	stopReqChan       chan int
	initHolepunchChan chan peerInfo
	stopListenChan    chan int
	connReadyChan     chan net.Conn
	closeChan         chan int
	reader            *bufio.Reader
	writer            *bufio.Writer
}

func NewHolepunch(config Config) (h *Holepunch, err error) {
	relayAddr := config.relayIP + ":" + config.relayPort

	raddr, err := net.ResolveTCPAddr("tcp", relayAddr)
	laddr, err := net.ResolveTCPAddr("tcp", config.listenAddr+":"+config.localPort)

	h = &Holepunch{
		raddr:             raddr,
		laddr:             laddr,
		wg:                new(sync.WaitGroup),
		uID:               config.uID,
		connRequestChan:   make(chan string, 1),
		stopReqChan:       make(chan int),
		initHolepunchChan: make(chan peerInfo),
		stopListenChan:    make(chan int),
		connReadyChan:     make(chan net.Conn),
		closeChan:         make(chan int),
	}

	return
}

func (h *Holepunch) Connect() (p2pConn net.Conn, err error) {
	//immediately start listening
	listenLoop(h.wg, h.laddr, h.stopListenChan, h.connReadyChan, h.closeChan)

	//connect to relayserver
	laddr := h.laddr.IP.String() + ":" + strconv.Itoa(h.laddr.Port)
	raddr := h.raddr.IP.String() + ":" + strconv.Itoa(h.raddr.Port)
	remoteConn, err := reuseport.Dial("tcp", laddr, raddr)

	if err != nil {
		fmt.Println(err)
		return
	}
	h.reader = bufio.NewReader(remoteConn)
	h.writer = bufio.NewWriter(remoteConn)

	h.writer.Write(createSessionPacket(h.uID, h.laddr.IP.String(), strconv.Itoa(h.laddr.Port)))
	h.writer.Flush()

	//start go routines
	readLoop(h.wg, h.reader, h.stopReqChan, h.initHolepunchChan, h.closeChan)
	connRequestLoop(h.wg, h.writer, h.connRequestChan, h.stopReqChan, h.closeChan)
	initHolepunch(h.wg, laddr, h.initHolepunchChan, h.stopListenChan, h.connReadyChan, h.closeChan)

	p2pConn = <-h.connReadyChan
	fmt.Println("Starting Teardown")
	teardown(h.wg, h.closeChan)
	return
}

func teardown(wg *sync.WaitGroup, closeChan chan int) {
	close(closeChan)

	fmt.Println("Waiting for go routines to exit ...")
	wg.Wait()
	fmt.Println("Exit")
}
