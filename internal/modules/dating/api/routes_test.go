package api

import (
	"net/http"
	"testing"
)

func TestSetupRoutes(t *testing.T) {
	mux := http.NewServeMux()
	handler := NewHandler(nil, nil)

	SetupRoutes(mux, handler)

	// Just verify that SetupRoutes doesn't panic
	// Actual route testing would require integration tests
}
