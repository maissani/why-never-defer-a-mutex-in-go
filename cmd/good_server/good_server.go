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
Repository contient les données partagées protégées par un mutex.
Cette structure simule un état partagé typique dans une application Go.

@fields:
  - mu: Mutex pour protéger l'accès concurrent aux données
  - counter: Compteur global des requêtes traitées
  - data: Map simulant des données métier partagées avec structure complexe
*/
type Repository struct {
	mu      sync.Mutex
	counter int
	data    map[string]*DataStruct
}

/*
NewRepository crée et initialise un nouveau repository.

@returns: *Repository - Nouvelle instance avec la map initialisée
*/
func NewRepository() *Repository {
	return &Repository{
		data: make(map[string]*DataStruct),
	}
}

/*
GoodHandler démontre la BONNE PRATIQUE d'utilisation des mutex.
Le mutex est libéré immédiatement après chaque opération critique,
permettant un maximum de parallélisme.

@params:
  - w: http.ResponseWriter pour envoyer la réponse
  - req: *http.Request contenant la requête HTTP

@behavior:
  1. Verrouille le mutex pour la lecture/copie des données
  2. Libère immédiatement le mutex après la copie
  3. Effectue le traitement lourd SANS le mutex
  4. Re-verrouille uniquement pour l'écriture finale
  5. Libère immédiatement après l'écriture

@performance: Cette approche maximise la concurrence et les performances
*/
func (r *Repository) GoodHandler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Première acquisition du mutex pour lecture
	r.mu.Lock()
	r.counter++
	currentCounter := r.counter
	dataCopy := make(map[string]*DataStruct)
	for k, v := range r.data {
		dataCopy[k] = &DataStruct{
			Identifier:   v.Identifier,
			Name:         v.Name,
			IsActive:     v.IsActive,
			Counter:      v.Counter,
			LastModified: v.LastModified,
		}
	}
	r.mu.Unlock() // Libération immédiate après la lecture

	// Traitement lourd SANS le mutex
	time.Sleep(10 * time.Millisecond) // Simule un traitement
	
	// Calcul intensif simulé
	result := 0
	for i := 0; i < 1000000; i++ {
		result += i
	}

	// Deuxième acquisition du mutex uniquement pour l'écriture
	key := fmt.Sprintf("request_%d", currentCounter)
	r.mu.Lock()
	r.data[key] = &DataStruct{
		Identifier:   key,
		Name:         fmt.Sprintf("Request %d", currentCounter),
		IsActive:     true,
		Counter:      result,
		LastModified: time.Now(),
	}
	r.mu.Unlock() // Libération immédiate après l'écriture

	response := map[string]interface{}{
		"method":   "good_no_defer",
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
main initialise et démarre le serveur HTTP démontrant la bonne pratique.

@behavior:
  - Crée un repository partagé
  - Configure les routes avec gorilla/mux
  - Démarre le serveur sur le port 8082
  - Affiche les endpoints disponibles

@endpoints:
  - GET /process : Handler avec bonne utilisation du mutex
  - GET /stats : Statistiques du serveur
*/
func main() {
	repo := NewRepository()
	
	r := mux.NewRouter()
	r.HandleFunc("/process", repo.GoodHandler).Methods("GET")
	r.HandleFunc("/stats", repo.StatsHandler).Methods("GET")

	fmt.Println("GOOD Server (sans defer) starting on :8082")
	fmt.Println("Endpoints:")
	fmt.Println("  GET /process - Bonne utilisation sans defer")
	fmt.Println("  GET /stats   - Voir les statistiques")
	
	if err := http.ListenAndServe(":8082", r); err != nil {
		panic(err)
	}
}