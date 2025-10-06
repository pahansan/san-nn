package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	Cols     = 28
	Rows     = 28
	CellSize = 10
	GridW    = Cols * CellSize
	GridH    = Rows * CellSize
)

type PixelGrid struct {
	widget.BaseWidget
	lock sync.Mutex

	cells   [Rows][Cols]bool
	img     *canvas.Image
	imgRGBA *image.RGBA

	onChange func()
}

func NewPixelGrid() *PixelGrid {
	g := &PixelGrid{}
	g.ExtendBaseWidget(g)
	g.imgRGBA = image.NewRGBA(image.Rect(0, 0, GridW, GridH))
	draw.Draw(g.imgRGBA, g.imgRGBA.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	g.img = canvas.NewImageFromImage(g.imgRGBA)
	g.img.FillMode = canvas.ImageFillContain
	return g
}

func (g *PixelGrid) CreateRenderer() fyne.WidgetRenderer {
	return &gridRenderer{grid: g, objects: []fyne.CanvasObject{g.img}}
}

func (g *PixelGrid) setCell(r, c int) {
	if r < 0 || r >= Rows || c < 0 || c >= Cols {
		return
	}
	g.lock.Lock()
	defer g.lock.Unlock()
	if g.cells[r][c] {
		return
	}
	g.cells[r][c] = true
	x0 := c * CellSize
	y0 := r * CellSize
	for y := y0; y < y0+CellSize; y++ {
		for x := x0; x < x0+CellSize; x++ {
			g.imgRGBA.Set(x, y, color.White)
		}
	}
	g.img.Image = g.imgRGBA
	g.img.Refresh()
	if g.onChange != nil {
		g.onChange()
	}
}

func (g *PixelGrid) Clear() {
	g.lock.Lock()
	defer g.lock.Unlock()
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			g.cells[r][c] = false
		}
	}
	draw.Draw(g.imgRGBA, g.imgRGBA.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	g.img.Image = g.imgRGBA
	g.img.Refresh()
	if g.onChange != nil {
		g.onChange()
	}
}

func (g *PixelGrid) Params() []float64 {
	g.lock.Lock()
	defer g.lock.Unlock()
	params := make([]float64, 10)
	for i := 0; i < 10; i++ {
		start := int(math.Floor(float64(i*Cols) / 10.0))
		end := int(math.Floor(float64((i+1)*Cols) / 10.0))
		if end <= start {
			end = start + 1
		}
		total := 0
		white := 0
		for c := start; c < end; c++ {
			for r := 0; r < Rows; r++ {
				total++
				if g.cells[r][c] {
					white++
				}
			}
		}
		if total == 0 {
			params[i] = 0
		} else {
			params[i] = float64(white) / float64(total)
		}
	}
	return params
}

func (g *PixelGrid) handlePoint(p fyne.Position) {
	x := int(p.X)
	y := int(p.Y)
	if x < 0 || x >= GridW || y < 0 || y >= GridH {
		return
	}
	col := x / CellSize
	row := y / CellSize
	g.setCell(row, col)
}

func (g *PixelGrid) Tapped(ev *fyne.PointEvent) {
	g.handlePoint(ev.Position)
}

func (g *PixelGrid) Dragged(ev *fyne.DragEvent) {
	g.handlePoint(ev.Position)
}

type gridRenderer struct {
	grid    *PixelGrid
	objects []fyne.CanvasObject
}

func (r *gridRenderer) Destroy()                     {}
func (r *gridRenderer) Layout(size fyne.Size)        { r.objects[0].Resize(fyne.NewSize(GridW, GridH)) }
func (r *gridRenderer) MinSize() fyne.Size           { return fyne.NewSize(GridW, GridH) }
func (r *gridRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *gridRenderer) Refresh()                     { canvas.Refresh(r.objects[0]) }

// ---------------- Radar -----------------

type Radar struct {
	widget.BaseWidget
	lock   sync.Mutex
	values []float64
	img    *canvas.Image
	w, h   int
}

func NewRadar(w, h int) *Radar {
	r := &Radar{
		values: make([]float64, 10),
		w:      w,
		h:      h,
	}
	r.ExtendBaseWidget(r)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{30, 30, 30, 255}}, image.Point{}, draw.Src)
	r.img = canvas.NewImageFromImage(img)
	r.img.FillMode = canvas.ImageFillContain
	return r
}

func (r *Radar) CreateRenderer() fyne.WidgetRenderer {
	return &radarRenderer{radar: r, objects: []fyne.CanvasObject{r.img}}
}

func (r *Radar) SetValues(vals []float64) {
	r.lock.Lock()
	copy(r.values, vals)
	r.lock.Unlock()
	r.redraw()
}

func (r *Radar) redraw() {
	img := image.NewRGBA(image.Rect(0, 0, r.w, r.h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{30, 30, 30, 255}}, image.Point{}, draw.Src)

	cx := float64(r.w) / 2
	cy := float64(r.h) / 2
	radius := math.Min(cx, cy) * 0.85
	n := len(r.values)

	// вспомогательная функция для линий
	drawLine := func(x0, y0, x1, y1 int, col color.Color) {
		dx := int(math.Abs(float64(x1 - x0)))
		dy := -int(math.Abs(float64(y1 - y0)))
		sx := 1
		if x0 >= x1 {
			sx = -1
		}
		sy := 1
		if y0 >= y1 {
			sy = -1
		}
		err := dx + dy
		for {
			if x0 >= 0 && x0 < r.w && y0 >= 0 && y0 < r.h {
				img.Set(x0, y0, col)
			}
			if x0 == x1 && y0 == y1 {
				break
			}
			e2 := 2 * err
			if e2 >= dy {
				err += dy
				x0 += sx
			}
			if e2 <= dx {
				err += dx
				y0 += sy
			}
		}
	}

	// оси и концентрические многоугольники
	for i := 0; i < n; i++ {
		angle := 2*math.Pi*float64(i)/float64(n) - math.Pi/2
		x := int(cx + radius*math.Cos(angle))
		y := int(cy + radius*math.Sin(angle))
		drawLine(int(cx), int(cy), x, y, color.RGBA{80, 80, 80, 255})
	}
	for k := 1; k <= 5; k++ {
		rad := radius * float64(k) / 5
		var px, py int
		for i := 0; i <= n; i++ {
			angle := 2*math.Pi*float64(i)/float64(n) - math.Pi/2
			x := int(cx + rad*math.Cos(angle))
			y := int(cy + rad*math.Sin(angle))
			if i > 0 {
				drawLine(px, py, x, y, color.RGBA{60, 60, 60, 255})
			}
			px, py = x, y
		}
	}

	r.lock.Lock()
	vals := append([]float64(nil), r.values...)
	r.lock.Unlock()

	// ограничим значения 0–1
	for i := range vals {
		if vals[i] < 0 {
			vals[i] = 0
		}
		if vals[i] > 1 {
			vals[i] = 1
		}
	}

	// построим полигоны значений
	var xs, ys []int
	for i := 0; i < n; i++ {
		angle := 2*math.Pi*float64(i)/float64(n) - math.Pi/2
		rad := radius * vals[i]
		x := int(cx + rad*math.Cos(angle))
		y := int(cy + rad*math.Sin(angle))
		xs = append(xs, x)
		ys = append(ys, y)
	}

	for i := 0; i < n; i++ {
		x0, y0 := xs[i], ys[i]
		x1, y1 := xs[(i+1)%n], ys[(i+1)%n]
		drawLine(x0, y0, x1, y1, color.White)
	}

	for i := 0; i < n; i++ {
		img.Set(xs[i], ys[i], color.White)
	}

	r.img.Image = img
	r.img.Refresh()
}

type radarRenderer struct {
	radar   *Radar
	objects []fyne.CanvasObject
}

func (rr *radarRenderer) Destroy() {}
func (rr *radarRenderer) Layout(size fyne.Size) {
	rr.objects[0].Resize(fyne.NewSize(float32(rr.radar.w), float32(rr.radar.h)))
}
func (rr *radarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(rr.radar.w), float32(rr.radar.h))
}
func (rr *radarRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *radarRenderer) Refresh()                     { rr.radar.redraw(); canvas.Refresh(rr.objects[0]) }

// ---------------- main ----------------

func main() {
	a := app.New()
	w := a.NewWindow("28x28 Grid + Radar")

	grid := NewPixelGrid()
	radar := NewRadar(320, 320)

	grid.onChange = func() {
		radar.SetValues(grid.Params())
	}

	clearBtn := widget.NewButton("Очистить", func() {
		grid.Clear()
	})

	saveBtn := widget.NewButton("Сохранить PNG", func() {
		f, err := os.Create("grid.png")
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		grid.lock.Lock()
		defer grid.lock.Unlock()
		png.Encode(f, grid.imgRGBA)
	})

	info := widget.NewLabel("Рисуйте, зажимая кнопку мыши. Радар обновляется автоматически.")
	content := container.NewBorder(info, container.NewHBox(clearBtn, saveBtn), nil, radar, grid.img)
	w.SetContent(content)
	w.Resize(fyne.NewSize(900, 350))

	// ✅ безопасный апдейтер радара через fyne.Do (универсальный способ)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			// vals := grid.Params()
			vals := []float64{1, 1, 1, 1, 1, 1, 1, 0.7, 0.5, 0}
			fyne.Do(func() {
				radar.SetValues(vals)
			})
		}
	}()

	w.ShowAndRun()
}
