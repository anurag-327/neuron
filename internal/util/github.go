package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type GitHubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Verified  bool   `json:"verified"`
}

type GitHubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

// FetchGitHubUserInfo gets GitHub user info (including verified primary email)
func FetchGitHubUserInfo(accessToken string) (*GitHubUser, error) {
	user, err := getGitHubUser(accessToken)
	if err != nil {
		return nil, err
	}

	data, err := getPrimaryVerifiedEmail(accessToken)
	if err != nil {
		return nil, err
	}
	user.Email = data.Email
	user.Verified = data.Verified

	return user, nil
}

func getGitHubUser(accessToken string) (*GitHubUser, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub /user API returned %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode /user response: %w", err)
	}
	return &user, nil
}

func getPrimaryVerifiedEmail(accessToken string) (*GitHubEmail, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call /emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub /emails API returned %d", resp.StatusCode)
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return nil, fmt.Errorf("failed to decode /emails response: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return &email, nil
		}
	}

	return &emails[0], nil // Return the first email if no primary verified email found
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

// Exchange the code for access token
func GetGitHubAccessToken(code string) (string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("failed to get access token: " + tokenResp.Error)
	}

	return tokenResp.AccessToken, nil
}
