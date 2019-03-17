package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeHealthCheckHandler(t *testing.T) {
	s := &AgentServer{}

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
			t.Fatal(err)
	}

	// creates a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.makeHealthCheckHandler())

	// asserts the service is ready
	s.unlockService()

	// calls the ServeHTTP method and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the expected status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf(
			"handler returned wrong status code: got %v but want %v",
			status, http.StatusOK)
	}

	// Check the expected response body
	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf(
			"handler returned unexpected body: got %v but want %v",
			rr.Body.String(), expected)
	}
}

func executeRequest(s *AgentServer, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.httpServeMux.ServeHTTP(rr, req)
	return rr
}
