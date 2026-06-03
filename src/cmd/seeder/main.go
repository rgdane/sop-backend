package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"jk-api/internal/config"
	"jk-api/internal/database/migrations"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/gorm"
)

type Config struct {
	DB    string
	Scale string
}

type SeedScale struct {
	DivisionCount int
	TitleCount    int
	SopCount      int
	SpkCount      int
	SpkJobCount   int
	SopJobCount   int
}

var scaleMap = map[string]SeedScale{
	"50k":   {DivisionCount: 150, TitleCount: 750, SopCount: 8000, SpkCount: 5000, SpkJobCount: 10000, SopJobCount: 26100},
	"250k":  {DivisionCount: 750, TitleCount: 3750, SopCount: 40000, SpkCount: 25000, SpkJobCount: 50000, SopJobCount: 130500},
	"1250k": {DivisionCount: 3750, TitleCount: 18750, SopCount: 200000, SpkCount: 125000, SpkJobCount: 250000, SopJobCount: 652500},
}

func main() {
	gofakeit.Seed(42)

	cfg := parseFlags()

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	scaleConfig := scaleMap[cfg.Scale]

	switch cfg.DB {
	case "postgres":
		if err := config.PostgresApp(); err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer closePostgres()
		runPostgresSeeder(config.DB, scaleConfig)

	case "neo4j":
		if err := config.Neo4jApp(); err != nil {
			log.Fatalf("Failed to connect to Neo4j: %v", err)
		}
		defer closeNeo4j()
		runNeo4jSeeder(config.GetNeo4j(), scaleConfig)
	}
}

func parseFlags() Config {
	db := flag.String("db", "postgres", "Database target: postgres or neo4j")
	scale := flag.String("scale", "50k", "Data scale: 50k, 250k, or 1250k")
	flag.Parse()

	if err := validateDB(*db); err != nil {
		log.Fatal(err)
	}
	if err := validateScale(*scale); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Memulai proses seeding untuk DB: %s dengan skala: %s...\n", *db, *scale)

	return Config{DB: *db, Scale: *scale}
}

func validateDB(db string) error {
	if db != "postgres" && db != "neo4j" {
		return fmt.Errorf("error: value \"%s\" is invalid for --db\nPilihan yang benar: postgres, neo4j", db)
	}
	return nil
}

func validateScale(scale string) error {
	if _, ok := scaleMap[scale]; !ok {
		return fmt.Errorf("error: value \"%s\" is invalid for --scale\nPilihan yang benar: 50k, 250k, 1250k", scale)
	}
	return nil
}

func closePostgres() {
	if sqlDB, err := config.DB.DB(); err == nil {
		sqlDB.Close()
		log.Println("PostgreSQL connection closed")
	}
}

func closeNeo4j() {
	config.CloseNeo4j()
}

func runPostgresSeeder(db *gorm.DB, scale SeedScale) {
	log.Println("Running migrations...")
	migrations.Migrate()

	log.Println("Seeding Flowcharts...")
	seedFlowchartsPG(db)
	log.Println("Flowchart seeding complete: 2 records")

	log.Println("Seeding Divisions...")
	divisionIDs := seedDivisionsPG(db, scale.DivisionCount)
	log.Printf("Division seeding complete: %d records", len(divisionIDs))

	log.Println("Seeding Titles...")
	titleIDs := seedTitlesPG(db, scale.TitleCount, divisionIDs)
	log.Printf("Title seeding complete: %d records", len(titleIDs))

	log.Println("Seeding SOPs...")
	sopIDs := seedSOPsPG(db, scale.SopCount, divisionIDs)
	log.Printf("SOP seeding complete: %d records", len(sopIDs))

	log.Println("Seeding SPKs...")
	spkIDs := seedSPKsPG(db, scale.SpkCount, titleIDs)
	log.Printf("SPK seeding complete: %d records", len(spkIDs))

	log.Println("Seeding SPK Jobs...")
	seedSpkJobsPG(db, scale.SpkJobCount, spkIDs, titleIDs, sopIDs)
	log.Printf("SPK Job seeding complete: %d records", scale.SpkJobCount)

	log.Println("Seeding SOP Jobs...")
	seedSopJobsPG(db, scale.SopJobCount, sopIDs, spkIDs, titleIDs)
	log.Printf("SOP Job seeding complete: %d records", scale.SopJobCount)

	log.Println("Postgres seeding complete!")
}

func runNeo4jSeeder(driver neo4j.DriverWithContext, scale SeedScale) {
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Neo4j connectivity check failed: %v", err)
	}

	setupNeo4jConstraints(driver)

	log.Println("Seeding Flowcharts...")
	seedFlowchartsNeo4j(driver)
	log.Println("Flowchart seeding complete: 2 records")

	log.Println("Seeding Divisions...")
	divisionIDs := seedDivisionsNeo4j(driver, scale.DivisionCount)
	log.Printf("Division seeding complete: %d records", len(divisionIDs))

	log.Println("Seeding Titles...")
	titleIDs, divisionTitleMap, titleDivisionMap := seedTitlesNeo4j(driver, scale.TitleCount, divisionIDs)
	log.Printf("Title seeding complete: %d records", len(titleIDs))

	log.Println("Creating Title-Division relationships...")
	seedTitleDivisionRelationsNeo4j(driver, scale.TitleCount, divisionIDs)
	log.Println("Title-Division relations complete")

	log.Println("Seeding SOPs...")
	sopIDs, sopDivisionMap := seedSOPsNeo4j(driver, scale.SopCount, divisionIDs)
	log.Printf("SOP seeding complete: %d records", len(sopIDs))

	log.Println("Seeding SPKs...")
	spkIDs, spkTitleMap := seedSPKsNeo4j(driver, scale.SpkCount, titleIDs)
	log.Printf("SPK seeding complete: %d records", len(spkIDs))

	log.Println("Seeding Jobs (SPK + SOP)...")
	seedJobsNeo4j(driver, scale.SpkJobCount, scale.SopJobCount, sopIDs, spkIDs, sopDivisionMap, divisionTitleMap, spkTitleMap, titleDivisionMap, titleIDs)
	log.Printf("Job seeding complete: %d records (SPK: %d, SOP: %d)", scale.SpkJobCount+scale.SopJobCount, scale.SpkJobCount, scale.SopJobCount)

	log.Println("Neo4j seeding complete!")
}

func setupNeo4jConstraints(driver neo4j.DriverWithContext) {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	constraints := []string{
		"CREATE CONSTRAINT division_id_unique IF NOT EXISTS FOR (d:Division) REQUIRE d.id IS UNIQUE",
		"CREATE CONSTRAINT title_id_unique IF NOT EXISTS FOR (t:Title) REQUIRE t.id IS UNIQUE",
		"CREATE CONSTRAINT flowchart_id_unique IF NOT EXISTS FOR (f:Flowchart) REQUIRE f.id IS UNIQUE",
		"CREATE CONSTRAINT sop_id_unique IF NOT EXISTS FOR (s:SOP) REQUIRE s.id IS UNIQUE",
		"CREATE CONSTRAINT spk_id_unique IF NOT EXISTS FOR (s:SPK) REQUIRE s.id IS UNIQUE",
		"CREATE CONSTRAINT job_id_unique IF NOT EXISTS FOR (j:Job) REQUIRE j.id IS UNIQUE",
		"CREATE INDEX title_division_id IF NOT EXISTS FOR (t:Title) ON (t.divisionId)",
	}

	for _, cypher := range constraints {
		_, err := session.Run(ctx, cypher, nil)
		if err != nil {
			log.Printf("Constraint creation (may already exist): %v", err)
		}
	}
}

func generateCode(prefix string, id int64) string {
	return fmt.Sprintf("%s%04d", prefix, id)
}

func generateDocName(prefix string) (name, alias string) {
	words := []string{
		gofakeit.VerbAction(),
		gofakeit.Noun(),
		gofakeit.Noun(),
	}
	words[0] = strings.ToUpper(words[0][:1]) + words[0][1:]

	phrase := strings.Join(words, " ")
	name = prefix + " " + phrase

	var initials []string
	for _, w := range words {
		if len(w) > 0 {
			initials = append(initials, strings.ToUpper(w[:1]))
		}
	}
	alias = prefix + " " + strings.Join(initials, "")
	return
}

func generateSopJobName() (name, alias string) {
	roles := []string{
		"Product Owner", "Project Manager", "Engineering Manager",
		"UI/UX Designer", "Software Engineer", "QA Engineer",
		"System Analyst", "Scrum Master", "Tech Lead", "CTO", "COO",
		"Business Analyst", "Data Engineer", "DevOps Engineer",
		"Product Designer", "Frontend Engineer", "Backend Engineer",
		"Full Stack Developer", "IT Support", "Security Engineer",
		"Database Administrator", "Network Engineer", "Solution Architect",
	}

	actions := []string{
		"melaksanakan meeting", "melakukan brainstorming",
		"menyusun dokumen", "mengevaluasi hasil",
		"membahas progress", "mereview deliverable",
		"mengkoordinasikan", "memimpin diskusi",
		"menyiapkan presentasi", "menganalisa kebutuhan",
		"membuat laporan", "mengembangkan fitur",
		"mengimplementasikan", "mengintegrasikan",
		"mengoptimalkan", "merefaktor kode",
		"mendokumentasikan", "memvalidasi",
		"mengaudit", "merancang arsitektur",
	}

	entities := []string{
		"COO", "CTO", "tim Engineering", "tim Product",
		"tim QA", "tim DevOps", "stakeholder", "client",
		"vendor", "tim Marketing", "tim Sales",
		"Divisi Developer", "Divisi Designer", "Data Team",
		"Infrastructure Team", "user", "management",
	}

	role := roles[gofakeit.Number(0, len(roles)-1)]
	action := actions[gofakeit.Number(0, len(actions)-1)]
	entity := entities[gofakeit.Number(0, len(entities)-1)]

	name = fmt.Sprintf("%s %s %s", role, action, entity)

	var roleAbbr string
	for _, part := range strings.Split(role, " ") {
		if len(part) > 0 {
			roleAbbr += strings.ToUpper(part[:1])
		}
	}
	alias = fmt.Sprintf("%s %s %s", roleAbbr, action, entity)

	return
}

func generateColor() string {
	colors := []string{"#FF5733", "#33FF57", "#3357FF", "#FF33F5", "#F3FF33", "#33FFF5", "#FF8C33", "#8C33FF"}
	return colors[gofakeit.Number(0, len(colors)-1)]
}

type FlowchartPG struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Type string `gorm:"type:text;not null"`
}

func (FlowchartPG) TableName() string { return "flowcharts" }

type DivisionPG struct {
	ID        int64  `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name      string `gorm:"type:varchar(255);not null"`
	Code      string `gorm:"type:varchar(50);uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (DivisionPG) TableName() string { return "divisions" }

type TitlePG struct {
	ID         int64  `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name       string `gorm:"type:varchar(255);not null"`
	Code       string `gorm:"type:varchar(255);uniqueIndex"`
	Color      string `gorm:"type:varchar(20)"`
	DivisionID int64  `gorm:"index"`
}

func (TitlePG) TableName() string { return "titles" }

type SopPG struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Code        string    `gorm:"type:varchar(255)"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SopPG) TableName() string { return "sops" }

type SopDivisionPG struct {
	SopID      int64 `gorm:"primaryKey"`
	DivisionID int64 `gorm:"primaryKey"`
}

func (SopDivisionPG) TableName() string { return "sop_divisions" }

type SpkPG struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Code        string    `gorm:"type:varchar(255);uniqueIndex"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SpkPG) TableName() string { return "spks" }

type SpkTitlePG struct {
	SpkID   int64 `gorm:"primaryKey"`
	TitleID int64 `gorm:"primaryKey"`
}

func (SpkTitlePG) TableName() string { return "spk_titles" }

type SopJobPG struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Alias       string    `gorm:"type:varchar(255)"`
	Type        *string   `gorm:"type:text"`
	Code        string    `gorm:"type:varchar(255)"`
	Description *string   `gorm:"type:text"`
	TitleID     *int64    `gorm:"index"`
	SopID       int64     `gorm:"index"`
	ReferenceID *int64
	Index       int       `gorm:"default:0"`
	IsPublished *bool     `gorm:"default:false"`
	IsHide      *bool     `gorm:"default:false"`
	FlowchartID *int64    `gorm:"default:1;index"`
	NextIndex   *int
	PrevIndex   *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SopJobPG) TableName() string { return "sop_jobs" }

type SpkJobPG struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;type:bigint"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description *string   `gorm:"type:text"`
	SpkID       int64     `gorm:"index"`
	SopID       *int64    `gorm:"index"`
	TitleID     *int64    `gorm:"index"`
	Index       int       `gorm:"default:0"`
	FlowchartID *int64    `gorm:"default:1;index"`
	NextIndex   *int
	PrevIndex   *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (SpkJobPG) TableName() string { return "spk_jobs" }

func seedFlowchartsPG(db *gorm.DB) {
	flowcharts := []FlowchartPG{
		{ID: 1, Type: "process"},
		{ID: 2, Type: "decision"},
	}
	if err := db.CreateInBatches(flowcharts, 2).Error; err != nil {
		log.Fatalf("Failed to insert flowcharts: %v", err)
	}
}

func seedFlowchartsNeo4j(driver neo4j.DriverWithContext) {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	rows := []map[string]interface{}{
		{"id": int64(1), "type": "process"},
		{"id": int64(2), "type": "decision"},
	}

	cypher := "UNWIND $batch AS row CREATE (f:Flowchart {id: row.id, type: row.type})"
	if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
		log.Fatalf("Failed to insert flowcharts: %v", err)
	}
}

func generateDivisionName(id int64) string {
	baseDivisions := []string{
		"Information Technology", "Human Resources", "Management", "Product",
		"Engineering", "Marketing", "Sales", "Finance", "Accounting",
		"Legal", "Operations", "Research & Development", "Customer Success",
		"Business Development", "Corporate Strategy", "Data Science",
		"Security", "Infrastructure", "Quality Assurance", "Design",
		"Communications", "Procurement", "Logistics", "Supply Chain",
		"Risk Management", "Compliance", "Internal Audit",
		"Public Relations", "Investor Relations", "Innovation",
		"Digital Transformation", "IT Support", "Network Operations",
		"Database Administration", "Cloud Services", "DevOps",
		"Platform Engineering", "Data Engineering", "Machine Learning",
		"AI Research", "Business Intelligence", "Analytics",
		"Product Design", "Brand Marketing", "Growth",
		"Partnerships", "Treasury", "Payroll", "Benefits",
		"Talent Acquisition", "Learning & Development",
		"Employee Relations", "Office Administration", "Facilities",
		"Sustainability", "Content Strategy", "Customer Experience",
		"Corporate Finance", "Tax", "Mergers & Acquisitions",
		"Product Operations", "Performance Marketing", "Alliances",
		"Corporate Communications", "Change Management",
		"Portfolio Management", "Vendor Management", "Asset Management",
		"Regulatory Affairs", "Quality Management", "Strategy & Planning",
	}

	n := len(baseDivisions)
	idx := int((id - 1) % int64(n))
	group := int((id - 1) / int64(n))

	if group == 0 {
		return baseDivisions[idx]
	}

	prefixes := []string{"Senior ", "Junior ", "Assistant ", "Deputy ", "Associate "}
	suffixes := []string{" I", " II", " III", " Alpha", " Beta", " Core", " Global", " Regional"}

	if group <= len(prefixes) {
		return prefixes[group-1] + baseDivisions[idx]
	}
	return baseDivisions[idx] + suffixes[(group-1-len(prefixes))%len(suffixes)]
}

var divisionNameCache = make(map[int64]string)

func getCachedDivisionName(id int64) string {
	if name, ok := divisionNameCache[id]; ok {
		return name
	}
	name := generateDivisionName(id)
	divisionNameCache[id] = name
	return name
}

func seedDivisionsPG(db *gorm.DB, count int) []int64 {
	startID := int64(1)
	divisions := make([]DivisionPG, 0, count)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		divisions = append(divisions, DivisionPG{
			ID:   id,
			Name: getCachedDivisionName(id),
			Code: generateCode("DIV", id),
		})
	}

	if err := db.CreateInBatches(divisions, 500).Error; err != nil {
		log.Fatalf("Failed to insert divisions: %v", err)
	}

	ids := make([]int64, count)
	for i := range divisions {
		ids[i] = divisions[i].ID
	}
	return ids
}

func seedTitlesPG(db *gorm.DB, count int, divisionIDs []int64) []int64 {
	if len(divisionIDs) == 0 {
		log.Fatal("No division IDs provided for title seeding")
	}

	startID := int64(1)
	titles := make([]TitlePG, 0, count)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]

		titles = append(titles, TitlePG{
			ID:         id,
			Name:       gofakeit.JobTitle(),
			Code:       generateCode("TTL", id),
			Color:      generateColor(),
			DivisionID: divisionID,
		})
	}

	if err := db.CreateInBatches(titles, 500).Error; err != nil {
		log.Fatalf("Failed to insert titles: %v", err)
	}

	ids := make([]int64, count)
	for i := range titles {
		ids[i] = titles[i].ID
	}
	return ids
}

func seedDivisionsNeo4j(driver neo4j.DriverWithContext, count int) []int64 {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 500
	var ids []int64
	startID := int64(1)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			rows = append(rows, map[string]interface{}{
				"id":   id,
				"name": getCachedDivisionName(id),
				"code": generateCode("DIV", id),
			})
			ids = append(ids, id)
		}

		cypher := "UNWIND $batch AS row CREATE (d:Division {id: row.id, name: row.name, code: row.code})"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert divisions: %v", err)
		}
	}

	return ids
}

func seedTitlesNeo4j(driver neo4j.DriverWithContext, count int, divisionIDs []int64) ([]int64, map[int64][]int64, map[int64]int64) {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 500
	var ids []int64
	startID := int64(1)
	divisionTitleMap := make(map[int64][]int64)
	titleDivisionMap := make(map[int64]int64, count)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]

			divisionTitleMap[divisionID] = append(divisionTitleMap[divisionID], id)
			titleDivisionMap[id] = divisionID

			rows = append(rows, map[string]interface{}{
				"id":         id,
				"name":       gofakeit.JobTitle(),
				"code":       generateCode("TTL", id),
				"color":      generateColor(),
				"divisionId": divisionID,
			})
			ids = append(ids, id)
		}

		cypher := "UNWIND $batch AS row CREATE (t:Title {id: row.id, name: row.name, code: row.code, color: row.color, divisionId: row.divisionId})"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert titles: %v", err)
		}
	}

	return ids, divisionTitleMap, titleDivisionMap
}

func seedTitleDivisionRelationsNeo4j(driver neo4j.DriverWithContext, titleCount int, divisionIDs []int64) {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 500
	startID := int64(1)

	for batchStart := 0; batchStart < titleCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > titleCount {
			batchEnd = titleCount
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]

			rows = append(rows, map[string]interface{}{
				"titleId":    id,
				"divisionId": divisionID,
			})
		}

		cypher := "UNWIND $batch AS row MATCH (t:Title {id: row.titleId}), (d:Division {id: row.divisionId}) CREATE (t)-[:HAS_DIVISION]->(d)"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to create title-division relations: %v", err)
		}
	}
}

func seedSOPsPG(db *gorm.DB, count int, divisionIDs []int64) []int64 {
	if len(divisionIDs) == 0 {
		log.Fatal("No division IDs provided for SOP seeding")
	}

	startID := int64(1)
	sops := make([]SopPG, 0, count)
	sopDivisions := make([]SopDivisionPG, 0, count)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]
		desc := gofakeit.Sentence(5)
		sopName, _ := generateDocName("SOP")

		sops = append(sops, SopPG{
			ID:          id,
			Name:        sopName,
			Code:        fmt.Sprintf("SOP-%s", strings.ReplaceAll(gofakeit.UUID(), "-", "")[:8]),
			Description: &desc,
		})

		sopDivisions = append(sopDivisions, SopDivisionPG{
			SopID:      id,
			DivisionID: divisionID,
		})
	}

	if err := db.CreateInBatches(sops, 1000).Error; err != nil {
		log.Fatalf("Failed to insert SOPs: %v", err)
	}

	if err := db.CreateInBatches(sopDivisions, 1000).Error; err != nil {
		log.Fatalf("Failed to insert SOP divisions: %v", err)
	}

	ids := make([]int64, count)
	for i := range sops {
		ids[i] = sops[i].ID
	}
	return ids
}

func seedSPKsPG(db *gorm.DB, count int, titleIDs []int64) []int64 {
	if len(titleIDs) == 0 {
		log.Fatal("No title IDs provided for SPK seeding")
	}

	startID := int64(1)
	spks := make([]SpkPG, 0, count)
	spkTitles := make([]SpkTitlePG, 0, count)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
		desc := gofakeit.Sentence(5)
		spkName, _ := generateDocName("SPK")

		spks = append(spks, SpkPG{
			ID:          id,
			Name:        spkName,
			Code:        generateCode("SPK", id),
			Description: &desc,
		})

		spkTitles = append(spkTitles, SpkTitlePG{
			SpkID:   id,
			TitleID: titleID,
		})
	}

	if err := db.CreateInBatches(spks, 1000).Error; err != nil {
		log.Fatalf("Failed to insert SPKs: %v", err)
	}

	if err := db.CreateInBatches(spkTitles, 1000).Error; err != nil {
		log.Fatalf("Failed to insert SPK titles: %v", err)
	}

	ids := make([]int64, count)
	for i := range spks {
		ids[i] = spks[i].ID
	}
	return ids
}

func seedSopJobsPG(db *gorm.DB, count int, sopIDs []int64, spkIDs []int64, titleIDs []int64) {
	if len(sopIDs) == 0 {
		log.Fatal("No SOP IDs provided for SopJob seeding")
	}
	if len(spkIDs) == 0 {
		log.Fatal("No SPK IDs provided for SopJob seeding")
	}
	if len(titleIDs) == 0 {
		log.Fatal("No Title IDs provided for SopJob seeding")
	}

	startID := int64(1)
	isPublished := true
	isHide := false

	jobs := make([]SopJobPG, 0, count)
	sopIndex := make(map[int64]int)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		sopID := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
		titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
		desc := gofakeit.Sentence(5)
		flowchartID := int64(gofakeit.Number(1, 2))
		jobType := gofakeit.RandomString([]string{"sop", "spk", "instruction"})

		sopIndex[sopID]++
		idx := sopIndex[sopID]

		jobName, jobAlias := generateSopJobName()

		var refID *int64
		switch jobType {
		case "sop":
			rid := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
			for rid == sopID {
				rid = sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
			}
			refID = &rid
		case "spk":
			rid := spkIDs[gofakeit.Number(0, len(spkIDs)-1)]
			refID = &rid
		}

		jobs = append(jobs, SopJobPG{
			ID:          id,
			Name:        jobName,
			Alias:       jobAlias,
			Type:        &jobType,
			Code:        fmt.Sprintf("P%04d", id),
			Description: &desc,
			TitleID:     &titleID,
			SopID:       sopID,
			ReferenceID: refID,
			Index:       idx,
			IsPublished: &isPublished,
			IsHide:      &isHide,
			FlowchartID: &flowchartID,
			NextIndex:   nil,
			PrevIndex:   nil,
		})
	}

	if err := db.CreateInBatches(jobs, 2000).Error; err != nil {
		log.Fatalf("Failed to insert SOP Jobs: %v", err)
	}
}

func seedSpkJobsPG(db *gorm.DB, count int, spkIDs []int64, titleIDs []int64, sopIDs []int64) {
	if len(spkIDs) == 0 {
		log.Fatal("No SPK IDs provided for SpkJob seeding")
	}
	if len(titleIDs) == 0 {
		log.Fatal("No Title IDs provided for SpkJob seeding")
	}
	if len(sopIDs) == 0 {
		log.Fatal("No SOP IDs provided for SpkJob seeding")
	}

	startID := int64(1)

	jobs := make([]SpkJobPG, 0, count)
	spkIndex := make(map[int64]int)

	for i := int64(0); i < int64(count); i++ {
		id := startID + i
		spkID := spkIDs[gofakeit.Number(0, len(spkIDs)-1)]
		sopID := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
		titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
		desc := gofakeit.Sentence(5)
		flowchartID := int64(gofakeit.Number(1, 2))

		spkIndex[spkID]++
		idx := spkIndex[spkID]

		jobName, _ := generateSopJobName()

		jobs = append(jobs, SpkJobPG{
			ID:          id,
			Name:        jobName,
			Description: &desc,
			SpkID:       spkID,
			SopID:       &sopID,
			TitleID:     &titleID,
			Index:       idx,
			FlowchartID: &flowchartID,
			NextIndex:   nil,
			PrevIndex:   nil,
		})
	}

	if err := db.CreateInBatches(jobs, 2000).Error; err != nil {
		log.Fatalf("Failed to insert SPK Jobs: %v", err)
	}
}

func seedSOPsNeo4j(driver neo4j.DriverWithContext, count int, divisionIDs []int64) ([]int64, map[int64]int64) {
	if len(divisionIDs) == 0 {
		log.Fatal("No division IDs provided for SOP seeding")
	}

	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 500
	var ids []int64
	startID := int64(1)
	sopDivisionMap := make(map[int64]int64, count)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]
			sopName, _ := generateDocName("SOP")

			sopDivisionMap[id] = divisionID

			rows = append(rows, map[string]interface{}{
				"id":          id,
				"name":        sopName,
				"code":        fmt.Sprintf("SOP-%s", strings.ReplaceAll(gofakeit.UUID(), "-", "")[:8]),
				"description": gofakeit.Sentence(5),
				"division_id": divisionID,
			})
			ids = append(ids, id)
		}

		cypher := "UNWIND $batch AS row " +
			"MATCH (d:Division {id: row.division_id}) " +
			"CREATE (s:SOP {id: row.id, name: row.name, code: row.code, description: row.description}) " +
			"CREATE (d)-[:HAS_SOP]->(s)"

		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert SOPs: %v", err)
		}
	}

	return ids, sopDivisionMap
}

func seedSPKsNeo4j(driver neo4j.DriverWithContext, count int, titleIDs []int64) ([]int64, map[int64]int64) {
	if len(titleIDs) == 0 {
		log.Fatal("No title IDs provided for SPK seeding")
	}

	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 500
	var ids []int64
	startID := int64(1)
	spkTitleMap := make(map[int64]int64, count)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]

			spkTitleMap[id] = titleID

			spkName, _ := generateDocName("SPK")

			rows = append(rows, map[string]interface{}{
				"id":          id,
				"name":        spkName,
				"code":        generateCode("SPK", id),
				"description": gofakeit.Sentence(5),
				"title_id":    titleID,
			})
			ids = append(ids, id)
		}

		cypher := "UNWIND $batch AS row " +
			"MATCH (t:Title {id: row.title_id}) " +
			"CREATE (s:SPK {id: row.id, name: row.name, code: row.code, description: row.description}) " +
			"CREATE (t)-[:HAS_SPK]->(s)"

		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert SPKs: %v", err)
		}
	}

	return ids, spkTitleMap
}

func seedJobsNeo4j(driver neo4j.DriverWithContext, spkJobCount int, sopJobCount int, sopIDs []int64, spkIDs []int64, sopDivisionMap map[int64]int64, divisionTitleMap map[int64][]int64, spkTitleMap map[int64]int64, titleDivisionMap map[int64]int64, titleIDs []int64) {
	if len(sopIDs) == 0 || len(spkIDs) == 0 {
		log.Fatal("No SOP/SPK IDs provided for Job seeding")
	}

	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	batchSize := 5000
	var jobIDCounter int64 = 1

	pickTitle := func(divisionID int64) int64 {
		titles := divisionTitleMap[divisionID]
		if len(titles) == 0 {
			return titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
		}
		return titles[gofakeit.Number(0, len(titles)-1)]
	}

	// --- SPK Jobs ---
	spkIndex := make(map[int64]int)
	for batchStart := 0; batchStart < spkJobCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > spkJobCount {
			batchEnd = spkJobCount
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			spkID := spkIDs[gofakeit.Number(0, len(spkIDs)-1)]
			flowchartID := int64(gofakeit.Number(1, 2))
			jobName, _ := generateSopJobName()
			jobType := gofakeit.RandomString([]string{"sop", "instruction"})

			spkIndex[spkID]++
			idx := spkIndex[spkID]

			// Determine ASSIGNED_TO: trace SPK -> Title -> Division -> random Title
			spkTitleID := spkTitleMap[spkID]
			spkDivisionID := titleDivisionMap[spkTitleID]
			assignedTitleID := pickTitle(spkDivisionID)

			// Determine HAS_REFERENCE — SpkJob only refs SOP, never SPK
			var refType string
			var refID int64
			hasRef := false
			if jobType == "sop" {
				rid := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
				refType = "SOP"
				refID = rid
				hasRef = true
			}

			row := map[string]interface{}{
				"id":              jobIDCounter,
				"name":            jobName,
				"description":     gofakeit.Sentence(5),
				"spk_id":          spkID,
				"flowchart_id":    flowchartID,
				"type":            jobType,
				"title_id":        assignedTitleID,
				"index":           idx,
			}
			if hasRef {
				row["ref_type"] = refType
				row["ref_id"] = refID
			}
			rows = append(rows, row)
			jobIDCounter++
		}

		// Create Job nodes with ASSIGNED_TO
		cypher := "UNWIND $batch AS row " +
			"MATCH (s:SPK {id: row.spk_id}), (f:Flowchart {id: row.flowchart_id}), (t:Title {id: row.title_id}) " +
			"CREATE (j:Job {id: row.id, name: row.name, description: row.description, type: row.type, index: row.index}) " +
			"CREATE (s)-[:HAS_JOB]->(j) " +
			"CREATE (j)-[:HAS_FLOWCHART]->(f) " +
			"CREATE (j)-[:ASSIGNED_TO]->(t)"

		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert SPK Jobs: %v", err)
		}

		// HAS_REFERENCE — split by ref_type to use correct label (SOP/SPK)
		sopRefRows := make([]map[string]interface{}, 0)
		spkRefRows := make([]map[string]interface{}, 0)
		for _, row := range rows {
			if rt, ok := row["ref_type"]; ok {
				r := map[string]interface{}{
					"job_id": row["id"],
					"ref_id": row["ref_id"],
				}
				if rt == "SOP" {
					sopRefRows = append(sopRefRows, r)
				} else {
					spkRefRows = append(spkRefRows, r)
				}
			}
		}
		if len(sopRefRows) > 0 {
			cypher := "UNWIND $batch AS row " +
				"MATCH (j:Job {id: row.job_id}), (ref:SOP {id: row.ref_id}) " +
				"CREATE (j)-[:HAS_REFERENCE]->(ref)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": sopRefRows}); err != nil {
				log.Fatalf("Failed to create SPK Job SOP references: %v", err)
			}
		}
		if len(spkRefRows) > 0 {
			cypher := "UNWIND $batch AS row " +
				"MATCH (j:Job {id: row.job_id}), (ref:SPK {id: row.ref_id}) " +
				"CREATE (j)-[:HAS_REFERENCE]->(ref)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": spkRefRows}); err != nil {
				log.Fatalf("Failed to create SPK Job SPK references: %v", err)
			}
		}
	}

	// --- SOP Jobs ---
	sopIndex := make(map[int64]int)
	for batchStart := 0; batchStart < sopJobCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > sopJobCount {
			batchEnd = sopJobCount
		}

		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		for i := int64(batchStart); i < int64(batchEnd); i++ {
			sopID := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
			flowchartID := int64(gofakeit.Number(1, 2))
			jobName, _ := generateSopJobName()
			jobType := gofakeit.RandomString([]string{"sop", "spk", "instruction"})

			sopIndex[sopID]++
			idx := sopIndex[sopID]

			// Determine ASSIGNED_TO: SOP -> Division -> random Title
			sopDivisionID := sopDivisionMap[sopID]
			assignedTitleID := pickTitle(sopDivisionID)

			// Determine HAS_REFERENCE
			var refType string
			var refID int64
			hasRef := false
			switch jobType {
			case "sop":
				rid := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
				for rid == sopID {
					rid = sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
				}
				refType = "SOP"
				refID = rid
				hasRef = true
			case "spk":
				rid := spkIDs[gofakeit.Number(0, len(spkIDs)-1)]
				refType = "SPK"
				refID = rid
				hasRef = true
			}

			row := map[string]interface{}{
				"id":              jobIDCounter,
				"name":            jobName,
				"description":     gofakeit.Sentence(5),
				"sop_id":          sopID,
				"flowchart_id":    flowchartID,
				"type":            jobType,
				"title_id":        assignedTitleID,
				"index":           idx,
			}
			if hasRef {
				row["ref_type"] = refType
				row["ref_id"] = refID
			}
			rows = append(rows, row)
			jobIDCounter++
		}

		// Create Job nodes with ASSIGNED_TO
		cypher := "UNWIND $batch AS row " +
			"MATCH (s:SOP {id: row.sop_id}), (f:Flowchart {id: row.flowchart_id}), (t:Title {id: row.title_id}) " +
			"CREATE (j:Job {id: row.id, name: row.name, description: row.description, type: row.type, index: row.index}) " +
			"CREATE (s)-[:HAS_JOB]->(j) " +
			"CREATE (j)-[:HAS_FLOWCHART]->(f) " +
			"CREATE (j)-[:ASSIGNED_TO]->(t)"

		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
			log.Fatalf("Failed to insert SOP Jobs: %v", err)
		}

		// HAS_REFERENCE — split by ref_type to use correct label (SOP/SPK)
		sopRefRows := make([]map[string]interface{}, 0)
		spkRefRows := make([]map[string]interface{}, 0)
		for _, row := range rows {
			if rt, ok := row["ref_type"]; ok {
				r := map[string]interface{}{
					"job_id": row["id"],
					"ref_id": row["ref_id"],
				}
				if rt == "SOP" {
					sopRefRows = append(sopRefRows, r)
				} else {
					spkRefRows = append(spkRefRows, r)
				}
			}
		}
		if len(sopRefRows) > 0 {
			cypher := "UNWIND $batch AS row " +
				"MATCH (j:Job {id: row.job_id}), (ref:SOP {id: row.ref_id}) " +
				"CREATE (j)-[:HAS_REFERENCE]->(ref)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": sopRefRows}); err != nil {
				log.Fatalf("Failed to create SOP Job SOP references: %v", err)
			}
		}
		if len(spkRefRows) > 0 {
			cypher := "UNWIND $batch AS row " +
				"MATCH (j:Job {id: row.job_id}), (ref:SPK {id: row.ref_id}) " +
				"CREATE (j)-[:HAS_REFERENCE]->(ref)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": spkRefRows}); err != nil {
				log.Fatalf("Failed to create SOP Job SPK references: %v", err)
			}
		}
	}
}
