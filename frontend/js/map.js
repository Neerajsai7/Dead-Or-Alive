/**
 * map.js — Live supply chain map using Leaflet
 * Vehicles on routes → trucks.png (rotated toward destination)
 * Vehicles off-route / delayed → aero.png
 */

let mapInstance = null;
let nodeMarkers = {};
let vehicleMarkers = {};
let routeLines = {};

// Track previous positions to compute heading
const prevPositions = {};

const NODE_COLORS = {
  active:    "#10b981",
  disrupted: "#ef4444",
  inactive:  "#64748b"
};

/* ── Heading helpers ──────────────────────────────────────── */
function bearingDeg(lat1, lng1, lat2, lng2) {
  const toRad = d => d * Math.PI / 180;
  const dLng  = toRad(lng2 - lng1);
  const φ1    = toRad(lat1), φ2 = toRad(lat2);
  const y     = Math.sin(dLng) * Math.cos(φ2);
  const x     = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(dLng);
  return (Math.atan2(y, x) * 180 / Math.PI + 360) % 360;
}

/* ── Build a Leaflet DivIcon for a vehicle ────────────────── */
function buildVehicleIcon(v, heading) {
  const isOnRoute  = v.status === 'in-transit';
  const imgFile    = isOnRoute ? 'trucks.png' : 'aero.png';

  // trucks.png sprite faces RIGHT (0°=East). aero.png faces UP (0°=North).
  // Leaflet bearing: 0°=North, 90°=East
  const baseRotation = isOnRoute ? -90 : 0;   // trucks need -90° correction
  const rotate       = heading + baseRotation;

  const size   = isOnRoute ? 52 : 40;   // trucks icon a bit bigger
  const half   = size / 2;

  const html = `
    <div style="
      width:${size}px;height:${size}px;
      transform:rotate(${rotate}deg);
      transform-origin:center;
      filter:drop-shadow(0 2px 6px rgba(0,0,0,0.45));
    ">
      <img src="${imgFile}" style="width:100%;height:100%;object-fit:contain;" />
    </div>`;

  return L.divIcon({
    className: '',
    html,
    iconSize:   [size, size],
    iconAnchor: [half, half]
  });
}

/* ── Map init ─────────────────────────────────────────────── */
function initMap() {
  if (mapInstance) return;

  mapInstance = L.map("map", {
    center: [20.5937, 78.9629],
    zoom: 5,
    zoomControl: true,
  });

  L.tileLayer("https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png", {
    attribution: "© OpenStreetMap © CartoDB",
    subdomains: "abcd",
    maxZoom: 19
  }).addTo(mapInstance);
}

/* ── Main update ──────────────────────────────────────────── */
function updateMap(nodes, vehicles, routes) {
  if (!mapInstance) initMap();

  // Build node lookup
  const nodeLookup = {};
  nodes.forEach(n => { nodeLookup[n.id] = n; });

  /* ── Routes ───────────────────────────────────────────── */
  Object.values(routeLines).forEach(l => mapInstance.removeLayer(l));
  routeLines = {};

  routes.forEach(rt => {
    const from = nodeLookup[rt.from];
    const to   = nodeLookup[rt.to];
    if (!from || !to) return;

    const isDisrupted = rt.status === 'disrupted';
    const line = L.polyline(
      [[from.lat, from.lng], [to.lat, to.lng]],
      {
        color:     isDisrupted ? '#ef4444' : '#94a3b8',
        weight:    isDisrupted ? 3 : 2,
        opacity:   0.65,
        dashArray: isDisrupted ? '6, 10' : null
      }
    ).addTo(mapInstance);

    routeLines[rt.id] = line;
  });

  /* ── Nodes ────────────────────────────────────────────── */
  nodes.forEach(node => {
    const color = NODE_COLORS[node.status] || '#64748b';
    const pct   = ((node.stock / node.capacity) * 100).toFixed(0);

    const icon = L.divIcon({
      className: '',
      html: `<div style="
        background:${color};border:2px solid #fff;border-radius:50%;
        width:16px;height:16px;box-shadow:0 0 8px ${color};
      "></div>`,
      iconSize: [16, 16], iconAnchor: [8, 8]
    });

    if (nodeMarkers[node.id]) {
      nodeMarkers[node.id].setIcon(icon);
      nodeMarkers[node.id].getPopup().setContent(nodePopup(node, pct));
    } else {
      nodeMarkers[node.id] = L.marker([node.lat, node.lng], { icon })
        .addTo(mapInstance)
        .bindPopup(nodePopup(node, pct));
    }
  });

  /* ── Vehicles ─────────────────────────────────────────── */
  vehicles.forEach(v => {
    const lat = v.lat, lng = v.lng;

    // Compute heading from previous position
    let heading = 90; // default → East
    if (prevPositions[v.id]) {
      const { lat: pLat, lng: pLng } = prevPositions[v.id];
      const moved = Math.abs(lat - pLat) + Math.abs(lng - pLng);
      if (moved > 0.0001) {
        heading = bearingDeg(pLat, pLng, lat, lng);
      } else {
        heading = prevPositions[v.id].heading || 90;
      }
    }
    prevPositions[v.id] = { lat, lng, heading };

    const icon = buildVehicleIcon(v, heading);

    if (vehicleMarkers[v.id]) {
      vehicleMarkers[v.id].setLatLng([lat, lng]);
      vehicleMarkers[v.id].setIcon(icon);
      vehicleMarkers[v.id].getPopup().setContent(vehiclePopup(v));
    } else {
      vehicleMarkers[v.id] = L.marker([lat, lng], { icon })
        .addTo(mapInstance)
        .bindPopup(vehiclePopup(v));
    }
  });
}

/* ── Popup helpers ────────────────────────────────────────── */
function nodePopup(node, pct) {
  return `
    <div style="color:#0f172a;min-width:160px">
      <b>${node.name}</b><br>
      Type: ${node.type}<br>
      Status: <b style="color:${NODE_COLORS[node.status]}">${node.status}</b><br>
      Stock: ${node.stock} / ${node.capacity} (${pct}%)
    </div>`;
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
    </div>`;
}