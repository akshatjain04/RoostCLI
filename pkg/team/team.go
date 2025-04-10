package team

type TeamListResponse struct {
	Teamlist []TeamList `json:"teams"`
	Count    int        `json:"count"`
}

type TeamList struct {
	TeamId         string `json:"team_id"`
	MemberId       string `json:"member_id"`
	MemberType     string `json:"member_type"`
	MemberRole     string `json:"member_role"`
	JoiningDate    string `json:"joining_date"`
	Isadmin        int    `json:"is_admin"`
	MakeAdminOn    string `json:"made_admin_on"`
	NamespaceRegex string `json:"namespace_regex"`
	NamespaceRole  string `json:"namespace_role"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Visibility     string `json:"visibility"`
	Organistation  string `json:"org"`
	MemberCount    string `json:"member_count"`
}

type CreateTeam struct {
	Description  string   `json:"description"`
	FirstMembers []string `json:"firstMembers"`
	Name         string   `json:"name"`
	Org          string   `json:"org"`
	Visibility   string   `json:"visibility"`
}

type DeleteTeam struct {
	TeamID string `json:"team_id"`
}

type InviteMembers struct {
	TeamID   string   `json:"team_id"`
	Username []string `json:"username"`
}

type RemoveMember struct {
	TeamID   string `json:"team_id"`
	MemberID string `json:"member_id"`
}

type TeamApiResponse struct {
	TeamRespMessage string `json:"message"`
}

type ClusterAdd struct {
	TeamId         string `json:"team_id"`
	ClusterId      int    `json:"cluster_id"`
	CustomerEmail  string `json:"customer_email"`
	CustomerToken  string `json:"customer_token"`
	RbacScope      string `json:"rbac_scope"`
	AwsCredentials `json:"aws_credentials"`
}

type Credentialfile struct {
	Filename    string `json:"file_name"`
	Filecontent string `json:"file_content"`
}

type AwsCredentials struct {
	CredentialInputType string `json:"credentials_input_type"`
	AccesskeyID         string `json:"access_key_id"`
	Credentialfile      `json:"credentials_file"`
	SecretAccesskey     string `json:"secret_access_key"`
	SessionToken        string `json:"session_token"`
}

type UpdateClusterInfo struct {
	TeamId     string `json:"team_id"`
	Teamconfig `json:"config"`
}

type TeamKubeConfigResponse struct {
	ClusterID  int    `json:"cluster_id"`
	Kubeconfig string `json:"kubeconfig"`
	PublicIP   string `json:"public_ip"`
}

type TeamKubeConfigObj struct {
	TeamId      string   `json:"team_id"`
	Clustertype []string `json:"cluster_type"`
}

type Teamconfig struct {
	Helmrepopwd        string `json:"helm_repo_pwd"`
	Helmrepousername   string `json:"helm_repo_username"`
	Restrictuseraccess bool   `json:"restrictuseraccess"`
	TeamClusterId      int    `json:"team_cluster_id"`
}
