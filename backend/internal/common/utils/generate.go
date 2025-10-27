package utils

import (
	"MathOverflow/internal/common/config"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	sf *snowflake.Node
)

func init() {
	var err error
	sf, err = snowflake.NewNode(int64(config.MachineID))
	if err != nil {
		fmt.Println(err)
		return
	}
}

// 生成 Snowflake ID
func GenerateSnowflakeID() int64 {
	return sf.Generate().Int64()
}

// 生成uuid
func GenerateUUID() string {
	return uuid.New().String() // 36位长度
}

// 生成密码哈希
func HashPassword(password string) (string, error) {
	hashed_pwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed_pwd), nil
}
