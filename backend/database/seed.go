package database

import "fmt"

func Seed() {
	// Skip if already seeded
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	if count > 0 {
		fmt.Println("✅ Database already seeded")
		return
	}

	// --- NODES ---
	nodes := []struct {
		id, name, typ, status string
		lat, lng              float64
		stock, capacity       int
	}{
		{"N001", "Mumbai Hub",       "warehouse",     "active",    19.0760, 72.8777, 1800, 2000},
		{"N002", "Delhi Center",     "distribution",  "active",    28.6139, 77.2090, 420,  500},
		{"N003", "Chennai Port",     "port",          "disrupted", 13.0827, 80.2707, 890,  1000},
		{"N004", "Kolkata Depot",    "depot",         "active",    22.5726, 88.3639, 50,   800},
		{"N005", "Hyderabad Node",   "warehouse",     "active",    17.3850, 78.4867, 670,  700},
		{"N006", "Bangalore Hub",    "distribution",  "active",    12.9716, 77.5946, 430,  600},
	}

	for _, n := range nodes {
		DB.Exec(
			"INSERT INTO nodes(id,name,lat,lng,type,status,stock,capacity) VALUES(?,?,?,?,?,?,?,?)",
			n.id, n.name, n.lat, n.lng, n.typ, n.status, n.stock, n.capacity,
		)
	}

	// --- VEHICLES ---
	vehicles := []struct {
		id, typ, status, originID, destID string
		load                              int
		lat, lng, progress                float64
	}{
		{"TRK-101", "truck", "in-transit", "N001", "N002", 85, 19.0760, 72.8777, 0.1},
		{"TRK-202", "truck", "in-transit", "N002", "N004", 60, 28.6139, 77.2090, 0.2},
		{"TRK-303", "truck", "delayed",    "N001", "N003", 90, 16.0000, 76.0000, 0.5},
		{"TRK-404", "van",   "in-transit", "N005", "N006", 45, 17.3850, 78.4867, 0.3},
		{"TRK-505", "van",   "idle",       "N006", "N006", 0,  12.9716, 77.5946, 0.0},
		{"TRK-606", "truck", "in-transit", "N003", "N005", 70, 13.0827, 80.2707, 0.4},
		{"AIR-001", "air",   "in-transit", "N001", "N004", 30, 22.0000, 80.0000, 0.6},
		{"TRK-707", "truck", "in-transit", "N004", "N001", 55, 24.0000, 83.0000, 0.3},
		{"VAN-001", "van",   "idle",       "N002", "N002", 0,  28.6139, 77.2090, 0.0},
		{"TRK-808", "truck", "in-transit", "N006", "N002", 80, 15.0000, 77.0000, 0.2},
		{"AIR-002", "air",   "in-transit", "N003", "N006", 25, 13.5000, 79.0000, 0.7},
		{"TRK-909", "truck", "delayed",    "N005", "N001", 65, 18.0000, 75.0000, 0.4},
	}

	for _, v := range vehicles {
		DB.Exec(
			"INSERT INTO vehicles(id,type,status,load,lat,lng,origin_id,dest_id,progress,eta) VALUES(?,?,?,?,?,?,?,?,?,?)",
			v.id, v.typ, v.status, v.load, v.lat, v.lng, v.originID, v.destID, v.progress, "Calculating...",
		)
	}

	// --- SEED EVENTS ---
	events := []struct{ time, typ, message, severity string }{
		{"09:12", "disruption", "Chennai Port route disrupted due to weather", "high"},
		{"08:45", "alert",      "FMCG inventory at Kolkata critically low",    "high"},
		{"08:30", "info",       "TRK-101 departed Mumbai Hub on schedule",     "low"},
		{"07:55", "alert",      "Stock below threshold at Delhi Center",       "medium"},
		{"07:20", "info",       "AIR-001 cleared customs, ETA 45 minutes",    "low"},
		{"06:50", "disruption", "TRK-303 delayed — road closure on NH-44",    "high"},
	}

	for _, e := range events {
		DB.Exec(
			"INSERT INTO events(time,type,message,severity) VALUES(?,?,?,?)",
			e.time, e.typ, e.message, e.severity,
		)
	}

	fmt.Println("✅ Database seeded with 6 nodes, 12 vehicles, 6 events")
}