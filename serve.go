package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
)

var STORIES_RE = regexp.MustCompile(`^(.*)_(\d+)$`)

type Run struct {
	RecordingPath string `json:"recording_path"`
	TraceHtmlPath string `json:"trace_html_path"`
}

type Story struct {
	Name string `json:"name"`
	Runs []*Run `json:"runs"`
}

type Info struct {
	Label   string   `json:"label"`
	Stories []*Story `json:"stories"`
}

var artifactsDir = flag.String("artifactsDir", "", "Specify [chromium checkout]/src/tools/perf/artifacts dir")

func genArtifactsPath(abs string) (string, error) {
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return "", err
		}

		log.Printf("Failed os.Stat(%q). err: %v", abs, err)
		return "", err
	}

	rel, err := filepath.Rel(*artifactsDir, abs)
	if err != nil {
		return "", err
	}

	return filepath.Join("artifacts", rel), nil
}

func getInfo(label string) (*Info, error) {
	info := &Info{
		Label: label,
	}

	stories := make(map[string]*Story)

	dir := filepath.Join(*artifactsDir, label)
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to read dir %q: %w", dir, err)
	}

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		ms := STORIES_RE.FindStringSubmatch(fi.Name())
		if ms == nil {
			continue
		}
		name := ms[1]

		story, ok := stories[name]
		if !ok {
			story = &Story{Name: name}
			stories[name] = story
		}

		recordingp, err := genArtifactsPath(filepath.Join(dir, fi.Name(), "recording.mp4"))
		if err != nil {
			continue
		}
		tracehtmlp, err := genArtifactsPath(filepath.Join(dir, fi.Name(), "trace/trace.html"))
		if err != nil {
			continue
		}

		run := &Run{
			RecordingPath: recordingp,
			TraceHtmlPath: tracehtmlp,
		}
		story.Runs = append(story.Runs, run)
	}
	for _, story := range stories {
		if len(story.Runs) == 0 {
			continue
		}

		info.Stories = append(info.Stories, story)
	}

	if len(info.Stories) == 0 {
		return nil, errors.New("Found no stories.")
	}
	return info, nil
}

func main() {
	flag.Parse()

	fis, err := ioutil.ReadDir(*artifactsDir)
	if err != nil {
		panic(err)
	}

	infos := make(map[string]*Info)
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		label := fi.Name()

		info, err := getInfo(label)
		if err != nil {
			log.Printf("getInfo(%q) failed: %v", label, err)
			continue
		}
		infos[label] = info
	}

	http.HandleFunc("/api/labels", func(w http.ResponseWriter, req *http.Request) {
		labels := make([]string, 0, len(infos))
		for _, i := range infos {
			labels = append(labels, i.Label)
		}
		sort.Strings(labels)

		bs, err := json.Marshal(labels)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.Write(bs)
	})
	http.HandleFunc("/api/info/", func(w http.ResponseWriter, req *http.Request) {
		label := path.Base(req.URL.Path)

		bs, err := json.Marshal(infos[label])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.Write(bs)
	})

	prefix := "/artifacts/"
	http.DefaultServeMux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(*artifactsDir))))
	http.DefaultServeMux.Handle("/", http.FileServer(http.Dir("assets")))

	http.ListenAndServe(":40080", nil)
}
