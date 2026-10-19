package palworld

import (
	"context"
	"fmt"
	"net/http"
)

// Save persists the world through the local REST API.
func (a *Palworld) Save(ctx context.Context) error {
	url := fmt.Sprintf("http://127.0.0.1:%d/v1/api/save", a.restPort)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create Palworld save request: %w", err)
	}
	req.SetBasicAuth("admin", a.adminPassword)
	resp, err := a.restClient.Do(req)
	if err != nil {
		return fmt.Errorf("save Palworld world: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Palworld save request returned HTTP %d", resp.StatusCode)
	}
	return nil
}
