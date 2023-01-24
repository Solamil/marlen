# today=$(date -d +"%Y%m%d")
# hour=$(date -d +"%H")

t="tomorrow"
tomorrow=$(date -d "$t" "+%Y%m%d")
formatted=$(date -d "$t" "+%d.%m. %A")
forecastType=${1:-"forecastPrecip"}
continent=${2:-"europa"}
url="https://en.sat24.com/image?type=${forecastType}&region=${continent}&timestamp="$tomorrow
prefix="img"
img_type="png"
dir="$HOME/repo/web/pics"
[ -d "$dir" ] || dir="$HOME/.local/src/marlen/web/pics"
#dir="web/pics"
dir_images="${dir}/sat24"
out="${dir}/forecast_tmrw.gif"
[ -d "$dir_images" ] || mkdir "$dir_images"

for i in "00" "03" "06" "09" "12" "15" "18" "21"; do
        clock="${i}00"
        pic="${dir_images}/${prefix}${tomorrow}${clock}.${img_type}"
        if [ ! -f "$pic" ]; then
                curl "${url}${clock}"  --output "$pic"
                convert "$pic" -gravity South -pointsize 18 -undercolor white -fill black  -annotate +0+0 "${formatted} ${i}:00" "$pic"
        fi
done

convert -delay 160 "${dir_images}/${prefix}${tomorrow}*${img_type}" "$out"

cwebp -resize 300 400 "${dir_images}/${prefix}${tomorrow}0000.${img_type}" -o "$dir/forecast_tmrw.webp"
rm -v "${dir_images}"/*png

# convert img0000.png -gravity South -pointsize 18 -fill yellow -annotate +0+0 "16.12.2022 00:00" img0001.png
