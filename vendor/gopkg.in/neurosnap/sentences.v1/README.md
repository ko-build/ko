[![Build Status](https://travis-ci.org/neurosnap/sentences.svg)](https://travis-ci.org/neurosnap/sentences)
[![GODOC](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/gopkg.in/neurosnap/sentences.v1)
![MIT](https://img.shields.io/packagist/l/doctrine/orm.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/neurosnap/sentences)](https://goreportcard.com/report/github.com/neurosnap/sentences)

Sentences - A command line sentence tokenizer
=============================================

This command line utility will convert a blob of text into a list of sentences.

* [Demo](http://sentences.erock.io)
* [Docs](https://godoc.org/gopkg.in/neurosnap/sentences.v1)

Install
-------

```
go get gopkg.in/neurosnap/sentences.v1
go install gopkg.in/neurosnap/sentences.v1/cmd/sentences
```

### Binaries

#### Linux

* [Linux 386](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_linux-386.tar.gz)
* [Linux AMD64](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_linux-amd64.tar.gz)

#### Mac

* [Darwin 386](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_darwin-386.tar.gz)
* [Darwin AMD64](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_darwin-amd64.tar.gz)

#### Windows

* [Windows 386](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_windows-386.tar.gz)
* [Windows AMD64](https://s3-us-west-2.amazonaws.com/sentence-binaries/sentences_windows-amd64.tar.gz)

Command
-------

![Command line](sentences.gif?raw=true)

Get it
------

```
go get gopkg.in/neurosnap/sentences.v1
```

Use it
------

```Go
import (
    "fmt"

    "gopkg.in/neurosnap/sentences.v1"
    "gopkg.in/neurosnap/sentences.v1/data"
)

func main() {
    text := `A perennial also-ran, Stallings won his seat when longtime lawmaker David Holmes
    died 11 days after the filing deadline. Suddenly, Stallings was a shoo-in, not
    the long shot. In short order, the Legislature attempted to pass a law allowing
    former U.S. Rep. Carolyn Cheeks Kilpatrick to file; Stallings challenged the
    law in court and won. Kilpatrick mounted a write-in campaign, but Stallings won.`

    // Compiling language specific data into a binary file can be accomplished
    // by using `make <lang>` and then loading the `json` data:
    b, _ := data.Asset("data/english.json");

    // load the training data
    training, _ := sentences.LoadTraining(b)

    // create the default sentence tokenizer
    tokenizer := sentences.NewSentenceTokenizer(training)
    sentences := tokenizer.Tokenize(text)

    for _, s := range sentences {
        fmt.Println(s.Text)
    }
}
```

English
-------

This package attempts to fix some problems I noticed for english.

```Go
import (
    "fmt"

    "gopkg.in/neurosnap/sentences.v1/english"
)

func main() {
    text := "Hi there. Does this really work?"

    tokenizer, err := english.NewSentenceTokenizer(nil)
    if err != nil {
        panic(err)
    }

    sentences := tokenizer.Tokenize(text)
    for _, s := range sentences {
        fmt.Println(s.Text)
    }
}
```

Contributing
------------

I need help maintaining this library.  If you are interested in contributing
to this library then please start by looking at the [golder-rules](https://github.com/neurosnap/sentences/tree/golden-rule) branch which
tests the [Golden Rules](https://github.com/diasks2/pragmatic_segmenter/blob/master/README.md#the-golden-rules)
for english sentence tokenization created by the [Pragmatic Segmenter](https://github.com/diasks2/pragmatic_segmenter)
library.

Create an issue for a particular failing test and submit an issue/PR.

I'm happy to help anyone willing to contribute.

Customizable
------------

Sentences was built around composability, most major components of this package
can be extended.

Eager to make adhoc changes but don't know how to start?
Have a look at `github.com/neurosnap/sentences/english` for a solid example.

Notice
------

I have not tested this tokenizer in any other language besides English.  By default
the command line utility loads english. I welcome anyone willing to test the
other languages to submit updates as needed.

A primary goal for this package is to be multilingual so I'm willing to help in
any way possible.

This library is a port of the [nltk's](http://www.nltk.org) punkt tokenizer.

A Punkt Tokenizer
-----------------

An unsupervised multilingual sentence boundary detection library for golang.
The way the punkt system accomplishes this goal is through training the tokenizer
with text in that given language.  Once the likelyhoods of abbreviations, collocations,
and sentence starters are determined, finding sentence boundaries becomes easier.

There are many problems that arise when tokenizing text into sentences, the primary
issue being abbreviations.  The punkt system attempts to determine whether a  word
is an abbrevation, an end to a sentence, or even both through training the system with text
in the given language.  The punkt system incorporates both token- and type-based
analysis on the text through two different phases of annotation.

[Unsupervised multilingual sentence boundary detection](http://citeseerx.ist.psu.edu/viewdoc/download;jsessionid=BAE5C34E5C3B9DC60DFC4D93B85D8BB1?doi=10.1.1.85.5017&rep=rep1&type=pdf)

Performance
-----------

Using [Brown Corpus](http://www.hit.uib.no/icame/brown/bcm.html) which is annotated American English
text, we compare this package with other libraries across multiple programming languages.

|Library    | Avg Speed (s, 10 runs) | Accuracy (%)
|:----------|:----------------------:|:-----------:
| Sentences | 1.96                   | 98.95
| NLTK      | 5.22                   | 99.21
