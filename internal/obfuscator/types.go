package obfuscator

import (
	mathrand "math/rand"
	"regexp"
)

type Options struct {
	InputFile    string
	OutputFile   string
	Level        int
	NoExec       bool
	UseStdin     bool
	UseStdout    bool
	Seed         int64
	Seeded       bool
	VarRename    bool
	MinFrag      int
	MinFragSet   bool
	MaxFrag      int
	MaxFragSet   bool
	Quiet        bool
	Pipeline     string
	StringDict   int
	StrEnc       string
	StrKeyHex    string
	NumEnc       bool
	IdenMode     string
	FormatMode   string
	CFOpaque     bool
	CFShuffle    bool
	DeadProb     int
	FragProfile  string
	Profile      string
	Fuzz         int
	Polymorphism int
}

type RNGProvider interface {
	Intn(n int) int
	Int63() int64
	Float64() float64
}

type TransformContext interface {
	RNG() RNGProvider
	Options() *Options
	InputHash() string
	AddHelper(name string)
	HasHelper(name string) bool
}

type Transform interface {
	Apply(ps string, ctx TransformContext) (string, error)
	Name() string
}

type Ctx struct {
	Rng       *mathrand.Rand
	Opts      *Options
	inputHash string
	Helpers   map[string]bool
}

func (c *Ctx) RNG() RNGProvider           { return c.Rng }
func (c *Ctx) Options() *Options          { return c.Opts }
func (c *Ctx) InputHash() string          { return c.inputHash }
func (c *Ctx) AddHelper(name string)      { c.Helpers[name] = true }
func (c *Ctx) HasHelper(name string) bool { return c.Helpers[name] }

// Deprecated: Use DefaultParser.Variables instead. Kept for backward compatibility.
var (
	reVar = regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)
	// Deprecated: Use DefaultParser.FunctionsParam instead.
	reFuncHeader = regexp.MustCompile(`(?i)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*\(`)
	// Deprecated: Use DefaultParser.Functions instead.
	reFuncNoParam = regexp.MustCompile(`(?i)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*{`)
	// Deprecated: Use DefaultParser.Variables instead.
	reParam = regexp.MustCompile(`(?i)\$[A-Za-z_][A-Za-z0-9_]*`)
	// Deprecated: Use DefaultParser.Numbers instead.
	reNum = regexp.MustCompile(`\b\d+\b`)
	// Deprecated: Use DefaultParser.StringsDQ instead.
	reDQ = regexp.MustCompile("\"(?:[^\"`]|``|'\")*\"")
	// Deprecated: Use DefaultParser.StringsSQ instead.
	reSQ = regexp.MustCompile("'(?:[^']|'')*'")
)
