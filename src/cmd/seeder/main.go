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

	scale := parseFlags()

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	scaleConfig := scaleMap[scale]

	if err := config.PostgresApp(); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	if err := config.Neo4jApp(); err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}

	defer func() {
		if sqlDB, err := config.DB.DB(); err == nil {
			sqlDB.Close()
			log.Println("PostgreSQL connection closed")
		}
		config.CloseNeo4j()
	}()

	runSeeder(config.DB, config.GetNeo4j(), scaleConfig)
}

func parseFlags() string {
	scale := flag.String("scale", "50k", "Data scale: 50k, 250k, or 1250k")
	flag.Parse()

	if err := validateScale(*scale); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Memulai proses seeding dual-write dengan skala: %s...\n", *scale)
	return *scale
}

func validateScale(scale string) error {
	if _, ok := scaleMap[scale]; !ok {
		return fmt.Errorf("error: value \"%s\" is invalid for --scale\nPilihan yang benar: 50k, 250k, 1250k", scale)
	}
	return nil
}

func runSeeder(db *gorm.DB, driver neo4j.DriverWithContext, scale SeedScale) {
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Neo4j connectivity check failed: %v", err)
	}

	log.Println("Running migrations...")
	migrations.Migrate()

	setupNeo4jConstraints(driver)

	log.Println("Seeding Flowcharts...")
	seedFlowcharts(db, driver)
	log.Println("Flowchart seeding complete: 2 records")

	log.Println("Seeding Divisions & Titles...")
	divisionIDs := seedDivisions(db, driver, scale.DivisionCount)
	log.Printf("Division seeding complete: %d records", len(divisionIDs))

	titleIDs := seedTitles(db, driver, scale.TitleCount, divisionIDs)
	log.Printf("Title seeding complete: %d records", len(titleIDs))

	// log.Println("Creating Title-Division relations in Neo4j...")
	// createTitleDivisionRelations(driver, scale.TitleCount)
	// log.Println("Title-Division relations complete")

	log.Println("Seeding SOPs & SPKs...")
	sopIDs := seedSOPs(db, driver, scale.SopCount, divisionIDs)
	log.Printf("SOP seeding complete: %d records", len(sopIDs))

	spkIDs := seedSPKs(db, driver, scale.SpkCount, titleIDs)
	log.Printf("SPK seeding complete: %d records", len(spkIDs))

	log.Println("Seeding Jobs (SOP Jobs + SPK Jobs)...")
	totalJobs := seedJobs(db, driver, scale.SpkJobCount, scale.SopJobCount, sopIDs, spkIDs, titleIDs)
	log.Printf("Job seeding complete: %d records (SPK: %d, SOP: %d)", totalJobs, scale.SpkJobCount, scale.SopJobCount)

	log.Println("Dual-write seeding complete!")
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
	}

	indexes := []string{
		"CREATE INDEX job_deleted_at IF NOT EXISTS FOR (j:Job) ON (j.deleted_at)",
		"CREATE INDEX job_type IF NOT EXISTS FOR (j:Job) ON (j.type)",
		"CREATE INDEX job_index ON (j:Job) ON (j.index)",
		"CREATE INDEX job_name IF NOT EXISTS FOR (j:Job) ON (j.name)",
		"CREATE INDEX sop_name IF NOT EXISTS FOR (s:SOP) ON (s.name)",
		"CREATE INDEX sop_deleted_at IF NOT EXISTS FOR (s:SOP) ON (s.deleted_at)",
		"CREATE INDEX division_name IF NOT EXISTS FOR (d:Division) ON (d.name)",
		"CREATE INDEX title_division_id IF NOT EXISTS FOR (t:Title) ON (t.divisionId)",
		"CREATE FULLTEXT INDEX sopNameIndex IF NOT EXISTS FOR (s:SOP) ON EACH [s.name]",
	}

	for _, cypher := range constraints {
		_, err := session.Run(ctx, cypher, nil)
		if err != nil {
			log.Printf("Constraint creation (may already exist): %v", err)
		}
	}

	for _, idx := range indexes {
		_, err := session.Run(ctx, idx, nil)
		if err != nil {
			log.Printf("Index creation (may already exist): %v", err)
		}
	}
}

func generateCode(prefix string, id int64) string {
	return fmt.Sprintf("%s%04d", prefix, id)
}

func generateJobTaskName() (name, alias string) {
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

func generateDocumentName(prefix string) string {
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

	return fmt.Sprintf("%s %s %s %s", prefix, role, action, entity)
}

func generateColor() string {
	colors := []string{"#FF5733", "#33FF57", "#3357FF", "#FF33F5", "#F3FF33", "#33FFF5", "#FF8C33", "#8C33FF"}
	return colors[gofakeit.Number(0, len(colors)-1)]
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
	CreatedAt  time.Time
	UpdatedAt  time.Time
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

func seedFlowcharts(db *gorm.DB, driver neo4j.DriverWithContext) {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	flowcharts := []FlowchartPG{
		{ID: 1, Type: "process"},
		{ID: 2, Type: "decision"},
	}
	if err := db.CreateInBatches(flowcharts, 2).Error; err != nil {
		log.Fatalf("Failed to insert flowcharts to PostgreSQL: %v", err)
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	rows := []map[string]interface{}{
		{"id": int64(1), "type": "process"},
		{"id": int64(2), "type": "decision"},
	}
	cypher := "UNWIND $batch AS row CREATE (f:Flowchart {id: row.id, type: row.type})"
	if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows}); err != nil {
		log.Fatalf("Failed to insert flowcharts to Neo4j: %v", err)
	}
}

func seedDivisions(db *gorm.DB, driver neo4j.DriverWithContext, count int) []int64 {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	batchSize := 500
	startID := int64(1)
	var neo4jIDs []int64

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		divisions := make([]DivisionPG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)

		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			name := getCachedDivisionName(id)
			code := generateCode("DIV", id)

			divisions = append(divisions, DivisionPG{
				ID:   id,
				Name: name,
				Code: code,
			})

			neo4jRows = append(neo4jRows, map[string]interface{}{
				"id":   id,
				"name": name,
				"code": code,
			})
			neo4jIDs = append(neo4jIDs, id)
		}

		if err := db.CreateInBatches(divisions, 500).Error; err != nil {
			log.Fatalf("Failed to insert divisions to PostgreSQL: %v", err)
		}

		cypher := "UNWIND $batch AS row CREATE (d:Division {id: row.id, name: row.name, code: row.code})"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": neo4jRows}); err != nil {
			log.Fatalf("Failed to insert divisions to Neo4j: %v", err)
		}
	}

	return neo4jIDs
}

func seedTitles(db *gorm.DB, driver neo4j.DriverWithContext, count int, divisionIDs []int64) []int64 {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	batchSize := 500
	startID := int64(1)
	var neo4jIDs []int64

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		titles := make([]TitlePG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)

		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]

			name := gofakeit.JobTitle()
			code := generateCode("TTL", id)
			color := generateColor()

			titles = append(titles, TitlePG{
				ID:         id,
				Name:       name,
				Code:       code,
				Color:      color,
				DivisionID: divisionID,
			})

			neo4jRows = append(neo4jRows, map[string]interface{}{
				"id":         id,
				"name":       name,
				"code":       code,
				"color":      color,
				"divisionId": divisionID,
			})
			neo4jIDs = append(neo4jIDs, id)
		}

		if err := db.CreateInBatches(titles, 500).Error; err != nil {
			log.Fatalf("Failed to insert titles to PostgreSQL: %v", err)
		}

		cypher := "UNWIND $batch AS row " +
			"CREATE (t:Title {id: row.id, name: row.name, code: row.code, color: row.color, divisionId: row.divisionId}) " +
			"WITH row, t MATCH (d:Division {id: row.divisionId}) " +
			"CREATE (d)-[:HAS_TITLE]->(t)"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": neo4jRows}); err != nil {
			log.Fatalf("Failed to insert titles to Neo4j: %v", err)
		}
	}

	return neo4jIDs
}

// func createTitleDivisionRelations(driver neo4j.DriverWithContext, titleCount int) {
// 	ctx := context.Background()
// 	dbName := config.AppConfig.Neo4jDB
// 	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
// 	defer session.Close(ctx)

// 	batchSize := 500
// 	startID := int64(1)

// 	for batchStart := 0; batchStart < titleCount; batchStart += batchSize {
// 		batchEnd := batchStart + batchSize
// 		if batchEnd > titleCount {
// 			batchEnd = titleCount
// 		}

// 		rows := make([]map[string]interface{}, 0, batchEnd-batchStart)
// 		for i := int64(batchStart); i < int64(batchEnd); i++ {
// 			rows = append(rows, map[string]interface{}{
// 				"titleId": startID + i,
// 			})
// 		}

// 		cypher := "UNWIND $batch AS row MATCH (t:Title {id: row.titleId}) MATCH (d:Division) WITH t, d RETURN t.id AS titleId, d.id AS divisionId LIMIT 1"
// 		_, err := session.Run(ctx, cypher, map[string]interface{}{"batch": rows})
// 		if err != nil {
// 		}
// 	}
// }

func seedSOPs(db *gorm.DB, driver neo4j.DriverWithContext, count int, divisionIDs []int64) []int64 {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	batchSize := 500
	startID := int64(1)
	var neo4jIDs []int64

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		sops := make([]SopPG, 0, batchEnd-batchStart)
		sopDivisions := make([]SopDivisionPG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)

		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			divisionID := divisionIDs[gofakeit.Number(0, len(divisionIDs)-1)]

			name := generateDocumentName("SOP")
			code := fmt.Sprintf("SOP-%s", strings.ReplaceAll(gofakeit.UUID(), "-", "")[:8])
			desc := gofakeit.Sentence(5)

			sops = append(sops, SopPG{
				ID:          id,
				Name:        name,
				Code:        code,
				Description: &desc,
			})

			sopDivisions = append(sopDivisions, SopDivisionPG{
				SopID:      id,
				DivisionID: divisionID,
			})

			neo4jRows = append(neo4jRows, map[string]interface{}{
				"id":          id,
				"name":        name,
				"code":        code,
				"description": desc,
				"division_id": divisionID,
			})
			neo4jIDs = append(neo4jIDs, id)
		}

		if err := db.CreateInBatches(sops, 1000).Error; err != nil {
			log.Fatalf("Failed to insert SOPs to PostgreSQL: %v", err)
		}

		if err := db.CreateInBatches(sopDivisions, 1000).Error; err != nil {
			log.Fatalf("Failed to insert SOP divisions to PostgreSQL: %v", err)
		}

		cypher := "UNWIND $batch AS row " +
			"MATCH (d:Division {id: row.division_id}) " +
			"CREATE (s:SOP {id: row.id, name: row.name, code: row.code, description: row.description}) " +
			"CREATE (d)-[:HAS_SOP]->(s)"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": neo4jRows}); err != nil {
			log.Fatalf("Failed to insert SOPs to Neo4j: %v", err)
		}
	}

	return neo4jIDs
}

func seedSPKs(db *gorm.DB, driver neo4j.DriverWithContext, count int, titleIDs []int64) []int64 {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	batchSize := 500
	startID := int64(1)
	var neo4jIDs []int64

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	for batchStart := 0; batchStart < count; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > count {
			batchEnd = count
		}

		spks := make([]SpkPG, 0, batchEnd-batchStart)
		spkTitles := make([]SpkTitlePG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)

		for i := int64(batchStart); i < int64(batchEnd); i++ {
			id := startID + i
			titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]

			name := generateDocumentName("SPK")
			code := generateCode("SPK", id)
			desc := gofakeit.Sentence(5)

			spks = append(spks, SpkPG{
				ID:          id,
				Name:        name,
				Code:        code,
				Description: &desc,
			})

			spkTitles = append(spkTitles, SpkTitlePG{
				SpkID:   id,
				TitleID: titleID,
			})

			neo4jRows = append(neo4jRows, map[string]interface{}{
				"id":          id,
				"name":        name,
				"code":        code,
				"description": desc,
				"title_id":    titleID,
			})
			neo4jIDs = append(neo4jIDs, id)
		}

		if err := db.CreateInBatches(spks, 1000).Error; err != nil {
			log.Fatalf("Failed to insert SPKs to PostgreSQL: %v", err)
		}

		if err := db.CreateInBatches(spkTitles, 1000).Error; err != nil {
			log.Fatalf("Failed to insert SPK titles to PostgreSQL: %v", err)
		}

		cypher := "UNWIND $batch AS row " +
			"MATCH (t:Title {id: row.title_id}) " +
			"CREATE (s:SPK {id: row.id, name: row.name, code: row.code, description: row.description}) " +
			"CREATE (t)-[:HAS_SPK]->(s)"
		if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": neo4jRows}); err != nil {
			log.Fatalf("Failed to insert SPKs to Neo4j: %v", err)
		}
	}

	return neo4jIDs
}

func seedJobs(db *gorm.DB, driver neo4j.DriverWithContext, spkJobCount int, sopJobCount int, sopIDs []int64, spkIDs []int64, titleIDs []int64) int {
	ctx := context.Background()
	dbName := config.AppConfig.Neo4jDB

	batchSize := 2000
	jobIDCounter := int64(1)
	pgJobIDCounter := int64(1)
	isPublished := true
	isHide := false

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite, DatabaseName: dbName})
	defer session.Close(ctx)

	var totalInserted int

	for batchStart := 0; batchStart < spkJobCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > spkJobCount {
			batchEnd = spkJobCount
		}

		spkJobs := make([]SpkJobPG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		spkIndex := make(map[int64]int)

		for i := int64(0); i < int64(batchEnd-batchStart); i++ {
			id := jobIDCounter
			pgID := pgJobIDCounter
			spkID := spkIDs[gofakeit.Number(0, len(spkIDs)-1)]
			sopID := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
			titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
			flowchartID := int64(gofakeit.Number(1, 2))

			jobName, _ := generateJobTaskName()
			desc := gofakeit.Sentence(5)

			spkIndex[spkID]++
			idx := spkIndex[spkID]

			spkJobs = append(spkJobs, SpkJobPG{
				ID:          pgID,
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

			neo4jRows = append(neo4jRows, map[string]interface{}{
				"id":           id,
				"name":         jobName,
				"description":  desc,
				"spk_id":       spkID,
				"title_id":     titleID,
				"flowchart_id": flowchartID,
				"index":        idx,
			})
			jobIDCounter++
			pgJobIDCounter++
		}

		if err := db.CreateInBatches(spkJobs, 2000).Error; err != nil {
			log.Fatalf("Failed to insert SPK Jobs to PostgreSQL: %v", err)
		}

		for _, row := range neo4jRows {
			cypher := "UNWIND $batch AS row " +
				"MATCH (s:SPK {id: row.spk_id}), (f:Flowchart {id: row.flowchart_id}), (t:Title {id: row.title_id}) " +
				"CREATE (j:Job {id: row.id, name: row.name, description: row.description, index: row.index}) " +
				"CREATE (s)-[:HAS_JOB]->(j) " +
				"CREATE (j)-[:HAS_FLOWCHART]->(f) " +
				"CREATE (j)-[:ASSIGNED_TO]->(t)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": []map[string]interface{}{row}}); err != nil {
				log.Fatalf("Failed to insert SPK Job to Neo4j: %v", err)
			}
		}
	}

	for batchStart := 0; batchStart < sopJobCount; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > sopJobCount {
			batchEnd = sopJobCount
		}

		sopJobs := make([]SopJobPG, 0, batchEnd-batchStart)
		neo4jRows := make([]map[string]interface{}, 0, batchEnd-batchStart)
		sopIndex := make(map[int64]int)

		for i := int64(0); i < int64(batchEnd-batchStart); i++ {
			id := jobIDCounter
			pgID := pgJobIDCounter
			sopID := sopIDs[gofakeit.Number(0, len(sopIDs)-1)]
			titleID := titleIDs[gofakeit.Number(0, len(titleIDs)-1)]
			flowchartID := int64(gofakeit.Number(1, 2))

			jobName, jobAlias := generateJobTaskName()
			jobType := gofakeit.RandomString([]string{"sop", "spk", "instruction"})
			desc := gofakeit.Sentence(5)

			sopIndex[sopID]++
			idx := sopIndex[sopID]

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

			sopJobs = append(sopJobs, SopJobPG{
				ID:          pgID,
				Name:        jobName,
				Alias:       jobAlias,
				Type:        &jobType,
				Code:        fmt.Sprintf("P%04d", pgID),
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

			neo4jRow := map[string]interface{}{
				"id":           id,
				"name":         jobName,
				"alias":        jobAlias,
				"type":         jobType,
				"code":         fmt.Sprintf("P%04d", id),
				"description":  desc,
				"sop_id":       sopID,
				"title_id":     titleID,
				"flowchart_id": flowchartID,
				"index":        idx,
			}
			if jobType != "instruction" {
				neo4jRow["ref_id"] = refID
			}
			neo4jRows = append(neo4jRows, neo4jRow)
			jobIDCounter++
			pgJobIDCounter++
		}

		if err := db.CreateInBatches(sopJobs, 2000).Error; err != nil {
			log.Fatalf("Failed to insert SOP Jobs to PostgreSQL: %v", err)
		}

		for _, row := range neo4jRows {
			jobType := row["type"].(string)

			cypher := "UNWIND $batch AS row " +
				"MATCH (s:SOP {id: row.sop_id}), (f:Flowchart {id: row.flowchart_id}), (t:Title {id: row.title_id}) " +
				"CREATE (j:Job {id: row.id, name: row.name, alias: row.alias, type: row.type, code: row.code, description: row.description, index: row.index}) " +
				"CREATE (s)-[:HAS_JOB]->(j) " +
				"CREATE (j)-[:HAS_FLOWCHART]->(f) " +
				"CREATE (j)-[:ASSIGNED_TO]->(t)"
			if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": []map[string]interface{}{row}}); err != nil {
				log.Fatalf("Failed to insert SOP Job to Neo4j: %v", err)
			}
			totalInserted++

			if jobType == "sop" || jobType == "spk" {
				refID := *row["ref_id"].(*int64)
				var refLabel string
				if jobType == "sop" {
					refLabel = "SOP"
				} else {
					refLabel = "SPK"
				}
				cypher = fmt.Sprintf("UNWIND $batch AS row "+
					"MATCH (j:Job {id: row.id}), (ref:%s {id: row.ref_id}) "+
					"CREATE (j)-[:HAS_REFERENCE]->(ref)", refLabel)
				refRow := map[string]interface{}{
					"id":     row["id"],
					"ref_id": refID,
				}
				if _, err := session.Run(ctx, cypher, map[string]interface{}{"batch": []map[string]interface{}{refRow}}); err != nil {
					log.Printf("Warning: failed to create reference relation: %v", err)
				}
			}
		}
	}

	return totalInserted
}
