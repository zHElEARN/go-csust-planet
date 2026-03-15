package dto

type ElectricityTaskOption struct {
	NotifyTime string `json:"notifyTime" binding:"required"`
	Campus     string `json:"campus" binding:"required"`
	Building   string `json:"building" binding:"required"`
	Room       string `json:"room" binding:"required"`
}

type SyncElectricityTaskRequest struct {
	DeviceToken string                  `json:"deviceToken" binding:"required"`
	Tasks       []ElectricityTaskOption `json:"tasks"`
}
