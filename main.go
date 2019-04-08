package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"syscall/js"
)

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

// There must be a better way to turn a JS array into its Go equivalent that
// doesn't involve extracting values element by element...
func typedArrayToByteSlice(arg js.Value) []byte {
	length := arg.Length()
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(arg.Index(i).Int())
	}
	return bytes
}

func processImage(this js.Value, args []js.Value) interface{} {
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

	return js.TypedArrayOf([]uint8(out))
}

func main() {
	done := make(chan struct{}, 0)

	log.Println("WASM Initialized")
	processImageFunc := js.FuncOf(processImage)
	js.Global().Set("processImage", processImageFunc)

	defer processImageFunc.Release()
	<-done
}
