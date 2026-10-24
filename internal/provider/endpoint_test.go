package provider

import "testing"

func TestCheckEndpoint(t *testing.T) {
	for endpoint, ok := range map[string]bool{
		"https://app.melt.cloud": true,
		"http://localhost:3000":  true,
		"http://127.0.0.1:3000":  true,
		"http://app.melt.cloud":  false,
		"ftp://app.melt.cloud":   false,
		"app.melt.cloud":         false,
	} {
		if err := checkEndpoint(endpoint); (err == nil) != ok {
			t.Errorf("checkEndpoint(%q) = %v; want ok=%v", endpoint, err, ok)
		}
	}
}
