package main

import (
	"github.com/vaishbgl/specula/modules"
)

// ModuleWeight defines the weight of each module in the overall grade.
type ModuleWeight struct {
	Name   string
	Weight float64 // 0.0 - 1.0
}

// defaultWeights defines the scoring distribution.
var defaultWeights = []ModuleWeight{
	{"Secret Scanner", 0.25},
	{"Dependency Audit", 0.20},
	{"Test Readiness", 0.20},
	{"Code & Context Lint", 0.15},
	{"Project Setup", 0.10},
	{"Docs & Env Check", 0.10},
}

// ModuleScore holds the computed score for a single module.
type ModuleScore struct {
	Name     string  `json:"name"`
	Score    int     `json:"score"`
	Status   string  `json:"status"` // PASS, WARN, FAIL, SKIPPED
	Findings int     `json:"findings"`
	TopMsg   string  `json:"top_finding"`
	Weight   float64 `json:"weight"`
}

// GradeReport is the full scored output.
type GradeReport struct {
	Overall      int           `json:"overall_score"`
	Grade        string        `json:"grade"`
	ModuleScores []ModuleScore `json:"modules"`
	TotalFails   int           `json:"total_fails"`
	TotalWarns   int           `json:"total_warns"`
	TotalInfos   int           `json:"total_infos"`
}

// ComputeGrade computes scores for all module results.
func ComputeGrade(results []modules.ModuleResult) GradeReport {
	report := GradeReport{}

	weightMap := make(map[string]float64)
	for _, w := range defaultWeights {
		weightMap[w.Name] = w.Weight
	}

	totalWeight := 0.0
	weightedSum := 0.0

	for _, r := range results {
		ms := scoreModule(r)

		// Get weight (default to 0 if unknown module)
		weight, exists := weightMap[r.Name]
		if !exists {
			weight = 0.0
		}
		ms.Weight = weight

		report.ModuleScores = append(report.ModuleScores, ms)

		if ms.Status != "SKIPPED" {
			totalWeight += weight
			weightedSum += float64(ms.Score) * weight
		}

		// Aggregate severity counts
		f, w, i := r.CountBySeverity()
		report.TotalFails += f
		report.TotalWarns += w
		report.TotalInfos += i
	}

	// Compute overall score
	if totalWeight > 0 {
		report.Overall = int(weightedSum / totalWeight)
	}

	// Clamp
	if report.Overall > 100 {
		report.Overall = 100
	}
	if report.Overall < 0 {
		report.Overall = 0
	}

	report.Grade = letterGrade(report.Overall)

	return report
}

// scoreModule computes the score for a single module.
func scoreModule(r modules.ModuleResult) ModuleScore {
	ms := ModuleScore{
		Name:     r.Name,
		Findings: len(r.Findings),
	}

	if len(r.Findings) == 0 {
		ms.Score = 100
		ms.Status = "PASS"
		ms.TopMsg = "No issues"
		return ms
	}

	// Aggregate deductions by rule to prevent one noisy rule from tanking score
	ruleDeductions := make(map[string]int)
	hasFail := false
	topMsg := ""

	for _, f := range r.Findings {
		if topMsg == "" {
			topMsg = f.Message
		}

		var deduction int
		var maxDeduction int

		switch f.Severity {
		case modules.SeverityFail:
			deduction = 25
			maxDeduction = 25
			hasFail = true
		case modules.SeverityWarn:
			deduction = 10
			maxDeduction = 20
		case modules.SeverityInfo:
			deduction = 3
			maxDeduction = 9
		}

		current := ruleDeductions[f.Rule]
		newVal := current + deduction
		if newVal > maxDeduction {
			newVal = maxDeduction
		}
		ruleDeductions[f.Rule] = newVal
	}

	// Sum all deductions
	totalDeduction := 0
	for _, d := range ruleDeductions {
		totalDeduction += d
	}

	// Floor at 25
	ms.Score = 100 - totalDeduction
	if ms.Score < 25 {
		ms.Score = 25
	}

	// Determine status
	if hasFail {
		ms.Status = "FAIL"
	} else if ms.Score < 75 {
		ms.Status = "WARN"
	} else {
		ms.Status = "PASS"
	}

	ms.TopMsg = topMsg
	return ms
}

// letterGrade converts a numeric score to a letter grade.
func letterGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 75:
		return "B"
	case score >= 60:
		return "C"
	case score >= 45:
		return "D"
	default:
		return "F"
	}
}
