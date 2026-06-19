package aaeditor

import (
	"fmt"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type AaEditorController struct {
	service *AaEditorService
}

func NewAaEditorController(service *AaEditorService) *AaEditorController {
	return &AaEditorController{service: service}
}

func (ctrl *AaEditorController) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "aa_editor/abilities", ctrl.listAbilities, nil),
		routes.RegisterRoute(http.MethodGet, "aa_editor/abilities/:id", ctrl.getAbility, nil),
		routes.RegisterRoute(http.MethodGet, "aa_editor/metadata", ctrl.getMetadata, nil),
		routes.RegisterRoute(http.MethodPut, "aa_editor/abilities", ctrl.createAbility, nil),
		routes.RegisterRoute(http.MethodPost, "aa_editor/export", ctrl.exportAbilities, nil),
		routes.RegisterRoute(http.MethodPost, "aa_editor/import/preview", ctrl.previewImport, nil),
		routes.RegisterRoute(http.MethodPost, "aa_editor/import/apply", ctrl.applyImport, nil),
		routes.RegisterRoute(http.MethodPost, "aa_editor/abilities/:id/duplicate", ctrl.duplicateAbility, nil),
		routes.RegisterRoute(http.MethodPatch, "aa_editor/abilities/:id", ctrl.saveAbility, nil),
		routes.RegisterRoute(http.MethodDelete, "aa_editor/abilities/:id", ctrl.deleteAbility, nil),
		routes.RegisterRoute(http.MethodPost, "aa_editor/ranks/preview", ctrl.previewRank, nil),
	}
}

func parseId(c echo.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

// listAaAbilities godoc
// @Id listAaEditorAbilities
// @Summary Lists AA abilities
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param search query string false "Name or id search"
// @Param category query int false "Category filter"
// @Param classes query int false "Class bitmask filter"
// @Param enabled query int false "Enabled filter (-1 any, 0 disabled, 1 enabled)"
// @Param page query int false "Pagination page"
// @Param limit query int false "Rows per page"
// @Param orderBy query string false "Order by field"
// @Param orderDirection query string false "Order direction"
// @Success 200 {object} AaAbilityListResult
// @Router /aa_editor/abilities [get]
func (ctrl *AaEditorController) listAbilities(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	category, _ := strconv.Atoi(c.QueryParam("category"))
	classes, _ := strconv.Atoi(c.QueryParam("classes"))
	enabled, err := strconv.Atoi(c.QueryParam("enabled"))
	if err != nil {
		enabled = -1
	}

	filter := AaListFilter{
		Search:   c.QueryParam("search"),
		Category: category,
		Classes:  classes,
		Enabled:  enabled,
		Page:     page,
		Limit:    limit,
		OrderBy:  c.QueryParam("orderBy"),
		OrderDir: c.QueryParam("orderDirection"),
	}

	result, err := ctrl.service.ListAbilities(c, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": result})
}

// getAaAbility godoc
// @Id getAaEditorAbilityFull
// @Summary Gets the full AA ability tree (ability + ranks + effects + prereqs + strings + spell)
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param id path int true "Id"
// @Success 200 {object} AaAbilityFull
// @Router /aa_editor/abilities/{id} [get]
func (ctrl *AaEditorController) getAbility(c echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	full, err := ctrl.service.GetFullAbility(c, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": full})
}

// getAaMetadata godoc
// @Id getAaEditorMetadata
// @Summary Gets AA editor metadata / constants for the GUI
// @Produce json
// @Tags AaEditor
// @Success 200 {object} AaMetadata
// @Router /aa_editor/metadata [get]
func (ctrl *AaEditorController) getMetadata(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"data": ctrl.service.GetMetadata()})
}

// createAaAbility godoc
// @Id createAaEditorAbilityFull
// @Summary Creates a full AA ability (ability + ranks + effects + prereqs + strings)
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param body body AaAbilityInput true "Ability tree"
// @Success 200 {object} AaAbilityFull
// @Router /aa_editor/abilities [put]
func (ctrl *AaEditorController) createAbility(c echo.Context) error {
	input := new(AaAbilityInput)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	full, err := ctrl.service.CreateFullAbility(c, *input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": full})
}

// exportAaBundle godoc
// @Id exportAaEditorBundle
// @Summary Exports selected AA abilities as a portable JSON bundle
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param body body AaExportRequest true "Ability ids"
// @Success 200 {object} AaExportBundle
// @Router /aa_editor/export [post]
func (ctrl *AaEditorController) exportAbilities(c echo.Context) error {
	req := new(AaExportRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	bundle, err := ctrl.service.ExportAbilities(c, *req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": bundle})
}

// previewAaImport godoc
// @Id previewAaEditorImport
// @Summary Previews an AA bundle import against the active connection
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param body body AaExportBundle true "AA export bundle"
// @Success 200 {object} AaImportPreviewResult
// @Router /aa_editor/import/preview [post]
func (ctrl *AaEditorController) previewImport(c echo.Context) error {
	bundle := new(AaExportBundle)
	if err := c.Bind(bundle); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	result, err := ctrl.service.PreviewImport(c, *bundle)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": result})
}

// applyAaImport godoc
// @Id applyAaEditorImport
// @Summary Applies an AA bundle import to the active connection
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param body body AaExportBundle true "AA export bundle"
// @Success 200 {object} AaImportApplyResult
// @Router /aa_editor/import/apply [post]
func (ctrl *AaEditorController) applyImport(c echo.Context) error {
	bundle := new(AaExportBundle)
	if err := c.Bind(bundle); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	result, err := ctrl.service.ApplyImport(c, *bundle)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": result})
}

// duplicateAaAbility godoc
// @Id duplicateAaEditorAbility
// @Summary Duplicates an existing AA ability (deep copy with re-linked rank chain)
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param id path int true "Source ability id"
// @Param body body DuplicateOptions true "Duplicate options"
// @Success 200 {object} AaAbilityFull
// @Router /aa_editor/abilities/{id}/duplicate [post]
func (ctrl *AaEditorController) duplicateAbility(c echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	opts := new(DuplicateOptions)
	_ = c.Bind(opts)
	full, err := ctrl.service.DuplicateAbility(c, id, *opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": full})
}

// saveAaAbility godoc
// @Id saveAaEditorAbilityFull
// @Summary Saves (full replace) an AA ability tree
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param id path int true "Ability id"
// @Param body body AaAbilityInput true "Ability tree"
// @Success 200 {object} AaAbilityFull
// @Router /aa_editor/abilities/{id} [patch]
func (ctrl *AaEditorController) saveAbility(c echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	input := new(AaAbilityInput)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	full, err := ctrl.service.SaveFullAbility(c, id, *input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"data": full})
}

// deleteAaAbility godoc
// @Id deleteAaEditorAbilityFull
// @Summary Deletes an AA ability and cascades to ranks/effects/prereqs (strings/spells preserved)
// @Produce json
// @Tags AaEditor
// @Param id path int true "Ability id"
// @Success 200 {string} string "Entity deleted successfully"
// @Router /aa_editor/abilities/{id} [delete]
func (ctrl *AaEditorController) deleteAbility(c echo.Context) error {
	id, err := parseId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if err := ctrl.service.DeleteFullAbility(c, id); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Ability deleted successfully"})
}

// previewRank godoc
// @Id previewAaEditorRank
// @Summary Validates an AA ability tree (server-side validation)
// @Accept json
// @Produce json
// @Tags AaEditor
// @Param body body AaAbilityInput true "Ability tree"
// @Success 200 {object} ValidationResult
// @Router /aa_editor/ranks/preview [post]
func (ctrl *AaEditorController) previewRank(c echo.Context) error {
	input := new(AaAbilityInput)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	result := ctrl.service.ValidateAbility(c, *input)
	return c.JSON(http.StatusOK, echo.Map{"data": result})
}
