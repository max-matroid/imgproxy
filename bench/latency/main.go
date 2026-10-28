// Command latency compares the latency of fetching images directly from an
// origin server against fetching them through imgproxy with various
// processing options.
//
// It starts a local origin HTTP server that serves a set of test images
// (generated on the fly unless -images is given), then for every image runs
// each scenario: a direct fetch from the origin, and several imgproxy URLs
// that point back at the same origin.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type scenario struct {
	Name string
	// Path is the imgproxy processing path; empty means no processing options.
	Path string
	// Ext is an optional output extension appended to the source URL.
	Ext string
}

var allScenarios = []scenario{
	{Name: "direct"},
	{Name: "raw", Path: "raw:1"},
	{Name: "reencode"},
	{Name: "resize", Path: "rs:fit:300:300"},
	{Name: "webp", Path: "rs:fit:300:300", Ext: "webp"},
	{Name: "avif", Path: "rs:fit:300:300", Ext: "avif"},
}

type sample struct {
	TTFB  time.Duration
	Total time.Duration
	Bytes int64
}

type result struct {
	Image     string  `json:"image"`
	Scenario  string  `json:"scenario"`
	Requests  int     `json:"requests"`
	Errors    int     `json:"errors"`
	RPS       float64 `json:"rps"`
	AvgBytes  int64   `json:"avg_bytes"`
	TTFBP50   float64 `json:"ttfb_p50_ms"`
	TotalMin  float64 `json:"total_min_ms"`
	TotalMean float64 `json:"total_mean_ms"`
	TotalP50  float64 `json:"total_p50_ms"`
	TotalP90  float64 `json:"total_p90_ms"`
	TotalP99  float64 `json:"total_p99_ms"`
	TotalMax  float64 `json:"total_max_ms"`
}

func main() {
	var (
		imgproxyURL = flag.String("imgproxy", "http://127.0.0.1:8080", "imgproxy base URL")
		originAddr  = flag.String("origin-listen", "0.0.0.0:9000", "address the local origin server listens on")
		originLocal = flag.String("origin-direct", "http://127.0.0.1:9000", "origin base URL used for direct fetches")
		originProxy = flag.String("origin-imgproxy", "http://host.docker.internal:9000", "origin base URL as seen from imgproxy")
		imagesDir   = flag.String("images", "", "directory with source images (default: generate test images)")
		scenarios   = flag.String("scenarios", "direct,raw,reencode,resize,webp,avif", "comma-separated scenarios to run")
		requests    = flag.Int("n", 200, "measured requests per image/scenario")
		warmup      = flag.Int("warmup", 10, "warm-up requests per image/scenario (not measured)")
		concurrency = flag.Int("c", 1, "concurrent requests")
		timeout     = flag.Duration("timeout", 30*time.Second, "per-request timeout")
		jsonOut     = flag.String("json", "", "write results as JSON to this file")
	)
	flag.Parse()

	selected, err := selectScenarios(*scenarios)
	if err != nil {
		log.Fatal(err)
	}

	dir := *imagesDir
	if dir == "" {
		dir, err = os.MkdirTemp("", "imgproxy-latency-*")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dir)
		if err := generateImages(dir); err != nil {
			log.Fatal(err)
		}
	}

	images, err := listImages(dir)
	if err != nil {
		log.Fatal(err)
	}
	if len(images) == 0 {
		log.Fatalf("no images found in %s", dir)
	}

	ln, err := net.Listen("tcp", *originAddr)
	if err != nil {
		log.Fatal(err)
	}
	go http.Serve(ln, http.FileServer(http.Dir(dir)))

	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			MaxIdleConns:        *concurrency * 2,
			MaxIdleConnsPerHost: *concurrency * 2,
			DisableCompression:  true,
		},
	}

	if err := checkReachable(client, *imgproxyURL+"/health"); err != nil {
		log.Fatalf("imgproxy is not reachable at %s: %v", *imgproxyURL, err)
	}

	var results []result
	for _, img := range images {
		for _, sc := range selected {
			url := buildURL(sc, img, *imgproxyURL, *originLocal, *originProxy)

			// Fail fast on a misconfigured scenario instead of measuring errors.
			if s, err := fetch(client, url); err != nil {
				log.Fatalf("%s/%s: %v (%s)", img, sc.Name, err, url)
			} else if s.Bytes == 0 {
				log.Fatalf("%s/%s: empty response (%s)", img, sc.Name, url)
			}

			run(client, url, *warmup, *concurrency)
			samples, errs, elapsed := run(client, url, *requests, *concurrency)

			r := summarize(img, sc.Name, samples, errs, elapsed)
			results = append(results, r)
			fmt.Fprintf(os.Stderr, "%-14s %-9s p50=%7.2fms p99=%7.2fms errors=%d\n",
				img, sc.Name, r.TotalP50, r.TotalP99, r.Errors)
		}
	}

	fmt.Println()
	printResults(results, *requests, *concurrency)

	if *jsonOut != "" {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(*jsonOut, data, 0o644); err != nil {
			log.Fatal(err)
		}
	}
}

func selectScenarios(list string) ([]scenario, error) {
	var out []scenario
	for _, name := range strings.Split(list, ",") {
		name = strings.TrimSpace(name)
		i := slices.IndexFunc(allScenarios, func(s scenario) bool { return s.Name == name })
		if i < 0 {
			return nil, fmt.Errorf("unknown scenario %q", name)
		}
		out = append(out, allScenarios[i])
	}
	return out, nil
}

func buildURL(sc scenario, img, imgproxyURL, originLocal, originProxy string) string {
	if sc.Name == "direct" {
		return originLocal + "/" + img
	}

	src := originProxy + "/" + img
	if sc.Ext != "" {
		src += "@" + sc.Ext
	}

	// No IMGPROXY_KEY is set, so any signature segment is accepted.
	parts := []string{imgproxyURL, "insecure"}
	if sc.Path != "" {
		parts = append(parts, sc.Path)
	}
	parts = append(parts, "plain", src)
	return strings.Join(parts, "/")
}

func checkReachable(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func fetch(client *http.Client, url string) (sample, error) {
	var ttfb time.Time
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() { ttfb = time.Now() },
	}

	req, err := http.NewRequestWithContext(
		httptrace.WithClientTrace(context.Background(), trace), http.MethodGet, url, nil,
	)
	if err != nil {
		return sample{}, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return sample{}, err
	}
	defer resp.Body.Close()

	n, err := io.Copy(io.Discard, resp.Body)
	total := time.Since(start)
	if err != nil {
		return sample{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return sample{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	return sample{TTFB: ttfb.Sub(start), Total: total, Bytes: n}, nil
}

// run issues n requests to url using c concurrent workers.
func run(client *http.Client, url string, n, c int) ([]sample, int, time.Duration) {
	var (
		mu      sync.Mutex
		samples = make([]sample, 0, n)
		errs    int
		wg      sync.WaitGroup
		jobs    = make(chan struct{})
	)

	start := time.Now()
	for range c {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				s, err := fetch(client, url)
				mu.Lock()
				if err != nil {
					errs++
				} else {
					samples = append(samples, s)
				}
				mu.Unlock()
			}
		}()
	}
	for range n {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()

	return samples, errs, time.Since(start)
}

func summarize(img, name string, samples []sample, errs int, elapsed time.Duration) result {
	r := result{
		Image:    img,
		Scenario: name,
		Requests: len(samples) + errs,
		Errors:   errs,
		RPS:      float64(len(samples)) / elapsed.Seconds(),
	}
	if len(samples) == 0 {
		return r
	}

	totals := make([]time.Duration, len(samples))
	ttfbs := make([]time.Duration, len(samples))
	var sum time.Duration
	var bytes int64
	for i, s := range samples {
		totals[i] = s.Total
		ttfbs[i] = s.TTFB
		sum += s.Total
		bytes += s.Bytes
	}
	slices.Sort(totals)
	slices.Sort(ttfbs)

	r.AvgBytes = bytes / int64(len(samples))
	r.TTFBP50 = ms(percentile(ttfbs, 50))
	r.TotalMin = ms(totals[0])
	r.TotalMean = ms(sum / time.Duration(len(samples)))
	r.TotalP50 = ms(percentile(totals, 50))
	r.TotalP90 = ms(percentile(totals, 90))
	r.TotalP99 = ms(percentile(totals, 99))
	r.TotalMax = ms(totals[len(totals)-1])
	return r
}

// percentile uses the nearest-rank method on a sorted slice.
func percentile(sorted []time.Duration, p float64) time.Duration {
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	return sorted[max(rank-1, 0)]
}

func ms(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}

func printResults(results []result, n, c int) {
	fmt.Printf("Latency (total time incl. body read), %d requests, concurrency %d\n\n", n, c)

	direct := map[string]float64{}
	for _, r := range results {
		if r.Scenario == "direct" {
			direct[r.Image] = r.TotalP50
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "image\tscenario\tbytes\tttfb p50\tp50\tp90\tp99\tmax\trps\tp50 vs direct\terrors\t")
	for _, r := range results {
		overhead := "-"
		if d, ok := direct[r.Image]; ok && r.Scenario != "direct" {
			overhead = fmt.Sprintf("+%.2fms (%.1fx)", r.TotalP50-d, r.TotalP50/d)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.0f\t%s\t%d\t\n",
			r.Image, r.Scenario, humanBytes(r.AvgBytes), r.TTFBP50,
			r.TotalP50, r.TotalP90, r.TotalP99, r.TotalMax, r.RPS, overhead, r.Errors)
	}
	w.Flush()
	fmt.Println("\nAll times in milliseconds.")
}

func humanBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1fMB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1fKB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%dB", b)
	}
}

func listImages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".jpg", ".jpeg", ".png", ".webp", ".avif", ".gif":
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// generateImages writes deterministic photo-like test images of several
// sizes. Gradients plus noise keep them from compressing unrealistically well.
func generateImages(dir string) error {
	specs := []struct {
		name string
		w, h int
	}{
		{"small.jpg", 640, 480},
		{"medium.jpg", 1920, 1080},
		{"large.jpg", 4032, 3024},
		{"medium.png", 1920, 1080},
	}

	for _, s := range specs {
		img := syntheticImage(s.w, s.h)
		f, err := os.Create(filepath.Join(dir, s.name))
		if err != nil {
			return err
		}
		if strings.HasSuffix(s.name, ".png") {
			err = png.Encode(f, img)
		} else {
			err = jpeg.Encode(f, img, &jpeg.Options{Quality: 85})
		}
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func syntheticImage(w, h int) image.Image {
	rng := rand.New(rand.NewSource(42))
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			fx, fy := float64(x)/float64(w), float64(y)/float64(h)
			n := float64(rng.Intn(24) - 12)
			img.SetRGBA(x, y, color.RGBA{
				R: clamp(255*fx + 40*math.Sin(fy*12) + n),
				G: clamp(255*fy + 40*math.Cos(fx*9) + n),
				B: clamp(128 + 100*math.Sin((fx+fy)*7) + n),
				A: 255,
			})
		}
	}
	return img
}

func clamp(v float64) uint8 {
	return uint8(max(0, min(255, v)))
}
