package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ─────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func parseBody(r *http.Request) map[string]string {
	var data map[string]string
	json.NewDecoder(r.Body).Decode(&data)
	return data
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// ─────────────────────────────────────────────
// SIGNUP
// POST /api/signup  { "email": "...", "password": "..." }
// ─────────────────────────────────────────────
func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	data := parseBody(r)
	email, password := data["email"], data["password"]

	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	if len(password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	_, err = db.Exec("INSERT INTO users(email, password) VALUES(?, ?)", email, string(hashed))
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}

	fmt.Printf("✅ New user registered: %s\n", email)
	writeJSON(w, http.StatusOK, map[string]string{"status": "user created"})
}

// ─────────────────────────────────────────────
// LOGIN
// POST /api/login  { "email": "...", "password": "..." }
// ─────────────────────────────────────────────
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	data := parseBody(r)
	email, password := data["email"], data["password"]

	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}

	var stored string
	err := db.QueryRow("SELECT password FROM users WHERE email = ?", email).Scan(&stored)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid password"})
		return
	}

	fmt.Printf("🔑 Login: %s\n", email)
	writeJSON(w, http.StatusOK, map[string]string{"status": "login success", "email": email})
}

// ─────────────────────────────────────────────
// SEND OTP  (for forgot-password flow)
// POST /api/send-otp  { "email": "..." }
//
// NOTE: In production, integrate an SMTP or third-party email
// service (SendGrid, Mailgun, AWS SES) here.
// For local dev the OTP is returned in the JSON response so
// you can paste it directly — remove that field before deploying!
// ─────────────────────────────────────────────
func sendOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	data := parseBody(r)
	email := data["email"]
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email required"})
		return
	}

	// Check user exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if count == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no account found with that email"})
		return
	}

	// Generate 6-digit OTP
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", rng.Intn(1000000))
	expiry := time.Now().Add(5 * time.Minute).Format(time.RFC3339)

	db.Exec("UPDATE users SET otp = ?, otp_expiry = ? WHERE email = ?", otp, expiry, email)

	// ── In production: send email here ──────────────────────
	// sendEmail(email, "LogiTwin OTP", "Your code is: " + otp)
	// ────────────────────────────────────────────────────────

	fmt.Printf("📧 OTP for %s: %s (expires %s)\n", email, otp, expiry)

	// DEV: return OTP in response for easy testing
	// PRODUCTION: remove "otp" from this response
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "otp sent",
		"otp":    otp, // ← remove in production
	})
}

// ─────────────────────────────────────────────
// VERIFY OTP
// POST /api/verify-otp  { "email": "...", "otp": "123456" }
// ─────────────────────────────────────────────
func verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	data := parseBody(r)
	email, otp := data["email"], data["otp"]

	if email == "" || otp == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and otp required"})
		return
	}

	var dbOTP, expiryStr string
	err := db.QueryRow("SELECT otp, otp_expiry FROM users WHERE email = ?", email).Scan(&dbOTP, &expiryStr)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	if dbOTP == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no OTP requested — please request a new one"})
		return
	}

	if otp != dbOTP {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid OTP"})
		return
	}

	expiry, _ := time.Parse(time.RFC3339, expiryStr)
	if time.Now().After(expiry) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "OTP has expired — please request a new one"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "otp verified"})
}

// ─────────────────────────────────────────────
// RESET PASSWORD
// POST /api/reset-password  { "email": "...", "password": "newpass" }
// ─────────────────────────────────────────────
func resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	data := parseBody(r)
	email, newPassword := data["email"], data["password"]

	if email == "" || newPassword == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	if len(newPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password must be at least 8 characters"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	result, _ := db.Exec(
		"UPDATE users SET password = ?, otp = NULL, otp_expiry = NULL WHERE email = ?",
		string(hashed), email,
	)

	rows, _ := result.RowsAffected()
	if rows == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	fmt.Printf("🔐 Password reset for: %s\n", email)
	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}