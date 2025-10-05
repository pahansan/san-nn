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
	tmp_input := n.Input

	for i := range n.Layers {
		if err := n.Layers[i].GetOutput(tmp_input); err != nil {
			panic(err)
		}
		tmp_input = n.Layers[i].Output
	}
	return nil
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

func (n *NN) BackProp(target []float64, theta float64) error {
	output := n.GetOutput()
	if len(target) != len(output) {
		return fmt.Errorf("got input with length %d, but expected %d", len(target), len(output))
	}

	for i := len(n.Layers) - 1; i >= 0; i-- {
		var dzdv float64
		var new_target_len int
		if i != 0 {
			new_target_len = len(n.Layers[i-1].Output)
		} else {
			new_target_len = len(n.Input)
		}
		var new_target = make([]float64, new_target_len)
		for j := range n.Layers[i].Weights {
			aj := n.Layers[i].Output[j]
			dadz := aj * (1 - aj)
			yj := target[j]
			dcda := aj - yj
			for k := range n.Layers[i].Weights[0] {
				if i != 0 {
					dzdv = n.Layers[i-1].Output[k]
				} else {
					dzdv = n.Input[k]
				}
				new_target[k] += n.Layers[i].Weights[j][k] * dadz * dcda
				n.Layers[i].Weights[j][k] -= theta * dzdv * dadz * dcda
			}
		}
		target = new_target
	}
	return nil
}

func (n *NN) GetCost(target []float64) (float64, error) {
	return nnmath.Cost(n.GetOutput(), target)
}
