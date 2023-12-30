package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"
	"net/http"
	"net/url"
	"strings"
	"path/filepath"
	//	"context"
	//	"html"
	"time"
	"text/template"

	"github.com/Solamil/marlen"
	"github.com/robfig/cron/v3"
)

type indexUrlParams struct {
	Lang     [1]string `json:"lang"`
	Location [1]string `json:"location"`
	Bg       [1]string `json:"bg"`
}

type indexDisplay struct {
	Bg             string
	Location       string
	WeatherInfo    string
	Ipv4address    string
	LocaleOptions  string
	Currency       string
	NameToday      string
	NameTmrw       string
	ForecastFirst  string
	ForecastSecond string
	WttrLink       string
	WttrSrc        string
	WttrInHolder   string
	BtcValue       string
	XmrValue       string
	LocalNews      string
	Tannoy         string
	// Crashnet       string
}
type feedsDisplay struct {
	Bg      string
	RssFeed string
}

const PORT = 8901
var WEB_DIR string = "web"
var STATIC_DIR string = filepath.Join(WEB_DIR, "static")
var indexBg string = "893531"
var fileSvatek string = filepath.Join(STATIC_DIR, "nameday_cz_sk_pretty.txt")
var svatekToday string = ""
var svatekTomorrow string = ""
var fileHolytrinity string = filepath.Join("rates", "svata_trojice.txt")
var location string = "Mnichovo Hradiště"
var lang string = "cs-CZ"

var wttrUrl string = "https://wttr.in"
var fakemoneyUrl string = "https://rate.sx"
var localtownUrl string = "https://www.mnhradiste.cz/rss"

// var wttrInHolders = map[string]string{
// 	"en": "Weather in...",
// 	"de": "Wetter für...",
// 	"cs": "Počasí v...",
// }

var countryFlags = map[string]string{
//	"en-US": "🇺🇸",
	"de-DE": "🇩🇪",
	"cs-CZ": "🇨🇿",
}

var indexTemplate *template.Template
var feedsTemplate *template.Template

func main() {
	port := flag.Int("port", PORT, "Port for the server to listen on")
	flag.Parse()
	
	startupScripts()
	cronJobs()

	pathTemplate := filepath.Join(WEB_DIR, "template")
	indexTemplate, _ = template.ParseFiles(filepath.Join(pathTemplate, "index.html"), 
					filepath.Join(pathTemplate, "timelocalization.html"),
					filepath.Join(pathTemplate, "footer.html"))
	feedsTemplate, _ = template.ParseFiles(filepath.Join(pathTemplate, "feeds.html"),
					filepath.Join(pathTemplate, "footer.html"))

	fs := http.FileServer(http.Dir(STATIC_DIR))
	http.Handle("/web/", http.StripPrefix("/web/", fs))
	http.HandleFunc("/index.html", index_handler)
	http.HandleFunc("/feeds.html", feeds_handler)
	http.HandleFunc("/index", index_handler)
	http.HandleFunc("/feeds", feeds_handler)
	http.HandleFunc("/", index_handler)

	marlen.PrepareSvatekList(fileSvatek)
	t := time.Now()
	svatekToday = "Dnes má svátek "+marlen.GetSvatekName(t, "cs-CZ")
	svatekTomorrow = "Zítra má svátek "+marlen.GetSvatekName(t.AddDate(0, 0, 1), "cs-CZ")
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	var bg string = indexBg 
	var weatherInfo string = ""
	var forecastFirst string = ""
	var forecastSecond string = ""

	handle_req_params(r, &location, &lang, &bg)

	prefix := strings.Split(lang, "-")[0]

	var wg sync.WaitGroup
	wg.Add(4)
	wttrin := fmt.Sprintf("%s/%s", wttrUrl, location)
	forecastCh := make(chan string)
	go marlen.GetForecast(wttrin, forecastCh, &wg)
	forecastStr := <-forecastCh
	forecasts := strings.Split(forecastStr, "\n")
	if len(forecasts) >= 3 {
		weatherInfo = forecasts[0]
		forecastFirst = forecasts[1]
		forecastSecond = forecasts[2]
	}
	sunMoonUrl := fmt.Sprintf(`%s?format="%s"`, wttrin, "%S+%s+%m")
	sunMoonCh := make(chan string)
	go marlen.GetDailyWttrInfo(sunMoonUrl, sunMoonCh, &wg)
	sunMoonStr := <-sunMoonCh
	if len(sunMoonStr) != 0 {
		sunMoon := strings.Split(sunMoonStr, " ")
		if len(sunMoon) == 3 && len(forecasts) > 0 {
			weatherInfo = "🌅 " + sunMoon[0] + " 🌇" + sunMoon[1] + " " + sunMoon[2] + " " + forecasts[0]
		} 
	}

	var i indexDisplay
	i.NameToday = svatekToday 
	i.NameTmrw = svatekTomorrow
	i.Bg = bg
	i.Location, _ = url.QueryUnescape(location)
	i.WeatherInfo = weatherInfo
	i.ForecastFirst = forecastFirst
	i.ForecastSecond = forecastSecond
	if len(r.Header["X-Real-Ip"]) > 0 {
		i.Ipv4address = r.Header["X-Real-Ip"][0]
	}
	currency := make(chan string)
	go marlen.CnbCurrency(fileHolytrinity, currency, &wg)
	i.Currency = <-currency
	i.WttrLink =  fmt.Sprintf("%s?lang=%s", wttrin, prefix)
//	i.WttrSrc = fmt.Sprintf("%s_0pq_transparency=255_background=%s_lang=%s.png", wttrin, bg, prefix)

	// i.WttrInHolder = wttrInHolders[prefix]
	i.LocaleOptions = getLocaleTags(lang) 
	cryptoCurr := marlen.FakeMoney(fakemoneyUrl)
	i.BtcValue = cryptoCurr[0]
	i.XmrValue = cryptoCurr[1]
	// i.CryptoCurrency = marlen.FakeMoney(fakemoneyUrl)
//	foneStr := make(chan string)
//	go marlen.RssCrashnet("https://www.crash.net/rss/f1", "Crash Net - F1", "https://crash.net", 5, foneStr, &wg )
//	motogpStr := make(chan string)
//	go marlen.RssCrashnet("https://www.crash.net/rss/motogp", "Crash Net - MotoGP", "https://crash.net", 5, motogpStr, &wg )
	// nitterStr := make(chan string)
	// go marlen.RssCrashnet("https://www.nitter.cz/jeremyclarkson/rss", "nitter - JC", "https://nitter.cz/JeremyClarkson", 3, nitterStr, &wg )
//	i.Crashnet = fmt.Sprintf("%s \n %s", <-foneStr, <-motogpStr)
	tannoy := make(chan string)
	go marlen.RssLocalplaceRoutine(localtownUrl, 2, true, true, tannoy, &wg)
	i.Tannoy = <-tannoy
	wg.Wait()	
	i.LocalNews = marlen.RssLocalplace(localtownUrl, 5, false, true)
	indexTemplate.Execute(w, i)

}

func feeds_handler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var rssFeed string = ""
	var location string = ""
	var lang string = "cs-CZ"
	var bg string = "442244"
	var i feedsDisplay

	handle_req_params(r, &location, &lang, &bg)

	if lang == "cs-CZ" {
		wg.Add(6)
		var ctkUrl string = "https://www.ceskenoviny.cz/sluzby/rss"
		ctkCr := make(chan string)
		go marlen.RssCtk(ctkUrl+"/cr.php", 5, true, ctkCr, &wg)
		ctkSvet := make(chan string)
		go marlen.RssCtk(ctkUrl+"/svet.php", 5, true, ctkSvet, &wg)
		ctkEko := make(chan string)
		go marlen.RssCtk(ctkUrl+"/ekonomika.php", 5, true, ctkEko, &wg)
		ctkSport := make(chan string)
		go marlen.RssCtk(ctkUrl+"/sport.php", 3, false, ctkSport, &wg)

		// ctkCr := marlen.RssCtk(ctkUrl+"/cr.php", 5, true)
		// ctkSvet := marlen.RssCtk(ctkUrl+"/svet.php", 5, true)
		// ctkEko := marlen.RssCtk(ctkUrl+"/ekonomika.php", 5, true)
		// ctkSport := marlen.RssCtk(ctkUrl+"/sport.php", 3, false)
		hrad := make(chan string)
		go marlen.RssCtk("https://www.hrad.cz/cs/pro-media/rss/tiskove-zpravy.xml", 5, false, hrad, &wg)
		neovlivni := make(chan string)
		go marlen.AtomFeed("https://neovlivni.cz/feed/atom/", neovlivni, &wg)

		render_feeds := fmt.Sprintf(`%s <br><hr> %s <br><hr>
			    %s <br><hr> %s <br><hr> %s <br><hr> %s`, <-neovlivni, <-hrad, <-ctkCr, <-ctkSvet, <-ctkEko, <-ctkSport )
		rssFeed = render_feeds
		wg.Wait()
	} else if lang == "de-DE" {
		wg.Add(1)
		taggeshau := make(chan string)
		go marlen.RssCtk("https://www.tagesschau.de/ausland/index~rss2.xml", 5, true, taggeshau, &wg)
		
		render_feeds := fmt.Sprintf(`%s <br><hr>`, <-taggeshau )
		wg.Wait()
		rssFeed = render_feeds
	} else if lang == "gb-GB" {
		wg.Add(1)
		theguardian := make(chan string)
		go marlen.RssCtk("https://www.theguardian.com/uk/rss", 7, true, theguardian, &wg)

		render_feeds := fmt.Sprintf(`%s <br><hr>`, <-theguardian)
		rssFeed = render_feeds
		wg.Wait()
	}
	i.Bg = "442244"
	i.RssFeed = rssFeed
	feedsTemplate.Execute(w, i)
}

func getLocaleTags(lang string) string {
	var localeTags string = ""
	var tag string = ""
	for key, value := range countryFlags {
		tag = getHTMLOptionTag(key, value, (key == lang))
		localeTags = strings.Join([]string{localeTags, tag}, "\n")
	}
	return localeTags
}

func handle_req_params(r *http.Request, location *string, lang *string, bg *string) {
	if c, err := r.Cookie("place"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		//		location, _ = url.QueryUnescape(value)
		*location = value
		//		fmt.Println(value)
	} else if err != nil {
		fmt.Println(err)
	}

	if c, err := r.Cookie("lang"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		*lang = value
	} else if err != nil {
		fmt.Println(err)
	}

	if c, err := r.Cookie("bgColor"); err == nil {
		value := strings.Split(c.String(), "=")[1]
		*bg = value
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
			*location = param.Location[0]
		}
		if len(param.Lang[0]) > 0 {
			*lang = param.Lang[0]
		}
		if len(param.Bg[0]) > 0 {
			*bg = param.Bg[0]
		}
	}

}

func startupScripts() {
	var wg sync.WaitGroup
	wg.Add(5)
//	go marlen.RunScriptRoutine(&wg, filepath.Join("scripts", "days-forecast.sh"), "0days")
//	go marlen.RunScriptRoutine(&wg, filepath.Join("scripts", "days-forecast.sh"), "0days", "forecastTemp")
//	go marlen.RunScriptRoutine(&wg, filepath.Join("scripts", "days-forecast.sh"), "0days", "forecastWind")
//	go marlen.RunScriptRoutine(&wg, filepath.Join("scripts", "days-forecast.sh"), "1days")
	go marlen.CalendarImgRoutine(&wg, "https://kalendar.beda.cz/pic/kalendar-m.png", 
					filepath.Join(STATIC_DIR, "pics", "kalendar-m.png"))
	// wg.Wait()

}
func cronJobs() {
	c := cron.New()
	c.AddFunc("40 14 * * 1-5", func() { marlen.RunScript(filepath.Join("scripts", "rates.sh")) })
	c.AddFunc("50 * * * *", func() { 
		marlen.RunScript(filepath.Join("scripts", "webcam.sh")) 
		// marlen.RunScript(filepath.Join("scripts", "sat-img.sh"))
	})
	c.Start()

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
