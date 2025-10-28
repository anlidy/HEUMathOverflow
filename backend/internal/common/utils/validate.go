package utils

import (
	"MathOverflow/internal/common/config"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var (
	emailRegex = regexp.MustCompile(config.EmailPattern) // 邮箱正则
	imageRegex = regexp.MustCompile(config.ImagePattern) // 图像正则
)

// 验证邮箱是否合法
func IsValidEmail(email string) bool {
	// 使用正则匹配邮箱
	return emailRegex.MatchString(email)
}

// 验证文件是否图片类型
func IsValidImage(filename string) bool {
	return imageRegex.MatchString(filename)
}

// 验证密码哈希是否相同
func ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
