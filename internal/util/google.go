package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/anurag-327/neuron/internal/dto"
	"google.golang.org/api/idtoken"
)

// GoogleUserInfo represents the normalized set of user attributes returned
// by Google after validating an ID token or calling the UserInfo API.

// GetGoogleUserInfo validates a Google-issued token and extracts user
// information. It supports two flows:
//
//  1. ID Token Flow: If the provided token is a Google ID token (JWT),
//     the function validates the token against the given Google client ID
//     and extracts user claims directly from the payload.
//  2. Access Token Flow: If the token is not a valid ID token, the function
//     falls back to treating it as an OAuth2 access token and queries
//     Google's UserInfo endpoint to retrieve user details.
//
// Parameters:
//   - token: The Google-issued token (ID token or access token).
//   - googleClientID: The OAuth2 client ID to validate ID tokens against.
//
// Returns:
//   - *GoogleUserInfo containing the user's email, name, picture, and
//     verification status.
//   - error if the token is invalid, expired, or user info cannot be retrieved.
//
// Example usage:
//
//	user, err := utils.GetGoogleUserInfo(idToken, clientID)
//	if err != nil {
//	    log.Fatalf("failed to validate token: %v", err)
//	}
//	fmt.Println("Logged in as:", user.Email)

func GetGoogleUserInfo(token, googleClientID string) (*dto.GoogleUserInfo, error) {
	// Try ID Token validation first
	payload, err := idtoken.Validate(context.Background(), token, googleClientID)
	if err == nil {
		email, ok := payload.Claims["email"].(string)
		if !ok {
			return nil, errors.New("email claim missing or invalid in ID token")
		}

		// Extract additional fields if present
		picture, _ := payload.Claims["picture"].(string)
		verified, _ := payload.Claims["email_verified"].(bool)
		name, _ := payload.Claims["name"].(string)
		firstName, _ := payload.Claims["given_name"].(string)
		lastName, _ := payload.Claims["family_name"].(string)

		return &dto.GoogleUserInfo{
			Email:     email,
			Picture:   picture,
			Name:      name,
			FirstName: firstName,
			LastName:  lastName,
			Verified:  verified,
		}, nil
	}

	// Fallback: Access token → call Google UserInfo API
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call userinfo endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	var userInfo struct {
		Email     string `json:"email"`
		Picture   string `json:"picture"`
		Name      string `json:"name"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Verified  bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &dto.GoogleUserInfo{
		Email:     userInfo.Email,
		Picture:   userInfo.Picture,
		Name:      userInfo.Name,
		FirstName: userInfo.FirstName,
		LastName:  userInfo.LastName,
		Verified:  userInfo.Verified,
	}, nil
}
