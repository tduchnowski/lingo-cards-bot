import argparse
import json
import os
import time
from translations.frequencies import extract_freqs, process_batch, extract_csv
from translations.translator import Chat, WordProcessor


def translate(cfg_path):
    from concurrent.futures import ThreadPoolExecutor
    # parse config file
    with open(cfg_path,"r") as f:
        cfg = json.load(f)
    max_words_per_wordlist = cfg["max_words_per_wordlist"]
    batch_size = cfg["batch_size"]
    chat_cfg = cfg["chat"]
    chat = Chat(chat_cfg["base_url"], chat_cfg["api_key"], chat_cfg["model"])
    rate_limit_per_minute = 500
    request_counter = 0
    wordlists = cfg["wordlists"]
    for wordlist in wordlists:
        word_freqs_all = extract_csv(wordlist["wordlist_path"])
        words_len_limit = min(max_words_per_wordlist, len(word_freqs_all))
        word_freqs = word_freqs_all[:words_len_limit]
        language = wordlist["language"]
        wp = WordProcessor(chat, language, "English")
        with ThreadPoolExecutor(max_workers=200) as executor:
            futures = {}
            for i in range(0, len(word_freqs), batch_size):
                if request_counter == rate_limit_per_minute:
                    time.sleep(60)
                    request_counter = 0
                max_index = min(len(word_freqs), i + batch_size)
                batch = word_freqs[i:max_index]
                batch_words = [ wf.word for wf in batch]
                futures[i] = executor.submit(wp.translate_words, batch_words)
                request_counter += 1
        results = {idx:future.result() for idx, future in futures.items()}
        missing_words_list = get_summary(results, word_freqs)
        all_incomplpete = missing_words_list + word_freqs_all[words_len_limit:]
        save_csv(all_incomplpete, wordlist["unprocessed_words_path"])
        save_translations(results, wordlist["output_path"])
        
def save_translations(tasks_results, output):
    """
    returns one json out of all the jsons returned by translator
    """
    print(f"Saving results to {output}")
    all_translations = {"words":[]}
    for _, response in tasks_results.items():
        if response:
            try:
                translations = response[0]["words"]
                all_translations["words"].extend(translations)
            except Exception as e:
                print(f"Error during saving a batch, skipping. Error: {e}")
    if all_translations["words"]:
        with open(output, "w") as f:
            json.dump(all_translations, f, ensure_ascii=False)
    print("Translations saved")


def get_summary(tasks_results, word_freqs_all):
    """
    calculates which entries from word_freqs_all are missing
    """
    failed_word_freqs = []
    for batch_start, result in tasks_results.items():
        missing = [ word_freqs_all[batch_start + missing_word_idx] for _, missing_word_idx in result[2].items() ]
        failed_word_freqs.extend(missing)
    return failed_word_freqs

def validate(path, destination, lang_code):
    frequencies = extract_freqs(path)
    filtered = process_batch(frequencies)    
    save_csv(filtered, destination)
    # save csv with a lang_code information (for database)
    directory = os.path.dirname(destination)
    frequencies_path = os.path.join(directory, f"{lang_code}_frequencies.csv")
    save_csv(filtered, frequencies_path, lang_code)

def save_csv(word_freqs, target_path, lang_code=None):
    print(f'Saving to {target_path}')
    if lang_code:
        header = "word,frequency,lang_code\n"
        lines = [f"{wf.word},{wf.freq},{lang_code}" for wf in word_freqs]
    else:
        header = "word,frequency\n"
        lines = [f"{wf.word},{wf.freq}" for wf in word_freqs]
    with open(target_path, 'w') as f:
        f.write(header)
        f.write('\n'.join(lines))
    print('Done')

def main():
    # cmdline arguments setup
    parser = argparse.ArgumentParser(description="Helper tool for translating or generating word frequency lists")
    subparsers = parser.add_subparsers(dest="command", help="List of commands, [command] -h prints the arguments for the command")
    # translate command
    translate_cmd = subparsers.add_parser('translate', help="translate word lists specified in a .json file")
    translate_cmd.add_argument("config", type=str, help="path to a config file containing chat and wordlists information")
    # lemmas command
    lemmas_cmd = subparsers.add_parser('validate', help="transform words into lemmas")
    lemmas_cmd.add_argument('wordlist_path', type=str, help='path to a file containing list of words')
    lemmas_cmd.add_argument('destination', type=str, help='path of a file where a result is going to be saved')
    lemmas_cmd.add_argument('lang_code', type=str, help='language code of a wordlist')
    lemmas_cmd.add_argument('--cpus', type=int, help='number of cpus used for generating lemmas')
    args = parser.parse_args()
    if args.command == 'translate':
        translate(args.config)
    elif args.command == 'validate':
        validate(args.wordlist_path, args.destination, args.lang_code)
    else:
        print("Type -h for available commands.")


if __name__ == "__main__":
    main()
