package dto

import (
	"time"

	"github.com/anurag-327/neuron/internal/models"
)

type GoogleUserInfo struct {
	Email     string // Primary Google account email
	Picture   string // Profile picture URL
	Name      string // User's display name
	FirstName string // User's first name
	LastName  string // User's last name
	Verified  bool   // True if Google has verified the email
}

type UserDetailsWithAuth struct {
	ID        string          `json:"id"`
	Email     string          `json:"email"`
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	AvatarURL string          `json:"avatarUrl"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Phone     *string         `json:"phone,omitempty"`
	Role      models.RoleType `json:"role"`
	Verified  bool            `json:"verified"`
}
