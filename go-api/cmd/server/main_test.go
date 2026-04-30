package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidatePlayerInput(t *testing.T) {
	input := playerInput{Name: " Player One ", FavoriteMove: "Rock", Rating: 1200}
	if err := validatePlayerInput(&input); err != nil {
		t.Fatalf("expected input to be valid: %v", err)
	}
	if input.Name != "Player One" {
		t.Fatalf("expected trimmed name, got %q", input.Name)
	}
	if input.FavoriteMove != "rock" {
		t.Fatalf("expected normalized favorite move, got %q", input.FavoriteMove)
	}
}

func TestValidatePlayerInputRejectsInvalidMove(t *testing.T) {
	input := playerInput{Name: "Player One", FavoriteMove: "lizard", Rating: 1200}
	if err := validatePlayerInput(&input); err == nil {
		t.Fatal("expected invalid favorite move to fail")
	}
}

func TestParsePlayerID(t *testing.T) {
	id, err := parsePlayerID("/api/players/42")
	if err != nil {
		t.Fatalf("expected valid id: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
}

func TestDocsAreBrowsable(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/docs", nil)
	response := httptest.NewRecorder()

	docs(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if body := response.Body.String(); body == "" || !strings.Contains(body, "Players CRUD API") {
		t.Fatalf("expected docs page to contain API title, got %q", body)
	}
}
