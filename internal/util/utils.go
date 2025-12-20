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
