package main

import (
	"fmt"
	"math/rand"
	"san-nn/internal/nn"
)

func main() {
	config := []int{2, 2, 1}
	neuro := nn.NewNN(config)
	neuro.InitWeightsRand()
	x := rand.Int() % 2
	y := rand.Int() % 2
	neuro.SetInput([]float64{float64(x), float64(y)})
	neuro.ForwardProp()
	start, _ := neuro.GetCost([]float64{float64(x ^ y)})
	for range 100000 {
		x := rand.Int() % 2
		y := rand.Int() % 2
		neuro.BackProp([]float64{float64(x ^ y)}, 1.0)
		neuro.ForwardProp()
		fmt.Printf("%d ^ %d = %f\n", x, y, neuro.GetOutput()[0])
	}
	end, _ := neuro.GetCost([]float64{float64(x ^ y)})

	fmt.Printf("Cost at start: %f, at end: %f\n", start, end)
}
