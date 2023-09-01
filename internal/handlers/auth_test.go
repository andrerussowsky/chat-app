package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"

	"github.com/dgrijalva/jwt-go"
)

func TestServeHome(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeHome(template.New("")))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestRegisterHandler_ServeForm(t *testing.T) {
	req, err := http.NewRequest("GET", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := RegisterHandler(template.New(""))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestRegisterHandler_SubmitForm(t *testing.T) {
	formValues := url.Values{
		"username": {"testuser_register"},
		"password": {"testpassword_register"},
	}

	req, err := http.NewRequest("POST", "/register", strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockTemplate := template.New("")
	rr := httptest.NewRecorder()
	handler := RegisterHandler(mockTemplate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}

func TestLoginHandler_ServeForm(t *testing.T) {
	req, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := LoginHandler(template.New(""))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestLoginHandler_SubmitForm(t *testing.T) {
	formValues := url.Values{
		"username": {"testuser_login"},
		"password": {"testpassword_login"},
	}

	req, err := http.NewRequest("POST", "/login", strings.NewReader(formValues.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	mockTemplate := template.New("")
	rr := httptest.NewRecorder()
	handler := LoginHandler(mockTemplate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}
}

func TestGenerateJWTToken(t *testing.T) {
	testCases := []struct {
		username string
	}{
		{"testuser1"},
		{"testuser2"},
	}

	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			tokenString, err := GenerateJWTToken(tc.username)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse the token to verify its contents
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verify that the signing method is as expected
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return jwtSecret, nil
			})
			if err != nil {
				t.Fatalf("unexpected error parsing token: %v", err)
			}

			// Verify claims
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if username, found := claims["username"]; !found || username != tc.username {
					t.Errorf("unexpected username in token claims: got %v, want %v", username, tc.username)
				}
			} else {
				t.Error("failed to retrieve claims from token")
			}
		})
	}
}

func GenerateToken(username string) string {
	token, _ := GenerateJWTToken(username)
	return token
}

func TestParseJWTToken(t *testing.T) {
	testCases := []struct {
		tokenString   string
		expectedUser  string
		shouldSucceed bool
	}{
		{GenerateToken("user1"), "user1", true},
		{GenerateToken("user2"), "user2", true},
		{"invalid-token", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.tokenString, func(t *testing.T) {
			username, err := ParseJWTToken(tc.tokenString)

			if tc.shouldSucceed {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if username != tc.expectedUser {
					t.Errorf("unexpected username: got %v, want %v", username, tc.expectedUser)
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			}
		})
	}
}
