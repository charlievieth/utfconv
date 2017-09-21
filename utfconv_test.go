package utfconv

import (
	"bytes"
	"reflect"
	"testing"
	"unicode/utf16"
	"unicode/utf8"
)

func TestConstants_UTF8(t *testing.T) {
	if runeSelf != utf8.RuneSelf {
		t.Errorf("runeSelf equals %d - utf8.RuneSelf equals %d", maxRune, utf8.MaxRune)
	}
	if maxRune != utf8.MaxRune {
		t.Errorf("maxRune equals %d - utf8.MaxRune equals %d", maxRune, utf8.MaxRune)
	}
	if runeError != utf8.RuneError {
		t.Errorf("runeError equals %d - utf8.RuneError equals %d", runeError, utf8.RuneError)
	}
	if n := utf8.RuneLen(runeError); n != runeErrorLen {
		t.Errorf("runeErrorLen equals %d want %d", runeErrorLen, n)
	}
}

var testStrings = []string{
	"\x00",
	"\x01",
	"\x7e",
	"\x7f",
	"\xc2\x80",
	"\xc2\x81",
	"\xc2\xbf",
	"\xc3\x80",
	"\xc3\x81",
	"\xc3\x88",
	"\xc3\x90",
	"\xc3\xa0",
	"\xc3\xb0",
	"\xc3\xb8",
	"\xc3\xbf",
	"\xc4\x80",
	"\xdf\xbf",
	"\xd0\x80",
	"\xe0\xa0\x80",
	"\xe0\xa0\x81",
	"\xe1\x80\x80",
	"\xed\x80\x80",
	"\xed\x9f\xbf", // last code point before surrogate half.
	"\xee\x80\x80", // first code point after surrogate half.
	"\xef\xbf\xbe",
	"\xef\xbf\xbf",
	"\xf0\x90\x80\x80",
	"\xf0\x90\x80\x81",
	"\xf1\x80\x80\x80",
	"\xf4\x8f\xbf\xbe",
	"\xf4\x8f\xbf\xbf",
	"\xef\xbf\xbd",

	// rune count tests
	"1,2,3,4",
	"evil 1: \xe2\x00",
	"evil 2: \xe2\x80",
	"evil 3: a\xe2\x80",

	"evil 1: \xe2\x00 suffix",
	"evil 2: \xe2\x80 suffix",
	"evil 3: a\xe2\x80 suffix",

	"abcd",
	"A ∆ B",
	"☺☻☹",
	"日a本b語ç日ð本Ê語þ日¥本¼語i日©",
	"日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©日a本b語ç日ð本Ê語þ日¥本¼語i日©",
	"\u007A",
	"\u6C34",
	"\uFEFF",
	"\U00010000",
	"\U0001D11E",
	"\U0010FFFD",
	string(rune(0xd7ff)),
	string(rune(0xd800)),
	string(rune(0xdc00)),
	string(rune(0xe000)),
	string(rune(0xdfff)),
	"\U0010FFFD" + string(rune(0xd800)) + string(rune(0xd800)) + "日a本b語ç" + "ABC",
	LargeTextSrcFile,
	LargeUnicodeSrcFile,
}

type RuneCountTest struct {
	in  string
	out int
}

var runecounttests = []RuneCountTest{
	{"abcd", 4},
	{"☺☻☹", 3},
	{"1,2,3,4", 7},
	{"\xe2\x00", 2},
	{"\xe2\x80", 2},
	{"a\xe2\x80", 3},
}

// expUTF16String should be used to test the results of decoding UTF16
func expUTF16String(s string) string {
	return string(utf16.Decode(utf16.Encode([]rune(s))))
}

func TestUTF8EncodedLen(t *testing.T) {
	for i, x := range testStrings {
		u := utf16.Encode([]rune(x))
		n := UTF8EncodedLen(u)
		exp := expUTF16String(x)
		if n != len(exp) {
			t.Errorf("UTF8EncodedLen (%d - %q) got %d want %d", i, exp, n, len(exp))
		}
	}
}

func TestUTF16ToBytes(t *testing.T) {
	for i, x := range testStrings {
		u := utf16.Encode([]rune(x))
		b := UTF16ToBytes(u)
		exp := expUTF16String(x)
		if !bytes.Equal(b, []byte(exp)) {
			if len(exp) < 128 {
				t.Errorf("UTF16ToString (%d - %q) got: %q want: %q", i, exp, string(b), exp)
			} else {
				t.Errorf("UTF16ToString failed test #%d", i)
			}
		}
	}
}

func TestUTF16ToString(t *testing.T) {
	for i, x := range testStrings {
		u := utf16.Encode([]rune(x))
		s := UTF16ToString(u)
		exp := expUTF16String(x)
		if s != exp {
			if len(exp) < 128 {
				t.Errorf("UTF16ToString (%d - %q) got: %q want: %q", i, exp, s, exp)
			} else {
				t.Errorf("UTF16ToString failed test #%d", i)
			}
		}
	}
}

func TestUTF16EncodedLen(t *testing.T) {
	for i, s := range testStrings {
		u := utf16.Encode([]rune(s))
		n := UTF16EncodedLen([]byte(s))
		if n != len(u) {
			t.Errorf("UTF16EncodedLen (%d - %q) got %d want %d", i, s, n, len(u))
		}
	}
}

func TestUTF16EncodedLenString(t *testing.T) {
	for i, s := range testStrings {
		u := utf16.Encode([]rune(s))
		n := UTF16EncodedLenString(s)
		if n != len(u) {
			if len(s) < 128 {
				t.Errorf("UTF16EncodedLenString (%d - %q) got %d want %d", i, s, n, len(u))
			} else {
				t.Errorf("UTF16EncodedLenString failed test #%d", i)
			}
		}
	}
}

func TestBytesToUTF16(t *testing.T) {
	// t.Skip("PENDING")
	for i, s := range testStrings {
		exp := utf16.Encode([]rune(s))
		u := BytesToUTF16([]byte(s))
		if !reflect.DeepEqual(u, exp) {
			if len(exp) < 128 {
				t.Errorf("BytesToUTF16 (%d - %q) got: %q want: %q", i, s,
					string(utf16.Decode(u)), string(utf16.Decode(exp)))
			} else {
				t.Errorf("BytesToUTF16 failed test #%d", i)
			}
		}
	}
}

func TestStringToUTF16(t *testing.T) {
	for i, s := range testStrings {
		exp := utf16.Encode([]rune(s))
		u := StringToUTF16(s)
		if !reflect.DeepEqual(u, exp) {
			if len(exp) < 128 {
				t.Errorf("StringToUTF16 (%d - %q) got: %q want: %q", i, s,
					string(utf16.Decode(u)), string(utf16.Decode(exp)))
			} else {
				t.Errorf("StringToUTF16 failed test #%d", i)
			}
		}
	}
}

var invalidSequenceTests = []string{
	"\xed\xa0\x80\x80", // surrogate min
	"\xed\xbf\xbf\x80", // surrogate max

	// xx
	"\x91\x80\x80\x80",

	// s1
	"\xC2\x7F\x80\x80",
	"\xC2\xC0\x80\x80",
	"\xDF\x7F\x80\x80",
	"\xDF\xC0\x80\x80",

	// s2
	"\xE0\x9F\xBF\x80",
	"\xE0\xA0\x7F\x80",
	"\xE0\xBF\xC0\x80",
	"\xE0\xC0\x80\x80",

	// s3
	"\xE1\x7F\xBF\x80",
	"\xE1\x80\x7F\x80",
	"\xE1\xBF\xC0\x80",
	"\xE1\xC0\x80\x80",

	//s4
	"\xED\x7F\xBF\x80",
	"\xED\x80\x7F\x80",
	"\xED\x9F\xC0\x80",
	"\xED\xA0\x80\x80",

	// s5
	"\xF0\x8F\xBF\xBF",
	"\xF0\x90\x7F\xBF",
	"\xF0\x90\x80\x7F",
	"\xF0\xBF\xBF\xC0",
	"\xF0\xBF\xC0\x80",
	"\xF0\xC0\x80\x80",

	// s6
	"\xF1\x7F\xBF\xBF",
	"\xF1\x80\x7F\xBF",
	"\xF1\x80\x80\x7F",
	"\xF1\xBF\xBF\xC0",
	"\xF1\xBF\xC0\x80",
	"\xF1\xC0\x80\x80",

	// s7
	"\xF4\x7F\xBF\xBF",
	"\xF4\x80\x7F\xBF",
	"\xF4\x80\x80\x7F",
	"\xF4\x8F\xBF\xC0",
	"\xF4\x8F\xC0\x80",
	"\xF4\x90\x80\x80",
}

func TestInvalidUTF8EncodedLen(t *testing.T) {
	for i, x := range invalidSequenceTests {
		u := utf16.Encode([]rune(x))
		n := UTF8EncodedLen(u)
		exp := expUTF16String(x)
		if n != len(exp) {
			t.Errorf("UTF8EncodedLen (%d - %q) got %d want %d", i, exp, n, len(exp))
		}
	}
}

func TestInvalidUTF16EncodedLen(t *testing.T) {
	for i, s := range invalidSequenceTests {
		u := utf16.Encode([]rune(s))
		n := UTF16EncodedLen([]byte(s))
		if n != len(u) {
			t.Errorf("UTF16EncodedLen (%d - %q) got %d want %d", i, s, n, len(u))
		}
	}
}

func TestInvalidUTF16EncodedLenString(t *testing.T) {
	for i, s := range invalidSequenceTests {
		u := utf16.Encode([]rune(s))
		n := UTF16EncodedLenString(s)
		if n != len(u) {
			t.Errorf("UTF16EncodedLenString (%d - %q) got %d want %d", i, s, n, len(u))
		}
	}
}

func TestInvalidBytesToUTF16(t *testing.T) {
	// WARN FIX THIS TEST
	t.Skip("BROKEN") // WARN

	for i, s := range invalidSequenceTests {
		exp := utf16.Encode([]rune(s))
		u := BytesToUTF16([]byte(s))
		if !reflect.DeepEqual(u, exp) {
			if len(exp) < 128 {
				t.Errorf("BytesToUTF16 (%d - %q) got: %q want: %q", i, s,
					string(utf16.Decode(u)), string(utf16.Decode(exp)))
			} else {
				t.Errorf("BytesToUTF16 failed test #%d", i)
			}
		}
	}
}

func TestInvalidStringToUTF16(t *testing.T) {
	// WARN FIX THIS TEST
	// t.Skip("Not working!")

	for i, s := range invalidSequenceTests {
		exp := utf16.Encode([]rune(s))
		u := StringToUTF16(s)
		if !reflect.DeepEqual(u, exp) {
			if len(exp) < 128 {
				t.Errorf("StringToUTF16 (%d - %q) got: %q want: %q", i, s,
					string(utf16.Decode(u)), string(utf16.Decode(exp)))
			} else {
				t.Errorf("StringToUTF16 failed test #%d", i)
			}
		}
	}
}

const TenASCIIChars = "0123456789"
const TenJapaneseChars = "日本語日本語日本語日"

var (
	LargeTextSrcFileUTF16    = utf16.Encode([]rune(LargeTextSrcFile))
	LargeUnicodeSrcFileUTF16 = utf16.Encode([]rune(LargeUnicodeSrcFile))
	TenASCIICharsUTF16       = utf16.Encode([]rune(TenASCIIChars))
	TenJapaneseCharsUTF16    = utf16.Encode([]rune(TenJapaneseChars))
)

// UTF8EncodedLen

func BenchmarkUTF8EncodedLen_TenASCIIChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF8EncodedLen(TenASCIICharsUTF16)
	}
}

func BenchmarkUTF8EncodedLen_Base_TenASCIIChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = len(string(utf16.Decode(TenASCIICharsUTF16)))
	}
}

func BenchmarkUTF8EncodedLen_TenJapaneseChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF8EncodedLen(TenJapaneseCharsUTF16)
	}
}

func BenchmarkUTF8EncodedLen_Base_TenJapaneseChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = len(string(utf16.Decode(TenJapaneseCharsUTF16)))
	}
}

func BenchmarkUTF8EncodedLen_LargeASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF8EncodedLen(LargeTextSrcFileUTF16)
	}
}

func BenchmarkUTF8EncodedLen_Base_LargeASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = len(string(utf16.Decode(LargeTextSrcFileUTF16)))
	}
}

func BenchmarkUTF8EncodedLen_LargeUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF8EncodedLen(LargeUnicodeSrcFileUTF16)
	}
}

func BenchmarkUTF8EncodedLen_Base_LargeUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = len(string(utf16.Decode(LargeUnicodeSrcFileUTF16)))
	}
}

// UTF16EncodedLen

func BenchmarkUTF16EncodedLen_TenASCIIChars(b *testing.B) {
	p := []byte(TenASCIIChars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLen(p)
	}
}

func BenchmarkUTF16EncodedLen_Base_TenASCIIChars(b *testing.B) {
	s := []byte(TenASCIIChars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(string(s))))
	}
}

func BenchmarkUTF16EncodedLen_TenJapaneseChars(b *testing.B) {
	p := []byte(TenJapaneseChars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLen(p)
	}
}

func BenchmarkUTF16EncodedLen_Base_TenJapaneseChars(b *testing.B) {
	s := []byte(TenJapaneseChars)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(string(s))))
	}
}

func BenchmarkUTF16EncodedLen_LargeASCII(b *testing.B) {
	p := []byte(LargeTextSrcFile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLen(p)
	}
}

// func BenchmarkUTF16EncodedLen_Base_LargeASCII(b *testing.B) {
// 	s := []byte(LargeTextSrcFile)
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(string(s))))
// 	}
// }

func BenchmarkUTF16EncodedLen_LargeUnicode(b *testing.B) {
	p := []byte(LargeUnicodeSrcFile)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLen(p)
	}
}

// func BenchmarkUTF16EncodedLen_Base_LargeUnicode(b *testing.B) {
// 	s := []byte(LargeUnicodeSrcFile)
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(string(s))))
// 	}
// }

// UTF16EncodedLenString

func BenchmarkUTF16EncodedLenString_TenASCIIChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLenString(TenASCIIChars)
	}
}

func BenchmarkUTF16EncodedLenString_Base_TenASCIIChars(b *testing.B) {
	s := TenASCIIChars
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(s)))
	}
}

func BenchmarkUTF16EncodedLenString_TenJapaneseChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16EncodedLenString(TenJapaneseChars)
	}
}

func BenchmarkUTF16EncodedLenString_Base_TenJapaneseChars(b *testing.B) {
	s := TenJapaneseChars
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(s)))
	}
}

// func BenchmarkUTF16EncodedLenString_LargeASCII(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = UTF16EncodedLenString(LargeTextSrcFile)
// 	}
// }

// func BenchmarkUTF16EncodedLenString_Base_LargeASCII(b *testing.B) {
// 	s := LargeTextSrcFile
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(s)))
// 	}
// }

// func BenchmarkUTF16EncodedLenString_LargeUnicode(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = UTF16EncodedLenString(LargeUnicodeSrcFile)
// 	}
// }

// func BenchmarkUTF16EncodedLenString_Base_LargeUnicode(b *testing.B) {
// 	s := LargeUnicodeSrcFile
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(s)))
// 	}
// }

// UTF16ToBytes

func BenchmarkUTF16ToBytes_TenASCIIChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToBytes(TenASCIICharsUTF16)
	}
}

func BenchmarkUTF16ToBytes_TenJapaneseChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToBytes(TenJapaneseCharsUTF16)
	}
}

func BenchmarkUTF16ToBytes_LargeASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToBytes(LargeTextSrcFileUTF16)
	}
}

func BenchmarkUTF16ToBytes_LargeUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToBytes(LargeUnicodeSrcFileUTF16)
	}
}

// UTF16ToString

func BenchmarkUTF16ToString_LargeASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToString(LargeTextSrcFileUTF16)
	}
}

func BenchmarkUTF16ToString_Base_LargeASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = string(utf16.Decode(LargeTextSrcFileUTF16))
	}
}

func BenchmarkUTF16ToString_LargeUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UTF16ToString(LargeUnicodeSrcFileUTF16)
	}
}

func BenchmarkUTF16ToString_Base_LargeUnicode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = string(utf16.Decode(LargeUnicodeSrcFileUTF16))
	}
}

// BytesToUTF16

func BenchmarkBytesToUTF16_TenASCIIChars(b *testing.B) {
	s := []byte(TenASCIIChars)
	for i := 0; i < b.N; i++ {
		_ = BytesToUTF16(s)
	}
}

func BenchmarkBytesToUTF16_Base_TenASCIIChars(b *testing.B) {
	s := []byte(TenASCIIChars)
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(string(s))))
	}
}

func BenchmarkBytesToUTF16_TenJapaneseChars(b *testing.B) {
	s := []byte(TenJapaneseChars)
	for i := 0; i < b.N; i++ {
		_ = BytesToUTF16(s)
	}
}

func BenchmarkBytesToUTF16_Base_TenJapaneseChars(b *testing.B) {
	s := []byte(TenJapaneseChars)
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(string(s))))
	}
}

// func BenchmarkBytesToUTF16_LargeASCII(b *testing.B) {
// 	s := []byte(LargeTextSrcFile)
// 	for i := 0; i < b.N; i++ {
// 		_ = BytesToUTF16(s)
// 	}
// }

// func BenchmarkBytesToUTF16_Base_LargeASCII(b *testing.B) {
// 	s := []byte(LargeTextSrcFile)
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(string(s))))
// 	}
// }

// func BenchmarkBytesToUTF16_LargeUnicode(b *testing.B) {
// 	s := []byte(LargeUnicodeSrcFile)
// 	for i := 0; i < b.N; i++ {
// 		_ = BytesToUTF16(s)
// 	}
// }

// func BenchmarkBytesToUTF16_Base_LargeUnicode(b *testing.B) {
// 	s := []byte(LargeUnicodeSrcFile)
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(string(s))))
// 	}
// }

// StringToUTF16

func BenchmarkStringToUTF16_TenASCIIChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToUTF16(TenASCIIChars)
	}
}

func BenchmarkStringToUTF16_Base_TenASCIIChars(b *testing.B) {
	s := TenASCIIChars
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(s)))
	}
}

func BenchmarkStringToUTF16_TenJapaneseChars(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToUTF16(TenJapaneseChars)
	}
}

func BenchmarkStringToUTF16_Base_TenJapaneseChars(b *testing.B) {
	s := TenJapaneseChars
	for i := 0; i < b.N; i++ {
		_ = len(utf16.Encode([]rune(s)))
	}
}

// func BenchmarkStringToUTF16_LargeASCII(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = StringToUTF16(LargeTextSrcFile)
// 	}
// }

// func BenchmarkStringToUTF16_Base_LargeASCII(b *testing.B) {
// 	s := LargeTextSrcFile
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(s)))
// 	}
// }

// func BenchmarkStringToUTF16_LargeUnicode(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = StringToUTF16(LargeUnicodeSrcFile)
// 	}
// }

// func BenchmarkStringToUTF16_Base_LargeUnicode(b *testing.B) {
// 	s := LargeUnicodeSrcFile
// 	for i := 0; i < b.N; i++ {
// 		_ = len(utf16.Encode([]rune(s)))
// 	}
// }
