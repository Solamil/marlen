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
