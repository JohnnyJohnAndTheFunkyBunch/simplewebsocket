// Copyright 2013 Jonathan Ma. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"net"
	"strings"
)

// Close codes
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseTLSHandshake            = 1015
)

// The message types are defined in RFC 6455, section 11.8.
const (
	TextFrame   = 1
	BinaryFrame = 2
	CloseFrame  = 8
	PingFrame   = 9
	PongFrame   = 10
)

var keyGUID []byte = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func initHandshake(s *Server, conn net.Conn) error {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		s.logger.Print("Error reading:", err.Error())
	}
	request := string(buf)
	key := getKey(request)
	// if not successful reutnr an error
	h := sha1.New()
	h.Write(key)
	h.Write(keyGUID)
	reponseKey := base64.StdEncoding.EncodeToString(h.Sum(nil))
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + reponseKey + "\r\n\r\n"
	conn.Write([]byte(response))
	return nil
}

func decodeFrames(frames []byte) ([]byte, error) {
	var payloadStart uint64 = 2
	if len(frames) < 3 {
		return nil, errors.New("decodeFrames Error: No payload in frames")
	}
	//b := frames[0]
	//fin := b & 0x80
	//opcode := b & 0x0f
	b2 := frames[1]
	mask := b2 & 0x80
	var length uint64
	lengthByte := b2 & 0x7f // low 7 bits

	if len(frames) < int(payloadStart)+4 {
		return nil, errors.New("decodeFrames Error: Frame length does not match payload length")
	} else if lengthByte < 126 {
		length = uint64(lengthByte)
	} else if lengthByte == 126 {
		length = uint64(binary.BigEndian.Uint16(frames[2:4]))
		payloadStart += 2
	} else if lengthByte == 127 {
		length = uint64(binary.BigEndian.Uint64(frames[2:10]))
		payloadStart += 8
	}

	maskBytes := make([]byte, 4)
	if mask != 0 {
		copy(maskBytes, frames[payloadStart:payloadStart+4])
		payloadStart += 4
	}

	if len(frames) < int(payloadStart)+int(length) {
		return nil, errors.New("decodeFrames Error: Frame length does not match payload length")
	}
	payload := frames[payloadStart : payloadStart+length]
	// unmask
	if mask != 0 {
		for i, b := range payload {
			payload[i] = b ^ maskBytes[i%4]
		}
	}
	maskBytes = nil
	return payload, nil
}

func encodeFrames(msg []byte) []byte {
	b1 := byte(0x80)
	b2 := byte(0x0)
	// can maybe be better if set a cap
	var buf []byte

	// send as text
	b1 = b1 | 0x1
	buf = append(buf, b1)
	length := uint64(len(msg))
	if length < 126 {
		b2 = b2 | byte(length)
		buf = append(buf, b2)
	} else if length < 65536 {
		b2 = b2 | 126
		buf = append(buf, b2)
		l := make([]byte, 2)
		binary.BigEndian.PutUint16(l, uint16(length))
		buf = append(buf, l...)
	} else {
		b2 = b2 | 127
		l := make([]byte, 8)
		binary.BigEndian.PutUint64(l, uint64(length))
		buf = append(buf, l...)
	}
	buf = append(buf, msg...)
	return buf
}

func getKey(resp string) []byte {
	needle := "Sec-WebSocket-Key"
	start := strings.Index(resp, needle)
	end := strings.Index(resp[start:], "\n") + start
	key := strings.Trim(resp[start+len(needle):end], " :\n\r")
	return []byte(key)
}
