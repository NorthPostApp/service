package dto

import "north-post/service/internal/infra"

type SystemInfoDTO struct {
	Health                     bool    `json:"health"`
	SystemCPUActivePercentage  float32 `json:"systemCpuActivePercentage"`
	SystemDiskTotalBytes       int64   `json:"systemDiskTotalBytes"`
	SystemDiskUsedBytes        int64   `json:"systemDiskUsedBytes"`
	SystemMemoryTotalBytes     int64   `json:"systemMemoryTotalBytes"`
	SystemMemoryUsedBytes      int64   `json:"systemMemoryUsedBytes"`
	SystemNetworkSentBytes     int64   `json:"systemNetworkSentBytes"`
	SystemNetworkReceivedBytes int64   `json:"systemNetworkReceivedBytes"`
}

type GetSystemInfoResponse struct {
	Data SystemInfoDTO `json:"data"`
}

func ToSystemInfoDTO(info *infra.TypesenseSystemInfo) SystemInfoDTO {
	return SystemInfoDTO{
		Health:                     info.Health,
		SystemCPUActivePercentage:  info.SystemCPUActivePercentage,
		SystemDiskTotalBytes:       info.SystemDiskTotalBytes,
		SystemDiskUsedBytes:        info.SystemDiskUsedBytes,
		SystemMemoryTotalBytes:     info.SystemMemoryTotalBytes,
		SystemMemoryUsedBytes:      info.SystemMemoryUsedBytes,
		SystemNetworkSentBytes:     info.SystemNetworkSentBytes,
		SystemNetworkReceivedBytes: info.SystemNetworkReceivedBytes,
	}
}
