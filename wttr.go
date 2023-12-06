package marlen

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"
	//	"os/exec"
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
	scriptFile := filepath.Join("scripts", "sb-forecast.sh")
	if output, _ := RunScript(scriptFile); len(output) > 0 {
		result = output
		Store(signature, output)
	} else {
		Store(signature, "")
	}
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
