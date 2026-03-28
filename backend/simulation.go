package main

import (
	"fmt"
	"math/rand"
	"time"

	"logitwin/graph"
)

var disruptedNodes = map[string]bool{}

func runSimulation() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		fmt.Println("🔄 Simulation engine started (ticking every 3s)")

		for range ticker.C {
			moveVehicles()
			depleteInventory()
			generateEvent()
		}
	}()
}

func moveVehicles() {
    // Changed database.DB to db
	rows, err := db.Query(
		"SELECT id, lat, lng, origin_id, dest_id, progress, status FROM vehicles",
	)
	if err != nil {
		return
	}
	defer rows.Close()

	type Vehicle struct {
		id, originID, destID, status string
		lat, lng, progress           float64
	}

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		rows.Scan(&v.id, &v.lat, &v.lng, &v.originID, &v.destID, &v.progress, &v.status)
		vehicles = append(vehicles, v)
	}

	for _, v := range vehicles {
		if v.status == "idle" {
			continue
		}

		if disruptedNodes[v.destID] || disruptedNodes[v.originID] {
			db.Exec(
				"UPDATE vehicles SET status=? WHERE id=?",
				"delayed", v.id,
			)
			continue
		}

		speed := 0.02 + rand.Float64()*0.01 
		newProgress := v.progress + speed

		var oLat, oLng, dLat, dLng float64
		db.QueryRow("SELECT lat,lng FROM nodes WHERE id=?", v.originID).Scan(&oLat, &oLng)
		db.QueryRow("SELECT lat,lng FROM nodes WHERE id=?", v.destID).Scan(&dLat, &dLng)

		if newProgress >= 1.0 {
			newProgress = 0.0
			newOrigin := v.destID
			newDest := pickNextDest(v.destID)

			db.Exec(
				"UPDATE vehicles SET lat=?,lng=?,progress=?,origin_id=?,dest_id=?,status=?,eta=? WHERE id=?",
				dLat, dLng, newProgress, newOrigin, newDest, "in-transit", "Calculating...", v.id,
			)
		} else {
			newLat, newLng := graph.Interpolate(oLat, oLng, dLat, dLng, newProgress)
			remaining := 1.0 - newProgress
			etaMins := int(remaining / speed * 3 / 60)
			eta := fmt.Sprintf("%dh %dm", etaMins/60, etaMins%60)

			db.Exec(
				"UPDATE vehicles SET lat=?,lng=?,progress=?,eta=?,status=? WHERE id=?",
				newLat, newLng, newProgress, eta, "in-transit", v.id,
			)
		}
	}
}

func depleteInventory() {
	rows, err := db.Query("SELECT id, stock FROM nodes")
	if err != nil {
		return
	}
	defer rows.Close()

	type NodeStock struct {
		id    string
		stock int
	}

	var nodes []NodeStock
	for rows.Next() {
		var n NodeStock
		rows.Scan(&n.id, &n.stock)
		nodes = append(nodes, n)
	}

	for _, n := range nodes {
		if disruptedNodes[n.id] {
			drain := rand.Intn(20) + 10
			newStock := n.stock - drain
			if newStock < 0 {
				newStock = 0
			}
			db.Exec("UPDATE nodes SET stock=? WHERE id=?", newStock, n.id)
		} else {
			drain := rand.Intn(5) + 1
			newStock := n.stock - drain
			if newStock < 0 {
				newStock = rand.Intn(200) + 100 
			}
			db.Exec("UPDATE nodes SET stock=? WHERE id=?", newStock, n.id)
		}
	}
}

func generateEvent() {
	if rand.Intn(5) != 0 { 
		return
	}

	messages := []struct{ typ, message, severity string }{
		{"info",       "Vehicle arrived at destination on schedule",         "low"},
		{"info",       "Inventory restocked at hub",                        "low"},
		{"alert",      "Stock running low at distribution center",          "medium"},
		{"alert",      "Vehicle load exceeding recommended capacity",       "medium"},
		{"disruption", "Weather delay reported on northern corridor",       "high"},
		{"disruption", "Road closure causing rerouting",                   "high"},
		{"info",       "New shipment dispatched from warehouse",            "low"},
		{"alert",      "Delivery ETA pushed back by 2 hours",              "medium"},
	}

	e := messages[rand.Intn(len(messages))]
	now := time.Now().Format("15:04")

	db.Exec(
		"INSERT INTO events(time,type,message,severity) VALUES(?,?,?,?)",
		now, e.typ, e.message, e.severity,
	)

	db.Exec(
		"DELETE FROM events WHERE id NOT IN (SELECT id FROM events ORDER BY id DESC LIMIT 50)",
	)
}

func pickNextDest(currentID string) string {
	allNodes := []string{"N001", "N002", "N003", "N004", "N005", "N006"}
	for {
		next := allNodes[rand.Intn(len(allNodes))]
		if next != currentID && !disruptedNodes[next] {
			return next
		}
	}
}

func simulateDisruption(nodeID string) {
	disruptedNodes[nodeID] = true
	db.Exec("UPDATE nodes SET status=? WHERE id=?", "disrupted", nodeID)

	now := time.Now().Format("15:04")
	db.Exec(
		"INSERT INTO events(time,type,message,severity) VALUES(?,?,?,?)",
		now, "disruption",
		fmt.Sprintf("Node %s manually disrupted — rerouting vehicles", nodeID),
		"high",
	)
}

func clearDisruption(nodeID string) {
	delete(disruptedNodes, nodeID)
	db.Exec("UPDATE nodes SET status=? WHERE id=?", "active", nodeID)

	now := time.Now().Format("15:04")
	db.Exec(
		"INSERT INTO events(time,type,message,severity) VALUES(?,?,?,?)",
		now, "info",
		fmt.Sprintf("Node %s disruption cleared — resuming normal operations", nodeID),
		"low",
	)
}