package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	Bold        = "\033[1m"
)

type BenchmarkResult struct {
	Name        string
	Concurrency int
	ReqPerSec   float64
	MsPerReq    float64
}

func main() {
	results := parseBenchmarkOutput()
	if len(results) == 0 {
		fmt.Println("Aucun résultat de benchmark trouvé")
		return
	}

	printFormattedResults(results)
}

func parseBenchmarkOutput() []BenchmarkResult {
	// Lire depuis stdin
	input := ""
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}
		input += string(buf[:n])
	}

	results := []BenchmarkResult{}
	lines := strings.Split(input, "\n")

	// Patterns pour extraire les données
	benchPattern := regexp.MustCompile(`Benchmark(Bad|Good)Server_Concurrency(\d+)`)
	reqPerSecPattern := regexp.MustCompile(`(\d+\.?\d*)\s+req/s`)
	msPerReqPattern := regexp.MustCompile(`(\d+\.?\d*)\s+ms/req`)

	for _, line := range lines {
		if matches := benchPattern.FindStringSubmatch(line); matches != nil {
			serverType := matches[1]
			concurrency, _ := strconv.Atoi(matches[2])

			reqPerSec := 0.0
			if m := reqPerSecPattern.FindStringSubmatch(line); m != nil {
				reqPerSec, _ = strconv.ParseFloat(m[1], 64)
			}

			msPerReq := 0.0
			if m := msPerReqPattern.FindStringSubmatch(line); m != nil {
				msPerReq, _ = strconv.ParseFloat(m[1], 64)
			}

			results = append(results, BenchmarkResult{
				Name:        serverType,
				Concurrency: concurrency,
				ReqPerSec:   reqPerSec,
				MsPerReq:    msPerReq,
			})
		}
	}

	return results
}

func printFormattedResults(results []BenchmarkResult) {
	fmt.Printf("\n%s%s╔═══════════════════════════════════════════════════════════════════╗%s\n", Bold, ColorCyan, ColorReset)
	fmt.Printf("%s%s║                    📊 TABLEAU RÉCAPITULATIF                       ║%s\n", Bold, ColorCyan, ColorReset)
	fmt.Printf("%s%s╚═══════════════════════════════════════════════════════════════════╝%s\n\n", Bold, ColorCyan, ColorReset)

	// Grouper par concurrence
	concurrencyLevels := []int{1, 10, 50, 100}

	fmt.Printf("%s%-12s │ %s%-20s%s │ %s%-20s%s │ %s%-15s%s\n",
		Bold, "Concurrence",
		ColorYellow, "Bad Server", ColorReset,
		ColorGreen, "Good Server", ColorReset,
		ColorBlue, "Amélioration", ColorReset)
	fmt.Println("─────────────┼──────────────────────┼──────────────────────┼─────────────────")

	for _, conc := range concurrencyLevels {
		var badResult, goodResult BenchmarkResult
		
		for _, r := range results {
			if r.Concurrency == conc {
				if r.Name == "Bad" {
					badResult = r
				} else if r.Name == "Good" {
					goodResult = r
				}
			}
		}

		if badResult.ReqPerSec > 0 && goodResult.ReqPerSec > 0 {
			improvement := ((goodResult.ReqPerSec - badResult.ReqPerSec) / badResult.ReqPerSec) * 100
			improvementColor := ColorGreen
			if improvement < 0 {
				improvementColor = ColorRed
			}

			fmt.Printf("%-12d │ %s%6.0f req/s (%5.1fms)%s │ %s%6.0f req/s (%5.1fms)%s │ %s%+6.1f%%%s\n",
				conc,
				ColorYellow, badResult.ReqPerSec, badResult.MsPerReq, ColorReset,
				ColorGreen, goodResult.ReqPerSec, goodResult.MsPerReq, ColorReset,
				improvementColor, improvement, ColorReset)
		}
	}

	fmt.Printf("\n%s💡 Interprétation:%s\n", Bold, ColorReset)
	fmt.Println("• Le serveur GOOD est plus performant sous charge concurrente")
	fmt.Println("• L'amélioration est plus marquée avec une concurrence élevée")
	fmt.Println("• Le defer dans le mutex crée un goulot d'étranglement significatif")
}