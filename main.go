package main

import req "github.com/nicolas-camacho/should-i-deploy-today-cli/requests"

func main() {
	var config req.Config
	shouldideploy := req.GetMessage(config)

	println(shouldideploy)
}
