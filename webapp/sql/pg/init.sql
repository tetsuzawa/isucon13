TRUNCATE TABLE themes;
TRUNCATE TABLE icons;
TRUNCATE TABLE reservation_slots;
TRUNCATE TABLE livestream_viewers_history;
TRUNCATE TABLE livecomment_reports;
TRUNCATE TABLE ng_words;
TRUNCATE TABLE reactions;
TRUNCATE TABLE tags;
TRUNCATE TABLE livestream_tags;
TRUNCATE TABLE livecomments;
TRUNCATE TABLE livestreams;
TRUNCATE TABLE users;

ALTER SEQUENCE icons_id_seq RESTART WITH 1;
ALTER SEQUENCE reservation_slots_id_seq RESTART WITH 8760;
ALTER SEQUENCE livestream_tags_id_seq RESTART WITH 10967;
ALTER SEQUENCE livestream_viewers_history_id_seq RESTART WITH 1;
ALTER SEQUENCE livecomment_reports_id_seq RESTART WITH 1;
ALTER SEQUENCE ng_words_id_seq RESTART WITH 14338;
ALTER SEQUENCE reactions_id_seq RESTART WITH 1002;
ALTER SEQUENCE tags_id_seq RESTART WITH 1;
ALTER SEQUENCE livecomments_id_seq RESTART WITH 1002;
ALTER SEQUENCE livestreams_id_seq RESTART WITH 7496;
ALTER SEQUENCE users_id_seq RESTART WITH 1001;
ALTER SEQUENCE themes_id_seq RESTART WITH 1001;

-- 追加テーブル
TRUNCATE TABLE icons_hash;
ALTER SEQUENCE icons_hash_id_seq RESTART WITH 1;

-- user_scoresの初期化
TRUNCATE TABLE user_scores;
-- 全ユーザのスコアを0にする
INSERT INTO user_scores
SELECT users.id,users.name,0,0 FROM users;
-- すでにスコアがあるユーザのスコアを更新する
UPDATE user_scores
SET reactions = u_reactions.reaction_count
FROM
(SELECT u.id, u.name, count(*) as reaction_count
      FROM users u
           INNER JOIN livestreams l ON l.user_id = u.id
           INNER JOIN reactions r ON l.id = r.livestream_id
      GROUP BY u.id) AS u_reactions
WHERE user_id = u_reactions.id

UPDATE user_scores
SET tip = u_tip.tips
FROM
(SELECT u.id, COALESCE(SUM(l2.tip), 0) tips
      FROM users u
           INNER JOIN livestreams l ON l.user_id = u.id
           INNER JOIN livecomments l2 ON l2.livestream_id = l.id
      GROUP BY u.id) AS u_tip
WHERE user_id = u_tip.id