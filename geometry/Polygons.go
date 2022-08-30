package geometry

import (
    "ray-tracing/materials"
    "ray-tracing/primitives"
)

type IGeometryObject interface {
    GetNormal(primitives.Vector) primitives.Vector
    GetTexturePoint(pos primitives.Vector) primitives.Vector
    GetBoundingBox() *BBox
    Intersect(ray *Ray) RayCoefIntersection
    GetMaterial() *materials.Material
}

type Triangle struct {
    points [3]primitives.Vector
    textureCoords [3]primitives.Vector
    material *materials.Material

    normal      primitives.Vector
    surfaceArea float64
    planeCoefficient float64
}

func (trg *Triangle) GetNormal(primitives.Vector) primitives.Vector {
    return trg.normal
}

func calculateArea(trg *Triangle) float64 {
    return trg.points[1].Sub(trg.points[0]).Cross(trg.points[2].Sub(trg.points[0])).Length()
}

func calculateNormal(trg *Triangle) float64 {
    trg.normal = trg.points[1].Sub(trg.points[0]).Cross(trg.points[2].Sub(trg.points[0])).Norm()
    return trg.normal.Dot(trg.points[0])
}

func NewTriangle(points [3]primitives.Vector, textureCoords [3]primitives.Vector, material *materials.Material) *Triangle {
    var triangle = &Triangle{points: points, textureCoords: textureCoords, material: material}
    triangle.surfaceArea = calculateArea(triangle)
    triangle.planeCoefficient = calculateNormal(triangle)
    return triangle
}

func (trg *Triangle) GetTexturePoint(pos primitives.Vector) primitives.Vector {
    tempPos := pos.Sub(trg.points[1])
    baseU := trg.textureCoords[2].Sub(trg.textureCoords[1])
    baseV := trg.textureCoords[0].Sub(trg.textureCoords[1])
    baseX := trg.points[2].Sub(trg.points[1])
    baseY := trg.points[0].Sub(trg.points[1])

    partU := baseU.Mult(tempPos.Dot(baseX)).Div(baseX.SqrLength())
    partV := baseV.Mult(tempPos.Dot(baseY)).Div(baseY.SqrLength())
    return partU.Add(partV)
}

func (trg *Triangle) Intersect(ray *Ray) RayCoefIntersection {
    coef := (trg.planeCoefficient - ray.Begin.Dot(trg.normal)) / (ray.Direction.Dot(trg.normal))
    point := ray.Begin.Add(ray.Direction.Mult(coef))

    area := 0.0

    for i := 0; i < 3; i++ {
        ind1, ind2 := i % 3, (i + 1) % 3
        area += point.Sub(trg.points[ind1]).Cross(point.Sub(trg.points[ind2])).Length()
    }

    if !primitives.Equal(area, trg.surfaceArea) || primitives.Less(area * trg.surfaceArea, 0) {
        return RayCoefIntersection{
            HasIntersection:  false,
            IntersectionCoef: 0,
        }
    }
    return RayCoefIntersection{
        HasIntersection:  true,
        IntersectionCoef: coef,
    }
}

func (trg *Triangle) GetBoundingBox() *BBox {
    return CreateFromPoints(trg.points[:])
}

func (trg *Triangle) GetMaterial() *materials.Material {
    return trg.material
}