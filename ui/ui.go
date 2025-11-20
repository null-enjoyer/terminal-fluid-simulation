package ui

import "github.com/gdamore/tcell/v2"

type MenuItem struct {
	Name string
	Type string // float, int, enum, action
	Val  any    // pointer to the config value, nil for action
	Step float64
	Fmt  string
}

func DrawText(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for i, c := range str {
		s.SetContent(x+i, y, c, nil, style)
	}
}

func DrawBox(s tcell.Screen, x, y, w, h int, style tcell.Style) {
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			s.SetContent(x+i, y+j, ' ', nil, style)
		}
	}
}

func DrawInputOverlay(s tcell.Screen, title, currentText string) {
	w, h := s.Size()
	boxW, boxH := 40, 5
	x, y := (w-boxW)/2, (h-boxH)/2

	style := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	DrawBox(s, x, y, boxW, boxH, style)
	DrawBox(s, x+1, y+1, boxW-2, boxH-2, style)
	DrawText(s, x+2, y+1, style, title)

	inputStyle := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	DrawBox(s, x+2, y+3, boxW-4, 1, inputStyle)
	DrawText(s, x+2, y+3, inputStyle, currentText+"_")
}
