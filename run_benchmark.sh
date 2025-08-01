#!/bin/bash

echo "Starting benchmark servers..."

# Démarrer le serveur "bad" en arrière-plan
go run bad_server.go &
BAD_PID=$!

# Démarrer le serveur "good" en arrière-plan
go run good_server.go &
GOOD_PID=$!

# Attendre que les serveurs soient prêts
echo "Waiting for servers to start..."
sleep 3

# Vérifier que les serveurs répondent
echo "Checking servers..."
curl -s http://localhost:8081/stats > /dev/null || { echo "Bad server not responding"; kill $BAD_PID $GOOD_PID; exit 1; }
curl -s http://localhost:8082/stats > /dev/null || { echo "Good server not responding"; kill $BAD_PID $GOOD_PID; exit 1; }

echo "Servers are ready. Running benchmarks..."
echo ""

# Lancer les benchmarks
go test -bench=. -benchtime=10s -v

# Lancer le test de latence comparative
go test -run TestLatencyComparison -v

# Tuer les serveurs
echo ""
echo "Stopping servers..."
kill $BAD_PID $GOOD_PID

echo "Benchmark completed!"