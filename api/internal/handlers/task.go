package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sol-armada/sol-bot/database"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/projects"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

func (h *Handler) listTaskActivity(ctx context.Context, taskID string) ([]taskActivityResponse, error) {
	db := database.Get()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	history, err := db.Queries.ListTaskHistory(ctx, taskID)
	if err != nil {
		return nil, err
	}

	activity := make([]taskActivityResponse, 0, len(history))
	for _, entry := range history {
		if len(entry.Details) == 0 {
			continue
		}

		var details map[string]any
		if err := json.Unmarshal(entry.Details, &details); err != nil {
			h.logger.Warn("Failed to parse task history details", "taskId", taskID, "historyId", entry.ID, "error", err)
			continue
		}

		summary, _ := details["summary"].(string)
		summary = strings.TrimSpace(summary)
		if summary == "" {
			continue
		}

		item := taskActivityResponse{
			ID:      strconv.FormatInt(entry.ID, 10),
			Summary: summary,
		}
		if entry.PerformedAt.Valid {
			item.Time = entry.PerformedAt.Time.UTC().Format(time.RFC3339)
		}

		activity = append(activity, item)
	}
	if len(activity) > 5 {
		activity = activity[:5]
	}
	return activity, nil
}

func (h *Handler) taskWithActivity(ctx context.Context, task *projects.Task) taskWithActivityResponse {
	response := taskWithActivityResponse{Task: task, Activity: []taskActivityResponse{}}
	if task == nil {
		return response
	}

	activity, err := h.listTaskActivity(ctx, task.Id)
	if err != nil {
		h.logger.Warn("Failed to list task activity", "taskId", task.Id, "error", err)
		return response
	}

	response.Activity = activity
	return response
}

func (h *Handler) ListTaskStatuses(c echo.Context) error {
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

	statuses, err := service.ListTaskStatuses(c.Request().Context(), projectUUID)
	if err != nil {
		h.logger.Error("Failed to list project task statuses", "projectId", projectId, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_statuses_list_failed",
			Message: "Failed to fetch project task statuses",
		})
	}

	return c.JSON(http.StatusOK, statuses)
}

func (h *Handler) CreateTask(c echo.Context) error {
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

	var req service.CreateProjectTaskInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	userId, _ := c.Get("user_id").(string)
	task, err := service.CreateProjectTask(c.Request().Context(), projectUUID, userId, req)
	if err != nil {
		h.logger.Error("Failed to create project task", "projectId", projectId, "error", err)
		if strings.Contains(strings.ToLower(err.Error()), "title is required") || strings.Contains(strings.ToLower(err.Error()), "invalid due date") {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_create_failed",
			Message: "Failed to create project task",
		})
	}

	return c.JSON(http.StatusCreated, h.taskWithActivity(c.Request().Context(), task))
}

func (h *Handler) UpdateTask(c echo.Context) error {
	userId, _ := c.Get("user_id").(string)
	member, err := members.Get(userId)
	if err != nil {
		if errors.Is(err, members.MemberNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "member_not_found",
				Message: "Member not found",
			})
		}
		h.logger.Error("Getting member details", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_fetch_failed",
			Message: "Failed to fetch member details",
		})
	}

	var projectId uuid.UUID
	if projectIdStr := strings.TrimSpace(c.Param("id")); projectIdStr != "" {
		var err error
		projectId, err = uuid.Parse(projectIdStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid project Id",
			})
		}
	}

	taskId := strings.TrimSpace(c.Param("taskId"))

	if projectId == uuid.Nil || taskId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Project Id and task Id are required",
		})
	}

	var req service.UpdateProjectTaskInput
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	task, err := projects.GetTask(c.Request().Context(), taskId)
	if err != nil {
		if errors.Is(err, projects.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Project task not found",
			})
		}
		h.logger.Error("Getting task for update", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_fetch_failed",
			Message: "Failed to fetch project task",
		})
	}
	beforeTask := *task // Create a copy of the task before updating

	var assignee *members.Member
	if req.Assignee != "" {
		assignee, err = members.Get(req.Assignee)
		if err != nil {
			if errors.Is(err, members.MemberNotFound) {
				return c.JSON(http.StatusNotFound, dto.ErrorResponse{
					Error:   "assignee_not_found",
					Message: "Assignee member not found",
				})
			}
			h.logger.Error("Getting assignee for task update", "error", err.Error())
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "assignee_fetch_failed",
				Message: "Failed to fetch assignee member",
			})
		}
	}

	var dueAt *time.Time
	if req.DueAt != nil && *req.DueAt != "" {
		parsed, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_date",
				Message: "Invalid due date format (expected RFC3339)",
			})
		}
		dueAt = &parsed
	}

	taskStatus, err := projects.GetTaskStatus(c.Request().Context(), projectId, string(req.Status))
	if err != nil {
		h.logger.Error("Getting task status", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "task_status_fetch_failed",
			Message: "Failed to fetch task status",
		})
	}

	var parentTaskId *string
	if req.ParentTaskId != nil && *req.ParentTaskId != "" {
		parentTaskId = req.ParentTaskId
	}

	var parentTask *projects.Task
	if parentTaskId != nil && *parentTaskId != "" {
		parentTask, err = projects.GetTask(c.Request().Context(), *parentTaskId)
		if err != nil {
			if errors.Is(err, projects.ErrTaskNotFound) {
				return c.JSON(http.StatusNotFound, dto.ErrorResponse{
					Error:   "parent_task_not_found",
					Message: "Parent task not found",
				})
			}
			h.logger.Error("Getting parent task for task update", "error", err.Error())
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "parent_task_fetch_failed",
				Message: "Failed to fetch parent task",
			})
		}
	}

	task.Title = req.Title
	task.Description = req.Description
	task.Priority = projects.TaskPriority(req.Priority)
	task.Assignee = assignee
	task.DueAt = dueAt
	task.Status = taskStatus
	task.ParentTask = parentTask

	if err := projects.UpdateTask(c.Request().Context(), &beforeTask, task, member); err != nil {
		h.logger.Error("Updating task", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_update_failed",
			Message: "Failed to update project task",
		})
	}

	return c.JSON(http.StatusOK, h.taskWithActivity(c.Request().Context(), task))
}

func (h *Handler) DeleteTask(c echo.Context) error {
	var projectId uuid.UUID
	var taskId string
	if projectIdStr := strings.TrimSpace(c.Param("id")); projectIdStr != "" {
		var err error
		projectId, err = uuid.Parse(projectIdStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid project Id",
			})
		}
	}
	if taskIdStr := strings.TrimSpace(c.Param("taskId")); taskIdStr != "" {
		taskId = taskIdStr
	}
	if taskId == "" {
		if taskIdStr := strings.TrimSpace(c.Param("ticketId")); taskIdStr != "" {
			taskId = taskIdStr
		}
	}

	if projectId == uuid.Nil || taskId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Project Id and task Id are required",
		})
	}

	userId, _ := c.Get("user_id").(string)
	task, err := projects.GetTask(c.Request().Context(), taskId)
	if err != nil {
		if errors.Is(err, projects.ErrTaskNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Project task not found",
			})
		}
		h.logger.Error("Getting task for delete", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_fetch_failed",
			Message: "Failed to fetch project task",
		})
	}

	member, err := members.Get(userId)
	if err != nil {
		if errors.Is(err, members.MemberNotFound) {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "member_not_found",
				Message: "Member not found",
			})
		}
		h.logger.Error("Getting member details", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_fetch_failed",
			Message: "Failed to fetch member details",
		})
	}

	if err := task.Delete(c.Request().Context(), member); err != nil {
		h.logger.Error("Deleting project task", "error", err.Error())
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "project_task_delete_failed",
			Message: "Failed to delete project task",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{"deleted": true})
}
