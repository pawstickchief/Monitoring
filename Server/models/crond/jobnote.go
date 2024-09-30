package crond

import (
	"context"
	"errors"
	"github.com/gorhill/cronexpr"
	"go-web-app/models"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronexpr"`
}
type JobMgr struct {
	Kv       clientv3.KV
	Lease    clientv3.Lease
	Clinet   *clientv3.Client
	Watcher  clientv3.Watcher
	JobEvent clientv3.Event
}
type ParameCrontab struct {
	ParameOption   string `json:"parameoption" bindding:"required"`
	models.Filelog `json:"filelog"`
	CrontabJob     `json:"crontabjob"`
	Job            `json:"job"`
	JobMgr         `json:"jobmgr"`
}
type CrontabJob struct {
	JobId        int    `json:"jobid"`
	JobCronExpr  string `json:"jobcronexpr"`
	JobName      string `json:"jobname"`
	JobShell     string `json:"jobshell"`
	JobStatus    int    `json:"jobstatus"`
	JobStartTime string `json:"jobstarttime"`
	JobStopTime  string `json:"jobstoptime"`
	JobInfo      string `json:"jobinfo"`
	JobRunning   string `json:"jobrunning"`
	JobErr       string `json:"joberr"`
}

const (
	JobDir         = "/cron/jobs/"
	JobKill        = "/cron/kill/"
	JobLock        = "/cron/lock/"
	JobEventSave   = 1
	JobEventDelete = 2
	JobKiller      = 3
)

type JobSchedulePlan struct {
	Job      *Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

type JobExecutingInfo struct {
	Job        *Job
	PlanTime   time.Time
	RealTime   time.Time
	CancleCtx  context.Context
	CancleFunc context.CancelFunc
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecutingInfo
	Output      []byte
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

var (
	ERR_LOCK_ALREADY_REQUIRED = errors.New("锁已被占用")
)
