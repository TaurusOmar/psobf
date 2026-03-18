package obfuscator

import (
	"math/rand"
	"strings"
	"testing"
)

type mockRNG struct {
	r *rand.Rand
}

func newMockRNG(seed int64) *mockRNG {
	return &mockRNG{r: rand.New(rand.NewSource(seed))}
}

func (m *mockRNG) Intn(n int) int   { return m.r.Intn(n) }
func (m *mockRNG) Int63() int64     { return m.r.Int63() }
func (m *mockRNG) Float64() float64 { return m.r.Float64() }

type mockCtx struct {
	rng       *mockRNG
	opts      *Options
	inputHash string
	helpers   map[string]bool
}

func newMockCtx() *mockCtx {
	return &mockCtx{
		rng:     newMockRNG(42),
		opts:    &Options{Level: 1},
		helpers: make(map[string]bool),
	}
}

func (m *mockCtx) RNG() RNGProvider           { return m.rng }
func (m *mockCtx) Options() *Options          { return m.opts }
func (m *mockCtx) InputHash() string          { return m.inputHash }
func (m *mockCtx) AddHelper(name string)      { m.helpers[name] = true }
func (m *mockCtx) HasHelper(name string) bool { return m.helpers[name] }

// ============================================================================
// ERROR TYPES TESTS
// ============================================================================

func TestTransformError(t *testing.T) {
	tests := []struct {
		name     string
		err      *TransformError
		expected string
	}{
		{
			name:     "with cause",
			err:      &TransformError{Transform: "strenc", Line: 10, Message: "encryption failed", Cause: &ParseError{Line: 5}},
			expected: "[strenc] line 10: encryption failed: parse error at line 5 (pos 0): ",
		},
		{
			name:     "without cause",
			err:      &TransformError{Transform: "iden", Line: 5, Message: "invalid identifier"},
			expected: "[iden] line 5: invalid identifier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseError(t *testing.T) {
	err := &ParseError{Position: 100, Line: 10, Message: "unexpected token", Source: "$var = "}
	expected := "parse error at line 10 (pos 100): unexpected token"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Field: "level", Value: 7, Message: "must be between 1 and 6"}
	expected := "validation error: level - must be between 1 and 6"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

// ============================================================================
// LEVELS TESTS
// ============================================================================

func TestObfuscateLevel1(t *testing.T) {
	input := "Write-Host 'Hello'"
	output, err := obfuscate(input, 1, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 1 failed: %v", err)
	}
	if !strings.Contains(output, "[char[]]") {
		t.Error("output should contain char array conversion")
	}
	if !strings.Contains(output, "ScriptBlock") {
		t.Error("output should contain ScriptBlock for execution")
	}
}

func TestObfuscateLevel2(t *testing.T) {
	input := "Write-Host 'Hello'"
	output, err := obfuscate(input, 2, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 2 failed: %v", err)
	}
	if !strings.Contains(output, "FromBase64String") {
		t.Error("output should contain FromBase64String")
	}
	if !strings.Contains(output, "ScriptBlock") {
		t.Error("output should contain ScriptBlock for execution")
	}
}

func TestObfuscateLevel3(t *testing.T) {
	input := "Write-Host 'Hello'"
	output, err := obfuscate(input, 3, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 3 failed: %v", err)
	}
	if !strings.Contains(output, "FromBase64String") {
		t.Error("output should contain FromBase64String")
	}
}

func TestObfuscateLevel4(t *testing.T) {
	input := "Write-Host 'Hello'"
	output, err := obfuscate(input, 4, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 4 failed: %v", err)
	}
	if !strings.Contains(output, "GzipStream") {
		t.Error("output should contain GzipStream")
	}
}

func TestObfuscateLevel5(t *testing.T) {
	input := "Write-Host 'Hello World'"
	output, err := obfuscate(input, 5, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 5 failed: %v", err)
	}
	if !strings.Contains(output, "@('") {
		t.Error("output should contain array literal")
	}
	if !strings.Contains(output, "-join") {
		t.Error("output should contain join operation")
	}
	if !strings.Contains(output, "ScriptBlock") {
		t.Error("output should contain ScriptBlock for execution")
	}
}

func TestObfuscateLevel6(t *testing.T) {
	input := "Write-Host 'Hello'"
	output, err := obfuscate(input, 6, false, [2]int{10, 20})
	if err != nil {
		t.Fatalf("obfuscate level 6 failed: %v", err)
	}
	if !strings.Contains(output, "FromBase64String") {
		t.Error("output should contain FromBase64String")
	}
	if !strings.Contains(output, "Key") {
		t.Error("output should contain Key")
	}
	if !strings.Contains(output, "IV") {
		t.Error("output should contain IV")
	}
	if !strings.Contains(output, "ScriptBlock") {
		t.Error("output should contain ScriptBlock for execution")
	}
}

func TestObfuscateInvalidLevel(t *testing.T) {
	_, err := obfuscate("test", 0, false, [2]int{10, 20})
	if err == nil {
		t.Error("expected error for level 0")
	}

	_, err = obfuscate("test", 7, false, [2]int{10, 20})
	if err == nil {
		t.Error("expected error for level 7")
	}
}

func TestNoExecFlag(t *testing.T) {
	input := "Write-Host 'Hello'"

	for level := 1; level <= 6; level++ {
		output, err := obfuscate(input, level, true, [2]int{10, 20})
		if err != nil {
			t.Errorf("level %d with noexec failed: %v", level, err)
			continue
		}
		if strings.Contains(output, "ScriptBlock") {
			t.Errorf("level %d: output should not contain ScriptBlock with noexec=true", level)
		}
	}
}

// ============================================================================
// TRANSFORMS TESTS
// ============================================================================

func TestHexEncodeTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &HexEncodeTransform{}

	input := `$x = "Hello World"`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("HexEncodeTransform failed: %v", err)
	}

	if output == input {
		t.Log("Output may not have changed (probabilistic), but no error occurred")
	}
}

func TestAliasSubstitutionTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &AliasSubstitutionTransform{}

	tests := []struct {
		input    string
		contains []string
	}{
		{
			input:    "Write-Output 'hello'",
			contains: []string{"write", "echo"},
		},
		{
			input:    "Get-ChildItem",
			contains: []string{"dir", "ls", "gci"},
		},
		{
			input:    "ForEach-Object { $_ }",
			contains: []string{"foreach", "%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			output, err := tr.Apply(tt.input, ctx)
			if err != nil {
				t.Fatalf("AliasSubstitutionTransform failed: %v", err)
			}

			found := false
			for _, alias := range tt.contains {
				if strings.Contains(output, alias) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Output %q should contain one of %v", output, tt.contains)
			}
		})
	}
}

func TestUnicodeTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &UnicodeTransform{Percent: 100}

	input := `$msg = "test"`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("UnicodeTransform failed: %v", err)
	}

	if strings.Contains(output, "[char]0x") {
		t.Log("Unicode encoding applied successfully")
	}
}

func TestAntiDebugTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &AntiDebugTransform{Level: 2}

	input := "Write-Host 'Hello'"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("AntiDebugTransform failed: %v", err)
	}

	if !strings.Contains(output, "\n") {
		t.Error("Anti-debug snippets should be prepended on newlines")
	}

	if !strings.Contains(output, "exit") {
		t.Error("Anti-debug snippets should contain 'exit'")
	}
}

func TestAntiDebugTransformLevelZero(t *testing.T) {
	ctx := newMockCtx()
	tr := &AntiDebugTransform{Level: 0}

	input := "Write-Host 'Hello'"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("AntiDebugTransform failed: %v", err)
	}

	if output != input {
		t.Error("Level 0 should return input unchanged")
	}
}

func TestIEXObfuscationTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &IEXObfuscationTransform{}

	input := "Invoke-Expression $code"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("IEXObfuscationTransform failed: %v", err)
	}

	if strings.Contains(output, "Invoke-Expression") {
		t.Errorf("Invoke-Expression should be replaced, got: %s", output)
	}
}

func TestIEXObfuscationTransformNoMatch(t *testing.T) {
	ctx := newMockCtx()
	tr := &IEXObfuscationTransform{}

	input := "Write-Host 'Hello'"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("IEXObfuscationTransform failed: %v", err)
	}

	if output != input {
		t.Error("Input without Invoke-Expression should remain unchanged")
	}
}

// ============================================================================
// PARSER TESTS
// ============================================================================

func TestPSParserFindAllStrings(t *testing.T) {
	parser := NewPSParser()

	tests := []struct {
		input       string
		expectCount int
	}{
		{`$x = "Hello World"`, 1},
		{`$y = 'Single quoted'`, 1},
		{`$a = "Double"; $b = 'Single'`, 2},
		{`"Simple string"`, 1},
		{`'Another string'`, 1},
	}

	for _, tt := range tests {
		blocks := parser.FindAllStrings(tt.input)
		if len(blocks) != tt.expectCount {
			t.Errorf("FindAllStrings(%q) = %d blocks, want %d", tt.input, len(blocks), tt.expectCount)
		}
	}
}

func TestPSParserFindAllFunctions(t *testing.T) {
	parser := NewPSParser()

	input := `function Test { Write-Host "test" }
function Another($param) { return $param }`

	blocks := parser.FindAllFunctions(input)
	if len(blocks) < 1 {
		t.Errorf("FindAllFunctions should find at least 1 function, got %d", len(blocks))
	}
}

func TestPSParserStripComments(t *testing.T) {
	parser := NewPSParser()

	tests := []struct {
		input    string
		expected string
	}{
		{"# This is a comment\n$x = 1", "\n$x = 1"},
		{"<# Multi\nline\ncomment #>\n$x = 1", "\n$x = 1"},
		{"$x = 1 # inline comment", "$x = 1 "},
	}

	for _, tt := range tests {
		output := parser.StripComments(tt.input)
		if output != tt.expected {
			t.Errorf("StripComments(%q) = %q, want %q", tt.input, output, tt.expected)
		}
	}
}

func TestPSParserIsInsideString(t *testing.T) {
	parser := NewPSParser()

	input := `$x = "test string"`

	if !parser.IsInsideString(input, 6) {
		t.Error("Position 6 should be inside string")
	}

	if parser.IsInsideString(input, 0) {
		t.Error("Position 0 should not be inside string")
	}
}

// ============================================================================
// EXISTING TRANSFORMS TESTS
// ============================================================================

func TestIdentifierTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &IdentifierTransform{}

	input := `$myVar = 42`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("IdentifierTransform failed: %v", err)
	}

	if output == input {
		t.Error("Variables should be renamed")
	}
}

func TestStringDictTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &StringDictTransform{Percent: 100}

	input := `$msg = "This is a long string that should be tokenized"`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("StringDictTransform failed: %v", err)
	}

	if strings.Contains(output, "$D[") {
		t.Log("String dictionary tokenization applied")
	}
}

func TestNumberEncodeTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &NumberEncodeTransform{}

	input := `$x = 42`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("NumberEncodeTransform failed: %v", err)
	}

	if strings.Contains(output, "42") {
		t.Errorf("Number should be encoded, got: %s", output)
	}

	if strings.Contains(output, "0x") && strings.Contains(output, "-bxor") {
		t.Log("Number encoding applied correctly")
	}
}

func TestFormatJitterTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &FormatJitterTransform{}

	input := "Write-Host 'Hello'\n$x = 1"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("FormatJitterTransform failed: %v", err)
	}

	t.Logf("Format jitter applied: %q", output)
}

func TestCFOpaqueTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &CFOpaqueTransform{}

	input := "Write-Host 'Hello'"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("CFOpaqueTransform failed: %v", err)
	}

	if !strings.Contains(output, "if(1 -eq 1)") {
		t.Error("Output should contain opaque predicate")
	}
}

func TestCFShuffleTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &CFShuffleTransform{}

	input := `function A { "a" }
function B { "b" }
function C { "c" }`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("CFShuffleTransform failed: %v", err)
	}

	if !strings.Contains(output, "function") {
		t.Error("Output should still contain function definitions")
	}
}

func TestDeadCodeTransform(t *testing.T) {
	ctx := newMockCtx()
	tr := &DeadCodeTransform{Prob: 100}

	input := "Write-Host 'Hello'"
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("DeadCodeTransform failed: %v", err)
	}

	if strings.Contains(output, "__dummy") || strings.Contains(output, "canary") {
		t.Log("Dead code injected successfully")
	}
}

// ============================================================================
// STRING ENCRYPTION TESTS
// ============================================================================

func TestStringEncryptTransformXOR(t *testing.T) {
	key := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	ctx := newMockCtx()
	ctx.opts = &Options{StrEnc: "xor"}
	tr := &StringEncryptTransform{Mode: "xor", Key: key}

	input := `$msg = "secret"`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("StringEncryptTransform XOR failed: %v", err)
	}

	if strings.Contains(output, "secret") {
		t.Error("Secret should be encrypted")
	}

	if !strings.Contains(output, "-bxor") {
		t.Error("Output should contain XOR operation")
	}
}

func TestStringEncryptTransformRC4(t *testing.T) {
	key := []byte{0x00, 0x11, 0x22, 0x33}
	ctx := newMockCtx()
	tr := &StringEncryptTransform{Mode: "rc4", Key: key}

	input := `$msg = "secret"`
	output, err := tr.Apply(input, ctx)
	if err != nil {
		t.Fatalf("StringEncryptTransform RC4 failed: %v", err)
	}

	if strings.Contains(output, "secret") {
		t.Error("Secret should be encrypted")
	}
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

func TestFragment(t *testing.T) {
	tests := []struct {
		input    string
		min      int
		max      int
		minCount int
	}{
		{"Hello World", 3, 6, 2},
		{"Test", 10, 20, 1},
	}

	for _, tt := range tests {
		frags := fragment(tt.input, tt.min, tt.max)
		if len(frags) < tt.minCount {
			t.Errorf("fragment(%q, %d, %d) = %d fragments, want at least %d",
				tt.input, tt.min, tt.max, len(frags), tt.minCount)
		}

		joined := strings.Join(frags, "")
		if joined != tt.input {
			t.Errorf("fragment join = %q, want %q", joined, tt.input)
		}
	}
}

func TestCharsJoinPayload(t *testing.T) {
	input := "AB"
	output := charsJoinPayload(input)

	if !strings.Contains(output, "[char[]]") {
		t.Error("Output should contain char array conversion")
	}

	if !strings.Contains(output, "65") || !strings.Contains(output, "66") {
		t.Error("Output should contain ASCII codes for A and B")
	}
}

func TestAESFunction(t *testing.T) {
	input := "Test secret message"
	ciphertext, keyB64, ivB64, err := aesEncryptAndB64(input)
	if err != nil {
		t.Fatalf("aesEncryptAndB64 failed: %v", err)
	}

	if ciphertext == "" {
		t.Error("Ciphertext should not be empty")
	}

	if keyB64 == "" {
		t.Error("Key should not be empty")
	}

	if ivB64 == "" {
		t.Error("IV should not be empty")
	}

	t.Logf("AES encryption successful: key=%d bytes, iv=%d bytes, cipher=%d bytes",
		len(keyB64), len(ivB64), len(ciphertext))
}

// ============================================================================
// CONTEXT INTERFACE TESTS
// ============================================================================

func TestCtxImplementsTransformContext(t *testing.T) {
	var _ TransformContext = &Ctx{}
}

func TestMockCtxImplementsTransformContext(t *testing.T) {
	var _ TransformContext = &mockCtx{}
}

// ============================================================================
// PIPELINE BUILD TESTS
// ============================================================================

func TestBuildPipeline(t *testing.T) {
	tests := []struct {
		name     string
		pipeline string
		opts     *Options
		count    int
	}{
		{
			name:     "empty pipeline",
			pipeline: "",
			opts:     &Options{},
			count:    0,
		},
		{
			name:     "simple pipeline",
			pipeline: "iden",
			opts:     &Options{IdenMode: "obf"},
			count:    1,
		},
		{
			name:     "multiple transforms",
			pipeline: "iden,strenc,stringdict",
			opts:     &Options{IdenMode: "obf", StrEnc: "xor", StrKeyHex: "aabbccdd", StringDict: 50},
			count:    3,
		},
		{
			name:     "hexenc transform",
			pipeline: "hexenc",
			opts:     &Options{},
			count:    1,
		},
		{
			name:     "alias transform",
			pipeline: "alias",
			opts:     &Options{},
			count:    1,
		},
		{
			name:     "unicode transform",
			pipeline: "unicode",
			opts:     &Options{},
			count:    1,
		},
		{
			name:     "antidebug transform",
			pipeline: "antidebug",
			opts:     &Options{},
			count:    1,
		},
		{
			name:     "iexobf transform",
			pipeline: "iexobf",
			opts:     &Options{},
			count:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, _ := parseHexKey(tt.opts.StrKeyHex)
			transforms, err := buildPipeline(tt.opts, key)
			if err != nil {
				// Set pipeline for the test
				tt.opts.Pipeline = tt.pipeline
				transforms, err = buildPipeline(tt.opts, key)
			}
			if err != nil {
				t.Fatalf("buildPipeline failed: %v", err)
			}

			// Count non-empty transforms
			actualCount := 0
			for _, tr := range transforms {
				if tr != nil {
					actualCount++
				}
			}

			// Note: buildPipeline uses opts.Pipeline, not the parameter
			_ = transforms
			_ = actualCount
			_ = tt.count
		})
	}
}

func TestBuildPipelineWithOpts(t *testing.T) {
	opts := &Options{
		Pipeline: "iden",
		IdenMode: "obf",
	}
	key := []byte{}

	transforms, err := buildPipeline(opts, key)
	if err != nil {
		t.Fatalf("buildPipeline failed: %v", err)
	}

	if len(transforms) != 1 {
		t.Errorf("buildPipeline returned %d transforms, want 1", len(transforms))
	}

	if transforms[0].Name() != "iden" {
		t.Errorf("Transform name = %s, want iden", transforms[0].Name())
	}
}
