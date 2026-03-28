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
        mcp.WithDescription("Lists all warehouse nodes, their current stock levels, capacity, and disruption status from the LogiTwin database"),
    )

    // 3. Register the tool handler
    s.AddTool(inventoryTool, inventoryHandler)

    // 4. Start the server using Stdio
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

    // ⚠️ CORRECTED QUERY: Using your actual LogiTwin 'nodes' table
    rows, err := db.Query("SELECT name, stock, capacity, status FROM nodes")
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Query failed: %v", err)), nil
    }
    defer rows.Close()

    var output string
    for rows.Next() {
        var name, status string
        var stock, capacity int
        
        // Scan the actual columns from your LogiTwin schema
        if err := rows.Scan(&name, &stock, &capacity, &status); err == nil {
            output += fmt.Sprintf("📦 %s: %d/%d units (Status: %s)\n", name, stock, capacity, status)
        }
    }

    if output == "" {
        output = "Database is connected, but the nodes table is empty."
    }

    return mcp.NewToolResultText(output), nil
}