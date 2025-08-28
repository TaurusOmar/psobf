package obfuscator

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

func obfuscate(ps string, level int, noExec bool, fragRange [2]int) (string, error) {
	switch level {
	case 1:
		payload := charsJoinPayload(ps)
		if noExec {
			return payload, nil
		}
		return fmt.Sprintf("$obfuscated = %s; Invoke-Expression $obfuscated", payload), nil
	case 2:
		enc := base64.StdEncoding.EncodeToString([]byte(ps))
		if noExec {
			return enc, nil
		}
		return fmt.Sprintf("$obfuscated = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('%s')); Invoke-Expression $obfuscated", enc), nil
	case 3:
		enc := base64.StdEncoding.EncodeToString([]byte(ps))
		if noExec {
			return enc, nil
		}
		return fmt.Sprintf("$e = [Convert]::FromBase64String('%s'); $obfuscated = [Text.Encoding]::UTF8.GetString($e); Invoke-Expression $obfuscated", enc), nil
	case 4:
		enc, err := gzipAndB64(ps)
		if err != nil {
			return "", err
		}
		if noExec {
			return enc, nil
		}
		return fmt.Sprintf("$compressed = '%s'; $bytes = [Convert]::FromBase64String($compressed); $ms = New-Object IO.MemoryStream(,$bytes); $gz = New-Object IO.Compression.GzipStream($ms,[IO.Compression.CompressionMode]::Decompress); $sr = New-Object IO.StreamReader($gz); $obfuscated = $sr.ReadToEnd(); Invoke-Expression $obfuscated", enc), nil
	case 5:
		frags := fragment(ps, fragRange[0], fragRange[1])
		joined := "@('" + strings.Join(escapePSFragments(frags), "','") + "')"
		if noExec {
			return joined, nil
		}
		return fmt.Sprintf("$fragments = %s; $script = $fragments -join ''; Invoke-Expression $script", joined), nil
	default:
		return "", fmt.Errorf("unsupported level: %d (valid 1..5)", level)
	}
}

func charsJoinPayload(s string) string {
	nums := make([]string, 0, len(s))
	for _, ch := range s {
		nums = append(nums, strconv.Itoa(int(ch)))
	}
	return fmt.Sprintf("$([char[]](%s) -join '')", strings.Join(nums, ","))
}

func gzipAndB64(s string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(s)); err != nil {
		_ = gz.Close()
		return "", fmt.Errorf("gzip write: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func fragment(s string, minFrag, maxFrag int) []string {
	if minFrag < 4 {
		minFrag = 4
	}
	if maxFrag < minFrag {
		maxFrag = minFrag + 6
	}
	var out []string
	for i := 0; i < len(s); {
		size := maxFrag
		if maxFrag > minFrag {
			size = minFrag + (len(s)-i)%(maxFrag-minFrag+1)
			if size < minFrag {
				size = minFrag
			}
		}
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		out = append(out, s[i:end])
		i = end
	}
	return out
}

func escapePSFragments(frags []string) []string {
	out := make([]string, len(frags))
	for i, f := range frags {
		out[i] = strings.ReplaceAll(f, "'", "''")
	}
	return out
}
