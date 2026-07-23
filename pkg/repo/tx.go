package repo

import (
	"context"
	"gopkg.in/go-mixed/kratos-packages.v2/pkg/db"
)

type transactionKey struct{}

// newTxContext 构造事务context
func newTxContext(ctx context.Context, db *db.DB) context.Context {
	return context.WithValue(ctx, &transactionKey{}, db)
}

// Transaction 开启事务
//
//	注意：在Transaction之前设置的Clauses、Select、Omit、Preloads、PreloadWithBuilder，在steps中是不生效的
func (repo *Repository[T]) Transaction(ctx context.Context, steps ...func(ctx context.Context) error) error {
	var err error
	tx := repo.db.Begin()
	defer func() {
		if err != nil {
			repo.logger.WithContext(ctx).Error(err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ctx = newTxContext(ctx, tx)

	for _, step := range steps {
		if err = step(ctx); err != nil {
			return err
		}
	}

	return nil
}
