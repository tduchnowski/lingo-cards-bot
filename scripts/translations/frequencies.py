import re
from collections import namedtuple


WordFreq = namedtuple('WordFreq', ['word', 'freq'])

def extract_freqs(path):
    with open(path) as f:
        lines = f.read()
    pattern = r"^\d+\t(.+)\t(\d+)$"
    p = re.compile(pattern, re.MULTILINE)
    return [WordFreq(m[0], int(m[1])) for m in p.findall(lines)]

def is_word_invalid(word):
    # I don't want short words
    if len(word) < 4:
        return True
    if any(chr.isdigit() for chr in word):
        return True
    invalid_chars = ",.:;\'\"+={}[]\\|?/><~`!@#$%^&*_"
    if any(chr in invalid_chars for chr in word):
        return True
    return False

def process_batch(word_frequencies):
    return [wf for wf in word_frequencies if not is_word_invalid(wf.word)]

# def expand_dict(dict_to_expand, dict_to_add):
#     for k, v in dict_to_add.items():
#         dict_to_expand[k] = dict_to_expand.get(k, 0) + v
#     return dict_to_expand

# def process_multicore(frequencies, cpus=1):
#     # this is useless, even on large datasets the processing 
#     # takes couple of seconds on a single core
#     # the overhead of created processes doesn't make sense in this case
#     processed = []
#     frequencies_len = len(frequencies)
#     batch_size = frequencies_len//cpus
#     todo = {}
#     print('\rStarted processing', frequencies_len, 'words')
#     with futures.ProcessPoolExecutor(cpus) as ppe:
#         for start in range(0, frequencies_len, batch_size):
#             stop = min(frequencies_len, start + batch_size)
#             future = ppe.submit(process_batch, frequencies[start:stop])
#             todo[future] = start
#         todo_iter = futures.as_completed(todo)
#         for completed in todo_iter:
#             filtered_words = completed.result()
#             processed.append(filtered_words)
#     print("Finished creating lemmas frequencies")
#     return chain.from_iterable(processed)

# def make_lemmas(frequencies, language, cpus):
#     lemmas_freq = {}
#     frequencies_len = len(frequencies)
#     batch_size = frequencies_len//cpus
#     lemma = Lemmanizator(language)
#     todo = {}
#     remaining_batches = frequencies_len/batch_size
#     completed_batches = 0
#     print('\rStarted processing', frequencies_len, 'words')
#     with futures.ProcessPoolExecutor(cpus) as ppe:
#         for start in range(0, frequencies_len, batch_size):
#             stop = min(frequencies_len, start + batch_size)
#             future = ppe.submit(process_batch, lemma, frequencies[start:stop])
#             todo[future] = start
#         todo_iter = futures.as_completed(todo)
#         for completed in todo_iter:
#             partial_lemmas_freq = completed.result()
#             expand_dict(lemmas_freq, partial_lemmas_freq)
#             completed_batches += 1
#             remaining_batches -= 1
#     print("Finished creating lemmas frequencies")
#     return lemmas_freq
#
# class Lemmanizator:
#     LANG_PACKAGES = {
#         "Polish":"pl_core_news_sm",
#         "Russian":"ru_core_news_sm"
#     }
#
#     def __init__(self, language):
#         self.language = language
#         self.nlp = spacy.load(self.LANG_PACKAGES[language])
#
#     def __repr__(self):
#         return f"Lemmanizator({self.language})"
#
#     def process_word(self, word):
#         if len(word) < 4 or any(char.isdigit() for char in word):
#             return ''
#         doc = self.nlp(word.lower())
#         return [token.lemma_ for token in doc][0]
#
#
# def process_batch(l:Lemmanizator, word_frequencies:list[WordFreq]):
#     lemmas_dict = {}
#     for wf in word_frequencies:
#         lemma = l.process_word(wf.word)
#         if not lemma:
#             continue
#         lemmas_dict[lemma] = lemmas_dict.get(lemma, 0) + wf.freq
#     return lemmas_dict
#
