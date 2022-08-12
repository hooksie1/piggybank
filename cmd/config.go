package cmd

import (
	_ "embed"

	"cuelang.org/go/cue"
	"github.com/nats-io/nats.go"
)

var (
	//go:embed schema.cue
	schema string
)

type Config struct {
	Config AppConfig
}

type AppConfig struct {
	URLs    string `json:"urls,omitempty"`
	Creds   string `json:"creds,omitempty"`
	User    string `json:"user,omitempty"`
	Pass    string `json:"pass,omitempty"`
	NKey    string `json:"nkey,omitempty"`
	TLSCert string `json:"tls_cert,omitempty"`
	TLSKey  string `json:"tls_key,omitempty"`
	TLSCA   string `json:"tls_ca,omitempty"`
}

func (c *Config) BuildConfig(v cue.Value) error {
	// must use the same context as the supplied config
	s := v.Context().CompileString(schema)

	u := s.Unify(v)

	if err := u.Decode(c); err != nil {
		return err
	}

	return nil
}

func (n *AppConfig) getOptions() ([]nats.Option, error) {
	var opts []nats.Option

	if n.Creds != "" {
		opts = append(opts, nats.UserCredentials(n.Creds))
	}

	if n.User != "" && n.Pass != "" {
		opts = append(opts, nats.UserInfo(n.User, n.Pass))
	}

	if n.TLSCert != "" && n.TLSKey != "" {
		opts = append(opts, nats.ClientCert(n.TLSCert, n.TLSKey))
	}

	if n.TLSCA != "" {
		opts = append(opts, nats.RootCAs(n.TLSCA))
	}

	if n.NKey != "" {
		opt, err := nats.NkeyOptionFromSeed(n.NKey)
		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	return opts, nil

}
