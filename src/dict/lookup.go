package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework AppKit
#import <Foundation/Foundation.h>
#import <AppKit/NSSpellChecker.h>

const char*
define(const char *s, int *start, int *len) {
  NSString *word = [[NSString alloc] initWithCString:s encoding:NSUTF8StringEncoding];
  CFRange termRange = DCSGetTermRangeInString(NULL, (CFStringRef)word, 0);

  NSString *definition = @"";
  *start = termRange.location;
  *len = termRange.length;

  if(*start != -1) {
    definition = (NSString*)DCSCopyTextDefinition(NULL, (CFStringRef)word, termRange);
    NSString *first_part = [word substringToIndex: *start];
    NSString *second_part = [word substringWithRange: NSMakeRange(*start, *len)];
    *start = [first_part lengthOfBytesUsingEncoding: NSUTF8StringEncoding];
    *len = [second_part lengthOfBytesUsingEncoding: NSUTF8StringEncoding];
  }

  return [definition UTF8String];
}

const char **
spell(const char * s, int *n) {
  NSString * term = [[NSString alloc] initWithCString:s encoding:NSUTF8StringEncoding];
  NSSpellChecker *spellChecker = [NSSpellChecker sharedSpellChecker];
  NSArray *guesses = [spellChecker guessesForWordRange:NSMakeRange(0, [term length])
                                                  inString:term
                                                  language:[spellChecker language]
                                    inSpellDocumentWithTag:0];
  int count = [guesses count];
  *n = count;
  const char** cArray = malloc(sizeof(const char *) * count);
  for (int i=0; i<count; ++i) {
    cArray[i] = [[guesses objectAtIndex:i] UTF8String];
  }
  return cArray;
}
*/
import "C"

import (
	"strings"
	"unsafe"
)

func define(word string) (string, string) {
	cs := C.CString(word)
	defer C.free(unsafe.Pointer(cs))
	var start, l C.int
	cdef := C.define(cs, &start, &l)
	if start != -1 {
		word = word[start : start+l]
	}
	return word, C.GoString(cdef)
}

func spell(word string) []string {
	cs := C.CString(word)
	defer C.free(unsafe.Pointer(cs))
	var n C.int
	r := C.spell(cs, &n)
	ar := ((*[1 << 30]*C.char)(unsafe.Pointer(r)))[:n]
	defer C.free(unsafe.Pointer(r))
	guesses := make([]string, 0)
	for _, s := range ar {
		guesses = append(guesses, C.GoString(s))
	}
	return guesses
}
func lookup(q string, limit int) [][]string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}

	words := spell(q)
	words = append(words, q)
	definitions := make([][]string, 0)
	for _, word := range words {
		if limit == 0 {
			break
		}
		subword, def := define(word)
		if def == "" {
			continue
		}
		def = strings.Replace(def, "\n", " ", -1)
		def = strings.TrimSpace(def)
		fields := strings.Fields(def)
		if len(fields) > 0 {
			fields = fields[1:]
		}
		def = strings.Join(fields, " ")
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}
		definitions = append(definitions, []string{subword, def})
		limit -= 1
	}

	limit = len(definitions)
	out := make([][]string, limit+1)
	i := 0
	for _, row := range definitions {
		i++
		word := row[0]
		def := row[1]
		if out[0] == nil && word == q {
			out[0] = []string{q, def}
			continue
		}
		if i == limit+1 {
			break
		}
		out[i] = []string{word, def}
	}

	if out[0] == nil {
		word, def := define(q)
		if def != "" {
			out[0] = []string{word, def}
		} else {
			out = append(out[1:], []string{})
		}
	}

	return out[0 : len(out)-1]
}
