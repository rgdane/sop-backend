@echo off
echo ========================================
echo  Generating Swagger Documentation
echo ========================================
echo.

cd src
if not exist docs mkdir docs

echo Generating swagger files...
swag init -g cmd/main.go -o docs --parseDependency --parseInternal

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo  SUCCESS! Swagger docs generated
    echo ========================================
    echo.
    echo Files generated:
    echo   - src\docs\docs.go
    echo   - src\docs\swagger.yaml
    echo   - src\docs\swagger.json
    echo.
    echo To view documentation:
    echo   1. Run your server
    echo   2. Open http://localhost:8080/swagger/index.html
    echo.
) else (
    echo.
    echo ========================================
    echo  ERROR: Failed to generate swagger docs
    echo ========================================
    echo.
)

cd ..
pause
