package geometry

import (
    "math"
    "ray-tracing/primitives"
    "sort"
)

type BBox struct {
    Left, Right primitives.Vector
}

func CreateFromPoints(points []primitives.Vector) *BBox {
    var res BBox
    res.Left, res.Right = points[0], points[0]
    for _, point := range points {
        res.Left = primitives.Min(res.Left, point)
        res.Right = primitives.Max(res.Right, point)
    }
    return &res
}

func (bbox *BBox) Expand(other *BBox) {
    bbox.Left = primitives.Min(bbox.Left, other.Left)
    bbox.Right = primitives.Max(bbox.Right, other.Right)
}

func (bbox *BBox) Contains(point primitives.Vector) bool {
    return bbox.Left.LessEqual(point) && bbox.Right.GreaterEqual(point)
}

func (bbox *BBox) Split(axisNumber int, value float64) [2]*BBox {
    if axisNumber > 2 {
        panic("Wrong axis " + string(axisNumber))
    }
    newLeft := bbox.Left
    newRight := bbox.Right
    switch axisNumber {
    case 0:
        newLeft.X = value
        newRight.X = value
    case 1:
        newLeft.Y = value
        newRight.Y = value
    case 2:
        newLeft.Z = value
        newRight.Z = value
    }
    return [2]*BBox{{bbox.Left, newRight}, {newLeft, bbox.Right}}
}

func (bbox *BBox) SurfaceArea() float64 {
    dir := bbox.Right.Sub(bbox.Left)
    return 2 * dir.Dot(dir)
}

func (bbox *BBox) CreatePlaneAndIntersect(ray *Ray, p1, p2, p3 primitives.Vector) RayCoefIntersection {
    normal := p2.Sub(p1).Cross(p3.Sub(p1)).Norm()
    if primitives.Equal(ray.Direction.Dot(normal), 0) {
        return RayCoefIntersection{}
    }
    coef := (normal.Dot(p1) - ray.Begin.Dot(normal)) / ray.Direction.Dot(normal)
    if !bbox.Contains(ray.Begin.Add(ray.Direction.Mult(coef))) {
        return RayCoefIntersection{}
    }
    return RayCoefIntersection{true, coef}
}

func (bbox *BBox) Intersect(ray *Ray) [2]RayCoefIntersection {
    var pu, pd, pl, pr, pn, pf RayCoefIntersection

    pd = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Left.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Left.X, bbox.Right.Y, bbox.Left.Z})
    pu = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Left.X, bbox.Left.Y, bbox.Right.Z},
        primitives.Vector{X: bbox.Right.X, Y: bbox.Left.Y, Z: bbox.Right.Z},
        primitives.Vector{bbox.Left.X, bbox.Right.Y, bbox.Right.Z})

    pl = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Left.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Left.X, bbox.Right.Y, bbox.Left.Z},
        primitives.Vector{bbox.Left.X, bbox.Right.Y, bbox.Right.Z})
    pr = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Right.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Right.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Right.Y, bbox.Right.Z})

    pn = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Left.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Left.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Left.Y, bbox.Right.Z})
    pf = bbox.CreatePlaneAndIntersect(ray,
        primitives.Vector{bbox.Left.X, bbox.Right.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Right.Y, bbox.Left.Z},
        primitives.Vector{bbox.Right.X, bbox.Right.Y, bbox.Right.Z})

    coefs := make([]float64, 0, 6)
    if pd.HasIntersection {
        coefs = append(coefs, pd.IntersectionCoef)
    }
    if pu.HasIntersection {
        coefs = append(coefs, pu.IntersectionCoef)
    }
    if pl.HasIntersection {
        coefs = append(coefs, pl.IntersectionCoef)
    }
    if pr.HasIntersection {
        coefs = append(coefs, pr.IntersectionCoef)
    }
    if pn.HasIntersection {
        coefs = append(coefs, pn.IntersectionCoef)
    }
    if pf.HasIntersection {
        coefs = append(coefs, pf.IntersectionCoef)
    }
    sort.Float64s(coefs)

    if len(coefs) == 0 || coefs[len(coefs) - 1] < 0 {
        return [2]RayCoefIntersection{{}, {}}
    }

    var firstCoef, lastCoef float64
    lastCoef = coefs[len(coefs) - 1]
    firstCoef = lastCoef
    for _, coef := range coefs {
        if coef > 0 {
            firstCoef = coef
            break
        }
    }
    return [2]RayCoefIntersection{{true, firstCoef}, {true, lastCoef}}
}

func (bbox *BBox) GetMin(axis int) float64 {
    switch axis {
    case 0:
        return math.Min(bbox.Left.X, bbox.Right.X)
    case 1:
        return math.Min(bbox.Left.Y, bbox.Right.Y)
    case 2:
        return math.Min(bbox.Left.Z, bbox.Right.Z)
    default:
        panic("Axis not in range [0:2]")
    }
}

func (bbox *BBox) GetMax(axis int) float64 {
    switch axis {
    case 0:
        return math.Max(bbox.Left.X, bbox.Right.X)
    case 1:
        return math.Max(bbox.Left.Y, bbox.Right.Y)
    case 2:
        return math.Max(bbox.Left.Z, bbox.Right.Z)
    default:
        panic("Axis not in range [0:2]")
    }
}
