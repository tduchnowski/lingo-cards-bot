-- TABLES SETUP

-- table containing words for which there exists 
-- a wiktionary definition. this should help with
-- eliminating words that are in frequency lists
-- but don't belong to a language
CREATE TABLE wiktionary_reference (
	id BIGSERIAL PRIMARY KEY,
	word VARCHAR(100) NOT NULL,
	pos VARCHAR(15) NOT NULL,
	lang_code VARCHAR(5) NOT NULL
);

COPY wiktionary_reference(word,pos,lang_code)
FROM '/docker-entrypoint-initdb.d/wiktionary/pl_wiktionary.csv'
DELIMITER ','
CSV HEADER;

COPY wiktionary_reference(word,pos,lang_code)
FROM '/docker-entrypoint-initdb.d/wiktionary/es_wiktionary.csv'
DELIMITER ','
CSV HEADER;

COPY wiktionary_reference(word,pos,lang_code)
FROM '/docker-entrypoint-initdb.d/wiktionary/ru_wiktionary.csv'
DELIMITER ','
CSV HEADER;

COPY wiktionary_reference(word,pos,lang_code)
FROM '/docker-entrypoint-initdb.d/wiktionary/it_wiktionary.csv'
DELIMITER ','
CSV HEADER;

COPY wiktionary_reference(word,pos,lang_code)
FROM '/docker-entrypoint-initdb.d/wiktionary/pt_wiktionary.csv'
DELIMITER ','
CSV HEADER;


CREATE TABLE tmp_translations (
  id BIGSERIAL PRIMARY KEY,
  word VARCHAR(50) NOT NULL,
  lang_code VARCHAR(5) NOT NULL,
  meaning TEXT,
  lemma VARCHAR(50),
  lemma_meaning TEXT,
  sentences TEXT,
--  UNIQUE (word, lang_code),
  CONSTRAINT not_empty_strings CHECK (word <> '' AND lang_code <> '' AND lemma <> '' AND lemma_meaning <> '')
);

-- indexes
CREATE INDEX idx_lang_code ON tmp_translations(lang_code);
CREATE INDEX idx_lemma ON tmp_translations(lemma);
CREATE INDEX idx_lemma_meaning ON tmp_translations(lemma_meaning);

COPY tmp_translations(word,lang_code,meaning,lemma,lemma_meaning,sentences)
FROM '/docker-entrypoint-initdb.d/dictionaries/pl_dictionary.csv'
DELIMITER '|'
CSV HEADER;

COPY tmp_translations(word,lang_code,meaning,lemma,lemma_meaning,sentences)
FROM '/docker-entrypoint-initdb.d/dictionaries/ru_dictionary.csv'
DELIMITER '|'
CSV HEADER;

COPY tmp_translations(word,lang_code,meaning,lemma,lemma_meaning,sentences)
FROM '/docker-entrypoint-initdb.d/dictionaries/es_dictionary.csv'
DELIMITER '|'
CSV HEADER;

COPY tmp_translations(word,lang_code,meaning,lemma,lemma_meaning,sentences)
FROM '/docker-entrypoint-initdb.d/dictionaries/pt_dictionary.csv'
DELIMITER '|'
CSV HEADER;

COPY tmp_translations(word,lang_code,meaning,lemma,lemma_meaning,sentences)
FROM '/docker-entrypoint-initdb.d/dictionaries/it_dictionary.csv'
DELIMITER '|'
CSV HEADER;

-- delete names, surnames and useless shit from the database
DELETE FROM tmp_translations
WHERE
  lemma_meaning LIKE '%a name%'
  OR lemma_meaning LIKE '%common name%'
  OR lemma_meaning LIKE '%given name%'
  OR lemma_meaning LIKE '%personal name%'
  OR lemma_meaning like '%male name%'
  OR lemma_meaning LIKE '%proper noun%'
  OR lemma_meaning LIKE '%name)%'
  OR lemma_meaning LIKE '%(name)%'
  OR lemma_meaning LIKE '%a surname%'
  OR lemma_meaning LIKE '%same as original%';

CREATE TABLE frequencies (
  id BIGSERIAL PRIMARY KEY,
  word VARCHAR(50),
  lang_code VARCHAR(5),
  frequency BIGINT
);

-- import frequencies for words in all available languages
COPY frequencies(word,frequency,lang_code)
FROM '/docker-entrypoint-initdb.d/frequencies/pl_frequencies.csv'
DELIMITER ','
CSV HEADER;

COPY frequencies(word,frequency,lang_code)
FROM '/docker-entrypoint-initdb.d/frequencies/ru_frequencies.csv'
DELIMITER ','
CSV HEADER;

COPY frequencies(word,frequency,lang_code)
FROM '/docker-entrypoint-initdb.d/frequencies/es_frequencies.csv'
DELIMITER ','
CSV HEADER;

COPY frequencies(word,frequency,lang_code)
FROM '/docker-entrypoint-initdb.d/frequencies/pt_frequencies.csv'
DELIMITER ','
CSV HEADER;

COPY frequencies(word,frequency,lang_code)
FROM '/docker-entrypoint-initdb.d/frequencies/it_frequencies.csv'
DELIMITER ','
CSV HEADER;

-- delete from tmp_translations where lemma is null;
-- delete from tmp_translations where lemma_meaning like '%a name%';
-- delete from tmp_translations where lemma_meaning like '%personal name%'
-- delete from tmp_translations where lemma_meaning like '%given name%'
-- delete from tmp_translations where lemma_meaning like '%male name%'
-- delete from tmp_translations where lemma_meaning like '%nickname%'
-- delete from tmp_translations where lemma_meaning like '%name)%'
-- delete from tmp_translations where lemma_meaning like '% surname%'
-- delete from tmp_translations where lemma_meaning like '%proper noun%'

-- translations + frequency
CREATE VIEW translations AS (
  SELECT tmp_translations.id, tmp_translations.lang_code, tmp_translations.word, meaning, lemma, lemma_meaning, sentences, frequency FROM frequencies
  JOIN tmp_translations
  ON tmp_translations.word=frequencies.word AND tmp_translations.lang_code=frequencies.lang_code
);

CREATE VIEW lemmas AS 
(
  WITH tmp_lemmas AS 
  (
	  SELECT MIN(id) AS id, lang_code, lemma, STRING_AGG(sentences, '') AS sentences, SUM(frequency) AS total_freq FROM translations
	  GROUP BY lemma, lang_code
  )
  SELECT tmp_lemmas.id, tmp_lemmas.lang_code, tmp_lemmas.lemma, lemma_meaning, tmp_lemmas.sentences, tmp_lemmas.total_freq 
  FROM
  translations JOIN tmp_lemmas ON translations.id=tmp_lemmas.id
  ORDER BY total_freq DESC
);

-- further cleaning, removal of the words that are unlikely to be actual words in a given language
create view lemmas_clean as (
	select distinct lemmas.lang_code, lemma, lemma_meaning, sentences, total_freq from lemmas
	join wiktionary_reference
	on lemmas.lemma=wiktionary_reference.word and lemmas.lang_code=wiktionary_reference.lang_code
);

create view lemmas_percentiles_per_lang as (
	select
	lang_code,
	percentile_cont(0.75) within group (order by total_freq) as percentile_75,
	percentile_cont(0.25) within group (order by total_freq) as percentile_25
	from lemmas_clean
	group by lang_code
);

-- POLISH
create materialized view words_pl_0 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pl' and total_freq > percentile_75
);
create materialized view words_pl_1 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pl' and total_freq <= percentile_75 and total_freq >= percentile_25
);
create materialized view words_pl_2 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pl' and total_freq < percentile_25
);

-- RUSSIAN
create materialized view words_ru_0 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='ru' and total_freq > percentile_75
);
create materialized view words_ru_1 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='ru' and total_freq <= percentile_75 and total_freq >= percentile_25
);
create materialized view words_ru_2 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='ru' and total_freq < percentile_25
);

-- SPANISH
create materialized view words_es_0 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='es' and total_freq > percentile_75
);
create materialized view words_es_1 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='es' and total_freq <= percentile_75 and total_freq >= percentile_25
);
create materialized view words_es_2 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='es' and total_freq < percentile_25
);

-- PORTUGUESE
create materialized view words_pt_0 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pt' and total_freq > percentile_75
);
create materialized view words_pt_1 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pt' and total_freq <= percentile_75 and total_freq >= percentile_25
);
create materialized view words_pt_2 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='pt' and total_freq < percentile_25
);

-- ITALIAN
create materialized view words_it_0 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='it' and total_freq > percentile_75
);
create materialized view words_it_1 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='it' and total_freq <= percentile_75 and total_freq >= percentile_25
);
create materialized view words_it_2 as (
	select lemma, lemma_meaning, sentences
	from lemmas_percentiles_per_lang
	join lemmas_clean
	on lemmas_clean.lang_code=lemmas_percentiles_per_lang.lang_code
	where lemmas_clean.lang_code='it' and total_freq < percentile_25
);

create materialized view language_codes as
(
	select distinct lang_code from lemmas_clean
	order by lang_code
);

-- TELEGRAM USERS TABLE SETUP
CREATE TABLE private_chats (
  username VARCHAR(100) PRIMARY KEY, -- telegram username if its a private chat
  chat_id BIGINT -- telegram chat id
);
