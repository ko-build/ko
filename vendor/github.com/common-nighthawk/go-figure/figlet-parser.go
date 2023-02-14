package figure

import (
	"strconv"
	"strings"
)

const signature = "flf2"
const reverseFlag = "1"

var charDelimiters = [3]string{"@", "#", "$"}
var hardblanksBlacklist = [2]byte{'a', '2'}

func getHeight(metadata string) int {
	datum := strings.Fields(metadata)[1]
	height, _ := strconv.Atoi(datum)
	return height
}

func getBaseline(metadata string) int {
	datum := strings.Fields(metadata)[2]
	baseline, _ := strconv.Atoi(datum)
	return baseline
}

func getHardblank(metadata string) byte {
	datum := strings.Fields(metadata)[0]
	hardblank := datum[len(datum)-1]
	if hardblank == hardblanksBlacklist[0] || hardblank == hardblanksBlacklist[1] {
		return ' '
	} else {
		return hardblank
	}
}

func getReverse(metadata string) bool {
	data := strings.Fields(metadata)
	return len(data) > 6 && data[6] == reverseFlag
}

func lastCharLine(text string, height int) bool {
	endOfLine, length := "  ", 2
	if height == 1 && len(text) > 0 {
		length = 1
	}
	if len(text) >= length {
		endOfLine = text[len(text)-length:]
	}
	return endOfLine == strings.Repeat(charDelimiters[0], length) ||
		endOfLine == strings.Repeat(charDelimiters[1], length) ||
		endOfLine == strings.Repeat(charDelimiters[2], length)
}
