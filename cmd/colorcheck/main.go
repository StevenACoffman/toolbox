package main

 import (
 	"fmt"
 	"image/color"
 	"math"
 	"os"
 	"strconv"
 )

 // find out if the two colors are the same
 func SameColor(color1, color2 color.RGBA) bool {
 	red1, green1, blue1, alpha1 := color1.RGBA()
 	red2, green2, blue2, alpha2 := color2.RGBA()

 	return red1 == red2 && green1 == green2 && blue1 == blue2 && alpha1 == alpha2
 }

// Luminance [0..255] of a given color - ignoring alpha
// We will use the formula taken from https://www.w3.org/TR/AERT#color-contrast
// Color brightness is determined by the following formula:
// ((Red value X 299) + (Green value X 587) + (Blue value X 114)) / 1000
 func Luminance(aColor color.RGBA) float64 {
 	fmt.Printf("Color %+v\n", aColor)

 	// because of
 	// http://stackoverflow.com/questions/35374300/why-does-golang-rgba-rgba-method-use-and
 	// we cannot use .RGBA() values. It will screw up the final luminance result
 	// red, green, blue, _ := aColor.RGBA()

 	//fmt.Println("red : ", red)
 	//fmt.Println("green : ", green)
 	//fmt.Println("blue : ", blue)

 	// instead, we will use the colors value from 0..255 for calculation purpose
 	red := float64(aColor.R)
 	green := float64(aColor.G)
 	blue := float64(aColor.B)

 	// need to convert uint32 to float64
 	return float64(float64(0.299)*red + float64(0.587)*green + float64(0.114)*blue)
 }

 // Compatible is the color Contrast(brightness) diff big enough?
 func Compatible(color1, color2 color.RGBA) bool {
	 fmt.Printf("1:%f\n", Luminance(color1))
	 fmt.Printf("2:%f\n", Luminance(color2))
 	ldiff := math.Abs(Luminance(color1)-Luminance(color2))
 	fmt.Printf("%f\n", ldiff)
 	return math.Abs(Luminance(color1)-Luminance(color2)) >= 128.0
 }

 func SaneColorCheck(arg string) int {
 	colorCode, err := strconv.Atoi(arg)
 	if err != nil {
 		fmt.Println("Input arguments must be integer from 0..255")
 		os.Exit(-1)
 	}

 	if (colorCode > 255) || (colorCode < 0) {
 		fmt.Println("Invalid color code : ", colorCode)
 		fmt.Println("Color values must be in range of 0..255")
 		os.Exit(-1)
 	}
 	return colorCode
 }

func ParseHexColor(s string) (c color.RGBA) {
	var err error
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 6:
		_, err = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	case 3:
		_, err = fmt.Sscanf(s, "%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default: // default to white
		c.R = 0xff
		c.G = 0xff
		c.B = 0xff
	}
	if err != nil { // default to white
		c.R = 0xff
		c.G = 0xff
		c.B = 0xff
	}
	return c
}

 func main() {
 	if len(os.Args) != 3 {
 		fmt.Println(len(os.Args))
 		fmt.Printf("Usage : %s <hex1> <hex2>\n", os.Args[0])
 		os.Exit(0)
 	}

	 hex1 := ParseHexColor(os.Args[1])
	 hex2 := ParseHexColor(os.Args[2])

 	//sanity checks
 	red1, green1, blue1, alpha1 := hex1.RGBA()
 	red2, green2, blue2, alpha2 := hex2.RGBA()

 	fmt.Println("Red 1 : ", red1)
 	fmt.Println("Green 1 : ", green1)
 	fmt.Println("Blue 1 : ", blue1)
 	fmt.Println("Alpha 1 : ", alpha1)

 	fmt.Println("Red 2 : ", red2)
 	fmt.Println("Green 2 : ", green2)
 	fmt.Println("Blue 2 : ", blue2)
 	fmt.Println("Alpha 2 : ", alpha2)

 	// create new colors from given arguments/command line parameters
 	color1 := color.RGBA{uint8(red1), uint8(green1), uint8(blue1), uint8(alpha1)}
 	color2 := color.RGBA{uint8(red2), uint8(green2), uint8(blue2), uint8(alpha2)}

 	fmt.Println("Color 1 = ", color1)
 	fmt.Println("Color 2 = ", color2)
 	fmt.Println("Color 1 same as Color 2 ? : ", SameColor(color1, color2))
 	fmt.Println("Luminance of color 1 = ", Luminance(color1))
 	fmt.Println("Luminance of color 2 = ", Luminance(color2))
 	fmt.Println("Is color 1 and color 2 compatible? : ", Compatible(color1, color2))

 }