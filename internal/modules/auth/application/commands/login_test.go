package commands

import (
	"context"
	"testing"
	"time"

	"github.com/elite4print/elite4print-go/internal/modules/auth/application"
	"github.com/elite4print/elite4print-go/internal/modules/auth/domain"
	identitydomain "github.com/elite4print/elite4print-go/internal/modules/identity/domain"
	"github.com/elite4print/elite4print-go/internal/shared/eventbus"
	"github.com/elite4print/elite4print-go/internal/shared/idgenerator"
	"github.com/elite4print/elite4print-go/internal/shared/password"
	"github.com/elite4print/elite4print-go/internal/shared/token"
	"github.com/elite4print/elite4print-go/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inMemoryUserRepository is a fake for identity users.
type inMemoryUserRepository struct {
	users []*identitydomain.User
}

func (r *inMemoryUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*identitydomain.User, error) {
	for _, u := range r.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) GetByEmail(ctx context.Context, email identitydomain.Email) (*identitydomain.User, error) {
	for _, u := range r.users {
		if u.Email.String() == email.String() {
			return u, nil
		}
	}
	return nil, nil
}

func (r *inMemoryUserRepository) EmailExists(ctx context.Context, email identitydomain.Email) (bool, error) {
	return false, nil
}

func (r *inMemoryUserRepository) Save(ctx context.Context, user *identitydomain.User) error {
	r.users = append(r.users, user)
	return nil
}

func (r *inMemoryUserRepository) Update(ctx context.Context, user *identitydomain.User) error {
	return nil
}

func (r *inMemoryUserRepository) List(ctx context.Context, offset, limit int) ([]*identitydomain.User, int, error) {
	return r.users, len(r.users), nil
}

// inMemorySessionRepository is a fake for sessions.
type inMemorySessionRepository struct {
	sessions []*domain.Session
}

func (r *inMemorySessionRepository) Create(ctx context.Context, s *domain.Session) error {
	r.sessions = append(r.sessions, s)
	return nil
}

func (r *inMemorySessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	for _, s := range r.sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, nil
}

func (r *inMemorySessionRepository) GetByRefreshJTI(ctx context.Context, jti string) (*domain.Session, error) {
	return nil, nil
}

func (r *inMemorySessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	var out []*domain.Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *inMemorySessionRepository) Update(ctx context.Context, s *domain.Session) error {
	for i, existing := range r.sessions {
		if existing.ID == s.ID {
			r.sessions[i] = s
			return nil
		}
	}
	return nil
}

func (r *inMemorySessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	for _, s := range r.sessions {
		if s.ID == id {
			now := time.Now().UTC()
			s.RevokedAt = &now
		}
	}
	return nil
}

func (r *inMemorySessionRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID, except uuid.UUID) error {
	now := time.Now().UTC()
	for _, s := range r.sessions {
		if s.UserID == userID && s.ID != except {
			s.RevokedAt = &now
		}
	}
	return nil
}

// inMemorySessionCache is a fake cache.
type inMemorySessionCache struct {
	data map[uuid.UUID]uuid.UUID
}

func (c *inMemorySessionCache) Set(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, ttlSeconds int) error {
	if c.data == nil {
		c.data = make(map[uuid.UUID]uuid.UUID)
	}
	c.data[sessionID] = userID
	return nil
}

func (c *inMemorySessionCache) Get(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	if uid, ok := c.data[sessionID]; ok {
		return uid, nil
	}
	return uuid.Nil, domain.ErrSessionNotFound
}

func (c *inMemorySessionCache) Delete(ctx context.Context, sessionID uuid.UUID) error {
	delete(c.data, sessionID)
	return nil
}

// inMemoryTokenBlacklist is a fake blacklist.
type inMemoryTokenBlacklist struct {
	items map[string]bool
}

func (b *inMemoryTokenBlacklist) Add(ctx context.Context, jti string, ttlSeconds int) error {
	if b.items == nil {
		b.items = make(map[string]bool)
	}
	b.items[jti] = true
	return nil
}

func (b *inMemoryTokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return b.items[jti], nil
}

func TestLogin_Success(t *testing.T) {
	hasher := password.NewArgon2id()
	hash, err := hasher.Hash("StrongPass123!")
	require.NoError(t, err)

	email, _ := identitydomain.NewEmail("alice@example.com")
	user := &identitydomain.User{
		ID:           idgenerator.NewUUIDv7(),
		Email:        email,
		PasswordHash: password.NewHashed(hash),
		FirstName:    "Alice",
		LastName:     "Smith",
		Role:         identitydomain.RoleCustomer,
		Status:       identitydomain.UserStatusActive,
	}
	userRepo := &inMemoryUserRepository{users: []*identitydomain.User{user}}
	sessionRepo := &inMemorySessionRepository{}
	cache := &inMemorySessionCache{}
	blacklist := &inMemoryTokenBlacklist{}
	tokenizer := token.NewJWT("super-secret-key-at-least-32-characters-long")

	uc := NewLogin(userRepo, sessionRepo, cache, blacklist, tokenizer, hasher, eventbus.Noop{}, validator.New(), LoginConfig{
		AccessTTL:     15 * time.Minute,
		RefreshTTL:    7 * 24 * time.Hour,
		MaxConcurrent: 5,
	})

	result, err := uc.Handle(context.Background(), application.LoginCommand{
		Email:     "alice@example.com",
		Password:  "StrongPass123!",
		IP:        "127.0.0.1",
		UserAgent: "test",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, 1, len(sessionRepo.sessions))
}

func TestLogin_InvalidPassword(t *testing.T) {
	hasher := password.NewArgon2id()
	hash, _ := hasher.Hash("StrongPass123!")

	email, _ := identitydomain.NewEmail("alice@example.com")
	user := &identitydomain.User{
		ID:           idgenerator.NewUUIDv7(),
		Email:        email,
		PasswordHash: password.NewHashed(hash),
		Role:         identitydomain.RoleCustomer,
		Status:       identitydomain.UserStatusActive,
	}
	userRepo := &inMemoryUserRepository{users: []*identitydomain.User{user}}

	uc := NewLogin(userRepo, &inMemorySessionRepository{}, &inMemorySessionCache{}, &inMemoryTokenBlacklist{}, token.NewJWT("secret"), hasher, eventbus.Noop{}, validator.New(), LoginConfig{})

	_, err := uc.Handle(context.Background(), application.LoginCommand{
		Email:    "alice@example.com",
		Password: "WrongPassword",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
