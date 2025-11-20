package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/null-enjoyer/terminal-fluid-simulation/config"
	"github.com/null-enjoyer/terminal-fluid-simulation/simulation"
	"log"
)

func (a *App) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btn := ev.Buttons()

	simX := x - simulation.SidebarWidth
	simY := y

	if simX >= 0 && simX < a.SimW && simY >= 0 && simY < a.SimH {
		a.CursorX = float64(simX)
		a.CursorY = float64(simY)
		a.MouseInBounds = true
	} else {
		a.MouseInBounds = false
	}

	if btn&tcell.Button1 != 0 {
		a.IsMouseDown = true
	} else {
		a.IsMouseDown = false
	}
}

func (a *App) HandleInput(ev *tcell.EventKey) bool {
	if a.InputMode {
		switch ev.Key() {
		case tcell.KeyEnter:
			name := a.InputText
			if name == "" {
				name = "Custom"
			}

			a.UIConfig.PaletteName = a.Palettes[a.UIConfig.PaletteIdx].Name
			a.AppConfig.Presets[name] = a.UIConfig

			if err := config.SaveSettings(a.ConfigPath, a.AppConfig); err != nil {
				log.Fatalf("Error saving settings (%s): %v", a.ConfigPath, err)
			}

			a.PresetNames = config.GetSortedPresetNames(a.AppConfig.Presets)
			a.ActivePresetName = name

			for i, n := range a.PresetNames {
				if n == a.ActivePresetName {
					a.ActivePresetIdx = i
					break
				}
			}

			a.ForceRedraw()
			a.InputMode = false
			a.InputText = ""
		case tcell.KeyEscape:
			a.InputMode = false
			a.InputText = ""
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(a.InputText) > 0 {
				a.InputText = a.InputText[:len(a.InputText)-1]
			}
		default:
			if ev.Rune() != 0 {
				a.InputText += string(ev.Rune())
			}
		}
		return false
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		return true
	case tcell.KeyEnter:
		item := a.MenuItems[a.SelectedItem]
		if item.Type == "action" && item.Name == "Save" {
			a.InputMode = true
			a.InputText = ""
		}
	case tcell.KeyTab:
		a.cycleMouseMode()
	case tcell.KeyLeft:
		a.CursorX -= 2.0
	case tcell.KeyRight:
		a.CursorX += 2.0
	case tcell.KeyUp:
		a.CursorY -= 1.0
	case tcell.KeyDown:
		a.CursorY += 1.0
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'm', 'M':
			a.cycleMouseMode()
		case 'q':
			return true
		case 'r':
			a.Sim.CmdChan <- func(s *simulation.Simulation) {
				s.Particles = s.Particles[:0]
			}
			a.ForceRedraw()
		case 'c', 'C':
			a.Sim.CmdChan <- func(s *simulation.Simulation) {
				for i := range s.Walls {
					s.Walls[i] = false
				}
			}
			a.ForceRedraw()
		case 'p', 'P':
			a.UIConfig.IsPaused = !a.UIConfig.IsPaused
			newCfg := a.UIConfig
			a.Sim.CmdChan <- func(s *simulation.Simulation) { s.Config = newCfg }
		case 'w', 'W':
			a.SelectedItem--
			if a.SelectedItem < 0 {
				a.SelectedItem = len(a.MenuItems) - 1
			}
		case 's', 'S':
			a.SelectedItem++
			if a.SelectedItem >= len(a.MenuItems) {
				a.SelectedItem = 0
			}
		case 'a', 'A':
			a.handleTweak(-1.0)
		case 'd', 'D':
			a.handleTweak(1.0)
		case ' ':
			cx, cy := a.CursorX, a.CursorY
			a.Sim.CmdChan <- func(s *simulation.Simulation) { s.Spawn(cx, cy) }
		}
	}

	// cursor bound checks
	if a.CursorX < 0 {
		a.CursorX = 0
	}
	if a.CursorX >= float64(a.SimW) {
		a.CursorX = float64(a.SimW - 1)
	}
	if a.CursorY < 0 {
		a.CursorY = 0
	}
	if a.CursorY >= float64(a.SimH) {
		a.CursorY = float64(a.SimH - 1)
	}

	return false
}

func (a *App) cycleMouseMode() {
	a.MouseMode++
	if a.MouseMode > ModeErase {
		a.MouseMode = ModeSpawn
	}
}

func (a *App) handleTweak(delta float64) {
	item := a.MenuItems[a.SelectedItem]
	isCustomizing := false

	switch item.Type {
	case "preset_enum":
		a.ActivePresetIdx += int(delta)
		if a.ActivePresetIdx < 0 {
			a.ActivePresetIdx = len(a.PresetNames) - 1
		}
		if a.ActivePresetIdx >= len(a.PresetNames) {
			a.ActivePresetIdx = 0
		}
		a.ActivePresetName = a.PresetNames[a.ActivePresetIdx]

		newP := a.AppConfig.Presets[a.ActivePresetName]
		a.UIConfig = newP
		a.SyncPalette()
		a.InitMenu()
		a.ForceRedraw()

	case "float":
		val := item.Val.(*float64)
		*val += delta * item.Step
		isCustomizing = true
	case "int":
		val := item.Val.(*int)
		*val += int(delta * item.Step)
		if *val < 1 {
			*val = 1
		}
		isCustomizing = true
	case "enum":
		val := item.Val.(*int)
		*val += int(delta)
		if *val < 0 {
			*val = len(a.Palettes) - 1
		}
		if *val >= len(a.Palettes) {
			*val = 0
		}
		isCustomizing = true
		a.ForceRedraw()
	}

	if isCustomizing {
		a.ActivePresetName = "Custom"
	}

	a.UIConfig.UpdateDerived()
	newCfg := a.UIConfig
	a.Sim.CmdChan <- func(s *simulation.Simulation) { s.Config = newCfg }
}
