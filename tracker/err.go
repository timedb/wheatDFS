package tracker

import "errors"

var (
	BucketErr   = errors.New("bucket nonexistent")
	HashErr     = errors.New("this hash doesn't exist")
	HostErr     = errors.New("host err")
	RegisterErr = errors.New("this interface can only register Tracker or Storage")
	DebugErr    = errors.New("debug is true unable to use log")
)
