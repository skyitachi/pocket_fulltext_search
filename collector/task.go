package collector

import "skyitachi/pocket_fulltext_search/pocket"

const (
  PENDING = iota
  RUNNING
  DONE
  FAILED
)

type ExecuteFunc func(int, int) ([]pocket.CompleteItem, error)

type Task struct {
  Status int
  Executor ExecuteFunc
  Params []int
  Error error
  Result []pocket.CompleteItem
  Done chan int
}

func (t *Task) Execute() {
  go func() {
    ret, err := t.Executor(t.Params[0], t.Params[1])
    if err != nil {
      t.Status = FAILED
      t.Error = err
      t.Done <- FAILED
    }
    t.Result = ret
    t.Status = DONE
    t.Done <- DONE
  }()
}

func NewTask(fn ExecuteFunc, params []int) *Task {
  task := &Task{
    Executor: fn,
    Params: params,
    Status: PENDING,
    Done: make(chan int),
  }
  return task
}
