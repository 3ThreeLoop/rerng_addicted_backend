package main

import (
	"fmt"
	"log"
	"os"
	"rerng_addicted_api/internal/admin/scraping"
	share "rerng_addicted_api/pkg/model"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// 1️⃣ Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set in environment")
	}

	// 2️⃣ Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Connected to database successfully")

	// 3️⃣ Create the repo instance
	scraping := scraping.NewScrapingRepoImpl(db, &share.UserContext{})

	// 4️⃣ Run seeding process
	scraping.Seed()

	fmt.Println("🎬 Done seeding all data.")
}
