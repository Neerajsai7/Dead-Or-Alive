package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	_ "modernc.org/sqlite" // CGO-free SQLite driver for Mac
)

func main() {
	// 1. Initialize the LogiTwin MCP Server
	s := server.NewMCPServer(
		"LogiTwin-Explorer",
		"1.0.0",
	)

	// 2. Define a tool to fetch inventory
	inventoryTool := mcp.NewTool("list_inventory",
		mcp.WithDescription("Lists all warehouse nodes, their current stock levels, capacity, and disruption status from the LogiTwin database"),
	)

	// 3. Register the tool handler
	s.AddTool(inventoryTool, inventoryHandler)

	// 4. Start the server using Stdio
	// Logging to Stderr so it shows up in Claude's "View Logs" button
	logger := log.New(os.Stderr, "[LogiTwin-MCP] ", log.LstdFlags)
	logger.Println("Server starting with Absolute Path logic...")

	if err := server.ServeStdio(s); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func inventoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // ⚠️ THE ABSOLUTE PATH WE JUST FOUND
    dbPath := "/Users/suyesh/Desktop/logitwin/logitwin.db"
    
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("DB Connection Error: %v", err)), nil
    }
    defer db.Close()

    // Querying your nodes table
    rows, err := db.Query("SELECT name, stock, capacity, status FROM nodes")
    if err != nil {
        // I added the dbPath to the error message so we can see it in Claude if it fails
        return mcp.NewToolResultError(fmt.Sprintf("SQL Query Failed: %v (Looked at: %s)", err, dbPath)), nil
    }
    defer rows.Close()

    var output string
    for rows.Next() {
        var name, status string
        var stock, capacity int
        if err := rows.Scan(&name, &stock, &capacity, &status); err == nil {
            output += fmt.Sprintf("🏠 %s | Stock: %d/%d | Status: %s\n", name, stock, capacity, status)
        }
    }

    if output == "" {
        output = "Database found, but the nodes table is empty."
    }

    return mcp.NewToolResultText(output), nil
}