package DNSBinarySearch

import "testing"

// a quick sanity check to make sure this library works as expected
func TestResult(t *testing.T) {

	output, err := DNSBinarySearch("test_data.txt", "amiccom.com.tw", DefaultLimits)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	} else {
		if len(output) != 6 {
			t.Fatalf("unexpected output length: %+v", output)
		}
	}
}
