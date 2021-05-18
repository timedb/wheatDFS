package app

import "errors"

const (
	MissingDataErr = "MissingData"
)

//pool

var (
	TimeOutErr = errors.New("timeout err")
	OnLinkErr  = errors.New("failed to connect to server") //连接失败
)
