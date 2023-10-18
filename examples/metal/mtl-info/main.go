// go:build darwin

// mtlinfo is a tool that displays information about Metal devices in the system.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hychul/gopen/internal/graphics/metal/mtl"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: mtlinfo")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	// Display the preferred system default Metal device.
	device, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		// An error here means Metal is not supported on this system.
		// Let the user know and stop here.
		fmt.Println(err)
		return nil
	}
	fmt.Println("preferred system default Metal device:", device.Name)

	// List all Metal devices in the system.
	allDevices := mtl.CopyAllDevices()
	for _, d := range allDevices {
		fmt.Println()
		printDeviceInfo(d)
	}

	return nil
}

func printDeviceInfo(d mtl.Device) {
	fmt.Println(d.Name + ":")
	fmt.Println("	• low-power:", yes(d.LowPower))
	fmt.Println("	• removable:", yes(d.Removable))
	fmt.Println("	• configured as headless:", yes(d.Headless))
	fmt.Println("	• registry ID:", d.RegistryID)
	fmt.Println()
	fmt.Println("	GPU Families:")
	fmt.Println("	• An Apple family 1 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple1)))
	fmt.Println("	• An Apple family 2 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple2)))
	fmt.Println("	• An Apple family 3 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple3)))
	fmt.Println("	• An Apple family 4 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple4)))
	fmt.Println("	• An Apple family 5 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple5)))
	fmt.Println("	• An Apple family 6 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple6)))
	fmt.Println("	• An Apple family 7 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple7)))
	fmt.Println("	• An Apple family 8 GPU:", supported(d.SupportsFamily(mtl.GPUFamilyApple8)))
	fmt.Println("	• A family 1 Mac GPU:", supported(d.SupportsFamily(mtl.GPUFamilyMac1)))
	fmt.Println("	• A family 2 Mac GPU:", supported(d.SupportsFamily(mtl.GPUFamilyMac2)))
	fmt.Println("	• Common family 1:", supported(d.SupportsFamily(mtl.GPUFamilyCommon1)))
	fmt.Println("	• Common family 2:", supported(d.SupportsFamily(mtl.GPUFamilyCommon2)))
	fmt.Println("	• Common family 3:", supported(d.SupportsFamily(mtl.GPUFamilyCommon3)))
	fmt.Println("	• Metal 3 features:", supported(d.SupportsFamily(mtl.GPUFamilyMetal3)))
	fmt.Println()
	fmt.Println("	Feature Sets (deprecated):")
	fmt.Println("	• macOS GPU family 1, version 1:", supported(d.SupportsFeatureSet(mtl.MacOSGPUFamily1V1)))
	fmt.Println("	• macOS GPU family 1, version 2:", supported(d.SupportsFeatureSet(mtl.MacOSGPUFamily1V2)))
	fmt.Println("	• macOS read-write texture, tier 2:", supported(d.SupportsFeatureSet(mtl.MacOSReadWriteTextureTier2)))
	fmt.Println("	• macOS GPU family 1, version 3:", supported(d.SupportsFeatureSet(mtl.MacOSGPUFamily1V3)))
	fmt.Println("	• macOS GPU family 1, version 4:", supported(d.SupportsFeatureSet(mtl.MacOSGPUFamily1V4)))
	fmt.Println("	• macOS GPU family 2, version 1:", supported(d.SupportsFeatureSet(mtl.MacOSGPUFamily2V1)))
}

func yes(v bool) string {
	switch v {
	case true:
		return "yes"
	case false:
		return "no"
	}
	panic("unreachable")
}

func supported(v bool) string {
	switch v {
	case true:
		return "✅ supported"
	case false:
		return "❌ unsupported"
	}
	panic("unreachable")
}
