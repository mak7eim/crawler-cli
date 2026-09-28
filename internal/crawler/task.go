package crawler

type Task struct {
	URL       string
	Depth     int
	ParentURL string
}

type ResultTask struct {
	Task   Task
	Status int
	Title  string
	Links  []string
	Err    error
}
