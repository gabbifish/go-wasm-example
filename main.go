package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"reflect"
	"syscall/js"
	"unsafe"
)

var imgBuf []uint8

func initMem(this js.Value, args []js.Value) interface{} {
	length := args[0].Int()
	imgBuf = make([]uint8, length)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&imgBuf))
	ptr := uintptr(unsafe.Pointer(hdr.Data))
	return int(ptr)
}

func makeGrayScale(imgPtr *image.Image) *image.Gray {
	imgSrc := *imgPtr
	// Create a new grayscale image
	bounds := imgSrc.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	grayScale := image.NewGray(
		image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{w, h},
		})
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			imageColor := imgSrc.At(x, y)
			rr, gg, bb, _ := imageColor.RGBA()
			r := math.Pow(float64(rr), 2.2)
			g := math.Pow(float64(gg), 2.2)
			b := math.Pow(float64(bb), 2.2)
			m := math.Pow(0.2125*r+0.7154*g+0.0721*b, 1/2.2)
			Y := uint16(m + 0.5)
			grayColor := color.Gray{uint8(Y >> 8)}
			grayScale.Set(x, y, grayColor)
		}
	}

	return grayScale
}

func typedArrayToByteSlice(arg js.Value) []byte {
	length := arg.Length()
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(arg.Index(i).Int())
	}
	return bytes
}

func processImage(this js.Value, args []js.Value) interface{} {
	//var []index
	log.Println(args[0])
	r := bytes.NewReader(typedArrayToByteSlice(args[0]))
	imgSrc, _, err := image.Decode(r)

	if err != nil {
		log.Fatal("Failed to decode image --", err)
	}

	grayScaleImgSrc := makeGrayScale(&imgSrc)

	w := new(bytes.Buffer)

	png.Encode(w, grayScaleImgSrc)
	out := w.Bytes()

	//hdr := (*reflect.SliceHeader)(unsafe.Pointer(&out))
	//ptr := uintptr(unsafe.Pointer(hdr.Data))

	//return []interface{}{int(ptr), len(out)}
	return js.TypedArrayOf([]uint8(out))
}

func registerCallbacks() {
	//js.Global().Set("initMem", js.FuncOf(initMem))
	js.Global().Set("processImage", js.FuncOf(processImage))
}

func main() {
	done := make(chan struct{}, 0)

	log.Println("WASM Initialized")
	//registerCallbacks()
	processImageFunc := js.FuncOf(processImage)
	js.Global().Set("processImage", processImageFunc)

	defer processImageFunc.Release()
	<-done
}
