package store


import (
	"errors"
	"sync"

	"ticket-system/internal/models"
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrTicketNotFound = errors.New("ticket not found")
)

type Store struct {
	mu sync.RWMutex

	usersByID    map[string]*models.User
	usersByEmail map[string]*models.User
	tickets      map[string]*models.Ticket

	nextUserID   int
	nextTicketID int
}

func New() *Store {
	return &Store{
		usersByID:    make(map[string]*models.User),
		usersByEmail: make(map[string]*models.User),
		tickets:      make(map[string]*models.Ticket),
	}
}

func (s *Store) CreateUser(u *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.usersByEmail[u.Email]; ok {
		return ErrUserExists
	}

	s.nextUserID++
	u.ID = idFromInt("usr", s.nextUserID)

	s.usersByID[u.ID] = u
	s.usersByEmail[u.Email] = u
	return nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Store) GetUserByID(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.usersByID[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Store) CreateTicket(t *models.Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextTicketID++
	t.ID = idFromInt("tkt", s.nextTicketID)
	s.tickets[t.ID] = t
}

func (s *Store) GetTicket(id string) (*models.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tickets[id]
	if !ok {
		return nil, ErrTicketNotFound
	}
	return t, nil
}

func (s *Store) ListTicketsByUser(userID string) []*models.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Ticket, 0)
	for _, t := range s.tickets {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Store) UpdateTicket(t *models.Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[t.ID] = t
}

func idFromInt(prefix string, n int) string {
	const hex = "0123456789abcdef"
	digits := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		digits[i] = hex[n%16]
		n /= 16
	}
	return prefix + "_" + string(digits)
}
