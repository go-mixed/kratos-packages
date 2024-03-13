package rbac

import "gopkg.in/go-mixed/kratos-packages.v2/pkg/db"

type PolicyModel struct {
	db.Model

	Ptype string `gorm:"size:10;uniqueIndex:unique_index" json:"ptype,omitempty"`
	V0    string `gorm:"size:50;uniqueIndex:unique_index" json:"v_0,omitempty"`
	V1    string `gorm:"size:100;uniqueIndex:unique_index" json:"v_1,omitempty"`
	V2    string `gorm:"size:100;uniqueIndex:unique_index" json:"v_2,omitempty"`
	V3    string `gorm:"size:100;uniqueIndex:unique_index" json:"v_3,omitempty"`
	V4    string `gorm:"size:100;uniqueIndex:unique_index" json:"v_4,omitempty"`
	V5    string `gorm:"size:100;uniqueIndex:unique_index" json:"v_5,omitempty"`
}

func (PolicyModel) TableName() string {
	return "policies"
}
