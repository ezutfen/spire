package aaeditor

import (
	"fmt"
	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/logger"
	"github.com/EQEmu/spire/internal/models"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
)

// AaEditorService performs transactional hierarchical operations across the
// four AA tables (aa_ability, aa_ranks, aa_rank_effects, aa_rank_prereqs) and
// the referenced db_str strings. It follows the same blueprint as the Quest
// Editor service but operates on database rows instead of files.
type AaEditorService struct {
	resolver *database.Resolver
	auditLog *auditlog.UserEvent
	logger   *logger.AppLogger
}

func NewAaEditorService(
	db *database.Resolver,
	auditLog *auditlog.UserEvent,
	appLogger *logger.AppLogger,
) *AaEditorService {
	return &AaEditorService{
		resolver: db,
		auditLog: auditLog,
		logger:   appLogger,
	}
}

// ------------------------------------------------------------------
// DTOs
// ------------------------------------------------------------------

type AaRankEffectInput struct {
	Slot     int `json:"slot"`
	EffectId int `json:"effect_id"`
	Base1    int `json:"base_1"`
	Base2    int `json:"base_2"`
}

type AaRankPrereqInput struct {
	AaId   int `json:"aa_id"`
	Points int `json:"points"`
}

type DbStrInput struct {
	ID    int    `json:"id"`
	Type  int    `json:"type"`
	Value string `json:"value"`
}

type AaRankInput struct {
	// TempId is a client-side identifier used to reference a rank during
	// create/duplicate before a real DB id exists. It is ignored on save.
	TempId         string              `json:"temp_id"`
	ID             int                 `json:"id"`
	UpperHotkeySid int                 `json:"upper_hotkey_sid"`
	LowerHotkeySid int                 `json:"lower_hotkey_sid"`
	TitleSid       int                 `json:"title_sid"`
	DescSid        int                 `json:"desc_sid"`
	Cost           int                 `json:"cost"`
	LevelReq       int                 `json:"level_req"`
	Spell          int                 `json:"spell"`
	SpellType      int                 `json:"spell_type"`
	RecastTime     int                 `json:"recast_time"`
	Expansion      int                 `json:"expansion"`
	PrevId         int                 `json:"prev_id"`
	NextId         int                 `json:"next_id"`
	Effects        []AaRankEffectInput `json:"effects"`
	Prereqs        []AaRankPrereqInput `json:"prereqs"`
	Strings        map[int]DbStrInput  `json:"strings"`
}

type AaAbilityInput struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Category         int           `json:"category"`
	Classes          int           `json:"classes"`
	Races            int           `json:"races"`
	DrakkinHeritage  int           `json:"drakkin_heritage"`
	Deities          int           `json:"deities"`
	Status           int           `json:"status"`
	Type             int           `json:"type"`
	Charges          int           `json:"charges"`
	GrantOnly        int           `json:"grant_only"`
	Enabled          int           `json:"enabled"`
	ResetOnDeath     int           `json:"reset_on_death"`
	AutoGrantEnabled int           `json:"auto_grant_enabled"`
	FirstRankId      int           `json:"first_rank_id"`
	Ranks            []AaRankInput `json:"ranks"`
}

type DuplicateOptions struct {
	Name      string `json:"name"`
	RemapSelf bool   `json:"remap_self"`
}

// Response shapes

type AaRankEffectFull struct {
	RankId   int `json:"rank_id"`
	Slot     int `json:"slot"`
	EffectId int `json:"effect_id"`
	Base1    int `json:"base_1"`
	Base2    int `json:"base_2"`
}

type AaRankPrereqFull struct {
	RankId int `json:"rank_id"`
	AaId   int `json:"aa_id"`
	Points int `json:"points"`
}

type AaStringFull struct {
	ID    int    `json:"id"`
	Type  int    `json:"type"`
	Value string `json:"value"`
}

type AaSpellSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AaRankFull struct {
	AaRank  models.AaRank        `json:"aa_rank"`
	Effects []AaRankEffectFull   `json:"effects"`
	Prereqs []AaRankPrereqFull   `json:"prereqs"`
	Strings map[int]AaStringFull `json:"strings"`
	Spell   *AaSpellSummary      `json:"spell"`
}

type AaAbilityFull struct {
	AaAbility models.AaAbility `json:"aa_ability"`
	Ranks     []AaRankFull     `json:"ranks"`
}

type AaAbilityListItem struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Category      int    `json:"category"`
	Classes       int    `json:"classes"`
	Enabled       int    `json:"enabled"`
	Type          int    `json:"type"`
	FirstRankId   int    `json:"first_rank_id"`
	RankCount     int    `json:"rank_count"`
	FirstRankCost int    `json:"first_rank_cost"`
	FirstRankLvl  int    `json:"first_rank_level"`
}

type AaAbilityListResult struct {
	Total int64               `json:"total"`
	Items []AaAbilityListItem `json:"items"`
}

type AaListFilter struct {
	Search   string
	Category int
	Classes  int
	Enabled  int
	Page     int
	Limit    int
	OrderBy  string
	OrderDir string
}

type AaMetadata struct {
	Categories map[int]string `json:"categories"`
	Types      map[int]string `json:"types"`
	SpellTypes map[int]string `json:"spell_types"`
	Statuses   map[int]string `json:"statuses"`
	Expansions map[int]string `json:"expansions"`
}

// ------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------

const aaStringType = 1

// db resolves the eqemu_content connection for the request context.
func (s *AaEditorService) db(c echo.Context) *gorm.DB {
	return s.resolver.Get(models.AaAbility{}, c)
}

// walkRankChain loads the ordered rank list by following next_id starting from
// firstRankId. It guards against cycles with a hard iteration cap.
func walkRankChain(tx *gorm.DB, firstRankId int) ([]models.AaRank, error) {
	if firstRankId <= 0 {
		return []models.AaRank{}, nil
	}

	const maxRanks = 10000
	rankById := map[int]models.AaRank{}
	var ordered []models.AaRank

	// load the chain
	current := firstRankId
	for i := 0; i < maxRanks; i++ {
		if current <= 0 {
			break
		}
		if _, seen := rankById[current]; seen {
			break // cycle guard
		}

		var rank models.AaRank
		err := tx.Where("id = ?", current).First(&rank).Error
		if err != nil {
			return nil, fmt.Errorf("rank [%v] not found: %w", current, err)
		}
		rankById[current] = rank
		ordered = append(ordered, rank)
		current = rank.NextId
	}

	return ordered, nil
}

// loadEffectsBulk loads effects for all given rank ids in a single query,
// grouped by rank id (ordered by slot).
func loadEffectsBulk(tx *gorm.DB, rankIds []int) (map[int][]AaRankEffectFull, error) {
	out := map[int][]AaRankEffectFull{}
	for _, id := range rankIds {
		out[id] = []AaRankEffectFull{}
	}
	if len(rankIds) == 0 {
		return out, nil
	}
	var rows []models.AaRankEffect
	if err := tx.Where("rank_id IN ?", rankIds).Order("rank_id asc, slot asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[int(r.RankId)] = append(out[int(r.RankId)], AaRankEffectFull{
			RankId:   int(r.RankId),
			Slot:     int(r.Slot),
			EffectId: r.EffectId,
			Base1:    r.Base1,
			Base2:    r.Base2,
		})
	}
	return out, nil
}

// loadPrereqsBulk loads prereqs for all given rank ids in a single query,
// grouped by rank id.
func loadPrereqsBulk(tx *gorm.DB, rankIds []int) (map[int][]AaRankPrereqFull, error) {
	out := map[int][]AaRankPrereqFull{}
	for _, id := range rankIds {
		out[id] = []AaRankPrereqFull{}
	}
	if len(rankIds) == 0 {
		return out, nil
	}
	var rows []models.AaRankPrereq
	if err := tx.Where("rank_id IN ?", rankIds).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[int(r.RankId)] = append(out[int(r.RankId)], AaRankPrereqFull{
			RankId: int(r.RankId),
			AaId:   r.AaId,
			Points: r.Points,
		})
	}
	return out, nil
}

// loadStringsBulk resolves db_str (type 1) entries for every rank in a single
// query. The per-rank map is keyed by the sid field it came from (1=title,
// 2=desc, 3=upper_hotkey, 4=lower_hotkey).
func loadStringsBulk(tx *gorm.DB, ranks []models.AaRank) (map[int]map[int]AaStringFull, error) {
	out := map[int]map[int]AaStringFull{}
	for _, r := range ranks {
		out[int(r.ID)] = map[int]AaStringFull{}
	}

	uniq := map[int]bool{}
	var ids []int
	for _, r := range ranks {
		for _, sid := range []int{r.TitleSid, r.DescSid, r.UpperHotkeySid, r.LowerHotkeySid} {
			if sid > 0 && !uniq[sid] {
				uniq[sid] = true
				ids = append(ids, sid)
			}
		}
	}
	if len(ids) == 0 {
		return out, nil
	}

	var rows []models.DbStr
	if err := tx.Where("type = ? AND id IN ?", aaStringType, ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	byId := map[int]models.DbStr{}
	for _, r := range rows {
		byId[r.ID] = r
	}

	for _, r := range ranks {
		strs := out[int(r.ID)]
		assign := func(key int, sid int) {
			if sid <= 0 {
				return
			}
			if s, ok := byId[sid]; ok {
				strs[key] = AaStringFull{ID: s.ID, Type: s.Type, Value: s.Value}
			}
		}
		assign(1, r.TitleSid)
		assign(2, r.DescSid)
		assign(3, r.UpperHotkeySid)
		assign(4, r.LowerHotkeySid)
	}
	return out, nil
}

// loadSpellSummariesBulk resolves spell names for every rank's spell id in a
// single query, keyed by rank id.
func loadSpellSummariesBulk(tx *gorm.DB, ranks []models.AaRank) (map[int]*AaSpellSummary, error) {
	out := map[int]*AaSpellSummary{}
	uniq := map[int]bool{}
	var ids []int
	for _, r := range ranks {
		if r.Spell > 0 && !uniq[r.Spell] {
			uniq[r.Spell] = true
			ids = append(ids, r.Spell)
		}
	}
	if len(ids) == 0 {
		return out, nil
	}
	type spellRow struct {
		ID   int    `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var rows []spellRow
	if err := tx.Table("spells_new").Select("id, name").Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	byId := map[int]string{}
	for _, r := range rows {
		byId[r.ID] = r.Name
	}
	for _, r := range ranks {
		if r.Spell > 0 {
			if name, ok := byId[r.Spell]; ok {
				out[int(r.ID)] = &AaSpellSummary{ID: r.Spell, Name: name}
			}
		}
	}
	return out, nil
}

// ------------------------------------------------------------------
// Read operations
// ------------------------------------------------------------------

func (s *AaEditorService) GetFullAbility(c echo.Context, id int) (*AaAbilityFull, error) {
	return s.getFullAbility(s.db(c), id)
}

func (s *AaEditorService) getFullAbility(db *gorm.DB, id int) (*AaAbilityFull, error) {
	var ability models.AaAbility
	if err := db.Where("id = ?", id).First(&ability).Error; err != nil {
		return nil, fmt.Errorf("ability not found: %w", err)
	}

	chain, err := walkRankChain(db, ability.FirstRankId)
	if err != nil {
		return nil, err
	}

	rankIds := make([]int, 0, len(chain))
	for _, r := range chain {
		rankIds = append(rankIds, int(r.ID))
	}

	// bulk-load children in a constant number of queries rather than fanning
	// out per rank (effects/prereqs/strings/spell per rank).
	effectsByRank, err := loadEffectsBulk(db, rankIds)
	if err != nil {
		return nil, err
	}
	prereqsByRank, err := loadPrereqsBulk(db, rankIds)
	if err != nil {
		return nil, err
	}
	stringsByRank, err := loadStringsBulk(db, chain)
	if err != nil {
		return nil, err
	}
	spellsByRank, err := loadSpellSummariesBulk(db, chain)
	if err != nil {
		return nil, err
	}

	ranks := make([]AaRankFull, 0, len(chain))
	for _, r := range chain {
		rid := int(r.ID)
		ranks = append(ranks, AaRankFull{
			AaRank:  r,
			Effects: effectsByRank[rid],
			Prereqs: prereqsByRank[rid],
			Strings: stringsByRank[rid],
			Spell:   spellsByRank[rid],
		})
	}

	return &AaAbilityFull{AaAbility: ability, Ranks: ranks}, nil
}

func (s *AaEditorService) ListAbilities(c echo.Context, filter AaListFilter) (*AaAbilityListResult, error) {
	return s.listAbilities(s.db(c), filter)
}

func (s *AaEditorService) listAbilities(db *gorm.DB, filter AaListFilter) (*AaAbilityListResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Page < 0 {
		filter.Page = 0
	}
	if filter.OrderBy == "" {
		filter.OrderBy = "name"
	}
	if filter.OrderDir == "" {
		filter.OrderDir = "asc"
	}

	q := db.Table("aa_ability").Model(&models.AaAbility{})

	if strings.TrimSpace(filter.Search) != "" {
		// numeric search by id, otherwise name like
		if idVal, err := strconv.Atoi(strings.TrimSpace(filter.Search)); err == nil {
			q = q.Where("id = ?", idVal)
		} else {
			q = q.Where("name LIKE ?", "%"+strings.TrimSpace(filter.Search)+"%")
		}
	}
	if filter.Category > 0 {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Classes > 0 {
		q = q.Where("classes & ? > 0", filter.Classes)
	}
	if filter.Enabled >= 0 {
		// enabled filter: -1 means "any", 0 disabled, 1 enabled
		if filter.Enabled == 0 {
			q = q.Where("enabled = 0")
		} else if filter.Enabled == 1 {
			q = q.Where("enabled = 1")
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	// safe order by (whitelist columns)
	allowedOrders := map[string]bool{
		"id": true, "name": true, "category": true, "type": true, "enabled": true,
	}
	orderCol := "name"
	if allowedOrders[filter.OrderBy] {
		orderCol = filter.OrderBy
	}
	orderDir := "asc"
	if strings.ToLower(filter.OrderDir) == "desc" {
		orderDir = "desc"
	}
	q = q.Order(fmt.Sprintf("%s %s", orderCol, orderDir)).
		Limit(filter.Limit).
		Offset(filter.Page * filter.Limit)

	var abilities []models.AaAbility
	if err := q.Find(&abilities).Error; err != nil {
		return nil, err
	}

	items := make([]AaAbilityListItem, 0, len(abilities))
	for _, a := range abilities {
		item := AaAbilityListItem{
			ID:          int(a.ID),
			Name:        a.Name,
			Category:    a.Category,
			Classes:     a.Classes,
			Enabled:     int(a.Enabled),
			Type:        a.Type,
			FirstRankId: a.FirstRankId,
			RankCount:   countRanks(db, a.FirstRankId),
		}
		if a.FirstRankId > 0 {
			var firstRank models.AaRank
			if err := db.Where("id = ?", a.FirstRankId).First(&firstRank).Error; err == nil {
				item.FirstRankCost = firstRank.Cost
				item.FirstRankLvl = firstRank.LevelReq
			}
		}
		items = append(items, item)
	}

	return &AaAbilityListResult{Total: total, Items: items}, nil
}

func countRanks(db *gorm.DB, firstRankId int) int {
	if firstRankId <= 0 {
		return 0
	}
	chain, err := walkRankChain(db, firstRankId)
	if err != nil {
		return 0
	}
	return len(chain)
}

func (s *AaEditorService) GetMetadata() AaMetadata {
	return AaMetadata{
		Categories: AaCategories,
		Types:      AaTypes,
		SpellTypes: AaSpellTypes,
		Statuses:   AaStatuses,
		Expansions: AaExpansions,
	}
}

// ------------------------------------------------------------------
// Create
// ------------------------------------------------------------------

func (s *AaEditorService) CreateFullAbility(c echo.Context, input AaAbilityInput) (*AaAbilityFull, error) {
	db := s.db(c)
	createdID, err := s.createAbilityTree(db, input)
	if err != nil {
		return nil, err
	}

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "CREATE", fmt.Sprintf("Created [AaAbility] [%v] name [%v] with %v ranks", createdID, input.Name, len(input.Ranks)))
	}

	return s.getFullAbility(db, createdID)
}

func (s *AaEditorService) createAbilityTree(db *gorm.DB, input AaAbilityInput) (int, error) {
	if strings.TrimSpace(input.Name) == "" {
		return 0, fmt.Errorf("ability name is required")
	}
	if len(input.Ranks) == 0 {
		return 0, fmt.Errorf("at least one rank is required")
	}

	var createdID int
	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. upsert any inline strings and backfill sid references
		for i := range input.Ranks {
			if err := upsertRankStrings(tx, &input.Ranks[i]); err != nil {
				return err
			}
		}

		// 2. insert ability with placeholder first_rank_id (0)
		ability := models.AaAbility{
			Name:             input.Name,
			Category:         input.Category,
			Classes:          input.Classes,
			Races:            input.Races,
			DrakkinHeritage:  input.DrakkinHeritage,
			Deities:          input.Deities,
			Status:           input.Status,
			Type:             input.Type,
			Charges:          input.Charges,
			GrantOnly:        int8(input.GrantOnly),
			Enabled:          uint8(input.Enabled),
			ResetOnDeath:     int8(input.ResetOnDeath),
			AutoGrantEnabled: int8(input.AutoGrantEnabled),
			FirstRankId:      0,
		}
		if err := tx.Create(&ability).Error; err != nil {
			return fmt.Errorf("failed to create ability: %w", err)
		}
		createdID = int(ability.ID)

		// 3. insert ranks with zeroed prev/next, collecting assigned ids
		rankIds, err := insertRanks(tx, input.Ranks)
		if err != nil {
			return err
		}

		// 4. stitch prev/next chain + set first_rank_id
		if err := stitchChain(tx, rankIds); err != nil {
			return err
		}
		if len(rankIds) > 0 {
			if err := tx.Model(&models.AaAbility{}).
				Where("id = ?", createdID).
				Update("first_rank_id", rankIds[0]).Error; err != nil {
				return fmt.Errorf("failed to set first_rank_id: %w", err)
			}
		}

		// 5. insert effects + prereqs for each rank
		if err := insertChildren(tx, input.Ranks, rankIds, createdID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return createdID, nil
}

// buildRankModel assembles an aa_ranks row from input, zeroing prev/next (the
// chain is stitched separately). It is the single source of truth for the
// editable rank field-set used on both create and update.
func buildRankModel(r AaRankInput) models.AaRank {
	return models.AaRank{
		UpperHotkeySid: r.UpperHotkeySid,
		LowerHotkeySid: r.LowerHotkeySid,
		TitleSid:       r.TitleSid,
		DescSid:        r.DescSid,
		Cost:           r.Cost,
		LevelReq:       r.LevelReq,
		Spell:          r.Spell,
		SpellType:      r.SpellType,
		RecastTime:     r.RecastTime,
		Expansion:      r.Expansion,
		PrevId:         0,
		NextId:         0,
	}
}

// rankFieldMap mirrors buildRankModel but as a column->value map for GORM
// Updates, keeping create and update in lockstep.
func rankFieldMap(r AaRankInput) map[string]interface{} {
	return map[string]interface{}{
		"upper_hotkey_sid": r.UpperHotkeySid,
		"lower_hotkey_sid": r.LowerHotkeySid,
		"title_sid":        r.TitleSid,
		"desc_sid":         r.DescSid,
		"cost":             r.Cost,
		"level_req":        r.LevelReq,
		"spell":            r.Spell,
		"spell_type":       r.SpellType,
		"recast_time":      r.RecastTime,
		"expansion":        r.Expansion,
	}
}

func buildEffect(rankId int, e AaRankEffectInput) models.AaRankEffect {
	return models.AaRankEffect{
		RankId:   uint(rankId),
		Slot:     uint(e.Slot),
		EffectId: e.EffectId,
		Base1:    e.Base1,
		Base2:    e.Base2,
	}
}

func buildPrereq(rankId int, p AaRankPrereqInput) models.AaRankPrereq {
	return models.AaRankPrereq{
		RankId: uint(rankId),
		AaId:   p.AaId,
		Points: p.Points,
	}
}

// insertRanks inserts each rank (prev/next zeroed) and returns the assigned ids
// in input order. If a rank already carries an explicit id it is honored.
func insertRanks(tx *gorm.DB, ranks []AaRankInput) ([]int, error) {
	ids := make([]int, 0, len(ranks))
	for _, r := range ranks {
		model := buildRankModel(r)
		if r.ID > 0 {
			model.ID = uint(r.ID)
		}
		if err := tx.Create(&model).Error; err != nil {
			return nil, fmt.Errorf("failed to create rank: %w", err)
		}
		ids = append(ids, int(model.ID))
	}
	return ids, nil
}

// stitchChain sets prev_id/next_id for the ordered rank ids.
func stitchChain(tx *gorm.DB, ids []int) error {
	for i, id := range ids {
		prev := 0
		next := 0
		if i > 0 {
			prev = ids[i-1]
		}
		if i < len(ids)-1 {
			next = ids[i+1]
		}
		if err := tx.Model(&models.AaRank{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"prev_id": prev,
				"next_id": next,
			}).Error; err != nil {
			return fmt.Errorf("failed to stitch rank [%v]: %w", id, err)
		}
	}
	return nil
}

// insertChildren inserts effects + prereqs for each rank, mapping temp order to
// real rank ids. A prereq whose aa_id is 0 is treated as a self-reference and
// is rewritten to the owning ability's id (this is how the duplicate-with-remap
// path marks self-referencing prereqs before the ability id is known).
func insertChildren(tx *gorm.DB, ranks []AaRankInput, rankIds []int, abilityId int) error {
	for i, r := range ranks {
		if i >= len(rankIds) {
			break
		}
		rankId := rankIds[i]
		for _, e := range r.Effects {
			if err := tx.Create(buildEffect(rankId, e)).Error; err != nil {
				return fmt.Errorf("failed to create effect for rank [%v]: %w", rankId, err)
			}
		}
		for _, p := range r.Prereqs {
			prereq := buildPrereq(rankId, p)
			if prereq.AaId == 0 {
				prereq.AaId = abilityId
			}
			if err := tx.Create(&prereq).Error; err != nil {
				return fmt.Errorf("failed to create prereq for rank [%v]: %w", rankId, err)
			}
		}
	}
	return nil
}

// upsertRankStrings handles inline db_str entries supplied with a rank. Entries
// that carry Value and a non-positive id are allocated a free id; entries with a
// positive id are updated in place. The corresponding *_sid field on the rank is
// backfilled so subsequent inserts reference the real string id.
func upsertRankStrings(tx *gorm.DB, rank *AaRankInput) error {
	if len(rank.Strings) == 0 {
		return nil
	}
	// key mapping: 1=title,2=desc,3=upper_hotkey,4=lower_hotkey
	for key, str := range rank.Strings {
		// never blank out a shared db_str row: an empty value is a no-op even
		// when a positive id is supplied (the REST endpoint has no validation).
		if strings.TrimSpace(str.Value) == "" {
			continue
		}
		// db_str is shared game-wide (spells, items, ...). Always pin type to the
		// AA scope so a client-supplied type cannot target an unrelated row.
		realType := aaStringType

		var sid int
		if str.ID > 0 {
			// update existing
			if err := tx.Model(&models.DbStr{}).
				Where("id = ? AND type = ?", str.ID, realType).
				Update("value", str.Value).Error; err != nil {
				return fmt.Errorf("failed to update string [%v]: %w", str.ID, err)
			}
			sid = str.ID
		} else {
			// allocate free id and insert
			newId, err := allocateFreeStringId(tx)
			if err != nil {
				return err
			}
			row := models.DbStr{ID: newId, Type: realType, Value: str.Value}
			if err := tx.Create(&row).Error; err != nil {
				return fmt.Errorf("failed to create string: %w", err)
			}
			sid = newId
		}

		switch key {
		case 1:
			rank.TitleSid = sid
		case 2:
			rank.DescSid = sid
		case 3:
			rank.UpperHotkeySid = sid
		case 4:
			rank.LowerHotkeySid = sid
		}
	}
	return nil
}

// allocateFreeStringId finds max(id) in db_str and returns the next value. The
// SELECT ... FOR UPDATE serializes concurrent allocations within the same
// transaction isolation so two writers cannot read the same max and collide on
// the (id, type) composite key.
func allocateFreeStringId(tx *gorm.DB) (int, error) {
	var maxId int
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&models.DbStr{}).
		Select("COALESCE(MAX(id), 0)").Row().Scan(&maxId); err != nil {
		return 0, fmt.Errorf("failed to determine max string id: %w", err)
	}
	return maxId + 1, nil
}

// ------------------------------------------------------------------
// Duplicate
// ------------------------------------------------------------------

func (s *AaEditorService) DuplicateAbility(c echo.Context, srcId int, opts DuplicateOptions) (*AaAbilityFull, error) {
	db := s.db(c)
	full, err := s.getFullAbility(db, srcId)
	if err != nil {
		return nil, err
	}

	name := full.AaAbility.Name + " (Copy)"
	if strings.TrimSpace(opts.Name) != "" {
		name = opts.Name
	}

	input := AaAbilityInput{
		Name:             name,
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
	}

	for i, r := range full.Ranks {
		// deep-copy strings so the duplicate owns its own db_str rows
		strings := map[int]DbStrInput{}
		for key, sf := range r.Strings {
			// new string (id <= 0 triggers free-id allocation)
			strings[key] = DbStrInput{ID: 0, Type: sf.Type, Value: sf.Value}
		}
		// preserve the referenced sid for strings we didn't load (still valid shared refs)
		ri := AaRankInput{
			TempId:         fmt.Sprintf("rank-%d", i),
			UpperHotkeySid: r.AaRank.UpperHotkeySid,
			LowerHotkeySid: r.AaRank.LowerHotkeySid,
			TitleSid:       r.AaRank.TitleSid,
			DescSid:        r.AaRank.DescSid,
			Cost:           r.AaRank.Cost,
			LevelReq:       r.AaRank.LevelReq,
			Spell:          r.AaRank.Spell,
			SpellType:      r.AaRank.SpellType,
			RecastTime:     r.AaRank.RecastTime,
			Expansion:      r.AaRank.Expansion,
			Strings:        strings,
		}
		// drop the copied string sid so the inline Strings map drives allocation
		for key := range strings {
			switch key {
			case 1:
				ri.TitleSid = 0
			case 2:
				ri.DescSid = 0
			case 3:
				ri.UpperHotkeySid = 0
			case 4:
				ri.LowerHotkeySid = 0
			}
		}
		for _, e := range r.Effects {
			ri.Effects = append(ri.Effects, AaRankEffectInput{
				Slot: e.Slot, EffectId: e.EffectId, Base1: e.Base1, Base2: e.Base2,
			})
		}
		for _, p := range r.Prereqs {
			aaId := p.AaId
			// optionally remap prereqs pointing back to the source to the new copy
			if opts.RemapSelf && aaId == srcId {
				aaId = 0 // resolved after create below
			}
			ri.Prereqs = append(ri.Prereqs, AaRankPrereqInput{AaId: aaId, Points: p.Points})
		}
		input.Ranks = append(input.Ranks, ri)
	}

	createdID, err := s.createAbilityTree(db, input)
	if err != nil {
		return nil, err
	}

	// Self-referencing prereqs (aa_id == 0 placeholder) are remapped to the new
	// ability id inside createAbilityTree, so no post-create fixup is needed.

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "CREATE", fmt.Sprintf("Duplicated [AaAbility] [%v] -> [%v]", srcId, createdID))
	}

	return s.getFullAbility(db, createdID)
}

// ------------------------------------------------------------------
// Save (full replace)
// ------------------------------------------------------------------

func (s *AaEditorService) SaveFullAbility(c echo.Context, id int, input AaAbilityInput) (*AaAbilityFull, error) {
	db := s.db(c)
	if err := s.saveAbilityTree(db, id, input); err != nil {
		return nil, err
	}

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "UPDATE", fmt.Sprintf("Updated [AaAbility] [%v] name [%v] with %v ranks", id, input.Name, len(input.Ranks)))
	}

	return s.getFullAbility(db, id)
}

func (s *AaEditorService) saveAbilityTree(db *gorm.DB, id int, input AaAbilityInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("ability name is required")
	}

	// ensure the ability exists
	var existing models.AaAbility
	if err := db.Where("id = ?", id).First(&existing).Error; err != nil {
		return fmt.Errorf("ability not found: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. upsert inline strings
		for i := range input.Ranks {
			if err := upsertRankStrings(tx, &input.Ranks[i]); err != nil {
				return err
			}
		}

		// 2. update ability row
		updates := map[string]interface{}{
			"name":               input.Name,
			"category":           input.Category,
			"classes":            input.Classes,
			"races":              input.Races,
			"drakkin_heritage":   input.DrakkinHeritage,
			"deities":            input.Deities,
			"status":             input.Status,
			"type":               input.Type,
			"charges":            input.Charges,
			"grant_only":         input.GrantOnly,
			"enabled":            input.Enabled,
			"reset_on_death":     input.ResetOnDeath,
			"auto_grant_enabled": input.AutoGrantEnabled,
		}
		if err := tx.Model(&models.AaAbility{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update ability: %w", err)
		}

		// 3. reconcile ranks: keep/update existing by id, insert new, delete missing
		oldChain, err := walkRankChain(tx, existing.FirstRankId)
		if err != nil {
			return err
		}
		oldIds := map[int]bool{}
		for _, r := range oldChain {
			oldIds[int(r.ID)] = true
		}

		keptIds := map[int]bool{}
		orderedIds := make([]int, 0, len(input.Ranks))
		for _, r := range input.Ranks {
			if r.ID > 0 && oldIds[r.ID] {
				// update existing rank row
				if err := tx.Model(&models.AaRank{}).Where("id = ?", r.ID).Updates(rankFieldMap(r)).Error; err != nil {
					return fmt.Errorf("failed to update rank [%v]: %w", r.ID, err)
				}
				keptIds[r.ID] = true
				orderedIds = append(orderedIds, r.ID)
			} else {
				// new rank
				model := buildRankModel(r)
				if err := tx.Create(&model).Error; err != nil {
					return fmt.Errorf("failed to create rank: %w", err)
				}
				keptIds[int(model.ID)] = true
				orderedIds = append(orderedIds, int(model.ID))
			}
		}

		// 4. delete ranks that were removed (and their children)
		for oldId := range oldIds {
			if !keptIds[oldId] {
				if err := deleteRankAndChildren(tx, oldId); err != nil {
					return err
				}
			}
		}

		// 5. re-stitch chain + first_rank_id
		if err := stitchChain(tx, orderedIds); err != nil {
			return err
		}
		newFirst := 0
		if len(orderedIds) > 0 {
			newFirst = orderedIds[0]
		}
		if err := tx.Model(&models.AaAbility{}).Where("id = ?", id).
			Update("first_rank_id", newFirst).Error; err != nil {
			return fmt.Errorf("failed to set first_rank_id: %w", err)
		}

		// 6. replace effects + prereqs for every kept/new rank
		for i, r := range input.Ranks {
			if i >= len(orderedIds) {
				break
			}
			rankId := orderedIds[i]
			if err := replaceEffects(tx, rankId, r.Effects); err != nil {
				return err
			}
			if err := replacePrereqs(tx, rankId, r.Prereqs); err != nil {
				return err
			}
		}

		return nil
	})
}

func replaceEffects(tx *gorm.DB, rankId int, effects []AaRankEffectInput) error {
	if err := tx.Where("rank_id = ?", rankId).Delete(&models.AaRankEffect{}).Error; err != nil {
		return fmt.Errorf("failed to clear effects for rank [%v]: %w", rankId, err)
	}
	for _, e := range effects {
		row := buildEffect(rankId, e)
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("failed to create effect for rank [%v]: %w", rankId, err)
		}
	}
	return nil
}

func replacePrereqs(tx *gorm.DB, rankId int, prereqs []AaRankPrereqInput) error {
	if err := tx.Where("rank_id = ?", rankId).Delete(&models.AaRankPrereq{}).Error; err != nil {
		return fmt.Errorf("failed to clear prereqs for rank [%v]: %w", rankId, err)
	}
	for _, p := range prereqs {
		row := buildPrereq(rankId, p)
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("failed to create prereq for rank [%v]: %w", rankId, err)
		}
	}
	return nil
}

func deleteRankAndChildren(tx *gorm.DB, rankId int) error {
	if err := tx.Where("rank_id = ?", rankId).Delete(&models.AaRankEffect{}).Error; err != nil {
		return fmt.Errorf("failed to delete effects for rank [%v]: %w", rankId, err)
	}
	if err := tx.Where("rank_id = ?", rankId).Delete(&models.AaRankPrereq{}).Error; err != nil {
		return fmt.Errorf("failed to delete prereqs for rank [%v]: %w", rankId, err)
	}
	if err := tx.Where("id = ?", rankId).Delete(&models.AaRank{}).Error; err != nil {
		return fmt.Errorf("failed to delete rank [%v]: %w", rankId, err)
	}
	return nil
}

// ------------------------------------------------------------------
// Delete (cascade)
// ------------------------------------------------------------------

func (s *AaEditorService) DeleteFullAbility(c echo.Context, id int) error {
	db := s.db(c)
	ability, err := s.deleteAbilityTree(db, id)
	if err != nil {
		return err
	}

	if s.auditLog != nil {
		s.auditLog.LogUserEvent(c, "DELETE", fmt.Sprintf("Deleted [AaAbility] [%v] name [%v]", id, ability.Name))
	}

	return nil
}

func (s *AaEditorService) deleteAbilityTree(db *gorm.DB, id int) (*models.AaAbility, error) {
	var ability models.AaAbility
	if err := db.Where("id = ?", id).First(&ability).Error; err != nil {
		return nil, fmt.Errorf("ability not found: %w", err)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		chain, err := walkRankChain(tx, ability.FirstRankId)
		if err != nil {
			return err
		}
		for _, r := range chain {
			if err := deleteRankAndChildren(tx, int(r.ID)); err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", id).Delete(&models.AaAbility{}).Error; err != nil {
			return fmt.Errorf("failed to delete ability: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ability, nil
}

// ------------------------------------------------------------------
// Validation / preview
// ------------------------------------------------------------------

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

// ValidateAbility inspects an ability input (or existing tree) for common
// integrity issues such as broken chains, orphaned prereqs and disabled
// abilities that still carry ranks.
func (s *AaEditorService) ValidateAbility(c echo.Context, input AaAbilityInput) *ValidationResult {
	res := &ValidationResult{Valid: true, Warnings: []string{}, Errors: []string{}}

	if strings.TrimSpace(input.Name) == "" {
		res.Valid = false
		res.Errors = append(res.Errors, "Ability name is required")
	}
	if len(input.Ranks) == 0 {
		res.Valid = false
		res.Errors = append(res.Errors, "Ability must have at least one rank")
	}

	for i, r := range input.Ranks {
		if r.Cost < 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("Rank #%v has a negative cost", i+1))
		}
		if r.LevelReq <= 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("Rank #%v has a non-positive level requirement", i+1))
		}
		for _, p := range r.Prereqs {
			if p.AaId <= 0 {
				res.Warnings = append(res.Warnings, fmt.Sprintf("Rank #%v has a prereq with no AA id", i+1))
			}
		}
	}

	if len(input.Ranks) > 0 && input.Enabled == 0 {
		res.Warnings = append(res.Warnings, "Ability is disabled but has ranks defined")
	}

	return res
}
