package DNSBinarySearch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	// for string reverse function
	"github.com/golang/example/stringutil"
)

// we expect every line to be less than 500 bytes (DNS only allows 255)
const MAXLINESIZE = 500

// scan backwards 10 kilobytes at a time looking for the edge of our matched string
const WALKBYTES = 10000

// we use 2 limits to limit runtime & output size of these library
type Limits struct {
	// the maximum distance to scan backwards (x 10kB)
	MaxScan int
	// the maximum number of lines of output
	MaxOutputLines int
}

var DefaultLimits = Limits{
	MaxScan:        100,    // 10MB
	MaxOutputLines: 100000, // 100,000 lines
}

// fetches a string buffer from a file
func getStringBuffer(f *os.File, offset int) (string, error) {
	_, err := f.Seek(int64(offset), 0)
	if err != nil {
		return "", err
	}
	returnBuf := make([]byte, MAXLINESIZE)
	_, err = io.ReadAtLeast(f, returnBuf, MAXLINESIZE)
	if err != nil {
		return "", err
	}

	return string(returnBuf), nil
}

// get the next line from a random string buffer
// (the first full line, newline char seperated)
func getNextLine(str string) string {
	// get the start of the next line
	lines := strings.Split(str, "\n")
	if len(lines) < 2 {
		// we expect the input file to be sufficiently large that we do not need to handle the EOF/start edge cases
		// we also expect that every line is less than 500 chars, that could also trigger this case
		return ""
	}

	// take out what we are going to compare
	// (the first line, after the next newline char, up to the length of the line we are trying to find)
	return lines[1]
}

// intermediary helper function to simplify code below
// takes a file, offset and string to search for
// returns a string compareLine which is fullLine truncated to len(searchStr)
// if err is set, result cannot be trusted
func getLineDetails(f *os.File, offset int, searchStr string) (compareLine string, err error) {
	// get the string buffer at this offset
	stringBuffer, err := getStringBuffer(f, offset)
	if err != nil {
		return "", err
	}

	// get the next line
	fullLine := getNextLine(stringBuffer)
	compareLine = fullLine
	if fullLine == "" {
		return "", fmt.Errorf("Failed to get next line from string buffer: %s\n", stringBuffer)
	}

	// filter out up to the length of the search string
	if len(compareLine) > len(searchStr) {
		compareLine = compareLine[0:len(searchStr)]
	}

	return
}

// pass a file path and search string to search for matches
// expects the file to sorted, with domain names at the start of the file, in reverse order
// example: "moc.elpmaxe.www,1.1.1.1"
// returns a list of matches
// example ["1.1.1.1,www.example.com"]
func DNSBinarySearch(filePath string, searchStr string, limit Limits) (ret []string, err error) {

	// reverse the search string
	searchStr = stringutil.Reverse(searchStr)

	// open the file & get it's size
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file")
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file")
	}

	// use sort.Search to find a line in our sorted file containing the search string
	// A possible enhancement here is to define our own sort.Search with an interface to
	// pass in the variables we need, rather than implicitly passing to the sub function here...
	foundByteLocation := sort.Search(int(fi.Size()), func(i int) bool {

		// use the intermediary function to get the line details at the offset we are currently considering
		searchLineCompare, err := getLineDetails(f, i, searchStr)
		if err != nil {
			// this should trigger an error in the next phase causing us to fail out quickly
			return false
		}

		// substring compare
		if strings.Compare(searchStr, searchLineCompare) > 0 {
			return false
		} else {
			return true
		}
	}) // end sort.Search

	// check if we found a match, if we did not, exit out
	stringBuffer, err := getStringBuffer(f, foundByteLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to get matched buffer from file?")
	}
	fullLine := getNextLine(stringBuffer)
	if fullLine == "" || !strings.HasPrefix(fullLine, searchStr) {
		return nil, fmt.Errorf("failed to find exact match via binary search")
	}

	// walk back 10 kilobytes bytes at a time, searching for a line that does not contain a match
	minSearchLocation := foundByteLocation
	maxScan := limit.MaxScan
	for {

		maxScan--
		if maxScan == 0 {
			return nil, fmt.Errorf("scan limit reached!")
		}

		// walk backwards in the file
		minSearchLocation = minSearchLocation - WALKBYTES
		if minSearchLocation < 0 {
			return nil, fmt.Errorf("scanned backwards too far! Reached start of file!")
		}

		// get the string buffer & next line at this offset
		searchLineCompare, err := getLineDetails(f, minSearchLocation, searchStr)
		if err != nil {
			return nil, fmt.Errorf("unexpected failure, failed to fetch next line while walking backwards?")
		}

		// we are looking for the first result that does not contain our substring
		if strings.Compare(searchStr, searchLineCompare) != 0 {
			break
		}
	}

	// seek to the minimum search location (this is likely unncessary to repeat as it was already done in getLineDetails above)
	_, err = f.Seek(int64(minSearchLocation), 0)
	if err != nil {
		return nil, err
	}

	// now that we have a min-location, use a bufio reader & ReadString('\n') to read the next line until they do not match!

	// call readString once to advance the pointer to the next \n
	reader := bufio.NewReader(f)
	_, err = reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	firstHit := false // bool flag is set to true once we start matching, once we stop matching and this flag is true, we can exit!
	maxOutputLines := limit.MaxOutputLines
	for {

		maxOutputLines--
		if maxOutputLines == 0 {
			// we likely could return what we have here already, but the result would be incomplete...
			return nil, fmt.Errorf("output limit reached!")
		}

		// this will read the next string up to the \n char
		nextLine, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		// remove the newline char
		nextLine = strings.TrimSuffix(nextLine, "\n")

		// filter out up to the length of the search string
		compareLine := nextLine
		if len(compareLine) > len(searchStr) {
			compareLine = compareLine[0:len(searchStr)]
		}

		// strings match!
		if strings.Compare(compareLine, searchStr) == 0 {
			// append the reversed line
			ret = append(ret, stringutil.Reverse(nextLine))

			// if this is our first hit, mark it as such!
			if firstHit == false {
				firstHit = true
			}
		} else if firstHit == true {
			// we've had a string match before, and they no longer match!
			// it's time to return
			break
		}

	}

	// and finally, close the file
	f.Close()

	return
}
