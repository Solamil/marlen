package main

import (
	"fmt"
	"net/http"
	"text/template"
	"io/ioutil"
	"strings"
	"strconv"
	"time"
	"unicode/utf8"
	"os/exec"
	"os"
)
var indexTemplate *template.Template
var responseTemplate *template.Template

type sunMoonInfo struct {
	Info string
	Day int
	Month time.Month
	Year int
}
var sunMoon = sunMoonInfo{}

type coinPriceInfo struct {
	Btc float64
	Xmr float64
}
var coinPrices = coinPriceInfo{0.0, 0.0}

type currPriceInfo struct {
	Json string
	Date string
}
var currPrices = currPriceInfo{}

type jsonData struct {
	foo string
}

func main() {
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/style.css")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})
	
	indexTemplate, _ = template.ParseFiles("web/index.html")	
	responseTemplate, _ = template.ParseFiles("web/response.html")
	http.HandleFunc("/base_info", base_handler)
	http.HandleFunc("/forecast_info", forecast_handler)
	fmt.Println(session.Name)
	http.ListenAndServe(":8900", nil)
}

func base_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	params := strings.Split(string(body), " ")
	var typeCurr string
	if params[0] == "" { 
		typeCurr = "usd"
	} else {
		typeCurr = params[0]
	}

	if btc := get_coin_price(typeCurr, "btc"); btc != 0 {
		coinPrices.Btc = btc
	}
	if xmr := get_coin_price(typeCurr, "xmr"); xmr != 0 {
		coinPrices.Xmr = xmr
	}
	var json string
//	params: CURRENCY conversion
	if len(params) > 1 && params[1] == "conversion" {
		json = fmt.Sprintf(`{"btc": "%f","xmr": "%f", "coinCode": "%s"}`, coinPrices.Btc, coinPrices.Xmr, typeCurr)
	} else {
		sunMoon := get_sun_moon_info()
		currency := get_currency_rates()
		hum_low_high := get_text_wttr_forecast()
		json = fmt.Sprintf(`{"btc": "%f","xmr": "%f", "coinCode": "%s",
			"sun_moon": %s, "hum_low_high": %s,  %s}`, 
		coinPrices.Btc, coinPrices.Xmr,typeCurr, sunMoon, hum_low_high, currency)
	}
	var session http.Cookie
	session.Name = "sessionid"
	session.Domain = "michalkukla.xyz"
	session.Path = "/startpage"
	session.HttpOnly = true
	session.Secure = true
	w.Write([]byte(json))

	
}

func forecast_handler(w http.ResponseWriter, r *http.Request) {

	weatherFile, err := os.Open("weatherreport")
	if err != nil {
		fmt.Println(err)
	}
	defer weatherFile.Close()
	byteWeather, _ := ioutil.ReadAll(weatherFile)
	w.Write(byteWeather)
}

func get_sun_moon_info() string {
	year, month, day := time.Now().Date()
	if sunMoon.Day != day || sunMoon.Month != month || sunMoon.Year != year {
		format := "%S+%s+%m"	
		if info := get_weather(format); info != "" {
			sunMoon.Info = info
			sunMoon.Day = day 
			sunMoon.Month = month 
			sunMoon.Year = year
		}
	}
	return sunMoon.Info
}
func get_weather(format string) string {
	place := "Zdar"
	url := fmt.Sprintf(`https://wttr.in/%s?format="%s"`, place, format)
	reqm, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	reqm.Header.Set("Content-Type", "text/html")
	content, err := http.DefaultClient.Do(reqm)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	weather, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(weather)
}
func get_text_wttr_forecast() string {

	output, err := exec.Command("/bin/sh", "sb-forecast.sh").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high := strings.Replace(string(output), "\n", "", 1)

	output, err = exec.Command("/bin/sh", "sb-forecast.sh", "23", "26").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next := strings.Replace(string(output), "\n", "", 1)
	output, err = exec.Command("/bin/sh", "sb-forecast.sh", "33", "36").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next2 := strings.Replace(string(output), "\n", "", 1)
	json := fmt.Sprintf(`["%s", "%s", "%s"]`, hum_low_high, hum_low_high_next, hum_low_high_next2)
	
	return json
}
func exec_shellscript(shellscript ...string) string {
	cmd := shellscript[0]
	var param1 string
	var param2 string
	if shellscript[1] != "" {
		param1 = shellscript[1]	
	}
	if shellscript[2] != "" {
		param1 = shellscript[2]	
	}
	output, err := exec.Command("/bin/sh", cmd, param1, param2).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	outputStr := string(output)
	return outputStr
}

func get_coin_price(showRates, coinCode string) float64 {

	url := fmt.Sprintf(`https://%s.rate.sx/1%s`, showRates, coinCode)
	reqm, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	reqm.Header.Set("Content-Type", "text/html")
	content, err := http.DefaultClient.Do(reqm)

	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	price, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	priceStr := string(price)
	priceStr = priceStr[:len(priceStr)-1]

	priceFloat, err := strconv.ParseFloat(priceStr, 32)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}
	return priceFloat
}
func get_currency_rates() string {
	rates := getCnbRates()
	exchRates := strings.Split(rates, "\n")


	gbpCurr := strings.Split(exchRates[33], "|")
	gbpValue := gbpCurr[len(gbpCurr)-1]
	gbpCode := gbpCurr[len(gbpCurr)-2]
	gbpVolume := gbpCurr[len(gbpCurr)-3]

	eurCurr := strings.Split(exchRates[7], "|")
	eurValue := eurCurr[len(eurCurr)-1]
	eurCode := eurCurr[len(eurCurr)-2]
	eurVolume := eurCurr[len(eurCurr)-3]

	usdCurr := strings.Split(exchRates[32], "|")
	usdValue := usdCurr[len(usdCurr)-1]
	usdCode := usdCurr[len(usdCurr)-2]
	usdVolume := usdCurr[len(usdCurr)-3]

	json := fmt.Sprintf(`"%s":{"volume": %s, "value": "%s"}, "%s":{"volume": %s, "value": "%s"}, "%s":{"volume": %s, "value": "%s"}`,
			usdCode, usdVolume, usdValue, eurCode, eurVolume, eurValue, gbpCode, gbpVolume, gbpValue)
	
	currPrices.Json = json
	currPrices.Date = strings.Split(exchRates[0], " ")[0]

	return currPrices.Json 
}

func getCnbRates() string {
	now := time.Now()
	if len(currPrices.Json) > 0 {
		dateStr := string(now.Day())+"."+string(int(now.Month()))+"."+string(now.Year())
		
		if currPrices.Date == dateStr || now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			return currPrices.Json
		}
	}
	url := "https://cnb.cz/cs/financni-trhy/devizovy-trh/kurzy-devizoveho-trhu/kurzy-devizoveho-trhu/denni_kurz.txt"
	reqm, _ := http.NewRequest("GET", url, nil)

	reqm.Header.Set("Content-Type", "text/html")
	content, err := http.DefaultClient.Do(reqm)

	if err != nil {
		fmt.Println(err)
		return currPrices.Json
	}
	b, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return currPrices.Json
	}
	return string(b)
}

func condenseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func truncateStrings(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for !utf8.ValidString(s[:n]) {
		n--
	}
	return s[:n]
}
