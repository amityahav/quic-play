package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"io"
	"log"
	"time"
)

type loggingWriter struct{ s quic.Stream }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Stream %d: Got '%s'\n", w.s.StreamID(), string(b))
	return 0, nil
}

func main() {
	crt, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	tlsConf := tls.Config{
		Certificates:       []tls.Certificate{crt},
		InsecureSkipVerify: true,
	}

	ln, err := quic.ListenAddr("localhost:1234", &tlsConf, nil)

	log.Println("starting server loop...")
	go serverLoop(ln)

	time.Sleep(1 * time.Second)

	// client
	cConn, err := quic.DialAddr(context.Background(), "localhost:1234", &tlsConf, nil)
	if err != nil {
		log.Fatal(err)
	}

	// open stream1
	s1, err := cConn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	_, err = s1.Write([]byte("Hello world1"))
	if err != nil {
		log.Fatal(err)
	}

	// open stream2
	s2, err := cConn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	_, err = s2.Write([]byte("Hello world2"))
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(5 * time.Second)
}

func serverLoop(ln *quic.Listener) {
	for {
		conn, err := ln.Accept(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn)
	}
}

func handleConn(conn quic.Connection) {
	log.Println("handling connection")
	for {
		s, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		go handleStream(s)
	}
}

func handleStream(s quic.Stream) {
	w := loggingWriter{s}
	log.Printf("handling stream %d\n", s.StreamID())
	io.Copy(w, s)
}
