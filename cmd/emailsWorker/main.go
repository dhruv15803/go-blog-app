package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/mailer"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

// this emailsWorker will run alongside the main REST API service
// it will pull of emailData from the emails queue(redis list) and process them one
// by one off the queue , if queue is empty, it will block until new emailData is ready

// this worker should connect to the same redis instance the server is working with

type mailerConfig struct {
	host     string
	port     int
	username string
	password string
}

type redisConfig struct {
	addr     string
	password string
}

type config struct {
	redisConfig  redisConfig
	mailerConfig mailerConfig
}

func loadConfig() (*config, error) {

	godotenv.Load()

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	mailerHost := os.Getenv("MAILER_HOST")
	mailerPort, err := strconv.ParseInt(os.Getenv("MAILER_PORT"), 10, 64)
	mailerUsername := os.Getenv("MAILER_USERNAME")
	mailerPassword := os.Getenv("MAILER_PASSWORD")
	if err != nil {
		return nil, err
	}

	if redisAddr == "" || redisPassword == "" {
		return nil, errors.New("REDIS_ADDR or REDIS_PASSWORD is empty")
	}

	redisCfg := redisConfig{
		addr:     redisAddr,
		password: redisPassword,
	}
	mailerCfg := mailerConfig{
		host:     mailerHost,
		port:     int(mailerPort),
		username: mailerUsername,
		password: mailerPassword,
	}

	return &config{redisConfig: redisCfg, mailerConfig: mailerCfg}, nil
}

type VerificationMailData struct {
	Subject       string `json:"subject"`
	Email         string `json:"email"`
	ActivationUrl string `json:"activation_url"`
}

const (
	MAX_RETRIES_PER_EMAIL = 3
)

func main() {

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("failed to load redis config: %v\n", err)
	}

	redisClient, err := newRedisConn(cfg.redisConfig.addr, cfg.redisConfig.password).createRedisInstance()
	if err != nil {
		log.Fatalf("failed to create redis worker instance: %v\n", err)
	}

	mailer := mailer.NewMailer(cfg.mailerConfig.host, cfg.mailerConfig.port, cfg.mailerConfig.username, cfg.mailerConfig.password)

	log.Println("Email Worker started!")
	for {
		var emailDataFromRedis VerificationMailData

		emailDataArr, err := redisClient.BRPop(context.Background(), 0, "emails").Result()
		if err != nil {
			log.Println("Invalid BRPop response")
			continue
		}
		emailDataStr := emailDataArr[1] // (actual stringified email data in json)
		emailDataJsonBytes := []byte(emailDataStr)

		if err := json.Unmarshal(emailDataJsonBytes, &emailDataFromRedis); err != nil {
			redisClient.LPush(context.Background(), "emails-dead", emailDataStr)
			continue
		}

		isEmailSent := false
		for i := 0; i < MAX_RETRIES_PER_EMAIL; i++ {

			if err := mailer.SendMailFromTemplate(emailDataFromRedis.Email, emailDataFromRedis.Subject, "./templates/verification.html", emailDataFromRedis); err != nil {
				log.Printf("failed to send verification email to %s , attempt=%d", emailDataFromRedis.Email, i+1)
				continue
			}
			isEmailSent = true
			break
		}

		if !isEmailSent {
			redisClient.LPush(context.Background(), "emails-dead", emailDataStr)
			continue
		}

		log.Printf("Verification Email Successfully sent, Email:%s", emailDataFromRedis.Email)
	}
}
