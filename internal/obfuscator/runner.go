package obfuscator

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"
)

func Run(opts Options) error {
	if !opts.Quiet {
		fmt.Print(banner)
	}
	if err := applyProfileDefaults(&opts); err != nil {
		return err
	}
	if err := applyFragProfile(&opts); err != nil {
		return err
	}
	if opts.Fuzz > 0 && opts.UseStdout {
		return errors.New("cannot use -fuzz with -stdout")
	}
	if err := requireInOut(opts); err != nil {
		return err
	}
	if opts.Level < 1 || opts.Level > 5 {
		return fmt.Errorf("invalid level: %d (valid 1..5)", opts.Level)
	}
	if opts.StringDict < 0 || opts.StringDict > 100 {
		return fmt.Errorf("invalid -stringdict: %d (0..100)", opts.StringDict)
	}
	if opts.DeadProb < 0 || opts.DeadProb > 100 {
		return fmt.Errorf("invalid -deadcode: %d (0..100)", opts.DeadProb)
	}
	if opts.StrEnc != "off" && opts.StrEnc != "xor" && opts.StrEnc != "rc4" {
		return fmt.Errorf("invalid -strenc: %s (off|xor|rc4)", opts.StrEnc)
	}
	if (opts.StrEnc == "xor" || opts.StrEnc == "rc4") && opts.StrKeyHex == "" {
		return errors.New("missing -strkey for -strenc xor|rc4")
	}
	key, err := parseHexKey(opts.StrKeyHex)
	if (opts.StrEnc == "xor" || opts.StrEnc == "rc4") && err != nil {
		return fmt.Errorf("invalid -strkey hex: %w", err)
	}
	if opts.Fuzz > 0 {
		data, err := readAllInput(opts)
		if err != nil {
			return fmt.Errorf("could not read input: %w", err)
		}
		for i := 1; i <= opts.Fuzz; i++ {
			tmp := opts
			tmp.Seeded = true
			tmp.Seed = time.Now().UnixNano() + int64(i*137)
			outName := fuzzOutName(opts.OutputFile, i)
			tmp.OutputFile = outName
			if err := processOnce(tmp, data, key); err != nil {
				return fmt.Errorf("fuzz variant %d failed: %w", i, err)
			}
			if !opts.Quiet {
				fmt.Println("Wrote:", outName)
			}
		}
		return nil
	}
	data, err := readAllInput(opts)
	if err != nil {
		return fmt.Errorf("could not read input: %w", err)
	}
	return processOnce(opts, data, key)
}

func processOnce(opts Options, data []byte, strKey []byte) error {
	r := InitRNG(&opts.Seed, opts.Seeded)
	ctx := &Ctx{
		Rng:       r,
		Opts:      &opts,
		InputHash: hex.EncodeToString(SumSha256(data)),
		Helpers:   map[string]bool{},
	}
	ps := string(data)
	if opts.Pipeline != "" || opts.Profile != "" {
		transforms, err := buildPipeline(&opts, strKey)
		if err != nil {
			return err
		}
		var errT error
		for _, t := range transforms {
			ps, errT = t.Apply(ps, ctx)
			if errT != nil {
				return fmt.Errorf("transform %s failed: %w", t.Name(), errT)
			}
		}
	}
	payload, err := obfuscate(ps, opts.Level, opts.NoExec, [2]int{opts.MinFrag, opts.MaxFrag})
	if err != nil {
		return err
	}
	if opts.UseStdout {
		_, err = os.Stdout.Write([]byte(payload))
		return err
	}
	return os.WriteFile(opts.OutputFile, []byte(payload), 0644)
}

func ObfuscateString(ps string, opts Options) (string, error) {
	r := InitRNG(&opts.Seed, opts.Seeded)
	ctx := &Ctx{
		Rng:       r,
		Opts:      &opts,
		InputHash: "",
		Helpers:   map[string]bool{},
	}
	key, err := parseHexKey(opts.StrKeyHex)
	if err != nil {
		return "", err
	}
	if opts.Pipeline != "" || opts.Profile != "" {
		if err := applyProfileDefaults(&opts); err != nil {
			return "", err
		}
		if err := applyFragProfile(&opts); err != nil {
			return "", err
		}
		transforms, err := buildPipeline(&opts, key)
		if err != nil {
			return "", err
		}
		for _, t := range transforms {
			ps, err = t.Apply(ps, ctx)
			if err != nil {
				return "", err
			}
		}
	}
	return obfuscate(ps, opts.Level, opts.NoExec, [2]int{opts.MinFrag, opts.MaxFrag})
}
