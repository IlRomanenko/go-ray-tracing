package primitives

import "math"

type Color struct {
    R, G, B float64
}

func (c Color) Mult(f float64) Color {
    return Color{c.R * f, c.G * f, c.B * f}
}

func (c Color) Add(o Color) Color {
    return Color{c.R + o.R, c.G + o.G, c.B + o.B}
}

func (c Color) L1Norm(o Color) float64 {
    return math.Abs(c.R - o.R) + math.Abs(c.G - o.G) + math.Abs(c.B - o.B)
}

func (c Color) Normalize() Color {
    return Color{math.Min(math.Abs(c.R), 1.0), math.Min(math.Abs(c.G), 1.0), math.Min(math.Abs(c.B), 1.0)}
}
