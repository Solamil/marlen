package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"strings"
//	"context"
//	"html"
	"strconv"
	"time"
	"os/exec"
	"crypto/md5"
	"text/template"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/beevik/etree"
	"os"
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
	CryptoCurrency string
}
type feedsDisplay struct {
	Bg string
	RssFeed string
}

const CACHESIZE int = 10000
const MIN_SIZE_FILE_CACHE = 80
var CACHE_DIR string = "cache"
const HASHSIZE int = md5.Size
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
var feedsTemplate *template.Template

func main() {
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
	http.HandleFunc("/pics/bitcoin-icon.svg", file_handler)
	http.HandleFunc("/pics/monero-icon.svg", file_handler)
	http.HandleFunc("/js/module-wttrin-widget.js", file_handler)
	http.HandleFunc("/cover.html", file_handler)
	indexTemplate, _ = template.ParseFiles("web/index.html")
	http.HandleFunc("/index.html", index_handler)
	http.HandleFunc("/feeds.html", feeds_handler)
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
	i.CryptoCurrency = getBtcXmr("https://rate.sx")
	indexTemplate, _ = template.ParseFiles("web/index.html")
	indexTemplate.Execute(w, i)

}

func feeds_handler(w http.ResponseWriter, r *http.Request) {
	var rssFeed string = ""
	var bg string = "442244"
	var i feedsDisplay
	var ctkUrl string = "https://www.ceskenoviny.cz/sluzby/rss"
	ctkCr := rss_feed_ctk(ctkUrl+"/cr.php", 5)
	ctkSvet := rss_feed_ctk(ctkUrl+"/svet.php", 5)
	ctkEko := rss_feed_ctk(ctkUrl+"/ekonomika.php", 5)
	ctkSport := rss_feed_ctk(ctkUrl+"/sport.php", 2)
	neovlivni := rss_feed_neovlivni("https://neovlivni.cz/feed/atom/")
	render_feeds := fmt.Sprintf(`%s <br> %s <br> %s <br> %s <br> %s`, neovlivni, ctkCr, ctkSvet, ctkEko, ctkSport)
	rssFeed = render_feeds
//	rssFeed = rss_feed_neovlivni("https://neovlivni.cz/feed/atom/")
	i.Bg = bg
	i.RssFeed = rssFeed
	feedsTemplate, _ = template.ParseFiles("web/feeds.html")
	feedsTemplate.Execute(w, i)
}

func file_handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, WEB_DIR+r.URL.Path)
}

func get_daily_wttr_info(url string) string {
	signature := fmt.Sprintf(`%s:%s`, url, "daily")
	var answer string = ""
	record, found := get(signature)	
	if found {
		yearNow, monthNow, dayNow := time.Now().Date()	
		year, month, day := record.expiry.Date()	
		if record.value != "" && dayNow == day && monthNow == month && yearNow == year {
			answer = record.value
			return answer
		}
	}
	value := get_weather_info(url)
	answer = value
	store(signature, value)
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
	var answer string = ""
	var lastRecord string = ""	
	if record, found := get(signature); found && record.value != "" {
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
		answer = value
		store(signature, value)
	} else {
		answer = lastRecord
	}
	
	return answer
}

func get_holy_trinity(url string) string {
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "trinity")
	if record, found := get(signature); found && record.value != "" {
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
		result = value
		store(signature, result)
	}
	return result
}

func get_name_day(url string) string {

	signature := fmt.Sprintf(`%s:%s`, url, "nameday")
	var answer string = ""
	
	if record, found := get(signature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		if d.Day() == now.Day() && d.Month() == now.Month() && d.Year() == now.Year() {
			answer = record.value
			return answer
		}
	}

	if value := new_request(url); value != "" {
		answer = string(value)
		store(signature,answer)
		
	}
	return answer 
}

func rss_feed_neovlivni(url string) string {
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
	if record, found := get(signature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 6)
		result = record.value
		if d.After(now) {
			return result 
		}
	}
	resp := new_request(url)	
	if resp == "" {
		return result
	}
	doc := etree.NewDocument()

	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)		
		return ""
	}
	
//	if err := doc.ReadFromFile("BwJLymVb.atom"); err != nil {
//		fmt.Println(err)		
//		return ""
//	}
	root := doc.SelectElement("feed")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").SelectAttrValue("href", "")
	result = fmt.Sprintf("<h3><a href=\"%s\" target=\"_blank\">%s</a></h3>\n<ul>", linkSite, mainTitle)
	for _, e := range root.SelectElements("entry") {
		title := e.SelectElement("title").Text()
		author := e.SelectElement("author")
		name := author.SelectElement("name").Text()
		published := e.SelectElement("published").Text()
		link := e.SelectElement("link").SelectAttrValue("href", "")
//		t, _ := time.Parse(time.RFC3339, published)
		date := fmt.Sprintf("<span class=\"date\">%s</span>", published)
// 	✏️ &#9999;📜&#128220;
		line := fmt.Sprintf(`<li><a href="%s" target="_blank">%s &#9999;%s &#128220;%s</a></li>`,link, date, name, title)
		result = fmt.Sprintf("%s\n%s", result, line)

	}
	result = fmt.Sprintf("%s\n</ul>", result)	
	store(signature, result)
	return result
}

func getBtcXmr(url string) string {
	var result string = ""
	btcStr := getCryptoCurrency(url, "btc")
	btc, _ := strconv.ParseFloat(btcStr, 64)
	xmrStr := getCryptoCurrency(url, "xmr")
	xmr, _ := strconv.ParseFloat(xmrStr, 64)
	result = fmt.Sprintf(`1<img src="./pics/bitcoin-icon.svg" loading=lazy> %.2f$
			      1<img src="./pics/monero-icon.svg" loading=lazy> %.2f$`,
		      		btc, xmr)
	return result
}

func getCryptoCurrency(url, code string) string {
	var result string = ""
	signature := fmt.Sprintf("%s:%s", url, code)
	if record, found := get(signature); record.value != "" && found {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 2)
		result = record.value
		if d.After(now) {
			return result
		}
	}
	url = fmt.Sprintf("%s/1%s", url, code)
	resp := new_request(url)
	result = strings.Split(resp, "\n")[0]
	store(signature, result)
	return result
}

func rss_feed_ctk(url string, nTitles int) string {
	var result string = ""
	signature := fmt.Sprintf(`%s:%s`, url, "rssFeed")
	if record, found := get(signature); found && record.value != "" {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 2)
		result = record.value
		if d.After(now) {
			return result 
		}
	}
	doc := etree.NewDocument()
//	if err := doc.ReadFromFile("cr.rss"); err != nil {
//		fmt.Println(err)		
//		return ""
//	}
	resp := new_request(url)
	if resp == "" {
		return result
	}
	if err := doc.ReadFromString(resp); err != nil {
		fmt.Println(err)		
		return ""
	}

	root := doc.SelectElement("rss").SelectElement("channel")
	mainTitle := root.SelectElement("title").Text()
	linkSite := root.SelectElement("link").Text()
	result = fmt.Sprintf("<h3><a href=\"%s\" target=\"_blank\">%s</a></h3>\n<ul>", linkSite, mainTitle)
	if nTitles < 1 {
		nTitles = 5
	}
	var size int = nTitles	
	for i, e := range root.SelectElements("item") {
		if i >= size {
			break
		}
		title := e.SelectElement("title").Text()
		published := e.SelectElement("pubDate").Text()
		description := e.SelectElement("description").Text()
		link := e.SelectElement("link").Text()
//		t, _ := time.Parse(time.RFC3339, published)
		date := fmt.Sprintf("<span class=\"date\">%s</span>", published)
// 	✏️ &#9999;📜&#128220;
		line := fmt.Sprintf(`<li><a href="%s" target="_blank">%s &#128220;%s 
				</a><p>%s<p></li>`,link, date, title, description)
		result = fmt.Sprintf("%s\n%s", result, line)
	}
	result = fmt.Sprintf("%s\n</ul>", result)	
	store(signature, result)
	return result
}

func new_request(url string) string {
	var answer string = ""
//	t := time.Now().Add(2 * time.Second)
//	ctx, cancel := context.WithCancel(context.TODO())
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 2 * time.Second,
			}).Dial,
		TLSHandshakeTimeout:   2 * time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		},
	}
	reqm, _ := http.NewRequest("GET", url, nil)
	reqm.Header.Set("Content-Type", "text/html")
	content, err := client.Do(reqm)
	if err != nil {
		fmt.Println(err)
		if content != nil {
			fmt.Println("statusCode: ", content.StatusCode)
		}
		return answer 
	}
	value, err := ioutil.ReadAll(content.Body)
	if err != nil {
		fmt.Println(err)
		return answer 
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

func store(signature, value string) string {
	cacheSignature := hash(signature)
	if len(value) >= MIN_SIZE_FILE_CACHE {
		value = storeInFile(signature, value)
	}
	cache.Add(cacheSignature, cacheRecord{value, time.Now()})
	return fmt.Sprintf("%x", cacheSignature) 
} 

func storeInFile(signature, value string) string {
	if _, err := os.Stat(CACHE_DIR); os.IsNotExist(err) {
		err = os.Mkdir(CACHE_DIR, 0755)
		if err != nil {
			fmt.Printf("error %s", err)
		}	
	}
	filename := fmt.Sprintf("file:%x.txt", hash(signature))
	err := os.WriteFile(CACHE_DIR+"/"+filename, []byte(value), 0644)
	if err != nil {
		fmt.Printf("error %s", err)
	}
	return filename 
}

func readFile(filename string) string {
	result, err := os.ReadFile(filename)	
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		fmt.Printf("error %s", err)
		return ""
	}
	return string(result)
}

func get(signature string) (cacheRecord, bool) {
	cacheSignature := hash(signature)
	record, found := cache.Get(cacheSignature);
	if found && record.value != "" {
		if strings.Compare(record.value, fmt.Sprintf("file:%x.txt", cacheSignature)) == 0 {
			filename := fmt.Sprintf("%s/file:%x.txt", CACHE_DIR, cacheSignature)
			record.value = readFile(filename)	
		}
	}
	return record, found
}

func hash(signature string) [HASHSIZE]byte {
	return md5.Sum([]byte(signature))
}
