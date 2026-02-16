package banner

import (
	"fmt"
	"strings"
)

// Rainbow renkleri (ANSI)
var rainbow = []int{35, 34, 36, 32, 33, 31} // M,B,C,G,Y,R

// PrintRainbow ASCII art'ı gökkuşağı renkleriyle yazdırır
func PrintRainbow(ascii string) {
	lines := strings.Split(strings.TrimSpace(ascii), "\n")
	for i, line := range lines {
		for j, r := range line {
			c := rainbow[(i+j)%len(rainbow)]
			fmt.Printf("\033[%dm%c\033[0m", c, r)
		}
		fmt.Println()
	}
}

// VGBotASCII VGBot logosu
const VGBotASCII = ` __      ______  ____        _   
 \ \    / / ___|| __ )  ___ | |_ 
  \ \  / / |  _ |  _ \ / _ \| __|
   \ \/ /| |_| || |_) | (_) | |_ 
    \__/  \____||____/ \___/ \__|
      Advanced SEO Traffic Bot v3.0`
