// Copyright 2014, Jonathan Ma. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package websocket implements the WebSocket protocol defined in RFC 6455.
//
// Overview
//
// The Conn type represents a WebSocket connection. The Server type represents
// a WebSocket server with configs that the user can set. This library
// implements in the style of using callbacks to handle incoming connections
// and messages. Here is and example on how you can use the library:
//
//  // optional logger if you want to see debug messages
//	logger := log.New(os.Stdout, "websocket: ", log.Ltime)
//	ws := websocket.CreateServer("localhost", "8000", logger)
//  // you can put your own user defined function to handle these events
//	ws.HandleConnected(onConnected)
//	ws.HandleDisconnected(onDisconnected)
//	ws.HandleTextMsg(onMsg)
//	ws.HandlePong(onPong)
//  // start a new go routine in a listen loop
//	go ws.ListenAndServe()
//
// Call the connection SendTextMsg and SendBinMsg methods to send  messages
// as a string or slice of bytes. This example code will shows how to send
// messages using these methods:
//
//  binMsg := []byte{1,2,3,4,5}
//  conn.SendBinMsg(binMsg)
//  conn.SendTextMsg("Hello!")
//
// Notes: Every 10 seconds the server will send a ping to the connection
// to see if it's still alive.
package websocket
