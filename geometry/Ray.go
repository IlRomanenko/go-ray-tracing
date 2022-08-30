package geometry

import (
    "ray-tracing/primitives"
)

type Ray struct {
    Begin     primitives.Vector
    Direction primitives.Vector
}

type RayCoefIntersection struct {
    HasIntersection  bool
    IntersectionCoef float64
}

type Intersection struct {
    Coefficient RayCoefIntersection
    Point       primitives.Vector
    Object      IGeometryObject
    Color       primitives.Color
}

func NewRay(begin primitives.Vector, end primitives.Vector) *Ray {
    return &Ray{Begin: begin, Direction: end.Sub(begin).Norm()}
}

func NewRayCoefIntersection(value float64) RayCoefIntersection {
    return RayCoefIntersection{HasIntersection: true, IntersectionCoef: value}
}

func (ray *Ray) GetLineCoef(point primitives.Vector) float64 {
    return point.Sub(ray.Begin).Dot(ray.Direction)
}

func (ray *Ray) Distance(point primitives.Vector) float64 {
    var npoint = point.Sub(ray.Begin)
    return npoint.Sub(ray.Direction.Mult(ray.Direction.Dot(npoint))).Length()
}

func (ray *Ray) GetReflectRay(point primitives.Vector, normal primitives.Vector) *Ray {
    dir := ray.Direction.Mult(-1.0)
    normDir := dir.Sub(normal.Mult(dir.Dot(normal)))
    newDir := dir.Sub(normDir.Mult(2.0))
    return &Ray{point, point.Add(newDir)}
}
