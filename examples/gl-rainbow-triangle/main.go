package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	// OR: github.com/go-gl/gl/v2.1/gl
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	float32Size = 4

	windowTitle = "OpenGL Rainbow Triangle"

	width, height = 800, 600

	// Use `#version 120` when use OpenGL 2.1
	vertexShaderSource = `
    #version 410

    layout (location = 0) in vec3 position;
    layout (location = 1) in vec3 color;
    
    out vec3 vertexColor;

    void main() {
        gl_Position = vec4(position, 1.0);
        vertexColor = color;
    }
` + "\x00"

	fragmentShaderSource = `
    #version 410

    out vec4 fragColor;
    in vec3 vertexColor;

    void main() {
        fragColor = vec4(vertexColor, 1);
    }
` + "\x00"
)

var (
	triangle = []float32{
		//  X, Y, Z, R, G, B
		0, 0.5, 0, 1, 0, 0, // top
		-0.5, -0.5, 0, 0, 1, 0, // left
		0.5, -0.5, 0, 0, 0, 1, // right
	}
)

func main() {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	program := newProgram(vertexShaderSource, fragmentShaderSource)

	vao := makeVao(triangle)

	for !window.ShouldClose() {
		drawTriangle(vao, program)

		glfw.PollEvents()
		window.SwapBuffers()
	}
}

// initGlfw initializes glfw and returns a Window to use.
func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, windowTitle, nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

// newProgram initializes OpenGL and returns an intiialized program.
func newProgram(vertexShaderSource, fragmentShaderSource string) uint32 {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return prog
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

// makeVao initializes and returns a vertex array from the points provided.
func makeVao(points []float32) uint32 {
	var vbo uint32 // Vertex Buffer Object
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32 // Vertex Array Object
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Enable vertex attribute and use points FLOAT array group by number of 3 as coords
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 6*float32Size, 0)

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 3, gl.FLOAT, false, 6*float32Size, 3*float32Size)

	return vao
}

func drawTriangle(vao uint32, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangle)/3))
}
