package websocket

import (
	"sync"

	"github.com/pafthang/arcanum/services/gate/internal/models"
)

// ACL wraps models.WSACL for the websocket package.
type ACL struct {
	inner *models.WSACL
}

// NewACL creates an ACL.
func NewACL() *ACL {
	return &ACL{inner: models.NewWSACL()}
}

// Set updates permissions.
func (a *ACL) Set(entry *models.WSACLEntry) {
	a.inner.Set(entry)
}

// Allow checks access.
func (a *ACL) Allow(subject, conn string) bool {
	return a.inner.Allow(subject, conn)
}

// Get returns the entry.
func (a *ACL) Get(subject string) (*models.WSACLEntry, bool) {
	return a.inner.Get(subject)
}

// Remove deletes the entry.
func (a *ACL) Remove(subject string) {
	a.inner.Remove(subject)
}

// ConnLimiter limits concurrent connections per subject.
type ConnLimiter struct {
	mu    sync.Mutex
	count map[string]int
	acl   *ACL
}

// NewConnLimiter creates a limiter.
func NewConnLimiter(acl *ACL) *ConnLimiter {
	return &ConnLimiter{
		count: make(map[string]int),
		acl:   acl,
	}
}

// Acquire tries to take a slot. Returns false if the limit is exceeded.
func (l *ConnLimiter) Acquire(subject string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.acl.Get(subject)
	max := 0
	if ok {
		max = entry.MaxConnections
	}
	if max > 0 && l.count[subject] >= max {
		return false
	}
	l.count[subject]++
	return true
}

// Release frees a slot.
func (l *ConnLimiter) Release(subject string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count[subject] > 0 {
		l.count[subject]--
	}
	if l.count[subject] == 0 {
		delete(l.count, subject)
	}
}
