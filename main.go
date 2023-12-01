package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	host     = "192.168.0.103"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

func setupDatabase() *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get generic database object:", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to database")
	return db
}

func listenForNotifications(connStr string) {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, reportProblem)
	err := listener.Listen("inv_channel")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start listening for PostgreSQL notifications on 'inv_channel'")
	for {
		select {
		case n := <-listener.Notify:
			u := url.URL{Scheme: "ws", Host: "192.168.0.103:8090", Path: "/api/inventory/notification"}
			log.Printf("connecting to %s", u.String())

			c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Println("dial:", err)
				return
			}
			defer c.Close()

			// Prepare the message to send
			msg := n.Extra

			// Send the message
			err = c.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("write:", err)
				return
			}
			fmt.Printf("Received notification: %s\n", n.Extra)
		case <-time.After(90 * time.Second):
			listener.Ping()
		}
	}
}

func main() {

	// connect websocket

	db := setupDatabase()
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)
	go listenForNotifications(connStr)

	// Prevent the main function from exiting
	select {}

}
