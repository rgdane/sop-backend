# JalanKerja Backend
### Full Documentation: [READ DOCS](https://jalankerja.gitbook.io/jalankerja)

## Project Structures
```bash
.
└── src/
    ├── api/                                 # Layer presentasi (HTTP/API layer)
    │   ├── controllers/
    │   │   └── v1/                          # Versi API v1
    │   │       ├── dto/                     # DTO (Data Transfer Object)
    │   │       │   └── squad_dto.go         # Struct DTO untuk operasi Squad
    │   │       ├── handler/                 # Handler yang menangani logic
    │   │       │   └── squad_handler.go     # Handler fungsi CRUD untuk Squad
    │   │       ├── mapper                   # (Opsional) Mapping antara model, DTO
    │   │       └── squad_controller.go      # Menghubungkan route ke handler Squad
    │   ├── presenters/                      # Output formatter
    │   ├── middleware/                      # Middleware seperti Auth, Logging, dll.
    │   └── routes/                          # Routing layer
    │       └── v1/                          # Versi API v1
    │          ├── routes.go                 # Entry point untuk semua route
    │          └── squad_routes.go           # Daftar route terkait Squad
    ├── cmd/                                 # Entry point aplikasi
    │   └── main.go                          # Fungsi utama untuk menjalankan server
    ├── internal/
    │   └── config/
    │   └── database/
    │   └── repository/    
    │   └── service/    
    └── pkg/
        ├── gorm/                      # Layer repository (akses data ke DB)
        │   └── builder/
        │       └── builder.go             # Repository interface Squad
        └── neo4j/                        # Logika bisnis
            └── builder/                          # Versi API v1
                └── builder.go         # Business logic squad
```

# Flows

![Flow Diagram](https://jam.dev/cdn-cgi/image/width=1000,quality=100,dpr=1/https://cdn-jam-screenshots.jam.dev/4bee731580457b0e55664da35511cc00/screenshot/da869424-f1e0-4435-a921-c790171b9c9d.png)

# Installations

1. Clone Repo

   ```bash
   git clone <repo>
   cd <repo>
   ```

2. Copy environtment

   ```bash
   cp src/.env.example src/.env
   ```

3. Run App

   - Development

   ```javascript
   cd src
   go mod tidy
   go run cmd/main.go
   ```

   - Production

   ```bash
   make build
   make run

   # for log
   make logs
   ```

App run on <http://localhost:5000>


### Generate Swagger Documentation
Ada 3 cara untuk generate dokumentasi Swagger:

#### 1. Menggunakan Makefile (Recommended untuk Linux/Mac)
```bash
make swagger
```

#### 2. Menggunakan Script
**Windows:**
```bash
generate-swagger.bat
```

**Linux/Mac:**
```bash
./generate-swagger.sh
```

#### 3. Manual Command
```bash
cd src
swag init -g cmd/main.go -o docs --parseDependency --parseInternal
```

## 📖 Viewing Documentation

1. Jalankan server:
```bash
go run src/cmd/main.go
```

2. Buka browser dan akses:
```
http://localhost:8080/swagger/index.html
```

