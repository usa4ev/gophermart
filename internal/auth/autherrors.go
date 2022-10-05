package auth

import "fmt"

var (
	ErrUserAlreadyExists = fmt.Errorf("user already exists")
	ErrUnathorized       = fmt.Errorf("wrong login or password")
)
