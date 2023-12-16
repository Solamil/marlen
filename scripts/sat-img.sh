#!/bin/bash
repo_path="$HOME/.local/src/marlen"
repo_pics_path="${repo_path}/web/static/pics"

the_server="https://api.sat24.com"
pic_link="${the_server}/mostrecent/DE"
gif_link="${the_server}/animated/DE"
time="Central%20Europe%20Standard%20Time"

curl --connect-timeout 5 --parallel --parallel-max 5 \
	${pic_link}/visual5hdcomplete --output "$repo_pics_path"/clouds.png \
	${pic_link}/rainTMC --output "$repo_pics_path"/rain.png \
	${gif_link}/rainTMC/3/${time}/5603172 --output "$repo_pics_path"/rain.gif \
	${gif_link}/visual/3/${time}/530005 --output  "$repo_pics_path"/clouds.gif \
	${the_server}/mostrecent/EU/visual5hdcomplete --output "$repo_pics_path"/clouds_eu.png \
	${the_server}/animated/EU/visual/2/${time}/4375097 --output "$repo_pics_path"/clouds_eu.gif



cwebp -resize 400 300 "$repo_pics_path"/rain.png -o "$repo_pics_path"/rain.webp &
cwebp -resize 400 300 "$repo_pics_path"/clouds.png -o "$repo_pics_path"/clouds.webp &
cwebp -resize 789 480 "$repo_pics_path"/clouds_eu.png -o "$repo_pics_path"/clouds_eu.webp &
wait
rm -v "$repo_pics_path"/{rain.png,clouds.png,clouds_eu.png} # /home/merlot/repo/web/pics/clouds.png
