package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/shynome/auto-tls/db"
	"github.com/shynome/auto-tls/deploy/aliyun"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

func bindDeploy(se *core.ServeEvent) error {
	se.App.OnRecordAfterCreateSuccess(db.TableTasks).BindFunc(func(e *core.RecordEvent) error {
		go deploy(e.App, e.Record)
		return e.Next()
	})
	se.App.Cron().Add("deploy", "0 1 * * *", func() {
		if !se.App.Settings().Meta.HideControls {
			return
		}
		genDeployTask(se.App)
	})
	return se.Next()
}

func deploy(app core.App, task *core.Record) (err error) {
	logger := app.Logger().With("task", task).With("started", time.Now())
	defer err0.Then(&err, nil, func() {
		logger.Error("执行部署任务失败", "error", err)
	})
	deploy := try.To1(app.FindRecordById(db.TableDeploys, task.GetString("deploy")))
	switch deploy.GetString("target") {
	case db.TargetAliyun:
		return deployAliyunTask(app, deploy, task)
	}
	return nil
}

func genDeployTask(app core.App) (err error) {
	q := dbx.Not(dbx.HashExp{"products": ""})
	deploys, err := app.FindAllRecords(db.TableDeploys, q)
	if err != nil {
		return
	}
	for _, d := range deploys {
		target := d.GetString("target")
		switch target {
		case db.TargetAliyun:
			genAliyunDeployTask(app, d)
		}
	}
	return
}

func genAliyunDeployTask(app core.App, deploy *core.Record) (err error) {
	logger := app.Logger().With("deploy", deploy).With("started", time.Now())
	defer err0.Then(&err, nil, func() {
		logger.Error("生成部署任务失败", "error", err)
	})

	value := deploy.GetString("value")
	config := try.To1(getAliyunConfig([]byte(value)))

	domain := try.To1(app.FindRecordById(db.TableDomains, deploy.GetString("domain")))
	pem, fn := getPEMTry(app, domain)
	fn = strings.ReplaceAll(fn, "*.", "wildcard_.")

	cas := aliyun.NewCas(config)
	var certId int64
	uploadedCert := try.To1(cas.First(fn))
	if uploadedCert == nil {
		resp := try.To1(cas.Upload(pem, fn))
		certId = resp.CertId
	} else {
		certId = uploadedCert.CertificateId
	}

	products := deploy.GetStringSlice("products")
	wildcard := domain.GetString("domain")
	suffix := strings.Replace(wildcard, "*.", "", 1)

	tasks := try.To1(app.FindCollectionByNameOrId(db.TableTasks))
	taskList := []*core.Record{}

	for _, product := range products {
		switch product {
		case db.ProductCDN:
			cdn := aliyun.NewCDN(config)
			items := try.To1(cdn.List(suffix))
			for _, item := range items {
				if item.SslProtocol != "on" {
					continue
				}
				if !certmagic.MatchWildcard(item.DomainName, wildcard) {
					continue
				}
				payload := AliyunTaskPayload{
					Domain: item.DomainName,
					CertID: certId,
				}
				p := try.To1(json.Marshal(payload))
				task := core.NewRecord(tasks)
				task.Load(map[string]any{
					"deploy":  deploy.Id,
					"product": product,
					"payload": types.JSONRaw(p),
				})
				taskList = append(taskList, task)
			}
		case db.ProductDCDN:
			dcdn := aliyun.NewDCDN(config)
			items := try.To1(dcdn.List(suffix))
			for _, item := range items {
				if item.SSLProtocol != "on" {
					continue
				}
				if !certmagic.MatchWildcard(item.DomainName, wildcard) {
					continue
				}
				payload := AliyunTaskPayload{
					Domain: item.DomainName,
					CertID: certId,
				}
				p := try.To1(json.Marshal(payload))
				task := core.NewRecord(tasks)
				task.Load(map[string]any{
					"deploy":  deploy.Id,
					"product": product,
					"payload": types.JSONRaw(p),
				})
				taskList = append(taskList, task)
			}
		}
	}

	err = app.RunInTransaction(func(tx core.App) error {
		for _, task := range taskList {
			q := "deploy = {:deploy} && product = {:product} && payload = {:payload}"
			p := dbx.Params{
				"deploy":  task.GetString("deploy"),
				"product": task.GetString("product"),
				"payload": task.GetString("payload"),
			}
			_, err := app.FindFirstRecordByFilter(db.TableTasks, q, p)
			if err == nil {
				continue
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if err := tx.Save(task); err != nil {
				return err
			}
		}
		return nil
	})
	try.To(err)

	return
}

type AliyunTaskPayload struct {
	Domain string `json:"domain"`
	CertID int64  `json:"cert_id"`
}

func getAliyunConfig(value []byte) (config aliyun.Config, err error) {
	try.To(json.Unmarshal([]byte(value), &config))
	if config.Key == "" || config.Secret == "" {
		return config, fmt.Errorf("缺少 key 或 secret")
	}
	if config.Region == "" {
		config.Region = "cn-hangzhou"
	}
	return config, nil
}

func deployAliyunTask(app core.App, deploy, task *core.Record) (err error) {
	defer err0.Then(&err, func() {
		task.Set("deployed", true)
		err = app.Save(task)
	}, func() {
		var s types.JSONRaw
		s, err = json.Marshal(err.Error())
		if err != nil {
			return
		}
		task.Set("result", s)
		err = app.Save(task)
	})

	value := deploy.GetString("value")
	config := try.To1(getAliyunConfig([]byte(value)))

	payloadStr := task.GetString("payload")
	switch task.GetString("product") {
	case db.ProductCDN:
		var payload AliyunTaskPayload
		try.To(json.Unmarshal([]byte(payloadStr), &payload))
		cdn := aliyun.NewCDN(config)
		resp := try.To1(cdn.Deploy(payload.Domain, int(payload.CertID)))
		var result types.JSONRaw = try.To1(json.Marshal(resp))
		task.Set("result", result)
	case db.ProductDCDN:
		var payload AliyunTaskPayload
		try.To(json.Unmarshal([]byte(payloadStr), &payload))
		dcdn := aliyun.NewDCDN(config)
		resp := try.To1(dcdn.Deploy(payload.Domain, int(payload.CertID)))
		var result types.JSONRaw = try.To1(json.Marshal(resp))
		task.Set("result", result)
	}
	return
}
