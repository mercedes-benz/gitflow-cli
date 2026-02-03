package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type SetupPyManager struct {
	projectPath string
	filePath    string
}

func NewSetupPyManager(projectPath string) (VersionManager, error) {
	filePath := filepath.Join(projectPath, "setup.py")
	if _, err := os.Stat(filePath); err != nil {
		return nil, fmt.Errorf("setup.py not found at %s", projectPath)
	}

	return &SetupPyManager{
		projectPath: projectPath,
		filePath:    filePath,
	}, nil
}

func (s *SetupPyManager) GetVersion() (string, error) {
	content, err := os.ReadFile(s.filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read setup.py: %v", err)
	}

	// Match version="..." or version='...'
	re := regexp.MustCompile(`version\s*=\s*["']([^"']+)["']`)
	matches := re.FindStringSubmatch(string(content))

	if len(matches) < 2 {
		return "", fmt.Errorf("version not found in setup.py")
	}

	return matches[1], nil
}

func (s *SetupPyManager) SetVersion(version string) error {
	content, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read setup.py: %v", err)
	}

	// Match version="..." or version='...' and preserve quote style
	re := regexp.MustCompile(`(version\s*=\s*)(["'])([^"']+)(["'])`)

	newContent := re.ReplaceAllString(string(content), fmt.Sprintf(`${1}${2}%s${4}`, version))

	if newContent == string(content) {
		return fmt.Errorf("version not found in setup.py")
	}

	err = os.WriteFile(s.filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write setup.py: %v", err)
	}

	return nil
}

func (s *SetupPyManager) GetName() string {
	return "setup.py"
}

func (s *SetupPyManager) GetFilePath() string {
	return filepath.Base(s.filePath)
}
