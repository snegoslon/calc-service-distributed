package tests

import (
	"distributed-calc/internal/application"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	orchestrator := application.NewOrchestrator()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orchestrator.CalculationHandler(w, r)
	})

	reqBody := `{"expression": "(11+2)*7"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected: %s, actual %s", http.StatusText(http.StatusCreated), http.StatusText(rr.Code))
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to parse json: %v", err)
	}
	if id, ok := resp["id"]; !ok || id == "" {
		t.Errorf("Unexpected id in response. Response: %v", resp)
	}
}