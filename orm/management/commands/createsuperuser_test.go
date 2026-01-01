package management

import (
	"testing"

	"github.com/anuragcarret/djang-drf-go/orm/db"
)

func TestCreateSuperuserCommand_Registration(t *testing.T) {
	cmd := NewCreateSuperuserCommand(&db.DB{})
	if cmd.Name() != "createsuperuser" {
		t.Errorf("Expected name 'createsuperuser', got %s", cmd.Name())
	}
	if cmd.Help() == "" {
		t.Error("Expected help string, got empty")
	}
}

func TestCreateSuperuserCommand_Execution(t *testing.T) {
	// In a real TDD scenario, we'd mock the DB and verify User.Save() is called
}
