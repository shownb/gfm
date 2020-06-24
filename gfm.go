//Package commonmark provides a Go wrapper for the CommonMark C Library
package gfm

/*
#cgo CFLAGS: -std=gnu99
#include <stdio.h>
#include <stdlib.h>
#include "cmark-gfm.h"
#include "cmark-gfm-extension_api.h"
#include "cmark-gfm-core-extensions.h"
*/
import "C"
import (
	"errors"
	"runtime"
	"strings"
	"unsafe"
)

// Converts Markdo--, er, CommonMark text to Html.
// Parameter mdtext contains CommonMark text.
// The return value is the HTML string
func Md2Html(mdtext string, options int) string {
	//The call to cmark will barf if the input string doesn't end with a newline
	if !strings.HasSuffix(mdtext, "\n") {
		mdtext += "\n"
	}
	mdCstr := C.CString(mdtext)
	strLen := C.size_t(len(mdtext))
	defer C.free(unsafe.Pointer(mdCstr))
	htmlString := C.cmark_markdown_to_html(mdCstr, strLen, C.int(options))
	defer C.free(unsafe.Pointer(htmlString))
	return C.GoString(htmlString)
}

//Wraps the cmark_doc_parser
type CMarkParser struct {
	parser *C.struct_cmark_parser
}

type CMarkExtension struct {
	*CMarkParser
	opts       int
	extensions *C.struct__cmark_llist
}

func (cme *CMarkExtension) ParserAttachExtension(exts ...string) {
	for _, ext := range exts {
		extStr := C.CString(ext)
		e := C.cmark_find_syntax_extension(extStr)
		C.cmark_parser_attach_syntax_extension(cme.parser, e)
		C.free(unsafe.Pointer(extStr))
	}
}

func (cme *CMarkExtension) ParserGetExtensions() {
	es, _ := C.cmark_parser_get_syntax_extensions(cme.parser)
	cme.extensions = es
}

func (cme *CMarkExtension) MarkdownToHtml(text string) string {
	s := len(text)
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.cmark_parser_feed(cme.parser, cstr, C.size_t(s))
	node := cme.Finish()
	defer node.Free()
	output := C.cmark_render_html(node.node, C.int(cme.opts), cme.extensions)
	defer C.free(unsafe.Pointer(output))
	return C.GoString(output)
}

func (cme *CMarkExtension) Close() {
	cme.Free()
}

func NewParserWithExts(options int, exts ...string) *CMarkExtension {
	C.cmark_gfm_core_extensions_ensure_registered()
	p := NewCmarkParser(options)
	cme := &CMarkExtension{
		CMarkParser: p,
		opts:        options,
	}
	cme.ParserAttachExtension(exts...)
	cme.ParserGetExtensions()
	return cme
}

// Retruns a new CMark Parser.
// You must call Free() on this thing when you're done with it!
// Please.
func NewCmarkParser(options int) *CMarkParser {
	p := &CMarkParser{
		parser: C.cmark_parser_new(C.int(options)),
	}
	runtime.SetFinalizer(p, (*CMarkParser).Free)
	return p
}

// Process some text
func (cmp *CMarkParser) Feed(text string) {
	s := len(text)
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.cmark_parser_feed(cmp.parser, cstr, C.size_t(s))
}

// Finish parsing and generate a document
// You must call Free() on the document when you're done with it!
func (cmp *CMarkParser) Finish() *CMarkNode {
	n := &CMarkNode{
		node: C.cmark_parser_finish(cmp.parser),
	}
	runtime.SetFinalizer(n, (*CMarkNode).Free)
	return n
}

// Cleanup the parser
// Once you call Free on this, you can't use it anymore
func (cmp *CMarkParser) Free() {
	if cmp.parser != nil {
		C.cmark_parser_free(cmp.parser)
	}
	cmp.parser = nil
}

// Generates a document directly from a string
func ParseDocument(buffer string, options int) *CMarkNode {
	if !strings.HasSuffix(buffer, "\n") {
		buffer += "\n"
	}
	Cstr := C.CString(buffer)
	Clen := C.size_t(len(buffer))
	defer C.free(unsafe.Pointer(Cstr))
	n := &CMarkNode{
		node: C.cmark_parse_document(Cstr, Clen, C.int(options)),
	}
	runtime.SetFinalizer(n, (*CMarkNode).Free)
	return n
}

// Parses a file and returns a CMarkNode
// Returns an error if the file can't be opened
func ParseFile(filename string, options int) (*CMarkNode, error) {
	fname := C.CString(filename)
	access := C.CString("r")
	defer C.free(unsafe.Pointer(fname))
	defer C.free(unsafe.Pointer(access))
	file := C.fopen(fname, access)
	if file == nil {
		return nil, errors.New("Unable to open file with name: " + filename)
	}
	defer C.fclose(file)
	n := &CMarkNode{
		node: C.cmark_parse_file(file, C.int(options)),
	}
	runtime.SetFinalizer(n, (*CMarkNode).Free)
	return n, nil
}

//Version information
func CMarkVersion() int {
	return int(C.cmark_version())
}
