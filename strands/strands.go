package strands

import (
	"fmt"
	"k8s.io/klog"
)

type Agent struct {
	Name string
}

func NewAgent(name string) *Agent {
	return &Agent{Name: name}
}

func (a *Agent) DetectDrift(data string) string {
	result := fmt.Sprintf("Drift detected in %s by agent %s", data, a.Name)
	klog.Infof("Strands Agent drift detection: %s", result)
	return result
}

func (a *Agent) Summarize(data string) string {
	result := fmt.Sprintf("Summary: %s processed by agent %s", data, a.Name)
	klog.Infof("Strands Agent summarization: %s", result)
	return result
}