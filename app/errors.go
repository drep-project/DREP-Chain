package app

import "errors"

var (
	ErrNotMatchedService = errors.New("the service added not match service interface")
	ErrConfigiNotFound   = errors.New("specify config file not exist")
	ErrServiceNotFound   = errors.New("Service not found")
)
