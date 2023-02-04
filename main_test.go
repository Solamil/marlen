package main

import (
	"net/http"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
	"os"
	"crypto/md5"
	"io/ioutil"
)

func TestCache(t *testing.T) {
	const hashsize int = md5.Size
	CACHE_DIR = "test_cache"
	tests := []struct {
		value string
		signature string 
		expFound bool
	}{
		{"Test hash", "TestCache", true},
		{`Test hash kdsafk sdkfakj j kfdksfkfkdskfkajdfsk jdskafjk ksjdfakdsf
		  dskfajksdfjakdfj dskafkdsjfkasdfk`, "TestCache1", true},
		{"", "TestCache2", true},
	}
	if HASHSIZE != hashsize {
		t.Errorf("Hash size is %d, but it is tested to %d.", HASHSIZE, hashsize)	
	}
	for _, test := range tests {
		store(test.signature, test.value)

		if got, found := get(test.signature); got.value != test.value {
			t.Errorf("Expected '%s', but got '%s'", test.value, got.value)
			if found != test.expFound {
				t.Errorf("Expected '%t', but got '%t'", test.expFound, found)
			}
		}
		if len(test.value) >= MIN_SIZE_FILE_CACHE {
			filename := fmt.Sprintf("%s/file:%x.txt", CACHE_DIR, hash(test.signature))
			if err := os.Remove(filename); err != nil {
				t.Errorf("error %s", err)
			}
		}
	}
}

func TestOptionTag(t *testing.T) {
	tests := []struct {
		value string
		selected bool
		exp string
	}{
		{"Hello", false, "<option value=\"Hello\">Hello</option>" },
		{"World", true, "<option value=\"World\" selected>World</option>" },
	}
	for _, test := range tests {
		if got := getHTMLOptionTag(test.value, test.value, test.selected); got != test.exp {
			t.Errorf("at input '%s' expected '%s', but got '%s'", test.value, test.exp, got)
		}
	}
}

func TestWeatherInfo(t *testing.T) {
	var exp string = "07:56:25 16:00:17 🌗"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `"07:56:25 16:00:17 🌗"`)
	}))
	defer ts.Close()

	if got := get_weather_info(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
		
	}
	
}

func TestFailWeatherInfo(t *testing.T) {
	var exp string = ""

	if got := get_weather_info("localhost:8080"); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
		
	}
	
}
func TestHolyTrinity(t *testing.T) {
	var exp string = "💵1€ 23.82Kč 1$ 21.93Kč 1£ 27.11Kč\n"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, "💵1€ 23.82Kč 1$ 21.93Kč 1£ 27.11Kč")
	}))
	defer ts.Close()
	got := get_holy_trinity(ts.URL)
	if got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
	//cached
	got = get_holy_trinity(ts.URL)
	if got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}

func TestRssFeedNeovlivni(t *testing.T) {
	var exp string = `<h3><a href="https://neovlivni.cz" target="_blank">Neovlivní – investigativní deník o vlivu a lidech</a></h3>
<ul>
<li><a href="https://neovlivni.cz/druhe-kolo-babis-vs-pavel-nerudova-vyzvala-k-podpore-generala/" target="_blank"><span class="date">2023-01-14T15:05:52Z</span> &#9999;-neo - &#128220;Druhé kolo: Babiš vs. Pavel. Nerudová vyzvala k podpoře generála</a></li>
<li><a href="https://neovlivni.cz/sabina-slonkova-zbabeleho-a-neschopneho-prezidenta-nepotrebujeme/" target="_blank"><span class="date">2023-01-13T06:10:19Z</span> &#9999;Sabina Slonková &#128220;Sabina Slonková: Zbabělého a neschopného prezidenta nepotřebujeme</a></li>
<li><a href="https://neovlivni.cz/klany-kolem-putina-jsou-jako-mafie-zalezi-na-tom-ktery-vyhraje/" target="_blank"><span class="date">2023-01-12T05:50:51Z</span> &#9999;Editor &#128220;Klany kolem Putina jsou jako mafie. Záleží na tom, který vyhraje</a></li>
<li><a href="https://neovlivni.cz/ucetni-skladka-ci-nepodpora-co-v-kampani-nezaznelo-ale-presto-vyvolalo-vasne/" target="_blank"><span class="date">2023-01-11T06:50:02Z</span> &#9999;Pavel Vrabec &#128220;Účetní, skládka či nepodpora. Co v kampani nezaznělo, ale přesto vyvolalo vášně</a></li>
<li><a href="https://neovlivni.cz/na-okraj-schuzky-s-macronem-co-maji-francouzi-na-babise/" target="_blank"><span class="date">2023-01-11T05:46:46Z</span> &#9999;Editor &#128220;Za kulisy schůzky s Macronem: Co mají Francouzi na Babiše</a></li>
</ul>`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		neoFile, err := os.Open("BwJLymVb_test.atom")
		if err != nil {
			fmt.Println(err)
		}
		defer neoFile.Close()
		byteWeather, _ := ioutil.ReadAll(neoFile)
		w.Write(byteWeather)		
	}))
	got := rss_feed_neovlivni(ts.URL)
	if got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}

}

func TestIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(index_handler)

	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestCookieIndexHandler(t *testing.T) {
 	req, err := http.NewRequest("GET", "/", nil)
 	if err != nil {
 		t.Fatal(err)
 	}
	date := time.Now()
	var place http.Cookie
	place.Name = "place"
	place.Value = "Prague"
	place.Expires = date.AddDate(1, 0, 0)
	place.Domain = "127.0.0.1"
	place.Path = "/"
	place.HttpOnly = false
	place.Secure = true
	req.AddCookie(&place)

	var lang http.Cookie
	lang.Name = "lang"
	lang.Value = "de-DE"
	lang.Expires = date.AddDate(1, 0, 0)
	lang.Domain = "127.0.0.1"
	lang.Path = "/"
	lang.HttpOnly = false
	lang.Secure = true
	req.AddCookie(&lang)

	var bg http.Cookie
	bg.Name = "bg"
	bg.Value = "442244"
	bg.Expires = date.AddDate(1, 0, 0)
	bg.Domain = "127.0.0.1"
	bg.Path = "/"
	bg.HttpOnly = false
	bg.Secure = true
	req.AddCookie(&bg)
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(index_handler)

	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestUrlParamsIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/?lang=de-DE&location=Prague&bg=442244", nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(index_handler)

	handler.ServeHTTP(recorder, req)
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestDailyWttrInfo(t *testing.T) {
	var exp string = "07:56:25 16:00:17 🌗"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `"07:56:25 16:00:17 🌗"`)
	}))
	defer ts.Close()

	if got := get_daily_wttr_info(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
//	reading from cache
	if got := get_daily_wttr_info(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}
func TestForecast(t *testing.T) {
	var exp string = "☔0% 🥶-12° 🌞-5°\n☔0% 🥶-8° 🌞0°\n☔0% 🥶-7° 🌞0°"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		weatherFile, err := os.Open("weatherreport_test")
		if err != nil {
			fmt.Println(err)
		}
		defer weatherFile.Close()
		byteWeather, _ := ioutil.ReadAll(weatherFile)
		w.Write(byteWeather)		
	}))
	defer ts.Close()
	if got := get_forecast(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}

func TestNameDay(t *testing.T) {
	var exp = "David"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "David")
	}))
	defer ts.Close()
	if got := get_name_day(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}
func TestCleanUpCache(t *testing.T) {
	CACHE_DIR = "test_cache"
	dirRead, _ := os.Open(CACHE_DIR)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range(dirFiles) {
		file := dirFiles[index]
		filename := file.Name()
		if err := os.Remove(CACHE_DIR+"/"+filename); err != nil {
			t.Errorf("error %s", err)
		}
	}
	if err := os.Remove(CACHE_DIR+"/"); err != nil {
		t.Errorf("error %s", err)
	}
}
