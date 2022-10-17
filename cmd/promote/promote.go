package promote

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/entity"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/glebarez/sqlite"
	"github.com/urfave/cli"
	"gorm.io/gorm"
)

func NewPromoteCommand() cli.Command {
	return cli.Command{
		Name:   "promote",
		Action: promote,
	}
}

func promote(cCtx *cli.Context) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(cfg.LiteDB), &gorm.Config{})
	if err != nil {
		return err
	}
	db.Debug()

	var user entity.User
	email := strings.ToLower(cCtx.Args().First())
	if email == "" {
		return fmt.Errorf("email is required")
	}
	result := db.First(&user, &entity.User{Email: email})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return fmt.Errorf("user %s not found", email)
	}

	fmt.Println(email, user, result.Error)

	role := model.Role(cCtx.Args().Get(1))
	if !role.IsValid() {
		return fmt.Errorf("role %s is invalid", role)
	}

	user.Roles = append(user.Roles, string(role))
	db.Save(&user)

	return nil
}
