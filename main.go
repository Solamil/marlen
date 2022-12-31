package main

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"strings"
//	"html"
	"strconv"
	"time"
	"os/exec"
	"crypto/md5"
	"text/template"
	"github.com/hashicorp/golang-lru/v2"
//	"os"
)

type cacheRecord struct {
	value string
	expiry time.Time

}

type indexUrlParams struct {
	Lang [1]string `json:"lang"`
	Location [1]string `json:"location"`
	Bg [1]string `json:"bg"`
}

type indexDisplay struct {
	Bg string
	Location string
	WeatherInfo string
	LocaleOptions string
	Currency string
	NameDay string
	ForecastFirst string
	ForecastSecond string
	WttrLink string
	WttrSrc string
	WttrInHolder string
}
const CACHESIZE int = 10000
const HASHSIZE int = 16
var cache, _ = lru.New[[HASHSIZE]byte, cacheRecord](CACHESIZE)

var wttrInHolders = map[string]string{
	"en": "Weather in...",
	"de": "Wetter für...",
	"cs": "Počasí v...",
}

var countryFlags = map[string]string{
	"en-US": "🇺🇸",
	"de-DE": "🇩🇪",
	"cs-CZ": "🇨🇿",
}

var currSymbols = map[string]string{
	"usd": "$",
	"eur": "€",
	"gbp": "£",
	"czk": "Kč",
	"btc": "BTC",
}

var indexTemplate *template.Template

func main() {
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/style.css")
	})
	http.HandleFunc("/pics/git-icon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/git-icon.svg")
	})
	http.HandleFunc("/pics/rain.webp", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/rain.webp")
	})
	http.HandleFunc("/pics/clouds.webp", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/clouds.webp")
	})
	http.HandleFunc("/pics/rain.gif", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/rain.gif")
	})
	http.HandleFunc("/pics/clouds.gif", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/clouds.gif")
	})
	http.HandleFunc("/js/module-wttrin-widget.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/js/module-wttrin-widget.js")
	})
	http.HandleFunc("/forecast.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/forecast.html")
	})
	http.HandleFunc("/cover.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cover.html")
	})
	indexTemplate, _ = template.ParseFiles("web/index.html")
	http.HandleFunc("/index.html", index_handler)
	http.HandleFunc("/", index_handler)
	http.ListenAndServe(":8901", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	var location string = "Zdar"
	var bg string = "893531"
	var lang string = "en-US"
	var nameDay string = ""

	if c, err := r.Cookie("place"); err == nil {
		value := strings.Split(c.String(), "=")[1]
//		location, _ = url.QueryUnescape(value)
		location = value
//		fmt.Println(value)
	} else if err != nil {
		fmt.Println(err)
	}

	if c, err := r.Cookie("lang"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		lang = value	
	} else if err != nil {
		fmt.Println(err)
	}

	q, _ := url.PathUnescape(r.URL.RawQuery)
	if len(q) != 0 {
		m, err := url.ParseQuery(q)
		if err != nil {
			fmt.Println(err)
		}
		
		js, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
		}
		var param *indexUrlParams
		json.Unmarshal(js, &param)		
		if len(param.Location[0]) > 0 {
			location = param.Location[0]
		}
		if len(param.Lang[0]) > 0 {
			lang = param.Lang[0]
		}
		if len(param.Bg[0]) > 0 {
			bg = param.Bg[0]
		}
	}
	wttrin := fmt.Sprintf("https://wttr.in/%s", location)
	prefix := strings.Split(lang, "-")[0]
	wttrPng := fmt.Sprintf("%s_0pq_transparency=255_background=%s_lang=%s.png",
				wttrin, bg, prefix)
	wttrLink := fmt.Sprintf("%s?lang=%s", wttrin, prefix)
	forecastStr := get_forecast(wttrin)
	forecasts := strings.Split(forecastStr, "\n")

	sunMoonUrl := fmt.Sprintf(`%s?format="%s"`, wttrin, "%S+%s+%m")
	sunMoonStr := get_daily_wttr_info(sunMoonUrl)
	sunMoon := strings.Split(sunMoonStr, " ")
	
	urlNameDay := fmt.Sprintf("https://svatek.michalkukla.xyz/today?country=%s", lang)
	nameDay = get_name_day(urlNameDay)

	var localeTags string = ""
	var tag string = ""
	for key, value := range countryFlags {
		if key == lang {
			tag = getHTMLOptionTag(key, value, true)
		} else {
			tag = getHTMLOptionTag(key, value, false)
		}
		localeTags = strings.Join([]string{localeTags, tag}, "\n")  	
	}

	urlCurr := fmt.Sprintf("https://czk.michalkukla.xyz/?code=%s", "USD")
	usdValue, _ := strconv.ParseFloat(get_cnb_info(urlCurr)[0], 64)
	urlCurr = fmt.Sprintf("https://czk.michalkukla.xyz/?code=%s", "EUR")
	eurValue, _ := strconv.ParseFloat(get_cnb_info(urlCurr)[0], 64)
	urlCurr = fmt.Sprintf("https://czk.michalkukla.xyz/?code=%s", "GBP")
	gbpValue, _ := strconv.ParseFloat(get_cnb_info(urlCurr)[0], 64)
	currency := fmt.Sprintf("1$ %.2fKč 1€ %.2fKč 1£ %.2fKč",  usdValue, eurValue, gbpValue)
		
	var i indexDisplay
	i.NameDay = nameDay 
	i.Bg = bg
	i.Location, _ = url.QueryUnescape(location)
	i.WeatherInfo = "🌅 "+sunMoon[0]+" 🌇"+sunMoon[1]+" "+sunMoon[2]+" "+forecasts[0]
	i.ForecastFirst = forecasts[1]
	i.ForecastSecond = forecasts[2]
	i.Currency = currency
	i.WttrLink = wttrLink
	i.WttrSrc = wttrPng
	i.WttrInHolder = wttrInHolders[prefix]
	i.LocaleOptions = localeTags
	indexTemplate, _ = template.ParseFiles("web/index.html")
	indexTemplate.Execute(w, i)

}

func get_daily_wttr_info(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "daily")
	cacheSignature := hash(signature)
	var answer string = ""
	record, found := get(cacheSignature)	
	if found {
		yearNow, monthNow, dayNow := time.Now().Date()	
		year, month, day := record.expiry.Date()	
		if record.value != "" && dayNow == day && monthNow == month && yearNow == year {
			answer = record.value
			return answer
		}
	}
	value := get_weather_info(url)
	answer = store(cacheSignature, value)
	return answer
}

func get_weather_info(url string) string {
	value := new_request(url)
	str_out := strings.ReplaceAll(string(value), "\"", "")
	str_out = strings.ReplaceAll(string(str_out), "\n", "")
	return string(str_out)
}

func get_forecast(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "forecast")
	cacheSignature := hash(signature)
	var answer string = ""
	
	if record, found := get(cacheSignature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 6)
		if d.After(now) {
			answer = record.value
			return answer
		}
	}
	output, err := exec.Command("/bin/sh", "sb-forecast.sh", url).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high := strings.Replace(string(output), "\n", "", 1)

	output, err = exec.Command("/bin/sh", "sb-forecast.sh", url, "23", "26").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next := strings.Replace(string(output), "\n", "", 1)
	output, err = exec.Command("/bin/sh", "sb-forecast.sh", url, "33", "36").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	hum_low_high_next2 := strings.Replace(string(output), "\n", "", 1)

	value := strings.Join([]string{hum_low_high, hum_low_high_next, hum_low_high_next2}, "\n")
	answer = store(cacheSignature, value)
	return answer

}

func get_cnb_info(url string) []string {

//	signature := fmt.Sprintf(`%s:%s`, url, "cnb-rates")
//	cacheSignature := hash(signature)
//	var answer string = ""
//
//	if record, found := get(cacheSignature); found && record.value != "" {
//		now := time.Now()
//		d := record.expiry
//		if d.Day() == now.Day() && d.Month() == now.Month() {
//			fmt.Println("cached")
//			answer = record.value
//			answerList := strings.Split(answer, "\n")
//			return answerList 
//		}
//	}
	value := new_request(url)
//	answer = store(cacheSignature,string(value))
	answerList := strings.Split(string(value), "\n")
	return answerList 
}

func get_name_day(url string) string {

	signature := fmt.Sprintf(`%s:%s`, url, "nameday")
	cacheSignature := hash(signature)
	var answer string = ""
	
	if record, found := get(cacheSignature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		if d.Day() == now.Day() && d.Month() == now.Month() && d.Year() == now.Year() {
			answer = record.value
			return answer
		}
	}

	value := new_request(url)
	answer = store(cacheSignature,string(value))
	return answer 
}

func new_request(url string) string {
	var answer string = ""
	reqm, _ := http.NewRequest("GET", url, nil)

	reqm.Header.Set("Content-Type", "text/html")
	content, err := http.DefaultClient.Do(reqm)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	value, err := ioutil.ReadAll(content.Body)
	if err != nil {
		fmt.Println(err)
		return "" 
	}
	answer = string(value)
	return answer	
}

func getHTMLOptionTag(value, symbol string, selected bool) string {
	var tag string = ""
	if selected {
		tag = fmt.Sprintf("<option value=\"%s\" %s>%s</option>", value, "selected", symbol)
	} else {
		tag = fmt.Sprintf("<option value=\"%s\">%s</option>", value, symbol)
	}
	return tag
}

func store(signature [HASHSIZE]byte, value string) string {
	cache.Add(signature, cacheRecord{value, time.Now()})
	return value
} 

func get(signature [HASHSIZE]byte) (cacheRecord, bool) {
	record, found := cache.Get(signature)
	return record, found
}

func hash(signature string) [HASHSIZE]byte {
	return md5.Sum([]byte(signature))
}
