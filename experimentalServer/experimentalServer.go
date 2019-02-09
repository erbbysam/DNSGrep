package main

import (
	. "dnsgrep/DNSBinarySearch"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// a struct for the metadata contained in the JSON
type MetaJSON struct {
	Runtime   string // not the most efficent way to convey this...
	Errors    []string
	FileNames []string // list of filenames scanned
	TOS       string
}

// a struct for the response json
type ResponseJSON struct {
	Meta   MetaJSON
	FDNS_A []string
	RDNS   []string
}

// fetch the DNS info from our files
func fetchDNSInfo(queryString string) (fdns_a []string, rdns []string, errors []string) {

	// fetch from our files
	fdns_a, err := DNSBinarySearch("fdns_a.sort.txt", queryString, DefaultLimits)
	if err != nil {
		errors = append(errors, fmt.Sprintf("fdns_a error: %+v", err))
	}
	rdns, err = DNSBinarySearch("rdns.sort.txt", queryString, DefaultLimits)
	if err != nil {
		errors = append(errors, fmt.Sprintf("rdns error: %+v", err))
	}

	return
}

// primary DNS handler
func DNSHandler(w http.ResponseWriter, r *http.Request) {

	vals := r.URL.Query()
	queryString, ok := vals["q"]
	if ok {

		// write out a JSON content-type
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// query the two large files
		before := time.Now()
		fdns_a, rdns, errors := fetchDNSInfo(queryString[0])

		// get runtime
		delta := time.Now().Sub(before)
		runtimeStr := fmt.Sprintf("%f seconds", delta.Seconds())

		// now put together our JSON!
		ret := ResponseJSON{
			FDNS_A: fdns_a,
			RDNS:   rdns,
		}
		ret.Meta.Runtime = runtimeStr
		ret.Meta.Errors = errors
		// TODO -- these really should come in via a config file
		ret.Meta.FileNames = []string{"2019-01-25-1548417890-fdns_a.json.gz", "2019-01-30-1548868121-rdns.json.gz"}
		ret.Meta.TOS = "The source of this data is Rapid7 Labs. Please review the Terms of Service: https://opendata.rapid7.com/about/"

		// finally, encode the json!
		jsonEncoded, err := json.MarshalIndent(ret, "", "\t")
		if err != nil {
			w.Write([]byte("Unexpected failure to encode json?\n"))
		} else {
			// success!
			w.Write(jsonEncoded)
		}

	} else {
		w.Write([]byte("Missing query string!\n"))
	}
}

// simple mux server startup
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/dns", DNSHandler)
	log.Fatal(http.ListenAndServe(":80", r))
}
