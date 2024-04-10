package metrics_test

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/matthew-hagemann/ubuntu-report/internal/helper"
	"github.com/matthew-hagemann/ubuntu-report/internal/metrics"
)

func TestGetIDS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		root string

		wantDistro  string
		wantVersion string
		wantErr     bool
	}{
		{"regular", "testdata/good", "ubuntu", "18.04", false},
		{"doesn't exist", "testdata/none", "", "", true},
		{"empty file", "testdata/empty", "", "", true},
		{"missing distro", "testdata/missing-fields/ids/distro", "", "", true},
		{"missing version", "testdata/missing-fields/ids/version", "", "", true},
		{"missing both", "testdata/missing-fields/ids/both", "", "", true},
		{"empty distro", "testdata/empty-fields/ids/distro", "", "", true},
		{"empty version", "testdata/empty-fields/ids/version", "", "", true},
		{"empty both", "testdata/empty-fields/ids/both", "", "", true},
		{"garbage content", "testdata/garbage", "", "", true},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			m := newTestMetrics(t, metrics.WithRootAt(tc.root))
			d, v, err := m.GetIDS()

			a.CheckWantedErr(err, tc.wantErr)
			a.Equal(d, tc.wantDistro)
			a.Equal(v, tc.wantVersion)
		})
	}
}

func TestCollect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		root             string
		caseGPU          string
		caseCPU          string
		caseScreen       string
		casePartition    string
		caseArchitecture string
		caseLibc6        string
		caseHwCap        string
		env              map[string]string

		// note that only an internal json package error can make it returning an error
		wantErr bool
	}{
		{"regular",
			"testdata/good", "one gpu", "regular", "one screen",
			"one partition", "regular", "regular", "regular",
			map[string]string{"XDG_CURRENT_DESKTOP": "some:thing", "XDG_SESSION_DESKTOP": "ubuntusession", "XDG_SESSION_TYPE": "x12", "LANG": "fr_FR.UTF-8", "LANGUAGE": "fr_FR.UTF-8"},
			false},
		{"empty",
			"testdata/none", "empty", "empty", "empty", "empty", "empty", "empty", "empty",
			nil,
			false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			cmdGPU, cancel := newMockShortCmd(t, "lspci", "-n", tc.caseGPU)
			defer cancel()
			cmdCPU, cancel := newMockShortCmd(t, "lscpu", "-J", tc.caseCPU)
			defer cancel()
			cmdScreen, cancel := newMockShortCmd(t, "xrandr", tc.caseScreen)
			defer cancel()
			cmdPartition, cancel := newMockShortCmd(t, "df", tc.casePartition)
			defer cancel()
			cmdArchitecture, cancel := newMockShortCmd(t, "dpkg", "--print-architecture", tc.caseArchitecture)
			defer cancel()
			cmdLibc6, cancel := newMockShortCmd(t, "dpkg", "--status", "libc6", tc.caseHwCap)
			defer cancel()
			cmdHwCap, cancel := newMockShortCmd(t, "/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2", "--help", tc.caseHwCap)
			defer cancel()

			m := newTestMetrics(t, metrics.WithRootAt(tc.root),
				metrics.WithGPUInfoCommand(cmdGPU),
				metrics.WithCPUInfoCommand(cmdCPU),
				metrics.WithScreenInfoCommand(cmdScreen),
				metrics.WithSpaceInfoCommand(cmdPartition),
				metrics.WithArchitectureCommand(cmdArchitecture),
				metrics.WithHwCapCommand(cmdHwCap),
				metrics.WithLibc6Command(cmdLibc6),
				metrics.WithMapForEnv(tc.env))
			got, err := m.Collect()

			want := helper.LoadOrUpdateGolden(t, filepath.Join(tc.root, "gold", "collect"), got, *metrics.Update)
			a.CheckWantedErr(err, tc.wantErr)
			a.Equal(got, want)
		})
	}
}

func TestRunCollectTwice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		root             string
		caseGPU          string
		caseCPU          string
		caseScreen       string
		casePartition    string
		caseArchitecture string
		caseLibc6        string
		caseHwCap        string
		env              map[string]string

		// note that only an internal json package error can make it returning an error
		wantErr bool
	}{
		{"regular",
			"testdata/good", "one gpu", "regular", "one screen",
			"one partition", "regular", "regular", "regular",
			map[string]string{"XDG_CURRENT_DESKTOP": "some:thing", "XDG_SESSION_DESKTOP": "ubuntusession", "XDG_SESSION_TYPE": "x12", "LANG": "fr_FR.UTF-8", "LANGUAGE": "fr_FR.UTF-8"},
			false},
		{"empty",
			"testdata/none", "empty", "empty", "empty", "empty", "empty", "empty", "empty",
			nil,
			false},
	}
	for _, tc := range testCases {
		tc := tc // capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := helper.Asserter{T: t}

			cmdGPU, cancel := newMockShortCmd(t, "lspci", "-n", tc.caseGPU)
			defer cancel()
			cmdCPU, cancel := newMockShortCmd(t, "lscpu", "-J", tc.caseCPU)
			defer cancel()
			cmdScreen, cancel := newMockShortCmd(t, "xrandr", tc.caseScreen)
			defer cancel()
			cmdPartition, cancel := newMockShortCmd(t, "df", tc.casePartition)
			defer cancel()
			cmdArchitecture, cancel := newMockShortCmd(t, "dpkg", "--print-architecture", tc.caseArchitecture)
			defer cancel()
			cmdLibc6, cancel := newMockShortCmd(t, "dpkg", "--status", "libc6", tc.caseHwCap)
			defer cancel()
			cmdHwCap, cancel := newMockShortCmd(t, "/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2", "--help", tc.caseHwCap)
			defer cancel()

			m := newTestMetrics(t, metrics.WithRootAt(tc.root),
				metrics.WithGPUInfoCommand(cmdGPU),
				metrics.WithCPUInfoCommand(cmdCPU),
				metrics.WithScreenInfoCommand(cmdScreen),
				metrics.WithSpaceInfoCommand(cmdPartition),
				metrics.WithArchitectureCommand(cmdArchitecture),
				metrics.WithHwCapCommand(cmdHwCap),
				metrics.WithLibc6Command(cmdLibc6),
				metrics.WithMapForEnv(tc.env))
			b1, err1 := m.Collect()

			cmdGPU, cancel = newMockShortCmd(t, "lspci", "-n", tc.caseGPU)
			defer cancel()
			cmdCPU, cancel = newMockShortCmd(t, "lscpu", "-J", tc.caseCPU)
			defer cancel()
			cmdScreen, cancel = newMockShortCmd(t, "xrandr", tc.caseScreen)
			defer cancel()
			cmdPartition, cancel = newMockShortCmd(t, "df", tc.casePartition)
			defer cancel()
			cmdArchitecture, cancel = newMockShortCmd(t, "dpkg", "--print-architecture", tc.caseArchitecture)
			defer cancel()
			cmdLibc6, cancel = newMockShortCmd(t, "dpkg", "--status", "libc6", tc.caseHwCap)
			defer cancel()
			cmdHwCap, cancel = newMockShortCmd(t, "/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2", "--help", tc.caseHwCap)
			defer cancel()
			m = newTestMetrics(t, metrics.WithRootAt(tc.root),
				metrics.WithGPUInfoCommand(cmdGPU),
				metrics.WithCPUInfoCommand(cmdCPU),
				metrics.WithScreenInfoCommand(cmdScreen),
				metrics.WithSpaceInfoCommand(cmdPartition),
				metrics.WithArchitectureCommand(cmdArchitecture),
				metrics.WithHwCapCommand(cmdHwCap),
				metrics.WithLibc6Command(cmdLibc6),
				metrics.WithMapForEnv(tc.env))
			b2, err2 := m.Collect()

			a.CheckWantedErr(err1, tc.wantErr)
			a.CheckWantedErr(err2, tc.wantErr)
			var got1, got2 json.RawMessage
			json.Unmarshal(b1, &got1)
			json.Unmarshal(b2, &got2)

			a.Equal(got1, got2)
		})
	}
}

func newTestMetrics(t *testing.T, fixtures ...func(m *metrics.Metrics) error) metrics.Metrics {
	t.Helper()
	m, err := metrics.New(fixtures...)
	if err != nil {
		t.Fatal("can't create metrics object", err)
	}
	return m
}

func newMockShortCmd(t *testing.T, s ...string) (*exec.Cmd, context.CancelFunc) {
	t.Helper()
	return helper.ShortProcess(t, "TestMetricsHelperProcess", s...)
}
