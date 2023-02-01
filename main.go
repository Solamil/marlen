package main

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"strings"
//	"html"
//	"strconv"
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
	OtherInfo string
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
var WEB_DIR string = "web"
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

// var currSymbols = map[string]string{
// 	"usd": "$",
// 	"eur": "€",
// 	"gbp": "£",
// 	"czk": "Kč",
// 	"btc": "BTC",
// }

var indexTemplate *template.Template

func main() {
	http.HandleFunc("/style.css", file_handler) 
	http.HandleFunc("/pics/git-icon.svg", file_handler)
	http.HandleFunc("/pics/rain.webp", file_handler) 
	http.HandleFunc("/pics/clouds.webp", file_handler) 
	http.HandleFunc("/pics/rain.gif", file_handler) 
	http.HandleFunc("/pics/clouds.gif", file_handler) 
	http.HandleFunc("/pics/forecastPrecip_0days.webp", file_handler)
	http.HandleFunc("/pics/forecastPrecip_0days.gif", file_handler) 
	http.HandleFunc("/pics/forecastTemp_0days.gif", file_handler)
	http.HandleFunc("/pics/forecastTemp_0days.webp", file_handler)
	http.HandleFunc("/pics/forecastWind_0days.gif", file_handler)
	http.HandleFunc("/pics/forecastWind_0days.webp", file_handler)
	http.HandleFunc("/pics/forecastPrecip_1days.webp", file_handler)
	http.HandleFunc("/pics/forecastPrecip_1days.gif", file_handler)
	http.HandleFunc("/pics/mhcam1.webp", file_handler)
	http.HandleFunc("/js/module-wttrin-widget.js", file_handler)
	http.HandleFunc("/cover.html", file_handler)
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
	var weatherInfo string = ""
	var forecastFirst string = ""
	var forecastSecond string = ""
	var otherInfo string = "🌐"

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
	if len(forecasts) >= 3 {
		forecastFirst = forecasts[1]
		forecastSecond = forecasts[2]
	}
	sunMoonUrl := fmt.Sprintf(`%s?format="%s"`, wttrin, "%S+%s+%m")
	
	if sunMoonStr := get_daily_wttr_info(sunMoonUrl); len(sunMoonStr) != 0 {
		sunMoon := strings.Split(sunMoonStr, " ")
		if len(sunMoon) == 3 && len(forecasts) > 0 {
			weatherInfo = "🌅 "+sunMoon[0]+" 🌇"+sunMoon[1]+" "+sunMoon[2]+" "+forecasts[0]
		}
	}
	
	urlNameDay := fmt.Sprintf("https://svatek.michalkukla.xyz/today?country=%s", lang)
	nameDay = get_name_day(urlNameDay)

	var localeTags string = ""
	var tag string = ""
	for key, value := range countryFlags {
		tag = getHTMLOptionTag(key, value, (key == lang))
		localeTags = strings.Join([]string{localeTags, tag}, "\n")  	
	}

	if len(r.Header["X-Real-Ip"]) > 0 {
		otherInfo = fmt.Sprintf("<a target=\"_blank\" href=\"https://www.whois.com/whois/%s\">🌐 %s</a>", 
					r.Header["X-Real-Ip"][0], r.Header["X-Real-Ip"][0])
	}
		
	var i indexDisplay
	i.NameDay = nameDay 
	i.Bg = bg
	i.Location, _ = url.QueryUnescape(location)
	i.WeatherInfo = weatherInfo 
	i.ForecastFirst = forecastFirst
	i.ForecastSecond = forecastSecond
	i.OtherInfo = otherInfo 
	i.Currency = get_holy_trinity("https://czk.michalkukla.xyz/holy_trinity?p")
	i.WttrLink = wttrLink
	i.WttrSrc = wttrPng
	i.WttrInHolder = wttrInHolders[prefix]
	i.LocaleOptions = localeTags
	indexTemplate, _ = template.ParseFiles("web/index.html")
	indexTemplate.Execute(w, i)

}

func file_handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, WEB_DIR+r.URL.Path)
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
	var result string = ""
	value := new_request(url)
	if len(value) > 0 {
		value = strings.ReplaceAll(value, "\"", "")
		result = strings.ReplaceAll(value, "\n", "")
	}
	return result 
}

func get_forecast(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "forecast")
	cacheSignature := hash(signature)
	var answer string = ""
	var lastRecord string = ""	
	if record, found := get(cacheSignature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 6)
		lastRecord = record.value
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
	if len(value) > 0 {
		answer = store(cacheSignature, value)
	} else {
		answer = lastRecord
	}
	
	return answer
}

func get_holy_trinity(url string) string {
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "trinity")
	cacheSignature := hash(signature)
	if record, found := get(cacheSignature); found && record.value != "" {
		now := time.Now()
		tUpdate := time.Date(now.Year(), now.Month(), now.Day(), 14, 45+1, 0, 0, now.Location())
		d := record.expiry
		if (now.Before(tUpdate) && now.Day() == d.Day() && now.Month() == d.Month() && now.Year() == d.Year())	|| 
			d.After(tUpdate) {
			result = record.value
			return result
		}
		
	}
	
	if value := new_request(url); len(value) > 0 {
		result = store(cacheSignature, value)
	}
	return result
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

	if value := new_request(url); value != "" {
		answer = store(cacheSignature,string(value))
	}
	return answer 
}

func new_request(url string) string {
	var answer string = ""
	client := &http.Client{Timeout: 2 * time.Second}
	reqm, _ := http.NewRequest("GET", url, nil)

	reqm.Header.Set("Content-Type", "text/html")
	content, err := client.Do(reqm)

	if err != nil || content.StatusCode >= http.StatusBadRequest {
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
