package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"ray-tracing/scene"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
)

func MaxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

type options struct {
	Filename string `long:"config" required:"true"`
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default&flags.HelpFlag)
	if _, err := parser.Parse(); err != nil {
		panic(err)
	}
	fmt.Println(MaxParallelism())
	curScene, err := scene.OpenScene(opts.Filename)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Initialisation: %.4fs\n", curScene.KDTree.TotalBuildingTime.Seconds())
	renderBegin := time.Now()
	curScene.Render()
	fmt.Println("Waiting")
	curScene.Wg.Wait()
	renderEnd := time.Now()
	fmt.Printf("Render time: %.4fs\n", renderEnd.Sub(renderBegin).Seconds())
	result := image.NewRGBA(image.Rect(0, 0, curScene.Viewport.Width, curScene.Viewport.Height))
	pixels := curScene.Pixels
	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[x]); y++ {
			result.SetRGBA(x, y, color.RGBA{
				R: uint8(255 * pixels[x][y].R),
				G: uint8(255 * pixels[x][y].G),
				B: uint8(255 * pixels[x][y].B),
				A: 255,
			})
		}
	}
	_ = os.Mkdir("results", os.ModePerm)
	writer, _ := os.Create("results/res.png")
	err = png.Encode(writer, result)
	if err != nil {
		panic(err)
	}
	_ = writer.Close()
}
