package main

import (
	"log"
	"log/slog"
	"nodes-grpc-be/config"
	"os"
	"os/signal"
	"time"
)

func main() {
	ginInstance := config.NewGin()
	psqlConn := config.NewPsqlConnection()
	redisClient := config.NewRedisConnection()
    defer psqlConn.Close()

	setRouters(ginInstance, psqlConn, redisClient)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		_ = <-sigs

		time.Sleep(1 * time.Second)

		// WARN: for development purpose only
		{
			slog.Info("dropping the table...")
			config.DropTable()
		}

		os.Exit(0)
	}()

	log.Fatal(ginInstance.Run(":8000"))
}
