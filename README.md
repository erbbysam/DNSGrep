# DNSGrep
A utility for quickly searching presorted DNS names. Built around the Rapid7 rdns & fdns dataset.

# How does it work?

This utility assumes the file provided is presorted (both alphabetical, and symbols).

The algorithm is pretty simple:
1) Use a binary search algorithm to seek through the file, looking for a substring match against the query.
2) Once a match is found, the file is scanned backwards in 10KB increments looking for a non-matching substring.
3) Once a non-matching substring is found, the file is scanned forwards until all exact matches are returned.

# Limits

There is a built-in limit system. This prevents 2 things:
1) scanning too far backwards (`MaxScan`)
2) scanning too far forwards after scanning backwards (`MaxOutputLines`)

This allows for any input while stopping requests that are taking too long.

Additionally, this utility does not handle the edge cases(start/end) of files and will return an error if encountered.

# Install

`go get` the following packages:

```
# used for dnsgrep cli flags
go get "github.com/jessevdk/go-flags"
# used by the experimental server for http routing
go get "github.com/gorilla/mux"
# pull in a string reversal function
go get "github.com/golang/example/stringutil"

```

# Run

The following steps were tested with Ubuntu 16.04 & go 1.11.5.

Generate fdns_a.sort.txt and rdns.sort.txt first using the scripts found in the scripts/ folder:
```
# Each of these scripts requires:
# * 3 hours+ on an SSD
# * 300GB+ temp disk space (under the same folder)
# * ~65GB  for output output (under the same folder)
# * jq to be installed
./scripts/fdns_a.sh
./scripts/rdns.sh
```


Run the command line utility:
```
go run dnsgrep.go -f DNSBinarySearch/test_data.txt -i "amiccom.com.tw"
```

Run the experimental server in the same folder as fdns_a.sort & rdns.sort.txt:
```
go run experimentalServer.go
```

# Docker 

You can also run the command line utility using Docker:
```
docker build -t dnsgrep .
docker run --rm -it -v "$PWD"/DNSBinarySearch:/files dnsgrep -f /files/test_data.txt -i ".amiccom.com.tw"
```

# Data Source
The source of this data referenced throughout this repository is Rapid7 Labs. Please review the Terms of Service:
https://opendata.rapid7.com/about/

https://opendata.rapid7.com/sonar.rdns_v2/

https://opendata.rapid7.com/sonar.fdns_v2/

# Stack Overflow References

via https://unix.stackexchange.com/a/35472
* we need to sort with LC_COLLATE=C to also sort ., chars

via https://unix.stackexchange.com/a/350068
 * To sort a large file: split it into chunks, sort the chunks and then simply merge the results



# License

See LICENSE file.
