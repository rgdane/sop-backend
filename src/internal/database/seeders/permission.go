package seeders

import (
	"time"

	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedPermissions(db *gorm.DB) error {
	permissionsMap := map[string][]string{
		"departments":         {"create", "update", "delete", "view", "viewOwn"},
		"divisions":           {"create", "update", "delete", "view", "viewOwn"},
		"holidays":            {"create", "update", "delete", "view", "viewOwn"},
		"leaves":              {"create", "update", "delete", "view", "viewOwn"},
		"levels":              {"create", "update", "delete", "view", "viewOwn"},
		"positions":           {"create", "update", "delete", "view", "viewOwn"},
		"projects":            {"create", "update", "delete", "view", "viewOwn"},
		"products":            {"create", "update", "delete", "view", "viewOwn"},
		"epics":               {"create", "update", "delete", "view", "viewOwn"},
		"features":            {"create", "update", "delete", "view", "viewOwn"},
		"languages":           {"create", "update", "delete", "view", "viewOwn"},
		"roles":               {"create", "update", "delete", "view", "viewOwn"},
		"permissions":         {"create", "update", "delete", "view", "viewOwn"},
		"squads":              {"create", "update", "delete", "view", "viewOwn"},
		"statuses":            {"create", "update", "delete", "view", "viewOwn"},
		"titles":              {"create", "update", "delete", "view", "viewOwn"},
		"users":               {"create", "update", "delete", "view", "viewOwn"},
		"backlogs":            {"create", "update", "delete", "view", "viewOwn"},
		"backlog_items":       {"create", "update", "delete", "view", "viewOwn"},
		"sprints":             {"create", "update", "delete", "view", "viewOwn"},
		"comments":            {"create", "update", "delete", "view", "viewOwn"},
		"cms_articles":        {"create", "update", "delete", "view", "viewOwn"},
		"cms_categories":      {"create", "update", "delete", "view", "viewOwn"},
		"cms_tags":            {"create", "update", "delete", "view", "viewOwn"},
		"documents":           {"create", "update", "delete", "view", "viewOwn"},
		"sops":                {"create", "update", "delete", "view", "viewOwn"},
		"spks":                {"create", "update", "delete", "view", "viewOwn"},
		"spk_jobs":            {"create", "update", "delete", "view", "viewOwn"},
		"spk_txs":             {"create", "update", "delete", "view", "viewOwn"},
		"sop_jobs":            {"create", "update", "delete", "view", "viewOwn"},
		"todos":               {"view", "viewOwn"},
		"sprint_lead_dailies": {"view", "viewOwn"},
		"color_palettes":      {"create", "update", "delete", "view"},
		"typographies":        {"create", "update", "delete", "view"},
		"sop_menus":           {"create", "update", "delete", "view"},
		"case_reports":        {"create", "update", "delete", "view", "viewOwn"},
		"case_categories":     {"create", "update", "delete", "view", "viewOwn"},
		"case_bug_features":   {"create", "update", "delete", "view", "viewOwn"},
		"case_statuses":       {"create", "update", "delete", "view", "viewOwn"},
		"design_systems":      {"create", "update", "delete", "view", "viewOwn"},
		"omnichannel":         {"create", "update", "delete", "view"},
	}

	for module, actions := range permissionsMap {
		for _, action := range actions {
			permName := module + "." + action

			var count int64
			db.Model(&models.Permission{}).Where("name = ?", permName).Count(&count)
			if count == 0 {
				db.Create(&models.Permission{
					Name:      permName,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				})
			}
		}
	}

	return nil
}
