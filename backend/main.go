package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// ─────────────────────────────────────────────
// GLOBAL DB — shared by main.go, auth.go, simulation.go
// ─────────────────────────────────────────────
var db *sql.DB

func main() {
	initDB()
	seedDB()

	go runSimulation() // background goroutine — see simulation.go

	// ── Auth routes (auth.go) ────────────────────────────────
	http.HandleFunc("/api/signup", corsMiddleware(signupHandler))
	http.HandleFunc("/api/login", corsMiddleware(loginHandler))
	http.HandleFunc("/api/send-otp", corsMiddleware(sendOTPHandler))
	http.HandleFunc("/api/verify-otp", corsMiddleware(verifyOTPHandler))
	http.HandleFunc("/api/reset-password", corsMiddleware(resetPasswordHandler))

	// ── Simulation / data routes (handlers below) ────────────
	http.HandleFunc("/health", corsMiddleware(healthHandler))
	http.HandleFunc("/nodes", corsMiddleware(nodesHandler))
	http.HandleFunc("/routes", corsMiddleware(routesHandler))
	http.HandleFunc("/vehicles", corsMiddleware(vehiclesHandler))
	http.HandleFunc("/inventory", corsMiddleware(inventoryHandler))
	http.HandleFunc("/events", corsMiddleware(eventsHandler))
	http.HandleFunc("/disrupt", corsMiddleware(disruptHandler))
	http.HandleFunc("/disrupt/clear", corsMiddleware(clearDisruptHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 LogiTwin backend running → http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ─────────────────────────────────────────────
// DATABASE INIT & SEED
// ─────────────────────────────────────────────

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./logitwin.db")
	if err != nil {
		log.Fatal("Failed to open DB:", err)
	}

	schema := `
    CREATE TABLE IF NOT EXISTS users (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        email      TEXT UNIQUE NOT NULL,
        password   TEXT NOT NULL,
        otp        TEXT,
        otp_expiry TEXT
    );
    CREATE TABLE IF NOT EXISTS nodes (
        id       TEXT PRIMARY KEY,
        name     TEXT NOT NULL,
        lat      REAL NOT NULL,
        lng      REAL NOT NULL,
        type     TEXT NOT NULL,
        status   TEXT NOT NULL,
        stock    INTEGER DEFAULT 0,
        capacity INTEGER DEFAULT 1000
    );
    CREATE TABLE IF NOT EXISTS vehicles (
        id        TEXT PRIMARY KEY,
        type      TEXT NOT NULL,
        status    TEXT NOT NULL,
        load      INTEGER DEFAULT 0,
        lat       REAL NOT NULL,
        lng       REAL NOT NULL,
        origin_id TEXT NOT NULL,
        dest_id   TEXT NOT NULL,
        progress  REAL DEFAULT 0,
        eta       TEXT DEFAULT ''
    );
    CREATE TABLE IF NOT EXISTS events (
        id       INTEGER PRIMARY KEY AUTOINCREMENT,
        time     TEXT NOT NULL,
        type     TEXT NOT NULL,
        message  TEXT NOT NULL,
        severity TEXT NOT NULL
    );`

	if _, err = db.Exec(schema); err != nil {
		log.Fatal("Schema error:", err)
	}
	fmt.Println("✅ Database initialised")
}

func seedDB() {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if count > 0 {
		fmt.Println("✅ Database already seeded")
		return
	}

	nodes := []struct {
		id, name, typ, status string
		lat, lng              float64
		stock, capacity       int
	}{
		{"N001", "Mumbai Hub", "warehouse", "active", 19.0760, 72.8777, 1800, 2000},
		{"N002", "Delhi Center", "distribution", "active", 28.6139, 77.2090, 420, 500},
		{"N003", "Chennai Port", "port", "disrupted", 13.0827, 80.2707, 890, 1000},
		{"N004", "Kolkata Depot", "depot", "active", 22.5726, 88.3639, 50, 800},
		{"N005", "Hyderabad Node", "warehouse", "active", 17.3850, 78.4867, 670, 700},
		{"N006", "Bangalore Hub", "distribution", "active", 12.9716, 77.5946, 430, 600},
	}
	for _, n := range nodes {
		db.Exec(
			"INSERT INTO nodes(id,name,lat,lng,type,status,stock,capacity) VALUES(?,?,?,?,?,?,?,?)",
			n.id, n.name, n.lat, n.lng, n.typ, n.status, n.stock, n.capacity,
		)
	}

	vehicles := []struct {
		id, typ, status, orig, dest string
		load                        int
		lat, lng, progress          float64
	}{
		{"TRK-101", "truck", "in-transit", "N001", "N002", 85, 19.0760, 72.8777, 0.1},
		{"TRK-202", "truck", "in-transit", "N002", "N004", 60, 28.6139, 77.2090, 0.2},
		{"TRK-303", "truck", "delayed", "N001", "N003", 90, 16.0000, 76.0000, 0.5},
		{"TRK-404", "van", "in-transit", "N005", "N006", 45, 17.3850, 78.4867, 0.3},
		{"TRK-505", "van", "idle", "N006", "N006", 0, 12.9716, 77.5946, 0.0},
		{"TRK-606", "truck", "in-transit", "N003", "N005", 70, 13.0827, 80.2707, 0.4},
		{"AIR-001", "air", "in-transit", "N001", "N004", 30, 22.0000, 80.0000, 0.6},
		{"TRK-707", "truck", "in-transit", "N004", "N001", 55, 24.0000, 83.0000, 0.3},
		{"VAN-001", "van", "idle", "N002", "N002", 0, 28.6139, 77.2090, 0.0},
		{"TRK-808", "truck", "in-transit", "N006", "N002", 80, 15.0000, 77.0000, 0.2},
		{"AIR-002", "air", "in-transit", "N003", "N006", 25, 13.5000, 79.0000, 0.7},
		{"TRK-909", "truck", "delayed", "N005", "N001", 65, 18.0000, 75.0000, 0.4},
	}
	for _, v := range vehicles {
		db.Exec(
			"INSERT INTO vehicles(id,type,status,load,lat,lng,origin_id,dest_id,progress,eta) VALUES(?,?,?,?,?,?,?,?,?,?)",
			v.id, v.typ, v.status, v.load, v.lat, v.lng, v.orig, v.dest, v.progress, "Calculating...",
		)
	}

	events := []struct{ time, typ, msg, sev string }{
		{"09:12", "disruption", "Chennai Port route disrupted due to weather", "high"},
		{"08:45", "alert", "FMCG inventory at Kolkata critically low", "high"},
		{"08:30", "info", "TRK-101 departed Mumbai Hub on schedule", "low"},
		{"07:55", "alert", "Stock below threshold at Delhi Center", "medium"},
		{"07:20", "info", "AIR-001 cleared customs, ETA 45 minutes", "low"},
		{"06:50", "disruption", "TRK-303 delayed — road closure on NH-44", "high"},
	}
	for _, e := range events {
		db.Exec("INSERT INTO events(time,type,message,severity) VALUES(?,?,?,?)", e.time, e.typ, e.msg, e.sev)
	}

	fmt.Println("✅ Database seeded: 6 nodes, 12 vehicles, 6 events")
}

// ─────────────────────────────────────────────
// DATA HANDLERS
// ─────────────────────────────────────────────

func healthHandler(w http.ResponseWriter, r *http.Request) {
	var total, active, delayed int
	db.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&total)
	db.QueryRow("SELECT COUNT(*) FROM vehicles WHERE status='in-transit'").Scan(&active)
	db.QueryRow("SELECT COUNT(*) FROM vehicles WHERE status='delayed'").Scan(&delayed)

	var disrupted int
	db.QueryRow("SELECT COUNT(*) FROM nodes WHERE status='disrupted'").Scan(&disrupted)

	var totalStock, totalCap int
	db.QueryRow("SELECT SUM(stock), SUM(capacity) FROM nodes").Scan(&totalStock, &totalCap)

	invHealth := 0.0
	if totalCap > 0 {
		invHealth = float64(totalStock) / float64(totalCap) * 100
	}
	onTime := 0.0
	if total > 0 {
		onTime = float64(active) / float64(total) * 100
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"activeVehicles":  active,
		"totalVehicles":   total,
		"delayedVehicles": delayed,
		"onTimeRate":      onTime,
		"disruptions":     disrupted,
		"inventoryHealth": invHealth,
	})
}

func nodesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id,name,lat,lng,type,status,stock,capacity FROM nodes")
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

func routesHandler(w http.ResponseWriter, r *http.Request) {
	type NodePos struct{ Lat, Lng float64 }
	nodeLookup := map[string]NodePos{}

	rows, _ := db.Query("SELECT id,lat,lng FROM nodes")
	defer rows.Close()
	for rows.Next() {
		var id string
		var lat, lng float64
		rows.Scan(&id, &lat, &lng)
		nodeLookup[id] = NodePos{lat, lng}
	}

	type Route struct{ From, To string }
	staticRoutes := []Route{
		{"N001", "N002"}, {"N002", "N004"}, {"N001", "N003"},
		{"N005", "N006"}, {"N006", "N002"}, {"N003", "N005"},
		{"N004", "N001"}, {"N005", "N001"},
	}

	var result []map[string]interface{}
	for i, rt := range staticRoutes {
		status := "active"
		if disruptedNodes[rt.From] || disruptedNodes[rt.To] {
			status = "disrupted"
		}
		result = append(result, map[string]interface{}{
			"id":     fmt.Sprintf("R%02d", i+1),
			"from":   rt.From,
			"to":     rt.To,
			"status": status,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func vehiclesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT id,type,status,load,lat,lng,origin_id,dest_id,progress,eta FROM vehicles",
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer rows.Close()

	var vehicles []map[string]interface{}
	for rows.Next() {
		var id, typ, status, origID, destID, eta string
		var load int
		var lat, lng, progress float64
		rows.Scan(&id, &typ, &status, &load, &lat, &lng, &origID, &destID, &progress, &eta)
		vehicles = append(vehicles, map[string]interface{}{
			"id": id, "type": typ, "status": status, "load": load,
			"lat": lat, "lng": lng, "origin_id": origID,
			"dest_id": destID, "progress": progress, "eta": eta,
		})
	}
	writeJSON(w, http.StatusOK, vehicles)
}

func inventoryHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id,name,stock,capacity,status FROM nodes")
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

		pct := float64(stock) / float64(capacity) * 100
		invStatus := "ok"
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

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
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
			"id":       id,
			"time":     t,
			"type":     typ,
			"message":  message,
			"severity": severity,
		})
	}
	writeJSON(w, http.StatusOK, events)
}

func disruptHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	nodeID := data["node_id"]
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id required"})
		return
	}
	simulateDisruption(nodeID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "disruption simulated",
		"node_id": nodeID,
		"impact":  "Vehicles rerouting, inventory depleting faster",
	})
}

func clearDisruptHandler(w http.ResponseWriter, r *http.Request) {
	data := parseBody(r)
	nodeID := data["node_id"]
	if nodeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "node_id required"})
		return
	}
	clearDisruption(nodeID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "disruption cleared",
		"node_id": nodeID,
	})
}
