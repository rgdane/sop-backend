package service

import (
	"os"

	"jk-api/internal/database/models"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/errors/bcrypt_err"
	"jk-api/pkg/errors/gorm_err"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(email, password string) (*models.User, error)
	GetProfile(token string) (*models.User, string, error)
	GenerateToken(userID int64, name string) (string, error)
	DecodeToken(token string) (jwt.MapClaims, error)
}

type authService struct {
	repo sql.UserRepository
}

func NewAuthService(userRepo sql.UserRepository) *authService {
	return &authService{repo: userRepo}
}

func (s *authService) Login(email, password string) (*models.User, error) {
	user, err := s.repo.
		WithPreloads("HasRoles.HasPermissions").
		FindUserByEmail(email)

	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, bcrypt_err.TranslateBcryptError(err)
	}

	return user, nil
}

func (s *authService) GetProfile(token string) (user *models.User, workspace string, err error) {
	claims, err := s.DecodeToken(token)
	if err != nil {
		return nil, "", err
	}
	userID, ok := claims["user_id"].(float64)

	if !ok {
		return nil, "", err
	}

	user, err = s.repo.WithPreloads("HasRoles.HasPermissions", "HasTitle.HasPosition.HasDivision").FindUserByID(int64(userID))
	if err != nil {
		return nil, "", err
	}

	return user, workspace, nil
}

func (s *authService) GenerateToken(userID int64, name string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"name":    name,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func (s *authService) DecodeToken(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	return claims, err
}
