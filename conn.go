// Copyright 2013 Jonathan Ma. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"net"
	"time"
)

// A WebSocket connection
type Conn struct {
	connection net.Conn
	latency    time.Duration // in ns
	server     *Server
	pingTime   chan time.Time
	// should have a buffer of fixed capacity, and resuse
}

// Send Message in Text format
func (conn *Conn) SendTextMsg(msg string) {
	msgBytes := encodeFrames([]byte(msg))
	conn.connection.Write(msgBytes)
}

// Send Message in Binary format
func (conn *Conn) SendBinMsg(msg []byte) {
	msgBytes := encodeFrames(msg)
	conn.connection.Write(msgBytes)
}

// Send Ping
func (conn *Conn) SendPing() {
	buf := make([]byte, 2)
	buf[0] = 0x80 | PingFrame
	buf[1] = 0x00
	conn.connection.Write(buf)
}

// Get the latency between server and connection
func (conn *Conn) Latency() time.Duration {
	return conn.latency
}

// Get the underlying TCP connection
func (conn *Conn) Connection() *net.Conn {
	return &conn.connection
}

// Close the WebSocket connection
func (conn *Conn) Close() {
	buf := make([]byte, 1)
	buf[0] = 0x80 | CloseFrame
	conn.connection.Write(buf)
	conn.connection.Close()
	if conn.server.handleDisconnected != nil {
		conn.server.handleDisconnected(conn)
	}
	delete(conn.server.connSet, conn)
}

func heartbeatLoop(s *Server, conn *Conn) {
	for {
		time.Sleep(2000 * time.Millisecond)
		conn.SendPing()
		conn.pingTime <- time.Now() // for the receiver
	}
}

func readLoop(s *Server, conn *Conn) {
	buf := make([]byte, 1024)
	for {
		length, err := conn.connection.Read(buf)
		if err != nil {
            if s.logger != nil {
                s.logger.Print("Error Reading", err.Error())
            }
			conn.connection.Close()
			if s.handleDisconnected != nil {
				s.handleDisconnected(conn)
			}
			return
		}
		if length == 0 {
			return
		}
		typeFrame := buf[0] & 0x0F

		if typeFrame == CloseFrame {
			closeFrame := make([]byte, 1)
			closeFrame[0] = 0x88
			conn.connection.Write(closeFrame)
			conn.connection.Close()
			if s.handleDisconnected != nil {
				s.handleDisconnected(conn)
			}
			return
		}
		if typeFrame == TextFrame {
			msg, err := decodeFrames(buf)
			if err != nil {
                if s.logger != nil {
                    s.logger.Print("Error decoding", err.Error())
                }
				continue
			}
			if s.handleTextMsg != nil {
				s.handleTextMsg(conn, string(msg))
			}
		}
		if typeFrame == BinaryFrame {
			msg, err := decodeFrames(buf)
			if err != nil {
                if s.logger != nil {
                    s.logger.Print("Error decoding", err.Error())
                }
				continue
			}
			if s.handleBinMsg != nil {
				s.handleBinMsg(conn, msg)
			}
		}
		if typeFrame == PingFrame {
			// send pong
			buf := make([]byte, 2)
			buf[0] = 0x80 | PingFrame
			buf[1] = 0x00
			conn.connection.Write(buf)
		}
		if typeFrame == PongFrame {
			conn.latency = time.Since(<-conn.pingTime)
			if s.handlePong != nil {
				s.handlePong(conn)
			}
		}
	}
}
