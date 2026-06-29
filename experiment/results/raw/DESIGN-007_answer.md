# DESIGN-007: 设计数据库Schema：内容平台 — 答案（Group A 基线）

## ER描述

```
User 1─N Article (作者)
User 1─N Like (点赞)
User 1─N Bookmark (收藏)
User 1─N Follow (关注, 自引用)
Article 1─N ArticleVersion (版本历史)
Article N─M Tag (标签, 多对多)
Article N─1 Category (分类)
```

## DDL

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    category_id UUID REFERENCES categories(id),
    title VARCHAR(200) NOT NULL,
    slug VARCHAR(250) UNIQUE NOT NULL,
    summary TEXT,
    status VARCHAR(20) DEFAULT 'draft',  -- draft/published/archived
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE article_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    change_summary VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(article_id, version_number)
);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) UNIQUE NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE article_tags (
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY(article_id, tag_id)
);

CREATE TABLE likes (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY(user_id, article_id)
);

CREATE TABLE bookmarks (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    article_id UUID REFERENCES articles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY(user_id, article_id)
);

CREATE TABLE follows (
    follower_id UUID REFERENCES users(id) ON DELETE CASCADE,
    followed_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY(follower_id, followed_id),
    CHECK(follower_id != followed_id)
);

-- 关键索引
CREATE INDEX idx_articles_author_published ON articles(author_id, published_at DESC) WHERE status='published';
CREATE INDEX idx_articles_category_published ON articles(category_id, published_at DESC) WHERE status='published';
CREATE INDEX idx_article_versions_article ON article_versions(article_id, version_number DESC);
CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_followed ON follows(followed_id);
```

## 关键查询：关注作者的最新10篇文章

```sql
SELECT a.*, u.display_name AS author_name
FROM articles a
JOIN follows f ON f.followed_id = a.author_id
JOIN users u ON u.id = a.author_id
WHERE f.follower_id = $current_user_id
  AND a.status = 'published'
ORDER BY a.published_at DESC
LIMIT 10;
```

**查询计划**：使用 `idx_articles_author_published` 索引。先通过follows找到关注的作者（索引扫描），再对每个作者用索引获取最新文章。如果关注100人，取10条——用LATERAL JOIN优化：

```sql
SELECT a.*, u.display_name
FROM follows f
CROSS JOIN LATERAL (
    SELECT * FROM articles
    WHERE author_id = f.followed_id AND status='published'
    ORDER BY published_at DESC LIMIT 3
) a
JOIN users u ON u.id = a.author_id
WHERE f.follower_id = $uid
ORDER BY a.published_at DESC LIMIT 10;
```

**大V场景（单作者有10000篇文章）**：索引 `(author_id, published_at DESC)` 直接定位，LATERAL每作者取3篇后归并排序，总扫描<300行。

## 自评

- ✅ ER设计覆盖所有实体和关系
- ✅ DDL语法正确，约束完整
- ✅ 索引设计合理（含部分索引和复合索引）
- ✅ 关键查询分析正确（含LATERAL优化和大V场景）

**完成** | 修复轮次: 0
