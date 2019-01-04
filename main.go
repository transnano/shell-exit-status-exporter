package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	configFile    = flag.String("config.file", "shell-exit-status-exporter.yml", "Shell exit status exporter configuration file.")
	listenAddress = flag.String("web.listen-address", ":9121", "The address to listen on for HTTP requests.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	shell         = flag.String("config.shell", "/bin/sh", "Shell to execute script")
)

type Config struct {
	Scripts []*Script `yaml:"scripts"`
}

type Script struct {
	Name    string `yaml:"name"`
	Content string `yaml:"script"`
	Timeout int64  `yaml:"timeout"`
}

type Measurement struct {
	Script     *Script
	Finished   int
	Duration   float64
	ExitStatus int
}

func runScript(script *Script) (error, int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(script.Timeout)*time.Second)
	exitStatus := 0
	defer cancel()

	bashCmd := exec.CommandContext(ctx, *shell)

	bashIn, err := bashCmd.StdinPipe()

	if err != nil {
		return err, exitStatus
	}

	if err = bashCmd.Start(); err != nil {
		return err, exitStatus
	}

	if _, err = bashIn.Write([]byte(script.Content)); err != nil {
		return err, exitStatus
	}

	bashIn.Close()

	if err := bashCmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
				log.Debugf("Exit Status: %d", exitStatus)
				if exitStatus < 0 {
					return err, exitStatus
				}
			} else {
				return err, exitStatus
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
			return err, exitStatus
		}
	}

	return nil, bashCmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
}

func runScripts(scripts []*Script) []*Measurement {
	measurements := make([]*Measurement, 0)

	ch := make(chan *Measurement)

	for _, script := range scripts {
		go func(script *Script) {
			start := time.Now()
			finished := 0
			err, exitStatus := runScript(script)
			duration := time.Since(start).Seconds()

			if err == nil {
				log.Debugf("OK: %s (after %fs).", script.Name, duration)
				finished = 1
			} else {
				log.Infof("ERROR: %s: %s (failed after %fs).", script.Name, err, duration)
			}

			ch <- &Measurement{
				Script:     script,
				Duration:   duration,
				Finished:   finished,
				ExitStatus: exitStatus,
			}
		}(script)
	}

	for i := 0; i < len(scripts); i++ {
		measurements = append(measurements, <-ch)
	}

	return measurements
}

func scriptFilter(scripts []*Script, name, pattern string) (filteredScripts []*Script, err error) {
	if name == "" && pattern == "" {
		err = errors.New("`name` or `pattern` required")
		return
	}

	var patternRegexp *regexp.Regexp

	if pattern != "" {
		patternRegexp, err = regexp.Compile(pattern)

		if err != nil {
			return
		}
	}

	for _, script := range scripts {
		if script.Name == name || (pattern != "" && patternRegexp.MatchString(script.Name)) {
			filteredScripts = append(filteredScripts, script)
		}
	}

	return
}

func scriptRunHandler(w http.ResponseWriter, r *http.Request, config *Config) {
	params := r.URL.Query()
	name := params.Get("name")
	pattern := params.Get("pattern")

	scripts, err := scriptFilter(config.Scripts, name, pattern)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	measurements := runScripts(scripts)

	for _, measurement := range measurements {
		fmt.Fprintf(w, "shell_exit_status_duration_seconds{script=\"%s\"} %f\n", measurement.Script.Name, measurement.Duration)
		fmt.Fprintf(w, "shell_exit_status_status{script=\"%s\"} %d\n", measurement.Script.Name, measurement.ExitStatus)
		fmt.Fprintf(w, "shell_exit_status_finished{script=\"%s\"} %d\n", measurement.Script.Name, measurement.Finished)
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("shell_exit_status_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("shell-exit-status-exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting shell-exit-status-exporter", version.Info())

	yamlFile, err := ioutil.ReadFile(*configFile)

	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	config := Config{}

	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	log.Infof("Loaded %d script configurations", len(config.Scripts))

	for _, script := range config.Scripts {
		if script.Timeout == 0 {
			script.Timeout = 15
		}
	}

	http.Handle(*metricsPath, prometheus.Handler())

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		scriptRunHandler(w, r, &config)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Shell exit status Exporter</title></head>
			<body>
			<h1>Shell exit status Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
