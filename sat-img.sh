#!/bin/bash
curl --parallel --parallel-max 5 \
	https://api.sat24.com/mostrecent/DE/visual5hdcomplete --output clouds.png \
	https://api.sat24.com/mostrecent/DE/rainTMC --output rain.png \
	https://api.sat24.com/animated/DE/rainTMC/3/Central%20Europe%20Standard%20Time/5603172 --output ./web/pics/rain.gif \
	https://api.sat24.com/animated/DE/visual/3/Central%20Europe%20Standard%20Time/530005 --output ./web/pics/clouds.gif


cwebp -resize 400 300 rain.png -o ./web/pics/rain.webp &
cwebp -resize 400 300 clouds.png -o ./web/pics/clouds.webp &
wait
cp ./web/pics/rain.webp ./web/pics/clouds.webp ./web/pics/rain.gif ./web/pics/clouds.gif /var/www/startpage/
rm rain.png clouds.png
