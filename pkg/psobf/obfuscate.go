package psobf

import (
	"github.com/TaurusOmar/psobf/v2/internal/obfuscator"
)

type Config = obfuscator.Options

func Obfuscate(source string, cfg Config) (string, error) {
	return obfuscator.ObfuscateString(source, cfg)
}
