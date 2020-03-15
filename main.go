package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zhuangsirui/binpacker"
)

// http://192.168.4.1/build/bundle.js
//var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "192.168.4.1:80", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				// return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	var i uint16 = 0

	for {
		select {
		case <-done:
			return
		case <-ticker.C:

			// fmt.Printf("Sending... %d\n", t.Unix())

			buffer := new(bytes.Buffer)
			packer := binpacker.NewPacker(binary.LittleEndian, buffer)
			packer.PushUint16(i)
			packer.PushUint8(99)  // c
			packer.PushUint8(124) // |
			// Bytes([]byte{0x02, 0x03})
			packer.PushString(fmt.Sprintf("id: %d", i))
			fmt.Printf("Sending... %s\n", buffer.Bytes())

			err = c.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("write:", err)
				// return
			}
			i++
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
