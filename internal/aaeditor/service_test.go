package aaeditor

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/EQEmu/spire/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// newTestService builds an AaEditorService whose DB-based methods can be
// exercised directly. The resolver/auditlog are nil because the db-methods
// under test do not touch them.
func newTestService() *AaEditorService {
	return &AaEditorService{
		resolver: nil,
		auditLog: nil,
		logger:   logger.NewAppLogger(),
	}
}

// testDsn resolves the MySQL DSN to use for integration tests. It honors
// AA_EDITOR_TEST_DSN first, then falls back to the standard eqemu dev env vars.
func testDsn() string {
	if dsn := os.Getenv("AA_EDITOR_TEST_DSN"); dsn != "" {
		return dsn
	}
	user := os.Getenv("MYSQL_EQEMU_USERNAME")
	if user == "" {
		user = "root"
	}
	pass := os.Getenv("MYSQL_EQEMU_PASSWORD")
	host := os.Getenv("MYSQL_EQEMU_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("MYSQL_EQEMU_PORT")
	if port == "" {
		port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		user, pass, host, port)
}

// testDB opens a connection to MySQL, creates a uniquely-named throwaway
// database containing the AA tables, and returns a gorm.DB pointing at it along
// with a cleanup function. Tests skip when MySQL is not reachable.
func testDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	dbName := fmt.Sprintf("spire_aa_test_%d", time.Now().UnixNano())

	mgmt, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                  testDsn(),
		DisableWithReturning: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		t.Skipf("mysql not available, skipping integration test: %v", err)
	}

	if err := mgmt.Exec("CREATE DATABASE IF NOT EXISTS " + dbName + " CHARACTER SET utf8mb4").Error; err != nil {
		t.Skipf("cannot create test database, skipping integration test: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		envOr("MYSQL_EQEMU_USERNAME", "root"),
		os.Getenv("MYSQL_EQEMU_PASSWORD"),
		envOr("MYSQL_EQEMU_HOST", "127.0.0.1"),
		envOr("MYSQL_EQEMU_PORT", "3306"),
		dbName,
	)
	if custom := os.Getenv("AA_EDITOR_TEST_DSN"); custom != "" {
		// custom DSN already points at a usable database; reuse it by name
		dsn = custom
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                  dsn,
		DisableWithReturning: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		_ = mgmt.Exec("DROP DATABASE IF EXISTS " + dbName)
		t.Skipf("cannot open test database, skipping integration test: %v", err)
	}

	if err := createAaTestSchema(db); err != nil {
		_ = mgmt.Exec("DROP DATABASE IF EXISTS " + dbName)
		t.Fatalf("create schema: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		if os.Getenv("AA_EDITOR_TEST_DSN") == "" {
			_ = mgmt.Exec("DROP DATABASE IF EXISTS " + dbName)
		}
		sqlMgmt, _ := mgmt.DB()
		if sqlMgmt != nil {
			_ = sqlMgmt.Close()
		}
	}

	return db, cleanup
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// createAaTestSchema creates the AA-related tables matching the EQEmu schema.
func createAaTestSchema(db *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE aa_ability (
			id INT NOT NULL AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL DEFAULT '',
			category INT NOT NULL DEFAULT 0,
			classes INT NOT NULL DEFAULT 0,
			races INT NOT NULL DEFAULT 0,
			drakkin_heritage INT NOT NULL DEFAULT 0,
			deities INT NOT NULL DEFAULT 0,
			status INT NOT NULL DEFAULT 0,
			type INT NOT NULL DEFAULT 0,
			charges INT NOT NULL DEFAULT 0,
			grant_only TINYINT NOT NULL DEFAULT 0,
			first_rank_id INT NOT NULL DEFAULT 0,
			enabled TINYINT UNSIGNED NOT NULL DEFAULT 0,
			reset_on_death TINYINT NOT NULL DEFAULT 0,
			auto_grant_enabled TINYINT NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE aa_ranks (
			id INT NOT NULL AUTO_INCREMENT,
			upper_hotkey_sid INT NOT NULL DEFAULT 0,
			lower_hotkey_sid INT NOT NULL DEFAULT 0,
			title_sid INT NOT NULL DEFAULT 0,
			desc_sid INT NOT NULL DEFAULT 0,
			cost INT NOT NULL DEFAULT 0,
			level_req INT NOT NULL DEFAULT 0,
			spell INT NOT NULL DEFAULT 0,
			spell_type INT NOT NULL DEFAULT 0,
			recast_time INT NOT NULL DEFAULT 0,
			expansion INT NOT NULL DEFAULT 0,
			prev_id INT NOT NULL DEFAULT 0,
			next_id INT NOT NULL DEFAULT 0,
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE aa_rank_effects (
			rank_id INT NOT NULL,
			slot INT NOT NULL,
			effect_id INT NOT NULL DEFAULT 0,
			base1 INT NOT NULL DEFAULT 0,
			base2 INT NOT NULL DEFAULT 0,
			PRIMARY KEY (rank_id, slot)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE aa_rank_prereqs (
			rank_id INT NOT NULL,
			aa_id INT NOT NULL,
			points INT NOT NULL DEFAULT 0,
			PRIMARY KEY (rank_id, aa_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE db_str (
			id INT NOT NULL,
			type INT NOT NULL,
			value TEXT,
			PRIMARY KEY (id, type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE spells_new (
			id INT NOT NULL,
			name VARCHAR(255) NOT NULL DEFAULT '',
			PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	return nil
}

func sampleAbility(name string, ranks int) AaAbilityInput {
	input := AaAbilityInput{
		Name:     name,
		Category: 1,
		Classes:  1,
		Enabled:  1,
		Type:     0,
	}
	for i := 0; i < ranks; i++ {
		r := AaRankInput{
			TempId:    fmt.Sprintf("r%d", i),
			Cost:      (i + 1) * 2,
			LevelReq:  50 + i,
			Expansion: 7,
		}
		r.Effects = append(r.Effects, AaRankEffectInput{Slot: 1, EffectId: 79, Base1: i + 1, Base2: 0})
		if i > 0 {
			r.Prereqs = append(r.Prereqs, AaRankPrereqInput{AaId: 999, Points: i})
		}
		input.Ranks = append(input.Ranks, r)
	}
	return input
}

// ------------------------------------------------------------------

func TestCreateFullAbility(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	id, err := svc.createAbilityTree(db, sampleAbility("Test Create", 3))
	if err != nil {
		t.Fatalf("createAbilityTree: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ability id, got %d", id)
	}

	full, err := svc.getFullAbility(db, id)
	if err != nil {
		t.Fatalf("getFullAbility: %v", err)
	}

	if full.AaAbility.Name != "Test Create" {
		t.Fatalf("unexpected name %q", full.AaAbility.Name)
	}
	if len(full.Ranks) != 3 {
		t.Fatalf("expected 3 ranks, got %d", len(full.Ranks))
	}

	// chain stitching
	if full.AaAbility.FirstRankId != int(full.Ranks[0].AaRank.ID) {
		t.Fatalf("first_rank_id mismatch: %d vs %d", full.AaAbility.FirstRankId, full.Ranks[0].AaRank.ID)
	}
	if full.Ranks[0].AaRank.PrevId != 0 {
		t.Fatalf("first rank prev_id should be 0, got %d", full.Ranks[0].AaRank.PrevId)
	}
	if full.Ranks[0].AaRank.NextId != int(full.Ranks[1].AaRank.ID) {
		t.Fatalf("rank0 next_id should point to rank1")
	}
	if full.Ranks[2].AaRank.NextId != 0 {
		t.Fatalf("last rank next_id should be 0, got %d", full.Ranks[2].AaRank.NextId)
	}
	if full.Ranks[1].AaRank.PrevId != int(full.Ranks[0].AaRank.ID) {
		t.Fatalf("rank1 prev_id should point to rank0")
	}

	// children
	if len(full.Ranks[0].Effects) != 1 {
		t.Fatalf("expected 1 effect on rank0, got %d", len(full.Ranks[0].Effects))
	}
	if len(full.Ranks[1].Prereqs) != 1 {
		t.Fatalf("expected 1 prereq on rank1, got %d", len(full.Ranks[1].Prereqs))
	}
}

func TestGetFullAbility(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	id, _ := svc.createAbilityTree(db, sampleAbility("Test Get", 2))

	// seed a string + spell to verify resolution
	db.Exec("INSERT INTO db_str (id, type, value) VALUES (?, 1, ?)", 5000, "Rank Title")
	db.Exec("UPDATE aa_ranks SET title_sid = 5000 WHERE id = ?", id)
	// find first rank id
	var firstRank int
	db.Raw("SELECT first_rank_id FROM aa_ability WHERE id = ?", id).Scan(&firstRank)
	db.Exec("UPDATE aa_ranks SET title_sid = 5000 WHERE id = ?", firstRank)
	db.Exec("INSERT INTO spells_new (id, name) VALUES (?, ?)", 1234, "Flame Strike")
	db.Exec("UPDATE aa_ranks SET spell = 1234 WHERE id = ?", firstRank)

	full, err := svc.getFullAbility(db, id)
	if err != nil {
		t.Fatalf("getFullAbility: %v", err)
	}
	if full.Ranks[0].Strings[1].Value != "Rank Title" {
		t.Fatalf("expected resolved title string, got %#v", full.Ranks[0].Strings[1])
	}
	if full.Ranks[0].Spell == nil || full.Ranks[0].Spell.Name != "Flame Strike" {
		t.Fatalf("expected resolved spell name, got %#v", full.Ranks[0].Spell)
	}
}

func TestSaveFullAbilityReorderAndReconcile(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	id, _ := svc.createAbilityTree(db, sampleAbility("Test Save", 3))
	full, _ := svc.getFullAbility(db, id)

	// build input: keep rank 0 + rank 2, drop rank 1, add a new rank, reorder
	var kept []AaRankFull
	kept = append(kept, full.Ranks[0])
	kept = append(kept, full.Ranks[2])

	input := AaAbilityInput{
		Name:    "Test Save (Edited)",
		Enabled: 1,
	}
	for _, r := range kept {
		ri := AaRankInput{
			ID:       int(r.AaRank.ID),
			Cost:     r.AaRank.Cost + 10,
			LevelReq: r.AaRank.LevelReq,
		}
		for _, e := range r.Effects {
			ri.Effects = append(ri.Effects, AaRankEffectInput{Slot: e.Slot, EffectId: e.EffectId, Base1: e.Base1, Base2: e.Base2})
		}
		input.Ranks = append(input.Ranks, ri)
	}
	// new rank at the end
	input.Ranks = append(input.Ranks, AaRankInput{Cost: 99, LevelReq: 90, Effects: []AaRankEffectInput{{Slot: 1, EffectId: 1}}})

	if err := svc.saveAbilityTree(db, id, input); err != nil {
		t.Fatalf("saveAbilityTree: %v", err)
	}

	after, err := svc.getFullAbility(db, id)
	if err != nil {
		t.Fatalf("getFullAbility: %v", err)
	}

	if after.AaAbility.Name != "Test Save (Edited)" {
		t.Fatalf("name not updated: %q", after.AaAbility.Name)
	}
	if len(after.Ranks) != 3 {
		t.Fatalf("expected 3 ranks after save, got %d", len(after.Ranks))
	}

	// chain integrity
	if after.Ranks[0].AaRank.NextId != int(after.Ranks[1].AaRank.ID) {
		t.Fatalf("chain broken after save")
	}
	if after.Ranks[2].AaRank.NextId != 0 {
		t.Fatalf("last rank next_id should be 0")
	}

	// dropped rank's children should be gone
	var count int64
	db.Table("aa_rank_effects").Where("rank_id = ?", full.Ranks[1].AaRank.ID).Count(&count)
	if count != 0 {
		t.Fatalf("dropped rank effects should be removed, got %d", count)
	}
}

func TestDeleteFullAbility(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	id, _ := svc.createAbilityTree(db, sampleAbility("Test Delete", 2))
	full, _ := svc.getFullAbility(db, id)

	// seed a string + spell that must survive deletion
	firstRank := full.Ranks[0].AaRank.ID
	db.Exec("INSERT INTO spells_new (id, name) VALUES (?, ?)", 5555, "Persisted Spell")
	db.Exec("INSERT INTO db_str (id, type, value) VALUES (?, 1, ?)", 6000, "Persisted String")

	if _, err := svc.deleteAbilityTree(db, id); err != nil {
		t.Fatalf("deleteAbilityTree: %v", err)
	}

	// ability + ranks + children gone
	var abilityCount int64
	db.Table("aa_ability").Where("id = ?", id).Count(&abilityCount)
	if abilityCount != 0 {
		t.Fatalf("ability should be deleted")
	}
	var rankCount int64
	db.Table("aa_ranks").Where("id = ?", firstRank).Count(&rankCount)
	if rankCount != 0 {
		t.Fatalf("rank should be deleted")
	}
	var effectCount int64
	db.Table("aa_rank_effects").Where("rank_id = ?", firstRank).Count(&effectCount)
	if effectCount != 0 {
		t.Fatalf("effects should be deleted")
	}

	// shared reference data preserved
	var spellCount int64
	db.Table("spells_new").Where("id = ?", 5555).Count(&spellCount)
	if spellCount != 1 {
		t.Fatalf("spells_new should be preserved")
	}
	var strCount int64
	db.Table("db_str").Where("id = ?", 6000).Count(&strCount)
	if strCount != 1 {
		t.Fatalf("db_str should be preserved")
	}
}

func TestDuplicateAbility(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	srcId, _ := svc.createAbilityTree(db, sampleAbility("Test Dup Source", 2))
	srcFull, _ := svc.getFullAbility(db, srcId)

	// duplicate via the create path with a deep-copied input
	input := AaAbilityInput{
		Name:     "Test Dup Copy",
		Category: srcFull.AaAbility.Category,
		Classes:  srcFull.AaAbility.Classes,
		Enabled:  1,
	}
	for _, r := range srcFull.Ranks {
		ri := AaRankInput{
			Cost:     r.AaRank.Cost,
			LevelReq: r.AaRank.LevelReq,
		}
		for _, e := range r.Effects {
			ri.Effects = append(ri.Effects, AaRankEffectInput{Slot: e.Slot, EffectId: e.EffectId, Base1: e.Base1, Base2: e.Base2})
		}
		input.Ranks = append(input.Ranks, ri)
	}

	newId, err := svc.createAbilityTree(db, input)
	if err != nil {
		t.Fatalf("createAbilityTree (dup): %v", err)
	}
	if newId == srcId {
		t.Fatalf("duplicate should have a new id")
	}

	newFull, err := svc.getFullAbility(db, newId)
	if err != nil {
		t.Fatalf("getFullAbility (dup): %v", err)
	}
	if len(newFull.Ranks) != len(srcFull.Ranks) {
		t.Fatalf("expected %d ranks in copy, got %d", len(srcFull.Ranks), len(newFull.Ranks))
	}
	if len(newFull.Ranks[0].Effects) != len(srcFull.Ranks[0].Effects) {
		t.Fatalf("child count mismatch")
	}

	// source untouched
	srcAgain, _ := svc.getFullAbility(db, srcId)
	if srcAgain.AaAbility.Name != "Test Dup Source" {
		t.Fatalf("source name changed: %q", srcAgain.AaAbility.Name)
	}
}

func TestTransactionRollback(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()

	// pre-validation failure (empty name) -> no rows
	input := sampleAbility("Test Rollback", 1)
	input.Name = ""
	_, err := svc.createAbilityTree(db, input)
	if err == nil {
		t.Fatalf("expected error for empty name")
	}

	// mid-transaction failure: create an ability, then attempt a second create
	// whose rank forces an explicit id colliding with the first ability's rank.
	seedId, _ := svc.createAbilityTree(db, sampleAbility("Rollback Seed", 1))
	seedFull, _ := svc.getFullAbility(db, seedId)
	existingRankId := int(seedFull.Ranks[0].AaRank.ID)

	collide := AaAbilityInput{
		Name:    "Rollback Collide",
		Enabled: 1,
		Ranks: []AaRankInput{
			{ID: existingRankId, Cost: 5, LevelReq: 50}, // dup PK -> tx rollback
		},
	}
	_, err = svc.createAbilityTree(db, collide)
	if err == nil {
		t.Fatalf("expected mid-transaction error from duplicate rank id")
	}

	// the collided ability must not have been persisted
	var count int64
	db.Table("aa_ability").Where("name = ?", "Rollback Collide").Count(&count)
	if count != 0 {
		t.Fatalf("expected rollback to prevent the collided ability, got %d rows", count)
	}

	// the seed ability is intact
	var seedCount int64
	db.Table("aa_ability").Where("id = ?", seedId).Count(&seedCount)
	if seedCount != 1 {
		t.Fatalf("seed ability should be intact after rollback")
	}
}

func TestListAbilities(t *testing.T) {
	db, cleanup := testDB(t)
	defer cleanup()

	svc := newTestService()
	svc.createAbilityTree(db, sampleAbility("Apple Ability", 1))
	svc.createAbilityTree(db, sampleAbility("Banana Ability", 2))

	res, err := svc.listAbilities(db, AaListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("listAbilities: %v", err)
	}
	if res.Total < 2 {
		t.Fatalf("expected at least 2 abilities, got %d", res.Total)
	}

	// search filter
	res, _ = svc.listAbilities(db, AaListFilter{Search: "Banana", Limit: 10})
	if res.Total != 1 {
		t.Fatalf("expected 1 result for 'Banana', got %d", res.Total)
	}
	if res.Items[0].RankCount != 2 {
		t.Fatalf("expected rank count 2, got %d", res.Items[0].RankCount)
	}
}
