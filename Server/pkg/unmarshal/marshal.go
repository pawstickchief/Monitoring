package unmarshal

import (
	"Server/models/tasktype"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

func UnPackJob(value []byte) (ret *tasktype.Job, err error) {

	var job *tasktype.Job
	job = &tasktype.Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, tasktype.JobDir)
}
func ExtractKillerName(jobKey string) string {
	return strings.TrimPrefix(jobKey, tasktype.JobKill)
}

type JobEvent struct {
	EventType int
	Job       *tasktype.Job
}

func BUildJobEvent(evenType int, job *tasktype.Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: evenType,
		Job:       job,
	}
}

func BuildJobSchedulePlan(job *tasktype.Job) (jobSchedulePlan *tasktype.JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &tasktype.JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}
