package client

import (
	"testing"
)

func TestConfigDownload(t *testing.T) {
	var err error

	if _, err = pullConfigFile(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigParse(t *testing.T) {
	var cfg *Config
	var err error

	if cfg, err = pullConfig(); err != nil {
		t.Fatal(err)
	}

	if cfg == nil {
		t.Fatal("Expecting non-nil config file.")
	}
}
