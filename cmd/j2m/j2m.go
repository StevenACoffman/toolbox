///usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)


func main() {

	str := ""
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (stat.Mode() & os.ModeNamedPipe) == 0 {
		fmt.Println("The command is intended to work with pipes but didn't get one. Assuming empty input")
	} else {
		stdInBytes, _ := ioutil.ReadAll(os.Stdin)
		str = string(stdInBytes)
	}

	// UnOrdered Lists
	re1 := regexp.MustCompile(`(?m)^[ \t]*(\*+)\s+`)
	first := ReplaceAllStringSubmatchFunc(re1, str, func(groups []string) string {
		_, stars := groups[0], groups[1]
		return strings.Repeat("  ", len(stars)-1) + "* "
	})
	// Ordered Lists
	re2 := regexp.MustCompile(`(?m)^[ \t]*(#+)\s+`)
	second := ReplaceAllStringSubmatchFunc(re2, first, func(groups []string) string {
		_, nums := groups[0], groups[1]
		return strings.Repeat("  ", len(nums)-1) + "1. "
	})
	// Headers 1-6
	re3 := regexp.MustCompile(`(?m)^h([0-6])\.(.*)$`)
	third := ReplaceAllStringSubmatchFunc(re3, second, func(groups []string) string {
		_, level, content := groups[0], groups[1], groups[2]
		i, _ := strconv.Atoi(level)
		return strings.Repeat("#", i) + content
	})
	// Bold
	re4 := regexp.MustCompile(`\*(\S.*)\*`)
	fourth := re4.ReplaceAllString(third, "**$1**")
	// Italic
	re5 := regexp.MustCompile(`\_(\S.*)\_`)
	fifth := re5.ReplaceAllString(fourth, "*$1*")
	// Monospaced text
	re6 := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	sixth := re6.ReplaceAllString(fifth, "`$1`")
	// Citations (buggy)
	re7 := regexp.MustCompile(`\?\?((?:.[^?]|[^?].)+)\?\?`)
	seventh := re7.ReplaceAllString(sixth, "<cite>$1</cite>")

	// Inserts
	re8 := regexp.MustCompile(`\+([^+]*)\+`)
	eighth := re8.ReplaceAllString(seventh, "<ins>$1</ins>")
	// Superscript
	re9:= regexp.MustCompile(`\^([^^]*)\^`)
	ninth := re9.ReplaceAllString(eighth, "<sup>$1</sup>")

	// Subscript
	re10:= regexp.MustCompile(`~([^~]*)~`)
	tenth := re10.ReplaceAllString(ninth, "<sub>$1</sub>")
	// Strikethrough
	re11:= regexp.MustCompile(`(\s+)-(\S+.*?\S)-(\s+)`)
	eleventh := re11.ReplaceAllString(tenth, "$1~~$2~~$3")
	// Code Block
	re12:= regexp.MustCompile(`\{code(:([a-z]+))?([:|]?(title|borderStyle|borderColor|borderWidth|bgColor|titleBGColor)=.+?)*\}`)
	twelfth := re12.ReplaceAllString(eleventh, "```$2")

	re13:= regexp.MustCompile(`{code}`)
	thirteenth := re13.ReplaceAllString(twelfth, "```")
     // Pre-formatted text
    re14:= regexp.MustCompile(`{noformat}`)
	fourteenth := re14.ReplaceAllString(thirteenth, "```")

	// Un-named Links
	re15:= regexp.MustCompile(`\[([^|]+)\]`)
	fifteenth := re15.ReplaceAllString(fourteenth, "<$1>")


	// Images
	re16:= regexp.MustCompile(`!(.+)!`)
	sixteenth := re16.ReplaceAllString(fifteenth, "![]($1)")
	// Named Links
	re17:= regexp.MustCompile(`\[(.+?)\|(.+)\]`)
	seventeenth := re17.ReplaceAllString(sixteenth, "[$1]($2)")

	// Single Paragraph Blockquote
	re18:= regexp.MustCompile(`(?m)^bq\.\s+`)
	eighteenth := re18.ReplaceAllString(seventeenth, "> ")

	// Remove color: unsupported in md
	re19:= regexp.MustCompile(`(?m)\{color:[^}]+\}(.*)\{color\}`)
	nineteenth := re19.ReplaceAllString(eighteenth, "$1")

	// panel into table
	re20:= regexp.MustCompile(`(?m)\{panel:title=([^}]*)\}\n?(.*?)\n?\{panel\}`)
	twentieth := re20.ReplaceAllString(nineteenth, "\n| $1 |\n| --- |\n| $2 |")

	// table header
	re21:= regexp.MustCompile(`(?m)^[ \t]*((?:\|\|.*?)+\|\|)[ \t]*$`)
	twentyfirst := ReplaceAllStringSubmatchFunc(re21, twentieth, func(groups []string) string {
		_, headers := groups[0], groups[1]
		reBarred:= regexp.MustCompile(`\|\|`)

		singleBarred :=  reBarred.ReplaceAllString(headers, "|")
		fillerRe := regexp.MustCompile(`\|[^|]+`)
		return "\n" + singleBarred + "\n" + fillerRe.ReplaceAllString(singleBarred, "| --- ")
	})

	// remove leading-space of table headers and rows
	re22:= regexp.MustCompile(`(?m)^[ \t]*\|`)
	twentysecond := re22.ReplaceAllString(twentyfirst, "|")


	fmt.Printf("%s\n", twentysecond)
}

// https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}