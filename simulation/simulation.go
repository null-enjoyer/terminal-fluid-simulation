package simulation

import (
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/null-enjoyer/terminal-fluid-simulation/config"
	"github.com/null-enjoyer/terminal-fluid-simulation/render"
)

const (
	MaxParticles = 45000
	CellShift    = 2
	SidebarWidth = 32
	SubSteps     = 4
)

type Simulation struct {
	Particles []Particle
	GridHeads []int
	GridNext  []int
	Walls     []bool

	GridCols int
	GridRows int
	Width    int
	Height   int

	CmdChan    chan func(*Simulation)
	RenderChan chan render.FrameSnapshot
	Config     config.PhysicsConfig

	NumCPU int
}

func NewSimulation(termW, termH int, cfg config.PhysicsConfig) *Simulation {
	cfg.UpdateDerived()

	sim := &Simulation{
		Particles:  make([]Particle, 0, MaxParticles),
		GridNext:   make([]int, MaxParticles),
		CmdChan:    make(chan func(*Simulation), 100),
		RenderChan: make(chan render.FrameSnapshot, 2),
		Config:     cfg,
		NumCPU:     runtime.NumCPU(),
	}

	sim.Resize(termW, termH)
	return sim
}

func (s *Simulation) ParallelFor(action func(start, end int)) {
	count := len(s.Particles)
	if count == 0 {
		return
	}

	// if too few particles, run serial to avoid scheduler overhead
	if count < 1000 {
		action(0, count)
		return
	}

	var wg sync.WaitGroup
	chunkSize := (count + s.NumCPU - 1) / s.NumCPU

	for i := 0; i < s.NumCPU; i++ {
		start := i * chunkSize
		if start >= count {
			break
		}
		end := start + chunkSize
		if end > count {
			end = count
		}

		wg.Add(1)
		go func(s, e int) {
			defer wg.Done()
			action(s, e)
		}(start, end)
	}
	wg.Wait()
}

func (s *Simulation) Resize(termW, termH int) {
	if termW <= SidebarWidth {
		termW = SidebarWidth + 10
	}

	newWidth := termW - SidebarWidth
	newHeight := termH

	oldWidth := s.Width
	oldHeight := s.Height
	oldWalls := s.Walls

	s.Width = newWidth
	s.Height = newHeight

	s.GridCols = (s.Width >> CellShift) + 1
	s.GridRows = (s.Height >> CellShift) + 1

	totalCells := s.GridCols * s.GridRows
	s.GridHeads = make([]int, totalCells)
	s.Walls = make([]bool, s.Width*s.Height)

	if len(oldWalls) > 0 && oldWidth > 0 {
		minH := oldHeight
		if s.Height < minH {
			minH = s.Height
		}
		minW := oldWidth
		if s.Width < minW {
			minW = s.Width
		}

		for y := 0; y < minH; y++ {
			srcStart := y * oldWidth
			destStart := y * s.Width
			copy(s.Walls[destStart:destStart+minW], oldWalls[srcStart:srcStart+minW])
		}
	}

	limitX := float64(s.Width) - 1.1
	limitY := float64(s.Height) - 1.1
	margin := 1.1

	for i := range s.Particles {
		p := &s.Particles[i]

		if p.Pos.X > limitX {
			p.Pos.X = limitX
			p.OldPos.X = limitX
		} else if p.Pos.X < margin {
			p.Pos.X = margin
			p.OldPos.X = margin
		}

		if p.Pos.Y > limitY {
			p.Pos.Y = limitY
			p.OldPos.Y = limitY
		} else if p.Pos.Y < margin {
			p.Pos.Y = margin
			p.OldPos.Y = margin
		}
	}
}

func (s *Simulation) Run() {
	ticker := time.NewTicker(time.Millisecond * 16)
	defer ticker.Stop()

	snapshot := make([]render.Point, 0, MaxParticles)

	for range ticker.C {
	ProcessCommands:
		for {
			select {
			case cmd := <-s.CmdChan:
				cmd(s)
			default:
				break ProcessCommands
			}
		}

		start := time.Now()

		if !s.Config.IsPaused {
			dt := 1.0 / float64(SubSteps)
			for step := 0; step < SubSteps; step++ {
				s.UpdateSpatialHash()
				s.Integration(dt)
				s.SolveViscosity()
				s.SolveFluid()
				s.EnforceBoundaries()
			}
		}

		calcTime := time.Since(start)

		// render snapshot
		snapshot = snapshot[:0]
		for i := range s.Particles {
			p := s.Particles[i]
			snapshot = append(snapshot, render.Point{X: int(p.Pos.X), Y: int(p.Pos.Y)})
		}

		pointsCopy := make([]render.Point, len(snapshot))
		copy(pointsCopy, snapshot)

		select {
		case s.RenderChan <- render.FrameSnapshot{Points: pointsCopy, CalcTime: calcTime}:
		default:
		}
	}
}

func (s *Simulation) Spawn(x, y float64) {
	ix, iy := int(x), int(y)
	if uint(ix) < uint(s.Width) && uint(iy) < uint(s.Height) {
		if s.Walls[ix+iy*s.Width] {
			return
		}
	}

	for i := 0; i < s.Config.SpawnCount; i++ {
		if len(s.Particles) >= MaxParticles {
			break
		}
		jx := x + (rand.Float64()*6 - 3)
		jy := y + (rand.Float64()*4 - 2)

		p := Particle{
			Pos:    Vector{X: jx, Y: jy},
			OldPos: Vector{X: jx, Y: jy},
		}
		p.OldPos.Y -= 0.5
		s.Particles = append(s.Particles, p)
	}
}

func (s *Simulation) SetWall(x, y int, isWall bool) {
	if uint(x) < uint(s.Width) && uint(y) < uint(s.Height) {
		s.Walls[x+y*s.Width] = isWall
	}
}
