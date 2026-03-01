package models

type Flowchart struct {
	ID   int64  `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('flowcharts_seq'::regclass)" json:"id"`
	Type string `gorm:"type:text;not null;index:idx_sop_job_type" json:"type"`

	SopJobs   []*SopJob `gorm:"foreignKey:FlowchartID;references:ID;constraint:OnDelete:CASCADE;" json:"has_sop_job,omitempty"`
	HasSpkJob []*SpkJob `gorm:"foreignKey:FlowchartID;references:ID;constraint:OnDelete:CASCADE;" json:"has_spk_job,omitempty"`
}

func (Flowchart) TableName() string {
	return "flowcharts"
}
