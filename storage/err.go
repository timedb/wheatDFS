package storage

import "errors"

var (
	HashErr             = errors.New("you mast give right hash")
	TokenEmptyErr       = errors.New("the token must not be empty")
	InvalidErr          = errors.New("invalid token")
	SmallFileSizeErr    = errors.New("this interface only accepts small file transfers")
	MaxFileSizeErr      = errors.New("this interface only accepts max file transfers")
	FileExpiredErr      = errors.New("the file has expired")
	OffsetNotProvideErr = errors.New("a correct offset is not provided")
	InvalidExpiredErr   = errors.New("the file has expired or is invalid")
	StateErr            = errors.New("state err")
	FileKeyErr          = errors.New("make fileKey err")
	UnableUseLogErr     = errors.New("debug is true unable to use log")
	MissDataErr         = errors.New("data missing")
	CreatFkErr          = errors.New("create fk err")
)
