package aaeditor

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/EQEmu/spire/internal/http/request"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const aaExportBundleVersion = "v1"

type AaExportRequest struct {
	AbilityIds []int `json:"ability_ids"`
}

type AaExportConnection struct {
	ID            uint   `json:"id,omitempty"`
	Name          string `json:"name"`
	ContentDbName string `json:"content_db_name,omitempty"`
	IsDefault     bool   `json:"is_default"`
}

type AaExportAbility struct {
	Ability  AaAbilityInput `json:"ability"`
	Warnings []string       `json:"warnings,omitempty"`
}

type AaExportBundle struct {
	Version          string             `json:"version"`
	ExportedAt       string             `json:"exported_at"`
	SourceConnection AaExportConnection `json:"source_connection"`
	Abilities        []AaExportAbility  `json:"abilities"`
}

type AaImportAbilityAction struct {
	AbilityId int      `json:"ability_id"`
	Name      string   `json:"name"`
	Action    string   `json:"action"`
	Reasons   []string `json:"reasons,omitempty"`
}

type AaImportPreviewResult struct {
	Version            string                  `json:"version"`
	Valid              bool                    `json:"valid"`
	Creates            int                     `json:"creates"`
	Updates            int                     `json:"updates"`
	Blocked            int                     `json:"blocked"`
	MissingPrereqAaIds []int                   `json:"missing_prereq_aa_ids"`
	MissingSpellIds    []int                   `json:"missing_spell_ids"`
	Actions            []AaImportAbilityAction `json:"actions"`
}

type AaImportApplyResult struct {
	Version            string                  `json:"version"`
	Valid              bool                    `json:"valid"`
	Creates            int                     `json:"creates"`
	Updates            int                     `json:"updates"`
	Blocked            int                     `json:"blocked"`
	MissingPrereqAaIds []int                   `json:"missing_prereq_aa_ids"`
	MissingSpellIds    []int                   `json:"missing_spell_ids"`
	AppliedAbilityIds  []int                   `json:"applied_ability_ids"`
	Actions            []AaImportAbilityAction `json:"actions"`
}

func (s *AaEditorService) ExportAbilities(c echo.Context, req AaExportRequest) (*AaExportBundle, error) {
	db := s.db(c)
	if len(req.AbilityIds) == 0 {
		return nil, fmt.Errorf("at least one ability id is required")
	}

	bundle := &AaExportBundle{
		Version:          aaExportBundleVersion,
		ExportedAt:       time.Now().UTC().Format(time.RFC3339),
		SourceConnection: s.getSourceConnection(c),
		Abilities:        make([]AaExportAbility, 0, len(req.AbilityIds)),
	}

	for _, id := range req.AbilityIds {
		if id <= 0 {
			return nil, fmt.Errorf("invalid ability id [%d]", id)
		}
		full, err := s.getFullAbility(db, id)
		if err != nil {
			return nil, err
		}
		input, warnings := normalizeAbilityForBundle(full)
		bundle.Abilities = append(bundle.Abilities, AaExportAbility{
			Ability:  input,
			Warnings: warnings,
		})
	}

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "EXPORT", fmt.Sprintf("Exported AA bundle with %d abilities [%s]", len(req.AbilityIds), joinIntList(req.AbilityIds)))
	}

	return bundle, nil
}

func (s *AaEditorService) PreviewImport(c echo.Context, bundle AaExportBundle) (*AaImportPreviewResult, error) {
	return s.previewImportBundle(s.db(c), bundle)
}

func (s *AaEditorService) ApplyImport(c echo.Context, bundle AaExportBundle) (*AaImportApplyResult, error) {
	result, err := s.applyImportBundle(s.db(c), bundle)
	if err != nil {
		return nil, err
	}

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "IMPORT", fmt.Sprintf("Imported AA bundle with %d abilities [%s]", len(result.AppliedAbilityIds), joinIntList(result.AppliedAbilityIds)))
	}

	return result, nil
}

func (s *AaEditorService) applyImportBundle(db *gorm.DB, bundle AaExportBundle) (*AaImportApplyResult, error) {
	preview, err := s.previewImportBundle(db, bundle)
	if err != nil {
		return nil, err
	}
	if !preview.Valid || preview.Blocked > 0 {
		return nil, fmt.Errorf("bundle import is blocked")
	}

	appliedIds := make([]int, 0, len(bundle.Abilities))
	err = db.Transaction(func(tx *gorm.DB) error {
		currentPreview, err := s.previewImportBundle(tx, bundle)
		if err != nil {
			return err
		}
		if !currentPreview.Valid || currentPreview.Blocked > 0 {
			return fmt.Errorf("bundle import is blocked")
		}

		existingAbilityIds, err := loadExistingAbilityIds(tx, bundleAbilityIds(bundle))
		if err != nil {
			return err
		}

		for _, exported := range bundle.Abilities {
			input := cloneAbilityInput(exported.Ability)
			if existingAbilityIds[input.ID] {
				existingFull, err := s.getFullAbility(tx, input.ID)
				if err != nil {
					return err
				}
				aligned := alignBundleInputToExisting(input, existingFull)
				if err := s.saveAbilityTreeTx(tx, input.ID, aligned); err != nil {
					return err
				}
			} else {
				if _, err := s.createAbilityTreeTx(tx, input, createAbilityTreeOptions{AllowExplicitAbilityID: true}); err != nil {
					return err
				}
			}
			appliedIds = append(appliedIds, input.ID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AaImportApplyResult{
		Version:            preview.Version,
		Valid:              true,
		Creates:            preview.Creates,
		Updates:            preview.Updates,
		Blocked:            0,
		MissingPrereqAaIds: []int{},
		MissingSpellIds:    []int{},
		AppliedAbilityIds:  appliedIds,
		Actions:            preview.Actions,
	}, nil
}

func (s *AaEditorService) previewImportBundle(db *gorm.DB, bundle AaExportBundle) (*AaImportPreviewResult, error) {
	if err := validateBundleEnvelope(bundle); err != nil {
		return nil, err
	}

	result := &AaImportPreviewResult{
		Version:            bundle.Version,
		Valid:              true,
		MissingPrereqAaIds: []int{},
		MissingSpellIds:    []int{},
		Actions:            make([]AaImportAbilityAction, 0, len(bundle.Abilities)),
	}

	bundleIds := bundleAbilityIds(bundle)
	existingAbilityIds, err := loadExistingAbilityIds(db, bundleIds)
	if err != nil {
		return nil, err
	}

	prereqIds := collectExternalPrereqIds(bundle)
	existingPrereqIds, err := loadExistingAbilityIds(db, prereqIds)
	if err != nil {
		return nil, err
	}

	spellIds := collectSpellIds(bundle)
	existingSpellIds, err := loadExistingSpellIds(db, spellIds)
	if err != nil {
		return nil, err
	}

	duplicateCounts := map[int]int{}
	for _, exported := range bundle.Abilities {
		duplicateCounts[exported.Ability.ID]++
	}

	missingPrereqs := map[int]bool{}
	missingSpells := map[int]bool{}

	for _, exported := range bundle.Abilities {
		input := cloneAbilityInput(exported.Ability)
		action := AaImportAbilityAction{
			AbilityId: input.ID,
			Name:      input.Name,
			Action:    "create",
		}
		if existingAbilityIds[input.ID] {
			action.Action = "update"
		}

		reasons := make([]string, 0)
		if input.ID <= 0 {
			reasons = append(reasons, "Ability id must be a positive integer")
		}
		if duplicateCounts[input.ID] > 1 {
			reasons = append(reasons, fmt.Sprintf("Bundle contains duplicate ability id [%d]", input.ID))
		}

		validation := s.ValidateAbility(nil, input)
		if validation != nil && !validation.Valid {
			reasons = append(reasons, validation.Errors...)
		}

		for _, rank := range input.Ranks {
			if rank.Spell > 0 && !existingSpellIds[rank.Spell] {
				missingSpells[rank.Spell] = true
				reasons = append(reasons, fmt.Sprintf("Referenced spell [%d] does not exist on the target", rank.Spell))
			}
			for _, prereq := range rank.Prereqs {
				if prereq.AaId <= 0 {
					continue
				}
				if bundleIds[prereq.AaId] || existingPrereqIds[prereq.AaId] {
					continue
				}
				missingPrereqs[prereq.AaId] = true
				reasons = append(reasons, fmt.Sprintf("Referenced prerequisite AA [%d] does not exist on the target or in the bundle", prereq.AaId))
			}
		}

		if len(reasons) > 0 {
			action.Action = "block"
			action.Reasons = dedupeStrings(reasons)
			result.Blocked++
			result.Valid = false
		} else if existingAbilityIds[input.ID] {
			result.Updates++
		} else {
			result.Creates++
		}

		result.Actions = append(result.Actions, action)
	}

	result.MissingPrereqAaIds = sortedIntKeys(missingPrereqs)
	result.MissingSpellIds = sortedIntKeys(missingSpells)

	return result, nil
}

func validateBundleEnvelope(bundle AaExportBundle) error {
	if strings.TrimSpace(bundle.Version) == "" {
		return fmt.Errorf("bundle version is required")
	}
	if bundle.Version != aaExportBundleVersion {
		return fmt.Errorf("unsupported bundle version [%s]", bundle.Version)
	}
	if len(bundle.Abilities) == 0 {
		return fmt.Errorf("bundle must contain at least one ability")
	}
	return nil
}

func normalizeAbilityForBundle(full *AaAbilityFull) (AaAbilityInput, []string) {
	input := AaAbilityInput{
		ID:               int(full.AaAbility.ID),
		Name:             full.AaAbility.Name,
		Category:         full.AaAbility.Category,
		Classes:          full.AaAbility.Classes,
		Races:            full.AaAbility.Races,
		DrakkinHeritage:  full.AaAbility.DrakkinHeritage,
		Deities:          full.AaAbility.Deities,
		Status:           full.AaAbility.Status,
		Type:             full.AaAbility.Type,
		Charges:          full.AaAbility.Charges,
		GrantOnly:        int(full.AaAbility.GrantOnly),
		Enabled:          int(full.AaAbility.Enabled),
		ResetOnDeath:     int(full.AaAbility.ResetOnDeath),
		AutoGrantEnabled: int(full.AaAbility.AutoGrantEnabled),
		FirstRankId:      0,
		Ranks:            make([]AaRankInput, 0, len(full.Ranks)),
	}

	warnings := append([]string{}, full.Warnings...)
	for index, rank := range full.Ranks {
		exportedRank := AaRankInput{
			TempId:         fmt.Sprintf("bundle-rank-%d", index),
			ID:             0,
			UpperHotkeySid: 0,
			LowerHotkeySid: 0,
			TitleSid:       0,
			DescSid:        0,
			Cost:           rank.AaRank.Cost,
			LevelReq:       rank.AaRank.LevelReq,
			Spell:          rank.AaRank.Spell,
			SpellType:      rank.AaRank.SpellType,
			RecastTime:     rank.AaRank.RecastTime,
			Expansion:      rank.AaRank.Expansion,
			PrevId:         0,
			NextId:         -1,
			Effects:        make([]AaRankEffectInput, 0, len(rank.Effects)),
			Prereqs:        make([]AaRankPrereqInput, 0, len(rank.Prereqs)),
			Strings:        map[int]DbStrInput{},
		}

		for _, effect := range rank.Effects {
			exportedRank.Effects = append(exportedRank.Effects, AaRankEffectInput{
				Slot:     effect.Slot,
				EffectId: effect.EffectId,
				Base1:    effect.Base1,
				Base2:    effect.Base2,
			})
		}

		for _, prereq := range rank.Prereqs {
			exportedRank.Prereqs = append(exportedRank.Prereqs, AaRankPrereqInput{
				AaId:   prereq.AaId,
				Points: prereq.Points,
			})
		}

		for key, str := range rank.Strings {
			exportedRank.Strings[key] = DbStrInput{
				ID:    0,
				Type:  str.Type,
				Value: str.Value,
			}
		}

		for _, slot := range []int{1, 2, 3, 4} {
			sid := rankStringSid(rank.AaRank, slot)
			if sid > 0 {
				if _, ok := rank.Strings[slot]; !ok {
					warnings = append(warnings, fmt.Sprintf("Rank #%d string slot [%d] references missing db_str id [%d].", index+1, slot, sid))
				}
			}
		}

		input.Ranks = append(input.Ranks, exportedRank)
	}

	return input, dedupeStrings(warnings)
}

func alignBundleInputToExisting(input AaAbilityInput, existing *AaAbilityFull) AaAbilityInput {
	aligned := cloneAbilityInput(input)
	for i := range aligned.Ranks {
		if i >= len(existing.Ranks) {
			continue
		}
		targetRank := existing.Ranks[i]
		aligned.Ranks[i].ID = int(targetRank.AaRank.ID)

		for slot, str := range aligned.Ranks[i].Strings {
			targetSid := rankStringSid(targetRank.AaRank, slot)
			if targetSid > 0 {
				str.ID = targetSid
				str.Type = aaStringType
				aligned.Ranks[i].Strings[slot] = str
			}
		}
	}

	return aligned
}

func cloneAbilityInput(input AaAbilityInput) AaAbilityInput {
	cloned := input
	cloned.Ranks = make([]AaRankInput, 0, len(input.Ranks))
	for _, rank := range input.Ranks {
		nextRank := rank
		nextRank.Effects = append([]AaRankEffectInput{}, rank.Effects...)
		nextRank.Prereqs = append([]AaRankPrereqInput{}, rank.Prereqs...)
		if rank.Strings != nil {
			nextRank.Strings = map[int]DbStrInput{}
			for key, str := range rank.Strings {
				nextRank.Strings[key] = str
			}
		} else {
			nextRank.Strings = map[int]DbStrInput{}
		}
		cloned.Ranks = append(cloned.Ranks, nextRank)
	}
	return cloned
}

func bundleAbilityIds(bundle AaExportBundle) map[int]bool {
	ids := map[int]bool{}
	for _, ability := range bundle.Abilities {
		if ability.Ability.ID > 0 {
			ids[ability.Ability.ID] = true
		}
	}
	return ids
}

func collectExternalPrereqIds(bundle AaExportBundle) map[int]bool {
	ids := map[int]bool{}
	bundleIds := bundleAbilityIds(bundle)
	for _, ability := range bundle.Abilities {
		for _, rank := range ability.Ability.Ranks {
			for _, prereq := range rank.Prereqs {
				if prereq.AaId > 0 && !bundleIds[prereq.AaId] {
					ids[prereq.AaId] = true
				}
			}
		}
	}
	return ids
}

func collectSpellIds(bundle AaExportBundle) map[int]bool {
	ids := map[int]bool{}
	for _, ability := range bundle.Abilities {
		for _, rank := range ability.Ability.Ranks {
			if rank.Spell > 0 {
				ids[rank.Spell] = true
			}
		}
	}
	return ids
}

func loadExistingAbilityIds(db *gorm.DB, ids map[int]bool) (map[int]bool, error) {
	out := map[int]bool{}
	if len(ids) == 0 {
		return out, nil
	}

	var abilityIds []int
	for id := range ids {
		abilityIds = append(abilityIds, id)
	}

	var rows []models.AaAbility
	if err := db.Select("id").Where("id IN ?", abilityIds).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[int(row.ID)] = true
	}
	return out, nil
}

func loadExistingSpellIds(db *gorm.DB, ids map[int]bool) (map[int]bool, error) {
	out := map[int]bool{}
	if len(ids) == 0 {
		return out, nil
	}

	var spellIds []int
	for id := range ids {
		spellIds = append(spellIds, id)
	}

	type spellRow struct {
		ID int `gorm:"column:id"`
	}
	var rows []spellRow
	if err := db.Table("spells_new").Select("id").Where("id IN ?", spellIds).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = true
	}
	return out, nil
}

func rankStringSid(rank models.AaRank, slot int) int {
	switch slot {
	case 1:
		return rank.TitleSid
	case 2:
		return rank.DescSid
	case 3:
		return rank.UpperHotkeySid
	case 4:
		return rank.LowerHotkeySid
	default:
		return 0
	}
}

func sortedIntKeys(values map[int]bool) []int {
	out := make([]int, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func joinIntList(values []int) string {
	if len(values) == 0 {
		return ""
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return strings.Join(parts, ", ")
}

func (s *AaEditorService) getSourceConnection(c echo.Context) AaExportConnection {
	info := AaExportConnection{
		Name:      "Local (Default)",
		IsDefault: true,
	}
	if s.resolver == nil || c == nil {
		return info
	}

	user := request.GetUser(c)
	if user.ID == 0 {
		return info
	}

	conn := s.resolver.GetUserConnection(user)
	if conn.ID == 0 {
		return info
	}

	info.ID = conn.ServerDatabaseConnection.ID
	info.Name = conn.ServerDatabaseConnection.Name
	info.IsDefault = false
	if conn.ServerDatabaseConnection.ContentDbName != "" {
		info.ContentDbName = conn.ServerDatabaseConnection.ContentDbName
	} else {
		info.ContentDbName = conn.ServerDatabaseConnection.DbName
	}

	return info
}
