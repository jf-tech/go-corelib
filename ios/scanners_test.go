package iohelper

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jf-tech/go-corelib/strs"
)

func TestNewScannerByDelim(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          io.Reader
		delim          string
		flags          ScannerByDelimFlag
		expectedTokens []string
	}{
		{
			name:           "multi-char delim | eof as delim | drop delim",
			input:          strings.NewReader("abc#123##efg####???##xyz##"),
			delim:          "##",
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{"abc#123", "efg", "", "???", "xyz"},
		},
		{
			name:           "CR LF delim | eof as delim | include delim",
			input:          strings.NewReader("\r\n\rabc\r"),
			delim:          "\r\n",
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagIncludeDelimInReturn,
			expectedTokens: []string{"\r\n", "\rabc\r"},
		},
		{
			name:           "empty reader",
			input:          strings.NewReader(""),
			delim:          "*",
			flags:          ScannerByDelimFlagDefault,
			expectedTokens: []string{},
		},
		{
			name:           "empty token",
			input:          strings.NewReader("*"),
			delim:          "*",
			flags:          ScannerByDelimFlagEofNotAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{""},
		},
		{
			name:           "trailing newlines",
			input:          strings.NewReader("*\n"),
			delim:          "*",
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagIncludeDelimInReturn,
			expectedTokens: []string{"*", "\n"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := NewScannerByDelim(test.input, test.delim, test.flags)
			tokens := []string{}
			for s.Scan() {
				tokens = append(tokens, s.Text())
			}
			assert.NoError(t, s.Err())
			assert.Equal(t, test.expectedTokens, tokens)
		})
	}
}

func TestNewScannerByDelim2(t *testing.T) {
	for _, test := range []struct {
		name           string
		input          io.Reader
		delim          string
		esc            rune
		flags          ScannerByDelimFlag
		expectedTokens []string
	}{
		{
			name:           "multi-char delim | with delim esc | eof as delim | drop delim",
			input:          strings.NewReader("abc#123##efg####???##xyz##"),
			delim:          "##",
			esc:            rune('?'),
			flags:          ScannerByDelimFlagEofAsDelim | ScannerByDelimFlagDropDelimInReturn,
			expectedTokens: []string{"abc#123", "efg", "", "???##xyz"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := NewScannerByDelim2(test.input, test.delim, strs.RunePtr(test.esc), test.flags)
			tokens := []string{}
			for s.Scan() {
				tokens = append(tokens, s.Text())
			}
			assert.NoError(t, s.Err())
			assert.Equal(t, test.expectedTokens, tokens)
		})
	}
}
