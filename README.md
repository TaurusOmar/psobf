# PowerShell Obfuscator

Its goal is to transform code to hinder analysis and static signatures, useful in labs and authorized Red Team/Pentesting engagements.

Supports 5 levels of obfuscation plus a transforms/pipeline architecture that allows stacking techniques such as string tokenization, light literal encryption, number masking, identifier morphing, format “jitter,” control-flow cosmetics, dead code injection, fragmentation profiles, and deterministic profiles.

⚠️ Responsible use: this tool is intended for research and authorized testing only.
**Do not** use for malicious purposes.

```powershell 

./psobf -h

	██████╗ ███████╗ ██████╗ ██████╗ ███████╗
	██╔══██╗██╔════╝██╔═══██╗██╔══██╗██╔════╝
	██████╔╝███████╗██║   ██║██████╔╝█████╗
	██╔═══╝ ╚════██║██║   ██║██╔══██╗██╔══╝
	██║     ███████║╚██████╔╝██████╔╝██║
	╚═╝     ╚══════╝ ╚═════╝ ╚═════╝ ╚═╝
	@TaurusOmar
	v.1.1.5											 	
	
Usage: ./psobf -i <inputFile> -o <outputFile> -level <1|2|3|4|5> [options]

  -cf-opaque		Enable opaque predicate wrapper.
  -cf-shuffle		Shuffle function blocks.
  -deadcode int		Dead-code injection probability (0..100).
  -fmt string		Format jitter: off|jitter. (default "off")
  -frag string		Fragmentation profile: profile=tight|medium|loose.
  -fuzz int			Generate N fuzzed variants (unique seeds).
  -i string			PowerShell script input file (use -stdin).
  -iden string		Identifier morphing: obf|keep. (default "keep")
  -level int		Obfuscation level (1..5). (default 1)
  -maxfrag int		Maximum fragment size (level 5). (default 20)
  -minfrag int		Minimum fragment size (level 5). (default 10)
  -noexec			Emit only payload without Invoke-Expression.
  -numenc			Enable number encoding.
  -o string			Output file (use -stdout). (default "obfuscated.ps1")
  -pipeline string	Comma-separated transforms: iden,strenc,stringdict,numenc,fmt,cf,dead
  -profile string	Preset: light|balanced|heavy.
  -q				Quiet mode (no banner).
  -seed int			RNG seed (reproducible). Overrides crypto/rand if set.
  -stdin			Read script from STDIN.
  -stdout			Write result to STDOUT.
  -strenc string	String encryption: off|xor|rc4. (default "off")
  -stringdict int	String tokenization percentage (0..100).
  -strkey string	Hex key for -strenc.
  -varrename		Deprecated: kept for backward-compatibility. Use -iden obf.
```


## Installation

```bash
go install github.com/TaurusOmar/psobf/cmd/psobf@latest
```

---

## Quick start

```bash
psobf -i input.ps1 -o out.ps1 -level 1..5 [options]
psobf -h   # full help
```


---

## Complete flags reference

| Flag          | Type / Values     |          Default | Description                                 | Example                                                      |                                 |                        |
| ------------- | ----------------- | ---------------: | ------------------------------------------- | ------------------------------------------------------------ | ------------------------------- | ---------------------- |
| `-i`          | string            |                — | Input PS1 (use `-stdin` to read from pipe)  | `-i script.ps1`                                              |                                 |                        |
| `-o`          | string            | `obfuscated.ps1` | Output (use `-stdout` to write to STDOUT)   | `-o out.ps1`                                                 |                                 |                        |
| `-level`      | 1..5              |                1 | Final packer (see Levels)                   | `-level 4`                                                   |                                 |                        |
| `-noexec`     | bool              |            false | Emit only payload (no `Invoke-Expression`)  | `-noexec`                                                    |                                 |                        |
| `-stdin`      | bool              |            false | Read PS from STDIN                          | `-stdin`                                                     |                                 |                        |
| `-stdout`     | bool              |            false | Write result to STDOUT                      | `-stdout`                                                    |                                 |                        |
| `-seed`       | int64             |           random | Reproducible randomness                     | `-seed 42`                                                   |                                 |                        |
| `-q`          | bool              |            false | Quiet (no banner)                           | `-q`                                                         |                                 |                        |
| `-pipeline`   | csv               |                — | Transforms to apply in order                | `-pipeline "iden,strenc,stringdict,numenc,fmt,cf,dead,frag"` |                                 |                        |
| `-iden`       | `keep`/`obf`      |           `keep` | Identifier morphing (vars & funcs)          | `-iden obf`                                                  |                                 |                        |
| `-strenc`     | `off`/`xor`/`rc4` |            `off` | String literal encryption                   | `-strenc rc4`                                                |                                 |                        |
| `-strkey`     | hex               |                — | Key for `-strenc`                           | `-strkey 0011223344556677`                                   |                                 |                        |
| `-stringdict` | 0..100            |                0 | Tokenize long strings; % chance per literal | `-stringdict 40`                                             |                                 |                        |
| `-numenc`     | bool              |            false | Encode numbers as arithmetic PS expressions | `-numenc`                                                    |                                 |                        |
| `-fmt`        | `off`/`jitter`    |            `off` | Randomize whitespace/line breaks            | `-fmt jitter`                                                |                                 |                        |
| `-cf-opaque`  | bool              |            false | Wrap in `if(1 -eq 1){...}`                  | `-cf-opaque`                                                 |                                 |                        |
| `-cf-shuffle` | bool              |            false | Reorder **function blocks**                 | `-cf-shuffle`                                                |                                 |                        |
| `-deadcode`   | 0..100            |                0 | Probability to inject dead code             | `-deadcode 20`                                               |                                 |                        |
| `-frag`       | \`profile=tight   |           medium | loose\`                                     | —                                                            | Fragmentation profile (level 5) | `-frag profile=medium` |
| `-minfrag`    | int               |               10 | Min fragment size (level 5)                 | `-minfrag 8`                                                 |                                 |                        |
| `-maxfrag`    | int               |               20 | Max fragment size (level 5)                 | `-maxfrag 16`                                                |                                 |                        |
| `-profile`    | \`light           |         balanced | heavy\`                                     | —                                                            | Presets for pipeline/seed/etc.  | `-profile heavy`       |
| `-fuzz`       | int               |                0 | Produce N variants (different seeds)        | `-fuzz 5`                                                    |                                 |                        |

> The **pipeline** runs **before** the final **`-level`** packing.

---

## Sample input script (safe)

To keep examples benign, we’ll use:

```powershell
Write-Host "Hello, World!"
$answer = 42
function Greet($name) { Write-Host ("Hi, " + $name) }
Greet "Ada"
```

---


## Obfuscation levels (1–5) + output snippets

> The following show the **shape** of outputs (snippets). Actual payloads will differ.

### Level 1 — Char join

```bash
psobf -i sample.ps1 -o out.ps1 -level 1
```

**Output (snippet):**

```powershell
$obfuscated = $([char[]](87,114,105,116,101,45,72,111,115,116,32,34,72,101,108,108,111,44,32,87,111,114,108,100,33,34,10,36,97,110,115,119,101,114,32,61,32,52,50,10,102,117,110,99,116,105,111,110,32,71,114,101,101,116,40,36,110,97,109,101,41,32,123,32,87,114,105,116,101,45,72,111,115,116,32,40,34,72,105,44,32,34,32,43,32,36,110,97,109,101,41,32,125,10,71,114,101,101,116,32,34,65,100,97,34,10) -join ''); Invoke-Expression $obfuscated
```

### Level 2 — Base64

```bash
psobf -i sample.ps1 -o out.ps1 -level 2
```

**Output (snippet):**

```powershell
$obfuscated = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String('V3JpdGUtSG9zdCAiSGVsbG8sIFdvcmxkISIKJGFuc3dlciA9IDQyCmZ1bmN0aW9uIEdyZWV0KCRuYW1lKSB7IFdyaXRlLUhvc3QgKCJIaSwgIiArICRuYW1lKSB9CkdyZWV0ICJBZGEiCg==')); Invoke-Expression $obfuscated
```

### Level 3 — Base64 (alt)

```bash
psobf -i sample.ps1 -o out.ps1 -level 3
```

**Output (snippet):**

```powershell
$e = [Convert]::FromBase64String('V3JpdGUtSG9zdCAiSGVsbG8sIFdvcmxkISIKJGFuc3dlciA9IDQyCmZ1bmN0aW9uIEdyZWV0KCRuYW1lKSB7IFdyaXRlLUhvc3QgKCJIaSwgIiArICRuYW1lKSB9CkdyZWV0ICJBZGEiCg=='); $obfuscated = [Text.Encoding]::UTF8.GetString($e); Invoke-Expression $obfuscated
```

### Level 4 — GZip + Base64

```bash
psobf -i sample.ps1 -o out.ps1 -level 4
```

**Output (snippet):**

```powershell
$compressed = 'H4sIAAAAAAAA/wovyixJ1fXILy5RUPJIzcnJ11EIzy/KSVFU4lJJzCsuTy1SsFUwMeJKK81LLsnMz1NwL0pNLdFQyUvMTdVUqFZA0q+h5JGpo6CkoK0Ala3lAitWUHJMSVTiAgQAAP//m+Ey2GoAAAA='; $bytes = [Convert]::FromBase64String($compressed); $ms = New-Object IO.MemoryStream(,$bytes); $gz = New-Object IO.Compression.GzipStream($ms,[IO.Compression.CompressionMode]::Decompress); $sr = New-Object IO.StreamReader($gz); $obfuscated = $sr.ReadToEnd(); Invoke-Expression $obfuscated
```

### Level 5 — Fragmentation

```bash
psobf -i sample.ps1 -o out.ps1 -level 5
```

**Output (snippet):**

```powershell
$fragments = @('Write-Host "Hello',', World!"
$','answer = 42','
function G','reet($name)',' { Write-Ho','st ("Hi, " ','+ $name) }
','Greet "Ada"','
'); $script = $fragments -join ''; Invoke-Expression $script
```

---

## Transforms (pipeline) — details & examples

> Use `-noexec` to inspect payloads without executing.

### Identifier (`-iden`)

* Renames variables and **functions** while preserving semantics.
* Protect anything you don’t want renamed with prefix `__$`.

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 -pipeline "iden" -iden obf -seed 11
```

**Output (snippet)**

```powershell
$WguE = 42
function QhZy($Chx){ Write-Host ("Hi, " + $Chx) }
QhZy "Ada"
```

---

### String Encryption (`-strenc xor|rc4`)

Encrypts **string literals only** (no API tampering). Decrypts just-in-time at runtime.
Flags: -strenc xor|rc4, -strkey <hex>.

#### XOR

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 -pipeline "strenc" -strenc xor -strkey a1b2c3d4 -seed 42
```

**Output (snippet)**

```powershell
$b=[Convert]::FromBase64String('EwAB...'); for($i=0;$i -lt $b.Length;$i++){$b[$i]=$b[$i] -bxor 0xA1}; [Text.Encoding]::UTF8.GetString($b)
```

#### RC4

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 2 -pipeline "strenc" -strenc rc4 -strkey 0011223344556677 -seed 7
```

**Output (snippet)**

```powershell
function __decGWREVT($k,[byte[]]$d){ $s=0..255; $j=0; for($i=0;$i -lt 256;$i++){ $j=($j+$s[$i]+$k[$i%$k.Length])%256; $t=$s[$i];$s[$i]=$s[$j];$s[$j]=$t } $i=0;$j=0; for($x=0;$x -lt $d.Length;$x++){ $i=($i+1)%256;$j=($j+$s[$i])%256; $t=$s[$i];$s[$i]=$s[$j];$s[$j]=$t; $d[$x]=$d[$x] -bxor $s[($s[$i]+$s[$j])%256] } [Text.Encoding]::UTF8.GetString($d) }
...
( __decGWREVT ([byte[]](0..(8-1)|%{[Convert]::ToByte('0011223344556677'.Substring($_*2,2),16)})) ([Convert]::FromBase64String('m7m7...')) )
```

---

### String Dictionary (`-stringdict`)

Tokenizes long strings into a `$D` array and rebuilds them at runtime. Reduces repetitive signatures.
Flag: -stringdict <0..100>

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 3 -pipeline "stringdict" -stringdict 40 -seed 1
```

**Output (snippet)**

```powershell
$D=@('Hello',', World','!','Hi, ', 'Ada');
Write-Host ($D[0]+$D[1]+$D[2])
function Greet($name){ Write-Host ($D[3] + $name) }
Greet $D[4]
```

---

### Number Encoding (`-numenc`)

Replaces plain numbers with equivalent arithmetic/bitwise expressions (outside strings).

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 2 -pipeline "numenc" -numenc -seed 1337
```

**Output (snippet)**

```powershell
$answer = ((0x2A -bxor 0x00)+0)
```
Caution: Redirects like 2>&1 must remain identical. If your source has unquoted redirects and you're experiencing problems, disable -numenc or encapsulate those redirects in strings in the source.

---

### Format Jitter (`-fmt`)

Randomizes spacing and newlines.

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 2 -pipeline "fmt" -fmt jitter -seed 20
```

**Output (snippet)**

```powershell
Write-Host   "Hello, World!"
$answer=42

function Greet($name) {  Write-Host ("Hi, "+$name) }
Greet  "Ada"
```

---

### Control Flow (`-cf-opaque`, `-cf-shuffle`)

* `-cf-opaque`: wraps the whole script in a never-false branch.
* `-cf-shuffle`: reorders **function blocks** (not single statements). You’ll notice changes only if your script defines functions.

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 -pipeline "cf" -cf-opaque -cf-shuffle -seed 77
```

**Output (snippet)**

```powershell
if(1 -eq 1){
  function Greet($name){ Write-Host ("Hi, " + $name) }
  Write-Host "Hello, World!"
  $answer = 42
  Greet "Ada"
}
```

---

### Dead Code (`-deadcode`)

Injects no-op functions, 0-iteration loops, harmless strings, etc. Controlled by probability.
Flag: -deadcode <0..100> (snippet injection probability).

**Command**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 -pipeline "dead" -deadcode 25 -seed 5
```

**Output (snippet)**

```powershell
function __dummyzQJxJk { return }
for($i=0;$i -lt 0;$i++){Start-Sleep -Milliseconds 0}
$x='canary';$y=$x+$x|Out-Null
Write-Host "Hello, World!"
...
```

---

### Fragmentation (`-frag`, `-minfrag`, `-maxfrag`)

Only affects **level 5** (string fragments + runtime join).

* **Profiles**:

  * `profile=tight` → small chunks (≈6–10)
  * `profile=medium` → medium chunks (≈10–18)
  * `profile=loose` → larger chunks (≈14–28)
* **Or** tune with `-minfrag` / `-maxfrag`.

**Commands**

```bash
# Profile based
psobf -i sample.ps1 -o out.ps1 -level 5 -frag profile=loose -seed 9

# Fine control
psobf -i sample.ps1 -o out.ps1 -level 5 -minfrag 8 -maxfrag 16 -seed 9
```

**Output (snippet)**

```powershell
$fragments=@('Write-Host "Hello,',' World!"',"`n", '$answer = 42',"`n",'function Greet($','name){ Write-Host ("Hi, "+$name)}',"`n",'Greet "Ada"');
$script=$fragments -join ''; Invoke-Expression $script
```

---

## Profiles (light, balanced, heavy)

Presets are convenient starting points. Any explicit flag you pass **overrides** the preset.
Any flag you pass explicitly takes precedence over the profile.

* **light**

  ```
  -pipeline "iden,stringdict,numenc,frag"
  -frag profile=tight
  -seed 1337
  ```
* **balanced**

  ```
  -pipeline "iden,strenc,stringdict,numenc,fmt,cf,dead,frag"
  -strenc xor -strkey a1b2c3d4
  -stringdict 30 -deadcode 10 -fmt jitter -frag profile=medium
  -seed 424242
  ```
* **heavy**

  ```
  -pipeline "iden,strenc,stringdict,numenc,fmt,cf,dead,frag"
  -strenc rc4 -strkey 00112233445566778899aabbccddeeff
  -stringdict 50 -deadcode 25 -fmt jitter -frag profile=loose
  -seed 987654321
  ```

---

## Seeds, reproducibility & fuzzing

* `-seed N` → deterministic output for a given config.
* No `-seed` → cryptographically seeded randomness.
* `-fuzz N` → produce **N** variants (`out.ps1.v1.ps1`, `out.ps1.v2.ps1`, …), great for diversity testing.

**Example**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 -profile heavy -fuzz 3
```

---

## STDIN/STDOUT and `-noexec`

* **Pipe in / out**

  ```bash
  cat sample.ps1 | psobf -stdin -stdout -level 2 > out.ps1
  ```
* **Audit only (no execution wrapper)**

  ```bash
  psobf -i sample.ps1 -o payload.txt -level 4 -noexec
  # payload.txt contains just the artifact (e.g., base64/gzip) without Invoke-Expression
  ```

---

## EDR/AV recipes (offensive  practice)

> The aim is to **diversify** artifacts and reduce stable signatures for **research** in authorized environments.

1. **Dense packing + literal encryption**

```bash
psobf -i sample.ps1 -o out.ps1 -level 4 \
  -pipeline "iden,strenc,stringdict" -iden obf -strenc rc4 -strkey 0011223344556677 -stringdict 40 \
  -seed 20250827
```

2. **Max diversity (format + fragmentation + dead code)**

```bash
psobf -i sample.ps1 -o out.ps1 -level 5 \
  -pipeline "fmt,frag,dead" -fmt jitter -frag profile=loose -deadcode 15 \
  -fuzz 5
```

3. **Balanced CI-friendly build**

```bash
psobf -i sample.ps1 -o out.ps1 -level 3 -profile balanced -seed 777
```

4. **Reduce static IOCs (numbers + dictionary)**

```bash
psobf -i sample.ps1 -o out.ps1 -level 2 -pipeline "numenc,stringdict" -numenc -stringdict 35 -seed 9
```

5. **RC4 + tight fragmentation (showing combined layers)**

```bash
psobf -i sample.ps1 -o out.ps1 -level 5 -pipeline "strenc,frag" -strenc rc4 -strkey 0011223344556677 -frag profile=tight -seed 44
```

---

## Best practices & defensive notes

* Rotate **`-strkey`** and **`-seed`** per build.
* Prefer combining layers: `-strenc` + `-stringdict` + `-fmt jitter` + fragmentation.
* Use `-fuzz` to generate families of variants for detection testing.
* Keep a clean, benign baseline and verify functional equivalence under sandbox before and after transforms.
* If your script relies on delicate PS syntax (e.g., redirections), keep them inside quotes or disable `-numenc`.

---

## Architecture diagram

```
       ┌──────────────┐
       │  input.ps1   │
       └──────┬───────┘
              │ read (-i / -stdin)
              ▼
       ┌──────────────┐
       │ Pipeline     │  order you choose
       │ iden         │  rename vars/funcs
       │ strenc       │  XOR/RC4 literals
       │ stringdict   │  tokenize + rejoin
       │ numenc       │  numeric masking
       │ fmt          │  whitespace jitter
       │ cf           │  opaque/shuffle
       │ dead         │  harmless noise
       └──────┬───────┘
              │ mutated script
              ▼
       ┌──────────────┐
       │  Level 1..5  │  final packing
       └──────┬───────┘
              │ + Invoke-Expression (unless -noexec)
              ▼
       ┌──────────────┐
       │   out.ps1    │
       └──────────────┘
```


---

### Quick cheat-sheet

```bash
# Deterministic, simple
psobf -i sample.ps1 -o out.ps1 -level 2 -seed 123

# RC4 (correct invocation shape)
psobf -i sample.ps1 -o out.ps1 -level 4 -pipeline "strenc" -strenc rc4 -strkey 0011223344556677

# Tokenization + numeric masking
psobf -i sample.ps1 -o out.ps1 -level 3 -pipeline "stringdict,numenc" -stringdict 40 -numenc

# Heavy combo
psobf -i sample.ps1 -o out.ps1 -level 5 \
  -pipeline "iden,strenc,stringdict,numenc,fmt,cf,dead,frag" \
  -iden obf -strenc xor -strkey a1b2c3d4 -stringdict 35 -numenc \
  -fmt jitter -cf-opaque -deadcode 15 -frag profile=medium -seed 777

# Inspect artifact only (no Invoke-Expression)
psobf -i sample.ps1 -o payload.txt -level 4 -noexec
```

---

## Legal

This project is for **educational** and **authorized** testing only. You are solely responsible for your usage. The authors and contributors assume no liability for direct or indirect damages.