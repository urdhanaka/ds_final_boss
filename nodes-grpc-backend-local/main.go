package main

import (
	"log"
	"nodes-grpc-backend-local/config"
	"nodes-grpc-backend-local/routers"
	"os"
	"os/signal"
	"time"
)

func main() {
	app := config.NewGin()
	db := config.NewPsql()
    redis := config.NewRedisConnection()

	routers.SetRouters(app, db, redis)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		_ = <-sigs

		time.Sleep(1 * time.Second)
        
		config.DropTable()
		os.Exit(0)
	}()

	log.Fatal(app.Run())
}
