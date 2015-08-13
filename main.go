package main

import (
	"actiontec"
	"bytes"
	"encoding/json"
	"flag"
	"insights"
	"log"
	"time"
)

// These functions are a little Insights-specific, although possibly still
// useful outside that context if you need JSON.
func createEvents(status *actiontec.Status, lines []actiontec.LineStats) ([]byte, error) {
	buffer := bytes.NewBufferString("[")

	for i, line := range lines {
		data, err := lineStatsToJSON(i, &line)
		if err != nil {
			return nil, err
		}

		buffer.Write(data)
		buffer.WriteRune(',')
	}

	data, err := statusToJSON(status)
	if err != nil {
		return nil, err
	}

	buffer.Write(data)
	buffer.WriteRune(']')

	return buffer.Bytes(), nil
}

func lineStatsToJSON(line int, stats *actiontec.LineStats) ([]byte, error) {
	return json.Marshal(struct {
		EventType             string `json:"eventType"`
		Line                  int
		RateUp                uint64
		RateDown              uint64
		SignalNoiseMarginUp   uint64
		SignalNoiseMarginDown uint64
		AttenuationUp         float64
		AttenuationDown       float64
		Retrains              uint64
	}{
		"LineStats",
		line,
		stats.Rates.Up,
		stats.Rates.Down,
		stats.SignalNoiseMargin.Up,
		stats.SignalNoiseMargin.Down,
		stats.Attenuation.Up,
		stats.Attenuation.Down,
		stats.Retrains,
	},
	)
}

func statusToJSON(status *actiontec.Status) ([]byte, error) {
	return json.Marshal(struct {
		EventType string `json:"eventType"`
		RateUp    uint64
		RateDown  uint64
		Retrains  uint64
	}{
		"ModemStats",
		status.TotalRate.Up,
		status.TotalRate.Down,
		status.TotalRetrains,
	},
	)
}

// Command line flags.
var account int
var apiKey string
var host string
var interval int
var password string
var username string

func init() {
	flag.IntVar(&account, "account", 0, "New Relic Insights account number")
	flag.StringVar(&apiKey, "apikey", "", "New Relic Insights API key")
	flag.StringVar(&host, "host", "", "router IP address or host name")
	flag.IntVar(&interval, "interval", 60, "interval between stat gathering (in seconds)")
	flag.StringVar(&password, "password", "", "router admin password")
	flag.StringVar(&username, "username", "admin", "router admin user name")
}

func main() {
	// Parse and check flags.
	flag.Parse()

	if account == 0 {
		log.Fatal("Insights account number must be provided.")
	}

	if apiKey == "" {
		log.Fatal("Insights API key must be provided.")
	}

	if host == "" {
		log.Fatal("Router host name or IP address must be provided.")
	}

	if interval < 30 {
		log.Fatal("Interval must be greater than or equal to 30.")
	}

	if password == "" {
		log.Fatal("Password must be provided.")
	}

	if username == "" {
		log.Fatal("User name must be provided.")
	}

	// Create our context for interacting with the router. Originally, a context
	// was created on each tick, but Go seemed to be unable to GC the open file
	// descriptors for the HTTP client, which is unfortunate.
	ctx, err := actiontec.NewContext(host)
	if err != nil {
		log.Fatalf("Error creating context: %v", err)
	}

	// Set up a ticker and gather and send data each tick. I haven't bothered
	// going this in a goroutine with fancy signal handling: you want to kill it,
	// just kill it via Ctrl-C or kill.
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for _ = range ticker.C {
		log.Print("Gathering data...")

		// We'll re-login every time: it doesn't hurt, and the Actiontec UI seems
		// to base the logout timeout on when you logged in, not your last
		// activity.
		if err := ctx.Login(username, password); err != nil {
			log.Fatalf("Error logging into router: %v", err)
		}

		status, stats, err := ctx.GetStatus()
		if err != nil {
			log.Fatalf("Error getting stats from router: %v", err)
		}

		events, err := createEvents(status, stats)
		if err != nil {
			log.Fatalf("Error buildling JSON: %v", err)
		}

		log.Print("Sending data to Insights...")
		if err := insights.Insert(account, apiKey, events); err != nil {
			log.Fatalf("Error inserting data to Insights: %v", err)
		}

		log.Print("Data inserted.")

		if err := ctx.Logout(); err != nil {
			log.Printf("Error logging out (will attempt to continue): %v", err)
		}
	}
}
