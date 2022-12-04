package main

import (
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"io/ioutil"
	"strings"
	"html"
	"strconv"
	"time"
	"unicode/utf8"
	"os/exec"
	"crypto/md5"
	"text/template"
	"github.com/hashicorp/golang-lru/v2"
//	"os"
)

type userBaseResponse struct {
	CoinPrices struct {
		Btc string `json:"btc"` 
		Xmr string `json:"xmr"`
		CoinCode string `json:"coin_code"`
	} `json:"coins"`

	Weather struct {
		SunMoon string `json:"sun_moon"`
		HumLowHigh []string `json:"hum_low_high"`
		Location string `json:"location"`
	} `json:"weather"`

	CurrPrices struct {
		Code []string `json:"code"`
		Volume []string `json:"volume"`
		Value []string `json:"value"`
		CoinCode string `json:"coin_code"`
		Date string `json:"date"`
	} `json:"currs"`
}

type cacheRecord struct {
	value string
	expiry time.Time

}


type urlParams struct {
	Lang [1]string `json:"lang"`
	Location [1]string `json:"location"`
	CoinCode [1]string `json:"coin_code"`
	Bg [1]string `json:"bg"`
}

type indexDisplay struct {
	Bg string
	Location string
	WeatherInfo string
	Coins string
	Currency string
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

var currSymbols = map[string]string{
	"usd": "$",
	"eur": "€",
	"czk": "Kč",
	"gbp": "£",
	"btc": "BTC",
}

var baseResp userBaseResponse
var coinPrices = &baseResp.CoinPrices
var weather = &baseResp.Weather

var indexTemplate *template.Template

type userBaseRequest struct {
	CoinCode string `json:"coin_code"`	
	Param	 string `json:"param"`
	Location string `json:"location"`
}

func main() {
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/style.css")
	})
	http.HandleFunc("/pics/bitcoin-icon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/bitcoin-icon.svg")
	})
	http.HandleFunc("/pics/monero-icon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pics/monero-icon.svg")
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

	indexTemplate, _ = template.ParseFiles("web/index.html")
	http.HandleFunc("/index.html", index_handler)
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/base_info", base_handler)
	http.ListenAndServe(":8900", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	var location string = "Zdar"
	var coinCode string = "usd"
	var bg string = "893531"
	var lang string = "en-US"

	if c, err := r.Cookie("coin_code"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		coinCode = value 
	} else if err != nil {
		fmt.Println(err)
	}
	if c, err := r.Cookie("place"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		location, _ = url.QueryUnescape(value)
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
		var param *urlParams
		json.Unmarshal(js, &param)		
		if len(param.Location[0]) > 0 {
			location = param.Location[0]
		}
		if len(param.Lang[0]) > 0 {
			lang = param.Lang[0]
		}
		if len(currSymbols[param.CoinCode[0]]) > 0 {
			coinCode = param.CoinCode[0]
		}
		if len(param.Bg[0]) > 0 {
			bg = param.Bg[0]
			fmt.Println(bg)
		}
	}
	prefix := strings.Split(lang, "-")[0]
	wttrSrc := "https://wttr.in/"+location+"_0pq_transparency=255_background="+bg+"_lang="+prefix+".png"
	wttrLink := "https://wttr.in/"+location+"?lang="+prefix
	forecastStr := get_forecast(location)
	forecasts := strings.Split(forecastStr, "\n")
	sunMoonStr := get_sun_moon_info(location)
	sunMoon := strings.Split(sunMoonStr, " ")
	btcStr := get_crypto_curr(coinCode, "btc")
	btc, _ := strconv.ParseFloat(btcStr, 32)
	xmrStr := get_crypto_curr(coinCode, "xmr")
	xmr, _ := strconv.ParseFloat(xmrStr, 32)
	coins := fmt.Sprintf("%s %.2f%s %s %.2f%s", 
			"<img src=\"/pics/bitcoin-icon.svg\">", btc, currSymbols[coinCode],
			"<img src=\"/pics/monero-icon.svg\">", xmr, currSymbols[coinCode])

	var i indexDisplay
	i.Bg = "#"+bg
	i.Location = location
	i.WeatherInfo = "🌅 "+sunMoon[0]+" 🌇"+sunMoon[1]+" "+sunMoon[2]+" "+forecasts[0]
	i.ForecastFirst = forecasts[1]
	i.ForecastSecond = forecasts[2]
	i.Coins = coins
	i.Currency = get_currency_rates()
	i.WttrLink = wttrLink
	i.WttrSrc = wttrSrc
	i.WttrInHolder = wttrInHolders[prefix]
	indexTemplate.Execute(w, i)

}
func base_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var baseRequest userBaseRequest
	json.Unmarshal(body, &baseRequest)

	if baseRequest.CoinCode  == "" { 
		baseRequest.CoinCode = "usd"
	}
	baseRequest.CoinCode = html.EscapeString(baseRequest.CoinCode)
	coinPrices.CoinCode = baseRequest.CoinCode

	coinPrices.Btc = get_crypto_curr(coinPrices.CoinCode, "btc") 
	coinPrices.Xmr = get_crypto_curr(coinPrices.CoinCode, "xmr")

	if len(coinPrices.CoinCode) > 1 && baseRequest.Param == "conversion" {
		raw, _ := json.Marshal(coinPrices)
		w.Write(raw)
	} else {
		baseRequest.Location = baseRequest.Location
		if baseRequest.Location == "" {
			baseRequest.Location = "Zdar"
		}
		get_weather(baseRequest.Location)
		get_currency_rates()
		raw, err := json.Marshal(&baseResp)
		if err != nil {
			fmt.Println(err)
		}
		w.Write(raw)
	}

	
}

func get_weather(location string) {
	forecast := get_forecast(location)
	weather.HumLowHigh = strings.Split(forecast, "\n")
	weather.Location = location
	weather.SunMoon = get_sun_moon_info(location)
}

func get_crypto_curr(coinCode, name string) string {
	signature := fmt.Sprintf(`%s:%s`, coinCode, name)
	cacheSignature := hash(signature)
	record, found := get(cacheSignature)
	var answer string = ""
	if found {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Minute * 5)

		if  record.value != "" && d.After(now) {
			answer = record.value
			return answer
		}
	}
	value := get_coin_price(coinCode, name)
	answer = store(cacheSignature, value)
	return answer
}

func get_sun_moon_info(location string) string {
	format := "%S+%s+%m"	
	signature := fmt.Sprintf(`%s:%s`, location, format)
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
	value := get_weather_info(format, location)
	answer = store(cacheSignature, value)
	return answer
}

func get_weather_info(format, location string) string {
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
	out, err := ioutil.ReadAll(content.Body)

	if err != nil {
		fmt.Println(err)
		return ""
	}
	str_out := strings.ReplaceAll(string(out), "\"", "")

	return string(str_out)
}

func get_forecast(location string) string {
	signature := fmt.Sprintf(`%s:%s`, location, "forecast")
	cacheSignature := hash(signature)
	var answer string = ""
	record, found := get(cacheSignature)

	if found {
		now := time.Now()
		d := record.expiry
		d = d.Add(time.Hour * 6)
		if record.value != "" && d.After(now) {
			answer = record.value
			return answer
		}
	}
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

	value := strings.Join([]string{hum_low_high, hum_low_high_next, hum_low_high_next2}, "\n")
	answer = store(cacheSignature, value)
	return answer

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
	var rates string = ""
	var answer string = ""
	signature := "cnb-rates"
	cacheSignature := hash(signature)

	record, found := get(cacheSignature)

	if found {
		dateStr := string(now.Day())+"."+string(int(now.Month()))+"."+string(now.Year())
		d := strings.Split(record.value, " ")[0]
		
		if d == dateStr || now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			rates = record.value
		} else {
			rates = getCnbRates()
			rates = store(cacheSignature, rates)
		}

	} else {
		rates = getCnbRates()
		rates = store(cacheSignature, rates)
	}
	exchRates := strings.Split(rates, "\n")


	gbpCurr := strings.Split(exchRates[33], "|")
	gbpValue := gbpCurr[len(gbpCurr)-1]
	gbpValue = strings.Replace(gbpValue, ",", ".", 1)
	gbpCode := gbpCurr[len(gbpCurr)-2]
	gbpVolume := gbpCurr[len(gbpCurr)-3]

	eurCurr := strings.Split(exchRates[7], "|")
	eurValue := eurCurr[len(eurCurr)-1]
	eurValue = strings.Replace(eurValue, ",", ".", 1)
	eurCode := eurCurr[len(eurCurr)-2]
	eurVolume := eurCurr[len(eurCurr)-3]

	usdCurr := strings.Split(exchRates[32], "|")
	usdValue := usdCurr[len(usdCurr)-1]
	usdValue = strings.Replace(usdValue, ",", ".", 1)
	usdCode := usdCurr[len(usdCurr)-2]
	usdVolume := usdCurr[len(usdCurr)-3]
	usdValueFloat, _ := strconv.ParseFloat(usdValue, 32)
	eurValueFloat, _ := strconv.ParseFloat(eurValue, 32)
	gbpValueFloat, _ := strconv.ParseFloat(gbpValue, 32)
	answer = fmt.Sprintf("%s$ %.2fKč %s€ %.2fKč %s£ %.2fKč", usdVolume, usdValueFloat,
								 eurVolume, eurValueFloat,
							         gbpVolume, gbpValueFloat)

	var currPrices = userBaseResponse{}.CurrPrices
	currPrices.Code = append(currPrices.Code, gbpCode)
	currPrices.Code = append(currPrices.Code, eurCode)
	currPrices.Code = append(currPrices.Code, usdCode)
	currPrices.Volume = append(currPrices.Volume, gbpVolume)
	currPrices.Volume = append(currPrices.Volume, eurVolume)
	currPrices.Volume = append(currPrices.Volume, usdVolume)
	currPrices.Value = append(currPrices.Value, gbpValue)
	currPrices.Value = append(currPrices.Value, eurValue)
	currPrices.Value = append(currPrices.Value, usdValue)
//	currPrices.Json = json
	currPrices.CoinCode = "czk"
	currPrices.Date = strings.Split(exchRates[0], " ")[0]

	baseResp.CurrPrices = currPrices	
	return answer
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

func download_sat_images() {
	output, err := exec.Command("/bin/sh", "sat-img.sh").Output()
	fmt.Println(output)
	if err != nil {
		fmt.Printf("error %s", err)
	}

}
