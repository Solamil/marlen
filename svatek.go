package marlen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const LINES int = 367
const COLUMNS int = 3

var svatekList = [LINES][COLUMNS]string{{}}

func PrepareSvatekList(name string) {
	for i, line := range readFile(name) {
		d := strings.Split(line, "|")
		for j, value := range d {
			svatekList[i][j] = value
		}
	}
}

func GetSvatekName(t time.Time, country string) string {
	var result string = ""
	names := getLineByDate(t)
	col := GetIndex(svatekList[0][:], country)
	if col > 0 && col < COLUMNS {
		result = names[col]
	}
	return result
}

func getLineByDate(t time.Time) []string {
	var result []string
	date := fmt.Sprintf("%d.%d", t.Day(), int(t.Month()))

	for i := 0; i < LINES; i++ {
		if date == svatekList[i][0] {
			result = svatekList[i][:]
			break
		}
	}
	return result
}

func getDate(name string, col int) string {
	var result string = ""
	for i := 0; i < LINES && col > 0 && col < COLUMNS; i++ {
		if name == svatekList[i][col] {
			result = svatekList[i][0]
			break
		}
		names := strings.Split(svatekList[i][col], "/")

		if j := GetIndex(names, name); j != -1 {
			result = svatekList[i][0]
			break
		}
	}
	return result
}

func readFile(name string) []string {
	file, err := os.Open(name)
	if err != nil {
		fmt.Sprintf("Failed to open %s", name)
		return []string{}
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string

	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()

	return text
}

func GetIndex(list []string, value string) int {
	var index int = -1
	for i := 0; i < len(list); i++ {
		if list[i] == value {
			index = i
			break
		}
	}
	return index
}

// Compute the exact day for each year
// credits to https://kalendar.beda.cz/vypocet-velikonocni-nedele-v-ruznych-programovacich-jazycich
func Velikonoce(rok int) time.Time {
	if rok <= 1583 {
		rok = 1584
	}
	zlateCislo := (rok % 19) + 1
	julEpakta := (11 * zlateCislo) % 30
	stoleti := int(rok/100) + 1
	slunecniOprava := int(3 * (stoleti - 16) / 4)
	mesicniOprava := int(8 * (stoleti - 15) / 25)
	epakta := (julEpakta - 10 - slunecniOprava + mesicniOprava) % 30
	if epakta < 0 {
		epakta += 30
	}
	tmp := epakta
	if epakta == 24 || (epakta == 25 && zlateCislo > 11) {
		tmp += 1
	}
	pfm := 0 // Paschal Full Moon
	if tmp < 24 {
		pfm = 44 - tmp
	} else {
		pfm = 74 - tmp
	}

	gregOprava := 10 + slunecniOprava
	denTydnePfm := (rok + (int)(rok/4) - gregOprava + pfm) % 7
	if denTydnePfm < 0 {
		denTydnePfm += 7
	}
	velNedele := pfm + 7 - denTydnePfm
	var t time.Time
	if velNedele < 32 {
		t = time.Date(rok, time.March, velNedele, 0, 0, 0, 0, time.UTC)
	} else {
		t = time.Date(rok, time.April, velNedele-31, 0, 0, 0, 0, time.UTC)
	}
	return t
}

func Denmatek(year int) time.Time {
	// Second sunday at the month of May
	if year < 1923 {
		year = 1923 // At Czechoslovakia started in 1923
	}

	t := time.Date(year, time.May, 1, 0, 0, 0, 0, time.UTC)

	if t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 7)
	} else {
		tmp := 7 - int(t.Weekday()) + 7
		t = t.AddDate(0, 0, tmp)
	}
	return t
}

func Denotcu(year int) time.Time {
	// Third sunday at the month of June
	if year < 1910 {
		year = 1910
	}
	t := time.Date(year, time.June, 1, 0, 0, 0, 0, time.UTC)

	if t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 7+7)
	} else {
		thirdSunday := 7 - int(t.Weekday()) + 7 + 7
		t = t.AddDate(0, 0, thirdSunday)
	}
	return t
}

func Summertime(year int, end bool) time.Time {
	if year < 1979 {
		year = 1979
	}
	month := time.March
	if end {
		month = time.October
	}
	return lastSundayofmonth(year, month)
}

func lastSundayofmonth(year int, month time.Month) time.Time {
	t := time.Date(year, month, 31, 0, 0, 0, 0, time.UTC)
	if t.Weekday() == time.Sunday {
		return t
	}
	return t.AddDate(0, 0, -int(t.Weekday()))
}
