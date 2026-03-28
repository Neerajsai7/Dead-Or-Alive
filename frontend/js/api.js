/**
 * api.js — All fetch() calls to the LogiTwin backend
 * Change API_BASE to your Render / production URL when deploying.
 */

const API_BASE = "http://localhost:8080";
// Production: const API_BASE = "https://logitwin.onrender.com";

const API = {

  async getHealth() {
    const res = await fetch(`${API_BASE}/health`);
    return res.json();
  },

  async getNodes() {
    const res = await fetch(`${API_BASE}/nodes`);
    return res.json();
  },

  async getRoutes() {
    const res = await fetch(`${API_BASE}/routes`);
    return res.json();
  },

  async getVehicles() {
    const res = await fetch(`${API_BASE}/vehicles`);
    return res.json();
  },

  async getInventory() {
    const res = await fetch(`${API_BASE}/inventory`);
    return res.json();
  },

  async getEvents() {
    const res = await fetch(`${API_BASE}/events`);
    return res.json();
  },

  async simulateDisruption(nodeId) {
    const res = await fetch(`${API_BASE}/disrupt`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ node_id: nodeId })
    });
    return res.json();
  },

  async clearDisruption(nodeId) {
    const res = await fetch(`${API_BASE}/disrupt/clear`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ node_id: nodeId })
    });
    return res.json();
  },

  async login(email, password) {
    const res = await fetch(`${API_BASE}/api/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password })
    });
    return { ok: res.ok, data: await res.json() };
  },

  async signup(email, password) {
    const res = await fetch(`${API_BASE}/api/signup`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password })
    });
    return { ok: res.ok, data: await res.json() };
  },

  async sendOTP(email) {
    const res = await fetch(`${API_BASE}/api/send-otp`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email })
    });
    return res.json();
  },

  async verifyOTP(email, otp) {
    const res = await fetch(`${API_BASE}/api/verify-otp`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, otp })
    });
    return { ok: res.ok, data: await res.json() };
  },

  async resetPassword(email, password) {
    const res = await fetch(`${API_BASE}/api/reset-password`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password })
    });
    return res.json();
  }

};