#!/bin/sh
url=${1:-"https://wttr.in/Mnichovo%20Hradiste"}
dir="$HOME/.local/src/marlen"
weatherreport="$dir/weatherreport"
place=$(echo "$place" | sed "s/ /%20/" )

parse_hum_temp(){
	line_temp=${1:-13}
	line_hum=${2:-16}
	printf "%s" "$(sed "${line_hum}q;d" "$weatherreport" |
		grep -wo "[0-9]*%" | sort -rn | sed "s/^/☔/g;1q" | tr -d '\n')"
	sed "${line_temp}q;d" "$weatherreport" | grep -o "m\\([-+]\\)*[0-9]\\+" | sed 's/+//g' | sort -n -t 'm' -k 2n | sed -e 1b -e '$!d' | tr '\n|m' ' ' | awk '{print " 🥶" $1 "°","🌞" $2 "°"}'

}

curl --connect-timeout 2 --fail --max-time 5 "$url" --output "$weatherreport"

parse_hum_temp 13 16
parse_hum_temp 23 26
parse_hum_temp 33 36

