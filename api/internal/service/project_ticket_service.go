package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sol-armada/sol-bot/members"
	"github.com/sol-armada/sol-bot/projects"
)

type CreateProjectTaskInput struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Priority     int32   `json:"priority"`
	Assignee     string  `json:"assignee"`
	DueAt        *string `json:"dueAt"`
	Status       string  `json:"status"`
	ParentTaskId *string `json:"parentTaskId"`
}

type UpdateProjectTaskInput struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Priority     int32   `json:"priority"`
	Assignee     string  `json:"assignee"`
	DueAt        *string `json:"dueAt"`
	Status       string  `json:"status"`
	ParentTaskId *string `json:"parentTaskId"`
}

func ListTaskStatuses(ctx context.Context, projectId uuid.UUID) ([]*projects.TaskStatus, error) {
	statuses, err := projects.ListTaskStatuses(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to list task statuses: %w", err)
	}

	return statuses, nil
}

func ListProjectTasks(ctx context.Context, projectId uuid.UUID) ([]*projects.Task, error) {
	tasks, err := projects.ListTasks(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to list project tasks: %w", err)
	}

	return tasks, nil
}

func CreateProjectTask(ctx context.Context, projectId uuid.UUID, performedById string, input CreateProjectTaskInput) (*projects.Task, error) {
	project, err := projects.GetProject(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}

	var dueAt *time.Time
	if input.DueAt != nil {
		parsedDueAt, err := time.Parse(time.RFC3339, *input.DueAt)
		if err != nil {
			return nil, fmt.Errorf("invalid due date format: %w", err)
		}
		dueAt = &parsedDueAt
	}

	var assignee *members.Member
	if strings.TrimSpace(input.Assignee) != "" && !strings.EqualFold(strings.TrimSpace(input.Assignee), "unassigned") {
		assignee, err = members.Get(input.Assignee)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch assignee: %w", err)
		}
		if assignee == nil {
			return nil, fmt.Errorf("assignee not found")
		}
	}

	var parentTaskId *string
	if input.ParentTaskId != nil && strings.TrimSpace(*input.ParentTaskId) != "" {
		parentTaskId = input.ParentTaskId
	}

	var parentTask *projects.Task
	if parentTaskId != nil && *parentTaskId != "" {
		parentTask, err = projects.GetTask(ctx, *parentTaskId)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent task: %w", err)
		}
		if parentTask == nil {
			return nil, fmt.Errorf("parent task not found")
		}
	}

	var status *projects.TaskStatus
	if strings.TrimSpace(string(input.Status)) != "" {
		status, err = projects.GetTaskStatus(ctx, projectId, string(input.Status))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch task status: %w", err)
		}
		if status == nil {
			return nil, fmt.Errorf("task status not found")
		}
	}

	newTask := projects.NewTask(
		projectId,
		input.Title,
		input.Description,
		0,
		projects.TaskPriority(input.Priority),
		dueAt,
		assignee,
		parentTask,
		status,
	)

	if err := project.AddTask(ctx, newTask); err != nil {
		return nil, fmt.Errorf("failed to add task to project: %w", err)
	}

	return newTask, nil
}
