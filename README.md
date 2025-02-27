# lingo-cards-bot

https://t.me/LingoCardsBot

The bot's purpose is to serve you word cards with translations and examples of sentences in which the word can appear. I made an effort for all the words to be in their root form (lemma). After you pick a language and frequency (most frequent, common, or less frequent out of the words in a bot's database) with a /menu command, the bot will start serving you random word cards for a given language and frequency, with a translation and examples hidden (through Telegram spoiler markings).

The frequency lists I used where based on the words taken from OpenSubtitles.org. Since those lists take words as they appear in a source text, they contain many words not in their root form (conjugated verbs etc.) or words that don't add much value for learning (peoples' names for example). I created a lemma database and examples with Chat GPT (gpt-4o-mini model). For every lemma, its count is a sum of counts of the words that have that lemma as its root.

So far the database contains 5 languages: Polish, Russian, Spanish, Poruguese and Italian. Each of them have around 10-13k lemmas in it. These lemmas are categorized into 3 groups: most frequent words (75th percentile and higher), pretty common (the middle ground - from 25th to 75th percentile), and less frequent words (below 25th percentile).

TODOs:
  - testing
  - rewriting the scripts for generating translations and lemmas in Go
  - maybe the ability for users to save word cards for later review
  - add more languages
