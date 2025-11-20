package render

import "time"

type FrameSnapshot struct {
	Points   []Point
	CalcTime time.Duration
}

type Point struct {
	X, Y int
}
