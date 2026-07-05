package handlers

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/sol-armada/sol-bot/database"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/projects"
	"github.com/sol-armada/website/internal/dto"
	"github.com/sol-armada/website/internal/service"
)

// AdminServiceInterface defines the interface for admin operations
type AdminServiceInterface interface {
	GetOverviewStats(context.Context) (*service.AdminOverviewStats, error)
	GetAttendanceRecords(context.Context, int, int, string) ([]service.AttendanceRecord, error)
	GetTokenLedger(context.Context, int, int, string) ([]service.TokenTransaction, error)
	GetTokenLedgerAnalytics(context.Context) (*service.TokenLedgerAnalytics, error)
	GetAttendanceAnalytics(context.Context) (*service.AttendanceAnalytics, error)
	GetMembers(context.Context, int, int, string) ([]service.MemberSummary, error)
	GetMembersByIds(context.Context, []string) (map[string]service.MemberSummary, error)
	CreateAttendanceRecord(context.Context, service.CreateAttendanceRecordInput) error
	GetAttendanceRecord(context.Context, string) (*service.AttendanceRecord, error)
	GetMembersByAttendance(context.Context, string) ([]service.MemberSummary, error)
	GetAttendanceEditPayload(context.Context, string) (*service.AttendanceEditPayload, error)
	UpdateAttendanceRecord(context.Context, string, service.UpdateAttendanceRecordInput) error
}

var _ AdminServiceInterface = (*service.AdminService)(nil)

type AdminHandler struct {
	adminService  AdminServiceInterface
	configService *service.ConfigService
	logger        *slog.Logger
}

type taskActivityResponse struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Time    string `json:"time,omitempty"`
}

type taskWithActivityResponse struct {
	*projects.Task
	Activity []taskActivityResponse `json:"activity"`
}

func NewAdminHandler(adminService AdminServiceInterface, configService *service.ConfigService, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		adminService:  adminService,
		configService: configService,
		logger:        logger,
	}
}

func (h *AdminHandler) listTaskActivity(ctx context.Context, taskID uuid.UUID) ([]taskActivityResponse, error) {
	db := database.Get()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	history, err := db.Queries.ListTaskHistory(ctx, taskID.String())
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
			h.logger.Warn("Failed to parse task history details", "taskId", taskID.String(), "historyId", entry.ID, "error", err)
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

func (h *AdminHandler) taskWithActivity(ctx context.Context, task *projects.Task) taskWithActivityResponse {
	response := taskWithActivityResponse{Task: task, Activity: []taskActivityResponse{}}
	if task == nil {
		return response
	}

	activity, err := h.listTaskActivity(ctx, task.Id)
	if err != nil {
		h.logger.Warn("Failed to list task activity", "taskId", task.Id.String(), "error", err)
		return response
	}

	response.Activity = activity
	return response
}

func (h *AdminHandler) ListProjects(c echo.Context) error {
	projects, err := service.ListProjects(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to list projects", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "projects_list_failed",
			Message: "Failed to fetch projects list",
		})
	}

	return c.JSON(http.StatusOK, projects)
}

func (h *AdminHandler) CreateProject(c echo.Context) error {
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
		parsed, err := time.Parse(time.RFC3339, *req.DueAt)
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

func (h *AdminHandler) ListProjectStatuses(c echo.Context) error {
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

func (h *AdminHandler) ListProjectTasks(c echo.Context) error {
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

func (h *AdminHandler) ListProjectTaskStatuses(c echo.Context) error {
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

func (h *AdminHandler) CreateProjectTask(c echo.Context) error {
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

func (h *AdminHandler) UpdateProjectTask(c echo.Context) error {
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

	var taskId uuid.UUID
	if taskIdStr := strings.TrimSpace(c.Param("taskId")); taskIdStr != "" {
		var err error
		taskId, err = uuid.Parse(taskIdStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid task Id",
			})
		}
	}

	if projectId == uuid.Nil || taskId == uuid.Nil {
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

	var parentTaskId *uuid.UUID
	if req.ParentTaskId != nil && *req.ParentTaskId != "" {
		parsedParentTaskId, err := uuid.Parse(*req.ParentTaskId)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid parent task Id",
			})
		}
		parentTaskId = &parsedParentTaskId
	}

	var parentTask *projects.Task
	if parentTaskId != nil && *parentTaskId != uuid.Nil {
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

func (h *AdminHandler) DeleteProjectTask(c echo.Context) error {
	var projectId, taskId uuid.UUID
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
		var err error
		taskId, err = uuid.Parse(taskIdStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: "Invalid task Id",
			})
		}
	}
	if taskId == uuid.Nil {
		if taskIdStr := strings.TrimSpace(c.Param("ticketId")); taskIdStr != "" {
			var err error
			taskId, err = uuid.Parse(taskIdStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
					Error:   "invalid_request",
					Message: "Invalid task Id",
				})
			}
		}
	}

	if projectId == uuid.Nil || taskId == uuid.Nil {
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

func (h *AdminHandler) GetOverview(c echo.Context) error {
	result, err := h.adminService.GetOverviewStats(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch admin overview", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "admin_overview_failed",
			Message: "Failed to fetch overview statistics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) GetAttendance(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := c.QueryParam("search")

	result, err := h.adminService.GetAttendanceRecords(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch attendance records", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance records",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"records": result,
		"page":    page,
		"limit":   limit,
	})
}

func (h *AdminHandler) GetTokenLedger(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := c.QueryParam("search")

	result, err := h.adminService.GetTokenLedger(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch token ledger", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_ledger_failed",
			Message: "Failed to fetch token ledger",
		})
	}

	// collect ids to turn into names
	userIds := make(map[string]struct{})
	for _, tx := range result {
		userIds[tx.MemberID] = struct{}{}
	}

	memberIds := make([]string, 0, len(userIds))
	for id := range userIds {
		memberIds = append(memberIds, id)
	}

	members, err := h.adminService.GetMembersByIds(c.Request().Context(), memberIds)
	if err != nil {
		h.logger.Error("Failed to fetch member details", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "member_fetch_failed",
			Message: "Failed to fetch member details",
		})
	}

	memberNames := make(map[string]string)
	for id, member := range members {
		memberNames[id] = member.Username
	}

	// enrich transactions with member names
	for i, tx := range result {
		if name, ok := memberNames[tx.MemberID]; ok {
			result[i].MemberName = tx.MemberID
			if name != "" {
				result[i].MemberName = name
			}
		}
	}

	slices.SortFunc(result, func(a, b service.TokenTransaction) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		} else if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return cmp.Compare(a.MemberName, b.MemberName)
	})

	return c.JSON(http.StatusOK, map[string]any{
		"records": result,
		"page":    page,
		"limit":   limit,
	})
}

func (h *AdminHandler) GetTokenLedgerAnalytics(c echo.Context) error {
	result, err := h.adminService.GetTokenLedgerAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch token ledger analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "token_analytics_failed",
			Message: "Failed to fetch token ledger analytics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) GetAttendanceAnalytics(c echo.Context) error {
	result, err := h.adminService.GetAttendanceAnalytics(c.Request().Context())
	if err != nil {
		h.logger.Error("Failed to fetch attendance analytics", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_analytics_failed",
			Message: "Failed to fetch attendance analytics",
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *AdminHandler) GetMembers(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	search := c.QueryParam("search")

	result, err := h.adminService.GetMembers(c.Request().Context(), limit, page, search)
	if err != nil {
		h.logger.Error("Failed to fetch members", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "members_fetch_failed",
			Message: "Failed to fetch members",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"members": result,
		"page":    page,
		"limit":   limit,
	})
}

func (h *AdminHandler) GetAvailableAttendanceNames(c echo.Context) error {
	attendanceNames, err := h.configService.GetAvailableAttendanceNames()
	if err != nil {
		h.logger.Error("Failed to fetch attendance names", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_names_failed",
			Message: "Failed to fetch attendance names",
		})
	}

	return c.JSON(http.StatusOK, attendanceNames)
}

func (h *AdminHandler) CreateAttendanceName(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind create attendance name request", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name is required",
		})
	}

	if err := h.configService.CreateAttendanceName(name); err != nil {
		h.logger.Error("Failed to create attendance name", "name", name, "error", err)
		if strings.Contains(strings.ToLower(err.Error()), "exist") {
			return c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   "attendance_name_exists",
				Message: "Attendance name already exists",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_name_create_failed",
			Message: "Failed to create attendance name",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance name created successfully",
		"name":    name,
	})
}

func (h *AdminHandler) DeleteAttendanceName(c echo.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind delete attendance name request", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name is required",
		})
	}

	if err := h.configService.DeleteAttendanceName(name); err != nil {
		h.logger.Error("Failed to delete attendance name", "name", name, "error", err)
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "attendance_name_not_found",
				Message: "Attendance name not found",
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_name_delete_failed",
			Message: "Failed to delete attendance name",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance name deleted successfully",
		"name":    name,
	})
}

func (h *AdminHandler) CreateAttendanceRecord(c echo.Context) error {
	// get the body
	req := service.CreateAttendanceRecordInput{}
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind request body", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	if err := h.adminService.CreateAttendanceRecord(c.Request().Context(), req); err != nil {
		h.logger.Error("Failed to create attendance record", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_creation_failed",
			Message: "Failed to create attendance record",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Attendance record created successfully",
	})
}

func (h *AdminHandler) GetAttendanceRecord(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
		})
	}

	record, err := h.adminService.GetAttendanceRecord(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch attendance record", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance record",
		})
	}

	if record == nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Attendance record not found",
		})
	}

	return c.JSON(http.StatusOK, record)
}

func (h *AdminHandler) GetMembersByAttendance(c echo.Context) error {
	attendanceId := c.Param("id")
	if attendanceId == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance Id is required",
		})
	}

	members, err := h.adminService.GetMembersByAttendance(c.Request().Context(), attendanceId)
	if err != nil {
		h.logger.Error("Failed to fetch members by attendance", "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "members_fetch_failed",
			Message: "Failed to fetch members for the given attendance",
		})
	}

	return c.JSON(http.StatusOK, members)
}

func (h *AdminHandler) GetAttendanceEditPayload(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
		})
	}

	payload, err := h.adminService.GetAttendanceEditPayload(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch attendance edit payload", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Failed to fetch attendance edit payload",
		})
	}

	if payload == nil {
		return c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: "Attendance record not found",
		})
	}

	return c.JSON(http.StatusOK, payload)
}

func (h *AdminHandler) UpdateAttendanceRecord(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance record Id is required",
		})
	}

	req := service.UpdateAttendanceRecordInput{}
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Failed to bind attendance update request", "attendanceId", id, "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	availableNames, err := h.configService.GetAvailableAttendanceNames()
	if err != nil {
		h.logger.Error("Failed to fetch available attendance names", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_names_failed",
			Message: "Failed to validate attendance name",
		})
	}

	if req.Name != "" && !slices.Contains(availableNames, req.Name) {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_request",
			Message: "Attendance name must be selected from the approved list",
		})
	}

	if err := h.adminService.UpdateAttendanceRecord(c.Request().Context(), id, req); err != nil {
		h.logger.Error("Failed to update attendance record", "attendanceId", id, "error", err)

		switch {
		case errors.Is(err, service.ErrInvalidAttendanceInput):
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "invalid_request",
				Message: err.Error(),
			})
		case errors.Is(err, service.ErrAttendanceRecordNotFound):
			return c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   "not_found",
				Message: "Attendance record not found",
			})
		default:
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "attendance_update_failed",
				Message: "Failed to update attendance record",
			})
		}
	}

	payload, err := h.adminService.GetAttendanceEditPayload(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("Failed to reload attendance after update", "attendanceId", id, "error", err)
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "attendance_fetch_failed",
			Message: "Attendance updated but reloading failed",
		})
	}

	return c.JSON(http.StatusOK, payload)
}
