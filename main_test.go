package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	//	"time"
)

func TestCache(t *testing.T) {
	const hashsize int = md5.Size
	CACHE_DIR = "test_cache"
	tests := []struct {
		value     string
		signature string
		expFound  bool
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
		value    string
		selected bool
		exp      string
	}{
		{"Hello", false, "<option value=\"Hello\">Hello</option>"},
		{"World", true, "<option value=\"World\" selected>World</option>"},
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
	var exp string = ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		neoFile, err := os.Open("test-data/neovlivni_test.atom")
		if err != nil {
			fmt.Println(err)
		}
		defer neoFile.Close()
		byteWeather, _ := io.ReadAll(neoFile)
		w.Write(byteWeather)
	}))
	testFile, err := os.ReadFile("test-data/neovlivni_test.txt")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	exp = strings.TrimSuffix(string(testFile), "\n")
	got := rss_feed_neovlivni(ts.URL)
	if strings.Compare(got, exp) != 0 {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
	// cached
	got = rss_feed_neovlivni(ts.URL)
	if strings.Compare(got, exp) != 0 {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}

/*
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
*/

/*
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
*/

/*
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
*/

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
		weatherFile, err := os.Open("test-data/weatherreport_test")
		if err != nil {
			fmt.Println(err)
		}
		defer weatherFile.Close()
		byteWeather, _ := io.ReadAll(weatherFile)
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

func TestBtcXmr(t *testing.T) {
	var exp = "1<b style=\"color: gold;\">BTC</b> 2345.43$" +
		" 1<b style=\"color: #999;\">XMR</b> 2345.43$"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "2345.432132\n")
	}))
	defer ts.Close()

	if got := getBtcXmr(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
	//	try cache
	if got := getBtcXmr(ts.URL); got != exp {
		t.Errorf("Expected '%s' but, got '%s'", exp, got)
	}
}

func TestRssCtk(t *testing.T) {
	var exp string = ""
	tests := []struct {
		file            string
		nTitles         int
		showDescription bool
	}{
		{"test-data/ctk_test1.txt", 101, false},
		{"test-data/ctk_test.txt", -1, true},
	}

	for _, test := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/xml")
			ctkRssFile, err := os.Open("test-data/cr_test.rss")
			if err != nil {
				t.Errorf("Error: %s", err)
			}
			defer ctkRssFile.Close()
			byteRss, _ := io.ReadAll(ctkRssFile)
			w.Write(byteRss)
		}))
		defer ts.Close()
		testFile, err := os.ReadFile(test.file)
		if err != nil {
			t.Errorf("Error: %s", err)
		}
		exp = strings.TrimSuffix(string(testFile), "\n")
		got := rss_feed_ctk(ts.URL, test.nTitles, test.showDescription)
		if strings.Compare(got, exp) != 0 {
			t.Errorf("Expected '%s' but, got '%s' '%t'", exp, got, test.showDescription)
		}
		//cached
		got = rss_feed_ctk(ts.URL, test.nTitles, test.showDescription)
		if strings.Compare(got, exp) != 0 {
			t.Errorf("Expected '%s' but, got '%s' '%t'", exp, got, test.showDescription)
		}
	}
}

func TestCleanUpCache(t *testing.T) {
	CACHE_DIR = "test_cache"
	dirRead, _ := os.Open(CACHE_DIR)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		file := dirFiles[index]
		filename := file.Name()
		if err := os.Remove(CACHE_DIR + "/" + filename); err != nil {
			t.Errorf("error %s", err)
		}
	}
	if err := os.Remove(CACHE_DIR + "/"); err != nil {
		t.Errorf("error %s", err)
	}
}
