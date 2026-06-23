package judge

import (
	"fmt"
	"strings"
)

// SanitizationError is returned when forbidden patterns are found in user code.
type SanitizationError struct {
	Language string
	Pattern  string
}

func (e *SanitizationError) Error() string {
	return fmt.Sprintf("forbidden pattern '%s' detected in %s code", e.Pattern, e.Language)
}

// forbiddenPatterns maps each language to its list of blocked patterns.
// This is the first line of defense (static analysis).
// Docker provides the second line (runtime isolation).
var forbiddenPatterns = map[string][]string{
	"python": {
		"import os", "import subprocess", "import shutil",
		"import socket", "import http", "import urllib",
		"from os", "from subprocess", "from shutil",
		"from socket", "from http", "from urllib",
		"open(", "exec(", "eval(", "__import__",
		"os.system", "os.popen",
	},
	"javascript": {
		"require('child_process')", "require(\"child_process\")",
		"require('fs')", "require(\"fs\")",
		"require('net')", "require(\"net\")",
		"require('http')", "require(\"http\")",
		"process.exit", "eval(",
		"execSync", "spawnSync",
	},
	"cpp": {
		"system(", "popen(", "execvp(", "execl(",
		"#include <fstream>",
		"#include <filesystem>",
		"#include <sys/socket.h>",
		"#include <netinet/",
		"#include <arpa/",
		"fork(",
	},
	"go": {
		"\"os/exec\"", "\"syscall\"",
		"\"net/", "\"net\"",
		"\"plugin\"",
		"os.Exit",
	},
	"java": {
		"Runtime.exec", "Runtime.getRuntime().exec",
		"ProcessBuilder",
		"System.exit",
		"java.net.", "java.io.File",
		"java.lang.reflect",
	},
}

// SupportedLanguages returns the list of languages the judge can handle.
func SupportedLanguages() []string {
	langs := make([]string, 0, len(forbiddenPatterns))
	for lang := range forbiddenPatterns {
		langs = append(langs, lang)
	}
	return langs
}

// IsSupported checks if a language is supported by the judge.
func IsSupported(language string) bool {
	_, ok := forbiddenPatterns[language]
	return ok
}

// Sanitize checks user code for dangerous patterns.
// Returns nil if the code is safe, or a SanitizationError if a forbidden pattern is found.
func Sanitize(language, code string) error {
	patterns, ok := forbiddenPatterns[language]
	if !ok {
		return fmt.Errorf("unsupported language: %s", language)
	}

	for _, pattern := range patterns {
		if strings.Contains(code, pattern) {
			return &SanitizationError{
				Language: language,
				Pattern:  pattern,
			}
		}
	}

	return nil
}
