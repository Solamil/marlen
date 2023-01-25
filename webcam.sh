dir="$HOME/repo"
[ -d "$dir" ] || dir="$HOME/.local/src/marlen"
cert_file="${dir}/mnhradiste-cz.pem"
NAME="mhcam1.jpg"
URL="https://www.mnhradiste.cz/kamera/${NAME}"
url=${1:-$URL}
name=${2-$NAME}
output="${dir}/web/pics/${name}"
curl --cacert "$cert_file" --connect-timeout 5 "$url" --output "$output"

cwebp -resize 340 220 "$output" -o "${dir}/web/pics/${name%.*}.webp"
rm -v "${output}"
