package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)

	var testConfig = defaultConfig
	testConfig.EXCEL_FILE = EXCEL_FILE

	res := ReturnAll(testConfig)

	t.Logf("Got result:\n%+v", res)

	return
}

func TestWriteFile(t *testing.T) {
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)

	var testConfig = defaultConfig
	testConfig.EXCEL_FILE = EXCEL_FILE

	res := ReturnAll(testConfig)
	var sheets = make(map[string][][]RowEntry)
	for _, month := range res {
		if len(month) > 0 {
			if len(month[0]) > 0 {
				sheets[month[0][0].SheetName] = month
			} else if len(month[2]) > 0 {
				sheets[month[2][0].SheetName] = month
			}
		}
	}

	sheets_new := make(map[string][][]RowEntry)
	sheets_new["01"] = sheets["01"]
	sheets = sheets_new
	date, err := time.Parse("02/01/2006", "04/01/2025")
	if err != nil {
		slog.Error("Failed to parse date", "error", err)
	}
	sheets["01"][3] = []RowEntry{
		{Day: "Sa", Date: date, Start: time.Now(), End: time.Now().Add(time.Duration(1) * time.Hour), Description: "Test entry"},
		{Day: "Sa", Date: date, Start: time.Now().Add(time.Duration(1) * time.Hour), End: time.Now().Add(time.Duration(2) * time.Hour), Description: "Test entry2"},
	}

	date, err = time.Parse("02/01/2006", "08/01/2025")
	if err != nil {
		slog.Error("Failed to parse date", "error", err)
	}
	sheets["01"][7] = []RowEntry{
		{Day: "Mi", Date: date, Start: time.Now(), End: time.Now().Add(time.Duration(1) * time.Hour), Description: "Test entry"},
		{Day: "Mi", Date: date, Start: time.Now().Add(time.Duration(1) * time.Hour), End: time.Now().Add(time.Duration(2) * time.Hour), Description: "Test entry2"},
	}
	WriteRowEntries(sheets, testConfig)

	return
}
