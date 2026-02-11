// Simple database connection test
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	fmt.Println("=== NER Studio Database Connection Test ===")
	fmt.Println()

	// Try 1: Direct PostgreSQL connection
	if databaseURL != "" {
		fmt.Println("Test 1: PostgreSQL Connection Pooler")
		fmt.Printf("URL: %s...\n", databaseURL[:50])
		testPostgres(databaseURL)
		fmt.Println()
	}

	// Try 2: Supabase Go Client
	if supabaseURL != "" && supabaseKey != "" {
		fmt.Println("Test 2: Supabase Go Client")
		fmt.Printf("URL: %s\n", supabaseURL)
		testSupabaseClient(supabaseURL, supabaseKey)
		fmt.Println()
	}

	fmt.Println("=== Tests Complete ===")
}

func testPostgres(databaseURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Printf("❌ Failed to create pool: %v\n", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Printf("❌ Failed to ping: %v\n", err)
		return
	}

	fmt.Println("✅ PostgreSQL connection successful!")

	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		log.Printf("⚠️  Could not get version: %v", err)
	} else {
		fmt.Printf("📦 PostgreSQL version: %s\n", version[:50])
	}
}

func testSupabaseClient(supabaseURL, supabaseKey string) {
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create Supabase client: %v\n", err)
		return
	}

	// Try to fetch data from a table using PostgREST
	data, count, err := client.From("organizations").Select("*", "exact", false).Execute()
	if err != nil {
		fmt.Printf("❌ Failed to query: %v\n", err)
		fmt.Println("   (Table might not exist yet - run migrations)")
		return
	}

	fmt.Println("✅ Supabase client connection successful!")
	fmt.Printf("📊 Found %d organizations (count: %d)\n", len(data), count)
}
