package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"schedulehelper/db"
	"schedulehelper/models"
	"testing"
)

func setupTestDB(t *testing.T) {
	err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize in-memory DB: %v", err)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	setupTestDB(t)
	defer db.CloseDB()

	// Register
	reqBody := RegisterRequest{
		Username: "testuser",
		Color:    "#ff0000",
		Emoji:    "😀",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(RegisterHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var user models.User
	json.NewDecoder(rr.Body).Decode(&user)

	if user.Username != "testuser" || user.ID == "" {
		t.Errorf("Unexpected user data: %+v", user)
	}

	// Login
	loginBody := LoginRequest{ID: user.ID}
	lBytes, _ := json.Marshal(loginBody)
	req2, _ := http.NewRequest("POST", "/api/login", bytes.NewReader(lBytes))
	rr2 := httptest.NewRecorder()

	loginHandler := http.HandlerFunc(LoginHandler)
	loginHandler.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusOK {
		t.Errorf("login returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var loggedIn models.User
	json.NewDecoder(rr2.Body).Decode(&loggedIn)
	if loggedIn.ID != user.ID {
		t.Errorf("Login returned wrong user")
	}
}