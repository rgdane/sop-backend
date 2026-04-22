package controllers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"

	"github.com/gofiber/fiber/v2"
)

func RunMasterDataSeeder(cn *container.AppContainer) fiber.Handler {
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

func RunParentDataSeeder(cn *container.AppContainer) fiber.Handler {
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

func RunJobDataSeeder(cn *container.AppContainer) fiber.Handler {
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