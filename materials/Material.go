package materials

import (
    "ray-tracing/primitives"
)

type MaterialType int8

const (
    ReflectDiffuse MaterialType = iota
    ReflectRefract
    Diffuse
    Transparent
)

type Material struct {
    Color                   primitives.Color
    Reflect, Refract, Alpha float64
    MaterialType            MaterialType

    MaterialId   int
    MaterialName *string
}

func NewMaterial(
    color primitives.Color, reflect float64, refract float64, alpha float64,
    materialId int, materialName *string) *Material {

    var materialType MaterialType
    if !primitives.Equal(alpha, 1) {
        materialType = Transparent
    } else if primitives.Equal(reflect, 0) && primitives.Equal(refract, 0) {
        materialType = Diffuse
    } else if primitives.Equal(refract, 0) {
        materialType = ReflectDiffuse
    } else {
        materialType = ReflectRefract
    }
    return &Material{
        Color: color, Reflect: reflect, Refract: refract, Alpha: alpha, MaterialType: materialType,
        MaterialId: materialId, MaterialName: materialName}
}