package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	LOG_FILE = "./debug.log"
)

func initLog() {
	var debug_file io.Writer
	debug_file, err := os.OpenFile(LOG_FILE, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("Could not create log file: ", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(debug_file, &slog.HandlerOptions{Level: slog.LevelDebug})))
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestRead(t *testing.T) {
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)

	f, err := excelize.OpenFile(EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", EXCEL_FILE, "error", err)
		os.Exit(1)
	}

	var testConfig = debugConfig
	testConfig.ExcelFileName = EXCEL_FILE
	testConfig.ExcelFile = f

	res := ReturnAll(testConfig)

	t.Logf("Got result:\n%+v", res)

	return
}

func TestWriteFile(t *testing.T) {

	initLog()
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)
	f, err := excelize.OpenFile(EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", EXCEL_FILE, "error", err)
		os.Exit(1)
	}

	var testConfig = debugConfig
	testConfig.ExcelFileName = EXCEL_FILE
	testConfig.ExcelFile = f
	testConfig.OutputFile = "./res/result.xlsx"

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
	sheets_new["02"] = sheets["02"]
	sheets = sheets_new
	date, err := time.Parse("02/01/2006", "04/01/2025")
	if err != nil {
		slog.Error("Failed to parse date", "error", err)
	}
	sheets["01"][3] = []RowEntry{
		{Day: "Sa", Date: date, Start: time.Now(), End: time.Now().Add(time.Duration(1) * time.Hour), Description: "Test entry", ProjectNr: "2024-1310"},
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

func TestWriteIdenticalFile(t *testing.T) {
	EXCEL_FILE := "./res/test_full.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)
	f, err := excelize.OpenFile(EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", EXCEL_FILE, "error", err)
		os.Exit(1)
	}

	var testConfig = debugConfig
	testConfig.ExcelFileName = EXCEL_FILE
	testConfig.ExcelFile = f
	testConfig.OutputFile = "./res/result_id.xlsx"

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

	WriteRowEntries(sheets, testConfig)

	return
}

func TestGetProjectNumbers(t *testing.T) {
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)
	f, err := excelize.OpenFile(EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", EXCEL_FILE, "error", err)
		os.Exit(1)
	}

	var testConfig = debugConfig
	testConfig.ExcelFileName = EXCEL_FILE
	testConfig.ExcelFile = f
	testConfig.OutputFile = "./res/result_id.xlsx"

	res, _, _ := GetProjectNumbers(testConfig)
	fmt.Println("Read project numbers:\n", res)
}
