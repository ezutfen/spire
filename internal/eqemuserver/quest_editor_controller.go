package eqemuserver

import (
	"fmt"
	"github.com/EQEmu/spire/internal/env"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/labstack/echo/v4"
	"net/http"
)

type QuestEditorController struct {
	service *QuestEditorService
}

func NewQuestEditorController(service *QuestEditorService) *QuestEditorController {
	return &QuestEditorController{service: service}
}

func (ctrl *QuestEditorController) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "eqemuserver/quests/capabilities", ctrl.getCapabilities, nil),
		routes.RegisterRoute(http.MethodGet, "eqemuserver/quests/tree", ctrl.getTree, nil),
		routes.RegisterRoute(http.MethodGet, "eqemuserver/quests/file", ctrl.getFile, nil),
		routes.RegisterRoute(http.MethodPut, "eqemuserver/quests/file", ctrl.saveFile, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/file/create", ctrl.createFile, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/folder/create", ctrl.createFolder, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/path/move", ctrl.movePath, nil),
		routes.RegisterRoute(http.MethodDelete, "eqemuserver/quests/path", ctrl.deletePath, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/file/format", ctrl.formatFile, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/file/validate", ctrl.validateFile, nil),
		routes.RegisterRoute(http.MethodPost, "eqemuserver/quests/file/reload", ctrl.reloadFile, nil),
	}
}

func (ctrl *QuestEditorController) ensureLocal(c echo.Context) error {
	if !env.IsAppEnvLocal() {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Quest Editor is only available in local environments"})
	}
	return nil
}

func (ctrl *QuestEditorController) getCapabilities(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, echo.Map{
		"data": ctrl.service.GetCapabilities(),
	})
}

func (ctrl *QuestEditorController) getTree(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	tree, err := ctrl.service.GetTree()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": tree})
}

func (ctrl *QuestEditorController) getFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	relativePath := c.QueryParam("path")
	if relativePath == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path query parameter is required"})
	}

	content, err := ctrl.service.GetFile(relativePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": echo.Map{
			"path":     relativePath,
			"contents": content,
		},
	})
}

type SaveFileRequest struct {
	Path     string `json:"path"`
	Contents string `json:"contents"`
}

func (ctrl *QuestEditorController) saveFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(SaveFileRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path is required"})
	}

	if err := ctrl.service.SaveFile(req.Path, req.Contents); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "File saved successfully"})
}

type CreateRequest struct {
	RelativePath string `json:"relative_path"`
	Content      string `json:"content"`
}

func (ctrl *QuestEditorController) createFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(CreateRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.RelativePath == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "relative_path is required"})
	}

	if err := ctrl.service.CreateFile(CreateFileRequest{
		RelativePath: req.RelativePath,
		Content:      req.Content,
		IsDirectory:  false,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "File created successfully"})
}

type CreateFolderRequest struct {
	RelativePath string `json:"relative_path"`
}

func (ctrl *QuestEditorController) createFolder(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(CreateFolderRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.RelativePath == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "relative_path is required"})
	}

	if err := ctrl.service.CreateFile(CreateFileRequest{
		RelativePath: req.RelativePath,
		IsDirectory:  true,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Folder created successfully"})
}

type MoveRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

func (ctrl *QuestEditorController) movePath(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(MoveRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.OldPath == "" || req.NewPath == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "old_path and new_path are required"})
	}

	if err := ctrl.service.MovePath(MovePathRequest{
		OldPath: req.OldPath,
		NewPath: req.NewPath,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Path moved successfully"})
}

func (ctrl *QuestEditorController) deletePath(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	relativePath := c.QueryParam("path")
	if relativePath == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path query parameter is required"})
	}

	if err := ctrl.service.DeletePath(relativePath); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Path deleted successfully"})
}

type FormatRequest struct {
	Path     string `json:"path"`
	Contents string `json:"contents"`
}

func (ctrl *QuestEditorController) formatFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(FormatRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path is required"})
	}

	result, err := ctrl.service.FormatFile(req.Path, req.Contents)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": result})
}

type ValidateRequest struct {
	Path     string `json:"path"`
	Contents string `json:"contents"`
}

func (ctrl *QuestEditorController) validateFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(ValidateRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path is required"})
	}

	result, err := ctrl.service.ValidateFile(req.Path, req.Contents)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": result})
}

type ReloadRequest struct {
	Path string `json:"path"`
}

func (ctrl *QuestEditorController) reloadFile(c echo.Context) error {
	if err := ctrl.ensureLocal(c); err != nil {
		return err
	}
	req := new(ReloadRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if req.Path == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "path is required"})
	}

	if err := ctrl.service.ReloadQuestsForPath(req.Path); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": fmt.Sprintf("Failed to reload quests: %v", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Quests reloaded successfully"})
}
