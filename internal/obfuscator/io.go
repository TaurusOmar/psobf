package obfuscator

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readAllInput(opts Options) ([]byte, error) {
	if opts.UseStdin {
		return io.ReadAll(bufio.NewReader(os.Stdin))
	}
	return os.ReadFile(opts.InputFile)
}

func fuzzOutName(base string, i int) string {
	if base == "" {
		return fmt.Sprintf("obfuscated.v%d.ps1", i)
	}
	if strings.HasSuffix(strings.ToLower(base), ".ps1") {
		return strings.TrimSuffix(base, ".ps1") + fmt.Sprintf(".v%d.ps1", i)
	}
	return base + fmt.Sprintf(".v%d", i)
}

func requireInOut(opts Options) error {
	if !opts.UseStdin && opts.InputFile == "" && opts.Fuzz == 0 {
		return errors.New("./psobf -i <inputFile> -o <outputFile> -level <1|2|3|4|5> [options]")
	}
	if !opts.UseStdout && opts.OutputFile == "" && opts.Fuzz == 0 {
		return errors.New("missing -o or -stdout")
	}
	return nil
}
