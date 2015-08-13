package insights

import (
	"bytes"
	"fmt"
	"net/http"
)

// New Relic Insights provides a REST API for inserting arbitrary events, in
// which case you can basically use it as a simple time series database.
//
// This function will take a JSON blob in the form that the API expects:
// https://docs.newrelic.com/docs/insights/new-relic-insights/adding-querying-data/inserting-custom-events-insights-api
func Insert(account int, apiKey string, events []byte) error {
	client := &http.Client{}

	url := fmt.Sprintf("https://insights-collector.newrelic.com/v1/accounts/%d/events", account)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(events))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insert-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Unexpected HTTP response code: %d", resp.StatusCode)
	}

	// Don't care about the response: if there wasn't an error, it was
	// successful.

	return nil
}
