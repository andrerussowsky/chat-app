package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

var templates = template.Must(template.ParseFiles("templates/register.html", "templates/login.html", "templates/chat.html", "static/index.html"))
var store = sessions.NewCookieStore([]byte("secret-key"))
var jwtSecret = []byte("secret-key")

func ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Serve the home page
		templates.ExecuteTemplate(w, "index.html", nil)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Implement user registration form submission
		username := r.FormValue("username")
		password := r.FormValue("password")

		session, err := store.Get(r, username)
		if err != nil {
			// Serve registration form
			templates.ExecuteTemplate(w, "register.html", nil)
			return
		}
		session.Values["authenticated"] = false
		session.Values["password"] = password
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		// Serve registration form
		templates.ExecuteTemplate(w, "register.html", nil)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Implement user login form submission
		username := r.FormValue("username")
		password := r.FormValue("password")

		// If login is successful, set session and redirect to chat
		session, err := store.Get(r, username)
		if err != nil {
			// Serve login form
			templates.ExecuteTemplate(w, "login.html", nil)
			return
		}
		if pass, ok := session.Values["password"].(string); !ok || pass != password {
			templates.ExecuteTemplate(w, "login.html", nil)
			return
		}
		session.Values["authenticated"] = true
		session.Save(r, w)

		token, err := GenerateJWTToken(username)
		if err != nil {
			// Serve login form
			templates.ExecuteTemplate(w, "login.html", nil)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/chat?token=%s", token), http.StatusSeeOther)
	} else {
		// Serve login form
		templates.ExecuteTemplate(w, "login.html", nil)
	}
}

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
