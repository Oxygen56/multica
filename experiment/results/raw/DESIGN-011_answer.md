# DESIGN-011: 设计用户认证系统 — 答案（Group A 基线）

## 架构

```
客户端 → API Gateway → Auth Service (OAuth2/OIDC)
                          ├─ /authorize (授权)
                          ├─ /token (令牌)
                          ├─ /userinfo (用户信息)
                          ├─ /revoke (撤销)
                          └─ /.well-known/openid-configuration
                       → 其他微服务 (验证JWT)
```

## Token流转

```
1. 用户登录 → Auth Service验证 → 返回access_token(JWT) + refresh_token
2. 客户端携带access_token访问业务API
3. 业务服务通过公钥验证JWT签名 → 提取user_id + scopes
4. access_token过期(15min) → 客户端用refresh_token换新access_token
5. refresh_token过期(7天) → 用户重新登录
```

**JWT结构**：
```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "scopes": ["read:tasks", "write:tasks"],
  "iat": 1719600000, "exp": 1719600900,
  "jti": "unique-token-id"
}
```

## 多租户

- JWT中携带`tenant_id`
- 业务服务从JWT提取tenant_id，所有查询附加`WHERE tenant_id=$jwt_tenant_id`
- 或者：Auth Service为每个租户签发独立的JWT（用户切换租户时重新授权）

## SSO

- Auth Service作为OIDC Provider
- 其他内部应用通过OAuth2 Authorization Code Flow接入
- 已登录用户在Auth Service有session → 其他应用授权时无需再次输入密码

## 安全考量

| 攻击 | 防护 |
|------|------|
| CSRF | OAuth2的state参数验证 |
| XSS窃取Token | access_token存内存（非localStorage），refresh_token httpOnly cookie |
| JWT重放 | jti(唯一ID)黑名单（Redis，TTL=exp时间） |
| Token泄露 | access_token短有效期(15min)+refresh_token rotation |

## 数据库设计

```sql
CREATE TABLE users (id UUID PK, email UNIQUE, password_hash TEXT, mfa_secret TEXT);
CREATE TABLE refresh_tokens (id UUID PK, user_id FK, token_hash TEXT UNIQUE, expires_at, revoked BOOLEAN);
CREATE TABLE jti_blacklist (jti TEXT PK, expires_at TIMESTAMPTZ);
```

## 自评

- ✅ OAuth2.0/OIDC流程正确
- ✅ JWT生命周期管理完整（签发/刷新/撤销/黑名单）
- ✅ 多租户隔离方案清晰（JWT携带+SQL过滤）
- ✅ 安全考量覆盖主要攻击向量

**完成** | 修复轮次: 0
