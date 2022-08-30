package scene

import (
    "ray-tracing/primitives"
)

type Viewport struct {
    Origin, TopLeft, BottomLeft, TopRight primitives.Vector
    Width, Height                         int
}

func (view *Viewport) GetWidthBase() primitives.Vector {
    return view.TopRight.Sub(view.TopLeft)
}

func (view *Viewport) GetHeightBase() primitives.Vector {
    return view.BottomLeft.Sub(view.TopLeft)
}