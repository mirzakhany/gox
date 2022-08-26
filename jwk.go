package gox

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

type PublicKeyProvider struct {
	url    string
	cache  *jwk.Cache
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPublicKeyProvider(url string, refreshInterval time.Duration) (*PublicKeyProvider, error) {
	p := &PublicKeyProvider{}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.ctx = ctx

	cache := jwk.NewCache(ctx)
	p.cache = cache

	err := cache.Register(url, jwk.WithMinRefreshInterval(refreshInterval))
	if err != nil {
		defer cancel()
		return nil, err
	}

	if _, err := cache.Refresh(ctx, url); err != nil {
		defer cancel()
		return nil, err
	}

	return p, nil
}

func (c *PublicKeyProvider) GetKeySet(ctx context.Context) (jwk.Set, error) {
	return c.cache.Get(ctx, c.url)
}

func (c *PublicKeyProvider) Close() error {
	c.cancel()
	return nil
}
