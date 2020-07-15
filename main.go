package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func main() {
	log.SetPrefix("gProxy: ")
	fmt.Println("Starting gProxy...")

	var localAddr string
	var remoteAddr string

	getArgs(&localAddr, &remoteAddr)

	addr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for connections...")

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		go proxyConn(conn, remoteAddr)
	}
}

func proxyConn(conn *net.TCPConn, remoteAddr string) {
	fmt.Println("New connection from", conn.RemoteAddr())

	defer conn.Close()

	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		panic(err)
	}

	defer rConn.Close()

	go func() {
		log.Println("waiting for data from client")
		data := make([]byte, 1024)
		for {
			n, err := conn.Read(data)
			if err != nil {
				break
			}
			_, err = rConn.Write(data[:n])
			if err != nil {
				break
			}
			log.Printf("sent:\n%v\nend of send\n", hex.Dump(data[:n]))
		}
	}()

	data := make([]byte, 1024)
	for {
		n, err := rConn.Read(data)
		if err != nil {
			break
		}
		_, err = conn.Write(data[:n])
		if err != nil {
			break
		}
		log.Printf("recv:\n%v\nend of recv\n", hex.Dump(data[:n]))
	}

	fmt.Println("Closed connection from", conn.RemoteAddr())
}

func getArgs(localAddr *string, remoteAddr *string) {
	flag.StringVar(localAddr, "l", "0.0.0.0:9999", "local address")
	flag.StringVar(remoteAddr, "r", "", "remote address")
	var dbg = flag.Bool("d", false, "debug info")

	flag.Parse()

	if *remoteAddr == "" {
		usage()
		os.Exit(1)
	}

	if *dbg == false {
		log.SetOutput(ioutil.Discard)
	}
}

func usage() {
	fmt.Println("./proxy -r [remote addr] (-d)")
}
