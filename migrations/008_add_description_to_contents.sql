-- 添加 description 字段到 contents 表
ALTER TABLE contents ADD COLUMN description VARCHAR(500) DEFAULT '';

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_contents_description ON contents(description);

-- 更新现有记录的描述字段（从标题截取前100个字符作为描述）
UPDATE contents 
SET description = CASE 
    WHEN LENGTH(title) > 100 THEN LEFT(title, 100) || '...'
    ELSE COALESCE(title, '')
END
WHERE description IS NULL OR description = '';
