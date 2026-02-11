// Database migration runner
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	fmt.Println("=== NER Studio Database Migration ===")
	fmt.Println()
	fmt.Printf("Connecting to: %s...\n", databaseURL[:50])

	// Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("❌ Failed to ping: %v", err)
	}

	fmt.Println("✅ Connected!")
	fmt.Println()

	// Find migration files
	migrationsDir := "supabase/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try from apps/api directory
		migrationsDir = "../../supabase/migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("❌ Cannot read migrations dir: %v", err)
	}

	fmt.Printf("Found %d migration files\n", len(files))
	fmt.Println()

	// Run each migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		fmt.Printf("→ Running %s... ", file.Name())

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			fmt.Printf("❌ Read error: %v\n", err)
			continue
		}

		// Execute migration
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			// Check if it's a "already exists" error
			errStr := err.Error()
			if strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "duplicate") ||
				strings.Contains(errStr, "42601") {
				fmt.Println("⚠️  Already applied")
			} else {
				fmt.Printf("❌ Error: %v\n", err)
			}
			continue
		}

		fmt.Println("✅ Success")
	}

	fmt.Println()
	fmt.Println("=== Migration Complete ===")
	fmt.Println()

	// Verify tables
	fmt.Println("Verifying tables...")
	rows, err := pool.Query(ctx, `
		SELECT tablename FROM pg_tables 
		WHERE schemaname = 'public'
		ORDER BY tablename
	`)
	if err != nil {
		log.Printf("⚠️  Could not list tables: %v", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var table string
			if err := rows.Scan(&table); err == nil {
				fmt.Printf("  ✓ %s\n", table)
				count++
			}
		}
		fmt.Printf("\nTotal tables: %d\n", count)
	}

	// Wait for user input
	fmt.Println()
	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
