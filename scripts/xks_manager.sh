#!/bin/bash

# Configuration
TIMEOUT=300  # 5 minutes
CHECK_INTERVAL=10

# Fonction pour vérifier le statut
check_status() {
    xks status 2>/dev/null | grep -q "running\|active\|started"
}

# Fonction pour attendre le démarrage
wait_for_start() {
    local elapsed=0
    echo "Démarrage en cours..."
    
    while [ $elapsed -lt $TIMEOUT ]; do
        if check_status; then
            echo "✓ Cluster démarré avec succès"
            return 0
        fi
        
        sleep $CHECK_INTERVAL
        elapsed=$((elapsed + CHECK_INTERVAL))
        echo "Attente... ($elapsed/${TIMEOUT}s)"
    done
    
    echo "✗ Timeout atteint (${TIMEOUT}s)"
    return 1
}

# Fonction pour attendre l'arrêt
wait_for_stop() {
    local elapsed=0
    echo "Arrêt en cours..."
    
    while [ $elapsed -lt $TIMEOUT ]; do
        if ! check_status; then
            echo "✓ Cluster arrêté avec succès"
            return 0
        fi
        
        sleep $CHECK_INTERVAL
        elapsed=$((elapsed + CHECK_INTERVAL))
        echo "Attente... ($elapsed/${TIMEOUT}s)"
    done
    
    echo "✗ Timeout atteint (${TIMEOUT}s)"
    return 1
}

# Commande start
xks_start() {
    if check_status; then
        echo "⚠ Cluster déjà démarré"
        return 0
    fi
    
    echo "Démarrage du cluster..."
    xks start
    wait_for_start
}

# Commande stop
xks_stop() {
    if ! check_status; then
        echo "⚠ Cluster déjà arrêté"
        return 0
    fi
    
    echo "Arrêt du cluster..."
    xks stop
    wait_for_stop
}

# Usage
case "$1" in
    start)
        xks_start
        ;;
    stop)
        xks_stop
        ;;
    status)
        if check_status; then
            echo "✓ Cluster en cours d'exécution"
        else
            echo "✗ Cluster arrêté"
        fi
        ;;
    *)
        echo "Usage: $0 {start|stop|status}"
        exit 1
        ;;
esac