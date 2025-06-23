package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubUser represents a GitHub user from the API
type GitHubUser struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Email     *string `json:"email"`
	Name      *string `json:"name"`
	AvatarURL string  `json:"avatar_url"`
}

// OAuthManager handles GitHub OAuth authentication
type OAuthManager struct {
	config *oauth2.Config
}

// NewOAuthManager creates a new OAuth manager
func NewOAuthManager(clientID, clientSecret, redirectURL string) *OAuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}

	return &OAuthManager{
		config: config,
	}
}

// GetAuthURL returns the GitHub OAuth authorization URL
func (om *OAuthManager) GetAuthURL(state string) string {
	return om.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (om *OAuthManager) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := om.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}

// GetUserInfo retrieves user information from GitHub using the access token
func (om *OAuthManager) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
	client := om.config.Client(ctx, token)
	
	// Get user info from GitHub API
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	// If email is null, try to get it from the emails endpoint
	if user.Email == nil {
		email, err := om.getPrimaryEmail(ctx, client)
		if err == nil && email != "" {
			user.Email = &email
		}
	}

	return &user, nil
}

// getPrimaryEmail retrieves the user's primary email from GitHub
func (om *OAuthManager) getPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get emails, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	// Find the primary email
	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	// If no primary email found, return the first one
	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}