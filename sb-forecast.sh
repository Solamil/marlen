weatherreport="./weatherreport"
place="Zdar"
line_temp=${1:-13}
line_hum=${2:-16}

[ -s "$weatherreport" ] && [ "$(stat -c %y "$weatherreport" 2>/dev/null | cut -d' ' -f1)" = "$(date '+%Y-%m-%d')" ] || \
	curl wttr.in/$place > $weatherreport
printf "%s" "$(sed "${line_hum}q;d" "$weatherreport" |
	grep -wo "[0-9]*%" | sort -rn | sed "s/^/☔/g;1q" | tr -d '\n')"
sed "${line_temp}q;d" "$weatherreport" | grep -o "m\\([-+]\\)*[0-9]\\+" | sed 's/+//g' | sort -n -t 'm' -k 2n | sed -e 1b -e '$!d' | tr '\n|m' ' ' | awk '{print " 🥶" $1 "°","🌞" $2 "°"}'
