package handler

type createTodoRequest struct {
	Text      string `json:"text"`
	DueDate   string `json:"due_date"`
	Completed *bool  `json:"completed"`
}

type replaceTodoRequest struct {
	Text      string `json:"text"`
	DueDate   string `json:"due_date"`
	Completed bool   `json:"completed"`
}

type patchTodoRequest struct {
	Text      *string `json:"text"`
	DueDate   *string `json:"due_date"`
	Completed *bool   `json:"completed"`
}

type todoResponse struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	DueDate   string `json:"due_date"`
	Completed bool   `json:"completed"`
}

type errorResponse struct {
	Error string `json:"error"`
}
