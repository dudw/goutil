package testutil

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Env mocking

// MockEnvValue will store old env value, set new val. will restore old value on end.
func MockEnvValue(key, val string, fn func(nv string)) {
	old := os.Getenv(key)
	err := os.Setenv(key, val)
	if err != nil {
		panic(err)
	}

	fn(os.Getenv(key))

	// if old is empty, unset key.
	if old == "" {
		err = os.Unsetenv(key)
	} else {
		err = os.Setenv(key, old)
	}
	if err != nil {
		panic(err)
	}
}

// MockEnvValues will store old env value, set new val. will restore old value on end.
func MockEnvValues(kvMap map[string]string, fn func()) {
	backups := make(map[string]string, len(kvMap))

	for key, val := range kvMap {
		backups[key] = os.Getenv(key)
		_ = os.Setenv(key, val)
	}

	fn()

	for key := range kvMap {
		if old := backups[key]; old == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, old)
		}
	}
}

// MockOsEnvByText by env multi line text string.
// Will **CLEAR** all old ENV data, use given data map,
// and will recover old ENV after fn run. see MockCleanOsEnv
//
// Usage:
//
//	testutil.MockOsEnvByText(`
//		APP_COMMAND = login
//		APP_ENV = dev
//		APP_DEBUG = true
//
//	`, func() {
//			// do something ...
//	})
func MockOsEnvByText(envText string, fn func()) {
	ss := strings.Split(envText, "\n")
	mp := make(map[string]string, len(ss))

	for _, line := range ss {
		if line = strings.TrimSpace(line); line == "" {
			continue
		}
		nodes := strings.SplitN(line, "=", 2)
		envKey := strings.TrimSpace(nodes[0])

		if len(nodes) < 2 {
			mp[envKey] = ""
		} else {
			mp[envKey] = strings.TrimSpace(nodes[1])
		}
	}

	MockCleanOsEnv(mp, fn)
}

// MockOsEnv by env map data. alias of MockCleanOsEnv
func MockOsEnv(mp map[string]string, fn func()) {
	MockCleanOsEnv(mp, fn)
}

var envGroupSet = make(map[string]map[string]string)

// SetOsEnvs by map data with a group key. should call RemoveTmpEnvs after tested.
//
// Usage:
//
//	tmpKey := testutil.SetOsEnvs(map[string]string{
//		"APP_COMMAND": "login",
//		"APP_ENV":     "dev",
//		"APP_DEBUG":   "true",
//	})
//	defer testutil.RemoveTmpEnvs(tmpKey)
func SetOsEnvs(mp map[string]string) string {
	timeStr := strconv.FormatInt(time.Now().UnixMicro(), 32)
	tmpKey := "g_" + timeStr
	envGroupSet[tmpKey] = mp

	for key, val := range mp {
		_ = os.Setenv(key, val)
	}
	return tmpKey
}

// RemoveTmpEnvs remove test set envs by SetOsEnvs
func RemoveTmpEnvs(tmpKey string) {
	if mp, ok := envGroupSet[tmpKey]; ok {
		for key := range mp {
			_ = os.Unsetenv(key)
		}
		// delete group key
		delete(envGroupSet, tmpKey)
	}
}

// backup os ENV
var envBak = os.Environ()

// ClearOSEnv info for some testing cases.
//
// Usage:
//
//	testutil.ClearOSEnv()
//	defer testutil.RevertOSEnv()
//	// do something ...
func ClearOSEnv() { os.Clearenv() }

// RevertOSEnv info
func RevertOSEnv() {
	os.Clearenv()
	for _, str := range envBak {
		nodes := strings.SplitN(str, "=", 2)
		_ = os.Setenv(nodes[0], nodes[1])
	}
}

// RunOnCleanEnv will CLEAR all old ENV, then run given func.
// will RECOVER old ENV after fn run.
func RunOnCleanEnv(runFn func()) {
	os.Clearenv()
	runFn()
	RevertOSEnv()
}

// MockCleanOsEnv by env map data.
//
// will CLEAR all old ENV data, use given a data map.
// will RECOVER old ENV after fn run.
func MockCleanOsEnv(mp map[string]string, fn func()) {
	os.Clearenv()
	for key, val := range mp {
		_ = os.Setenv(key, val)
	}

	fn()

	os.Clearenv()
	for _, str := range envBak {
		nodes := strings.SplitN(str, "=", 2)
		_ = os.Setenv(nodes[0], nodes[1])
	}
}
