package main

import (
	"flag"

	req "github.com/nicolas-camacho/should-i-deploy-today-cli/requests"
)

func main() {
	var config req.Config

	flag.StringVar(&config.Tzone, "tz", "UTC", "Timezone to use")
	flag.StringVar(&config.Date, "date", "", "Date to use")

	flag.Parse()

	shouldideploy := req.GetMessage(config)

	println(shouldideploy)
}
