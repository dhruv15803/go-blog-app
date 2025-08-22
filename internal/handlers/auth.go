package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
	AuthUserId = "AuthUserId"
)

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {

	var registerUserPayload RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&registerUserPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(registerUserPayload.Email))
	userPlainTextPassword := strings.TrimSpace(registerUserPayload.Password)

	if userEmail == "" || userPlainTextPassword == "" {
		writeJSONError(w, "email and password are required", http.StatusBadRequest)
		return
	}
	//	check if email and password are valid and strong respectively
	if !isPasswordStrong(userPlainTextPassword) {
		writeJSONError(w, "weak password", http.StatusBadRequest)
		return
	}
	if !isValidEmail(userEmail) {
		writeJSONError(w, "invalid email", http.StatusBadRequest)
		return
	}

	// after checks , check if there is a verified user that already exists by the email
	existingVerifiedUser, err := h.storage.GetVerifiedUserByEmail(userEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get verified user by email: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if existingVerifiedUser != nil {
		writeJSONError(w, "user already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPlainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash password: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	plainTextToken, hashedTokenStr, err := generateToken(32)
	if err != nil {
		log.Printf("failed to generate token: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	inviteExpiration := time.Now().Add(time.Minute * 15)

	//	create user and invite
	user, err := h.storage.CreateUserAndInvite(userEmail, string(hashedPassword), hashedTokenStr, inviteExpiration)
	if err != nil {
		log.Printf("failed to create user: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type VerificationMailData struct {
		Subject       string
		Email         string
		ActivationUrl string
	}
	verificationMailData := VerificationMailData{
		Subject:       "Verify your account",
		Email:         user.Email,
		ActivationUrl: fmt.Sprintf("%s/activate-account/%s", h.clientUrl, plainTextToken),
	}

	//	send verification mail to user
	maxRetries := 3
	retryCount := 0
	isMailSent := false

	for retryCount < maxRetries {
		if err := h.mailer.SendMailFromTemplate(verificationMailData.Email, verificationMailData.Subject, "./templates/verification.html", verificationMailData); err != nil {
			retryCount++
			continue
		}
		isMailSent = true
		break
	}
	if !isMailSent {
		log.Printf("failed to send verification mail: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "registered user successfully", User: *user}, http.StatusCreated); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {

	plainTextToken := chi.URLParam(r, "token")

	hashedToken := sha256.Sum256([]byte(plainTextToken))
	hashedTokenStr := hex.EncodeToString(hashedToken[:])

	user, err := h.storage.ActivateUser(hashedTokenStr)
	if err != nil {
		log.Printf("failed to activate user: %v\n", err)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//	create jwt token with user.Id payload and expiration and set using cookie
	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWT_SECRET)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var sameSiteConfig http.SameSite

	if os.Getenv("GO_ENV") == "development" {
		sameSiteConfig = http.SameSiteLaxMode
	} else {
		sameSiteConfig = http.SameSiteNoneMode
	}

	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		Path:     "/",
		Secure:   os.Getenv("GO_ENV") == "production",
		MaxAge:   60 * 60 * 24 * 2,
		SameSite: sameSiteConfig,
	}

	http.SetCookie(w, &cookie)

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err = writeJSON(w, Response{Success: true, Message: "activated user successfully", User: *user}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {

	var loginUserPayload LoginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&loginUserPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(loginUserPayload.Email))
	userPlainTextPassword := strings.TrimSpace(loginUserPayload.Password)

	if userEmail == "" || userPlainTextPassword == "" {
		writeJSONError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetVerifiedUserByEmail(userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "invalid email or password", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userPlainTextPassword)); err != nil {
		writeJSONError(w, "invalid email or password", http.StatusBadRequest)
	}

	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWT_SECRET)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var sameSiteConfig http.SameSite

	if os.Getenv("GO_ENV") == "development" {
		sameSiteConfig = http.SameSiteLaxMode
	} else {
		sameSiteConfig = http.SameSiteNoneMode
	}

	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		Path:     "/",
		Secure:   os.Getenv("GO_ENV") == "production",
		MaxAge:   60 * 60 * 24 * 2,
		SameSite: sameSiteConfig,
	}

	http.SetCookie(w, &cookie)

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "logged in user successfully", User: *user}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) GetUserHandler(w http.ResponseWriter, r *http.Request) {

	//	authenticated endpoint - auth middleware

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user does not exist", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	type Response struct {
		Success bool         `json:"success"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, User: *user}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := cookie.Value

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return JWT_SECRET, nil
		})
		if err != nil {
			log.Printf("failed to parse token string with secret: %v\n", err)
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			exp, ok := claims["exp"].(float64)
			if !ok {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if time.Now().Unix() > int64(exp) {
				log.Printf("token is expired")
				writeJSONError(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userIdFloat, ok := claims["sub"].(float64)
			if !ok {
				log.Println("failed to parse user id from token")
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
			userId := int(userIdFloat)

			ctx := context.WithValue(r.Context(), AuthUserId, userId)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		} else {
			log.Println("invalid token")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})
}

func (h *Handler) AdminAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//	admin role will be checked after the auth middleware decodes and validates the jwt

		userId, ok := r.Context().Value(AuthUserId).(int)
		if !ok {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		user, err := h.storage.GetUserById(userId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {

				writeJSONError(w, "user does not exist", http.StatusBadRequest)
				return

			} else {

				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if user.Role != storage.RoleAdmin {
			writeJSONError(w, "user does not have admin role", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isPasswordStrong(password string) bool {
	//	strong password characteristics:
	//	1] minimum length = 6
	//	2] atleast one special character
	//	3] atleast 1 uppercase character
	//	4] atleast 1 lowercase char
	//	5] atleast 1 numerical char

	if utf8.RuneCountInString(password) < 6 {
		return false
	}

	const SPECIAL_CHARS = "!@#$%^&*()_+-="
	const NUMERICAL_CHARS = "0123456789"
	const UPPERCASE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const LOWERCASE_CHARS = "abcdefghijklmnopqrstuvwxyz"
	hasSpecialChar := false
	hasNumericalChar := false
	hasUpperChar := false
	hasLowerChar := false

	for _, c := range password {
		if hasSpecialChar && hasNumericalChar && hasUpperChar && hasLowerChar {
			return true
		}

		if !hasSpecialChar && strings.Contains(SPECIAL_CHARS, string(c)) {
			hasSpecialChar = true
		}

		if !hasNumericalChar && strings.Contains(NUMERICAL_CHARS, string(c)) {
			hasNumericalChar = true
		}

		if !hasUpperChar && strings.Contains(UPPERCASE_CHARS, string(c)) {
			hasUpperChar = true
		}

		if !hasLowerChar && strings.Contains(LOWERCASE_CHARS, string(c)) {
			hasLowerChar = true
		}
	}

	return hasSpecialChar && hasNumericalChar && hasUpperChar && hasLowerChar
}

func isValidEmail(email string) bool {

	if !strings.Contains(email, "@") {
		return false
	}

	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return false
	}
	firstPart, secondPart := emailParts[0], emailParts[1]
	if firstPart == "" || secondPart == "" {
		return false
	}

	if !strings.Contains(secondPart, ".") {
		return false
	}

	if len(strings.Split(secondPart, ".")) != 2 {
		return false
	}

	return true
}

func generateToken(n int) (string, string, error) {
	token := make([]byte, 32)
	_, err := rand.Read(token)
	if err != nil {
		return "", "", err
	}

	plainTextToken := hex.EncodeToString(token)
	hashedToken := sha256.Sum256([]byte(plainTextToken))
	hashedTokenStr := hex.EncodeToString(hashedToken[:])

	return plainTextToken, hashedTokenStr, nil
}
