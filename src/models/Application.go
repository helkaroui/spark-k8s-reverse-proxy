package models

type Stats struct {
	Application StatsCount
}

type StatsCount struct {
	Running   int
	Completed int
}

type Application struct {
	Id          string
	Name        string
	Driver      string
	Status      string
	StartTime   string
	EndTime     string
	Duration    string
	Labels      map[string]string
	Annotations map[string]string
}

type HomePageData struct {
	Applications []Application
	Stats        Stats
}

type StartTimeSorter []Application

func (a StartTimeSorter) Len() int           { return len(a) }
func (a StartTimeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StartTimeSorter) Less(i, j int) bool { return a[i].StartTime > a[j].StartTime }
