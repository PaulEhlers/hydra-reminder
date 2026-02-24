package timer

import (
	"log"
	"sync"
	"time"
)

type State int

const (
	StateStopped State = iota
	StateRunning
	StateAlerting
)

type Manager struct {
	mu        sync.Mutex
	state     State
	timer     *time.Timer
	duration  time.Duration
	startTime time.Time
	onStart   func()
	onAlert   func()
	onStop    func()
}

func NewManager(onStart func(), onAlert func(), onStop func()) *Manager {
	return &Manager{
		state:   StateStopped,
		onStart: onStart,
		onAlert: onAlert,
		onStop:  onStop,
	}
}

func (m *Manager) Start(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.duration = d
	m.stopInternal()

	m.state = StateRunning
	m.startTime = time.Now()

	// Create a new timer
	m.timer = time.AfterFunc(m.duration, func() {
		m.triggerAlert()
	})

	log.Printf("Timer started for %v", m.duration)
	if m.onStart != nil {
		m.onStart()
	}
}

func (m *Manager) triggerAlert() {
	m.mu.Lock()
	if m.state != StateRunning {
		m.mu.Unlock()
		return
	}
	m.state = StateAlerting
	m.mu.Unlock()

	if m.onAlert != nil {
		m.onAlert()
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopInternal()
	m.state = StateStopped

	if m.onStop != nil {
		m.onStop()
	}
}

// Reset will restart the timer with the current duration, turning off any alert state.
func (m *Manager) Reset() {
	m.mu.Lock()
	d := m.duration
	// Prevent resetting if we haven't ever set a duration
	if d == 0 {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()

	m.Start(d)
}

func (m *Manager) Toggle() {
	m.mu.Lock()
	state := m.state
	d := m.duration
	m.mu.Unlock()

	if state == StateRunning || state == StateAlerting {
		m.Stop()
	} else if d > 0 {
		m.Start(d)
	}
}

// GetState returns current timer state
func (m *Manager) GetState() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// TimeRemaining returns the time left before alert
func (m *Manager) TimeRemaining() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state != StateRunning {
		return 0
	}
	remaining := m.duration - time.Since(m.startTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Internal func, assumes lock is held
func (m *Manager) stopInternal() {
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
}
