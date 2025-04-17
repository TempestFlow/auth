package dep

import (
	"context"
	"fmt"

	"users/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/jackc/pgx/v4/stdlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gtracing "gorm.io/plugin/opentelemetry/tracing"
)

type Gorm struct {
	// TODO wrapped database client
	DB     *gorm.DB
	Logger log.Logger
}

func openDB(ctx context.Context, c *conf.Data, tp trace.TracerProvider) (*gorm.DB, error) {
	_, span := tp.Tracer("gorm").Start(ctx, "openDB")
	defer span.End()
	span.SetAttributes(attribute.String("driver", c.Database.Driver))
	span.SetAttributes(attribute.String("source", c.Database.Source))
	retries := -1
	var err error = nil
	var db *gorm.DB = nil
	for {
		retries++
		if retries > 3 {
			if err != nil {
				break
			}
			return nil, err
		}
		span.SetAttributes(attribute.Int("retry", retries))
		db, err = gorm.Open(
			postgres.New(
				postgres.Config{
					DriverName:           c.Database.Driver,
					DSN:                  c.Database.Source,
					PreferSimpleProtocol: true,
				},
			),
			&gorm.Config{},
		)
		if err != nil {
			continue
		}
		if db == nil {
			continue
		}
		err := db.Use(gtracing.NewPlugin(gtracing.WithTracerProvider(tp)))
		if err != nil {
			continue
		}
		return db, nil
	}
	return nil, err
}

func NewGorm(c *conf.Data, logger log.Logger, tp trace.TracerProvider) (*Gorm, func(), error) {
	lg := log.NewHelper(logger)
	if c.GetDatabase() == nil || c.GetDatabase().GetSource() == "" {
		lg.Warn("No Database configuration found, skipping gorm initialization")
		return nil, nil, nil
	}
	db, err := openDB(context.Background(), c, tp)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		lg.Info("closing gorm data resources")
	}

	return &Gorm{
		DB:     db,
		Logger: logger,
	}, cleanup, nil
}

func GormMigrate(ctx context.Context, c *conf.Data, logger log.Logger, models ...interface{}) error {
	if c == nil || c.Database == nil || logger == nil || len(models) == 0 {
		return errors.InternalServer("invalid input parameters", "nil input parameters")
	}

	ctx, span := otel.Tracer("data").Start(ctx, "Migrate")
	defer span.End()
	l := log.NewHelper(logger)

	var db *gorm.DB
	var err error

	// Try to connect to database
	for i := 0; i < 3; i++ {
		span.SetAttributes(attribute.Int("retry", i))
		if db, err = openDB(ctx, c, otel.GetTracerProvider()); err == nil {
			break
		}
		l.Errorf("database connection attempt %d failed: %v", i+1, err)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying database: %w", err)
	}
	defer sqlDB.Close()

	// Migrate models
	var m []interface{}
	for _, model := range models {
		if model == nil {
			continue
		}
		m = append(m, model)
	}
	if err := db.Migrator().AutoMigrate(m...); err != nil {
		reason := fmt.Sprintf("failed to migrate %T", m)
		l.Error(reason, err)
		return errors.InternalServer(reason, err.Error())
	}
	l.Infof("migrated %T", m)

	l.Info("migration completed successfully")
	return nil
}
