package wunderground

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	cfg "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	log "github.com/sirupsen/logrus"
)

const (
	BaseURL   = "https://weatherstation.wunderground.com/weatherstation/updateweatherstation.php"
	ApiFormat = "%s?ID=%s&PASSWORD=%s&dateutc=now&action=updateraw&softwaretype=wunderground-uploader"
)

func Start(ctx context.Context, wg *sync.WaitGroup, cfg cfg.Config, dataCh <-chan map[string]string) <-chan struct{} {
	out := make(chan struct{})

	wg.Add(1)

	go func() {
		defer close(out)
		defer wg.Done()

		for {
			select {
			case data, ok := <-dataCh:
				if !ok {
					return
				}

				if err := uploadWeather(data, cfg); err != nil {
					log.Errorf("Upload failed %v", err)
					continue
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func uploadWeather(data map[string]string, cfg cfg.Config) error {
	req, err := buildRequestString(data, cfg.StationID, cfg.StationKey)
	if err != nil {
		return err
	}

	log.Debug(req)

	res, err := http.Get(req)
	if err != nil {
		return err
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	return nil
}

func buildRequestString(data map[string]string, id string, password string) (string, error) {
	req := fmt.Sprintf(ApiFormat, BaseURL, id, password)

	if len(data) == 0 {
		return "", errors.New("no data to build request with")
	}

	for k, v := range data {
		req += fmt.Sprintf("&%s=%s", k, v)
	}

	return req, nil
}
