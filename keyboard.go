// Copyright (c) 2019, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import "strconv"

// See these two pages:
// https://www.win.tue.nl/~aeb/linux/kbd/scancodes-1.html
// https://www.millisecond.com/support/docs/v5/html/language/scancodes.htm

var asciiCodes = [128]uint8{
	'1': 2, '2': 3, '3': 4, '4': 5, '5': 6, '6': 7, '7': 8, '8': 9, '9': 10, '0': 11, '-': 12, '=': 13,
	'\t': 15, 'q': 16, 'w': 17, 'e': 18, 'r': 19, 't': 20, 'y': 21, 'u': 22, 'i': 23, 'o': 24, 'p': 25, '[': 26, ']': 27,
	'a': 30, 's': 31, 'd': 32, 'f': 33, 'g': 34, 'h': 35, 'j': 36, 'k': 37, 'l': 38, ';': 39, '\'': 40, '`': 41, '\\': 43,
	'z': 44, 'x': 45, 'c': 46, 'v': 47, 'b': 48, 'n': 49, 'm': 50, ',': 51, '.': 52, '/': 53, ' ': 57,
}

// ascii turns a string into a press-down keycode sequence.
func ascii(s string) (cs []uint8) {
	for _, r := range s {
		if r >= 128 {
			fatalf("non-ascii: %q", s)
		}
		code := asciiCodes[r]
		if code == 0 {
			fatalf("no code for %q", r)
		}
		cs = append(cs, code)
	}
	return cs
}

// winRun is the press-down keycode sequence for WIN+r.
var winRun = []uint8{0xe0, 0x5b, asciiCodes['r']}

// winRun is the press-down keycode sequence for ENTER.
var enter = []uint8{0x1c}

// alt adds the ALT pres-down keycode to the beginning of a sequence.
func alt(seq []uint8) []uint8 {
	return append([]uint8{0x38}, seq...)
}

// codes returns the hex strings required to perform a number of press-down
// keycode sequences. Each press-down sequence is followed by its release
// keycode sequence, to mimic how a human would press keys on a keyboard.
func codes(seqs ...[]uint8) (hexs []string) {
	for _, seq := range seqs {
		// press the keys
		for _, code := range seq {
			hexs = append(hexs, strconv.FormatUint(uint64(code), 16))
		}
		// release the keys
		for _, code := range seq {
			hexs = append(hexs, strconv.FormatUint(uint64(code|0x80), 16))
		}
	}
	return hexs
}
