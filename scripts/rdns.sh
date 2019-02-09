# fetch the rdns file
wget -O rdns.gz https://opendata.rapid7.com/sonar.rdns_v2/2019-01-30-1548868121-rdns.json.gz

# extract and format our data
gunzip -c rdns.gz | jq -r '.name + ","+ .value' | tr '[:upper:]' '[:lower:]' | rev > rdns.rev.lowercase.txt

# split the data into chunks ot sort
split -b100M rdns.rev.lowercase.txt fileChunk

# remove the old files
rm rdns.gz
rm rdns.rev.lowercase.txt

## Sort each of the pieces and delete the unsorted one
for f in fileChunk*; do LC_COLLATE=C sort "$f" > "$f".sorted && rm "$f"; done

## merge the sorted files with local tmp directory
mkdir -p sorttmp
LC_COLLATE=C sort -T sorttmp/ -muo rdns.sort.txt fileChunk*.sorted

# clean up
rm fileChunk*
