package main_test

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

// Couleurs ANSI
const (
	// Couleurs
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	
	// Styles
	Bold      = "\033[1m"
	Underline = "\033[4m"
)

const (
	badServerURL     = "http://localhost:8081/process"
	goodServerURL    = "http://localhost:8082/process"
	syncmapServerURL = "http://localhost:8083/process"
)

/*
benchmarkServer effectue des tests de charge sur un serveur HTTP.
Mesure le throughput et la latence sous différents niveaux de concurrence.

@params:
  - b: *testing.B instance du benchmark
  - url: string URL du serveur à tester
  - concurrency: int nombre de goroutines concurrentes

@metrics:
  - req/s: Requêtes par seconde (throughput)
  - ms/req: Millisecondes par requête (latence moyenne)
*/
func benchmarkServer(b *testing.B, url string, concurrency int) {
	b.ResetTimer()
	
	var wg sync.WaitGroup
	requests := b.N
	requestsPerGoroutine := requests / concurrency
	
	start := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			
			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(url)
				if err != nil {
					b.Errorf("Request failed: %v", err)
					continue
				}
				io.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}()
	}
	
	wg.Wait()
	
	duration := time.Since(start)
	b.ReportMetric(float64(requests)/duration.Seconds(), "req/s")
	b.ReportMetric(float64(duration.Milliseconds())/float64(requests), "ms/req")
}

/*
BenchmarkBadServer_Concurrency1 teste le serveur "bad" avec 1 seule goroutine.
Ce test sert de baseline sans contention sur le mutex.

@expected: Performances similaires au serveur "good" car pas de concurrence
*/
func BenchmarkBadServer_Concurrency1(b *testing.B) {
	benchmarkServer(b, badServerURL, 1)
}

/*
BenchmarkBadServer_Concurrency10 teste avec 10 goroutines concurrentes.
@expected: Début de la dégradation des performances due au mutex
*/
func BenchmarkBadServer_Concurrency10(b *testing.B) {
	benchmarkServer(b, badServerURL, 10)
}

/*
BenchmarkBadServer_Concurrency50 teste avec 50 goroutines concurrentes.
@expected: Dégradation significative, goulot d'étranglement évident
*/
func BenchmarkBadServer_Concurrency50(b *testing.B) {
	benchmarkServer(b, badServerURL, 50)
}

/*
BenchmarkBadServer_Concurrency100 teste avec 100 goroutines concurrentes.
@expected: Performances catastrophiques, système quasi-séquentiel
*/
func BenchmarkBadServer_Concurrency100(b *testing.B) {
	benchmarkServer(b, badServerURL, 100)
}

/*
BenchmarkGoodServer_Concurrency1 teste le serveur "good" avec 1 seule goroutine.
@expected: Baseline identique au serveur "bad"
*/
func BenchmarkGoodServer_Concurrency1(b *testing.B) {
	benchmarkServer(b, goodServerURL, 1)
}

/*
BenchmarkGoodServer_Concurrency10 teste avec 10 goroutines concurrentes.
@expected: Amélioration massive du throughput grâce au parallélisme
*/
func BenchmarkGoodServer_Concurrency10(b *testing.B) {
	benchmarkServer(b, goodServerURL, 10)
}

/*
BenchmarkGoodServer_Concurrency50 teste avec 50 goroutines concurrentes.
@expected: Performances maintenues, bonne scalabilité
*/
func BenchmarkGoodServer_Concurrency50(b *testing.B) {
	benchmarkServer(b, goodServerURL, 50)
}

/*
BenchmarkGoodServer_Concurrency100 teste avec 100 goroutines concurrentes.
@expected: Légère dégradation mais toujours bien meilleur que "bad"
*/
func BenchmarkGoodServer_Concurrency100(b *testing.B) {
	benchmarkServer(b, goodServerURL, 100)
}

/*
BenchmarkSyncMapServer_Concurrency1 teste le serveur "syncmap" avec 1 seule goroutine.
@expected: Baseline comparable aux autres serveurs
*/
func BenchmarkSyncMapServer_Concurrency1(b *testing.B) {
	benchmarkServer(b, syncmapServerURL, 1)
}

/*
BenchmarkSyncMapServer_Concurrency10 teste avec 10 goroutines concurrentes.
@expected: Bonnes performances sans gestion manuelle de mutex
*/
func BenchmarkSyncMapServer_Concurrency10(b *testing.B) {
	benchmarkServer(b, syncmapServerURL, 10)
}

/*
BenchmarkSyncMapServer_Concurrency50 teste avec 50 goroutines concurrentes.
@expected: sync.Map optimisée pour les lectures concurrentes
*/
func BenchmarkSyncMapServer_Concurrency50(b *testing.B) {
	benchmarkServer(b, syncmapServerURL, 50)
}

/*
BenchmarkSyncMapServer_Concurrency100 teste avec 100 goroutines concurrentes.
@expected: Performance stable avec haute concurrence
*/
func BenchmarkSyncMapServer_Concurrency100(b *testing.B) {
	benchmarkServer(b, syncmapServerURL, 100)
}

/*
TestLatencyComparison effectue une comparaison détaillée des latences.
Génère un tableau comparatif montrant l'amélioration de performance.

@params:
  - t: *testing.T instance du test

@output: Tableau formaté avec latences et pourcentages d'amélioration
*/
func TestLatencyComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}
	
	concurrencyLevels := []int{1, 10, 50, 100}
	
	fmt.Printf("\n%s%s=== 🚀 COMPARAISON DE LATENCE DES 3 SERVEURS ===%s\n", Bold, ColorCyan, ColorReset)
	fmt.Printf("%s%-12s | %-15s | %-16s | %-17s | %s%s\n", 
		Bold, "Concurrency", "Bad (defer) ms", "Good (no defer) ms", "SyncMap (no mutex) ms", "Best Improvement", ColorReset)
	fmt.Println("━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━")
	
	for _, concurrency := range concurrencyLevels {
		badLatency := measureAverageLatency(badServerURL, concurrency, 100)
		goodLatency := measureAverageLatency(goodServerURL, concurrency, 100)
		syncmapLatency := measureAverageLatency(syncmapServerURL, concurrency, 100)
		
		// Calculer les améliorations
		goodImprovement := ((badLatency - goodLatency) / badLatency) * 100
		syncmapImprovement := ((badLatency - syncmapLatency) / badLatency) * 100
		
		// Colorer les latences selon les valeurs
		badColor := ColorRed
		if badLatency < 100 {
			badColor = ColorYellow
		}
		goodColor := ColorGreen
		if goodLatency > 100 {
			goodColor = ColorYellow
		}
		syncmapColor := ColorPurple
		if syncmapLatency > 100 {
			syncmapColor = ColorYellow
		}
		
		// Déterminer la meilleure amélioration
		bestImprovement := goodImprovement
		bestServer := "GOOD"
		if syncmapImprovement > goodImprovement {
			bestImprovement = syncmapImprovement
			bestServer = "SYNC.MAP"
		}
		
		// Colorer l'amélioration
		improvementStr := ""
		if bestImprovement > 0 {
			improvementStr = fmt.Sprintf("%s%s+%.1f%% (%s)%s", ColorGreen, Bold, bestImprovement, bestServer, ColorReset)
		} else {
			improvementStr = fmt.Sprintf("%s%.1f%%%s", ColorRed, bestImprovement, ColorReset)
		}
		
		fmt.Printf("%s%-12d%s ┃ %s%-15.2f%s ┃ %s%-17.2f%s ┃ %s%-20.2f%s ┃ %s\n", 
			ColorWhite, concurrency, ColorReset,
			badColor, badLatency, ColorReset,
			goodColor, goodLatency, ColorReset,
			syncmapColor, syncmapLatency, ColorReset,
			improvementStr)
	}
	
	fmt.Printf("\n%s%sLégende:%s\n", Bold, ColorBlue, ColorReset)
	fmt.Printf("• %sBad Server%s: Mutex avec defer (bloque pendant tout le traitement)\n", ColorRed, ColorReset)
	fmt.Printf("• %sGood Server%s: Mutex sans defer (libération immédiate)\n", ColorGreen, ColorReset)
	fmt.Printf("• %sSyncMap Server%s: sync.Map (pas de mutex manuel)\n", ColorPurple, ColorReset)
}

/*
measureAverageLatency calcule la latence moyenne pour un serveur donné.

@params:
  - url: string URL du serveur à mesurer
  - concurrency: int nombre de clients concurrents
  - totalRequests: int nombre total de requêtes à effectuer

@returns: float64 latence moyenne en millisecondes

@behavior:
  - Distribue les requêtes équitablement entre les goroutines
  - Mesure le temps de chaque requête individuellement
  - Calcule la moyenne sur toutes les requêtes réussies
*/
func measureAverageLatency(url string, concurrency int, totalRequests int) float64 {
	var wg sync.WaitGroup
	latencies := make(chan time.Duration, totalRequests)
	requestsPerGoroutine := totalRequests / concurrency
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			
			for j := 0; j < requestsPerGoroutine; j++ {
				start := time.Now()
				resp, err := client.Get(url)
				if err == nil {
					io.ReadAll(resp.Body)
					resp.Body.Close()
					latencies <- time.Since(start)
				}
			}
		}()
	}
	
	wg.Wait()
	close(latencies)
	
	var totalLatency time.Duration
	count := 0
	for latency := range latencies {
		totalLatency += latency
		count++
	}
	
	if count == 0 {
		return 0
	}
	
	return float64(totalLatency.Milliseconds()) / float64(count)
}