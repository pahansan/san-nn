package main

import (
	"fmt"
	"math/rand"
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
	return float64(correctCount) / 10000 * 100

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
	mnist := nn.NewNN([]int{784, 32, 16, 10})
	mnist.InitWeightsRand()
	accuracy := 0.0
	targetAccuracy := 95.0
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

	fmt.Println("Validation...")
	fmt.Println("Parsing...")
	fmt.Println("Accuracy: ", countAccuracy(test, mnist), "%")
}
