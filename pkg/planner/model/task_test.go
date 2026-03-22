package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weibaohui/nanobot-go/pkg/planner/model"
)

func TestSubTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    *model.SubTask
		wantErr bool
	}{
		{
			name: "valid task",
			task: &model.SubTask{
				ID:          "test",
				Name:        "Test Task",
				Description: "A test task",
				Type:        model.SubTaskTypeAnalyze,
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			task: &model.SubTask{
				Name:        "Test Task",
				Description: "A test task",
				Type:        model.SubTaskTypeAnalyze,
			},
			wantErr: true,
		},
		{
			name: "missing name",
			task: &model.SubTask{
				ID:          "test",
				Description: "A test task",
				Type:        model.SubTaskTypeAnalyze,
			},
			wantErr: true,
		},
		{
			name: "missing type",
			task: &model.SubTask{
				ID:          "test",
				Name:        "Test Task",
				Description: "A test task",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskList_ValidateAll(t *testing.T) {
	tests := []struct {
		name    string
		tasks   []*model.SubTask
		wantErr bool
	}{
		{
			name: "valid tasks",
			tasks: []*model.SubTask{
				{
					ID:   "task1",
					Name: "Task 1",
					Type: model.SubTaskTypeAnalyze,
				},
				{
					ID:        "task2",
					Name:      "Task 2",
					Type:      model.SubTaskTypeCreate,
					DependsOn: []string{"task1"},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate task ID",
			tasks: []*model.SubTask{
				{
					ID:   "task1",
					Name: "Task 1",
					Type: model.SubTaskTypeAnalyze,
				},
				{
					ID:   "task1",
					Name: "Task 2",
					Type: model.SubTaskTypeCreate,
				},
			},
			wantErr: true,
		},
		{
			name: "non-existent dependency",
			tasks: []*model.SubTask{
				{
					ID:        "task1",
					Name:      "Task 1",
					Type:      model.SubTaskTypeAnalyze,
					DependsOn: []string{"nonexistent"},
				},
			},
			wantErr: true,
		},
		{
			name: "cyclic dependency",
			tasks: []*model.SubTask{
				{
					ID:        "task1",
					Name:      "Task 1",
					Type:      model.SubTaskTypeAnalyze,
					DependsOn: []string{"task2"},
				},
				{
					ID:        "task2",
					Name:      "Task 2",
					Type:      model.SubTaskTypeCreate,
					DependsOn: []string{"task1"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := &model.TaskList{Tasks: tt.tasks}
			err := tl.ValidateAll()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaskList_GetTask(t *testing.T) {
	tasks := []*model.SubTask{
		{
			ID:   "task1",
			Name: "Task 1",
			Type: model.SubTaskTypeAnalyze,
		},
		{
			ID:   "task2",
			Name: "Task 2",
			Type: model.SubTaskTypeCreate,
		},
	}

	tl := &model.TaskList{Tasks: tasks}

	task, ok := tl.GetTask("task1")
	assert.True(t, ok)
	assert.Equal(t, "task1", task.ID)

	task, ok = tl.GetTask("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, task)
}

func TestSubTask_HasDependency(t *testing.T) {
	task := &model.SubTask{
		ID:        "task2",
		Name:      "Task 2",
		Type:      model.SubTaskTypeCreate,
		DependsOn: []string{"task1", "task0"},
	}

	assert.True(t, task.HasDependency("task1"))
	assert.True(t, task.HasDependency("task0"))
	assert.False(t, task.HasDependency("task3"))
}
