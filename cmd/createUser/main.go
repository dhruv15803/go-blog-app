package main

import (
	"errors"
	"flag"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/dhruv15803/go-blog-app/scripts"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

type dbConfig struct {
	dbConnStr string
}

// executable to create a user for the go-blog-app
func loadPostgresDbConfig() (*dbConfig, error) {

	godotenv.Load()

	dbConnStr := os.Getenv("POSTGRES_DB_CONN")
	if dbConnStr == "" {
		return nil, errors.New("POSTGRES_DB_CONN env variable not set")
	}

	return &dbConfig{
		dbConnStr: dbConnStr,
	}, nil
}

func main() {
	cfg, err := loadPostgresDbConfig()
	if err != nil {
		log.Fatal(err)
	}

	// database connection to go-blog-db (required)
	db, err := connectToPostgresDb(cfg.dbConnStr)
	if err != nil {
		log.Fatalf("Error connecting to postgres db: %v\n", err)
	}
	//	create a storage layer instance
	storage := storage.NewStorage(db)
	//	pass storage layer instance to scripts instance
	scripts := scripts.NewScript(storage)

	//	get the email and plain text password from the CLI args
	emailPtr := flag.String("email", "", "user email")
	passwordPtr := flag.String("password", "", "user password")
	flag.Parse()

	email := *emailPtr
	password := *passwordPtr

	//	run the scripts.CreateVerifiedUser() func
	user, err := scripts.CreateVerifiedUser(email, password)
	if err != nil {
		log.Fatalf("Error creating verified user: %v\n", err)
	}

	log.Printf("Created User: %v\n", user)
}

func connectToPostgresDb(dbConnStr string) (*sqlx.DB, error) {

	db, err := sqlx.Open("postgres", dbConnStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
