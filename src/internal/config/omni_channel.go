package config

import (
	"context"
	"strings"
	"sync"

	"jk-api/pkg/httpclient"
)

type OmniChannel struct {
	client *httpclient.Client
	token  string
}

var (
	omniChannelClient *OmniChannel
	omniChannelOnce   sync.Once
)

func newOmniChannelClient() *OmniChannel {
	return &OmniChannel{
		client: httpclient.New(strings.TrimRight(AppConfig.OmniChannelURI, "/")),
	}
}

func InitOmniChannelClient() {
	omniChannelOnce.Do(func() {
		omniChannelClient = newOmniChannelClient()
	})
}

func OmniChannelClient() *OmniChannel {
	if omniChannelClient == nil {
		InitOmniChannelClient()
	}

	return omniChannelClient
}

func (c *OmniChannel) WithBearerToken(token string) *OmniChannel {
	c.token = strings.TrimSpace(token)
	return c
}

func (c *OmniChannel) Get(endpoint string) ([]byte, int, error) {
	return c.client.Get(context.Background(), endpoint, c.token)
}

func (c *OmniChannel) Post(endpoint string, body any) ([]byte, int, error) {
	return c.client.Post(context.Background(), endpoint, c.token, body)
}

func (c *OmniChannel) Put(endpoint string, body any) ([]byte, int, error) {
	return c.client.Put(context.Background(), endpoint, c.token, body)
}

func (c *OmniChannel) Patch(endpoint string, body any) ([]byte, int, error) {
	return c.client.Patch(context.Background(), endpoint, c.token, body)
}

func (c *OmniChannel) Delete(endpoint string, body any) ([]byte, int, error) {
	return c.client.Delete(context.Background(), endpoint, c.token, body)
}
