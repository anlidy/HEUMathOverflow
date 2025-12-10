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

}
