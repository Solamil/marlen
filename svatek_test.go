package svatek

import (
	"time"
	"testing"
)

func TestVelikonoce(t *testing.T) {
	tests := []struct {
		year int 
		expected time.Time 
	}{
		{1583, time.Date(1584, time.April, 1, 0, 0, 0, 0, time.UTC)},
		{2016, time.Date(2016, time.March, 27, 0, 0, 0, 0, time.UTC)},
		{2100, time.Date(2100, time.March, 28, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		if got := Velikonoce(test.year); got != test.expected {
			t.Errorf("at input '%d' expected '%s', but got '%s'", test.year, test.expected, got)
		}

	}
}

func TestDenmatek(t *testing.T) {
	tests := []struct {
		year int 
		expected time.Time 
	}{
		{1583, time.Date(1923, time.May, 13, 0, 0, 0, 0, time.UTC)},
		{2016, time.Date(2016, time.May, 8, 0, 0, 0, 0, time.UTC)},
		{2100, time.Date(2100, time.May, 9, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		if got := Denmatek(test.year); got != test.expected {
			t.Errorf("at input '%d' expected '%s', but got '%s'", test.year, test.expected, got)
		}

	}
}
 
func TestDenotcu(t *testing.T) {
	tests := []struct {
		year int 
		expected time.Time 
	}{
		{1583, time.Date(1910, time.June, 19, 0, 0, 0, 0, time.UTC)},
		{1958, time.Date(1958, time.June, 15, 0, 0, 0, 0, time.UTC)},
		{2016, time.Date(2016, time.June, 19, 0, 0, 0, 0, time.UTC)},
		{2100, time.Date(2100, time.June, 20, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		if got := Denotcu(test.year); got != test.expected {
			t.Errorf("at input '%d' expected '%s', but got '%s'", test.year, test.expected, got)
		}

	}
}

func TestSummertime(t *testing.T) {
	tests := []struct {
		year int 
		end bool
		expected time.Time 
	}{
		{2016, false, time.Date(2016, time.March, 27, 0, 0, 0, 0, time.UTC)},
		{2016, true, time.Date(2016, time.October, 30, 0, 0, 0, 0, time.UTC)},
		{-2000, false, time.Date(1979, time.March, 25, 0, 0, 0, 0, time.UTC)},
		{-2000, true, time.Date(1979, time.October, 28, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		if got := Summertime(test.year, test.end); got != test.expected {
			t.Errorf("at input '%d', '%t' expected '%s', but got '%s'", test.year, test.end, test.expected, got)
		}

	}
}

func TestLastSundayofmonth(t *testing.T) {
	tests := []struct {
		year int 
		month time.Month	
		expected time.Time 
	}{
		{2016, time.January, time.Date(2016, time.January, 31, 0, 0, 0, 0, time.UTC)},
		{2016, time.December, time.Date(2016, time.December, 25, 0, 0, 0, 0, time.UTC)},
	}
	for _, test := range tests {
		if got := lastSundayofmonth(test.year, test.month); got != test.expected {
			t.Errorf("at input '%d', '%d' expected '%s', but got '%s'", test.year, test.month, test.expected, got)
		}

	}
}
