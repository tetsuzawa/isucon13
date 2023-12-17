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
WHERE user_id = u_reactions.id;

UPDATE user_scores
SET tip = u_tip.tips
FROM
(SELECT u.id, COALESCE(SUM(l2.tip), 0) tips
    FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id
        INNER JOIN livecomments l2 ON l2.livestream_id = l.id
    GROUP BY u.id) AS u_tip
WHERE user_id = u_tip.id;