package main

import "testing"

func TestRead(t *testing.T) {
	EXCEL_FILE := "./res/test.xlsx"
	t.Log("Trying to read file: ", EXCEL_FILE)

	var testConfig = defaultConfig
	testConfig.EXCEL_FILE = EXCEL_FILE

	res := ReturnAll(testConfig)

	t.Logf("Got result:\n%+v", res)

	return
}
