package handlers

import (
	"sync"
	"time"

	"vpnmanager/internal/model"
)

type CreateDraft struct {
	Name     string
	Server   *model.Server
	Protocol string
}

type EditDraft struct {
	Config *model.ConfigListView
}

type Session struct {
	State     WizardState
	Create    CreateDraft
	Edit      EditDraft
	updatedAt time.Time
}

type Sessions struct {
	mu sync.Mutex
	m  map[int64]*Session
}

func (s *Sessions) Update(userID int64, fn func(*Session)) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.m[userID]
	if !ok {
		sess = &Session{}
		s.m[userID] = sess
	}

	fn(sess)
	sess.updatedAt = time.Now()

	return *sess
}

func (s *Sessions) Get(userID int64) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.m[userID]; ok {
		return *sess
	}

	return Session{}
}
