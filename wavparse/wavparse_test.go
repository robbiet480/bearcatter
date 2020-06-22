package wavparse

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/stretchr/testify/assert"
)

type WavPlayerTime struct {
	time.Time
}

const wavPlayerTimeFormat = "1/02/2006 03:04:05 PM"

// Convert the internal date as CSV string
func (date *WavPlayerTime) MarshalCSV() (string, error) {
	return date.Time.Format(wavPlayerTimeFormat), nil
}

// Convert the CSV string as internal date
func (date *WavPlayerTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(wavPlayerTimeFormat, csv)
	return err
}

type WavPlayerDuration struct {
	time.Duration
}

// Convert the internal duration as CSV string
func (clock *WavPlayerDuration) MarshalCSV() (string, error) {
	return clock.Duration.String(), nil
}

// Convert the CSV string as internal duration
func (clock *WavPlayerDuration) UnmarshalCSV(csv string) (err error) {
	split := strings.Split(csv, ":")
	clock.Duration, err = time.ParseDuration(fmt.Sprintf("%sh%sm%ss", split[0], split[1], split[2]))
	return err
}

type WavPlayerEntry struct {
	FilePath       string            `csv:"File path"`
	FileName       string            `csv:"File name"`
	Product        string            `csv:"Scanner type"`
	DateAndTime    WavPlayerTime     `csv:"Date and time"`
	Duration       WavPlayerDuration `csv:"Duration"`
	ScanMode       string            `csv:"Scan mode"`
	SystemType     string            `csv:"Type"`
	Frequency      float64           `csv:"Frequency"`
	Code           string            `csv:"Code"`
	FavoriteName   string            `csv:"Favorite name"`
	SystemName     string            `csv:"System name"`
	DepartmentName string            `csv:"Department name"`
	ChannelName    string            `csv:"Channel name"`
	SiteName       string            `csv:"Site"`
	TGID           string            `csv:"TGID"`
	UnitID         int64             `csv:"UID"`
	UnitIDName     string            `csv:"UID Name"`
	Latitude       float64           `csv:"Latitude"`
	Longitude      float64           `csv:"Longitude"`
}

func TestDecodeRecording(t *testing.T) {
	testCaseFile, openErr := os.OpenFile("fixtures.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if openErr != nil {
		panic(openErr)
	}
	defer testCaseFile.Close()

	testCases := []*WavPlayerEntry{}

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		return r
	})

	if unmarshalErr := gocsv.UnmarshalFile(testCaseFile, &testCases); unmarshalErr != nil { // Load WavPlayerEntry from file
		panic(unmarshalErr)
	}

	for _, testCase := range testCases {
		if testCase == nil {
			t.Log("Refusing to run a nil parsed fixture")
			continue
		}
		t.Run(testCase.FileName, testDecodeRecordingCase(fmt.Sprintf("fixtures/%s", testCase.FileName), *testCase))
	}
}

func testDecodeRecordingCase(path string, expected WavPlayerEntry) func(t *testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)

		parsed, parsedErr := DecodeRecording(path)
		if parsedErr != nil {
			t.Fatalf("error when parsing file: %v", parsedErr)
		}

		assert.Equal(expected.FileName, parsed.File, "File names should be equal")

		assert.Equal(parsed.Duration, expected.Duration.Duration, "Duration should be equal")

		assert.Equal(expected.Product, parsed.Public.Product, "Products (public) should be equal")

		assert.Equal(parsed.Public.Timestamp, &expected.DateAndTime.Time, "Timestamps (public) should be equal")

		assert.Equal(expected.SystemType, parsed.Private.System.Type, "System types (private) should be equal")

		assert.Equal(expected.Frequency, parsed.Private.Frequency, "Frequencies (private) should be equal")

		assert.Equal(expected.FavoriteName, parsed.Public.FavoriteListName, "Favorite List Names (public) should be equal")
		assert.Equal(expected.FavoriteName, parsed.Private.FavoriteList.Name, "Favorite List Names (private) should be equal")

		assert.Equal(expected.SystemName, parsed.Public.System, "System Names (public) should be equal")
		assert.Equal(expected.SystemName, parsed.Private.System.Name, "System Names (private) should be equal")

		assert.Equal(expected.DepartmentName, parsed.Public.Department, "Department Names (public) should be equal")
		assert.Equal(expected.DepartmentName, parsed.Private.Department, "Department Names (private) should be equal")

		assert.Equal(expected.ChannelName, parsed.Public.Channel, "Channel Names (public) should be equal")
		assert.Equal(expected.ChannelName, parsed.Private.Channel, "Channel Names (private) should be equal")

		assert.Equal(expected.SiteName, parsed.Private.Site.Name, "Site Names (private) should be equal")

		assert.Equal(expected.TGID, parsed.Public.TGIDFreq, "TGID (public) should be equal")
		assert.Equal(expected.TGID, parsed.Private.TGID, "TGID (private) should be equal")

		assert.Equal(expected.UnitID, parsed.Public.UnitID, "UnitID (public) should be equal")
		assert.Equal(expected.UnitID, parsed.Private.UnitID, "UnitID (private) should be equal")

		assert.Equal(expected.Latitude, parsed.Private.Location.Latitude, "Latitude (private) should be equal")
		assert.Equal(expected.Longitude, parsed.Private.Location.Longitude, "Longitude (private) should be equal")
	}
}
