package marathon

import (
	"net/url"

	"github.com/gambol99/go-marathon"
)

// AppFetcher fetches apps from Marathon
type AppFetcher struct {
	m marathon.Marathon
}

// NewAppFetcher makes a new AppFetcher with the specified Marathon location
func NewAppFetcher(u string) (*AppFetcher, error) {
	mc, err := marathon.NewClient(marathon.Config{
		URL:            u,
		RequestTimeout: 5,
	})
	if err != nil {
		return nil, err
	}

	return &AppFetcher{
		m: mc,
	}, nil
}

// Apps fetches all apps from marathon.
func (a *AppFetcher) Apps() ([]marathon.Application, error) {
	mv := url.Values{}
	mv.Set("embed", "apps.tasks")
	ma, err := a.m.Applications(mv)
	if err != nil {
		return nil, err
	}

	return ma.Apps, nil
}

// FetchApps fetches apps with specific label set to specific value.
func (a *AppFetcher) FetchApps(labelKey string, labelValue string) ([]marathon.Application, error) {
	mv := url.Values{}
	mv.Set("embed", "apps.tasks")
	mv.Set("label", labelKey+"=="+labelValue)

	ma, err := a.m.Applications(mv)
	if err != nil {
		return nil, err
	}

	return ma.Apps, nil
}
