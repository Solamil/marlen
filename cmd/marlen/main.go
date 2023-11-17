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
	"text/template"
	"github.com/Solamil/marlen"
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
	OtherInfo      string
	LocaleOptions  string
	Currency       string
	NameDay        string
	ForecastFirst  string
	ForecastSecond string
	WttrLink       string
	WttrSrc        string
	WttrInHolder   string
	CryptoCurrency string
	LocalNews      string
	Tannoy         string
	Crashnet       string
}
type feedsDisplay struct {
	Bg      string
	RssFeed string
}

const PORT = 8901
var indexBg string = "893531"
var svatekFile string = filepath.Join("web", "nameday_cz_sk_pretty.txt")
var location string = "Mnichovo Hradiště"
var lang string = "cs-CZ"
// var svatekUrl string = "http://localhost:7903/today?pp"
// var currencyUrl string = "http://localhost:7902/holy_trinity?p"
var wttrUrl string = "https://wttr.in"
var fakemoneyUrl string = "https://rate.sx"
var localtownUrl string = "https://www.mnhradiste.cz/rss"

var WEB_DIR string = "web"
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
	http.HandleFunc("/pics/clouds_eu.webp", file_handler)
	http.HandleFunc("/pics/clouds_eu.gif", file_handler)
	http.HandleFunc("/pics/mhcam1.webp", file_handler)
	http.HandleFunc("/pics/snap-klub.webp", file_handler)
	http.HandleFunc("/pics/kalendar-m.png", file_handler)
	http.HandleFunc("/js/module-wttrin-widget.js", file_handler)
	http.HandleFunc("/cover.html", file_handler)
	http.HandleFunc("/traffic.html", file_handler)
	http.HandleFunc("/f1.html", file_handler)
	http.HandleFunc("/motogp.html", file_handler)
	http.HandleFunc("/cnb-rates.html", file_handler)
	http.HandleFunc("/svatek.html", file_handler)
	http.HandleFunc("/artix_arch.sh", file_handler)
	http.HandleFunc("/nameday_cz_sk_pretty.txt", file_handler)

	indexTemplate, _ = template.ParseFiles("web/index.html")
	http.HandleFunc("/index.html", index_handler)
	http.HandleFunc("/feeds.html", feeds_handler)
	http.HandleFunc("/", index_handler)

	marlen.PrepareSvatekList(svatekFile)
	// marlen.NewImgRequest("https://kalendar.beda.cz/pic/kalendar-m.png", "./web/pics/kalendar-m.png" )
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup
	var bg string = indexBg 
	var weatherInfo string = ""
	var forecastFirst string = ""
	var forecastSecond string = ""

	handle_req_params(r, &location, &lang, &bg)

	prefix := strings.Split(lang, "-")[0]

	wg.Add(6)
	wttrin := fmt.Sprintf("%s/%s", wttrUrl, location)
	forecastCh := make(chan string)
	go marlen.GetForecast(wttrin, forecastCh, &wg)
	forecastStr := <-forecastCh
	forecasts := strings.Split(forecastStr, "\n")
	if len(forecasts) >= 3 {
		forecastFirst = forecasts[1]
		forecastSecond = forecasts[2]
		weatherInfo = forecasts[0]
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
	i.NameDay = "📆Dnes má svátek "+marlen.GetSvatekNameToday("cs-CZ")
	i.Bg = bg
	i.Location, _ = url.QueryUnescape(location)
	i.WeatherInfo = weatherInfo
	i.ForecastFirst = forecastFirst
	i.ForecastSecond = forecastSecond
	i.OtherInfo = req_ip_address(r)
	currency := make(chan string)
	go marlen.CnbCurr(currency, &wg)
	i.Currency = <-currency
	i.WttrLink =  fmt.Sprintf("%s?lang=%s", wttrin, prefix)
	i.WttrSrc = fmt.Sprintf("%s_0pq_transparency=255_background=%s_lang=%s.png", wttrin, bg, prefix)

	// i.WttrInHolder = wttrInHolders[prefix]
	i.LocaleOptions = getLocaleTags(lang) 
	i.CryptoCurrency = marlen.FakeMoney(fakemoneyUrl)
	foneStr := make(chan string)
	go marlen.RssCrashnet("https://www.crash.net/rss/f1", "Crash Net - F1", "https://crash.net", 5, foneStr, &wg )
	motogpStr := make(chan string)
	go marlen.RssCrashnet("https://www.crash.net/rss/motogp", "Crash Net - MotoGP", "https://crash.net", 5, motogpStr, &wg )
	nitterStr := make(chan string)
	go marlen.RssCrashnet("https://www.nitter.cz/jeremyclarkson/rss", "nitter - JC", "https://nitter.cz/JeremyClarkson", 3, nitterStr, &wg )
	i.Crashnet = fmt.Sprintf("%s \n %s \n %s", <-foneStr, <-motogpStr, <-nitterStr)
	wg.Wait()	
	i.Tannoy = marlen.RssLocalplace(localtownUrl, 2, true, true)
	i.LocalNews = marlen.RssLocalplace(localtownUrl, 5, false, true)
	indexTemplate, _ = template.ParseFiles("web/index.html")
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
	feedsTemplate, _ = template.ParseFiles("web/feeds.html")
	feedsTemplate.Execute(w, i)
}

func file_handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Clean(WEB_DIR+r.URL.Path))
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

func req_ip_address(r *http.Request) string {
	if len(r.Header["X-Real-Ip"]) > 0 {
		return fmt.Sprintf("<a target=\"_blank\" href=\"https://www.whois.com/whois/%s\">🌐 %s</a>",
			r.Header["X-Real-Ip"][0], r.Header["X-Real-Ip"][0])
	}
	return "🌐 IPv4 address"

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
