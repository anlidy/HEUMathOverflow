package service

import (
	"MathOverflow/internal/common/config"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/user-service/model"
	syserror "MathOverflow/internal/user-service/model/error"
	"MathOverflow/internal/user-service/model/request"
	"MathOverflow/internal/user-service/model/response"
	"MathOverflow/internal/user-service/repository"
	"context"
	"log"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// 用户服务接口
type UserService interface {
	UserRegister(ctx context.Context, req request.UserRegister) (model.Session, response.UserInfo, syserror.Error)
	UserLogin(ctx context.Context, req request.UserLogin) (model.Session, response.UserInfo, syserror.Error)
	UserUploadAvatar(ctx context.Context, userID int64, file model.File) (string, syserror.Error)
	UserDownloadAvatar(ctx context.Context, filename string) (*model.File, syserror.Error)
}

// 用户服务类型
type userService struct {
	cfg         config.Config
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
	fileRepo    repository.FileRepo
	serviceName string
}

func NewUserService(cfg config.Config, userRepo repository.UserRepo, sessionRepo repository.SessionRepo, fileRepo repository.FileRepo) UserService {
	return &userService{cfg: cfg, userRepo: userRepo, sessionRepo: sessionRepo, fileRepo: fileRepo, serviceName: "User-Service"}
}

// 用户注册
func (s *userService) UserRegister(ctx context.Context, req request.UserRegister) (model.Session, response.UserInfo, syserror.Error) {
	// 校验用户邮箱格式
	if !utils.IsValidEmail(req.Email) {
		return model.Session{}, response.UserInfo{}, syserror.EmailError
	}
	// 密码哈希加密
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Printf("[%s] %v", s.serviceName, err)
		return model.Session{}, response.UserInfo{}, syserror.InternalError
	}

	// user结构体
	var user = model.User{
		ID:           utils.GenerateSnowflakeID(), // 生成userID
		Email:        req.Email,
		Username:     req.Username,
		Role:         model.Student,
		PasswordHash: passwordHash,
		LastLogin:    time.Now(),
	}

	// 查询用户是否已经存在
	_, err = s.userRepo.FindUserByEmail(user.Email)
	if err == nil {
		return model.Session{}, response.UserInfo{}, syserror.EmailExistsError
	}
	_, err = s.userRepo.FindUserByUsername(user.Username)
	if err == nil {
		return model.Session{}, response.UserInfo{}, syserror.NameExistsError
	}
	// 将新用户存入数据库
	err = s.userRepo.CreateUser(&user)
	if err != nil {
		log.Printf("[%s] %v", s.serviceName, err)
		return model.Session{}, response.UserInfo{}, syserror.InternalError
	}

	// 将sessionID -> Session 存入redis
	var sessionID = utils.GenerateUUID()
	var session = model.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		Remember:  false,
		TTL:       time.Hour * 24,
	}
	err = s.sessionRepo.SetSession(ctx, session) // 默认有效期为1天
	if err != nil {
		log.Printf("[%s] %v", s.serviceName, err)
		return model.Session{}, response.UserInfo{}, syserror.InternalError
	}

	// 返回用户基本信息
	var info = response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     model.GetRoleName(user.Role),
	}
	return session, info, syserror.NoError
}

// 用户登录
func (s *userService) UserLogin(ctx context.Context, req request.UserLogin) (model.Session, response.UserInfo, syserror.Error) {
	// 校验用户邮箱格式
	if !utils.IsValidEmail(req.Email) {
		return model.Session{}, response.UserInfo{}, syserror.EmailError
	}
	// 查询邮箱所属用户
	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.Session{}, response.UserInfo{}, syserror.NotFoundError
		}
		log.Printf("[%s] %v\n", s.serviceName, err)
		return model.Session{}, response.UserInfo{}, syserror.InternalError
	}
	// 校验密码是否正确
	if !utils.ValidatePassword(user.PasswordHash, req.Password) {
		return model.Session{}, response.UserInfo{}, syserror.PasswordError
	}

	// 将sessionID -> Session 存入redis
	var sessionID = utils.GenerateUUID()
	var ttl time.Duration
	if req.Remember {
		ttl = time.Hour * 24 * 30 // 保存30天
	} else {
		ttl = time.Hour * 24 //保存1天
	}
	var session = model.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		Remember:  req.Remember,
		TTL:       ttl,
	}
	err = s.sessionRepo.SetSession(ctx, session)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return model.Session{}, response.UserInfo{}, syserror.InternalError
	}

	var info = response.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      model.GetRoleName(user.Role),
		AvatarUrl: user.AvatarUrl,
	}
	return session, info, syserror.NoError
}

// 上传用户头像
func (s *userService) UserUploadAvatar(ctx context.Context, userID int64, file model.File) (string, syserror.Error) {
	// 将文件存入minio
	url, err := s.fileRepo.UploadFile(ctx, s.cfg.Minio.Bucket, file)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return "", syserror.InternalError
	}
	// 更新用户信息
	_, err = s.userRepo.UpdateColumn(userID, "avatar_url", url)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		err := s.fileRepo.DeleteFile(ctx, file.Filename) // 删除上传的文件
		if err != nil {
			log.Printf("[%s] %v\n", s.serviceName, err)
		}
		return "", syserror.InternalError
	}
	return url, syserror.NoError
}

// 下载用户头像
func (s *userService) UserDownloadAvatar(ctx context.Context, filename string) (*model.File, syserror.Error) {
	// 检查文件是否存在
	exists, err := s.fileRepo.FileExists(ctx, filename)
	if err != nil {
		return nil, syserror.InternalError
	} else if !exists {
		return nil, syserror.NotFoundError
	}
	// 从minio流式读取文件
	file, err := s.fileRepo.DownloadFile(ctx, filename)
	if err != nil {
		return nil, syserror.InternalError
	}
	return file, syserror.NoError
}

// 修改用户信息
func (s *userService) UserUpdateProfie(ctx context.Context, userID int64, req request.UserProfie) syserror.Error {
	// 获取用户记录
	user, err := s.userRepo.FindUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}
	// 覆盖修改的字段
	err = copier.CopyWithOption(&user, &req, copier.Option{
		IgnoreEmpty: true, // 忽略 req 中的空值
	})
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}
	// 更新数据库的用户记录
	_, err = s.userRepo.UpdateUser(&user)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}
	return syserror.NoError
}

// 更新用户密码
func (s *userService) UserUpdatePassword(ctx context.Context, userID int64, req request.UserPassword) syserror.Error {
	// 获取用户记录
	user, err := s.userRepo.FindUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}

	// 校验密码哈希是否相同
	if !utils.ValidatePassword(user.PasswordHash, req.OldPassword) {
		return syserror.PasswordError
	}

	// 生成新密码哈希
	newPwdHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}

	// 将新密码哈希存入用户记录
	user.PasswordHash = newPwdHash
	_, err = s.userRepo.UpdateUser(&user)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}
	return syserror.NoError
}

// 更新用户角色
func (s *userService) UpdateUserRole(ctx context.Context, opID int64, req request.UserRole) syserror.Error {
	operator, err1 := s.userRepo.FindUserByID(opID)
	target, err2 := s.userRepo.FindUserByID(req.ID)
	if err1 == gorm.ErrRecordNotFound || err2 == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err1 != nil || err2 != nil {
		log.Printf("[%s] %v %v\n", s.serviceName, err1, err2)
		return syserror.InternalError
	}
	// 无效的Role, 操作者role不是Admin且role权限低于目标权限时,报错返回
	if !model.ValidateRole(req.NewRole) || operator.Role != model.Admin && operator.Role <= req.NewRole {
		return syserror.PermissionDeniedError
	}
	target.Role = req.NewRole
	// 更新数据库
	_, err := s.userRepo.UpdateUser(&target)
	if err != nil {
		log.Printf("[%s] %v\n", s.serviceName, err)
		return syserror.InternalError
	}

	// 更新redis
	err = s.sessionRepo.UpdateUserRole(ctx, target.ID, target.Role)
	if err != nil {
		return syserror.InternalError
	}

	return syserror.NoError
}
