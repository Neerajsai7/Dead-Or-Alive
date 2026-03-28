package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	_ "modernc.org/sqlite" // CGO-free SQLite driver for Mac/Linux/Windows
)

func main() {
	// 1. Initialize the LogiTwin MCP Server
	s := server.NewMCPServer(
		"LogiTwin-Explorer",
		"1.0.0",
	)

	// 2. Define a tool to fetch inventory
	inventoryTool := mcp.NewTool("list_inventory",
		mcp.WithDescription("Lists all items and their quantities from the LogiTwin database"),
	)

	// 3. Register the tool handler
	s.AddTool(inventoryTool, inventoryHandler)

	// 4. Start the server using Stdio
	// CRITICAL: We use Stderr for logging because Stdout is reserved for the JSON-RPC protocol
	logger := log.New(os.Stderr, "[LogiTwin-MCP] ", log.LstdFlags)
	logger.Println("Server starting...")

	if err := server.ServeStdio(s); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func inventoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Open the database (located in the root folder, one level up)
	db, err := sql.Open("sqlite", "../logitwin.db")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("DB Error: %v", err)), nil
	}
	defer db.Close()

	// Example Query: Adjust 'inventory' to your actual table name
	rows, err := db.Query("SELECT item_name, quantity FROM inventory")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query failed: %v", err)), nil
	}
	defer rows.Close()

	var output string
	for rows.Next() {
		var name string
		var qty int
		if err := rows.Scan(&name, &qty); err == nil {
			output += fmt.Sprintf("📦 %s: %d\n", name, qty)
		}
	}

	if output == "" {
		output = "Database is connected, but the inventory table is empty."
	}

	return mcp.NewToolResultText(output), nil
}