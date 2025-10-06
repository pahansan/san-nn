package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
	"san-nn/internal/nn"
	"san-nn/internal/parser"
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

	cells         [Rows][Cols]float64
	img           *canvas.Image
	imgRGBA       *image.RGBA
	brushSize     int
	brushStrength float64
	lastRow       int
	lastCol       int
}

var _ fyne.Draggable = (*PixelGrid)(nil)

func NewPixelGrid() *PixelGrid {
	g := &PixelGrid{
		brushSize:     1,
		brushStrength: 0.3,
		lastRow:       -1,
		lastCol:       -1,
	}
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

func (g *PixelGrid) applyBrush(r, c int) {
	if r < 0 || r >= Rows || c < 0 || c >= Cols {
		return
	}

	radius := g.brushSize
	strength := g.brushStrength

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nr := r + dy
			nc := c + dx
			if nr < 0 || nr >= Rows || nc < 0 || nc >= Cols {
				continue
			}
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			intens := strength * math.Exp(-dist*dist/2)
			g.cells[nr][nc] += intens
			if g.cells[nr][nc] > 1 {
				g.cells[nr][nc] = 1
			}
		}
	}
	g.redraw()
}

func (g *PixelGrid) drawBrushLine(r0, c0, r1, c1 int) {
	if r0 == r1 && c0 == c1 {
		return // no need to apply if same cell
	}

	dr := int(math.Abs(float64(r1 - r0)))
	dc := int(math.Abs(float64(c1 - c0)))
	sr := 1
	if r0 >= r1 {
		sr = -1
	}
	sc := 1
	if c0 >= c1 {
		sc = -1
	}
	err := dr - dc
	r, c := r0, c0

	for {
		g.applyBrush(r, c)
		if r == r1 && c == c1 {
			break
		}
		e2 := 2 * err
		if e2 > -dc {
			err -= dc
			r += sr
		}
		if e2 < dr {
			err += dr
			c += sc
		}
	}
}

func (g *PixelGrid) redraw() {
	draw.Draw(g.imgRGBA, g.imgRGBA.Bounds(), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			intens := uint8(g.cells[r][c] * 255)
			cellColor := color.RGBA{intens, intens, intens, 255}
			x0 := c * CellSize
			y0 := r * CellSize
			for y := y0; y < y0+CellSize; y++ {
				for x := x0; x < x0+CellSize; x++ {
					g.imgRGBA.Set(x, y, cellColor)
				}
			}
		}
	}
	lineColor := color.RGBA{40, 40, 40, 255}
	for c := 0; c <= Cols; c++ {
		x := c * CellSize
		for y := 0; y < GridH; y++ {
			g.imgRGBA.Set(x, y, lineColor)
		}
	}
	for r := 0; r <= Rows; r++ {
		y := r * CellSize
		for x := 0; x < GridW; x++ {
			g.imgRGBA.Set(x, y, lineColor)
		}
	}
	g.img.Refresh()
}

func (g *PixelGrid) Clear() {
	g.lock.Lock()
	defer g.lock.Unlock()
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			g.cells[r][c] = 0
		}
	}
	g.lastRow = -1
	g.lastCol = -1
	g.redraw()
}

func (g *PixelGrid) Params() []float64 {
	g.lock.Lock()
	defer g.lock.Unlock()
	params := []float64{}
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			params = append(params, g.cells[i][j])
		}
	}
	return params
}

func (g *PixelGrid) handlePoint(p fyne.Position, isTap bool) {
	x := int(p.X)
	y := int(p.Y)
	if x < 0 || x >= GridW || y < 0 || y >= GridH {
		return
	}
	col := x / CellSize
	row := y / CellSize

	g.lock.Lock()
	if isTap || g.lastRow < 0 {
		g.applyBrush(row, col)
	} else if row != g.lastRow || col != g.lastCol {
		g.drawBrushLine(g.lastRow, g.lastCol, row, col)
	}
	g.lastRow = row
	g.lastCol = col
	g.lock.Unlock()
}

func (g *PixelGrid) Tapped(ev *fyne.PointEvent) {
	g.handlePoint(ev.Position, true)
}

func (g *PixelGrid) Dragged(ev *fyne.DragEvent) {
	g.handlePoint(ev.Position, false)
}

func (g *PixelGrid) DragEnd() {
	g.lock.Lock()
	g.lastRow = -1
	g.lastCol = -1
	g.lock.Unlock()
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

type Radar struct {
	widget.BaseWidget
	lock   sync.Mutex
	values []float64
	img    *canvas.Image
	labels []*canvas.Text
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

	for i := 0; i < 10; i++ {
		label := canvas.NewText(fmt.Sprintf("%d", i), color.White)
		label.TextSize = 14
		label.Alignment = fyne.TextAlignCenter
		r.labels = append(r.labels, label)
	}

	return r
}

func (r *Radar) CreateRenderer() fyne.WidgetRenderer {
	objects := []fyne.CanvasObject{r.img}
	for _, label := range r.labels {
		objects = append(objects, label)
	}
	return &radarRenderer{radar: r, objects: objects}
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

	for i := range vals {
		if vals[i] < 0 {
			vals[i] = 0
		}
		if vals[i] > 1 {
			vals[i] = 1
		}
	}

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
		drawLine(xs[i], ys[i], xs[(i+1)%n], ys[(i+1)%n], color.White)
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
	rr.objects[0].Resize(size)
	cx := size.Width / 2
	cy := size.Height / 2
	radius := math.Min(float64(size.Width)/2, float64(size.Height)/2)*0.85 + 15

	for i, label := range rr.radar.labels {
		angle := 2*math.Pi*float64(i)/10 - math.Pi/2
		x := cx + float32(radius*math.Cos(angle))
		y := cy + float32(radius*math.Sin(angle))
		labelMin := label.MinSize()
		label.Resize(labelMin)
		label.Move(fyne.NewPos(x-labelMin.Width/2, y-labelMin.Height/2))
	}
}

func (rr *radarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(rr.radar.w), float32(rr.radar.h))
}
func (rr *radarRenderer) Objects() []fyne.CanvasObject { return rr.objects }
func (rr *radarRenderer) Refresh()                     { rr.radar.redraw(); canvas.Refresh(rr.objects[0]) }

func formatTarget(t int) []float64 {
	tmp := make([]float64, 10)
	tmp[t] = 1
	return tmp
}

func maxIndex(arr []float64) int {
	max := 0.0
	var idx int
	for i, num := range arr {
		if num > max {
			max = num
			idx = i
		}
	}
	return idx
}

func shuffle(slice [][]float64) {
	for i := range slice {
		j := rand.Intn(len(slice))
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func prepareData(data [][]float64) {
	for _, ex := range data {
		input := ex[1:]
		for j := range input {
			input[j] = input[j] / 255
		}
	}
}

func countAccuracy(data [][]float64, model nn.NN) float64 {
	correctCount := 0
	for _, ex := range data {
		input := ex[1:]
		model.SetInput(input)
		model.ForwardProp()
		ans := maxIndex(model.GetOutput())
		if ans == int(ex[0]) {
			correctCount++
		}
	}
	return float64(correctCount) / float64(len(data)) * 100
}

func main() {
	fmt.Println("Parsing...")
	strs, _ := parser.ReadCSV("mnist_train.csv")
	train := parser.ParseLines(strs)
	prepareData(train)
	strs, _ = parser.ReadCSV("mnist_test.csv")
	test := parser.ParseLines(strs)
	prepareData(test)
	fmt.Println("Train...")
	mnist := nn.NewNN([]int{784, 16, 16, 10})
	mnist.InitWeightsRand()
	accuracy := 0.0
	targetAccuracy := 94.0
	for j := 0; accuracy <= targetAccuracy; j++ {
		shuffle(train)
		for i, ex := range train {
			input := ex[1:]
			mnist.SetInput(input)
			mnist.ForwardProp()
			mnist.BackProp(formatTarget(int(ex[0])), 0.1)
			if i%10000 == 0 {
				cost, _ := mnist.GetCost(formatTarget(int(ex[0])))
				accuracy = countAccuracy(test, mnist)
				fmt.Println("Iteration:", i+j*60000, "Cost:", cost, "Accuracy:", accuracy, "%")
				if accuracy >= targetAccuracy {
					break
				}
			}
		}
	}

	a := app.NewWithID("nn.demo")
	w := a.NewWindow("san-nn")

	grid := NewPixelGrid()
	grid.Clear()
	radar := NewRadar(280, 280)

	clearBtn := widget.NewButton("Очистить", func() { grid.Clear() })

	mainContainer := container.NewGridWithColumns(2, grid, radar)

	mainContainer = container.NewBorder(
		mainContainer,
		clearBtn,
		nil,
		nil,
	)

	content := container.NewPadded(mainContainer)
	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 320))

	go func() {
		for {
			time.Sleep(15 * time.Millisecond)
			vals := grid.Params()
			err := mnist.SetInput(vals)
			if err != nil {
				fmt.Println(err)
			}
			mnist.ForwardProp()
			fyne.Do(func() { radar.SetValues(mnist.GetOutput()) })
		}
	}()

	w.ShowAndRun()
}
