package service

import "errors"

var (
	errCustomerEmailExists = errors.New("customer with this email already exists")
	errInvalidPoints       = errors.New("loyalty points must be greater than 0")
)
