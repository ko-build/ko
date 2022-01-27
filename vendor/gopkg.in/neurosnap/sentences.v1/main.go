/*
Package sentences is a golang package will convert a blob of text into a list of sentences.

This package attempts to support a multitude of languages:
czech, danish, dutch, english, estonian, finnish,
french, german, greek, italian, norwegian, polish,
portuguese, slovene, spanish, swedish, and turkish.

An unsupervised multilingual sentence boundary detection library for golang.
The goal of this library is to be able to break up any text into a list of
sentences in multiple languages.  The way the punkt system accomplishes this goal is
through training the tokenizer with text in that given language.
Once the likelyhoods of abbreviations, collocations, and sentence starters are
determined, finding sentence boundaries becomes easier.

There are many problems that arise when tokenizing text into sentences,
the primary issue being abbreviations. The punkt system attempts to determine
whether a word is an abbrevation, an end to a sentence, or even both through
training the system with text in the given language. The punkt system
incorporates both token- and type-based analysis on the text through two
different phases of annotation.

Original research article: http://citeseerx.ist.psu.edu/viewdoc/download;jsessionid=BAE5C34E5C3B9DC60DFC4D93B85D8BB1?doi=10.1.1.85.5017&rep=rep1&type=pdf
*/
package sentences
