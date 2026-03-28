/**
 * map.js — Live supply chain map using Leaflet
 * Vehicles move in real-time, nodes show status colors
 */

let mapInstance = null;
let nodeMarkers = {};
let vehicleMarkers = {};
let routeLines = {};

const NODE_COLORS = {
  active:    "#3b82f6",
  disrupted: "#dc2626",
  inactive:  "#64748b"
};

const VEHICLE_ICONS = {
  truck: "🚛",
  van:   "🚐",
  air:   "✈️"
};

function initMap() {
  if (mapInstance) return;

  mapInstance = L.map("map", {
    center: [20.5937, 78.9629], // Center of India
    zoom: 5,
    zoomControl: true,
  });

  // Dark tile layer
  L.tileLayer("https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png", {
    attribution: "© OpenStreetMap © CartoDB",
    subdomains: "abcd",
    maxZoom: 19
  }).addTo(mapInstance);
}

function updateMap(nodes, vehicles, routes) {
  if (!mapInstance) initMap();

  // ── Draw Routes ─────────────────────────────────────────
  // Build node lookup
  const nodeLookup = {};
  nodes.forEach(n => { nodeLookup[n.id] = n; });

  // Clear old route lines
  Object.values(routeLines).forEach(l => mapInstance.removeLayer(l));
  routeLines = {};

  routes.forEach(rt => {
    const from = nodeLookup[rt.from];
    const to   = nodeLookup[rt.to];
    if (!from || !to) return;

    const color  = rt.status === "disrupted" ? "#dc2626" : "#1e40af";
    const dashed = rt.status === "disrupted" ? [8, 6] : [0];

    const line = L.polyline(
      [[from.lat, from.lng], [to.lat, to.lng]],
      { color, weight: 2, opacity: 0.7, dashArray: dashed }
    ).addTo(mapInstance);

    routeLines[rt.id] = line;
  });

  // ── Draw / Update Nodes ──────────────────────────────────
  nodes.forEach(node => {
    const color = NODE_COLORS[node.status] || "#64748b";
    const pct   = ((node.stock / node.capacity) * 100).toFixed(0);

    const icon = L.divIcon({
      className: "",
      html: `
        <div style="
          background:${color};
          border:2px solid #fff;
          border-radius:50%;
          width:16px;height:16px;
          box-shadow:0 0 8px ${color};
        "></div>
      `,
      iconSize: [16, 16],
      iconAnchor: [8, 8]
    });

    if (nodeMarkers[node.id]) {
      nodeMarkers[node.id].setIcon(icon);
      nodeMarkers[node.id]
        .getPopup()
        .setContent(nodePopup(node, pct));
    } else {
      const marker = L.marker([node.lat, node.lng], { icon })
        .addTo(mapInstance)
        .bindPopup(nodePopup(node, pct));
      nodeMarkers[node.id] = marker;
    }
  });

  // ── Draw / Update Vehicles ───────────────────────────────
  vehicles.forEach(v => {
    const emoji = VEHICLE_ICONS[v.type] || "🚛";
    const color = v.status === "delayed"    ? "#f97316"
                : v.status === "in-transit" ? "#4ade80"
                : "#64748b";

    const icon = L.divIcon({
      className: "",
      html: `<div style="font-size:18px;filter:drop-shadow(0 0 4px ${color});">${emoji}</div>`,
      iconSize: [24, 24],
      iconAnchor: [12, 12]
    });

    if (vehicleMarkers[v.id]) {
      vehicleMarkers[v.id].setLatLng([v.lat, v.lng]);
      vehicleMarkers[v.id].setIcon(icon);
      vehicleMarkers[v.id]
        .getPopup()
        .setContent(vehiclePopup(v));
    } else {
      const marker = L.marker([v.lat, v.lng], { icon })
        .addTo(mapInstance)
        .bindPopup(vehiclePopup(v));
      vehicleMarkers[v.id] = marker;
    }
  });
}

function nodePopup(node, pct) {
  return `
    <div style="color:#0f172a;min-width:160px">
      <b>${node.name}</b><br>
      Type: ${node.type}<br>
      Status: <b style="color:${NODE_COLORS[node.status]}">${node.status}</b><br>
      Stock: ${node.stock} / ${node.capacity} (${pct}%)
    </div>
  `;
}

function vehiclePopup(v) {
  return `
    <div style="color:#0f172a;min-width:140px">
      <b>${v.id}</b><br>
      Type: ${v.type}<br>
      Status: <b>${v.status}</b><br>
      Load: ${v.load}%<br>
      ETA: ${v.eta}<br>
      Progress: ${(v.progress * 100).toFixed(0)}%
    </div>
  `;
}