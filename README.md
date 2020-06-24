a Go (golang) wrapper for the cmark-gfm https://github.com/github/cmark-gfm C library

and thanks https://github.com/rhinoman/go-commonmark

example

```golang
package main

import (
	"fmt"
	"gfm"
	"io/ioutil"
)

func main() {
	b, err := ioutil.ReadFile("markdown.md")
    if err != nil {
        fmt.Print(err)
    }
	md := gfm.NewParserWithExts(2, "autolink", "table", "strikethrough", "tagfilter")
	o := md.MarkdownToHtml(string(b))
	fmt.Printf("%v\n", o)
	o = md.MarkdownToHtml("ok")
	fmt.Printf("%v\n", o)
	md.Close()
}

```
