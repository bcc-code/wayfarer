package members

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 3,
}

func sendRequest[t any](ctx context.Context, client *Client, req *http.Request) (*result[t], error) {
	body, err := client.breaker.Execute(func() ([]byte, error) {
		token, err := client.tokenProvider.GetToken(ctx, client.domain)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve members token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		res, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = res.Body.Close()
		}()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if 200 > res.StatusCode || res.StatusCode > 299 {
			return nil, newStatusError(res.StatusCode, body)
		}

		return body, err
	})

	if err != nil {
		return nil, err
	}

	var data result[t]
	err = json.Unmarshal(body, &data)
	return &data, err
}

func get[t any](ctx context.Context, client *Client, endpoint string) (*t, error) {
	url := fmt.Sprintf("https://%s/%s", client.domain, endpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := sendRequest[t](ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return &res.Data, nil
}
