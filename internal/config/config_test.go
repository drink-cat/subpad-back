package config

import "testing"

func TestLoadQuoteToken(t *testing.T) {
	cfg, err := Load("../../etc/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.QuoteToken.LocalUsdc != "0xcB0f2a13098f8e841e6Adfa5B17Ec00508b27665" {
		t.Fatalf("localUsdc = %s", cfg.QuoteToken.LocalUsdc)
	}
	if cfg.QuoteToken.SepoliaUsdc != "0x5728d6521217108001057f09271311987e81d5a0" {
		t.Fatalf("sepoliaUsdc = %s", cfg.QuoteToken.SepoliaUsdc)
	}
	if len(cfg.SyncLog) != 2 || cfg.SyncLog[0].LaunchContract != "0x88D1aF96098a928eE278f162c1a84f339652f95b" {
		t.Fatalf("synclog = %+v", cfg.SyncLog)
	}
}
