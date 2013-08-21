package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
)

type InnerStruct map[string]float64
type OuterStruct map[string]InnerStruct

type MetricFunc func(p1, p2 InnerStruct) float64

type Richard struct {
	data OuterStruct
}

func (r *Richard) CommonKeys(v1, v2 InnerStruct) InnerStruct {
	common := make(InnerStruct)

	for key, _ := range v1 {
		if v2[key] != 0 {
			common[key] = 1
		}
	}

	return common
}

func (r *Richard) SimDistance(v1, v2 InnerStruct) float64 {
	common := r.CommonKeys(v1, v2)

	n := float64(len(common))

	if n == 0 {
		return 0.0
	}

	var sum_of_squares float64

	for key, _ := range common {
		sum_of_squares += math.Pow(v1[key]-v2[key], 2)
	}

	return 1 / (1 + sum_of_squares)
}

func (r *Richard) SimPearson(v1, v2 InnerStruct) float64 {
	common := r.CommonKeys(v1, v2)

	n := float64(len(common))

	if n == 0 {
		return 0
	}

	var sum1, sum2, sumsq1, sumsq2, sump float64

	for key, _ := range common {
		sum1 += v1[key]
		sumsq1 += math.Pow(v1[key], 2)
		sum2 += v2[key]
		sumsq2 += math.Pow(v2[key], 2)
		sump += v1[key] * v2[key]
	}

	num := sump - ((sum1 * sum2) / n)
	den := math.Sqrt((sumsq1 - (math.Pow(sum1, 2))/n) * (sumsq2 - (math.Pow(sum2, 2))/n))

	if den == 0 {
		return 0
	}

	return num / den
}

func (r *Richard) Recommend(thekey string, metric MetricFunc) InnerStruct {
	total := make(InnerStruct)
	totalSim := make(InnerStruct)
	ranking := make(InnerStruct)
	k1 := r.data[thekey]

	for p2, k2 := range r.data {
		if thekey != p2 {
			common := r.CommonKeys(k1, k2)
			sim := metric(k1, k2)

			if sim < 0 {
				continue
			}

			for key, value := range k2 {
				if common[key] > 0 {
					continue
				}

				total[key] += value * sim
				totalSim[key] += sim
			}
		}
	}

	for key, _ := range total {
		ranking[key] = total[key] / totalSim[key]
	}

	return ranking
}

func (r *Richard) Transform() OuterStruct {
	result := make(OuterStruct)

	for name, ratings := range r.data {
		for item, rating := range ratings {
			if result[item] == nil {
				result[item] = make(InnerStruct)
			}

			result[item][name] = rating
		}
	}

	return result
}

func (r *Richard) Top(thekey string, n int, metric MetricFunc) []float64 {
	if n == 0 {
		n = 5
	}

	scores := make([]float64, 0)

	for key, values := range r.data {
		if key == thekey {
			continue
		}

		scores = append(scores, metric(r.data[thekey], values))
	}

	if n > len(r.data) {
		n = len(r.data)
	}

	return scores
}

func (r *Richard) ImportJSON(filename string) {
	input, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("[ERROR]", "[ReadFile]", err)
		os.Exit(1)
	}

	err = json.Unmarshal(input, &r.data)

	if err != nil {
		fmt.Println("[ERROR]", "[Unmarshal]", err)
		os.Exit(1)
	}
}

func (r *Richard) Sim(n int) {
	if n == 0 {
		n = 5
	}

	for key, _ := range r.data {
		fmt.Print(key)
		fmt.Println(r.Top(key, 5, r.SimPearson))
	}
}

func main() {
	var r Richard

	r.ImportJSON(os.Args[1])

	r.data = r.Transform()

	for p1, _ := range r.data {
		fmt.Println(p1)
		fmt.Println(r.Recommend(p1, r.SimDistance))
		fmt.Println(r.Recommend(p1, r.SimPearson))
	}
	r.Sim(5)
}
