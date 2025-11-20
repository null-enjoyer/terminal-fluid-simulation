package simulation

import (
	"math"
)

const (
	MaxVelocity = 3.0
)

func (s *Simulation) UpdateSpatialHash() {
	for i := range s.GridHeads {
		s.GridHeads[i] = -1
	}

	cols := s.GridCols
	rows := s.GridRows

	for i := range s.Particles {
		p := &s.Particles[i]

		gx := int(p.Pos.X) >> CellShift
		gy := int(p.Pos.Y) >> CellShift

		if gx < 0 {
			gx = 0
		} else if gx >= cols {
			gx = cols - 1
		}
		if gy < 0 {
			gy = 0
		} else if gy >= rows {
			gy = rows - 1
		}

		cellIdx := gx + gy*cols
		s.GridNext[i] = s.GridHeads[cellIdx]
		s.GridHeads[cellIdx] = i
	}
}

func (s *Simulation) IsWallSafe(x, y float64) bool {
	ix, iy := int(x), int(y)
	if uint(ix) >= uint(s.Width) || uint(iy) >= uint(s.Height) {
		return true
	}
	return s.Walls[ix+iy*s.Width]
}

func (s *Simulation) Integration(dt float64) {
	wLimit := float64(s.Width) - 1.1
	hLimit := float64(s.Height) - 1.1
	margin := 1.1
	stepGravity := s.Config.Gravity * dt
	damping := s.Config.Damping
	width := s.Width
	uintW, uintH := uint(s.Width), uint(s.Height)

	s.ParallelFor(func(start, end int) {
		for i := start; i < end; i++ {
			p := &s.Particles[i]

			vx := (p.Pos.X - p.OldPos.X) * damping
			vy := (p.Pos.Y - p.OldPos.Y) * damping
			vy += stepGravity

			vSq := vx*vx + vy*vy
			if vSq > MaxVelocity*MaxVelocity {
				scale := MaxVelocity / math.Sqrt(vSq)
				vx *= scale
				vy *= scale
			}

			// anti-tunneling raycast, check multiple points along the path to ensure we don't jump a wall
			steps := int(math.Sqrt(vSq)) + 1
			if steps > 5 {
				steps = 5
			}

			startPos := p.Pos
			collided := false

			for k := 1; k <= steps; k++ {
				t := float64(k) / float64(steps)
				testX := startPos.X + (vx * t)
				testY := startPos.Y + (vy * t)

				ix, iy := int(testX), int(testY)
				isWall := false
				if uint(ix) >= uintW || uint(iy) >= uintH {
					isWall = true
				} else {
					isWall = s.Walls[ix+iy*width]
				}

				if isWall {
					p.OldPos = p.Pos
					collided = true
					break
				}

				p.Pos.X = testX
				p.Pos.Y = testY
			}

			if collided {
				continue
			}

			p.OldPos = startPos

			// screen boundaries
			if p.Pos.X < margin {
				p.Pos.X = margin
				p.OldPos.X = margin
			} else if p.Pos.X > wLimit {
				p.Pos.X = wLimit
				p.OldPos.X = wLimit
			}

			if p.Pos.Y < margin {
				p.Pos.Y = margin
				p.OldPos.Y = margin
			} else if p.Pos.Y > hLimit {
				p.Pos.Y = hLimit
				p.OldPos.Y = p.Pos.Y
			}
		}
	})
}

func (s *Simulation) SolveFluid() {
	cols := s.GridCols
	rows := s.GridRows
	radSq := s.Config.InteractionRadSq
	invRad := s.Config.InvInteractionRad
	width := s.Width
	uintW, uintH := uint(s.Width), uint(s.Height)
	stiffness := s.Config.Stiffness
	stiffNear := s.Config.StiffnessNear
	restDens := s.Config.RestDensity

	s.ParallelFor(func(start, end int) {
		type Neighbor struct {
			Index int
			Q     float64
		}
		var neighbors [64]Neighbor

		for i := start; i < end; i++ {
			p := &s.Particles[i]

			gx := int(p.Pos.X) >> CellShift
			gy := int(p.Pos.Y) >> CellShift

			density := 0.0
			nearDensity := 0.0
			neighborCount := 0

			startX := gx - 1
			if startX < 0 {
				startX = 0
			}
			endX := gx + 1
			if endX >= cols {
				endX = cols - 1
			}
			startY := gy - 1
			if startY < 0 {
				startY = 0
			}
			endY := gy + 1
			if endY >= rows {
				endY = rows - 1
			}

			for x := startX; x <= endX; x++ {
				yOffset := x
				for y := startY; y <= endY; y++ {
					cellIdx := yOffset + y*cols
					nj := s.GridHeads[cellIdx]

					for nj != -1 {
						if nj != i {
							pj := &s.Particles[nj]
							dx := pj.Pos.X - p.Pos.X
							dy := pj.Pos.Y - p.Pos.Y
							rSq := dx*dx + dy*dy

							if rSq < radSq && rSq > 1e-6 {
								r := math.Sqrt(rSq)
								q := 1.0 - (r * invRad)
								q2 := q * q

								density += q2
								nearDensity += q2 * q

								if neighborCount < 64 {
									neighbors[neighborCount] = Neighbor{nj, q}
									neighborCount++
								}
							}
						}
						nj = s.GridNext[nj]
					}
				}
			}

			pressure := stiffness * (density - restDens)
			nearPressure := stiffNear * nearDensity

			pVecX, pVecY := 0.0, 0.0

			for k := 0; k < neighborCount; k++ {
				n := neighbors[k]
				pj := &s.Particles[n.Index]

				dm := (pressure * n.Q) + (nearPressure * n.Q * n.Q)

				dx := pj.Pos.X - p.Pos.X
				dy := pj.Pos.Y - p.Pos.Y
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist > 1e-4 {
					invDist := 1.0 / dist
					moveX := (dx * invDist) * dm * 0.5
					moveY := (dy * invDist) * dm * 0.5

					pVecX -= moveX
					pVecY -= moveY
				}
			}

			// high pressure can cause massive jumps, so clamp to 1.0 pixel max per step
			pVecLenSq := pVecX*pVecX + pVecY*pVecY
			if pVecLenSq > 1.0 {
				scale := 1.0 / math.Sqrt(pVecLenSq)
				pVecX *= scale
				pVecY *= scale
			}

			targetPX := p.Pos.X + pVecX
			targetPY := p.Pos.Y + pVecY

			// check wall collision
			ix, iy := int(targetPX), int(targetPY)
			isWall := false
			if uint(ix) >= uintW || uint(iy) >= uintH {
				isWall = true
			} else {
				isWall = s.Walls[ix+iy*width]
			}

			if !isWall {
				p.Pos.X = targetPX
				p.Pos.Y = targetPY
			}
		}
	})
}

func (s *Simulation) SolveViscosity() {
	cols := s.GridCols
	rows := s.GridRows
	radSq := s.Config.InteractionRadSq
	invRad := s.Config.InvInteractionRad
	visc := s.Config.Viscosity

	s.ParallelFor(func(start, end int) {
		for i := start; i < end; i++ {
			p := &s.Particles[i]

			gx := int(p.Pos.X) >> CellShift
			gy := int(p.Pos.Y) >> CellShift

			startX := gx - 1
			if startX < 0 {
				startX = 0
			}
			endX := gx + 1
			if endX >= cols {
				endX = cols - 1
			}
			startY := gy - 1
			if startY < 0 {
				startY = 0
			}
			endY := gy + 1
			if endY >= rows {
				endY = rows - 1
			}

			for x := startX; x <= endX; x++ {
				yOffset := x
				for y := startY; y <= endY; y++ {
					cellIdx := yOffset + y*cols
					nj := s.GridHeads[cellIdx]

					for nj != -1 {
						if nj != i {
							pj := &s.Particles[nj]
							dx := pj.Pos.X - p.Pos.X
							dy := pj.Pos.Y - p.Pos.Y
							rSq := dx*dx + dy*dy

							if rSq < radSq && rSq > 1e-6 {
								r := math.Sqrt(rSq)

								invR := 1.0 / r
								nx, ny := dx*invR, dy*invR

								v1x := p.Pos.X - p.OldPos.X
								v1y := p.Pos.Y - p.OldPos.Y
								v2x := pj.Pos.X - pj.OldPos.X
								v2y := pj.Pos.Y - pj.OldPos.Y

								velAlongNormal := (v1x-v2x)*nx + (v1y-v2y)*ny

								if velAlongNormal > 0 {
									impulse := velAlongNormal * (1 - (r * invRad)) * visc
									ix, iy := nx*impulse, ny*impulse
									p.OldPos.X -= ix
									p.OldPos.Y -= iy
								}
							}
						}
						nj = s.GridNext[nj]
					}
				}
			}
		}
	})
}

func (s *Simulation) EnforceBoundaries() {
	uintW, uintH := uint(s.Width), uint(s.Height)
	width := s.Width

	s.ParallelFor(func(start, end int) {
		for i := start; i < end; i++ {
			p := &s.Particles[i]

			ix, iy := int(p.Pos.X), int(p.Pos.Y)
			isWall := false
			if uint(ix) >= uintW || uint(iy) >= uintH {
				isWall = true
			} else {
				isWall = s.Walls[ix+iy*width]
			}

			if isWall {
				oix, oiy := int(p.OldPos.X), int(p.OldPos.Y)
				oldSafe := false
				if uint(oix) < uintW && uint(oiy) < uintH {
					if !s.Walls[oix+oiy*width] {
						oldSafe = true
					}
				}

				if oldSafe {
					p.Pos = p.OldPos
				} else {
					foundSafe := false

					// check 4 cardinal directions
					dirs := []struct{ dx, dy float64 }{
						{0, -1}, {0, 1}, {-1, 0}, {1, 0},
					}

					for _, d := range dirs {
						nx, ny := p.Pos.X+d.dx, p.Pos.Y+d.dy
						nix, niy := int(nx), int(ny)
						if uint(nix) < uintW && uint(niy) < uintH {
							if !s.Walls[nix+niy*width] {
								p.Pos.X = nx
								p.Pos.Y = ny
								foundSafe = true
								break
							}
						}
					}

					if !foundSafe {
						p.Pos = p.OldPos
					}
				}
				p.OldPos = p.Pos
			}
		}
	})
}
