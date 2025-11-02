package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// CoverageHelper provides utilities for code coverage testing
type CoverageHelper struct {
	t                *testing.T
	coverageDir      string
	coverageFile     string
	profileFile      string
	htmlReportDir    string
	excludePatterns  []*regexp.Regexp
	thresholds       CoverageThresholds
}

// CoverageThresholds defines coverage thresholds
type CoverageThresholds struct {
	Overall    float64
	Packages   map[string]float64
	Functions  float64
	Statements float64
	Branches   float64
	Lines      float64
}

// CoverageReport represents a coverage report
type CoverageReport struct {
	Overall    CoverageMetrics
	Packages   map[string]CoverageMetrics
	Functions  []FunctionCoverage
	Uncovered  []UncoveredLine
	Summary    CoverageSummary
}

// CoverageMetrics represents coverage metrics for a package or overall
type CoverageMetrics struct {
	Percent    float64
	Covered    int
	Total      int
	Functions  int
	Statements int
	Branches   int
	Lines      int
}

// FunctionCoverage represents coverage for a specific function
type FunctionCoverage struct {
	Name       string
	File       string
	StartLine  int
	EndLine    int
	Percent    float64
	Statements int
	Covered    int
}

// UncoveredLine represents an uncovered line of code
type UncoveredLine struct {
	File       string
	LineNumber int
	Line       string
	Function   string
}

// CoverageSummary represents a summary of coverage information
type CoverageSummary struct {
	TotalPackages    int
	CoveredPackages  int
	TotalFunctions   int
	CoveredFunctions int
	TotalLines       int
	CoveredLines     int
}

// NewCoverageHelper creates a new coverage helper
func NewCoverageHelper(t *testing.T) *CoverageHelper {
	return &CoverageHelper{
		t:               t,
		coverageDir:      "coverage",
		coverageFile:     "coverage.out",
		profileFile:      "coverage.prof",
		htmlReportDir:    "coverage.html",
		excludePatterns:  make([]*regexp.Regexp, 0),
		thresholds: CoverageThresholds{
			Overall:    80.0,
			Packages:   make(map[string]float64),
			Functions:  80.0,
			Statements: 80.0,
			Branches:   80.0,
			Lines:      80.0,
		},
	}
}

// WithCoverageDir sets the coverage directory
func (ch *CoverageHelper) WithCoverageDir(dir string) *CoverageHelper {
	ch.coverageDir = dir
	return ch
}

// WithCoverageFile sets the coverage file name
func (ch *CoverageHelper) WithCoverageFile(file string) *CoverageHelper {
	ch.coverageFile = file
	return ch
}

// WithHTMLReportDir sets the HTML report directory
func (ch *CoverageHelper) WithHTMLReportDir(dir string) *CoverageHelper {
	ch.htmlReportDir = dir
	return ch
}

// WithExcludePattern adds an exclude pattern
func (ch *CoverageHelper) WithExcludePattern(pattern string) *CoverageHelper {
	regex, err := regexp.Compile(pattern)
	require.NoError(ch.t, err, "Failed to compile exclude pattern")
	ch.excludePatterns = append(ch.excludePatterns, regex)
	return ch
}

// WithThresholds sets coverage thresholds
func (ch *CoverageHelper) WithThresholds(thresholds CoverageThresholds) *CoverageHelper {
	ch.thresholds = thresholds
	return ch
}

// SetupCoverage sets up coverage collection
func (ch *CoverageHelper) SetupCoverage() {
	// Create coverage directory
	err := os.MkdirAll(ch.coverageDir, 0755)
	require.NoError(ch.t, err, "Failed to create coverage directory")
	
	// Set environment variables for coverage
	os.Setenv("GOCOVERDIR", ch.coverageDir)
	os.Setenv("GOCOVERPROFILE", ch.coverageFile)
}

// RunTestsWithCoverage runs tests with coverage collection
func (ch *CoverageHelper) RunTestsWithCoverage(testPattern string, args ...string) error {
	// Build test command with coverage
	cmdArgs := []string{
		"test",
		"-coverprofile=" + filepath.Join(ch.coverageDir, ch.coverageFile),
		"-covermode=count",
		"-v",
	}
	
	if testPattern != "" {
		cmdArgs = append(cmdArgs, "-run", testPattern)
	}
	
	cmdArgs = append(cmdArgs, args...)
	
	// Run tests
	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// GenerateCoverageReport generates coverage report
func (ch *CoverageHelper) GenerateCoverageReport() (*CoverageReport, error) {
	coveragePath := filepath.Join(ch.coverageDir, ch.coverageFile)
	
	// Check if coverage file exists
	if _, err := os.Stat(coveragePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("coverage file not found: %s", coveragePath)
	}
	
	// Parse coverage file
	report, err := ch.parseCoverageFile(coveragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage file: %w", err)
	}
	
	// Apply exclusions
	ch.applyExclusions(report)
	
	// Calculate summary
	ch.calculateSummary(report)
	
	return report, nil
}

// parseCoverageFile parses the coverage file
func (ch *CoverageHelper) parseCoverageFile(filePath string) (*CoverageReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	report := &CoverageReport{
		Packages:  make(map[string]CoverageMetrics),
		Functions: make([]FunctionCoverage, 0),
		Uncovered: make([]UncoveredLine, 0),
	}
	
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var coverage struct {
			Mode     string `json:"mode"`
			Files     []struct {
				File       string `json:"file"`
				Functions []struct {
					Name       string `json:"name"`
					Start      int    `json:"start"`
					End        int    `json:"end"`
					Statements int    `json:"statements"`
					Covered    int    `json:"covered"`
				} `json:"functions"`
				Lines []struct {
					LineNumber int    `json:"line_number"`
					Statements int    `json:"statements"`
					Covered    int    `json:"covered"`
					Line       string `json:"line"`
				} `json:"lines"`
			} `json:"files"`
		}
		
		err := decoder.Decode(&coverage)
		if err != nil {
			return nil, err
		}
		
		// Process coverage data
		for _, file := range coverage.Files {
			packageMetrics := CoverageMetrics{}
			
			for _, fn := range file.Functions {
				percent := float64(fn.Covered) / float64(fn.Statements) * 100
				report.Functions = append(report.Functions, FunctionCoverage{
					Name:       fn.Name,
					File:       file.File,
					StartLine:  fn.Start,
					EndLine:    fn.End,
					Percent:    percent,
					Statements: fn.Statements,
					Covered:    fn.Covered,
				})
				
				packageMetrics.Functions++
				packageMetrics.Statements += fn.Statements
				packageMetrics.Covered += fn.Covered
			}
			
			for _, line := range file.Lines {
				if line.Covered == 0 {
					report.Uncovered = append(report.Uncovered, UncoveredLine{
						File:       file.File,
						LineNumber: line.LineNumber,
						Line:       line.Line,
						Function:   ch.findFunctionForLine(report.Functions, file.File, line.LineNumber),
					})
				}
				
				packageMetrics.Lines++
				if line.Covered > 0 {
					packageMetrics.Covered++
				}
			}
			
			if packageMetrics.Statements > 0 {
				packageMetrics.Percent = float64(packageMetrics.Covered) / float64(packageMetrics.Statements) * 100
			}
			
			packageName := ch.extractPackageName(file.File)
			report.Packages[packageName] = packageMetrics
		}
	}
	
	return report, nil
}

// applyExclusions applies exclusion patterns to coverage report
func (ch *CoverageHelper) applyExclusions(report *CoverageReport) {
	if len(ch.excludePatterns) == 0 {
		return
	}
	
	// Filter functions
	var filteredFunctions []FunctionCoverage
	for _, fn := range report.Functions {
		if !ch.shouldExclude(fn.File) {
			filteredFunctions = append(filteredFunctions, fn)
		}
	}
	report.Functions = filteredFunctions
	
	// Filter uncovered lines
	var filteredUncovered []UncoveredLine
	for _, line := range report.Uncovered {
		if !ch.shouldExclude(line.File) {
			filteredUncovered = append(filteredUncovered, line)
		}
	}
	report.Uncovered = filteredUncovered
	
	// Filter packages
	filteredPackages := make(map[string]CoverageMetrics)
	for pkg, metrics := range report.Packages {
		if !ch.shouldExcludePackage(pkg) {
			filteredPackages[pkg] = metrics
		}
	}
	report.Packages = filteredPackages
}

// shouldExclude checks if a file should be excluded
func (ch *CoverageHelper) shouldExclude(filePath string) bool {
	for _, pattern := range ch.excludePatterns {
		if pattern.MatchString(filePath) {
			return true
		}
	}
	return false
}

// shouldExcludePackage checks if a package should be excluded
func (ch *CoverageHelper) shouldExcludePackage(packageName string) bool {
	for _, pattern := range ch.excludePatterns {
		if pattern.MatchString(packageName) {
			return true
		}
	}
	return false
}

// extractPackageName extracts package name from file path
func (ch *CoverageHelper) extractPackageName(filePath string) string {
	// Extract package name from file path
	parts := strings.Split(filePath, string(filepath.Separator))
	for i, part := range parts {
		if part == "internal" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "unknown"
}

// findFunctionForLine finds the function containing a specific line
func (ch *CoverageHelper) findFunctionForLine(functions []FunctionCoverage, file string, line int) string {
	for _, fn := range functions {
		if fn.File == file && line >= fn.StartLine && line <= fn.EndLine {
			return fn.Name
		}
	}
	return ""
}

// calculateSummary calculates summary statistics
func (ch *CoverageHelper) calculateSummary(report *CoverageReport) {
	summary := CoverageSummary{
		TotalPackages:    len(report.Packages),
		TotalFunctions:   len(report.Functions),
		TotalLines:       0,
		CoveredLines:     0,
	}
	
	// Calculate overall coverage
	var totalStatements, totalCovered int
	for _, metrics := range report.Packages {
		totalStatements += metrics.Statements
		totalCovered += metrics.Covered
		
		if metrics.Percent > 0 {
			summary.CoveredPackages++
		}
		
		summary.TotalLines += metrics.Lines
		summary.CoveredLines += metrics.Covered
	}
	
	if totalStatements > 0 {
		report.Overall = CoverageMetrics{
			Percent:    float64(totalCovered) / float64(totalStatements) * 100,
			Covered:    totalCovered,
			Total:      totalStatements,
			Functions:  summary.TotalFunctions,
			Statements: totalStatements,
			Lines:      summary.TotalLines,
			Branches:   0, // Not available in this format
		}
	}
	
	// Count covered functions
	for _, fn := range report.Functions {
		if fn.Percent > 0 {
			summary.CoveredFunctions++
		}
	}
	
	report.Summary = summary
}

// GenerateHTMLReport generates HTML coverage report
func (ch *CoverageHelper) GenerateHTMLReport(report *CoverageReport) error {
	// Create HTML report directory
	err := os.MkdirAll(ch.htmlReportDir, 0755)
	if err != nil {
		return err
	}
	
	// Generate HTML content
	htmlContent := ch.generateHTMLContent(report)
	
	// Write HTML file
	htmlFile := filepath.Join(ch.htmlReportDir, "index.html")
	return os.WriteFile(htmlFile, []byte(htmlContent), 0644)
}

// generateHTMLContent generates HTML content for coverage report
func (ch *CoverageHelper) generateHTMLContent(report *CoverageReport) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Coverage Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .summary { display: flex; gap: 20px; margin-bottom: 20px; }
        .metric { background-color: #e8f4fd; padding: 15px; border-radius: 5px; text-align: center; }
        .metric-value { font-size: 24px; font-weight: bold; color: #2c5282; }
        .metric-label { font-size: 14px; color: #666; }
        .packages { margin-top: 20px; }
        .package { background-color: #f9f9f9; padding: 15px; margin-bottom: 10px; border-radius: 5px; }
        .package-name { font-weight: bold; margin-bottom: 10px; }
        .coverage-bar { width: 100%; height: 20px; background-color: #e0e0e0; border-radius: 10px; overflow: hidden; }
        .coverage-fill { height: 100%; background-color: #4caf50; }
        .low-coverage { background-color: #f44336; }
        .medium-coverage { background-color: #ff9800; }
        .high-coverage { background-color: #4caf50; }
        .uncovered { background-color: #ffebee; padding: 10px; margin-top: 10px; border-radius: 5px; }
        .uncovered-line { font-family: monospace; margin-bottom: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Code Coverage Report</h1>
        <p>Generated on: ` + time.Now().Format(time.RFC3339) + `</p>
    </div>
    
    <div class="summary">
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%.1f%%", report.Overall.Percent) + `</div>
            <div class="metric-label">Overall Coverage</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Summary.TotalPackages) + `</div>
            <div class="metric-label">Total Packages</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Summary.CoveredPackages) + `</div>
            <div class="metric-label">Covered Packages</div>
        </div>
        <div class="metric">
            <div class="metric-value">` + fmt.Sprintf("%d", report.Summary.TotalFunctions) + `</div>
            <div class="metric-label">Total Functions</div>
        </div>
    </div>
    
    <div class="packages">
        <h2>Packages</h2>`
	
	for pkg, metrics := range report.Packages {
		coverageClass := "high-coverage"
		if metrics.Percent < 50 {
			coverageClass = "low-coverage"
		} else if metrics.Percent < 80 {
			coverageClass = "medium-coverage"
		}
		
		html += `
        <div class="package">
            <div class="package-name">` + pkg + `</div>
            <div class="coverage-bar">
                <div class="coverage-fill ` + coverageClass + `" style="width: ` + fmt.Sprintf("%.1f%%", metrics.Percent) + `"></div>
            </div>
            <div>` + fmt.Sprintf("%.1f%% (%d/%d)", metrics.Percent, metrics.Covered, metrics.Total) + `</div>
        </div>`
	}
	
	html += `
    </div>
    
    <div class="uncovered">
        <h2>Uncovered Lines</h2>`
	
	for _, line := range report.Uncovered {
		html += `
        <div class="uncovered-line">
            <strong>` + line.File + `:` + fmt.Sprintf("%d", line.LineNumber) + `</strong> in ` + line.Function + `<br>
            <code>` + line.Line + `</code>
        </div>`
	}
	
	html += `
    </div>
</body>
</html>`
	
	return html
}

// AssertCoverageThresholds asserts that coverage meets thresholds
func (ch *CoverageHelper) AssertCoverageThresholds(report *CoverageReport) {
	// Check overall coverage
	if report.Overall.Percent < ch.thresholds.Overall {
		ch.t.Errorf("Overall coverage %.1f%% is below threshold %.1f%%", 
			report.Overall.Percent, ch.thresholds.Overall)
	}
	
	// Check package-specific thresholds
	for pkg, threshold := range ch.thresholds.Packages {
		if metrics, exists := report.Packages[pkg]; exists {
			if metrics.Percent < threshold {
				ch.t.Errorf("Package %s coverage %.1f%% is below threshold %.1f%%", 
					pkg, metrics.Percent, threshold)
			}
		}
	}
	
	// Check function coverage
	if ch.thresholds.Functions > 0 {
		coveredFunctions := 0
		for _, fn := range report.Functions {
			if fn.Percent > 0 {
				coveredFunctions++
			}
		}
		
		functionCoverage := float64(coveredFunctions) / float64(len(report.Functions)) * 100
		if functionCoverage < ch.thresholds.Functions {
			ch.t.Errorf("Function coverage %.1f%% is below threshold %.1f%%", 
				functionCoverage, ch.thresholds.Functions)
		}
	}
}

// GenerateCoverageBadge generates a coverage badge
func (ch *CoverageHelper) GenerateCoverageBadge(report *CoverageReport) (string, error) {
	percent := report.Overall.Percent
	color := "red"
	if percent >= 80 {
		color = "green"
	} else if percent >= 60 {
		color = "yellow"
	}
	
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="20">
    <rect width="100" height="20" fill="#555"/>
    <rect width="70" height="20" fill="#%s"/>
    <text x="35" y="14" font-family="Arial, sans-serif" font-size="12" fill="white" text-anchor="middle">coverage</text>
    <text x="85" y="14" font-family="Arial, sans-serif" font-size="12" fill="white" text-anchor="middle">%.1f%%</text>
</svg>`, color, percent)
	
	badgeFile := filepath.Join(ch.coverageDir, "coverage.svg")
	err := os.WriteFile(badgeFile, []byte(svg), 0644)
	if err != nil {
		return "", err
	}
	
	return badgeFile, nil
}

// UploadCoverageToCodecov uploads coverage to Codecov
func (ch *CoverageHelper) UploadCoverageToCodecov() error {
	coveragePath := filepath.Join(ch.coverageDir, ch.coverageFile)
	
	// Check if codecov is available
	_, err := exec.LookPath("codecov")
	if err != nil {
		ch.t.Logf("Codecov not found, skipping upload")
		return nil
	}
	
	// Upload coverage
	cmd := exec.Command("codecov", "--file", coveragePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// UploadCoverageToCoveralls uploads coverage to Coveralls
func (ch *CoverageHelper) UploadCoverageToCoveralls() error {
	coveragePath := filepath.Join(ch.coverageDir, ch.coverageFile)
	
	// Check if goveralls is available
	_, err := exec.LookPath("goveralls")
	if err != nil {
		ch.t.Logf("goveralls not found, skipping upload")
		return nil
	}
	
	// Upload coverage
	cmd := exec.Command("goveralls", "-coverprofile", coveragePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// MergeCoverageFiles merges multiple coverage files
func (ch *CoverageHelper) MergeCoverageFiles(files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no coverage files to merge")
	}
	
	// Create merge command
	args := []string{"tool", "cover", "-mode=count", "-o", filepath.Join(ch.coverageDir, "merged.out")}
	args = append(args, files...)
	
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// DiffCoverage compares coverage between two reports
func (ch *CoverageHelper) DiffCoverage(oldReport, newReport *CoverageReport) CoverageDiff {
	diff := CoverageDiff{
		OverallChange:    newReport.Overall.Percent - oldReport.Overall.Percent,
		PackageChanges:   make(map[string]float64),
		AddedFunctions:    make([]string, 0),
		RemovedFunctions:  make([]string, 0),
		ModifiedFunctions: make([]FunctionChange, 0),
	}
	
	// Compare packages
	for pkg, newMetrics := range newReport.Packages {
		if oldMetrics, exists := oldReport.Packages[pkg]; exists {
			diff.PackageChanges[pkg] = newMetrics.Percent - oldMetrics.Percent
		} else {
			diff.PackageChanges[pkg] = newMetrics.Percent
		}
	}
	
	// Compare functions
	oldFunctions := make(map[string]FunctionCoverage)
	for _, fn := range oldReport.Functions {
		key := fn.File + ":" + fn.Name
		oldFunctions[key] = fn
	}
	
	for _, newFn := range newReport.Functions {
		key := newFn.File + ":" + newFn.Name
		if oldFn, exists := oldFunctions[key]; exists {
			if newFn.Percent != oldFn.Percent {
				diff.ModifiedFunctions = append(diff.ModifiedFunctions, FunctionChange{
					Name:    newFn.Name,
					File:    newFn.File,
					OldPercent: oldFn.Percent,
					NewPercent: newFn.Percent,
				})
			}
			delete(oldFunctions, key)
		} else {
			diff.AddedFunctions = append(diff.AddedFunctions, newFn.Name)
		}
	}
	
	for _, oldFn := range oldFunctions {
		diff.RemovedFunctions = append(diff.RemovedFunctions, oldFn.Name)
	}
	
	return diff
}

// CoverageDiff represents a difference between two coverage reports
type CoverageDiff struct {
	OverallChange    float64
	PackageChanges   map[string]float64
	AddedFunctions    []string
	RemovedFunctions  []string
	ModifiedFunctions []FunctionChange
}

// FunctionChange represents a change in function coverage
type FunctionChange struct {
	Name       string
	File       string
	OldPercent float64
	NewPercent float64
}

// PrintCoverageReport prints a coverage report to console
func (ch *CoverageHelper) PrintCoverageReport(report *CoverageReport) {
	fmt.Printf("\n=== Coverage Report ===\n")
	fmt.Printf("Overall Coverage: %.1f%% (%d/%d)\n", 
		report.Overall.Percent, report.Overall.Covered, report.Overall.Total)
	fmt.Printf("Packages: %d/%d covered\n", 
		report.Summary.CoveredPackages, report.Summary.TotalPackages)
	fmt.Printf("Functions: %d/%d covered\n", 
		report.Summary.CoveredFunctions, report.Summary.TotalFunctions)
	fmt.Printf("Lines: %d/%d covered\n", 
		report.Summary.CoveredLines, report.Summary.TotalLines)
	
	fmt.Printf("\n=== Package Coverage ===\n")
	for pkg, metrics := range report.Packages {
		status := "✓"
		if metrics.Percent < ch.thresholds.Overall {
			status = "✗"
		}
		fmt.Printf("%s %s: %.1f%% (%d/%d)\n", 
			status, pkg, metrics.Percent, metrics.Covered, metrics.Total)
	}
	
	if len(report.Uncovered) > 0 {
		fmt.Printf("\n=== Uncovered Lines ===\n")
		for _, line := range report.Uncovered[:10] { // Show first 10
			fmt.Printf("%s:%d in %s\n", line.File, line.LineNumber, line.Function)
		}
		if len(report.Uncovered) > 10 {
			fmt.Printf("... and %d more\n", len(report.Uncovered)-10)
		}
	}
	
	fmt.Printf("\n=== Coverage Thresholds ===\n")
	fmt.Printf("Overall: %.1f%% (threshold: %.1f%%) %s\n", 
		report.Overall.Percent, ch.thresholds.Overall, 
		ch.getStatusIcon(report.Overall.Percent >= ch.thresholds.Overall))
	
	for pkg, threshold := range ch.thresholds.Packages {
		if metrics, exists := report.Packages[pkg]; exists {
			fmt.Printf("%s: %.1f%% (threshold: %.1f%%) %s\n", 
				pkg, metrics.Percent, threshold, 
				ch.getStatusIcon(metrics.Percent >= threshold))
		}
	}
}

// getStatusIcon returns status icon based on condition
func (ch *CoverageHelper) getStatusIcon(passed bool) string {
	if passed {
		return "✓"
	}
	return "✗"
}

// SaveCoverageReport saves coverage report to JSON file
func (ch *CoverageHelper) SaveCoverageReport(report *CoverageReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// LoadCoverageReport loads coverage report from JSON file
func (ch *CoverageHelper) LoadCoverageReport(filename string) (*CoverageReport, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var report CoverageReport
	err = json.Unmarshal(data, &report)
	if err != nil {
		return nil, err
	}
	
	return &report, nil
}

// Cleanup cleans up coverage files
func (ch *CoverageHelper) Cleanup() {
	os.RemoveAll(ch.coverageDir)
	os.RemoveAll(ch.htmlReportDir)
}