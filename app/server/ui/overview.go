package ui

type Overview struct {
	Instances OverviewInstanceArray
}

type OverviewInstance struct {
	Docker    OverviewDocker
	Server    OverviewServer
	Resources OverviewResources
}

type OverviewInstanceArray []OverviewInstance

type OverviewServer struct {
	Name      string
	Host      string
	Role      string
	Agents    []string
	CountCPU  int
	AmountRAM uint64
}

type OverviewDocker struct {
	Version string
	Host    string
}

type OverviewResources struct {
	Containers JSON
	Images     JSON
	Volumes    JSON
	Networks   JSON
}
