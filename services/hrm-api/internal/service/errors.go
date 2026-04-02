package service

import "errors"

var (
	ErrEmployeeEmailExists = errors.New("employee with this email already exists")
	ErrInvalidSalary       = errors.New("base salary must be greater than 0")
	ErrInvalidRole         = errors.New("invalid employee role")
)
