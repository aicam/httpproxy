package server

import (
	"encoding/json"
	"github.com/aicam/jsonconfig"
	"io/ioutil"
	"os"
)

func ReadConfig(filename string) string {
	file, _ := ioutil.ReadFile(filename)
	return string(file)
}

func WriteConfig(filename string, body jsonconfig.Configuration) {
	os.Remove(filename)
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	data, _ := json.Marshal(body)
	_, _ = file.Write(data)
	file.Close()
}

func GetInfo(categories map[uint]string, filename string, configuration jsonconfig.Configuration) []byte {
	file, _ := ioutil.ReadFile(filename)
	type Log struct {
		Host       string `json:"host"`
		Path       string `json:"path"`
		Fragment   string `json:"fragment"`
		CategoryID uint   `json:"category_id"`
	}
	var logArray []Log
	file = append([]byte{byte('[')}, file...)
	file = file[:len(file)-1]
	file = append(file, byte(']'))
	_ = json.Unmarshal(file, &logArray)
	type categoryGroupedStruct struct {
		Title string `json:"title"`
		Count int    `json:"count"`
	}
	categoryGrouped := make(map[string]int)
	var responseStruct []categoryGroupedStruct
	for _, category := range configuration.Categories {
		for _, selfLog := range logArray {
			if category.ID == selfLog.CategoryID {
				categoryGrouped[category.Title] += 1
			}
		}
	}
	for title, i := range categoryGrouped {
		responseStruct = append(responseStruct, categoryGroupedStruct{
			Title: title,
			Count: i,
		})
	}
	respJS, _ := json.Marshal(responseStruct)
	return respJS
}
