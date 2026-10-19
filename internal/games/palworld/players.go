package palworld

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var _maxResponseSize = 1 << 20

// PlayerCount returns the number of players reported by the local REST API.
func (a *Palworld) PlayerCount(ctx context.Context) (int, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/v1/api/players", a.restPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create Palworld player request: %w", err)
	}
	req.SetBasicAuth("admin", a.adminPassword)
	req.Header.Set("Accept", "application/json")
	resp, err := a.restClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("query Palworld players: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Palworld player request returned HTTP %d", resp.StatusCode)
	}

	body := &io.LimitedReader{R: resp.Body, N: int64(_maxResponseSize) + 1}
	decoder := json.NewDecoder(body)
	var result struct {
		Players []struct{} `json:"players"`
	}
	err = decoder.Decode(&result)
	if body.N == 0 {
		return 0, fmt.Errorf("Palworld player response exceeds 1 MiB")
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return 0, fmt.Errorf("read Palworld player response: %w", err)
	}
	if err != nil || result.Players == nil {
		return 0, fmt.Errorf("Palworld player response must contain a valid players array")
	}
	return len(result.Players), nil
}
