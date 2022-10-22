package main

import (
	"fmt"
	"net/http"
	"text/template"
	"encoding/json"
	"io/ioutil"
	"strings"
//	"strconv"
	"time"
	"unicode/utf8"
	"os/exec"
//	"os"
)
var indexTemplate *template.Template
var responseTemplate *template.Template

type sunMoonInfo struct {
	Info string
	Day int
	Month int 
	Year int
}
var sunMoon = sunMoonInfo{}

type coinPriceInfo struct {
	Btc string 
	Xmr string
}
var coinPrices = coinPriceInfo{}

type currPriceInfo struct {
	Json string
	Date string
}
var currPrices = currPriceInfo{}

type userBaseRequest struct {
	CoinCode string `json:"coin_code"`	
	Param	 string `json:"param"`
	Location string `json:"location"`
}

func main() {
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/style.css")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/startpage.html")
	})
	
	indexTemplate, _ = template.ParseFiles("web/index.html")	
	responseTemplate, _ = template.ParseFiles("web/response.html")

//	jsonFile, err := os.Open("testfile.json")
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println("reading testfile.json")
//	defer jsonFile.Close()
//
//	byteValue, _ := ioutil.ReadAll(jsonFile)
//	var base userBaseRequest 
//	json.Unmarshal(byteValue, &base)
//	fmt.Println(base.CoinCode)
	
	http.HandleFunc("/base_info", base_handler)
	http.HandleFunc("/forecast_info", forecast_handler)
	http.ListenAndServe(":8900", nil)
}

func base_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var base userBaseRequest
	json.Unmarshal(body, &base)

	if base.CoinCode  == "" { 
		base.CoinCode = "usd"
	}

	if btc := get_coin_price(base.CoinCode, "btc"); btc != "" {
		coinPrices.Btc = btc
	}
	if xmr := get_coin_price(base.CoinCode, "xmr"); xmr != "" {
		coinPrices.Xmr = xmr
	}
	var json string
	if len(base.CoinCode) > 1 && base.Param == "conversion" {
		json = fmt.Sprintf(`{"btc": "%s","xmr": "%s", "coinCode": "%s"}`, coinPrices.Btc, coinPrices.Xmr, base.CoinCode)
	} else {
		if base.Location == "" {
			base.Location = "Zdar"
		}
		sunMoon := get_sun_moon_info(base.Location)
		hum_low_high := get_text_wttr_forecast(base.Location)
		currency := get_currency_rates()
		json = fmt.Sprintf(`{"btc": "%s","xmr": "%s", "coinCode": "%s",
			"sun_moon": %s, "hum_low_high": %s,  %s}`, 
		coinPrices.Btc, coinPrices.Xmr,base.CoinCode, sunMoon, hum_low_high, currency)
	}
//	var session http.Cookie
//	session.Name = "sessionid"
//	session.Domain = "michalkukla.xyz"
//	session.Path = "/startpage"
//	session.HttpOnly = true
//	session.Secure = true
	w.Write([]byte(json))

	
}

func forecast_handler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var base userBaseRequest
	json.Unmarshal(body, &base)
	fmt.Println(base.CoinCode)
//	weatherFile, err := os.Open("weatherreport")
//	if err != nil {
//		fmt.Println(err)
//	}
//	defer weatherFile.Close()
//	byteWeather, _ := ioutil.ReadAll(weatherFile)
//	w.Write(byteWeather)
}

func get_sun_moon_info(location string) string {
	year, month, day := time.Now().Date()
	if sunMoon.Day != day || time.Month(sunMoon.Month) != month || sunMoon.Year != year {
		format := "%S+%s+%m"	
		if info := get_weather(format, location); info != "" {
			sunMoon.Info = info
			sunMoon.Day = day 
			sunMoon.Month = int(month) 
			sunMoon.Year = year
		}
	}
	return sunMoon.Info
}
func get_weather(format, location string) string {
	url := fmt.Sprintf(`https://wttr.in/%s?format="%s"`, location, format)
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
func get_text_wttr_forecast(location string) string {
	output, err := exec.Command("/bin/sh", "sb-forecast.sh", location).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high := strings.Replace(string(output), "\n", "", 1)

	output, err = exec.Command("/bin/sh", "sb-forecast.sh", location, "23", "26").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next := strings.Replace(string(output), "\n", "", 1)
	output, err = exec.Command("/bin/sh", "sb-forecast.sh", location, "33", "36").Output()
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

func get_coin_price(showRates, coinCode string) string {

	url := fmt.Sprintf(`https://%s.rate.sx/1%s`, showRates, coinCode)
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
	price, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	priceStr := string(price)
	priceStr = priceStr[:len(priceStr)-1]
	return priceStr
}
func get_currency_rates() string {
	now := time.Now()
	if len(currPrices.Json) > 0 {
		dateStr := string(now.Day())+"."+string(int(now.Month()))+"."+string(now.Year())
		
		if currPrices.Date == dateStr || now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			return currPrices.Json
		}
	}
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
	url := "https://cnb.cz/cs/financni-trhy/devizovy-trh/kurzy-devizoveho-trhu/kurzy-devizoveho-trhu/denni_kurz.txt"
	reqm, _ := http.NewRequest("GET", url, nil)

	reqm.Header.Set("Content-Type", "text/html")
	content, err := http.DefaultClient.Do(reqm)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	b, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return "" 
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
