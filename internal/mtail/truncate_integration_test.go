// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.
// +build integration

package mtail_test

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/testutil"
)

func TestTruncatedLogRead(t *testing.T) {
	tmpDir, rmTmpDir := testutil.TestTempDir(t)
	defer rmTmpDir()

	logDir := path.Join(tmpDir, "logs")
	progDir := path.Join(tmpDir, "progs")
	err := os.Mkdir(logDir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Mkdir(progDir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	m, stopM := mtail.TestStartServer(t, 0, false, mtail.ProgramPath(progDir), mtail.LogPathPatterns(logDir+"/log"))
	defer stopM()

	startLineCount := mtail.TestGetMetric(t, m.Addr(), "lines_total")
	startLogCount := mtail.TestGetMetric(t, m.Addr(), "log_count")

	logFile := path.Join(logDir, "log")
	f := testutil.TestOpenFile(t, logFile)
	time.Sleep(time.Second)

	n, err := f.WriteString("1\n")
	if err != nil {
		t.Fatal(err)
	}
	glog.Infof("Wrote %d bytes", n)
	time.Sleep(time.Second)
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.OpenFile(logFile, os.O_TRUNC|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
	n, err = f.WriteString("2\n")
	if err != nil {
		t.Fatal(err)
	}
	glog.Infof("Wrote %d bytes", n)
	time.Sleep(time.Second)

	endLineCount := mtail.TestGetMetric(t, m.Addr(), "lines_total")
	endLogCount := mtail.TestGetMetric(t, m.Addr(), "log_count")

	mtail.ExpectMetricDelta(t, endLineCount, startLineCount, 2)
	mtail.ExpectMetricDelta(t, endLogCount, startLogCount, 1)
}
