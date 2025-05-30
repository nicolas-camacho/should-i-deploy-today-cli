package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Config struct {
	Tzone string
	Date  string
}

type Response struct {
	Timezone      string
	Date          string
	Message       string
	Shouldideploy bool
}

type model struct {
	spinner spinner.Model
	config  Config
	loading bool
	message string
}

type requestMessage string

func initialModel(config Config) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner: s,
		loading: true,
		config:  config,
	}
}

func fetchMessage(config Config) tea.Cmd {
	return func() tea.Msg {
		URL := "https://shouldideploy.today/api" + "?tz=" + config.Tzone

		if config.Date != "" {
			URL += "&date=" + config.Date
		}

		content, err := http.Get(URL)
		if err != nil {
			return requestMessage("Error: " + err.Error())
		}

		body, err := io.ReadAll(content.Body)
		if err != nil {
			return requestMessage("Error: " + err.Error())
		}

		content.Body.Close()

		if content.StatusCode > 299 {
			return requestMessage("The main request failed with status code: " + strconv.Itoa(content.StatusCode) + " and body: " + string(body))
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
	case requestMessage:
		m.loading = false
		m.message = string(msg)
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	s := ""
	if m.loading {
		s += m.spinner.View() + " should you?..."
	} else {
		s += m.message
	}

	return s
}

func main() {
	var config Config

	flag.StringVar(&config.Tzone, "tz", "UTC", "Timezone to use")
	flag.StringVar(&config.Date, "date", "", "Date to use")

	flag.Parse()

	p := tea.NewProgram(initialModel(config))
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
