package fileKeyTorch

import "errors"

var (
	MaxFileSizeErr   = errors.New("this interface only handles large files")
	SmallFileSizeErr = errors.New("large files cannot use this interface")
	OffsetErr        = errors.New("the offset cannot be considered negative")
)
