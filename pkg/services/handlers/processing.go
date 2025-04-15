package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/ASRafalsky/telemetry/internal/types"
)

func counterPostDataHandler(repo repository, key string, value string) error {
	newValue, err := types.ParseCounter(value)
	if err != nil {
		return err
	}
	if previousValue, ok := repo.Get(strings.ToLower(key)); ok {
		newValue += types.BytesToCounter(previousValue)
	}
	repo.Set(strings.ToLower(key), types.CounterToBytes(newValue))
	return nil
}

func gaugeGetDataHandler(repo repository, key string) (string, error) {
	if value, ok := repo.Get(strings.ToLower(key)); ok {
		return types.BytesToGauge(value).String(), nil
	}
	return "", errors.New("gauge value not found")
}

func counterGetDataHandler(repo repository, key string) (string, error) {
	if value, ok := repo.Get(strings.ToLower(key)); ok {
		return types.BytesToCounter(value).String(), nil
	}
	return "", errors.New("counter value not found")
}

func gaugePostDataHandler(repo repository, key string, value string) error {
	newValue, err := types.ParseGauge(value)
	if err != nil {
		return err
	}
	repo.Set(strings.ToLower(key), types.GaugeToBytes(newValue))
	return nil
}

func getKeyList(repos ...repository) []string {
	totalEntryCnt := 0
	for _, repo := range repos {
		totalEntryCnt += repo.Size()
	}

	result := make([]string, totalEntryCnt)
	for _, repo := range repos {
		_ = repo.ForEach(context.Background(), func(k string, _ []byte) error {
			result = append(result, k)
			return nil
		})
	}
	return result
}
