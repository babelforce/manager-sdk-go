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
	// Scripts is the task-scripts namespace (/api/v3/tasks/scripts).
	Scripts *TaskScriptsResource
	// Secrets is the task-secrets namespace (/api/v3/tasks/configurations/secrets).
	Secrets *TaskSecretsResource
	// SelectionConfig is the account task-selection configuration (/api/v3/tasks/configurations/selection).
	SelectionConfig *TaskSelectionConfigResource
	// Metrics is the task & agent metrics namespace (/api/v3/tasks/metrics).
	Metrics *TaskMetricsResource
}

// Usage returns the task usage time series.
func (r *TasksResource) Usage(ctx context.Context) (*taskautomationapi.TaskTimeSeriesResponse, error) {
	resp, err := r.ta.TaskUsageWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// UsageTypes returns the available task usage types.
func (r *TasksResource) UsageTypes(ctx context.Context) (*taskautomationapi.TaskUsageTypesResponse, error) {
	resp, err := r.ta.TaskUsageTypesWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Logs returns the customer task logs.
func (r *TasksResource) Logs(ctx context.Context) (*taskautomationapi.LogsResponse, error) {
	resp, err := r.ta.ListWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TestAction tests a task action without dispatching it.
func (r *TasksResource) TestAction(ctx context.Context, body taskautomationapi.TestAction) (*taskautomationapi.TestActionResponse, error) {
	resp, err := r.ta.TestingWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ChangeState transitions a task to a new state.
//
// Deprecated: deprecated by the API; prefer Interrupt.
func (r *TasksResource) ChangeState(ctx context.Context, taskID string, taskState taskautomationapi.TaskState) error {
	resp, err := r.ta.ChangeTaskStateWithResponse(ctx, taskID, taskState)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// AgentAction performs an agent action on a task (accept / reject / complete).
func (r *TasksResource) AgentAction(ctx context.Context, taskID string, agentAction taskautomationapi.AgentActions, action taskautomationapi.ManualActionRequest) (*taskautomationapi.TaskResponse, error) {
	resp, err := r.ta.AgentActionOnTaskWithResponse(ctx, taskID, agentAction, action)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// SetAgentLock changes the agent task-locking state.
func (r *TasksResource) SetAgentLock(ctx context.Context, lockState taskautomationapi.AgentLocking) (*taskautomationapi.AgentLockState, error) {
	resp, err := r.ta.ChangeAgentLockWithResponse(ctx, lockState)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TaskMetricsResource is the task & agent metrics namespace (/api/v3/tasks/metrics).
type TaskMetricsResource struct {
	ta *taskautomationapi.ClientWithResponses
}

// TaskJournal returns the journal (event timeline) for a single task.
func (r *TaskMetricsResource) TaskJournal(ctx context.Context, taskID string) (*taskautomationapi.JournalResponse, error) {
	resp, err := r.ta.TaskJournalWithResponse(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AgentJournal returns the interaction journal for an agent.
func (r *TaskMetricsResource) AgentJournal(ctx context.Context, agentID string) (*taskautomationapi.AgentJournalResponse, error) {
	resp, err := r.ta.AgentInteractionsWithResponse(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// AgentInteractionDurations returns interaction durations for an agent.
func (r *TaskMetricsResource) AgentInteractionDurations(ctx context.Context, agentID string) (*taskautomationapi.AgentInteractionDurationsResponse, error) {
	resp, err := r.ta.AgentInteractionDurationWithResponse(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// TaskScriptsResource is the task-scripts namespace (/api/v3/tasks/scripts).
type TaskScriptsResource struct {
	ta *taskautomationapi.ClientWithResponses
}

// List lists scripts of a given type.
func (r *TaskScriptsResource) List(ctx context.Context, scriptType taskautomationapi.ScriptType) (*taskautomationapi.ScriptListResponse, error) {
	resp, err := r.ta.ListScriptsWithResponse(ctx, scriptType, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Get returns a script by type and code id.
func (r *TaskScriptsResource) Get(ctx context.Context, scriptType taskautomationapi.ScriptType, codeID string) (*taskautomationapi.ScriptResponse, error) {
	resp, err := r.ta.GetScriptWithResponse(ctx, scriptType, codeID, nil)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Submit creates a script of a given type.
func (r *TaskScriptsResource) Submit(ctx context.Context, scriptType taskautomationapi.ScriptType, body taskautomationapi.Script) (*taskautomationapi.ScriptResponse, error) {
	resp, err := r.ta.SubmitScriptWithResponse(ctx, scriptType, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Update updates a script.
func (r *TaskScriptsResource) Update(ctx context.Context, scriptType taskautomationapi.ScriptType, codeID string, body taskautomationapi.Script) (*taskautomationapi.ScriptResponse, error) {
	resp, err := r.ta.UpdateScriptWithResponse(ctx, scriptType, codeID, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Delete deletes a script.
func (r *TaskScriptsResource) Delete(ctx context.Context, scriptType taskautomationapi.ScriptType, codeID string) error {
	resp, err := r.ta.DeleteScriptWithResponse(ctx, scriptType, codeID)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// TaskSecretsResource is the task-secrets namespace (/api/v3/tasks/configurations/secrets).
type TaskSecretsResource struct {
	ta *taskautomationapi.ClientWithResponses
}

// ListPrefixes lists the secret prefixes.
func (r *TaskSecretsResource) ListPrefixes(ctx context.Context) (*taskautomationapi.SecretPrefixes, error) {
	resp, err := r.ta.ListSecretPrefixesWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// ListKeys lists the secret keys under a prefix.
func (r *TaskSecretsResource) ListKeys(ctx context.Context, prefix string) (*taskautomationapi.SecretKeys, error) {
	resp, err := r.ta.ListSecretKeysWithResponse(ctx, prefix)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Create creates secrets under a prefix.
func (r *TaskSecretsResource) Create(ctx context.Context, prefix string, secrets taskautomationapi.CreateSecretsJSONRequestBody) error {
	resp, err := r.ta.CreateSecretsWithResponse(ctx, prefix, secrets)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// Patch merges secrets under a prefix.
func (r *TaskSecretsResource) Patch(ctx context.Context, prefix string, secrets taskautomationapi.PatchSecretsJSONRequestBody) error {
	resp, err := r.ta.PatchSecretsWithResponse(ctx, prefix, secrets)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// DeleteKeys deletes the given secret keys under a prefix.
func (r *TaskSecretsResource) DeleteKeys(ctx context.Context, prefix string, keys taskautomationapi.SecretKeys) error {
	resp, err := r.ta.DeleteSecretKeysWithResponse(ctx, prefix, keys)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

// TaskSelectionConfigResource is the account task-selection configuration
// (/api/v3/tasks/configurations/selection).
type TaskSelectionConfigResource struct {
	ta *taskautomationapi.ClientWithResponses
}

// Read reads the current selection configuration.
func (r *TaskSelectionConfigResource) Read(ctx context.Context) (*taskautomationapi.SelectionConfigurationResponse, error) {
	resp, err := r.ta.ReadSelectionConfigurationWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Create creates the selection configuration.
func (r *TaskSelectionConfigResource) Create(ctx context.Context, body taskautomationapi.CreateSelectionConfigurationJSONRequestBody) (*taskautomationapi.SelectionConfigurationResponse, error) {
	resp, err := r.ta.CreateSelectionConfigurationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON201, resp.HTTPResponse, resp.Body)
}

// Update updates the selection configuration.
func (r *TaskSelectionConfigResource) Update(ctx context.Context, body taskautomationapi.UpdateSelectionConfigurationJSONRequestBody) (*taskautomationapi.SelectionConfigurationResponse, error) {
	resp, err := r.ta.UpdateSelectionConfigurationWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// Delete deletes the selection configuration.
func (r *TaskSelectionConfigResource) Delete(ctx context.Context) error {
	resp, err := r.ta.DeleteSelectionConfigurationWithResponse(ctx)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
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
