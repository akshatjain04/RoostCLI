package eaas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZB-io/internal/roostcli/pkg/spinner"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type GetLogsReq struct {
	TriggerID       string `json:"trigger_id"`
}

type TriggerEAASObj struct {
	Branch     string `json:"branch"`
	OwnerName  string `json:"owner_name"`
	RepoName   string `json:"repo_name"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	UserName   string `json:"user_name"`
	WorkflowID string `json:"workflow_id"`
}

type DeleteAppObj struct {
	AppID                     string `json:"app_id"`
	DeleteAssociatedWorkFlows bool   `json:"delete_associated_workflows"`
	GitTokenID                string `json:"git_token_id"`
}

type ListAppsObj struct {
	AppID      string `json:"app_id"`
	GetForAll  bool   `json:"get_for_all"`
	SearchTerm *string `json:"searchTerm"`
	Skip       int    `json:"skip"`
	SortBy     string `json:"sortBy"`
	Take       *int    `json:"take"`
}

type GetLogsRes struct {
	ClusterLogs        string `json:"cluster_logs"`
	DeployLogs         string `json:"deploy_logs"`
	BuildLogs          string `json:"build_logs"`
	UninstallLogs      string `json:"uninstall_logs"`
	TerraformLogs      string `json:"terraform_logs"`
	CloudFormationLogs string `json:"cloud_formation_logs"`
	TerraformState     string `json:"terraform_state"`
	CDKLogs            string `json:"cdk_logs"`
	PulumiLogs         string `json:"pulumi_logs"`
	APIResp
}

type APIResp struct {
	ResponseCode        int32  `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
}

type EaaslistResp struct {
	Data []Eaaslistdata `json:"data"`
	Count int `json:"count"`
}

type EAASAPIResp struct{
	Msg string `json:"msg"`
}

type Eaaslistdata struct{
	ID string `json:"id"`
	Appname   string `json:"app_name"`
	CodeRepo  string `json:"type"`
	AppRepoName string `json:"app_repo_name"`
	AppRepoBranch string `json:"app_repo_branch"`
	CreatedBy string `json:"created_by"`
	CreatedOn string `json:"created_on"`
}

type GetWorkFlowIDReq struct {
	AppID string `json:"app_id"`
	GitTokenID string `json:"git_token_id"`
}

type workflowIDResp struct {
	Data []workFlowID `json:"data"`
}

type workFlowID struct {
	WorkFlowID string `json:"id"`
}

func GetWorkFlowID(req GetWorkFlowIDReq) string {
	apiEndPoint := "/api/application/client/git/workflow/get"
	reqBuff, err := json.Marshal(req)
	cobra.CheckErr(err)

	body := bytes.NewReader(reqBuff)
    authkey:="Bearer "+viper.Get("roost_jwt_token").(string)
	status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
	cobra.CheckErr(err)
    if status != 201 {
        fmt.Println(status)
    }
	var resp workflowIDResp
	err = json.Unmarshal(respbody, &resp)
	cobra.CheckErr(err)

	return resp.Data[0].WorkFlowID
}

func GetEaasList (isSetAll bool) (list EaaslistResp) {
	eaasObj := ListAppsObj{
		AppID: "zbio",
		GetForAll: isSetAll,
		Skip: 0,
		SortBy: "date",
		Take: nil,
	    SearchTerm: nil,
	}

	var getEaasList EaaslistResp
	spinner := spinner.NewSpinner()
	spinner.Start("Fetching EAAS applications list")
	apiEndPoint := "/api/application/client/git/token/get"
	reqBuff, err := json.Marshal(eaasObj)
	cobra.CheckErr(err)

	body := bytes.NewReader(reqBuff)
    authkey:="Bearer "+viper.Get("roost_jwt_token").(string)
	status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
	cobra.CheckErr(err)
    if status != 201 {
		spinner.Stop(false)
        fmt.Println(status)
    }

	err = json.Unmarshal(resp, &getEaasList)
	cobra.CheckErr(err)
	spinner.Stop(true)
	return getEaasList
}

type ListEnvReq struct {
	AppID string `json:"app_id"`
	EventFilter []string `json:"event_filter"`
	GitTokenID *string `json:"git_token_id"`
	SearchTerm *string `json:"searchTerm"`
	Skip int `json:"skip"`
	SortBy string `json:"sortBy"`
	StatusFilter []string `json:"status_filter"`
	Take int `json:"take"`
	TimeFilter *string `json:"time_filter"`
}

type EnvDetails struct {
	TriggerID string `json:"trigger_id"`
	AssignedCluster int `json:"assigned_cluster_id"`
	AssignedNS string `json:"assigned_namespace"`
	Status string `json:"current_status"`
	StatusDetails string `json:"status_details"`
	StatusUpdated string `json:"status_updated_on"`
	ApplicationEndPoints string `json:"application_end_points"`
	StopStatus int `json:"stop_status"`
	AutoExpiry int `json:"auto_expiry_hrs"`
	TokenType string `json:"token_type"`
	AppName string `json:"app_name"`
	RepoName string `json:"repo_name"`
	BranchName string `json:"branch_name"`
	Action string `json:"action"`
	Date string `json:"string"`
	UserName string `json:"user_name"`
}


type ListEnvResp struct {
	Data []EnvDetails `json:"data"`
	Count int `json:"count"`
}

func GetEaasListEnv(take int) ListEnvResp{
	eaasObj := ListEnvReq{
		AppID: "zbio",
		EventFilter: []string{"pr-open", "pr-reopen", "pr-merge", "on-demand", "push", "manual-trigger", "gh-actions", "circle-ci", "release-publish"},
		GitTokenID: nil,
		SearchTerm: nil,
		Skip: 0,
		SortBy: "date",
		StatusFilter: []string{"In-Queue", "In-Progress", "Skipped", "Failed", "Aborted", "Timed-Out", "Completed", "Stopped"},
		Take: take,
		TimeFilter: nil,
	}
	spinner := spinner.NewSpinner()
	spinner.Start("Fetching environments")
	apiEndPoint := "/api/application/client/git/eaas/get"

	reqBuff, err := json.Marshal(eaasObj)
	cobra.CheckErr(err)

	body := bytes.NewReader(reqBuff)
    authkey:="Bearer "+viper.Get("roost_jwt_token").(string)
	status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
	cobra.CheckErr(err)
    if status != 201 {
		spinner.Stop(false)
        fmt.Println(status)
    }
	var getEaasList ListEnvResp

	err = json.Unmarshal(resp, &getEaasList)
	cobra.CheckErr(err)
	spinner.Stop(true)
	return getEaasList
}