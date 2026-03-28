package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"logitwin/database"
	"logitwin/graph"
	"logitwin/simulation"

	"golang.org/x/crypto/bcrypt"
)

// ----------------------------
// HELPERS
// ----------------------------
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

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
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

// ----------------------------
// HEALTH / KPIs
// ----------------------------
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	var totalVehicles, activeVehicles, delayedVehicles int
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&totalVehicles)
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicles WHERE status='in-transit'").Scan(&activeVehicles)
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicles WHERE status='delayed'").Scan(&delayedVehicles)

	var disruptedNodes int
	database.DB.QueryRow("SELECT COUNT(*) FROM nodes WHERE status='disrupted'").Scan(&disruptedNodes)

	var totalStock, totalCapacity int
	database.DB.QueryRow("SELECT SUM(stock), SUM(capacity) FROM nodes").Scan(&totalStock, &totalCapacity)

	inventoryHealth := 0.0
	if totalCapacity > 0 {
		inventoryHealth = float64(totalStock) / float64(totalCapacity) * 100
	}

	onTimeRate := 0.0
	if totalVehicles > 0 {
		onTimeRate = float64(activeVehicles) / float64(totalVehicles) * 100
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"activeVehicles":  activeVehicles,
		"totalVehicles":   totalVehicles,
		"delayedVehicles": delayedVehicles,
		"onTimeRate":      onTimeRate,
		"disruptions":     disruptedNodes,
		"inventoryHealth": inventoryHealth,
	})
}

// ----------------------------
// NODES
// ----------------------------
func NodesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id,name,lat,lng,type,status,stock,capacity FROM nodes")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer rows.Close()

	var nodes []map[string]interface{}
	for rows.Next() {
		var id, name, typ, status string
		var lat, lng float64
		var stock, capacity int
		rows.Scan(&id, &name, &lat, &lng, &typ, &status, &stock, &capacity)
		nodes = append(nodes, map[string]interface{}{
			"id": id, "name": name, "lat": lat, "lng": lng,
			"type": typ, "status": status, "stock": stock, "capacity": capacity,
		})
	}
	writeJSON(w, http.StatusOK, nodes)
}

// ----------------------------
// ROUTES
// ----------------------------
func RoutesHandler(w http.ResponseWriter, r *http.Request) {
	nodeRows, _ := database.DB.Query("SELECT id,lat,lng,name FROM nodes")
	defer nodeRows.Close()

	var nodes []graph.Node
	for nodeRows.Next() {
		var n graph.Node
		nodeRows.Scan(&n.ID, &n.Lat, &n.Lng, &n.Name)
		nodes = append(nodes, n)
	}

	routes := graph.BuildRoutes(nodes)

	var result []map[string]interface{}
	for i, rt := range routes {
		// Check if either node is disrupted
		status := "active"
		if simulation.DisruptedNodes[rt.From] || simulation.DisruptedNodes[rt.To] {
			status = "disrupted"
		}
		result = append(result, map[string]interface{}{
			"id":          "R" + string(rune('0'+i+1)),
			"from":        rt.From,
			"to":          rt.To,
			"distance_km": int(rt.DistanceKm),
			"status":      status,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// ----------------------------
// VEHICLES
// ----------------------------
func VehiclesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(
		"SELECT id,type,status,load,lat,lng,origin_id,dest_id,progress,eta FROM vehicles",
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer rows.Close()

	var vehicles []map[string]interface{}
	for rows.Next() {
		var id, typ, status, originID, destID, eta string
		var load int
		var lat, lng, progress float64
		rows.Scan(&id, &typ, &status, &load, &lat, &lng, &originID, &destID, &progress, &eta)
		vehicles = append(vehicles, map[string]interface{}{
			"id": id, "type": typ, "status": status, "load": load,
			"lat": lat, "lng": lng, "origin_id": originID,
			"dest_id": destID, "progress": progress, "eta": eta,
		})
	}
	writeJSON(w, http.StatusOK, vehicles)
}

// ----------------------------
// INVENTORY
// ----------------------------
func InventoryHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id,name,stock,capacity,status FROM nodes")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer rows.Close()

	var inventory []map[string]interface{}
	for rows.Next() {
		var id, name, status string
		var stock, capacity int
		rows.Scan(&id, &name, &stock, &capacity, &status)

		invStatus := "ok"
		pct := float64(stock) / float64(capacity) * 100
		if pct < 15 {
			invStatus = "critical"
		} else if pct < 35 {
			invStatus = "low"
		}

		inventory = append(inventory, map[string]interface{}{
			"id":       id,
			"name":     name,
			"node":     id,
			"stock":    stock,
			"capacity": capacity,
			"status":   invStatus,
			"pct":      pct,
		})
	}
	writeJSON(w, http.StatusOK, inventory)
}

// ----------------------------
// EVENTS
// ----------------------------
func EventsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(
		"SELECT id,time,type,message,severity FROM events ORDER BY id DESC LIMIT 20",
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id int
		var t, typ, message, severity string
		rows.Scan(&id, &t, &typ, &message, &severity)
		events = append(events, map[string]interface{}{
			"id": id, "time": t, "type": typ,
			"message": message, "severity": severity,
		})
	}
	writeJSON(w, http.StatusOK, events)
}

// ----------------------------
// SIMULATE DISRUPTION
// ----------------------------
func DisruptHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	nodeID := data["node_id"]
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id required"})
		return
	}
	simulation.SimulateDisruption(nodeID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "disruption simulated",
		"node_id": nodeID,
		"impact":  "Vehicles rerouting, inventory depleting faster",
	})
}

// ----------------------------
// CLEAR DISRUPTION
// ----------------------------
func ClearDisruptHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	nodeID := data["node_id"]
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id required"})
		return
	}
	simulation.ClearDisruption(nodeID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "disruption cleared",
		"node_id": nodeID,
	})
}

// ----------------------------
// AUTH — SIGNUP
// ----------------------------
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	email, password := data["email"], data["password"]
	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := database.DB.Exec("INSERT INTO users(email,password) VALUES(?,?)", email, string(hashed))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user already exists"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "user created"})
}

// ----------------------------
// AUTH — LOGIN
// ----------------------------
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	email, password := data["email"], data["password"]
	if email == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
		return
	}
	var stored string
	err := database.DB.QueryRow("SELECT password FROM users WHERE email=?", email).Scan(&stored)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(stored), []byte(password)) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid password"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "login success"})
}

// ----------------------------
// AUTH — SEND OTP
// ----------------------------
func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	email := data["email"]
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email required"})
		return
	}
	otp := string(rune('0'+rand.Intn(10))) + string(rune('0'+rand.Intn(10))) +
		string(rune('0'+rand.Intn(10))) + string(rune('0'+rand.Intn(10))) +
		string(rune('0'+rand.Intn(10))) + string(rune('0'+rand.Intn(10)))
	expiry := time.Now().Add(5 * time.Minute).Format(time.RFC3339)
	database.DB.Exec("UPDATE users SET otp=?,otp_expiry=? WHERE email=?", otp, expiry, email)
	writeJSON(w, http.StatusOK, map[string]string{"status": "otp sent", "otp": otp})
}

// ----------------------------
// AUTH — VERIFY OTP
// ----------------------------
func VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	email, otp := data["email"], data["otp"]
	var dbOTP, expiryStr string
	err := database.DB.QueryRow("SELECT otp,otp_expiry FROM users WHERE email=?", email).Scan(&dbOTP, &expiryStr)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if otp != dbOTP {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid otp"})
		return
	}
	expiry, _ := time.Parse(time.RFC3339, expiryStr)
	if time.Now().After(expiry) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "otp expired"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "otp verified"})
}

// ----------------------------
// AUTH — RESET PASSWORD
// ----------------------------
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	email, newPassword := data["email"], data["password"]
	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	database.DB.Exec(
		"UPDATE users SET password=?,otp=NULL,otp_expiry=NULL WHERE email=?",
		string(hashed), email,
	)
	writeJSON(w, http.StatusOK, map[string]string{"status": "password updated"})
}