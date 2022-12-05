package sqldb

import (
	"context"
	"testing"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreateTask(t *testing.T) {
	ctx := context.Background()
	tdb := createDBForTesting(t)
	tx := tdb.NoTxn(ctx)
	email := "user@example.com"
	userID, err0 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(email), "User's Name", email)
	noErrDuringSetup(t, err0)

	taskID, err := tdb.CreateTask(tx, userID)
	if err != nil {
		t.Fatalf("creating task: %v", err)
	}

	actual, err := tdb.Task(tx, taskID)
	if err != nil {
		t.Fatalf("getting task: %v", err)
	}
	expected := &todo.Task{
		ID:        taskID,
		CreatedBy: userID,
		Name:      defaultTaskName,
		Body:      defaultTaskBody,
		Tags:      todo.Tags{},
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Fatalf("unexpected diff (-want +got)\n%s", diff)
	}
}

func TestUpdateTask(t *testing.T) {
	ctx := context.Background()
	tdb := createDBForTesting(t)
	tx := tdb.NoTxn(ctx)
	email := "user@example.com"
	userID, err0 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(email), "User's Name", email)
	taskID, err1 := tdb.CreateTask(tx, userID)
	noErrDuringSetup(t, err0, err1)

	taskName := "Stop Climate Change"
	taskBody := "Too long to describe succinctly, read 'Drawdown' for some strategies"
	tagA := "High Priority"
	tagB := "Can wait for tomorrow"
	err := tdb.UpdateTask(tx, taskID,
		db.SetTaskBody(taskBody),
		db.SetTaskName(taskName),
		db.AddTaskTag(tagA),
		db.RemoveTaskTag(tagB))
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	actual, err := tdb.Task(tx, taskID)
	if err != nil {
		t.Fatalf("getting task: %v", err)
	}
	expected := &todo.Task{
		ID:        taskID,
		CreatedBy: userID,
		Name:      taskName,
		Body:      taskBody,
		Tags:      todo.Tags{tagA},
	}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Fatalf("unexpected diff (-want +got)\n%s", diff)
	}
}

func TestListTasks(t *testing.T) {
	ctx := context.Background()
	tdb := createDBForTesting(t)
	tx := tdb.NoTxn(ctx)
	emailA := "plankton@example.com"
	emailB := "krabbs@example.com"
	userIDA, err0 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(emailA), "User's Name", emailA)
	userIDB, err1 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(emailB), "User's Name", emailB)
	taskA1, err2 := tdb.CreateTask(tx, userIDA)
	taskA2, err3 := tdb.CreateTask(tx, userIDA)
	taskB1, err4 := tdb.CreateTask(tx, userIDB)
	nameA1 := "Conquer the World"
	nameA2 := "Seek Vengance"
	nameB1 := "Make Money"
	err5 := tdb.UpdateTask(tx, taskA1, db.SetTaskName(nameA1))
	err6 := tdb.UpdateTask(tx, taskA2, db.SetTaskName(nameA2))
	err7 := tdb.UpdateTask(tx, taskB1, db.SetTaskName(nameB1))
	noErrDuringSetup(t, err0, err1, err2, err3, err4, err5, err6, err7)

	actual, err := tdb.TasksByCreator(tx, userIDA)
	if err != nil {
		t.Fatalf("listing tasks: %v", err)
	}
	expected := []*todo.Task{{
		ID:        taskA1,
		CreatedBy: userIDA,
		Name:      nameA1,
		Body:      defaultTaskBody,
	}, {
		ID:        taskA2,
		CreatedBy: userIDA,
		Name:      nameA2,
		Body:      defaultTaskBody,
	}}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Fatalf("unexpected diff (-want +got)\n%s", diff)
	}
}

func TestDeleteTask(t *testing.T) {
	ctx := context.Background()
	tdb := createDBForTesting(t)
	tx := tdb.NoTxn(ctx)
	emailA := "plankton@example.com"
	emailB := "krabbs@example.com"
	userIDA, err0 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(emailA), "User's Name", emailA)
	userIDB, err1 := tdb.CreateUser(tx, authn.EmailAndPass, authn.UserID(emailB), "User's Name", emailB)
	taskA1, err2 := tdb.CreateTask(tx, userIDA)
	taskA2, err3 := tdb.CreateTask(tx, userIDA)
	taskB1, err4 := tdb.CreateTask(tx, userIDB)
	nameA1 := "Conquer the World"
	nameA2 := "Seek Vengance"
	nameB1 := "Make Money"
	err5 := tdb.UpdateTask(tx, taskA1, db.SetTaskName(nameA1))
	err6 := tdb.UpdateTask(tx, taskA2, db.SetTaskName(nameA2))
	err7 := tdb.UpdateTask(tx, taskB1, db.SetTaskName(nameB1))
	noErrDuringSetup(t, err0, err1, err2, err3, err4, err5, err6, err7)

	err := tdb.DeleteTask(tx, taskA2)
	if err != nil {
		t.Fatalf("deleting task: %v", err)
	}

	actual, err := tdb.TasksByCreator(tx, userIDA)
	if err != nil {
		t.Fatalf("listing tasks: %v", err)
	}
	expected := []*todo.Task{{
		ID:        taskA1,
		CreatedBy: userIDA,
		Name:      nameA1,
		Body:      defaultTaskBody,
	}}
	if diff := cmp.Diff(expected, actual, taskCmpOpts()); diff != "" {
		t.Fatalf("unexpected diff (-want +got)\n%s", diff)
	}
}

func taskCmpOpts() cmp.Option {
	userIDLessFn := func(a, b todo.TaskID) bool {
		return a < b
	}
	groupLessFn := func(a, b *todo.Task) bool {
		return a.ID < b.ID
	}
	return cmp.Options{
		cmpopts.EquateEmpty(),
		cmpopts.EquateApproxTime(time.Second),
		cmpopts.SortSlices(userIDLessFn),
		cmpopts.SortSlices(groupLessFn),
	}
}
