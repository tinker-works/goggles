// Package zones keeps mouse hit testing behind a small presentation boundary.
package zones

import (
	"sort"
	"strconv"
	"sync"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone"
)

// Zone describes a rectangular interactive region.
type Zone struct {
	ID     string
	StartX int
	StartY int
	EndX   int
	EndY   int
}

var global = New()

const (
	HeaderProject = "header-project"
	ModalSubmit   = "modal-submit"
	ModalCancel   = "modal-cancel"
	ModalBody     = "modal-body"
)

// Init resets the package-level manager used by components that do not own a
// screen-level manager.
func Init() { global = New() }

// Mark registers a package-level marker.
func Mark(id, content string) string { return global.Mark(id, content) }

// Scan removes package-level markers and records their bounds.
func Scan(view string) string { return global.Scan(view) }

// In reports whether a mouse point falls inside the named package-level zone.
func In(mouse tea.Mouse, id string) bool {
	area, ok := global.Get(id)
	if !ok {
		return false
	}
	return mouse.X >= area.StartX && mouse.X <= area.EndX &&
		mouse.Y >= area.StartY && mouse.Y <= area.EndY
}

// Bounds returns the package-level zone origin as x, y, and ok.
func Bounds(id string) (int, int, bool) {
	zone, ok := global.Get(id)
	if !ok {
		return 0, 0, false
	}
	return zone.StartX, zone.StartY, true
}

// EpicTreeRow is the stable ID for an issue-tree row.
func EpicTreeRow(index int) string { return "epic-tree-row-" + strconv.Itoa(index) }

func ModalField(index int) string  { return "modal-field-" + strconv.Itoa(index) }
func ModalOption(index int) string { return "modal-option-" + strconv.Itoa(index) }
func ModalChoice(index int) string { return "modal-choice-" + strconv.Itoa(index) }

func AttentionRow(index int) string { return "attention-" + strconv.Itoa(index) }

func ProjectRow(index int) string { return "project-row-" + strconv.Itoa(index) }

func OrganisationRow(index int) string { return "organisation-row-" + strconv.Itoa(index) }

func BoardEpicCard(index int) string { return "board-epic-" + strconv.Itoa(index) }

func BoardIssueCard(lane, column, index int) string {
	return "board-issue-" + strconv.Itoa(lane) + "-" + strconv.Itoa(column) + "-" + strconv.Itoa(index)
}

func BoardRunnerRow(index int) string { return "board-runner-" + strconv.Itoa(index) }

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
	mu sync.RWMutex

	// zones contains zones registered directly with Set. Rendered zones are
	// kept separately so Scan can replace only the zones found in the view.
	zones    map[string]Zone
	rendered map[string]Zone

	// markers maps BubbleZone's generated marker sequences back to the IDs
	// exposed by this package.
	markers map[string]string

	scanMu  sync.Mutex
	manager *zone.Manager
}

// New creates an empty manager.
func New() *Manager {
	return &Manager{
		zones:    make(map[string]Zone),
		rendered: make(map[string]Zone),
		markers:  make(map[string]string),
		manager:  zone.New(),
	}
}

// Close stops the marker scanner's worker.
func (m *Manager) Close() {
	if m == nil {
		return
	}
	m.scanMu.Lock()
	if m.manager != nil {
		m.manager.Close()
	}
	m.scanMu.Unlock()
}

// Mark wraps content in a zero-width marker for id. The markers are removed
// by Scan after their coordinates have been recorded.
func (m *Manager) Mark(id, content string) string {
	if m == nil || id == "" || content == "" {
		return content
	}

	m.scanMu.Lock()
	defer m.scanMu.Unlock()
	m.initialize()
	marked := m.manager.Mark(id, content)
	if len(marked) <= len(content) {
		return marked
	}

	markerSize := (len(marked) - len(content)) / 2
	marker := marked[:markerSize]
	if marker == "" || len(marked) != len(content)+markerSize*2 || marked[len(marked)-markerSize:] != marker {
		return marked
	}

	m.mu.Lock()
	m.markers[marker] = id
	m.mu.Unlock()
	return marked
}

// Scan removes markers from view and records the zones they delimit. It is
// intended to wrap the outermost rendered view.
func (m *Manager) Scan(view string) string {
	if m == nil {
		return view
	}

	m.scanMu.Lock()
	defer m.scanMu.Unlock()
	m.initialize()
	clean := m.manager.Scan(view)
	rendered := m.scan(view)

	m.mu.Lock()
	m.rendered = rendered
	m.mu.Unlock()
	return clean
}

func (m *Manager) initialize() {
	if m.manager != nil {
		return
	}
	m.manager = zone.New()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.zones == nil {
		m.zones = make(map[string]Zone)
	}
	if m.rendered == nil {
		m.rendered = make(map[string]Zone)
	}
	if m.markers == nil {
		m.markers = make(map[string]string)
	}
}

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
	m.rendered = make(map[string]Zone)
	m.mu.Unlock()
}

// Clear removes all zones.
func (m *Manager) Clear() {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.zones = make(map[string]Zone)
	m.rendered = make(map[string]Zone)
	m.mu.Unlock()
}

// Get returns a zone by ID.
func (m *Manager) Get(id string) (Zone, bool) {
	if m == nil {
		return Zone{}, false
	}
	m.mu.RLock()
	zone, ok := m.rendered[id]
	if !ok {
		zone, ok = m.zones[id]
	}
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
	candidates := make([]Zone, 0, len(m.rendered)+len(m.zones))
	for _, zone := range m.rendered {
		candidates = append(candidates, zone)
	}
	for id, zone := range m.zones {
		if _, rendered := m.rendered[id]; !rendered {
			candidates = append(candidates, zone)
		}
	}
	m.mu.RUnlock()

	var hit Zone
	found := false
	for _, zone := range candidates {
		if zone.InBounds(msg) && (!found || zone.ID < hit.ID) {
			hit, found = zone, true
		}
	}
	return hit, found
}

type marker struct {
	value string
	id    string
}

func (m *Manager) scan(view string) map[string]Zone {
	m.mu.RLock()
	markers := make([]marker, 0, len(m.markers))
	for value, id := range m.markers {
		markers = append(markers, marker{value: value, id: id})
	}
	m.mu.RUnlock()
	sort.Slice(markers, func(i, j int) bool {
		return len(markers[i].value) > len(markers[j].value)
	})

	active := make(map[string]Zone)
	result := make(map[string]Zone)
	line := ""
	y := 0
	for pos := 0; pos < len(view); {
		if matched, ok := scanMarker(view[pos:], markers); ok {
			x := ansi.StringWidth(line)
			if start, exists := active[matched.id]; exists {
				start.EndX = x - 1
				start.EndY = y
				result[start.ID] = start
				delete(active, matched.id)
			} else {
				active[matched.id] = Zone{ID: matched.id, StartX: x, StartY: y}
			}
			pos += len(matched.value)
			continue
		}

		r, width := utf8.DecodeRuneInString(view[pos:])
		if r == '\n' {
			line = ""
			y++
		} else {
			line += view[pos : pos+width]
		}
		pos += width
	}
	return result
}

func scanMarker(view string, markers []marker) (marker, bool) {
	for _, marker := range markers {
		if len(view) >= len(marker.value) && view[:len(marker.value)] == marker.value {
			return marker, true
		}
	}
	return marker{}, false
}
