#!/bin/bash

echo "========================================"
echo " Generating Swagger Documentation"
echo "========================================"
echo ""

cd src

# Create docs directory if it doesn't exist
mkdir -p docs

echo "Generating swagger files..."
swag init -g cmd/main.go -o docs --parseDependency --parseInternal

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo " SUCCESS! Swagger docs generated"
    echo "========================================"
    echo ""
    echo "Files generated:"
    echo "  - src/docs/docs.go"
    echo "  - src/docs/swagger.yaml"
    echo "  - src/docs/swagger.json"
    echo ""
    echo "To view documentation:"
    echo "  1. Run your server"
    echo "  2. Open http://localhost:8080/swagger/index.html"
    echo ""
else
    echo ""
    echo "========================================"
    echo " ERROR: Failed to generate swagger docs"
    echo "========================================"
    echo ""
    exit 1
fi

cd ..
