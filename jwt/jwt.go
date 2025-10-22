package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/Amierza/worker-service/dto"
	"github.com/golang-jwt/jwt/v5"
)

type (
	IJWT interface {
		GenerateToken(userID string, role string) (string, string, error)
		ValidateToken(token string) (*jwt.Token, error)
		GetUserIDByToken(tokenString string) (string, error)
		GetUserRoleByToken(tokenString string) (string, error)
	}

	jwtCustomClaim struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
		jwt.RegisteredClaims
	}

	JWT struct {
		secretKey string
		issuer    string
	}
)

func NewJWT() *JWT {
	return &JWT{
		secretKey: getSecretKey(),
		issuer:    "Template",
	}
}

func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "Template"
	}

	return secretKey
}

func (j *JWT) GenerateToken(userID string, role string) (string, string, error) {
	accessClaims := jwtCustomClaim{
		userID,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 3600 * 24 * 7)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", dto.ErrGenerateAccessToken
	}

	refreshClaims := jwtCustomClaim{
		userID,
		role,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 3600 * 24 * 7)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", "", dto.ErrGenerateRefreshToken
	}

	return accessTokenString, refreshTokenString, nil
}

func (j *JWT) parseToken(t_ *jwt.Token) (any, error) {
	if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, dto.ErrUnexpectedSigningMethod
	}

	return []byte(j.secretKey), nil
}

func (j *JWT) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, j.parseToken)
	if err != nil {
		return nil, err
	}

	return token, err
}

func (j *JWT) GetUserIDByToken(tokenString string) (string, error) {
	token, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", dto.ErrValidateToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", dto.ErrTokenInvalid
	}

	userID := fmt.Sprintf("%v", claims["user_id"])

	return userID, nil
}

func (j *JWT) GetUserRoleByToken(tokenString string) (string, error) {
	token, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", dto.ErrValidateToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", dto.ErrTokenInvalid
	}

	role := fmt.Sprintf("%v", claims["role"])

	return role, nil
}
