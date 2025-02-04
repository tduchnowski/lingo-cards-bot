import re
import spacy
from collections import namedtuple
from concurrent import futures

WordFreq = namedtuple('WordFreq', ['word', 'freq'])

class Lemmanizator:
    LANG_PACKAGES = {
        "Polish":"pl_core_news_sm",
        "Russian":"ru_core_news_sm"
    }

    def __init__(self, language):
        self.language = language
        self.nlp = spacy.load(self.LANG_PACKAGES[language])

    def __repr__(self):
        return f"Lemmanizator({self.language})"

    def process_word(self, word):
        if len(word) < 4 or any(char.isdigit() for char in word):
            return ''
        doc = self.nlp(word)
        return [token.lemma_ for token in doc][0]


def extract_freqs(path):
    with open(path) as f:
        lines = f.read()
    pattern = r"^\d+\t(.+)\t(\d+)$"
    p = re.compile(pattern, re.MULTILINE)
    return [WordFreq(m[0], int(m[1])) for m in p.findall(lines)]

def process_batch(l:Lemmanizator, word_frequencies:list[WordFreq]):
    lemmas_dict = {}
    for wf in word_frequencies:
        lemma = l.process_word(wf.word)
        if not lemma:
            continue
        lemmas_dict[lemma] = lemmas_dict.get(lemma, 0) + wf.freq
    return lemmas_dict

def expand_dict(dict_to_expand, dict_to_add):
    for k, v in dict_to_add.items():
        dict_to_expand[k] = dict_to_expand.get(k, 0) + v
    return dict_to_expand

def make_lemmas(frequencies_list_path, language, cpus):
    lemmas_freq = {}
    frequencies = extract_freqs(frequencies_list_path)
    frequencies_len = len(frequencies)
    batch_size = frequencies_len//cpus
    lemma = Lemmanizator(language)
    todo = {}
    remaining_batches = frequencies_len/batch_size
    completed_batches = 0
    print('\rStarted processing', frequencies_len, 'words')
    with futures.ProcessPoolExecutor(cpus) as ppe:
        for start in range(0, frequencies_len, batch_size):
            stop = min(frequencies_len, start + batch_size)
            future = ppe.submit(process_batch, lemma, frequencies[start:stop])
            todo[future] = start
        todo_iter = futures.as_completed(todo)
        for completed in todo_iter:
            partial_lemmas_freq = {}
            partial_lemmas_freq = completed.result()
            expand_dict(lemmas_freq, partial_lemmas_freq)
            completed_batches += 1
            remaining_batches -= 1
    print("Finished creating lemmas frequencies")
    return lemmas_freq
