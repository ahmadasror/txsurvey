package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/ahmadasror/txsurvey/internal/config"
	"github.com/ahmadasror/txsurvey/internal/model"
	"github.com/ahmadasror/txsurvey/pkg/apperror"
)

// googleUserinfoURL is the OIDC userinfo endpoint; var (not const) so tests can
// point it at a stub server.
var googleUserinfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

// UserRepository is the persistence dependency of AuthService (interface for
// testability — concrete impl is repository.UserRepo).
type UserRepository interface {
	UpsertByGoogleSub(ctx context.Context, p model.GoogleProfile) (*model.User, error)
	UpsertByGoogleSubCapped(ctx context.Context, p model.GoogleProfile, maxUsers int) (*model.User, bool, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
}

// AuthService runs the Google sign-in handshake and resolves the creator.
type AuthService struct {
	oauthCfg *oauth2.Config
	states   *stateStore
	users    UserRepository
	maxUsers int
}

func NewAuthService(cfg *config.Config, users UserRepository) *AuthService {
	return &AuthService{
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		states:   newStateStore(10 * time.Minute),
		users:    users,
		maxUsers: cfg.MaxUsers,
	}
}

// AuthURL builds the Google consent URL, generating and storing a fresh state +
// PKCE verifier. The state is the CSRF guard; PKCE binds the code exchange.
func (s *AuthService) AuthURL() string {
	state := randomToken()
	verifier := oauth2.GenerateVerifier()
	s.states.put(state, verifier)
	return s.oauthCfg.AuthCodeURL(state,
		oauth2.AccessTypeOnline,
		oauth2.S256ChallengeOption(verifier),
	)
}

// HandleCallback verifies state, exchanges the code (with PKCE), fetches the
// Google profile, and upserts the creator.
func (s *AuthService) HandleCallback(ctx context.Context, state, code string) (*model.User, error) {
	if state == "" || code == "" {
		return nil, apperror.New(http.StatusBadRequest, "OAUTH_INVALID", "missing state or code")
	}
	verifier, ok := s.states.take(state)
	if !ok {
		return nil, apperror.New(http.StatusBadRequest, "OAUTH_STATE", "invalid or expired login state")
	}

	token, err := s.oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, apperror.New(http.StatusBadRequest, "OAUTH_EXCHANGE", "failed to exchange authorization code")
	}

	profile, err := s.fetchProfile(ctx, token)
	if err != nil {
		return nil, err
	}
	if profile.Sub == "" || profile.Email == "" {
		return nil, apperror.New(http.StatusBadGateway, "OAUTH_PROFILE", "incomplete profile from Google")
	}

	user, capped, err := s.users.UpsertByGoogleSubCapped(ctx, *profile, s.maxUsers)
	if err != nil {
		return nil, err
	}
	if capped {
		return nil, apperror.New(http.StatusForbidden, "REGISTRATION_FULL",
			"sign-ups are full for now — contact the owner")
	}
	return user, nil
}

func (s *AuthService) fetchProfile(ctx context.Context, token *oauth2.Token) (*model.GoogleProfile, error) {
	client := s.oauthCfg.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserinfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build userinfo request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, apperror.New(http.StatusBadGateway, "OAUTH_PROFILE", "failed to reach Google userinfo")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, apperror.New(http.StatusBadGateway, "OAUTH_PROFILE", "Google userinfo returned an error")
	}
	var p model.GoogleProfile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode userinfo: %w", err)
	}
	return &p, nil
}

// CurrentUser returns the signed-in creator by id (nil when not found).
func (s *AuthService) CurrentUser(ctx context.Context, userID string) (*model.User, error) {
	return s.users.GetByID(ctx, userID)
}

// DevLogin upserts a deterministic test creator from an email, bypassing Google.
// It exists ONLY for local/E2E auth and must never be reachable in production
// (the route is mounted only when APP_ENV != "production"; the handler re-checks).
func (s *AuthService) DevLogin(ctx context.Context, email, name string) (*model.User, error) {
	if email == "" {
		return nil, apperror.New(http.StatusBadRequest, "DEV_LOGIN_EMAIL", "email is required")
	}
	if name == "" {
		name = email
	}
	return s.users.UpsertByGoogleSub(ctx, model.GoogleProfile{
		Sub:   "dev|" + email, // stable synthetic subject so reruns reuse one user
		Email: email,
		Name:  name,
	})
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
