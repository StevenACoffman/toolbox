package main

import (
	"fmt"
	"github.com/StevenACoffman/toolbox/cmd/colory/colors"
)

func main() {
	// Basic ANSI colors 0 - 15
	fmt.Println(colors.String("Basic ANSI colors").Bold())


	for i := int64(0); i < 16; i++ {
		if i%8 == 0 {
			fmt.Println()
		}

		// background color
		bg := colors.ANSIColor(i)
		c := colors.ConvertToRGB(bg)

		out := colors.String(fmt.Sprintf(" %2d %s ", i, c.Hex()))

		// apply colors
		if i < 5 {
			out = out.Foreground(colors.ANSIColor(7))
		} else {
			out = out.Foreground(colors.ANSIColor(0))
		}
		out = out.Background(bg)

		fmt.Print(out.String()[:])
	}
	fmt.Printf("\n\n")

	fmt.Printf("\n\n")

	fmt.Printf("\n\n")
}
