package utfconv

// UTF16 constants

const (
	runeError    = '\uFFFD' // Unicode replacement character
	runeErrorLen = 3
)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd800
	surr2 = 0xdc00
	surr3 = 0xe000

	surrSelf = 0x10000
)

// UTF8 constants

// Numbers fundamental to the encoding.
const (
	runeSelf = 0x80         // characters below Runeself are represented as themselves in a single byte.
	maxRune  = '\U0010FFFF' // Maximum valid Unicode code point.
)

// Code points in the surrogate range are not valid for UTF-8.
const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
)

const (
	rune1Max = 1<<7 - 1  // 0x7F
	rune2Max = 1<<11 - 1 // 0x7FF
	rune3Max = 1<<16 - 1 // 0xFFFF
)

const (
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	// The default lowest and highest continuation byte.
	locb = 0x80 // 1000 0000
	hicb = 0xBF // 1011 1111
)

// UTF8EncodedLen returns the number of bytes required to encode UTF16 slice s
// as UTF8.
func UTF8EncodedLen(s []uint16) int {
	ns := len(s)
	n := ns
	i := 0
	for ; i < ns; i++ {
		if s[i] >= runeSelf {
			goto Loop
		}
	}
	return ns

Loop:
	for ; i < ns; i++ {
		if s[i] < runeSelf {
			continue Loop
		}
		switch r := rune(s[i]); {
		case r < surr1, surr3 <= r:
			// normal rune
			if r <= rune2Max {
				n += 1
			} else {
				n += 2
			}
			continue Loop
		case surr1 <= r && r < surr2 && i+1 < ns &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			n += 2
			i++
			continue Loop
		default:
			// invalid surrogate sequence
			n += runeErrorLen - 1
			continue Loop
		}
	}
	return n
}

func UTF16ToBytes(s []uint16) []byte {
	na := UTF8EncodedLen(s)
	a := make([]byte, na)
	ns := len(s)

	// ASCII fast path
	if na == ns {
		for i, c := range s {
			a[i] = byte(c)
		}
		return a
	}

	n := 0
	for i := 0; i < ns; i++ {
		switch r := uint32(s[i]); {
		case r < runeSelf:
			// ASCII fast path
			a[n] = byte(r)
			n++
		case r < surr1, surr3 <= r:
			// normal rune
			if r <= rune2Max {
				_ = a[n+1] // eliminate bounds checks
				a[n+0] = t2 | byte(r>>6)
				a[n+1] = tx | byte(r)&maskx
				n += 2
			} else {
				_ = a[n+2] // eliminate bounds checks
				a[n+0] = t3 | byte(r>>12)
				a[n+1] = tx | byte(r>>6)&maskx
				a[n+2] = tx | byte(r)&maskx
				n += 3
			}
		case surr1 <= r && r < surr2 && i+1 < ns &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence

			r = (r-surr1)<<10 | (uint32(s[i+1]) - surr2) + surrSelf
			i++

			_ = a[n+3] // eliminate bounds checks
			a[n+0] = t4 | byte(r>>18)
			a[n+1] = tx | byte(r>>12)&maskx
			a[n+2] = tx | byte(r>>6)&maskx
			a[n+3] = tx | byte(r)&maskx
			n += 4
		default:
			// invalid surrogate sequence
			n += copy(a[n:], "\uFFFD") // replacement char
		}
	}
	return a
}

func UTF16ToString(s []uint16) string {
	var buf [32]byte
	var a []byte

	na := UTF8EncodedLen(s)
	if na <= len(buf) {
		a = buf[0:]
	} else {
		a = make([]byte, na)
	}

	// ASCII fast path
	ns := len(s)
	if na == ns {
		for i := 0; i < ns; i++ {
			a[i] = byte(s[i])
		}
		return string(a[:na])
	}

	n := 0
	for i := 0; i < ns; i++ {
		switch r := uint32(s[i]); {
		case r < runeSelf:
			// ASCII fast path
			a[n] = byte(r)
			n++
		case r < surr1, surr3 <= r:
			// normal rune
			if r <= rune2Max {
				_ = a[n+1] // eliminate bounds checks
				a[n+0] = t2 | byte(r>>6)
				a[n+1] = tx | byte(r)&maskx
				n += 2
			} else {
				_ = a[n+2] // eliminate bounds checks
				a[n+0] = t3 | byte(r>>12)
				a[n+1] = tx | byte(r>>6)&maskx
				a[n+2] = tx | byte(r)&maskx
				n += 3
			}
		case surr1 <= r && r < surr2 && i+1 < ns &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			r = (r-surr1)<<10 | (uint32(s[i+1]) - surr2) + surrSelf
			i++
			_ = a[n+3] // eliminate bounds checks
			a[n+0] = t4 | byte(r>>18)
			a[n+1] = tx | byte(r>>12)&maskx
			a[n+2] = tx | byte(r>>6)&maskx
			a[n+3] = tx | byte(r)&maskx
			n += 4
		default:
			// invalid surrogate sequence
			n += copy(a[n:], "\uFFFD") // replacement char
		}
	}

	return string(a[:na])
}

func UTF16EncodedLen(p []byte) int {
	// Regarding the inner switch statements, there are equivalent expressions
	// that are simpler, but when benchmarked with Go 1.9 were slower. Even the
	// minimal check of 'if surrSelf <= r {...' was ~28% slower than the switch
	// statement.

	n := 0
Loop:
	for i := 0; i < len(p); n++ {
		if p[i] < runeSelf {
			i++
			continue Loop
		}
		switch s := p[i:]; {
		case t2 <= s[0] && s[0] < t3:
			if len(s) > 1 && (locb <= s[1] && s[1] <= hicb) {
				r := rune(s[0]&mask2)<<6 | rune(s[1]&maskx)
				if rune1Max < r {
					i += 2
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						// ok
					case surrSelf <= r && r <= maxRune:
						n++
					}
					continue Loop
				}
			}
		case t3 <= s[0] && s[0] < t4:
			if len(s) > 2 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) {
				r := rune(s[0]&mask3)<<12 | rune(s[1]&maskx)<<6
				if rune2Max < r && !(surrogateMin <= r && r <= surrogateMax) {
					i += 3
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						// ok
					case surrSelf <= r && r <= maxRune:
						n++
					}
					continue Loop
				}
			}
		case t4 <= s[0] && s[0] < t5:
			if len(s) > 3 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] &&
				s[2] <= hicb) && (locb <= s[3] && s[3] <= hicb) {
				r := rune(s[0]&mask4)<<18 | rune(s[1]&maskx)<<12 | rune(s[2]&maskx)<<6
				if rune3Max < r && r <= maxRune {
					i += 4
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						// ok
					case surrSelf <= r && r <= maxRune:
						n++
					}
					continue Loop
				}
			}
		}
		i++
	}

	return n
}

func BytesToUTF16(p []byte) []uint16 {
	na := UTF16EncodedLen(p)
	a := make([]uint16, na)
	n := 0
Loop:
	for i := 0; i < len(p); n++ {
		if p[i] < runeSelf {
			a[n] = uint16(p[i])
			i++
			continue Loop
		}
		switch s := p[i:]; {
		case t2 <= s[0] && s[0] < t3:
			if len(s) > 1 && (locb <= s[1] && s[1] <= hicb) {
				r := rune(s[0]&mask2)<<6 | rune(s[1]&maskx)
				if rune1Max < r {
					i += 2
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						a[n] = uint16(r)
					case surrSelf <= r && r <= maxRune:
						r -= surrSelf
						a[n] = uint16(surr1 + (r>>10)&0x3ff)
						a[n+1] = uint16(surr2 + r&0x3ff)
						n++
					default:
						a[n] = uint16(runeError)
					}
					continue Loop
				}
			}
		case t3 <= s[0] && s[0] < t4:
			if len(s) > 2 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) {
				r := rune(s[0]&mask3)<<12 | rune(s[1]&maskx)<<6 | rune(s[2]&maskx)
				if rune2Max < r && !(surrogateMin <= r && r <= surrogateMax) {
					i += 3
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						a[n] = uint16(r)
					case surrSelf <= r && r <= maxRune:
						r -= surrSelf
						a[n] = uint16(surr1 + (r>>10)&0x3ff)
						a[n+1] = uint16(surr2 + r&0x3ff)
						n++
					default:
						a[n] = uint16(runeError)
					}
					continue Loop
				}
			}
		case t4 <= s[0] && s[0] < t5:
			if len(s) > 3 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] &&
				s[2] <= hicb) && (locb <= s[3] && s[3] <= hicb) {
				r := rune(s[0]&mask4)<<18 | rune(s[1]&maskx)<<12 | rune(s[2]&maskx)<<6 | rune(s[3]&maskx)
				if rune3Max < r && r <= maxRune {
					i += 4
					switch {
					case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
						a[n] = uint16(r)
					case surrSelf <= r && r <= maxRune:
						r -= surrSelf
						a[n] = uint16(surr1 + (r>>10)&0x3ff)
						a[n+1] = uint16(surr2 + r&0x3ff)
						n++
					default:
						a[n] = uint16(runeError)
					}
					continue Loop
				}
			}
		}
		a[n] = uint16(runeError)
		i++
	}
	return a
}

func UTF16EncodedLenString(p string) int {
	n := 0
Loop:
	for i := 0; i < len(p); i++ {
		n++
		if p[i] < runeSelf {
			continue Loop
		}
		switch s := p[i:]; {
		case t2 <= s[0] && s[0] < t3:
			if len(s) > 1 && (locb <= s[1] && s[1] <= hicb) {
				r := rune(s[0]&mask2)<<6 | rune(s[1]&maskx)
				if rune1Max < r {
					i += 1
					continue Loop
				}
			}
		case t3 <= s[0] && s[0] < t4:
			if len(s) > 2 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) {
				r := rune(s[0]&mask3)<<12 | rune(s[1]&maskx)<<6
				if rune2Max < r && !(surrogateMin <= r && r <= surrogateMax) {
					i += 2
					continue Loop
				}
			}
		case t4 <= s[0] && s[0] < t5:
			if len(s) > 3 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] &&
				s[2] <= hicb) && (locb <= s[3] && s[3] <= hicb) {
				r := rune(s[0]&mask4)<<18 | rune(s[1]&maskx)<<12 | rune(s[2]&maskx)<<6
				if rune3Max < r && r <= maxRune {
					i += 3
					if surrSelf <= r {
						n++
					}
					continue Loop
				}
			}
		}
	}
	return n
}

func encodedLenString(p string) (int, bool) {
	n := 0
	i := 0
	for ; i < len(p); i++ {
		if p[i] >= runeSelf {
			n = i
			goto Loop
		}
	}
	return i, true

Loop:
	for ; i < len(p); i++ {
		n++
		if p[i] < runeSelf {
			continue Loop
		}
		switch s := p[i:]; {
		case t2 <= s[0] && s[0] < t3:
			if len(s) > 1 && (locb <= s[1] && s[1] <= hicb) {
				r := rune(s[0]&mask2)<<6 | rune(s[1]&maskx)
				if rune1Max < r {
					i += 1
					continue Loop
				}
			}
		case t3 <= s[0] && s[0] < t4:
			if len(s) > 2 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) {
				r := rune(s[0]&mask3)<<12 | rune(s[1]&maskx)<<6
				if rune2Max < r && !(surrogateMin <= r && r <= surrogateMax) {
					i += 2
					continue Loop
				}
			}
		case t4 <= s[0] && s[0] < t5:
			if len(s) > 3 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] &&
				s[2] <= hicb) && (locb <= s[3] && s[3] <= hicb) {

				r := rune(s[0]&mask4)<<18 | rune(s[1]&maskx)<<12 | rune(s[2]&maskx)<<6
				if rune3Max < r && r <= maxRune {
					i += 3
					if surrSelf <= r {
						n++
					}
					continue Loop
				}
			}
		}
	}

	return n, false
}

func StringToUTF16(s string) []uint16 {
	na, ok := encodedLenString(s)
	a := make([]uint16, na)
	if ok && na == len(s) {
		for i := 0; i < len(s); i++ {
			a[i] = uint16(s[i])
		}
		return a
	}
	n := 0
	for _, r := range s {
		switch {
		case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
			// normal rune
			a[n] = uint16(r)
			n++
		case surrSelf <= r && r <= maxRune:
			// needs surrogate sequence
			r -= surrSelf
			_ = a[n+1] // eliminate bounds checks
			a[n] = uint16(surr1 + (r>>10)&0x3ff)
			a[n+1] = uint16(surr2 + r&0x3ff)
			n += 2
		default:
			a[n] = runeError
			n++
		}
	}
	return a[:n]
}
