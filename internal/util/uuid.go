package util

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

func UUID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id, "-", "")
}

func GenerateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, length)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
