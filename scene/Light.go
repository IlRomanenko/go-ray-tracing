package scene

import (
    "ray-tracing/primitives"
)

type Reference struct {
    Power float64
    Distance float64
}

type Light struct {
    Ref      Reference
    Power    float64
    Position primitives.Vector
}
