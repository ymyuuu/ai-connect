package domain

import (
	"encoding/json"
	"github.com/dhbin/ai-connect/templates"
	"github.com/spf13/cobra"
	"io"
)

type Me struct {
	Amr                      []interface{} `json:"amr"`
	Created                  int           `json:"created"`
	Email                    string        `json:"email"`
	Groups                   []interface{} `json:"groups"`
	HasPaygProjectSpendLimit bool          `json:"has_payg_project_spend_limit"`
	Id                       string        `json:"id"`
	MfaFlagEnabled           bool          `json:"mfa_flag_enabled"`
	Name                     string        `json:"name"`
	Object                   string        `json:"object"`
	Orgs                     struct {
		Data []struct {
			Created                        int           `json:"created"`
			Description                    string        `json:"description"`
			Geography                      interface{}   `json:"geography"`
			Groups                         []interface{} `json:"groups"`
			Id                             string        `json:"id"`
			IsDefault                      bool          `json:"is_default"`
			IsScaleTierAuthorizedPurchaser interface{}   `json:"is_scale_tier_authorized_purchaser"`
			IsScimManaged                  bool          `json:"is_scim_managed"`
			Name                           string        `json:"name"`
			Object                         string        `json:"object"`
			ParentOrgId                    interface{}   `json:"parent_org_id"`
			Personal                       bool          `json:"personal"`
			Projects                       struct {
				Data   []interface{} `json:"data"`
				Object string        `json:"object"`
			} `json:"projects"`
			Role     string `json:"role"`
			Settings struct {
				DisableUserApiKeys       bool   `json:"disable_user_api_keys"`
				ThreadsUiVisibility      string `json:"threads_ui_visibility"`
				UsageDashboardVisibility string `json:"usage_dashboard_visibility"`
			} `json:"settings"`
			Title string `json:"title"`
		} `json:"data"`
		Object string `json:"object"`
	} `json:"orgs"`
	PhoneNumber interface{} `json:"phone_number"`
	Picture     string      `json:"picture"`
}

type CheckGpts struct {
	Kind     string `json:"kind"`
	Referrer string `json:"referrer"`
}

var GptsInfoInject map[string]interface{}

func init() {
	f, err := templates.TemplateFs.Open("chatgpt/gpts_info_inject.json")
	cobra.CheckErr(err)
	bs, err := io.ReadAll(f)
	GptsInfoInject = make(map[string]interface{})
	err = json.Unmarshal(bs, &GptsInfoInject)
	cobra.CheckErr(err)
}
