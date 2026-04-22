package controllers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"

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