package actiontec

import (
	"testing"
	"time"
)

func TestStringToAttenuationFloat(t *testing.T) {
	successCases := []struct {
		input string
		f     float64
	}{
		{
			"(DS1)10.0",
			10.0,
		},
		{
			"(US1)0.0",
			0.0,
		},
	}

	for _, c := range successCases {
		f, err := stringToAttenuationFloat(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.f != f {
			t.Errorf("Invalid state: got %v; expected %v", f, c.f)
		}
	}

	errorCases := []string{
		"",
		"foo",
		"0.0",
		"(DS1)",
		"(DS1)foo",
	}

	for _, c := range errorCases {
		_, err := stringToAttenuationFloat(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToChannelType(t *testing.T) {
	successCases := []struct {
		input string
		ct    ChannelType
	}{
		{
			"0",
			Interleaved,
		},
		{
			"1",
			FastChannel,
		},
	}

	for _, c := range successCases {
		ct, err := stringToChannelType(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.ct != ct {
			t.Errorf("Invalid state: got %v; expected %v", ct, c.ct)
		}
	}

	errorCases := []string{
		"",
		"2",
		"foo",
	}

	for _, c := range errorCases {
		_, err := stringToChannelType(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToLineRate(t *testing.T) {
	successCases := []struct {
		input string
		rate  LineRate
	}{
		{
			"Up|100|200",
			LineRate{Rates{100, 200}, Up},
		},
		{
			"Down|0|0",
			LineRate{Rates{0, 0}, Down},
		},
	}

	for _, c := range successCases {
		rate, err := stringToLineRate(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.rate != rate {
			t.Errorf("Invalid pair: got %v; expected %v", rate, c.rate)
		}
	}

	errorCases := []string{
		"",
		"||",
		"0||",
		"Up|0|",
	}

	for _, c := range errorCases {
		_, err := stringToLineRate(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToLineStats(t *testing.T) {
	// Test coverage here is minimal because most of this is delegated down to
	// other functions that are also tested.
	successCases := []struct {
		input string
		stats LineStats
	}{
		{
			"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1|2|1000|0",
			LineStats{
				Up,
				Rates{2000, 10000},
				UintPair{7, 9},
				FloatPair{13.1, 26.6},
				2,
				time.Duration(1000) * time.Second,
			},
		},
		{
			"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1, (US2)70.0 |2|1000|0",
			LineStats{
				Up,
				Rates{2000, 10000},
				UintPair{7, 9},
				FloatPair{13.1, 26.6},
				2,
				time.Duration(1000) * time.Second,
			},
		},
	}

	for _, c := range successCases {
		stats, err := stringToLineStats(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.stats != stats {
			t.Errorf("Invalid pair: got %v; expected %v", stats, c.stats)
		}
	}

	errorCases := []string{
		// Basic malformed strings.
		"",
		"Up",
		"|||||||",
		"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1|2",
		// Invalid formats within fields. Not intended to be exhaustive, but just
		// enough to check error propagation.
		"Foo|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1|2|1000|0",
		"Up||10fo|2000|9/7|(DS1)26.6 /(US1)13.1|2|1000|0",
		"Up|PTM|10000|bar|9/7|(DS1)26.6 /(US1)13.1|2|1000|0",
		"Up|PTM|10000|2000|9-7|(DS1)26.6 /(US1)13.1|bar|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)26.6/(US1)13.1|2|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)xxxx /(US1)13.1|2|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)xxxx|2|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US13.1|2|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1|-1|1000|0",
		"Up|PTM|10000|2000|9/7|(DS1)26.6 /(US1)13.1|2|bar|0",
	}

	for _, c := range errorCases {
		_, err := stringToLineStats(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToPackets(t *testing.T) {
	successCases := []struct {
		input   string
		packets PacketPair
	}{
		{
			"0|0|0|0",
			PacketPair{
				Packets{0, 0},
				Packets{0, 0},
			},
		},
		{
			"20|0|30|0",
			PacketPair{
				Packets{20, 0},
				Packets{30, 0},
			},
		},
		{
			"100|1|200|2",
			PacketPair{
				Packets{100, 1},
				Packets{200, 2},
			},
		},
	}

	for _, c := range successCases {
		packets, err := stringToPackets(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.packets != packets {
			t.Errorf("Invalid pair: got %v; expected %v", packets, c.packets)
		}
	}

	errorCases := []string{
		"",
		"|||",
		"0|||",
		"0|0|0|",
		"foo|0|0|0",
	}

	for _, c := range errorCases {
		_, err := stringToPackets(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToState(t *testing.T) {
	successCases := []struct {
		input string
		state State
	}{
		{
			"Up",
			Up,
		},
		{
			"EstablishingLink",
			EstablishingLink,
		},
		{
			"Down",
			Down,
		},
	}

	for _, c := range successCases {
		state, err := stringToState(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.state != state {
			t.Errorf("Invalid state: got %v; expected %v", state, c.state)
		}
	}

	errorCases := []string{
		"",
		"foo",
	}

	for _, c := range errorCases {
		_, err := stringToState(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}

func TestStringToUintPair(t *testing.T) {
	successCases := []struct {
		input string
		pair  UintPair
	}{
		{
			"0/0",
			UintPair{0, 0},
		},
		{
			"1/2",
			UintPair{2, 1},
		},
		{
			"1000000/4200000",
			UintPair{4200000, 1000000},
		},
	}

	for _, c := range successCases {
		pair, err := stringToUintPair(c.input)

		if err != nil {
			t.Errorf("Got an error when one wasn't expected")
		}

		if c.pair != pair {
			t.Errorf("Invalid pair: got %v; expected %v", pair, c.pair)
		}
	}

	errorCases := []string{
		"",
		"1/",
		"a/b",
		"1/2/3",
		"/",
	}

	for _, c := range errorCases {
		_, err := stringToUintPair(c)

		if err == nil {
			t.Errorf("Expected an error; got none")
		}
	}
}
