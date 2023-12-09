#!/bin/sh
title="Kurz devizového trhu"
url="https://cnb.cz/cs/financni-trhy/devizovy-trh/kurzy-devizoveho-trhu/kurzy-devizoveho-trhu/denni_kurz.txt"
path="$HOME/.local/src/marlen"
dir="${path}/rates"
file="${path}/denni_kurz.txt"
list_file="$dir/list.txt"
date_file="$dir/date.txt"
number_file="$dir/number.txt"
trinity_file="$dir/svata_trojice.txt"
web_dir="${path}/web"
index="${web_dir}/cnb-rates.html"

[ -d "$web_dir" ] || mkdir -pv "$web_dir"
[ -d "$dir" ] || mkdir -pv "$dir"

download_rates() { curl -sLf "$url" -o "$file"; }

parse_rates() {

	grep "|" "$file" | cut -d"|" -f4 > "$list_file"
	head -n 1 "$file" | cut -d" " -f1 > "$date_file"
	head -n 1 "$file" | cut -d"#" -f2 > "$number_file"

	grep -v "kód" < "$list_file" | while IFS= read -r code 
	do
		line=$(grep "$code" "$file")
		value=$(echo "$line" | grep -o "\|[^\|]*$" | tr "," ".")
		printf "%.2f\n" "$value" > "$dir/$code.txt"
		echo "$line" | cut -d"|" -f3 >> "$dir/$code.txt"
	done

#	option_tags=""
#	links_code=""
#	codes=$(grep -v "^kód" "$list_file")
#	for i in $codes; do
#		value=$(cat "${dir}/${i}.txt")
#		option_tags=$option_tags" <option value=\"$i\">$i</option>"
#		links_code=$links_code" <a href=\"/?code=$i\"><abbr title=\"$value\">$i</abbr></a>"
#	done
	printf "1zł %.2fKč 1$ %.2fKč 1€ %.2fKč 1£ %.2fKč" "$(head -n 1 "$dir/PLN.txt")" "$(head -n 1 "$dir/USD.txt")" "$(head -n 1 "$dir/EUR.txt")" "$(head -n 1 "$dir/GBP.txt")"  > "$trinity_file"
}

render_html() { 
	echo "<!DOCTYPE html>
	<html>
	<head>
	<title>$title</title>
	<link rel=\"shortcut icon\" href=\"./pics/favicon.svg\" type=\"image/svg+xml\" />
	<meta charset=\"utf-8\"/>
	<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">
	</head>
	<body>
	
		<p>$(cat "$trinity_file")</p>
		<pre>
$(cat "$file")		
		</pre>
	</body>
	</html>
	" > "$index"

}

#			<p>$links_code</p>
#			<a href=\"/json\">JSON</a>
#			<a href=\"/list\">list</a>
#			<a href=\"/date\"><abbr title=\"$(cat "$date_file" 2>/dev/null)\">date</abbr></a>
#			<a href=\"/number\"><abbr title=\"$(cat "$number_file" 2>/dev/null)\">number</abbr></a>
#			<a href=\"/denni_kurz.txt\">denni_kurz.txt</a>
#			<a href=\"/svata_trojice\">Svatá trojice</a>
#			<a href=\"/holy_trinity\">Holy Trinity</a>
#			<a href=\"/holy_trinity?p\">Pretty</a>

if [ -f "$file" ]; then
	current_date=$(date +"%d.%m.%Y")
	if [ "$(cat "$date_file")" != "$current_date" ]; then
			download_rates
	else
		echo "Up to date."
	fi
else
	download_rates
fi
parse_rates
render_html
