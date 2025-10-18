package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	ID          string
	Payload     any
	Result      any
	Error       error
	Retries     int
	MaxRetries  int
	CreatedAt   time.Time
	CompletedAt *time.Time
}

func main() {
	newJob := &Job{
		ID:          "klasjkalsdj",
		Payload:     nil,
		Result:      nil,
		Error:       nil,
		Retries:     4,
		MaxRetries:  5,
		CreatedAt:   time.Now(),
		CompletedAt: new(time.Time),
	}

	bytes, err := json.Marshal(newJob)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("json marshal", string(bytes))
}
