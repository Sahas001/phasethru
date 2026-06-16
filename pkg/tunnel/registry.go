package tunnel

import (
	"errors"
	"sync"

	"github.com/hashicorp/yamux"
)

var (
	// ErrSubdomainConflict is returned when a client tries to request a subdomain that is already active.
	ErrSubdomainConflict = errors.New("subdomain is already in use")
)

// TunnelRegistry maintains a thread-safe registry of active client subdomains to their multiplexed sessions.
type TunnelRegistry struct {
	mu      sync.RWMutex
	tunnels map[string]*yamux.Session
}

// NewTunnelRegistry creates a new instance of TunnelRegistry.
func NewTunnelRegistry() *TunnelRegistry {
	return &TunnelRegistry{
		tunnels: make(map[string]*yamux.Session),
	}
}

// Register registers a new subdomain and its active Yamux session, returning an error if it's already in use.
func (r *TunnelRegistry) Register(subdomain string, session *yamux.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tunnels[subdomain]; exists {
		return ErrSubdomainConflict
	}
	r.tunnels[subdomain] = session
	return nil
}

// Put updates or inserts a subdomain session directly, bypassing the duplicate check.
func (r *TunnelRegistry) Put(subdomain string, session *yamux.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tunnels[subdomain] = session
}

// Unregister removes a subdomain from the registry.
func (r *TunnelRegistry) Unregister(subdomain string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tunnels, subdomain)
}

// Get retrieves the active Yamux session for a subdomain, returning false if it is not registered.
func (r *TunnelRegistry) Get(subdomain string) (*yamux.Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, exists := r.tunnels[subdomain]
	return session, exists
}

// CloseAll closes all active sessions and clears the registry.
func (r *TunnelRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for subdomain, session := range r.tunnels {
		if session != nil {
			_ = session.Close()
		}
		delete(r.tunnels, subdomain)
	}
}
