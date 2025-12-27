package util

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsValidObjectID(id string) (primitive.ObjectID, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	return objectId, err
}

func SetCreditsLeftHeader(c *gin.Context, credits int64) {
	c.Header("X-Credit-Balance", fmt.Sprintf("%d", credits))
}

// secureCompare performs constant-time comparison to prevent timing attacks
func SecureCompare(a, b string) bool {
	// Import crypto/subtle at the top of the file
	// Use subtle.ConstantTimeCompare for timing-attack resistance
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
