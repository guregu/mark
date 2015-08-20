package mark

import (
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	cases := map[string]string{
		"foobar":               "<p>foobar</p>",
		"  foo bar":            "<p>  foo bar</p>",
		"foo|bar":              "<p>foo|bar</p>",
		"foo  \nbar":           "<p>foo<br>bar</p>",
		"__bar__ foo":          "<p><strong>bar</strong> foo</p>",
		"**bar** foo __bar__":  "<p><strong>bar</strong> foo <strong>bar</strong></p>",
		"**bar**__baz__":       "<p><strong>bar</strong><strong>baz</strong></p>",
		"**bar**foo__bar__":    "<p><strong>bar</strong>foo<strong>bar</strong></p>",
		"_bar_baz":             "<p><em>bar</em>baz</p>",
		"_foo_~~bar~~ baz":     "<p><em>foo</em><del>bar</del> baz</p>",
		"~~baz~~ _baz_":        "<p><del>baz</del> <em>baz</em></p>",
		"`bool` and thats it.": "<p><code>bool</code> and thats it.</p>",
		// Emphasis mixim
		"___foo___":       "<p><strong><em>foo</em></strong></p>",
		"__foo _bar___":   "<p><strong>foo <em>bar</em></strong></p>",
		"__*foo*__":       "<p><strong><em>foo</em></strong></p>",
		"_**mixim**_":     "<p><em><strong>mixim</strong></em></p>",
		"~~__*mixim*__~~": "<p><del><strong><em>mixim</em></strong></del></p>",
		"~~*mixim*~~":     "<p><del><em>mixim</em></del></p>",
		// Paragraph
		"1  \n2  \n3":        "<p>1<br>2<br>3</p>",
		"1\n\n2":             "<p>1</p>\n<p>2</p>",
		"1\n\n\n2":           "<p>1</p>\n<p>2</p>",
		"1\n\n\n\n\n\n\n\n2": "<p>1</p>\n<p>2</p>",
		// Heading
		"# 1\n## 2":                   "<h1 id=\"1\">1</h1>\n<h2 id=\"2\">2</h2>",
		"# 1\np\n## 2\n### 3\n4\n===": "<h1 id=\"1\">1</h1>\n<p>p</p>\n<h2 id=\"2\">2</h2>\n<h3 id=\"3\">3</h3>\n<h1 id=\"4\">4</h1>",
		"Hello\n===":                  "<h1 id=\"hello\">Hello</h1>",
		// Links
		"[text](link \"title\")": "<p><a href=\"link\" title=\"title\">text</a></p>",
		"[text](link)":           "<p><a href=\"link\">text</a></p>",
		"[](link)":               "<p><a href=\"link\"></a></p>",
		"Link: [example](#)":     "<p>Link: <a href=\"#\">example</a></p>",
		"Link: [not really":      "<p>Link: [not really</p>",
		"http://localhost:3000":  "<p><a href=\"http://localhost:3000\">http://localhost:3000</a></p>",
		"Link: http://yeah.com":  "<p>Link: <a href=\"http://yeah.com\">http://yeah.com</a></p>",
		"<http://foo.com>":       "<p><a href=\"http://foo.com\">http://foo.com</a></p>",
		"Link: <http://l.co>":    "<p>Link: <a href=\"http://l.co\">http://l.co</a></p>",
		"Link: <not really":      "<p>Link: &lt;not really</p>",
		// CodeBlock
		"\tfoo\n\tbar": "<pre><code>foo\nbar</code></pre>",
		"\tfoo\nbar":   "<pre><code>foo\n</code></pre>\n<p>bar</p>",
		// GfmCodeBlock
		"```js\nvar a;\n```":  "<pre><code class=\"lang-js\">var a;</code></pre>",
		"~~~\nvar b;~~~":      "<pre><code>var b;</code></pre>",
		"~~~js\nlet d = 1~~~": "<pre><code class=\"lang-js\">let d = 1</code></pre>",
		// Hr
		"foo\n****\nbar": "<p>foo</p>\n<hr>\n<p>bar</p>",
		"foo\n___":       "<p>foo</p>\n<hr>",
		// Images
		"![name](url)":           "<p><img src=\"url\" alt=\"name\"></p>",
		"![name](url \"title\")": "<p><img src=\"url\" alt=\"name\" title=\"title\"></p>",
		"img: ![name]()":         "<p>img: <img src=\"\" alt=\"name\"></p>",
		// Lists
		"- foo\n- bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		"* foo\n* bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		"+ foo\n+ bar": "<ul>\n<li>foo</li>\n<li>bar</li>\n</ul>",
		// // Ordered Lists
		"1. one\n2. two\n3. three": "<ol>\n<li>one</li>\n<li>two</li>\n<li>three</li>\n</ol>",
		"1. one\n 1. one of one":   "<ol>\n<li>one<ol>\n<li>one of one</li>\n</ol></li>\n</ol>",
		"2. two\n 3. three":        "<ol>\n<li>two<ol>\n<li>three</li>\n</ol></li>\n</ol>",
		// Special characters escaping
		"< hello":   "<p>&lt; hello</p>",
		"hello >":   "<p>hello &gt;</p>",
		"foo & bar": "<p>foo &amp; bar</p>",
		"'foo'":     "<p>&#39;foo&#39;</p>",
		"\"foo\"":   "<p>&quot;foo&quot;</p>",
		"&copy;":    "<p>&copy;</p>",
		// Backslash escaping
		"\\**foo\\**":       "<p>*<em>foo*</em></p>",
		"\\*foo\\*":         "<p>*foo*</p>",
		"\\_underscores\\_": "<p>_underscores_</p>",
		"\\## header":       "<p>## header</p>",
		"header\n\\===":     "<p>header\n\\===</p>",
	}
	for input, expected := range cases {
		if actual := Render(input); actual != expected {
			t.Errorf("%s: got\n%+v\nexpected\n%+v", input, actual, expected)
		}
	}
}

func TestData(t *testing.T) {
	var testFiles []string
	files, err := ioutil.ReadDir("test")
	if err != nil {
		t.Error("Couldn't open 'test' directory")
	}
	for _, file := range files {
		if name := file.Name(); strings.HasSuffix(name, ".text") {
			testFiles = append(testFiles, "test/"+strings.TrimSuffix(name, ".text"))
		}
	}
	re := regexp.MustCompile(`\n`)
	for _, file := range testFiles {
		html, err := ioutil.ReadFile(file + ".html")
		if err != nil {
			t.Errorf("Error to read html file: %s", file)
		}
		text, err := ioutil.ReadFile(file + ".text")
		if err != nil {
			t.Errorf("Error to read text file: %s", file)
		}
		// Remove '\n'
		sHTML := re.ReplaceAllLiteralString(string(html), "")
		output := Render(string(text))
		opts := DefaultOptions()
		if strings.Contains(file, "smartypants") {
			opts.Smartypants = true
			output = New(string(text), opts).Render()
		}
		if strings.Contains(file, "smartyfractions") {
			opts.Fractions = true
			output = New(string(text), opts).Render()
		}
		sText := re.ReplaceAllLiteralString(output, "")
		if sHTML != sText {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%+v", file, sText, sHTML)
		}
	}
}

// TODO: Add more tests for it.
func TestRenderFn(t *testing.T) {
	m := New("hello world", nil)
	m.AddRenderFn(NodeParagraph, func(n Node) (s string) {
		if p, ok := n.(*ParagraphNode); ok {
			s += "<p class=\"mv-msg\">"
			for _, pp := range p.Nodes {
				s += pp.Render()
			}
			s += "</p>"
		}
		return
	})
	expected := "<p class=\"mv-msg\">hello world</p>"
	if actual := m.Render(); actual != expected {
		t.Errorf("RenderFn: got\n\t%+v\nexpected\n\t%+v", actual, expected)
	}
}

type CommonMarkSpec struct {
	name     string
	input    string
	expected string
}

var CMCases = []CommonMarkSpec{
	{"6", "- `one\n- two`", "<ul><li>`one</li><li>two`</li></ul>"},
	{"7", "***\n---\n___", "<hr><hr><hr>"},
	{"8", "+++", "<p>+++</p>"},
	{"9", "===", "<p>===</p>"},
	{"10", "--\n**\n__", "<p>--**__</p>"},
	{"11", " ***\n  ***\n   ***", "<hr><hr><hr>"},
	{"12", "    ***", "<pre><code>***</code></pre>"},
	{"14", "_____________________________________", "<hr>"},
	{"15", " - - -", "<hr>"},
	{"16", " **  * ** * ** * **", "<hr>"},
	{"17", "-     -      -      -", "<hr>"},
	{"18", "- - - -    ", "<hr>"},
	{"21", "- foo\n***\n- bar", "<ul>\n<li>foo</li>\n</ul>\n<hr>\n<ul>\n<li>bar</li>\n</ul>"},
	{"22", "Foo\n***\nbar", "<p>Foo</p><hr><p>bar</p>"},
	{"23", "Foo\n---\nbar", "<h2>Foo</h2><p>bar</p>"},
	{"24", "* Foo\n* * *\n* Bar", "<ul>\n<li>Foo</li>\n</ul>\n<hr>\n<ul>\n<li>Bar</li>\n</ul>"},
	{"25", "- Foo\n- * * *", "<ul>\n<li>Foo</li>\n<li>\n<hr>\n</li>\n</ul>"},
	{"26", `# foo
## foo
### foo
#### foo
##### foo
###### foo`, `<h1>foo</h1>
<h2>foo</h2>
<h3>foo</h3>
<h4>foo</h4>
<h5>foo</h5>
<h6>foo</h6>`},
	{"27", "####### foo", "<p>####### foo</p>"},
	{"28", "#5 bolt\n\n#foobar", "<p>#5 bolt</p>\n<p>#foobar</p>"},
	{"29", "\\## foo", "<p>## foo</p>"},
	{"31", "#                  foo                     ", "<h1>foo</h1>"},
	{"32", ` ### foo
  ## foo
   # foo`, `<h3>foo</h3>
<h2>foo</h2>
<h1>foo</h1>`},
	{"33", "    # foo", "<pre><code># foo</code></pre>"},
	{"35", `## foo ##
  ###   bar    ###`, `<h2>foo</h2>
<h3>bar</h3>`},
	{"36", `# foo ##################################
##### foo ##`, `<h1>foo</h1>
<h5>foo</h5>`},
	{"37", "### foo ###     ", "<h3>foo</h3>"},
	{"38", "### foo ### b", "<h3>foo ### b</h3>"},
	{"41", `****
## foo
****`, `<hr>
<h2>foo</h2>
<hr>`},
	{"42", `Foo bar
# baz
Bar foo`, `<p>Foo bar</p>
<h1>baz</h1>
<p>Bar foo</p>`},
	{"45", `Foo
-------------------------

Foo
=`, `<h2>Foo</h2>
<h1>Foo</h1>`},
	{"46", `   Foo
---

  Foo
-----

  Foo
  ===`, `<h2>Foo</h2>
<h2>Foo</h2>
<h1>Foo</h1>`},
	{"47", `    Foo
    ---

    Foo
---`, `<pre><code>Foo
---

Foo
</code></pre>
<hr>`},
	{"48", `Foo
   ----      `, "<h2>Foo</h2>"},
	{"50", `Foo
= =

Foo
--- -`, `<p>Foo
= =</p>
<p>Foo</p>
<hr>`},
	{"51", `Foo  
-----`, "<h2>Foo</h2>"},
	{"52", `Foo\
----`, "<h2>Foo\\</h2>"},
	{"53", "`Foo\n----\n`\n\n<a title=\"a lot\n---\nof dashes\"/>", "<h2>`Foo</h2>\n<p>`</p>\n<h2>&lt;a title=&quot;a lot</h2>\n<p>of dashes&quot;/&gt;</p>"},
	{"55", `- Foo
---`, `<ul>
<li>Foo</li>
</ul>
<hr>`},
	{"57", `---
Foo
---
Bar
---
Baz`, `<hr>
<h2>Foo</h2>
<h2>Bar</h2>
<p>Baz</p>`},
	{"58", "====", "<p>====</p>"},
	{"59", `---
---`, "<hr><hr>"},
	{"60", `- foo
-----`, `<ul>
<li>foo</li>
</ul>
<hr>`},
	{"61", `    foo
---`, `<pre><code>foo
</code></pre>
<hr>`},
	{"64", `    a simple
      indented code block`, `<pre><code>a simple
  indented code block
</code></pre>`},
	{"66", `1.  foo

    - bar`, `<ol>
<li>
<p>foo</p>
<ul>
<li>bar</li>
</ul>
</li>
</ol>`},
	{"67", `    <a/>
    *hi*

    - one`, `<pre><code>&lt;a/&gt;
*hi*

- one
</code></pre>`},
	{"69", `    chunk1
      
      chunk2`, `<pre><code>chunk1
  
  chunk2
</code></pre>`},
	{"71", `    foo
bar`, `<pre><code>foo
</code></pre>
<p>bar</p>`},
	{"72", `# Header
    foo
Header
------
    foo
----`, `<h1>Header</h1>
<pre><code>foo
</code></pre>
<h2>Header</h2>
<pre><code>foo
</code></pre>
<hr>`},
	{"73", `        foo
    bar`, `<pre><code>    foo
bar
</code></pre>`},
	{"74", `    
    foo
    `, `<pre><code>foo
</code></pre>`},
	{"75", "    foo  ", `<pre><code>foo  
</code></pre>`},
	{"76", "```\n< \n>\n```", `<pre><code>&lt;
 &gt;
</code></pre>`},
	{"77", `~~~
<
 >
~~~`, `<pre><code>&lt;
 &gt;
</code></pre>`},
	{"78", "```\naaa\n~~~\n```", `<pre><code>aaa
~~~
</code></pre>`},
	{"79", "~~~\naaa\n```\n~~~", "<pre><code>aaa\n```\n</code></pre>"},
	{"86", "```\n```", `<pre><code></code></pre>`},
	{"90", "    ```\n    aaa\n    ```", "<pre><code>```\naaa\n```\n</code></pre>"},
	{"91", "```\naaa\n  ```", `<pre><code>aaa
</code></pre>`},
	{"92", "   ```\naaa\n  ```", `<pre><code>aaa
</code></pre>`},
	{"96", "foo\n```\nbar\n```\nbaz", `<p>foo</p>
<pre><code>bar
</code></pre>
<p>baz</p>`},
	{"97", `foo
---
~~~
bar
~~~
# baz`, `<h2>foo</h2>
<pre><code>bar
</code></pre>
<h1>baz</h1>`},
	{"102", "```\n``` aaa\n```", "<pre><code>``` aaa\n</code></pre>"},
	{"103", `
<table>
  <tr>
    <td>
           hi
    </td>
  </tr>
</table>

okay.`, `
<table>
  <tr>
    <td>
           hi
    </td>
  </tr>
</table>
<p>okay.</p>`},
	// Move out the id, beacuse the regexp below
	{"107", `
<div
  class="bar">
</div>`, `
<div
  class="bar">
</div>`},
	{"108", `
<div class="bar
  baz">
</div>`, `
<div class="bar
  baz">
</div>`},
	{"113", `<div><a href="bar">*foo*</a></div>`, `<div><a href="bar">*foo*</a></div>`},
	{"114", `
<table><tr><td>
foo
</td></tr></table>`, `
<table><tr><td>
foo
</td></tr></table>`},
	{"117", `
<Warning>
*bar*
</Warning>`, `
<Warning>
*bar*
</Warning>`},
	{"121", "<del>*foo*</del>", "<p><del><em>foo</em></del></p>"},
	{"122", `
<pre language="haskell"><code>
import Text.HTML.TagSoup

main :: IO ()
main = print $ parseTags tags
</code></pre>`, `
<pre language="haskell"><code>
import Text.HTML.TagSoup

main :: IO ()
main = print $ parseTags tags
</code></pre>`},
	{"123", `
<script type="text/javascript">
// JavaScript example

document.getElementById("demo").innerHTML = "Hello JavaScript!";
</script>`, `
<script type="text/javascript">
// JavaScript example

document.getElementById("demo").innerHTML = "Hello JavaScript!";
</script>`},
	{"124", `
<style
  type="text/css">
h1 {color:red;}

p {color:blue;}
</style>`, `
<style
  type="text/css">
h1 {color:red;}

p {color:blue;}
</style>`},
	{"127", `
- <div>
- foo`, `
<ul>
<li>
<div>
</li>
<li>foo</li>
</ul>`},
	{"137", `
Foo
<div>
bar
</div>`, `
<p>Foo</p>
<div>
bar
</div>`},
	{"139", `
Foo
<a href="bar">
baz`, `
<p>Foo
<a href="bar">
baz</p>`},
	{"141", `
<div>
*Emphasized* text.
</div>`, `
<div>
*Emphasized* text.
</div>
`},
	{"142", `
<table>

<tr>

<td>
Hi
</td>

</tr>

</table>`, `
<table>
<tr>
<td>
Hi
</td>
</tr>
</table>
`},
	{"144", `
[foo]: /url "title"

[foo]`, `<p><a href="/url" title="title">foo</a></p>`},
	{"148", `
[foo]: /url '
title
line1
line2
'

[foo]`, `
<p><a href="/url" title="
title
line1
line2
">foo</a></p>`},
	{"151", `
[foo]:

[foo]`, `
<p>[foo]:</p>
<p>[foo]</p>`},
	{"153", `
[foo]

[foo]: url`, `<p><a href="url">foo</a></p>`},
	{"154", `
[foo]

[foo]: first
[foo]: second`, `<p><a href="first">foo</a></p>`},
	{"155", `
[FOO]: /url

[Foo]`, `<p><a href="/url">Foo</a></p>`},
	{"157", "[foo]: /url", ""},
	{"158", `
[
foo
]: /url
bar`, "<p>bar</p>"},
	{"159", `[foo]: /url "title" ok`, "<p>[foo]: /url &quot;title&quot; ok</p>"},
	{"160", `
[foo]: /url
"title" ok`, "<p>&quot;title&quot; ok</p>"},
	{"161", `
    [foo]: /url "title"

[foo]`, `
<pre><code>[foo]: /url &quot;title&quot;
</code></pre>
<p>[foo]</p>`},
	{"162", "```\n[foo]: /url\n```\n\n[foo]", `
<pre><code>[foo]: /url
</code></pre>
<p>[foo]</p>`},
	{"166", `
[foo]

> [foo]: /url`, `
<p><a href="/url">foo</a></p>
<blockquote>
</blockquote>`},
	{"167", `
aaa

bbb`, `
<p>aaa</p>
<p>bbb</p>`},
	{"168", `
aaa
bbb

ccc
ddd`, `
<p>aaa
bbb</p>
<p>ccc
ddd</p>`},
	{"169", `
aaa


bbb`, `
<p>aaa</p>
<p>bbb</p>`},
	{"173", `
    aaa
bbb`, `
<pre><code>aaa
</code></pre>
<p>bbb</p>`},
	{"175", `
  

aaa
  

# aaa

  `, `
<p>aaa</p>
<h1>aaa</h1>`},
	{"176", `
> # Foo
> bar
> baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"177", `
># Foo
>bar
> baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"179", `
    > # Foo
    > bar
    > baz`, `
<pre><code>&gt; # Foo
&gt; bar
&gt; baz
</code></pre>`},
	{"180", `
> # Foo
> bar
baz`, `
<blockquote>
<h1>Foo</h1>
<p>bar
baz</p>
</blockquote>`},
	{"181", `
> bar
baz
> foo`, `
<blockquote>
<p>bar
baz
foo</p>
</blockquote>`},
	{"188", `
>
> foo
>  `, `
<blockquote>
<p>foo</p>
</blockquote>`},
	{"190", `
> foo
> bar`, `
<blockquote>
<p>foo
bar</p>
</blockquote>`},
	{"191", `
> foo
>
> bar`, `
<blockquote>
<p>foo</p>
<p>bar</p>
</blockquote>`},
	{"192", `
foo
> bar`, `
<p>foo</p>
<blockquote>
<p>bar</p>
</blockquote>`},
	{"194", `
> bar
baz`, `
<blockquote>
<p>bar
baz</p>
</blockquote>`},
	{"195", `
> bar

baz`, `
<blockquote>
<p>bar</p>
</blockquote>
<p>baz</p>`},
	{"197", `
> > > foo
bar`, `
<blockquote>
<blockquote>
<blockquote>
<p>foo
bar</p>
</blockquote>
</blockquote>
</blockquote>`},
	{"198", `
>>> foo
> bar
>>baz`, `
<blockquote>
<blockquote>
<blockquote>
<p>foo
bar
baz</p>
</blockquote>
</blockquote>
</blockquote>`},
	{"200", `
A paragraph
with two lines.

    indented code

> A block quote.`, `
<p>A paragraph
with two lines.</p>
<pre><code>indented code
</code></pre>
<blockquote>
<p>A block quote.</p>
</blockquote>`},
	{"201", `
1.  A paragraph
    with two lines.

        indented code

    > A block quote.`, `
<ol>
<li>
<p>A paragraph
with two lines.</p>
<pre><code>indented code
</code></pre>
<blockquote>
<p>A block quote.</p>
</blockquote>
</li>
</ol>`},
	{"203", `
- one

  two`, `
<ul>
<li>
<p>one</p>
<p>two</p>
</li>
</ul>`},
	{"206", `
   > > 1.  one
>>
>>     two`, `
<blockquote>
<blockquote>
<ol>
<li>
<p>one</p>
<p>two</p>
</li>
</ol>
</blockquote>
</blockquote>`},
	{"208", `-one

2.two`, `
<p>-one</p>
<p>2.two</p>`},
	{"210", `
1.  foo

    ~~~
    bar
    ~~~

    baz

    > bam`, `
<ol>
<li>
<p>foo</p>
<pre><code>bar
</code></pre>
<p>baz</p>
<blockquote>
<p>bam</p>
</blockquote>
</li>
</ol>`},
	{"212", `1234567890. not ok`, `<p>1234567890. not ok</p>`},
	{"215", `-1. not ok`, `<p>-1. not ok</p>`},
	{"216", `
- foo

      bar`, `
<ul>
<li>
<p>foo</p>
<pre><code>bar
</code></pre>
</li>
</ul>`},
	{"218", `
    indented code

paragraph

    more code`, `
<pre><code>indented code
</code></pre>
<p>paragraph</p>
<pre><code>more code
</code></pre>`},
	{"223", `
-  foo

   bar`, `
<ul>
<li>
<p>foo</p>
<p>bar</p>
</li>
</ul>`},
	{"226", `
- foo
-   
- bar`, `
<ul>
<li>foo</li>
<li></li>
<li>bar</li>
</ul>`},
	{"232", `
    1.  A paragraph
        with two lines.

            indented code

        > A block quote.`, `
<pre><code>1.  A paragraph
    with two lines.

        indented code

    &gt; A block quote.
</code></pre>`},
	{"234", `
  1.  A paragraph
    with two lines.`, `
<ol>
<li>A paragraph
with two lines.</li>
</ol>`},
	{"235", `
> 1. > Blockquote
continued here.`, `
<blockquote>
<ol>
<li>
<blockquote>
<p>Blockquote
continued here.</p>
</blockquote>
</li>
</ol>
</blockquote>`},
	{"236", `
> 1. > Blockquote
continued here.`, `
<blockquote>
<ol>
<li>
<blockquote>
<p>Blockquote
continued here.</p>
</blockquote>
</li>
</ol>
</blockquote>`},
	{"237", `
- foo
  - bar
    - baz`, `
<ul>
<li>foo
<ul>
<li>bar
<ul>
<li>baz</li>
</ul>
</li>
</ul>
</li>
</ul>`},
	{"241", "- - foo", `
<ul>
<li>
<ul>
<li>foo</li>
</ul>
</li>
</ul>`},
	{"243", `
- # Foo
- Bar
  ---
  baz`, `
<ul>
<li>
<h1>Foo</h1>
</li>
<li>
<h2>Bar</h2>
baz</li>
</ul>`},
	{"246", `
Foo
- bar
- baz`, `
<p>Foo</p>
<ul>
<li>bar</li>
<li>baz</li>
</ul>`},
	{"248", `
- foo

- bar


- baz`, `
<ul>
<li>
<p>foo</p>
</li>
<li>
<p>bar</p>
</li>
</ul>
<ul>
<li>baz</li>
</ul>`},
	{"250", `
- foo
  - bar
    - baz


      bim`, `
<ul>
<li>foo
<ul>
<li>bar
<ul>
<li>baz</li>
</ul>
</li>
</ul>
</li>
</ul>
<pre><code>  bim
</code></pre>`},
	{"251", `
- foo
- bar


- baz
- bim`, `
<ul>
<li>foo</li>
<li>bar</li>
</ul>
<ul>
<li>baz</li>
<li>bim</li>
</ul>`},
	{"252", `
-   foo

    notcode

-   foo


    code`, `
<ul>
<li>
<p>foo</p>
<p>notcode</p>
</li>
<li>
<p>foo</p>
</li>
</ul>
<pre><code>code
</code></pre>`},
	{"261", `
* a
  > b
  >
* c`, `
<ul>
<li>a
<blockquote>
<p>b</p>
</blockquote>
</li>
<li>c</li>
</ul>`},
	{"263", "- a", `
<ul>
<li>a</li>
</ul>`},
	{"264", `
- a
  - b`, `
<ul>
<li>a
<ul>
<li>b</li>
</ul>
</li>
</ul>`},
	{"266", `
* foo
  * bar

  baz`, `
<ul>
<li>
<p>foo</p>
<ul>
<li>bar</li>
</ul>
<p>baz</p>
</li>
</ul>`},
	{"267", `
- a
  - b
  - c

- d
  - e
  - f`, `
<ul>
<li>
<p>a</p>
<ul>
<li>b</li>
<li>c</li>
</ul>
</li>
<li>
<p>d</p>
<ul>
<li>e</li>
<li>f</li>
</ul>
</li>
</ul>`},
	{"268", "`hi`lo`", "<p><code>hi</code>lo`</p>"},
	{"275", `    \[\]`, `<pre><code>\[\]
</code></pre>`},
	{"276", `
~~~
\[\]
~~~`, `
<pre><code>\[\]
</code></pre>`},
}

func TestCommonMark(t *testing.T) {
	reId := regexp.MustCompile(` +?id=".*"`)
	for _, c := range CMCases {
		// Remove the auto-hashing until it'll be in the configuration
		actual := reId.ReplaceAllString(Render(c.input), "")
		if strings.Replace(actual, "\n", "", -1) != strings.Replace(c.expected, "\n", "", -1) {
			t.Errorf("\ninput:%s\ngot:\n%s\nexpected:\n%s\nlink: http://spec.commonmark.org/0.21/#example-%s\n",
				c.input, actual, c.expected, c.name)
		}
	}
}
