package main

type VPSInfo struct {
	NodeDatacenter   string `json:"node_datacenter,omitempty"`
	VeStatus         string `json:"ve_status,omitempty"`
	VeDiskQuotaGb    string `json:"ve_disk_quota_gb,omitempty"`
	VeUsedDiskSpaceB int64  `json:"ve_used_disk_space_b,omitempty"`
	PlanMonthlyData  int64  `json:"plan_monthly_data,omitempty"`
	DataCounter      int64  `json:"data_counter,omitempty"`
	DataNextReset    int64  `json:"data_next_reset,omitempty"`
}
