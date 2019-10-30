package broker

import (
	"errors"
)

var (
	ErrWrongNode     = errors.New("Trying to subscribe to wrong edge node")
	ErrNotRegistered = errors.New("Edge node not registered")
	ErrRegistered    = errors.New("Edge node already registered")
	ErrUnsubscribe   = errors.New("Unsubscribe failed")
	ErrLoadCert      = errors.New("Certificate loading error")
	ErrConnect       = errors.New("Could not connect to server")
)
