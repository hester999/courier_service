package cli

import (
	"errors"
	"flag"
	"os"

	"avito/pkg"
)

type FlagsCLI struct {
	port string
	log  pkg.Logger
}

func NewFlagsCLI(logger pkg.Logger) *FlagsCLI {
	return &FlagsCLI{
		log: logger,
	}
}

var ErrSetEnv = errors.New("set environment error")

func (f *FlagsCLI) SetFlags() error {
	flag.StringVar(&f.port, "port", "", "port to listen on")
	flag.Parse()
	if f.port == "" {
		f.log.Info("port not set, read from environment variable PORT")
		return nil
	}
	err := os.Setenv("PORT", f.port)
	if err != nil {
		f.log.Error("Error setting env var: %v", err)
		return ErrSetEnv
	}
	return nil
}
