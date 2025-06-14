package termenv

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/xo/terminfo"
)

// ColorLevel is the color level supported by a terminal.
type ColorLevel uint8

const (
	TermColorNone ColorLevel = iota // not support color
	TermColor16                     // 16(4bit) ANSI color supported
	TermColor256                    // 256(8bit) color supported
	TermColorTrue                   // support TRUE(RGB) color
)

// String returns the string name of the color level.
func (l ColorLevel) String() string {
	switch l {
	case TermColor16:
		return "ansi"
	case TermColor256:
		return "256"
	case TermColorTrue:
		return "true"
	default:
		return "none"
	}
}

// NoColor returns true if the NO_COLOR environment variable is set.
func NoColor() bool { return noColor }

// IsSupportColor returns true if the terminal supports color.
func IsSupportColor() bool { return colorLevel != TermColorNone }

// IsSupport256Color returns true if the terminal supports 256 colors.
func IsSupport256Color() bool { return colorLevel >= TermColor256 }

// IsSupportTrueColor returns true if the terminal supports true color.
func IsSupportTrueColor() bool { return colorLevel == TermColorTrue }

//
// ---------------- for testing ----------------
//

var backOldVal bool

// ForceEnableColor setting value. TIP: use for unit testing.
//
// Usage:
//
//	ccolor.ForceEnableColor()
//	defer ccolor.RevertColorSupport()
func ForceEnableColor() {
	noColor = false
	backOldVal = supportColor
	// force enables color
	supportColor = true
	// return colorLevel
}

// RevertColorSupport value
func RevertColorSupport() {
	supportColor = backOldVal
	noColor = os.Getenv("NO_COLOR") == ""
}

/*************************************************************
 * helper methods for detect color supports
 *************************************************************/

// DetectColorLevel for current env
//
// NOTICE: The method will detect terminal info each time.
//
//	if only want to get current color level, please direct call IsSupportColor() or TermColorLevel()
func DetectColorLevel() ColorLevel {
	level, _ := detectTermColorLevel()
	return level
}

// on TERM=screen: not support true-color
const noTrueColorTerm = "screen"

// detect terminal color support level
//
// refer https://github.com/Delta456/box-cli-maker
func detectTermColorLevel() (level ColorLevel, needVTP bool) {
	isWin := runtime.GOOS == "windows"
	termVal := os.Getenv("TERM")

	if termVal != noTrueColorTerm {
		// On JetBrains Terminal
		// - TERM value not set, but support true-color
		// env:
		// 	TERMINAL_EMULATOR=JetBrains-JediTerm
		val := os.Getenv("TERMINAL_EMULATOR")
		if val == "JetBrains-JediTerm" {
			debugf("True Color support on JetBrains-JediTerm, is win: %v", isWin)
			return TermColorTrue, false
		}
	}

	level = detectColorLevelFromEnv(termVal, isWin)
	debugf("color level by detectColorLevelFromEnv: %s", level.String())

	// fallback: simply detect by TERM value string.
	if level == TermColorNone {
		debugf("level none - fallback check special term color support")
		// on Windows: enable VTP as it has True Color support
		level, needVTP = detectSpecialTermColor(termVal)
	}
	return
}

// detectColorFromEnv returns the color level COLORTERM, FORCE_COLOR,
// TERM_PROGRAM, or determined from the TERM environment variable.
//
// refer the terminfo.ColorLevelFromEnv()
// https://en.wikipedia.org/wiki/Terminfo
func detectColorLevelFromEnv(termVal string, isWin bool) ColorLevel {
	if termVal == noTrueColorTerm { // on TERM=screen: not support true-color
		return TermColor256
	}

	// check for overriding environment variables
	colorTerm, termProg, forceColor := os.Getenv("COLORTERM"), os.Getenv("TERM_PROGRAM"), os.Getenv("FORCE_COLOR")
	switch {
	case strings.Contains(colorTerm, "truecolor") || strings.Contains(colorTerm, "24bit"):
		return TermColorTrue
	case colorTerm != "" || forceColor != "":
		return TermColor16
	case termProg == "Apple_Terminal":
		return TermColor256
	case termProg == "Terminus" || termProg == "Hyper":
		return TermColorTrue
	case termProg == "iTerm.app":
		// check iTerm version
		termVer := os.Getenv("TERM_PROGRAM_VERSION")
		if termVer != "" {
			i, err := strconv.Atoi(strings.Split(termVer, ".")[0])
			if err != nil {
				setLastErr(terminfo.ErrInvalidTermProgramVersion)
				return TermColor256 // return TermColorNone
			}
			if i == 3 {
				return TermColorTrue
			}
		}
		return TermColor256
	}

	// otherwise determine from TERM's max_colors capability
	if !isWin && termVal != "" {
		debugf("TERM=%s - TODO check color level by load terminfo file", termVal)
		return TermColor16
	}

	// no TERM env value. default return none level
	return TermColorNone
}

var (
	detectedWSL bool
	wslContents string
)

// https://github.com/Microsoft/WSL/issues/423#issuecomment-221627364
func detectWSL() bool {
	if !detectedWSL {
		detectedWSL = true

		b := make([]byte, 1024)
		// `cat /proc/version`
		// on mac:
		// 	!not the file!
		// on linux(debian,ubuntu,alpine):
		//	Linux version 4.19.121-linuxkit (root@18b3f92ade35) (gcc version 9.2.0 (Alpine 9.2.0)) #1 SMP Thu Jan 21 15:36:34 UTC 2021
		// on win git bash, conEmu:
		// 	MINGW64_NT-10.0-19042 version 3.1.7-340.x86_64 (@WIN-N0G619FD3UK) (gcc version 9.3.0 (GCC) ) 2020-10-23 13:08 UTC
		// on WSL:
		//  Linux version 4.4.0-19041-Microsoft (Microsoft@Microsoft.com) (gcc version 5.4.0 (GCC) ) #488-Microsoft Mon Sep 01 13:43:00 PST 2020
		f, err := os.Open("/proc/version")
		if err == nil {
			_, _ = f.Read(b) // ignore error
			if err = f.Close(); err != nil {
				setLastErr(err)
			}

			wslContents = string(b)
			return strings.Contains(wslContents, "Microsoft")
		}
	}
	return false
}
