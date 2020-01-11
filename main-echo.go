package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/olahol/melody.v1"
	"log"
	"net/http"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"unsafe"
)

func main() {
	e := echo.New()
	m := melody.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		http.ServeFile(c.Response().Writer, c.Request(), "index.html")
		return nil
	})

	e.GET("/channel/:name", func(c echo.Context) error {
		http.ServeFile(c.Response().Writer, c.Request(), "channel.html")
		return nil
	})

	e.GET("/ws/channel/:name", func(c echo.Context) error {
		m.HandleRequest(c.Response().Writer, c.Request())
		return nil
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		log.Println("HandleMessage")
		go SaveMessage(msg)
		m.BroadcastFilter(msg, func(q *melody.Session) bool {
			log.Println("BroadcastFilter")
			return q.Request.URL.Path == s.Request.URL.Path
		})
	})
	// 新たにwebsocketを通して送られてきたデータをどのように処理するかを記述する。
	m.HandleConnect(func(s *melody.Session) {
		str := `{"name":"運営","message":"user login"}`
		byte := *(*[]byte)(unsafe.Pointer(&str))
		m.BroadcastFilter(byte, func(q *melody.Session) bool {
			return q.Request.URL.Path == s.Request.URL.Path
		})
	})

	// なんらかの原因でwebsocketが切れたときに生じる処理を記述。
	m.HandleDisconnect(func(s *melody.Session) {
		str := `{"name":"運営","message":"user logout"}`
		byte := *(*[]byte)(unsafe.Pointer(&str))
		m.BroadcastFilter(byte, func(q *melody.Session) bool {
			return q.Request.URL.Path == s.Request.URL.Path
		})
	})
	Migrate()
	e.Logger.Fatal(e.Start(":8080"))
}

type Chat struct {
	gorm.Model
	ChatRoomId string `json: chatRoomId`
	UserId uint `json: userId`
	Message string `json: message`
}

/**
 * gormを使って、MySqlと接続する
 * @see http://gorm.io/ja_JP/docs
 */
func ConnectMySql() *gorm.DB {
	db, err := gorm.Open("mysql", "root:@(localhost)/chat_example?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Println("データベースへの接続に失敗しました")
		log.Fatal(err)
		panic("データベースへの接続に失敗しました")
	}
	return db
}

func Migrate() (*gorm.DB, error) {
	conn := ConnectMySql()
	defer conn.Close()

	conn.AutoMigrate(&Chat{})
	log.Println("Migration has been processed")
	return conn, nil
}

func SaveMessage (chat []byte) {
	var chatJson Chat
	// []byte => Chat型にmapping
	json.Unmarshal(chat, &chatJson)
	conn := ConnectMySql()
	defer conn.Close()
	conn.Save( &Chat{ChatRoomId: chatJson.ChatRoomId, UserId: chatJson.UserId, Message: chatJson.Message})
}