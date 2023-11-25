-- ユーザ (配信者、視聴者)
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  display_name VARCHAR(255) NOT NULL,
  password VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  UNIQUE (name)
);

-- プロフィール画像
CREATE TABLE icons (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  image BYTEA NOT NULL
);

-- ユーザごとのカスタムテーマ
CREATE TABLE themes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  dark_mode BOOLEAN NOT NULL
);

-- ライブ配信
CREATE TABLE livestreams (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  playlist_url VARCHAR(255) NOT NULL,
  thumbnail_url VARCHAR(255) NOT NULL,
  start_at BIGINT NOT NULL,
  end_at BIGINT NOT NULL
);

-- ライブ配信予約枠
CREATE TABLE reservation_slots (
  id BIGSERIAL PRIMARY KEY,
  slot BIGINT NOT NULL,
  start_at BIGINT NOT NULL,
  end_at BIGINT NOT NULL
);

-- ライブストリームに付与される、サービスで定義されたタグ
CREATE TABLE tags (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  UNIQUE (name)
);

-- ライブ配信とタグの中間テーブル
CREATE TABLE livestream_tags (
  id BIGSERIAL PRIMARY KEY,
  livestream_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL
);

-- ライブ配信視聴履歴
CREATE TABLE livestream_viewers_history (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  livestream_id BIGINT NOT NULL,
  created_at BIGINT NOT NULL
);

-- ライブ配信に対するライブコメント
CREATE TABLE livecomments (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  livestream_id BIGINT NOT NULL,
  comment VARCHAR(255) NOT NULL,
  tip BIGINT NOT NULL DEFAULT 0,
  created_at BIGINT NOT NULL
);

-- ユーザからのライブコメントのスパム報告
CREATE TABLE livecomment_reports (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  livestream_id BIGINT NOT NULL,
  livecomment_id BIGINT NOT NULL,
  created_at BIGINT NOT NULL
);

-- 配信者からのNGワード登録
CREATE TABLE ng_words (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  livestream_id BIGINT NOT NULL,
  word VARCHAR(255) NOT NULL,
  created_at BIGINT NOT NULL
);
CREATE INDEX ng_words_word ON ng_words(word);

-- ライブ配信に対するリアクション
CREATE TABLE reactions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  livestream_id BIGINT NOT NULL,
  emoji_name VARCHAR(255) NOT NULL,
  created_at BIGINT NOT NULL
);