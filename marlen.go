package marlen

import (
	"strconv"
	"time"
	"fmt"
	"sync"
	"strings"
	"path/filepath"
)

var ratesDir string = "rates"
var pathHolytrinity string = filepath.Join(ratesDir, "svata_trojice.txt")

func FakeMoney(url string) string {
	var wg sync.WaitGroup
	var result string = ""
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
	result = fmt.Sprintf("1<b style=\"color: gold;\">BTC</b> %.2f$"+
		" 1<b style=\"color: #999;\">XMR</b> %.2f$",
		btc, xmr)
	return result
}

func getCryptoCurrency(url, code string, answer chan string, wg* sync.WaitGroup) string {
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

func CnbCurr(answer chan string, wg *sync.WaitGroup) string {
	defer wg.Done()	
	result := readAllFile(pathHolytrinity)
	answer <- result 
	return result 
}

// Deprecated
// func CnbCurrency(url string, answer chan string, wg *sync.WaitGroup) string {
// 	defer wg.Done()
// 	var result string = ""
// 	signature := fmt.Sprintf(`%s:%s`, url, "currency")
// 	if record, found := Get(signature); found {
// 		now := time.Now()
// 		tUpdate := time.Date(now.Year(), now.Month(), now.Day(), 14, 45+1, 0, 0, now.Location())
// 		d := record.Expiry
// 		if record.Value != "" && ((now.Before(tUpdate) && now.Day() == d.Day() && now.Month() == d.Month() && now.Year() == d.Year()) ||
// 			d.After(tUpdate)) {
// 			result = record.Value
// 			answer <- result
// 			return result
// 		} else if d = d.Add(time.Minute * 35); record.Value == "" && d.After(now) {
// 			result = record.Value
// 			answer <- result
// 			return result
// 		}
// 
// 	}
// 
// 	value := NewRequest(url)
// 	result = value
// 	Store(signature, result)
// 	answer <- result
// 	return result
// }
