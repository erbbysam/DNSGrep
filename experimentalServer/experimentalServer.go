package main

import (
	. "dnsgrep/DNSBinarySearch"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"io/ioutil"
	"github.com/gorilla/mux"
)
const (
	configJSON = "/home/ubuntu/go/src/dnsgrep/experimentalServer/config.json"
)
// a struct for the metadata contained in the JSON
type MetaJSON struct {
	Runtime   string // not the most efficent way to convey this...
	Errors    []string
	Message   string `json:"Message"` // custom message to send
	FileNames []string `json:"FileNames"`// list of filenames scanned
	TOS       string `json:"TOS"`
}

// a struct for the response json
type ResponseJSON struct {
	Meta   MetaJSON
	FDNS_A []string
	RDNS   []string
}
// load config
func GetMeta(path string) (MetaConfig *MetaJSON) {
	MetaConfig = new(MetaJSON)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
 	err = json.Unmarshal(data, &MetaConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling config file: %v",err)
	}
	return MetaConfig
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
// homepage handler
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
// primary DNS handler
func DNSHandler(w http.ResponseWriter, r *http.Request) {
	MetaCfg := GetMeta(configJSON)
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
		ret.Meta.Message = MetaCfg.Message
                ret.Meta.FileNames = MetaCfg.FileNames
                ret.Meta.TOS = MetaCfg.TOS

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
	r.HandleFunc("/",IndexHandler)
	r.HandleFunc("/dns", DNSHandler)
	log.Fatal(http.ListenAndServe(":80", r))
}

