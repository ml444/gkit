package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	gws "github.com/gorilla/websocket"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/websocket", websocket.Handler(echoHandler))
	http.HandleFunc("/web", web)
	log.Println("===>starting ws")

	// 将HTTP服务器监听在端口8000上，准备接收WebSocket连接
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		panic("fail" + err.Error())
	}
}

func web(w http.ResponseWriter, r *http.Request) {
	fmt.Println("===> haha")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, nil)
	}
}

func echoHandler(ws *websocket.Conn) {
	log.Printf("==>time: %d \n", time.Now().Unix())
	defer func() {
		log.Println("exist ws!!!")
		ws.Close()
	}()
	for {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Received: %s\n", msg)
		if err := websocket.Message.Send(ws, msg); err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("Sent: %s\n", msg)
	}
}

func main0() {
	http.HandleFunc("/socket", socketHandler)
	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    ":3001",
		Handler: h2c.NewHandler(http.DefaultServeMux, h2s),
	}
	log.Println("===>starting ws")
	log.Fatal(h1s.ListenAndServe())
}

var upgrader = gws.Upgrader{}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		log.Printf("Received: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error during message writing:", err)
			break
		}
	}
}
