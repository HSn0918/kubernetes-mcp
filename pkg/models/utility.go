package models

// SearchResult 搜索结果数据
type SearchResult struct {
	Kind         string `json:"kind"`
	APIVersion   string `json:"apiVersion"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Labels       string `json:"labels,omitempty"`
	Annotations  string `json:"annotations,omitempty"`
	MatchedBy    string `json:"matchedBy,omitempty"`
	MatchedValue string `json:"matchedValue,omitempty"`
	CreationTime string `json:"creationTime,omitempty"`
}

// SearchResults 搜索结果列表
type SearchResults struct {
	Items       []SearchResult `json:"items"`
	TotalCount  int            `json:"totalCount"`
	SearchQuery string         `json:"searchQuery"`
	TypesCount  int            `json:"typesCount"`
}

// EventInfo 事件信息
type EventInfo struct {
	LastSeen    string `json:"lastSeen"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Object      string `json:"object"`
	Message     string `json:"message"`
	FullMessage string `json:"fullMessage,omitempty"`
}

// EventsResult 事件查询结果
type EventsResult struct {
	Items       []EventInfo `json:"items"`
	ResourceRef struct {
		Kind      string `json:"kind"`
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"resourceRef"`
	Count int `json:"count"`
}

// DiffResult 差异比较结果
type DiffResult struct {
	Kind         string       `json:"kind"`
	Name         string       `json:"name"`
	Namespace    string       `json:"namespace"`
	ApiVersion   string       `json:"apiVersion"`
	Exists       bool         `json:"exists"`
	DiffCount    int          `json:"diffCount"`
	DiffDetails  []DiffDetail `json:"diffDetails,omitempty"`
	IsNewResurce bool         `json:"isNewResource"`
}

// DiffDetail 差异详情
type DiffDetail struct {
	Field    string `json:"field"`
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue,omitempty"`
	Action   string `json:"action"` // "add", "remove", "change"
}

// ValidationResult 验证结果
type ValidationResult struct {
	Valid     bool   `json:"valid"`
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Document  int    `json:"document,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ValidationResults 验证结果列表
type ValidationResults struct {
	Items      []ValidationResult `json:"items"`
	ValidCount int                `json:"validCount"`
	ErrorCount int                `json:"errorCount"`
	TotalCount int                `json:"totalCount"`
}

// ClusterInfo 集群信息
type ClusterInfo struct {
	Version      string `json:"version"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Platform     string `json:"platform"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	Compiler     string `json:"compiler"`
	Namespace    string `json:"namespace,omitempty"`
}

// ApplyResult 应用清单的结果
type ApplyResult struct {
	Kind          string `json:"kind"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace,omitempty"`
	ApiVersion    string `json:"apiVersion"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Document      int    `json:"document"`
	ClusterScoped bool   `json:"clusterScoped"`
}

// ApplyResults 应用清单结果列表
type ApplyResults struct {
	Items        []ApplyResult `json:"items"`
	SuccessCount int           `json:"successCount"`
	ErrorCount   int           `json:"errorCount"`
	DryRun       bool          `json:"dryRun"`
}

// ResourceDef API资源定义
type ResourceDef struct {
	Kind         string   `json:"kind"`
	GroupVersion string   `json:"groupVersion"`
	Name         string   `json:"name"`
	Namespaced   bool     `json:"namespaced"`
	Verbs        []string `json:"verbs"`
	ShortNames   []string `json:"shortNames,omitempty"`
}

// APIResourceGroup API资源组
type APIResourceGroup struct {
	GroupVersion string        `json:"groupVersion"`
	Resources    []ResourceDef `json:"resources"`
}

// APIResourceList API资源列表
type APIResourceList struct {
	Groups []APIResourceGroup `json:"groups"`
}
