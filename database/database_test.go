package database

import (
	"errors"
	"path/filepath"
	"strconv"
	"testing"

	"datalchemist/models"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

func TestBootstrapAndResetAdmin(t *testing.T) {
	dbGorm = nil
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
	defer func() { dbGorm = nil }()

	if err := Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	hasUsers, err := HasUsers()
	if err != nil {
		t.Fatalf("check users: %v", err)
	}
	if hasUsers {
		t.Fatal("a new database must not contain a default user")
	}

	if err := BootstrapAdmin("operator", "Initial-Password1"); err != nil {
		t.Fatalf("bootstrap administrator: %v", err)
	}
	if err := BootstrapAdmin("second", "Another-Password1"); err == nil {
		t.Fatal("bootstrapping must fail after the first user exists")
	}

	user, err := UserGet("operator")
	if err != nil {
		t.Fatalf("load bootstrapped administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("Initial-Password1")); err != nil {
		t.Fatalf("verify bootstrapped password: %v", err)
	}
	isAdmin, err := UserIdIsAdmin(user.ID)
	if err != nil || !isAdmin {
		t.Fatalf("bootstrapped user is not an administrator: admin=%t, err=%v", isAdmin, err)
	}

	if err := ResetAdminPassword("operator", "Replacement-Password1"); err != nil {
		t.Fatalf("reset administrator password: %v", err)
	}
	user, err = UserGet("operator")
	if err != nil {
		t.Fatalf("load reset administrator: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("Replacement-Password1")); err != nil {
		t.Fatalf("verify reset password: %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	rejected := map[string]string{
		"too short":    "Short-1a",
		"no lowercase": "UPPERCASE-123",
		"no uppercase": "lowercase-123",
		"no digit":     "NoDigitHere-x",
		"no special":   "NoSpecialHere1",
		"empty":        "",
	}
	for reason, password := range rejected {
		if err := ValidatePassword(password); err == nil {
			t.Errorf("password %q was accepted (%s)", password, reason)
		}
	}

	for _, password := range []string{"Initial-Password1", "Sup3r Sécurisé!"} {
		if err := ValidatePassword(password); err != nil {
			t.Errorf("compliant password %q was rejected: %v", password, err)
		}
	}
}

func TestHashPasswordRejectsWeakPassword(t *testing.T) {
	if _, err := HashPassword("too-short"); err == nil {
		t.Fatal("weak administrator password was accepted")
	}
}

func TestDuplicateCopiesEntityAndLinks(t *testing.T) {
	dbGorm = nil
	viper.Set("database", filepath.Join(t.TempDir(), "test.sqlite"))
	defer func() { dbGorm = nil }()

	if err := Init(); err != nil {
		t.Fatalf("initialize database: %v", err)
	}

	baseID, err := SourceUpdate(models.Sources{Name: "dup base", JSON: `{"src":"text"}`})
	if err != nil {
		t.Fatalf("create base source: %v", err)
	}
	sourceID, err := SourceUpdate(models.Sources{Name: "dup source", Parameters: "params", JSON: `{"src":"url"}`})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	SourceAddRequire(models.Source_require{Source: sourceID, Require: baseID})

	copyID, err := SourceDuplicate(strconv.Itoa(int(sourceID)), "")
	if err != nil {
		t.Fatalf("duplicate source: %v", err)
	}
	if copyID == sourceID || copyID == 0 {
		t.Fatalf("duplicate source id = %d", copyID)
	}
	sourceCopy, err := SourceGet(strconv.Itoa(int(copyID)))
	if err != nil {
		t.Fatalf("load duplicated source: %v", err)
	}
	if sourceCopy.Name != "dup source_1" {
		t.Fatalf("duplicated source name = %q", sourceCopy.Name)
	}
	if sourceCopy.Parameters != "params" || sourceCopy.JSON != `{"src":"url"}` {
		t.Fatalf("duplicated source content = %+v", sourceCopy)
	}
	requires, err := SourceRequire(strconv.Itoa(int(copyID)))
	if err != nil {
		t.Fatalf("load duplicated source requires: %v", err)
	}
	if len(requires) != 1 || requires[0].ID != baseID {
		t.Fatalf("duplicated source requires = %+v, want [%d]", requires, baseID)
	}

	// Un second doublon sans nom doit être suffixé « _2 ».
	secondID, err := SourceDuplicate(strconv.Itoa(int(sourceID)), "")
	if err != nil {
		t.Fatalf("duplicate source twice: %v", err)
	}
	second, err := SourceGet(strconv.Itoa(int(secondID)))
	if err != nil {
		t.Fatalf("load second duplicate: %v", err)
	}
	if second.Name != "dup source_2" {
		t.Fatalf("second duplicate name = %q", second.Name)
	}

	itemID, err := ItemUpdate(models.Items{Name: "dup item", Parameters: "p", Template: "tpl", Javascript: "js"})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	ItemAddRequire(models.Item_sources{Item: itemID, Source: sourceID})

	itemCopyID, err := ItemDuplicate(strconv.Itoa(int(itemID)), "custom name")
	if err != nil {
		t.Fatalf("duplicate item: %v", err)
	}
	itemCopy, err := ItemGet(strconv.Itoa(int(itemCopyID)))
	if err != nil {
		t.Fatalf("load duplicated item: %v", err)
	}
	if itemCopy.Name != "custom name" || itemCopy.Template != "tpl" || itemCopy.Javascript != "js" || itemCopy.Parameters != "p" {
		t.Fatalf("duplicated item = %+v", itemCopy)
	}
	itemSources, err := ItemSources(strconv.Itoa(int(itemCopyID)))
	if err != nil {
		t.Fatalf("load duplicated item sources: %v", err)
	}
	if len(itemSources) != 1 || itemSources[0].ID != sourceID {
		t.Fatalf("duplicated item sources = %+v, want [%d]", itemSources, sourceID)
	}

	viewID, err := ViewAdd(models.Views{Name: "dup view", Parameters: `{"version":2,"items":[{"itemid":1}]}`, Protected: true})
	if err != nil {
		t.Fatalf("create view: %v", err)
	}
	viewCopyID, err := ViewDuplicate(strconv.Itoa(int(viewID)), "")
	if err != nil {
		t.Fatalf("duplicate view: %v", err)
	}
	viewCopy, err := ViewGet(strconv.Itoa(int(viewCopyID)))
	if err != nil {
		t.Fatalf("load duplicated view: %v", err)
	}
	if viewCopy.Name != "dup view_1" || viewCopy.Parameters != `{"version":2,"items":[{"itemid":1}]}` || !viewCopy.Protected {
		t.Fatalf("duplicated view = %+v", viewCopy)
	}

	// Un nom demandé déjà pris doit être refusé.
	if _, err := SourceDuplicate(strconv.Itoa(int(sourceID)), "dup base"); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("duplicate with taken name: err = %v, want ErrNameTaken", err)
	}

	if _, err := SourceDuplicate("404404", ""); err == nil {
		t.Fatal("duplicating a missing source must fail")
	}
}
