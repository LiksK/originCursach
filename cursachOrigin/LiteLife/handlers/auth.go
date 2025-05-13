package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"github.com/gorilla/sessions"
	"litelife/models"
	"litelife/database"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))
var templates *template.Template

func InitTemplates() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := models.PageData{}
	templates.ExecuteTemplate(w, "login.html", data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	data := models.PageData{}
	templates.ExecuteTemplate(w, "register.html", data)
}

func LoginProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var storedPassword string
	err := database.DB.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&storedPassword)

	if err == sql.ErrNoRows {
		data := models.PageData{Error: "User not found"}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	} else if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		data := models.PageData{Error: "Incorrect password"}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	session, _ := store.Get(r, "session-name")
	session.Values["username"] = username
	session.Save(r, w)

	if username == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/user", http.StatusSeeOther)
	}
}

func RegisterProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		data := models.PageData{Error: "Username already taken"}
		templates.ExecuteTemplate(w, "register.html", data)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["username"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
} 