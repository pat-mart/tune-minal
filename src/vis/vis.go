package vis

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"os/signal"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"gonum.org/v1/gonum/dsp/fourier"
)

const (
	sampleRate      = 22050
	framesPerBuffer = 1024
	numBars         = 64
)

var (
	mu       sync.Mutex
	spectrum []float64
	prevBars []float64
	fft      = fourier.NewFFT(framesPerBuffer)
	barColor = color.RGBA{0, 0, 255, 255}
	bgColor  = color.RGBA{10, 10, 20, 255}
	windowW  = 800
	windowH  = 400
)

// --------------------- AUDIO LOOP ------------------------
func startAudioCapture() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	buffer := make([]float32, framesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(buffer), &buffer)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	stream.Start()
	defer stream.Stop()

	for {
		if err := stream.Read(); err != nil {
			fmt.Println("Read error:", err)
			continue
		}

		data := make([]float64, len(buffer))
		for i, v := range buffer {
			data[i] = float64(v)
		}

		complexSpectrum := fft.Coefficients(nil, data)
		magnitudes := make([]float64, len(complexSpectrum)/2)
		for i := range magnitudes {
			re := real(complexSpectrum[i])
			im := imag(complexSpectrum[i])
			magnitudes[i] = math.Sqrt(re*re + im*im)
		}

		bars := make([]float64, numBars)
		n := len(magnitudes)
		chunk := n / numBars
		for i := 0; i < numBars; i++ {
			sum := 0.0
			for j := 0; j < chunk; j++ {
				k := i*chunk + j

				freq := float64(k) * float64(sampleRate) / float64(n)

				// reduces bass dominance
				gain := math.Sqrt(freq / 200.0)
				if gain < 1 {
					gain = 1
				}
				sum += magnitudes[k] * gain
			}
			avg := sum / float64(chunk)
			bars[i] = math.Log10(avg+1) * 30
		}

		if prevBars == nil {
			prevBars = make([]float64, len(bars))
		}

		alpha := 0.1

		for i := range bars {
			prevBars[i] = prevBars[i]*(1-alpha) + bars[i]*alpha
		}

		mu.Lock()
		spectrum = bars
		mu.Unlock()
	}
}

// Visualizer --------------------- GUI LOOP ------------------------
type Visualizer struct{}

func (v *Visualizer) Update() error {
	return nil
}

func (v *Visualizer) Draw(screen *ebiten.Image) {
	screen.Fill(bgColor)

	mu.Lock()
	bars := make([]float64, len(spectrum)/4)
	copy(bars, spectrum)
	mu.Unlock()

	// draws bars on screen
	barWidth := float64(windowW) / float64(len(bars))
	for i, val := range bars {
		h := val * 5
		if h > float64(windowH) {
			h = float64(windowH)
		}
		x := float64(i) * barWidth

		h = math.Round(h/5) * 5
		if h < 1 {
			h = 1
		}

		y := float64(windowH) - h

		rect := ebiten.NewImage(int(barWidth)-2, int(h))
		rect.Fill(barColor)

		var ops ebiten.DrawImageOptions
		ops.GeoM.Translate(float64(x), float64(y))
		screen.DrawImage(rect, &ebiten.DrawImageOptions{
			GeoM:       ops.GeoM,
			ColorScale: ops.ColorScale,
		})
	}

	ebitenutil.DebugPrint(screen, "pat-mart.com")
}

func (v *Visualizer) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowW, windowH
}

func Start() {
	// handles ctrl c
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Println("\nExiting...")
		os.Exit(0)
	}()

	go startAudioCapture()

	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowTitle("vis")
	if err := ebiten.RunGame(&Visualizer{}); err != nil {
		panic(err)
	}
}
