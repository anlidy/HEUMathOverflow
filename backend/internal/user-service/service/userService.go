package service

import (
	"MathOverflow/internal/common/config"
	common "MathOverflow/internal/common/model"
	"MathOverflow/internal/common/utils"
	"MathOverflow/internal/user-service/model"
	syserror "MathOverflow/internal/user-service/model/error"
	"MathOverflow/internal/user-service/model/request"
	"MathOverflow/internal/user-service/model/response"
	"MathOverflow/internal/user-service/repository"
	userpb "MathOverflow/proto/user"
	"context"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// 用户服务接口
type UserService interface {
	RPCGetUserInfo(ctx context.Context, userID int64) (*userpb.GetUserResponse, syserror.Error)
	RPCBatchGetUserInfo(ctx context.Context, userIDs []int64) (map[int64]*userpb.GetUserResponse, syserror.Error)
	UserRegister(ctx context.Context, req request.UserRegister) (*model.Session, *response.UserInfo, syserror.Error)
	UserLogin(ctx context.Context, req request.UserLogin) (*model.Session, *response.UserInfo, syserror.Error)
	UserUploadAvatar(ctx context.Context, userID int64, file common.File) (string, syserror.Error)
	UserDownloadAvatar(ctx context.Context, filename string) (*common.File, syserror.Error)
	UserUpdateProfie(ctx context.Context, userID int64, req request.UserProfie) syserror.Error
	UserUpdatePassword(ctx context.Context, userID int64, req request.UserPassword) syserror.Error
	UpdateUserRole(ctx context.Context, opID int64, req request.UserRole) syserror.Error
}

// 用户服务类型
type userService struct {
	cfg         config.Config
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
	fileRepo    repository.FileRepo
	servName    string
}

func NewUserService(cfg config.Config, userRepo repository.UserRepo, sessionRepo repository.SessionRepo, fileRepo repository.FileRepo) UserService {
	return &userService{cfg: cfg, userRepo: userRepo, sessionRepo: sessionRepo, fileRepo: fileRepo, servName: "User-Service"}
}

// 获取单个用户信息
func (s *userService) RPCGetUserInfo(ctx context.Context, userID int64) (*userpb.GetUserResponse, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	user, err := s.userRepo.FindUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, syserror.NotFoundError
		}
		logger.WithError(err).Error("find user by id failed")
		return nil, syserror.InternalError
	}
	info := &userpb.GetUserResponse{
		UserId:    user.ID,
		Username:  user.Username,
		Role:      int64(user.Role),
		AvatarUrl: user.AvatarUrl,
	}
	return info, syserror.NoError
}

// 批量获取用户信息
func (s *userService) RPCBatchGetUserInfo(ctx context.Context, userIDs []int64) (map[int64]*userpb.GetUserResponse, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	users, err := s.userRepo.BatchFindUserByID(userIDs)
	if err != nil {
		logger.WithError(err).Error("batch find user by id failed")
		return nil, syserror.InternalError
	}
	infoMap := make(map[int64]*userpb.GetUserResponse, len(users))
	for _, user := range users {
		infoMap[user.ID] = &userpb.GetUserResponse{
			UserId:    user.ID,
			Username:  user.Username,
			Role:      int64(user.Role),
			AvatarUrl: user.AvatarUrl,
		}
	}
	return infoMap, syserror.NoError
}

// 用户注册
func (s *userService) UserRegister(ctx context.Context, req request.UserRegister) (*model.Session, *response.UserInfo, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 校验用户邮箱格式
	if !utils.IsValidEmail(req.Email) {
		return nil, nil, syserror.EmailError
	}
	// 密码哈希加密
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.WithError(err).Error("hash password failed")
		return nil, nil, syserror.InternalError
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
		return nil, nil, syserror.EmailExistsError
	}
	_, err = s.userRepo.FindUserByUsername(user.Username)
	if err == nil {
		return nil, nil, syserror.NameExistsError
	}
	// 将新用户存入数据库
	err = s.userRepo.CreateUser(&user)
	if err != nil {
		logger.WithError(err).Error("create user failed")
		return nil, nil, syserror.InternalError
	}

	// 将sessionID -> Session 存入redis
	var sessionID = utils.GenerateUUID()
	var session = &model.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		Remember:  false,
		TTL:       time.Hour * 24,
	}
	err = s.sessionRepo.SetSession(ctx, *session) // 默认有效期为1天
	if err != nil {
		logger.WithError(err).Error("set session failed")
		return nil, nil, syserror.InternalError
	}

	// 返回用户基本信息
	var info = &response.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}
	return session, info, syserror.NoError
}

// 用户登录
func (s *userService) UserLogin(ctx context.Context, req request.UserLogin) (*model.Session, *response.UserInfo, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 校验用户邮箱格式
	if !utils.IsValidEmail(req.Email) {
		return nil, nil, syserror.EmailError
	}

	// 查询邮箱所属用户
	user, err := s.userRepo.FindUserByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, syserror.NotFoundError
		}
		logger.WithError(err).Error("find user by email failed")
		return nil, nil, syserror.InternalError
	}
	// 校验密码是否正确
	if !utils.ValidatePassword(user.PasswordHash, req.Password) {
		return nil, nil, syserror.PasswordError
	}

	// 将sessionID -> Session 存入redis
	var sessionID = utils.GenerateUUID()
	var ttl time.Duration
	if req.Remember {
		ttl = time.Hour * 24 * 30 // 保存30天
	} else {
		ttl = time.Hour * 24 //保存1天
	}
	var session = &model.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Role:      user.Role,
		Remember:  req.Remember,
		TTL:       ttl,
	}
	err = s.sessionRepo.SetSession(ctx, *session)
	if err != nil {
		logger.WithError(err).Error("set session failed")
		return nil, nil, syserror.InternalError
	}

	// 更新用户上次登录时间
	_, err = s.userRepo.UpdateColumn(user.ID, "last_login", time.Now())
	if err != nil {
		logger.WithError(err).Warn("update last_login failed")
	}

	// 返回用户信息
	var info = &response.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		AvatarUrl: user.AvatarUrl,
	}
	return session, info, syserror.NoError
}

// 上传用户头像
func (s *userService) UserUploadAvatar(ctx context.Context, userID int64, file common.File) (string, syserror.Error) {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 将文件存入minio
	url, err := s.fileRepo.UploadFile(ctx, s.cfg.Minio.Bucket, file)
	if err != nil {
		logger.WithError(err).Error("upload avatar failed")
		return "", syserror.InternalError
	}
	// 更新用户信息
	_, err = s.userRepo.UpdateColumn(userID, "avatar_url", url)
	if err != nil {
		logger.WithError(err).Error("update avatar_url failed")
		err := s.fileRepo.DeleteFile(ctx, file.Filename) // 删除上传的文件
		if err != nil {
			logger.WithError(err).Error("delete avatar file failed")
		}
		return "", syserror.InternalError
	}
	return url, syserror.NoError
}

// 下载用户头像
func (s *userService) UserDownloadAvatar(ctx context.Context, filename string) (*common.File, syserror.Error) {
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
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 获取用户记录
	user, err := s.userRepo.FindUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err != nil {
		logger.WithError(err).Error("find user by id failed")
		return syserror.InternalError
	}
	// 覆盖修改的字段
	err = copier.CopyWithOption(&user, &req, copier.Option{
		IgnoreEmpty: true, // 忽略 req 中的空值
	})
	if err != nil {
		logger.WithError(err).Error("copy profile fields failed")
		return syserror.InternalError
	}
	// 更新数据库的用户记录
	_, err = s.userRepo.UpdateUser(&user)
	if err != nil {
		logger.WithError(err).Error("update user failed")
		return syserror.InternalError
	}
	return syserror.NoError
}

// 更新用户密码
func (s *userService) UserUpdatePassword(ctx context.Context, userID int64, req request.UserPassword) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	// 获取用户记录
	user, err := s.userRepo.FindUserByID(userID)
	if err == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err != nil {
		logger.WithError(err).Error("find user by id failed")
		return syserror.InternalError
	}

	// 校验密码哈希是否相同
	if !utils.ValidatePassword(user.PasswordHash, req.OldPassword) {
		return syserror.PasswordError
	}

	// 生成新密码哈希
	newPwdHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		logger.WithError(err).Error("hash new password failed")
		return syserror.InternalError
	}

	// 将新密码哈希存入用户记录
	user.PasswordHash = newPwdHash
	_, err = s.userRepo.UpdateUser(&user)
	if err != nil {
		logger.WithError(err).Error("update user failed")
		return syserror.InternalError
	}
	return syserror.NoError
}

// 更新用户角色
func (s *userService) UpdateUserRole(ctx context.Context, opID int64, req request.UserRole) syserror.Error {
	logger := utils.WithContext(ctx).WithField("service", s.servName)
	operator, err1 := s.userRepo.FindUserByID(opID)
	target, err2 := s.userRepo.FindUserByID(req.ID)
	if err1 == gorm.ErrRecordNotFound || err2 == gorm.ErrRecordNotFound {
		return syserror.NotFoundError
	} else if err1 != nil || err2 != nil {
		logger.WithError(err1).WithField("target_error", err2).Error("find operator or target user failed")
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
		logger.WithError(err).Error("update user role failed")
		return syserror.InternalError
	}

	// 更新redis
	err = s.sessionRepo.UpdateUserRole(ctx, target.ID, target.Role)
	if err != nil {
		return syserror.InternalError
	}

	return syserror.NoError
}
