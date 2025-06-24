package config

import "github.com/go-resty/resty/v2"

func NewResty() *resty.Client {
    client := resty.New()

    // client.SetDebug(true)

    return client
}
