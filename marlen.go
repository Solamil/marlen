package marlen

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func FakeMoney(url string) []string {
	var wg sync.WaitGroup
	var result []string
	wg.Add(2)
	btcStr := make(chan string)
	//	btcStr := getCryptoCurrency(url, "btc")
	go getCryptoCurrency(url, "btc", btcStr, &wg)
	//	xmrStr := getCryptoCurrency(url, "xmr")
	xmrStr := make(chan string)
	go getCryptoCurrency(url, "xmr", xmrStr, &wg)

	btc, _ := strconv.ParseFloat(<-btcStr, 64)
	xmr, _ := strconv.ParseFloat(<-xmrStr, 64)
	wg.Wait()
	result = append(result, fmt.Sprintf("%.2f", btc))
	result = append(result, fmt.Sprintf("%.2f", xmr))
	return result
}

func getCryptoCurrency(url, code string, answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()
	var result string = ""
	signature := fmt.Sprintf("%s:%s", url, code)
	if record, found := Get(signature); found {
		now := time.Now()
		d := record.Expiry
		if record.Value != "" {
			d = d.Add(time.Hour * 6)
		} else {
			d = d.Add(time.Minute * 30)
		}
		result = record.Value
		if d.After(now) {
			answer <- result
			return result
		}
	}
	url = fmt.Sprintf("%s/1%s", url, code)
	resp := NewRequest(url)
	result = strings.Split(resp, "\n")[0]
	Store(signature, result)
	answer <- result
	return result
}

// Deprecated
// func Nameday(url string) string {
//
// 	signature := fmt.Sprintf(`%s:%s`, url, "nameday")
// 	var answer string = ""
//
// 	if record, found := Get(signature); found {
// 		now := time.Now()
// 		d := record.Expiry
// 		if record.Value != "" &&
// 			d.Day() == now.Day() && d.Month() == now.Month() && d.Year() == now.Year() {
// 			answer = record.Value
// 			return answer
// 		} else if d = d.Add(time.Minute * 35); record.Value == "" && d.After(now) {
// 			answer = record.Value
// 			return answer
// 		}
//
// 	}
//
// 	if value := NewRequest(url); value != "" {
// 		answer = value
// 		Store(signature, answer)
//
// 	}
// 	return answer
// }

// func CnbCurrAsync(answer chan string, wg *sync.WaitGroup) string {
// 	defer wg.Done()
// 	result := readAllFile(pathHolytrinity)
// 	answer <- result
// 	return result
// }

func CnbCurrency(filepath string, answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, filepath, "currency")
	if record, found := Get(signature); found {
		now := time.Now()
		tUpdate := time.Date(now.Year(), now.Month(), now.Day(), 14, 45+1, 0, 0, now.Location())
		d := record.Expiry
		if record.Value != "" && ((now.Before(tUpdate) && now.Day() == d.Day() && now.Month() == d.Month() && now.Year() == d.Year()) ||
			d.After(tUpdate)) {
			result = record.Value
			answer <- result
			return result
		} else if d = d.Add(time.Minute * 35); record.Value == "" && d.After(now) {
			result = record.Value
			answer <- result
			return result
		}

	}

	value := readAllFile(filepath)
	result = value
	Store(signature, result)
	answer <- result
	return result
}

// Download image calendar for today
func CalendarImgRoutine(wg *sync.WaitGroup, url, filedest string) {
	defer wg.Done()
	CalendarImg(url, filedest)

}
func CalendarImg(url, filedest string) {
	fileinfo, err := os.Stat(filedest)
	t := time.Now()
	if err != nil || fileinfo.ModTime().Day() != t.Day() {
		NewImgRequest(url, filedest)
		return
	}
	fmt.Println("Calendar up to date.")
}
func RunScriptRoutine(wg *sync.WaitGroup, args ...string) {
	defer wg.Done()
	RunScript(args...)
}

func RunScript(args ...string) (string, error) {
	shell := "/bin/sh"
	output, err := exec.Command(shell, args...).Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(string(output))
	return string(output), nil
}
