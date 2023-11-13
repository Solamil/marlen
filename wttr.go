package marlen

import (
	"time"
	"fmt"
	"sync"
	"os/exec"
	"strings"
)

func GetDailyWttrInfo(url string, answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()
	signature := fmt.Sprintf(`%s:%s`, url, "daily")
	var result string = ""

	if record, found := Get(signature); found {
		now := time.Now()
		yearNow, monthNow, dayNow := now.Date()
		year, month, day := record.Expiry.Date()
		d := record.Expiry
		if record.Value != "" && dayNow == day && monthNow == month && yearNow == year {
			result = record.Value
			answer <- result
			return result
		} else if d = d.Add(time.Minute * 35); record.Value == "" && d.After(now) {
			result = record.Value
			answer <- result
			return result
		}
	}
	value := getWeatherInfo(url)
	result = value
	Store(signature, value)
	answer <- result
	return result
}

func GetForecast(url string, answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()
	signature := fmt.Sprintf(`%s:%s`, url, "forecast")
	shell := "/bin/sh"
	scriptFile := "./scripts/sb-forecast.sh"
	var result string = ""
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 6)
		} else {
			d = d.Add(time.Minute * 35)
		}
		if d.After(now) {
			result = record.Value
			answer <- result
			return result
		}
	}
	output, err := exec.Command(shell, scriptFile, url).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high := strings.Replace(string(output), "\n", "", 1)

	output, err = exec.Command(shell, scriptFile, url, "23", "26").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next := strings.Replace(string(output), "\n", "", 1)
	output, err = exec.Command(shell, scriptFile, url, "33", "36").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next2 := strings.Replace(string(output), "\n", "", 1)
	var value string = ""

	if len(hum_low_high) > 0 {
		value = hum_low_high
	}
	if len(hum_low_high_next) > 0 {
		value = fmt.Sprintf("%s\n%s", value, hum_low_high_next)
	}
	if len(hum_low_high_next2) > 0 {
		value = fmt.Sprintf("%s\n%s", value, hum_low_high_next2)
	}
	result = value
	Store(signature, value)
	answer <- result
	return result
}

func getWeatherInfo(url string) string {
	var result string = ""
	value := NewRequest(url)
	if len(value) > 0 {
		value = strings.ReplaceAll(value, "\"", "")
		result = strings.ReplaceAll(value, "\n", "")
	}
	return result
}
