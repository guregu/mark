package mark

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// type position
type Pos int

// itemType identifies the type of lex items.
type itemType int

// Item represent a token or text string returned from the scanner
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

const eof = -1 // Zero value so closed channel delivers EOF

const (
	itemError itemType = iota // Error occurred; value is text of error
	// Intersting things
	itemNewLine
	itemHTML
	// Block Elements
	itemText
	itemLineBreak
	itemHeading
	itemLHeading
	itemBlockQuote
	itemList
	itemCodeBlock
	itemGfmCodeBlock
	itemHr
	itemTable
	// Span Elements
	itemLinks
	itemStrong
	itemItalic
	itemStrike
	itemCode
	itemImages
)

var (
	reEmphasise = "^_{%[1]d}([\\s\\S]+?)_{%[1]d}|^\\*{%[1]d}([\\s\\S]+?)\\*{%[1]d}"
	reGfmCode   = "^%s{3,} *(\\S+)? *\n([\\s\\S]+?)\\s*%s{3,}$*(?:\n+|$)"
)

// Block Grammer
var block = map[itemType]*regexp.Regexp{
	itemHeading:   regexp.MustCompile("^ *(#{1,6}) *([^\n]+?) *#* *(?:\n+|$)"),
	itemHr:        regexp.MustCompile("^( *[-*_]){3,} *(?:\n+|$)"),
	itemCodeBlock: regexp.MustCompile("^( {4}[^\n]+\n*)+"),
	// Backreferences is unavailable
	itemGfmCodeBlock: regexp.MustCompile(fmt.Sprintf(reGfmCode, "`", "`") + "|" + fmt.Sprintf(reGfmCode, "~", "~")),
}

// Inline Grammer
var span = map[itemType]*regexp.Regexp{
	itemItalic: regexp.MustCompile(fmt.Sprintf(reEmphasise, 1)),
	itemStrong: regexp.MustCompile(fmt.Sprintf(reEmphasise, 2)),
	itemStrike: regexp.MustCompile("^~{2}([\\s\\S]+?)~{2}"),
	// itemMixed(e.g: ***str***, ~~*str*~~) will be part of the parser
	itemCode: regexp.MustCompile("^`{1,2}\\s*([\\s\\S]*?[^`])\\s*`{1,2}"),
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// lex creates a new lexer for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

// next return the next rune in the input
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// lexAny scans non-space items.
func lexAny(l *lexer) stateFn {
	switch r := l.next(); r {
	case eof:
		return nil
	case '*', '-':
		if l.peek() == '*' || l.peek() == '-' {
			l.backup()
			return lexHr
		} else {
			l.backup()
			return lexList
		}
	case '#':
		l.backup()
		return lexHeading
	case ' ':
		// Should be here ?
		if block[itemCodeBlock].MatchString(l.input[l.pos-1:]) {
			l.backup()
			return lexCode
		}
		fallthrough
	case '`', '~':
		// if it's gfm-code
		c := l.input[l.pos : l.pos+2]
		if c == "``" || c == "~~" {
			l.backup()
			return lexGfmCode
		}
		fallthrough
	default:
		l.backup()
		return lexText
	}
}

// lexHeading scans heading items.
func lexHeading(l *lexer) stateFn {
	if block[itemHeading].MatchString(l.input[l.pos:]) {
		match := block[itemHeading].FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemHeading)
		return lexAny
	}
	return lexText
}

// lexHr scans horizontal rules items.
func lexHr(l *lexer) stateFn {
	if block[itemHr].MatchString(l.input[l.pos:]) {
		match := block[itemHr].FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemHr)
		return lexAny
	}
	return lexText
}

// lexGfmCode scans GFM code block.
func lexGfmCode(l *lexer) stateFn {
	re := block[itemGfmCodeBlock]
	if re.MatchString(l.input[l.pos:]) {
		match := re.FindString(l.input[l.pos:])
		l.pos += Pos(len(match))
		l.emit(itemGfmCodeBlock)
		return lexAny
	}
	return lexText
}

// lexCode scans code block.
func lexCode(l *lexer) stateFn {
	match := block[itemCodeBlock].FindString(l.input[l.pos:])
	l.pos += Pos(len(match))
	l.emit(itemCodeBlock)
	return lexAny
}

// lexList scans ordered and unordered lists.
func lexList(l *lexer) stateFn {
	// ...
	return lexText
}

// lexText scans until eol(\n)
func lexText(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case r == eof:
			break Loop
		case r == '\n' && l.peek() == '\n' || r == ' ' && l.peek() == ' ':
			// length of new-line
			l.pos++
			l.emit(itemNewLine)
			break Loop
		// if it's start as an emphasis
		case r == '_', r == '*', r == '~', r == '`':
			l.backup()
			input := l.input[l.pos:]
			// Strong
			if m := span[itemStrong].FindString(input); m != "" {
				l.pos += Pos(len(m))
				l.emit(itemStrong)
				break
			}
			// Italic
			if m := span[itemItalic].FindString(input); m != "" {
				l.pos += Pos(len(m))
				l.emit(itemItalic)
				break
			}
			// Strike
			if m := span[itemStrike].FindString(input); m != "" {
				l.pos += Pos(len(m))
				l.emit(itemStrike)
				break
			}
			// InlineCode
			if m := span[itemCode].FindString(input); m != "" {
				l.pos += Pos(len(m))
				l.emit(itemCode)
				break
			}
			// ~backup()
			l.pos += l.width
			fallthrough
		default:
			l.emit(itemText)
		}
	}
	return lexAny
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}
