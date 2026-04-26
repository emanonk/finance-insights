package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/manoskammas/finance-insights/apps/api/internal/domain"
)

// MerchantRepository persists and queries merchant records and their tags.
type MerchantRepository struct {
	pool *pgxpool.Pool
}

// NewMerchantRepository returns a MerchantRepository bound to the given pool.
func NewMerchantRepository(pool *pgxpool.Pool) *MerchantRepository {
	return &MerchantRepository{pool: pool}
}

// TopIdentifiers returns the most frequent merchant_identifier values from
// the transactions table, joined with any existing merchant record.
func (r *MerchantRepository) TopIdentifiers(ctx context.Context, limit int) ([]domain.IdentifierCount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			t.merchant_identifier,
			COUNT(*) AS cnt,
			m.id,
			m.default_title,
			pt.id   AS pt_id,
			pt.name AS pt_name,
			pt.type AS pt_type
		FROM transactions t
		LEFT JOIN merchants m  ON m.identifier_name = t.merchant_identifier
		LEFT JOIN tags      pt ON pt.id = m.primary_tag_id
		WHERE t.merchant_identifier IS NOT NULL
		GROUP BY t.merchant_identifier, m.id, m.default_title, pt.id, pt.name, pt.type
		ORDER BY cnt DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("top identifiers query: %w", err)
	}
	defer rows.Close()

	var out []domain.IdentifierCount
	for rows.Next() {
		var (
			identifier string
			count      int
			mID        *int64
			mTitle     *string
			ptID       *int64
			ptName     *string
			ptType     *string
		)
		if err := rows.Scan(&identifier, &count, &mID, &mTitle, &ptID, &ptName, &ptType); err != nil {
			return nil, fmt.Errorf("scan top identifier: %w", err)
		}
		ic := domain.IdentifierCount{Identifier: identifier, Count: count}
		if mID != nil {
			m := &domain.Merchant{
				ID:             *mID,
				IdentifierName: identifier,
				DefaultTitle:   mTitle,
			}
			if ptID != nil {
				m.PrimaryTag = domain.Tag{ID: *ptID, Name: *ptName, Type: *ptType}
			}
			ic.Merchant = m
		}
		out = append(out, ic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top identifiers: %w", err)
	}

	for i, ic := range out {
		if ic.Merchant == nil {
			continue
		}
		tags, err := r.loadSecondaryTags(ctx, ic.Merchant.ID)
		if err != nil {
			return nil, err
		}
		out[i].Merchant.SecondaryTags = tags
	}

	return out, nil
}

func (r *MerchantRepository) loadSecondaryTags(ctx context.Context, merchantID int64) ([]domain.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.name, t.type
		FROM merchant_secondary_tags mst
		JOIN tags t ON t.id = mst.tag_id
		WHERE mst.merchant_id = $1
		ORDER BY t.name
	`, merchantID)
	if err != nil {
		return nil, fmt.Errorf("load secondary tags: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Type); err != nil {
			return nil, fmt.Errorf("scan secondary tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// Upsert creates or updates a merchant record, setting its primary tag and
// replacing all secondary tags. Tags are created if they don't exist yet.
func (r *MerchantRepository) Upsert(
	ctx context.Context,
	identifierName string,
	primaryTagName string,
	secondaryTagNames []string,
) (*domain.Merchant, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var ptID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO tags (name, type) VALUES ($1, 'primary')
		ON CONFLICT (name, type) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, primaryTagName).Scan(&ptID); err != nil {
		return nil, fmt.Errorf("upsert primary tag: %w", err)
	}

	var mID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO merchants (identifier_name, primary_tag_id)
		VALUES ($1, $2)
		ON CONFLICT (identifier_name) DO UPDATE SET primary_tag_id = EXCLUDED.primary_tag_id
		RETURNING id
	`, identifierName, ptID).Scan(&mID); err != nil {
		return nil, fmt.Errorf("upsert merchant: %w", err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM merchant_secondary_tags WHERE merchant_id = $1`, mID); err != nil {
		return nil, fmt.Errorf("clear secondary tags: %w", err)
	}

	secTags := make([]domain.Tag, 0, len(secondaryTagNames))
	for _, name := range secondaryTagNames {
		if name == "" {
			continue
		}
		var stID int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO tags (name, type) VALUES ($1, 'secondary')
			ON CONFLICT (name, type) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, name).Scan(&stID); err != nil {
			return nil, fmt.Errorf("upsert secondary tag %q: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO merchant_secondary_tags (merchant_id, tag_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, mID, stID); err != nil {
			return nil, fmt.Errorf("link secondary tag: %w", err)
		}
		secTags = append(secTags, domain.Tag{ID: stID, Name: name, Type: "secondary"})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &domain.Merchant{
		ID:             mID,
		IdentifierName: identifierName,
		PrimaryTag:     domain.Tag{ID: ptID, Name: primaryTagName, Type: "primary"},
		SecondaryTags:  secTags,
	}, nil
}
