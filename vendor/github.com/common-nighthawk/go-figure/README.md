[![Build Status](https://travis-ci.com/common-nighthawk/go-figure.svg?branch=master)](https://travis-ci.com/common-nighthawk/go-figure)

# Go Figure

## Description
Go Figure prints beautiful ASCII art from text.
It supports [FIGlet](http://www.figlet.org/) files,
and most of its features.

This package was inspired by the Ruby gem [artii](https://github.com/miketierney/artii),
but built from scratch and with a different feature set.

## Installation
`go get github.com/common-nighthawk/go-figure`

## Basic Example
```go
package main

import("github.com/common-nighthawk/go-figure")

func main() {
  myFigure := figure.NewFigure("Hello World", "", true)
  myFigure.Print()
}
```

```txt
  _   _          _   _          __        __                 _       _ 
 | | | |   ___  | | | |   ___   \ \      / /   ___    _ __  | |   __| |
 | |_| |  / _ \ | | | |  / _ \   \ \ /\ / /   / _ \  | '__| | |  / _` |
 |  _  | |  __/ | | | | | (_) |   \ V  V /   | (_) | | |    | | | (_| |
 |_| |_|  \___| |_| |_|  \___/     \_/\_/     \___/  |_|    |_|  \__,_|
```

You can also make colorful figures:

```go
func main() {
  myFigure := figure.NewColorFigure("Hello World", "", "green", true)
  myFigure.Print()
}
```

## Documentation
### Create a Figure
There are three ways to create a Figure. These are--
the method `func NewFigure`,
the method `func NewColorFigure`, and
the method `func NewFigureWithFont`.

Each constructor takes the arguments: the text, font, and strict mode.
The "color" constructor takes a color as an additional arg.
The "with font" specifies the font differently.
The method signature are:
```
func NewFigure(phrase, fontName string, strict bool) figure
func NewColorFigure(phrase, fontName string, color string, strict bool) figure
func NewFigureWithFont(phrase string, reader io.Reader, strict bool) figure
```

`NewFigure` requires only the name of the font, and uses the font file shipped
with this package stored in bindata.

If passed an empty string for the font name, a default is provided.
That is, these are both valid--

`myFigure := figure.NewFigure("Foo Bar", "", true)`

`myFigure := figure.NewFigure("Foo Bar", "alphabet", true)`

Please note that font names are case sensitive.

`NewFigureWithFont`, on the other hand, accepts the font file directly.
This allows you to BYOF (bring your own font).
Provide the absolute path to the flf.
You can point to a file the comes with this project
or you can store the file anywhere you'd like and use that location.

The font files are available in the [fonts folder](https://github.com/common-nighthawk/go-figure/tree/master/fonts)
and on [figlet.org](http://www.figlet.org/fontdb.cgi).

Here are two examples--

`myFigure := figure.NewFigureWithFont("Foo Bar", "/home/ubuntu/go/src/github.com/common-nighthawk/go-figure/fonts/alphabet.flf", true)`

`myFigure := figure.NewFigureWithFont("Foo Bar", "/home/lib/fonts/alaphabet.flf", true)`

You can also make colorful figures! The current supported colors are:
blue, cyan, gray, green, purple, red, white, yellow.

An example--

`myFigure := figure.NewColorFigure("Foo Bar", "", "green", true)

Strict mode dictates how to handle characters outside of standard ASCII.
When set to true, a non-ASCII character (outside character codes 32-127)
will cause the program to panic.
When set to false, these characters are replaced with a question mark ('?').
Examples of each--

`figure.NewFigure("Foo 👍  Bar", "alphabet", true).Print()`

```txt
2016/12/01 19:35:38 invalid input.
```

`figure.NewFigure("Foo 👍  Bar", "alphabet", false).Print()`

```txt
 _____                     ___     ____                 
 |  ___|   ___     ___     |__ \   | __ )    __ _   _ __ 
 | |_     / _ \   / _ \      / /   |  _ \   / _` | | '__|
 |  _|   | (_) | | (_) |    |_|    | |_) | | (_| | | |   
 |_|      \___/   \___/     (_)    |____/   \__,_| |_|   
```

### Methods: stdout
#### Print()
The most basic, and common, method is func Print.
A figure responds to Print(), and will write the output to the terminal.
There is no return value.

`myFigure.Print()`

But if you're feeling adventurous,
explore the methods below.

#### Blink(duration, timeOn, timeOff int)
A figure responds to the func Blink, taking three arguments.
`duration` is the total time the banner will display, in milliseconds.
`timeOn` is the length of time the text will blink on (also in ms).
`timeOff` is the length of time the text will blink off (ms).
For an even blink, set `timeOff` to -1
(same as setting `timeOff` to the value of `timeOn`).
There is no return value.

`myFigure.Blink(5000, 1000, 500)`

`myFigure.Blink(5000, 1000, -1)`

#### Scroll(duration, stillness int, direction string)
A figure responds to the func Scroll, taking three arguments.
`duration` is the total time the banner will display, in milliseconds.
`stillness` is the length of time the text will not move (also in ms).
Therefore, the lower the stillness the faster the scroll speed.
`direction` can be either "right" or "left" (case insensitive).
The direction will be left if an invalid option (e.g. "foo") is passed.
There is no return value.

`myFigure.Scroll(5000, 200, "right")`

`myFigure.Scroll(5000, 100, "left")`

#### Dance(duration, freeze int)
A figure responds to the func Dance, taking two arguments.
`duration` is the total time the banner will display, in milliseconds.
`freeze` is the length of time between dance moves (also in ms).
Therefore, the lower the freeze the faster the dancing.
There is no return value.

`myFigure.Dance(5000, 800)`

### Methods: Writers
#### Write(w io.Writer, fig figure)
Unlike the above methods that operate on a figure value,
func Write is a function that takes two arguments.
`w` is a value that implements all the methods in the io.Writer interface.
`fig` is the figure that will be written.

`figure.Write(w, myFigure)`

This method would be useful, for example, to add a nifty banner to a web page--
```go
func landingPage(w http.ResponseWriter, r *http.Request) {
  figure.Write(w, myFigure)
}
```

### Methods: Misc
#### Slicify() ([]string)
If you want to do something outside of the created methods,
you can grab the internal slice.
This gives you a good start to build anything
with the ASCII art, if manually.

A figure responds to the func Slicify,
and will return the slice of strings.

`myFigure.Slicify()`

returns

```txt
["FFFF           BBBB         ",
 "F              B   B        ",
 "FFF  ooo ooo   BBBB   aa rrr",
 "F    o o o o   B   B a a r  ",
 "F    ooo ooo   BBBB  aaa r  "]
```

## More Examples
`figure.NewFigure("Go-Figure", "isometric1", true).Print()`

```
      ___           ___           ___                       ___           ___           ___           ___     
     /\  \         /\  \         /\  \          ___        /\  \         /\__\         /\  \         /\  \    
    /::\  \       /::\  \       /::\  \        /\  \      /::\  \       /:/  /        /::\  \       /::\  \   
   /:/\:\  \     /:/\:\  \     /:/\:\  \       \:\  \    /:/\:\  \     /:/  /        /:/\:\  \     /:/\:\  \  
  /:/  \:\  \   /:/  \:\  \   /::\~\:\  \      /::\__\  /:/  \:\  \   /:/  /  ___   /::\~\:\  \   /::\~\:\  \ 
 /:/__/_\:\__\ /:/__/ \:\__\ /:/\:\ \:\__\  __/:/\/__/ /:/__/_\:\__\ /:/__/  /\__\ /:/\:\ \:\__\ /:/\:\ \:\__\
 \:\  /\ \/__/ \:\  \ /:/  / \/__\:\ \/__/ /\/:/  /    \:\  /\ \/__/ \:\  \ /:/  / \/_|::\/:/  / \:\~\:\ \/__/
  \:\ \:\__\    \:\  /:/  /       \:\__\   \::/__/      \:\ \:\__\    \:\  /:/  /     |:|::/  /   \:\ \:\__\  
   \:\/:/  /     \:\/:/  /         \/__/    \:\__\       \:\/:/  /     \:\/:/  /      |:|\/__/     \:\ \/__/  
    \::/  /       \::/  /                    \/__/        \::/  /       \::/  /       |:|  |        \:\__\    
     \/__/         \/__/                                   \/__/         \/__/         \|__|         \/__/    
```

`figure.NewFigure("Foo Bar Pop", "smkeyboard", true).Print()`

```
 ____  ____  ____  ____  ____  ____  ____  ____  ____ 
||F ||||o ||||o ||||B ||||a ||||r ||||P ||||o ||||p ||
||__||||__||||__||||__||||__||||__||||__||||__||||__||
|/__\||/__\||/__\||/__\||/__\||/__\||/__\||/__\||/__\|
```

`figure.NewFigure("Keep Your Eyes On Me", "rectangles", true).Print()`

```
                                                                                          
 _____                 __ __                 _____                 _____       _____      
|  |  | ___  ___  ___ |  |  | ___  _ _  ___ |   __| _ _  ___  ___ |     | ___ |     | ___ 
|    -|| -_|| -_|| . ||_   _|| . || | ||  _||   __|| | || -_||_ -||  |  ||   || | | || -_|
|__|__||___||___||  _|  |_|  |___||___||_|  |_____||_  ||___||___||_____||_|_||_|_|_||___|
                 |_|                               |___|                                  
```

`figure.NewFigure("ABCDEFGHIJ", "eftichess", true).Print()`

```
#########         #########   ___   #########         #########                           
##[`'`']#  \`~'/  ##'\v/`##  /\*/\  ##|`+'|##  '\v/`  ##\`~'/##  [`'`']   '\v/`    \`~'/  
###|  |##  (o o)  ##(o 0)## /(o o)\ ##(o o)##  (o 0)  ##(o o)##   |  |    (o 0)    (o o)  
###|__|##   \ / \ ###(_)###   (_)   ###(_)###   (_)   ###\ / \#   |__|     (_)      \ / \ 
#########    "    #########         #########         ####"####                      "    
```

`figure.NewFigure("Give your reasons", "doom", true).Blink(10000, 500, -1)`

![blink](docs/blink.gif "blink")

`figure.NewFigure("I mean, I could...", "basic", true).Scroll(10000, 200, "right")`

`figure.NewFigure("But why would I want to?", "basic", true).Scroll(10000, 200, "left")`

![scroll](docs/scroll.gif "scroll")

`figure.NewFigure("It's been waiting for you", "larry3d", true).Dance(10000, 500)`

![dance](docs/dance.gif "dance")

`figure.Write(w, figure.NewFigure("Hello, It's Me", "puffy", true))`

![web](docs/web.png "web")


## Supported Fonts
* 3-d
* 3x5
* 5lineoblique
* acrobatic
* alligator
* alligator2
* alphabet
* avatar
* banner
* banner3-D
* banner3
* banner4
* barbwire
* basic
* bell
* big
* bigchief
* binary
* block
* bubble
* bulbhead
* calgphy2
* caligraphy
* catwalk
* chunky
* coinstak
* colossal
* computer
* contessa
* contrast
* cosmic
* cosmike
* cricket
* cursive
* cyberlarge
* cybermedium
* cybersmall
* diamond
* digital
* doh
* doom
* dotmatrix
* drpepper
* eftichess
* eftifont
* eftipiti
* eftirobot
* eftitalic
* eftiwall
* eftiwater
* epic
* fender
* fourtops
* fuzzy
* goofy
* gothic
* graffiti
* hollywood
* invita
* isometric1
* isometric2
* isometric3
* isometric4
* italic
* ivrit
* jazmine
* jerusalem
* katakana
* kban
* larry3d
* lcd
* lean
* letters
* linux
* lockergnome
* madrid
* marquee
* maxfour
* mike
* mini
* mirror
* mnemonic
* morse
* moscow
* nancyj-fancy
* nancyj-underlined
* nancyj
* nipples
* ntgreek
* o8
* ogre
* pawp
* peaks
* pebbles
* pepper
* poison
* puffy
* pyramid
* rectangles
* relief
* relief2
* rev
* roman
* rot13
* rounded
* rowancap
* rozzo
* runic
* runyc
* sblood
* script
* serifcap
* shadow
* short
* slant
* slide
* slscript
* small
* smisome1
* smkeyboard
* smscript
* smshadow
* smslant
* smtengwar
* speed
* stampatello
* standard
* starwars
* stellar
* stop
* straight
* tanja
* tengwar
* term
* thick
* thin
* threepoint
* ticks
* ticksslant
* tinker-toy
* tombstone
* trek
* tsalagi
* twopoint
* univers
* usaflag
* wavy
* weird

## Contributing
Because this project is small, we can dispense with formality.
Submit a pull request, open an issue, request a change.
All good!

## Wanna Say Thanks?
GitHub stars are helpful.
Most importantly, they help with discoverability.
Projects with more stars are displayed higher
in search results when people are looking for packages.
Also--they make contributors feel good :)

If you are feeling especially generous,
give a shout to [@cmmn_nighthawk](https://twitter.com/cmmn_nighthawk).

## TODO
* Add proper support for spaces
* More animations
* Implement graceful line-wrapping and smushing
* Deep-copy font for Dance (current implementation is destructive)
