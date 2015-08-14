package actiontec

// Parsing functions for the refresh status "API" are here.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Fake enums, since Go doesn't support real ones.

type ChannelType int

const (
	Interleaved ChannelType = iota
	FastChannel
)

type State int

const (
	Up State = iota
	EstablishingLink
	Down
)

// Various structures representing the data we get back in a more structured
// form. (No pun intended.)

type LinkFailures struct {
	Power  uint64
	Signal uint64
	Margin uint64
	Train  uint64
}

type Packets struct {
	Count  uint64
	Errors uint64
}

type PacketPair struct {
	Received    Packets
	Transmitted Packets
}

type FloatPair struct {
	Up   float64
	Down float64
}

type UintPair struct {
	Up   uint64
	Down uint64
}

type Rates UintPair

type LineRate struct {
	Rates
	State State
}

type LineStats struct {
	State             State
	Rates             Rates
	SignalNoiseMargin UintPair
	Attenuation       FloatPair
	Retrains          uint64
	Uptime            time.Duration
}

// Interesting bits of the status. I've deliberately omitted things like CRC
// errors, because they're honestly not that interesting to me. There's
// duplication around things like line rates, but this is a relatively close
// mapping to the underlying data structure, which feels like it has grown
// organically rather than anybody ever thinking about a "design".
type Status struct {
	TotalRate          Rates
	SoftwareVersion    string
	LineStats          LineStats
	TotalRetrains      uint64
	Failures           LinkFailures
	UnavailableSeconds time.Duration
	ChannelType        ChannelType
	ModemUptime        time.Duration
	Packets            PacketPair
	LineRates          []LineRate
}

// Given a blob of status data, parse into a status object.
func ParseStatus(input string) (status *Status, err error) {
	status = new(Status)

	fields := strings.Split(input, "+")
	if len(fields) < 27 {
		return nil, fmt.Errorf("Unexpected number of fields: %d", len(fields))
	}

	status.TotalRate.Up, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}

	status.TotalRate.Down, err = strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return nil, err
	}

	status.SoftwareVersion = fields[3]

	status.LineStats, err = stringToLineStats(fields[4])
	if err != nil {
		return nil, err
	}

	status.TotalRetrains, err = strconv.ParseUint(fields[5], 10, 64)
	if err != nil {
		return nil, err
	}

	status.Failures.Power, err = strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		return nil, err
	}

	status.Failures.Signal, err = strconv.ParseUint(fields[7], 10, 64)
	if err != nil {
		return nil, err
	}

	status.Failures.Margin, err = strconv.ParseUint(fields[8], 10, 64)
	if err != nil {
		return nil, err
	}

	status.Failures.Train, err = strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return nil, err
	}

	status.UnavailableSeconds, err = stringSecondsToDuration(fields[10])
	if err != nil {
		return nil, err
	}

	status.ChannelType, err = stringToChannelType(fields[11])
	if err != nil {
		return nil, err
	}

	status.ModemUptime, err = stringSecondsToDuration(fields[12])
	if err != nil {
		return nil, err
	}

	status.Packets, err = stringToPackets(fields[13])
	if err != nil {
		return nil, err
	}

	for i := 25; i < len(fields)-1; i++ {
		rate, err := stringToLineRate(fields[i])
		if err != nil {
			return nil, err
		}

		status.LineRates = append(status.LineRates, rate)
	}

	return status, nil
}

// Lots of internal parsing functions below: the top level status information
// is delimited by + characters, but many of those fields are then themselves
// delimited in various ad hoc ways. These functions take those fields and turn
// them into structured data.

func stringToAttenuation(s string) (pair FloatPair, err error) {
	// Firmware T2200H-31.128L.03 adds a third field, separated by a comma, which
	// right now we're not interested in (it appears to be an attempt to add the
	// encapsulation of line 2, except it's a completely out of range value).
	// We'll ignore it if present.
	fields := strings.Split(strings.Split(s, ",")[0], " /")
	if len(fields) != 2 {
		err = fmt.Errorf("Unexpected number of fields in a pair: %d", len(fields))
		return
	}

	pair.Down, err = stringToAttenuationFloat(fields[0])
	if err != nil {
		return
	}

	pair.Up, err = stringToAttenuationFloat(fields[1])

	return
}

func stringToAttenuationFloat(s string) (f float64, err error) {
	fields := strings.Split(s, ")")
	if len(fields) != 2 {
		err = fmt.Errorf("Unexpected number of fields in attenuation string: %d", len(fields))
		return
	}

	return strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
}

func stringSecondsToDuration(s string) (d time.Duration, err error) {
	seconds, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		d = time.Duration(seconds) * time.Second
	}

	return
}

func stringToChannelType(s string) (ChannelType, error) {
	if s == "0" {
		return Interleaved, nil
	} else if s == "1" {
		return FastChannel, nil
	}

	return Interleaved, fmt.Errorf("Unknown channel type: %s", s)
}

func stringToLineRate(s string) (rate LineRate, err error) {
	fields := strings.Split(s, "|")
	if len(fields) < 3 {
		err = fmt.Errorf("Unexpected number of fields in line rate: %d", len(fields))
		return
	}

	rate.State, err = stringToState(fields[0])
	if err != nil {
		return
	}

	rate.Up, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return
	}

	rate.Down, err = strconv.ParseUint(fields[2], 10, 64)

	return
}

func stringToLineStats(s string) (stats LineStats, err error) {
	fields := strings.Split(s, "|")
	if len(fields) < 8 {
		err = fmt.Errorf("Unexpected number of fields in line stats: %d", len(fields))
		return
	}

	stats.State, err = stringToState(fields[0])
	if err != nil {
		return
	}

	stats.Rates.Down, err = strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return
	}

	stats.Rates.Up, err = strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return
	}

	stats.SignalNoiseMargin, err = stringToUintPair(fields[4])
	if err != nil {
		return
	}

	stats.Attenuation, err = stringToAttenuation(fields[5])
	if err != nil {
		return
	}

	stats.Retrains, err = strconv.ParseUint(fields[6], 10, 64)
	if err != nil {
		return
	}

	stats.Uptime, err = stringSecondsToDuration(fields[7])

	return
}

func stringToPackets(s string) (packets PacketPair, err error) {
	fields := strings.Split(s, "|")

	packets.Received.Count, err = strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return
	}

	packets.Received.Errors, err = strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return
	}

	packets.Transmitted.Count, err = strconv.ParseUint(fields[2], 10, 64)
	if err != nil {
		return
	}

	packets.Transmitted.Errors, err = strconv.ParseUint(fields[3], 10, 64)

	return
}

func stringToState(s string) (State, error) {
	if s == "Up" {
		return Up, nil
	} else if s == "EstablishingLink" {
		return EstablishingLink, nil
	} else if s == "Down" {
		return Down, nil
	}

	return Down, fmt.Errorf("Unknown state: %s", s)
}

func stringToUintPair(s string) (pair UintPair, err error) {
	fields := strings.Split(s, "/")
	if len(fields) != 2 {
		err = fmt.Errorf("Unexpected number of fields in a pair: %d", len(fields))
		return
	}

	pair.Down, err = strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return
	}

	pair.Up, err = strconv.ParseUint(fields[1], 10, 64)

	return
}
