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
		Width(len(message) + 4).
		PaddingLeft(2).
		PaddingTop(1).
		PaddingBottom(1).
		Foreground(lipgloss.Color("#FAFAFA"))

	if shouldideploy {
		style = style.Background(lipgloss.Color("#16C47F"))
	} else {
		style = style.Background(lipgloss.Color("#AF3E3E"))
	}

	fmt.Println(style.Render(message))
}
