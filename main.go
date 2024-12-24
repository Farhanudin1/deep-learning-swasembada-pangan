package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// initialize project with user account
func InitializeFirebaseApp() (*firebase.App, error) {
	opt := option.WithCredentialsFile("testing-e3e8d-firebase-adminsdk-eoic0-21df8f66f2.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %v", err)
	}
	return app, nil
}

// login system
func AuthenticateWithEmail(email, password string) (string, error) {
	client := &http.Client{}
	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=AIzaSyAec4x7qAj27adl3KEtTdGy1nxpD7_dRAM"

	payload := map[string]interface{}{
		"email":             email,
		"password":          password,
		"returnSecureToken": true,
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error authenticating with email and password: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed, status code: %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)

	token, ok := respBody["idToken"].(string)
	if !ok {
		return "", fmt.Errorf("idToken not found in response")
	}

	return token, nil
}

// try to send data to realtime database
func SendDataToRealtimeDatabase(token, path string, data interface{}) error {
	client := &http.Client{}

	uniqueID := uuid.New().String()

	// Append the unique ID to the base path
	pathWithUniqueID := fmt.Sprintf("%s/%s", path, uniqueID)

	url := fmt.Sprintf("https://testing-e3e8d-default-rtdb.asia-southeast1.firebasedatabase.app/%s.json?auth=%s", pathWithUniqueID, token)

	payloadBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", url, bytes.NewReader(payloadBytes))

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending data to realtime database: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("error ketika kirim data  : ", err)
		return fmt.Errorf("failed to send data, status code: %d", resp.StatusCode)
	}

	return nil
}

// Handler untuk melayani file HTML
func ServeHTML(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handler untuk login
// Handler untuk login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Ambil data dari form
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Menampilkan email dan password yang diterima
	fmt.Println("Email:", email, "Password:", password)

	// Autentikasi dengan Firebase
	token, err := AuthenticateWithEmail(email, password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Login failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Kirim respon sukses
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
		"token":   token,
	})
}
func main() {
	_, err := InitializeFirebaseApp()
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	// Endpoint untuk melayani file HTML
	http.HandleFunc("/", ServeHTML)

	// Endpoint untuk login
	http.HandleFunc("/login", LoginHandler)

	// Jalankan server
	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
