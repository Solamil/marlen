#!/bin/sh
url=${1:-""}
weatherreport="./weatherreport"
line_temp=${2:-13}
line_hum=${3:-16}
place=$(echo "$place" | sed "s/ /%20/" )
curl --connect-timeout 2 --max-time 5 "$url" > $weatherreport

printf "%s" "$(sed "${line_hum}q;d" "$weatherreport" |
	grep -wo "[0-9]*%" | sort -rn | sed "s/^/☔/g;1q" | tr -d '\n')"
sed "${line_temp}q;d" "$weatherreport" | grep -o "m\\([-+]\\)*[0-9]\\+" | sed 's/+//g' | sort -n -t 'm' -k 2n | sed -e 1b -e '$!d' | tr '\n|m' ' ' | awk '{print " 🥶" $1 "°","🌞" $2 "°"}'
