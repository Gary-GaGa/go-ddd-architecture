package gameclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Client 是對後端遊戲 API 的極薄封裝。
type Client struct {
	base string
	hc   *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		base: baseURL,
		hc:   &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) GetViewModel(ctx context.Context) (ViewModel, error) {
	var vm ViewModel
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/api/v1/game/viewmodel", nil)
	resp, err := c.hc.Do(req)
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return vm, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return vm, err
	}
	return vm, nil
}

func (c *Client) PostClaimOffline(ctx context.Context, asOf string) (ClaimOfflineResponse, error) {
	var out ClaimOfflineResponse
	body, _ := json.Marshal(ClaimOfflineRequest{AsOf: asOf})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/game/claim-offline", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return out, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) PostStartPractice(ctx context.Context) (ViewModel, error) {
	var vm ViewModel
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/game/start-practice", nil)
	resp, err := c.hc.Do(req)
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return vm, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return vm, err
	}
	return vm, nil
}

func (c *Client) PostTryFinish(ctx context.Context) (FinishResponse, error) {
	var out FinishResponse
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/game/try-finish", nil)
	resp, err := c.hc.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return out, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) PostUpgradeKnowledge(ctx context.Context) (ViewModel, error) {
	var vm ViewModel
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/game/upgrade-knowledge", nil)
	resp, err := c.hc.Do(req)
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return vm, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return vm, err
	}
	return vm, nil
}

func (c *Client) PostSelectLanguage(ctx context.Context, lang string) (ViewModel, error) {
	var vm ViewModel
	body, _ := json.Marshal(map[string]string{"language": lang})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/api/v1/game/select-language", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		return vm, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return vm, decodeAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(&vm); err != nil {
		return vm, err
	}
	return vm, nil
}

func decodeAPIError(resp *http.Response) error {
	var env ErrorEnvelope
	_ = json.NewDecoder(resp.Body).Decode(&env)
	if env.Error.Code == "" {
		return errors.New(resp.Status)
	}
	return &APIErrorErr{Code: env.Error.Code, Message: env.Error.Message}
}

// APIErrorErr 是型別化的 API 錯誤，便於前端依錯誤碼做細緻處理。
type APIErrorErr struct {
	Code    string
	Message string
}

func (e *APIErrorErr) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	return e.Code + ": " + e.Message
}
