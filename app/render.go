package app

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/null-enjoyer/terminal-fluid-simulation/simulation"
	"github.com/null-enjoyer/terminal-fluid-simulation/ui"
)

func (a *App) Render() {
	const wallID = 999

	renderStart := time.Now()
	a.FpsCounter++
	if time.Since(a.FpsTimer) >= time.Second {
		a.Fps = a.FpsCounter
		a.FpsCounter = 0
		a.FpsTimer = time.Now()
	}

	// reset grid
	for x := 0; x < a.SimW; x++ {
		for y := 0; y < a.SimH; y++ {
			a.CurrentGrid[x][y] = 0
		}
	}

	// mark walls
	wallLen := len(a.Sim.Walls)
	if wallLen == a.SimW*a.SimH {
		for x := 0; x < a.SimW; x++ {
			for y := 0; y < a.SimH; y++ {
				if a.Sim.Walls[x+y*a.SimW] {
					a.CurrentGrid[x][y] = wallID
				}
			}
		}
	}

	// add particles to grid
	for _, p := range a.CurrentParticles {
		if p.X >= 0 && p.X < a.SimW && p.Y >= 0 && p.Y < a.SimH {
			if a.CurrentGrid[p.X][p.Y] != wallID {
				a.CurrentGrid[p.X][p.Y] += 3
				if p.X+1 < a.SimW && a.CurrentGrid[p.X+1][p.Y] != wallID {
					a.CurrentGrid[p.X+1][p.Y] += 1
				}
				if p.X-1 >= 0 && a.CurrentGrid[p.X-1][p.Y] != wallID {
					a.CurrentGrid[p.X-1][p.Y] += 1
				}
				if p.Y+1 < a.SimH && a.CurrentGrid[p.X][p.Y+1] != wallID {
					a.CurrentGrid[p.X][p.Y+1] += 1
				}
			}
		}
	}

	palette := a.Palettes[a.UIConfig.PaletteIdx].Colors

	for x := 0; x < a.SimW; x++ {
		for y := 0; y < a.SimH; y++ {
			newVal := a.CurrentGrid[x][y]
			oldVal := a.LastGrid[x][y]

			isCursor := int(a.CursorX) == x && int(a.CursorY) == y
			wasCursor := a.LastCursorX == x && a.LastCursorY == y

			if newVal != oldVal || isCursor || wasCursor {
				screenX := x + simulation.SidebarWidth
				screenY := y

				if isCursor {
					cursorChar := '▼'
					color := tcell.ColorRed
					if a.MouseMode == ModeWall {
						cursorChar = '■'
						color = tcell.ColorGray
					} else if a.MouseMode == ModeErase {
						cursorChar = 'X'
						color = tcell.ColorRed
					}
					a.Screen.SetContent(screenX, screenY, cursorChar, nil, tcell.StyleDefault.Foreground(color))
				} else if newVal == 999 {
					a.Screen.SetContent(screenX, screenY, '█', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
				} else if newVal > 0 {
					idx := newVal / 2
					if idx >= len(palette) {
						idx = len(palette) - 1
					}
					a.Screen.SetContent(screenX, screenY, '█', nil, tcell.StyleDefault.Foreground(palette[idx]))
				} else {
					a.Screen.SetContent(screenX, screenY, ' ', nil, tcell.StyleDefault)
				}
				a.LastGrid[x][y] = newVal
			}
		}
	}

	a.LastCursorX = int(a.CursorX)
	a.LastCursorY = int(a.CursorY)

	a.DrawMenu()
	a.Screen.Show()
	a.LastRenderTime = time.Since(renderStart)
}

func (a *App) DrawMenu() {
	_, h := a.Screen.Size()
	for y := 0; y < h; y++ {
		a.Screen.SetContent(simulation.SidebarWidth-1, y, '│', nil, a.StyleBorder)
	}
	ui.DrawBox(a.Screen, 0, 0, simulation.SidebarWidth-1, h, a.StyleMenuBg)

	ui.DrawText(a.Screen, 2, 1, a.StyleMenuBg, "FLUID SIMULATION")
	ui.DrawText(a.Screen, 2, 2, a.StyleMenuBg, "----------------")

	yPos := 4
	for i, item := range a.MenuItems {
		style := a.StyleMenuBg
		prefix := " "
		if i == a.SelectedItem {
			style = a.StyleMenuSel
			prefix = ">"
		}

		valStr := ""
		switch item.Type {
		case "preset_enum":
			valStr = a.ActivePresetName
		case "float":
			valStr = fmt.Sprintf(item.Fmt, *item.Val.(*float64))
		case "int":
			valStr = fmt.Sprintf(item.Fmt, *item.Val.(*int))
		case "enum":
			idx := *item.Val.(*int)
			valStr = fmt.Sprintf(item.Fmt, a.Palettes[idx].Name)
		case "action":
			valStr = item.Fmt
		}

		line := fmt.Sprintf("%s %-9s %s", prefix, item.Name, valStr)
		ui.DrawText(a.Screen, 2, yPos, style, line)
		yPos++
	}

	yPos += 2
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, "CONTROLS:")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [Tab] Cycle Mode")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [Mouse LB] Spawn/Draw/Erase")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [ C ] Clear Walls")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [Space] Spawn Fluid")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [ P ] Pause")
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, " [ R ] Clear Fluid")

	yPos += 2
	status := "RUNNING"
	if a.UIConfig.IsPaused {
		status = "PAUSED"
	}
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, fmt.Sprintf("Status: %s", status))

	yPos++
	modeStr := "SPAWN"
	if a.MouseMode == ModeWall {
		modeStr = "WALLS"
	} else if a.MouseMode == ModeErase {
		modeStr = "ERASE"
	}

	ui.DrawText(a.Screen, 2, yPos, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow), fmt.Sprintf("MODE:   %s", modeStr))
	yPos += 2
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, fmt.Sprintf("Particles: %d", len(a.CurrentParticles)))
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, fmt.Sprintf("FPS: %d", a.Fps))
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, fmt.Sprintf("Physics: %v", a.LastPhysTime.Round(time.Microsecond)))
	yPos++
	ui.DrawText(a.Screen, 2, yPos, a.StyleMenuBg, fmt.Sprintf("Render: %v", a.LastRenderTime.Round(time.Microsecond)))

	if a.InputMode {
		ui.DrawInputOverlay(a.Screen, "Save Preset As:", a.InputText)
	}
}

func (a *App) ForceRedraw() {
	a.Screen.Clear()
	for i := range a.LastGrid {
		for j := range a.LastGrid[i] {
			a.LastGrid[i][j] = -1
		}
	}
}
