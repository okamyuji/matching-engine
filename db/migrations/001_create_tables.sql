-- Matching Engine Database Schema (PostgreSQL 18)
-- 設計書: claude/06_DATABASE_SCHEMA.md に準拠
-- 列挙値は PostgreSQL の enum 型ではなく text + CHECK 制約で表現する（ALTER TYPE を避けるため）
-- updated_at の自動更新は持たない。更新クエリ側で明示的に now() を設定する

-- 2. Dating スキーマ

-- 2.1 dating_users
CREATE TABLE IF NOT EXISTS dating_users (
    id              TEXT PRIMARY KEY,
    nickname        TEXT NOT NULL,
    gender          TEXT NOT NULL CHECK (gender IN ('male', 'female', 'other')),
    birth_date      DATE NOT NULL,
    prefecture      TEXT NOT NULL,
    verified        BOOLEAN NOT NULL DEFAULT FALSE,
    elo_rating      INTEGER NOT NULL DEFAULT 1000,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_active_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dating_users_gender ON dating_users (gender);
CREATE INDEX IF NOT EXISTS idx_dating_users_prefecture ON dating_users (prefecture);
CREATE INDEX IF NOT EXISTS idx_dating_users_birth_date ON dating_users (birth_date);
CREATE INDEX IF NOT EXISTS idx_dating_users_verified ON dating_users (verified);
CREATE INDEX IF NOT EXISTS idx_dating_users_elo_rating ON dating_users (elo_rating);
CREATE INDEX IF NOT EXISTS idx_dating_users_last_active ON dating_users (last_active_at);

-- 2.2 dating_profiles
CREATE TABLE IF NOT EXISTS dating_profiles (
    user_id           TEXT PRIMARY KEY REFERENCES dating_users(id) ON DELETE CASCADE,
    height            INTEGER,
    body_type         TEXT CHECK (body_type IN ('slim', 'average', 'athletic', 'large')),
    education         TEXT CHECK (education IN ('high_school', 'vocational', 'university', 'graduate')),
    occupation        TEXT,
    income_level      INTEGER,
    marriage_desire   TEXT CHECK (marriage_desire IN ('want_soon', 'want_eventually', 'undecided', 'not_want')),
    children_desire   TEXT CHECK (children_desire IN ('want', 'not_want', 'undecided')),
    smoking           TEXT CHECK (smoking IN ('non_smoker', 'occasional', 'smoker')),
    drinking          TEXT CHECK (drinking IN ('non_drinker', 'social', 'regular')),
    self_introduction TEXT,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dating_profiles_height ON dating_profiles (height);
CREATE INDEX IF NOT EXISTS idx_dating_profiles_income_level ON dating_profiles (income_level);

-- 2.3 dating_profile_tags
CREATE TABLE IF NOT EXISTS dating_profile_tags (
    id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id  TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    tag      TEXT NOT NULL,
    UNIQUE (user_id, tag)
);
CREATE INDEX IF NOT EXISTS idx_dating_profile_tags_tag ON dating_profile_tags (tag);

-- 2.4 dating_profile_photos
CREATE TABLE IF NOT EXISTS dating_profile_photos (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id       TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    url           TEXT NOT NULL,
    is_primary    BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_dating_profile_photos_user ON dating_profile_photos (user_id);
CREATE INDEX IF NOT EXISTS idx_dating_profile_photos_primary ON dating_profile_photos (user_id, is_primary);

-- 2.5 dating_preferences
CREATE TABLE IF NOT EXISTS dating_preferences (
    user_id    TEXT PRIMARY KEY REFERENCES dating_users(id) ON DELETE CASCADE,
    age_min    INTEGER,
    age_max    INTEGER,
    height_min INTEGER,
    height_max INTEGER,
    income_min INTEGER,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2.6 dating_preference_prefectures
CREATE TABLE IF NOT EXISTS dating_preference_prefectures (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    prefecture TEXT NOT NULL,
    UNIQUE (user_id, prefecture)
);

-- 2.7 dating_preference_educations
CREATE TABLE IF NOT EXISTS dating_preference_educations (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id   TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    education TEXT NOT NULL CHECK (education IN ('high_school', 'vocational', 'university', 'graduate')),
    UNIQUE (user_id, education)
);

-- 2.8 dating_preference_marriage_desires
CREATE TABLE IF NOT EXISTS dating_preference_marriage_desires (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    marriage_desire TEXT NOT NULL CHECK (marriage_desire IN ('want_soon', 'want_eventually', 'undecided', 'not_want')),
    UNIQUE (user_id, marriage_desire)
);

-- 2.9 dating_preference_smoking_statuses
CREATE TABLE IF NOT EXISTS dating_preference_smoking_statuses (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    smoking_status TEXT NOT NULL CHECK (smoking_status IN ('non_smoker', 'occasional', 'smoker')),
    UNIQUE (user_id, smoking_status)
);

-- 2.10 dating_preference_drinking_statuses
CREATE TABLE IF NOT EXISTS dating_preference_drinking_statuses (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    drinking_status TEXT NOT NULL CHECK (drinking_status IN ('non_drinker', 'social', 'regular')),
    UNIQUE (user_id, drinking_status)
);

-- 2.11 dating_likes
CREATE TABLE IF NOT EXISTS dating_likes (
    id           TEXT PRIMARY KEY,
    from_user_id TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    to_user_id   TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_user_id, to_user_id)
);
CREATE INDEX IF NOT EXISTS idx_dating_likes_to_user ON dating_likes (to_user_id);

-- 2.12 dating_matches
CREATE TABLE IF NOT EXISTS dating_matches (
    id         TEXT PRIMARY KEY,
    user_id_a  TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    user_id_b  TEXT NOT NULL REFERENCES dating_users(id) ON DELETE CASCADE,
    score      DOUBLE PRECISION NOT NULL,
    breakdown  JSONB,
    matched_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id_a, user_id_b)
);
CREATE INDEX IF NOT EXISTS idx_dating_matches_user_a ON dating_matches (user_id_a);
CREATE INDEX IF NOT EXISTS idx_dating_matches_user_b ON dating_matches (user_id_b);

-- 複合インデックス（候補検索最適化）
CREATE INDEX IF NOT EXISTS idx_dating_candidates ON dating_users (gender, verified, prefecture, birth_date);

-- 3. M&A スキーマ

-- 3.1 ma_companies
CREATE TABLE IF NOT EXISTS ma_companies (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    industry         TEXT NOT NULL CHECK (industry IN (
        'technology', 'finance', 'healthcare', 'manufacturing',
        'retail', 'real_estate', 'energy', 'education',
        'entertainment', 'logistics')),
    location         TEXT NOT NULL,
    employee_count   INTEGER NOT NULL,
    founded          DATE NOT NULL,
    listing_status   TEXT NOT NULL DEFAULT 'private' CHECK (listing_status IN ('public', 'private')),
    matching_purpose TEXT NOT NULL CHECK (matching_purpose IN ('buyer', 'seller')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ma_companies_industry ON ma_companies (industry);
CREATE INDEX IF NOT EXISTS idx_ma_companies_location ON ma_companies (location);
CREATE INDEX IF NOT EXISTS idx_ma_companies_purpose ON ma_companies (matching_purpose);
CREATE INDEX IF NOT EXISTS idx_ma_companies_employee_count ON ma_companies (employee_count);

-- 3.2 ma_financials
CREATE TABLE IF NOT EXISTS ma_financials (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id        TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    fiscal_year       INTEGER NOT NULL,
    revenue           BIGINT NOT NULL,
    ebitda            BIGINT NOT NULL,
    net_income        BIGINT NOT NULL,
    total_assets      BIGINT NOT NULL,
    total_liabilities BIGINT NOT NULL,
    equity            BIGINT NOT NULL,
    roe               DOUBLE PRECISION,
    roa               DOUBLE PRECISION,
    debt_equity_ratio DOUBLE PRECISION,
    current_ratio     DOUBLE PRECISION,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (company_id, fiscal_year)
);
CREATE INDEX IF NOT EXISTS idx_ma_financials_revenue ON ma_financials (revenue);
CREATE INDEX IF NOT EXISTS idx_ma_financials_ebitda ON ma_financials (ebitda);

-- 3.3 ma_matching_criteria
CREATE TABLE IF NOT EXISTS ma_matching_criteria (
    company_id            TEXT PRIMARY KEY REFERENCES ma_companies(id) ON DELETE CASCADE,
    purpose               TEXT NOT NULL CHECK (purpose IN ('buyer', 'seller')),
    revenue_min           BIGINT,
    revenue_max           BIGINT,
    ebitda_min            BIGINT,
    employee_min          INTEGER,
    employee_max          INTEGER,
    max_debt_equity_ratio DOUBLE PRECISION,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 3.4 ma_criteria_industries
CREATE TABLE IF NOT EXISTS ma_criteria_industries (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    industry   TEXT NOT NULL,
    UNIQUE (company_id, industry)
);

-- 3.5 ma_company_technologies
CREATE TABLE IF NOT EXISTS ma_company_technologies (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    technology TEXT NOT NULL,
    UNIQUE (company_id, technology)
);
CREATE INDEX IF NOT EXISTS idx_ma_company_technologies_tech ON ma_company_technologies (technology);

-- 3.6 ma_company_markets
CREATE TABLE IF NOT EXISTS ma_company_markets (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    company_id TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    market     TEXT NOT NULL,
    UNIQUE (company_id, market)
);
CREATE INDEX IF NOT EXISTS idx_ma_company_markets_market ON ma_company_markets (market);

-- 3.7 ma_interests
CREATE TABLE IF NOT EXISTS ma_interests (
    id              TEXT PRIMARY KEY,
    from_company_id TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    to_company_id   TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_company_id, to_company_id)
);
CREATE INDEX IF NOT EXISTS idx_ma_interests_to_company ON ma_interests (to_company_id);

-- 3.8 ma_matches
CREATE TABLE IF NOT EXISTS ma_matches (
    id              TEXT PRIMARY KEY,
    company_id_a    TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    company_id_b    TEXT NOT NULL REFERENCES ma_companies(id) ON DELETE CASCADE,
    score           DOUBLE PRECISION NOT NULL,
    breakdown       JSONB,
    synergy_summary JSONB,
    matched_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (company_id_a, company_id_b)
);
CREATE INDEX IF NOT EXISTS idx_ma_matches_company_a ON ma_matches (company_id_a);
CREATE INDEX IF NOT EXISTS idx_ma_matches_company_b ON ma_matches (company_id_b);

-- 複合インデックス（M&A候補検索最適化）
CREATE INDEX IF NOT EXISTS idx_ma_candidates ON ma_companies (matching_purpose, industry, location);
