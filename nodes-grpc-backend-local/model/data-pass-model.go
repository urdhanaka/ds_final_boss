package model

type DataPass struct {
	Data    any
	Message string
	Error   error
	Code    int
	IsError bool
}
