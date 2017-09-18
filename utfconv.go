package utfconv

// UTF16 constants

const (
	runeError    = '\uFFFD' // Unicode replacement character
	runeErrorLen = 3        // Unicode replacement character
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

// RuneLen returns the number of bytes required to encode the rune.
// It returns -1 if the rune is not a valid value to encode in UTF-8.
func runeLenUTF8(r uint32) int {
	switch {
	case r <= rune1Max:
		return 1
	case r <= rune2Max:
		return 2
	case surrogateMin <= r && r <= surrogateMax:
		return runeErrorLen
	case r <= rune3Max:
		return 3
	case r <= maxRune:
		return 4
	}
	return runeErrorLen
}

// UTF8EncodedLen returns the number of bytes required to encode UTF16 slice s
// as UTF8.
func UTF8EncodedLen(s []uint16) int {
	ns := len(s)
	n := 0
	for i := 0; i < ns; i++ {
		switch r := uint32(s[i]); {
		case r < runeSelf:
			// ASCII fast path
			n++
		case r < surr1, surr3 <= r:
			// normal rune
			switch {
			case r <= rune2Max:
				n += 2
			case r <= rune3Max:
				n += 3
			case r <= maxRune:
				n += 4
			default:
				n += runeErrorLen
			}
		case surr1 <= r && r < surr2 && i+1 < ns &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence
			n += runeLenUTF8((r-surr1)<<10 | (uint32(s[i+1]) - surr2) + surrSelf)
			i++
		default:
			// invalid surrogate sequence
			n += runeErrorLen
		}
	}
	return n
}

// encoderune writes into p (which must be large enough) the UTF-8 encoding of the rune.
// It returns the number of bytes written.
func encoderune(p []byte, r uint32) int {
	const (
		t1 = 0x00 // 0000 0000
		tx = 0x80 // 1000 0000
		t2 = 0xC0 // 1100 0000
		t3 = 0xE0 // 1110 0000
		t4 = 0xF0 // 1111 0000

		maskx = 0x3F // 0011 1111
	)
	// Negative values are erroneous. Making it unsigned addresses the problem.
	switch {
	case r <= rune1Max:
		p[0] = byte(r)
		return 1
	case r <= rune2Max:
		_ = p[1] // eliminate bounds checks
		p[0] = t2 | byte(r>>6)
		p[1] = tx | byte(r)&maskx
		return 2
	case r > maxRune, surrogateMin <= r && r <= surrogateMax:
		r = runeError
		fallthrough
	case r <= rune3Max:
		_ = p[2] // eliminate bounds checks
		p[0] = t3 | byte(r>>12)
		p[1] = tx | byte(r>>6)&maskx
		p[2] = tx | byte(r)&maskx
		return 3
	default:
		_ = p[3] // eliminate bounds checks
		p[0] = t4 | byte(r>>18)
		p[1] = tx | byte(r>>12)&maskx
		p[2] = tx | byte(r>>6)&maskx
		p[3] = tx | byte(r)&maskx
		return 4
	}
}

func UTF16ToBytes(s []uint16) []byte {
	const (
		t1 = 0x00 // 0000 0000
		tx = 0x80 // 1000 0000
		t2 = 0xC0 // 1100 0000
		t3 = 0xE0 // 1110 0000
		t4 = 0xF0 // 1111 0000

		maskx = 0x3F // 0011 1111
	)
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
			// n += encoderune(a[n:], r)
			p := a[n:]
			switch {
			case r <= rune2Max:
				_ = p[1] // eliminate bounds checks
				p[0] = t2 | byte(r>>6)
				p[1] = tx | byte(r)&maskx
				n += 2
			case r <= rune3Max:
				_ = p[2] // eliminate bounds checks
				p[0] = t3 | byte(r>>12)
				p[1] = tx | byte(r>>6)&maskx
				p[2] = tx | byte(r)&maskx
				n += 3
			default:
				_ = p[3] // eliminate bounds checks
				p[0] = t4 | byte(r>>18)
				p[1] = tx | byte(r>>12)&maskx
				p[2] = tx | byte(r>>6)&maskx
				p[3] = tx | byte(r)&maskx
				n += 4
			}
		case surr1 <= r && r < surr2 && i+1 < ns &&
			surr2 <= s[i+1] && s[i+1] < surr3:
			// valid surrogate sequence

			// n += encoderune(a[n:], (r-surr1)<<10|(uint32(s[i+1])-surr2)+surrSelf)
			// i++

			r = (r-surr1)<<10 | (uint32(s[i+1]) - surr2) + surrSelf
			p := a[n:]
			switch {
			case r <= rune2Max:
				_ = p[1] // eliminate bounds checks
				p[0] = t2 | byte(r>>6)
				p[1] = tx | byte(r)&maskx
				n += 2
			case r <= rune3Max:
				_ = p[2] // eliminate bounds checks
				p[0] = t3 | byte(r>>12)
				p[1] = tx | byte(r>>6)&maskx
				p[2] = tx | byte(r)&maskx
				n += 3
			default:
				_ = p[3] // eliminate bounds checks
				p[0] = t4 | byte(r>>18)
				p[1] = tx | byte(r>>12)&maskx
				p[2] = tx | byte(r>>6)&maskx
				p[3] = tx | byte(r)&maskx
				n += 4
			}
			i++
		default:
			// invalid surrogate sequence
			n += copy(a[n:], []byte{239, 191, 189}) // replacementChar bytes
		}
	}
	return a
}

// func UTF16ToBytes(s []uint16) []byte {
// 	na := UTF8EncodedLen(s)
// 	a := make([]byte, na)
// 	ns := len(s)

// 	// ASCII fast path
// 	if na == ns {
// 		for i, c := range s {
// 			a[i] = byte(c)
// 		}
// 		return a
// 	}

// 	n := 0
// 	for i := 0; i < ns; i++ {
// 		switch r := rune(s[i]); {
// 		case r < runeSelf:
// 			// ASCII fast path
// 			a[n] = byte(r)
// 			n++
// 		case r < surr1, surr3 <= r:
// 			// normal rune
// 			n += utf8.EncodeRune(a[n:], r)
// 			// n += encoderune(a[n:], r)
// 		case surr1 <= r && r < surr2 && i+1 < ns &&
// 			surr2 <= s[i+1] && s[i+1] < surr3:
// 			// valid surrogate sequence
// 			n += utf8.EncodeRune(a[n:], (r-surr1)<<10|(rune(s[i+1])-surr2)+surrSelf)
// 			// n += encoderune(a[n:], (r-surr1)<<10|(rune(s[i+1])-surr2)+surrSelf)
// 			i++
// 		default:
// 			// invalid surrogate sequence
// 			n += copy(a[n:], []byte{239, 191, 189}) // replacementChar bytes
// 		}
// 	}
// 	return a
// }

func UTF16ToString(s []uint16) string {
	return string(UTF16ToBytes(s))
}

// This is the decoderune function found at runtime/utf8.go modified to take
// a byte slice as its argument.
//
// Original comment (note it refers to strings and here we use a byte slice):
//
//    decoderune returns the non-ASCII rune at the start of
//    s[k:] and the index after the rune in s.
//
//    decoderune assumes that caller has checked that
//    the to be decoded rune is a non-ASCII rune.
//
//    If the string appears to be incomplete or decoding problems
//    are encountered (runeerror, k + 1) is returned to ensure
//    progress when decoderune is used to iterate over a string.
//
func decoderune(s []byte, k int) (r rune, pos int) {
	const (
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

	pos = k
	s = s[k:]

	switch {
	case t2 <= s[0] && s[0] < t3:
		// 0080-07FF two byte sequence
		if len(s) > 1 && (locb <= s[1] && s[1] <= hicb) {
			r = rune(s[0]&mask2)<<6 | rune(s[1]&maskx)
			pos += 2
			if rune1Max < r {
				return
			}
		}
	case t3 <= s[0] && s[0] < t4:
		// 0800-FFFF three byte sequence
		if len(s) > 2 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) {
			r = rune(s[0]&mask3)<<12 | rune(s[1]&maskx)<<6 | rune(s[2]&maskx)
			pos += 3
			if rune2Max < r && !(surrogateMin <= r && r <= surrogateMax) {
				return
			}
		}
	case t4 <= s[0] && s[0] < t5:
		// 10000-1FFFFF four byte sequence
		if len(s) > 3 && (locb <= s[1] && s[1] <= hicb) && (locb <= s[2] && s[2] <= hicb) &&
			(locb <= s[3] && s[3] <= hicb) {

			r = rune(s[0]&mask4)<<18 | rune(s[1]&maskx)<<12 |
				rune(s[2]&maskx)<<6 | rune(s[3]&maskx)
			pos += 4
			if rune3Max < r && r <= maxRune {
				return
			}
		}
	}

	return runeError, k + 1
}

func UTF16EncodedLen(s []byte) int {
	n := 0
	ns := len(s)
	for i := 0; i < ns; n++ {
		if s[i] < runeSelf {
			i++
		} else {
			v, idx := decoderune(s, i)
			if surrSelf <= v && v <= maxRune {
				n++
			}
			i = idx
		}
	}
	return n
}

func UTF16EncodedLenString(s string) int {
	n := 0
	for _, v := range s {
		if surrSelf <= v && v <= maxRune {
			n++
		}
		n++
	}
	return n
}

func BytesToUTF16(s []byte) []uint16 {
	ns := len(s)
	na := UTF16EncodedLen(s)
	a := make([]uint16, na)
	if na == ns {
		for i, c := range s {
			a[i] = uint16(c)
		}
		return a
	}
	n := 0
	for i := 0; i < ns; n++ {
		if s[i] < runeSelf {
			a[n] = uint16(s[i])
			i++
		} else {
			r, idx := decoderune(s, i)
			i = idx
			switch {
			case 0 <= r && r < surr1, surr3 <= r && r < surrSelf:
				// normal rune
				a[n] = uint16(r)
			case surrSelf <= r && r <= maxRune:
				// needs surrogate sequence
				r -= surrSelf
				_ = a[n+1] // eliminate bounds checks
				a[n] = uint16(surr1 + (r>>10)&0x3ff)
				a[n+1] = uint16(surr2 + r&0x3ff)
				n++
			default:
				a[n] = runeError
			}
		}
	}
	return a[:n]
}

func StringToUTF16(s string) []uint16 {
	ns := len(s)
	na := UTF16EncodedLenString(s)
	a := make([]uint16, na)
	if na == ns {
		for i, c := range s {
			a[i] = uint16(c)
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
