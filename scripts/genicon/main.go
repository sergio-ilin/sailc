package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

// Generates high quality multi-size icon for SailC
func main() {
	sizes := []int{256, 128, 64, 48, 32, 16}
	pngBlobs := make([][]byte, len(sizes))

	for i, sz := range sizes {
		img := drawIcon(sz)
		var buf bytes.Buffer
		_ = png.Encode(&buf, img)
		pngBlobs[i] = buf.Bytes()
	}

	_ = os.MkdirAll("assets", 0755)

	// Save main PNG (256x256)
	_ = os.WriteFile(filepath.Join("assets", "sailc.png"), pngBlobs[0], 0644)

	// Build .ico file containing all sizes
	icoBuf := new(bytes.Buffer)
	// ICONDIR header
	binary.Write(icoBuf, binary.LittleEndian, uint16(0)) // Reserved
	binary.Write(icoBuf, binary.LittleEndian, uint16(1)) // ICO type
	binary.Write(icoBuf, binary.LittleEndian, uint16(len(sizes)))

	offset := uint32(6 + len(sizes)*16)
	for i, sz := range sizes {
		w := byte(sz)
		h := byte(sz)
		if sz == 256 {
			w = 0
			h = 0
		}
		binary.Write(icoBuf, binary.LittleEndian, w)
		binary.Write(icoBuf, binary.LittleEndian, h)
		binary.Write(icoBuf, binary.LittleEndian, byte(0)) // color count
		binary.Write(icoBuf, binary.LittleEndian, byte(0)) // reserved
		binary.Write(icoBuf, binary.LittleEndian, uint16(1))  // planes
		binary.Write(icoBuf, binary.LittleEndian, uint16(32)) // bit count
		binary.Write(icoBuf, binary.LittleEndian, uint32(len(pngBlobs[i])))
		binary.Write(icoBuf, binary.LittleEndian, offset)
		offset += uint32(len(pngBlobs[i]))
	}

	for _, blob := range pngBlobs {
		icoBuf.Write(blob)
	}

	_ = os.WriteFile(filepath.Join("assets", "sailc.ico"), icoBuf.Bytes(), 0644)
	println("Successfully generated assets/sailc.ico and assets/sailc.png")
}

func drawIcon(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)

	// Rounded squircle background
	cornerRadius := s * 0.22
	bgDark := color.RGBA{R: 0x0E, G: 0x11, B: 0x17, A: 0xFF}
	bgBlue := color.RGBA{R: 0x16, G: 0x24, B: 0x3E, A: 0xFF}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			if insideSquircle(fx, fy, s, cornerRadius) {
				t := (fx + fy) / (s * 2)
				r := uint8(float64(bgDark.R)*(1-t) + float64(bgBlue.R)*t)
				g := uint8(float64(bgDark.G)*(1-t) + float64(bgBlue.G)*t)
				b := uint8(float64(bgDark.B)*(1-t) + float64(bgBlue.B)*t)
				img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 0xFF})
			}
		}
	}

	// Draw stylized vector Sail
	// Main Sail: triangle (0.22*s, 0.85*s) -> (0.48*s, 0.16*s) -> (0.48*s, 0.80*s)
	p1A := pt(0.22*s, 0.82*s)
	p1B := pt(0.48*s, 0.16*s)
	p1C := pt(0.48*s, 0.80*s)
	colSailMain := color.RGBA{R: 0x3B, G: 0x82, B: 0xF6, A: 0xFF} // Vivid Azure
	fillTriangle(img, p1A, p1B, p1C, colSailMain)

	// Jib / Right Sail: triangle (0.55*s, 0.80*s) -> (0.55*s, 0.35*s) -> (0.80*s, 0.80*s)
	p2A := pt(0.55*s, 0.80*s)
	p2B := pt(0.55*s, 0.35*s)
	p2C := pt(0.80*s, 0.80*s)
	colSailJib := color.RGBA{R: 0x60, G: 0xA5, B: 0xFA, A: 0xFF} // Sky Blue
	fillTriangle(img, p2A, p2B, p2C, colSailJib)

	// Base Hull: trapezoid (0.18*s, 0.86*s) -> (0.82*s, 0.86*s) -> (0.72*s, 0.93*s) -> (0.28*s, 0.93*s)
	colHull := color.RGBA{R: 0x93, G: 0xC5, B: 0xFD, A: 0xFF}
	h1 := pt(0.18*s, 0.86*s)
	h2 := pt(0.82*s, 0.86*s)
	h3 := pt(0.72*s, 0.93*s)
	h4 := pt(0.28*s, 0.93*s)
	fillTriangle(img, h1, h2, h3, colHull)
	fillTriangle(img, h1, h3, h4, colHull)

	return img
}

type point struct {
	x, y float64
}

func pt(x, y float64) point {
	return point{x, y}
}

func insideSquircle(x, y, size, r float64) bool {
	margin := 2.0
	minX, minY := margin, margin
	maxX, maxY := size-margin, size-margin

	if x < minX || x > maxX || y < minY || y > maxY {
		return false
	}

	// Check 4 corners
	cx1, cy1 := minX+r, minY+r
	cx2, cy2 := maxX-r, minY+r
	cx3, cy3 := minX+r, maxY-r
	cx4, cy4 := maxX-r, maxY-r

	if x < cx1 && y < cy1 {
		return math.Hypot(x-cx1, y-cy1) <= r
	}
	if x > cx2 && y < cy2 {
		return math.Hypot(x-cx2, y-cy2) <= r
	}
	if x < cx3 && y > cy3 {
		return math.Hypot(x-cx3, y-cy3) <= r
	}
	if x > cx4 && y > cy4 {
		return math.Hypot(x-cx4, y-cy4) <= r
	}
	return true
}

func sign(p1, p2, p3 point) float64 {
	return (p1.x-p3.x)*(p2.y-p3.y) - (p2.x-p3.x)*(p1.y-p3.y)
}

func pointInTriangle(pt, v1, v2, v3 point) bool {
	d1 := sign(pt, v1, v2)
	d2 := sign(pt, v2, v3)
	d3 := sign(pt, v3, v1)
	hasNeg := (d1 < 0) || (d2 < 0) || (d3 < 0)
	hasPos := (d1 > 0) || (d2 > 0) || (d3 > 0)
	return !(hasNeg && hasPos)
}

func fillTriangle(img *image.RGBA, v1, v2, v3 point, col color.RGBA) {
	minX := int(math.Floor(math.Min(v1.x, math.Min(v2.x, v3.x))))
	maxX := int(math.Ceil(math.Max(v1.x, math.Max(v2.x, v3.x))))
	minY := int(math.Floor(math.Min(v1.y, math.Min(v2.y, v3.y))))
	maxY := int(math.Ceil(math.Max(v1.y, math.Max(v2.y, v3.y))))

	bounds := img.Bounds()
	if minX < bounds.Min.X {
		minX = bounds.Min.X
	}
	if maxX > bounds.Max.X {
		maxX = bounds.Max.X
	}
	if minY < bounds.Min.Y {
		minY = bounds.Min.Y
	}
	if maxY > bounds.Max.Y {
		maxY = bounds.Max.Y
	}

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			p := point{float64(x) + 0.5, float64(y) + 0.5}
			if pointInTriangle(p, v1, v2, v3) {
				img.SetRGBA(x, y, col)
			}
		}
	}
}
