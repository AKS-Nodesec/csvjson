package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	out, errx := exec.Command("wmic", "cpu", "get", "caption,", "deviceid,", "name,", "numberofcores,", "maxclockspeed,", "status", "/format:csv").Output()
	if errx != nil {
		log.Fatal(errx)
	}

	strTok := strings.Split(string(out), "\r")

	outStr := ""
	for _, i := range strTok {
		outStr += i + "\n"
	}

	err2 := ioutil.WriteFile("data.csv", []byte(outStr), 0)

	if err2 != nil {
		log.Fatal(err2)
	}

	data, err := readAndParseCsv("data.csv")
	if err != nil {
		panic(fmt.Sprintf("error while handling csv file: %s\n", err))
	}

	json, err := csvToJson(data)
	if err != nil {
		panic(fmt.Sprintf("error while converting csv to json file: %s\n", err))
	}

	errj := ioutil.WriteFile("data.json", []byte(json), 0)

	if errj != nil {
		log.Fatal(errj)
	}

}

func readAndParseCsv(path string) ([][]string, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error opening %s", path)
	}

	var rows [][]string

	reader := csv.NewReader(csvFile)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return rows, fmt.Errorf("failed to parse csv: %s", err)
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func csvToJson(rows [][]string) (string, error) {
	var entries []map[string]interface{}
	attributes := rows[0]
	for _, row := range rows[1:] {
		entry := map[string]interface{}{}
		for i, value := range row {
			attribute := attributes[i]
			// split csv header key for nested objects
			objectSlice := strings.Split(attribute, ".")
			internal := entry
			for index, val := range objectSlice {
				// split csv header key for array objects
				key, arrayIndex := arrayContentMatch(val)
				if arrayIndex != -1 {
					if internal[key] == nil {
						internal[key] = []interface{}{}
					}
					internalArray := internal[key].([]interface{})
					if index == len(objectSlice)-1 {
						internalArray = append(internalArray, value)
						internal[key] = internalArray
						break
					}
					if arrayIndex >= len(internalArray) {
						internalArray = append(internalArray, map[string]interface{}{})
					}
					internal[key] = internalArray
					internal = internalArray[arrayIndex].(map[string]interface{})
				} else {
					if index == len(objectSlice)-1 {
						internal[key] = value
						break
					}
					if internal[key] == nil {
						internal[key] = map[string]interface{}{}
					}
					internal = internal[key].(map[string]interface{})
				}
			}
		}
		entries = append(entries, entry)
	}

	bytes, err := json.MarshalIndent(entries, "", "	")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func arrayContentMatch(str string) (string, int) {
	i := strings.Index(str, "[")
	if i >= 0 {
		j := strings.Index(str, "]")
		if j >= 0 {
			index, _ := strconv.Atoi(str[i+1 : j])
			return str[0:i], index
		}
	}
	return str, -1
}
