// go:build darwin

// movingtriangle is an example Metal program that displays a moving triangle in a window.
// It opens a window and renders a triangle that follows the mouse cursor.
package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/hychul/gopen/examples/metal/mtl"
	"github.com/hychul/gopen/examples/metal/mtl/appkit"
	"github.com/hychul/gopen/examples/metal/mtl/coreanim"
	"golang.org/x/image/math/f32"
)

const (
	float32Size = 4

	windowTitle = "Metal Rainbow Triangle"

	windowWidth, windowHeight = 800, 600

	metalShaderSource = `
    #include <metal_stdlib>

    using namespace metal;

    struct Vertex {
        float4 position [[position]];
        float4 color;
    };

    vertex Vertex VertexShader(
        uint vertexID [[vertex_id]],
        device Vertex * vertices [[buffer(0)]],
        constant int2 * windowSize [[buffer(1)]],
        constant float2 * pos [[buffer(2)]]
    ) {
        Vertex out = vertices[vertexID];
        out.position.xy += *pos;
        float2 viewportSize = float2(*windowSize);
        out.position.xy = float2(-1 + out.position.x / (0.5 * viewportSize.x),
                                1 - out.position.y / (0.5 * viewportSize.y));
        return out;
    }

    fragment float4 FragmentShader(Vertex in [[stage_in]]) {
        return in.color;
    }
    `
)

type Vertex struct {
	Position f32.Vec4
	Color    f32.Vec4
}

var (
	vertices = [...]Vertex{
		{f32.Vec4{0, -100, 0, 1}, f32.Vec4{1, 0, 0, 1}},
		{f32.Vec4{-100, 100, 0, 1}, f32.Vec4{0, 1, 0, 1}},
		{f32.Vec4{100, 100, 0, 1}, f32.Vec4{0, 0, 1, 1}},
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

	// Create MTL command queue
	cq := device.MakeCommandQueue()

	fpsCounter := startFPSCounter()

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
		rce.DrawPrimitives(mtl.PrimitiveTypeTriangle, 0, 3)
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

func startFPSCounter() chan struct{} {
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
