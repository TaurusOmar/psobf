package obfuscator

import (
    "encoding/base64"
    "encoding/hex"
    "fmt"
    mathrand "math/rand"
    "regexp"
    "sort"
    "strconv"
    "strings"
)


type IdentifierTransform struct{}

func (t *IdentifierTransform) Name() string { return "iden" }

func (t *IdentifierTransform) Apply(ps string, ctx *Ctx) (string, error) {
	mapping := map[string]string{}
	reservedPrefix := "$__"
	rename := func(name string) string {
		if strings.HasPrefix(name, reservedPrefix) {
			return name
		}
		if v, ok := mapping[name]; ok {
			return v
		}
		n := RandIdent(ctx.Rng, len(name))
		if strings.HasPrefix(name, "$") && !strings.HasPrefix(n, "$") {
			n = "$" + n
		}
		mapping[name] = n
		return n
	}
	funcNames := map[string]string{}
	for _, m := range reFuncHeader.FindAllStringSubmatch(ps, -1) {
		fn := m[1]
		if _, ok := funcNames[fn]; !ok {
			funcNames[fn] = RandIdent(ctx.Rng, len(fn))
		}
	}
	for _, m := range reFuncNoParam.FindAllStringSubmatch(ps, -1) {
		fn := m[1]
		if _, ok := funcNames[fn]; !ok {
			funcNames[fn] = RandIdent(ctx.Rng, len(fn))
		}
	}
	for orig, neo := range funcNames {
		re := regexpQuoteWord(orig)
		ps = re.ReplaceAllString(ps, neo)
	}
	ps = reVar.ReplaceAllStringFunc(ps, func(v string) string { return rename(v) })
	return ps, nil
}

type StringDictTransform struct {
	Percent int
}

func (t *StringDictTransform) Name() string { return "stringdict" }

func (t *StringDictTransform) Apply(ps string, ctx *Ctx) (string, error) {
	dq := reDQ.FindAllStringIndex(ps, -1)
	sq := reSQ.FindAllStringIndex(ps, -1)
	type span struct{ s, e int; dbl bool }
	var spans []span
	for _, p := range dq {
		spans = append(spans, span{p[0], p[1], true})
	}
	for _, p := range sq {
		spans = append(spans, span{p[0], p[1], false})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].s < spans[j].s })
	var tokens []string
	tokenMap := map[string]int{}
	buildTokens := func(s string) []int {
		var idxs []int
		minTok := 3
		for i := 0; i < len(s); {
			left := len(s) - i
			if left <= minTok {
				chunk := s[i:]
				idxs = append(idxs, addTok(chunk, &tokens, tokenMap))
				break
			}
			maxTok := 6
			size := ctx.Rng.Intn(maxTok-minTok+1) + minTok
			if size > left {
				size = left
			}
			chunk := s[i : i+size]
			idxs = append(idxs, addTok(chunk, &tokens, tokenMap))
			i += size
		}
		return idxs
	}
	var out strings.Builder
	cursor := 0
	var injected bool
	for _, sp := range spans {
		out.WriteString(ps[cursor:sp.s])
		raw := ps[sp.s:sp.e]
		lit := raw
		if strings.HasPrefix(lit, "\"") && strings.HasSuffix(lit, "\"") {
			lit = lit[1 : len(lit)-1]
		} else if strings.HasPrefix(lit, "'") && strings.HasSuffix(lit, "'") {
			lit = lit[1 : len(lit)-1]
		}
		if t.Percent > 0 && len(lit) >= 10 && ctx.Rng.Intn(100) < t.Percent {
			idxs := buildTokens(lit)
			var parts []string
			for _, id := range idxs {
				parts = append(parts, fmt.Sprintf("$D[%d]", id))
			}
			out.WriteString("(" + strings.Join(parts, "+") + ")")
			injected = true
		} else {
			out.WriteString(raw)
		}
		cursor = sp.e
	}
	out.WriteString(ps[cursor:])
	res := out.String()
	if injected && len(tokens) > 0 {
		var sb strings.Builder
		sb.WriteString("$D=@(")
		for i, tk := range tokens {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("'" + strings.ReplaceAll(tk, "'", "''") + "'")
		}
		sb.WriteString(");\n")
		res = sb.String() + res
	}
	return res, nil
}

func addTok(t string, arr *[]string, idx map[string]int) int {
	if id, ok := idx[t]; ok {
		return id
	}
	id := len(*arr)
	*arr = append(*arr, t)
	idx[t] = id
	return id
}

type StringEncryptTransform struct {
	Mode string
	Key  []byte
}

func (t *StringEncryptTransform) Name() string { return "strenc" }

func (t *StringEncryptTransform) Apply(ps string, ctx *Ctx) (string, error) {
	if t.Mode == "xor" {
		return encryptStrings(ps, ctx,
			func(b []byte) (enc string, helper string) {
				xb := make([]byte, len(b))
				for i := range b {
					xb[i] = b[i] ^ t.Key[i%len(t.Key)]
				}
				enc = base64.StdEncoding.EncodeToString(xb)
				return enc, ""
			},
			func(enc string) string {
				khex := strings.ToUpper(hex.EncodeToString(t.Key))
				return fmt.Sprintf(
					`(&{[byte[]]$k=0..(%d-1)|%%{[Convert]::ToByte('%s'.Substring($_*2,2),16)};`+
						`[byte[]]$b=[Convert]::FromBase64String('%s');`+
						`for($i=0;$i -lt $b.Length;$i++){$b[$i]=$b[$i] -bxor $k[$i%%$k.Length]};`+
						`[Text.Encoding]::UTF8.GetString($b)})`,
					len(t.Key), khex, enc)
			})
	}
	if t.Mode == "rc4" {
		fn := "__dec" + RandIdent(ctx.Rng, 6)
		if !ctx.Helpers["rc4"] {
			rc4Func := fmt.Sprintf(
				`function %s($k,[byte[]]$d){$s=0..255;$j=0;for($i=0;$i -lt 256;$i++){`+
					`$j=($j+$s[$i]+$k[$i%%$k.Length])%%256;$t=$s[$i];$s[$i]=$s[$j];$s[$j]=$t}`+
					`$i=0;$j=0;for($x=0;$x -lt $d.Length;$x++){`+
					`$i=($i+1)%%256;$j=($j+$s[$i])%%256;$t=$s[$i];$s[$i]=$s[$j];$s[$j]=$t;`+
					`$d[$x]=$d[$x] -bxor $s[($s[$i]+$s[$j])%%256]}`+
					`[Text.Encoding]::UTF8.GetString($d)}`, fn)
			ps = rc4Func + "\n" + ps
			ctx.Helpers["rc4"] = true
		}
		khex := strings.ToUpper(hex.EncodeToString(t.Key))
		return encryptStrings(ps, ctx,
			func(b []byte) (string, string) {
				s := make([]byte, 256)
				for i := 0; i < 256; i++ {
					s[i] = byte(i)
				}
				j := 0
				for i := 0; i < 256; i++ {
					j = (j + int(s[i]) + int(t.Key[i%len(t.Key)])) % 256
					s[i], s[j] = s[j], s[i]
				}
				i, j2 := 0, 0
				enc := make([]byte, len(b))
				for x := 0; x < len(b); x++ {
					i = (i + 1) % 256
					j2 = (j2 + int(s[i])) % 256
					s[i], s[j2] = s[j2], s[i]
					keystream := s[(int(s[i])+int(s[j2]))%256]
					enc[x] = b[x] ^ keystream
				}
				return base64.StdEncoding.EncodeToString(enc), ""
			},
			func(enc string) string {
				return fmt.Sprintf(
					`(%s ([byte[]](0..(%d-1)|%%{[Convert]::ToByte('%s'.Substring($_*2,2),16)})) ([Convert]::FromBase64String('%s')))`,
					fn, len(t.Key), khex, enc)
			})
	}
	return ps, nil
}

func encryptStrings(ps string, ctx *Ctx, encfn func([]byte) (string, string), psExpr func(string) string) (string, error) {
	idxs := reDQ.FindAllStringIndex(ps, -1)
	idxs2 := reSQ.FindAllStringIndex(ps, -1)
	type span struct{ s, e int }
	var spans []span
	for _, p := range idxs {
		spans = append(spans, span{p[0], p[1]})
	}
	for _, p := range idxs2 {
		spans = append(spans, span{p[0], p[1]})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].s < spans[j].s })
	var out strings.Builder
	cursor := 0
	for _, sp := range spans {
		out.WriteString(ps[cursor:sp.s])
		raw := ps[sp.s:sp.e]
		lit := raw[1 : len(raw)-1]
		enc, _ := encfn([]byte(lit))
		out.WriteString(psExpr(enc))
		cursor = sp.e
	}
	out.WriteString(ps[cursor:])
	return out.String(), nil
}

type NumberEncodeTransform struct{}

func (t *NumberEncodeTransform) Name() string { return "numenc" }

func shouldSkipNumberContext(s string, start, end int) bool {
	if end < len(s) && end+1 <= len(s) && strings.HasPrefix(s[end:], ">&") {
		return true
	}
	if start >= 2 {
		pfx := s[start-2 : start]
		if pfx == ">&" || pfx == "<&" {
			return true
		}
	}
	if (start > 0 && s[start-1] == '.') || (end < len(s) && s[end] == '.') {
		return true
	}
	return false
}

func encodeNumber(n int, r *mathrand.Rand) string {
	c := r.Intn(5) + 1
	b := r.Intn(0x7FFF)
	a := (n + c) ^ b
	return fmt.Sprintf("((0x%X -bxor 0x%X)-%d)", a, b, c)
}

func replaceNumsSafe(seg string, r *mathrand.Rand) string {
	var out strings.Builder
	last := 0
	idxs := reNum.FindAllStringIndex(seg, -1)
	for _, p := range idxs {
		start, end := p[0], p[1]
		out.WriteString(seg[last:start])
		if shouldSkipNumberContext(seg, start, end) {
			out.WriteString(seg[start:end])
		} else {
			n, _ := strconv.Atoi(seg[start:end])
			out.WriteString(encodeNumber(n, r))
		}
		last = end
	}
	out.WriteString(seg[last:])
	return out.String()
}

func (t *NumberEncodeTransform) Apply(ps string, ctx *Ctx) (string, error) {
	type seg struct{ s, e int }
	var spans []seg
	for _, p := range reDQ.FindAllStringIndex(ps, -1) {
		spans = append(spans, seg{p[0], p[1]})
	}
	for _, p := range reSQ.FindAllStringIndex(ps, -1) {
		spans = append(spans, seg{p[0], p[1]})
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].s < spans[j].s })

	var out strings.Builder
	cursor := 0
	for _, sp := range spans {
		before := ps[cursor:sp.s]
		before = replaceNumsSafe(before, ctx.Rng)
		out.WriteString(before)
		out.WriteString(ps[sp.s:sp.e])
		cursor = sp.e
	}
	rest := ps[cursor:]
	rest = replaceNumsSafe(rest, ctx.Rng)
	out.WriteString(rest)
	return out.String(), nil
}

type FormatJitterTransform struct{}

func (t *FormatJitterTransform) Name() string { return "fmt" }

func (t *FormatJitterTransform) Apply(ps string, ctx *Ctx) (string, error) {
	lines := strings.Split(ps, "\n")
	for i := range lines {
		if ctx.Rng.Intn(100) < 35 {
			lines[i] = strings.TrimSpace(lines[i])
		}
		if ctx.Rng.Intn(100) < 30 {
			lines[i] = " " + lines[i]
		}
		if ctx.Rng.Intn(100) < 30 {
			lines[i] = lines[i] + " "
		}
	}
	return strings.Join(lines, strings.Repeat("\n", 1+ctx.Rng.Intn(2))), nil
}

type CFOpaqueTransform struct{}

func (t *CFOpaqueTransform) Name() string { return "cf-opaque" }

func (t *CFOpaqueTransform) Apply(ps string, ctx *Ctx) (string, error) {
	return fmt.Sprintf("if(1 -eq 1){\n%s\n}", ps), nil
}

type CFShuffleTransform struct{}

func (t *CFShuffleTransform) Name() string { return "cf-shuffle" }

func (t *CFShuffleTransform) Apply(ps string, ctx *Ctx) (string, error) {
	type fb struct{ start, end int }
	var blocks []fb
	locs := reFuncNoParam.FindAllStringIndex(ps, -1)
	for _, st := range locs {
		start := st[0]
		end := findMatchingBrace(ps, start)
		if end > start {
			blocks = append(blocks, fb{start, end})
		}
	}
	if len(blocks) == 0 {
		return ps, nil
	}
	var out strings.Builder
	var mids []string
	cursor := 0
	for _, b := range blocks {
		out.WriteString(ps[cursor:b.start])
		mids = append(mids, ps[b.start:b.end])
		cursor = b.end
	}
	out.WriteString(ps[cursor:])
	RandPerm(ctx.Rng, mids)
	var buf strings.Builder
	buf.WriteString(out.String())
	for _, m := range mids {
		buf.WriteString("\n" + m + "\n")
	}
	return buf.String(), nil
}

func findMatchingBrace(s string, start int) int {
	i := strings.Index(s[start:], "{")
	if i < 0 {
		return -1
	}
	depth := 0
	for pos := start + i; pos < len(s); pos++ {
		switch s[pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return pos + 1
			}
		}
	}
	return -1
}

type DeadCodeTransform struct{ Prob int }

func (t *DeadCodeTransform) Name() string { return "deadcode" }

func (t *DeadCodeTransform) Apply(ps string, ctx *Ctx) (string, error) {
	if ctx.Rng.Intn(100) >= t.Prob {
		return ps, nil
	}
	fn := "__dummy" + RandIdent(ctx.Rng, 6)
	snippets := []string{
		fmt.Sprintf("function %s{ return }", fn),
		"for($i=0;$i -lt 0;$i++){Start-Sleep -Milliseconds 0}",
		"$x='canary';$y=$x+$x|Out-Null",
	}
	var out strings.Builder
	out.WriteString(ps)
	for _, s := range snippets {
		if ctx.Rng.Intn(100) < t.Prob {
			out.WriteString("\n" + s + "\n")
		}
	}
	return out.String(), nil
}

func regexpQuoteWord(word string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(word) + `\b`)
}
