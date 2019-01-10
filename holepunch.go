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
	}

	return
}

func (h *Holepunch) Connect() (err error) {
	//immediately start listening
	listenLoop(h.wg, h.laddr)

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
	readLoop(h.wg, h.reader, h.stopReqChan, h.initHolepunchChan)
	connRequestLoop(h.wg, h.writer, h.connRequestChan, h.stopReqChan)
	initHolepunch(h.wg, laddr, h.initHolepunchChan)

	h.wg.Wait()
	return
}
