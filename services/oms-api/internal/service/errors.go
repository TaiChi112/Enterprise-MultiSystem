package service

import "errors"

var (
	errInvalidStatus   = errors.New("invalid order status")
	errInvalidItemData = errors.New("item quantity and unit price must be greater than 0")
)
