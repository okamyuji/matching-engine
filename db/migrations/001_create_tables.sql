-- Matching Engine Database Schema - Dating Module
-- 設計書: claude/06_DATABASE_SCHEMA.md に準拠

-- 2.1 dating_users テーブル
CREATE TABLE IF NOT EXISTS dating_users (
    id VARCHAR(36) PRIMARY KEY,
    nickname VARCHAR(50) NOT NULL,
    gender ENUM('male', 'female', 'other') NOT NULL,
    birth_date DATE NOT NULL,
    prefecture VARCHAR(10) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    elo_rating INT NOT NULL DEFAULT 1000,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_gender (gender),
    INDEX idx_prefecture (prefecture),
    INDEX idx_birth_date (birth_date),
    INDEX idx_verified (verified),
    INDEX idx_elo_rating (elo_rating),
    INDEX idx_last_active (last_active_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.2 dating_profiles テーブル
CREATE TABLE IF NOT EXISTS dating_profiles (
    user_id VARCHAR(36) PRIMARY KEY,
    height INT,
    body_type ENUM('slim', 'average', 'athletic', 'large'),
    education ENUM('high_school', 'vocational', 'university', 'graduate'),
    occupation VARCHAR(100),
    income_level INT,
    marriage_desire ENUM('want_soon', 'want_eventually', 'undecided', 'not_want'),
    children_desire ENUM('want', 'not_want', 'undecided'),
    smoking ENUM('non_smoker', 'occasional', 'smoker'),
    drinking ENUM('non_drinker', 'social', 'regular'),
    self_introduction TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    INDEX idx_height (height),
    INDEX idx_income_level (income_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.3 dating_profile_tags テーブル
CREATE TABLE IF NOT EXISTS dating_profile_tags (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    tag VARCHAR(50) NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_tag (user_id, tag),
    INDEX idx_tag (tag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.4 dating_profile_photos テーブル
CREATE TABLE IF NOT EXISTS dating_profile_photos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    url VARCHAR(512) NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_is_primary (user_id, is_primary)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.5 dating_preferences テーブル
CREATE TABLE IF NOT EXISTS dating_preferences (
    user_id VARCHAR(36) PRIMARY KEY,
    age_min INT,
    age_max INT,
    height_min INT,
    height_max INT,
    income_min INT,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.6 dating_preference_prefectures テーブル
CREATE TABLE IF NOT EXISTS dating_preference_prefectures (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    prefecture VARCHAR(10) NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_pref (user_id, prefecture)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.7 dating_preference_educations テーブル
CREATE TABLE IF NOT EXISTS dating_preference_educations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    education ENUM('high_school', 'vocational', 'university', 'graduate') NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_education (user_id, education)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.8 dating_preference_marriage_desires テーブル
CREATE TABLE IF NOT EXISTS dating_preference_marriage_desires (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    marriage_desire ENUM('want_soon', 'want_eventually', 'undecided', 'not_want') NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_marriage (user_id, marriage_desire)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.9 dating_preference_smoking_statuses テーブル
CREATE TABLE IF NOT EXISTS dating_preference_smoking_statuses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    smoking_status ENUM('non_smoker', 'occasional', 'smoker') NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_smoking (user_id, smoking_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.10 dating_preference_drinking_statuses テーブル
CREATE TABLE IF NOT EXISTS dating_preference_drinking_statuses (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    drinking_status ENUM('non_drinker', 'social', 'regular') NOT NULL,

    FOREIGN KEY (user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_user_drinking (user_id, drinking_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.11 dating_likes テーブル
CREATE TABLE IF NOT EXISTS dating_likes (
    id VARCHAR(36) PRIMARY KEY,
    from_user_id VARCHAR(36) NOT NULL,
    to_user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (from_user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_from_to (from_user_id, to_user_id),
    INDEX idx_to_user (to_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2.12 dating_matches テーブル
CREATE TABLE IF NOT EXISTS dating_matches (
    id VARCHAR(36) PRIMARY KEY,
    user_id_a VARCHAR(36) NOT NULL,
    user_id_b VARCHAR(36) NOT NULL,
    score DECIMAL(5,4) NOT NULL,
    breakdown JSON,
    matched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id_a) REFERENCES dating_users(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id_b) REFERENCES dating_users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_users (user_id_a, user_id_b),
    INDEX idx_user_a (user_id_a),
    INDEX idx_user_b (user_id_b)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 複合インデックス（候補検索最適化）
CREATE INDEX idx_dating_candidates
ON dating_users(gender, verified, prefecture, birth_date);

-- 3. M&Aスキーマ
-- 設計書: claude/06_DATABASE_SCHEMA.md セクション3 に準拠

-- 3.1 ma_companies テーブル
CREATE TABLE IF NOT EXISTS ma_companies (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    industry ENUM(
        'technology', 'finance', 'healthcare', 'manufacturing',
        'retail', 'real_estate', 'energy', 'education',
        'entertainment', 'logistics'
    ) NOT NULL,
    location VARCHAR(10) NOT NULL,
    employee_count INT NOT NULL,
    founded DATE NOT NULL,
    listing_status ENUM('public', 'private') NOT NULL DEFAULT 'private',
    matching_purpose ENUM('buyer', 'seller') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_industry (industry),
    INDEX idx_location (location),
    INDEX idx_purpose (matching_purpose),
    INDEX idx_employee_count (employee_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.2 ma_financials テーブル
CREATE TABLE IF NOT EXISTS ma_financials (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL,
    fiscal_year INT NOT NULL,
    revenue BIGINT NOT NULL,
    ebitda BIGINT NOT NULL,
    net_income BIGINT NOT NULL,
    total_assets BIGINT NOT NULL,
    total_liabilities BIGINT NOT NULL,
    equity BIGINT NOT NULL,
    roe DECIMAL(10,4),
    roa DECIMAL(10,4),
    debt_equity_ratio DECIMAL(10,4),
    current_ratio DECIMAL(10,4),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_company_year (company_id, fiscal_year),
    INDEX idx_revenue (revenue),
    INDEX idx_ebitda (ebitda)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.3 ma_matching_criteria テーブル
CREATE TABLE IF NOT EXISTS ma_matching_criteria (
    company_id VARCHAR(36) PRIMARY KEY,
    purpose ENUM('buyer', 'seller') NOT NULL,
    revenue_min BIGINT,
    revenue_max BIGINT,
    ebitda_min BIGINT,
    employee_min INT,
    employee_max INT,
    max_debt_equity_ratio DECIMAL(10,4),
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (company_id) REFERENCES ma_companies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.4 ma_criteria_industries テーブル
CREATE TABLE IF NOT EXISTS ma_criteria_industries (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL,
    industry VARCHAR(50) NOT NULL,

    FOREIGN KEY (company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_company_industry (company_id, industry)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.5 ma_company_technologies テーブル
CREATE TABLE IF NOT EXISTS ma_company_technologies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL,
    technology VARCHAR(100) NOT NULL,

    FOREIGN KEY (company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_company_tech (company_id, technology),
    INDEX idx_technology (technology)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.6 ma_company_markets テーブル
CREATE TABLE IF NOT EXISTS ma_company_markets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    company_id VARCHAR(36) NOT NULL,
    market VARCHAR(100) NOT NULL,

    FOREIGN KEY (company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_company_market (company_id, market),
    INDEX idx_market (market)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.7 ma_interests テーブル
CREATE TABLE IF NOT EXISTS ma_interests (
    id VARCHAR(36) PRIMARY KEY,
    from_company_id VARCHAR(36) NOT NULL,
    to_company_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (from_company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    FOREIGN KEY (to_company_id) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_from_to (from_company_id, to_company_id),
    INDEX idx_to_company (to_company_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3.8 ma_matches テーブル
CREATE TABLE IF NOT EXISTS ma_matches (
    id VARCHAR(36) PRIMARY KEY,
    company_id_a VARCHAR(36) NOT NULL,
    company_id_b VARCHAR(36) NOT NULL,
    score DECIMAL(5,4) NOT NULL,
    breakdown JSON,
    synergy_summary JSON,
    matched_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (company_id_a) REFERENCES ma_companies(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id_b) REFERENCES ma_companies(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_companies (company_id_a, company_id_b),
    INDEX idx_company_a (company_id_a),
    INDEX idx_company_b (company_id_b)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 複合インデックス（M&A候補検索最適化）
CREATE INDEX idx_ma_candidates
ON ma_companies(matching_purpose, industry, location);
