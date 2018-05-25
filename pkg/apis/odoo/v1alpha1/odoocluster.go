package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type OdooClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []OdooCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type OdooCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              OdooClusterSpec   `json:"spec"`
	Status            OdooClusterStatus `json:"status,omitempty"`
}

type OdooClusterSpec struct {
	Images        []OdooClusterSpecImage  `json:images`
	PqSpec        PgNamespace             `json:pgSpec`
	ResourceSpec  OdooClusterResourceSpec `json:resourceSpec`
	AdminPassword string                  `json:"adminPassword"`
	ConfigMap     string                  `json:"configMap"`
	DeployModel   OdooClusterMode         `json:deployModel`
	NodeSelector  string                  `json:"nodeSelector"`
	// Replicas         int                      `json:"replicas"`

	// MailServer  bool `json:"mailServer"`
	// OnlyOffice  bool `json:"onlyOffice"`
	// Mattermost  bool `json:"mattermost"`
	// Nuxeo       bool `json:"nuxeo"`
	// BpmnEngine  bool `json:"bpmnEngine"`
	// OpenProject bool `json:"openProject"`
}

type OdooClusterSpecImage struct {
	Registry string `json:"registry"`
	Name     string `json:"image"`
	Tag      string `json:"tag"`
}

type PgNamespace struct {
	User      string              `json:"user"`
	Password  string              `json:"password"`
	UserQuota int                 `json:"userQuota"`
	PgCluster PgClusterConnection `json:"pgCluster"`
}

type PgClusterConnection struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type OdooClusterResourceSpec struct {
	Cpu  int `json:"cpu"`
	Ram  int `json:"ram"`
	Disk int `json:"disk"`
}

// OdooClusterMode ...
type OdooClusterMode string

const (
	// OdooClusterModeRemote ...
	OdooClusterModeRemote OdooClusterMode = "remote"
	// OdooClusterModeLocal ...
	OdooClusterModeLocal OdooClusterMode = "local"
	// OdooClusterModeHybrid ...
	OdooClusterModeHybrid OdooClusterMode = "hybrid"
)

type OdooClusterStatus struct {
	DbUserQuotaUsage string           `json:"dbUserQuotaUsage,omitempty"`
	DiskUsage        string           `json:"diskUsage,omitempty"`
	State            OdooClusterState `json:"state,omitempty"`
	Message          string           `json:"message,omitempty"`
	CurrentImage     string           `json:"currentImage,omitempty"`
	ImageLoadShare   []ImageLoad      `json:"imageLoadShare,omitempty"`
	// Replicas     int               `json:"replicas,omitempty"`
}

type ImageLoad struct {
	Name      string `json:"name"`
	Instances int    `json:"instances"`
}

// OdooClusterState ...
type OdooClusterState string

const (
	// OdooClusterStateCreated ...
	OdooClusterStateCreated OdooClusterState = "Created"
	// OdooClusterStateReconciled ...
	OdooClusterStateReconciled OdooClusterState = "Reconciled"
)
