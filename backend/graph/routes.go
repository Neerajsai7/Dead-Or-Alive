package graph

import "math"

type Node struct {
	ID   string
	Lat  float64
	Lng  float64
	Name string
}

type Edge struct {
	From       string
	To         string
	DistanceKm float64
}

// Haversine distance between two lat/lng points in km
func Distance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// Interpolate position between two points given progress (0.0 to 1.0)
func Interpolate(lat1, lng1, lat2, lng2, progress float64) (float64, float64) {
	lat := lat1 + (lat2-lat1)*progress
	lng := lng1 + (lng2-lng1)*progress
	return lat, lng
}

// BuildRoutes returns all edges between nodes
func BuildRoutes(nodes []Node) []Edge {
	routes := []Edge{
		{"N001", "N002", Distance(19.0760, 72.8777, 28.6139, 77.2090)},
		{"N002", "N004", Distance(28.6139, 77.2090, 22.5726, 88.3639)},
		{"N001", "N003", Distance(19.0760, 72.8777, 13.0827, 80.2707)},
		{"N005", "N006", Distance(17.3850, 78.4867, 12.9716, 77.5946)},
		{"N006", "N002", Distance(12.9716, 77.5946, 28.6139, 77.2090)},
		{"N003", "N005", Distance(13.0827, 80.2707, 17.3850, 78.4867)},
		{"N004", "N001", Distance(22.5726, 88.3639, 19.0760, 72.8777)},
		{"N005", "N001", Distance(17.3850, 78.4867, 19.0760, 72.8777)},
	}
	return routes
}