package projects

type Project struct {
	Id           int      `json:"id"`
	Name         string   `json:"name"`
	Author       string   `json:"author"`
	Participants []string `json:"participants"`
}

type NewProjectRequest struct {
	Name string `json:"name"`
}
