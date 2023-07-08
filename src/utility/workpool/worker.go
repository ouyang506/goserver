package workpool

type Worker struct {
	task func()
}

func newWorker(task func()) *Worker {
	worker := &Worker{
		task: task,
	}
	return worker
}

func (w *Worker) run() {
	if w.task == nil {
		return
	}

	w.task()
}
