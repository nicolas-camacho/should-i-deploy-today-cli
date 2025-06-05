package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	Tzone       string
	Date        string
	Interactive bool
}

type Response struct {
	Timezone      string
	Date          string
	Message       string
	Shouldideploy bool
}

type errorResponse struct {
	Message string
	Code    int
	Type    string
}

type badRequest struct {
	Error errorResponse
}

type requestMessage string
type requestError error
type interactiveMode bool

type TimezoneListItem struct {
	Name string
}

func (t TimezoneListItem) Title() string {
	return t.Name
}

func (t TimezoneListItem) Description() string {
	return "Select this timezone"
}

func (t TimezoneListItem) FilterValue() string {
	return t.Name
}

type model struct {
	spinner   spinner.Model
	config    Config
	loading   bool
	message   string
	timezones list.Model
	err       requestError
}

func initialModel(config Config) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	items := []list.Item{
		TimezoneListItem{Name: "UTC"},
		TimezoneListItem{Name: "America/New_York"},
		TimezoneListItem{Name: "Europe/London"},
		TimezoneListItem{Name: "Europe/Paris"},
		TimezoneListItem{Name: "Asia/Tokyo"},
		TimezoneListItem{Name: "Australia/Sydney"},
		TimezoneListItem{Name: "America/Los_Angeles"},
		TimezoneListItem{Name: "America/Chicago"},
		TimezoneListItem{Name: "America/Denver"},
		TimezoneListItem{Name: "Asia/Kolkata"},
		TimezoneListItem{Name: "Asia/Shanghai"},
		TimezoneListItem{Name: "Asia/Singapore"},
		TimezoneListItem{Name: "Europe/Berlin"},
		TimezoneListItem{Name: "Europe/Moscow"},
		TimezoneListItem{Name: "America/Sao_Paulo"},
	}
	return model{
		spinner:   s,
		loading:   true,
		config:    config,
		timezones: list.New(items, list.NewDefaultDelegate(), 0, 0),
	}
}

func fetchMessage(config Config) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second) // Simulate a delay for the spinner

		URL := "https://shouldideploy.today/api" + "?tz=" + config.Tzone

		if config.Date != "" {
			URL += "&date=" + config.Date
		}

		content, err := http.Get(URL)
		if err != nil {
			return requestError(fmt.Errorf("error fetching data: %w", err))
		}

		body, err := io.ReadAll(content.Body)
		if err != nil {
			return requestError(fmt.Errorf("error reading response body: %w", err))
		}

		content.Body.Close()

		if content.StatusCode == http.StatusBadRequest {
			var badReq badRequest
			if err := json.Unmarshal(body, &badReq); err != nil {
				return requestError(fmt.Errorf("error parsing error response: %w", err))
			}
			return requestError(fmt.Errorf("%s (code: %d, type: %s)", badReq.Error.Message, badReq.Error.Code, badReq.Error.Type))
		}

		if content.StatusCode > 299 {
			return requestError(fmt.Errorf("received non-2xx status code: %d, error: %s", content.StatusCode, string(body)))
		}

		var response Response
		json.Unmarshal(body, &response)

		messageStyle := lipgloss.NewStyle().
			Bold(true).
			Width(len(response.Message) + 4).
			PaddingLeft(2).
			PaddingTop(1).
			PaddingBottom(2).
			Foreground(lipgloss.Color("#FAFAFA"))

		if response.Shouldideploy {
			messageStyle = messageStyle.Background(lipgloss.Color("#16C47F"))
		} else {
			messageStyle = messageStyle.Background(lipgloss.Color("#AF3E3E"))
		}

		return requestMessage(messageStyle.Render(response.Message))
	}
}

func (m model) Init() tea.Cmd {

	if m.config.Interactive {
		m.timezones.Title = "Select a timezone (use arrow keys to navigate, enter to select)"
		return func() tea.Msg {
			return interactiveMode(true)
		}
	}

	return tea.Batch(
		m.spinner.Tick,
		fetchMessage(m.config),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		} else if msg.String() == "enter" && m.config.Interactive {
			selectedItem := m.timezones.SelectedItem()
			if item, ok := selectedItem.(TimezoneListItem); ok {
				m.config.Tzone = item.Name
				m.config.Interactive = false
				m.loading = true
				return m, tea.Quit
			}
		}
	case requestMessage:
		m.loading = false
		m.message = string(msg)
		return m, tea.Quit
	case requestError:
		m.loading = false
		m.err = msg
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.timezones.SetSize(msg.Width, msg.Height)
	case interactiveMode:
		m.loading = false
	}

	var cmd tea.Cmd
	m.timezones, cmd = m.timezones.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := ""
	if m.loading {
		s += m.spinner.View() + " should you?..."
	} else if m.config.Interactive {
		return m.timezones.View()
	} else if m.err != nil {
		os.Stderr.WriteString("Error: " + m.err.Error() + "\n")
		os.Exit(1)
	} else {
		s += m.message
	}

	return s
}

func main() {
	var config Config
	var interactiveMode bool

	flag.StringVar(&config.Tzone, "tz", "UTC", "Timezone to use")
	flag.StringVar(&config.Date, "date", "", "Date to use")
	flag.BoolVar(&interactiveMode, "i", false, "Run in interactive mode")

	flag.Parse()

	config.Interactive = interactiveMode

	var p *tea.Program

	if config.Interactive {
		p = tea.NewProgram(initialModel(config), tea.WithAltScreen())
	} else {
		p = tea.NewProgram(initialModel(config))
	}

	finalModel, err := p.Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok && interactiveMode {
		p = tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
}
