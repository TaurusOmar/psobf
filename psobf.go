package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func obfuscate(word string, level int) string {
	switch level {
	case 1:
		concatenated := ""
		for _, char := range word {
			concatenated += obfuscateCharacter(char)
		}
		concatenated = strings.TrimSuffix(concatenated, ",")
		return fmt.Sprintf("$obfuscated = $([char[]](%s) -join ''); Invoke-Expression $obfuscated", concatenated)
	case 2:
		encoded := base64.StdEncoding.EncodeToString([]byte(word))
		return fmt.Sprintf("$obfuscated = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')); Invoke-Expression $obfuscated", encoded)
	case 3:
		encoded := base64.StdEncoding.EncodeToString([]byte(word))
		return fmt.Sprintf("$e = [System.Convert]::FromBase64String('%s'); $obfuscated = [System.Text.Encoding]::UTF8.GetString($e); Invoke-Expression $obfuscated", encoded)
	case 4:
		encoded := compressAndEncode(word)
		return fmt.Sprintf("$compressed = '%s'; $bytes = [System.Convert]::FromBase64String($compressed); $stream = New-Object IO.MemoryStream(, $bytes); $decompressed = New-Object IO.Compression.GzipStream($stream, [IO.Compression.CompressionMode]::Decompress); $reader = New-Object IO.StreamReader($decompressed); $obfuscated = $reader.ReadToEnd(); Invoke-Expression $obfuscated", encoded)
	case 5:
		fragments := fragmentScript(word)
		return fmt.Sprintf("$fragments = @('%s'); $script = $fragments -join ''; Invoke-Expression $script", strings.Join(fragments, "','"))
	default:
		log.Panicf("Unsupported obfuscation level: %d", level)
		return ""
	}
}

func obfuscateCharacter(char rune) string {
	switch char {
	case '\n':
		return "\"`n\","
	case '\'':
		return "\"'\","
	case '`':
		return "\"``\","
	case '$':
		return "\"`$\","
	case '(':
		return "\"`(\","
	case ')':
		return "\"`)\","
	case '|':
		return "\"`|\","
	default:
		return fmt.Sprintf("'%s',", string(char))
	}
}

func compressAndEncode(word string) string {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write([]byte(word))
	if err != nil {
		log.Panicf("Failed to compress: %v", err)
	}
	w.Close()
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func fragmentScript(script string) []string {
	rand.Seed(time.Now().UnixNano())
	length := len(script)
	fragmentSize := rand.Intn(10) + 10

	var fragments []string
	for i := 0; i < length; i += fragmentSize {
		end := i + fragmentSize
		if end > length {
			end = length
		}

		fragment := script[i:end]
		fragment = strings.ReplaceAll(fragment, "'", "''")
		fragments = append(fragments, fragment)
	}
	return fragments
}

func obfuscateVariables(script string) string {
	variables := findVariables(script)
	for _, variable := range variables {
		obfuscated := randomString(len(variable))
		script = strings.ReplaceAll(script, variable, obfuscated)
	}
	return script
}

func findVariables(script string) []string {
	variables := make(map[string]bool)
	lines := strings.Split(script, "\n")
	for _, line := range lines {
		if strings.Contains(line, "$") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "$") {
					variables[part] = true
				}
			}
		}
	}
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	return keys
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func main() {
	fmt.Println(`
	██████╗ ███████╗ ██████╗ ██████╗ ███████╗
	██╔══██╗██╔════╝██╔═══██╗██╔══██╗██╔════╝
	██████╔╝███████╗██║   ██║██████╔╝█████╗  
	██╔═══╝ ╚════██║██║   ██║██╔══██╗██╔══╝  
	██║     ███████║╚██████╔╝██████╔╝██║     
	╚═╝     ╚══════╝ ╚═════╝ ╚═════╝ ╚═╝     
	@TaurusOmar 
	v.1.0											 								
	  `)
	inputFile := flag.String("i", "", "Name of the PowerShell script file.")
	outputFile := flag.String("o", "obfuscated.ps1", "Name of the output file for the obfuscated script.")
	level := flag.Int("level", 1, "Obfuscation level (1 to 5).")

	flag.Usage = func() {
		fmt.Println("Usage: ./obfuscator -i <inputFile> -o <outputFile> -level <1|2|3|4|5>")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nObfuscation levels:")
		fmt.Println("  1: Basic obfuscation by splitting the script into individual characters.")
		fmt.Println("  2: Base64 encoding of the script.")
		fmt.Println("  3: Alternative Base64 encoding with a different PowerShell decoding method.")
		fmt.Println("  4: Compression and Base64 encoding of the script will be decoded and decompressed at runtime.")
		fmt.Println("  5: Fragmentation of the script into multiple parts and reconstruction at runtime.")
	}

	flag.Parse()

	if *inputFile == "" {
		flag.Usage()
		return
	}

	psScript, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Panicf("Could not read the file: %v", err)
	}

	obfuscated := obfuscate(string(psScript), *level)

	err = os.WriteFile(*outputFile, []byte(obfuscated), 0644)
	if err != nil {
		log.Panicf("Could not write to the file: %v", err)
	}

	fmt.Println("The obfuscated script has been written to", *outputFile)
}
