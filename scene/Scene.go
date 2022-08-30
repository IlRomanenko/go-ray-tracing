package scene

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"ray-tracing/geometry"
	"ray-tracing/kd_tree"
	"ray-tracing/materials"
	"ray-tracing/primitives"
	"sync"
	"sync/atomic"

	"github.com/udhos/gwob"
)

const ANTIALIASING_CONST float64 = 0.2
const ANTIALIASING_POINT_COUNT int = 5
const MAX_RAY_TRACING_DEPTH int = 10

type SceneSerialisable struct {
	Lights    []Light
	Viewport  Viewport
	ModelName string
}

type Scene struct {
	objects  []geometry.IGeometryObject
	KDTree   *kd_tree.KDTree
	Lights   []Light
	Viewport Viewport

	Pixels [][]primitives.Color

	Wg         sync.WaitGroup
	RaysCasted atomic.Value
}

type renderInput struct {
	x, y         int
	antialiasing bool
}

func OpenScene(filename string) (*Scene, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var sceneData SceneSerialisable
	err = json.Unmarshal(data, &sceneData)
	if err != nil {
		return nil, err
	}
	options := gwob.ObjParserOptions{IgnoreNormals: true}
	obj, err := gwob.NewObjFromFile(filepath.Join(filepath.Dir(filename), sceneData.ModelName), &options)
	if err != nil {
		return nil, err
	}

	mtlib, err := gwob.ReadMaterialLibFromFile(filepath.Join(filepath.Dir(filename), obj.Mtllib), &gwob.ObjParserOptions{})
	if err != nil {
		return nil, err
	}

	triangles := make([]geometry.IGeometryObject, 0)

	for _, g := range obj.Groups {
		group_lib := mtlib.Lib[g.Usemtl]
		// Incorrect, but fast for simple use
		material := materials.NewMaterial(
			primitives.Color{
				R: float64(group_lib.Kd[0]),
				G: float64(group_lib.Kd[1]),
				B: float64(group_lib.Kd[2]),
			}, 0, 0, 1, 0, &group_lib.Name,
		)

		for ind := g.IndexBegin; ind < g.IndexBegin+g.IndexCount; ind += 3 {
			v1 := primitives.VectorFromFloat32(obj.VertexCoordinates(obj.Indices[ind]))
			v2 := primitives.VectorFromFloat32(obj.VertexCoordinates(obj.Indices[ind+1]))
			v3 := primitives.VectorFromFloat32(obj.VertexCoordinates(obj.Indices[ind+2]))
			triangle := geometry.NewTriangle([3]primitives.Vector{v1, v2, v3}, [3]primitives.Vector{}, material)
			triangles = append(triangles, triangle)
		}
	}

	return NewScene(triangles, sceneData.Lights, sceneData.Viewport), nil
}

func NewScene(objects []geometry.IGeometryObject, lights []Light, viewport Viewport) *Scene {
	scene := Scene{objects: objects, Lights: lights, Viewport: viewport}
	scene.KDTree = new(kd_tree.KDTree)
	scene.KDTree.BuildTree(objects)
	scene.Pixels = make([][]primitives.Color, viewport.Width)
	for ind := 0; ind < viewport.Width; ind++ {
		scene.Pixels[ind] = make([]primitives.Color, viewport.Height)
	}
	return &scene
}

func (scene *Scene) Render() {
	scene.Wg.Add(16)

	inputChannel := make(chan renderInput)

	for i := 0; i < 16; i++ {
		go renderWorker(scene, inputChannel)
	}

	for x := 0; x < scene.Viewport.Width; x++ {
		for y := 0; y < scene.Viewport.Height; y++ {
			inputChannel <- renderInput{x, y, false}
		}
	}
	close(inputChannel)
}

func (scene *Scene) GetPixels() [][]primitives.Color {
	return scene.Pixels
}

func renderWorker(scene *Scene, input chan renderInput) {
	base_w := scene.Viewport.GetWidthBase().Div(float64(scene.Viewport.Width))
	base_h := scene.Viewport.GetHeightBase().Div(float64(scene.Viewport.Height))
	origin := scene.Viewport.Origin
	defaultOffset := base_w.Div(2.0).Add(base_h.Div(2))

	var color primitives.Color

	count := 0
	for obj := range input {
		basePoint := scene.Viewport.TopLeft.Add(base_w.Mult(float64(obj.x)).Add(base_h.Mult(float64(obj.y))))
		if !obj.antialiasing {
			screenPoint := basePoint.Add(defaultOffset)
			newRay := geometry.NewRay(origin, screenPoint)
			color = scene.traceRay(newRay)
		}
		scene.Pixels[obj.x][obj.y] = color
		count++
	}
	//fmt.Printf("Done after %v\n", count)
	scene.Wg.Done()
}

func (scene *Scene) castRayKD(ray *geometry.Ray) geometry.Intersection {
	newRay := *ray
	newRay.Begin = newRay.Begin.Add(newRay.Direction.Mult(1e-5))
	return scene.KDTree.CastRay(&newRay)
}

func (scene *Scene) castRay(ray *geometry.Ray, additionalLight float64, depth int) geometry.Intersection {
	if depth > MAX_RAY_TRACING_DEPTH {
		return geometry.Intersection{}
	}
	//TODO fix this fucking shit
	additionalLight = 0
	intersection := scene.castRayKD(ray)

	if !intersection.Coefficient.HasIntersection {
		return intersection
	}
	material := intersection.Object.GetMaterial()
	intersectionNormal := intersection.Object.GetNormal(intersection.Point)

	var refractColor, reflectColor primitives.Color
	var Kr, Kt float64 // reflection and refraction mix value

	Kr = fresnel(ray.Direction, intersectionNormal, material.Refract)
	Kt = 1 - Kr

	lightIntensity := scene.getLightIntensity(intersection.Point, intersection.Object) + additionalLight
	normalizedLight := math.Min(1, lightIntensity)

	// texturePoint := intersection.Object.GetTexturePoint(intersection.Point)
	switch material.MaterialType {
	case materials.ReflectDiffuse:
		{
			//texturePoint
			materialColor := material.Color.Mult(1 - material.Reflect)
			reflectRay := ray.GetReflectRay(intersection.Point, intersectionNormal)
			reflectInter := scene.castRay(reflectRay, lightIntensity, depth+1)
			if reflectInter.Coefficient.HasIntersection {
				reflectColor = reflectInter.Color.Mult(material.Reflect)
			}
			intersection.Color = materialColor.Mult(normalizedLight).Add(reflectColor)
		}
	case materials.ReflectRefract:
		{
			//reflection
			reflectRay := ray.GetReflectRay(intersection.Point, intersectionNormal)
			reflectInter := scene.castRay(reflectRay, lightIntensity, depth+1)
			if reflectInter.Coefficient.HasIntersection {
				reflectColor = reflectInter.Color.Mult(Kr)
			}

			//refraction
			refractDirection := refract(ray, intersectionNormal, material.Refract)
			refractRay := geometry.NewRay(intersection.Point, intersection.Point.Add(refractDirection))
			refractInter := scene.castRay(refractRay, lightIntensity, depth+1)
			if refractInter.Coefficient.HasIntersection {
				refractColor = refractInter.Color.Mult(Kt)
			}

			intersection.Color = reflectColor.Add(refractColor)
		}
	case materials.Diffuse:
		{
			//texturePoint
			materialColor := material.Color.Mult(normalizedLight)
			intersection.Color = materialColor
		}
	case materials.Transparent:
		// texturePoint
		materialColor := material.Color.Mult(material.Alpha)

		refractDirection := refract(ray, intersectionNormal, material.Refract)

		refractRay := geometry.NewRay(intersection.Point, intersection.Point.Add(refractDirection))
		refractInter := scene.castRay(refractRay, lightIntensity, depth+1)
		if refractInter.Coefficient.HasIntersection {
			refractColor = refractInter.Color.Mult(1 - material.Alpha)
		}

		intersection.Color = materialColor.Mult(normalizedLight).Add(refractColor)
	}

	intersection.Color = intersection.Color.Normalize()
	return intersection
}

func (scene *Scene) traceRay(ray *geometry.Ray) primitives.Color {
	intersection := scene.castRay(ray, 0, 0)
	if !intersection.Coefficient.HasIntersection {
		return primitives.Color{R: 0.2, G: 0.2, B: 0.2}
	}
	return intersection.Color.Normalize()
}

func (scene *Scene) getLightIntensity(point primitives.Vector, object geometry.IGeometryObject) float64 {
	lightIntensity := 0.0
	for _, light := range scene.Lights {
		newRay := geometry.NewRay(point, light.Position)
		lightIntersection := scene.castRayKD(newRay)
		lightPositionCoef := newRay.GetLineCoef(light.Position)
		if !lightIntersection.Coefficient.HasIntersection ||
			primitives.Greater(lightIntersection.Coefficient.IntersectionCoef, lightPositionCoef) {
			lightVector := light.Position.Sub(point)
			lightSqrLength := lightVector.SqrLength()
			normLightVector := lightVector.Norm()

			currentPointLightContribution := object.GetNormal(point).Dot(normLightVector)
			distanceForOriginPoint := light.Ref.Distance / light.Ref.Power
			currentPower := light.Power / distanceForOriginPoint / lightSqrLength

			currentPointLightContribution *= currentPower

			lightIntensity += math.Max(currentPointLightContribution, 0.0)
		}
	}
	return math.Min(lightIntensity+0.2, 1.0)
}

func refract(ray *geometry.Ray, normal primitives.Vector, ior float64) primitives.Vector {
	cosi := primitives.Clamp(-1, 1, normal.Dot(ray.Direction))
	etai, etat := 1.0, ior
	if primitives.Less(cosi, 0) {
		cosi *= -1
	} else {
		etai, etat = etat, etai
		normal = normal.Mult(-1)
	}
	eta := etai / etat
	k := 1 - eta*eta*(1-cosi*cosi)
	if primitives.Less(k, 0) {
		// total internal reflection
		return primitives.Vector{}
	}
	return ray.Direction.Mult(eta).Add(normal.Mult(eta*cosi - math.Sqrt(k)))
}

func fresnel(I, N primitives.Vector, ior float64) float64 {
	var Kr float64
	cosi := primitives.Clamp(-1, 1, N.Dot(I))
	etai, etat := 1.0, ior
	if primitives.Greater(cosi, 0) {
		etai, etat = etat, etai
	}
	sint := etai / etat * math.Sqrt(math.Max(0, 1-cosi*cosi))

	if primitives.GreaterEqual(sint, 0) {
		//total inherit reflection
		Kr = 1
	} else {
		cost := math.Sqrt(math.Max(0, 1-sint*sint))
		cosi = math.Abs(cosi)
		Rs := ((etat * cosi) - (etai * cost)) / ((etat * cosi) + (etai * cost))
		Rp := ((etai * cosi) - (etat * cost)) / ((etai * cosi) + (etat * cost))
		Kr = (Rs*Rs + Rp*Rp) / 2
	}
	return Kr
}
