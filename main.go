package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func main() {

	var (
		inputfile   string
		outputfile  string
		debugoutput bool
	)

	flag.StringVar(&inputfile, "in", "test.xlsx", "Excel file to work with")
	flag.StringVar(&outputfile, "out", "out.xlsx", "File to save the results to")
	flag.BoolVar(&debugoutput, "debug", false, "Decides whether debug output should be logged")

	flag.Parse()

	// Create debug output
	var debug_file io.Writer
	debug_file, err := os.OpenFile("./run.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Println("Could not create log file: ", err)
	}

	var loglevel slog.Level
	loglevel = slog.LevelDebug.Level()
	if debugoutput {
		slog.Info("Setting logger to DEBUG level")
		loglevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(debug_file, &slog.HandlerOptions{Level: loglevel})))
	slog.SetLogLoggerLevel(slog.LevelDebug)

	slog.Info("Starting Excel-Editor...")
	var config Configuration = Configuration{
		EXCEL_FILE:          inputfile,
		COL_ID_DATE:         0,
		COL_ID_HOURS_START:  2,
		COL_ID_HOURS_END:    3,
		COL_ID_HOURS_PAUSE:  4,
		ROW_ID_ENTRY_START:  6, // sixth row contains first entries
		OutputFile:          outputfile,
		ProjectNumbersSheet: "Projektnummern",
	}

	slog.Debug("Using config", "config", config)
	Start(config)
}
