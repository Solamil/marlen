
t="tomorrow"
tomorrow=$(date -d "$t" "+%Y%m%d")
continent="europa"
url="https://en.sat24.com/image?type=forecastPrecip&region=${continent}&timestamp="$tomorrow
prefix="img"
img_type="png"

dir="web/pics"
dir_images="${dir}/sat24"
clock="0000"
pic="${dir_images}/${prefix}${tomorrow}${clock}.${img_type}"

curl "${url}${clock}"  --output "$pic"

cwebp -resize 300 400 "${pic}" -o "$dir/forecast_tmrw.webp"
