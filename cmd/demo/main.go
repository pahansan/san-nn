package main

import (
	"fmt"
	"san-nn/internal/nn"
)

func main() {
	config := []int{2, 2, 1}
	neuro := nn.NewNN(config)
	neuro.InitWeightsRand()
	neuro.SetInput([]float64{.4, .2})
	neuro.ForwardProp()
	fmt.Printf("Output: %f\n", neuro.GetOutput()[0])
}
