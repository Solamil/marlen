package marlen

import (
	"time"
	"fmt"
	"os/exec"
	"strings"
)

func GetDailyWttrInfo(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "daily")
	var answer string = ""

	if record, found := Get(signature); found {
		now := time.Now()
		yearNow, monthNow, dayNow := now.Date()
		year, month, day := record.Expiry.Date()
		d := record.Expiry
		if record.Value != "" && dayNow == day && monthNow == month && yearNow == year {
			answer = record.Value
			return answer
		} else if d = d.Add(time.Minute * 35); record.Value == "" && d.After(now) {
			answer = record.Value
			return answer
		}
	}
	value := getWeatherInfo(url)
	answer = value
	Store(signature, value)
	return answer
}

func GetForecast(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "forecast")
	shell := "/bin/sh"
	scriptFile := "./scripts/sb-forecast.sh"
	var answer string = ""
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 6)
		} else {
			d = d.Add(time.Minute * 35)
		}
		if d.After(now) {
			answer = record.Value
			return answer
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
	answer = value
	Store(signature, value)

	return answer
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
