package services

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JwtService struct {
	secretKey string
	issuer    string
}

type jwtCustomClaim struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJwtService() *JwtService {
	return &JwtService{
		secretKey: "loremipsum",
		issuer:    "loremipsum",
	}
}

func (s *JwtService) GenerateToken(userID uuid.UUID) string {
	claims := &jwtCustomClaim{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 120)),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		slog.Error("could not create jwt signed string",
			"error", err,
		)
	}

	return t
}

func (s *JwtService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t_ *jwt.Token) (any, error) {
		if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t_.Header["alg"])
		}

		return []byte(s.secretKey), nil
	})
}

func (s *JwtService) GetUserIDByToken(token string) (uuid.UUID, error) {
	t_Token, err := s.ValidateToken(token)
	if err != nil {
		return uuid.Nil, err
	}
	claims := t_Token.Claims.(jwt.MapClaims)
	id := fmt.Sprintf("%v", claims["user_id"])
	userId, _ := uuid.Parse(id)

	return userId, nil
}
