package main

import (
	"fmt"
	"san-nn/internal/nn"
	"san-nn/internal/parser"
)

func formatTarget(t int) []float64 {
	tmp := make([]float64, 10)
	tmp[t] = 1
	return tmp
}

func main() {
	fmt.Println("Parsing...")
	strs, _ := parser.ReadCSV("mnist_train.csv")
	nums := parser.ParseLines(strs)
	fmt.Println("Train...")
	mnist := nn.NewNN([]int{784, 1000, 10})
	for i, ex := range nums {
		mnist.SetInput(ex[1:])
		mnist.ForwardProp()
		mnist.BackProp(formatTarget(int(ex[0])), 0.1)
		if i%10000 == 0 {
			fmt.Println(mnist.GetCost(formatTarget(int(ex[0]))))
			fmt.Println(formatTarget(int(ex[0])))
			fmt.Println(mnist.GetOutput())
		}
	}
}
