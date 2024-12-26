package main

import (
	"errors"
	"log/slog"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type RowEntry struct {
	SheetName   string
	RowIndex    int
	Date        time.Time // Date of entry
	Day         string
	Start       time.Time
	End         time.Time
	Pause       time.Duration
	ProjectNr   string
	Project     string
	Customer    string
	Description string
	Hours       time.Duration
	Vacation    time.Duration
	Sickness    time.Duration
	Note        string
	RawRow      []string
	Styles      []excelize.Style
	Formulas    []string
}

type Configuration struct {
	EXCEL_FILE         string
	COL_ID_DATE        int
	COL_ID_HOURS_START int
	COL_ID_HOURS_END   int
	COL_ID_HOURS_PAUSE int

	ROW_ID_ENTRY_START int
}

var excelEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)

func excelDateToDate(excelDate string) time.Time {
	var days, _ = strconv.Atoi(excelDate)
	return excelEpoch.Add(time.Second * time.Duration(days*86400))
}

// func ReadEntryFromRowAt(f *excelize.File, sheet string, row int, col int) (RowEntry, error){
func ReadEntryFromRowAt(currentRow []string, sheet string, rowIdx, colIdx int) (RowEntry, error) {

	var res = RowEntry{}
	res.SheetName = sheet
	res.RowIndex = rowIdx

	//internalOffset := 0
	slog.Debug("Trying to parse following row", "row", currentRow)
	// res.Date, err = time.Parse("01/02/2006", currentRow[colIdx+0])
	// if err != nil {
	//   slog.Error("Could not parse date", "err", err)
	// }
	if len(currentRow) == 0 {
		return RowEntry{}, errors.New("trying to read from empty row")
	}
	res.Date = excelDateToDate(currentRow[colIdx+0])

	res.Day = currentRow[colIdx+1]
	if res.Day == "" {
		slog.Error("No day provided", "day", res.Day)
	}
	if res.Day == "Sa" || res.Day == "So" {
		return res, nil
	}

	//res.Start, err = time.Parse(time.TimeOnly, currentRow[colIdx+2])
	res.Start = calcTimeFromFloat(res.Date, currentRow[colIdx+2])
	res.End = calcTimeFromFloat(res.Date, currentRow[colIdx+3])
	res.Pause = calcDurationFromFloat(currentRow[colIdx+4])

	res.ProjectNr = currentRow[colIdx+5]
	res.Project = currentRow[colIdx+6]
	res.Customer = currentRow[colIdx+7]
	res.Description = currentRow[colIdx+8]
	res.Hours = calcDurationFromFloat(currentRow[colIdx+9])
	if len(currentRow) <= colIdx+10 {
		return res, nil
	}
	res.Vacation = calcDurationFromFloat(currentRow[colIdx+10])
	if len(currentRow) <= colIdx+11 {
		return res, nil
	}
	res.Sickness = calcDurationFromFloat(currentRow[colIdx+11])
	if len(currentRow) <= colIdx+12 {
		return res, nil
	}
	res.Note = currentRow[colIdx+12]

	// cellID := fmt.Sprintf("%c%d", rune(int(colIdx)+internalOffset), row)
	// res.Date = time.Parse("02/01/2016", )

	res.RawRow = currentRow

	return res, nil
}

// func ReadEntryFromRow(f *excelize.File, sheet string, row int) (RowEntry, error){
func ReadEntryFromRow(currentRow []string, sheet string, rowIdx int) (RowEntry, error) {
	return ReadEntryFromRowAt(currentRow, sheet, rowIdx, 0)
}

func calcTimeFromFloat(date time.Time, f string) time.Time {
	if f == "" {
		return date
	}
	tmp, err := strconv.ParseFloat(f, 32)
	if err != nil {
		slog.Error("Failed to parse float: ", "string", f)
	}
	return date.Add(time.Duration(int(tmp*24))*time.Hour + time.Duration(int(math.Mod(tmp*24, 1.0)*60))*time.Minute)
}

func calcDurationFromFloat(f string) time.Duration {
	if f == "" {
		return time.Duration(0)
	}
	tmp, err := strconv.ParseFloat(f, 32)
	if err != nil {
		slog.Error("Failed to parse float: ", "string", f)
	}
	return time.Duration(int(tmp*24))*time.Hour + time.Duration(int(math.Mod(tmp*24, 1.0)*60))*time.Minute
}

func ReturnAll(config Configuration) [][][]RowEntry {
	f, err := excelize.OpenFile(config.EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", config.EXCEL_FILE)
		os.Exit(1)
	}
	sheetMap := f.GetSheetMap()
	sheetNames := make([]string, 0, len(sheetMap))
	for _, v := range sheetMap {
		sheetNames = append(sheetNames, v)
	}
	sort.Strings(sheetNames)

	var allEntries [][][]RowEntry = make([][][]RowEntry, 12)
	for i := 0; i < 12; i++ {
		sheetIndex, err := f.GetSheetIndex(sheetNames[i])
		if err != nil {
			slog.Error("Failed to read sheet name: ", "index", i, "sheetNames", sheetNames)
		}
		allEntries[i] = ReturnMonth(sheetNames[sheetIndex], config)
	}

	return allEntries
}

func ReturnMonth(month string, config Configuration) [][]RowEntry {
	slog.Debug("Reading ", "file", config.EXCEL_FILE)
	f, err := excelize.OpenFile(config.EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to read excel file: ", "err", err)
		os.Exit(1)
	}

	// defer f.Save()
	defer f.Close()

	// sheetName := strconv.Itoa(currentMonth)
	sheetName := month

	//fmt.Println("First entry row: ", rows[ROW_ID_ENTRY_START])
	rows, err := f.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		// slog.Error("Failed to get rows of sheet", "sheet", sheetName, "err", err)
		return [][]RowEntry{}
	}

	formula, err := f.GetCellFormula(sheetName, "G7")
	if err != nil {
		slog.Error("Could not read cell formula: ", "err", err.Error())
	}
	slog.Info("Read formula", "formula", formula)

	style, err := f.GetCellStyle(sheetName, "G7")
	if err != nil {
		slog.Error("Could not read cell style", "err", err.Error())
	}
	slog.Info("Read style", "style", style)

	var rowEntries []RowEntry
	for i, row := range rows[config.ROW_ID_ENTRY_START:] {
		rowEntry, err := ReadEntryFromRow(row, sheetName, i)
		slog.Debug("Parsed row: ", "row", rowEntry, "error", err)
		if err != nil {
			continue
		}
		rowEntries = append(rowEntries, rowEntry)
	}

	var res = make([][]RowEntry, 31)
	for _, entry := range rowEntries {
		res[entry.Date.Day()-1] = append(res[entry.Date.Day()-1], entry)
	}
	return res
}
