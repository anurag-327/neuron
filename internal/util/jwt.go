package util

import (
	"time"

	"github.com/anurag-327/neuron/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserClaims struct {
	Sub   primitive.ObjectID `json:"sub"`
	Email string             `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWTToken(user *models.User, expiresIn time.Duration, secret []byte) (string, error) {
	user.Password = ""
	claims := &UserClaims{
		Sub:   user.ID,
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Issuer:    "Neuron",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ParseJWTToken(tokenString string, secret []byte) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}

func ValidateHeaderToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	const bearerPrefix = "Bearer "
	if len(token) > len(bearerPrefix) && token[:len(bearerPrefix)] == bearerPrefix {
		token = token[len(bearerPrefix):]
		return token, true
	}
	return "", false
}

func ValidateObjectId(id string) (primitive.ObjectID, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return objectId, nil
}
