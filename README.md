# VideoSXS

VideoSXS is a simple HTML5 based tool that you can use to compare page load video capture taken from Chrome on Android.

## Usage

VideoSXS takes a [catapult/telemetry](https://chromium.googlesource.com/catapult/+/HEAD/telemetry/README.md) benchmark output, and you can compare visual loading experience across different runs.

* Run page load benchmarks with `--capture-screen-video` option to generate mp4 screen recordings of a page load.
   * Command line example:
 ```bash
  $ tools/perf/run_benchmark run --browser android-chromium \
      --results-label=foobar --pageset-repeat=1 \
      --use-live-sites --capture-screen-video UNSCHEDULED_loading.mbi
```
* `go run serve.go -artifactsDir [artifactsDir]`. The `[artifactsDir]` should be at `[Chromium checkout dir]/tools/perf/arctifact`, unless you configured it separately.
* Navigate your browser to `http://localhost:40080`.
