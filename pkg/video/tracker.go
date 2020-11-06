package video

import (
	"fmt"
	"image/color"
	"io"
	"os/exec"
	"strconv"

	"gocv.io/x/gocv"
)

type tracker struct {
	data   <-chan []byte
	cancel chan interface{}
}

func NewTracker(data <-chan []byte, cancel chan interface{}) *tracker {
	if data == nil || cancel == nil {
		panic("tracker: args not initialized!")
	}
	return &tracker{
		data:   data,
		cancel: cancel,
	}
}

const (
	frameX    = 400
	frameY    = 300
	frameSize = frameX * frameY * 3
)

func (t *tracker) Start() {
	// ffmpeg command to decode video stream from drone
	ffmpeg := exec.Command("ffmpeg", "-hwaccel", "auto", "-hwaccel_device", "opencl", "-i", "pipe:0",
		"-pix_fmt", "bgr24", "-s", strconv.Itoa(frameX)+"x"+strconv.Itoa(frameY), "-f", "rawvideo", "pipe:1")
	ffmpegIn, _ := ffmpeg.StdinPipe()
	ffmpegOut, _ := ffmpeg.StdoutPipe()

	window := gocv.NewWindow("Tello")
	defer window.Close()

	if err := ffmpeg.Start(); err != nil {
		fmt.Printf("tracker: error starting stream: %v", err)
		close(t.cancel)
		return
	}

	haarcascade := "/usr/local/share/opencv4/haarcascades/haarcascade_frontalcatface.xml"
	classifier := gocv.NewCascadeClassifier()
	classifier.Load(haarcascade)
	defer classifier.Close()

	color := color.RGBA{0, 255, 0, 0}
	go func() {
		for {
			// get next frame from stream
			buf := make([]byte, frameSize)
			if _, err := io.ReadFull(ffmpegOut, buf); err != nil {
				fmt.Printf("tracker: error reading stream: %v", err)
				continue
			}
			img, err := gocv.NewMatFromBytes(frameY, frameX, gocv.MatTypeCV8UC3, buf)
			if err != nil {
				fmt.Printf("tracker: error get mat: %v", err)
			}
			if img.Empty() {
				continue
			}
			rects := classifier.DetectMultiScale(img)
			for _, r := range rects {
				fmt.Println("detected", r)
				gocv.Rectangle(&img, r, color, 3)
			}

			window.IMShow(img)
			window.WaitKey(50)
		}
	}()

videoInLoop:
	for {
		select {
		case _, ok := <-t.cancel:
			if !ok {
				fmt.Printf("mplayer: cancelled!\n")
				_ = ffmpeg.Process.Kill()
				break videoInLoop
			}
		case data := <-t.data:
			if _, err := ffmpegIn.Write(data); err != nil {
				fmt.Println(err)
			}
		}
	}

	fmt.Println("mplayer: leaving loop!")
}
