# POC - Impact du defer sur les performances des Mutex en Go

## 🎯 Objectif

Ce POC (Proof of Concept) démontre pourquoi l'utilisation de `defer` avec les mutex est une **mauvaise pratique** qui peut drastiquement impacter les performances d'une application Go sous charge concurrente.

## 🚨 Le Problème

Beaucoup de développeurs Go utilisent systématiquement `defer` pour libérer les mutex :

```go
mu.Lock()
defer mu.Unlock()  // ❌ MAUVAISE PRATIQUE
// ... traitement long ...
```

Cette approche maintient le mutex verrouillé pendant **toute la durée de la fonction**, créant un goulot d'étranglement critique.

## ✅ La Solution

Libérer le mutex immédiatement après les opérations critiques :

```go
mu.Lock()
// ... opération critique rapide ...
mu.Unlock()  // ✅ BONNE PRATIQUE

// ... traitement long SANS le mutex ...
```

## 📊 Résultats du Benchmark

Les tests comparent deux serveurs HTTP identiques, avec la seule différence étant la gestion des mutex :

### Latence moyenne par requête

| Concurrence | Bad Server (defer) | Good Server | **Amélioration** |
|-------------|-------------------|-------------|------------------|
| 1 goroutine | 11.12 ms | 11.20 ms | -0.7% |
| 10 goroutines | **106.97 ms** | 11.89 ms | **88.9%** |
| 50 goroutines | **419.25 ms** | 13.84 ms | **96.7%** |
| 100 goroutines | **558.35 ms** | 16.17 ms | **97.1%** |

### Throughput (requêtes/seconde)

- **Bad Server** : ~87-90 req/s (constant, peu importe la concurrence)
- **Good Server** : 
  - 1 goroutine : 88 req/s
  - 10 goroutines : **687 req/s** 
  - 50 goroutines : 367 req/s
  - 100 goroutines : 316 req/s

## 🔍 Analyse

1. **Sans concurrence** (1 goroutine) : Performances identiques, pas de contention
2. **Avec concurrence** : Le serveur "bad" devient un **goulot d'étranglement** car une seule goroutine peut traiter à la fois
3. **Impact exponentiel** : Plus la concurrence augmente, plus la dégradation est importante (jusqu'à **97% plus lent**)

## 🏗️ Structure du Projet

- `bad_server.go` : Serveur HTTP avec mutex + defer (port 8081)
- `good_server.go` : Serveur HTTP avec mutex bien utilisés (port 8082)
- `benchmark_test.go` : Tests de charge comparatifs
- `run_benchmark.sh` : Script d'automatisation des tests

## 🚀 Installation et Exécution

### Prérequis
- Go 1.21 ou supérieur
- Git (pour cloner le projet)

### Installation

1. Cloner le projet :
```bash
git clone <url-du-repo>
cd poc
```

2. Installer les dépendances :
```bash
go mod download
go mod tidy
```

### Exécution des Benchmarks

#### Méthode 1 : Script automatique (Recommandé)

Le script `run_benchmark.sh` gère automatiquement le démarrage des serveurs et l'exécution des tests :

```bash
chmod +x run_benchmark.sh
./run_benchmark.sh
```

Le script va :
1. Démarrer les deux serveurs en arrière-plan
2. Vérifier qu'ils répondent correctement
3. Lancer les benchmarks pendant 10 secondes par test
4. Afficher les résultats de latence comparative
5. Arrêter proprement les serveurs

#### Méthode 2 : Exécution manuelle

Si vous préférez contrôler chaque étape :

1. **Terminal 1** - Démarrer le serveur "bad" :
```bash
go run bad_server.go
# Le serveur écoute sur http://localhost:8081
```

2. **Terminal 2** - Démarrer le serveur "good" :
```bash
go run good_server.go
# Le serveur écoute sur http://localhost:8082
```

3. **Terminal 3** - Vérifier que les serveurs fonctionnent :
```bash
# Tester le serveur "bad"
curl http://localhost:8081/stats

# Tester le serveur "good"
curl http://localhost:8082/stats
```

4. **Terminal 3** - Lancer les benchmarks :
```bash
# Benchmarks complets (10 secondes par test)
go test -bench=. -benchtime=10s benchmark_test.go

# Version rapide (1 seconde par test)
go test -bench=. -benchtime=1s benchmark_test.go

# Uniquement les tests de latence
go test -run TestLatencyComparison -v benchmark_test.go
```

### Interpréter les Résultats

Les benchmarks affichent :
- **ns/op** : Nanosecondes par opération
- **ms/req** : Millisecondes par requête (plus facile à lire)
- **req/s** : Requêtes par seconde (throughput)

Plus la concurrence augmente, plus la différence entre les deux approches devient évidente.

## 💡 Leçons Clés

1. **N'utilisez `defer` avec les mutex que pour des opérations très courtes**
2. **Libérez les mutex dès que possible** pour permettre le parallélisme
3. **Copiez les données nécessaires** puis libérez le mutex avant le traitement
4. **L'impact sur les performances peut être catastrophique** (jusqu'à 97% de dégradation)

## 📝 Conclusion

Ce POC prouve que l'utilisation systématique de `defer` avec les mutex est une anti-pattern qui peut transformer votre application en système mono-thread de facto, annulant tous les bénéfices de la concurrence Go.