package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPath(t *testing.T) {
	PathSeparator = "\\"
	req, err := http.NewRequest("GET", "/ls/", nil)

	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("path", "")
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnDirectoryListingAtPath)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheck)
	handler.ServeHTTP(rr, req)

	checkResponseCode(t, http.StatusOK, rr.Code)

	var response HealthResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	expected := "OK"
	if response.Status != expected {
		t.Errorf("Health endpoint returned an unexpected body or status: got %v want %v", rr.Body.String(), expected)
	}
}

func TestPathNotFound(t *testing.T) {
	PathSeparator = "\\"
	req, err := http.NewRequest("GET", "/ls/", nil)

	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("path", "asdjmlk123l1n34lknasd")
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(returnDirectoryListingAtPath)
	handler.ServeHTTP(rr, req)

	//if status := rr.Code; status != http.StatusBadRequest {
	//	t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	//}

	checkResponseCode(t, http.StatusBadRequest, rr.Code)

	expected := "Specified path not found"
	if rr.Body.String() != expected {
		t.Errorf("Health endpoint returned an unexpected body or status: got %v want %v", rr.Body.String(), expected)
	}
}

func TestErroneusRoute(t *testing.T) {
	req, err := http.NewRequest("GET", "/unknown", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(unexpectedRoute)
	handler.ServeHTTP(rr, req)

	checkResponseCode(t, http.StatusNotFound, rr.Code)

	expected := "Invalid, unknown, or unauthorized API call"
	if rr.Body.String() != expected {
		t.Errorf("Health endpoint returned an unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected HTTP response code %d. Got %d\n", expected, actual)
	}
}
