package main

import (
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/handlers"
	"github.com/dhruv15803/go-blog-app/internal/mailer"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type dbConfig struct {
	dbConnStr       string
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifeTime time.Duration
	connMaxIdleTime time.Duration
}

type mailerConfig struct {
	host     string
	port     int
	username string
	password string
}

type config struct {
	addr                string
	readRequestTimeout  time.Duration
	writeRequestTimeout time.Duration
	clientUrl           string
	dbConfig            dbConfig
	mailerConfig        mailerConfig
}

func loadConfig() (*config, error) {

	godotenv.Load()

	port := os.Getenv("PORT")
	dbConnStr := os.Getenv("POSTGRES_DB_CONN")
	mailerHost := os.Getenv("MAILER_HOST")
	mailerPort, err := strconv.ParseInt(os.Getenv("MAILER_PORT"), 10, 64)
	mailerUsername := os.Getenv("MAILER_USERNAME")
	mailerPassword := os.Getenv("MAILER_PASSWORD")
	clientUrl := os.Getenv("CLIENT_URL")
	if err != nil {
		return nil, err
	}
	if port == "" || dbConnStr == "" {
		return nil, errors.New("PORT or POSTGRES_DB_CONN not set")
	}

	cfg := &config{
		addr:                port,
		readRequestTimeout:  time.Second * 15,
		writeRequestTimeout: time.Second * 15,
		clientUrl:           clientUrl,
		dbConfig: dbConfig{
			dbConnStr:       dbConnStr,
			maxOpenConns:    50,
			maxIdleConns:    25,
			connMaxLifeTime: time.Hour,
			connMaxIdleTime: time.Minute * 5,
		},
		mailerConfig: mailerConfig{
			host:     mailerHost,
			port:     int(mailerPort),
			username: mailerUsername,
			password: mailerPassword,
		},
	}

	return cfg, nil
}

func main() {

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading server config: %v\n", err)
	}

	pgConn := newPostgresDbConn(cfg.dbConfig.dbConnStr, cfg.dbConfig.maxOpenConns, cfg.dbConfig.maxIdleConns, cfg.dbConfig.connMaxLifeTime, cfg.dbConfig.connMaxIdleTime)
	db, err := pgConn.connect()
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	defer db.Close()

	mailer := mailer.NewMailer(cfg.mailerConfig.host, cfg.mailerConfig.port, cfg.mailerConfig.username, cfg.mailerConfig.password)

	//layers
	storage := storage.NewStorage(db)
	handler := handlers.NewHandler(storage, mailer, cfg.clientUrl)

	server := newServer(cfg.addr, cfg.readRequestTimeout, cfg.writeRequestTimeout, handler)

	log.Printf("starting server on port %v\n", server.addr)

	if err := server.run(); err != nil {
		log.Fatalf("Error starting server on port %v\n", server.addr)
	}
}
