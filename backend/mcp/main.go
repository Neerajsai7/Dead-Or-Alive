package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Define a struct that matches the JSON coming from your backend /nodes endpoint
type NodeData struct {
	Name     string `json:"name"`
	Stock    int    `json:"stock"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

func main() {
	// 1. Initialize the LogiTwin MCP Server
	s := server.NewMCPServer(
		"LogiTwin-Explorer",
		"1.1.0", // Bumped version for the API upgrade!
	)

	// 2. Define the tool
	inventoryTool := mcp.NewTool("list_inventory",
		mcp.WithDescription("Lists all warehouse nodes, their current stock levels, capacity, and disruption status from the live LogiTwin API"),
	)

	// 3. Register the tool handler
	s.AddTool(inventoryTool, inventoryHandler)

	// 4. Start the server
	logger := log.New(os.Stderr, "[LogiTwin-MCP] ", log.LstdFlags)
	logger.Println("Server starting in API-Consumer mode...")

	if err := server.ServeStdio(s); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func inventoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ⚠️ CHANGE THIS TO YOUR RENDER URL ONCE IT IS LIVE
	// For local testing, leave it as http://localhost:8080/nodes
	apiURL := "http://localhost:8080/nodes"

	// Make an HTTP GET request to your backend
	resp, err := http.Get(apiURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("API Connection Error. Make sure the backend is running at %s. Details: %v", apiURL, err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mcp.NewToolResultError(fmt.Sprintf("API returned bad status: %d", resp.StatusCode)), nil
	}

	// Read and parse the JSON response
	body, _ := io.ReadAll(resp.Body)
	var nodes []NodeData
	if err := json.Unmarshal(body, &nodes); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse API data: %v", err)), nil
	}

	// Format the output for Claude
	var output string
	for _, node := range nodes {
		output += fmt.Sprintf("🏠 %s | Stock: %d/%d | Status: %s\n", node.Name, node.Stock, node.Capacity, node.Status)
	}

	if output == "" {
		output = "API connected successfully, but returned an empty inventory list."
	}

	return mcp.NewToolResultText(output), nil
}