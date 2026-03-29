LogiTwin: AI-Powered Supply Chain Digital TwinLogi

Twin is a high-fidelity Digital Twin platform for modern logistics. It bridges the gap between physical supply chain assets and digital intelligence by combining a high-concurrency Go backend with a Retrieval-Augmented Generation (RAG) AI assistant.

Live Demo: [logitwin.vercel.app] 
Backend API: [https://logitwin-api.onrender.com] 
Technical Highlights 
RAG-Powered AI Intelligence
Unlike static chatbots, the LogiTwin Assistant uses a custom RAG pipeline. When a user asks a question, the Go backend queries the live SQLite database, injects real-time network state (disruptions, delays, stock levels), and provides the context to Gemini 2.5 Flash for highly accurate, state-aware responses.

High-Performance Telemetry (60 FPS)
Implemented a custom Linear Interpolation (Lerp) Engine on the frontend. Instead of "snapping" icons every time the API polls (3s), LogiTwin calculates sub-pixel movement 60 times per second, resulting in ultra-smooth vehicle gliding across the map.

Chaos Engineering 
SimulatorA built-in Disruption Simulator allows users to manually trigger "failures" in specific nodes (Mumbai, Chennai, etc.). The system dynamically reroutes traffic, updates inventory health metrics, and alerts the AI assistant instantly.

System Architecture
Frontend: Single Page Application (SPA) with modern Glassmorphism UI, interactive Leaflet.js geospatial mapping, and real-time state synchronization.
Backend: Concurrent Go (Golang) server handling RESTful endpoints, background simulation routines, and Gemini API orchestration.
Database: Persistent SQLite3 for user authentication and supply chain state.
AI Link: Model Context Protocol (MCP) server included for connecting local LLMs (like Claude Desktop) to the live data stream.

Tech Stack
Layer       Technology
Backend     Go (Golang),SQLite3
Frontend    HTML5,CSS3 (Glassmorphic),Vanilla JavaScript
AI/ML       Google Gemini 2.5 Flash API,RAG Pipeline
Mapping     Leaflet.js,OpenStreetMap API
Deployment  Render (API), Vercel (Frontend), GitHub Actions
Installation & Setup
1. Prerequisites
      Go 1.21+
      A Google AI Studio API Key (Gemini)
2.Backend Setup
  Bash
  cd backend
  go mod tidy
  export GEMINI_API_KEY="your_api_key"
  go run .
3. Frontend Setup
Update the API_BASE in dashboard.html to http://localhost:8080 and open the file in any modern browser.