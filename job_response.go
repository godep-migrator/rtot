package rtot

type jobResponse struct {
	Jobs []*jobJSON `json:"jobs"`
}

func newJobResponse(jobs []*job, fields *map[string]int) *jobResponse {
	mapped := []*jobJSON{}
	for _, j := range jobs {
		mapped = append(mapped, j.toJSON(fields))
	}
	return &jobResponse{Jobs: mapped}
}
