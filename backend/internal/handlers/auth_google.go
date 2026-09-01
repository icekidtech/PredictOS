package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"predictos-backend/internal/config"
	"predictos-backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoogleHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewGoogleHandler(db *gorm.DB, cfg *config.Config) *GoogleHandler {
	return &GoogleHandler{db: db, cfg: cfg}
}

// GET /api/v1/auth/google/login — redirect to Google OAuth
func (h *GoogleHandler) Login(c *fiber.Ctx) error {
	if h.cfg.GoogleClientID == "" {
		return c.Status(500).JSON(fiber.Map{"error": "google oauth not configured"})
	}
	state := uuid.NewString()
	params := url.Values{
		"client_id":     {h.cfg.GoogleClientID},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}
	c.Cookie(&fiber.Cookie{Name: "oauth_state", Value: state, Path: "/", MaxAge: 600, HTTPOnly: true})
	return c.Redirect("https://accounts.google.com/o/oauth2/v2/auth?"+params.Encode(), 302)
}

// GET /api/v1/auth/google/callback — handle Google redirect
func (h *GoogleHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(400).JSON(fiber.Map{"error": "missing code"})
	}

	// Exchange code for tokens
	tokenResp, err := h.exchangeCode(code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token exchange failed: " + err.Error()})
	}

	// Fetch user info
	profile, err := h.fetchProfile(tokenResp["access_token"].(string))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch profile: " + err.Error()})
	}

	googleID, _ := profile["sub"].(string)
	email, _ := profile["email"].(string)
	name, _ := profile["name"].(string)
	picture, _ := profile["picture"].(string)

	if googleID == "" || email == "" {
		return c.Status(500).JSON(fiber.Map{"error": "incomplete google profile"})
	}

	// Find or create user
	var user models.User
	err = h.db.Where("google_id = ?", googleID).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		// Also check by email — link if exists
		err = h.db.Where("email = ?", email).First(&user).Error
		if err == gorm.ErrRecordNotFound {
			user = models.User{
				BaseModel:    models.BaseModel{ID: uuid.New()},
				Username:     name,
				Email:        email,
				GoogleID:     googleID,
				AvatarURL:    picture,
				AuthProvider: "google",
			}
			if err := h.db.Create(&user).Error; err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error()})
			}
			h.db.Create(&models.UserSettings{BaseModel: models.BaseModel{ID: uuid.New()}, UserID: user.ID})
		} else {
			// Link google to existing email user
			h.db.Model(&user).Updates(map[string]interface{}{
				"google_id": googleID, "avatar_url": picture, "auth_provider": "both",
			})
		}
	} else if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Issue JWT and redirect to frontend
	token, _ := signTokenStatic(user.ID, user.WalletAddress, h.cfg.JWTSecret)
	frontendURL := h.cfg.FrontendURL + "/auth/callback?token=" + token
	return c.Redirect(frontendURL, 302)
}

func (h *GoogleHandler) exchangeCode(code string) (map[string]interface{}, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {h.cfg.GoogleClientID},
		"client_secret": {h.cfg.GoogleClientSecret},
		"redirect_uri":  {h.cfg.GoogleRedirectURL},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (h *GoogleHandler) fetchProfile(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
