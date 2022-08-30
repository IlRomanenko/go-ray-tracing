package kd_tree

import (
    "math"
    "ray-tracing/geometry"
    "ray-tracing/primitives"
    "sort"
    "time"
)

type SplitPlane struct {
    value float64
    index int
}

type KDTreeNode struct {
    left, right *KDTreeNode
    bbox *geometry.BBox
    splitPlane SplitPlane
    nodeSize int
    objects []geometry.IGeometryObject
}

type KDTree struct {
    root *KDTreeNode
    TotalBuildingTime time.Duration
}

type BBoxSplit struct {
    value float64
    axis int
    cost float64
}

type kdTreeNodeInput struct {
    node *KDTreeNode
    objects []geometry.IGeometryObject
    bbox *geometry.BBox
}

const (
	TRAVERSAL_COEF    = 2
	INTERSECTION_COEF = 1
)

func costFunction(leftProbability float64, leftSize int, rightProbability float64, rightSize int) float64 {
    fullProb := leftProbability * float64(leftSize) + rightProbability * float64(rightSize)
    return TRAVERSAL_COEF + INTERSECTION_COEF * fullProb
}

func getBoundingBox(objects []geometry.IGeometryObject) *geometry.BBox {
    result := objects[0].GetBoundingBox()
    for _, obj := range objects {
        result.Expand(obj.GetBoundingBox())
    }
    return result
}

func surfaceAreaHeuristic(planeCoord float64, axis int, voxel *geometry.BBox, leftSize, middleSize, rightSize int) float64 {
    boxes := voxel.Split(axis, planeCoord)
    leftProb := boxes[0].SurfaceArea() / voxel.SurfaceArea()
    rightProb := boxes[1].SurfaceArea() / voxel.SurfaceArea()

    leftPartCost := costFunction(leftProb, leftSize + middleSize, rightProb, rightSize)
    rightPartCost := costFunction(leftProb, leftSize, rightProb, middleSize + rightSize)

    if primitives.Equal(leftPartCost, 0) || primitives.Equal(rightPartCost, 0) {
        return math.MaxFloat64
    }
    return math.Max(leftPartCost, rightPartCost)
}

func findPlane(objects []geometry.IGeometryObject, boundingBox *geometry.BBox) BBoxSplit {
    type ObjectLocationType int
    const (
        END ObjectLocationType = 0
        BELONGS ObjectLocationType = 1
        BEGIN ObjectLocationType = 2
    )
    type IntersectionEvent struct {
        locType ObjectLocationType
        value float64
    }

    var split BBoxSplit
    minCost := math.MaxFloat64

    for dim := 0; dim < 3; dim++ {
        events := make([]IntersectionEvent, 0)

        for _, obj := range objects {
            bbox := obj.GetBoundingBox()
            if primitives.Equal(bbox.GetMin(dim), bbox.GetMax(dim)) {
                events = append(events, IntersectionEvent{BELONGS, bbox.GetMin(dim)})
            } else {
                events = append(events, IntersectionEvent{BEGIN, bbox.GetMin(dim)})
                events = append(events, IntersectionEvent{END, bbox.GetMax(dim)})
            }
        }

        sort.Slice(events, func(i, j int) bool {
            return primitives.Less(events[i].value, events[j].value) ||
                (primitives.Equal(events[i].value, events[j].value) && events[i].locType < events[j].locType)
        })

        var leftSize, middleSize, rightSize = 0, 0, len(objects)

        for i := 0; i < len(objects); i++ {
            var pBegin, pBelong, pEnd = 0, 0, 0
            currentPlaneValue := events[i].value

            for ; i < len(events) && primitives.Equal(currentPlaneValue, events[i].value); i++ {
                switch events[i].locType {
                case BEGIN:
                    pBegin++
                case BELONGS:
                    pBelong++
                case END:
                    pEnd++
                }
            }
            i--
            middleSize = pBelong
            rightSize -= pBelong + pEnd

            currentResult := surfaceAreaHeuristic(currentPlaneValue, dim, boundingBox, leftSize, middleSize, rightSize)
            if primitives.Less(currentResult, minCost) {
                minCost = currentResult
                split.value = currentPlaneValue
                split.axis = dim
                split.cost = minCost
            }
            leftSize += pBegin + pBelong
        }
    }
    return split
}

func classify(objects []geometry.IGeometryObject, split BBoxSplit) [2][]geometry.IGeometryObject {
    leftPart := make([]geometry.IGeometryObject, 0)
    rightPart := make([]geometry.IGeometryObject, 0)
    for _, obj := range objects {
        bbox := obj.GetBoundingBox()
        if primitives.Less(bbox.GetMin(split.axis), split.value +primitives.EPS) {
            leftPart = append(leftPart, obj)
        }
        if primitives.Less(split.value -primitives.EPS, bbox.GetMax(split.axis)) {
            rightPart = append(rightPart, obj)
        }
    }
    return [2][]geometry.IGeometryObject{leftPart, rightPart}
}

func recBuild(input kdTreeNodeInput, sync chan int) {
    root, objects, bbox := input.node, input.objects, input.bbox
    root.bbox = bbox
    root.nodeSize = len(objects)

    split := findPlane(objects, bbox)

    if primitives.GreaterEqual(split.cost, float64(len(objects) * INTERSECTION_COEF)) {
        root.objects = objects
        sync <- -1
        return
    }

    boxes := bbox.Split(split.axis, split.value)
    parts := classify(objects, split)

    root.splitPlane = SplitPlane{split.value, split.axis}

    root.left = new(KDTreeNode)
    root.right = new(KDTreeNode)
    sync <- 2
    go recBuild(kdTreeNodeInput{root.left, parts[0], boxes[0]}, sync)
    go recBuild(kdTreeNodeInput{root.right, parts[1], boxes[1]}, sync)
    sync <- -1
    return
}

func findIntersection(node *KDTreeNode, ray *geometry.Ray) geometry.Intersection {
    if node.nodeSize == 0 || !node.bbox.Intersect(ray)[0].HasIntersection {
        return geometry.Intersection{}
    }
    if len(node.objects) != 0 {
        var intersection geometry.Intersection
        currentCoef := math.MaxFloat64

        for _, obj := range node.objects {
            objIntersection := obj.Intersect(ray)
            if objIntersection.HasIntersection && primitives.Less(objIntersection.IntersectionCoef, currentCoef) &&
                primitives.Greater(objIntersection.IntersectionCoef, 0) {
                currentCoef = objIntersection.IntersectionCoef

                intersection = geometry.Intersection{
                    Coefficient: geometry.RayCoefIntersection{IntersectionCoef: currentCoef, HasIntersection: true},
                    Point:       ray.Begin.Add(ray.Direction.Mult(currentCoef)),
                    Object:      obj,
                }
            }
        }
        return intersection
    }
    left := findIntersection(node.left, ray)
    right := findIntersection(node.right, ray)
    if left.Coefficient.HasIntersection {
        if right.Coefficient.HasIntersection && primitives.Greater(left.Coefficient.IntersectionCoef, right.Coefficient.IntersectionCoef) {
            return right
        }
        return left
    }
    return right
}

func (tree *KDTree) BuildTree(objects []geometry.IGeometryObject) {
    buildingBegin := time.Now()
    sync := make(chan int)
    tree.root = new(KDTreeNode)
    go recBuild(kdTreeNodeInput{tree.root, objects, getBoundingBox(objects)}, sync)
    sum := 1
    for obj := range sync {
        sum += obj
        if sum == 0 {
            // All nodes already constructed
            break
        }
    }
    buildingEnd := time.Now()
    tree.TotalBuildingTime = buildingEnd.Sub(buildingBegin)
}

func (tree *KDTree) CastRay(ray *geometry.Ray) geometry.Intersection {
    if !tree.root.bbox.Intersect(ray)[0].HasIntersection {
        return geometry.Intersection{}
    }
    return findIntersection(tree.root, ray)
}

