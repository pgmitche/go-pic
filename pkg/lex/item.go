package lex

import (
	"fmt"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ  itemType // The type of this item.
	pos  Pos      // The starting position, in bytes, of this item in the input string.
	val  string   // The value of this item.
	line int      // The line number at the start of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"

	case i.typ == itemError:
		return i.val
	}

	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError      itemType = iota // error occurred; value is text of error
	itemBool                       // boolean constant
	itemChar                       // printable ASCII character; grab bag for comma etc.
	itemComplex                    // complex constant (1+2i); imaginary is just a number
	itemEOF                        // end of file
	itemEOL                        // end of line
	itemIdentifier                 // Name of PIC or group
	itemNumber                     // simple number, including imaginary
	itemSpace                      // run of spaces separating arguments
	itemDot                        // the cursor, spelled '.'
	itemOCCURS                     // OCCURS keyword
	itemPIC                        // PIC keyword
	itemREDEFINES                  // REDEFINES keyword
	itemEnum                       // enum example: 'Y' 'N' 'T' 'F'
)

const (
	eof        = -1
	picLeft    = 'P'
	picRight   = '.'
	leftParen  = '('
	rightParen = ')'
)

var (
	picChars = map[rune]struct{}{
		'P': {}, 'I': {}, 'C': {}, leftParen: {}, rightParen: {}, 'X': {}, '9': {}, 'S': {}, 'V': {},
	}

	picTypes = map[rune]struct{}{
		'X': {}, '9': {}, 'S': {}, 'V': {},
	}
)
