package primitives

import "math"

const EPS = 1e-9

type Vector struct {
    X, Y, Z float64
}

func VectorFromFloat32(x, y, z float32) Vector {
    return Vector{float64(x), float64(y), float64(z)}
}

func (v Vector) Add(q Vector) Vector {
    return Vector{v.X + q.X, v.Y + q.Y, v.Z + q.Z}
}

func (v Vector) Sub(q Vector) Vector {
    return Vector{v.X - q.X, v.Y - q.Y, v.Z - q.Z}
}

func (v Vector) Mult(f float64) Vector {
    return Vector{v.X * f, v.Y * f, v.Z * f}
}

func (v Vector) Div(f float64) Vector {
    return Vector{v.X / f, v.Y / f, v.Z / f}
}

func (v Vector) Dot(q Vector) float64 {
    return v.X * q.X + v.Y * q.Y + v.Z * q.Z
}

func (v Vector) Cross(q Vector) Vector {
    return Vector{v.Y* q.Z - v.Z* q.Y, -(v.X* q.Z - v.Z* q.X), v.X* q.Y - v.Y* q.X}
}

func (v Vector) SqrLength() float64 {
    return v.X* v.X + v.Y* v.Y + v.Z* v.Z
}

func (v Vector) Length() float64 {
    return math.Sqrt(v.SqrLength())
}

func (v Vector) Norm() Vector {
    var length = v.Length()
    return Vector{v.X / length, v.Y / length, v.Z / length}
}

func (v Vector) LessEqual(q Vector) bool {
    return LessEqual(v.X, q.X) && LessEqual(v.Y, q.Y) && LessEqual(v.Z, q.Z)
}

func (v Vector) GreaterEqual(q Vector) bool {
    return GreaterEqual(v.X, q.X) && GreaterEqual(v.Y, q.Y) && GreaterEqual(v.Z, q.Z)
}

func Min(v1, v2 Vector) Vector {
    return Vector{math.Min(v1.X, v2.X), math.Min(v1.Y, v2.Y), math.Min(v1.Z, v2.Z)}
}

func Max(v1, v2 Vector) Vector {
    return Vector{math.Max(v1.X, v2.X), math.Max(v1.Y, v2.Y), math.Max(v1.Z, v2.Z)}
}

func Equal(a, b float64) bool {
    return math.Abs(a - b) < EPS
}

func Less(a, b float64) bool {
    return a +EPS < b
}

func Greater(a, b float64) bool {
    return a -EPS > b
}

func LessEqual(a, b float64) bool {
    return Less(a, b) || Equal(a, b)
}

func GreaterEqual(a, b float64) bool {
    return Greater(a, b) || Equal(a, b)
}

func Clamp(low, high, value float64) float64 {
    return math.Max(low, math.Min(high, value))
}