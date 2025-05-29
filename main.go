package main

import (
	"flag"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	req "github.com/nicolas-camacho/should-i-deploy-today-cli/requests"
)

func main() {
	var config req.Config

	flag.StringVar(&config.Tzone, "tz", "UTC", "Timezone to use")
	flag.StringVar(&config.Date, "date", "", "Date to use")

	flag.Parse()

	shouldideploy, message := req.GetMessage(config)

	var style = lipgloss.NewStyle().
		Bold(true).
		PaddingLeft(4).
		Width(20)

	if shouldideploy {
		style = style.Foreground(lipgloss.Color("#222831")).Background(lipgloss.Color("#FAFAFA"))
	} else {
		style = style.Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#AF3E3E"))
	}

	fmt.Println(style.Render(message))
}
