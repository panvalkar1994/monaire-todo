package handler

type createTodoRequest struct {
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Completed   *bool  `json:"completed"`
}

type replaceTodoRequest struct {
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Completed   bool   `json:"completed"`
}

type patchTodoRequest struct {
	Description *string `json:"description"`
	DueDate     *string `json:"due_date"`
	Completed   *bool   `json:"completed"`
}

type todoResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Completed   bool   `json:"completed"`
}

type errorResponse struct {
	Error string `json:"error"`
}
