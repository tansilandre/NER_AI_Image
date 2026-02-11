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
)

func main() {
	// Load .env
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	fmt.Println("Testing database connection...")
	fmt.Printf("URL: %s\n", databaseURL[:50]+"...")
	fmt.Println()

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	fmt.Println("✅ Database connection successful!")
	fmt.Println()

	// Query version
	var version string
	if err := pool.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		log.Printf("⚠️  Could not get version: %v", err)
	} else {
		fmt.Printf("PostgreSQL version: %s\n", version[:50])
	}

	// List tables
	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename IN ('organizations', 'profiles', 'generations', 'providers')
	`)
	if err != nil {
		log.Printf("⚠️  Could not list tables: %v", err)
	} else {
		defer rows.Close()
		fmt.Println()
		fmt.Println("Tables found:")
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err == nil {
				fmt.Printf("  - %s\n", table)
			}
		}
	}

	fmt.Println()
	fmt.Println("✅ Database is ready!")
}
