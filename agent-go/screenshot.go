package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"syscall"
	"unsafe"
)

var (
	user32                     = syscall.NewLazyDLL("user32.dll")
	gdi32                      = syscall.NewLazyDLL("gdi32.dll")
	procGetDC                  = user32.NewProc("GetDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
)

const (
	smCxScreen = 0
	smCyScreen = 1
	srccopy    = 0x00CC0020
	captureBlt = 0x40000000
	bi_rgb     = 0
)

type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type bitmapInfo struct {
	BmiHeader bitmapInfoHeader
	BmiColors [1]uint32
}

func CaptureScreen() (string, error) {
	width, _, _ := procGetSystemMetrics.Call(uintptr(smCxScreen))
	height, _, _ := procGetSystemMetrics.Call(uintptr(smCyScreen))
	w := int(width)
	h := int(height)

	if w == 0 || h == 0 {
		return "", fmt.Errorf("could not determine screen dimensions")
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return "", fmt.Errorf("GetDC failed")
	}
	defer procReleaseDC.Call(0, hdc)

	memDC, _, _ := procCreateCompatibleDC.Call(hdc)
	if memDC == 0 {
		return "", fmt.Errorf("CreateCompatibleDC failed")
	}
	defer procDeleteDC.Call(memDC)

	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	if hBitmap == 0 {
		return "", fmt.Errorf("CreateCompatibleBitmap failed")
	}
	defer procDeleteObject.Call(hBitmap)

	procSelectObject.Call(memDC, hBitmap)

	ret, _, _ := procBitBlt.Call(
		memDC, 0, 0, uintptr(w), uintptr(h),
		hdc, 0, 0,
		uintptr(srccopy|captureBlt),
	)
	if ret == 0 {
		return "", fmt.Errorf("BitBlt failed")
	}

	bi := bitmapInfo{}
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	bi.BmiHeader.BiWidth = int32(w)
	bi.BmiHeader.BiHeight = -int32(h)
	bi.BmiHeader.BiPlanes = 1
	bi.BmiHeader.BiBitCount = 32
	bi.BmiHeader.BiCompression = bi_rgb

	buf := make([]byte, w*h*4)
	ret, _, _ = procGetDIBits.Call(
		memDC, hBitmap, 0, uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bi)),
		0,
	)
	if ret == 0 {
		return "", fmt.Errorf("GetDIBits failed")
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			b := buf[i]
			g := buf[i+1]
			r := buf[i+2]
			pi := img.PixOffset(x, y)
			img.Pix[pi] = r
			img.Pix[pi+1] = g
			img.Pix[pi+2] = b
			img.Pix[pi+3] = 255
		}
	}

	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 75}); err != nil {
		return "", fmt.Errorf("jpeg encode failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(jpegBuf.Bytes()), nil
}
