-- Test Data for Matching Engine - Dating Module (PostgreSQL)
-- 設計書: claude/06_DATABASE_SCHEMA.md に準拠

-- テストユーザー (dating_users)
INSERT INTO dating_users (id, nickname, gender, birth_date, prefecture, verified, elo_rating, last_active_at, created_at) VALUES
('user1', 'TestUser1', 'male', '1995-06-15', 'Tokyo', TRUE, 1200, now(), now()),
('user2', 'TestUser2', 'female', '1998-03-22', 'Tokyo', TRUE, 1300, now(), now()),
('user3', 'TestUser3', 'male', '1992-11-08', 'Osaka', TRUE, 1150, now(), now()),
('user4', 'TestUser4', 'female', '1996-07-30', 'Tokyo', FALSE, 1000, now(), now()),
('user5', 'TestUser5', 'male', '1994-02-14', 'Kyoto', TRUE, 1400, now(), now())
ON CONFLICT (id) DO UPDATE SET
    nickname = EXCLUDED.nickname,
    gender = EXCLUDED.gender,
    birth_date = EXCLUDED.birth_date,
    prefecture = EXCLUDED.prefecture,
    verified = EXCLUDED.verified,
    elo_rating = EXCLUDED.elo_rating;

-- テストプロフィール (dating_profiles)
INSERT INTO dating_profiles (user_id, height, body_type, income_level, education, marriage_desire, children_desire, smoking, drinking) VALUES
('user1', 175, 'average', 600, 'university', 'want_soon', 'want', 'non_smoker', 'social'),
('user2', 165, 'slim', 400, 'university', 'want_soon', 'want', 'non_smoker', 'non_drinker'),
('user3', 180, 'athletic', 800, 'graduate', 'undecided', 'undecided', 'non_smoker', 'regular'),
('user4', 160, 'average', 300, 'high_school', 'want_soon', 'want', 'non_smoker', 'non_drinker'),
('user5', 172, 'slim', 1000, 'graduate', 'want_eventually', 'not_want', 'non_smoker', 'social')
ON CONFLICT (user_id) DO UPDATE SET
    height = EXCLUDED.height,
    body_type = EXCLUDED.body_type,
    income_level = EXCLUDED.income_level,
    education = EXCLUDED.education,
    marriage_desire = EXCLUDED.marriage_desire,
    children_desire = EXCLUDED.children_desire,
    smoking = EXCLUDED.smoking,
    drinking = EXCLUDED.drinking;

-- プロフィールタグ (dating_profile_tags)
INSERT INTO dating_profile_tags (user_id, tag) VALUES
('user1', 'sports'),
('user1', 'travel'),
('user2', 'reading'),
('user2', 'music'),
('user3', 'sports'),
('user3', 'cooking'),
('user4', 'movies'),
('user4', 'art'),
('user5', 'technology'),
('user5', 'travel')
ON CONFLICT (user_id, tag) DO NOTHING;

-- テスト設定 (dating_preferences)
INSERT INTO dating_preferences (user_id, age_min, age_max, height_min, height_max, income_min) VALUES
('user1', 25, 35, 160, 175, 300),
('user2', 25, 35, 170, 185, 500),
('user3', 23, 33, 155, 170, 300)
ON CONFLICT (user_id) DO UPDATE SET
    age_min = EXCLUDED.age_min,
    age_max = EXCLUDED.age_max,
    height_min = EXCLUDED.height_min,
    height_max = EXCLUDED.height_max,
    income_min = EXCLUDED.income_min;

-- 希望都道府県 (dating_preference_prefectures)
INSERT INTO dating_preference_prefectures (user_id, prefecture) VALUES
('user1', 'Tokyo'),
('user1', 'Kanagawa'),
('user2', 'Tokyo'),
('user2', 'Saitama'),
('user3', 'Osaka'),
('user3', 'Kyoto')
ON CONFLICT (user_id, prefecture) DO NOTHING;
