package psobf

import (
	"github.com/taurusomar/psobf/internal/obfuscator"
)

type Config = obfuscator.Options

func Obfuscate(source string, cfg Config) (string, error) {
	return obfuscator.ObfuscateString(source, cfg)
}
