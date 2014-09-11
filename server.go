// Copyright 2013 Jonathan Ma. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"log"
	"net"
	"os"
	"time"
)

// A WebSocket Server
type Server struct {
	port               string
	host               string
	handleConnected    func(*Conn)
	handleDisconnected func(*Conn)
	handleTextMsg      func(*Conn, string)
	handleBinMsg       func(*Conn, []byte)
	handlePong         func(*Conn)
	connSet            map[*Conn]bool
	logger             *log.Logger
}

// Start the listen loop and will create a goroutine for each connection
func (s *Server) ListenAndServe() {
	s.logger.Print("Listenening on ", s.host, ":", s.port)
	l, err := net.Listen("tcp", s.host+":"+s.port)
	if err != nil {
        if s.logger != nil {
            s.logger.Print("Error listening", err.Error())
        }
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
            if s.logger != nil {
                s.logger.Print("Error accepting: ", err.Error())
            }
			os.Exit(1)
		}
		go connHandler(s, conn)
	}
}

// Create server and configures it
func CreateServer(host string, port string, logger *log.Logger) Server {
	m := make(map[*Conn]bool)
	return Server{
		port:    port,
		host:    host,
		connSet: m,
		logger:  logger,
	}
}

// Function callback for new connections
func (s *Server) HandleConnected(handler func(*Conn)) {
	s.handleConnected = handler
}

// Function callback for disconnects
func (s *Server) HandleDisconnected(handler func(*Conn)) {
	s.handleDisconnected = handler
}

// Function callback for handling text type messages
func (s *Server) HandleTextMsg(handler func(*Conn, string)) {
	s.handleTextMsg = handler
}

// Function callback for handling binary type messages
func (s *Server) HandleBinMsg(handler func(*Conn, []byte)) {
	s.handleBinMsg = handler
}

// Function callback for handling pong messages
func (s *Server) HandlePong(handler func(*Conn)) {
	s.handlePong = handler
}

func connHandler(s *Server, conn net.Conn) {
	err := initHandshake(s, conn)
	if err != nil {
		return
	}
	newConn := Conn{
		connection: conn,
		latency:    0,
		server:     s,
		pingTime:   make(chan time.Time),
	}
	if s.handleConnected != nil {
		s.handleConnected(&newConn)
	}
	go heartbeatLoop(s, &newConn)
	readLoop(s, &newConn)
}
