package obfuscator

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

func ParseFlags() Options {
	opts := Options{}
	flag.StringVar(&opts.InputFile, "i", "", "PowerShell script input file (use -stdin).")
	flag.StringVar(&opts.OutputFile, "o", "obfuscated.ps1", "Output file (use -stdout).")
	flag.IntVar(&opts.Level, "level", 1, "Obfuscation level (1..5).")
	flag.BoolVar(&opts.NoExec, "noexec", false, "Emit only payload without Invoke-Expression.")
	flag.BoolVar(&opts.UseStdin, "stdin", false, "Read script from STDIN.")
	flag.BoolVar(&opts.UseStdout, "stdout", false, "Write result to STDOUT.")
	flag.BoolVar(&opts.VarRename, "varrename", false, "Deprecated: kept for backward-compatibility. Use -iden obf.")
	flag.IntVar(&opts.MinFrag, "minfrag", 10, "Minimum fragment size (level 5).")
	flag.IntVar(&opts.MaxFrag, "maxfrag", 20, "Maximum fragment size (level 5).")
	flag.BoolVar(&opts.Quiet, "q", false, "Quiet mode (no banner).")
	flag.StringVar(&opts.Pipeline, "pipeline", "", "Comma-separated transforms: iden,strenc,stringdict,numenc,fmt,cf,dead")
	flag.IntVar(&opts.StringDict, "stringdict", 0, "String tokenization percentage (0..100).")
	flag.StringVar(&opts.StrEnc, "strenc", "off", "String encryption: off|xor|rc4.")
	flag.StringVar(&opts.StrKeyHex, "strkey", "", "Hex key for -strenc.")
	flag.BoolVar(&opts.NumEnc, "numenc", false, "Enable number encoding.")
	flag.StringVar(&opts.IdenMode, "iden", "keep", "Identifier morphing: obf|keep.")
	flag.StringVar(&opts.FormatMode, "fmt", "off", "Format jitter: off|jitter.")
	flag.BoolVar(&opts.CFOpaque, "cf-opaque", false, "Enable opaque predicate wrapper.")
	flag.BoolVar(&opts.CFShuffle, "cf-shuffle", false, "Shuffle function blocks.")
	flag.IntVar(&opts.DeadProb, "deadcode", 0, "Dead-code injection probability (0..100).")
	flag.StringVar(&opts.FragProfile, "frag", "", "Fragmentation profile: profile=tight|medium|loose.")
	flag.StringVar(&opts.Profile, "profile", "", "Preset: light|balanced|heavy.")
	flag.IntVar(&opts.Fuzz, "fuzz", 0, "Generate N fuzzed variants (unique seeds).")
	var seed int64
	flag.Int64Var(&seed, "seed", 0, "RNG seed (reproducible). Overrides crypto/rand if set.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  psobf -i input.ps1 -o out.ps1 -level 1..5 [options]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	opts.Seeded = flag.Lookup("seed").Value.String() != "0"
	if opts.Seeded {
		opts.Seed = seed
	}
	if flag.Lookup("minfrag") != nil {
		opts.MinFragSet = true
	}
	if flag.Lookup("maxfrag") != nil {
		opts.MaxFragSet = true
	}
	return opts
}

func applyProfileDefaults(opts *Options) error {
	switch strings.ToLower(opts.Profile) {
	case "":
	case "light":
		if opts.Pipeline == "" {
			opts.Pipeline = "iden,stringdict,numenc,frag"
		}
		if !opts.Seeded {
			opts.Seeded = true
			opts.Seed = 1337
		}
		if opts.FragProfile == "" {
			opts.FragProfile = "profile=tight"
		}
	case "balanced":
		if opts.Pipeline == "" {
			opts.Pipeline = "iden,strenc,stringdict,numenc,fmt,cf,dead,frag"
		}
		if !opts.Seeded {
			opts.Seeded = true
			opts.Seed = 424242
		}
		if opts.FragProfile == "" {
			opts.FragProfile = "profile=medium"
		}
		if opts.StrEnc == "off" {
			opts.StrEnc = "xor"
			if opts.StrKeyHex == "" {
				opts.StrKeyHex = "a1b2c3d4"
			}
		}
		if opts.StringDict == 0 {
			opts.StringDict = 30
		}
		if opts.DeadProb == 0 {
			opts.DeadProb = 10
		}
	case "heavy":
		if opts.Pipeline == "" {
			opts.Pipeline = "iden,strenc,stringdict,numenc,fmt,cf,dead,frag"
		}
		if !opts.Seeded {
			opts.Seeded = true
			opts.Seed = 987654321
		}
		if opts.FragProfile == "" {
			opts.FragProfile = "profile=loose"
		}
		if opts.StrEnc == "off" {
			opts.StrEnc = "rc4"
			if opts.StrKeyHex == "" {
				opts.StrKeyHex = "00112233445566778899aabbccddeeff"
			}
		}
		if opts.StringDict == 0 {
			opts.StringDict = 50
		}
		if opts.DeadProb == 0 {
			opts.DeadProb = 25
		}
	default:
		return fmt.Errorf("invalid -profile: %s", opts.Profile)
	}
	return nil
}

func applyFragProfile(opts *Options) error {
	if opts.FragProfile == "" {
		return nil
	}
	kv := strings.SplitN(opts.FragProfile, "=", 2)
	if len(kv) != 2 || kv[0] != "profile" {
		return fmt.Errorf("invalid -frag value: %s", opts.FragProfile)
	}
	switch strings.ToLower(kv[1]) {
	case "tight":
		if !opts.MinFragSet && !opts.MaxFragSet {
			opts.MinFrag, opts.MaxFrag = 6, 10
		}
	case "medium":
		if !opts.MinFragSet && !opts.MaxFragSet {
			opts.MinFrag, opts.MaxFrag = 10, 18
		}
	case "loose":
		if !opts.MinFragSet && !opts.MaxFragSet {
			opts.MinFrag, opts.MaxFrag = 14, 28
		}
	default:
		return fmt.Errorf("unknown fragment profile: %s", kv[1])
	}
	return nil
}

func parseHexKey(h string) ([]byte, error) {
	if h == "" {
		return nil, nil
	}
	if len(h)%2 != 0 {
		return nil, errors.New("hex key length must be even")
	}
	return hex.DecodeString(h)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if pp != "" {
			out = append(out, pp)
		}
	}
	return out
}

func buildPipeline(opts *Options, strKey []byte) ([]Transform, error) {
	var out []Transform
	items := splitCSV(opts.Pipeline)
	for _, it := range items {
		switch strings.ToLower(strings.TrimSpace(it)) {
		case "iden":
			if strings.ToLower(opts.IdenMode) == "obf" || opts.VarRename {
				out = append(out, &IdentifierTransform{})
			}
		case "strenc":
			if opts.StrEnc == "xor" {
				out = append(out, &StringEncryptTransform{Mode: "xor", Key: strKey})
			} else if opts.StrEnc == "rc4" {
				out = append(out, &StringEncryptTransform{Mode: "rc4", Key: strKey})
			}
		case "stringdict":
			if opts.StringDict > 0 {
				out = append(out, &StringDictTransform{Percent: opts.StringDict})
			}
		case "numenc":
			if opts.NumEnc {
				out = append(out, &NumberEncodeTransform{})
			}
		case "fmt":
			if strings.ToLower(opts.FormatMode) == "jitter" {
				out = append(out, &FormatJitterTransform{})
			}
		case "cf":
			if opts.CFOpaque {
				out = append(out, &CFOpaqueTransform{})
			}
			if opts.CFShuffle {
				out = append(out, &CFShuffleTransform{})
			}
		case "dead":
			if opts.DeadProb > 0 {
				out = append(out, &DeadCodeTransform{Prob: opts.DeadProb})
			}
		case "frag":
		default:
			if it != "" {
				return nil, fmt.Errorf("unknown pipeline item: %s", it)
			}
		}
	}
	return out, nil
}
