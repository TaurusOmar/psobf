package obfuscator

import (
	mathrand "math/rand"
	"regexp"
)

type Options struct {
	InputFile   string
	OutputFile  string
	Level       int
	NoExec      bool
	UseStdin    bool
	UseStdout   bool
	Seed        int64
	Seeded      bool
	VarRename   bool
	MinFrag     int
	MinFragSet  bool
	MaxFrag     int
	MaxFragSet  bool
	Quiet       bool
	Pipeline    string
	StringDict  int
	StrEnc      string
	StrKeyHex   string
	NumEnc      bool
	IdenMode    string
	FormatMode  string
	CFOpaque    bool
	CFShuffle   bool
	DeadProb    int
	FragProfile string
	Profile     string
	Fuzz        int
}

type Transform interface {
	Apply(ps string, ctx *Ctx) (string, error)
	Name() string
}

type Ctx struct {
	Rng       *mathrand.Rand
	Opts      *Options
	InputHash string
	Helpers   map[string]bool
}

var (
	reVar         = regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)
	reFuncHeader  = regexp.MustCompile(`(?i)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*\(`)
	reFuncNoParam = regexp.MustCompile(`(?i)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*{`)
	reParam       = regexp.MustCompile(`(?i)\$[A-Za-z_][A-Za-z0-9_]*`)
	reNum         = regexp.MustCompile(`\b\d+\b`)
	reDQ          = regexp.MustCompile("\"(?:[^\"`]|``|`\")*\"")
	reSQ          = regexp.MustCompile("'(?:[^']|'')*'")
)
