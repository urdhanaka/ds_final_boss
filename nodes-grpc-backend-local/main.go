package main

import (
	"nodes-grpc-backend-local/config"
)

func main() {
	app := config.NewGin()
    db := config.NewPsql()
}
