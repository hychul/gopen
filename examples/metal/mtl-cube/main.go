package main

import (
	"fmt"
	"go/build"
	"image"
	"image/draw"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hychul/gopen/internal/graphics/metal/appkit"
	"github.com/hychul/gopen/internal/graphics/metal/coreanim"
	"github.com/hychul/gopen/internal/graphics/metal/mtl"
)

const (
	windowTitle = "Metal Cube"

	windowWidth, windowHeight = 800, 600

	metalShaderSource = `
    #include <metal_stdlib>

    using namespace metal;

    struct Vertex {
        float4 position [[ position ]];
		float4 color;
        float2 texCoord;
    };

    vertex Vertex VertexShader(
        uint vertexID [[ vertex_id ]],
        const device Vertex * vertices [[ buffer(0) ]],
        constant int2 * windowSize [[ buffer(1) ]],
        constant float2 * pos [[ buffer(2) ]]
    ) {
        Vertex out;

		out.position = vertices[vertexID].position;
		out.color = vertices[vertexID].color;
		out.texCoord = vertices[vertexID].texCoord;

        // out.position.xy += *pos;
        // float2 viewportSize = float2(*windowSize);
        // out.position.xy = float2(-1 + out.position.x / (0.5 * viewportSize.x),
        //                          1 - out.position.y / (0.5 * viewportSize.y));
		
        return out;
    }

    fragment float4 FragmentShader(
        Vertex in [[ stage_in ]],
        texture2d<float, access::sample> texture [[ texture(0) ]]
    ) {
        constexpr sampler textureSampler(mip_filter::linear, mag_filter::linear, min_filter::linear);
        return texture.sample(textureSampler, in.texCoord);

		// return in.color;
    }
    `
)

type Vertex struct {
	Position [4]float32 // x, y, z, w
	Color    [4]float32 // r, g, b, a
	TexCoord [4]float32 // u, v, _, _
}

var (
	vertices = [...]Vertex{
		//              X,     Y,     Z,     W                  R,     G,     B,     A                  U,     V
		{[4]float32{+0.00, +1.00, +0.00, +1.00}, [4]float32{+1.00, +0.00, +0.00, +1.00}, [4]float32{+1.00, +0.00}}, // top
		{[4]float32{-0.75, -0.75, +0.00, +1.00}, [4]float32{+0.00, +1.00, +0.00, +1.00}, [4]float32{+1.00, +1.00}}, // left
		{[4]float32{+0.75, -0.75, +0.00, +1.00}, [4]float32{+0.00, +0.00, +1.00, +1.00}, [4]float32{+0.00, +1.00}}, // right
	}
	// vertices = [...]float32{
	// 	//  X,     Y,     Z,     W,     R,     G,     B,     A,     U,     V,     _,     _
	// 	+0.00, +1.00, +0.00, +1.00, +1.00, +1.00, +1.00, +1.00, +1.00, +0.00, +0.00, +0.00, // top
	// 	-0.75, -0.75, +0.00, +1.00, +1.00, +1.00, +1.00, +1.00, +1.00, +1.00, +0.00, +0.00, // left
	// 	+0.75, -0.75, +0.00, +1.00, +1.00, +1.00, +1.00, +1.00, +0.00, +1.00, +0.00, +0.00, // right
	// }

	indices = [...]uint16{
		0, 1, 2,
	}

	cubeVertices = [...]float32{
		//  X, Y, Z, U, V
		// Bottom
		-1.0, -1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, -1.0, 1.0, 0.0,
		-1.0, -1.0, 1.0, 0.0, 1.0,
		1.0, -1.0, -1.0, 1.0, 0.0,
		1.0, -1.0, 1.0, 1.0, 1.0,
		-1.0, -1.0, 1.0, 0.0, 1.0,

		// Top
		-1.0, 1.0, -1.0, 0.0, 0.0,
		-1.0, 1.0, 1.0, 0.0, 1.0,
		1.0, 1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, -1.0, 1.0, 0.0,
		-1.0, 1.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0, 1.0,

		// Front
		-1.0, -1.0, 1.0, 1.0, 0.0,
		1.0, -1.0, 1.0, 0.0, 0.0,
		-1.0, 1.0, 1.0, 1.0, 1.0,
		1.0, -1.0, 1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 0.0, 1.0,
		-1.0, 1.0, 1.0, 1.0, 1.0,

		// Back
		-1.0, -1.0, -1.0, 0.0, 0.0,
		-1.0, 1.0, -1.0, 0.0, 1.0,
		1.0, -1.0, -1.0, 1.0, 0.0,
		1.0, -1.0, -1.0, 1.0, 0.0,
		-1.0, 1.0, -1.0, 0.0, 1.0,
		1.0, 1.0, -1.0, 1.0, 1.0,

		// Left
		-1.0, -1.0, 1.0, 0.0, 1.0,
		-1.0, 1.0, -1.0, 1.0, 0.0,
		-1.0, -1.0, -1.0, 0.0, 0.0,
		-1.0, -1.0, 1.0, 0.0, 1.0,
		-1.0, 1.0, 1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0, 1.0, 0.0,

		// Right
		1.0, -1.0, 1.0, 1.0, 1.0,
		1.0, -1.0, -1.0, 1.0, 0.0,
		1.0, 1.0, -1.0, 0.0, 0.0,
		1.0, -1.0, 1.0, 1.0, 1.0,
		1.0, 1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 0.0, 1.0,
	}
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	device, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		return err
	}
	fmt.Println("Metal device:", device.Name)

	// GLFW setting
	err = glfw.Init()
	if err != nil {
		return err
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		return err
	}
	defer window.Destroy()

	// Metal setting
	ml := coreanim.MakeMetalLayer()
	ml.SetDevice(device)
	ml.SetPixelFormat(mtl.PixelFormatBGRA8UNorm)
	ml.SetDrawableSize(window.GetFramebufferSize())
	ml.SetMaximumDrawableCount(3)
	ml.SetDisplaySyncEnabled(true)
	cv := appkit.NewWindow(window.GetCocoaWindow()).ContentView()
	cv.SetLayer(ml)
	cv.SetWantsLayer(true)

	// Set window callbacks
	window.SetFramebufferSizeCallback(func(_ *glfw.Window, width, height int) {
		ml.SetDrawableSize(width, height)
	})
	var windowSize = [2]int32{windowWidth, windowHeight}
	window.SetSizeCallback(func(_ *glfw.Window, width, height int) {
		windowSize[0], windowSize[1] = int32(width), int32(height)
	})
	var pos [2]float32
	window.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		pos[0], pos[1] = float32(x), float32(y)
	})

	// Load Texture
	file := "diamond_ore.png"
	imgFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("texture %q not found on disk: %v", file, err)
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return err
	}

	sz := img.Bounds()

	nrgba := imageToNRGBA(img)

	td := mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatRGBA8UNorm,
		Width:       sz.Max.X - sz.Min.X,
		Height:      sz.Max.Y - sz.Min.Y,
	}
	texture := device.MakeTexture(td)
	region := mtl.RegionMake2D(0, 0, texture.Width, texture.Height)
	bytesPerRow := 4 * texture.Width
	texture.ReplaceRegion(region, 0, &nrgba.Pix[0], uintptr(bytesPerRow))

	// Compile shader
	lib, err := device.MakeLibrary(metalShaderSource, mtl.CompileOptions{})
	if err != nil {
		return err
	}
	vs, err := lib.MakeFunction("VertexShader")
	if err != nil {
		return err
	}
	fs, err := lib.MakeFunction("FragmentShader")
	if err != nil {
		return err
	}
	var rpld mtl.RenderPipelineDescriptor
	rpld.VertexFunction = vs
	rpld.FragmentFunction = fs
	rpld.ColorAttachments[0].PixelFormat = ml.PixelFormat()
	rps, err := device.MakeRenderPipelineState(rpld)
	if err != nil {
		return err
	}

	// Make vertex buffer
	vertexBuffer := device.MakeBuffer(unsafe.Pointer(&vertices[0]), unsafe.Sizeof(vertices), mtl.ResourceStorageModeManaged)
	indexBuffer := device.MakeBuffer(unsafe.Pointer(&indices[0]), unsafe.Sizeof(indices), mtl.ResourceStorageModeManaged)

	// Create MTL command queue
	cq := device.MakeCommandQueue()

	fpsCounter := makeFPSCounter()

	for !window.ShouldClose() {
		// Create a drawable to render into
		drawable, err := ml.NextDrawable()
		if err != nil {
			panic(err)
		}

		// Create command buffer
		cb := cq.MakeCommandBuffer()

		// Encode all render commands
		var rpd mtl.RenderPassDescriptor
		rpd.ColorAttachments[0].LoadAction = mtl.LoadActionClear
		rpd.ColorAttachments[0].StoreAction = mtl.StoreActionStore
		rpd.ColorAttachments[0].ClearColor = mtl.ClearColor{Red: 0, Green: 0, Blue: 0, Alpha: 1}
		rpd.ColorAttachments[0].Texture = drawable.Texture()
		rce := cb.MakeRenderCommandEncoder(rpd)
		rce.SetRenderPipelineState(rps)
		rce.SetVertexBuffer(vertexBuffer, 0, 0)
		rce.SetVertexBytes(unsafe.Pointer(&windowSize[0]), unsafe.Sizeof(windowSize), 1)
		rce.SetVertexBytes(unsafe.Pointer(&pos[0]), unsafe.Sizeof(pos), 2)
		rce.SetFragmentTexture(texture, 0)
		rce.DrawIndexedPrimitives(mtl.PrimitiveTypeTriangle, len(indices), mtl.IndexTypeUInt16, indexBuffer, 0) // TODO
		rce.EndEncoding()

		// Commit command
		cb.PresentDrawable(drawable)
		cb.Commit()

		// Count frame
		fpsCounter <- struct{}{}

		// Poll window events
		glfw.PollEvents()
	}

	return nil
}

func imageToNRGBA(src image.Image) *image.NRGBA {
	if dst, ok := src.(*image.NRGBA); ok {
		return dst
	}

	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func makeFPSCounter() chan struct{} {
	frame := make(chan struct{}, 4)
	go func() {
		second := time.Tick(time.Second)
		frames := 0
		for {
			select {
			case <-second:
				fmt.Println("fps:", frames)
				frames = 0
			case <-frame:
				frames++
			}
		}
	}()
	return frame
}

func init() {
	dir, err := importPathToDir("github.com/hychul/gopen/examples/metal/mtl-cube")
	if err != nil {
		log.Fatalln("Unable to find Go package in your GOPATH, it's needed to load assets:", err)
	}
	err = os.Chdir(dir)
	if err != nil {
		log.Panicln("os.Chdir:", err)
	}
}

func importPathToDir(importPath string) (string, error) {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
