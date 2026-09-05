package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var version = "3.0"

// ── ANSI colors ───────────────────────────────────────────────────────────────

const (
	cReset  = "\033[0m"
	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cCyan   = "\033[36m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
)

// ── Output formats ────────────────────────────────────────────────────────────

type outputFormat int

const (
	formatText outputFormat = iota
	formatJSON
	formatBurp
)

// ── Result structures ─────────────────────────────────────────────────────────

type SourceStats struct {
	Count      int    `json:"count"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type HuntResult struct {
	Domain     string                 `json:"domain"`
	Timestamp  string                 `json:"timestamp"`
	Total      int                    `json:"total"`
	Sources    map[string]SourceStats `json:"sources"`
	Subdomains []string               `json:"subdomains"`
}

// Burp Suite advanced scope JSON
type burpItem struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	File     string `json:"file"`
}
type burpScope struct {
	AdvancedMode bool       `json:"advanced_mode"`
	Exclude      []burpItem `json:"exclude"`
	Include      []burpItem `json:"include"`
}
type burpTarget struct {
	Scope burpScope `json:"scope"`
}
type burpProject struct {
	Target burpTarget `json:"target"`
}

// ── Globals ───────────────────────────────────────────────────────────────────

var (
	quietMode  bool
	httpClient *http.Client
	updateCh   = make(chan string, 1)
)

// ── Logging ───────────────────────────────────────────────────────────────────

func logInfo(f string, a ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, cBlue+"[*]"+cReset+" "+f+"\n", a...)
	}
}

func logOK(f string, a ...interface{}) {
	if !quietMode {
		fmt.Fprintf(os.Stderr, cGreen+"[+]"+cReset+" "+f+"\n", a...)
	}
}

func logWarn(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, cYellow+"[!]"+cReset+" "+f+"\n", a...)
}

func srcLine(ok bool, name string, detail string) {
	if quietMode {
		return
	}
	if ok {
		fmt.Fprintf(os.Stderr, "    "+cGreen+"✔"+cReset+"  %-14s %s\n", name, detail)
	} else {
		fmt.Fprintf(os.Stderr, "    "+cRed+"✗"+cReset+"  %-14s %s\n", name, detail)
	}
}

// ── Domain validation ─────────────────────────────────────────────────────────

var validDomainRE = regexp.MustCompile(
	`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`,
)

func cleanDomainInput(d string) (string, error) {
	d = strings.TrimSpace(d)
	d = strings.ToLower(d)
	// Strip protocol prefix
	for _, pfx := range []string{"https://", "http://"} {
		d = strings.TrimPrefix(d, pfx)
	}
	// Strip trailing slash/dot
	d = strings.TrimRight(d, "/.")
	if !validDomainRE.MatchString(d) {
		return "", fmt.Errorf("invalid domain: %q", d)
	}
	return d, nil
}

// ── HTTP helper ───────────────────────────────────────────────────────────────

func httpGet(rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "crt.sh/"+version)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited (HTTP 429)")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// ── Source: crt.sh ────────────────────────────────────────────────────────────

func fetchCrtsh(domain string) ([]string, error) {
	u := "https://crt.sh/?q=%25." + url.QueryEscape(domain) + "&output=json"
	data, err := httpGet(u)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		NameValue  string `json:"name_value"`
		CommonName string `json:"common_name"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		// crt.sh sometimes returns HTML on overload
		if strings.Contains(string(data[:min(100, len(data))]), "<html") {
			return nil, fmt.Errorf("server returned HTML (overloaded)")
		}
		return nil, fmt.Errorf("JSON parse: %w", err)
	}

	seen := make(map[string]struct{})
	var out []string
	for _, r := range rows {
		for _, field := range []string{r.NameValue, r.CommonName} {
			for _, name := range strings.Split(field, "\n") {
				name = strings.TrimSpace(strings.ToLower(name))
				if name != "" {
					if _, ok := seen[name]; !ok {
						seen[name] = struct{}{}
						out = append(out, name)
					}
				}
			}
		}
	}
	return out, nil
}

// ── Source: crt.sh org search ─────────────────────────────────────────────────

func fetchCrtshOrg(org string) ([]string, error) {
	u := "https://crt.sh/?O=" + url.QueryEscape(org) + "&output=json"
	data, err := httpGet(u)
	if err != nil {
		return nil, err
	}

	var rows []struct {
		NameValue  string `json:"name_value"`
		CommonName string `json:"common_name"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		if strings.Contains(string(data[:min(100, len(data))]), "<html") {
			return nil, fmt.Errorf("server returned HTML (overloaded)")
		}
		return nil, fmt.Errorf("JSON parse: %w", err)
	}

	seen := make(map[string]struct{})
	var out []string
	for _, r := range rows {
		for _, field := range []string{r.NameValue, r.CommonName} {
			for _, name := range strings.Split(field, "\n") {
				name = strings.TrimSpace(strings.ToLower(name))
				if name != "" && !strings.HasPrefix(name, "*") {
					if _, ok := seen[name]; !ok {
						seen[name] = struct{}{}
						out = append(out, name)
					}
				}
			}
		}
	}
	return out, nil
}

// ── Hunt by organization name (crt.sh only) ───────────────────────────────────

func huntOrg(org string) HuntResult {
	result := HuntResult{
		Domain:    org,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Sources:   make(map[string]SourceStats),
	}

	t0 := time.Now()
	names, err := fetchCrtshOrg(org)
	ms := time.Since(t0).Milliseconds()

	stat := SourceStats{DurationMs: ms}
	if err != nil {
		stat.Error = err.Error()
		result.Sources["crt.sh"] = stat
		srcLine(false, "crt.sh", fmt.Sprintf(cDim+"→ error: %v"+cReset+" "+cDim+"[%dms]"+cReset, err, ms))
		return result
	}

	seen := make(map[string]struct{})
	var allNames []string
	for _, n := range names {
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			allNames = append(allNames, n)
		}
	}
	sort.Strings(allNames)

	stat.Count = len(allNames)
	result.Sources["crt.sh"] = stat
	if len(allNames) > 0 {
		srcLine(true, "crt.sh",
			fmt.Sprintf(cDim+"→"+cReset+" "+cBold+"%d"+cReset+" "+cDim+"found"+cReset+" "+cDim+"[%dms]"+cReset,
				len(allNames), ms))
	} else {
		srcLine(false, "crt.sh", fmt.Sprintf(cDim+"→ no results [%dms]"+cReset, ms))
	}

	result.Subdomains = allNames
	result.Total = len(allNames)
	return result
}

// ── Source: certspotter (with pagination) ─────────────────────────────────────

func fetchCertspotter(domain string) ([]string, error) {
	seen := make(map[string]struct{})
	var out []string
	after := ""
	page := 0

	for {
		page++
		if page > 1 && !quietMode {
			fmt.Fprintf(os.Stderr, cDim+"         certspotter  page %d...\r"+cReset, page)
		}
		u := "https://api.certspotter.com/v1/issuances?domain=" +
			url.QueryEscape(domain) +
			"&include_subdomains=true&expand=dns_names"
		if after != "" {
			u += "&after=" + url.QueryEscape(after)
		}

		data, err := httpGet(u)
		if err != nil {
			return out, err
		}

		var entries []struct {
			ID       string   `json:"id"`
			DNSNames []string `json:"dns_names"`
		}
		if err := json.Unmarshal(data, &entries); err != nil {
			return out, fmt.Errorf("JSON parse: %w", err)
		}

		if len(entries) == 0 {
			break
		}
		for _, e := range entries {
			for _, name := range e.DNSNames {
				name = strings.ToLower(strings.TrimSpace(name))
				if _, ok := seen[name]; !ok {
					seen[name] = struct{}{}
					out = append(out, name)
				}
			}
			after = e.ID
		}
		if len(entries) < 100 {
			break
		}
	}
	return out, nil
}

// ── Source: crt.name ──────────────────────────────────────────────────────────

func fetchCrtname(domain string) ([]string, error) {
	u := "https://crt.name/v1/search?apex=" + url.QueryEscape(domain)
	data, err := httpGet(u)
	if err != nil {
		return nil, err
	}

	var out []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		name := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if name != "" {
			out = append(out, name)
		}
	}
	return out, scanner.Err()
}

// ── Source: Shodan CTL ────────────────────────────────────────────────────────

func fetchShodanCTL(domain string) ([]string, error) {
	u := "https://ctl.shodan.io/api/v1/domain/" + url.QueryEscape(domain) + "/hostnames"
	data, err := httpGet(u)
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("JSON parse: %w", err)
	}
	return names, nil
}

// ── Filter helpers ────────────────────────────────────────────────────────────

func filterSource(names []string, domain string) []string {
	re := regexp.MustCompile(`(?i)(^|\.)` + regexp.QuoteMeta(domain) + `$`)
	seen := make(map[string]struct{})
	var out []string
	for _, name := range names {
		name = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(name, "*.")))
		if name == "" || !re.MatchString(name) {
			continue
		}
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}

func globalDedup(names []string, seen map[string]struct{}) []string {
	var out []string
	for _, name := range names {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Source registry ───────────────────────────────────────────────────────────

type srcDef struct {
	name string
	fn   func(string) ([]string, error)
}

var allSources = []srcDef{
	{"crt.sh", fetchCrtsh},
	{"certspotter", fetchCertspotter},
	{"crt.name", fetchCrtname},
	{"shodan-ctl", fetchShodanCTL},
}

type srcResult struct {
	name       string
	names      []string
	err        error
	durationMs int64
}

// ── Hunt a single domain ──────────────────────────────────────────────────────

func huntDomain(domain string, skip map[string]bool, onFresh func([]string)) HuntResult {
	result := HuntResult{
		Domain:    domain,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Sources:   make(map[string]SourceStats),
	}

	activeSources := make([]srcDef, 0, len(allSources))
	for _, s := range allSources {
		if !skip[s.name] {
			activeSources = append(activeSources, s)
		}
	}
	n := len(activeSources)

	// Assign a fixed line index to each source for ANSI cursor targeting
	sourceIdx := make(map[string]int, n)
	for i, s := range activeSources {
		sourceIdx[s.name] = i
	}

	// Print pending placeholders so the source block is reserved upfront
	if !quietMode {
		for _, s := range activeSources {
			fmt.Fprintf(os.Stderr, "    "+cDim+"○"+cReset+"  %-14s\n", s.name)
		}
	}

	ch := make(chan srcResult, n)
	var wg sync.WaitGroup
	for _, s := range activeSources {
		wg.Add(1)
		go func(s srcDef) {
			defer wg.Done()
			t0 := time.Now()
			names, err := s.fn(domain)
			ch <- srcResult{s.name, names, err, time.Since(t0).Milliseconds()}
		}(s)
	}
	go func() { wg.Wait(); close(ch) }()

	globalSeen := make(map[string]struct{})
	var allNames []string
	subLinesCount := 0 // lines printed below source block (header + subdomain lines)
	headerPrinted := false

	for r := range ch {
		stat := SourceStats{DurationMs: r.durationMs}

		if !quietMode {
			idx := sourceIdx[r.name]
			goUp := (n - idx) + subLinesCount
			fmt.Fprintf(os.Stderr, "\033[%dA\033[2K\r", goUp)
			if r.err != nil {
				fmt.Fprintf(os.Stderr, "    "+cRed+"✗"+cReset+"  %-14s "+cDim+"→ error: %v [%dms]"+cReset,
					r.name, r.err, r.durationMs)
				fmt.Fprintf(os.Stderr, "\033[%dB\r", goUp)
				stat.Error = r.err.Error()
				result.Sources[r.name] = stat
				continue
			}
			filtered := filterSource(r.names, domain)
			if len(filtered) > 0 {
				fmt.Fprintf(os.Stderr, "    "+cGreen+"✔"+cReset+"  %-14s "+cDim+"→"+cReset+" "+cBold+"%d"+cReset+" "+cDim+"found [%dms]"+cReset,
					r.name, len(filtered), r.durationMs)
			} else {
				fmt.Fprintf(os.Stderr, "    "+cRed+"✗"+cReset+"  %-14s "+cDim+"→ no results [%dms]"+cReset,
					r.name, r.durationMs)
			}
			fresh := globalDedup(filtered, globalSeen)
			stat.Count = len(filtered)
			result.Sources[r.name] = stat
			allNames = append(allNames, fresh...)
			fmt.Fprintf(os.Stderr, "\033[%dB\r", goUp)
			if len(fresh) > 0 {
				if !headerPrinted {
					fmt.Fprintf(os.Stderr, "\n"+cBold+cGreen+"[+]"+cReset+" "+cBold+"Results"+cReset+"\n")
					subLinesCount += 2
					headerPrinted = true
				}
				for _, sub := range fresh {
					fmt.Fprintln(os.Stderr, sub)
					subLinesCount++
				}
			}
		} else {
			// Quiet mode: stream via onFresh callback
			if r.err != nil {
				stat.Error = r.err.Error()
				result.Sources[r.name] = stat
				continue
			}
			filtered := filterSource(r.names, domain)
			fresh := globalDedup(filtered, globalSeen)
			stat.Count = len(filtered)
			result.Sources[r.name] = stat
			allNames = append(allNames, fresh...)
			if onFresh != nil && len(fresh) > 0 {
				onFresh(fresh)
			}
			continue
		}
	}

	sort.Strings(allNames)
	result.Subdomains = allNames
	result.Total = len(allNames)
	return result
}

// ── Output formatters ─────────────────────────────────────────────────────────

func writeText(w io.Writer, results []HuntResult) {
	for _, r := range results {
		for _, sub := range r.Subdomains {
			fmt.Fprintln(w, sub)
		}
	}
}

func writeJSON(w io.Writer, results []HuntResult) error {
	var out interface{}
	if len(results) == 1 {
		out = results[0]
	} else {
		type multiResult struct {
			Timestamp string       `json:"timestamp"`
			Total     int          `json:"total"`
			Results   []HuntResult `json:"results"`
		}
		total := 0
		for _, r := range results {
			total += r.Total
		}
		out = multiResult{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Total:     total,
			Results:   results,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writeBurp(w io.Writer, results []HuntResult) error {
	proj := burpProject{
		Target: burpTarget{
			Scope: burpScope{
				AdvancedMode: true,
				Exclude:      []burpItem{},
				Include:      []burpItem{},
			},
		},
	}
	seen := make(map[string]struct{})
	for _, r := range results {
		for _, sub := range r.Subdomains {
			if _, ok := seen[sub]; ok {
				continue
			}
			seen[sub] = struct{}{}
			proj.Target.Scope.Include = append(proj.Target.Scope.Include, burpItem{
				Enabled:  true,
				Host:     "^" + regexp.QuoteMeta(sub) + "$",
				Port:     "",
				Protocol: "any",
				File:     "",
			})
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(proj)
}

// ── Banner ────────────────────────────────────────────────────────────────────

func printBanner() {
	if quietMode {
		return
	}
	fmt.Fprintf(os.Stderr, "\n"+cCyan+cBold+
		"             __         __  \n"+
		"  __________/ /_  _____/ /_ \n"+
		" / ___/ ___/ __/ / ___/ __ \\\n"+
		"/ /__/ /  / /__ (__  ) / / /\n"+
		"\\___/_/   \\__(_)____/_/ /_/  "+cReset+cBold+"v"+version+cReset+"\n\n"+
		cDim+"                    az7rb\n\n"+cReset)
}

// ── Update system ─────────────────────────────────────────────────────────────

func checkUpdate() {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/az7rb/crt.sh/releases/latest", nil)
	if err != nil {
		updateCh <- ""
		return
	}
	req.Header.Set("User-Agent", "crt.sh/"+version)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		updateCh <- ""
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		updateCh <- ""
		return
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		updateCh <- ""
		return
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(version, "v")
	if latest != "" && latest != current {
		updateCh <- fmt.Sprintf(
			cYellow+"[!]"+cReset+" New version "+cBold+"v%s"+cReset+
				" available → "+cCyan+"crt.sh -update"+cReset+"\n", latest)
	} else {
		updateCh <- ""
	}
}

func printUpdateNotice() {
	if quietMode {
		return
	}
	select {
	case msg := <-updateCh:
		if msg != "" {
			fmt.Fprint(os.Stderr, "\n"+msg)
		}
	default:
	}
}

func selfUpdate() {
	printBanner()
	fmt.Fprintf(os.Stderr, cBlue+"[*]"+cReset+" Fetching latest release info...\n")

	client := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequest("GET", "https://api.github.com/repos/az7rb/crt.sh/releases/latest", nil)
	req.Header.Set("User-Agent", "crt.sh/"+version)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Failed to reach GitHub: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var rel struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Parse error: %v\n", err)
		os.Exit(1)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	current := strings.TrimPrefix(version, "v")
	if latest == current {
		fmt.Fprintf(os.Stderr, cGreen+"[+]"+cReset+" Already on latest version "+cBold+"v%s"+cReset+"\n", current)
		return
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var assetURL, assetName string
	for _, a := range rel.Assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, goos) && strings.Contains(lower, goarch) {
			assetURL = a.BrowserDownloadURL
			assetName = a.Name
			break
		}
	}
	if assetURL == "" {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" No binary found for %s/%s\n", goos, goarch)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, cBlue+"[*]"+cReset+" Downloading "+cBold+"v%s"+cReset+" (%s)...\n", latest, assetName)
	dlResp, err := client.Get(assetURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Download failed: %v\n", err)
		os.Exit(1)
	}
	defer dlResp.Body.Close()

	var binData []byte
	switch {
	case strings.HasSuffix(assetName, ".tar.gz"):
		binData, err = extractTarGz(dlResp.Body)
	case strings.HasSuffix(assetName, ".zip"):
		raw, rerr := io.ReadAll(dlResp.Body)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Read error: %v\n", rerr)
			os.Exit(1)
		}
		binData, err = extractZip(raw)
	default:
		binData, err = io.ReadAll(dlResp.Body)
	}
	if err != nil || len(binData) == 0 {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Extraction failed: %v\n", err)
		os.Exit(1)
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Cannot find executable path: %v\n", err)
		os.Exit(1)
	}
	execPath, _ = filepath.EvalSymlinks(execPath)

	tmpPath := execPath + ".new"
	if err := os.WriteFile(tmpPath, binData, 0755); err != nil {
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Write failed: %v\n", err)
		os.Exit(1)
	}

	oldPath := execPath + ".old"
	os.Remove(oldPath)
	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Cannot replace binary: %v\n", err)
		os.Exit(1)
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Rename(oldPath, execPath)
		os.Remove(tmpPath)
		fmt.Fprintf(os.Stderr, cRed+"[!]"+cReset+" Cannot replace binary: %v\n", err)
		os.Exit(1)
	}
	os.Remove(oldPath)

	fmt.Fprintf(os.Stderr, cGreen+"[+]"+cReset+" Updated to "+cBold+"v%s"+cReset+" ✓\n", latest)
}

func extractTarGz(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		base := filepath.Base(hdr.Name)
		if base == "crt.sh" || base == "crt.sh.exe" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("binary not found in archive")
}

func extractZip(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		base := filepath.Base(f.Name)
		if base == "crt.sh" || base == "crt.sh.exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("binary not found in archive")
}

// ── Auto output filename ──────────────────────────────────────────────────────

func autoFilename(domains []string, fmt outputFormat) string {
	base := strings.Join(domains, "_")
	if len(base) > 40 {
		base = base[:40]
	}
	switch fmt {
	case formatJSON:
		return base + "_ct.json"
	case formatBurp:
		return base + "_burp_scope.json"
	default:
		return base + "_subs.txt"
	}
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	domainFlag  := flag.String("d", "", "Target domain (comma-separated for multiple)")
	listFlag    := flag.String("l", "", "File containing one domain per line")
	orgFlag     := flag.String("O", "", "Organization name (use + for spaces, e.g. -O \"Google LLC\")")
	outputFlag  := flag.String("o", "", "Output file (auto-named if -f is set without -o)")
	appendFlag  := flag.Bool("a", false, "Append to output file")
	formatFlag  := flag.String("f", "txt", "Output format: txt | json | burp")
	timeoutFlag := flag.Int("t", 30, "HTTP timeout per source in seconds")
	skipFlag    := flag.String("s", "", "Skip sources: comma-separated (e.g. crt.sh,crt.name)")
	flag.BoolVar(&quietMode, "q", false, "Quiet: only print results, no UI (useful for piping)")
	updateFlag  := flag.Bool("update", false, "Update crt.sh to the latest version")

	flag.Usage = func() {
		printBanner()
		fmt.Fprintf(os.Stderr, "\n"+cBold+"USAGE"+cReset+"\n")
		fmt.Fprintf(os.Stderr, "  crt.sh [options]\n\n")
		fmt.Fprintf(os.Stderr, cBold+"OPTIONS"+cReset+"\n")
		fmt.Fprintf(os.Stderr, "  -d  <domain>   Target domain (e.g. example.com)\n")
		fmt.Fprintf(os.Stderr, "                 Multiple: -d foo.com,bar.com\n")
		fmt.Fprintf(os.Stderr, "  -l  <file>     File with one domain per line\n")
		fmt.Fprintf(os.Stderr, "  -O  <org>      Organization name search via crt.sh (e.g. -O \"Google LLC\")\n")
		fmt.Fprintf(os.Stderr, "  -o  <file>     Output file (default: auto-named when -f is used)\n")
		fmt.Fprintf(os.Stderr, "  -a             Append to output file instead of overwrite\n")
		fmt.Fprintf(os.Stderr, "  -f  <format>   Output format:\n")
		fmt.Fprintf(os.Stderr, "                   txt   Plain text, one subdomain per line (default)\n")
		fmt.Fprintf(os.Stderr, "                   json  Structured JSON with source stats and timestamps\n")
		fmt.Fprintf(os.Stderr, "                   burp  Burp Suite advanced scope JSON\n")
		fmt.Fprintf(os.Stderr, "  -s  <sources>  Skip sources: crt.sh, certspotter, crt.name, shodan-ctl\n")
		fmt.Fprintf(os.Stderr, "  -t  <sec>      HTTP timeout per source (default: 30)\n")
		fmt.Fprintf(os.Stderr, "  -q             Quiet mode: only print found subdomains\n")
		fmt.Fprintf(os.Stderr, "  -update        Update to the latest version\n")
		fmt.Fprintf(os.Stderr, "  -h             Show this help\n\n")
	}
	flag.Parse()


	if *updateFlag {
		selfUpdate()
		return
	}

	// Check for updates in background (non-blocking)
	go checkUpdate()

	// Parse output format
	var outFmt outputFormat
	switch strings.ToLower(*formatFlag) {
	case "txt", "text", "":
		outFmt = formatText
	case "json":
		outFmt = formatJSON
	case "burp":
		outFmt = formatBurp
	default:
		fmt.Fprintf(os.Stderr, "Unknown format %q — use: txt, json, burp\n", *formatFlag)
		os.Exit(1)
	}

	// Parse skip list
	skipSources := make(map[string]bool)
	if *skipFlag != "" {
		for _, s := range strings.Split(*skipFlag, ",") {
			skipSources[strings.TrimSpace(s)] = true
		}
	}

	// Collect and validate domains
	var domains []string
	add := func(raw string) {
		d, err := cleanDomainInput(raw)
		if err != nil {
			logWarn("%v — skipping", err)
			return
		}
		domains = append(domains, d)
	}

	if *domainFlag != "" {
		for _, d := range strings.Split(*domainFlag, ",") {
			add(d)
		}
	}
	if *listFlag != "" {
		f, err := os.Open(*listFlag)
		if err != nil {
			logWarn("Cannot open list file: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		sc := bufio.NewScanner(f)
		first := true
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if first {
				// Strip UTF-8 BOM (0xEF 0xBB 0xBF) that Windows tools add
				line = strings.TrimPrefix(line, "\ufeff")
				first = false
			}
			if line != "" && !strings.HasPrefix(line, "#") {
				add(line)
			}
		}
	}

	// Org mode: query crt.sh by organization, then exit
	if *orgFlag != "" {
		t := *timeoutFlag
		if t <= 0 {
			t = 30
		}
		httpClient = &http.Client{Timeout: time.Duration(t) * time.Second}
		printBanner()
		logInfo("Organization: %s%s%s", cBold, *orgFlag, cReset)
		r := huntOrg(*orgFlag)
		logOK("Found %s%d%s unique hostnames for org %s%s%s",
			cBold, r.Total, cReset, cCyan, *orgFlag, cReset)
		for _, sub := range r.Subdomains {
			fmt.Println(sub)
		}
		printUpdateNotice()
		return
	}

	if len(domains) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// HTTP client
	// Enforce sane timeout floor
	t := *timeoutFlag
	if t <= 0 {
		t = 30
	}
	httpClient = &http.Client{Timeout: time.Duration(t) * time.Second}

	// Determine output destination
	outPath := *outputFlag
	if outPath == "" && (outFmt == formatJSON || outFmt == formatBurp) {
		outPath = autoFilename(domains, outFmt)
	}

	var outFile *os.File
	if outPath != "" {
		flags := os.O_CREATE | os.O_WRONLY
		if *appendFlag {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		f, err := os.OpenFile(outPath, flags, 0644)
		if err != nil {
			logWarn("Cannot open output file: %v", err)
			os.Exit(1)
		}
		defer f.Close()
		outFile = f
	}

	printBanner()

	if outPath != "" && !quietMode {
		logInfo("Output → %s%s%s", cBold, outPath, cReset)
	}

	// Run
	var results []HuntResult
	totalStart := time.Now()

	for i, domain := range domains {
		if i > 0 && !quietMode {
			fmt.Fprintln(os.Stderr)
		}
		logInfo("Scanning %s%s%s", cBold, domain, cReset)

		// In quiet mode: stream subdomains immediately (pipe-friendly)
		// In normal mode: collect and print as clean block after status lines
		var onFresh func([]string)
		if outFmt == formatText && quietMode {
			onFresh = func(fresh []string) {
				for _, sub := range fresh {
					fmt.Println(sub)
				}
			}
		}

		r := huntDomain(domain, skipSources, onFresh)
		results = append(results, r)

		if outFmt == formatText {
			if !quietMode {
				fmt.Fprintln(os.Stderr)
				logOK("Found %s%d%s unique subdomains for %s%s%s %s[%.1fs]%s",
					cBold, r.Total, cReset,
					cCyan, domain, cReset,
					cDim, time.Since(totalStart).Seconds(), cReset)
				if outFile != nil {
					for _, sub := range r.Subdomains {
						fmt.Fprintln(outFile, sub)
					}
				}
			}
			continue
		}

		logOK("Found %s%d%s unique subdomains for %s%s%s %s[%.1fs]%s",
			cBold, r.Total, cReset,
			cCyan, domain, cReset,
			cDim, time.Since(totalStart).Seconds(), cReset)
	}

	// For non-text formats: write all results at once
	if outFmt != formatText {
		var w io.Writer = os.Stdout
		if outFile != nil {
			w = outFile
		}
		var err error
		switch outFmt {
		case formatJSON:
			err = writeJSON(w, results)
		case formatBurp:
			err = writeBurp(w, results)
		}
		if err != nil {
			logWarn("Error writing output: %v", err)
			os.Exit(1)
		}
	}

	// Final summary
	if len(domains) > 1 && !quietMode {
		fmt.Fprintln(os.Stderr)
		total := 0
		for _, r := range results {
			total += r.Total
		}
		logOK("Total: %s%d%s unique subdomains across %d domains %s[%.1fs total]%s",
			cBold, total, cReset, len(domains),
			cDim, time.Since(totalStart).Seconds(), cReset)
	}
	if outPath != "" && !quietMode {
		logOK("Results saved → %s%s%s", cBold, outPath, cReset)
	}
	printUpdateNotice()
}
