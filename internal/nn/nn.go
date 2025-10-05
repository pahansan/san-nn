package nn

import (
	"fmt"
	"san-nn/internal/layer"
)

type NN struct {
	Layers []layer.Layer
	Input  []float64
}

func NewNN(layersConfig []int) NN {
	tmp := NN{
		Layers: make([]layer.Layer, len(layersConfig)-1),
		Input:  make([]float64, layersConfig[0]),
	}

	for i := range tmp.Layers {
		tmp.Layers[i] = layer.NewLayer(layersConfig[i], layersConfig[i+1])
	}

	return tmp
}

func (n *NN) InitWeightsRand() {
	for i := range n.Layers {
		n.Layers[i].InitWeightsRand()
	}
}

func (n *NN) ForwardProp() {
	tmp_input := n.Input

	for i := range n.Layers {
		n.Layers[i].GetOutput(tmp_input)
		tmp_input = n.Layers[i].Output
	}
}

func (n *NN) SetInput(input []float64) error {
	if len(input) != len(n.Input) {
		return fmt.Errorf("got input with length %d, but expected %d", len(input), len(n.Input))
	}
	n.Input = input
	return nil
}

func (n *NN) GetOutput() []float64 {
	return n.Layers[len(n.Layers)-1].Output
}
