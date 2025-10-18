package model

type Auth struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}
