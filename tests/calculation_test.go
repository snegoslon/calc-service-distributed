package tests

import (
	"distributed-calc/pkg/calculation"
	"testing"
)

func TestCompute(t *testing.T) {
	tests := []struct {
		op        string
		a, b      float64
		expected  float64
		shouldErr bool
	}{
		{"+", 3, 3, 6, false},
		{"-", 3, 3, 0, false},
		{"*", 3, 3, 9, false},
		{"/", 3, 3, 1, false},
		{"/", 3, 0, 0, true},
		{"^", 3, 3, 0, true},
	}
	for _, tc := range tests {
		result, err := calculation.Evaluate(tc.op, tc.a, tc.b)
		if tc.shouldErr && err == nil {
			t.Errorf("Error expected! Expression: %s", tc.op)
		}

		if !tc.shouldErr && result != tc.expected {
			t.Errorf("Evaluated(%s, %f, %f) = %f. Expected: %f", tc.op, tc.a, tc.b, result, tc.expected)
		}
	}
}