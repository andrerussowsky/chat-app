package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"text/template"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))
var jwtSecret = []byte("secret-key")

// ServeHome serves the home page
func ServeHome(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			templates.ExecuteTemplate(w, "index.html", nil)
		}
	}
}

// RegisterHandler serves the registration page
func RegisterHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			session, err := store.Get(r, username)
			if err != nil {
				http.Redirect(w, r, "/register", http.StatusSeeOther)
				return
			}
			session.Values["authenticated"] = false
			session.Values["password"] = GetMD5Hash(password)
			session.Save(r, w)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			templates.ExecuteTemplate(w, "register.html", nil)
		}
	}
}

// LoginHandler serves the login page
func LoginHandler(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			session, err := store.Get(r, username)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if pass, ok := session.Values["password"].(string); !ok || pass != GetMD5Hash(password) {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			session.Values["authenticated"] = true
			session.Save(r, w)

			token, err := GenerateJWTToken(username)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// If login is successful, set session and redirect to chat
			http.Redirect(w, r, fmt.Sprintf("/chat?token=%s", token), http.StatusSeeOther)
		} else {
			templates.ExecuteTemplate(w, "login.html", nil)
		}
	}
}

// GenerateJWTToken generates a JWT token
func GenerateJWTToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWTToken parses a JWT token
func ParseJWTToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, exists := claims["username"].(string); exists {
			return username, nil
		}
	}

	return "", errors.New("invalid token")
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
