package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sol-armada/sol-bot/database"
	"github.com/sol-armada/sol-bot/projects"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) ListProjects(c echo.Context) error {
	prjcts, err := projects.ListProjects(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to list projects", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "projects_list_failed",
			Message: "Failed to fetch projects list",
		})
	}

	for _, project := range prjcts {
		tasks, err := projects.ListTasks(c.Request().Context(), project.Id)
		if err != nil {
			h.logger.Error("Failed to list tasks for project", "projectId", project.Id.String(), "error", err)
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "project_tasks_list_failed",
				Message: "Failed to fetch tasks for project",
			})
		}
		project.Tasks = tasks
	}

	return c.JSON(http.StatusOK, prjcts)
}

func (h *Handler) CreateProject(c echo.Context) error {
	var req dto.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Parse optional due date
	var dueAt *time.Time
	if req.DueAt != nil && *req.DueAt != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid due date format (expected RFC3339)",
			})
		}
		dueAt = &parsed
	}

	project, err := service.CreateProject(req.Name, req.Description, req.StatusID, req.OwnerID, dueAt)
	if err != nil {
		h.logger.Error("Failed to create project", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_create_failed",
			Message: "Failed to create project",
		})
	}

	return c.JSON(http.StatusCreated, project)
}

func (h *Handler) ListProjectStatuses(c echo.Context) error {
	db := database.Get()
	if db == nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "database_error",
			Message: "Database not initialized",
		})
	}

	ctx := context.Background()
	statuses, err := db.Queries.ListProjectStatuses(ctx)
	if err != nil {
		h.logger.Error("Failed to list project statuses", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "statuses_fetch_failed",
			Message: "Failed to fetch project statuses",
		})
	}

	type StatusResponse struct {
		Id   int32  `json:"id"`
		Name string `json:"name"`
	}

	response := make([]StatusResponse, len(statuses))
	for i, status := range statuses {
		response[i] = StatusResponse{
			Id:   status.ID,
			Name: status.Name,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *Handler) ListTasks(c echo.Context) error {
	projectId := strings.TrimSpace(c.Param("id"))
	if projectId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Project Id is required",
		})
	}

	projectUUID, err := uuid.Parse(projectId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid project Id",
		})
	}

	tasks, err := service.ListProjectTasks(c.Request().Context(), projectUUID)
	if err != nil {
		h.logger.Error("Failed to list project tasks", "projectId", projectId, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_tasks_list_failed",
			Message: "Failed to fetch project tasks",
		})
	}

	response := make([]taskWithActivityResponse, 0, len(tasks))
	for _, task := range tasks {
		response = append(response, h.taskWithActivity(c.Request().Context(), task))
	}

	return c.JSON(http.StatusOK, map[string]any{"tasks": response})
}
