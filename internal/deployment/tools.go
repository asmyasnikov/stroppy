/* Copyright 2021 The Stroppy Authors. All rights reserved         *
 * Use of this source code is governed by the 2-Clause BSD License *
 * that can be found in the LICENSE file.                          */

package deployment

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	llog "github.com/sirupsen/logrus"

	"gitlab.com/picodata/stroppy/pkg/engine/stroppy"
	"gitlab.com/picodata/stroppy/pkg/state"

	"github.com/ansel1/merry"
	"github.com/tidwall/gjson"
	"gitlab.com/picodata/stroppy/pkg/database/cluster"
	"gitlab.com/picodata/stroppy/pkg/database/config"
)

const dateFormat = "02-01-2006_15_04_05"

//nolint:nonamedreturns // should be fixed in future
func (sh *shell) executeRemotePay(
	settings *config.DatabaseSettings,
) (beginTime, endTime int64, err error) {
	llog.Debugf("DBURL: %s", settings.DBURL)

	payTestCommand := []string{
		stroppyBinaryPath,
		"pay",
		"--dir",
		stroppyHomePath,
		"--run-as-pod",
		"--url", fmt.Sprintf("%v", settings.DBURL),
		"--check",
		"--count", fmt.Sprintf("%v", settings.Count),
		"-r", fmt.Sprintf("%v", settings.BanRangeMultiplier),
		"-w", fmt.Sprintf("%v", settings.Workers),
		"--dbtype", sh.state.Settings.DatabaseSettings.DBType,
		"--log-level", sh.state.Settings.LogLevel,
	}

	llog.Tracef("Stroppy remote command '%s'", strings.Join(payTestCommand, " "))

	logFileName := fmt.Sprintf("%v_pay_%v_%v_zipfian_%v_%v.log",
		settings.DBType, settings.Count, settings.BanRangeMultiplier,
		settings.Zipfian, time.Now().Format(dateFormat))

	beginTime, endTime, err = sh.k.ExecuteRemoteCommand(
		stroppy.StroppyClientPodName,
		"",
		payTestCommand,
		logFileName,
		&sh.state,
	)
	if err != nil {
		err = merry.Prepend(err, "failed to execute remote transfer test")
	}

	return
}

// executePay - выполнить тест переводов внутри удаленного пода stroppy
func (sh *shell) executePay(shellState *state.State) error {
	var (
		settings *config.DatabaseSettings
		err      error
	)

	if settings, err = sh.readDatabaseConfig("pay"); err != nil {
		return merry.Prepend(err, "failed to read config")
	}

	var beginTime, endTime int64

	if sh.state.Settings.TestSettings.UseCloudStroppy {
		if beginTime, endTime, err = sh.executeRemotePay(settings); err != nil {
			return merry.Prepend(err, "failed to executeRemotePay")
		}
	} else {
		beginTime = (time.Now().UTC().UnixNano() / int64(time.Millisecond)) - 20000
		if err = sh.payload.Pay(shellState); err != nil {
			return merry.Prepend(err, "failed to execut local pay")
		}
		endTime = (time.Now().UTC().UnixNano() / int64(time.Millisecond)) - 20000
	}
	llog.Infof("pay test start time: '%d', end time: '%d'", beginTime, endTime)

	monImagesArchName := fmt.Sprintf("%v_pay_%v_%v_zipfian_%v_%v.tar.gz",
		settings.DBType, settings.Count, settings.BanRangeMultiplier,
		settings.Zipfian, time.Now().Format(dateFormat))

	// таймаут, чтобы не получать пустое место на графиках
	time.Sleep(20 * time.Second)

	if err = sh.k.Engine.CollectMonitoringData(
		beginTime,
		endTime,
		sh.k.MonitoringPort.Port,
		monImagesArchName,
		&sh.state,
	); err != nil {
		return merry.Prepend(err, "failed to get monitoring images for pay test")
	}

	return nil
}

// executePop - выполнить загрузку счетов в указанную БД внутри удаленного пода stroppy
func (sh *shell) executePop(shellState *state.State) error {
	var (
		settings *config.DatabaseSettings
		err      error
	)

	if settings, err = sh.readDatabaseConfig("pop"); err != nil {
		return merry.Prepend(err, "failed to read config")
	}

	llog.Debugf(
		"Stroppy executed on remote host: %v",
		sh.state.Settings.TestSettings.UseCloudStroppy,
	)

	var beginTime, endTime int64

	if sh.state.Settings.TestSettings.UseCloudStroppy {
		if beginTime, endTime, err = sh.executeRemotePop(settings); err != nil {
			return merry.Prepend(err, "failed to executeRemotePop")
		}
	} else {
		beginTime = (time.Now().UTC().UnixNano() / int64(time.Millisecond)) - 20000
		if err = sh.payload.Pop(shellState); err != nil {
			return merry.Prepend(err, "failed to execut local Pop")
		}
		endTime = (time.Now().UTC().UnixNano() / int64(time.Millisecond)) - 20000
	}

	llog.Infof("Pop test start time: '%d', end time: '%d'", beginTime, endTime)

	monImagesArchName := fmt.Sprintf("%v_pop_%v_%v_zipfian_%v_%v.tar.gz",
		settings.DBType, settings.Count, settings.BanRangeMultiplier,
		settings.Zipfian, time.Now().Format(dateFormat))

	// таймаут, чтобы не получать пустое место на графиках
	time.Sleep(20 * time.Second)

	if err = sh.k.Engine.CollectMonitoringData(
		beginTime,
		endTime,
		sh.k.MonitoringPort.Port,
		monImagesArchName,
		&sh.state,
	); err != nil {
		return merry.Prepend(err, "failed to get monitoring images for pop test")
	}

	return nil
}

//nolint:nonamedreturns // should be fixed in future
func (sh *shell) executeRemotePop(
	settings *config.DatabaseSettings,
) (beginTime, endTime int64, err error) {
	llog.Debugf("DBURL: %s", settings.DBURL)

	popTestCommand := []string{
		stroppyBinaryPath,
		"pop",
		"--dir",
		stroppyHomePath,
		"--run-as-pod",
		"--url", settings.DBURL,
		"--count", fmt.Sprintf("%v", settings.Count),
		"-r", fmt.Sprintf("%v", settings.BanRangeMultiplier),
		"-w", fmt.Sprintf("%v", settings.Workers),
		"--dbtype", sh.state.Settings.DatabaseSettings.DBType,
		"--log-level", sh.state.Settings.LogLevel,
	}

	llog.Tracef("Stroppy remote command '%s'", strings.Join(popTestCommand, " "))

	if settings.Sharded {
		popTestCommand = append(popTestCommand, "sharded")
	}

	logFileName := fmt.Sprintf("%v_pop_%v_%v_zipfian_%v_%v.log",
		settings.DBType, settings.Count, settings.BanRangeMultiplier,
		settings.Zipfian, time.Now().Format(dateFormat))

	if beginTime, endTime, err = sh.k.ExecuteRemoteCommand(
		stroppy.StroppyClientPodName,
		"",
		popTestCommand,
		logFileName,
		&sh.state,
	); err != nil {
		return 0, 0, merry.Prepend(err, "failed to execute remote populate test")
	}

	return beginTime, endTime, nil
}

// readDatabaseConfig
// прочитать конфигурационный файл test_config.json
func (sh *shell) readDatabaseConfig(cmdType string) (settings *config.DatabaseSettings, err error) {
	var data []byte

	llog.Debugf(
		"Expected test config file path %s",
		filepath.Join(sh.workingDirectory, testConfDir, configFileName),
	)

	configFilePath := filepath.Join(sh.workingDirectory, testConfDir, configFileName)
	if data, err = ioutil.ReadFile(configFilePath); err != nil {
		err = merry.Prepend(err, "failed to read config file")
		return
	}

	settings = config.DatabaseDefaults()
	settings.BanRangeMultiplier = gjson.Parse(string(data)).Get("banRangeMultiplier").Float()
	settings.DBType = sh.state.Settings.DatabaseSettings.DBType

	switch sh.state.Settings.DatabaseSettings.DBType {
	case cluster.Postgres:
		settings.DBURL = "postgres://stroppy:stroppy@acid-postgres-cluster/stroppy?sslmode=disable"

	case cluster.Foundation:
		settings.DBURL = "fdb.cluster"

	case cluster.MongoDB:
		settings.DBURL = "mongodb://stroppy:stroppy@sample-cluster-name-mongos.default.svc.cluster.local/admin?ssl=false"

	case cluster.Cockroach:
		settings.DBURL = "postgres://stroppy:stroppy@/stroppy?sslmode=disable"

	case cluster.Cartridge:
		settings.DBURL = "http://routers:8081"

	case cluster.YandexDB:
		settings.DBURL = "grpc://stroppy-ydb-database-grpc:2135/root/stroppy-ydb-database"

	default:
		err = merry.Errorf("unknown db type '%s'", sh.state.Settings.DatabaseSettings.DBType)
		return
	}

	switch cmdType {
	case "pop":
		settings.Count = int(gjson.Parse(string(data)).Get("cmd.0").Get("pop").Get("count").Int())
	case "pay":
		settings.Count = int(
			gjson.Parse(string(data)).Get("cmd.1").Get("pay").Get("count").Int(),
		)
		settings.Check = gjson.Parse(string(data)).Get("cmd.1").Get("pay").Get("Check").Bool()
		settings.Zipfian = gjson.Parse(string(data)).
			Get("cmd.1").
			Get("pay").
			Get("zipfian").
			Bool()
		settings.Oracle = gjson.Parse(string(data)).Get("cmd.1").Get("pay").Get("oracle").Bool()
	}

	return
}
