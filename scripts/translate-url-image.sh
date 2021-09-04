# shellcheck disable=SC2162
echo Enter the image URL:
read inputURL
./manga-translator -url=true "$inputURL"
