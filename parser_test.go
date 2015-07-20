package mark

// import (
// 	"fmt"
// 	"github.com/k0kubun/pp"
// 	"testing"
// )

// func jTestParser(t *testing.T) {
// 	l := lex("1", "foo bar baz\nhello world")
// 	p := &Tree{lex: l}
// 	fmt.Println("peek")
// 	item := p.peek()
// 	fmt.Println(tokenNames[item.typ], "-->", item.val)
// 	item = p.peek()
// 	fmt.Println(tokenNames[item.typ], "-->", item.val)
// 	fmt.Println("next")
// 	item = p.next()
// 	fmt.Println(tokenNames[item.typ], "-->", item.val)
// }

// func ParseFn(*testing.T) {
// 	l := lex("2", "hello\nworld. **ariel**foo  \nenter hahaha  \n~~hello~~ world  \n_bar_  \n This is my code:`javascript`")
// 	p := &Tree{lex: l}
// 	p.parse()

// 	pp.Printf("[Message]: Tree Node List After Compile\n\n")
// 	pp.Println(p.Nodes)
// 	pp.Println("Length of nodes:", len(p.Nodes))
// 	p.render()
// 	pp.Printf(p.output + "\n")

// 	l = lex("3", "```js\nMy Code Block\nbla bla\n```\n    block code    \n    yeah with more rows    \n")
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("4", "fooo  \n***\nafter hr")
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("5", "#foo bar")
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("6", `
// this is header
// ===
// And then we have some dummy text...

// ## This is H2!!!
// `)
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("7", `
// paragraph
// <http://autolink.com>
// [text](http://localhost.com "Ariel")
// [text](http://google.com)
// done
// https://github.link done!
// `)
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("8", `
// ![name](http://github.com/foo.gif "Title")
// ![name](http://only-url)
// paragraph
// `)
// 	p = &Tree{lex: l}
// 	p.parse()
// 	p.render()
// 	pp.Printf("\n" + p.output + "\n")

// 	l = lex("8", `
// - This is listItem
// - And another listItem
// `)

// 	for item := range l.items {
// 		pp.Println(item)
// 	}
// }
//
