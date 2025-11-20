package render

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/null-enjoyer/terminal-fluid-simulation/config"
)

type Palette struct {
	Name   string
	Colors []tcell.Color
}

func ParsePalettes(hexPalettes []config.HexPalette) []Palette {
	var palettes []Palette
	for _, r := range hexPalettes {
		p := Palette{Name: r.Name}
		for _, hex := range r.Colors {
			val, _ := strconv.ParseInt(cleanHex(hex), 16, 32)
			p.Colors = append(p.Colors, tcell.NewHexColor(int32(val)))
		}
		palettes = append(palettes, p)
	}
	return palettes
}

func cleanHex(s string) string {
	if len(s) > 0 && s[0] == '#' {
		return s[1:]
	}
	return s
}
