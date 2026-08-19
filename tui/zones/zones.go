// Package zones keeps mouse hit testing behind a small presentation boundary.
package zones

import (
	"sync"

	tea "charm.land/bubbletea/v2"
)

// Zone describes a rectangular interactive region.
type Zone struct {
	ID     string
	StartX int
	StartY int
	EndX   int
	EndY   int
}

// InBounds reports whether a mouse event is inside the zone.
func (z Zone) InBounds(msg tea.MouseMsg) bool {
	if msg == nil || z.StartX > z.EndX || z.StartY > z.EndY {
		return false
	}
	point := msg.Mouse()
	return point.X >= z.StartX && point.X <= z.EndX && point.Y >= z.StartY && point.Y <= z.EndY
}

// Pos returns coordinates relative to the top-left corner, or (-1, -1) when
// the event is outside the zone.
func (z Zone) Pos(msg tea.MouseMsg) (int, int) {
	if !z.InBounds(msg) {
		return -1, -1
	}
	point := msg.Mouse()
	return point.X - z.StartX, point.Y - z.StartY
}

// Manager stores zones for one rendered view. It is safe to use from a view
// renderer and an update loop running on separate goroutines.
type Manager struct {
	mu    sync.RWMutex
	zones map[string]Zone
}

// New creates an empty manager.
func New() *Manager { return &Manager{zones: make(map[string]Zone)} }

// Set records a zone, replacing any earlier zone with the same ID.
func (m *Manager) Set(zone Zone) {
	if m == nil || zone.ID == "" {
		return
	}
	m.mu.Lock()
	if m.zones == nil {
		m.zones = make(map[string]Zone)
	}
	m.zones[zone.ID] = zone
	m.mu.Unlock()
}

// Replace replaces all recorded zones with zones.
func (m *Manager) Replace(zones []Zone) {
	if m == nil {
		return
	}
	next := make(map[string]Zone, len(zones))
	for _, zone := range zones {
		if zone.ID != "" {
			next[zone.ID] = zone
		}
	}
	m.mu.Lock()
	m.zones = next
	m.mu.Unlock()
}

// Clear removes all zones.
func (m *Manager) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.zones = make(map[string]Zone)
	m.mu.Unlock()
}

// Get returns a zone by ID.
func (m *Manager) Get(id string) (Zone, bool) {
	if m == nil {
		return Zone{}, false
	}
	m.mu.RLock()
	zone, ok := m.zones[id]
	m.mu.RUnlock()
	return zone, ok
}

// Hit returns the first zone containing msg, ordered by ID for deterministic
// behavior when zones overlap.
func (m *Manager) Hit(msg tea.MouseMsg) (Zone, bool) {
	if m == nil {
		return Zone{}, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var hit Zone
	found := false
	for _, zone := range m.zones {
		if zone.InBounds(msg) && (!found || zone.ID < hit.ID) {
			hit, found = zone, true
		}
	}
	return hit, found
}
