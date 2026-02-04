package repository

import (
	"context"
	"sync"

	"github.com/albenik/uber-fx-based-service-example/internal/core/domain"
)

// MemoryUserRepository is an in-memory implementation of UserRepository.
type MemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *MemoryUserRepository) Save(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID] = user
	return nil
}

func (r *MemoryUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *MemoryUserRepository) FindAll(_ context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

func (r *MemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return domain.ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}
