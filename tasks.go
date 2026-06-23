package manager

import (
	"context"
	"iter"

	taskautomationapi "github.com/babelforce/manager-sdk-go/gen/taskautomation"
	taskscheduleapi "github.com/babelforce/manager-sdk-go/gen/taskschedule"
)

// InterruptTarget is the state a manager-interrupt transitions a task to.
type InterruptTarget = taskautomationapi.InterruptionTargetStates

const (
	InterruptCancel   InterruptTarget = taskautomationapi.InterruptionTargetStatesCanceled
	InterruptComplete InterruptTarget = taskautomationapi.InterruptionTargetStatesCompleted
	InterruptFail     InterruptTarget = taskautomationapi.InterruptionTargetStatesFailed
)

// TasksResource is the task-automation namespace (/api/v3/tasks).
type TasksResource struct {
	ta *taskautomationapi.ClientWithResponses
	// Schedules is the recurring task-schedule namespace (/api/v3/tasks/schedules).
	Schedules *TaskSchedulesResource
}

// ListTasksQuery filters a task listing.
type ListTasksQuery struct {
	// Filter is a server-side filter expression.
	Filter *string
	// PageSize is the page size (1..100); defaults to 100.
	PageSize int
}

// List returns an iterator over tasks, auto-paginating across pages.
func (r *TasksResource) List(ctx context.Context, q ListTasksQuery) iter.Seq2[taskautomationapi.Task, error] {
	return func(yield func(taskautomationapi.Task, error) bool) {
		var zero taskautomationapi.Task
		pageSize := q.PageSize
		if pageSize <= 0 {
			pageSize = 100
		}
		for page := 1; ; page++ {
			pg := taskautomationapi.QueryPage(page)
			ps := taskautomationapi.QueryPageSize(pageSize)
			params := &taskautomationapi.TasksParams{Page: &pg, PageSize: &ps}
			if q.Filter != nil {
				f := taskautomationapi.QueryFilter(*q.Filter)
				params.Filter = &f
			}

			resp, err := r.ta.TasksWithResponse(ctx, params)
			if err != nil {
				yield(zero, err)
				return
			}
			data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
			if err != nil {
				yield(zero, err)
				return
			}

			var records []taskautomationapi.Task
			if data.Records != nil {
				records = *data.Records
			}
			for _, t := range records {
				if !yield(t, nil) {
					return
				}
			}
			if len(records) == 0 {
				return
			}
			pages := 0
			if data.Metadata != nil && data.Metadata.PageCount != nil {
				pages = *data.Metadata.PageCount
			}
			if pages > 0 {
				if page >= pages {
					return
				}
			} else if len(records) < pageSize {
				return
			}
		}
	}
}

// ListAll collects every task into a slice.
func (r *TasksResource) ListAll(ctx context.Context, q ListTasksQuery) ([]taskautomationapi.Task, error) {
	var tasks []taskautomationapi.Task
	for t, err := range r.List(ctx, q) {
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Create creates a task.
func (r *TasksResource) Create(ctx context.Context, task taskautomationapi.SubmitTask) (*taskautomationapi.Task, error) {
	resp, err := r.ta.SubmitTaskWithResponse(ctx, task)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// CreateFromTemplate creates a task from a template, with overrides.
func (r *TasksResource) CreateFromTemplate(ctx context.Context, template string, overrides taskautomationapi.TemplateOverride) (*taskautomationapi.Task, error) {
	resp, err := r.ta.SubmitTaskTemplateWithResponse(ctx, template, nil, overrides)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Get returns a task by id.
func (r *TasksResource) Get(ctx context.Context, taskID string) (*taskautomationapi.Task, error) {
	resp, err := r.ta.TaskWithResponse(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Update updates a task.
func (r *TasksResource) Update(ctx context.Context, taskID string, task taskautomationapi.Task) (*taskautomationapi.Task, error) {
	resp, err := r.ta.UpdateTaskWithResponse(ctx, taskID, task)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Interrupt manager-interrupts a task, transitioning it to the given target state.
func (r *TasksResource) Interrupt(ctx context.Context, taskID string, target InterruptTarget, action taskautomationapi.ManualActionRequest) error {
	resp, err := r.ta.ManagerInterruptOnTaskWithResponse(ctx, taskID, target, action)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// TaskSchedulesResource is the recurring task-schedule namespace (/api/v3/tasks/schedules).
type TaskSchedulesResource struct {
	ts *taskscheduleapi.ClientWithResponses
}

// List returns all task schedules.
func (r *TaskSchedulesResource) List(ctx context.Context) (*taskscheduleapi.TaskScheduleList, error) {
	resp, err := r.ts.GetTaskSchedulesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Create creates a task schedule.
func (r *TaskSchedulesResource) Create(ctx context.Context, schedule taskscheduleapi.SubmitTaskSchedule) error {
	resp, err := r.ts.SubmitTaskScheduleWithResponse(ctx, schedule)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Get returns a task schedule by name.
func (r *TaskSchedulesResource) Get(ctx context.Context, name string) (*taskscheduleapi.TaskSchedule, error) {
	resp, err := r.ts.GetTaskScheduleWithResponse(ctx, name)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes a task schedule by name.
func (r *TaskSchedulesResource) Delete(ctx context.Context, name string) error {
	resp, err := r.ts.DeleteScheduleTaskWithResponse(ctx, name)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}
