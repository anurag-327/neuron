package registry

import (
	"fmt"
)

type LanguageConfig struct {
	Name string

	// API layer
	Validator func(code string) error

	// Execution layer
	DockerImage string
	BaseName    string
	Ext         string
	RunCmd      func(n FileNames) string

	// Billing
	CreditCost int64
}

type FileNames struct {
	BaseName string // "main" or "Main"
	FullName string // "main.cpp"
	PathBase string // "/host/job/main"
	PathFull string // "/host/job/main.cpp"
}

var LanguageRegistry = map[string]LanguageConfig{
	"cpp": {
		Name:        "cpp",
		Validator:   nil,
		DockerImage: "gcc:latest",
		BaseName:    "main",
		Ext:         "cpp",
		RunCmd: func(n FileNames) string {
			return fmt.Sprintf(
				"g++ %s -o %s && ./%s < input.txt",
				n.FullName, n.BaseName, n.BaseName,
			)
		},
		CreditCost: 5,
	},

	"go": {
		Name:        "go",
		Validator:   nil,
		DockerImage: "golang:1.23-alpine",
		BaseName:    "main",
		Ext:         "go",
		RunCmd: func(n FileNames) string {
			return fmt.Sprintf(
				"go build -o %s %s && ./%s",
				n.BaseName, n.FullName, n.BaseName,
			)
		},
		CreditCost: 4,
	},

	"python": {
		Name:        "python",
		Validator:   nil,
		DockerImage: "python:3.12-alpine",
		BaseName:    "main",
		Ext:         "py",
		RunCmd: func(n FileNames) string {
			return fmt.Sprintf("python3 %s < input.txt", n.FullName)
		},
		CreditCost: 6,
	},

	"java": {
		Name:        "java",
		Validator:   ValidateAndSanitizeJava,
		DockerImage: "eclipse-temurin:21-jdk-alpine",
		BaseName:    "Main",
		Ext:         "java",
		RunCmd: func(n FileNames) string {
			return fmt.Sprintf(
				"javac %s && java %s < input.txt",
				n.FullName, n.BaseName,
			)
		},
		CreditCost: 7,
	},

	"javascript": {
		Name:        "javascript",
		Validator:   nil,
		DockerImage: "node:22-alpine",
		BaseName:    "main",
		Ext:         "js",
		RunCmd: func(n FileNames) string {
			return fmt.Sprintf("node %s < input.txt", n.FullName)
		},
		CreditCost: 5,
	},
}
