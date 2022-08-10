package kindleland

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// This code works on the Kindle 4 with 8 bit gray values.

func NewFrameBuffer(device string, width, height int) (*FrameBuffer, error) {
	file, err := os.OpenFile(device, os.O_RDWR, 0)
	size := 480000
	defer file.Close()
	if err != nil {
		panic(err)
	}
	fd := int(file.Fd())
	fb, err := syscall.Mmap(fd, 0, size, syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	return &FrameBuffer{
		device: device,
		buffer: fb,
		Width:  width,
		Height: height,
	}, nil
}

// Pixel sets the value of the pixel at x, y.
func (fb *FrameBuffer) Pixel(x, y int, level uint8) error {
	offset := x + (y * fb.Width)
	if offset >= len(fb.buffer) {
		return fmt.Errorf("%d is out of range; max is %d; x: %d, y: %d", offset, len(fb.buffer)-1, x, y)
	}
	fb.buffer[offset] = byte(255 - level)
	return nil
}

type FrameBuffer struct {
	device string
	buffer []byte
	Width  int
	Height int
}

// ApplyImage copies each pixel value from img and places it at the same position in the framebuffer.
// img and the framebuffer are expected to have the same dimensions. No checks are done to verify this.
func (fb *FrameBuffer) ApplyImage(img image.Image) error {
	//colorModel := img.ColorModel()
	for y := 0; y < fb.Height; y++ {
		for x := 0; x < fb.Width; x++ {
			colorAt := img.At(x, y)
			err := fb.Pixel(x, y, color.GrayModel.Convert(colorAt).(color.Gray).Y)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ClearScreen clears the screen and empties the framebuffer
func (fb *FrameBuffer) ClearScreen() error {
	file, err := os.OpenFile(fb.device, os.O_WRONLY, 0)
	defer file.Close()
	if err != nil {
		return err
	}

	return unix.IoctlSetInt(int(file.Fd()), FBIOEinkClearScreen, 0)
}

// UpdateScreen flushes any changes to the framebuffer to the display.
func (fb *FrameBuffer) UpdateScreen() error {
	file, err := os.OpenFile(fb.device, os.O_WRONLY, 0)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = unix.IoctlGetInt(int(file.Fd()), FBIOEinkUpdateDisplay)
	if err != nil {
		return err
	}
	return nil
}

func (fb *FrameBuffer) UpdateScreenFx(mode UpdateMode) error {
	file, err := os.OpenFile(fb.device, os.O_WRONLY, 0)
	defer file.Close()
	if err != nil {
		return err
	}

	err = unix.IoctlSetPointerInt(int(file.Fd()), FBIOEinkUpdateDisplayFx, int(mode))
	if err != nil {
		return err
	}
	return nil
}

// At returns the value of the pixel at x, y.
func (fb *FrameBuffer) At(x, y int) (color.Gray, error) {
	offset := x + (y * fb.Width)
	if offset >= len(fb.buffer) {
		return color.Gray{}, fmt.Errorf("%d is out of range; max is %d; x: %d, y: %d", offset, len(fb.buffer)-1, x, y)
	}

	// Flip the bits so that 0 is black and 255 is white
	bits := 255 - uint8(fb.buffer[offset])

	return color.Gray{Y: bits}, nil
}

// Image returns an image.Image containing the current value of the framebuffer.Image
// This should be in sync with the framebuffer, but there is no guarantee that the screen and the framebuffer are
// in sync unless UpdateScreen() was just called.
func (fb *FrameBuffer) Image() image.Image {
	rect := image.Rect(0, 0, fb.Width, fb.Height)
	img := image.NewGray(rect)
	for y := 0; y < fb.Height; y++ {
		for x := 0; x < fb.Width; x++ {
			gray, _ := fb.At(x, y)
			img.SetGray(x, y, gray)
		}
	}
	return img
}
