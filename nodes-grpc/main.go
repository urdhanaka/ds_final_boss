package main

import (
	"nodes-grpc/services/db"
)

func main() {
    _ = db.InitDB()
}
