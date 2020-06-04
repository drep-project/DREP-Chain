package app

import "errors"

var (
	// ErrNotMatchedService print error msg
	ErrNotMatchedService = errors.New("the service added not match service interface")
	// ErrConfigiNotFound print error msg
	ErrConfigiNotFound = errors.New("specify config file not exist")
	// ErrServiceNotFound print error msg
	ErrServiceNotFound = errors.New("Service not found")
)
