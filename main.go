package main

import (
	"github.com/gin-gonic/gin"
	melody "gopkg.in/olahol/melody.v1"
	"log"
	"net/http"
)

func main()  {
	// WEbアプリフレームワーク
	r := gin.Default()
	// Websocketを簡単に扱うことのできるパッケージ
	// melodyオブジェクトを生成している。
	m := melody.New()
	// byte型... a.k.a unit8
	// unit8 ... 1byte (8bit)  (0 ~ 255)
	 // golang  では、[]byteを使うことがおおい。string型よりもprimitiveなデータ型でmutable, subsetの置換などの取り扱いがしやすい。
	m.HandleMessage(func(s *melody.Session, msg []byte){
		// HandleMessage() WebSocketを通して、送られてきたデータをどのように処理するかを記述する。
		// 今回はBroadcast()することで接続されているデバイス全てにそのまま転送している。
		m.Broadcast(msg)
	})

	// 新たにwebsocketを通して送られてきたデータをどのように処理するかを記述する。
	m.HandleConnect(func (s *melody.Session) {
		log.Printf("websocket connection open. [session: %#v]\n", s)
	})

	// なんらかの原因でwebsocketが切れたときに生じる処理を記述。
	m.HandleDisconnect(func (s *melody.Session){
		log.Printf("websocket connection close. [session: %#v]\n", s)
	})

	r.GET("/", indexHandler)
	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})
	r.Static("/static", "/static")
	r.Run(":8080")
}

func indexHandler(c *gin.Context) {
	http.ServeFile(c.Writer, c.Request, "index.html")
}