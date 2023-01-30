package main

import (
	"net/http"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
	"os"
	"io/ioutil"
)

func TestCache(t *testing.T) {
	const hashsize int = 16
	tests := []struct {
		value string
		expSignature [hashsize]byte
	}{
		{"Test hash", [hashsize]byte{19, 151, 200, 224, 151, 66, 63, 32, 153, 210, 159, 199, 33, 67, 179, 230}},
	}
	if HASHSIZE != hashsize {
		t.Errorf("Hash size is %d, but it is tested to %d.", HASHSIZE, hashsize)	
	}
	for _, test := range tests {
		if got := hash(test.value); got != test.expSignature {
			t.Errorf("at input '%s' expected '%b', but got '%s'", test.value, test.expSignature, got)
		}
		store(test.expSignature, test.value)

		if got, found := get(test.expSignature); got.value != test.value && found {
			t.Errorf("at input '%b' expected '%s', but got '%s'", test.expSignature, test.value, got.value)
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
