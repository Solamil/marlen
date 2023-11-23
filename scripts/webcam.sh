dir="$HOME/.local/src/marlen/web/pics"
curl  --connect-timeout 5 --parallel --parallel-max 5 \
	"https://kamery.mh2net.cz/img/snap-klub.jpeg" --output "${dir}/snap-klub.jpeg" \
	"https://www.mnhradiste.cz/kamera/mhcam1.jpg" --output "${dir}/mhcam1.jpg"


cwebp -resize 340 220 "${dir}/mhcam1.jpg" -o "${dir}/mhcam1.webp"
cwebp -resize 340 220  "${dir}/snap-klub.jpeg" -o "${dir}/snap-klub.webp"

rm -v "${dir}/snap-klub.jpeg" "${dir}/mhcam1.jpg"
