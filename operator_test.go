package goflow

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommand(t *testing.T) {
	result, _ := Command{Cmd: "sh", Args: []string{"-c", "echo $((2 + 4))"}}.Run()
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

	client := &http.Client{}
	result, _ := Get{client, srv.URL}.Run()

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

	client := &http.Client{}
	_, err := Get{client, srv.URL}.Run()

	if err == nil {
		t.Errorf("Expected an error")
	}
}

func TestPostSuccess(t *testing.T) {
	expected := "OK"
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte(expected))
		}))
	defer srv.Close()

	client := &http.Client{}
	result, _ := Post{client, srv.URL, bytes.NewBuffer([]byte(""))}.Run()

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPostNotFound(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
			w.Write([]byte("Page not found"))
		}))
	defer srv.Close()

	client := &http.Client{}
	_, err := Post{client, srv.URL, bytes.NewBuffer([]byte(""))}.Run()

	if err == nil {
		t.Errorf("Expected an error")
	}
}
