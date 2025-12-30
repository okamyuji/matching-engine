-- Test Data for Matching Engine - Dating Module
-- 設計書: claude/06_DATABASE_SCHEMA.md に準拠

-- テストユーザー (dating_users)
INSERT INTO dating_users (id, nickname, gender, birth_date, prefecture, verified, elo_rating, last_active_at, created_at) VALUES
('user1', 'TestUser1', 'male', '1995-06-15', 'Tokyo', TRUE, 1200, NOW(), NOW()),
('user2', 'TestUser2', 'female', '1998-03-22', 'Tokyo', TRUE, 1300, NOW(), NOW()),
('user3', 'TestUser3', 'male', '1992-11-08', 'Osaka', TRUE, 1150, NOW(), NOW()),
('user4', 'TestUser4', 'female', '1996-07-30', 'Tokyo', FALSE, 1000, NOW(), NOW()),
('user5', 'TestUser5', 'male', '1994-02-14', 'Kyoto', TRUE, 1400, NOW(), NOW())
ON DUPLICATE KEY UPDATE
    nickname = VALUES(nickname),
    gender = VALUES(gender),
    birth_date = VALUES(birth_date),
    prefecture = VALUES(prefecture),
    verified = VALUES(verified),
    elo_rating = VALUES(elo_rating);

-- テストプロフィール (dating_profiles)
INSERT INTO dating_profiles (user_id, height, body_type, income_level, education, marriage_desire, children_desire, smoking, drinking) VALUES
('user1', 175, 'average', 600, 'university', 'want_soon', 'want', 'non_smoker', 'social'),
('user2', 165, 'slim', 400, 'university', 'want_soon', 'want', 'non_smoker', 'non_drinker'),
('user3', 180, 'athletic', 800, 'graduate', 'undecided', 'undecided', 'non_smoker', 'regular'),
('user4', 160, 'average', 300, 'high_school', 'want_soon', 'want', 'non_smoker', 'non_drinker'),
('user5', 172, 'slim', 1000, 'graduate', 'want_eventually', 'not_want', 'non_smoker', 'social')
ON DUPLICATE KEY UPDATE
    height = VALUES(height),
    body_type = VALUES(body_type),
    income_level = VALUES(income_level),
    education = VALUES(education),
    marriage_desire = VALUES(marriage_desire),
    children_desire = VALUES(children_desire),
    smoking = VALUES(smoking),
    drinking = VALUES(drinking);

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
ON DUPLICATE KEY UPDATE
    tag = VALUES(tag);

-- テスト設定 (dating_preferences)
INSERT INTO dating_preferences (user_id, age_min, age_max, height_min, height_max, income_min) VALUES
('user1', 25, 35, 160, 175, 300),
('user2', 25, 35, 170, 185, 500),
('user3', 23, 33, 155, 170, 300)
ON DUPLICATE KEY UPDATE
    age_min = VALUES(age_min),
    age_max = VALUES(age_max),
    height_min = VALUES(height_min),
    height_max = VALUES(height_max),
    income_min = VALUES(income_min);

-- 希望都道府県 (dating_preference_prefectures)
INSERT INTO dating_preference_prefectures (user_id, prefecture) VALUES
('user1', 'Tokyo'),
('user1', 'Kanagawa'),
('user2', 'Tokyo'),
('user2', 'Saitama'),
('user3', 'Osaka'),
('user3', 'Kyoto')
ON DUPLICATE KEY UPDATE
    prefecture = VALUES(prefecture);
