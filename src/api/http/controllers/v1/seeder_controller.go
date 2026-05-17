package controllers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RunMasterDataSeederV1(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// 1. SEEDING DIVISIONS
		// ==========================================
		divisionsInput := dto.BulkCreateDivisionDto{
			Data: []*dto.CreateDivisionDto{
				&dto.CreateDivisionDto{Name: "Human Resources & General Affairs", Code: "HRGA"},
				&dto.CreateDivisionDto{Name: "Information Technology", Code: "IT"},
				&dto.CreateDivisionDto{Name: "Finance & Accounting", Code: "FIN"},
				&dto.CreateDivisionDto{Name: "Operations & Logistics", Code: "OPR"},
				&dto.CreateDivisionDto{Name: "Marketing & Sales", Code: "MKT"},
				&dto.CreateDivisionDto{Name: "Research & Development", Code: "RND"},
				&dto.CreateDivisionDto{Name: "Legal & Compliance", Code: "LGL"},
				&dto.CreateDivisionDto{Name: "Customer Support", Code: "CS"},
			},
		}

		// Panggil Handler Bulk Create secara langsung (Bypass HTTP Request body)
		divisionsRes, errDiv := cn.DivisionHandler.BulkCreateHandler(&divisionsInput)
		if errDiv != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding Divisions: "+errDiv.Error())
		}

		// ==========================================
		// 2. SEEDING TITLES (JABATAN)
		// ==========================================
		titlesInput := dto.BulkCreateTitleDto{
			Data: []*dto.CreateTitleDto{
				&dto.CreateTitleDto{Name: "Board of Directors", Code: "BOD", Color: "#FF0000"},
				&dto.CreateTitleDto{Name: "General Manager", Code: "GM", Color: "#FF8C00"},
				&dto.CreateTitleDto{Name: "Division Head", Code: "DH", Color: "#FFD700"},
				&dto.CreateTitleDto{Name: "Department Manager", Code: "DM", Color: "#32CD32"},
				&dto.CreateTitleDto{Name: "Branch Manager", Code: "BM", Color: "#00CED1"},
				&dto.CreateTitleDto{Name: "Supervisor", Code: "SPV", Color: "#1E90FF"},
				&dto.CreateTitleDto{Name: "Senior Specialist", Code: "SS", Color: "#8A2BE2"},
				&dto.CreateTitleDto{Name: "Junior Staff", Code: "JS", Color: "#FF1493"},
				&dto.CreateTitleDto{Name: "Intern", Code: "INT", Color: "#A9A9A9"},
			},
		}

		// Panggil Handler Bulk Create Title
		titlesRes, errTitle := cn.TitleHandler.BulkCreateHandler(&titlesInput)
		if errTitle != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding Titles: "+errTitle.Error())
		}

		// ==========================================
		// 3. KEMBALIKAN RESPONSE SUKSES
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Seeding Data Master (Divisions & Titles) Berhasil!",
			"data": fiber.Map{
				"divisions_inserted": len(divisionsInput.Data),
				"titles_inserted":    len(titlesInput.Data),
				"raw_divisions_res":  divisionsRes, // Menampilkan data yang baru saja berhasil dibuat
				"raw_titles_res":     titlesRes,
			},
		})
	}
}

func RunMasterDataSeeder(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// 1. GENERATE 30 DIVISIONS
		// ==========================================
		divisionNames := []string{
			"Human Resources", "General Affairs", "Information Technology", "Finance", "Accounting",
			"Operations", "Logistics", "Marketing", "Sales", "Research & Development",
			"Legal", "Compliance", "Customer Support", "Customer Success", "Procurement",
			"Quality Assurance", "Public Relations", "Facilities Management", "Internal Audit", "Business Development",
			"Product Management", "Data Engineering", "Data Science", "Information Security", "Corporate Strategy",
			"Supply Chain", "Manufacturing", "Corporate Communications", "Investor Relations", "Risk Management",
		}

		var divisionsData []*dto.CreateDivisionDto
		for i, name := range divisionNames {
			// Bikin kode singkatan simple, misal: DIV-1, DIV-2, dst.
			code := fmt.Sprintf("DIV-%03d", i+1)
			divisionsData = append(divisionsData, &dto.CreateDivisionDto{
				Name: name,
				Code: code,
			})
		}

		divisionsInput := dto.BulkCreateDivisionDto{Data: divisionsData}
		divisionsRes, errDiv := cn.DivisionHandler.BulkCreateHandler(&divisionsInput)
		if errDiv != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding Divisions: "+errDiv.Error())
		}

		// ==========================================
		// 2. GENERATE 100 TITLES (JABATAN)
		// ==========================================
		// Kita akan kombinasikan Level dan Spesialisasi untuk dapat 100 kombinasi
		levels := []string{"Intern", "Junior", "Staff", "Senior", "Lead", "Supervisor", "Manager", "Director"}
		specialties := []string{
			"Software Engineer", "Data Analyst", "Product Designer", "Accountant", 
			"HR Specialist", "Marketing Exec", "Sales Rep", "System Admin", 
			"Legal Counsel", "Operations Spec", "QA Tester", "Support Agent",
		}

		var titlesData []*dto.CreateTitleDto

		// 8 levels x 12 specialties = 96 Titles
		counter := 1
		for _, level := range levels {
			for _, spec := range specialties {
				titleName := fmt.Sprintf("%s %s", level, spec)
				code := fmt.Sprintf("TTL-%03d", counter)
				titlesData = append(titlesData, &dto.CreateTitleDto{
					Name:  titleName,
					Code:  code,
					Color: helper.GenerateRandomHexColor(),
				})
				counter++
			}
		}

		// Tambahkan 4 C-Level Executive biar pas 100
		cLevels := []string{"Chief Executive Officer", "Chief Technology Officer", "Chief Financial Officer", "Chief Operating Officer"}
		for i, cName := range cLevels {
			titlesData = append(titlesData, &dto.CreateTitleDto{
				Name:  cName,
				Code:  fmt.Sprintf("C-LVL-%02d", i+1),
				Color: helper.GenerateRandomHexColor(),
			})
		}

		titlesInput := dto.BulkCreateTitleDto{Data: titlesData}
		titlesRes, errTitle := cn.TitleHandler.BulkCreateHandler(&titlesInput)
		if errTitle != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding Titles: "+errTitle.Error())
		}

		// ==========================================
		// 3. RETURN RESPONSE
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Seeding Data Master (30 Divisions & 100 Titles) Berhasil!",
			"data": fiber.Map{
				"divisions_inserted": len(divisionsInput.Data),
				"titles_inserted":    len(titlesInput.Data),
				"raw_divisions_res":  divisionsRes,
				"raw_titles_res":     titlesRes,
			},
		})
	}
}

func RunParentDataSeederV1(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// 1. SEEDING SOPs (Standard Operating Procedures)
		// ==========================================
		sopsInput := dto.BulkCreateSopsDto{
			Data: []*dto.CreateSopDto{
				{
					Name:         "SOP Rekrutmen Karyawan Baru",
					Description:  helper.StrPtr("Prosedur standar dari proses seleksi, wawancara, hingga onboarding karyawan baru."),
					Code:         "SOP-HR-01",
					HasDivisions: []int64{1}, // Relasi ke ID 1: HRGA
					ParentJobID:  nil,        // Nil karena ini parent/root SOP
				},
				{
					Name:         "SOP Penggajian Bulanan",
					Description:  helper.StrPtr("Prosedur perhitungan absensi, pajak, dan pencairan gaji karyawan tiap bulan."),
					Code:         "SOP-FIN-01",
					HasDivisions: []int64{1, 3}, // Relasi ke ID 1 (HRGA) & ID 3 (Finance)
					ParentJobID:  nil,
				},
				{
					Name:         "SOP Maintenance Server Berkala",
					Description:  helper.StrPtr("Prosedur pemeliharaan server fisik dan backup cloud setiap akhir bulan."),
					Code:         "SOP-IT-01",
					HasDivisions: []int64{2}, // Relasi ke ID 2: IT
					ParentJobID:  nil,
				},
				{
					Name:         "SOP Audit Internal",
					Description:  helper.StrPtr("Prosedur pelaksanaan audit internal tahunan perusahaan untuk memastikan kepatuhan."),
					Code:         "SOP-AUD-01",
					HasDivisions: []int64{3, 7}, // Relasi ke ID 3 (Finance) & ID 7 (Legal/Compliance)
					ParentJobID:  nil,
				},
				{
					Name:         "SOP Penanganan Komplain Pelanggan",
					Description:  helper.StrPtr("Prosedur eskalasi dan penyelesaian keluhan dari pelanggan B2B maupun B2C."),
					Code:         "SOP-CS-01",
					HasDivisions: []int64{8}, // Relasi ke ID 8: Customer Support
					ParentJobID:  nil,
				},
			},
		}

		// Panggil Handler Bulk Create SOP
		sopsRes, errSop := cn.SopHandler.BulkCreateSopsHandler(&sopsInput)
		if errSop != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SOPs: "+errSop.Error())
		}

		// ==========================================
		// 2. SEEDING SPKs (Surat Perintah Kerja)
		// ==========================================
		spksInput := dto.BulkCreateSpksDto{
			Data: []*dto.CreateSpkDto{
				{
					Name:        "SPK Pengembangan Aplikasi Internal",
					Description: helper.StrPtr("Surat perintah kerja untuk pembuatan aplikasi portal HR versi 2.0."),
					Code:        "SPK-IT-001",
					HasTitles:   []int64{6, 7, 8}, // Dikerjakan oleh ID: Supervisor, Senior Spec, Junior Staff
				},
				{
					Name:        "SPK Penyusunan Laporan Pajak Tahunan",
					Description: helper.StrPtr("Surat perintah kerja untuk rekonsiliasi dan pelaporan pajak badan perusahaan."),
					Code:        "SPK-FIN-001",
					HasTitles:   []int64{4, 6, 7}, // Dikerjakan oleh ID: Dept Manager, Supervisor, Senior Spec
				},
				{
					Name:        "SPK Kampanye Marketing Kuartal 3",
					Description: helper.StrPtr("Pelaksanaan kampanye promosi produk baru di berbagai platform digital."),
					Code:        "SPK-MKT-001",
					HasTitles:   []int64{3, 4, 6}, // Dikerjakan oleh ID: Div Head, Dept Mgr, Supervisor
				},
				{
					Name:        "SPK Audit Kepatuhan Regulasi",
					Description: helper.StrPtr("Pemeriksaan dokumen legal dan izin usaha perusahaan sesuai regulasi terbaru."),
					Code:        "SPK-LGL-001",
					HasTitles:   []int64{7, 8}, // Dikerjakan oleh ID: Senior Spec, Junior Staff
				},
			},
		}

		// Panggil Handler Bulk Create SPK
		spksRes, errSpk := cn.SpkHandler.BulkCreateSpksHandler(&spksInput)
		if errSpk != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SPKs: "+errSpk.Error())
		}

		// ==========================================
		// 3. KEMBALIKAN RESPONSE SUKSES
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Seeding Data Parent (SOPs & SPKs) Berhasil!",
			"data": fiber.Map{
				"sops_inserted": len(sopsInput.Data),
				"spks_inserted": len(spksInput.Data),
				"raw_sops_res":  sopsRes, // Menampilkan data SOP (termasuk relasi Graph/SQL) yg berhasil dicreate
				"raw_spks_res":  spksRes, // Menampilkan data SPK yg berhasil dicreate
			},
		})
	}
}

func RunParentDataSeeder(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		// Kamus kata untuk bikin nama dinamis
		departments := []string{"Keuangan", "Teknologi", "SDM", "Operasional", "Produksi", "Pemasaran", "Legal", "Logistik"}
		actions := []string{"Pemeliharaan", "Audit", "Evaluasi", "Perencanaan", "Pelaporan", "Pengawasan", "Implementasi"}

		// ==========================================
		// 1. GENERATE 50 SOPs
		// ==========================================
		totalSOPs := 50
		var sopsData []*dto.CreateSopDto

		for i := 1; i <= totalSOPs; i++ {
			dept := departments[r.Intn(len(departments))]
			action := actions[r.Intn(len(actions))]
			codePrefix := strings.ToUpper(dept[:3])

			sopsData = append(sopsData, &dto.CreateSopDto{
				Name:         fmt.Sprintf("SOP %s %s v%d.0", action, dept, r.Intn(5)+1),
				Description:  helper.StrPtr(fmt.Sprintf("Dokumen standar operasional resmi untuk kegiatan %s di departemen %s.", action, dept)),
				Code:         fmt.Sprintf("SOP-%s-%03d", codePrefix, i),
				HasDivisions: helper.GenerateRandomIDs(1, 30, r.Intn(4)+1), // Tiap SOP dipegang 1-4 Divisi acak (ID 1-30)
				ParentJobID:  nil,
			})
		}

		sopsInput := dto.BulkCreateSopsDto{Data: sopsData}
		sopsRes, errSop := cn.SopHandler.BulkCreateSopsHandler(&sopsInput)
		if errSop != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SOPs: "+errSop.Error())
		}

		// ==========================================
		// 2. GENERATE 150 SPKs
		// ==========================================
		totalSPKs := 150
		var spksData []*dto.CreateSpkDto

		for i := 1; i <= totalSPKs; i++ {
			dept := departments[r.Intn(len(departments))]
			action := actions[r.Intn(len(actions))]
			codePrefix := strings.ToUpper(dept[:3])

			spksData = append(spksData, &dto.CreateSpkDto{
				Name:        fmt.Sprintf("SPK %s %s - Kuartal %d", action, dept, r.Intn(4)+1),
				Description: helper.StrPtr(fmt.Sprintf("Surat perintah kerja untuk eksekusi %s pada area %s.", action, dept)),
				Code:        fmt.Sprintf("SPK-%s-%04d", codePrefix, i),
				HasTitles:   helper.GenerateRandomIDs(1, 100, r.Intn(6)+1), // Tiap SPK dikerjakan oleh 1-6 Jabatan acak (ID 1-100)
			})
		}

		spksInput := dto.BulkCreateSpksDto{Data: spksData}
		spksRes, errSpk := cn.SpkHandler.BulkCreateSpksHandler(&spksInput)
		if errSpk != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SPKs: "+errSpk.Error())
		}

		// ==========================================
		// 3. RETURN RESPONSE
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": fmt.Sprintf("Seeding Parent (%d SOPs & %d SPKs) Berhasil!", totalSOPs, totalSPKs),
			"data": fiber.Map{
				"sops_inserted": len(sopsInput.Data),
				"spks_inserted": len(spksInput.Data),
				"raw_sops_res":  sopsRes,
				"raw_spks_res":  spksRes,
			},
		})
	}
}

func RunJobDataSeederV1(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// 1. SEEDING SOP JOBS
		// ==========================================
		sopJobsInput := dto.BulkCreateSopJobs{
			Data: []*dto.CreateSopJobDto{
				// -- Jobs untuk SOP 1: Rekrutmen (ID: 1) --
				{
					Name:        "Screening CV Kandidat",
					Alias:       "screening-cv",
					Description: helper.StrPtr("Mengecek kesesuaian CV dengan kualifikasi yang dibutuhkan."),
					Type:        helper.StrPtr("instruction"), // <-- DIUBAH SESUAI CONSTRAINT
					TitleID:     helper.Int64Ptr(6), // Supervisor
					SopID:       1,
					ReferenceID: nil, // Node awal
					IsPublished: helper.BoolPtr(true),
				},
				{
					Name:        "Interview HR",
					Alias:       "interview-hr",
					Description: helper.StrPtr("Melakukan wawancara awal untuk menilai culture fit."),
					Type:        helper.StrPtr("sop"), // <-- DIUBAH SESUAI CONSTRAINT
					TitleID:     helper.Int64Ptr(4), // Department Manager
					SopID:       1,
					ReferenceID: nil,
					IsPublished: helper.BoolPtr(true),
				},
				// -- Jobs untuk SOP 2: Penggajian Bulanan (ID: 2) --
				{
					Name:        "Rekap Absensi Karyawan",
					Alias:       "rekap-absen",
					Description: helper.StrPtr("Menarik data absensi dari mesin fingerprint ke sistem Excel."),
					Type:        helper.StrPtr("instruction"), // <-- DIUBAH SESUAI CONSTRAINT
					TitleID:     helper.Int64Ptr(8), // Junior Staff
					SopID:       2,
					IsPublished: helper.BoolPtr(true),
				},
				{
					Name:        "Perhitungan PPh 21",
					Alias:       "hitung-pajak",
					Description: helper.StrPtr("Menghitung potongan pajak penghasilan masing-masing karyawan."),
					Type:        helper.StrPtr("instruction"), // <-- DIUBAH SESUAI CONSTRAINT
					TitleID:     helper.Int64Ptr(7), // Senior Specialist
					SopID:       2,
					IsPublished: helper.BoolPtr(true),
				},
				// -- Jobs untuk SOP 3: Maintenance Server (ID: 3) --
				{
					Name:        "Backup Database Utama",
					Alias:       "backup-db",
					Description: helper.StrPtr("Melakukan dump database dan upload ke Cloud Storage AWS."),
					Type:        helper.StrPtr("spk"), // <-- DIUBAH SESUAI CONSTRAINT
					TitleID:     helper.Int64Ptr(7), // Senior Specialist (IT)
					SopID:       3,
					IsPublished: helper.BoolPtr(true),
				},
			},
		}

		// Panggil Handler Bulk Create SOP Jobs
		sopJobsRes, errSopJob := cn.SopJobHandler.BulkCreateSopJobsHandler(&sopJobsInput)
		if errSopJob != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SOP Jobs: "+errSopJob.Error())
		}

		// ==========================================
		// 2. SEEDING SPK JOBS
		// ==========================================
		spkJobsInput := dto.BulkCreateSpkJobsDto{
			Data: []*dto.CreateSpkJobDto{
				// -- Jobs untuk SPK 1: App Internal (ID: 1) --
				{
					Name:        "Setup Repository & Environment",
					Description: helper.StrPtr("Membuat repo di GitHub dan inisialisasi project Golang & React."),
					SpkID:       1,
					TitleID:     helper.Int64Ptr(7),
					Index:       1,
				},
				{
					Name:        "Pembuatan Schema Database",
					Description: helper.StrPtr("Merancang relasi tabel dan migration SQL & Neo4j."),
					SpkID:       1,
					TitleID:     helper.Int64Ptr(7),
					Index:       2,
				},
				// -- Jobs untuk SPK 2: Laporan Pajak (ID: 2) --
				{
					Name:        "Pengumpulan Faktur Pajak Masukan",
					Description: helper.StrPtr("Merekap seluruh faktur pajak dari vendor selama 1 tahun."),
					SpkID:       2,
					TitleID:     helper.Int64Ptr(8),
					Index:       1,
				},
				{
					Name:        "Rekonsiliasi Bank vs Buku Besar",
					Description: helper.StrPtr("Mencocokkan saldo bank dengan pencatatan akuntansi internal."),
					SpkID:       2,
					TitleID:     helper.Int64Ptr(6),
					Index:       2,
				},
			},
		}

		// Panggil Handler Bulk Create SPK Jobs
		spkJobsRes, errSpkJob := cn.SpkJobHandler.BulkCreateSpkJobsHandler(&spkJobsInput)
		if errSpkJob != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SPK Jobs: "+errSpkJob.Error())
		}

		// ==========================================
		// 3. KEMBALIKAN RESPONSE SUKSES
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Seeding Data Jobs (SOP Jobs & SPK Jobs) Berhasil!",
			"data": fiber.Map{
				"sop_jobs_inserted": len(sopJobsInput.Data),
				"spk_jobs_inserted": len(spkJobsInput.Data),
				"raw_sop_jobs_res":  sopJobsRes,
				"raw_spk_jobs_res":  spkJobsRes,
			},
		})
	}
}

func RunJobDataSeeder(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		
		// Enum type yang diizinkan database-mu
		jobTypes := []string{"instruction", "sop", "spk"}

		// ==========================================
		// 1. GENERATE SOP JOBS (LINKED-LIST GRAPH)
		// ==========================================
		totalSOPs := 50
		sopJobsInserted := 0

		for i := 1; i <= totalSOPs; i++ {
			// Setiap SOP punya 3-7 langkah kerja
			stepsCount := r.Intn(5) + 3 
			var lastInsertedID *int64 = nil

			for j := 1; j <= stepsCount; j++ {
				// Pilih tipe job acak
				jType := jobTypes[r.Intn(len(jobTypes))]
				
				jobDto := &dto.CreateSopJobDto{
					Name:        fmt.Sprintf("Langkah %d untuk SOP ID %d", j, i),
					Alias:       fmt.Sprintf("step-%d-sop-%d", j, i),
					Description: helper.StrPtr(fmt.Sprintf("Ini adalah instruksi detail untuk eksekusi langkah ke-%d", j)),
					Type:        helper.StrPtr(jType), 
					TitleID:     helper.Int64Ptr(int64(r.Intn(100) + 1)), // Acak Jabatan 1-100
					SopID:       int64(i),
					ReferenceID: lastInsertedID, // MENGHUBUNGKAN KE JOB SEBELUMNYA (Linked-List)
					IsPublished: helper.BoolPtr(true),
				}

				// Insert SATU per SATU untuk mendapatkan ID yang akan dipakai node selanjutnya
				input := dto.BulkCreateSopJobs{Data: []*dto.CreateSopJobDto{jobDto}}
				
				res, err := cn.SopJobHandler.BulkCreateSopJobsHandler(&input)
				if err != nil {
					return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, fmt.Sprintf("Gagal seeding SOP Job di SOP %d: %v", i, err))
				}

				// Ambil ID dari record yang baru terbuat dan jadikan lastInsertedID
				if len(res) > 0 {
					newID := int64(res[0].ID)
					lastInsertedID = &newID
					sopJobsInserted++
				}
			}
		}

		// ==========================================
		// 2. GENERATE SPK JOBS
		// ==========================================
		// totalSPKs := 150
		// var spksJobsData []*dto.CreateSpkJobDto

		// for i := 1; i <= totalSPKs; i++ {
		// 	// Setiap SPK punya 2-6 pekerjaan
		// 	jobsCount := r.Intn(5) + 2

		// 	for index := 1; index <= jobsCount; index++ {
		// 		spksJobsData = append(spksJobsData, &dto.CreateSpkJobDto{
		// 			Name:        fmt.Sprintf("Tugas SPK %d - %d", i, index),
		// 			Description: helper.StrPtr(fmt.Sprintf("Pekerjaan indeks %d yang harus diselesaikan untuk SPK %d", index, i)),
		// 			SpkID:       int64(i),
		// 			SopID:       helper.Int64Ptr(int64(r.Intn(50) + 1)), // Relasi acak ke SOP 1-50
		// 			TitleID:     helper.Int64Ptr(int64(r.Intn(100) + 1)), // Acak Jabatan 1-100
		// 			Index:       index, // Menggunakan urutan Index
		// 		})
		// 	}
		// }

		// // SPK Job tidak pakai ReferenceID, jadi aman kita bulk insert ratusan sekaligus!
		// spkJobsInput := dto.BulkCreateSpkJobsDto{Data: spksJobsData}
		// _, errSpkJob := cn.SpkJobHandler.BulkCreateSpkJobsHandler(&spkJobsInput)
		// if errSpkJob != nil {
		// 	return presenters.SendErrorResponseWithMessage(c, fiber.StatusInternalServerError, "Gagal seeding SPK Jobs: "+errSpkJob.Error())
		// }

		// ==========================================
		// 3. RETURN RESPONSE
		// ==========================================
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Seeding Data Jobs (Linked-List & Indexed) Berhasil!",
			"data": fiber.Map{
				"sop_jobs_inserted": sopJobsInserted,
				//"spk_jobs_inserted": len(spkJobsInput.Data),
				// Kita tidak return raw_res karena datanya terlalu masif (bisa bikin crash response)
			},
		})
	}
}