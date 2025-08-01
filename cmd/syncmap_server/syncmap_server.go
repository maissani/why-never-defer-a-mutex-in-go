package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
)

/*
DataStruct représente une structure de données complexe.

@fields:
  - Identifier: Identifiant unique
  - Name: Nom de l'élément
  - IsActive: État actif/inactif
  - Counter: Compteur d'accès
  - LastModified: Timestamp de dernière modification
*/
type DataStruct struct {
	Identifier   string    `json:"identifier"`
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	Counter      int       `json:"counter"`
	LastModified time.Time `json:"last_modified"`
}

/*
Repository utilise sync.Map pour une gestion thread-safe sans mutex explicite.
sync.Map est optimisée pour deux cas d'usage:
1) Peu d'écritures mais beaucoup de lectures
2) Plusieurs goroutines lisent/écrivent des clés disjointes

@fields:
  - counter: Compteur atomique pour éviter les mutex
  - data: sync.Map pour stocker les données de manière thread-safe
*/
type Repository struct {
	counter int64    // Utilise atomic pour éviter le mutex
	data    sync.Map // Thread-safe map sans mutex manuel
}

/*
NewRepository crée et initialise un nouveau repository avec sync.Map.

@returns: *Repository - Nouvelle instance utilisant sync.Map
*/
func NewRepository() *Repository {
	return &Repository{}
}

/*
SyncMapHandler démontre l'utilisation de sync.Map pour la concurrence.
sync.Map gère automatiquement la synchronisation sans mutex explicite.

@params:
  - w: http.ResponseWriter pour envoyer la réponse
  - req: *http.Request contenant la requête HTTP

@behavior:
  1. Incrémente le compteur atomiquement
  2. Lit les données via sync.Map.Range (thread-safe)
  3. Effectue le traitement lourd sans bloquer d'autres opérations
  4. Écrit les résultats dans sync.Map (thread-safe)

@performance: sync.Map optimise automatiquement l'accès concurrent
*/
func (r *Repository) SyncMapHandler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Incrémentation atomique du compteur
	currentCounter := atomic.AddInt64(&r.counter, 1)

	// Lecture des données avec sync.Map.Range (thread-safe)
	dataCopy := make(map[string]*DataStruct)
	r.data.Range(func(key, value interface{}) bool {
		if ds, ok := value.(*DataStruct); ok {
			dataCopy[key.(string)] = &DataStruct{
				Identifier:   ds.Identifier,
				Name:         ds.Name,
				IsActive:     ds.IsActive,
				Counter:      ds.Counter,
				LastModified: ds.LastModified,
			}
		}
		return true // Continue l'itération
	})

	// Traitement lourd (pas de mutex à gérer)
	time.Sleep(10 * time.Millisecond) // Simule un traitement
	
	// Calcul intensif simulé
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i
	}

	// Écriture dans sync.Map (thread-safe automatiquement)
	key := fmt.Sprintf("request_%d", currentCounter)
	r.data.Store(key, &DataStruct{
		Identifier:   key,
		Name:         fmt.Sprintf("Request %d", currentCounter),
		IsActive:     true,
		Counter:      result,
		LastModified: time.Now(),
	})

	response := map[string]interface{}{
		"method":   "sync_map",
		"counter":  currentCounter,
		"result":   result,
		"duration": time.Since(start).Microseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*
StatsHandler retourne les statistiques actuelles du serveur.
Utilise sync.Map et atomic pour un accès thread-safe.

@params:
  - w: http.ResponseWriter pour envoyer la réponse
  - req: *http.Request contenant la requête HTTP

@returns: JSON contenant total_requests et data_size
*/
func (r *Repository) StatsHandler(w http.ResponseWriter, req *http.Request) {
	// Lecture atomique du compteur
	counter := atomic.LoadInt64(&r.counter)
	
	// Comptage des éléments dans sync.Map
	dataSize := 0
	r.data.Range(func(key, value interface{}) bool {
		dataSize++
		return true
	})

	stats := map[string]interface{}{
		"total_requests": counter,
		"data_size":      dataSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

/*
main initialise et démarre le serveur HTTP utilisant sync.Map.

@behavior:
  - Crée un repository avec sync.Map
  - Configure les routes avec gorilla/mux
  - Démarre le serveur sur le port 8083
  - Affiche les endpoints disponibles

@endpoints:
  - GET /process : Handler avec sync.Map (pas de mutex manuel)
  - GET /stats : Statistiques du serveur
*/
func main() {
	repo := NewRepository()
	
	r := mux.NewRouter()
	r.HandleFunc("/process", repo.SyncMapHandler).Methods("GET")
	r.HandleFunc("/stats", repo.StatsHandler).Methods("GET")

	fmt.Println("SYNC.MAP Server (sans mutex manuel) starting on :8083")
	fmt.Println("Endpoints:")
	fmt.Println("  GET /process - Utilisation avec sync.Map")
	fmt.Println("  GET /stats   - Voir les statistiques")
	
	if err := http.ListenAndServe(":8083", r); err != nil {
		panic(err)
	}
}