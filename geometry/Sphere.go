package geometry

import (
    "math"
    "ray-tracing/primitives"
)

type Sphere struct {
    Center primitives.Vector
    Radius float64
}

func (s Sphere) GetNormal(pos primitives.Vector) primitives.Vector {
    norm := pos.Sub(s.Center)
    if primitives.Greater(norm.Length(), s.Radius) {
        panic("normal length more than Radius")
    }
    return norm.Norm()
}

func (s Sphere) GetTexturePoint(pos primitives.Vector) primitives.Vector {
    panic("implement me")
}

func (s Sphere) GetBoundingBox() *BBox {
    radiusVector := primitives.Vector{s.Radius, s.Radius, s.Radius}
    return &BBox{s.Center.Sub(radiusVector), s.Center.Add(radiusVector)}
}

func (s Sphere) Intersect(ray *Ray) RayCoefIntersection {
    distance := ray.Distance(s.Center)

    if primitives.Greater(distance, s.Radius) {
        return RayCoefIntersection{}
    }

    scalarDistance := ray.GetLineCoef(s.Center)
    halfSphereDistance := math.Sqrt(math.Abs(s.Radius* s.Radius - distance * distance))
    rayD := math.Min(scalarDistance - halfSphereDistance, scalarDistance + halfSphereDistance)
    if primitives.Less(rayD, 0) {
        rayD = math.Max(scalarDistance - halfSphereDistance, scalarDistance + halfSphereDistance)
    }
    return RayCoefIntersection{IntersectionCoef: rayD, HasIntersection: true}
}
