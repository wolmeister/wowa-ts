package spinny

import (
	"github.com/gosuri/uilive"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
)

type spinnerStatus int

const (
	// Status constants
	active spinnerStatus = iota
	succeeded
	warning
	failed
	info
	stooped
	removed
)

type Spinner struct {
	text   string
	status spinnerStatus
	mu     *sync.Mutex
}

type Manager struct {
	writer        *uilive.Writer
	frameDuration time.Duration
	runes         []rune

	spinners    []*Spinner
	mu          *sync.Mutex
	ticker      *time.Ticker
	currentRune int
	done        chan struct{}
}

// ManagerOption is a function that configures a Manager.
type ManagerOption func(*Manager)

// WithWriter sets the writer for the Manager.
func WithWriter(out io.Writer) ManagerOption {
	return func(m *Manager) {
		m.writer = uilive.New()
		m.writer.Out = out
	}
}

// WithFrameDuration sets the FrameDuration for the Manager.
func WithFrameDuration(duration time.Duration) ManagerOption {
	return func(m *Manager) {
		m.frameDuration = duration
	}
}

// WithRunes sets the runes for the Manager.
func WithRunes(runes []rune) ManagerOption {
	return func(m *Manager) {
		m.runes = runes
	}
}

// NewManager creates a new Manager with optional configurations.
func NewManager(options ...ManagerOption) *Manager {
	manager := &Manager{
		frameDuration: 80 * time.Millisecond,
		writer:        uilive.New(),
		mu:            &sync.Mutex{},
		done:          make(chan struct{}),
		currentRune:   -1,
		runes: []rune{
			'⠋',
			'⠙',
			'⠹',
			'⠸',
			'⠼',
			'⠴',
			'⠦',
			'⠧',
			'⠇',
			'⠏',
		},
	}

	for _, option := range options {
		option(manager)
	}

	return manager
}

func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ticker != nil {
		// Already started
		return
	}

	m.ticker = time.NewTicker(m.frameDuration)
	go func() {
		for {
			select {
			case <-m.ticker.C:
				m.render()
			case <-m.done:
				m.render()
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	//defer m.mu.Unlock()

	if m.ticker != nil {
		m.ticker.Stop()
		m.ticker = nil
		close(m.done)
		m.mu.Unlock()
		m.render()
	} else {
		m.mu.Unlock()
	}

}

func (m *Manager) NewSpinner(text string) *Spinner {
	m.mu.Lock()
	defer m.mu.Unlock()

	spinner := &Spinner{
		text:   text,
		status: active,
		mu:     &sync.Mutex{},
	}
	m.spinners = append(m.spinners, spinner)
	return spinner
}

func (s *Spinner) Text(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.text = text
}

func (s *Spinner) Stop(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = stooped
	if text != "" {
		s.text = text
	}
}

func (s *Spinner) Succeed(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = succeeded
	if text != "" {
		s.text = text
	}
}

func (s *Spinner) Warn(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = warning
	if text != "" {
		s.text = text
	}
}

func (s *Spinner) Info(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = info
	if text != "" {
		s.text = text
	}
}

func (s *Spinner) Fail(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = failed
	if text != "" {
		s.text = text
	}
}

func (s *Spinner) Remove() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = removed
}

func (m *Manager) nextRune() rune {
	nextRune := m.currentRune + 1
	if nextRune >= len(m.runes) {
		nextRune = 0
	}
	m.currentRune = nextRune
	return m.runes[nextRune]
}

func (m *Manager) render() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var builder strings.Builder

	nextRune := m.nextRune()

	for _, spinner := range m.spinners {
		spinner.mu.Lock()

		switch spinner.status {
		case active:
			builder.WriteRune(nextRune)
			builder.WriteRune(' ')
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case succeeded:
			builder.WriteString(ansiGreen)
			builder.WriteString("✔ ")
			builder.WriteString(ansiReset)
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case warning:
			builder.WriteString(ansiYellow)
			builder.WriteString("⚠ ")
			builder.WriteString(ansiReset)
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case failed:
			builder.WriteString(ansiRed)
			builder.WriteString("✕ ")
			builder.WriteString(ansiReset)
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case info:
			builder.WriteString(ansiBlue)
			builder.WriteString("ℹ ")
			builder.WriteString(ansiReset)
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case stooped:
			builder.WriteString(spinner.text)
			builder.WriteString("\n")
		case removed:
			//	Do nothing
		}

		spinner.mu.Unlock()
	}

	_, _ = m.writer.Write([]byte(builder.String()))
	_ = m.writer.Flush()
}
