package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Invoker abilities with their combinations
var abilities = map[string]string{
	"Cold Snap":       "qqq",
	"Ghost Walk":      "qqw",
	"Ice Wall":        "qqe",
	"EMP":             "www",
	"Tornado":         "wwq",
	"Alacrity":        "wwe",
	"Sun Strike":      "eee",
	"Forge Spirit":    "eeq",
	"Chaos Meteor":    "eew",
	"Deafening Blast": "qwe",
}

// Get all ability names as slice
func getAbilityNames() []string {
	names := make([]string, 0, len(abilities))
	for name := range abilities {
		names = append(names, name)
	}
	return names
}

// App modes
type Mode int

const (
	MenuMode Mode = iota
	TimerMode
	FreestyleMode
)

// Timer durations
type TimerDuration int

const (
	Timer15s TimerDuration = iota
	Timer30s
)

func (t TimerDuration) Duration() time.Duration {
	switch t {
	case Timer15s:
		return 15 * time.Second
	case Timer30s:
		return 30 * time.Second
	default:
		return 15 * time.Second
	}
}

func (t TimerDuration) String() string {
	switch t {
	case Timer15s:
		return "15 seconds"
	case Timer30s:
		return "30 seconds"
	default:
		return "15 seconds"
	}
}

// Model represents the application state
type model struct {
	mode           Mode
	timerDuration  TimerDuration
	currentAbility string
	userInput      string
	score          int
	timeLeft       time.Duration
	gameStarted    bool
	gameOver       bool
	message        string
	abilityNames   []string
	startTime      time.Time
	totalAttempts  int
	correctAnswers int
}

// Messages
type tickMsg time.Time
type gameOverMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case MenuMode:
			return m.handleMenuInput(msg)
		case TimerMode:
			return m.handleTimerInput(msg)
		case FreestyleMode:
			return m.handleFreestyleInput(msg)
		}

	case tickMsg:
		if m.mode == TimerMode && m.gameStarted && !m.gameOver {
			m.timeLeft -= 100 * time.Millisecond
			if m.timeLeft <= 0 {
				m.gameOver = true
				return m, nil
			}
			return m, tick()
		}

	case gameOverMsg:
		m.gameOver = true
	}

	return m, nil
}

func (m model) handleMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "1":
		m.mode = TimerMode
		m.timerDuration = Timer15s
		return m.startGame(), tick()
	case "2":
		m.mode = TimerMode
		m.timerDuration = Timer30s
		return m.startGame(), tick()
	case "3":
		m.mode = FreestyleMode
		return m.startGame(), nil
	}
	return m, nil
}

func (m model) handleTimerInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.resetToMenu(), nil
	case "r":
		if m.gameOver {
			return m.resetToMenu(), nil
		}
		return m.checkAnswer(), nil
	case "backspace":
		if len(m.userInput) > 0 {
			m.userInput = m.userInput[:len(m.userInput)-1]
		}
	default:
		if !m.gameOver && len(msg.String()) == 1 {
			char := strings.ToLower(msg.String())
			if char == "q" || char == "w" || char == "e" {
				if len(m.userInput) >= 3 {
					// Keep only the last 2 characters and add the new one
					m.userInput = m.userInput[1:] + char
				} else {
					m.userInput += char
				}
			}
		}
	}
	return m, nil
}

func (m model) handleFreestyleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.resetToMenu(), nil
	case "r":
		return m.checkAnswer(), nil
	case "backspace":
		if len(m.userInput) > 0 {
			m.userInput = m.userInput[:len(m.userInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			char := strings.ToLower(msg.String())
			if char == "q" || char == "w" || char == "e" {
				if len(m.userInput) >= 3 {
					// Keep only the last 2 characters and add the new one
					m.userInput = m.userInput[1:] + char
				} else {
					m.userInput += char
				}
			}
		}
	}
	return m, nil
}

func (m model) startGame() model {
	m.gameStarted = true
	m.gameOver = false
	m.score = 0
	m.totalAttempts = 0
	m.correctAnswers = 0
	m.userInput = ""
	m.message = ""
	m.startTime = time.Now()

	if m.mode == TimerMode {
		m.timeLeft = m.timerDuration.Duration()
	}

	return m.nextAbility()
}

func (m model) resetToMenu() model {
	return model{
		mode:         MenuMode,
		abilityNames: m.abilityNames,
	}
}

func (m model) nextAbility() model {
	rand.Seed(time.Now().UnixNano())
	m.currentAbility = m.abilityNames[rand.Intn(len(m.abilityNames))]
	m.userInput = ""
	return m
}

func (m model) checkAnswer() model {
	m.totalAttempts++
	correctCombo := abilities[m.currentAbility]

	if strings.ToLower(m.userInput) == correctCombo {
		m.score++
		m.correctAnswers++
		m.message = fmt.Sprintf("✓ Correct! %s = %s", m.currentAbility, correctCombo)

		if m.mode == TimerMode && !m.gameOver {
			return m.nextAbility()
		} else if m.mode == FreestyleMode {
			return m.nextAbility()
		}
	} else {
		m.message = fmt.Sprintf("✗ Wrong! %s = %s (you entered: %s)",
			m.currentAbility, correctCombo, m.userInput)
		m.userInput = ""
	}

	return m
}

func (m model) renderOrbs() string {
	// Orb styles
	quasStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A90E2")).Bold(true)  // Blue
	wexStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6")).Bold(true)   // Purple
	exortStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F1C40F")).Bold(true) // Yellow
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2C2C2C"))            // Dark gray

	orbs := make([]string, 3)

	// Fill orbs based on current input, pad with empty orbs
	for i := 0; i < 3; i++ {
		if i < len(m.userInput) {
			switch m.userInput[i] {
			case 'q':
				orbs[i] = quasStyle.Render("●")
			case 'w':
				orbs[i] = wexStyle.Render("●")
			case 'e':
				orbs[i] = exortStyle.Render("●")
			}
		} else {
			orbs[i] = emptyStyle.Render("○")
		}
	}

	return fmt.Sprintf("%s %s %s", orbs[0], orbs[1], orbs[2])
}

func (m model) View() string {
	var s strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B35")).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	abilityStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ECDC4")).
		Padding(1)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1).
		Width(20)

	correctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#90EE90"))
	wrongStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))

	switch m.mode {
	case MenuMode:
		s.WriteString(titleStyle.Render("🔥 INVOKER ABILITY TRAINER 🔥"))
		s.WriteString("\n\n")
		s.WriteString("Choose your training mode:\n\n")
		s.WriteString("1. Timer Mode - 15 seconds\n")
		s.WriteString("2. Timer Mode - 30 seconds\n")
		s.WriteString("3. Freestyle Mode\n\n")
		s.WriteString("Press 'q' or Ctrl+C to quit")

	case TimerMode:
		if m.gameOver {
			s.WriteString(titleStyle.Render("⏰ TIMER MODE - GAME OVER!"))
			s.WriteString("\n\n")
			s.WriteString(fmt.Sprintf("Time: %s\n", m.timerDuration.String()))
			s.WriteString(fmt.Sprintf("Final Score: %d\n", m.score))
			s.WriteString(fmt.Sprintf("Total Attempts: %d\n", m.totalAttempts))
			if m.totalAttempts > 0 {
				accuracy := float64(m.correctAnswers) / float64(m.totalAttempts) * 100
				s.WriteString(fmt.Sprintf("Accuracy: %.1f%%\n", accuracy))
			}
			s.WriteString("\nPress R to return to menu")
		} else {
			s.WriteString(titleStyle.Render("⏰ TIMER MODE"))
			s.WriteString("\n\n")
			s.WriteString(fmt.Sprintf("Time Left: %.1fs\n", m.timeLeft.Seconds()))
			s.WriteString(fmt.Sprintf("Score: %d\n\n", m.score))

			s.WriteString("Current Ability:\n")
			s.WriteString(abilityStyle.Render(m.currentAbility))
			s.WriteString("\n\n")

			s.WriteString("Orbs: ")
			s.WriteString(m.renderOrbs())
			s.WriteString("\n\n")

			s.WriteString("Enter combination (Q/W/E):\n")
			s.WriteString(inputStyle.Render(strings.ToUpper(m.userInput)))
			s.WriteString("\n\n")

			if m.message != "" {
				if strings.Contains(m.message, "✓") {
					s.WriteString(correctStyle.Render(m.message))
				} else {
					s.WriteString(wrongStyle.Render(m.message))
				}
				s.WriteString("\n\n")
			}

			s.WriteString("Press R to invoke spell, ESC to return to menu")
		}

	case FreestyleMode:
		s.WriteString(titleStyle.Render("🎯 FREESTYLE MODE"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Score: %d/%d", m.correctAnswers, m.totalAttempts))
		if m.totalAttempts > 0 {
			accuracy := float64(m.correctAnswers) / float64(m.totalAttempts) * 100
			s.WriteString(fmt.Sprintf(" (%.1f%%)", accuracy))
		}
		s.WriteString("\n\n")

		s.WriteString("Current Ability:\n")
		s.WriteString(abilityStyle.Render(m.currentAbility))
		s.WriteString("\n\n")

		s.WriteString("Orbs: ")
		s.WriteString(m.renderOrbs())
		s.WriteString("\n\n")

		s.WriteString("Enter combination (Q/W/E):\n")
		s.WriteString(inputStyle.Render(strings.ToUpper(m.userInput)))
		s.WriteString("\n\n")

		if m.message != "" {
			if strings.Contains(m.message, "✓") {
				s.WriteString(correctStyle.Render(m.message))
			} else {
				s.WriteString(wrongStyle.Render(m.message))
			}
			s.WriteString("\n\n")
		}

		s.WriteString("Press R to invoke spell, ESC to return to menu")
	}

	return s.String()
}

func main() {
	initialModel := model{
		mode:         MenuMode,
		abilityNames: getAbilityNames(),
	}

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}
