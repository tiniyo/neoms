package models

type RecordJob struct {
	Args struct {
		FileName string `json:"file_name"`
		FilePath string `json:"file_path"`
		JobID    int64  `json:"job_id"`
	} `json:"args"`
	ID   string `json:"id"`
	Name string `json:"name"`
	T    int64  `json:"t"`
}

