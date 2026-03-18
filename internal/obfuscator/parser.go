package obfuscator

import (
	"regexp"
)

type PSBlock struct {
	Type    string
	Content string
	Start   int
	End     int
	Line    int
}

type PSParser struct {
	Comments       *regexp.Regexp
	StringsDQ      *regexp.Regexp
	StringsSQ      *regexp.Regexp
	HereStrings    *regexp.Regexp
	Variables      *regexp.Regexp
	Functions      *regexp.Regexp
	FunctionsParam *regexp.Regexp
	Numbers        *regexp.Regexp
	Cmdlets        *regexp.Regexp
	Operators      *regexp.Regexp
}

var DefaultParser *PSParser

func init() {
	DefaultParser = NewPSParser()
}

func NewPSParser() *PSParser {
	return &PSParser{
		Comments:       regexp.MustCompile(`(?s)(?:#[^\n]*|<#.*?#>)`),
		StringsDQ:      regexp.MustCompile("\"(?:[^\"`]|``|`\")*\""),
		StringsSQ:      regexp.MustCompile(`'(?:[^']|'')*'`),
		HereStrings:    regexp.MustCompile(`(?s)@\s*"([^"]|"[^@])*"\s*@|@\s*'([^']|'[^@])*'\s*@`),
		Variables:      regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_:]*\}|\$[A-Za-z_][A-Za-z0-9_]*`),
		Functions:      regexp.MustCompile(`(?im)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*\{`),
		FunctionsParam: regexp.MustCompile(`(?im)\bfunction\s+([A-Za-z_][A-Za-z0-9_-]*)\s*\(([^)]*)\)\s*\{`),
		Numbers:        regexp.MustCompile(`\b(?:0x[0-9A-Fa-f]+|\d+(?:\.\d+)?)\b`),
		Cmdlets:        regexp.MustCompile(`\b[A-Z][a-z]+-[A-Z][A-Za-z]+\b`),
		Operators:      regexp.MustCompile(`\b(?:-eq|-ne|-gt|-lt|-ge|-le|-like|-notlike|-match|-notmatch|-contains|-notcontains|-in|-notin|-replace|-split|-join|-and|-or|-xor|-band|-bor|-bxor|-shl|-shr)\b`),
	}
}

func (p *PSParser) FindAllStrings(src string) []PSBlock {
	var blocks []PSBlock
	hereMatches := p.HereStrings.FindAllStringIndex(src, -1)
	for _, m := range hereMatches {
		blocks = append(blocks, PSBlock{
			Type:    "string",
			Content: src[m[0]:m[1]],
			Start:   m[0],
			End:     m[1],
		})
	}
	processed := make([]bool, len(src))
	for _, b := range blocks {
		for i := b.Start; i < b.End && i < len(src); i++ {
			processed[i] = true
		}
	}
	dqMatches := p.StringsDQ.FindAllStringIndex(src, -1)
	for _, m := range dqMatches {
		if !processed[m[0]] {
			blocks = append(blocks, PSBlock{
				Type:    "string",
				Content: src[m[0]:m[1]],
				Start:   m[0],
				End:     m[1],
			})
		}
	}
	sqMatches := p.StringsSQ.FindAllStringIndex(src, -1)
	for _, m := range sqMatches {
		if !processed[m[0]] {
			blocks = append(blocks, PSBlock{
				Type:    "string",
				Content: src[m[0]:m[1]],
				Start:   m[0],
				End:     m[1],
			})
		}
	}
	return blocks
}

func (p *PSParser) FindAllFunctions(src string) []PSBlock {
	var blocks []PSBlock
	matches := p.Functions.FindAllStringSubmatchIndex(src, -1)
	for _, m := range matches {
		blocks = append(blocks, PSBlock{
			Type:    "function",
			Content: src[m[0]:m[1]],
			Start:   m[0],
			End:     m[1],
		})
	}
	return blocks
}

func (p *PSParser) StripComments(src string) string {
	return p.Comments.ReplaceAllString(src, "")
}

func (p *PSParser) IsInsideString(src string, pos int) bool {
	strings := p.FindAllStrings(src)
	for _, s := range strings {
		if pos >= s.Start && pos < s.End {
			return true
		}
	}
	return false
}
