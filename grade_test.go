package main

import (
	"testing"

	"github.com/vaishbgl/specula/modules"
)

func TestComputeGrade_AllClean(t *testing.T) {
	results := []modules.ModuleResult{
		{Name: "Secret Scanner", Findings: nil},
		{Name: "Dependency Audit", Findings: nil},
		{Name: "Test Readiness", Findings: nil},
		{Name: "Code & Context Lint", Findings: nil},
	}

	report := ComputeGrade(results)
	if report.Overall != 100 {
		t.Errorf("expected overall 100 for clean project, got %d", report.Overall)
	}
	if report.Grade != "A" {
		t.Errorf("expected grade 'A', got %q", report.Grade)
	}
	if report.TotalFails != 0 || report.TotalWarns != 0 || report.TotalInfos != 0 {
		t.Errorf("expected 0 issues, got fails=%d warns=%d infos=%d",
			report.TotalFails, report.TotalWarns, report.TotalInfos)
	}
}

func TestComputeGrade_WithDeductions(t *testing.T) {
	results := []modules.ModuleResult{
		{
			Name: "Secret Scanner",
			Findings: []modules.Finding{
				{Rule: "secret-pattern", Severity: modules.SeverityFail, Message: "AWS key found"},
			},
		},
		{
			Name: "Code & Context Lint",
			Findings: []modules.Finding{
				{Rule: "todo-markers", Severity: modules.SeverityWarn, Message: "TODO found"},
				{Rule: "todo-markers", Severity: modules.SeverityWarn, Message: "TODO 2 found"},
			},
		},
	}

	report := ComputeGrade(results)

	// Secret scanner: 100 - 25 = 75 (FAIL status)
	// Code lint: 100 - 20 (capped at 20 for multiple warns of same rule) = 80 (PASS status)
	// Overall should be lower than 100
	if report.Overall >= 100 {
		t.Errorf("expected overall score < 100, got %d", report.Overall)
	}
	if report.TotalFails != 1 {
		t.Errorf("expected 1 fail, got %d", report.TotalFails)
	}
	if report.TotalWarns != 2 {
		t.Errorf("expected 2 warns, got %d", report.TotalWarns)
	}
}

func TestScoreModule_FloorAt25(t *testing.T) {
	// A module with 10 distinct failure rules
	var findings []modules.Finding
	for i := 0; i < 10; i++ {
		findings = append(findings, modules.Finding{
			Rule:     string(rune('a' + i)),
			Severity: modules.SeverityFail,
			Message:  "critical issue",
		})
	}

	res := modules.ModuleResult{
		Name:     "Secret Scanner",
		Findings: findings,
	}

	ms := scoreModule(res)
	if ms.Score != 25 {
		t.Errorf("expected score to floor at 25, got %d", ms.Score)
	}
	if ms.Status != "FAIL" {
		t.Errorf("expected status 'FAIL', got %q", ms.Status)
	}
}

func TestLetterGrade(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{95, "A"},
		{90, "A"},
		{85, "B"},
		{75, "B"},
		{70, "C"},
		{60, "C"},
		{55, "D"},
		{45, "D"},
		{40, "F"},
		{0, "F"},
	}

	for _, tc := range tests {
		got := letterGrade(tc.score)
		if got != tc.want {
			t.Errorf("letterGrade(%d) = %q, want %q", tc.score, got, tc.want)
		}
	}
}
