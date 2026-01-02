package management

import (
	"context"
	"fmt"
	"reflect"

	"github.com/anuragcarret/djang-drf-go/contrib/auth"
	"github.com/anuragcarret/djang-drf-go/orm/db"
	"github.com/anuragcarret/djang-drf-go/orm/queryset"
)

type CreateSuperuserCommand[T auth.Authenticatable] struct {
	database *db.DB
}

func NewCreateSuperuserCommand[T auth.Authenticatable](database *db.DB) *CreateSuperuserCommand[T] {
	return &CreateSuperuserCommand[T]{database: database}
}

func (c *CreateSuperuserCommand[T]) Name() string { return "createsuperuser" }
func (c *CreateSuperuserCommand[T]) Help() string { return "Create a superuser" }

func (c *CreateSuperuserCommand[T]) Run(ctx context.Context, args []string) error {
	var username, email, password string

	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Email address: ")
	fmt.Scanln(&email)
	fmt.Print("Password: ")
	fmt.Scanln(&password)

	// Use reflection to create a new instance of T
	var user T
	userType := reflect.TypeOf(user)
	if userType.Kind() == reflect.Ptr {
		userType = userType.Elem()
		user = reflect.New(userType).Interface().(T)
	} else {
		// Should generally be a pointer, but handle value type if needed
		user = reflect.New(userType).Elem().Interface().(T)
	}

	// We need a way to set these fields generically.
	// Since T is Authenticatable, we might need a more specific interface or just use reflection.
	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Helper to set field by name if it exists
	setField := func(name string, val interface{}) {
		f := v.FieldByName(name)
		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(val))
		} else {
			// Try finding in embedded User if it's there
			for i := 0; i < v.NumField(); i++ {
				field := v.Field(i)
				if field.Kind() == reflect.Struct {
					subF := field.FieldByName(name)
					if subF.IsValid() && subF.CanSet() {
						subF.Set(reflect.ValueOf(val))
						return
					}
				}
			}
		}
	}

	setField("Username", username)
	setField("Email", email)
	setField("IsSuperuser", true)
	setField("IsStaff", true)
	setField("IsActive", true)

	user.SetPassword(password)

	qs := queryset.NewQuerySet[T](c.database)
	if err := qs.Create(user); err != nil {
		return fmt.Errorf("failed to create superuser: %v", err)
	}

	fmt.Printf("Superuser '%s' created successfully.\n", username)
	return nil
}
