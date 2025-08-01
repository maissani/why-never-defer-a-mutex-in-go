package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

/*
Repository contient les données partagées protégées par un mutex.
Cette structure simule un état partagé typique dans une application Go.

@fields:
  - mu: Mutex pour protéger l'accès concurrent aux données
  - counter: Compteur global des requêtes traitées
  - data: Map simulant des données métier partagées
*/
type Repository struct {
	mu      sync.Mutex
	counter int
	data    map[string]int
}

/*
NewRepository crée et initialise un nouveau repository.

@returns: *Repository - Nouvelle instance avec la map initialisée
*/
func NewRepository() *Repository {
	return &Repository{
		data: make(map[string]int),
	}
}

/*
BadHandler démontre la MAUVAISE PRATIQUE d'utilisation des mutex avec defer.
Le mutex reste verrouillé pendant toute la durée du traitement, incluant
les opérations coûteuses qui n'ont pas besoin de protection.

@params:
  - w: http.ResponseWriter pour envoyer la réponse
  - req: *http.Request contenant la requête HTTP

@behavior:
  1. Verrouille le mutex avec defer (reste verrouillé jusqu'à la fin)
  2. Effectue des opérations sur les données partagées
  3. Effectue un traitement lourd AVEC le mutex verrouillé
  4. Le mutex n'est libéré qu'à la fin de la fonction

@performance: Cette approche crée un goulot d'étranglement majeur
*/
func (r *Repository) BadHandler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Mauvaise pratique: le mutex reste verrouillé pendant TOUT le traitement
	r.mu.Lock()
	defer r.mu.Unlock()

	// Lecture et copie des données
	r.counter++
	currentCounter := r.counter
	dataCopy := make(map[string]int)
	for k, v := range r.data {
		dataCopy[k] = v
	}

	// Simulation d'un traitement lourd (calcul, appel API, etc.)
	// Le mutex reste verrouillé pendant ce temps !
	time.Sleep(10 * time.Millisecond) // Simule un traitement
	
	// Calcul intensif simulé
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i
	}

	// Écriture des résultats
	r.data[fmt.Sprintf("request_%d", currentCounter)] = result

	response := map[string]interface{}{
		"method":   "bad_defer",
		"counter":  currentCounter,
		"result":   result,
		"duration": time.Since(start).Microseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*
StatsHandler retourne les statistiques actuelles du serveur.
Utilise correctement le mutex uniquement pour la lecture des données.

@params:
  - w: http.ResponseWriter pour envoyer la réponse
  - req: *http.Request contenant la requête HTTP

@returns: JSON contenant total_requests et data_size
*/
func (r *Repository) StatsHandler(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	stats := map[string]interface{}{
		"total_requests": r.counter,
		"data_size":      len(r.data),
	}
	r.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

/*
main initialise et démarre le serveur HTTP démontrant la mauvaise pratique.

@behavior:
  - Crée un repository partagé
  - Configure les routes avec gorilla/mux
  - Démarre le serveur sur le port 8081
  - Affiche les endpoints disponibles

@endpoints:
  - GET /process : Handler avec mauvaise utilisation du mutex
  - GET /stats : Statistiques du serveur
*/
func main() {
	repo := NewRepository()
	
	r := mux.NewRouter()
	r.HandleFunc("/process", repo.BadHandler).Methods("GET")
	r.HandleFunc("/stats", repo.StatsHandler).Methods("GET")

	fmt.Println("BAD Server (avec defer) starting on :8081")
	fmt.Println("Endpoints:")
	fmt.Println("  GET /process - Mauvaise utilisation avec defer")
	fmt.Println("  GET /stats   - Voir les statistiques")
	
	if err := http.ListenAndServe(":8081", r); err != nil {
		panic(err)
	}
}