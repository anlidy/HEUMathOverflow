package main

import "gorm.io/gorm"

func InitTriggers(db *gorm.DB) {
	// 插入回帖时更新原帖的回复时间
	sql := `
        CREATE OR REPLACE FUNCTION update_post_after_insert_reply_func()
        RETURNS TRIGGER AS $$
        BEGIN
            UPDATE posts
            SET last_reply_at = NEW.created_at
            WHERE id = NEW.post_id;
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;

        DROP TRIGGER IF EXISTS update_post_after_insert_reply ON replies;

        CREATE TRIGGER update_post_after_insert_reply
        AFTER INSERT ON replies
        FOR EACH ROW
        EXECUTE FUNCTION update_post_after_insert_reply_func();
        `

	if err := db.Exec(sql).Error; err != nil {
		panic(err)
	}

	// 保证 post_stars 幂等（去重 + 建唯一约束），避免重复收藏导致统计漂移。
	// 说明：Postgres 的 UNIQUE 对 NULL 不敏感（允许多条 NULL），因此加 partial unique index 仅约束有效 post_id。
	sql = `
		-- 删除重复收藏（保留 created_at 最早的一条）
		WITH ranked AS (
			SELECT
				id,
				ROW_NUMBER() OVER (PARTITION BY post_id, user_id ORDER BY created_at ASC) AS rn
			FROM post_stars
			WHERE post_id IS NOT NULL
		)
		DELETE FROM post_stars
		WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

		-- 建立唯一索引：同一用户同一帖子最多一条收藏记录
		CREATE UNIQUE INDEX IF NOT EXISTS idx_post_stars_post_user
		ON post_stars (post_id, user_id)
		WHERE post_id IS NOT NULL;
	`
	if err := db.Exec(sql).Error; err != nil {
		panic(err)
	}
}
