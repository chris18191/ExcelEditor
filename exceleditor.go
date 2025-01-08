package main

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
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
	OutputFile         string
}

var excelEpoch = time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)

func excelDateToDate(excelDate string) time.Time {
	var days, _ = strconv.Atoi(excelDate)
	return excelEpoch.Add(time.Second * time.Duration(days*86400))
}
func dateToExcelDate(date time.Time) string {
	dur := date.Sub(excelEpoch)
	return fmt.Sprint(math.Floor(dur.Abs().Hours()))
}

// func ReadEntryFromRowAt(f *excelize.File, sheet string, row int, col int) (RowEntry, error){
func ReadEntryFromRowAt(currentRow []string, sheet string, rowIdx, colIdx int) (RowEntry, error) {

	var res = RowEntry{}
	res.SheetName = sheet
	res.RowIndex = rowIdx

	//internalOffset := 0
	// slog.Info("Trying to parse following row", "row", currentRow)
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
	tmp, err := strconv.ParseFloat(f, 64)
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

	// formula, err := f.GetCellFormula(sheetName, "G7")
	// if err != nil {
	// 	slog.Error("Could not read cell formula: ", "err", err.Error())
	// }
	// slog.Info("Read formula", "formula", formula)

	// style, err := f.GetCellStyle(sheetName, "G7")
	// if err != nil {
	// 	slog.Error("Could not read cell style", "err", err.Error())
	// }
	// slog.Info("Read style", "style", style)

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

func WriteRowEntry(f *excelize.File, sheetname string, row int, entry RowEntry) {
	// slog.Info("vvvvvvvv")
	// defer slog.Info("^^^^^^^")

	// var res string
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("A%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("B%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("C%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("D%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("E%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("F%d", row))
	// slog.Info("Read value: ", "value", res)
	// res, _ = f.GetCellValue(sheetname, fmt.Sprintf("I%d", row))
	// slog.Info("Read value: ", "value", res)

	// slog.Info("Writing entry", "entry", entry)
	// f.SetCellValue(sheetname, fmt.Sprintf("A%d", row), entry.Date)
	if entry.Start == entry.End {
		for char := range []string{"C", "D", "E", "F", "I", "J"} {
			f.SetCellValue(sheetname, fmt.Sprintf("%c%d", char, row), "")
		}
		f.SetCellFormula(sheetname, fmt.Sprintf("J%d", row), fmt.Sprintf("=IF(D%[1]d>=C%[1]d,(D%[1]d-C%[1]d-E%[1]d)*24)", row))
		return
	}

	// res, err := f.CalcCellValue(sheetname, fmt.Sprintf("B%d", row))
	// if err != nil {
	// 	slog.Error("Failed to calc day for date", "entry", entry)
	// }

	// f.SetCellValue(sheetname, fmt.Sprintf("B%d", row), entry.Day)
	f.SetCellFormula(sheetname, fmt.Sprintf("B%d", row), fmt.Sprintf("=LEFT(TEXT(A%d,\"TTTT\"),2)", row))
	f.UpdateLinkedValue()
	// f.UpdateLinkedValue()
	f.SetCellStr(sheetname, fmt.Sprintf("C%d", row), entry.Start.Format("15:04"))
	f.SetCellValue(sheetname, fmt.Sprintf("D%d", row), entry.End.Format("15:04"))
	if entry.Pause > time.Duration(0) {
		f.SetCellValue(sheetname, fmt.Sprintf("E%d", row), entry.Pause)
	} else {
		f.SetCellValue(sheetname, fmt.Sprintf("E%d", row), "")
	}
	f.SetCellValue(sheetname, fmt.Sprintf("F%d", row), entry.ProjectNr)
	f.SetCellValue(sheetname, fmt.Sprintf("I%d", row), entry.Description)
	// d := entry.End.Sub(entry.Start)
	// hour := int(d.Hours())
	// minute := int(d.Minutes()) % 60
	// f.SetCellValue(sheetname, fmt.Sprintf("J%d", row), fmt.Sprintf("%02d:%02d\n", hour, minute))
	f.SetCellValue(sheetname, fmt.Sprintf("J%d", row), "")
	f.SetCellFormula(sheetname, fmt.Sprintf("J%d", row), fmt.Sprintf("=IF(D%[1]d>=C%[1]d,(D%[1]d-C%[1]d-E%[1]d)*24)", row))
	f.UpdateLinkedValue()

}

// func EraseExtraEntry(f *excelize.File, sheetname string, lastEntry int)

func WriteRowEntries(entries map[string][][]RowEntry, config Configuration) {
	f, err := excelize.OpenFile(config.EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Could not read from file during export", "error", err)
	}

	for sheetname, month := range entries {
		slog.Info("Writing entries for month", "month", sheetname, "#days", len(month))
		var currentRowIndex = config.ROW_ID_ENTRY_START
		for _, day := range month {

			if len(day) == 0 {
				slog.Info("Skipping line", "rowIndex", currentRowIndex)
				currentRowIndex += 1
				continue
			}
			style, _ := f.GetCellStyle(sheetname, fmt.Sprintf("A%d", currentRowIndex))
			writtenDateStr, _ := f.GetCellValue(sheetname, fmt.Sprintf("A%d", currentRowIndex))
			writtenDate := excelDateToDate(writtenDateStr)
			slog.Info("Writing line", "rowIndex", currentRowIndex, "writtenDay", writtenDate, "entryDate", day[0].Date, "style", style)

			if writtenDate.Before(day[0].Date) {
				f.RemoveRow(sheetname, currentRowIndex)
			}

			WriteRowEntry(f, sheetname, currentRowIndex, day[0])
			for _, entry := range day[1:] {
				f.DuplicateRow(sheetname, currentRowIndex)
				currentRowIndex += 1
				writtenDateStr, _ := f.GetCellValue(sheetname, fmt.Sprintf("A%d", currentRowIndex))
				writtenDate = excelDateToDate(writtenDateStr)
				slog.Info("Writing line", "rowIndex", currentRowIndex, "writtenDay", writtenDate, "entryDate", entry.Date, "style", style)
				WriteRowEntry(f, sheetname, currentRowIndex, entry)
			}

			currentRowIndex += 1

		}
	}

	indx, _ := f.GetSheetIndex("Gesamt")

	// Fix formulas in overview
	for row := 4; row <= 15; row++ {
		for _, col := range []string{"F", "G", "H"} {
			sheetNum := row - 3
			cell := fmt.Sprintf("%s%d", col, row)
			formula, _ := f.GetCellFormula("Gesamt", cell)
			newFormula := strings.ReplaceAll(formula, fmt.Sprintf("%02d!", sheetNum), fmt.Sprintf("$'%02d'.", sheetNum))
			newFormula = newFormula[:len(newFormula)-1]
			f.SetCellValue("Gesamt", cell, "")
			f.SetCellFormula("Gesamt", cell, newFormula)
			res, _ := f.GetCellFormula("Gesamt", cell)
			slog.Info("Wrote formula", "cell", cell, "old", formula, "new", res)
			f.UpdateLinkedValue()
		}
	}

	f.SetActiveSheet(indx)
	f.SaveAs(config.OutputFile, excelize.Options{RawCellValue: true})
}

func WriteExampleFile(config Configuration) {
	f, err := excelize.OpenFile(config.EXCEL_FILE, excelize.Options{RawCellValue: true})
	if err != nil {
		slog.Error("Failed to open file", "file", config.EXCEL_FILE)
		return
	}

	f.SetCellValue("01", "C6", "This is a test")
	f.DuplicateRowTo("01", 10, 15)
	f.RemoveRow("01", 20)
	f.SetCellStyle("01", fmt.Sprintf("A%d", config.ROW_ID_ENTRY_START), fmt.Sprintf("A%d", config.ROW_ID_ENTRY_START), 6)

	if err := f.SaveAs("./result.xlsx", excelize.Options{RawCellValue: true}); err != nil {
		slog.Error("Failed to save file", "error", err)
	}

	slog.Info("Save file successfully")
}
