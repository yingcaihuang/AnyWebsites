-- 添加 deleted_at 字段到 contents 表
ALTER TABLE contents ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_contents_deleted_at ON contents(deleted_at);
CREATE INDEX IF NOT EXISTS idx_contents_is_active_deleted_at ON contents(is_active, deleted_at);

-- 添加注释
COMMENT ON COLUMN contents.deleted_at IS '软删除时间，NULL表示未删除';
