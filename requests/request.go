package requests

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type Config struct {
	Tzone string `default:"UTC"`
	Date  string `default:""`
}

type Response struct {
	Timezone      string
	Date          string
	Message       string
	Shouldideploy bool
}

func GetMessage(config Config) string {
	URL := "https://shouldideploy.today/api" + "?tz=" + config.Tzone + "&date="
	if config.Date != "" {
		URL += config.Date
	}

	content, err := http.Get(URL)
	if err != nil {
		return "Error: " + err.Error()
	}

	body, err := io.ReadAll(content.Body)
	if err != nil {
		return "Error: " + err.Error()
	}

	content.Body.Close()

	if content.StatusCode > 299 {
		return "The main request failed with status code: " + strconv.Itoa(content.StatusCode) + " and body: " + string(body)
	}

	var response Response
	json.Unmarshal(body, &response)

	return response.Message
}
