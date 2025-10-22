package routerproxy

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// FetchRouterStats scrapes or proxies the local I2P router console.
// cfgRouterURL should be something like "http://127.0.0.1:7657"
func FetchRouterStats(cfgRouterURL string) (map[string]any, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(cfgRouterURL + "/json")
	if err != nil {
		// If router doesn't expose /json, attempt to fetch root and basic parse (simplified)
		resp2, err2 := client.Get(cfgRouterURL + "/")
		if err2 != nil {
			return nil, err
		}
		defer resp2.Body.Close()
		b, _ := io.ReadAll(resp2.Body)
		// naive parse stub
		return map[string]any{"html": string(b)}, nil
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
