package graph

import (
	"context"
	"testing"

	"github.com/Silicon-Ally/silicon-starter/cmd/server/model"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreateTask(t *testing.T) {
	r, env := setup(t)
	_, ctx := createUserForTest(t, env)

	anonCtx := context.Background()
	_, err := r.Mutation().CreateTask(anonCtx)
	if err == nil {
		t.Fatalf("expected an error when creating a task as an anonymous user, but got none")
	}

	taskID, err := r.Mutation().CreateTask(ctx)
	if err != nil {
		t.Fatalf("expected no error when creating a task with a logged-in context, but got %v", err)
	}

	actual, err := r.Query().Task(anonCtx, taskID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}
	expected := &model.Task{
		ID: string(taskID),
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestSetTaskName(t *testing.T) {
	r, env := setup(t)
	_, ctx := createUserForTest(t, env)
	taskID, err0 := r.Mutation().CreateTask(ctx)
	noErrDuringSetup(t, err0)

	name := "name would go here for example"
	_, err := r.Mutation().SetTaskName(ctx, taskID, name)
	if err != nil {
		t.Fatalf("setting task name: %v", err)
	}

	actual, err := r.Query().Task(ctx, taskID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}
	expected := &model.Task{
		ID:   string(taskID),
		Name: name,
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestSetTaskBody(t *testing.T) {
	r, env := setup(t)
	_, ctx := createUserForTest(t, env)
	taskID, err0 := r.Mutation().CreateTask(ctx)
	noErrDuringSetup(t, err0)

	body := "a body would go here"
	_, err := r.Mutation().SetTaskBody(ctx, taskID, body)
	if err != nil {
		t.Fatalf("setting task name: %v", err)
	}

	actual, err := r.Query().Task(ctx, taskID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}
	expected := &model.Task{
		ID:   string(taskID),
		Body: body,
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestAddTaskTags(t *testing.T) {
	r, env := setup(t)
	_, ctx := createUserForTest(t, env)
	taskID, err0 := r.Mutation().CreateTask(ctx)
	noErrDuringSetup(t, err0)

	tagA := "hi world"
	tagB := "oh hi"
	_, err := r.Mutation().AddTaskTag(ctx, taskID, tagA)
	if err != nil {
		t.Fatalf("adding task tag: %v", err)
	}
	_, err = r.Mutation().AddTaskTag(ctx, taskID, tagB)
	if err != nil {
		t.Fatalf("adding task tag: %v", err)
	}

	actual, err := r.Query().Task(ctx, taskID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}
	expected := &model.Task{
		ID:   string(taskID),
		Tags: []*string{&tagA, &tagB},
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestRemoveTaskTag(t *testing.T) {
	r, env := setup(t)
	_, ctx := createUserForTest(t, env)
	taskID, err0 := r.Mutation().CreateTask(ctx)
	tagA := "Guaddo"
	tagB := "Vaccinated for rabies"
	_, err1 := r.Mutation().AddTaskTag(ctx, taskID, tagA)
	_, err2 := r.Mutation().AddTaskTag(ctx, taskID, tagB)
	noErrDuringSetup(t, err0, err1, err2)

	_, err := r.Mutation().RemoveTaskTag(ctx, taskID, tagB)
	if err != nil {
		t.Fatalf("adding task tag: %v", err)
	}

	actual, err := r.Query().Task(ctx, taskID)
	if err != nil {
		t.Fatalf("reading task: %v", err)
	}
	expected := &model.Task{
		ID:   string(taskID),
		Tags: []*string{&tagA},
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestTagsByCreator(t *testing.T) {
	r, env := setup(t)
	userIDA, ctxA := createUserForTest(t, env)
	_, ctxB := createUserForTest(t, env)
	taskIDA1, err0 := r.Mutation().CreateTask(ctxA)
	taskIDA2, err1 := r.Mutation().CreateTask(ctxA)
	taskIDB, err2 := r.Mutation().CreateTask(ctxB)
	tagA1 := "A1"
	tagA2 := "A2"
	tagB := "B"
	_, err3 := r.Mutation().AddTaskTag(ctxA, taskIDA1, tagA1)
	_, err4 := r.Mutation().AddTaskTag(ctxA, taskIDA2, tagA2)
	_, err5 := r.Mutation().AddTaskTag(ctxB, taskIDB, tagB)
	noErrDuringSetup(t, err0, err1, err2, err3, err4, err5)

	actual, err := r.Query().TasksByCreator(ctxB, string(userIDA))
	if err != nil {
		t.Fatalf("tasks by creator: %v", err)
	}

	expected := []*model.Task{{
		ID:   string(taskIDA1),
		Tags: []*string{&tagA1},
	}, {
		ID:   string(taskIDA2),
		Tags: []*string{&tagA2},
	}}

	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func TestDeleteTask(t *testing.T) {
	r, env := setup(t)
	userID, ctx := createUserForTest(t, env)
	taskID1, err0 := r.Mutation().CreateTask(ctx)
	taskID2, err1 := r.Mutation().CreateTask(ctx)
	taskID3, err2 := r.Mutation().CreateTask(ctx)
	tag1 := "1"
	tag2 := "2"
	tag3 := "3"
	_, err3 := r.Mutation().AddTaskTag(ctx, taskID1, tag1)
	_, err4 := r.Mutation().AddTaskTag(ctx, taskID2, tag2)
	_, err5 := r.Mutation().AddTaskTag(ctx, taskID3, tag3)
	noErrDuringSetup(t, err0, err1, err2, err3, err4, err5)

	_, err := r.Mutation().DeleteTask(ctx, taskID2)
	if err != nil {
		t.Fatalf("deleting task: %v", err)
	}

	actual, err := r.Query().TasksByCreator(ctx, string(userID))
	if err != nil {
		t.Fatalf("tasks by creator: %v", err)
	}

	expected := []*model.Task{{
		ID:   string(taskID1),
		Tags: []*string{&tag1},
	}, {
		ID:   string(taskID3),
		Tags: []*string{&tag3},
	}}
	opts := cmpopts.SortSlices(func(a, b *model.Task) bool {
		return a.ID < b.ID
	})
	if diff := cmp.Diff(expected, actual, opts); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}

func taskCmpOpts() cmp.Option {
	return cmp.Options{
		cmpopts.SortSlices(func(a, b *model.Task) bool {
			return a.ID < b.ID
		}),
		cmpopts.EquateEmpty(),
	}
}
