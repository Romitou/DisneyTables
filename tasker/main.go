package tasker

import (
	"github.com/go-co-op/gocron"
	"log"
	"time"
)

var tasker *Tasker

type Tasker struct {
	scheduler *gocron.Scheduler
	tasks     []*Task
}

func Get() *Tasker {
	if tasker == nil {
		tasker = &Tasker{}
	}
	return tasker
}

type Task struct {
	Cron        string
	Immediately bool
	Run         func()
}

func (t *Tasker) RegisterTasks(tasks ...*Task) {
	for _, task := range tasks {
		t.tasks = append(t.tasks, task)
	}
}

func (t *Tasker) Start() {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		location = time.UTC
	}

	t.scheduler = gocron.NewScheduler(location)
	for _, task := range t.tasks {
		job := t.scheduler.Cron(task.Cron)
		if task.Immediately {
			job.StartImmediately()
		}
		_, taskErr := job.Do(task.Run)
		if taskErr != nil {
			log.Println(taskErr)
			continue
		}
	}

	t.scheduler.StartBlocking()
}
