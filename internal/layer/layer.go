package layer

import (
	"fmt"
	"math/rand"
	"san-nn/internal/nnmath"
)

type Layer struct {
	Weights [][]float64
	Output  []float64
}

func (l *Layer) GetOutput(input []float64) error {
	if len(l.Weights) == 0 {
		return fmt.Errorf("layer has no weights")
	}
	if len(l.Weights[0]) == 0 {
		return fmt.Errorf("layer has zero input size")
	}
	if len(input) != len(l.Weights[0]) {
		return fmt.Errorf("got input size %d but expected %d", len(input), len(l.Weights[0]))
	}

	if len(l.Output) != len(l.Weights) {
		l.Output = make([]float64, len(l.Weights))
	}

	for i, row := range l.Weights {
		sum := 0.0
		for j, weight := range row {
			sum += weight * input[j]
		}
		l.Output[i] = nnmath.Sigmoid(sum)
	}
	return nil
}

func NewLayer(inputSize, outputSize int) Layer {
	l := Layer{
		Weights: make([][]float64, outputSize),
		Output:  make([]float64, outputSize),
	}
	for i := range l.Weights {
		l.Weights[i] = make([]float64, inputSize)
	}
	return l
}

func (l *Layer) InitWeightsRand() {
	for i := range l.Weights {
		for j := range l.Weights[i] {
			l.Weights[i][j] = rand.Float64()
		}
	}
}
