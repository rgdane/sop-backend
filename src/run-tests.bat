@echo off
setlocal enabledelayedexpansion

echo === JK-API Unit Tests ===
echo.

if "%1"=="services" (
    echo Running Service Layer Tests...
    go test ./test/services -v
    goto :end
)

if "%1"=="handlers" (
    echo Running Handler Layer Tests...
    go test ./test/handlers -v
    goto :end
)

if "%1"=="controllers" (
    echo Running Controller Layer Tests...
    go test ./test/controllers -v
    goto :end
)

if "%1"=="coverage" (
    echo Running All Tests with Coverage...
    go test ./test/... -coverprofile=coverage.out
    echo.
    echo Coverage Report:
    go tool cover -func=coverage.out
    echo.
    echo Opening HTML coverage report...
    go tool cover -html=coverage.out
    goto :end
)

echo Running All Tests...
echo.
go test ./test/... -v

if !ERRORLEVEL! EQU 0 (
    echo.
    echo ✅ All tests passed!
) else (
    echo.
    echo ❌ Some tests failed!
    exit /b 1
)

:end
echo.
echo Usage:
echo   run-tests.bat          - Run all tests
echo   run-tests.bat services - Run service layer tests only
echo   run-tests.bat handlers - Run handler layer tests only
echo   run-tests.bat coverage - Run tests with coverage report
