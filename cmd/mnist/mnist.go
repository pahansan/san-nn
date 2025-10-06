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

func main() {
	fmt.Println("Parsing...")
	strs, _ := parser.ReadCSV("mnist_train.csv")
	nums := parser.ParseLines(strs)
	fmt.Println("Train...")
	mnist := nn.NewNN([]int{784, 16, 16, 10})
	mnist.InitWeightsRand()
	for i, ex := range nums {
		input := ex[1:]
		for j := range input {
			input[j] = input[j] / 255
		}
		mnist.SetInput(input)
		mnist.ForwardProp()
		mnist.BackProp(formatTarget(int(ex[0])), 0.3)
		if i%10000 == 0 {
			cost, _ := mnist.GetCost(formatTarget(int(ex[0])))
			fmt.Println("Iteration:", i, "Cost:", cost)
		}
	}

	fmt.Println("Validation...")
	fmt.Println("Parsing...")
	strs, _ = parser.ReadCSV("mnist_test.csv")
	nums = parser.ParseLines(strs)
	correctCount := 0
	for _, ex := range nums {
		input := ex[1:]
		for j := range input {
			input[j] = input[j] / 255
		}
		mnist.SetInput(input)
		mnist.ForwardProp()
		ans := maxIndex(mnist.GetOutput())
		if ans == int(ex[0]) {
			correctCount++
		}
	}
	fmt.Println("Accuracy: ", float64(correctCount)/10000*100, "%")
}
