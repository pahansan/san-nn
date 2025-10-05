package nn

import (
	"fmt"
	"san-nn/internal/layer"
	"san-nn/internal/nnmath"
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

func (n *NN) ForwardProp() error {
	tmpInput := n.Input
	for i := range n.Layers {
		if err := (&n.Layers[i]).GetOutput(tmpInput); err != nil {
			return err
		}
		tmpInput = n.Layers[i].Output
	}
	return nil
}

func (n *NN) SetInput(input []float64) error {
	if len(input) != len(n.Input) {
		return fmt.Errorf("got input with length %d, but expected %d", len(input), len(n.Input))
	}
	n.Input = append([]float64(nil), input...)
	return nil
}

func (n *NN) GetOutput() []float64 {
	if len(n.Layers) == 0 {
		return nil
	}
	return n.Layers[len(n.Layers)-1].Output
}

func (n *NN) BackProp(target []float64, theta float64) error {
	if len(n.Layers) == 0 {
		return fmt.Errorf("network has no layers")
	}
	output := n.GetOutput()
	if len(target) != len(output) {
		return fmt.Errorf("got target with length %d, but expected %d", len(target), len(output))
	}

	delta := make([]float64, len(output))
	for j := range output {
		delta[j] = output[j] - target[j]
	}

	for i := len(n.Layers) - 1; i >= 0; i-- {
		layer := &n.Layers[i]

		prevAct := n.Input
		if i > 0 {
			prevAct = n.Layers[i-1].Output
		}

		prevDelta := make([]float64, len(prevAct))

		for j := 0; j < len(layer.Weights); j++ {
			a := layer.Output[j]
			da_dz := a * (1 - a)
			deltaZ := delta[j] * da_dz
			for k := 0; k < len(layer.Weights[j]); k++ {
				oldW := layer.Weights[j][k]
				grad := deltaZ * prevAct[k]
				layer.Weights[j][k] -= theta * grad
				prevDelta[k] += oldW * deltaZ
			}
		}
		delta = prevDelta
	}
	return nil
}

func (n *NN) GetCost(target []float64) (float64, error) {
	return nnmath.Cost(n.GetOutput(), target)
}
