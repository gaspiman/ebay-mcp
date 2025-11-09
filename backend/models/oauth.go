package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OAuthClient represents a third-party application that wants to access user data
type OAuthClient struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	ClientSecret string         `gorm:"not null" json:"-"`
	Name         string         `gorm:"not null" json:"name"`
	RedirectURIs string         `gorm:"type:text;not null" json:"redirect_uris"` // JSON array of allowed redirect URIs
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (c *OAuthClient) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// OAuthAuthorizationCode represents a temporary authorization code
type OAuthAuthorizationCode struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex;not null" json:"code"`
	ClientID    string    `gorm:"not null;index" json:"client_id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	RedirectURI string    `gorm:"not null" json:"redirect_uri"`
	Scope       string    `gorm:"type:text" json:"scope"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
	Used        bool      `gorm:"default:false;index" json:"used"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Client OAuthClient `gorm:"foreignKey:ClientID" json:"-"`
	User   User        `gorm:"foreignKey:UserID" json:"-"`
}

// OAuthAccessToken represents an access token for API access
type OAuthAccessToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ClientID  string    `gorm:"not null;index" json:"client_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Scope     string    `gorm:"type:text" json:"scope"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Client OAuthClient `gorm:"foreignKey:ClientID" json:"-"`
	User   User        `gorm:"foreignKey:UserID" json:"-"`
}

// OAuthRefreshToken represents a refresh token for obtaining new access tokens
type OAuthRefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ClientID  string    `gorm:"not null;index" json:"client_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Scope     string    `gorm:"type:text" json:"scope"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Client OAuthClient `gorm:"foreignKey:ClientID" json:"-"`
	User   User        `gorm:"foreignKey:UserID" json:"-"`
}
