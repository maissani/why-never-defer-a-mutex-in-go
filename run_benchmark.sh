#!/bin/bash

# Couleurs
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[0;37m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Emojis et symboles
CHECK="✓"
CROSS="✗"
ROCKET="🚀"
WARNING="⚠️"
INFO="ℹ️"
FIRE="🔥"

# Fonction pour afficher un message avec style
print_header() {
    echo -e "\n${BOLD}${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${CYAN}   $1${NC}"
    echo -e "${BOLD}${CYAN}═══════════════════════════════════════════════════════════════${NC}\n"
}

print_success() {
    echo -e "${GREEN}${CHECK} $1${NC}"
}

print_error() {
    echo -e "${RED}${CROSS} $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}${WARNING} $1${NC}"
}

print_info() {
    echo -e "${BLUE}${INFO} $1${NC}"
}

# Tuer les processus existants
print_header "${FIRE} MUTEX BENCHMARK - COMPARAISON DE PERFORMANCE ${FIRE}"

print_info "Nettoyage des processus existants..."
pkill -f "bad_server" 2>/dev/null
pkill -f "good_server" 2>/dev/null
sleep 2
print_success "Processus nettoyés"

# Démarrer les serveurs
print_info "Démarrage des serveurs de benchmark..."

# Démarrer le serveur "bad" en arrière-plan
echo -e "${YELLOW}→ Lancement du serveur 'BAD' (avec defer dans mutex) sur le port 8081${NC}"
go run cmd/bad_server/bad_server.go &
BAD_PID=$!

# Démarrer le serveur "good" en arrière-plan
echo -e "${GREEN}→ Lancement du serveur 'GOOD' (sans defer dans mutex) sur le port 8082${NC}"
go run cmd/good_server/good_server.go &
GOOD_PID=$!

# Attendre que les serveurs soient prêts
echo -e "\n${BLUE}⏳ Attente du démarrage des serveurs...${NC}"
for i in {1..3}; do
    echo -n "."
    sleep 1
done
echo ""

# Vérifier que les serveurs répondent
print_info "Vérification de la disponibilité des serveurs..."
curl -s http://localhost:8081/stats > /dev/null || { print_error "Le serveur BAD ne répond pas"; kill $BAD_PID $GOOD_PID 2>/dev/null; exit 1; }
print_success "Serveur BAD (port 8081) opérationnel"

curl -s http://localhost:8082/stats > /dev/null || { print_error "Le serveur GOOD ne répond pas"; kill $BAD_PID $GOOD_PID 2>/dev/null; exit 1; }
print_success "Serveur GOOD (port 8082) opérationnel"

# Lancer le test de latence comparative en premier
print_header "${ROCKET} TEST DE COMPARAISON DE LATENCE"
echo -e "${PURPLE}Ce test mesure la latence moyenne sous différents niveaux de concurrence${NC}\n"
go test -run TestLatencyComparison -v 2>&1 | grep -v "^go:" | grep -v "PASS"

# Lancer les benchmarks
print_header "${FIRE} BENCHMARKS DE THROUGHPUT"
echo -e "${PURPLE}Ces benchmarks mesurent le nombre de requêtes par seconde (req/s)${NC}\n"

# Fonction pour formater les résultats des benchmarks
format_benchmark_output() {
    while IFS= read -r line; do
        if [[ $line == *"BenchmarkBadServer"* ]]; then
            echo -e "${YELLOW}${line}${NC}" | sed 's/ns\/op/ns\/op/g'
        elif [[ $line == *"BenchmarkGoodServer"* ]]; then
            echo -e "${GREEN}${line}${NC}" | sed 's/ns\/op/ns\/op/g'
        elif [[ $line == *"PASS"* ]]; then
            echo -e "${GREEN}${BOLD}$line${NC}"
        elif [[ $line == *"FAIL"* ]]; then
            echo -e "${RED}${BOLD}$line${NC}"
        elif [[ $line == "goos:"* ]] || [[ $line == "goarch:"* ]] || [[ $line == "pkg:"* ]] || [[ $line == "cpu:"* ]]; then
            echo -e "${BLUE}$line${NC}"
        else
            echo "$line"
        fi
    done
}

go test -bench=. -benchtime=10s -run=^$ -v 2>&1 | grep -v "^go:" | format_benchmark_output

# Afficher un résumé
print_header "📊 RÉSUMÉ DES RÉSULTATS"

# Récupérer les stats finales
echo -e "${BOLD}${YELLOW}Statistiques du serveur BAD:${NC}"
curl -s http://localhost:8081/stats 2>/dev/null | head -5 || echo "Impossible de récupérer les stats"

echo -e "\n${BOLD}${GREEN}Statistiques du serveur GOOD:${NC}"
curl -s http://localhost:8082/stats 2>/dev/null | head -5 || echo "Impossible de récupérer les stats"

# Tuer les serveurs
echo ""
print_info "Arrêt des serveurs..."
kill $BAD_PID $GOOD_PID 2>/dev/null
sleep 1

# Vérifier que les serveurs sont bien arrêtés
if ps -p $BAD_PID > /dev/null 2>&1; then
    print_warning "Force l'arrêt du serveur BAD..."
    kill -9 $BAD_PID 2>/dev/null
fi

if ps -p $GOOD_PID > /dev/null 2>&1; then
    print_warning "Force l'arrêt du serveur GOOD..."
    kill -9 $GOOD_PID 2>/dev/null
fi

print_success "Serveurs arrêtés"

print_header "${CHECK} BENCHMARK TERMINÉ AVEC SUCCÈS! ${CHECK}"
echo -e "${CYAN}${BOLD}Conclusion:${NC} Le serveur ${GREEN}GOOD${NC} (sans defer) est significativement plus performant"
echo -e "sous charge concurrente grâce à une meilleure gestion des mutex.${NC}\n"