package sqldb

import (
	"fmt"

	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"github.com/jackc/pgx/v4"
)

func (db *DB) Task(tx db.Tx, id todo.TaskID) (*todo.Task, error) {
	row := db.queryRow(tx, `
		SELECT id, name, body, tags, created_by 
		FROM task
		WHERE id = $1;
		`, id)
	task, err := rowToTask(row)
	if err != nil {
		return nil, fmt.Errorf("reading task: %w", err)
	}
	return task, nil
}

func (db *DB) TasksByCreator(tx db.Tx, creatorID todo.UserID) ([]*todo.Task, error) {
	rows, err := db.query(tx, `
		SELECT id, name, body, tags, created_by
		FROM task
		WHERE created_by = $1;`, creatorID)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	tasks, err := rowsToTasks(rows)
	if err != nil {
		return nil, fmt.Errorf("reading task: %w", err)
	}
	return tasks, nil
}

const taskIDNamespace = "task"

const defaultTaskName = "Unnamed Task"
const defaultTaskBody = "New Task Body"

func (db *DB) CreateTask(tx db.Tx, creatorID todo.UserID) (todo.TaskID, error) {
	id := todo.TaskID(db.randomID(taskIDNamespace))
	name := defaultTaskName
	body := defaultTaskBody
	tags := todo.Tags{}.ToStored()
	err := db.exec(tx, `
		INSERT INTO task 
			(id, name, body, tags, created_by)
			VALUES
			($1, $2, $3, $4, $5);
		`, id, name, body, tags, creatorID)
	if err != nil {
		return "", fmt.Errorf("creating task row: %w", err)
	}
	return id, nil
}

func (d *DB) UpdateTask(
	tx db.Tx,
	taskID todo.TaskID,
	taskMutations ...db.UpdateTaskFn) error {
	err := d.RunOrContinueTransaction(tx, func(tx db.Tx) error {
		task, err := d.Task(tx, taskID)
		if err != nil {
			return fmt.Errorf("reading task pre-mutations: %w", err)
		}
		for i, m := range taskMutations {
			err := m(task)
			if err != nil {
				return fmt.Errorf("running mutation #%d: %w", i, err)
			}
		}
		err = d.putTask(tx, task)
		if err != nil {
			return fmt.Errorf("writing task post-mutations: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("running update task txn: %w", err)
	}
	return nil
}

func (d *DB) DeleteTask(tx db.Tx, taskID todo.TaskID) error {
	err := d.RunOrContinueTransaction(tx, func(tx db.Tx) error {
		err := d.exec(tx, "DELETE FROM task WHERE id = $1;", taskID)
		if err != nil {
			return fmt.Errorf("deleting task: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("running delete task transaction: %w", err)
	}
	return nil
}

func (db *DB) putTask(tx db.Tx, task *todo.Task) error {
	err := db.exec(tx, `
		UPDATE task SET
			name = $2,
			body = $3,
			tags = $4
		WHERE id = $1;
		`, task.ID, task.Name, task.Body, task.Tags.ToStored())
	if err != nil {
		return fmt.Errorf("updating task writable fields: %w", err)
	}
	return nil
}

func rowToTask(s rowScanner) (*todo.Task, error) {
	tagsAsStr := ""
	t := &todo.Task{}
	err := s.Scan(&t.ID, &t.Name, &t.Body, &tagsAsStr, &t.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("scanning into task: %w", err)
	}
	if len(tagsAsStr) > 0 {
		t.Tags = todo.TagsFromStored(tagsAsStr)
	}
	return t, nil
}

func rowsToTasks(rows pgx.Rows) ([]*todo.Task, error) {
	defer rows.Close()
	var ts []*todo.Task
	for rows.Next() {
		u, err := rowToTask(rows)
		if err != nil {
			return nil, fmt.Errorf("converting row to task: %w", err)
		}
		ts = append(ts, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("while processing task rows: %w", err)
	}
	return ts, nil
}
