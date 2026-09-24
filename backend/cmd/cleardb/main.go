package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	// Strip surrounding quotes if present
	dbURL = strings.Trim(dbURL, "'\"")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	tables := []string{
		"content_chunk_tags",
		"content_tags",
		"content_chunks",
		"ingestion_results",
		"content",
		"spaces",
		"tags",
		"users",
	}

	for _, t := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", t))
		if err != nil {
			log.Printf("WARN truncating %s: %v", t, err)
		} else {
			fmt.Printf("✓ Cleared %s\n", t)
		}
	}

	fmt.Println("\nDatabase cleared. All demo data removed.")
}
