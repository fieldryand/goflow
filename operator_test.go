package goflow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var engine = operatorTestEngine()

func TestBash(t *testing.T) {
	result, _ := BashOp("sh", "-c", "echo $((2 + 4))").Run()
	resultStr := fmt.Sprintf("%v", result)
	expected := "6\n"

	if resultStr != expected {
		t.Errorf("Expected %s, got %s", expected, resultStr)
	}
}

func TestGetSuccess(t *testing.T) {
	expected := "OK"
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(expected))
		}))
	defer srv.Close()

	result, _ := GetOp(srv.URL).Run()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetNotFound(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
			w.Write([]byte("Page not found"))
		}))
	defer srv.Close()

	_, err := GetOp(srv.URL).Run()

	if err == nil {
		t.Errorf("Expected an error")
	}
}

func operatorTestEngine() *Engine {
	g := NewEngine()
	g.Use(DefaultLogger())
	g.addRoutes()
	return g
}
