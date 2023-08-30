package utils

import "fmt"

func CalcPercStr(value int, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%%)", int(100*float64(value)/float64(total)))
}

func CalcDiffPercStr(curValue int, prevValue int) string {
	diff := curValue - prevValue
	if prevValue == 0 {
		return "       "
	}
	sig := ""
	if diff > 0 {
		sig = "+"
	}
	return fmt.Sprintf(" (%s%d%%)", sig, int(100*float64(diff)/float64(prevValue)))
}

func CalcDiffStr(curValue int, prevValue int) string {
	diff := curValue - prevValue
	if prevValue == 0 {
		return "       "
	}
	sig := ""
	if diff > 0 {
		sig = "+"
	}
	return fmt.Sprintf(" (%s%d)", sig, diff)
}
