package eqemuserver

import (
	"fmt"
	"github.com/EQEmu/spire/internal/env"
	"github.com/EQEmu/spire/internal/filepathcheck"
	"github.com/EQEmu/spire/internal/logger"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type QuestEditorService struct {
	logger    *logger.AppLogger
	pathmgmt  *pathmgmt.PathManagement
	serverApi *Client
}

func NewQuestEditorService(
	logger *logger.AppLogger,
	pathmgmt *pathmgmt.PathManagement,
	serverApi *Client,
) *QuestEditorService {
	return &QuestEditorService{
		logger:    logger,
		pathmgmt:  pathmgmt,
		serverApi: serverApi,
	}
}

type QuestFileInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Extension    string `json:"extension"`
	QuestScope   string `json:"quest_scope"`
	Size         int64  `json:"size"`
	ModifiedTime int64  `json:"modified_time"`
	IsDirectory  bool   `json:"is_directory"`
}

func (s *QuestEditorService) GetQuestsDir() string {
	return s.pathmgmt.GetQuestsDir()
}

func (s *QuestEditorService) ensureLocalEnvironment() error {
	if !env.IsAppEnvLocal() {
		return fmt.Errorf("quest editor is only available in local environments")
	}
	return nil
}

func (s *QuestEditorService) cleanRelativePath(relativePath string) (string, error) {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	cleaned := filepath.Clean(relativePath)
	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("path is required")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected")
	}

	for _, part := range strings.Split(cleaned, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		if filepathcheck.IsHiddenFile(part) {
			return "", fmt.Errorf("hidden files are not allowed")
		}
	}

	return cleaned, nil
}

func (s *QuestEditorService) ensureLuaFilePath(relativePath string) error {
	if strings.ToLower(filepath.Ext(relativePath)) != ".lua" {
		return fmt.Errorf("only .lua files are supported")
	}
	return nil
}

func (s *QuestEditorService) validateNoSymlinkEscape(questsDir string, fullPath string) error {
	absQuestsDir, err := filepath.Abs(questsDir)
	if err != nil {
		return fmt.Errorf("failed to resolve quest root: %v", err)
	}

	current := absQuestsDir
	relToRoot, err := filepath.Rel(absQuestsDir, fullPath)
	if err != nil {
		return fmt.Errorf("failed to resolve relative path: %v", err)
	}
	if relToRoot == "." {
		return nil
	}

	for _, part := range strings.Split(relToRoot, string(filepath.Separator)) {
		current = filepath.Join(current, part)

		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if info.Mode()&os.ModeSymlink == 0 {
			continue
		}

		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			return fmt.Errorf("failed to evaluate symlink: %v", err)
		}
		absResolved, err := filepath.Abs(resolved)
		if err != nil {
			return fmt.Errorf("failed to resolve symlink target: %v", err)
		}
		if absResolved != absQuestsDir && !strings.HasPrefix(absResolved, absQuestsDir+string(filepath.Separator)) {
			return fmt.Errorf("path is outside quest root")
		}
	}

	return nil
}

func (s *QuestEditorService) validateQuestPath(relativePath string) (string, error) {
	if err := s.ensureLocalEnvironment(); err != nil {
		return "", err
	}

	questsDir := s.pathmgmt.GetQuestsDir()
	if questsDir == "" {
		return "", fmt.Errorf("quests directory is not configured")
	}

	cleaned, err := s.cleanRelativePath(relativePath)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(questsDir, cleaned)
	if err := filepathcheck.ValidateSafePath(questsDir, fullPath); err != nil {
		return "", err
	}

	absQuestsDir, err := filepath.Abs(questsDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve quest root: %v", err)
	}
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve full path: %v", err)
	}
	if !strings.HasPrefix(absFullPath, absQuestsDir+string(filepath.Separator)) && absFullPath != absQuestsDir {
		return "", fmt.Errorf("path is outside quest root")
	}

	if err := s.validateNoSymlinkEscape(absQuestsDir, absFullPath); err != nil {
		return "", err
	}

	return cleaned, nil
}

func (s *QuestEditorService) resolvePath(relativePath string) (string, error) {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.pathmgmt.GetQuestsDir(), cleaned), nil
}

func (s *QuestEditorService) determineScope(relPath string) string {
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) == 0 {
		return ""
	}
	top := parts[0]
	switch top {
	case "global", "plugins", "lua_modules":
		return top
	default:
		return "zone"
	}
}

func (s *QuestEditorService) GetTree() ([]QuestFileInfo, error) {
	if err := s.ensureLocalEnvironment(); err != nil {
		return nil, err
	}

	questsDir := s.pathmgmt.GetQuestsDir()
	if questsDir == "" {
		return nil, fmt.Errorf("quests directory is not configured")
	}

	if _, err := os.Stat(questsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("quests directory does not exist")
	}

	var files []QuestFileInfo
	err := filepath.Walk(questsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, relErr := filepath.Rel(questsDir, path)
		if relErr != nil {
			return nil
		}
		if relPath == "." {
			return nil
		}

		if filepathcheck.IsHiddenFile(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if info.IsDir() {
			files = append(files, QuestFileInfo{
				Name:         info.Name(),
				Path:         relPath,
				RelativePath: relPath,
				Extension:    "",
				QuestScope:   s.determineScope(relPath),
				Size:         0,
				ModifiedTime: info.ModTime().Unix(),
				IsDirectory:  true,
			})
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext != ".lua" {
			return nil
		}

		files = append(files, QuestFileInfo{
			Name:         info.Name(),
			Path:         relPath,
			RelativePath: relPath,
			Extension:    ext,
			QuestScope:   s.determineScope(relPath),
			Size:         info.Size(),
			ModifiedTime: info.ModTime().Unix(),
			IsDirectory:  false,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDirectory != files[j].IsDirectory {
			return files[i].IsDirectory
		}
		return strings.ToLower(files[i].Path) < strings.ToLower(files[j].Path)
	})

	return files, nil
}

func (s *QuestEditorService) GetFile(relativePath string) (string, error) {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return "", err
	}
	if err := s.ensureLuaFilePath(cleaned); err != nil {
		return "", err
	}

	fullPath, err := s.resolvePath(relativePath)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file does not exist")
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (s *QuestEditorService) SaveFile(relativePath string, content string) error {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return err
	}
	if err := s.ensureLuaFilePath(cleaned); err != nil {
		return err
	}

	fullPath, err := s.resolvePath(relativePath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	tmpFile := fullPath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := os.Rename(tmpFile, fullPath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

type CreateFileRequest struct {
	RelativePath string `json:"relative_path"`
	Content      string `json:"content"`
	IsDirectory  bool   `json:"is_directory"`
}

func (s *QuestEditorService) CreateFile(req CreateFileRequest) error {
	cleaned, err := s.validateQuestPath(req.RelativePath)
	if err != nil {
		return err
	}
	if !req.IsDirectory {
		if err := s.ensureLuaFilePath(cleaned); err != nil {
			return err
		}
	}

	fullPath, err := s.resolvePath(req.RelativePath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("path already exists")
	}

	if req.IsDirectory {
		return os.MkdirAll(fullPath, 0755)
	}

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	return os.WriteFile(fullPath, []byte(req.Content), 0644)
}

type MovePathRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

func (s *QuestEditorService) MovePath(req MovePathRequest) error {
	cleanOld, err := s.validateQuestPath(req.OldPath)
	if err != nil {
		return fmt.Errorf("invalid source path: %v", err)
	}
	cleanNew, err := s.validateQuestPath(req.NewPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %v", err)
	}

	oldFull, err := s.resolvePath(req.OldPath)
	if err != nil {
		return fmt.Errorf("invalid source path: %v", err)
	}
	newFull, err := s.resolvePath(req.NewPath)
	if err != nil {
		return fmt.Errorf("invalid destination path: %v", err)
	}

	oldInfo, err := os.Stat(oldFull)
	if os.IsNotExist(err) {
		return fmt.Errorf("source path does not exist")
	}
	if err != nil {
		return err
	}

	if !oldInfo.IsDir() {
		if err := s.ensureLuaFilePath(cleanOld); err != nil {
			return err
		}
	}

	if newInfo, err := os.Stat(newFull); err == nil && newInfo.IsDir() {
		newFull = filepath.Join(newFull, filepath.Base(oldFull))
		cleanNew = filepath.Join(cleanNew, filepath.Base(cleanOld))
	}
	if !oldInfo.IsDir() {
		if err := s.ensureLuaFilePath(cleanNew); err != nil {
			return err
		}
	}

	newDir := filepath.Dir(newFull)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	return os.Rename(oldFull, newFull)
}

func (s *QuestEditorService) DeletePath(relativePath string) error {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return err
	}

	fullPath, err := s.resolvePath(relativePath)
	if err != nil {
		return err
	}

	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist")
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if err := s.ensureLuaFilePath(cleaned); err != nil {
			return err
		}
	}

	return os.RemoveAll(fullPath)
}

type FormatResult struct {
	Formatted string `json:"formatted"`
	Error     string `json:"error,omitempty"`
}

func (s *QuestEditorService) FormatFile(relativePath string, content string) (*FormatResult, error) {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureLuaFilePath(cleaned); err != nil {
		return nil, err
	}

	_, err = s.resolvePath(relativePath)
	if err != nil {
		return nil, err
	}

	styluaPath, _ := exec.LookPath("stylua")
	if styluaPath == "" {
		return nil, fmt.Errorf("stylua is not installed or not on PATH")
	}

	cmd := exec.Command(styluaPath, "-")
	cmd.Stdin = strings.NewReader(content)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &FormatResult{
			Formatted: content,
			Error:     fmt.Sprintf("stylua error: %v\n%s", err, stderr.String()),
		}, nil
	}

	return &FormatResult{
		Formatted: stdout.String(),
	}, nil
}

type ValidateResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

func (s *QuestEditorService) ValidateFile(relativePath string, content string) (*ValidateResult, error) {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return nil, err
	}
	if err := s.ensureLuaFilePath(cleaned); err != nil {
		return nil, err
	}

	_, err = s.resolvePath(relativePath)
	if err != nil {
		return nil, err
	}

	luacPath, _ := exec.LookPath("luac")
	if luacPath == "" {
		return nil, fmt.Errorf("luac is not installed or not on PATH")
	}

	tmpFile, err := os.CreateTemp("", "quest-validate-*.lua")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command(luacPath, "-p", tmpFile.Name())
	var stderr strings.Builder
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return &ValidateResult{
			Valid:  false,
			Errors: []string{strings.TrimSpace(stderr.String())},
		}, nil
	}

	return &ValidateResult{
		Valid: true,
	}, nil
}

type Capabilities struct {
	FormatAvailable   bool `json:"format_available"`
	ValidateAvailable bool `json:"validate_available"`
}

func (s *QuestEditorService) GetCapabilities() Capabilities {
	_, fmtOk := exec.LookPath("stylua")
	_, valOk := exec.LookPath("luac")
	return Capabilities{
		FormatAvailable:   fmtOk == nil,
		ValidateAvailable: valOk == nil,
	}
}

func (s *QuestEditorService) ReloadQuestsForPath(relativePath string) error {
	cleaned, err := s.validateQuestPath(relativePath)
	if err != nil {
		return err
	}
	if err := s.ensureLuaFilePath(cleaned); err != nil {
		return err
	}
	if s.serverApi == nil {
		return fmt.Errorf("server API is not available")
	}

	relativePath = filepath.ToSlash(cleaned)
	parts := strings.Split(relativePath, "/")

	reloadTarget := "all"

	if len(parts) >= 1 {
		topDir := parts[0]
		matched, _ := regexp.MatchString(`^(global|plugins|lua_modules)$`, topDir)
		if matched {
			reloadTarget = "all"
		} else {
			reloadTarget = topDir
		}
	}

	if reloadTarget == "all" {
		return s.serverApi.ReloadQuestsForZone("all")
	}
	return s.serverApi.ReloadQuestsForZone(reloadTarget)
}
