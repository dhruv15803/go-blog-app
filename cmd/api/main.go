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

type redisConfig struct {
	addr     string //"HOSTNAME:PORT"
	password string
}

type mailerConfig struct {
	host     string
	port     int
	username string
	password string
}

type cloudinaryConfig struct {
	cloudinaryUrl string
}

type config struct {
	addr                string
	readRequestTimeout  time.Duration
	writeRequestTimeout time.Duration
	clientUrl           string
	dbConfig            dbConfig
	mailerConfig        mailerConfig
	redisConfig         redisConfig
	cloudinaryConfig    cloudinaryConfig
}

func loadConfig() (*config, error) {

	godotenv.Load()

	port := os.Getenv("PORT")
	dbConnStr := os.Getenv("POSTGRES_DB_CONN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	mailerHost := os.Getenv("MAILER_HOST")
	mailerPort, err := strconv.ParseInt(os.Getenv("MAILER_PORT"), 10, 64)
	mailerUsername := os.Getenv("MAILER_USERNAME")
	mailerPassword := os.Getenv("MAILER_PASSWORD")
	clientUrl := os.Getenv("CLIENT_URL")
	cloudinaryUrl := os.Getenv("CLOUDINARY_URL")
	if err != nil {
		return nil, err
	}
	if port == "" || dbConnStr == "" {
		return nil, errors.New("PORT or POSTGRES_DB_CONN not set")
	}
	if redisAddr == "" || redisPassword == "" {
		return nil, errors.New("REDIS_ADDR or REDIS_PASSWORD not set")
	}
	if cloudinaryUrl == "" {
		return nil, errors.New("CLOUDINARY_URL not set")
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
		redisConfig: redisConfig{
			addr:     redisAddr,
			password: redisPassword,
		},
		cloudinaryConfig: cloudinaryConfig{
			cloudinaryUrl: cloudinaryUrl,
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

	redisConn := newRedisConn(cfg.redisConfig.addr, cfg.redisConfig.password)
	redisClient, err := redisConn.createRedisInstance()
	if err != nil {
		log.Fatalf("Error creating redis client: %v\n", err)
	}

	cld, err := newCloudinaryApi(cfg.cloudinaryConfig.cloudinaryUrl).createInstance()
	if err != nil {
		log.Fatalf("Error creating cloudinary client: %v\n", err)
	}

	mailer := mailer.NewMailer(cfg.mailerConfig.host, cfg.mailerConfig.port, cfg.mailerConfig.username, cfg.mailerConfig.password)

	//layers
	storage := storage.NewStorage(db)
	handler := handlers.NewHandler(storage, mailer, redisClient, cld, cfg.clientUrl)

	server := newServer(cfg.addr, cfg.readRequestTimeout, cfg.writeRequestTimeout, handler)

	log.Printf("starting server on port %v\n", server.addr)

	if err := server.run(); err != nil {
		log.Fatalf("Error starting server on port %v\n", server.addr)
	}
}
