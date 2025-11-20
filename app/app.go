package app

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/null-enjoyer/terminal-fluid-simulation/config"
	"github.com/null-enjoyer/terminal-fluid-simulation/render"
	"github.com/null-enjoyer/terminal-fluid-simulation/simulation"
	"github.com/null-enjoyer/terminal-fluid-simulation/ui"
)

// mouse modes
const (
	ModeSpawn = iota
	ModeWall
	ModeErase
)

type App struct {
	Screen    tcell.Screen
	Sim       *simulation.Simulation
	AppConfig *config.AppConfig
	Palettes  []render.Palette

	ConfigPath string
	UIConfig   config.PhysicsConfig

	MenuItems   []ui.MenuItem
	PresetNames []string

	// Input
	CursorX, CursorY float64
	SelectedItem     int
	ActivePresetName string
	ActivePresetIdx  int
	InputMode        bool
	InputText        string

	// Mouse State
	MouseMode     int
	IsMouseDown   bool
	MouseInBounds bool

	// Particles
	CurrentParticles []render.Point
	CurrentGrid      [][]int
	LastGrid         [][]int
	SimW, SimH       int
	LastCursorX      int
	LastCursorY      int

	// UI Styling
	StyleBorder  tcell.Style
	StyleMenuBg  tcell.Style
	StyleMenuSel tcell.Style

	// Debug Info
	Fps            int
	FpsCounter     int
	FpsTimer       time.Time
	LastPhysTime   time.Duration
	LastRenderTime time.Duration
}

func New(configPath string) *App {
	var appConfig *config.AppConfig
	var err error

	if configPath != "" {
		appConfig, err = config.LoadSettings(configPath)
		if err != nil {
			log.Fatalf("Failed to load settings: %v", err)
		}
	} else {
		appConfig = config.NewDefaultConfig()
	}

	if len(appConfig.Palettes) == 0 {
		log.Fatal("No palettes found")
	}
	palettes := render.ParsePalettes(appConfig.Palettes)

	defaultCfg, ok := appConfig.Presets["Default"]
	if !ok {
		log.Fatal("Default config not found in settings.json")
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	screen.EnableMouse()

	w, h := screen.Size()
	sim := simulation.NewSimulation(w, h, defaultCfg)

	app := &App{
		Screen:           screen,
		Sim:              sim,
		AppConfig:        appConfig,
		Palettes:         palettes,
		ConfigPath:       configPath,
		UIConfig:         sim.Config,
		CursorX:          float64((w - simulation.SidebarWidth) / 2),
		CursorY:          float64(10),
		ActivePresetName: "Default",
		ActivePresetIdx:  0,
		MouseMode:        ModeSpawn,
		SimW:             sim.Width,
		SimH:             sim.Height,
		LastCursorX:      -1,
		LastCursorY:      -1,
		StyleBorder:      tcell.StyleDefault.Foreground(tcell.ColorWhite),
		StyleMenuBg:      tcell.StyleDefault.Background(tcell.ColorBlack),
		StyleMenuSel:     tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow),
		FpsTimer:         time.Now(),
	}

	app.CurrentGrid = make([][]int, app.SimW)
	app.LastGrid = make([][]int, app.SimW)
	for i := range app.CurrentGrid {
		app.CurrentGrid[i] = make([]int, app.SimH)
		app.LastGrid[i] = make([]int, app.SimH)
	}

	app.PresetNames = config.GetSortedPresetNames(appConfig.Presets)
	for i, name := range app.PresetNames {
		if name == app.ActivePresetName {
			app.ActivePresetIdx = i
			break
		}
	}

	app.SyncPalette()
	app.InitMenu()

	return app
}

func (a *App) Run() {
	go a.Sim.Run()

	events := make(chan tcell.Event)
	go func() {
		for {
			events <- a.Screen.PollEvent()
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 16)
	defer ticker.Stop()

	for {
		select {
		case ev := <-events:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				a.Resize()
			case *tcell.EventKey:
				if a.HandleInput(ev) {
					return
				}
			case *tcell.EventMouse:
				a.HandleMouse(ev)
			}
		case snapshot := <-a.Sim.RenderChan:
			a.CurrentParticles = snapshot.Points
			a.LastPhysTime = snapshot.CalcTime
		case <-ticker.C:
			a.HandleContinuousInput()
			a.Render()
		}
	}
}

func (a *App) HandleContinuousInput() {
	if a.IsMouseDown && a.MouseInBounds {
		cx, cy := a.CursorX, a.CursorY

		switch a.MouseMode {
		case ModeSpawn:
			a.Sim.CmdChan <- func(s *simulation.Simulation) { s.Spawn(cx, cy) }
		case ModeWall:
			ix, iy := int(cx), int(cy)
			a.Sim.CmdChan <- func(s *simulation.Simulation) { s.SetWall(ix, iy, true) }
		case ModeErase:
			ix, iy := int(cx), int(cy)
			a.Sim.CmdChan <- func(s *simulation.Simulation) { s.SetWall(ix, iy, false) }
		}
	}
}

func (a *App) SyncPalette() {
	found := false
	for i, p := range a.Palettes {
		if p.Name == a.UIConfig.PaletteName {
			a.UIConfig.PaletteIdx = i
			found = true
			break
		}
	}
	if !found {
		a.UIConfig.PaletteIdx = 0
	}
}

func (a *App) InitMenu() {
	a.MenuItems = []ui.MenuItem{
		{Name: "Preset", Type: "preset_enum", Val: &a.ActivePresetIdx, Step: 1.0, Fmt: "%s"},
		{Name: "Palette", Type: "enum", Val: &a.UIConfig.PaletteIdx, Step: 1.0, Fmt: "%s"},
		{Name: "SpawnQty", Type: "int", Val: &a.UIConfig.SpawnCount, Step: 5.0, Fmt: "%d"},
		{Name: "Gravity", Type: "float", Val: &a.UIConfig.Gravity, Step: 0.01, Fmt: "%.2f"},
		{Name: "RestDens", Type: "float", Val: &a.UIConfig.RestDensity, Step: 0.5, Fmt: "%.1f"},
		{Name: "Stiffness", Type: "float", Val: &a.UIConfig.Stiffness, Step: 0.01, Fmt: "%.2f"},
		{Name: "StiffNear", Type: "float", Val: &a.UIConfig.StiffnessNear, Step: 0.01, Fmt: "%.2f"},
		{Name: "Viscosity", Type: "float", Val: &a.UIConfig.Viscosity, Step: 0.001, Fmt: "%.3f"},
		{Name: "Damping", Type: "float", Val: &a.UIConfig.Damping, Step: 0.005, Fmt: "%.3f"},
		{Name: "InteractionRad", Type: "float", Val: &a.UIConfig.InteractionRad, Step: 0.2, Fmt: "%.1f"},
	}

	if a.ConfigPath != "" {
		a.MenuItems = append(a.MenuItems, ui.MenuItem{Name: "Save", Type: "action", Val: nil, Step: 0, Fmt: "> ENTER <"})
	}
}

func (a *App) Resize() {
	a.Screen.Sync()
	w, h := a.Screen.Size()

	a.Sim.CmdChan <- func(s *simulation.Simulation) { s.Resize(w, h) }

	a.SimW = w - simulation.SidebarWidth
	a.SimH = h

	a.CurrentGrid = make([][]int, a.SimW)
	a.LastGrid = make([][]int, a.SimW)
	for i := range a.CurrentGrid {
		a.CurrentGrid[i] = make([]int, a.SimH)
		a.LastGrid[i] = make([]int, a.SimH)
	}
	a.Screen.Clear()
}
