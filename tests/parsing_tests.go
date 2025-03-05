package tests

import (
	"distributed-calc/internal/parsing"
	"fmt"
	"testing"
)

func Evaluate(node *parsing.Node) (float64, error) {
	if node.IsLeaf {
		return node.Value, nil
	}
	left, err := Evaluate(node.Left)
	if err != nil {
		return 0, err
	}
	right, err := Evaluate(node.Right)
	if err != nil {
		return 0, err
	}
	switch node.Operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", node.Operator)
	}
}

func TestParsingValid(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"7+2+9", 18},
		{"(9-8)*3", 3},
		{"14/1-1", 13},
		{"(14/1)-1", 13},
		{"20+5*4", 40},
		{"-9", -9},
	}
	for _, tc := range tests {
		ast, err := parsing.ParseExpression(tc.expr)
		if err != nil {
			t.Errorf("Error: %s. %v", tc.expr, err)
			continue
		}
		result, err := Evaluate(ast)
		if err != nil {
			t.Errorf("Error. Expression: %s. Message: %v", tc.expr, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("Wrong results. Expression: %s. Expected %f but got %f", tc.expr, tc.expected, result)
		}
	}
}

func TestParsingInvalid(t *testing.T) {
	expressions := []string{
		"",
		"17+",
		"--2",
		"1222++8888",
		"abc",
	}
	for _, expr := range expressions {
		_, err := parsing.ParseExpression(expr)
		if err == nil {
			t.Errorf("Expression %s is invalid but no error", expr)
		}
	}
}