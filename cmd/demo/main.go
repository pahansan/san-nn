package main

import (
	"fmt"
	"math/rand"
	"san-nn/internal/nn"
)

func main() {
	config := []int{2, 4, 1}
	neuro := nn.NewNN(config)
	neuro.InitWeightsRand()

	fmt.Println("Training...")

	for i := 0; i < 100000; i++ {
		x := rand.Int() % 2
		y := rand.Int() % 2
		target := float64(x ^ y)

		neuro.SetInput([]float64{float64(x), float64(y)})
		neuro.ForwardProp()
		neuro.BackProp([]float64{target}, 0.1)

		if i%10000 == 0 {
			cost, _ := neuro.GetCost([]float64{target})
			fmt.Printf("Iter %d, cost=%.6f\n", i, cost)
		}
	}

	tests := [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}}
	for _, t := range tests {
		neuro.SetInput([]float64{float64(t[0]), float64(t[1])})
		neuro.ForwardProp()
		fmt.Printf("%d ^ %d -> %.4f\n", t[0], t[1], neuro.GetOutput()[0])
	}
}
