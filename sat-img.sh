#!/bin/bash
repo_path="$HOME/repo"
repo_pics_path=$repo_path"/web/pics"
curl --parallel --parallel-max 5 \
	https://api.sat24.com/mostrecent/DE/visual5hdcomplete --output "$repo_pics_path"/clouds.png \
	https://api.sat24.com/mostrecent/DE/rainTMC --output "$repo_pics_path"/rain.png \
	https://api.sat24.com/animated/DE/rainTMC/3/Central%20Europe%20Standard%20Time/5603172 --output "$repo_pics_path"/rain.gif \
	https://api.sat24.com/animated/DE/visual/3/Central%20Europe%20Standard%20Time/530005 --output  "$repo_pics_path"/clouds.gif


cwebp -resize 400 300 "$repo_pics_path"/rain.png -o "$repo_pics_path"/rain.webp &
cwebp -resize 400 300 "$repo_pics_path"/clouds.png -o "$repo_pics_path"/clouds.webp &
wait
rm "$repo_pics_path"/{rain.png,clouds.png} # /home/merlot/repo/web/pics/clouds.png
