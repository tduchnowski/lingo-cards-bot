CREATE TABLE words (
  id bigserial PRIMARY KEY,
  word VARCHAR(50) NOT NULL,
  lang_code VARCHAR(5) NOT NULL,
  language VARCHAR(20) NOT NULL,
  meaning TEXT,
  usage TEXT,
  part_of_speech VARCHAR(10),
  frequency INTEGER CHECK (frequency >= 0),
  UNIQUE (word, lang_code),
  CONSTRAINT not_empty_strings CHECK (word <> '' AND lang_code <> '' AND language <> '')
);

-- indexes
CREATE INDEX idx_lang_code ON words(lang_code);
CREATE INDEX idx_language ON words(language);
CREATE INDEX idx_frequency ON words(frequency);

-- import CSVs
COPY words(word, lang_code, language, frequency)
FROM '/home/wordlist.csv'
WITH (FORMAT CSV, HEADER);

-- views
CREATE VIEW words_pl AS
SELECT * FROM words
WHERE lang_code = 'pl';

CREATE VIEW words_ru AS
SELECT * FROM words
WHERE lang_code = 'ru';
