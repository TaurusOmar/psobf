package obfuscator

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

var iexAlternativesNames = []string{
	"Invoke-Expression",
	"iex",
	"IEX",
}

var iexAlternativeMethods = []string{
	"$ExecutionContext.InvokeCommand.InvokeScript($script)",
	"[ScriptBlock]::Create($script).Invoke()",
	". ([ScriptBlock]::Create($script))",
	"& ([ScriptBlock]::Create($script))",
	"Invoke-Command -ScriptBlock ([ScriptBlock]::Create($script))",
}

func getStealthyIEX(r *rand.Rand) string {
	return iexAlternativeMethods[r.Intn(len(iexAlternativeMethods))]
}

func getVarName(r *rand.Rand) string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	n := 3 + r.Intn(4)
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(chars[r.Intn(len(chars))])
	}
	return b.String()
}

func obfuscateMethodName(name string, r *rand.Rand) string {
	chars := strings.Split(name, "")
	for i := range chars {
		if r.Intn(100) < 30 {
			chars[i] = strings.ToUpper(chars[i])
		} else if r.Intn(100) < 30 {
			chars[i] = strings.ToLower(chars[i])
		} else if r.Intn(100) < 20 {
			chars[i] = "`" + chars[i]
		}
	}
	return strings.Join(chars, "")
}

func obfuscate(ps string, level int, noExec bool, fragRange [2]int) (string, error) {
	seed := int64(42)
	rng := rand.New(rand.NewSource(seed))

	switch level {
	case 1:
		payload := charsJoinPayload(ps)
		if noExec {
			return payload, nil
		}
		v := getVarName(rng)
		return fmt.Sprintf("$%s = %s; . ([ScriptBlock]::Create($%s))", v, payload, v), nil
	case 2:
		enc := base64.StdEncoding.EncodeToString([]byte(ps))
		if noExec {
			return enc, nil
		}
		v := getVarName(rng)
		return fmt.Sprintf("$%s = '%s'; $%s = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($%s)); . ([ScriptBlock]::Create($%s))", v, enc, v, v, v), nil
	case 3:
		enc := base64.StdEncoding.EncodeToString([]byte(ps))
		if noExec {
			return enc, nil
		}
		v1 := getVarName(rng)
		v2 := getVarName(rng)
		return fmt.Sprintf("$%s = '%s'; $%s = [Convert]::FromBase64String($%s); $%s = [Text.Encoding]::UTF8.GetString($%s); . ([ScriptBlock]::Create($%s))", v1, enc, v2, v1, v2, v2, v2), nil
	case 4:
		enc, err := gzipAndB64(ps)
		if err != nil {
			return "", err
		}
		if noExec {
			return enc, nil
		}
		v := getVarName(rng)
		return fmt.Sprintf("$c='%s';$b=[Convert]::FromBase64String($c);$m=New-Object IO.MemoryStream(,$b);$g=New-Object IO.Compression.GzipStream($m,[IO.Compression.CompressionMode]::Decompress);$r=New-Object IO.StreamReader($g);$%s=$r.ReadToEnd();. ([ScriptBlock]::Create($%s))", enc, v, v), nil
	case 5:
		frags := fragment(ps, fragRange[0], fragRange[1])
		joined := "@('" + strings.Join(escapePSFragments(frags), "','") + "')"
		if noExec {
			return joined, nil
		}
		v := getVarName(rng)
		return fmt.Sprintf("$%s = %s; $%s = $%s -join ''; . ([ScriptBlock]::Create($%s))", v, joined, v, v, v), nil
	case 6:
		enc, keyB64, ivB64, err := aesEncryptAndB64(ps)
		if err != nil {
			return "", fmt.Errorf("aes encryption: %w", err)
		}
		if noExec {
			return enc, nil
		}
		v1 := getVarName(rng)
		v2 := getVarName(rng)
		ps1 := fmt.Sprintf("$k=[Convert]::FromBase64String('%s');$iv=[Convert]::FromBase64String('%s');$e=[Convert]::FromBase64String('%s')", keyB64, ivB64, enc)
		ps2 := fmt.Sprintf("$a=New-Object (\"Secu\"+\"rity.Cryptography.Aes\"+\"Managed\");$a.Mode='CBC';$a.Padding='PKCS7';$a.Key=$k;$a.IV=$iv")
		ps3 := fmt.Sprintf("$d=$a.CreateDecryptor();$%s=$d.TransformFinalBlock($e,0,$e.Length);$%s=[Text.Encoding]::UTF8.GetString($%s);. ([ScriptBlock]::Create($%s))", v1, v2, v1, v2)
		return ps1 + ";" + ps2 + ";" + ps3, nil
	default:
		return "", fmt.Errorf("unsupported level: %d (valid 1..6)", level)
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

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func aesEncryptAndB64(s string) (ciphertext, keyB64, ivB64 string, err error) {
	key := make([]byte, 32)
	if _, err := cryptorand.Read(key); err != nil {
		return "", "", "", fmt.Errorf("generate key: %w", err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := cryptorand.Read(iv); err != nil {
		return "", "", "", fmt.Errorf("generate iv: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", "", fmt.Errorf("create cipher: %w", err)
	}
	plaintext := pkcs7Pad([]byte(s), aes.BlockSize)
	ciphertextBytes := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertextBytes, plaintext)
	keyB64 = base64.StdEncoding.EncodeToString(key)
	ivB64 = base64.StdEncoding.EncodeToString(iv)
	ciphertext = base64.StdEncoding.EncodeToString(ciphertextBytes)
	return ciphertext, keyB64, ivB64, nil
}
