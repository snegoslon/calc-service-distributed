package calculation

import "errors"

var (
	ErrDivisionByZero  = errors.New("Division by zero")
	ErrInvalidOperator = errors.New("Invalid operator")
)
