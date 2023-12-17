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