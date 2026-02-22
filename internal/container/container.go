package container

type State string

const (
	StateCreated State = "created"
	StateRunning State = "running"
	StateStopped State = "stopped"
)

type ResourceLimits struct {
	MemoryBytes int64   `json:"memory_bytes"`
	CPUQuota    float64 `json:"cpu_quota"`
	PidsMax     int     `json:"pids_max"`
}

type Config struct {
	Image     string
	Command   []string
	Hostname  string
	Resources ResourceLimits
	NetMode   string
	Volumes   []string
}
