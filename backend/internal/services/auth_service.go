package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"onechat/internal/models"
)

type AuthService struct {
	db        *gorm.DB
	jwtSecret string
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(phone, username, password string) (*models.User, string, string, error) {
	// Check if user exists
	var existingUser models.User
	if err := s.db.Where("phone = ? OR username = ?", phone, username).First(&existingUser).Error; err == nil {
		return nil, "", "", errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	// Create user
	user := &models.User{
		Phone:    phone,
		Username: username,
		Password: string(hashedPassword),
		Status:   "Hey there! I'm using OneChat",
		IsOnline: true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, "", "", err
	}

	// Generate tokens
	accessToken, err := s.generateToken(user.ID, user.Phone, 24*time.Hour)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateToken(user.ID, user.Phone, 7*24*time.Hour)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) Login(phone, password string) (*models.User, string, string, error) {
	var user models.User
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid credentials")
	}

	// Update online status
	now := time.Now()
	user.IsOnline = true
	user.LastSeen = &now
	s.db.Save(&user)

	// Generate tokens
	accessToken, err := s.generateToken(user.ID, user.Phone, 24*time.Hour)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.generateToken(user.ID, user.Phone, 7*24*time.Hour)
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(oldToken string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(oldToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	// Generate new access token
	return s.generateToken(claims.UserID, claims.Phone, 24*time.Hour)
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) UpdateProfile(userID uint, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) SearchUsers(query string, currentUserID uint) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("(username LIKE ? OR phone LIKE ?) AND id != ?", 
		"%"+query+"%", "%"+query+"%", currentUserID).
		Limit(20).
		Find(&users).Error
	
	return users, err
}

func (s *AuthService) generateToken(userID uint, phone string, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
