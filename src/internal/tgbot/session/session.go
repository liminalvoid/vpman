package session

import (
	"sync"
	"time"

	"vpnmanager/internal/model"
)

type CreateDraft struct {
	Name     string
	Server   *model.Server
	Protocol string
	Format   string
}

type EditDraft struct {
	Config *model.ConfigListView
}

type EditUserDraft struct {
	User         model.User
	Username     string
	NewKeysLimit int
}

type SetPastDueDraft struct {
	Message string
	Amount int
}

type WizardState int

type Session struct {
	State      WizardState
	Create     CreateDraft
	Edit       EditDraft
	EditUser   EditUserDraft
	SetPastDue SetPastDueDraft
	IsAdmin    bool
	updatedAt  time.Time
}

type Sessions struct {
	mu sync.Mutex
	m  map[int64]*Session
}

const (
	StateIdle WizardState = iota

	StateCreateConfig
	StateCreateConfigEnterName
	StateCreateConfigSelectLocation
	StateCreateConfigSelectProtocol
	StateCreateConfigSelectFormat
	StateCreateConfigPreview
	StateCreateConfigConfirm
	StateCreateConfigEdit

	StateEditConfig
	StateEditConfigSelect
	StateEditConfigEditName
	StateEditConfigEnterName
	StateEditConfigDelete

	StateCreateInviteCode
	StateCreateInviteCodeGenerate

	StateSetPastDue
	StateSetPastDueEnterMessage
	StateSetPastDueAmount
	StateSetPastDueEnterAmount
	StateSetPastDuePreview
	StateSetPastDueConfirm

	StateManageUsers
	StateManageUsersSelect
	StateManageUsersEditKeysAmount
	StateManageUsersEnterKeysAmount

	StateConfirmPayment
)

func NewStore() *Sessions {
	return &Sessions{m: make(map[int64]*Session)}
}

func (s *Sessions) Get(userID int64) Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.m[userID]; ok {
		return *sess
	}

	return Session{}
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
