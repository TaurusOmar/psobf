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
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func obfuscate(word string, level int) string {
	switch level {
	case 1:
		return obfuscateLevel1(word)
	case 2:
		return obfuscateLevel2(word)
	case 3:
		return obfuscateLevel3(word)
	case 4:
		return obfuscateLevel4(word)
	case 5:
		return obfuscateLevel5(word)
	default:
		log.Panicf("Unsupported obfuscation level: %d", level)
		return ""
	}
}

func obfuscateLevel1(word string) string {
	concatenated := ""
	for _, char := range word {
		concatenated += obfuscateCharacter(char)
	}
	concatenated = strings.TrimSuffix(concatenated, ",")
	return fmt.Sprintf("$obfuscated = $([char[]](%s) -join ''); Invoke-Expression $obfuscated", concatenated)
}

func obfuscateLevel2(word string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(word))
	return fmt.Sprintf("$obfuscated = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s')); Invoke-Expression $obfuscated", encoded)
}

func obfuscateLevel3(word string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(word))
	return fmt.Sprintf("$e = [System.Convert]::FromBase64String('%s'); $obfuscated = [System.Text.Encoding]::UTF8.GetString($e); Invoke-Expression $obfuscated", encoded)
}

func obfuscateLevel4(word string) string {
	encoded := compressAndEncode(word)
	return fmt.Sprintf(`$compressed = '%s'; $bytes = [System.Convert]::FromBase64String($compressed); $stream = New-Object IO.MemoryStream(, $bytes); $decompressed = New-Object IO.Compression.GzipStream($stream, [IO.Compression.CompressionMode]::Decompress); $reader = New-Object IO.StreamReader($decompressed); $obfuscated = $reader.ReadToEnd(); Invoke-Expression $obfuscated`, encoded)
}

func obfuscateLevel5(word string) string {
	fragments := fragmentScript(word)
	return fmt.Sprintf(`$fragments = @('%s'); $script = $fragments -join ''; Invoke-Expression $script`, strings.Join(fragments, "','"))
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

func copyToClipboard(text string) error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("xsel", "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	default:
		return fmt.Errorf("clipboard functionality is not supported on this OS: %s", runtime.GOOS)
	}
}

func processInput(inputFile, command string) (string, error) {
	if command != "" {
		return command, nil
	} else if inputFile != "" {
		content, err := os.ReadFile(inputFile)
		if err != nil {
			return "", fmt.Errorf("could not read the input file: %w", err)
		}
		return string(content), nil
	} else {
		return "", fmt.Errorf("no input provided")
	}
}

func main() {
	fmt.Println(`
	██████╗ ███████╗ ██████╗ ██████╗ ███████╗
	██╔══██╗██╔════╝██╔═══██╗██╔══██╗██╔════╝
	██████╔╝███████╗██║   ██║██████╔╝█████╗  
	██╔═══╝ ╚════██║██║   ██║██╔══██╗██╔══╝  
	██║     ███████║╚██████╔╝██████╔╝██║     
	╚═╝     ╚══════╝ ╚═════╝ ╚═╝     
	@TaurusOmar 
	v.1.3						  
	`)

	inputFile := flag.String("i", "", "Name of the PowerShell script file.")
	inputFileAlias := flag.String("input", "", "Alias for the name of the PowerShell script file.")
	outputFile := flag.String("o", "obfuscated.ps1", "Name of the output file for the obfuscated script.")
	outputFileAlias := flag.String("output", "", "Alias for the name of the output file for the obfuscated script.")
	level := flag.Int("l", 1, "Obfuscation level (1 to 5).")
	levelAlias := flag.Int("level", 0, "Alias for the obfuscation level.")
	command := flag.String("c", "", "PowerShell command to obfuscate.")
	commandAlias := flag.String("command", "", "Alias for the PowerShell command to obfuscate.")
	toClipboard := flag.Bool("clipboard", false, "Copy the obfuscated script to the clipboard instead of writing to a file.")
	toClipboardAlias := flag.Bool("clip", false, "Alias to copy the obfuscated script to the clipboard.")

	flag.Usage = func() {
		fmt.Println("Usage: ./psobf [-i <inputFile> | --input <inputFile> | -c <command> | --command <command>] [-o <outputFile> | --output <outputFile>] [-l <1|2|3|4|5> | --level <1|2|3|4|5>] [--clipboard | -clip]")
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

	if *inputFileAlias != "" {
		*inputFile = *inputFileAlias
	}
	if *outputFileAlias != "" {
		*outputFile = *outputFileAlias
	}
	if *levelAlias != 0 {
		*level = *levelAlias
	}
	if *commandAlias != "" {
		*command = *commandAlias
	}
	if *toClipboardAlias {
		*toClipboard = true
	}

	if *inputFile == "" && *command == "" {
		fmt.Println("Error: Either -i (input file) or -c (command) must be provided.")
		flag.Usage()
		return
	}

	psScript, err := processInput(*inputFile, *command)
	if err != nil {
		log.Panic(err)
	}

	obfuscated := obfuscate(psScript, *level)

	if *toClipboard {
		err := copyToClipboard(obfuscated)
		if err != nil {
			log.Panicf("Could not copy to clipboard: %v", err)
		}
		fmt.Println("The obfuscated script has been copied to the clipboard.")
	} else {
		err = os.WriteFile(*outputFile, []byte(obfuscated), 0644)
		if err != nil {
			log.Panicf("Could not write to the file: %v", err)
		}
		fmt.Println("The obfuscated script has been written to", *outputFile)
	}
}
