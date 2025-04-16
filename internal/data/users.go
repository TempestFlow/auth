package data

import (
	"context"
	"fmt"

	"users/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Username string    `gorm:"not null;uniqueIndex"`
	Email    string    `gorm:"not null;uniqueIndex"`
	Phone    string
	Picture  string
}

type usersRepo struct {
	db  *gorm.DB
	log *log.Helper
}

func NewUsersRepo(data *Data) biz.UsersRepo {
	return &usersRepo{
		db:  data.gorm,
		log: log.NewHelper(data.logger),
	}
}

func (r usersRepo) Save(ctx context.Context, u *biz.User) (string, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.Save")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "user",
		Value: attribute.StringValue(u.Username + " " + u.Email + " " + u.Phone),
	})
	user := Users{
		Username: u.Username,
		Email:    u.Email,
		Phone:    "123",
		Picture:  "",
		ID:       uuid.New(),
		// Phone:    u.Phone,
		// Picture:  u.Picture,
	}
	r.log.Debug(user)
	r.log.Debug("saving user")

	res := r.db.Create(&user)
	r.log.Debug("sent query")
	if res.Error != nil {
		r.log.Error("failed to save user", res.Error)
		return "", res.Error
	}
	r.log.Debug("saved user")
	if res.RowsAffected == 0 {
		r.log.Error("failed to save user", "err was empty but insertions failed")
		return "", errors.InternalServer("failed to save user", "err was empty but insertions failed")
	}
	return user.ID.String(), nil
}

func (r usersRepo) GetByID(ctx context.Context, id string) (*biz.User, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.GetByID")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "id",
		Value: attribute.StringValue(id),
	})
	uid, err := uuid.Parse(id)
	if err != nil {
		r.log.Error("failed to parse user id", err)
		return nil, err
	}
	user := &Users{
		ID: uid,
	}

	res := r.db.WithContext(ctx).First(&user)
	if res.Error != nil {
		r.log.Error("failed to get user", res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		r.log.Error("failed to get user", "err was empty but insertions failed")
		return nil, errors.InternalServer("failed to get user", "err was empty but insertions failed")
	}
	return &biz.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Picture:  user.Picture,
	}, nil
}

func (r usersRepo) List(ctx context.Context, pagination *biz.Pagination) ([]*biz.User, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.List")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "pagination",
		Value: attribute.StringValue(fmt.Sprintf("Page: %d Size: %d", pagination.Page, pagination.Size)),
	})
	offset := pagination.Page * pagination.Size
	take := pagination.Size
	if offset < 0 {
		offset = 0
	}
	if take < 0 {
		take = 0
	}

	var users []Users
	res := r.db.WithContext(ctx).Model(&Users{}).Offset(int(offset)).Limit(int(take)).Find(&users)

	if res.Error != nil {
		r.log.Error("failed to list users", res.Error)
		return nil, res.Error
	}
	if len(users) == 0 {
		err := errors.NotFound("users", "no users found")
		r.log.Error("failed to list users", err)
		return nil, err
	}

	var usersRes []*biz.User
	for _, user := range users {
		usersRes = append(usersRes, &biz.User{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Picture:  user.Picture,
		})
	}
	return usersRes, nil
}

func (r usersRepo) Update(ctx context.Context, u *biz.User) (*biz.User, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.Update")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "user",
		Value: attribute.StringValue(u.Username + " " + u.Email + " " + u.Phone),
	})
	uid, err := uuid.Parse(u.ID)
	if err != nil {
		r.log.Error("failed to parse user id", err)
		return nil, err
	}
	user := Users{
		ID:       uid,
		Username: u.Username,
		Email:    u.Email,
		Phone:    u.Phone,
		Picture:  u.Picture,
	}
	res := r.db.WithContext(ctx).Save(&user)
	if res.Error != nil {
		r.log.Error("failed to update user", res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		r.log.Error("failed to update user", "err was empty but insertions failed")
		return nil, errors.InternalServer("failed to update user", "err was empty but insertions failed")
	}
	return &biz.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Picture:  user.Picture,
	}, nil
}

func (r usersRepo) Delete(ctx context.Context, id string) (*biz.User, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.Delete")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "id",
		Value: attribute.StringValue(id),
	})
	uid, err := uuid.Parse(id)
	if err != nil {
		r.log.Error("failed to parse user id", err)
		return nil, err
	}
	user := Users{
		ID: uid,
	}
	res := r.db.WithContext(ctx).Delete(&user)
	if res.Error != nil {
		r.log.Error("failed to delete user", res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		r.log.Error("failed to delete user", "err was empty but insertions failed")
		return nil, errors.InternalServer("failed to delete user", "err was empty but insertions failed")
	}
	return &biz.User{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Phone:    user.Phone,
		Picture:  user.Picture,
	}, nil
}

func (r usersRepo) Search(ctx context.Context, keyword string, pagination *biz.Pagination) ([]*biz.User, error) {
	ctx, span := otel.Tracer("users").Start(ctx, "usersRepo.Search")
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "keyword",
		Value: attribute.StringValue(keyword),
	})
	span.SetAttributes(attribute.KeyValue{
		Key:   "pagination",
		Value: attribute.StringValue(fmt.Sprintf("Page: %d Size: %d", pagination.Page, pagination.Size)),
	})
	var users []Users
	res := r.db.WithContext(ctx).Where("username LIKE ?", "%"+keyword+"%").Find(&users)
	if res.Error != nil {
		r.log.Error("failed to search users", res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		r.log.Error("failed to search users", "err was empty but insertions failed")
		return nil, errors.InternalServer("failed to search users", "err was empty but insertions failed")
	}
	var usersRes []*biz.User
	for _, user := range users {
		usersRes = append(usersRes, &biz.User{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Picture:  user.Picture,
		})
	}
	return usersRes, nil
}
