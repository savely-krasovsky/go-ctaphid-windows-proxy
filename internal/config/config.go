package config

import "github.com/savely-krasovsky/go-ctaphid-windows-proxy/internal/infra/transport"

type Config struct {
	Transport *transport.Config
}
