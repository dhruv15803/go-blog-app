package scripts

import (
	"errors"
	"github.com/dhruv15803/go-blog-app/internal/storage"
	"github.com/dhruv15803/go-blog-app/utils"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type Script struct {
	storage *storage.Storage
}

func NewScript(storage *storage.Storage) *Script {
	return &Script{
		storage: storage,
	}
}

// CreateVerifiedUser password passed in is plain text
func (s *Script) CreateVerifiedUser(email string, password string) (*storage.User, error) {

	userEmail := strings.ToLower(strings.TrimSpace(email))
	userPassword := strings.TrimSpace(password)

	// validate and create verified user
	if userEmail == "" || userPassword == "" {
		return nil, errors.New("email or password is empty")
	}

	if !utils.IsValidEmail(userEmail) {
		return nil, errors.New("invalid email")
	}

	if !utils.IsPasswordStrong(userPassword) {
		return nil, errors.New("weak password")
	}

	//	hash user password (plain text to hashed)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashPasswordStr := string(hashedPassword)

	user, err := s.storage.CreateVerifiedUser(userEmail, hashPasswordStr)
	if err != nil {
		return nil, err
	}

	return user, nil
}
